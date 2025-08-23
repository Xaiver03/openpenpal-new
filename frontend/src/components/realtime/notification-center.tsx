'use client'

import React, { useState } from 'react'
import { Bell, X, Check, CheckCheck, Trash2, Info, AlertTriangle, CheckCircle, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { 
  DropdownMenu, 
  DropdownMenuContent, 
  DropdownMenuHeader, 
  DropdownMenuTrigger,
  DropdownMenuItem,
  DropdownMenuSeparator 
} from '@/components/ui/dropdown-menu'
import { useRealtimeNotifications } from '@/hooks/use-realtime'
import { SafeTimestamp } from '@/components/ui/safe-timestamp'

interface NotificationCenterProps {
  className?: string
}

export function NotificationCenter({ className = '' }: NotificationCenterProps) {
  const { 
    notifications, 
    unreadCount, 
    markAsRead, 
    markAllAsRead, 
    clearNotifications 
  } = useRealtimeNotifications()
  
  const [isOpen, setIsOpen] = useState(false)
  const [bellAnimation, setBellAnimation] = useState('')
  const [lastNotificationId, setLastNotificationId] = useState<string | null>(null)

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'success':
        return <CheckCircle className="w-4 h-4 text-green-500" />
      case 'warning':
        return <AlertTriangle className="w-4 h-4 text-yellow-500" />
      case 'error':
        return <XCircle className="w-4 h-4 text-red-500" />
      default:
        return <Info className="w-4 h-4 text-blue-500" />
    }
  }

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high':
        return 'bg-red-100 text-red-800 border-red-200'
      case 'normal':
        return 'bg-blue-100 text-blue-800 border-blue-200'
      case 'low':
        return 'bg-gray-100 text-gray-800 border-gray-200'
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  const handleNotificationClick = (notification: any) => {
    if (!notification.read) {
      markAsRead(notification.id)
    }
    
    if (notification.actionUrl) {
      window.location.href = notification.actionUrl
    }
  }

  // 监听新通知，触发Bell动画
  React.useEffect(() => {
    if (notifications.length > 0) {
      const latestNotification = notifications[0]
      if (latestNotification.id !== lastNotificationId) {
        setLastNotificationId(latestNotification.id)
        setBellAnimation('bell-ring')
        
        // 根据优先级选择不同动画
        if (latestNotification.priority === 'high') {
          setBellAnimation('bell-pulse')
        }
        
        setTimeout(() => setBellAnimation(''), 2000)
      }
    }
  }, [notifications, lastNotificationId])

  return (
    <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className={`relative ${className}`}
        >
          <Bell className={`w-5 h-5 ${bellAnimation}`} />
          {unreadCount > 0 && (
            <Badge 
              variant="destructive" 
              className="absolute -top-1 -right-1 w-5 h-5 text-xs flex items-center justify-center p-0"
            >
              {unreadCount > 99 ? '99+' : unreadCount}
            </Badge>
          )}
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" className="w-80 max-h-96 overflow-hidden">
        <DropdownMenuHeader className="px-4 py-2 border-b">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold">通知中心</h3>
            <div className="flex items-center gap-2">
              {unreadCount > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={markAllAsRead}
                  className="h-6 px-2 text-xs"
                >
                  <CheckCheck className="w-3 h-3 mr-1" />
                  全部已读
                </Button>
              )}
              {notifications.length > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={clearNotifications}
                  className="h-6 px-2 text-xs text-red-600 hover:text-red-700"
                >
                  <Trash2 className="w-3 h-3 mr-1" />
                  清空
                </Button>
              )}
            </div>
          </div>
        </DropdownMenuHeader>

        <div className="max-h-80 overflow-y-auto">
          {notifications.length === 0 ? (
            <div className="p-8 text-center text-gray-500">
              <Bell className="w-8 h-8 mx-auto mb-2 opacity-50" />
              <p className="text-sm">暂无通知</p>
            </div>
          ) : (
            <div className="space-y-1 p-2">
              {notifications.map((notification, index) => (
                <Card
                  key={notification.id}
                  className={`p-3 cursor-pointer transition-all duration-300 hover:bg-gray-50 hover:shadow-md hover:scale-[1.02] ${
                    !notification.read ? 'bg-blue-50 border-blue-200 shadow-sm' : ''
                  } ${index === 0 && notification.id === lastNotificationId ? 'notification-bounce-in' : ''}`}
                  onClick={() => handleNotificationClick(notification)}
                >
                  <div className="flex items-start gap-3">
                    <div className="flex-shrink-0 mt-0.5">
                      {getNotificationIcon(notification.type)}
                    </div>
                    
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between mb-1">
                        <h4 className="text-sm font-medium text-gray-900 truncate">
                          {notification.title}
                        </h4>
                        {!notification.read && (
                          <div className="flex-shrink-0 ml-2">
                            <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
                          </div>
                        )}
                      </div>
                      
                      <p className="text-xs text-gray-600 mb-2 line-clamp-2">
                        {notification.content}
                      </p>
                      
                      <div className="flex items-center justify-between">
                        <SafeTimestamp 
                          date={notification.created_at} 
                          format="relative" 
                          fallback="刚刚"
                          className="text-xs text-gray-500"
                        />
                        
                        {notification.priority !== 'normal' && (
                          <Badge 
                            variant="outline" 
                            className={`text-xs ${getPriorityColor(notification.priority)}`}
                          >
                            {notification.priority === 'high' ? '紧急' : '低优先级'}
                          </Badge>
                        )}
                      </div>
                    </div>
                  </div>

                  {!notification.read && (
                    <div className="flex justify-end mt-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation()
                          markAsRead(notification.id)
                        }}
                        className="h-6 px-2 text-xs"
                      >
                        <Check className="w-3 h-3 mr-1" />
                        标记已读
                      </Button>
                    </div>
                  )}
                </Card>
              ))}
            </div>
          )}
        </div>

        {notifications.length > 0 && (
          <>
            <DropdownMenuSeparator />
            <div className="p-2">
              <Button
                variant="ghost"
                size="sm"
                className="w-full justify-center text-xs"
                onClick={() => {
                  setIsOpen(false)
                  // 这里可以跳转到完整的通知页面
                }}
              >
                查看所有通知
              </Button>
            </div>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

// 浮动通知组件
export function FloatingNotification({ 
  notification, 
  onClose, 
  duration = 5000 
}: {
  notification: any
  onClose: () => void
  duration?: number
}) {
  const [isVisible, setIsVisible] = useState(true)
  const [isExiting, setIsExiting] = useState(false)

  React.useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        handleClose()
      }, duration)
      return () => clearTimeout(timer)
    }
  }, [duration])

  const handleClose = () => {
    setIsExiting(true)
    setTimeout(() => {
      setIsVisible(false)
      onClose()
    }, 300) // 等待动画完成
  }

  if (!isVisible) return null

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'success':
        return <CheckCircle className="w-5 h-5 text-green-500" />
      case 'warning':
        return <AlertTriangle className="w-5 h-5 text-yellow-500" />
      case 'error':
        return <XCircle className="w-5 h-5 text-red-500" />
      default:
        return <Info className="w-5 h-5 text-blue-500" />
    }
  }

  return (
    <Card className={`fixed top-20 right-4 z-50 w-80 p-4 shadow-lg border-l-4 ${
      notification.type === 'success' ? 'border-l-green-500 bg-green-50' :
      notification.type === 'warning' ? 'border-l-yellow-500 bg-yellow-50' :
      notification.type === 'error' ? 'border-l-red-500 bg-red-50' :
      'border-l-blue-500 bg-blue-50'
    } ${isExiting ? 'notification-slide-out' : 'notification-bounce-in'} ${
      notification.priority === 'high' ? 'shadow-xl ring-2 ring-red-200' : ''
    }`}>
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0">
          {getNotificationIcon(notification.type)}
        </div>
        
        <div className="flex-1 min-w-0">
          <h4 className="text-sm font-medium text-gray-900 mb-1">
            {notification.title}
          </h4>
          <p className="text-sm text-gray-600">
            {notification.content}
          </p>
          
          {/* 进度条显示自动关闭倒计时 */}
          {duration > 0 && !isExiting && (
            <div className="mt-2">
              <div className="h-1 bg-gray-200 rounded-full overflow-hidden">
                <div 
                  className={`h-full transition-all linear ${
                    notification.type === 'success' ? 'bg-green-500' :
                    notification.type === 'warning' ? 'bg-yellow-500' :
                    notification.type === 'error' ? 'bg-red-500' :
                    'bg-blue-500'
                  }`}
                  style={{
                    animation: `shrink ${duration}ms linear`,
                    width: '100%'
                  }}
                />
              </div>
            </div>
          )}
        </div>
        
        <Button
          variant="ghost"
          size="sm"
          onClick={handleClose}
          className="flex-shrink-0 h-6 w-6 p-0 hover:bg-gray-200 transition-colors"
        >
          <X className="w-4 h-4" />
        </Button>
      </div>
    </Card>
  )
}

// 系统通知管理器
export function NotificationManager() {
  const { notifications } = useRealtimeNotifications()
  const [floatingNotifications, setFloatingNotifications] = useState<any[]>([])

  // 显示新的浮动通知
  React.useEffect(() => {
    const latestNotification = notifications[0]
    if (latestNotification && !latestNotification.read) {
      // 只显示高优先级的通知作为浮动通知
      if (latestNotification.priority === 'high') {
        setFloatingNotifications(prev => [
          latestNotification,
          ...prev.slice(0, 2) // 最多同时显示3个浮动通知
        ])
      }
    }
  }, [notifications])

  const removeFloatingNotification = (notificationId: string) => {
    setFloatingNotifications(prev => 
      prev.filter(n => n.id !== notificationId)
    )
  }

  return (
    <div>
      {floatingNotifications.map((notification, index) => (
        <div
          key={notification.id}
          style={{ top: `${5 + index * 6}rem` }}
          className="fixed right-4 z-50"
        >
          <FloatingNotification
            notification={notification}
            onClose={() => removeFloatingNotification(notification.id)}
            duration={notification.priority === 'high' ? 8000 : 5000}
          />
        </div>
      ))}
    </div>
  )
}