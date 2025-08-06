/**
 * 登录自动跳转Hook
 * 解决用户反馈的"登录成功，页面不显示内容"问题
 */

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/contexts/auth-context-new'
import { parseUserRole } from '@/lib/auth/role-system'

interface UseAuthRedirectOptions {
  /** 跳转到用户的默认页面 */
  redirectToDefault?: boolean
  /** 自定义跳转路径 */
  customPath?: string
  /** 只在首次登录时跳转 */
  onlyOnFirstLogin?: boolean
  /** 跳转延迟（毫秒） */
  delay?: number
}

export function useAuthRedirect(options: UseAuthRedirectOptions = {}) {
  const { user, isAuthenticated, isLoading } = useAuth()
  const router = useRouter()
  
  const {
    redirectToDefault = true,
    customPath,
    onlyOnFirstLogin = false,
    delay = 500
  } = options

  useEffect(() => {
    if (isLoading) return
    
    if (isAuthenticated && user) {
      const shouldRedirect = !onlyOnFirstLogin || !sessionStorage.getItem('hasRedirected')
      
      if (shouldRedirect) {
        const targetPath = customPath || (redirectToDefault ? getUserDefaultPath(user) : null)
        
        if (targetPath) {
          // 标记已经跳转过
          if (onlyOnFirstLogin) {
            sessionStorage.setItem('hasRedirected', 'true')
          }
          
          // 延迟跳转，让用户看到登录成功状态
          setTimeout(() => {
            router.push(targetPath)
          }, delay)
        }
      }
    }
  }, [user, isAuthenticated, isLoading, router, customPath, redirectToDefault, onlyOnFirstLogin, delay])

  return {
    isRedirecting: isAuthenticated && !isLoading,
    targetPath: customPath || (user ? getUserDefaultPath(user) : null)
  }
}

/**
 * 获取用户的默认跳转路径
 */
function getUserDefaultPath(user: any): string {
  try {
    const { defaultRoute } = parseUserRole(user)
    return defaultRoute
  } catch (error) {
    console.error('Failed to get user default path:', error)
    return '/dashboard'
  }
}

/**
 * 清除跳转标记（用于登出后重置）
 */
export function clearRedirectFlag() {
  sessionStorage.removeItem('hasRedirected')
}