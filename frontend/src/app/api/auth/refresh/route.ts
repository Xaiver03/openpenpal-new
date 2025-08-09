import { NextRequest, NextResponse } from 'next/server'
import { JWTUtils } from '@/lib/auth/jwt-utils'
import { TestDataManager, TestUserAccount } from '@/lib/auth/test-data-manager'

// 全局用户数据存储
let PRODUCTION_USERS: Record<string, TestUserAccount> = {}

// 初始化用户数据
async function initializeUserData() {
  try {
    PRODUCTION_USERS = await TestDataManager.initializeTestUsers()
  } catch (error) {
    console.error('Failed to initialize user data:', error)
    PRODUCTION_USERS = {}
  }
}

// 初始化用户数据
initializeUserData().catch(console.error)

export async function POST(request: NextRequest) {
  try {
    // 确保用户数据已初始化
    if (Object.keys(PRODUCTION_USERS).length === 0) {
      await initializeUserData()
    }
    
    const body = await request.json()
    const { refreshToken } = body
    
    // 输入验证
    if (!refreshToken) {
      return NextResponse.json({
        code: 400,
        message: 'Refresh token不能为空',
        data: null
      }, { status: 400 })
    }

    // 验证refresh token
    let refreshPayload: { userId: string; jti: string }
    try {
      refreshPayload = JWTUtils.verifyRefreshToken(refreshToken)
    } catch (error) {
      return NextResponse.json({
        code: 401,
        message: 'Refresh token无效或已过期',
        data: null
      }, { status: 401 })
    }

    // 首先尝试网关服务
    const gatewayUrl = process.env.NEXT_PUBLIC_GATEWAY_URL || 'http://localhost:8080'
    
    try {
      const response = await fetch(`${gatewayUrl}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refreshToken }),
        signal: AbortSignal.timeout(5000)
      })
      
      const result = await response.json()
      
      if (result.success) {
        return NextResponse.json({
          code: 0,
          message: 'Token刷新成功',
          data: {
            accessToken: result.data.accessToken || result.data.token,
            refreshToken: result.data.refreshToken || refreshToken,
            expiresAt: result.data.expiresAt || new Date(Date.now() + 15 * 60 * 1000).toISOString(),
            tokenType: 'Bearer'
          }
        })
      } else {
        throw new Error('网关刷新失败，使用本地刷新')
      }
    } catch (error) {
      console.log('网关服务不可用，使用本地刷新')
      
      // 本地刷新逻辑
      // 根据userId查找用户 (userId实际就是username)
      const user = PRODUCTION_USERS[refreshPayload.userId]
      
      if (!user) {
        return NextResponse.json({
          code: 401,
          message: '用户不存在',
          data: null
        }, { status: 401 })
      }

      // 检查账户状态
      if (user.status !== 'active') {
        return NextResponse.json({
          code: 403,
          message: '账户已被禁用，请重新登录',
          data: null
        }, { status: 403 })
      }

      // 生成新的JWT令牌对
      const tokenPair = JWTUtils.generateTokenPair({
        userId: user.id,
        username: user.username,
        email: user.email,
        role: user.role,
        permissions: user.permissions,
        schoolCode: user.school_code
      })

      return NextResponse.json({
        code: 0,
        message: 'Token刷新成功',
        data: {
          accessToken: tokenPair.accessToken,
          refreshToken: tokenPair.refreshToken,
          expiresAt: tokenPair.expiresAt,
          tokenType: 'Bearer'
        }
      })
    }
    
  } catch (error) {
    console.error('Token刷新错误:', error)
    return NextResponse.json({
      code: 500,
      message: 'Token刷新失败，请重新登录',
      data: null
    }, { status: 500 })
  }
}