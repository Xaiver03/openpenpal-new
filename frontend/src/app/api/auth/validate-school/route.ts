import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const schoolCode = searchParams.get('schoolCode')

  if (!schoolCode) {
    return NextResponse.json(
      {
        code: 400,
        message: '学校编码参数缺失',
        data: null
      },
      { status: 400 }
    )
  }

  // 模拟学校编码验证 - 在实际环境中这会查询学校数据库
  const validSchoolCodes = [
    'BJUT2024',    // 北京理工大学
    'THU2024',     // 清华大学
    'PKU2024',     // 北京大学
    'BUAA2024',    // 北京航空航天大学
    'BNU2024',     // 北京师范大学
    'USTC2024',    // 中国科学技术大学
    'SJTU2024',    // 上海交通大学
    'FDU2024',     // 复旦大学
    'ZJU2024',     // 浙江大学
    'NJU2024'      // 南京大学
  ]
  
  const isValid = validSchoolCodes.includes(schoolCode.toUpperCase())

  return NextResponse.json({
    code: 0,
    message: 'success',
    data: {
      schoolCode,
      valid: isValid,
      message: isValid ? '学校编码有效' : '学校编码无效，请联系管理员获取正确编码'
    }
  })
}