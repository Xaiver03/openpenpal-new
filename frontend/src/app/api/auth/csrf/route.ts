import { NextResponse } from 'next/server'
import { CSRFServer } from '@/lib/security/csrf'
import { StandardApiResponse } from '@/lib/api/response'

/**
 * GET /api/auth/csrf
 * 获取CSRF令牌
 */
export async function GET() {
  try {
    console.log('🔧 CSRF Endpoint: Generating new CSRF token...')
    
    // Create a simple response data object
    const responseData = {
      code: 0,
      message: '操作成功',
      data: {
        token: '',
        expiresIn: 86400 // 24 hours in seconds
      },
      timestamp: new Date().toISOString()
    }
    
    // Create a response object to get headers
    const tempResponse = new NextResponse()
    
    // Generate and set CSRF token using server-side utility
    console.log('🔧 CSRF Endpoint: Calling CSRFServer.generateAndSet...')
    const token = CSRFServer.generateAndSet(tempResponse.headers)
    
    console.log('🔧 CSRF Endpoint: Token generated:', token.substring(0, 16) + '...')
    console.log('🔧 CSRF Endpoint: Response headers:', [...tempResponse.headers.entries()])
    
    // Update the response data with the generated token
    responseData.data.token = token
    
    const finalHeaders = {
      'Content-Type': 'application/json',
      ...Object.fromEntries(tempResponse.headers.entries())
    }
    
    console.log('🔧 CSRF Endpoint: Final response headers:', finalHeaders)
    console.log('🔧 CSRF Endpoint: Response data:', responseData)
    
    // Create the final response
    const response = new NextResponse(JSON.stringify(responseData), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
      }
    })
    
    // Set the cookie directly on the response with explicit domain
    const isProduction = process.env.NODE_ENV === 'production'
    response.cookies.set('csrf-token', token, {
      httpOnly: false, // Allow JavaScript to read
      secure: isProduction, // HTTPS only in production
      sameSite: 'lax',
      path: '/',
      maxAge: 86400, // 24 hours
      domain: undefined // Let browser handle domain
    })
    
    // Also set via headers for better compatibility
    response.headers.set('X-CSRF-Token', token)
    
    return response
  } catch (error) {
    console.error('🚨 CSRF token generation failed:', error)
    return StandardApiResponse.error(500, 'Failed to generate CSRF token', error)
  }
}