'use client'

import React, { Suspense } from 'react'

// 确保这些导入存在并正确导出
const WebSocketProvider = React.lazy(() => 
  import('@/contexts/websocket-context').then(module => ({
    default: module.WebSocketProvider
  }))
)

const PerformanceMonitor = React.lazy(() => 
  import('@/components/optimization/performance-monitor').then(module => ({
    default: module.PerformanceMonitor
  }))
)

const AuthDebugPanel = React.lazy(() => 
  import('@/components/debug/auth-debug-panel').then(module => ({
    default: module.AuthDebugPanel
  }))
)

const NotificationManager = React.lazy(() => 
  import('@/components/realtime/notification-center').then(module => ({
    default: module.NotificationManager
  }))
)

interface ClientBoundaryProps {
  children: React.ReactNode
}

// 客户端渲染的边界组件
export function ClientBoundary({ children }: ClientBoundaryProps) {
  // 确保只在客户端渲染
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    // 服务端渲染时返回占位符，保持结构一致
    return (
      <div className="relative flex min-h-screen flex-col">
        <main className="flex-1">{children}</main>
      </div>
    )
  }

  // 客户端渲染时加载所有组件
  return (
    <Suspense fallback={
      <div className="relative flex min-h-screen flex-col">
        <main className="flex-1">{children}</main>
      </div>
    }>
      <WebSocketProvider>
        <div className="relative flex min-h-screen flex-col">
          <main className="flex-1">{children}</main>
          <NotificationManager />
          <PerformanceMonitor />
          <AuthDebugPanel />
        </div>
      </WebSocketProvider>
    </Suspense>
  )
}