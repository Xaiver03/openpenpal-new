import { NextRequest, NextResponse } from 'next/server'
import { JWTUtils } from '@/lib/auth/jwt-utils'
import { findCourierTestAccount } from '@/lib/auth/user-utils'

export async function GET(request: NextRequest) {
  try {
    const authHeader = request.headers.get('Authorization')
    
    if (!authHeader?.startsWith('Bearer ')) {
      return NextResponse.json({
        code: 401,
        message: '未提供授权令牌',
        data: null
      }, { status: 401 })
    }
    
    const token = authHeader.substring(7)
    
    try {
      const payload = JWTUtils.verifyAccessToken(token)
      
      // 简化：直接使用信使测试账号
      const courierTestAccount = findCourierTestAccount(payload.userId || '', payload.username || '')
      
      if (courierTestAccount) {
        return NextResponse.json({
          code: 0,
          message: '获取用户信息成功',
          data: {
            id: payload.userId,
            username: payload.username,
            email: payload.email,
            realName: courierTestAccount.levelName,
            role: payload.role,
            permissions: payload.permissions || [],
            schoolCode: payload.schoolCode,
            schoolName: courierTestAccount.zoneCode.includes('BJDX') ? '北京大学' : '系统测试',
            status: 'active',
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            courierInfo: {
              level: courierTestAccount.level,
              zoneCode: courierTestAccount.zoneCode,
              zoneType: courierTestAccount.zoneType,
              status: 'active',
              points: Math.floor(Math.random() * 1000) + 500,
              taskCount: Math.floor(Math.random() * 50) + 20
            }
          }
        })
      }
      
      return NextResponse.json({
        code: 404,
        message: '用户不存在',
        data: null
      }, { status: 404 })
      
    } catch (jwtError) {
      return NextResponse.json({
        code: 401,
        message: '无效的认证令牌',
        data: null
      }, { status: 401 })
    }
    
  } catch (error) {
    console.error('获取用户信息错误:', error)
    return NextResponse.json({
      code: 500,
      message: '服务器内部错误',
      data: null
    }, { status: 500 })
  }
}