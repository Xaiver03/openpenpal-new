#!/usr/bin/env node

/**
 * 生产级Mock服务启动脚本
 * 专为启动系统集成优化，移除阻塞问题
 */

import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import { SERVICES } from './src/config/services.js';
import { createServiceRouter, createHealthRouter } from './src/router.js';

// 简化且稳定的日志函数
function log(level, message, data = null) {
  const timestamp = new Date().toISOString();
  const levelColor = {
    info: '\x1b[34m',    // 蓝色
    success: '\x1b[32m', // 绿色  
    warn: '\x1b[33m',    // 黄色
    error: '\x1b[31m'    // 红色
  };
  
  const color = levelColor[level] || '\x1b[0m';
  const reset = '\x1b[0m';
  
  console.log(`${color}[${timestamp}] [${level.toUpperCase()}] ${message}${reset}`);
  if (data) {
    console.log(JSON.stringify(data, null, 2));
  }
}

/**
 * 简化的CORS配置
 */
const corsOptions = {
  origin: ['http://localhost:3000', 'http://localhost:3001'],
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With']
};

/**
 * 创建服务实例（生产优化版）
 */
function createServiceInstance(serviceName, config) {
  log('info', `创建 ${serviceName} 服务实例`);
  
  const app = express();
  
  // 基础安全中间件
  app.use(helmet({
    contentSecurityPolicy: false,
    crossOriginEmbedderPolicy: false
  }));
  
  // CORS
  app.use(cors(corsOptions));
  
  // 请求解析
  app.use(express.json({ limit: '10mb' }));
  app.use(express.urlencoded({ extended: true, limit: '10mb' }));
  
  // 简单的请求日志（非阻塞）
  app.use((req, res, next) => {
    const start = Date.now();
    res.on('finish', () => {
      const duration = Date.now() - start;
      log('info', `${req.method} ${req.path} - ${res.statusCode} (${duration}ms)`);
    });
    next();
  });
  
  // 统一响应格式
  app.use((req, res, next) => {
    const originalJson = res.json;
    res.json = function(data) {
      // 如果数据已经是标准格式，直接返回
      if (data && typeof data === 'object' && 'code' in data && 'msg' in data) {
        return originalJson.call(this, data);
      }
      
      // 否则包装成标准格式
      return originalJson.call(this, {
        code: res.statusCode >= 400 ? -1 : 0,
        msg: res.statusCode >= 400 ? '请求失败' : '操作成功',
        data: data,
        timestamp: new Date().toISOString()
      });
    };
    next();
  });
  
  // 健康检查路由
  app.use('/health', createHealthRouter());
  
  // 服务特定路由
  try {
    if (serviceName === 'gateway') {
      log('info', '设置Gateway路由');
      // Gateway路由设置
      app.use('/api/auth', createServiceRouter('auth'));
      app.use('/api/write', createServiceRouter('write-service'));
      app.use('/api/courier', createServiceRouter('courier-service'));
      app.use('/api/admin', createServiceRouter('admin-service'));
      app.use('/api/users', createServiceRouter('main-backend'));
      app.use('/api/ocr', createServiceRouter('ocr-service'));
      
      // 根路径
      app.get('/', (req, res) => {
        res.json({
          service: 'API Gateway',
          status: 'running',
          version: '1.0.0',
          endpoints: [
            '/api/auth - 认证服务',
            '/api/write - 写信服务',
            '/api/courier - 信使服务',
            '/api/admin - 管理服务',
            '/api/users - 用户服务',
            '/api/ocr - OCR服务'
          ]
        });
      });
      
    } else {
      // 单服务路由
      const basePath = config.basePath || '/api';
      app.use(basePath, createServiceRouter(serviceName));
      
      app.get('/', (req, res) => {
        res.json({
          service: config.name,
          status: 'running',
          version: '1.0.0',
          basePath: basePath
        });
      });
    }
    
    log('success', `${serviceName} 路由设置完成`);
    
  } catch (error) {
    log('error', `设置 ${serviceName} 路由失败:`, error.message);
    throw error;
  }
  
  // 404处理
  app.use('*', (req, res) => {
    res.status(404).json({
      code: -1,
      msg: '接口不存在',
      data: null,
      path: req.path
    });
  });
  
  // 错误处理
  app.use((error, req, res, next) => {
    log('error', `服务错误: ${error.message}`);
    res.status(500).json({
      code: -1,
      msg: '服务器内部错误',
      data: null,
      error: process.env.NODE_ENV === 'development' ? error.message : undefined
    });
  });
  
  return app;
}

/**
 * 启动单个服务（异步非阻塞）
 */
function startService(serviceName, config) {
  return new Promise((resolve, reject) => {
    try {
      log('info', `正在启动 ${serviceName} 服务...`);
      
      const app = createServiceInstance(serviceName, config);
      const port = config.port;
      
      const server = app.listen(port, '0.0.0.0', () => {
        log('success', `✅ ${serviceName} 启动成功!`);
        log('info', `   服务名称: ${config.name}`);
        log('info', `   端口: ${port}`);
        log('info', `   健康检查: http://localhost:${port}/health`);
        log('info', `   服务地址: http://localhost:${port}/`);
        log('info', '');
        
        resolve({ server, port, serviceName });
      });
      
      server.on('error', (error) => {
        if (error.code === 'EADDRINUSE') {
          log('error', `端口 ${port} 已被占用`);
        } else {
          log('error', `${serviceName} 启动失败:`, error.message);
        }
        reject(error);
      });
      
      // 设置超时
      setTimeout(() => {
        reject(new Error(`${serviceName} 启动超时`));
      }, 10000);
      
    } catch (error) {
      log('error', `创建 ${serviceName} 服务失败:`, error.message);
      reject(error);
    }
  });
}

/**
 * 解析命令行参数
 */
function parseArgs() {
  const args = process.argv.slice(2);
  const options = {
    services: [],
    port: null,
    help: false
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
      case '--help':
      case '-h':
        options.help = true;
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
OpenPenPal Mock Services (生产版)

用法:
  node production-start.js [选项]

选项:
  --service, -s <name>   启动指定服务 (可重复)
  --port, -p <port>      指定端口 (仅单服务时有效)
  --help, -h             显示帮助信息

可用服务:
  ${Object.keys(SERVICES).map(name => `  ${name} - ${SERVICES[name].name}`).join('\n')}

示例:
  node production-start.js --service gateway
  node production-start.js --service gateway --port 8000
  node production-start.js --service write-service --service courier-service
`);
}

/**
 * 主函数
 */
async function main() {
  const options = parseArgs();
  
  if (options.help) {
    showHelp();
    return;
  }
  
  const runningServices = [];
  
  try {
    log('info', '🚀 OpenPenPal Mock Services (生产版) 启动中...');
    log('info', `Node.js版本: ${process.version}`);
    
    let servicesToStart;
    
    if (options.services.length > 0) {
      servicesToStart = options.services;
      log('info', `指定启动服务: ${servicesToStart.join(', ')}`);
    } else {
      servicesToStart = Object.keys(SERVICES).filter(name => SERVICES[name].enabled);
      log('info', `启动所有已启用服务: ${servicesToStart.join(', ')}`);
    }
    
    // 验证服务名称
    for (const serviceName of servicesToStart) {
      if (!SERVICES[serviceName]) {
        throw new Error(`未知服务: ${serviceName}`);
      }
    }
    
    // 逐个启动服务
    for (const serviceName of servicesToStart) {
      const config = { ...SERVICES[serviceName] };
      
      // 如果指定了端口且只有一个服务，使用指定端口
      if (options.port && servicesToStart.length === 1) {
        config.port = options.port;
      }
      
      const serviceInfo = await startService(serviceName, config);
      runningServices.push(serviceInfo);
      
      // 服务间启动间隔
      if (servicesToStart.length > 1) {
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    }
    
    log('success', '🎉 所有服务启动完成!');
    log('info', '');
    log('info', '📋 运行中的服务:');
    runningServices.forEach(({ serviceName, port }) => {
      const config = SERVICES[serviceName];
      log('info', `   • ${config.name}: http://localhost:${port}`);
    });
    log('info', '');
    log('info', '💡 常用操作:');
    log('info', '   • 健康检查: curl http://localhost:{port}/health');
    log('info', '   • 停止服务: Ctrl+C');
    log('info', '');
    
    // 优雅关闭处理
    const shutdown = () => {
      log('info', '🛑 正在关闭所有服务...');
      
      Promise.all(
        runningServices.map(({ server, serviceName }) => 
          new Promise(resolve => {
            server.close(() => {
              log('success', `${serviceName} 已关闭`);
              resolve();
            });
          })
        )
      ).then(() => {
        log('success', '✅ 所有服务已安全关闭');
        process.exit(0);
      });
    };
    
    process.on('SIGINT', shutdown);
    process.on('SIGTERM', shutdown);
    
  } catch (error) {
    log('error', `启动失败: ${error.message}`);
    
    // 清理已启动的服务
    if (runningServices.length > 0) {
      log('info', '清理已启动的服务...');
      runningServices.forEach(({ server, serviceName }) => {
        server.close();
        log('info', `${serviceName} 已清理`);
      });
    }
    
    process.exit(1);
  }
}

// 启动应用
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch(error => {
    console.error('未捕获的错误:', error);
    process.exit(1);
  });
}