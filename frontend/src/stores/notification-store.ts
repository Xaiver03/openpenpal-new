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

// 已移除所有 mock 数据生成器，现在完全使用真实 API

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
            const { getNotifications } = await import('@/lib/api/notification')
            const response = await getNotifications(query)
            
            set({
              notifications: response.notifications,
              loading: false,
              last_fetched: Date.now()
            })
            
            // Update stats
            get().refreshStats()
          } catch (error) {
            console.error('Failed to load notifications from API:', error)
            set({
              notifications: [], // 不使用 mock 数据，显示空列表
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
            const { markAsRead } = await import('@/lib/api/notification')
            await Promise.all(ids.map(id => markAsRead(id)))
            
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
            const { deleteNotification } = await import('@/lib/api/notification')
            await deleteNotification(id)
            
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
            const { updatePreferences } = await import('@/lib/api/notification')
            await updatePreferences(newPrefs)
            
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