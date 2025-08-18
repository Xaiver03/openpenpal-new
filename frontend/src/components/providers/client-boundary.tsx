'use client'

import React, { Suspense, useState, useEffect } from 'react'

// 直接导入而不是懒加载以避免hydration问题
import { WebSocketProvider } from '@/contexts/websocket-context'

// 保持懒加载用于可选组件
const PerformanceMonitor = React.lazy(() => 
  import('@/components/optimization/performance-monitor').then(module => ({
    default: module.PerformanceMonitor || (() => React.createElement(React.Fragment))
  })).catch(() => ({ default: () => React.createElement(React.Fragment) }))
)

const AuthDebugPanel = React.lazy(() => 
  import('@/components/debug/auth-debug-panel').then(module => ({
    default: module.AuthDebugPanel || (() => React.createElement(React.Fragment))
  })).catch(() => ({ default: () => React.createElement(React.Fragment) }))
)

const NotificationManager = React.lazy(() => 
  import('@/components/realtime/notification-center').then(module => ({
    default: module.NotificationManager || module.NotificationCenter || (() => React.createElement(React.Fragment))
  })).catch(() => ({ default: () => React.createElement(React.Fragment) }))
)

interface ClientBoundaryProps {
  children: React.ReactNode
}

// 客户端组件的容器，避免hydration不匹配
function ClientOnlyComponents() {
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return null
  }

  return (
    <Suspense fallback={null}>
      <NotificationManager />
      <PerformanceMonitor />
      <AuthDebugPanel />
    </Suspense>
  )
}

// 客户端渲染的边界组件
export function ClientBoundary({ children }: ClientBoundaryProps) {
  return (
    <WebSocketProvider>
      <div className="relative flex min-h-screen flex-col" suppressHydrationWarning>
        <main className="flex-1">{children}</main>
        <ClientOnlyComponents />
      </div>
    </WebSocketProvider>
  )
}