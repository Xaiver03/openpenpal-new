/**
 * 信件状态枚举
 */
export type LetterStatus = 
  | 'draft'        // 草稿
  | 'generated'    // 已生成编号
  | 'collected'    // 已收取
  | 'in_transit'   // 在途
  | 'delivered'    // 已送达
  | 'read'         // 已查看

/**
 * 信件样式枚举
 */
export type LetterStyle = 
  | 'classic'      // 经典
  | 'modern'       // 现代
  | 'vintage'      // 复古
  | 'elegant'      // 优雅
  | 'casual'       // 休闲

/**
 * 信件草稿
 */
export interface LetterDraft {
  id: string
  title?: string
  content: string
  style: LetterStyle
  reply_to?: string
  created_at: Date
  updated_at: Date
}

/**
 * 信件编号
 */
export interface LetterCode {
  id: string
  letter_id: string
  code: string
  qr_code_url?: string
  generated_at: Date
  expires_at?: Date
}

/**
 * 信件
 */
export interface Letter {
  id: string
  user_id: string
  title?: string
  content: string
  style: LetterStyle
  status: LetterStatus
  reply_to?: string
  code?: LetterCode
  created_at: Date
  updated_at: Date
}

/**
 * 信件照片
 */
export interface LetterPhoto {
  id: string
  letter_id: string
  image_url: string
  is_public: boolean
  created_at: Date
}

/**
 * 状态更新日志
 */
export interface StatusLog {
  id: string
  code_id: string
  status: LetterStatus
  updated_by: string
  location?: string
  note?: string
  created_at: Date
}

/**
 * 发送的信件
 */
export interface SentLetter extends Letter {
  status_logs: StatusLog[]
  photos: LetterPhoto[]
}

/**
 * 收到的信件
 */
export interface ReceivedLetter extends Letter {
  sender_nickname?: string
  read_at?: Date
}

/**
 * 创建信件请求
 */
export interface CreateLetterRequest {
  title?: string
  content: string
  style: LetterStyle
  reply_to?: string
}

/**
 * 更新信件状态请求
 */
export interface UpdateLetterStatusRequest {
  status: LetterStatus
  location?: string
  note?: string
}

/**
 * 信件列表查询参数
 */
export interface LetterListParams {
  page?: number
  limit?: number
  status?: LetterStatus
  style?: LetterStyle
  search?: string
  sort_by?: 'created_at' | 'updated_at'
  sort_order?: 'asc' | 'desc'
}

/**
 * 信件统计
 */
export interface LetterStats {
  total_sent: number
  total_received: number
  in_transit: number
  delivered: number
  drafts: number
}

/**
 * 信件模板
 */
export interface LetterTemplate {
  id: string
  name: string
  description?: string
  category: string
  preview_image?: string
  content_template: string
  style_config?: any
  is_premium: boolean
  is_active: boolean
  usage_count: number
  rating: number
  created_by?: string
  created_at: string
  updated_at: string
}