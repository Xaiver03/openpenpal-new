import { NextRequest, NextResponse } from 'next/server'
import { JWTUtils, PasswordUtils, SecurityUtils } from '@/lib/auth/jwt-utils'
import { RedisClient, SessionManager } from '@/lib/redis/redis-client'
import { DatabaseUserService } from '@/lib/services/database-user-service'
import { TestDataManager, TestUserAccount } from '@/lib/auth/test-data-manager'
import { rateLimiters } from '@/lib/security/rate-limit'
import { CSRFServer } from '@/lib/security/csrf'

// 全局用户数据存储（fallback）
let PRODUCTION_USERS: Record<string, TestUserAccount> = {}
let DATABASE_AVAILABLE = false

// 检查数据库可用性并初始化用户数据
async function initializeUserData() {
  try {
    // 首先尝试检查数据库中是否有测试用户
    const hasUsers = await DatabaseUserService.hasTestUsers()
    if (hasUsers) {
      DATABASE_AVAILABLE = true
      console.log('✅ 使用数据库中的测试用户')
      return
    }
  } catch (error) {
    console.log('⚠️  数据库不可用，使用内存存储')
  }

  // 如果数据库不可用，使用内存存储
  try {
    PRODUCTION_USERS = await TestDataManager.initializeTestUsers()
    console.log(`📄 内存初始化完成：${Object.keys(PRODUCTION_USERS).length} 个账户`)
  } catch (error) {
    console.error('❌ 内存用户数据初始化失败:', error)
    PRODUCTION_USERS = {}
  }
}

// 初始化用户数据
initializeUserData().catch(console.error)

export async function POST(request: NextRequest) {
  // Apply rate limiting if enabled
  const rateLimitEnabled = process.env.RATE_LIMIT_ENABLED === 'true' || process.env.NODE_ENV === 'production'
  
  if (rateLimitEnabled) {
    return rateLimiters.auth.middleware()(request, async (req) => {
      return await handleLogin(req)
    })
  } else {
    return await handleLogin(request)
  }
}

async function handleLogin(request: NextRequest | Request) {
    try {
      // Validate CSRF token
      const isDevelopment = process.env.NODE_ENV === 'development'
      const skipCSRF = isDevelopment; // Skip CSRF in development only
      const isValidCSRF = skipCSRF || CSRFServer.validate(request)
      
      if (!isValidCSRF) {
        return NextResponse.json(
          { 
            success: false, 
            error: 'Invalid CSRF token',
            code: 'CSRF_VALIDATION_FAILED'
          },
          { status: 403 }
        )
      }

      const body = await request.json()
      const { username, password } = body
      
      // 输入验证
      if (!username || !password) {
        console.log('Login attempt with missing credentials')
        
        return NextResponse.json({
          code: 400,
          message: '用户名和密码不能为空',
          data: null
        }, { status: 400 })
      }

    // 用户名格式验证
    const usernameValidation = SecurityUtils.validateUsername(username)
    if (!usernameValidation.isValid) {
      return NextResponse.json({
        code: 400,
        message: usernameValidation.errors[0],
        data: null
      }, { status: 400 })
    }

    // 检查Redis连接，如果失败则使用fallback认证
    let useRedis = true
    try {
      await RedisClient.get('connection_test')
    } catch (error) {
      console.warn('Redis不可用，使用fallback认证:', error)
      useRedis = false
    }

    // 首先尝试网关服务
    const gatewayUrl = process.env.NEXT_PUBLIC_GATEWAY_URL || 'http://localhost:8080'
    
    try {
      const response = await fetch(`${gatewayUrl}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
        signal: AbortSignal.timeout(5000)
      })
      
      const result = await response.json()
      
      if (result.success) {
        // 如果网关认证成功，使用网关返回的数据生成JWT
        const tokenPair = JWTUtils.generateTokenPair({
          userId: result.data.user.id,
          username: result.data.user.username,
          email: result.data.user.email,
          role: result.data.user.role,
          permissions: result.data.user.permissions || [],
          schoolCode: result.data.user.schoolCode
        })

        // 创建会话（如果Redis可用）
        if (useRedis) {
          const sessionId = SecurityUtils.generateSessionId()
          await SessionManager.createSession(sessionId, result.data.user.id, {
            username: result.data.user.username,
            role: result.data.user.role,
            loginTime: new Date().toISOString(),
            lastActivity: new Date().toISOString()
          })
        }

        return NextResponse.json({
          code: 0,
          message: '登录成功',
          data: {
            ...tokenPair,
            user: result.data.user
          }
        })
      } else {
        throw new Error('网关登录失败，使用本地认证')
      }
    } catch (error) {
      console.log('网关服务不可用，使用本地认证')
      
      // 优先使用数据库认证，如果不可用则使用内存认证
      if (DATABASE_AVAILABLE) {
        try {
          const user = await DatabaseUserService.authenticate(username, password)
          
          if (!user) {
            // 记录失败尝试（如果Redis可用）
            if (useRedis) {
              const failKey = `login_fail:${username}`
              await RedisClient.incr(failKey)
              await RedisClient.expire(failKey, 300) // 5分钟过期
            }
            
            return NextResponse.json({
              code: 401,
              message: '用户名或密码错误',
              data: null
            }, { status: 401 })
          }

          // 检查账户状态
          if (user.status !== 'active') {
            return NextResponse.json({
              code: 403,
              message: '账户已被禁用，请联系管理员',
              data: null
            }, { status: 403 })
          }

          // 检查登录失败次数（如果Redis可用）
          if (useRedis) {
            const failKey = `login_fail:${username}`
            const failCount = await RedisClient.get(failKey)
            if (failCount && parseInt(failCount) >= 5) {
              return NextResponse.json({
                code: 429,
                message: '登录失败次数过多，请5分钟后重试',
                data: null
              }, { status: 429 })
            }
          }

          // 清除失败记录
          if (useRedis) {
            await RedisClient.del(`login_fail:${username}`)
          }

          // 更新最后登录时间
          await DatabaseUserService.updateLastLogin(username)

          // 生成JWT令牌对
          const tokenPair = JWTUtils.generateTokenPair({
            userId: user.id,
            username: user.username,
            email: user.email,
            role: user.role,
            permissions: user.permissions,
            schoolCode: (user as any).schoolCode || (user as any).school_code
          })

          // 创建会话
          if (useRedis) {
            const sessionId = SecurityUtils.generateSessionId()
            await SessionManager.createSession(sessionId, user.id, {
              username: user.username,
              role: user.role,
              loginTime: new Date().toISOString(),
              lastActivity: new Date().toISOString()
            })
          }

          // 准备返回的用户数据（不包含密码哈希）
          const { password_hash, ...safeUserData } = user
          
          return NextResponse.json({
            code: 0,
            message: '登录成功',
            data: {
              ...tokenPair,
              user: {
                id: safeUserData.id,
                username: safeUserData.username,
                email: safeUserData.email,
                realName: safeUserData.realName,
                role: safeUserData.role,
                permissions: safeUserData.permissions,
                schoolCode: safeUserData.schoolCode,
                schoolName: safeUserData.school_name,
                school_code: safeUserData.schoolCode,
                school_name: safeUserData.school_name,
                status: safeUserData.status,
                createdAt: safeUserData.createdAt,
                updatedAt: safeUserData.updatedAt,
                courierLevel: safeUserData.courier_level,
                courierInfo: safeUserData.courier_info
              }
            }
          })
        } catch (dbError) {
          console.log('数据库认证失败，降级到内存认证:', dbError)
          DATABASE_AVAILABLE = false
        }
      }
      
      // 内存认证逻辑（fallback）
      const user = PRODUCTION_USERS[username]
      
      if (!user) {
        // 记录失败尝试（如果Redis可用）
        if (useRedis) {
          const failKey = `login_fail:${username}`
          await RedisClient.incr(failKey)
          await RedisClient.expire(failKey, 300) // 5分钟过期
        }
        
        return NextResponse.json({
          code: 401,
          message: '用户名或密码错误',
          data: null
        }, { status: 401 })
      }

      // 检查账户状态
      if (user.status !== 'active') {
        return NextResponse.json({
          code: 403,
          message: '账户已被禁用，请联系管理员',
          data: null
        }, { status: 403 })
      }

      // 检查登录失败次数（如果Redis可用）
      if (useRedis) {
        const failKey = `login_fail:${username}`
        const failCount = await RedisClient.get(failKey)
        if (failCount && parseInt(failCount) >= 5) {
          return NextResponse.json({
            code: 429,
            message: '登录失败次数过多，请5分钟后重试',
            data: null
          }, { status: 429 })
        }
      }

      // 验证密码
      const isPasswordValid = await PasswordUtils.comparePassword(password, user.passwordHash || '')
      
      if (!isPasswordValid) {
        // 记录失败尝试
        if (useRedis) {
          const failKey = `login_fail:${username}`
          await RedisClient.incr(failKey)
          await RedisClient.expire(failKey, 300)
        }
        
        return NextResponse.json({
          code: 401,
          message: '用户名或密码错误',
          data: null
        }, { status: 401 })
      }

      // 清除失败记录
      if (useRedis) {
        await RedisClient.del(`login_fail:${username}`)
      }

      // 生成JWT令牌对
      const tokenPair = JWTUtils.generateTokenPair({
        userId: user.id,
        username: user.username,
        email: user.email,
        role: user.role,
        permissions: user.permissions,
        schoolCode: user.school_code
      })

      // 创建会话
      if (useRedis) {
        const sessionId = SecurityUtils.generateSessionId()
        await SessionManager.createSession(sessionId, user.id, {
          username: user.username,
          role: user.role,
          loginTime: new Date().toISOString(),
          lastActivity: new Date().toISOString()
        })
      }

      // 准备返回的用户数据（不包含密码哈希）
      const { passwordHash, ...safeUserData } = user
      
      return NextResponse.json({
        code: 0,
        message: '登录成功',
        data: {
          ...tokenPair,
          user: safeUserData
        }
      })
    }
    
    } catch (error) {
      console.error('登录错误:', error)
      return NextResponse.json({
        code: 500,
        message: '登录失败，请稍后重试',
        data: null
      }, { status: 500 })
    }
}