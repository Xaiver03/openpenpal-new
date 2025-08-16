const axios = require('axios');
const colors = require('colors/safe');

const API_URL = 'http://localhost:8080';

// 测试账号列表
const TEST_ACCOUNTS = [
  { username: 'admin', password: 'admin123', expectedRole: 'super_admin' },
  { username: 'alice', password: 'secret', expectedRole: 'user' },
  { username: 'courier_level1', password: 'secret', expectedRole: 'courier_level1' },
  { username: 'courier_level2', password: 'secret', expectedRole: 'courier_level2' },
  { username: 'courier_level3', password: 'secret', expectedRole: 'courier_level3' },
  { username: 'courier_level4', password: 'secret', expectedRole: 'courier_level4' },
];

// 统计结果
let stats = {
  total: 0,
  success: 0,
  failed: 0
};

// 打印分隔线
function printSeparator(char = '=', length = 60) {
  console.log(char.repeat(length));
}

// 测试健康检查
async function testHealth() {
  try {
    const response = await axios.get(`${API_URL}/health`);
    console.log(colors.green('✅ 健康检查通过'));
    console.log(`   服务: ${response.data.service}`);
    console.log(`   版本: ${response.data.version}`);
    console.log(`   数据库: ${response.data.database}`);
    return true;
  } catch (error) {
    console.log(colors.red('❌ 健康检查失败'));
    console.log(`   错误: ${error.message}`);
    return false;
  }
}

// 测试单个用户登录
async function testUserLogin(username, password, expectedRole) {
  stats.total++;
  
  try {
    const startTime = Date.now();
    const response = await axios.post(`${API_URL}/api/v1/auth/login`, {
      username,
      password
    });
    const endTime = Date.now();
    
    const data = response.data.data;
    const actualRole = data.user.role;
    
    if (actualRole === expectedRole) {
      stats.success++;
      console.log(colors.green(`✅ ${username.padEnd(15)} - 登录成功 (${endTime - startTime}ms)`));
      console.log(colors.gray(`   角色: ${actualRole}, Token: ${data.token.substring(0, 20)}...`));
      return { success: true, token: data.token, user: data.user };
    } else {
      stats.failed++;
      console.log(colors.yellow(`⚠️  ${username.padEnd(15)} - 角色不匹配`));
      console.log(colors.gray(`   期望: ${expectedRole}, 实际: ${actualRole}`));
      return { success: false, error: 'Role mismatch' };
    }
  } catch (error) {
    stats.failed++;
    console.log(colors.red(`❌ ${username.padEnd(15)} - 登录失败`));
    console.log(colors.gray(`   错误: ${error.response?.data?.error || error.message}`));
    return { success: false, error: error.message };
  }
}

// 测试错误密码
async function testWrongPassword() {
  console.log('\n测试错误密码处理:');
  try {
    await axios.post(`${API_URL}/api/v1/auth/login`, {
      username: 'admin',
      password: 'wrongpassword'
    });
    console.log(colors.red('❌ 错误密码测试失败 - 不应该成功'));
  } catch (error) {
    if (error.response?.status === 401) {
      console.log(colors.green('✅ 错误密码正确返回 401'));
    } else {
      console.log(colors.red('❌ 错误密码返回了意外的状态码:', error.response?.status));
    }
  }
}

// 测试使用 Token
async function testTokenUsage(token, username) {
  try {
    const response = await axios.get(`${API_URL}/api/v1/users/profile`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    console.log(colors.green(`✅ Token 验证成功 (${username})`));
    return true;
  } catch (error) {
    console.log(colors.red(`❌ Token 验证失败 (${username}): ${error.response?.status || error.message}`));
    return false;
  }
}

// 主测试函数
async function main() {
  console.log(colors.cyan.bold('\nOpenPenPal 登录系统完整测试'));
  printSeparator();
  console.log(`时间: ${new Date().toLocaleString()}`);
  console.log(`API: ${API_URL}`);
  console.log();
  
  // 1. 健康检查
  console.log(colors.yellow.bold('1. 系统健康检查'));
  printSeparator('-');
  const isHealthy = await testHealth();
  if (!isHealthy) {
    console.log(colors.red('\n⚠️  系统不健康，中止测试'));
    return;
  }
  
  // 2. 测试所有用户登录
  console.log(colors.yellow.bold('\n2. 测试用户登录'));
  printSeparator('-');
  const tokens = {};
  
  for (const account of TEST_ACCOUNTS) {
    const result = await testUserLogin(account.username, account.password, account.expectedRole);
    if (result.success) {
      tokens[account.username] = result.token;
    }
    await new Promise(r => setTimeout(r, 500)); // 避免速率限制
  }
  
  // 3. 测试错误密码
  console.log(colors.yellow.bold('\n3. 测试安全性'));
  printSeparator('-');
  await testWrongPassword();
  
  // 4. 测试 Token 使用
  console.log(colors.yellow.bold('\n4. 测试 Token 认证'));
  printSeparator('-');
  for (const [username, token] of Object.entries(tokens)) {
    await testTokenUsage(token, username);
    await new Promise(r => setTimeout(r, 500));
  }
  
  // 5. 总结
  console.log(colors.yellow.bold('\n5. 测试总结'));
  printSeparator();
  console.log(`总测试数: ${stats.total}`);
  console.log(colors.green(`成功: ${stats.success}`));
  console.log(colors.red(`失败: ${stats.failed}`));
  console.log(`成功率: ${((stats.success / stats.total) * 100).toFixed(1)}%`);
  
  if (stats.failed === 0) {
    console.log(colors.green.bold('\n🎉 所有测试通过！登录系统工作正常。'));
  } else {
    console.log(colors.red.bold(`\n⚠️  有 ${stats.failed} 个测试失败，请检查系统。`));
  }
}

// 安装 colors 包
const { exec } = require('child_process');
exec('npm list colors', (error) => {
  if (error) {
    console.log('Installing colors package...');
    exec('npm install colors', (installError) => {
      if (installError) {
        console.log('Failed to install colors package, running without colors');
        // 如果安装失败，提供一个简单的 fallback
        global.colors = {
          green: (text) => text,
          red: (text) => text,
          yellow: (text) => text,
          cyan: (text) => text,
          gray: (text) => text,
          bold: (text) => text
        };
      }
      main().catch(console.error);
    });
  } else {
    main().catch(console.error);
  }
});