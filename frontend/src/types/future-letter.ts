/**
 * 未来信状态枚举
 */
export type FutureLetterStatus = 
  | 'scheduled'   // 已安排
  | 'sent'        // 已发送
  | 'cancelled'   // 已取消

/**
 * 投递方式枚举
 */
export type DeliveryMethod = 
  | 'system'      // 系统投递
  | 'courier'     // 信使投递

/**
 * 未来信
 */
export interface FutureLetter {
  id: string
  letter_id: string
  sender_id: string
  recipient_id?: string
  recipient_op_code?: string
  status: FutureLetterStatus
  scheduled_date: Date
  delivery_method: DeliveryMethod
  reminder_enabled: boolean
  reminder_days: number
  last_reminder_sent?: Date
  sent_at?: Date
  created_at: Date
  updated_at: Date

  // 关联对象
  letter?: {
    id: string
    title?: string
    content: string
    style: string
    created_at: Date
  }
  sender?: {
    id: string
    nickname?: string
    avatar_url?: string
  }
  recipient?: {
    id: string
    nickname?: string
    avatar_url?: string
    op_code?: string
  }
}

/**
 * 创建未来信请求
 */
export interface CreateFutureLetterRequest {
  letter_id: string
  scheduled_date: Date
  recipient_id?: string
  recipient_op_code?: string
  delivery_method?: DeliveryMethod
  reminder_enabled?: boolean
  reminder_days?: number
}

/**
 * 未来信列表参数
 */
export interface FutureLetterListParams {
  page?: number
  limit?: number
  status?: FutureLetterStatus
  delivery_method?: DeliveryMethod
  date_from?: string
  date_to?: string
  sort_by?: 'created_at' | 'scheduled_date' | 'sent_at'
  sort_order?: 'asc' | 'desc'
}

/**
 * 未来信统计信息
 */
export interface FutureLetterStats {
  pending_count: number        // 待发送数量
  upcoming_24h_count: number   // 24小时内即将发送数量
  total_system_pending: number // 系统总待发送数量
  sent_count?: number          // 已发送数量
  cancelled_count?: number     // 已取消数量
  total_count: number          // 总计数量
}

/**
 * 更新未来信请求
 */
export interface UpdateFutureLetterRequest {
  scheduled_date?: Date
  delivery_method?: DeliveryMethod
  reminder_enabled?: boolean
  reminder_days?: number
}

/**
 * 未来信响应
 */
export interface FutureLetterResponse {
  id: string
  letter: {
    id: string
    title?: string
    content: string
    style: string
    created_at: Date
  }
  status: FutureLetterStatus
  scheduled_date: Date
  delivery_method: DeliveryMethod
  reminder_enabled: boolean
  reminder_days: number
  last_reminder_sent?: Date
  sent_at?: Date
  recipient?: {
    id: string
    nickname?: string
    avatar_url?: string
    op_code?: string
  }
  created_at: Date
}

/**
 * 未来信提醒设置
 */
export interface ReminderSettings {
  enabled: boolean
  days_before: number
  notification_types: ('email' | 'push' | 'sms')[]
}

/**
 * 未来信日历事件
 */
export interface FutureLetterCalendarEvent {
  id: string
  title: string
  scheduled_date: Date
  status: FutureLetterStatus
  recipient_name?: string
  preview_content: string
}