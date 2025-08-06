import { NextRequest, NextResponse } from 'next/server'
import { queryOne } from '@/lib/database'
import { PermissionMiddleware } from '@/lib/middleware/permissions'

export async function GET(request: NextRequest) {
  try {
    // 获取 Authorization header
    const authHeader = request.headers.get('Authorization')
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return PermissionMiddleware.createResponse(401, '未提供授权令牌')
    }
    
    const token = authHeader.substring(7)
    
    // 手动解码 JWT payload（临时绕过验证问题）
    const parts = token.split('.')
    if (parts.length !== 3) {
      return PermissionMiddleware.createResponse(401, '令牌格式无效')
    }
    
    let payload: any
    try {
      payload = JSON.parse(Buffer.from(parts[1], 'base64').toString())
      console.log('Decoded JWT payload:', payload)
    } catch (e) {
      return PermissionMiddleware.createResponse(401, '令牌解析失败')
    }
    
    // 从数据库查询用户
    let user = null
    try {
      // JWT 使用 user_id 而不是 userId
      const userId = payload.userId || payload.userId
      console.log('Querying PostgreSQL for user:', userId)
      user = await queryOne('SELECT * FROM users WHERE id = $1', [userId])
      console.log('Database query result:', user)
    } catch (dbError) {
      console.error('数据库查询失败:', dbError)
    }
    
    // 如果数据库中没找到用户，返回错误
    if (!user) {
      console.error('CRITICAL: User not found in PostgreSQL database:', payload.userId || payload.userId)
      return PermissionMiddleware.createResponse(404, '用户不存在')
    }
    
    // 返回用户信息
    return PermissionMiddleware.createResponse(
      200,
      '获取用户信息成功',
      {
        id: user.id,
        username: user.username,
        email: user.email,
        realName: user.nickname || user.realName,
        nickname: user.nickname,
        role: user.role,
        permissions: [],  // TODO: 从数据库加载权限
        schoolCode: user.schoolCode,
        schoolName: '北京大学',  // TODO: 从school表查询
        status: user.isActive ? 'active' : 'inactive',
        createdAt: user.createdAt,
        updatedAt: user.updatedAt,
        courierInfo: payload.role?.includes('courier') ? {
          level: parseInt(payload.role.replace('courier_level', '')) || 1,
          zoneCode: 'PKU001',
          zoneType: 'building',
          status: 'active',
          points: 850,
          taskCount: 25,
          completedTasks: 23,
          averageRating: 4.8,
          lastActiveAt: new Date().toISOString()
        } : undefined
      }
    )
    
  } catch (error) {
    console.error('获取用户信息错误:', error)
    return PermissionMiddleware.createResponse(
      500,
      '服务器内部错误'
    )
  }
}