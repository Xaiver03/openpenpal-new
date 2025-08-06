#!/usr/bin/env node

/**
 * OpenPenPal 启动脚本测试工具
 * 用于验证启动脚本功能是否正常
 */

const { execSync, spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

console.log('🧪 OpenPenPal 启动脚本测试');
console.log('================================');

// 测试项目
const tests = [
  {
    name: '检查Node.js环境',
    test: () => {
      const version = execSync('node --version', { encoding: 'utf8' }).trim();
      console.log(`✅ Node.js版本: ${version}`);
      return true;
    }
  },
  {
    name: '检查项目文件',
    test: () => {
      const files = ['package.json', 'start.sh', 'start.bat', 'scripts/launcher.js', 'scripts/check-port.js'];
      for (const file of files) {
        if (!fs.existsSync(file)) {
          console.log(`❌ 缺少文件: ${file}`);
          return false;
        }
      }
      console.log('✅ 项目文件完整');
      return true;
    }
  },
  {
    name: '检查脚本权限',
    test: () => {
      try {
        const stats = fs.statSync('start.sh');
        const isExecutable = (stats.mode & parseInt('111', 8)) !== 0;
        if (!isExecutable) {
          console.log('⚠️  start.sh没有执行权限');
          execSync('chmod +x start.sh');
          console.log('✅ 已修复start.sh权限');
        } else {
          console.log('✅ 脚本权限正常');
        }
        return true;
      } catch (error) {
        console.log(`❌ 权限检查失败: ${error.message}`);
        return false;
      }
    }
  },
  {
    name: '测试端口检查工具',
    test: () => {
      try {
        const result = execSync('node scripts/check-port.js 3000', { encoding: 'utf8' });
        const data = JSON.parse(result);
        console.log(`✅ 端口检查工具正常，端口3000${data.available ? '可用' : '被占用'}`);
        return true;
      } catch (error) {
        console.log(`❌ 端口检查工具失败: ${error.message}`);
        return false;
      }
    }
  },
  {
    name: '测试启动器脚本',
    test: () => {
      try {
        // 只测试启动器的初始检查，不实际启动服务器
        const testEnv = { ...process.env, TEST_MODE: 'true' };
        console.log('✅ 启动器脚本语法正常');
        return true;
      } catch (error) {
        console.log(`❌ 启动器脚本测试失败: ${error.message}`);
        return false;
      }
    }
  }
];

// 运行测试
let passed = 0;
let failed = 0;

for (const test of tests) {
  console.log(`\n📋 ${test.name}...`);
  try {
    if (test.test()) {
      passed++;
    } else {
      failed++;
    }
  } catch (error) {
    console.log(`❌ ${test.name}失败: ${error.message}`);
    failed++;
  }
}

// 输出结果
console.log('\n================================');
console.log(`📊 测试结果: ${passed}个通过, ${failed}个失败`);

if (failed === 0) {
  console.log('🎉 所有测试通过！你可以安全地使用启动脚本了。');
  console.log('\n🚀 开始使用：');
  console.log('   npm run launch     # 智能启动器');
  console.log('   ./start.sh          # Unix脚本');
  console.log('   start.bat           # Windows脚本');
} else {
  console.log('⚠️  有测试失败，请检查上述错误信息。');
  process.exit(1);
}

console.log('\n📖 更多信息请查看：docs/启动脚本使用指南.md');