'use client'

import React, { createContext, useContext, useEffect, useRef, useState, useCallback } from 'react'
import { useToken } from './token-context'
import { TokenManager } from '@/lib/auth/cookie-token-manager'

// WebSocket消息类型
export type EventType = 
  | 'LETTER_STATUS_UPDATE'
  | 'LETTER_CREATED'
  | 'LETTER_READ'
  | 'LETTER_DELIVERED'
  | 'COURIER_LOCATION_UPDATE'
  | 'NEW_TASK_ASSIGNMENT'
  | 'TASK_STATUS_UPDATE'
  | 'COURIER_ONLINE'
  | 'COURIER_OFFLINE'
  | 'USER_ONLINE'
  | 'USER_OFFLINE'
  | 'NOTIFICATION'
  | 'SYSTEM_MESSAGE'
  | 'HEARTBEAT'
  | 'ERROR'
  | 'CONNECTED'
  | 'DISCONNECTED'

// WebSocket消息结构
export interface WebSocketMessage {
  id: string
  type: EventType
  data: Record<string, any>
  timestamp: string
  user_id?: string
  room?: string
}

// 连接状态
export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'reconnecting' | 'error'

// WebSocket上下文类型
interface WebSocketContextType {
  connectionStatus: ConnectionStatus
  lastMessage: WebSocketMessage | null
  sendMessage: (type: EventType, data: Record<string, any>, room?: string) => void
  sendDirectMessage: (targetUserId: string, type: EventType, data: Record<string, any>) => void
  subscribe: (eventType: EventType, handler: (message: WebSocketMessage) => void) => () => void
  subscribeToRoom: (room: string, handler: (message: WebSocketMessage) => void) => () => void
  connect: () => void
  disconnect: () => void
  isConnected: boolean
  connectionInfo: any
  stats: any
}

const WebSocketContext = createContext<WebSocketContextType | null>(null)

// WebSocket配置
const WS_CONFIG = {
  reconnectInterval: 3000,
  maxReconnectAttempts: 10,
  heartbeatInterval: 30000,
  connectionTimeout: 10000,
}

interface WebSocketProviderProps {
  children: React.ReactNode
}

export function WebSocketProvider({ children }: WebSocketProviderProps) {
  const { token, userId } = useToken()
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('disconnected')
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null)
  const [connectionInfo, setConnectionInfo] = useState<any>(null)
  const [stats, setStats] = useState<any>(null)

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>()
  const heartbeatTimeoutRef = useRef<NodeJS.Timeout>()
  const reconnectAttemptsRef = useRef(0)
  const eventHandlersRef = useRef<Map<EventType, Set<(message: WebSocketMessage) => void>>>(new Map())
  const roomHandlersRef = useRef<Map<string, Set<(message: WebSocketMessage) => void>>>(new Map())

  // 获取WebSocket URL
  const getWebSocketUrl = useCallback(() => {
    if (typeof window === 'undefined') return null
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    // Remove protocol and use the base URL
    const baseUrl = apiUrl.replace(/^https?:/, '')
    const token = TokenManager.get()
    
    if (!token) {
      console.warn('No token available for WebSocket connection')
      return null
    }
    
    // Check if baseUrl already contains /api/v1 to avoid duplication
    const wsPath = baseUrl.includes('/api/v1') ? '/ws/connect' : '/api/v1/ws/connect'
    return `${protocol}${baseUrl}${wsPath}?token=${encodeURIComponent(token)}`
  }, [])

  // 发送心跳
  const sendHeartbeat = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      const heartbeatMessage: WebSocketMessage = {
        id: `heartbeat_${Date.now()}`,
        type: 'HEARTBEAT',
        data: { client_time: new Date().toISOString() },
        timestamp: new Date().toISOString(),
      }
      wsRef.current.send(JSON.stringify(heartbeatMessage))
    }
  }, [])

  // 启动心跳
  const startHeartbeat = useCallback(() => {
    heartbeatTimeoutRef.current = setInterval(sendHeartbeat, WS_CONFIG.heartbeatInterval)
  }, [sendHeartbeat])

  // 停止心跳
  const stopHeartbeat = useCallback(() => {
    if (heartbeatTimeoutRef.current) {
      clearInterval(heartbeatTimeoutRef.current)
      heartbeatTimeoutRef.current = undefined
    }
  }, [])

  // 处理接收到的消息
  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const message: WebSocketMessage = JSON.parse(event.data)
      setLastMessage(message)

      // 处理特殊消息类型
      switch (message.type) {
        case 'CONNECTED':
          setConnectionStatus('connected')
          setConnectionInfo(message.data)
          reconnectAttemptsRef.current = 0
          startHeartbeat()
          console.log('WebSocket连接成功', message.data)
          break
          
        case 'HEARTBEAT':
          // 心跳响应，更新延迟信息
          const clientTime = new Date(message.data.client_time).getTime()
          const serverTime = new Date(message.data.server_time).getTime()
          const latency = Date.now() - clientTime
          console.debug('心跳延迟:', latency, 'ms')
          break
          
        case 'ERROR':
          console.error('WebSocket错误:', message.data)
          break
          
        default:
          break
      }

      // 触发事件处理器
      const typeHandlers = eventHandlersRef.current.get(message.type)
      if (typeHandlers) {
        typeHandlers.forEach(handler => handler(message))
      }

      // 触发房间处理器
      if (message.room) {
        const roomHandlers = roomHandlersRef.current.get(message.room)
        if (roomHandlers) {
          roomHandlers.forEach(handler => handler(message))
        }
      }
    } catch (error) {
      console.error('解析WebSocket消息失败:', error)
    }
  }, [startHeartbeat])

  // 连接WebSocket
  const connect = useCallback(() => {
    if (typeof window === 'undefined') {
      setConnectionStatus('disconnected')
      return
    }

    const token = TokenManager.get()
    if (!userId || !token) {
      console.warn('用户未登录，无法建立WebSocket连接')
      setConnectionStatus('disconnected')
      return
    }

    // 验证token是否过期
    if (TokenManager.isExpired(token)) {
      console.warn('Token已过期，无法建立WebSocket连接')
      setConnectionStatus('disconnected')
      return
    }

    if (wsRef.current?.readyState === WebSocket.OPEN) {
      console.warn('WebSocket已连接')
      return
    }

    try {
      setConnectionStatus('connecting')
      const wsUrl = getWebSocketUrl()
      if (!wsUrl) {
        setConnectionStatus('disconnected')
        return
      }
      console.log('正在连接WebSocket:', wsUrl)
      
      wsRef.current = new WebSocket(wsUrl)

      wsRef.current.onopen = () => {
        console.log('WebSocket连接已建立')
        setConnectionStatus('connected')
        reconnectAttemptsRef.current = 0
        startHeartbeat()
      }

      wsRef.current.onmessage = handleMessage

      wsRef.current.onclose = (event) => {
        console.log('WebSocket连接已关闭:', event.code, event.reason)
        setConnectionStatus('disconnected')
        stopHeartbeat()
        
        // 检查是否仍然有有效的认证
        const token = TokenManager.get()
        if (!token || TokenManager.isExpired(token)) {
          console.log('📡 无有效token，停止WebSocket重连')
          reconnectAttemptsRef.current = 0
          return
        }
        
        // 自动重连
        if (!event.wasClean && reconnectAttemptsRef.current < WS_CONFIG.maxReconnectAttempts && userId) {
          setConnectionStatus('reconnecting')
          reconnectAttemptsRef.current++
          console.log(`尝试重连 (${reconnectAttemptsRef.current}/${WS_CONFIG.maxReconnectAttempts})`)
          
          reconnectTimeoutRef.current = setTimeout(() => {
            connect()
          }, WS_CONFIG.reconnectInterval)
        }
      }

      wsRef.current.onerror = (error) => {
        console.error('WebSocket错误:', error)
        setConnectionStatus('error')
        
        // 不要因为WebSocket错误就触发logout
        // WebSocket连接失败不应该影响用户的认证状态
        console.log('📡 WebSocket连接失败，但不影响用户认证状态')
      }

    } catch (error) {
      console.error('创建WebSocket连接失败:', error)
      setConnectionStatus('error')
    }
  }, [userId, getWebSocketUrl, handleMessage, stopHeartbeat])

  // 断开连接
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = undefined
    }
    
    stopHeartbeat()
    
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    
    setConnectionStatus('disconnected')
    reconnectAttemptsRef.current = 0
  }, [stopHeartbeat])

  // 发送消息
  const sendMessage = useCallback((type: EventType, data: Record<string, any>, room?: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      const message: WebSocketMessage = {
        id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        type,
        data,
        timestamp: new Date().toISOString(),
        room,
      }
      wsRef.current.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket未连接，无法发送消息')
    }
  }, [])

  // 发送定向消息（通过HTTP API）
  const sendDirectMessage = useCallback(async (targetUserId: string, type: EventType, data: Record<string, any>) => {
    try {
      const response = await fetch('/api/v1/ws/direct', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${TokenManager.get()}`,
        },
        body: JSON.stringify({
          target_user_id: targetUserId,
          type,
          data,
        }),
      })

      if (!response.ok) {
        throw new Error('发送定向消息失败')
      }
    } catch (error) {
      console.error('发送定向消息失败:', error)
    }
  }, [])

  // 订阅事件
  const subscribe = useCallback((eventType: EventType, handler: (message: WebSocketMessage) => void) => {
    if (!eventHandlersRef.current.has(eventType)) {
      eventHandlersRef.current.set(eventType, new Set())
    }
    
    const handlers = eventHandlersRef.current.get(eventType)!
    handlers.add(handler)

    // 返回取消订阅函数
    return () => {
      handlers.delete(handler)
      if (handlers.size === 0) {
        eventHandlersRef.current.delete(eventType)
      }
    }
  }, [])

  // 订阅房间消息
  const subscribeToRoom = useCallback((room: string, handler: (message: WebSocketMessage) => void) => {
    if (!roomHandlersRef.current.has(room)) {
      roomHandlersRef.current.set(room, new Set())
    }
    
    const handlers = roomHandlersRef.current.get(room)!
    handlers.add(handler)

    // 返回取消订阅函数
    return () => {
      handlers.delete(handler)
      if (handlers.size === 0) {
        roomHandlersRef.current.delete(room)
      }
    }
  }, [])

  // 用户登录时自动连接 - 只在客户端执行
  useEffect(() => {
    if (typeof window === 'undefined') return
    
    // 延迟连接，避免认证初始化未完成
    const connectTimer = setTimeout(() => {
      if (userId && TokenManager.get()) {
        connect()
      }
    }, 1000)

    return () => {
      clearTimeout(connectTimer)
      if (typeof window !== 'undefined') {
        disconnect()
      }
    }
  }, [userId, connect, disconnect])

  // 页面卸载时断开连接 - 只在客户端执行
  useEffect(() => {
    if (typeof window === 'undefined') return

    const handleBeforeUnload = () => {
      disconnect()
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload)
      if (typeof window !== 'undefined') {
        disconnect()
      }
    }
  }, [disconnect])

  // 定期获取统计信息 - 只在客户端执行
  useEffect(() => {
    if (typeof window === 'undefined' || connectionStatus !== 'connected' || !TokenManager.get()) return

    const fetchStats = async () => {
      try {
        const token = TokenManager.get()
        if (!token || TokenManager.isExpired(token)) {
          console.log('📡 Token已过期，跳过WebSocket统计获取')
          return
        }
        
        // Temporarily disabled - endpoint not implemented
        // const response = await fetch('/api/v1/ws/stats', {
        //   headers: {
        //     'Authorization': `Bearer ${token}`,
        //   },
        // })
        
        // if (response.ok) {
        //   const statsData = await response.json()
        //   setStats(statsData)
        // } else if (response.status === 401) {
        //   // 401错误，但不要影响用户认证状态
        //   console.log('📡 WebSocket统计API返回401，但不影响用户认证状态')
        //   // 清除统计数据，但不清除认证状态
        //   setStats(null)
        // } else if (response.status === 500) {
        //   // 500错误可能是后端问题，不影响认证
        //   console.log('📡 WebSocket统计API返回500，可能是后端服务问题')
        //   setStats(null)
        // }
      } catch (error) {
        // 网络错误或其他问题，不影响认证状态
        console.log('📡 获取WebSocket统计信息失败，但不影响用户认证:', error)
        setStats(null)
      }
    }

    fetchStats()
    const interval = setInterval(fetchStats, 60000) // 每分钟更新一次

    return () => clearInterval(interval)
  }, [connectionStatus])

  const value: WebSocketContextType = {
    connectionStatus,
    lastMessage,
    sendMessage,
    sendDirectMessage,
    subscribe,
    subscribeToRoom,
    connect,
    disconnect,
    isConnected: connectionStatus === 'connected',
    connectionInfo,
    stats,
  }

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  )
}

// Hook for using WebSocket context
export function useWebSocket() {
  const context = useContext(WebSocketContext)
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider')
  }
  return context
}