/**
 * NotificationBell - Notification Bell with Dropdown
 * 通知铃铛组件 - 带下拉菜单的通知提醒
 */

'use client'

import React, { useState, useRef, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Bell, Check, Settings, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import { useNotifications, initializeNotificationStore, cleanupNotificationStore } from '@/stores/notification-store'
import NotificationItem from './notification-item'
import type { NotificationBellProps } from '@/types/notification'

export function NotificationBell({ 
  className,
  show_count = true,
  max_count = 99
}: NotificationBellProps) {
  const router = useRouter()
  const [isOpen, setIsOpen] = useState(false)
  const { 
    notifications, 
    unreadCount, 
    loading,
    loadNotifications,
    markAsRead,
    markAllAsRead,
    deleteNotification
  } = useNotifications()

  // Initialize notification polling
  useEffect(() => {
    initializeNotificationStore()
    
    return () => {
      cleanupNotificationStore()
    }
  }, [])

  // Load notifications when dropdown opens
  useEffect(() => {
    if (isOpen) {
      loadNotifications()
    }
  }, [isOpen, loadNotifications])

  const handleNotificationClick = async (notificationId: string, url?: string) => {
    // Mark as read
    await markAsRead([notificationId])
    
    // Navigate if URL provided
    if (url) {
      router.push(url)
    }
    
    // Close dropdown
    setIsOpen(false)
  }

  const displayCount = unreadCount > max_count ? `${max_count}+` : unreadCount

  return (
    <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className={cn('relative', className)}
        >
          <Bell className="h-5 w-5" />
          {show_count && unreadCount > 0 && (
            <Badge 
              variant="destructive" 
              className="absolute -top-1 -right-1 h-5 min-w-[20px] px-1 text-xs"
            >
              {displayCount}
            </Badge>
          )}
        </Button>
      </DropdownMenuTrigger>
      
      <DropdownMenuContent 
        align="end" 
        className="w-[380px] p-0"
        sideOffset={5}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <h3 className="font-semibold">通知</h3>
          <div className="flex items-center gap-1">
            {unreadCount > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => markAllAsRead()}
                className="h-8 text-xs"
              >
                <Check className="h-3.5 w-3.5 mr-1" />
                全部已读
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.push('/settings/notifications')}
              className="h-8 w-8 p-0"
            >
              <Settings className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Notification List */}
        <ScrollArea className="h-[400px]">
          {loading && notifications.length === 0 ? (
            <div className="p-8 text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4" />
              <p className="text-sm text-muted-foreground">加载中...</p>
            </div>
          ) : notifications.length === 0 ? (
            <div className="p-8 text-center">
              <Bell className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-sm font-medium mb-1">暂无通知</p>
              <p className="text-xs text-muted-foreground">
                新的通知会在这里显示
              </p>
            </div>
          ) : (
            <div className="divide-y">
              {notifications.map((notification) => (
                <NotificationItem
                  key={notification.id}
                  notification={notification}
                  on_click={() => handleNotificationClick(notification.id, notification.metadata?.url)}
                  on_delete={() => deleteNotification(notification.id)}
                  show_actions
                />
              ))}
            </div>
          )}
        </ScrollArea>

        {/* Footer */}
        {notifications.length > 0 && (
          <>
            <DropdownMenuSeparator className="m-0" />
            <div className="p-2">
              <Button
                variant="ghost"
                className="w-full justify-center text-sm"
                onClick={() => {
                  setIsOpen(false)
                  router.push('/notifications')
                }}
              >
                查看全部通知
              </Button>
            </div>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export default NotificationBell