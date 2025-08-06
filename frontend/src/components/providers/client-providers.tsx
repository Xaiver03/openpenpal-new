'use client'

import dynamic from 'next/dynamic'

// 动态加载WebSocketProvider以避免SSR问题
export const WebSocketProvider = dynamic(
  () => import('@/contexts/websocket-context').then(mod => ({ default: mod.WebSocketProvider })),
  { 
    ssr: false,
    loading: () => <div className="hidden"></div> // 静默加载
  }
)

// 动态加载性能监控组件
export const PerformanceMonitor = dynamic(
  () => import('@/components/optimization/performance-monitor').then(mod => ({ default: mod.PerformanceMonitor })),
  { ssr: false }
)