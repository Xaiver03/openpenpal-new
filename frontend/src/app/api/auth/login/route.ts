import { NextRequest, NextResponse } from 'next/server'

const BACKEND_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000'

/**
 * 简化的前端登录代理 - 专注于CSRF + JWT集成
 */
export async function POST(request: NextRequest) {
  try {
    console.log('🔄 Frontend login proxy - Start')
    
    // 1. 解析请求体
    const body = await request.json()
    const { username, password } = body
    
    if (!username || !password) {
      return NextResponse.json({
        code: 400,
        message: '用户名和密码不能为空',
        data: null
      }, { status: 400 })
    }
    
    // 2. 获取客户端的CSRF token（如果有）
    const csrfToken = request.headers.get('X-CSRF-Token') || request.headers.get('x-csrf-token')
    console.log('🔑 CSRF Token from client:', csrfToken ? csrfToken.substring(0, 16) + '...' : 'none')
    
    // 3. 获取客户端cookies并转发
    const cookieHeader = request.headers.get('cookie')
    console.log('🍪 Cookies from client:', cookieHeader ? 'present' : 'none')
    
    // 4. 构建转发给后端的headers
    const forwardHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
    }
    
    // 转发CSRF token（如果有）
    if (csrfToken) {
      forwardHeaders['X-CSRF-Token'] = csrfToken
    }
    
    // 转发cookies（如果有）
    if (cookieHeader) {
      forwardHeaders['Cookie'] = cookieHeader
    }
    
    console.log('📤 Forwarding to backend with headers:', Object.keys(forwardHeaders))
    
    // 5. 调用后端登录API
    const backendResponse = await fetch(`${BACKEND_URL}/api/v1/auth/login`, {
      method: 'POST',
      headers: forwardHeaders,
      body: JSON.stringify({ username, password }),
      signal: AbortSignal.timeout(10000), // 10秒超时
      // 在开发环境中避免代理问题
      ...(process.env.NODE_ENV === 'development' ? { 
        // @ts-ignore - Node.js fetch特定选项
        agent: undefined 
      } : {})
    })
    
    console.log('📥 Backend response status:', backendResponse.status)
    
    // 6. 处理后端响应
    let backendData
    try {
      const responseText = await backendResponse.text()
      console.log('📥 Raw response length:', responseText.length)
      console.log('📥 First 100 chars:', responseText.substring(0, 100))
      console.log('📥 Response headers:', Object.fromEntries(backendResponse.headers.entries()))
      
      // 检查是否是空响应
      if (!responseText || responseText.trim() === '') {
        throw new Error('Empty response from backend')
      }
      
      // 尝试找到JSON开始的位置（处理可能的代理问题）
      const jsonStart = responseText.indexOf('{')
      if (jsonStart > 0) {
        console.log('🔧 Found JSON at position:', jsonStart, 'Prefix:', JSON.stringify(responseText.substring(0, jsonStart)))
        backendData = JSON.parse(responseText.substring(jsonStart))
      } else if (jsonStart === 0) {
        backendData = JSON.parse(responseText)
      } else {
        throw new Error('No JSON found in response')
      }
    } catch (parseError) {
      console.error('❌ Failed to parse backend response:', parseError)
      throw new Error(`Failed to parse backend response: ${parseError instanceof Error ? parseError.message : String(parseError)}`)
    }
    console.log('📥 Backend response data:', backendData.success ? 'success' : 'failed')
    
    if (!backendResponse.ok) {
      // 转发后端错误
      return NextResponse.json({
        code: backendData.code || backendResponse.status,
        message: backendData.message || '登录失败',
        data: null
      }, { status: backendResponse.status })
    }
    
    if (!backendData.success) {
      return NextResponse.json({
        code: backendData.code || 401,
        message: backendData.message || '登录失败',
        data: null
      }, { status: 401 })
    }
    
    // 7. 转换后端响应格式为前端期望格式
    const responseData = {
      code: 0,
      message: '登录成功',
      data: {
        accessToken: backendData.data.token,
        refreshToken: backendData.data.refreshToken || backendData.data.token,
        expiresAt: backendData.data.expiresAt,
        tokenType: 'Bearer',
        user: {
          id: backendData.data.user.id,
          username: backendData.data.user.username,
          email: backendData.data.user.email,
          nickname: backendData.data.user.nickname,
          role: backendData.data.user.role,
          schoolCode: backendData.data.user.schoolCode,
          createdAt: backendData.data.user.createdAt,
          updatedAt: backendData.data.user.updatedAt,
          lastLoginAt: backendData.data.user.lastLoginAt,
          isActive: backendData.data.user.isActive
        }
      },
      timestamp: new Date().toISOString()
    }
    
    // 8. 创建响应并转发后端设置的cookies
    const response = NextResponse.json(responseData)
    
    // 转发后端设置的认证cookies
    const setCookieHeaders = backendResponse.headers.get('set-cookie')
    if (setCookieHeaders) {
      response.headers.set('Set-Cookie', setCookieHeaders)
      console.log('🍪 Forwarding auth cookies from backend')
    }
    
    console.log('✅ Frontend login proxy - Success')
    return response
    
  } catch (error) {
    console.error('❌ Frontend login proxy error:', error)
    
    // 更详细的错误信息
    let errorMessage = '服务器内部错误，请稍后重试'
    let errorDetails = null
    
    if (error instanceof Error) {
      errorMessage = error.message
      errorDetails = {
        name: error.name,
        message: error.message,
        stack: process.env.NODE_ENV === 'development' ? error.stack : undefined
      }
    }
    
    console.error('Error details:', errorDetails)
    
    return NextResponse.json({
      code: 500,
      message: errorMessage,
      data: null,
      error: process.env.NODE_ENV === 'development' ? errorDetails : undefined
    }, { status: 500 })
  }
}