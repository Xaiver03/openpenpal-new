import { NextRequest, NextResponse } from 'next/server'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

export async function GET(
  request: NextRequest,
  { params }: { params: { username: string } }
) {
  try {
    const { username } = params
    const { searchParams } = new URL(request.url)
    const publicOnly = searchParams.get('public') === 'true'

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

    // 构建查询参数
    const queryParams = new URLSearchParams()
    if (publicOnly) {
      queryParams.append('public', 'true')
    }
    
    const queryString = queryParams.toString()
    const url = `${API_URL}/users/${encodeURIComponent(username)}/letters${queryString ? `?${queryString}` : ''}`

    // 调用后端API获取用户信件
    const response = await fetch(url, {
      method: 'GET',
      headers
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      
      if (response.status === 404) {
        return NextResponse.json(
          { error: '用户不存在或无信件' },
          { status: 404 }
        )
      } else if (response.status === 403) {
        return NextResponse.json(
          { error: '无权限访问该用户的信件' },
          { status: 403 }
        )
      } else {
        return NextResponse.json(
          { error: errorData.error || '获取用户信件失败' },
          { status: response.status }
        )
      }
    }

    const data = await response.json()
    
    // 处理信件数据，只返回必要的信息
    const letters = (data.data || data.letters || []).map((letter: any) => ({
      id: letter.id,
      title: letter.title,
      content_preview: letter.content_preview || (letter.content ? letter.content.substring(0, 100) + '...' : ''),
      created_at: letter.created_at,
      status: letter.status,
      recipient: letter.recipient_username,
      sender: letter.sender_username,
      is_public: letter.is_public || false
    }))
    
    return NextResponse.json({
      success: true,
      data: letters,
      count: letters.length,
      message: '获取用户信件成功'
    })

  } catch (error) {
    console.error('获取用户信件失败:', error)
    return NextResponse.json(
      { error: '内部服务器错误' },
      { status: 500 }
    )
  }
}