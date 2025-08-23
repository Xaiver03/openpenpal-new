/**
 * Hydration Helper Utilities
 * 水合辅助工具
 * 
 * Common causes of hydration mismatch:
 * 1. Using Date.now() or new Date() in render
 * 2. Math.random() in render
 * 3. typeof window checks that change render output
 * 4. Reading from localStorage/sessionStorage
 * 5. User-specific data without proper handling
 */

/**
 * Check if code is running on client side
 * 检查代码是否在客户端运行
 */
export const isClient = typeof window !== 'undefined'

/**
 * Check if code is running on server side
 * 检查代码是否在服务器端运行
 */
export const isServer = !isClient

/**
 * Safe window access
 * 安全访问 window 对象
 */
export const safeWindow = isClient ? window : undefined

/**
 * Safe document access
 * 安全访问 document 对象
 */
export const safeDocument = isClient ? document : undefined

/**
 * Get stable ID for SSR/CSR
 * 获取 SSR/CSR 稳定的 ID
 */
let idCounter = 0
export function getStableId(prefix = 'id'): string {
  if (isServer) {
    // Use a deterministic ID on server
    return `${prefix}-ssr-${++idCounter}`
  }
  // Use a different pattern on client to avoid conflicts
  return `${prefix}-csr-${++idCounter}`
}

/**
 * Suppress hydration warning for a component
 * 抑制组件的水合警告
 */
export function suppressHydrationWarning<T extends React.HTMLAttributes<any>>(
  props: T
): T {
  return {
    ...props,
    suppressHydrationWarning: true
  }
}

/**
 * Create a client-only wrapper component
 * 创建仅客户端渲染的包装组件
 */
export function createClientOnly<P extends object>(
  Component: React.ComponentType<P>,
  fallback?: React.ReactNode
) {
  return function ClientOnlyComponent(props: P) {
    const [mounted, setMounted] = React.useState(false)

    React.useEffect(() => {
      setMounted(true)
    }, [])

    if (!mounted) {
      return fallback ? React.createElement(React.Fragment, null, fallback) : null
    }

    return React.createElement(Component, props)
  }
}

/**
 * Hook to detect hydration mismatches in development
 * 开发环境下检测水合不匹配的 Hook
 */
export function useHydrationCheck(componentName: string) {
  if (process.env.NODE_ENV === 'development') {
    const [isHydrated, setIsHydrated] = React.useState(false)
    
    React.useEffect(() => {
      setIsHydrated(true)
      
      // Check for common hydration issues
      const warnings: string[] = []
      
      // Check for date/time in component
      const componentElement = document.querySelector(`[data-hydration-check="${componentName}"]`)
      if (componentElement) {
        const text = componentElement.textContent || ''
        
        // Check for ISO date strings
        if (/\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/.test(text)) {
          warnings.push('Component contains date/time that might cause hydration mismatch')
        }
        
        // Check for random values
        if (/Math\.random|Date\.now/.test(text)) {
          warnings.push('Component might be using random values')
        }
      }
      
      if (warnings.length > 0) {
        console.warn(`[Hydration Check] ${componentName}:`, warnings)
      }
    }, [componentName])
    
    return isHydrated
  }
  
  return true
}

// Import React for the helper functions
import * as React from 'react'