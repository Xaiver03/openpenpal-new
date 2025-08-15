/**
 * 积分系统类型定义 - 与后端模型严格对应
 */

/**
 * 积分任务类型
 */
export type CreditTaskType = 
  // 信件相关任务
  | 'letter_created'     // 创建信件
  | 'letter_generated'   // 生成编号
  | 'letter_delivered'   // 信件送达
  | 'letter_read'        // 信件被阅读
  | 'receive_letter'     // 收到信件
  | 'public_like'        // 公开信被点赞
  
  // 写作与挑战任务
  | 'writing_challenge'  // 写作挑战
  | 'ai_interaction'     // AI互动
  
  // 信使相关任务
  | 'courier_first'      // 信使首次任务
  | 'courier_delivery'   // 信使送达
  
  // 博物馆相关任务
  | 'museum_submit'      // 博物馆提交
  | 'museum_approved'    // 博物馆审核通过
  | 'museum_liked'       // 博物馆点赞
  
  // 系统管理任务
  | 'opcode_approval'    // OP Code审核
  | 'community_badge'    // 社区徽章
  | 'admin_reward'       // 管理员奖励

/**
 * 积分任务状态
 */
export type CreditTaskStatus = 
  | 'pending'    // 等待执行
  | 'scheduled'  // 已计划
  | 'executing'  // 执行中
  | 'completed'  // 已完成
  | 'failed'     // 执行失败
  | 'cancelled'  // 已取消
  | 'skipped'    // 已跳过(达到限制)

/**
 * 用户积分信息
 */
export interface UserCredit {
  id: string
  user_id: string
  total: number        // 总积分
  available: number    // 可用积分
  used: number         // 已使用积分
  earned: number       // 获得积分
  level: number        // 用户等级
  created_at: string
  updated_at: string
}

/**
 * 积分交易记录
 */
export interface CreditTransaction {
  id: string
  user_id: string
  type: 'earn' | 'spend'  // 获得或消费
  amount: number          // 积分数量（正数为获得，负数为消费）
  description: string     // 描述
  reference: string       // 关联对象ID
  created_at: string
}

/**
 * 积分任务
 */
export interface CreditTask {
  id: string
  task_type: CreditTaskType
  user_id: string
  status: CreditTaskStatus
  points: number
  description: string
  reference: string
  metadata?: string
  
  // 执行控制
  priority: number
  max_attempts: number
  attempts: number
  scheduled_at?: string
  executed_at?: string
  completed_at?: string
  failed_at?: string
  error_message?: string
  
  // 限制条件
  daily_limit: number
  weekly_limit: number
  constraints?: string
  
  created_at: string
  updated_at: string
}

/**
 * 积分任务统计
 */
export interface CreditTaskStatistics {
  task_type: CreditTaskType
  total_tasks: number
  completed_tasks: number
  failed_tasks: number
  total_points: number
  success_rate: number
  avg_points: number
}

/**
 * 积分任务规则
 */
export interface CreditTaskRule {
  id: string
  task_type: CreditTaskType
  points: number
  daily_limit: number
  weekly_limit: number
  is_active: boolean
  auto_execute: boolean
  description: string
  constraints?: string
  created_at: string
  updated_at: string
}

/**
 * 积分概览信息
 */
export interface CreditSummary {
  credit: UserCredit
  today_earned: number
  today_transactions: number
  week_earned: number
  month_earned: number
  pending_tasks: number
  completed_tasks_today: number
  current_level_progress: {
    current_points: number
    next_level_required: number
    progress_percentage: number
  }
}

/**
 * 积分历史查询参数
 */
export interface CreditHistoryParams {
  page?: number
  limit?: number
  type?: 'earn' | 'spend' | 'all'
  date_from?: string
  date_to?: string
}

/**
 * 积分历史响应
 */
export interface CreditHistoryResponse {
  transactions: CreditTransaction[]
  total: number
  page: number
  limit: number
}

/**
 * 积分任务列表查询参数
 */
export interface CreditTaskListParams {
  page?: number
  limit?: number
  status?: CreditTaskStatus
  task_type?: CreditTaskType
  date_from?: string
  date_to?: string
}

/**
 * 积分任务列表响应
 */
export interface CreditTaskListResponse {
  tasks: CreditTask[]
  total: number
  page: number
  limit: number
}

/**
 * 积分常量 - FSD规格
 */
export const CREDIT_POINTS = {
  // 信件相关积分
  LETTER_CREATED: 10,
  LETTER_GENERATED: 10,
  LETTER_DELIVERED: 20,
  LETTER_READ: 15,
  RECEIVE_LETTER: 5,
  PUBLIC_LETTER_LIKE: 1,
  
  // 写作与挑战相关积分
  WRITING_CHALLENGE: 15,
  AI_INTERACTION: 3,
  
  // 信使相关积分
  COURIER_FIRST_TASK: 20,
  COURIER_DELIVERY: 5,
  
  // 博物馆相关积分
  MUSEUM_SUBMIT: 25,
  MUSEUM_APPROVED: 100,
  MUSEUM_LIKED: 5,
  
  // 系统管理相关积分
  OPCODE_APPROVAL: 10,
  COMMUNITY_BADGE: 50,
} as const

/**
 * 积分等级升级所需积分
 */
export const LEVEL_UP_POINTS = [0, 100, 300, 600, 1000, 1500]

/**
 * 每日积分限制
 */
export const DAILY_LIMITS = {
  letter_created: 3,
  receive_letter: 5,
  public_like: 20,
  writing_challenge: 1,
  ai_interaction: 3,
} as const

/**
 * 积分任务描述映射
 */
export const TASK_DESCRIPTIONS: Record<CreditTaskType, string> = {
  letter_created: '创建信件',
  letter_generated: '生成信件编号',
  letter_delivered: '信件送达',
  letter_read: '信件被阅读',
  receive_letter: '收到信件',
  public_like: '公开信被点赞',
  writing_challenge: '参与写作挑战',
  ai_interaction: 'AI互动',
  courier_first: '信使首次任务完成',
  courier_delivery: '信使成功送达',
  museum_submit: '提交作品到博物馆',
  museum_approved: '博物馆作品审核通过',
  museum_liked: '博物馆作品获得点赞',
  opcode_approval: '点位申请审核成功',
  community_badge: '获得社区贡献徽章',
  admin_reward: '管理员奖励',
}

/**
 * 积分任务状态描述映射
 */
export const TASK_STATUS_DESCRIPTIONS: Record<CreditTaskStatus, string> = {
  pending: '等待执行',
  scheduled: '已计划',
  executing: '执行中',
  completed: '已完成',
  failed: '执行失败',
  cancelled: '已取消',
  skipped: '已跳过',
}