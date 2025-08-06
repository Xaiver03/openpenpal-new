/**
 * 权限实时通知Hook - 管理SSE连接和权限变更事件
 */

import { useEffect, useRef, useState, useCallback } from 'react'
import { useUserStore } from '@/stores/user-store'

export interface PermissionChangeEvent {
  type: 'permission_updated' | 'permission_reset' | 'config_imported' | 'user_affected' | 'connected' | 'heartbeat'
  data?: {
    target: string
    targetType: 'role' | 'courier-level' | 'system'
    affectedUsers?: number
    modifiedBy: string
    timestamp: string
    changes?: {
      added: string[]
      removed: string[]
    }
  }
  timestamp?: string
}

export interface PermissionNotificationState {
  connected: boolean
  lastEvent: PermissionChangeEvent | null
  connectionError: string | null
  eventCount: number
}

export function usePermissionNotifications() {
  const [state, setState] = useState<PermissionNotificationState>({
    connected: false,
    lastEvent: null,
    connectionError: null,
    eventCount: 0
  })

  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const { refreshPermissions, user } = useUserStore()

  // 处理权限变更事件
  const handlePermissionChange = useCallback(async (event: PermissionChangeEvent) => {
    console.log('收到权限变更事件:', event)

    setState(prev => ({
      ...prev,
      lastEvent: event,
      eventCount: prev.eventCount + 1
    }))

    // 如果当前用户受到影响，刷新权限
    if (event.type === 'permission_updated' || event.type === 'permission_reset' || event.type === 'config_imported') {
      const shouldRefresh = shouldRefreshUserPermissions(event, user)
      if (shouldRefresh) {
        console.log('当前用户权限受到影响，刷新权限缓存')
        await refreshPermissions()
      }
    }
  }, [refreshPermissions, user])

  // 建立SSE连接
  const connect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    console.log('建立权限通知连接...')
    
    const eventSource = new EventSource('/api/admin/permissions/notifications')
    eventSourceRef.current = eventSource

    eventSource.onopen = () => {
      console.log('权限通知连接已建立')
      setState(prev => ({
        ...prev,
        connected: true,
        connectionError: null
      }))
    }

    eventSource.onmessage = (event) => {
      try {
        const data: PermissionChangeEvent = JSON.parse(event.data)
        handlePermissionChange(data)
      } catch (error) {
        console.error('解析权限变更事件失败:', error)
      }
    }

    eventSource.onerror = (error) => {
      console.error('权限通知连接错误:', error)
      setState(prev => ({
        ...prev,
        connected: false,
        connectionError: '连接中断'
      }))

      // 自动重连（指数退避）
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      
      reconnectTimeoutRef.current = setTimeout(() => {
        console.log('尝试重新连接权限通知...')
        connect()
      }, Math.min(1000 * Math.pow(2, state.eventCount % 5), 30000)) // 最大30秒
    }
  }, [handlePermissionChange, state.eventCount])

  // 断开连接
  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
    
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }

    setState(prev => ({
      ...prev,
      connected: false,
      connectionError: null
    }))
  }, [])

  // 手动刷新权限
  const refreshUserPermissions = useCallback(async () => {
    try {
      await refreshPermissions()
      console.log('手动刷新权限成功')
    } catch (error) {
      console.error('手动刷新权限失败:', error)
    }
  }, [refreshPermissions])

  // 组件挂载时自动连接
  useEffect(() => {
    connect()
    
    return () => {
      disconnect()
    }
  }, []) // 只在挂载时连接一次

  // 页面可见性变化时重连
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible' && !state.connected) {
        console.log('页面重新可见，重新连接权限通知')
        connect()
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [connect, state.connected])

  return {
    ...state,
    connect,
    disconnect,
    refreshUserPermissions
  }
}

// 判断当前用户是否受权限变更影响
function shouldRefreshUserPermissions(
  event: PermissionChangeEvent, 
  user: any
): boolean {
  if (!event.data || !user) return false

  const { targetType, target } = event.data

  switch (targetType) {
    case 'role':
      // 如果变更的是当前用户的角色
      return user.role === target

    case 'courier-level':
      // 如果变更的是当前用户的信使等级
      return user.courierInfo?.level === parseInt(target)

    case 'system':
      // 系统级变更影响所有用户
      return true

    default:
      return false
  }
}