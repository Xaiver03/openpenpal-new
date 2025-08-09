import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { apiClient } from '@/lib/api-client'
import type { Courier, CourierTask } from '@/lib/api/courier'

interface CourierState {
  courier: Courier | null
  tasks: CourierTask[]
  loading: boolean
  error: string | null
  
  // Actions
  setCourier: (courier: Courier | null) => void
  setTasks: (tasks: CourierTask[]) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  
  // Async actions
  fetchCourierInfo: () => Promise<void>
  fetchTasks: (status?: string) => Promise<void>
  acceptTask: (taskId: string) => Promise<boolean>
  completeTask: (taskId: string) => Promise<boolean>
  scanQRCode: (code: string, action: string) => Promise<boolean>
}

export const useCourierStore = create<CourierState>()(
  devtools(
    persist(
      (set, get) => ({
        courier: null,
        tasks: [],
        loading: false,
        error: null,
        
        setCourier: (courier) => set({ courier }),
        setTasks: (tasks) => set({ tasks }),
        setLoading: (loading) => set({ loading }),
        setError: (error) => set({ error }),
        
        fetchCourierInfo: async () => {
          set({ loading: true, error: null })
          try {
            const response = await apiClient.get('/api/v1/courier/info')
            set({ courier: (response as any).data, loading: false })
          } catch (error) {
            set({ 
              error: error instanceof Error ? error.message : '获取信使信息失败',
              loading: false 
            })
          }
        },
        
        fetchTasks: async (status?: string) => {
          set({ loading: true, error: null })
          try {
            const response = await apiClient.get('/api/v1/courier/tasks' + (status ? `?status=${status}` : ''))
            set({ tasks: (response as any).data || [], loading: false })
          } catch (error) {
            set({ 
              error: error instanceof Error ? error.message : '获取任务列表失败',
              loading: false 
            })
          }
        },
        
        acceptTask: async (taskId: string) => {
          set({ loading: true, error: null })
          try {
            await apiClient.post(`/api/v1/courier/tasks/${taskId}/accept`)
            await get().fetchTasks()
            set({ loading: false })
            return true
          } catch (error) {
            set({ 
              error: error instanceof Error ? error.message : '接受任务失败',
              loading: false 
            })
            return false
          }
        },
        
        completeTask: async (taskId: string) => {
          set({ loading: true, error: null })
          try {
            await apiClient.post(`/api/v1/courier/tasks/${taskId}/complete`)
            await get().fetchTasks()
            set({ loading: false })
            return true
          } catch (error) {
            set({ 
              error: error instanceof Error ? error.message : '完成任务失败',
              loading: false 
            })
            return false
          }
        },
        
        scanQRCode: async (code: string, action: string) => {
          set({ loading: true, error: null })
          try {
            await apiClient.post('/api/v1/courier/scan', { code, action })
            await get().fetchTasks()
            set({ loading: false })
            return true
          } catch (error) {
            set({ 
              error: error instanceof Error ? error.message : '扫码操作失败',
              loading: false 
            })
            return false
          }
        }
      }),
      {
        name: 'courier-storage',
        partialize: (state) => ({ courier: state.courier })
      }
    )
  )
)