/**
 * API Response Types
 * API响应类型定义
 */

export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  message?: string
  error?: string
  code?: number
}

export interface ApiError {
  code: number
  message: string
  details?: any
}

export interface PaginatedResponse<T = any> {
  success: boolean
  data: T[]
  pagination: {
    current_page: number
    per_page: number
    total: number
    total_pages: number
    has_next: boolean
    has_prev: boolean
  }
  message?: string
}