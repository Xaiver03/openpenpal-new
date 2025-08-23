/**
 * Notification Channel Preferences
 * 通知渠道偏好设置类型定义
 */

import type { NotificationType } from './notification'

/**
 * 通知渠道偏好 - 每种通知类型的渠道设置
 */
export interface NotificationChannelPreferences {
  email: Record<NotificationType, boolean>
  push: Record<NotificationType, boolean>
}

/**
 * API 响应格式 - 后端返回的格式
 */
export interface NotificationPreferencesResponse {
  emailEnabled: boolean
  smsEnabled: boolean
  pushEnabled: boolean
  frequency: string
  language: string
  timezone: string
}