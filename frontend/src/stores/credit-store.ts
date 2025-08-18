/**
 * Credit Store - 积分系统状态管理
 * Unified state management for credit data, tasks, transactions, and statistics
 */

import { create } from 'zustand'
import { devtools } from 'zustand/middleware'
import type {
  UserCredit,
  UserCreditLeaderboard,
  CreditTransaction,
  CreditTask,
  CreditTaskStatistics,
  CreditSummary,
  CreditHistoryParams,
  CreditTaskListParams,
  CreditTaskType,
  CreditTaskStatus
} from '@/types/credit'
import * as creditApi from '@/lib/api/credit'

interface CreditState {
  // 积分信息
  userCredit: UserCredit | null
  creditSummary: CreditSummary | null
  leaderboard: UserCreditLeaderboard[]
  
  // 积分历史
  transactions: CreditTransaction[]
  transactionTotal: number
  transactionPage: number
  transactionLimit: number
  
  // 积分任务
  tasks: CreditTask[]
  taskTotal: number
  taskPage: number
  taskLimit: number
  taskStatistics: CreditTaskStatistics[]
  
  // 加载状态
  loading: {
    credit: boolean
    summary: boolean
    leaderboard: boolean
    transactions: boolean
    tasks: boolean
    statistics: boolean
  }
  
  // 错误状态
  error: string | null
  
  // Actions
  fetchUserCredit: () => Promise<void>
  fetchCreditSummary: () => Promise<void>
  fetchLeaderboard: (limit?: number) => Promise<void>
  fetchTransactions: (params?: CreditHistoryParams) => Promise<void>
  fetchTasks: (params?: CreditTaskListParams) => Promise<void>
  fetchTaskStatistics: (timeRange?: 'today' | 'week' | 'month' | 'all') => Promise<void>
  
  // Refresh actions
  refreshAll: () => Promise<void>
  clearError: () => void
  
  // Utility actions
  getTasksByType: (taskType: CreditTaskType) => CreditTask[]
  getTasksByStatus: (status: CreditTaskStatus) => CreditTask[]
  getTotalEarnedToday: () => number
  getPendingTasksCount: () => number
}

export const useCreditStore = create<CreditState>()(
  devtools(
    (set, get) => ({
      // Initial state
      userCredit: null,
      creditSummary: null,
      leaderboard: [],
      transactions: [],
      transactionTotal: 0,
      transactionPage: 1,
      transactionLimit: 20,
      tasks: [],
      taskTotal: 0,
      taskPage: 1,
      taskLimit: 20,
      taskStatistics: [],
      
      loading: {
        credit: false,
        summary: false,
        leaderboard: false,
        transactions: false,
        tasks: false,
        statistics: false,
      },
      
      error: null,
      
      // Fetch user credit
      fetchUserCredit: async () => {
        set((state) => ({
          loading: { ...state.loading, credit: true },
          error: null,
        }))
        
        try {
          const credit = await creditApi.getUserCredit()
          set({ userCredit: credit })
        } catch (error) {
          set({ error: error instanceof Error ? error.message : 'Failed to fetch credit' })
        } finally {
          set((state) => ({
            loading: { ...state.loading, credit: false }
          }))
        }
      },
      
      // Fetch credit summary
      fetchCreditSummary: async () => {
        set((state) => ({
          loading: { ...state.loading, summary: true },
          error: null,
        }))
        
        try {
          const summary = await creditApi.getCreditSummary()
          set({ creditSummary: summary })
        } catch (error) {
          set({ error: error instanceof Error ? error.message : 'Failed to fetch credit summary' })
        } finally {
          set((state) => ({
            loading: { ...state.loading, summary: false }
          }))
        }
      },
      
      // Fetch leaderboard
      fetchLeaderboard: async (limit = 10) => {
        set((state) => ({
          loading: { ...state.loading, leaderboard: true },
          error: null,
        }))
        
        try {
          const leaderboard = await creditApi.getCreditLeaderboard(limit)
          set({ leaderboard })
        } catch (error) {
          set({ error: error instanceof Error ? error.message : 'Failed to fetch leaderboard' })
        } finally {
          set((state) => ({
            loading: { ...state.loading, leaderboard: false }
          }))
        }
      },
      
      // Fetch transactions
      fetchTransactions: async (params = {}) => {
        set((state) => ({
          loading: { ...state.loading, transactions: true },
          error: null,
        }))
        
        try {
          const defaultParams = {
            page: get().transactionPage,
            limit: get().transactionLimit,
            ...params
          }
          
          const response = await creditApi.getCreditHistory(defaultParams)
          set({
            transactions: response.transactions,
            transactionTotal: response.total,
            transactionPage: response.page,
            transactionLimit: response.limit,
          })
        } catch (error) {
          set({ error: error instanceof Error ? error.message : 'Failed to fetch transactions' })
        } finally {
          set((state) => ({
            loading: { ...state.loading, transactions: false }
          }))
        }
      },
      
      // Fetch tasks
      fetchTasks: async (params = {}) => {
        set((state) => ({
          loading: { ...state.loading, tasks: true },
          error: null,
        }))
        
        try {
          const defaultParams = {
            page: get().taskPage,
            limit: get().taskLimit,
            ...params
          }
          
          const response = await creditApi.getCreditTasks(defaultParams)
          set({
            tasks: response.tasks,
            taskTotal: response.total,
            taskPage: response.page,
            taskLimit: response.limit,
          })
        } catch (error) {
          set({ error: error instanceof Error ? error.message : 'Failed to fetch tasks' })
        } finally {
          set((state) => ({
            loading: { ...state.loading, tasks: false }
          }))
        }
      },
      
      // Fetch task statistics
      fetchTaskStatistics: async (timeRange = 'week') => {
        set((state) => ({
          loading: { ...state.loading, statistics: true },
          error: null,
        }))
        
        try {
          const statistics = await creditApi.getCreditTaskStatistics(timeRange)
          set({ taskStatistics: statistics })
        } catch (error) {
          set({ error: error instanceof Error ? error.message : 'Failed to fetch task statistics' })
        } finally {
          set((state) => ({
            loading: { ...state.loading, statistics: false }
          }))
        }
      },
      
      // Refresh all data
      refreshAll: async () => {
        const { fetchUserCredit, fetchCreditSummary, fetchTransactions, fetchTasks } = get()
        await Promise.allSettled([
          fetchUserCredit(),
          fetchCreditSummary(),
          fetchTransactions(),
          fetchTasks(),
        ])
      },
      
      // Clear error
      clearError: () => set({ error: null }),
      
      // Utility functions
      getTasksByType: (taskType: CreditTaskType) => {
        return get().tasks.filter(task => task.task_type === taskType)
      },
      
      getTasksByStatus: (status: CreditTaskStatus) => {
        return get().tasks.filter(task => task.status === status)
      },
      
      getTotalEarnedToday: () => {
        const today = new Date().toISOString().split('T')[0]
        return get().transactions
          .filter(transaction => 
            transaction.type === 'earn' && 
            transaction.created_at.startsWith(today)
          )
          .reduce((total, transaction) => total + transaction.amount, 0)
      },
      
      getPendingTasksCount: () => {
        return get().tasks.filter(task => 
          task.status === 'pending' || task.status === 'scheduled'
        ).length
      },
    }),
    {
      name: 'credit-store', // 用于devtools调试
    }
  )
)

// Selector hooks for better performance
export const useCreditInfo = () => useCreditStore((state) => ({
  userCredit: state.userCredit,
  creditSummary: state.creditSummary,
  loading: state.loading.credit || state.loading.summary,
  error: state.error,
}))

export const useCreditTransactions = () => useCreditStore((state) => ({
  transactions: state.transactions,
  total: state.transactionTotal,
  page: state.transactionPage,
  limit: state.transactionLimit,
  loading: state.loading.transactions,
  error: state.error,
}))

export const useCreditTasks = () => useCreditStore((state) => ({
  tasks: state.tasks,
  total: state.taskTotal,
  page: state.taskPage,
  limit: state.taskLimit,
  statistics: state.taskStatistics,
  loading: state.loading.tasks || state.loading.statistics,
  error: state.error,
}))

export const useCreditLeaderboard = () => useCreditStore((state) => ({
  leaderboard: state.leaderboard,
  loading: state.loading.leaderboard,
  error: state.error,
}))

export const useCreditStatistics = () => useCreditStore((state) => ({
  statistics: state.taskStatistics,
  loading: state.loading.statistics,
  error: state.error,
}))

// Helper function to initialize credit store
export const initializeCreditStore = async () => {
  const store = useCreditStore.getState()
  await store.refreshAll()
}