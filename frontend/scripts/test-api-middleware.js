#!/usr/bin/env node

/**
 * API中间件测试脚本
 * API Middleware Testing Script
 */

const colors = {
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
  reset: '\x1b[0m',
  bold: '\x1b[1m'
}

class APIMiddlewareTest {
  constructor() {
    this.baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:3000'
    this.testResults = []
  }

  log(message, color = 'white') {
    console.log(`${colors[color]}${message}${colors.reset}`)
  }

  async runTest(name, testFn) {
    this.log(`\n🧪 测试: ${name}`, 'cyan')
    try {
      const result = await testFn()
      this.testResults.push({ name, status: 'PASS', result })
      this.log(`✅ 通过: ${name}`, 'green')
      return result
    } catch (error) {
      this.testResults.push({ name, status: 'FAIL', error: error.message })
      this.log(`❌ 失败: ${name} - ${error.message}`, 'red')
      throw error
    }
  }

  async testUnauthorizedAccess() {
    const response = await fetch(`${this.baseUrl}/api/courier/me`, {
      method: 'GET'
    })

    if (response.status !== 401) {
      throw new Error(`期望状态码 401，但收到 ${response.status}`)
    }

    const data = await response.json()
    
    if (!data.message || !data.code) {
      throw new Error('响应格式不正确，缺少统一的响应结构')
    }

    return { status: response.status, data }
  }

  async testInvalidToken() {
    const response = await fetch(`${this.baseUrl}/api/courier/me`, {
      method: 'GET',
      headers: {
        'Authorization': 'Bearer invalid_token_12345'
      }
    })

    if (response.status !== 401) {
      throw new Error(`期望状态码 401，但收到 ${response.status}`)
    }

    const data = await response.json()
    
    if (!data.message || !data.code) {
      throw new Error('响应格式不正确，缺少统一的响应结构')
    }

    return { status: response.status, data }
  }

  async testMalformedToken() {
    const response = await fetch(`${this.baseUrl}/api/courier/me`, {
      method: 'GET',
      headers: {
        'Authorization': 'InvalidHeader token123'
      }
    })

    if (response.status !== 401) {
      throw new Error(`期望状态码 401，但收到 ${response.status}`)
    }

    const data = await response.json()
    return { status: response.status, data }
  }

  async testStandardizedResponses() {
    // 测试多个端点的响应格式是否统一
    const endpoints = [
      '/api/courier/me',
      '/api/auth/me'
    ]

    const responses = []
    
    for (const endpoint of endpoints) {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: 'GET'
      })
      
      const data = await response.json()
      responses.push({ endpoint, status: response.status, data })
      
      // 检查是否有统一的响应格式
      if (!data.hasOwnProperty('code') || !data.hasOwnProperty('message')) {
        throw new Error(`端点 ${endpoint} 的响应格式不符合统一标准`)
      }
      
      // 检查是否有timestamp字段
      if (!data.timestamp) {
        throw new Error(`端点 ${endpoint} 的响应缺少timestamp字段`)
      }
    }

    return responses
  }

  async testHealthCheck() {
    // 测试公共端点是否正常工作
    try {
      const response = await fetch(`${this.baseUrl}/api/health`, {
        method: 'GET'
      })
      
      return { 
        status: response.status, 
        available: response.status < 500 
      }
    } catch (error) {
      return { 
        status: null, 
        available: false, 
        error: error.message 
      }
    }
  }

  generateReport() {
    this.log('\n' + '='.repeat(80), 'cyan')
    this.log('API中间件测试报告', 'bold')
    this.log('='.repeat(80), 'cyan')

    const passCount = this.testResults.filter(r => r.status === 'PASS').length
    const failCount = this.testResults.filter(r => r.status === 'FAIL').length

    this.log(`\n📊 测试总结:`, 'cyan')
    this.log(`   ✅ 通过: ${passCount}`, 'green')
    this.log(`   ❌ 失败: ${failCount}`, 'red')
    this.log(`   📈 成功率: ${Math.round((passCount / this.testResults.length) * 100)}%`, 'yellow')

    if (failCount === 0) {
      this.log('\n🎉 所有测试通过！API中间件工作正常', 'green')
    } else {
      this.log('\n⚠️ 发现问题需要修复:', 'red')
      this.testResults
        .filter(r => r.status === 'FAIL')
        .forEach(result => {
          this.log(`   • ${result.name}: ${result.error}`, 'red')
        })
    }

    this.log('\n📋 测试详情:', 'cyan')
    this.testResults.forEach(result => {
      const statusIcon = result.status === 'PASS' ? '✅' : '❌'
      const statusColor = result.status === 'PASS' ? 'green' : 'red'
      this.log(`   ${statusIcon} ${result.name}`, statusColor)
    })

    this.log('\n' + '='.repeat(80), 'cyan')
    
    return failCount === 0
  }

  async run() {
    this.log('🚀 开始API中间件测试...', 'cyan')
    
    try {
      // 检查服务器是否运行
      await this.runTest('服务器健康检查', () => this.testHealthCheck())

      // 测试认证中间件
      await this.runTest('未授权访问测试', () => this.testUnauthorizedAccess())
      await this.runTest('无效令牌测试', () => this.testInvalidToken())
      await this.runTest('错误令牌格式测试', () => this.testMalformedToken())
      
      // 测试响应格式标准化
      await this.runTest('标准化响应格式测试', () => this.testStandardizedResponses())

    } catch (error) {
      // 某些测试失败是正常的，继续运行其他测试
    }

    return this.generateReport()
  }
}

// 运行测试
if (require.main === module) {
  const tester = new APIMiddlewareTest()
  tester.run().then(success => {
    process.exit(success ? 0 : 1)
  }).catch(error => {
    console.error('测试运行失败:', error)
    process.exit(1)
  })
}

module.exports = APIMiddlewareTest