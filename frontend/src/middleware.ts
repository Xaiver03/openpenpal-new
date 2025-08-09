/**
 * Next.js Middleware for Route Protection
 * 前端路由保护中间件
 */

import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { JWTUtils } from '@/lib/auth/jwt-utils'
import { securityMiddleware } from './middleware-security'

// 需要认证的路由路径
const PROTECTED_ROUTES = [
  '/write', 
  '/plaza',
  '/museum',
  '/shop',
  '/profile',
  '/settings',
  '/courier',
  '/admin'
]

// 公开路由路径（不需要认证）
const PUBLIC_ROUTES = [
  '/',
  '/login',
  '/register',
  '/about',
  '/contact',
  '/privacy',
  '/terms',
  '/ai'  // AI页面改为公开访问
]

// API路由不需要在这里处理
const API_ROUTES = [
  '/api'
]

/**
 * 检查路径是否需要认证
 */
function isProtectedRoute(pathname: string): boolean {
  return PROTECTED_ROUTES.some(route => 
    pathname.startsWith(route) || pathname === route
  )
}

/**
 * 检查路径是否为公开路由
 */
function isPublicRoute(pathname: string): boolean {
  return PUBLIC_ROUTES.some(route => 
    pathname === route || pathname.startsWith(route)
  )
}

/**
 * 检查路径是否为API路由
 */
function isApiRoute(pathname: string): boolean {
  return API_ROUTES.some(route => pathname.startsWith(route))
}

/**
 * 获取并验证token
 */
function getAndValidateToken(request: NextRequest): {
  isValid: boolean
  token: string | null
  payload: any | null
} {
  // 1. 尝试从Cookie获取token
  let token = request.cookies.get('openpenpal_auth_token')?.value
  
  // 2. 如果Cookie没有，尝试从Authorization header获取
  if (!token) {
    const authHeader = request.headers.get('authorization')
    if (authHeader && authHeader.startsWith('Bearer ')) {
      token = authHeader.substring(7)
    }
  }
  
  if (!token) {
    return { isValid: false, token: null, payload: null }
  }
  
  try {
    // 只检查token格式和过期时间，不验证签名（签名验证由API路由处理）
    const payload = JWTUtils.decodeToken(token)
    if (!payload || !payload.exp) {
      console.log('🔒 Token validation failed: Invalid format')
      return { isValid: false, token, payload: null }
    }
    
    // 检查是否过期
    if (Date.now() >= payload.exp * 1000) {
      console.log('🔒 Token validation failed: Expired')
      return { isValid: false, token, payload: null }
    }
    
    return { isValid: true, token, payload }
  } catch (error) {
    console.log('🔒 Token validation failed:', error)
    return { isValid: false, token, payload: null }
  }
}

/**
 * 创建登录重定向响应
 */
function createLoginRedirect(request: NextRequest, reason?: string): NextResponse {
  const loginUrl = new URL('/login', request.url)
  
  // 保存原始URL以便登录后重定向
  loginUrl.searchParams.set('redirect', request.nextUrl.pathname + request.nextUrl.search)
  
  if (reason) {
    loginUrl.searchParams.set('reason', reason)
  }
  
  console.log('🔒 Redirecting to login:', {
    from: request.nextUrl.pathname,
    to: loginUrl.toString(),
    reason
  })
  
  const response = NextResponse.redirect(loginUrl)
  
  // 清理过期的认证cookie
  response.cookies.delete('openpenpal_auth_token')
  response.cookies.delete('openpenpal_user')
  
  return response
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  
  console.log('🛡️ Middleware processing:', pathname)
  
  // Apply security middleware first (HTTPS, security headers)
  const securityResponse = securityMiddleware(request)
  
  // 跳过API路由（由API路由自己处理认证）
  if (isApiRoute(pathname)) {
    console.log('🛡️ Skipping API route:', pathname)
    return securityResponse
  }
  
  // 跳过静态资源
  if (pathname.startsWith('/_next') || 
      pathname.startsWith('/favicon') ||
      pathname.includes('.')) {
    return securityResponse
  }
  
  // 公开路由直接通过
  if (isPublicRoute(pathname)) {
    console.log('🛡️ Public route allowed:', pathname)
    return securityResponse
  }
  
  // 检查是否为需要保护的路由
  if (isProtectedRoute(pathname)) {
    const { isValid, token, payload } = getAndValidateToken(request)
    
    if (!isValid) {
      console.log('🔒 Protected route access denied - invalid token:', pathname)
      return createLoginRedirect(request, token ? 'token_expired' : 'no_token')
    }
    
    // Token有效，检查用户角色和权限（如果需要）
    if (pathname.startsWith('/admin')) {
      const userRole = payload?.role
      const adminRoles = ['super_admin', 'platform_admin']
      
      if (!adminRoles.includes(userRole)) {
        console.log('🔒 Admin route access denied - insufficient role:', {
          pathname,
          userRole,
          requiredRoles: adminRoles
        })
        
        // 重定向到首页而不是登录页
        return NextResponse.redirect(new URL('/', request.url))
      }
    }
    
    if (pathname.startsWith('/courier')) {
      const userRole = payload?.role
      const courierRoles = ['courier_level1', 'courier_level2', 'courier_level3', 'courier_level4', 'super_admin']
      
      if (!courierRoles.includes(userRole)) {
        console.log('🔒 Courier route access denied - insufficient role:', {
          pathname,
          userRole,
          requiredRoles: courierRoles
        })
        
        return NextResponse.redirect(new URL('/', request.url))
      }
    }
    
    console.log('✅ Protected route access granted:', {
      pathname,
      userId: payload?.userId,
      role: payload?.role
    })
    
    // 在请求头中添加用户信息，便于页面组件使用
    const requestHeaders = new Headers(request.headers)
    requestHeaders.set('x-user-id', payload.userId)
    requestHeaders.set('x-user-role', payload.role)
    
    const response = NextResponse.next({
      request: {
        headers: requestHeaders
      }
    })
    
    // Add security headers to authenticated responses
    return securityMiddleware(request)
  }
  
  // 其他路由默认允许通过
  console.log('🛡️ Route allowed by default:', pathname)
  return securityResponse
}

// 配置中间件匹配的路径
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
}