import { useEffect, useState, useCallback } from 'react'
import { useWebSocket, WebSocketMessage, EventType } from '@/contexts/websocket-context'
import { useAuth } from '@/contexts/auth-context-new'

// 实时通知Hook
export function useRealtimeNotifications() {
  const { subscribe } = useWebSocket()
  const [notifications, setNotifications] = useState<any[]>([])
  const [unreadCount, setUnreadCount] = useState(0)

  useEffect(() => {
    const unsubscribe = subscribe('NOTIFICATION', (message: WebSocketMessage) => {
      const notification = {
        id: message.data.notification_id,
        title: message.data.title,
        content: message.data.content,
        type: message.data.type,
        priority: message.data.priority,
        actionUrl: message.data.action_url,
        createdAt: new Date(message.data.createdAt),
        read: false,
      }

      setNotifications(prev => [notification, ...prev.slice(0, 49)]) // 保留最近50条
      setUnreadCount(prev => prev + 1)
    })

    return unsubscribe
  }, [subscribe])

  const markAsRead = useCallback((notificationId: string) => {
    setNotifications(prev => 
      prev.map(n => n.id === notificationId ? { ...n, read: true } : n)
    )
    setUnreadCount(prev => Math.max(0, prev - 1))
  }, [])

  const markAllAsRead = useCallback(() => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })))
    setUnreadCount(0)
  }, [])

  const clearNotifications = useCallback(() => {
    setNotifications([])
    setUnreadCount(0)
  }, [])

  return {
    notifications,
    unreadCount,
    markAsRead,
    markAllAsRead,
    clearNotifications,
  }
}

// 信件状态实时更新Hook
export function useLetterStatusUpdates(letterId?: string) {
  const { subscribe, subscribeToRoom } = useWebSocket()
  const [statusUpdates, setStatusUpdates] = useState<any[]>([])
  const [currentStatus, setCurrentStatus] = useState<string | null>(null)

  useEffect(() => {
    const unsubscribes: (() => void)[] = []

    // 订阅全局信件状态更新
    unsubscribes.push(
      subscribe('LETTER_STATUS_UPDATE', (message: WebSocketMessage) => {
        if (!letterId || message.data.letterId === letterId) {
          const update = {
            id: message.id,
            letterId: message.data.letterId,
            code: message.data.code,
            status: message.data.status,
            location: message.data.location,
            courierId: message.data.courierId,
            courierName: message.data.courier_name,
            updatedAt: new Date(message.data.updatedAt),
            message: message.data.message,
          }

          setStatusUpdates(prev => [update, ...prev.slice(0, 19)]) // 保留最近20条
          setCurrentStatus(update.status)
        }
      })
    )

    // 如果指定了特定信件，订阅信件专用房间
    if (letterId) {
      unsubscribes.push(
        subscribeToRoom(`letter:${letterId}`, (message: WebSocketMessage) => {
          if (message.type === 'LETTER_STATUS_UPDATE') {
            const update = {
              id: message.id,
              letterId: message.data.letterId,
              code: message.data.code,
              status: message.data.status,
              location: message.data.location,
              courierId: message.data.courierId,
              courierName: message.data.courier_name,
              updatedAt: new Date(message.data.updatedAt),
              message: message.data.message,
            }

            setStatusUpdates(prev => [update, ...prev.slice(0, 19)])
            setCurrentStatus(update.status)
          }
        })
      )
    }

    return () => {
      unsubscribes.forEach(unsubscribe => unsubscribe())
    }
  }, [subscribe, subscribeToRoom, letterId])

  return {
    statusUpdates,
    currentStatus,
  }
}

// 信使实时位置Hook
export function useCourierTracking() {
  const { subscribe } = useWebSocket()
  const [courierLocations, setCourierLocations] = useState<Map<string, any>>(new Map())
  const [onlineCouriers, setOnlineCouriers] = useState<Set<string>>(new Set())

  useEffect(() => {
    const unsubscribes: (() => void)[] = []

    // 信使位置更新
    unsubscribes.push(
      subscribe('COURIER_LOCATION_UPDATE', (message: WebSocketMessage) => {
        const location = {
          courierId: message.data.courierId,
          courierName: message.data.courier_name,
          latitude: message.data.latitude,
          longitude: message.data.longitude,
          accuracy: message.data.accuracy,
          timestamp: new Date(message.data.timestamp),
          status: message.data.status,
        }

        setCourierLocations(prev => new Map(prev.set(location.courierId, location)))
        
        if (location.status === 'online') {
          setOnlineCouriers(prev => new Set(prev.add(location.courierId)))
        }
      })
    )

    // 信使上线
    unsubscribes.push(
      subscribe('COURIER_ONLINE', (message: WebSocketMessage) => {
        setOnlineCouriers(prev => new Set(prev.add(message.data.courierId)))
      })
    )

    // 信使离线
    unsubscribes.push(
      subscribe('COURIER_OFFLINE', (message: WebSocketMessage) => {
        setOnlineCouriers(prev => {
          const newSet = new Set(prev)
          newSet.delete(message.data.courierId)
          return newSet
        })
      })
    )

    return () => {
      unsubscribes.forEach(unsubscribe => unsubscribe())
    }
  }, [subscribe])

  const getCourierLocation = useCallback((courierId: string) => {
    return courierLocations.get(courierId)
  }, [courierLocations])

  const isCourierOnline = useCallback((courierId: string) => {
    return onlineCouriers.has(courierId)
  }, [onlineCouriers])

  return {
    courierLocations: Array.from(courierLocations.values()),
    onlineCouriers: Array.from(onlineCouriers),
    getCourierLocation,
    isCourierOnline,
  }
}

// 任务分配实时更新Hook
export function useTaskAssignments() {
  const { subscribe } = useWebSocket()
  const { user } = useAuth()
  const [newTasks, setNewTasks] = useState<any[]>([])
  const [taskUpdates, setTaskUpdates] = useState<any[]>([])

  useEffect(() => {
    if (!user) return

    const unsubscribes: (() => void)[] = []

    // 新任务分配
    unsubscribes.push(
      subscribe('NEW_TASK_ASSIGNMENT', (message: WebSocketMessage) => {
        if (message.data.courierId === user.id) {
          const task = {
            id: message.data.taskId,
            courierId: message.data.courierId,
            letterId: message.data.letterId,
            priority: message.data.priority,
            deadline: new Date(message.data.deadline),
            pickupLocation: message.data.pickup_location,
            deliveryLocation: message.data.delivery_location,
            reward: message.data.reward,
            assignedAt: new Date(message.data.assigned_at),
          }

          setNewTasks(prev => [task, ...prev.slice(0, 9)]) // 保留最近10条
        }
      })
    )

    // 任务状态更新
    unsubscribes.push(
      subscribe('TASK_STATUS_UPDATE', (message: WebSocketMessage) => {
        const update = {
          id: message.id,
          taskId: message.data.taskId,
          status: message.data.status,
          updatedAt: new Date(message.data.updatedAt),
          message: message.data.message,
        }

        setTaskUpdates(prev => [update, ...prev.slice(0, 19)])
      })
    )

    return () => {
      unsubscribes.forEach(unsubscribe => unsubscribe())
    }
  }, [subscribe, user])

  const markTaskAsRead = useCallback((taskId: string) => {
    setNewTasks(prev => prev.filter(task => task.id !== taskId))
  }, [])

  return {
    newTasks,
    taskUpdates,
    markTaskAsRead,
  }
}

// 用户在线状态Hook
export function useUserPresence() {
  const { subscribe } = useWebSocket()
  const [onlineUsers, setOnlineUsers] = useState<Map<string, any>>(new Map())

  useEffect(() => {
    const unsubscribes: (() => void)[] = []

    // 用户上线
    unsubscribes.push(
      subscribe('USER_ONLINE', (message: WebSocketMessage) => {
        const user = {
          userId: message.data.user.userId,
          username: message.data.user.username,
          status: message.data.user.status,
          lastSeen: new Date(message.data.user.last_seen),
        }

        setOnlineUsers(prev => new Map(prev.set(user.userId, user)))
      })
    )

    // 用户离线
    unsubscribes.push(
      subscribe('USER_OFFLINE', (message: WebSocketMessage) => {
        const userId = message.data.user.userId
        setOnlineUsers(prev => {
          const newMap = new Map(prev)
          newMap.delete(userId)
          return newMap
        })
      })
    )

    return () => {
      unsubscribes.forEach(unsubscribe => unsubscribe())
    }
  }, [subscribe])

  const isUserOnline = useCallback((userId: string) => {
    return onlineUsers.has(userId)
  }, [onlineUsers])

  const getUserStatus = useCallback((userId: string) => {
    return onlineUsers.get(userId)
  }, [onlineUsers])

  return {
    onlineUsers: Array.from(onlineUsers.values()),
    isUserOnline,
    getUserStatus,
  }
}

// 系统消息Hook
export function useSystemMessages() {
  const { subscribe } = useWebSocket()
  const [systemMessages, setSystemMessages] = useState<any[]>([])

  useEffect(() => {
    const unsubscribe = subscribe('SYSTEM_MESSAGE', (message: WebSocketMessage) => {
      const systemMessage = {
        id: message.data.message_id,
        level: message.data.level,
        title: message.data.title,
        content: message.data.content,
        createdAt: new Date(message.data.createdAt),
      }

      setSystemMessages(prev => [systemMessage, ...prev.slice(0, 9)]) // 保留最近10条
    })

    return unsubscribe
  }, [subscribe])

  const dismissMessage = useCallback((messageId: string) => {
    setSystemMessages(prev => prev.filter(msg => msg.id !== messageId))
  }, [])

  return {
    systemMessages,
    dismissMessage,
  }
}

// 通用实时数据Hook
export function useRealtimeData<T>(
  eventType: EventType,
  transformer?: (data: any) => T,
  filter?: (message: WebSocketMessage) => boolean
) {
  const { subscribe } = useWebSocket()
  const [data, setData] = useState<T[]>([])
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null)

  useEffect(() => {
    const unsubscribe = subscribe(eventType, (message: WebSocketMessage) => {
      if (filter && !filter(message)) {
        return
      }

      const transformedData = transformer ? transformer(message.data) : message.data as T
      
      setData(prev => [transformedData, ...prev.slice(0, 99)]) // 保留最近100条
      setLastUpdate(new Date())
    })

    return unsubscribe
  }, [subscribe, eventType, transformer, filter])

  return {
    data,
    lastUpdate,
  }
}