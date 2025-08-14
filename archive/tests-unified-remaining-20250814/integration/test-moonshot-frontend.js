#!/usr/bin/env node
/**
 * 测试前端AI组件与Moonshot API的完整调用链路
 * 
 * 这个脚本将测试：
 * 1. 前端AI组件的API调用
 * 2. Next.js API代理的工作状态
 * 3. 后端AI处理器的响应
 * 4. Moonshot API的实际调用（如果配置正确）
 */

// 使用内置fetch (Node.js 18+)
const fetch = globalThis.fetch

// 测试配置
const FRONTEND_BASE_URL = 'http://localhost:3000'
const BACKEND_BASE_URL = 'http://localhost:8080'

// 测试用户凭据
const TEST_USER = {
  username: 'admin',  // 使用已存在的admin用户
  password: 'admin123'
}

class MoonshotFrontendTester {
  constructor() {
    this.authToken = null
    this.testResults = []
  }

  async log(message, type = 'info') {
    const timestamp = new Date().toISOString()
    const prefix = type === 'error' ? '❌' : type === 'success' ? '✅' : 'ℹ️'
    console.log(`${timestamp} ${prefix} ${message}`)
  }

  async test(name, testFn) {
    try {
      await this.log(`开始测试: ${name}`)
      const result = await testFn()
      this.testResults.push({ name, success: true, result })
      await this.log(`测试通过: ${name}`, 'success')
      return result
    } catch (error) {
      this.testResults.push({ name, success: false, error: error.message })
      await this.log(`测试失败: ${name} - ${error.message}`, 'error')
      throw error
    }
  }

  // 1. 测试用户登录获取token
  async testLogin() {
    return this.test('用户登录认证', async () => {
      const response = await fetch(`${BACKEND_BASE_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(TEST_USER)
      })

      if (!response.ok) {
        throw new Error(`登录失败: ${response.status} ${response.statusText}`)
      }

      const data = await response.json()
      if (!data.success || !data.data?.token) {
        throw new Error(`登录响应无效: ${JSON.stringify(data)}`)
      }

      this.authToken = data.data.token
      return { token: this.authToken.substring(0, 20) + '...' }
    })
  }

  // 2. 测试前端API代理（通过Next.js）
  async testFrontendProxy() {
    return this.test('前端API代理测试', async () => {
      if (!this.authToken) {
        throw new Error('需要先登录获取token')
      }

      // 测试每日灵感API（现在需要认证）
      const response = await fetch(`${FRONTEND_BASE_URL}/api/ai/daily-inspiration`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.authToken}`
        }
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`代理请求失败: ${response.status} - ${errorText}`)
      }

      const data = await response.json()
      if (!data.success) {
        throw new Error(`代理响应失败: ${JSON.stringify(data)}`)
      }

      return {
        theme: data.data.theme,
        hasPrompt: !!data.data.prompt,
        hasTips: Array.isArray(data.data.tips)
      }
    })
  }

  // 3. 测试需要认证的AI功能
  async testAuthenticatedAI() {
    return this.test('认证AI功能测试', async () => {
      if (!this.authToken) {
        throw new Error('需要先登录获取token')
      }

      // 测试写作灵感生成
      const response = await fetch(`${FRONTEND_BASE_URL}/api/ai/inspiration`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.authToken}`
        },
        body: JSON.stringify({
          theme: '校园生活',
          count: 2
        })
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`认证AI请求失败: ${response.status} - ${errorText}`)
      }

      const data = await response.json()
      if (!data.success) {
        throw new Error(`认证AI响应失败: ${JSON.stringify(data)}`)
      }

      return {
        inspirationCount: data.data.inspirations?.length || 0,
        hasTheme: !!data.data.inspirations?.[0]?.theme,
        hasPrompt: !!data.data.inspirations?.[0]?.prompt
      }
    })
  }

  // 4. 测试AI人设功能
  async testAIPersonas() {
    return this.test('AI人设功能测试', async () => {
      if (!this.authToken) {
        throw new Error('需要先登录获取token')
      }

      const response = await fetch(`${FRONTEND_BASE_URL}/api/ai/personas`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.authToken}`
        }
      })

      if (!response.ok) {
        throw new Error(`人设请求失败: ${response.status} ${response.statusText}`)
      }

      const data = await response.json()
      
      // 检查响应格式，支持不同的API响应格式
      let personas, total
      if (data.success && data.data) {
        // 标准格式：{ success: true, data: { personas: [...], total: 8 } }
        personas = data.data.personas
        total = data.data.total
      } else if (data.personas) {
        // 直接格式：{ personas: [...], total: 8 }
        personas = data.personas
        total = data.total
      } else {
        throw new Error(`无效的人设响应格式: ${JSON.stringify(data)}`)
      }

      return {
        personaCount: personas?.length || 0,
        totalCount: total || 0,
        hasPoet: personas?.some(p => p.id === 'poet') || false
      }
    })
  }

  // 5. 测试AI使用统计
  async testAIStats() {
    return this.test('AI使用统计测试', async () => {
      const response = await fetch(`${FRONTEND_BASE_URL}/api/ai/stats`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': this.authToken ? `Bearer ${this.authToken}` : undefined
        }
      })

      if (!response.ok) {
        throw new Error(`统计请求失败: ${response.status} ${response.statusText}`)
      }

      const data = await response.json()
      
      // 检查响应格式，支持不同的API响应格式
      let usage, limits, remaining
      if (data.success && data.data) {
        // 标准格式：{ success: true, data: { usage: {...}, limits: {...}, remaining: {...} } }
        usage = data.data.usage
        limits = data.data.limits
        remaining = data.data.remaining
      } else if (data.usage || data.limits || data.remaining) {
        // 直接格式：{ usage: {...}, limits: {...}, remaining: {...} }
        usage = data.usage
        limits = data.limits
        remaining = data.remaining
      } else {
        throw new Error(`无效的统计响应格式: ${JSON.stringify(data)}`)
      }

      return {
        hasUsage: !!usage,
        hasLimits: !!limits,
        hasRemaining: !!remaining
      }
    })
  }

  // 6. 测试Moonshot特定功能（如果可以找到相关接口）
  async testMoonshotSpecific() {
    return this.test('Moonshot特定功能测试', async () => {
      // 尝试测试可能使用Moonshot API的功能
      // 由于我们没有直接的Moonshot端点，我们通过现有AI功能间接测试
      
      if (!this.authToken) {
        throw new Error('需要先登录获取token')
      }

      // 测试AI回信建议功能（这个可能会调用Moonshot）
      const response = await fetch(`${BACKEND_BASE_URL}/api/v1/ai/reply-advice`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.authToken}`
        },
        body: JSON.stringify({
          letter_id: 'test-letter-123',
          persona_type: 'friend',
          persona_name: '知心朋友',
          persona_desc: '一个温暖的朋友',
          relationship: '好朋友',
          delivery_days: 1
        })
      })

      // 即使失败也记录响应，因为这能告诉我们Moonshot配置的状态
      const responseText = await response.text()
      let data
      try {
        data = JSON.parse(responseText)
      } catch {
        data = { rawResponse: responseText }
      }

      return {
        status: response.status,
        ok: response.ok,
        hasData: !!data.data,
        response: data
      }
    })
  }

  // 7. 检查后端AI服务配置
  async testBackendAIConfig() {
    return this.test('后端AI配置检查', async () => {
      // 测试后端是否有AI配置端点
      const response = await fetch(`${BACKEND_BASE_URL}/api/v1/admin/ai/config`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': this.authToken ? `Bearer ${this.authToken}` : undefined
        }
      })

      // 即使未认证也要记录响应状态
      const responseText = await response.text()
      let data
      try {
        data = JSON.parse(responseText)
      } catch {
        data = { rawResponse: responseText }
      }

      return {
        status: response.status,
        hasProviders: !!data.data?.providers,
        hasMoonshot: !!data.data?.providers?.moonshot,
        hasSiliconflow: !!data.data?.providers?.siliconflow,
        configData: data.data || data
      }
    })
  }

  // 8. 测试前端组件能否正确处理API响应
  async testFrontendComponentAPI() {
    return this.test('前端组件API处理测试', async () => {
      // 模拟前端组件会发起的请求序列
      const requests = [
        {
          name: 'daily-inspiration',
          url: `${FRONTEND_BASE_URL}/api/ai/daily-inspiration`,
          method: 'GET'
        },
        {
          name: 'personas',
          url: `${FRONTEND_BASE_URL}/api/ai/personas`,
          method: 'GET'
        },
        {
          name: 'stats',
          url: `${FRONTEND_BASE_URL}/api/ai/stats`,
          method: 'GET'
        }
      ]

      const results = {}
      for (const req of requests) {
        try {
          const response = await fetch(req.url, {
            method: req.method,
            headers: {
              'Content-Type': 'application/json',
              'Authorization': this.authToken ? `Bearer ${this.authToken}` : undefined
            }
          })
          
          const data = await response.json()
          results[req.name] = {
            status: response.status,
            success: data.success,
            hasData: !!data.data
          }
        } catch (error) {
          results[req.name] = {
            status: 'error',
            error: error.message
          }
        }
      }

      return results
    })
  }

  // 运行所有测试
  async runAllTests() {
    await this.log('🚀 开始Moonshot前端集成测试')
    
    try {
      // 1. 基础认证测试
      await this.testLogin()
      
      // 2. 前端代理测试
      await this.testFrontendProxy()
      
      // 3. 认证AI功能测试
      await this.testAuthenticatedAI()
      
      // 4. AI人设测试
      await this.testAIPersonas()
      
      // 5. AI统计测试
      await this.testAIStats()
      
      // 6. Moonshot特定测试
      await this.testMoonshotSpecific()
      
      // 7. 后端配置检查
      await this.testBackendAIConfig()
      
      // 8. 前端组件API测试
      await this.testFrontendComponentAPI()
      
    } catch (error) {
      await this.log(`测试过程中发生错误: ${error.message}`, 'error')
    }

    // 输出测试总结
    await this.printSummary()
  }

  async printSummary() {
    await this.log('\n📊 测试结果总结:')
    
    const successful = this.testResults.filter(r => r.success).length
    const failed = this.testResults.filter(r => !r.success).length
    
    console.log(`✅ 成功: ${successful}`)
    console.log(`❌ 失败: ${failed}`)
    console.log(`📈 成功率: ${(successful / this.testResults.length * 100).toFixed(1)}%`)
    
    await this.log('\n📋 详细结果:')
    this.testResults.forEach(result => {
      const status = result.success ? '✅' : '❌'
      console.log(`${status} ${result.name}`)
      if (!result.success) {
        console.log(`   错误: ${result.error}`)
      }
    })

    // 输出关键发现
    await this.log('\n🔍 关键发现:')
    
    const authSuccess = this.testResults.find(r => r.name === '用户登录认证')?.success
    if (authSuccess) {
      console.log('✅ 用户认证系统正常工作')
    } else {
      console.log('❌ 用户认证存在问题')
    }

    const proxySuccess = this.testResults.find(r => r.name === '前端API代理测试')?.success
    if (proxySuccess) {
      console.log('✅ Next.js API代理正常工作')
    } else {
      console.log('❌ Next.js API代理存在问题')
    }

    const aiSuccess = this.testResults.find(r => r.name === '认证AI功能测试')?.success
    if (aiSuccess) {
      console.log('✅ 认证AI功能可以正常调用')
    } else {
      console.log('❌ 认证AI功能存在问题')
    }

    // 输出建议
    await this.log('\n💡 建议和下一步:')
    
    if (!authSuccess) {
      console.log('1. 检查后端认证服务是否运行正常')
      console.log('2. 验证测试用户账户是否存在')
    }

    if (!proxySuccess) {
      console.log('1. 检查Next.js前端服务是否在端口3000运行')
      console.log('2. 验证[...path]/route.ts代理配置是否正确')
    }

    if (!aiSuccess) {
      console.log('1. 检查后端AI服务配置')
      console.log('2. 验证Moonshot API密钥是否配置正确')
      console.log('3. 检查网络连接到Moonshot API服务器')
    }

    console.log('\n✨ 测试完成！')
  }
}

// 运行测试
async function main() {
  const tester = new MoonshotFrontendTester()
  await tester.runAllTests()
}

if (require.main === module) {
  main().catch(console.error)
}

module.exports = MoonshotFrontendTester