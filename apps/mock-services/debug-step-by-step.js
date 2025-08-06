#!/usr/bin/env node

/**
 * 逐步调试Mock服务启动
 */

console.log('🔍 逐步调试Mock服务启动...');

async function step1_imports() {
  console.log('\n📦 Step 1: 测试所有导入...');
  
  try {
    console.log('  导入 express...');
    const express = await import('express');
    
    console.log('  导入 cors...');
    const cors = await import('cors');
    
    console.log('  导入 helmet...');
    const helmet = await import('helmet');
    
    console.log('  导入 morgan...');
    const morgan = await import('morgan');
    
    console.log('  导入 services config...');
    const services = await import('./src/config/services.js');
    
    console.log('  导入 middleware/response...');
    const response = await import('./src/middleware/response.js');
    
    console.log('  导入 router...');
    const router = await import('./src/router.js');
    
    console.log('  导入 utils/logger...');
    const logger = await import('./src/utils/logger.js');
    
    console.log('✅ Step 1: 所有导入成功');
    return { express, cors, helmet, morgan, services, response, router, logger };
  } catch (error) {
    console.error('❌ Step 1 失败:', error.message);
    throw error;
  }
}

async function step2_middleware(modules) {
  console.log('\n🔧 Step 2: 测试中间件创建...');
  
  try {
    const { express, cors, helmet, morgan, services, response } = modules;
    
    console.log('  创建Express应用...');
    const app = express.default();
    
    console.log('  添加helmet中间件...');
    app.use(helmet.default({
      contentSecurityPolicy: false,
      crossOriginEmbedderPolicy: false
    }));
    
    console.log('  添加CORS中间件...');
    app.use(cors.default(services.DEFAULT_CONFIG.cors));
    
    console.log('  添加morgan日志中间件...');
    if (services.DEFAULT_CONFIG.logging.requests) {
      app.use(morgan.default(services.DEFAULT_CONFIG.logging.format));
    }
    
    console.log('  添加body解析中间件...');
    app.use(express.default.json({ limit: '10mb' }));
    app.use(express.default.urlencoded({ extended: true, limit: '10mb' }));
    
    console.log('  添加自定义中间件...');
    app.use(response.requestLogger());
    app.use(response.formatResponse());
    app.use(response.simulateDelay('gateway'));
    app.use(response.simulateErrors());
    
    console.log('✅ Step 2: 中间件创建成功');
    return app;
  } catch (error) {
    console.error('❌ Step 2 失败:', error.message);
    throw error;
  }
}

async function step3_routes(app, modules) {
  console.log('\n🛤️  Step 3: 测试路由创建...');
  
  try {
    const { router } = modules;
    
    console.log('  创建健康检查路由...');
    app.use('/health', router.createHealthRouter());
    
    console.log('  创建基础路由...');
    app.get('/', (req, res) => {
      res.json({ 
        message: 'Gateway service is running',
        timestamp: new Date().toISOString(),
        service: 'gateway'
      });
    });
    
    console.log('✅ Step 3: 路由创建成功');
    return app;
  } catch (error) {
    console.error('❌ Step 3 失败:', error.message);
    throw error;
  }
}

async function step4_server(app) {
  console.log('\n🚀 Step 4: 测试服务器启动...');
  
  try {
    const port = 8000;
    
    console.log(`  启动服务器在端口 ${port}...`);
    const server = app.listen(port, () => {
      console.log(`✅ 服务器启动成功在端口 ${port}`);
      console.log(`   健康检查: http://localhost:${port}/health`);
      console.log(`   主页: http://localhost:${port}/`);
    });
    
    // 测试连接
    setTimeout(async () => {
      try {
        console.log('  测试连接...');
        const response = await fetch(`http://localhost:${port}/health`);
        const data = await response.json();
        console.log('✅ Step 4: 连接测试成功', data);
      } catch (error) {
        console.error('❌ 连接测试失败:', error.message);
      }
    }, 1000);
    
    // 优雅关闭
    process.on('SIGINT', () => {
      console.log('\n🛑 正在关闭服务器...');
      server.close(() => {
        console.log('✅ 服务器已关闭');
        process.exit(0);
      });
    });
    
    console.log('\n🎉 所有步骤完成! 按 Ctrl+C 停止服务');
    return server;
  } catch (error) {
    console.error('❌ Step 4 失败:', error.message);
    throw error;
  }
}

async function main() {
  try {
    const modules = await step1_imports();
    const app = await step2_middleware(modules);
    await step3_routes(app, modules);
    await step4_server(app);
  } catch (error) {
    console.error('\n💥 调试失败:', error.message);
    console.error('🔍 错误堆栈:', error.stack);
    process.exit(1);
  }
}

main();