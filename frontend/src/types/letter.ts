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
 * 信件可见性枚举
 */
export type LetterVisibility =
  | 'private'      // 私有
  | 'public'       // 公开
  | 'friends'      // 好友可见

/**
 * FSD 条码状态枚举
 */
export type BarcodeStatus =
  | 'unactivated'  // 未激活
  | 'bound'        // 已绑定
  | 'in_transit'   // 投递中
  | 'delivered'    // 已送达
  | 'expired'      // 已过期
  | 'cancelled'    // 已取消

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

// 删除重复的LetterStats定义，使用下方的增强版

/**
 * 信件编号 - 增强支持FSD条码系统规格
 */
export interface LetterCode {
  id: string
  letter_id: string
  code: string
  qr_code_url?: string
  qr_code_path?: string
  expires_at?: Date
  created_at: Date
  updated_at: Date
  
  // FSD条码系统增强字段
  status: BarcodeStatus           // 条码状态
  recipient_code?: string         // 收件人OP Code
  envelope_id?: string           // 关联信封ID
  bound_at?: Date               // 绑定时间
  delivered_at?: Date           // 送达时间
  last_scanned_by?: string      // 最后扫码人
  last_scanned_at?: Date        // 最后扫码时间
  scan_count: number            // 扫码次数
  
  // 关联对象
  letter?: Letter
  envelope?: any  // Envelope interface to be defined
}

/**
 * 信件 - 完整模型定义
 */
export interface Letter {
  id: string
  user_id: string
  author_id?: string              // 作者ID（可能与user_id不同）
  title?: string
  content: string
  style: LetterStyle
  status: LetterStatus
  visibility: LetterVisibility    // 可见性控制
  like_count: number             // 点赞数
  share_count: number            // 分享数
  view_count: number             // 浏览数
  
  // OP Code System - 核心地址标识
  recipient_op_code?: string     // 收件人OP Code，如: PK5F3D
  sender_op_code?: string        // 发件人OP Code（可选）
  
  reply_to?: string             // 回复的信件ID
  envelope_id?: string          // 关联信封ID
  
  created_at: Date
  updated_at: Date
  
  // 关联对象
  user?: any                    // User object
  author?: any                  // Author User object
  code?: LetterCode            // 信件编号
  status_logs?: StatusLog[]     // 状态日志
  photos?: LetterPhoto[]       // 信件照片
  envelope?: any               // Envelope object
  likes?: any[]                // 点赞记录
  shares?: any[]               // 分享记录
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
  visibility?: LetterVisibility   // 可见性设置
  recipient_op_code?: string     // 收件人OP Code
  sender_op_code?: string        // 发件人OP Code
  reply_to?: string
}

/**
 * FSD 条码绑定请求
 */
export interface BindBarcodeRequest {
  barcode: string
  envelope_id: string
  recipient_op_code?: string
}

/**
 * 更新条码状态请求
 */
export interface UpdateBarcodeStatusRequest {
  status: BarcodeStatus
  location?: string
  note?: string
  scanned_by?: string
}

/**
 * 条码验证请求
 */
export interface ValidateBarcodeRequest {
  barcode: string
  operation: 'bind' | 'scan' | 'deliver'
  current_op_code?: string
}

/**
 * 信封与条码响应
 */
export interface EnvelopeWithBarcodeResponse {
  envelope_id: string
  barcode: string
  status: BarcodeStatus
  recipient_op_code?: string
  bound_at?: Date
  estimated_delivery?: Date
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
 * 信件列表查询参数 - 增强版
 */
export interface LetterListParams {
  page?: number
  limit?: number
  status?: LetterStatus
  style?: LetterStyle
  visibility?: LetterVisibility
  search?: string
  recipient_op_code?: string      // 按收件人OP Code筛选
  sender_op_code?: string         // 按发件人OP Code筛选
  date_from?: string
  date_to?: string
  sort_by?: 'created_at' | 'updated_at' | 'like_count' | 'view_count'
  sort_order?: 'asc' | 'desc'
}

/**
 * 信件搜索参数
 */
export interface LetterSearchParams {
  query: string
  tags: string[]
  date_from?: string
  date_to?: string
  visibility?: LetterVisibility
  sort_by: string
  sort_order: 'asc' | 'desc'
  page: number
  limit: number
}

/**
 * 信件统计 - 增强版
 */
export interface LetterStats {
  total_sent: number
  total_received: number
  in_transit: number
  delivered: number
  drafts: number
  
  // 按可见性统计
  public_letters: number
  private_letters: number
  friends_letters: number
  
  // 按OP Code统计
  op_code_letters: number
  
  // FSD条码系统统计
  unactivated_barcodes: number
  bound_barcodes: number
  in_transit_barcodes: number
  delivered_barcodes: number
  expired_barcodes: number
  cancelled_barcodes: number
  
  // 交互统计
  total_likes: number
  total_shares: number
  total_views: number
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