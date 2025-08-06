import { NextRequest, NextResponse } from 'next/server'
import { UserRegistrationRequest } from '@/types/auth'

// 声明全局用户存储
declare global {
  var users: Map<string, any> | undefined
}

// 初始化全局用户数据存储
if (!global.users) {
  global.users = new Map<string, any>()
}

export async function POST(request: NextRequest) {
  try {
    const registrationData: UserRegistrationRequest = await request.json()
    
    // 验证必填字段
    const { username, email, password, confirmPassword, schoolCode, realName, verificationCode, agreeToTerms, agreeToPrivacy } = registrationData

    if (!username || !email || !password || !confirmPassword || !schoolCode || !realName) {
      return NextResponse.json(
        {
          code: 400,
          message: '请填写所有必填字段',
          data: null
        },
        { status: 400 }
      )
    }

    if (!verificationCode) {
      return NextResponse.json(
        {
          code: 400,
          message: '验证码不能为空',
          data: null
        },
        { status: 400 }
      )
    }

    // 验证密码匹配
    if (password !== confirmPassword) {
      return NextResponse.json(
        {
          code: 400,
          message: '两次输入的密码不一致',
          data: null
        },
        { status: 400 }
      )
    }

    // 验证协议同意
    if (!agreeToTerms || !agreeToPrivacy) {
      return NextResponse.json(
        {
          code: 400,
          message: '请同意用户协议和隐私政策',
          data: null
        },
        { status: 400 }
      )
    }

    // 检查邮箱是否已存在
    const existingUserByEmail = Array.from(global.users!.values()).find(user => user.email === email)
    if (existingUserByEmail) {
      return NextResponse.json(
        {
          code: 400,
          message: '该邮箱已被注册',
          data: null
        },
        { status: 400 }
      )
    }

    // 检查用户名是否已存在
    const existingUserByUsername = Array.from(global.users!.values()).find(user => user.username === username)
    if (existingUserByUsername) {
      return NextResponse.json(
        {
          code: 400,
          message: '该用户名已被占用',
          data: null
        },
        { status: 400 }
      )
    }

    // 生成用户ID
    const userId = `user_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    const now = new Date().toISOString()

    // 创建用户记录
    const newUser = {
      id: userId,
      username,
      email,
      realName,
      schoolCode,
      studentId: registrationData.studentId || null,
      major: registrationData.major || null,
      grade: registrationData.grade || null,
      className: registrationData.className || null,
      phone: registrationData.phone || null,
      status: 'ACTIVE',
      registeredAt: now,
      createdAt: now,
      updatedAt: now
    }

    // 存储用户（在实际环境中这会保存到数据库）
    global.users!.set(userId, newUser)

    console.log(`New user registered: ${username} (${email})`)

    return NextResponse.json({
      code: 0,
      message: 'success',
      data: {
        userId,
        username,
        email,
        status: 'ACTIVE',
        registeredAt: now,
        nextStep: '请登录开始使用OpenPenPal'
      }
    })
  } catch (error) {
    console.error('Registration error:', error)
    return NextResponse.json(
      {
        code: 500,
        message: '注册失败，请稍后重试',
        data: null
      },
      { status: 500 }
    )
  }
}