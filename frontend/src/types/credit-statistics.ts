/**
 * 积分统计扩展类型定义
 */

import type { CreditTaskStatistics } from './credit'

/**
 * 每日统计数据
 */
export interface DailyStatistic {
  date: string
  earned: number
  spent: number
  tasks: number
  points: number  // 用于图表显示的积分数（可能等于 earned）
}

/**
 * 任务类型统计
 */
export interface TaskTypeStatistic {
  type: string
  count: number
  points: number
  percentage: number
}

/**
 * 完整的积分统计信息
 */
export interface CreditStatisticsData {
  // 基础统计
  totalEarned: number
  earnGrowth: number
  tasksCompleted: number
  tasksTotal: number
  dailyAverage: number
  maxDaily: number
  currentRank: number
  
  // 扩展统计
  totalUsers?: number
  dailyBreakdown?: DailyStatistic[]
  taskTypeBreakdown?: TaskTypeStatistic[]
  
  // 时间段统计
  todayEarned?: number
  weekEarned?: number
  monthEarned?: number
  
  // 任务状态统计
  tasksExecuting?: number
  tasksFailed?: number
  
  // 用户活跃度
  avgResponseTime?: number
  activeDays?: number
  streakDays?: number
}

/**
 * 积分统计响应
 */
export interface CreditStatisticsResponse {
  statistics: CreditStatisticsData
  period: string
  updated_at: string
}