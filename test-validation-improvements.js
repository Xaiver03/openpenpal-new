#!/usr/bin/env node

const axios = require('axios');

// Configuration
const BASE_URL = 'http://localhost:8080';

// Colors for console output
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  cyan: '\x1b[36m',
  magenta: '\x1b[35m'
};

const log = (message, color = 'reset') => {
  const timestamp = new Date().toISOString();
  console.log(`${colors[color]}[${timestamp}] ${message}${colors.reset}`);
};

// Configure axios
const api = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
});

async function testValidationImprovements() {
  log('🔍 Testing API Validation Error Handling Improvements', 'cyan');
  log(`🌐 Target: ${BASE_URL}`, 'cyan');
  
  const testCases = [
    {
      name: '用户注册验证 - 缺少必填字段',
      endpoint: '/api/v1/auth/register',
      method: 'POST',
      data: {
        // Missing required fields: username, email, password, nickname, school_code
      },
      expectedImprovement: '应该显示具体缺少哪些字段的中文错误信息'
    },
    {
      name: '用户注册验证 - 用户名太短',
      endpoint: '/api/v1/auth/register',
      method: 'POST', 
      data: {
        username: 'ab',  // Too short (min=3)
        email: 'test@example.com',
        password: 'password123',
        nickname: 'Test User',
        school_code: 'BJDX01'
      },
      expectedImprovement: '应该显示"用户名长度不能少于3个字符"'
    },
    {
      name: '用户注册验证 - 邮箱格式错误',
      endpoint: '/api/v1/auth/register',
      method: 'POST',
      data: {
        username: 'testuser',
        email: 'invalid-email',  // Invalid email format
        password: 'password123',
        nickname: 'Test User',
        school_code: 'BJDX01'
      },
      expectedImprovement: '应该显示"请输入有效的邮箱地址"'
    },
    {
      name: '用户登录验证 - 缺少必填字段',
      endpoint: '/api/v1/auth/login',
      method: 'POST',
      data: {
        // Missing username and password
      },
      expectedImprovement: '应该显示登录信息验证失败的详细错误'
    },
    {
      name: '信件创建验证 - 无效JSON格式',
      endpoint: '/api/v1/letters/draft',
      method: 'POST',
      data: 'invalid-json-data',  // Invalid JSON
      expectedImprovement: '应该显示JSON格式错误的友好提示'
    },
    {
      name: 'AI请求验证 - 缺少参数',
      endpoint: '/api/v1/ai/match',
      method: 'POST',
      data: {
        // Missing required AI match parameters
      },
      expectedImprovement: '应该显示AI请求参数验证失败'
    }
  ];

  let improvementCount = 0;
  let totalTests = testCases.length;

  for (const testCase of testCases) {
    try {
      log(`\n📋 测试: ${testCase.name}`, 'cyan');
      log(`   期望改进: ${testCase.expectedImprovement}`, 'magenta');

      const config = { validateStatus: () => true };
      let response;

      if (testCase.method === 'POST') {
        response = await api.post(testCase.endpoint, testCase.data, config);
      } else {
        response = await api.get(testCase.endpoint, config);
      }

      // Check if response has improved validation format
      if (response.status === 400 && response.data) {
        const data = response.data;
        
        // Check for new validation response structure
        if (data.error_code === 'VALIDATION_ERROR' && data.details && Array.isArray(data.details)) {
          log(`✅ 改进成功: 新的验证错误格式`, 'green');
          log(`   错误代码: ${data.error_code}`, 'green');
          log(`   错误信息: ${data.message}`, 'green');
          log(`   字段详情: ${data.details.length} 个字段错误`, 'green');
          
          // Show field-specific errors
          data.details.forEach((detail, index) => {
            log(`     ${index + 1}. ${detail.field}: ${detail.message}`, 'green');
          });
          
          improvementCount++;
        } else if (data.success === false && data.message && data.message.includes('验证失败')) {
          log(`✅ 部分改进: 使用了中文错误信息`, 'yellow');
          log(`   错误信息: ${data.message}`, 'yellow');
          improvementCount += 0.5;
        } else {
          log(`⚠️  旧格式: 仍使用旧的验证错误格式`, 'yellow');
          log(`   响应: ${JSON.stringify(data).substring(0, 150)}...`, 'yellow');
        }
      } else {
        log(`📡 状态码: ${response.status}`, 'yellow');
        log(`   响应: ${JSON.stringify(response.data).substring(0, 100)}...`, 'yellow');
      }

    } catch (error) {
      log(`❌ 请求失败: ${error.message}`, 'red');
    }

    // Small delay between requests
    await new Promise(resolve => setTimeout(resolve, 500));
  }

  // Summary
  log(`\n📊 验证改进测试结果:`, 'cyan');
  log(`✅ 改进的端点: ${improvementCount}/${totalTests}`, improvementCount === totalTests ? 'green' : 'yellow');
  
  const improvementPercentage = (improvementCount / totalTests * 100).toFixed(1);
  log(`📈 改进率: ${improvementPercentage}%`, improvementPercentage >= 80 ? 'green' : 'yellow');
  
  if (improvementCount >= totalTests * 0.8) {
    log(`🎉 验证错误处理显著改进!`, 'green');
    log(`   • 统一的错误响应格式`, 'green');
    log(`   • 中文友好错误信息`, 'green');
    log(`   • 字段级别的详细错误`, 'green');
    log(`   • 结构化的错误数据`, 'green');
  } else if (improvementCount >= totalTests * 0.5) {
    log(`⚠️  验证错误处理有所改进，但还需要更多工作`, 'yellow');
  } else {
    log(`❌ 验证错误处理改进有限，需要继续优化`, 'red');
  }

  // Additional detailed analysis
  log(`\n🔍 详细分析:`, 'cyan');
  log(`   • 新验证系统的主要优势:`, 'cyan');
  log(`     - 统一的 error_code: "VALIDATION_ERROR"`, 'cyan');
  log(`     - 中文错误信息 (用户名不能为空, 请输入有效的邮箱地址)`, 'cyan');
  log(`     - 字段级别错误详情 (field, message, code)`, 'cyan');
  log(`     - 时间戳和请求ID支持`, 'cyan');
  log(`     - 一致的JSON响应结构`, 'cyan');
}

// Execute test
testValidationImprovements().catch(error => {
  log(`💥 测试执行失败: ${error.message}`, 'red');
  console.error(error);
  process.exit(1);
});