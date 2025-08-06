#!/usr/bin/env node

/**
 * 修复版Mock服务启动脚本
 * 移除可能导致阻塞的组件
 */

import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import morgan from 'morgan';
import { DEFAULT_CONFIG, SERVICES } from './src/config/services.js';
import { formatResponse, simulateDelay, simulateErrors, requestLogger } from './src/middleware/response.js';
import { createServiceRouter, createHealthRouter } from './src/router.js';

// 简化的日志函数
function log(level, message) {
  const timestamp = new Date().toISOString();
  console.log(`[${timestamp}] [${level.toUpperCase()}] ${message}`);
}

/**
 * 创建服务实例（简化版）
 */
function createServiceInstance(serviceName, config) {
  log('info', `创建 ${serviceName} 服务实例...`);
  
  const app = express();
  
  // 基础中间件
  app.use(helmet({
    contentSecurityPolicy: false,
    crossOriginEmbedderPolicy: false
  }));
  
  app.use(cors(DEFAULT_CONFIG.cors));
  
  // 请求解析
  app.use(express.json({ limit: '10mb' }));
  app.use(express.urlencoded({ extended: true, limit: '10mb' }));
  
  // 自定义中间件（简化版）
  app.use((req, res, next) => {
    log('debug', `${req.method} ${req.path}`);
    next();
  });
  
  app.use(formatResponse());
  
  // 健康检查路由
  app.use('/health', createHealthRouter());
  
  // 服务特定路由
  if (serviceName === 'gateway') {
    log('info', '设置Gateway路由...');
    // 简化的Gateway路由
    app.use('/api/auth', createServiceRouter('auth'));
    app.use('/api/write', createServiceRouter('write-service'));
    app.use('/api/courier', createServiceRouter('courier-service'));
    app.use('/api/admin', createServiceRouter('admin-service'));
    app.use('/api/users', createServiceRouter('main-backend'));
    log('info', 'Gateway路由设置完成');
  } else {
    // 单服务路由
    const basePath = config.basePath || '/api';
    app.use(basePath, createServiceRouter(serviceName));
  }
  
  // 默认路由
  app.get('/', (req, res) => {
    res.json({
      service: serviceName,
      status: 'running',
      timestamp: new Date().toISOString(),
      version: '1.0.0'
    });
  });
  
  log('info', `${serviceName} 服务实例创建完成`);
  return app;
}

/**
 * 启动单个服务（简化版）
 */
async function startService(serviceName, config) {
  return new Promise((resolve, reject) => {
    try {
      log('info', `正在启动 ${serviceName} 服务...`);
      
      const app = createServiceInstance(serviceName, config);
      const port = config.port;
      
      const server = app.listen(port, () => {
        log('info', `✅ ${serviceName} 服务启动成功!`);
        log('info', `   端口: ${port}`);
        log('info', `   健康检查: http://localhost:${port}/health`);
        log('info', `   主页: http://localhost:${port}/`);
        resolve(server);
      });
      
      server.on('error', (error) => {
        log('error', `${serviceName} 服务启动失败: ${error.message}`);
        reject(error);
      });
      
    } catch (error) {
      log('error', `创建 ${serviceName} 服务失败: ${error.message}`);
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
    port: null
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
    }
  }
  
  return options;
}

/**
 * 主函数
 */
async function main() {
  try {
    log('info', '🚀 OpenPenPal Mock Services (修复版) 启动中...');
    
    const options = parseArgs();
    log('info', `启动选项: ${JSON.stringify(options)}`);
    
    if (options.services.length > 0) {
      log('info', `启动指定服务: ${options.services.join(', ')}`);
      
      for (const serviceName of options.services) {
        if (!SERVICES[serviceName]) {
          throw new Error(`未知服务: ${serviceName}`);
        }
        
        const config = { ...SERVICES[serviceName] };
        if (options.port && options.services.length === 1) {
          config.port = options.port;
        }
        
        await startService(serviceName, config);
        log('info', `${serviceName} 启动完成，等待下一个服务...`);
      }
    } else {
      log('info', '启动所有服务...');
      // 启动所有服务
      const serviceNames = Object.keys(SERVICES).filter(name => SERVICES[name].enabled);
      for (const serviceName of serviceNames) {
        await startService(serviceName, SERVICES[serviceName]);
      }
    }
    
    log('info', '🎉 所有服务启动完成!');
    
    // 优雅关闭处理
    process.on('SIGINT', () => {
      log('info', '🛑 正在关闭服务...');
      process.exit(0);
    });
    
    process.on('SIGTERM', () => {
      log('info', '🛑 正在关闭服务...');
      process.exit(0);
    });
    
  } catch (error) {
    log('error', `启动失败: ${error.message}`);
    log('error', `错误堆栈: ${error.stack}`);
    process.exit(1);
  }
}

// 启动应用
if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}