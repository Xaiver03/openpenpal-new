// Import CourierLevel from user types
import { CourierLevel } from './user'

/**
 * 任务状态枚举
 */
export type TaskStatus = 
  | 'pending'      // 待处理
  | 'in_progress'  // 进行中
  | 'completed'    // 已完成
  | 'failed'       // 失败

/**
 * 任务类型枚举
 */
export type TaskType = 
  | 'collect'      // 收取
  | 'transfer'     // 转交
  | 'deliver'      // 投递

/**
 * 信使任务
 */
export interface CourierTask {
  id: string
  courier_id: string
  code_id: string
  task_type: TaskType
  status: TaskStatus
  zone?: string
  priority: number
  assigned_at: Date
  completed_at?: Date
  note?: string
  description?: string
}

/**
 * 信使信息
 */
export interface Courier {
  id: string
  user_id: string
  level: CourierLevel
  zone: string
  school_code: string
  is_active: boolean
  score: number
  completed_tasks: number
  created_at: Date
  updated_at: Date
}

/**
 * 信使统计
 */
export interface CourierStats {
  total_tasks: number
  completed_tasks: number
  pending_tasks: number
  success_rate: number
  average_delivery_time: number // 小时
  ranking: number
  points: number
}

/**
 * 扫码记录
 */
export interface ScanRecord {
  id: string
  courier_id: string
  code: string
  action: 'collect' | 'deliver'
  location?: string
  timestamp: Date
  note?: string
}

/**
 * 信使申请
 */
export interface CourierApplication {
  id: string
  user_id: string
  level: CourierLevel
  zone: string
  reason: string
  status: 'pending' | 'approved' | 'rejected'
  applied_at: Date
  reviewed_at?: Date
  reviewed_by?: string
  review_note?: string
}

/**
 * 创建信使任务请求
 */
export interface CreateTaskRequest {
  code_id: string
  task_type: TaskType
  zone?: string
  priority?: number
  note?: string
}

/**
 * 更新任务状态请求
 */
export interface UpdateTaskStatusRequest {
  status: TaskStatus
  note?: string
  location?: string
}

/**
 * 信使绩效
 */
export interface CourierPerformance {
  courier_id: string
  courier_name: string
  level: CourierLevel
  zone: string
  period: {
    start: Date
    end: Date
  }
  metrics: {
    total_tasks: number
    completed_tasks: number
    failed_tasks: number
    average_completion_time: number
    success_rate: number
    user_rating: number
  }
  ranking: number
  rewards: {
    points: number
    badges: string[]
    bonuses: number
  }
}