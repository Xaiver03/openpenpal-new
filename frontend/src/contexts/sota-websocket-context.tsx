/**
 * SOTA WebSocket Context
 * 使用新的WebSocket管理器，实现优雅的连接管理
 */

'use client'

import React, { createContext, useContext, useEffect, useState } from 'react'
import { useUserStore } from '@/stores/user-store'
import { sotaWebSocketManager, ConnectionState } from '@/lib/websocket/sota-websocket-manager'

interface SOTAWebSocketContextType {
  connectionState: ConnectionState
  isConnected: boolean
  connect: () => void
  disconnect: () => void
  subscribe: (eventType: string, listener: (data: any) => void) => () => void
}

const SOTAWebSocketContext = createContext<SOTAWebSocketContextType | null>(null)

export function SOTAWebSocketProvider({ children }: { children: React.ReactNode }) {
  const [connectionState, setConnectionState] = useState<ConnectionState>(ConnectionState.DISCONNECTED)
  const { user, isAuthenticated } = useUserStore()

  // 监听连接状态变化
  useEffect(() => {
    const unsubscribe = sotaWebSocketManager.onStateChange(setConnectionState)
    return unsubscribe
  }, [])

  // 根据用户认证状态管理连接
  useEffect(() => {
    if (isAuthenticated && user) {
      console.log('📡 User authenticated, attempting WebSocket connection')
      sotaWebSocketManager.connect()
    } else {
      console.log('📡 User not authenticated, disconnecting WebSocket')
      sotaWebSocketManager.disconnect()
    }
  }, [isAuthenticated, user])

  // 页面可见性变化时的处理
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (!document.hidden && isAuthenticated) {
        // 页面变为可见且用户已认证时，尝试连接
        if (connectionState === ConnectionState.DISCONNECTED || 
            connectionState === ConnectionState.ERROR) {
          console.log('📡 Page visible and user authenticated, reconnecting')
          sotaWebSocketManager.connect()
        }
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange)
  }, [isAuthenticated, connectionState])

  // 组件卸载时清理
  useEffect(() => {
    return () => {
      sotaWebSocketManager.disconnect()
    }
  }, [])

  const contextValue: SOTAWebSocketContextType = {
    connectionState,
    isConnected: connectionState === ConnectionState.CONNECTED,
    connect: () => sotaWebSocketManager.connect(),
    disconnect: () => sotaWebSocketManager.disconnect(),
    subscribe: (eventType: string, listener: (data: any) => void) => 
      sotaWebSocketManager.subscribe(eventType, listener)
  }

  return (
    <SOTAWebSocketContext.Provider value={contextValue}>
      {children}
    </SOTAWebSocketContext.Provider>
  )
}

export function useSOTAWebSocket() {
  const context = useContext(SOTAWebSocketContext)
  if (!context) {
    throw new Error('useSOTAWebSocket must be used within a SOTAWebSocketProvider')
  }
  return context
}

export { ConnectionState }