/**
 * WebSocket Provider Component
 * WebSocket集成提供者组件
 */

'use client'

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { wsManager, SERVICE_EVENTS } from '@/lib/services'
import { useAuth } from '@/contexts/auth-context-new'
import { toast } from '@/components/ui/use-toast'

interface WebSocketContextType {
  isConnected: boolean
  connectionStatus: 'disconnected' | 'connecting' | 'connected' | 'reconnecting'
  subscribe: (eventType: string, handler: Function) => void
  unsubscribe: (eventType: string, handler: Function) => void
  send: (type: string, data: any) => void
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined)

interface WebSocketProviderProps {
  children: ReactNode
}

export function WebSocketProvider({ children }: WebSocketProviderProps) {
  const { isAuthenticated, user } = useAuth()
  const [isConnected, setIsConnected] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState<WebSocketContextType['connectionStatus']>('disconnected')

  useEffect(() => {
    if (!isAuthenticated || !user) {
      // 用户未登录时断开连接
      wsManager.disconnect()
      setIsConnected(false)
      setConnectionStatus('disconnected')
      return
    }

    let reconnectAttempts = 0
    const maxReconnectAttempts = 5
    const reconnectDelay = 1000

    const connect = async () => {
      try {
        setConnectionStatus('connecting')
        await wsManager.connect()
        setIsConnected(true)
        setConnectionStatus('connected')
        reconnectAttempts = 0
        
        console.log('WebSocket connected successfully')
        
        // 发送在线状态
        wsManager.send(SERVICE_EVENTS.USER_ONLINE, {
          userId: user.id,
          timestamp: new Date().toISOString()
        })
        
      } catch (error) {
        console.error('WebSocket connection failed:', error)
        setIsConnected(false)
        
        // 尝试重连
        if (reconnectAttempts < maxReconnectAttempts) {
          setConnectionStatus('reconnecting')
          reconnectAttempts++
          
          setTimeout(() => {
            console.log(`Attempting to reconnect (${reconnectAttempts}/${maxReconnectAttempts})`)
            connect()
          }, reconnectDelay * Math.pow(2, reconnectAttempts - 1))
        } else {
          setConnectionStatus('disconnected')
          toast({
            title: '连接失败',
            description: 'WebSocket连接失败，某些实时功能可能不可用',
            variant: 'destructive'
          })
        }
      }
    }

    // 建立连接
    connect()

    // 设置全局事件监听
    const setupGlobalEventHandlers = () => {
      // 信件相关事件
      wsManager.subscribe(SERVICE_EVENTS.LETTER_DELIVERED, (data: any) => {
        toast({
          title: '信件已送达',
          description: `您的信件「${data.title || '无标题'}」已成功送达`,
        })
      })

      wsManager.subscribe(SERVICE_EVENTS.LETTER_READ, (data: any) => {
        toast({
          title: '信件已阅读',
          description: `您的信件「${data.title || '无标题'}」已被阅读`,
        })
      })

      wsManager.subscribe(SERVICE_EVENTS.LETTER_REPLIED, (data: any) => {
        toast({
          title: '收到回信',
          description: `${data.sender_name} 回复了您的信件`,
        })
      })

      // 信使任务事件
      wsManager.subscribe(SERVICE_EVENTS.TASK_ASSIGNED, (data: any) => {
        if (data.courierId === user?.id) {
          toast({
            title: '新任务分配',
            description: `您有一个新的投递任务: ${data.letter_code}`,
          })
        }
      })

      wsManager.subscribe(SERVICE_EVENTS.TASK_COMPLETED, (data: any) => {
        if (data.sender_id === user?.id) {
          toast({
            title: '投递完成',
            description: `您的信件 ${data.letter_code} 已投递完成`,
          })
        }
      })

      // 用户任命事件
      wsManager.subscribe(SERVICE_EVENTS.USER_APPOINTED, (data: any) => {
        if (data.userId === user?.id) {
          toast({
            title: '角色变更',
            description: `恭喜！您已被任命为${data.new_role}`,
          })
          
          // 触发用户信息刷新
          window.dispatchEvent(new CustomEvent('auth:role-changed', { 
            detail: { newRole: data.new_role } 
          }))
        }
      })

      // 系统通知事件
      wsManager.subscribe(SERVICE_EVENTS.NOTIFICATION, (data: any) => {
        toast({
          title: data.title,
          description: data.content,
          variant: data.type === 'error' ? 'destructive' : 'default'
        })
      })

      // 系统维护通知
      wsManager.subscribe(SERVICE_EVENTS.SYSTEM_MAINTENANCE, (data: any) => {
        if (data.type === 'start') {
          toast({
            title: '系统维护',
            description: `系统将于 ${data.start_time} 进行维护，预计持续 ${data.duration}`,
            variant: 'destructive'
          })
        }
      })

      // 安全警报
      wsManager.subscribe(SERVICE_EVENTS.SECURITY_ALERT, (data: any) => {
        if (data.userId === user?.id) {
          toast({
            title: '安全提醒',
            description: data.message,
            variant: 'destructive'
          })
        }
      })
    }

    setupGlobalEventHandlers()

    // 清理函数
    return () => {
      // 发送离线状态
      if (isConnected) {
        wsManager.send(SERVICE_EVENTS.USER_OFFLINE, {
          userId: user.id,
          timestamp: new Date().toISOString()
        })
      }
      
      wsManager.disconnect()
      setIsConnected(false)
      setConnectionStatus('disconnected')
    }
  }, [isAuthenticated, user?.id])

  // 监听页面可见性变化
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden) {
        // 页面隐藏时发送离线状态
        if (isConnected && user) {
          wsManager.send(SERVICE_EVENTS.USER_OFFLINE, {
            userId: user.id,
            timestamp: new Date().toISOString()
          })
        }
      } else {
        // 页面显示时发送在线状态
        if (isConnected && user) {
          wsManager.send(SERVICE_EVENTS.USER_ONLINE, {
            userId: user.id,
            timestamp: new Date().toISOString()
          })
        }
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [isConnected, user?.id])

  // 监听网络状态变化
  useEffect(() => {
    const handleOnline = () => {
      if (isAuthenticated && !isConnected) {
        // 网络恢复时尝试重连
        console.log('Network restored, attempting to reconnect WebSocket')
        wsManager.connect().catch(console.error)
      }
    }

    const handleOffline = () => {
      console.log('Network lost, WebSocket will be disconnected')
      setConnectionStatus('disconnected')
    }

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [isAuthenticated, isConnected])

  const subscribe = (eventType: string, handler: Function) => {
    wsManager.subscribe(eventType, handler)
  }

  const unsubscribe = (eventType: string, handler: Function) => {
    wsManager.unsubscribe(eventType, handler)
  }

  const send = (type: string, data: any) => {
    wsManager.send(type, data)
  }

  const value: WebSocketContextType = {
    isConnected,
    connectionStatus,
    subscribe,
    unsubscribe,
    send
  }

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  )
}

export function useWebSocket() {
  const context = useContext(WebSocketContext)
  if (context === undefined) {
    throw new Error('useWebSocket must be used within a WebSocketProvider')
  }
  return context
}

// WebSocket连接状态指示器组件
export function WebSocketStatus() {
  const { connectionStatus, isConnected } = useWebSocket()
  
  if (connectionStatus === 'connected') {
    return null // 连接正常时不显示
  }

  const getStatusInfo = () => {
    switch (connectionStatus) {
      case 'connecting':
        return { text: '正在连接...', color: 'bg-yellow-500' }
      case 'reconnecting':
        return { text: '重新连接中...', color: 'bg-orange-500' }
      case 'disconnected':
        return { text: '连接断开', color: 'bg-red-500' }
      default:
        return { text: '未知状态', color: 'bg-gray-500' }
    }
  }

  const { text, color } = getStatusInfo()

  return (
    <div className="fixed bottom-4 right-4 z-50">
      <div className={`px-3 py-2 rounded-md text-white text-sm flex items-center space-x-2 ${color}`}>
        <div className={`w-2 h-2 rounded-full bg-white animate-pulse`} />
        <span>{text}</span>
      </div>
    </div>
  )
}

export default WebSocketProvider