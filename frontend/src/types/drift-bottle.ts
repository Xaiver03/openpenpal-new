/**
 * 漂流瓶状态枚举
 */
export type DriftBottleStatus = 
  | 'floating'    // 漂流中
  | 'collected'   // 已被捞取
  | 'expired'     // 已过期

/**
 * 漂流瓶主题枚举
 */
export type DriftBottleTheme = 
  | 'friendship'  // 友谊
  | 'love'        // 爱情
  | 'confession'  // 表白
  | 'wish'        // 许愿
  | 'gratitude'   // 感谢
  | 'memory'      // 回忆
  | 'anonymous'   // 匿名
  | 'random'      // 随机

/**
 * 漂流瓶
 */
export interface DriftBottle {
  id: string
  letter_id: string
  sender_id: string
  collector_id?: string
  status: DriftBottleStatus
  theme?: string
  region?: string
  collected_at?: Date
  expires_at: Date
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
  collector?: {
    id: string
    nickname?: string
    avatar_url?: string
  }
}

/**
 * 创建漂流瓶请求
 */
export interface CreateDriftBottleRequest {
  letter_id: string
  theme?: string
  region?: string
  days?: number  // 漂流天数，默认7天
}

/**
 * 漂流瓶列表参数
 */
export interface DriftBottleListParams {
  page?: number
  limit?: number
  theme?: DriftBottleTheme
  region?: string
  status?: DriftBottleStatus
  sort_by?: 'created_at' | 'expires_at' | 'collected_at'
  sort_order?: 'asc' | 'desc'
}

/**
 * 漂流瓶统计信息
 */
export interface DriftBottleStats {
  sent_count: number       // 发送的漂流瓶数
  collected_count: number  // 捞取的漂流瓶数
  floating_count: number   // 正在漂流的瓶子数
  expired_count?: number   // 过期的瓶子数
}

/**
 * 漂流瓶响应
 */
export interface DriftBottleResponse {
  id: string
  letter: {
    id: string
    title?: string
    content: string
    style: string
    created_at: Date
  }
  status: DriftBottleStatus
  theme?: string
  region?: string
  expires_at: Date
  collected_at?: Date
  collector?: {
    id: string
    nickname?: string
    avatar_url?: string
  }
  created_at: Date
}

/**
 * 捞取漂流瓶响应
 */
export interface CatchDriftBottleResponse extends DriftBottleResponse {
  is_new_catch: boolean
  catch_timestamp: Date
}