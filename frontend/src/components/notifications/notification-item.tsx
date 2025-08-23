/**
 * NotificationItem - Individual Notification Display Component
 * 通知项组件 - 单个通知的显示组件
 */

'use client'

import React from 'react'
import { 
  Bell, 
  MessageSquare, 
  Heart, 
  Users, 
  Mail, 
  Award,
  Package,
  AlertCircle,
  Trash2
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { SafeTimestamp } from '@/components/ui/safe-timestamp'
import { cn } from '@/lib/utils'
import type { NotificationItemProps, NotificationType } from '@/types/notification'

export function NotificationItem({
  notification,
  on_click,
  on_delete,
  show_actions = false,
  className
}: NotificationItemProps) {
  const getIcon = (type: NotificationType) => {
    switch (type) {
      case 'follow':
        return Users
      case 'comment':
      case 'comment_reply':
        return MessageSquare
      case 'like':
        return Heart
      case 'letter_received':
        return Mail
      case 'achievement':
        return Award
      case 'courier_task':
        return Package
      case 'system':
        return AlertCircle
      default:
        return Bell
    }
  }

  const getIconColor = (type: NotificationType) => {
    switch (type) {
      case 'follow':
        return 'text-blue-600 bg-blue-100'
      case 'comment':
      case 'comment_reply':
        return 'text-green-600 bg-green-100'
      case 'like':
        return 'text-red-600 bg-red-100'
      case 'letter_received':
        return 'text-purple-600 bg-purple-100'
      case 'achievement':
        return 'text-yellow-600 bg-yellow-100'
      case 'courier_task':
        return 'text-orange-600 bg-orange-100'
      case 'system':
        return 'text-gray-600 bg-gray-100'
      default:
        return 'text-primary bg-primary/10'
    }
  }

  const Icon = getIcon(notification.type)
  const iconColor = getIconColor(notification.type)
  const isUnread = notification.status === 'unread'

  return (
    <div
      className={cn(
        'relative flex gap-3 p-4 transition-colors cursor-pointer',
        isUnread && 'bg-muted/50',
        'hover:bg-muted/30',
        className
      )}
      onClick={() => on_click?.(notification)}
    >
      {/* Unread indicator */}
      {isUnread && (
        <div className="absolute left-1 top-1/2 -translate-y-1/2 w-2 h-2 bg-primary rounded-full" />
      )}

      {/* Icon or Avatar */}
      <div className="flex-shrink-0">
        {notification.metadata?.actor_user ? (
          <div className="relative">
            <Avatar className="h-10 w-10">
              <AvatarImage src={notification.metadata.actor_user.avatar_url} />
              <AvatarFallback>
                {notification.metadata.actor_user.nickname?.charAt(0) || 
                 notification.metadata.actor_user.username.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div className={cn(
              'absolute -bottom-1 -right-1 h-5 w-5 rounded-full flex items-center justify-center',
              iconColor
            )}>
              <Icon className="h-3 w-3" />
            </div>
          </div>
        ) : (
          <div className={cn(
            'h-10 w-10 rounded-full flex items-center justify-center',
            iconColor
          )}>
            <Icon className="h-5 w-5" />
          </div>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <p className={cn(
          'text-sm',
          isUnread ? 'font-medium' : 'text-muted-foreground'
        )}>
          {notification.content}
        </p>
        <SafeTimestamp 
          date={notification.created_at} 
          format="relative" 
          fallback="刚刚"
          className="text-xs text-muted-foreground mt-1 block"
        />
      </div>

      {/* Actions */}
      {show_actions && on_delete && (
        <div className="flex-shrink-0">
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={(e) => {
              e.stopPropagation()
              on_delete(notification.id)
            }}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  )
}

export default NotificationItem