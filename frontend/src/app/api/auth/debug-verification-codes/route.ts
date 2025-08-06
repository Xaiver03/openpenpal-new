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

  const codes: any = {}
  
  if (global.verificationCodes) {
    global.verificationCodes.forEach((data, email) => {
      codes[email] = {
        code: data.code,
        timestamp: new Date(data.timestamp).toLocaleString('zh-CN'),
        attempts: data.attempts,
        expired: Date.now() - data.timestamp > 5 * 60 * 1000
      }
    })
  }

  return NextResponse.json({
    code: 0,
    message: 'success',
    data: {
      verification_codes: codes,
      cooldowns: global.cooldowns ? Object.fromEntries(
        Array.from(global.cooldowns.entries()).map(([email, until]) => [
          email,
          {
            until: new Date(until).toLocaleString('zh-CN'),
            remaining_seconds: Math.max(0, Math.ceil((until - Date.now()) / 1000))
          }
        ])
      ) : {}
    }
  })
}