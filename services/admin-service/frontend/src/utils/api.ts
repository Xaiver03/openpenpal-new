import axios from 'axios'
import type { AxiosResponse, AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

// 创建axios实例 - SOTA管理后台统一: 指向Go后端
export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1/admin',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - SOTA管理后台统一: 兼容Go后端JWT
api.interceptors.request.use(
  (config) => {
    // 优先使用admin_token，fallback到Go后端的token
    const token = localStorage.getItem('admin_token') || localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  (error: AxiosError) => {
    if (error.response) {
      const { status, data } = error.response as any
      
      switch (status) {
        case 401:
          ElMessage.error('登录已过期，请重新登录')
          localStorage.removeItem('admin_token')
          localStorage.removeItem('admin_user')
          router.push('/login')
          break
        case 403:
          ElMessage.error(data?.msg || '权限不足')
          break
        case 404:
          ElMessage.error(data?.msg || '请求的资源不存在')
          break
        case 500:
          ElMessage.error(data?.msg || '服务器内部错误')
          break
        default:
          ElMessage.error(data?.msg || '请求失败')
      }
    } else if (error.code === 'ECONNABORTED') {
      ElMessage.error('请求超时，请稍后重试')
    } else {
      ElMessage.error('网络错误，请检查网络连接')
    }
    
    return Promise.reject(error)
  }
)

// API响应类型
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
  timestamp: string
  error?: {
    type: string
    details: string
    field?: string
    traceId?: string
  }
}

// 分页响应类型
export interface PageResponse<T = any> {
  items: T[]
  pagination: {
    page: number
    limit: number
    total: number
    pages: number
    hasNext: boolean
    hasPrev: boolean
  }
}

// 常用API方法封装
export const userApi = {
  // 获取用户列表
  getUsers: (params: any) => api.get<ApiResponse<PageResponse>>('/users', { params }),
  
  // 获取用户详情
  getUser: (id: string) => api.get<ApiResponse>(`/users/${id}`),
  
  // 更新用户
  updateUser: (id: string, data: any) => api.put<ApiResponse>(`/users/${id}`, data),
  
  // 删除用户
  deleteUser: (id: string) => api.delete<ApiResponse>(`/users/${id}`),
  
  // 解锁用户
  unlockUser: (id: string) => api.post<ApiResponse>(`/users/${id}/unlock`),
  
  // 重置密码
  resetPassword: (id: string, password: string) => 
    api.post<ApiResponse>(`/users/${id}/reset-password`, { password }),
  
  // 用户统计
  getUserStats: () => api.get<ApiResponse>('/users/stats/role')
}

export const letterApi = {
  // 获取信件列表
  getLetters: (params: any) => api.get<ApiResponse<PageResponse>>('/letters', { params }),
  
  // 获取信件详情
  getLetter: (id: string) => api.get<ApiResponse>(`/letters/${id}`),
  
  // 更新信件状态
  updateLetterStatus: (id: string, data: any) => 
    api.put<ApiResponse>(`/letters/${id}/status`, data),
  
  // 批量更新状态
  batchUpdateStatus: (data: any) => api.put<ApiResponse>('/letters/batch/status', data),
  
  // 信件统计
  getLetterStats: () => api.get<ApiResponse>('/letters/stats/overview')
}

export const courierApi = {
  // 获取信使列表
  getCouriers: (params: any) => api.get<ApiResponse<PageResponse>>('/couriers', { params }),
  
  // 获取信使详情
  getCourier: (id: string) => api.get<ApiResponse>(`/couriers/${id}`),
  
  // 更新信使状态
  updateCourierStatus: (id: string, status: string) => 
    api.put<ApiResponse>(`/couriers/${id}/status`, { status }),
  
  // 信使统计
  getCourierStats: () => api.get<ApiResponse>('/couriers/stats/overview')
}

export const museumApi = {
  // 展览管理
  getExhibitions: (params?: any) => api.get<ApiResponse<PageResponse>>('/museum/exhibitions', { params }),
  createExhibition: (data: any) => api.post<ApiResponse>('/museum/exhibitions', data),
  updateExhibition: (id: string, data: any) => api.put<ApiResponse>(`/museum/exhibitions/${id}`, data),
  updateExhibitionStatus: (id: string, data: any) => api.put<ApiResponse>(`/museum/exhibitions/${id}/status`, data),
  deleteExhibition: (id: string) => api.delete<ApiResponse>(`/museum/exhibitions/${id}`),
  
  // 内容审核
  getModerationStats: () => api.get<ApiResponse>('/museum/moderation/statistics'),
  getModerationTasks: (params?: any) => api.get<ApiResponse<PageResponse>>('/museum/moderation/tasks', { params }),
  approveModerationTask: (id: string) => api.post<ApiResponse>(`/museum/moderation/tasks/${id}/approve`),
  rejectModerationTask: (id: string, data: any) => api.post<ApiResponse>(`/museum/moderation/tasks/${id}/reject`, data),
  batchModerationApprove: (data: any) => api.post<ApiResponse>('/museum/moderation/batch/approve', data),
  batchModerationReject: (data: any) => api.post<ApiResponse>('/museum/moderation/batch/reject', data),
  
  // 敏感词管理
  getSensitiveWords: (params?: any) => api.get<ApiResponse<PageResponse>>('/museum/sensitive-words', { params }),
  addSensitiveWord: (data: any) => api.post<ApiResponse>('/museum/sensitive-words', data),
  updateSensitiveWord: (id: string, data: any) => api.put<ApiResponse>(`/museum/sensitive-words/${id}`, data),
  updateSensitiveWordStatus: (id: string, data: any) => api.put<ApiResponse>(`/museum/sensitive-words/${id}/status`, data),
  deleteSensitiveWord: (id: string) => api.delete<ApiResponse>(`/museum/sensitive-words/${id}`),
  batchDeleteSensitiveWords: (data: any) => api.delete<ApiResponse>('/museum/sensitive-words/batch', { data }),
  
  // 举报管理
  getReports: (params?: any) => api.get<ApiResponse<PageResponse>>('/museum/reports', { params }),
  handleReport: (id: string, data: any) => api.post<ApiResponse>(`/museum/reports/${id}/handle`, data),
  batchHandleReports: (data: any) => api.post<ApiResponse>('/museum/reports/batch/handle', data)
}

export const systemApi = {
  // 系统配置
  getSystemConfigs: () => api.get<ApiResponse>('/system/config'),
  updateSystemConfig: (key: string, data: any) => api.put<ApiResponse>(`/system/config/${key}`, data),
  batchUpdateConfigs: (data: any) => api.put<ApiResponse>('/system/config/batch', data),
  
  // 系统信息
  getSystemInfo: () => api.get<ApiResponse>('/system/info'),
  getSystemHealth: () => api.get<ApiResponse>('/system/health'),
  
  // 权限管理
  getPermissions: () => api.get<ApiResponse>('/system/permissions'),
  getRoles: () => api.get<ApiResponse>('/system/roles'),
  updateRolePermissions: (roleId: string, data: any) => api.put<ApiResponse>(`/system/roles/${roleId}/permissions`, data)
}