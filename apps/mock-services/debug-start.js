#!/usr/bin/env node

/**
 * Debug启动脚本
 * 用于调试mock服务启动问题
 */

console.log('🔍 开始调试启动过程...');
console.log('Node.js版本:', process.version);
console.log('命令行参数:', process.argv);

try {
  // 导入主模块
  console.log('📦 导入模块...');
  const { createServiceInstance, startService } = await import('./src/index.js');
  console.log('✓ 模块导入成功');

  // 检查服务配置
  console.log('📋 检查服务配置...');
  const { SERVICES, getServiceConfig } = await import('./src/config/services.js');
  console.log('✓ 可用服务:', Object.keys(SERVICES));

  // 获取gateway配置
  const serviceName = process.argv[3] || 'gateway';
  console.log('🎯 目标服务:', serviceName);
  
  const config = getServiceConfig(serviceName);
  if (!config) {
    throw new Error(`服务 ${serviceName} 配置不存在`);
  }
  console.log('✓ 服务配置:', config);

  // 尝试启动服务
  console.log('🚀 启动服务...');
  console.log('启动参数:', { serviceName, port: config.port });
  
  // 手动创建简单的Express服务器来测试
  const express = await import('express');
  const app = express.default();
  
  app.get('/health', (req, res) => {
    res.json({ status: 'healthy', service: serviceName, timestamp: new Date().toISOString() });
  });
  
  app.get('/', (req, res) => {
    res.json({ message: `${serviceName} service is running`, timestamp: new Date().toISOString() });
  });
  
  const server = app.listen(config.port, () => {
    console.log(`✅ ${serviceName} 服务启动成功!`);
    console.log(`   端口: ${config.port}`);
    console.log(`   健康检查: http://localhost:${config.port}/health`);
    console.log(`   主页: http://localhost:${config.port}/`);
    console.log('');
    console.log('🎉 启动完成! 按 Ctrl+C 停止服务');
  });
  
  // 优雅关闭
  process.on('SIGINT', () => {
    console.log('\n🛑 正在关闭服务...');
    server.close(() => {
      console.log('✅ 服务已关闭');
      process.exit(0);
    });
  });

} catch (error) {
  console.error('❌ 启动失败:', error.message);
  console.error('🔍 错误详情:', error.stack);
  process.exit(1);
}