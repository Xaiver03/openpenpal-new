import { NextRequest, NextResponse } from 'next/server'

// 使用全局冷却时间存储
declare global {
  var cooldowns: Map<string, number> | undefined
}

if (!global.cooldowns) {
  global.cooldowns = new Map()
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

  const now = Date.now()
  const cooldownUntil = global.cooldowns?.get(email) || 0
  const canSend = now >= cooldownUntil
  const cooldownSeconds = canSend ? 0 : Math.ceil((cooldownUntil - now) / 1000)

  return NextResponse.json({
    code: 0,
    message: 'success',
    data: {
      email,
      canSend,
      cooldownSeconds,
      message: canSend ? '可以发送验证码' : `请等待 ${cooldownSeconds} 秒后再发送验证码`
    }
  })
}