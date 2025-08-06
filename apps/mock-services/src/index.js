/**
 * Mock 服务主入口
 * 统一启动所有微服务的 Mock 实例
 */

import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import { WebSocketServer } from 'ws';
import http from 'http';
// import morgan from 'morgan';  // 暂时注释掉，可能导致阻塞
import { DEFAULT_CONFIG, SERVICES, getEnabledServices } from './config/services.js';
import { formatResponse, simulateDelay, simulateErrors, requestLogger } from './middleware/response.js';
import { createServiceRouter, createHealthRouter } from './router.js';
import { logStartup, logShutdown, setLogLevel, createLogger } from './utils/logger.js';
import { verifyToken } from './middleware/auth.js';
import * as letterApi from './api/write/letters.js';

const logger = createLogger('main');

// 简化的日志函数
function simpleLog(message) {
  console.log(`[${new Date().toISOString()}] ${message}`);
}

// 服务实例存储
const serviceInstances = new Map();

/**
 * 创建单个服务实例
 */
function createServiceInstance(serviceName, config) {
  const app = express();
  
  // 基础中间件
  app.use(helmet({
    contentSecurityPolicy: false, // 开发环境可以关闭
    crossOriginEmbedderPolicy: false
  }));
  
  // CORS 配置
  app.use(cors(DEFAULT_CONFIG.cors));
  
  // 请求日志（简化版）
  if (DEFAULT_CONFIG.logging.requests) {
    app.use((req, res, next) => {
      simpleLog(`${req.method} ${req.path}`);
      next();
    });
  }
  
  // 请求解析
  app.use(express.json({ limit: '10mb' }));
  app.use(express.urlencoded({ extended: true, limit: '10mb' }));
  
  // 自定义中间件
  app.use(requestLogger());
  app.use(formatResponse());
  app.use(simulateDelay(serviceName));
  app.use(simulateErrors());
  
  // 健康检查路由
  app.use('/health', createHealthRouter());
  
  // 服务特定路由
  if (serviceName === 'gateway') {
    // API Gateway 路由所有服务
    setupGatewayRoutes(app);
  } else {
    // 单服务路由
    const basePath = config.basePath || '/api';
    app.use(basePath, createServiceRouter(serviceName));
  }
  
  // 404 处理
  app.use('*', (req, res) => {
    res.error(404, '接口不存在', `路径 ${req.originalUrl} 未找到`);
  });
  
  // 全局错误处理
  app.use((error, req, res, next) => {
    logger.error('全局错误处理:', error);
    
    if (res.headersSent) {
      return next(error);
    }
    
    res.error(500, '服务器内部错误', error.message);
  });
  
  return app;
}

/**
 * 设置 API Gateway 路由
 */
function setupGatewayRoutes(app) {
  // 公开路由 - 不需要认证，必须在其他路由之前定义
  app.get('/api/v1/letters/public', (req, res, next) => {
    // 添加响应格式化中间件到请求对象
    formatResponse()(req, res, () => {
      letterApi.getPublicLetters(req, res, next);
    });
  });
  
  // 认证相关路由
  app.use('/api/auth', createServiceRouter('auth'));
  
  // 各微服务路由
  app.use('/api/write', createServiceRouter('write-service'));
  app.use('/api/courier', createServiceRouter('courier-service'));
  app.use('/api/admin', createServiceRouter('admin-service'));
  app.use('/api/users', createServiceRouter('main-backend'));
  app.use('/api/ocr', createServiceRouter('ocr-service'));
  
  // 代理路由到其他服务（如果需要）
  app.use('/api/v1', createServiceRouter('main-backend'));
  
  logger.info('API Gateway 路由设置完成');
}

/**
 * 设置WebSocket服务器
 */
function setupWebSocketServer(server) {
  const wss = new WebSocketServer({ 
    server,
    path: '/ws'
  });
  
  logger.info('WebSocket服务器已设置，路径: /ws');
  
  wss.on('connection', (ws, req) => {
    logger.info('新的WebSocket连接');
    
    // 从查询参数获取token
    const url = new URL(req.url, 'http://localhost');
    const token = url.searchParams.get('token');
    
    if (!token) {
      logger.warn('WebSocket连接缺少token，关闭连接');
      ws.close(1008, 'Missing token');
      return;
    }
    
    // 验证token
    const decoded = verifyToken(token);
    if (!decoded) {
      logger.warn('WebSocket token验证失败，关闭连接');
      ws.close(1008, 'Invalid token');
      return;
    }
    
    logger.info(`WebSocket用户 ${decoded.username} 连接成功`);
    
    // 保存用户信息到WebSocket连接
    ws.user = decoded;
    
    // 发送欢迎消息
    ws.send(JSON.stringify({
      type: 'welcome',
      message: '连接成功',
      user: {
        id: decoded.id,
        username: decoded.username,
        role: decoded.role
      },
      timestamp: new Date().toISOString()
    }));
    
    // 处理消息
    ws.on('message', (data) => {
      try {
        const message = JSON.parse(data.toString());
        logger.debug('收到WebSocket消息:', message);
        
        // 处理不同类型的消息
        switch (message.type) {
          case 'ping':
            ws.send(JSON.stringify({
              type: 'pong',
              timestamp: new Date().toISOString()
            }));
            break;
            
          case 'task_update':
            // 模拟任务更新广播
            broadcastToRole('courier', {
              type: 'task_notification',
              data: message.data,
              timestamp: new Date().toISOString()
            });
            break;
            
          default:
            logger.warn('未知的WebSocket消息类型:', message.type);
        }
      } catch (error) {
        logger.error('处理WebSocket消息错误:', error);
      }
    });
    
    // 连接关闭处理
    ws.on('close', (code, reason) => {
      logger.info(`WebSocket用户 ${decoded.username} 断开连接: ${code} ${reason}`);
    });
    
    // 错误处理
    ws.on('error', (error) => {
      logger.error('WebSocket错误:', error);
    });
  });
  
  // 广播消息到指定角色的所有连接
  function broadcastToRole(role, message) {
    wss.clients.forEach((client) => {
      if (client.readyState === client.OPEN && client.user && client.user.role === role) {
        client.send(JSON.stringify(message));
      }
    });
  }
  
  // 广播消息到所有连接
  function broadcastToAll(message) {
    wss.clients.forEach((client) => {
      if (client.readyState === client.OPEN) {
        client.send(JSON.stringify(message));
      }
    });
  }
  
  return wss;
}

/**
 * 启动单个服务
 */
async function startService(serviceName, config) {
  try {
    const app = createServiceInstance(serviceName, config);
    const port = config.port;
    
    // 创建HTTP服务器
    const server = http.createServer(app);
    
    // 如果是gateway服务，添加WebSocket支持
    let wss = null;
    if (serviceName === 'gateway') {
      wss = setupWebSocketServer(server);
    }
    
    server.listen(port, () => {
      console.log(`✅ ${serviceName} 服务启动成功! 端口: ${port}`);
      console.log(`   健康检查: http://localhost:${port}/health`);
      console.log(`   服务地址: http://localhost:${port}/`);
      if (wss) {
        console.log(`   WebSocket: ws://localhost:${port}/ws`);
      }
    });
    
    // 优雅关闭处理
    server.on('close', () => {
      if (wss) {
        wss.close();
      }
      logShutdown(serviceName);
    });
    
    serviceInstances.set(serviceName, { app, server, port, wss });
    
    return server;
    
  } catch (error) {
    logger.error(`启动 ${serviceName} 服务失败:`, error);
    throw error;
  }
}

/**
 * 启动所有服务
 */
async function startAllServices() {
  try {
    logger.info('🚀 OpenPenPal Mock Services 启动中...');
    
    // 设置日志级别
    setLogLevel(DEFAULT_CONFIG.logging.level);
    
    const enabledServices = getEnabledServices();
    const startPromises = [];
    
    for (const [serviceName, config] of Object.entries(enabledServices)) {
      startPromises.push(startService(serviceName, config));
    }
    
    await Promise.all(startPromises);
    
    logger.info('');
    logger.info('🎉 所有 Mock 服务启动完成！');
    logger.info('');
    logger.info('📋 服务列表:');
    
    for (const [serviceName, instance] of serviceInstances) {
      const config = enabledServices[serviceName];
      logger.info(`   • ${config.name}: http://localhost:${instance.port}`);
    }
    
    logger.info('');
    logger.info('📖 快速开始:');
    logger.info('   1. 登录获取 token: POST http://localhost:8000/api/auth/login');
    logger.info('   2. 创建信件: POST http://localhost:8000/api/write/letters');
    logger.info('   3. 查看任务: GET http://localhost:8000/api/courier/tasks');
    logger.info('');
    logger.info('🔧 测试工具:');
    logger.info('   • npm run test:permissions - 运行权限测试');
    logger.info('   • npm test - 运行所有测试');
    logger.info('');
    
  } catch (error) {
    logger.error('启动服务失败:', error);
    process.exit(1);
  }
}

/**
 * 优雅关闭所有服务
 */
async function shutdownAllServices() {
  logger.info('🛑 正在关闭所有 Mock 服务...');
  
  const shutdownPromises = [];
  
  for (const [serviceName, instance] of serviceInstances) {
    shutdownPromises.push(
      new Promise((resolve) => {
        instance.server.close(() => {
          logShutdown(serviceName);
          resolve();
        });
      })
    );
  }
  
  await Promise.all(shutdownPromises);
  logger.info('✅ 所有服务已安全关闭');
  process.exit(0);
}

/**
 * 信号处理
 */
process.on('SIGTERM', shutdownAllServices);
process.on('SIGINT', shutdownAllServices);

// 未捕获异常处理
process.on('uncaughtException', (error) => {
  logger.error('未捕获异常:', error);
  shutdownAllServices();
});

process.on('unhandledRejection', (reason, promise) => {
  logger.error('未处理的 Promise 拒绝:', reason);
  shutdownAllServices();
});

/**
 * 命令行参数处理
 */
function parseArgs() {
  const args = process.argv.slice(2);
  const options = {
    services: [],
    port: null,
    logLevel: DEFAULT_CONFIG.logging.level
  };
  
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    
    switch (arg) {
      case '--service':
      case '-s':
        if (args[i + 1]) {
          options.services.push(args[i + 1]);
          i++;
        }
        break;
      case '--port':
      case '-p':
        if (args[i + 1]) {
          options.port = parseInt(args[i + 1], 10);
          i++;
        }
        break;
      case '--log-level':
      case '-l':
        if (args[i + 1]) {
          options.logLevel = args[i + 1];
          i++;
        }
        break;
      case '--help':
      case '-h':
        showHelp();
        process.exit(0);
        break;
    }
  }
  
  return options;
}

/**
 * 显示帮助信息
 */
function showHelp() {
  console.log(`
OpenPenPal Mock Services

用法: node src/index.js [选项]

选项:
  -s, --service <name>    启动指定服务 (可重复使用)
  -p, --port <number>     指定端口 (仅单服务模式)
  -l, --log-level <level> 设置日志级别 (debug|info|warn|error)
  -h, --help              显示帮助信息

示例:
  node src/index.js                           # 启动所有服务
  node src/index.js -s gateway                # 只启动 API Gateway
  node src/index.js -s write-service -s courier-service  # 启动指定服务
  node src/index.js -s gateway -p 3000        # 启动 Gateway 并指定端口

可用服务:
  ${Object.keys(SERVICES).map(name => `  • ${name}`).join('\n')}
`);
}

/**
 * 主函数
 */
async function main() {
  const options = parseArgs();
  
  // 设置日志级别
  setLogLevel(options.logLevel);
  
  // 如果指定了特定服务，只启动这些服务
  if (options.services.length > 0) {
    for (const serviceName of options.services) {
      if (!SERVICES[serviceName]) {
        logger.error(`未知服务: ${serviceName}`);
        process.exit(1);
      }
      
      const config = { ...SERVICES[serviceName] };
      
      // 如果指定了端口且只有一个服务，使用指定端口
      if (options.port && options.services.length === 1) {
        config.port = options.port;
      }
      
      await startService(serviceName, config);
    }
  } else {
    // 启动所有服务
    await startAllServices();
  }
}

// 启动应用
if (import.meta.url.endsWith('index.js') || process.argv[1].endsWith('index.js')) {
  main().catch((error) => {
    logger.error('应用启动失败:', error);
    process.exit(1);
  });
}