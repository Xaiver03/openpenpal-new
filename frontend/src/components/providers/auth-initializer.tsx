'use client'

import { useEffect } from 'react'
import { initializeAuth } from '@/lib/auth/auth-initializer'
import { getAuthSyncService } from '@/lib/auth/auth-sync-service'

export function AuthInitializer({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    // 初始化传统的认证
    initializeAuth()
    
    // 初始化强化的认证同步服务
    getAuthSyncService().initialize()
    
    // 清理函数
    return () => {
      getAuthSyncService().destroy()
    }
  }, [])

  return <>{children}</>
}