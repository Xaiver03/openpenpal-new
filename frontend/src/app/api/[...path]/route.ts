import { NextRequest, NextResponse } from 'next/server'

// 如果你有外部后端，在这里配置
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080'

// API 代理处理器
async function handler(req: NextRequest, { params }: { params: { path: string[] } }) {
  const path = params.path.join('/')
  // Map frontend routes to backend routes
  const routePath = path.startsWith('ai/') ? path : `v1/${path}`
  const url = `${BACKEND_URL}/api/${routePath}${req.nextUrl.search}`

  try {
    // Clean up headers - remove Next.js specific headers that might cause issues
    const headers = new Headers()
    req.headers.forEach((value, key) => {
      // Skip Next.js internal headers and problematic headers
      if (!key.startsWith('x-') && 
          !key.startsWith('next-') && 
          key !== 'host' && 
          key !== 'connection' &&
          key !== 'transfer-encoding' &&
          key !== 'content-length') {
        headers.set(key, value)
      }
    })
    
    // Set the correct host header
    headers.set('host', new URL(BACKEND_URL).host)

    // 构建请求选项
    const fetchOptions: RequestInit = {
      method: req.method,
      headers: headers,
      // Important: don't follow redirects automatically
      redirect: 'manual',
    }

    // 只有在有请求体的方法中才添加body
    if (req.method !== 'GET' && req.method !== 'HEAD') {
      const contentType = req.headers.get('content-type')
      if (contentType?.includes('application/json')) {
        try {
          const body = await req.json()
          fetchOptions.body = JSON.stringify(body)
          // Ensure content-type is set
          headers.set('content-type', 'application/json')
        } catch (e) {
          console.error('Failed to parse JSON body:', e)
          return NextResponse.json(
            { error: 'Invalid JSON in request body' },
            { status: 400 }
          )
        }
      } else {
        // 对于非JSON内容，直接传递请求体
        fetchOptions.body = await req.text()
      }
    }

    // 转发请求到后端
    const response = await fetch(url, fetchOptions)

    // Handle redirects
    if (response.status >= 300 && response.status < 400) {
      const location = response.headers.get('location')
      if (location) {
        return NextResponse.redirect(new URL(location, req.url))
      }
    }

    // 处理响应
    const contentType = response.headers.get('content-type')
    
    // 如果是JSON响应
    if (contentType?.includes('application/json')) {
      const data = await response.json()
      return NextResponse.json(data, { 
        status: response.status,
        headers: {
          'content-type': 'application/json',
          'cache-control': 'no-store',
        }
      })
    }
    
    // 对于非JSON响应，返回原始内容
    const text = await response.text()
    return new NextResponse(text, { 
      status: response.status,
      headers: {
        'content-type': contentType || 'text/plain',
        'cache-control': 'no-store',
      }
    })
  } catch (error) {
    console.error('API Proxy Error:', error)
    console.error('Failed URL:', url)
    console.error('Request method:', req.method)
    console.error('Error details:', error instanceof Error ? error.stack : error)
    
    // Check if it's a connection error
    if (error instanceof Error && error.message.includes('fetch failed')) {
      return NextResponse.json(
        { 
          error: 'Backend connection failed',
          message: 'Unable to connect to the backend service. Please ensure the backend is running.',
          path: path,
          backend_url: url
        },
        { status: 503 }
      )
    }
    
    // 返回更详细的错误信息
    return NextResponse.json(
      { 
        error: 'Internal Server Error',
        message: error instanceof Error ? error.message : 'Unknown error',
        path: path
      },
      { status: 500 }
    )
  }
}

export const GET = handler
export const POST = handler
export const PUT = handler
export const DELETE = handler
export const PATCH = handler