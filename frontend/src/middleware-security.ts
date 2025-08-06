import { NextRequest, NextResponse } from 'next/server'
import { SecurityHeaders, HTTPSRedirect } from '@/lib/security/https-config'

/**
 * Enhanced security middleware for production
 * 生产环境增强安全中间件
 */
export function securityMiddleware(request: NextRequest) {
  // 1. HTTPS redirect in production
  if (HTTPSRedirect.shouldRedirect(request)) {
    return HTTPSRedirect.redirect(request)
  }

  // 2. Skip security headers for certain paths
  const pathname = request.nextUrl.pathname
  const skipSecurityPaths = [
    '/api/health',
    '/favicon.ico',
    '/_next/static/',
    '/_next/image/',
    '/images/',
    '/icons/'
  ]

  const shouldSkipSecurity = skipSecurityPaths.some(path => 
    pathname.startsWith(path)
  )

  if (shouldSkipSecurity) {
    return NextResponse.next()
  }

  // 3. Create response with security headers
  const response = NextResponse.next()
  
  // Apply security headers
  SecurityHeaders.apply(response)

  // 4. Additional security measures for API routes
  if (pathname.startsWith('/api/')) {
    // Add API-specific headers
    response.headers.set('X-Robots-Tag', 'noindex, nofollow, noarchive, nosnippet')
    response.headers.set('Cache-Control', 'no-store, no-cache, must-revalidate, private')
    response.headers.set('Pragma', 'no-cache')
    response.headers.set('Expires', '0')
  }

  // 5. Log security events in production
  if (process.env.NODE_ENV === 'production') {
    const userAgent = request.headers.get('user-agent') || 'unknown'
    const ip = request.headers.get('x-forwarded-for') || 
               request.headers.get('x-real-ip') || 
               'unknown'
    
    // Log suspicious patterns
    const suspiciousPatterns = [
      /\.(php|asp|jsp|cgi)$/i,
      /\/(wp-admin|admin|phpmyadmin)/i,
      /(union|select|insert|delete|drop|update).*from/i,
      /<script|javascript:|vbscript:|onload=/i
    ]

    const isSuspicious = suspiciousPatterns.some(pattern => 
      pattern.test(pathname) || pattern.test(request.url)
    )

    if (isSuspicious) {
      console.warn(`🚨 Suspicious request detected:`, {
        ip,
        userAgent: userAgent.substring(0, 100),
        path: pathname,
        timestamp: new Date().toISOString()
      })
    }
  }

  return response
}

/**
 * Integrate with existing middleware
 * 与现有中间件集成
 */
export function enhanceExistingMiddleware(
  existingMiddleware: (request: NextRequest) => NextResponse | Promise<NextResponse>
) {
  return async (request: NextRequest) => {
    // Apply security first
    const securityResponse = securityMiddleware(request)
    if (securityResponse instanceof NextResponse && securityResponse.status !== 200) {
      return securityResponse // Return redirects or blocks immediately
    }

    // Run existing middleware
    const existingResponse = await existingMiddleware(request)
    
    // Apply security headers to existing response
    if (existingResponse instanceof NextResponse) {
      SecurityHeaders.apply(existingResponse)
    }

    return existingResponse
  }
}