/**
 * 信使等级枚举
 */
export type CourierLevel = 
  | 'level1'       // 1级信使 (楼栋)
  | 'level2'       // 2级信使 (片区)
  | 'level3'       // 3级信使 (校区)
  | 'level4'       // 4级信使 (城市总代)

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
  courierId: string
  codeId: string
  taskType: TaskType
  status: TaskStatus
  zone?: string
  priority: number
  assignedAt: Date
  completedAt?: Date
  note?: string
  description?: string
}

/**
 * 信使信息
 */
export interface Courier {
  id: string
  userId: string
  level: CourierLevel
  zone: string
  schoolCode: string
  isActive: boolean
  score: number
  completedTasks: number
  createdAt: Date
  updatedAt: Date
}

/**
 * 信使统计
 */
export interface CourierStats {
  totalTasks: number
  completedTasks: number
  pendingTasks: number
  successRate: number
  averageDeliveryTime: number // 小时
  ranking: number
  points: number
}

/**
 * 扫码记录
 */
export interface ScanRecord {
  id: string
  courierId: string
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
  userId: string
  level: CourierLevel
  zone: string
  reason: string
  status: 'pending' | 'approved' | 'rejected'
  appliedAt: Date
  reviewedAt?: Date
  reviewedBy?: string
  reviewNote?: string
}

/**
 * 创建信使任务请求
 */
export interface CreateTaskRequest {
  codeId: string
  taskType: TaskType
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
  courierId: string
  courierName: string
  level: CourierLevel
  zone: string
  period: {
    start: Date
    end: Date
  }
  metrics: {
    totalTasks: number
    completedTasks: number
    failedTasks: number
    averageCompletionTime: number
    successRate: number
    userRating: number
  }
  ranking: number
  rewards: {
    points: number
    badges: string[]
    bonuses: number
  }
}