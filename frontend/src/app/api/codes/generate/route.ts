import { NextRequest, NextResponse } from 'next/server'
import QRCode from 'qrcode'
import { ApiResponseBuilder } from '@/lib/api/response'

// 生成唯一编号
function generateLetterCode(): string {
  const prefix = 'OP'
  const timestamp = Date.now().toString(36).toUpperCase()
  const random = Math.random().toString(36).substring(2, 8).toUpperCase()
  return `${prefix}${timestamp}${random}`
}

// 生成二维码
async function generateQRCode(code: string): Promise<string> {
  try {
    const url = `${process.env.NEXT_PUBLIC_BASE_URL || 'http://localhost:3000'}/read/${code}`
    const qrDataUrl = await QRCode.toDataURL(url, {
      width: 200,
      margin: 2,
      color: {
        dark: '#2c1810',
        light: '#fdfcf9'
      }
    })
    return qrDataUrl
  } catch (error) {
    console.error('生成二维码失败:', error)
    throw new Error('生成二维码失败')
  }
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { title, content, style, userId } = body

    // 验证必需字段
    if (!content || !content.trim()) {
      return ApiResponseBuilder.error(
        400,
        '信件内容不能为空'
      )
    }

    // 生成唯一编号
    const letterCode = generateLetterCode()
    
    // 生成二维码
    const qrCode = await generateQRCode(letterCode)

    // 这里应该将信件信息保存到数据库
    // 目前使用模拟数据
    const letterData = {
      id: letterCode,
      code: letterCode,
      title: title || '无标题',
      content,
      style: style || 'classic',
      userId: userId || 'anonymous',
      status: 'draft', // draft, sent, delivered, read
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      qrCode
    }

    // 模拟保存到数据库的延迟
    await new Promise(resolve => setTimeout(resolve, 500))

    return ApiResponseBuilder.success({
      letterCode,
      qrCode,
      readUrl: `${process.env.NEXT_PUBLIC_BASE_URL || 'http://localhost:3000'}/read/${letterCode}`,
      letter: letterData
    }, '信件编号生成成功')

  } catch (error) {
    console.error('生成编号失败:', error)
    return ApiResponseBuilder.serverError(
      '生成编号失败，请稍后重试',
      error
    )
  }
}

export async function GET() {
  return ApiResponseBuilder.error(
    405,
    '请使用 POST 方法生成信件编号'
  )
}