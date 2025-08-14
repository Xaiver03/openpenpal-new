/**
 * Notification Store - SOTA State Management
 * 通知系统状态管理 - 支持实时更新、批量操作、偏好设置
 */

import { create } from 'zustand'
import { devtools, persist, subscribeWithSelector } from 'zustand/middleware'
import { toast } from 'sonner'
import type {
  Notification,
  NotificationState,
  NotificationActions,
  NotificationListQuery,
  NotificationPreferences,
  NotificationStats,
  NotificationType
} from '@/types/notification'

interface NotificationStoreState extends NotificationState, NotificationActions {
  _isInitialized: boolean
  _pollInterval: NodeJS.Timeout | null
}

// Mock data generator for development
const generateMockNotifications = (): Notification[] => {
  const types: NotificationType[] = ['follow', 'comment', 'like', 'letter_received', 'achievement']
  const now = Date.now()
  
  return Array.from({ length: 5 }, (_, i) => ({
    id: `notif-${i + 1}`,
    user_id: 'current-user',
    type: types[i % types.length],
    status: i < 2 ? 'unread' : 'read',
    title: getNotificationTitle(types[i % types.length]),
    content: getNotificationContent(types[i % types.length]),
    metadata: {
      actor_user: {
        id: `user-${i + 10}`,
        username: `user${i + 10}`,
        nickname: `User ${i + 10}`
      },
      url: getNotificationUrl(types[i % types.length])
    },
    created_at: new Date(now - i * 60 * 60 * 1000).toISOString(),
    read_at: i >= 2 ? new Date(now - (i - 1) * 60 * 60 * 1000).toISOString() : undefined
  }))
}

function getNotificationTitle(type: NotificationType): string {
  switch (type) {
    case 'follow': return '新的关注者'
    case 'comment': return '收到新评论'
    case 'comment_reply': return '评论被回复'
    case 'like': return '收到点赞'
    case 'letter_received': return '收到新信件'
    case 'achievement': return '获得新成就'
    case 'courier_task': return '信使任务更新'
    case 'system': return '系统通知'
    default: return '新通知'
  }
}

function getNotificationContent(type: NotificationType): string {
  switch (type) {
    case 'follow': return '@alice 关注了你'
    case 'comment': return '@bob 在你的信件下发表了评论'
    case 'comment_reply': return '@charlie 回复了你的评论'
    case 'like': return '@david 赞了你的信件'
    case 'letter_received': return '你收到了一封来自远方的信'
    case 'achievement': return '恭喜！你获得了"笔友达人"成就'
    case 'courier_task': return '你有新的信使任务待处理'
    case 'system': return '系统维护将于今晚10点开始'
    default: return '点击查看详情'
  }
}

function getNotificationUrl(type: NotificationType): string {
  switch (type) {
    case 'follow': return '/followers'
    case 'comment':
    case 'comment_reply':
    case 'like': return '/letters'
    case 'letter_received': return '/inbox'
    case 'achievement': return '/profile#achievements'
    case 'courier_task': return '/courier'
    case 'system': return '/announcements'
    default: return '/'
  }
}

const DEFAULT_PREFERENCES: NotificationPreferences = {
  email_enabled: true,
  push_enabled: true,
  types: {
    follow: true,
    comment: true,
    comment_reply: true,
    like: true,
    letter_received: true,
    achievement: true,
    courier_task: true,
    system: true
  }
}

export const useNotificationStore = create<NotificationStoreState>()(
  devtools(
    persist(
      subscribeWithSelector((set, get) => ({
        // Initial state
        notifications: [],
        stats: null,
        preferences: DEFAULT_PREFERENCES,
        loading: false,
        error: null,
        last_fetched: null,
        _isInitialized: false,
        _pollInterval: null,

        // Load notifications
        loadNotifications: async (query?: NotificationListQuery) => {
          set({ loading: true, error: null })
          
          try {
            // TODO: Replace with real API call
            // const response = await fetch(`/api/notifications?${new URLSearchParams(query)}`)
            // const data = await response.json()
            
            // Mock implementation
            await new Promise(resolve => setTimeout(resolve, 500))
            const mockNotifications = generateMockNotifications()
            
            set({
              notifications: mockNotifications,
              loading: false,
              last_fetched: Date.now()
            })
            
            // Update stats
            get().refreshStats()
          } catch (error) {
            set({
              loading: false,
              error: error instanceof Error ? error.message : 'Failed to load notifications'
            })
          }
        },

        // Mark notifications as read
        markAsRead: async (ids: string[]) => {
          // Optimistic update
          set(state => ({
            notifications: state.notifications.map(notif =>
              ids.includes(notif.id)
                ? { ...notif, status: 'read', read_at: new Date().toISOString() }
                : notif
            )
          }))
          
          try {
            // TODO: Replace with real API call
            // await fetch('/api/notifications/mark-read', {
            //   method: 'POST',
            //   headers: { 'Content-Type': 'application/json' },
            //   body: JSON.stringify({ notification_ids: ids, status: 'read' })
            // })
            
            // Update stats after marking as read
            get().refreshStats()
          } catch (error) {
            // Rollback on error
            await get().loadNotifications()
            throw error
          }
        },

        // Mark all as read
        markAllAsRead: async () => {
          const unreadIds = get().notifications
            .filter(n => n.status === 'unread')
            .map(n => n.id)
          
          if (unreadIds.length > 0) {
            await get().markAsRead(unreadIds)
          }
        },

        // Delete notification
        deleteNotification: async (id: string) => {
          // Optimistic update
          set(state => ({
            notifications: state.notifications.filter(n => n.id !== id)
          }))
          
          try {
            // TODO: Replace with real API call
            // await fetch(`/api/notifications/${id}`, { method: 'DELETE' })
            
            get().refreshStats()
          } catch (error) {
            // Rollback on error
            await get().loadNotifications()
            throw error
          }
        },

        // Update preferences
        updatePreferences: async (preferences: Partial<NotificationPreferences>) => {
          const currentPrefs = get().preferences!
          const newPrefs = { ...currentPrefs, ...preferences }
          
          // Optimistic update
          set({ preferences: newPrefs })
          
          try {
            // TODO: Replace with real API call
            // await fetch('/api/notifications/preferences', {
            //   method: 'PUT',
            //   headers: { 'Content-Type': 'application/json' },
            //   body: JSON.stringify(newPrefs)
            // })
            
            toast.success('通知设置已更新')
          } catch (error) {
            // Rollback on error
            set({ preferences: currentPrefs })
            toast.error('更新通知设置失败')
            throw error
          }
        },

        // Refresh stats
        refreshStats: async () => {
          const notifications = get().notifications
          const unreadCount = notifications.filter(n => n.status === 'unread').length
          
          const byType = notifications.reduce((acc, notif) => {
            acc[notif.type] = (acc[notif.type] || 0) + 1
            return acc
          }, {} as Record<NotificationType, number>)
          
          const stats: NotificationStats = {
            unread_count: unreadCount,
            total_count: notifications.length,
            by_type: byType
          }
          
          set({ stats })
        },

        // Clear all notifications
        clearNotifications: () => {
          set({
            notifications: [],
            stats: null,
            last_fetched: null
          })
        }
      })),
      {
        name: 'openpenpal-notification-store',
        partialize: (state) => ({
          preferences: state.preferences,
          _isInitialized: state._isInitialized
        })
      }
    ),
    {
      name: 'notification-store'
    }
  )
)

// Convenience hooks
export const useNotifications = () => {
  const store = useNotificationStore()
  
  return {
    notifications: store.notifications,
    unreadCount: store.stats?.unread_count || 0,
    loading: store.loading,
    error: store.error,
    
    loadNotifications: store.loadNotifications,
    markAsRead: store.markAsRead,
    markAllAsRead: store.markAllAsRead,
    deleteNotification: store.deleteNotification
  }
}

export const useNotificationPreferences = () => {
  const preferences = useNotificationStore(state => state.preferences)
  const updatePreferences = useNotificationStore(state => state.updatePreferences)
  
  return {
    preferences,
    updatePreferences
  }
}

export const useUnreadNotificationCount = () => {
  return useNotificationStore(state => state.stats?.unread_count || 0)
}

// Initialize store and start polling
export const initializeNotificationStore = () => {
  const store = useNotificationStore.getState()
  
  if (!store._isInitialized) {
    store.loadNotifications()
    
    // Start polling for new notifications every 30 seconds
    const interval = setInterval(() => {
      store.loadNotifications()
    }, 30000)
    
    useNotificationStore.setState({
      _isInitialized: true,
      _pollInterval: interval
    })
  }
}

// Cleanup function
export const cleanupNotificationStore = () => {
  const store = useNotificationStore.getState()
  
  if (store._pollInterval) {
    clearInterval(store._pollInterval)
    useNotificationStore.setState({
      _pollInterval: null,
      _isInitialized: false
    })
  }
}

export default useNotificationStore