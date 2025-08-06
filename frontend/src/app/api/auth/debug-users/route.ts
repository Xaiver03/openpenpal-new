import { NextRequest, NextResponse } from 'next/server'

// 调试端点 - 仅用于开发测试
export async function GET() {
  // 检查是否为开发环境
  if (process.env.NODE_ENV !== 'development') {
    return NextResponse.json({
      code: 403,
      message: 'Forbidden - Only available in development',
      data: null
    }, { status: 403 })
  }

  const users: any = {}
  
  if (global.users) {
    global.users.forEach((userData, userId) => {
      users[userId] = {
        username: userData.username,
        email: userData.email,
        realName: userData.realName,
        schoolCode: userData.schoolCode,
        status: userData.status,
        registeredAt: userData.registeredAt
      }
    })
  }

  return NextResponse.json({
    code: 0,
    message: 'success',
    data: {
      users,
      total: global.users?.size || 0
    }
  })
}