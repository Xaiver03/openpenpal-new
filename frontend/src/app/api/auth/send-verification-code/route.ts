import { NextRequest, NextResponse } from 'next/server'

// 模拟验证码存储 - 在实际环境中这会使用 Redis 或数据库
declare global {
  var verificationCodes: Map<string, { code: string; timestamp: number; attempts: number }> | undefined
  var cooldowns: Map<string, number> | undefined
}

if (!global.verificationCodes) {
  global.verificationCodes = new Map()
}

if (!global.cooldowns) {
  global.cooldowns = new Map()
}

export async function POST(request: NextRequest) {
  try {
    const requestBody = await request.json()
    console.log('📥 发送验证码请求体:', requestBody)
    
    const { email } = requestBody

    if (!email) {
      console.log('❌ 邮箱参数缺失')
      return NextResponse.json(
        {
          code: 400,
          message: '邮箱参数缺失',
          data: null
        },
        { status: 400 }
      )
    }

    console.log('📧 处理邮箱验证码发送:', email)

    // 检查冷却时间
    const now = Date.now()
    const cooldownUntil = global.cooldowns?.get(email) || 0
    
    if (now < cooldownUntil) {
      const remainingSeconds = Math.ceil((cooldownUntil - now) / 1000)
      return NextResponse.json(
        {
          code: 400,
          message: `请等待 ${remainingSeconds} 秒后再重新发送`,
          data: null
        },
        { status: 400 }
      )
    }

    // 生成6位验证码
    const code = Math.floor(100000 + Math.random() * 900000).toString()
    
    // 存储验证码（5分钟有效期）
    global.verificationCodes?.set(email, {
      code,
      timestamp: now,
      attempts: 0
    })
    
    // 设置60秒冷却时间
    global.cooldowns?.set(email, now + 60000)

    // 尝试使用真实后端服务，如果失败则使用模拟
    try {
      const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000';
      const backendResponse = await fetch(`${backendUrl}/api/v1/auth/send-verification-code`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email }),
      });

      if (backendResponse.ok) {
        const backendData = await backendResponse.json();
        console.log('Real email sent via backend service');
        return NextResponse.json(backendData);
      }
    } catch (backendError) {
      console.log('Backend service unavailable, using mock email service');
    }

    // 后端不可用时的模拟邮件发送
    console.log(`🔥 Mock email sent to ${email}: Verification code is ${code}`)
    console.log('📧 请在浏览器控制台查看验证码，或启动后端服务以发送真实邮件')
    console.log('🎯 验证码已生成:', code)

    return NextResponse.json({
      code: 0,
      message: 'success',
      data: {
        email,
        message: '验证码已发送，请查收邮件',
        expiryMinutes: 5,
        cooldownSeconds: 60
      }
    })
  } catch (error) {
    console.error('❌ 发送验证码API错误:', error)
    return NextResponse.json(
      {
        code: 500,
        message: '验证码发送失败: ' + (error instanceof Error ? error.message : '未知错误'),
        data: null
      },
      { status: 500 }
    )
  }
}