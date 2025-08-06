import { NextRequest, NextResponse } from 'next/server'

// 模拟验证码存储 - 与 send-verification-code 共享
declare global {
  var verificationCodes: Map<string, { code: string; timestamp: number; attempts: number }> | undefined
}

if (!global.verificationCodes) {
  global.verificationCodes = new Map()
}

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const email = searchParams.get('email')
  const code = searchParams.get('code')

  if (!email || !code) {
    return NextResponse.json(
      {
        code: 400,
        message: '邮箱或验证码参数缺失',
        data: null
      },
      { status: 400 }
    )
  }

  const storedData = global.verificationCodes?.get(email)
  
  if (!storedData) {
    return NextResponse.json({
      code: 0,
      message: 'success',
      data: {
        email,
        isValid: false,
        message: '验证码不存在或已过期'
      }
    })
  }

  // 检查是否过期（5分钟）
  const now = Date.now()
  const isExpired = now - storedData.timestamp > 5 * 60 * 1000
  
  if (isExpired) {
    global.verificationCodes?.delete(email)
    return NextResponse.json({
      code: 0,
      message: 'success',
      data: {
        email,
        isValid: false,
        message: '验证码已过期'
      }
    })
  }

  // 检查尝试次数（最多5次）
  if (storedData.attempts >= 5) {
    global.verificationCodes?.delete(email)
    return NextResponse.json({
      code: 0,
      message: 'success',
      data: {
        email,
        isValid: false,
        message: '验证次数过多，请重新获取验证码'
      }
    })
  }

  // 验证码验证
  const isValid = storedData.code === code
  
  // 增加尝试次数
  storedData.attempts++
  
  console.log(`🔍 验证码验证: 邮箱=${email}, 输入=${code}, 存储=${storedData.code}, 结果=${isValid}`)
  
  if (isValid) {
    // 验证成功，删除验证码
    global.verificationCodes?.delete(email)
    console.log(`✅ 验证码验证成功，已删除存储的验证码`)
  } else {
    console.log(`❌ 验证码验证失败，尝试次数: ${storedData.attempts}/5`)
  }

  return NextResponse.json({
    code: 0,
    message: 'success',
    data: {
      email,
      isValid,
      message: isValid ? '验证码验证成功' : '验证码不正确'
    }
  })
}