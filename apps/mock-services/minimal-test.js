#!/usr/bin/env node

/**
 * 最小化测试 - 找出确切的卡住点
 */

console.log('🔍 开始最小化测试...');

async function test() {
  try {
    console.log('Step 1: 导入Express...');
    const express = await import('express');
    console.log('✅ Express导入成功');
    
    console.log('Step 2: 创建应用...');
    const app = express.default();
    console.log('✅ 应用创建成功');
    
    console.log('Step 3: 设置基础路由...');
    app.get('/', (req, res) => {
      res.json({ status: 'ok', timestamp: new Date().toISOString() });
    });
    app.get('/health', (req, res) => {
      res.json({ status: 'healthy' });
    });
    console.log('✅ 路由设置成功');
    
    console.log('Step 4: 启动服务器...');
    const port = 8000;
    
    const server = app.listen(port, () => {
      console.log(`✅ 服务器启动成功在端口 ${port}`);
      console.log(`访问: http://localhost:${port}/`);
      console.log('测试完成! 按 Ctrl+C 退出');
    });
    
    server.on('error', (error) => {
      console.error('❌ 服务器错误:', error.message);
    });
    
    // 优雅关闭
    process.on('SIGINT', () => {
      console.log('\n🛑 关闭服务器...');
      server.close(() => {
        console.log('✅ 服务器已关闭');
        process.exit(0);
      });
    });
    
  } catch (error) {
    console.error('❌ 测试失败:', error.message);
    console.error('堆栈:', error.stack);
    process.exit(1);
  }
}

test();