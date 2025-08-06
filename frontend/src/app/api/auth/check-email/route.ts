import { NextRequest, NextResponse } from 'next/server'

// 声明全局用户存储类型（与注册API共享）
declare global {
  var users: Map<string, any> | undefined
}

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const email = searchParams.get('email')

  if (!email) {
    return NextResponse.json(
      {
        code: 400,
        message: '邮箱参数缺失',
        data: null
      },
      { status: 400 }
    )
  }

  // 简单的邮箱格式验证
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(email)) {
    return NextResponse.json(
      {
        code: 400,
        message: '邮箱格式不正确',
        data: null
      },
      { status: 400 }
    )
  }

  // 检查已注册的邮箱
  const mockTakenEmails = ['test@example.com', 'admin@openpenpal.com']
  
  // 检查注册API中存储的用户数据（共享全局状态）
  let registeredInSystem = false
  if (global.users) {
    registeredInSystem = Array.from(global.users.values()).some((user: any) => 
      user.email?.toLowerCase() === email.toLowerCase()
    )
  }
  
  const isTaken = mockTakenEmails.includes(email.toLowerCase()) || registeredInSystem

  return NextResponse.json({
    code: 0,
    message: 'success',
    data: {
      email,
      available: !isTaken,
      message: isTaken ? '该邮箱已被注册' : '邮箱可用'
    }
  })
}

export async function POST(request: NextRequest) {
  try {
    const { email } = await request.json()
    
    if (!email) {
      return NextResponse.json({
        code: 400,
        message: '邮箱参数缺失',
        data: null
      }, { status: 400 })
    }
    
    // 简单的邮箱格式验证
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(email)) {
      return NextResponse.json({
        code: 400,
        message: '邮箱格式不正确',
        data: null
      }, { status: 400 })
    }
    
    // 检查已注册的邮箱
    const mockTakenEmails = ['test@example.com', 'admin@openpenpal.com']
    
    // 检查注册API中存储的用户数据（共享全局状态）
    let registeredInSystem = false
    if (global.users) {
      registeredInSystem = Array.from(global.users.values()).some((user: any) => 
        user.email?.toLowerCase() === email.toLowerCase()
      )
    }
    
    const isTaken = mockTakenEmails.includes(email.toLowerCase()) || registeredInSystem
    
    console.log(`📧 邮箱可用性检查 (POST): ${email} -> ${!isTaken ? '可用' : '已被注册'}`)
    
    return NextResponse.json({
      code: 0,
      message: 'success',
      data: {
        email,
        available: !isTaken,
        message: isTaken ? '该邮箱已被注册' : '邮箱可用'
      }
    })
    
  } catch (error) {
    console.error('邮箱检查API错误:', error)
    return NextResponse.json({
      code: 500,
      message: '邮箱检查失败',
      data: null
    }, { status: 500 })
  }
}