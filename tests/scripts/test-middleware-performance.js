#!/usr/bin/env node

/**
 * 中间件性能测试脚本
 * Middleware Performance Testing Script
 */

const BACKEND_URL = 'http://localhost:8080'
const FRONTEND_URL = 'http://localhost:3001'

class PerformanceTest {
  constructor() {
    this.results = {
      backend: {},
      frontend: {}
    }
  }

  async getAuthToken() {
    console.log('🔑 获取认证令牌...')
    
    try {
      const response = await fetch(`${BACKEND_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: 'admin',
          password: 'admin123'
        })
      })

      if (!response.ok) {
        throw new Error(`Login failed: ${response.status}`)
      }

      const data = await response.json()
      return data.data.token
    } catch (error) {
      console.error('❌ 获取令牌失败:', error.message)
      return null
    }
  }

  async testBackendMiddleware(token) {
    console.log('🧪 测试后端中间件性能...')
    
    const tests = [
      {
        name: '认证中间件',
        endpoint: '/api/v1/users/profile',
        iterations: 50
      },
      {
        name: '频率限制中间件',
        endpoint: '/api/v1/letters/public',
        iterations: 20
      },
      {
        name: '健康检查',
        endpoint: '/health',
        iterations: 100,
        noAuth: true
      }
    ]

    for (const test of tests) {
      console.log(`\n📊 测试 ${test.name}...`)
      
      const times = []
      const cacheHits = []
      let successCount = 0
      let errorCount = 0

      for (let i = 0; i < test.iterations; i++) {
        const start = performance.now()
        
        try {
          const headers = {
            'Content-Type': 'application/json'
          }
          
          if (!test.noAuth && token) {
            headers['Authorization'] = `Bearer ${token}`
          }

          const response = await fetch(`${BACKEND_URL}${test.endpoint}`, {
            method: 'GET',
            headers
          })

          const end = performance.now()
          const duration = end - start

          times.push(duration)
          
          // 检查缓存命中
          const cacheHit = response.headers.get('X-Cache-Hit')
          if (cacheHit) {
            cacheHits.push(cacheHit === 'true')
          }

          if (response.ok) {
            successCount++
          } else {
            errorCount++
            if (response.status !== 429 && response.status !== 401) {
              console.warn(`  ⚠️ Unexpected status: ${response.status}`)
            }
          }
        } catch (error) {
          const end = performance.now()
          times.push(end - start)
          errorCount++
        }

        // 避免过快请求触发限流
        if (test.name.includes('频率限制')) {
          await new Promise(resolve => setTimeout(resolve, 100))
        }
      }

      // 计算统计信息
      const avgTime = times.reduce((sum, time) => sum + time, 0) / times.length
      const minTime = Math.min(...times)
      const maxTime = Math.max(...times)
      const cacheHitRate = cacheHits.length > 0 ? 
        (cacheHits.filter(hit => hit).length / cacheHits.length * 100) : 0

      this.results.backend[test.name] = {
        avgTime: avgTime.toFixed(2),
        minTime: minTime.toFixed(2),
        maxTime: maxTime.toFixed(2),
        successRate: (successCount / test.iterations * 100).toFixed(1),
        errorRate: (errorCount / test.iterations * 100).toFixed(1),
        cacheHitRate: cacheHitRate.toFixed(1)
      }

      console.log(`  ✅ 平均响应时间: ${avgTime.toFixed(2)}ms`)
      console.log(`  📈 最小/最大时间: ${minTime.toFixed(2)}ms / ${maxTime.toFixed(2)}ms`)
      console.log(`  🎯 成功率: ${(successCount / test.iterations * 100).toFixed(1)}%`)
      if (cacheHits.length > 0) {
        console.log(`  🚀 缓存命中率: ${cacheHitRate.toFixed(1)}%`)
      }
    }
  }

  async testFrontendMiddleware(token) {
    console.log('\n🧪 测试前端中间件性能...')
    
    const tests = [
      {
        name: 'API认证中间件',
        endpoint: '/api/auth/me',
        iterations: 30
      }
    ]

    for (const test of tests) {
      console.log(`\n📊 测试 ${test.name}...`)
      
      const times = []
      const cacheHits = []
      let successCount = 0
      let errorCount = 0

      for (let i = 0; i < test.iterations; i++) {
        const start = performance.now()
        
        try {
          const response = await fetch(`${FRONTEND_URL}${test.endpoint}`, {
            method: 'GET',
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json'
            }
          })

          const end = performance.now()
          const duration = end - start

          times.push(duration)
          
          // 检查缓存命中
          const cacheHit = response.headers.get('X-Cache-Hit')
          if (cacheHit) {
            cacheHits.push(cacheHit === 'true')
          }

          if (response.ok) {
            successCount++
          } else {
            errorCount++
          }
        } catch (error) {
          const end = performance.now()
          times.push(end - start)
          errorCount++
        }

        // 避免过快请求
        await new Promise(resolve => setTimeout(resolve, 50))
      }

      // 计算统计信息
      const avgTime = times.reduce((sum, time) => sum + time, 0) / times.length
      const minTime = Math.min(...times)
      const maxTime = Math.max(...times)
      const cacheHitRate = cacheHits.length > 0 ? 
        (cacheHits.filter(hit => hit).length / cacheHits.length * 100) : 0

      this.results.frontend[test.name] = {
        avgTime: avgTime.toFixed(2),
        minTime: minTime.toFixed(2),
        maxTime: maxTime.toFixed(2),
        successRate: (successCount / test.iterations * 100).toFixed(1),
        errorRate: (errorCount / test.iterations * 100).toFixed(1),
        cacheHitRate: cacheHitRate.toFixed(1)
      }

      console.log(`  ✅ 平均响应时间: ${avgTime.toFixed(2)}ms`)
      console.log(`  📈 最小/最大时间: ${minTime.toFixed(2)}ms / ${maxTime.toFixed(2)}ms`)
      console.log(`  🎯 成功率: ${(successCount / test.iterations * 100).toFixed(1)}%`)
      if (cacheHits.length > 0) {
        console.log(`  🚀 缓存命中率: ${cacheHitRate.toFixed(1)}%`)
      }
    }
  }

  async testConcurrentRequests(token) {
    console.log('\n🧪 测试并发请求处理...')
    
    const concurrentUsers = 10
    const requestsPerUser = 5
    
    console.log(`📊 模拟 ${concurrentUsers} 个并发用户，每用户 ${requestsPerUser} 个请求`)
    
    const startTime = performance.now()
    
    const promises = Array.from({ length: concurrentUsers }, async (_, userIndex) => {
      const userResults = []
      
      for (let i = 0; i < requestsPerUser; i++) {
        const requestStart = performance.now()
        
        try {
          const response = await fetch(`${BACKEND_URL}/api/v1/letters/public?limit=3`, {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json'
            }
          })
          
          const requestEnd = performance.now()
          userResults.push({
            success: response.ok,
            time: requestEnd - requestStart,
            status: response.status
          })
        } catch (error) {
          const requestEnd = performance.now()
          userResults.push({
            success: false,
            time: requestEnd - requestStart,
            error: error.message
          })
        }
        
        // 用户请求间隔
        await new Promise(resolve => setTimeout(resolve, 100))
      }
      
      return userResults
    })
    
    const allResults = await Promise.all(promises)
    const endTime = performance.now()
    
    // 分析结果
    const flatResults = allResults.flat()
    const successCount = flatResults.filter(r => r.success).length
    const totalRequests = flatResults.length
    const avgResponseTime = flatResults.reduce((sum, r) => sum + r.time, 0) / flatResults.length
    const totalTime = endTime - startTime
    const throughput = totalRequests / (totalTime / 1000) // requests per second
    
    console.log(`  ✅ 总请求数: ${totalRequests}`)
    console.log(`  🎯 成功率: ${(successCount / totalRequests * 100).toFixed(1)}%`)
    console.log(`  ⚡ 平均响应时间: ${avgResponseTime.toFixed(2)}ms`)
    console.log(`  🚀 吞吐量: ${throughput.toFixed(2)} req/s`)
    console.log(`  ⏱️ 总耗时: ${totalTime.toFixed(2)}ms`)
    
    this.results.concurrent = {
      totalRequests,
      successRate: (successCount / totalRequests * 100).toFixed(1),
      avgResponseTime: avgResponseTime.toFixed(2),
      throughput: throughput.toFixed(2),
      totalTime: totalTime.toFixed(2)
    }
  }

  generateReport() {
    console.log('\n📋 性能测试报告')
    console.log('=' * 50)
    
    console.log('\n🔧 后端中间件性能:')
    for (const [name, stats] of Object.entries(this.results.backend)) {
      console.log(`\n  ${name}:`)
      console.log(`    平均响应时间: ${stats.avgTime}ms`)
      console.log(`    成功率: ${stats.successRate}%`)
      if (parseFloat(stats.cacheHitRate) > 0) {
        console.log(`    缓存命中率: ${stats.cacheHitRate}%`)
      }
    }
    
    console.log('\n🎨 前端中间件性能:')
    for (const [name, stats] of Object.entries(this.results.frontend)) {
      console.log(`\n  ${name}:`)
      console.log(`    平均响应时间: ${stats.avgTime}ms`)
      console.log(`    成功率: ${stats.successRate}%`)
      if (parseFloat(stats.cacheHitRate) > 0) {
        console.log(`    缓存命中率: ${stats.cacheHitRate}%`)
      }
    }
    
    if (this.results.concurrent) {
      console.log('\n🚀 并发性能:')
      console.log(`    吞吐量: ${this.results.concurrent.throughput} req/s`)
      console.log(`    成功率: ${this.results.concurrent.successRate}%`)
      console.log(`    平均响应时间: ${this.results.concurrent.avgResponseTime}ms`)
    }
    
    // 性能评估
    console.log('\n📊 性能评估:')
    const backendAvgTime = Object.values(this.results.backend)
      .reduce((sum, stats) => sum + parseFloat(stats.avgTime), 0) / 
      Object.keys(this.results.backend).length
    
    if (backendAvgTime < 100) {
      console.log('  🟢 后端性能: 优秀 (< 100ms)')
    } else if (backendAvgTime < 300) {
      console.log('  🟡 后端性能: 良好 (< 300ms)')
    } else {
      console.log('  🔴 后端性能: 需要优化 (> 300ms)')
    }
    
    const throughput = this.results.concurrent ? parseFloat(this.results.concurrent.throughput) : 0
    if (throughput > 100) {
      console.log('  🟢 并发处理: 优秀 (> 100 req/s)')
    } else if (throughput > 50) {
      console.log('  🟡 并发处理: 良好 (> 50 req/s)')
    } else if (throughput > 0) {
      console.log('  🔴 并发处理: 需要优化 (< 50 req/s)')
    }
  }

  async run() {
    console.log('🚀 启动中间件性能测试...\n')
    
    const token = await this.getAuthToken()
    if (!token) {
      console.log('❌ 无法获取认证令牌，终止测试')
      return
    }
    
    try {
      await this.testBackendMiddleware(token)
      await this.testFrontendMiddleware(token)
      await this.testConcurrentRequests(token)
      
      this.generateReport()
      
      console.log('\n✅ 性能测试完成!')
    } catch (error) {
      console.error('❌ 测试过程中出现错误:', error)
    }
  }
}

// 检查fetch支持
if (typeof fetch === 'undefined') {
  console.log('❌ 此测试需要 Node.js 18+ 或 fetch polyfill')
  process.exit(1)
}

// 运行测试
const test = new PerformanceTest()
test.run().catch(console.error)