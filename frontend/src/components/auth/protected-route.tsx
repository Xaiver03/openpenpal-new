'use client'

import { useEffect } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { useAuth } from '@/contexts/auth-context-new'
import { Loader2 } from 'lucide-react'

interface ProtectedRouteProps {
  children: React.ReactNode
  requiredRole?: string | string[]
  redirectTo?: string
}

export function ProtectedRoute({ 
  children, 
  requiredRole,
  redirectTo = '/login' 
}: ProtectedRouteProps) {
  const { user, isAuthenticated, isLoading } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      // 保存当前路径，登录后可以重定向回来
      const returnUrl = encodeURIComponent(pathname || '/')
      router.push(`${redirectTo}?returnUrl=${returnUrl}`)
    }
  }, [isLoading, isAuthenticated, router, pathname, redirectTo])

  // 角色检查
  useEffect(() => {
    if (!isLoading && isAuthenticated && requiredRole && user) {
      const allowedRoles = Array.isArray(requiredRole) ? requiredRole : [requiredRole]
      
      if (!allowedRoles.includes(user.role)) {
        // 如果没有权限，重定向到首页或403页面
        router.push('/403')
      }
    }
  }, [isLoading, isAuthenticated, requiredRole, user, router])

  // 加载中显示
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-amber-600" />
      </div>
    )
  }

  // 未认证或无权限时不渲染子组件
  if (!isAuthenticated) {
    return null
  }

  if (requiredRole && user) {
    const allowedRoles = Array.isArray(requiredRole) ? requiredRole : [requiredRole]
    if (!allowedRoles.includes(user.role)) {
      return null
    }
  }

  return <>{children}</>
}