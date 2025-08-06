import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const username = searchParams.get('username')

  if (!username) {
    return NextResponse.json(
      {
        code: 400,
        message: '用户名参数缺失',
        data: null
      },
      { status: 400 }
    )
  }

  // 用户名长度和格式验证
  if (username.length < 3 || username.length > 20) {
    return NextResponse.json(
      {
        code: 400,
        message: '用户名长度应在3-20个字符之间',
        data: null
      },
      { status: 400 }
    )
  }

  const usernameRegex = /^[a-zA-Z0-9_\u4e00-\u9fa5]+$/
  if (!usernameRegex.test(username)) {
    return NextResponse.json(
      {
        code: 400,
        message: '用户名只能包含字母、数字、下划线和中文字符',
        data: null
      },
      { status: 400 }
    )
  }

  // 模拟用户名检查 - 在实际环境中这会查询数据库
  const mockTakenUsernames = ['admin', 'test', 'user', 'openpenpal', '管理员']
  const isTaken = mockTakenUsernames.includes(username.toLowerCase())

  return NextResponse.json({
    code: 0,
    message: 'success',
    data: {
      username,
      available: !isTaken,
      message: isTaken ? '该用户名已被占用' : '用户名可用'
    }
  })
}