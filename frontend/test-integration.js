/**
 * 前端后端集成测试脚本
 * 测试漂流瓶和未来信API的基本功能
 */

const API_BASE = 'http://localhost:8080/api/v1'

// 获取CSRF token
async function getCSRFToken() {
  try {
    const response = await fetch(`${API_BASE}/auth/csrf`)
    if (!response.ok) {
      throw new Error(`Failed to get CSRF token: ${response.status}`)
    }
    const data = await response.json()
    return data.data?.token || data.token
  } catch (error) {
    console.error('Failed to get CSRF token:', error)
    return null
  }
}

// 模拟用户登录并获取token
async function loginAndGetToken() {
  try {
    // 首先获取CSRF token
    const csrfToken = await getCSRFToken()
    if (!csrfToken) {
      throw new Error('Failed to obtain CSRF token')
    }

    const response = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': csrfToken
      },
      body: JSON.stringify({
        username: 'alice',
        password: 'Secret123!'
      })
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(`Login failed: ${response.status} - ${errorData.message || 'Unknown error'}`)
    }

    const data = await response.json()
    return data.data?.token || data.token
  } catch (error) {
    console.error('Login failed:', error)
    return null
  }
}

// 测试漂流瓶API
async function testDriftBottleAPI(token) {
  console.log('🧪 Testing Drift Bottle API...')
  
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  }

  try {
    // 1. 测试获取漂流中的瓶子
    console.log('  📍 Testing GET /drift-bottles/floating')
    const floatingResponse = await fetch(`${API_BASE}/drift-bottles/floating?limit=5`, {
      headers
    })
    
    if (floatingResponse.ok) {
      const floatingData = await floatingResponse.json()
      console.log('  ✅ Floating bottles:', floatingData.data?.length || 0)
    } else {
      console.log('  ❌ Floating bottles API failed:', floatingResponse.status)
    }

    // 2. 测试获取我的漂流瓶
    console.log('  📍 Testing GET /drift-bottles/my')
    const myBottlesResponse = await fetch(`${API_BASE}/drift-bottles/my?page=1&limit=10`, {
      headers
    })
    
    if (myBottlesResponse.ok) {
      const myBottlesData = await myBottlesResponse.json()
      console.log('  ✅ My bottles:', myBottlesData.data?.total || 0)
    } else {
      console.log('  ❌ My bottles API failed:', myBottlesResponse.status)
    }

    // 3. 测试统计信息
    console.log('  📍 Testing GET /drift-bottles/stats')
    const statsResponse = await fetch(`${API_BASE}/drift-bottles/stats`, {
      headers
    })
    
    if (statsResponse.ok) {
      const statsData = await statsResponse.json()
      console.log('  ✅ Stats retrieved successfully')
      console.log('    - Sent:', statsData.data?.sent_count || 0)
      console.log('    - Collected:', statsData.data?.collected_count || 0)
      console.log('    - Floating:', statsData.data?.floating_count || 0)
    } else {
      console.log('  ❌ Stats API failed:', statsResponse.status)
    }

  } catch (error) {
    console.error('  ❌ Drift Bottle API error:', error.message)
  }
}

// 测试未来信API
async function testFutureLetterAPI(token) {
  console.log('🧪 Testing Future Letter API...')
  
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  }

  try {
    // 1. 测试获取已安排的未来信
    console.log('  📍 Testing GET /future-letters')
    const scheduledResponse = await fetch(`${API_BASE}/future-letters?page=1&limit=10`, {
      headers
    })
    
    if (scheduledResponse.ok) {
      const scheduledData = await scheduledResponse.json()
      console.log('  ✅ Scheduled letters:', scheduledData.data?.total || 0)
    } else {
      console.log('  ❌ Scheduled letters API failed:', scheduledResponse.status)
    }

    // 2. 测试统计信息
    console.log('  📍 Testing GET /future-letters/stats')
    const statsResponse = await fetch(`${API_BASE}/future-letters/stats`, {
      headers
    })
    
    if (statsResponse.ok) {
      const statsData = await statsResponse.json()
      console.log('  ✅ Stats retrieved successfully')
      console.log('    - Pending:', statsData.data?.pending_count || 0)
      console.log('    - Upcoming 24h:', statsData.data?.upcoming_24h_count || 0)
    } else {
      console.log('  ❌ Stats API failed:', statsResponse.status)
    }

  } catch (error) {
    console.error('  ❌ Future Letter API error:', error.message)
  }
}

// 检查服务器健康状态
async function checkServerHealth() {
  try {
    const response = await fetch('http://localhost:8080/health')
    if (response.ok) {
      console.log('✅ Backend server is healthy')
      return true
    } else {
      console.log('❌ Backend server health check failed:', response.status)
      return false
    }
  } catch (error) {
    console.log('❌ Backend server is not accessible:', error.message)
    return false
  }
}

// 简单的API路由存在性测试（无需认证）
async function testAPIRoutesExist() {
  console.log('🧪 Testing API Routes Exist...')
  
  const routes = [
    '/api/v1/drift-bottles/floating',
    '/api/v1/drift-bottles/stats',
    '/api/v1/future-letters/stats'
  ]
  
  for (const route of routes) {
    try {
      console.log(`  📍 Testing ${route}`)
      const response = await fetch(`http://localhost:8080${route}`)
      
      if (response.status === 401 || response.status === 403) {
        console.log(`  ✅ Route exists (${response.status} - auth required)`)
      } else if (response.status === 404) {
        console.log(`  ❌ Route not found (${response.status})`)
      } else {
        console.log(`  ✅ Route accessible (${response.status})`)
      }
    } catch (error) {
      console.log(`  ❌ Route error: ${error.message}`)
    }
  }
}

// 主测试函数
async function runIntegrationTest() {
  console.log('🚀 Starting Frontend-Backend Integration Test')
  console.log('=' .repeat(50))

  // 1. 检查服务器健康状态
  const isHealthy = await checkServerHealth()
  if (!isHealthy) {
    console.log('❌ Cannot proceed without healthy backend server')
    return
  }

  // 2. 测试API路由是否存在
  await testAPIRoutesExist()

  // 3. 尝试登录获取token（如果失败也继续）
  console.log('🔐 Attempting login...')
  const token = await loginAndGetToken()
  if (!token) {
    console.log('⚠️  Authentication failed, but continuing with route tests')
  } else {
    console.log('✅ Login successful')
    
    // 4. 测试漂流瓶API
    await testDriftBottleAPI(token)

    // 5. 测试未来信API
    await testFutureLetterAPI(token)
  }

  console.log('=' .repeat(50))
  console.log('🎉 Integration test completed!')
}

// 运行测试
if (require.main === module) {
  runIntegrationTest().catch(console.error)
}

module.exports = {
  runIntegrationTest,
  testDriftBottleAPI,
  testFutureLetterAPI,
  checkServerHealth
}