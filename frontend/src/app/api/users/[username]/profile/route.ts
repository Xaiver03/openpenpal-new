import { NextRequest, NextResponse } from 'next/server'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

export async function GET(
  request: NextRequest,
  { params }: { params: { username: string } }
) {
  try {
    const { username } = params

    if (!username) {
      return NextResponse.json(
        { error: '用户名不能为空' },
        { status: 400 }
      )
    }

    // 获取认证信息
    const authHeader = request.headers.get('authorization')
    const cookieHeader = request.headers.get('cookie')

    // 构建请求头
    const headers: Record<string, string> = {
      'Content-Type': 'application/json'
    }

    if (authHeader) {
      headers['Authorization'] = authHeader
    }
    if (cookieHeader) {
      headers['Cookie'] = cookieHeader
    }

    // 调用后端API获取用户资料
    const response = await fetch(`${API_URL}/users/${encodeURIComponent(username)}/profile`, {
      method: 'GET',
      headers
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      
      if (response.status === 404) {
        return NextResponse.json(
          { error: '用户不存在' },
          { status: 404 }
        )
      } else if (response.status === 403) {
        return NextResponse.json(
          { error: '该用户资料未公开或无权限访问' },
          { status: 403 }
        )
      } else {
        return NextResponse.json(
          { error: errorData.error || '获取用户资料失败' },
          { status: response.status }
        )
      }
    }

    const data = await response.json()
    
    // 返回用户资料数据
    return NextResponse.json({
      success: true,
      data: data.data || data,
      message: '获取用户资料成功'
    })

  } catch (error) {
    console.error('获取用户资料失败:', error)
    return NextResponse.json(
      { error: '内部服务器错误' },
      { status: 500 }
    )
  }
}