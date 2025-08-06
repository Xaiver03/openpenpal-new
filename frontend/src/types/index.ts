/**
 * Unified Type Definitions for OpenPenPal Frontend
 * OpenPenPal前端统一类型定义
 * 
 * This file consolidates all type definitions to ensure consistency
 * across the codebase and eliminate duplicate interfaces.
 */

// Re-export existing type modules for backward compatibility
export * from './user'
export * from './auth'
export * from './courier'
export * from './api'

// ================================
// Core Entity Types
// ================================

/** 
 * 用户基础信息接口
 * Basic user information interface
 */
export interface BaseUser {
  id: string
  username: string
  nickname: string
  email: string
  avatar?: string
  bio?: string
  createdAt: string
  updatedAt: string
  last_login_at?: string
  status: 'active' | 'inactive' | 'banned'
  is_active?: boolean
}

/**
 * 学校信息接口
 * School information interface
 */
export interface School {
  id: string
  code: string
  name: string
  address?: string
  city?: string
  province?: string
  contact_email?: string
  contact_phone?: string
  status: 'active' | 'inactive'
  createdAt: string
  updatedAt: string
}

/**
 * 信件信息接口
 * Letter information interface
 */
export interface Letter {
  id: string
  code: string
  title?: string
  content: string
  style: 'classic' | 'modern' | 'elegant' | 'casual'
  sender_id: string
  recipient_info?: {
    name?: string
    address?: string
    phone?: string
  }
  status: 'draft' | 'sent' | 'in_transit' | 'delivered' | 'read'
  qr_code?: string
  read_url?: string
  createdAt: string
  updatedAt: string
  delivered_at?: string
  read_at?: string
}

/**
 * 邮政编码信息接口
 * Postal code information interface
 */
export interface PostalCode {
  code: string
  province: string
  city: string
  district: string
  area?: string
  full_address: string
  latitude?: number
  longitude?: number
}

/**
 * 地址信息接口
 * Address information interface
 */
export interface Address {
  id?: string
  province: string
  city: string
  district: string
  street: string
  building?: string
  room?: string
  postal_code?: string
  full_address: string
  coordinates?: {
    latitude: number
    longitude: number
  }
}

// ================================
// Form and Input Types
// ================================

/**
 * 登录表单数据
 * Login form data
 */
export interface LoginFormData {
  username: string
  password: string
  remember_me?: boolean
}

/**
 * 注册表单数据
 * Registration form data
 */
export interface RegisterFormData {
  username: string
  password: string
  confirm_password: string
  email: string
  nickname?: string
  school_code?: string
}

/**
 * 个人资料更新数据
 * Profile update data
 */
export interface ProfileUpdateData {
  nickname?: string
  email?: string
  avatar?: string
  bio?: string
  address?: string
}

// ================================
// UI Component Types
// ================================

/**
 * 分页信息接口
 * Pagination information interface
 */
export interface PaginationInfo {
  current_page: number
  per_page: number
  total: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

/**
 * 排序选项接口
 * Sort option interface
 */
export interface SortOption {
  value: string
  label: string
  direction?: 'asc' | 'desc'
}

/**
 * 筛选选项接口
 * Filter option interface
 */
export interface FilterOption {
  value: string | number
  label: string
  count?: number
}

/**
 * 搜索参数接口
 * Search parameters interface
 */
export interface SearchParams {
  query?: string
  filters?: Record<string, any>
  sort?: {
    field: string
    direction: 'asc' | 'desc'
  }
  pagination?: {
    page: number
    per_page: number
  }
}

/**
 * 表格列定义接口
 * Table column definition interface
 */
export interface TableColumn<T = any> {
  key: keyof T | string
  title: string
  dataIndex?: keyof T
  width?: number | string
  align?: 'left' | 'center' | 'right'
  sortable?: boolean
  filterable?: boolean
  render?: (value: any, record: T, index: number) => React.ReactNode
}

// ================================
// Error and Loading Types
// ================================

/**
 * 错误信息接口
 * Error information interface
 */
export interface ErrorInfo {
  code: string | number
  message: string
  details?: any
  timestamp?: string
  stack?: string
}

/**
 * 加载状态接口
 * Loading state interface
 */
export interface LoadingState {
  isLoading: boolean
  isRefreshing?: boolean
  error?: string | null
  lastUpdated?: number | null
}

/**
 * 异步操作结果接口
 * Async operation result interface
 */
export interface AsyncResult<T = any> {
  success: boolean
  data?: T
  error?: ErrorInfo
  message?: string
}

// ================================
// Event and Callback Types
// ================================

/**
 * 事件处理器类型
 * Event handler types
 */
export type EventHandler<T = any> = (event: T) => void
export type AsyncEventHandler<T = any> = (event: T) => Promise<void>
export type ChangeHandler<T = any> = (value: T) => void
export type SubmitHandler<T = any> = (data: T) => void | Promise<void>

/**
 * 通用回调函数类型
 * Generic callback function types
 */
export type Callback<T = void> = () => T
export type AsyncCallback<T = void> = () => Promise<T>
export type CallbackWithParams<P = any, R = void> = (params: P) => R
export type AsyncCallbackWithParams<P = any, R = void> = (params: P) => Promise<R>

// ================================
// Utility Types
// ================================

/**
 * 可选字段类型
 * Optional fields type
 */
export type Optional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>

/**
 * 必需字段类型
 * Required fields type
 */
export type RequiredFields<T, K extends keyof T> = T & Required<Pick<T, K>>

/**
 * 深度部分类型
 * Deep partial type
 */
export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P]
}

/**
 * 时间戳类型
 * Timestamp types
 */
export type Timestamp = string | number | Date
export type ISOString = string

/**
 * ID类型
 * ID types
 */
export type ID = string | number
export type UUID = string

// ================================
// Configuration Types
// ================================

/**
 * 应用配置接口
 * Application configuration interface
 */
export interface AppConfig {
  app_name: string
  version: string
  environment: 'development' | 'staging' | 'production'
  api_base_url: string
  websocket_url: string
  features: {
    debug_panel: boolean
    performance_monitoring: boolean
    error_boundary: boolean
    [key: string]: boolean
  }
}

/**
 * 主题配置接口
 * Theme configuration interface
 */
export interface ThemeConfig {
  mode: 'light' | 'dark' | 'auto'
  primary_color: string
  font_family: string
  font_size: 'small' | 'medium' | 'large'
  border_radius: 'none' | 'small' | 'medium' | 'large'
}

// ================================
// Environment and Context Types
// ================================

/**
 * 浏览器环境信息
 * Browser environment information
 */
export interface BrowserInfo {
  user_agent: string
  is_mobile: boolean
  is_tablet: boolean
  is_desktop: boolean
  browser_name: string
  browser_version: string
  os_name: string
  os_version: string
  screen_resolution: {
    width: number
    height: number
  }
}

/**
 * 设备信息接口
 * Device information interface
 */
export interface DeviceInfo {
  type: 'mobile' | 'tablet' | 'desktop'
  os: 'ios' | 'android' | 'windows' | 'macos' | 'linux' | 'unknown'
  browser: 'chrome' | 'firefox' | 'safari' | 'edge' | 'opera' | 'unknown'
  supports_touch: boolean
  screen_size: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl'
  preferred_input: 'touch' | 'mouse' | 'keyboard'
}

// ================================
// Export Declarations Complete
// All types are exported via interface declarations above
// ================================