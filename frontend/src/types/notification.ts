/**
 * Notification type definitions
 * 通知系统类型定义
 */

export type NotificationType = 
  | 'follow'           // 新关注
  | 'comment'          // 新评论
  | 'comment_reply'    // 评论回复
  | 'like'             // 点赞
  | 'letter_received'  // 收到信件
  | 'achievement'      // 获得成就
  | 'courier_task'     // 信使任务
  | 'system'           // 系统通知

export type NotificationStatus = 'unread' | 'read' | 'archived'

export interface NotificationUser {
  id: string
  username: string
  nickname?: string
  avatar_url?: string
}

export interface Notification {
  id: string
  user_id: string          // 接收通知的用户
  type: NotificationType
  status: NotificationStatus
  title: string
  content: string
  metadata?: {
    actor_user?: NotificationUser    // 触发通知的用户
    letter_id?: string
    comment_id?: string
    achievement_id?: string
    task_id?: string
    url?: string                     // 点击后跳转的链接
  }
  created_at: string
  read_at?: string
}

export interface NotificationStats {
  unread_count: number
  total_count: number
  by_type: Record<NotificationType, number>
}

export interface NotificationPreferences {
  email_enabled: boolean
  push_enabled: boolean
  types: {
    follow: boolean
    comment: boolean
    comment_reply: boolean
    like: boolean
    letter_received: boolean
    achievement: boolean
    courier_task: boolean
    system: boolean
  }
}

// API Request/Response types
export interface NotificationListQuery {
  page?: number
  limit?: number
  type?: NotificationType
  status?: NotificationStatus
  start_date?: string
  end_date?: string
}

export interface NotificationListResponse {
  notifications: Notification[]
  pagination: {
    page: number
    limit: number
    total: number
    pages: number
  }
}

export interface MarkNotificationRequest {
  notification_ids: string[]
  status: NotificationStatus
}

export interface NotificationBatchResponse {
  success: string[]
  failed: string[]
}

// Component Props
export interface NotificationBellProps {
  className?: string
  show_count?: boolean
  max_count?: number
}

export interface NotificationListProps {
  max_height?: string
  show_header?: boolean
  show_mark_all?: boolean
  on_notification_click?: (notification: Notification) => void
  className?: string
}

export interface NotificationItemProps {
  notification: Notification
  on_click?: (notification: Notification) => void
  on_mark_read?: (id: string) => void
  on_delete?: (id: string) => void
  show_actions?: boolean
  className?: string
}

// Store types
export interface NotificationState {
  notifications: Notification[]
  stats: NotificationStats | null
  preferences: NotificationPreferences | null
  loading: boolean
  error: string | null
  last_fetched: number | null
}

export interface NotificationActions {
  loadNotifications: (query?: NotificationListQuery) => Promise<void>
  markAsRead: (ids: string[]) => Promise<void>
  markAllAsRead: () => Promise<void>
  deleteNotification: (id: string) => Promise<void>
  updatePreferences: (preferences: Partial<NotificationPreferences>) => Promise<void>
  refreshStats: () => Promise<void>
  clearNotifications: () => void
}