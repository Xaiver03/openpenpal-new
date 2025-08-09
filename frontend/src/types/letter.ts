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
  replyTo?: string
  createdAt: Date
  updatedAt: Date
}

/**
 * 信件编号
 */
export interface LetterCode {
  id: string
  letterId: string
  code: string
  qrCodeUrl?: string
  generatedAt: Date
  expiresAt?: Date
}

/**
 * 信件
 */
export interface Letter {
  id: string
  userId: string
  title?: string
  content: string
  style: LetterStyle
  status: LetterStatus
  replyTo?: string
  code?: LetterCode
  createdAt: Date
  updatedAt: Date
}

/**
 * 信件照片
 */
export interface LetterPhoto {
  id: string
  letterId: string
  imageUrl: string
  isPublic: boolean
  createdAt: Date
}

/**
 * 状态更新日志
 */
export interface StatusLog {
  id: string
  codeId: string
  status: LetterStatus
  updatedBy: string
  location?: string
  note?: string
  createdAt: Date
}

/**
 * 发送的信件
 */
export interface SentLetter extends Letter {
  statusLogs: StatusLog[]
  photos: LetterPhoto[]
}

/**
 * 收到的信件
 */
export interface ReceivedLetter extends Letter {
  senderNickname?: string
  readAt?: Date
}

/**
 * 创建信件请求
 */
export interface CreateLetterRequest {
  title?: string
  content: string
  style: LetterStyle
  replyTo?: string
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
  sortBy?: 'createdAt' | 'updatedAt'
  sortOrder?: 'asc' | 'desc'
}

/**
 * 信件统计
 */
export interface LetterStats {
  totalSent: number
  totalReceived: number
  inTransit: number
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