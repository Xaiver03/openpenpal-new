/**
 * 用户角色枚举 - 符合PRD的简化版本
 */
export type UserRole = 
  | 'user'                // 普通用户
  | 'courier_level1'      // 一级信使（基础投递信使）
  | 'courier_level2'      // 二级信使（片区协调员）
  | 'courier_level3'      // 三级信使（校区负责人）
  | 'courier_level4'      // 四级信使（城市负责人）
  | 'platform_admin'      // 平台管理员
  | 'super_admin'         // 超级管理员

/**
 * 用户状态枚举
 */
export type UserStatus = 
  | 'active'       // 活跃
  | 'inactive'     // 非活跃
  | 'banned'       // 被禁用

/**
 * 信使等级 (1-4级)
 */
export type CourierLevel = 1 | 2 | 3 | 4

/**
 * 信使信息
 */
export interface CourierInfo {
  level: CourierLevel
  zoneCode: string
  zoneType: 'city' | 'school' | 'zone' | 'building'
  status: string
  points: number
  taskCount: number
  completedTasks: number
  averageRating: number
  lastActiveAt: string
  school_code: string
  username: string
  school_name: string
}

/**
 * 用户信息 - 与后端JSON完全一致
 */
export interface User {
  id: string
  username: string
  email: string
  nickname: string
  avatar: string
  role: UserRole
  school_code: string           // Matches backend JSON tag
  school_name?: string          // School display name
  bio?: string                  // User biography
  is_active: boolean           // Matches backend JSON tag
  last_login_at?: string       // Matches backend JSON tag (time.Time -> string)
  created_at: string           // Matches backend JSON tag
  updated_at: string           // Matches backend JSON tag
  
  // 信使信息（仅信使角色拥有）
  courierInfo?: CourierInfo
  
  // 关联数据 (optional, matches backend gorm tags)
  sent_letters?: any[]
  authored_letters?: any[]
}

/**
 * 用户档案 - 扩展用户信息
 */
export interface UserProfile {
  // 基础用户信息
  id: string
  username: string
  email: string
  nickname: string
  avatar: string
  role: UserRole
  school_code: string
  is_active: boolean
  last_login_at?: string
  created_at: string
  updated_at: string
  
  // 档案扩展信息
  phone?: string
  bio?: string
  address?: string
}

/**
 * 注册请求 - 与后端API一致
 */
export interface RegisterRequest {
  username: string
  email: string
  password: string
  nickname: string
  school_code: string        // Matches backend expected field
  school_name?: string       // Optional school name
}

/**
 * 登录请求
 */
export interface LoginRequest {
  username: string
  password: string
}

/**
 * 登录响应 - 与后端API一致
 */
export interface LoginResponse {
  token: string
  refresh_token: string        // Matches backend JSON field
  expires_at: string          // Matches backend JSON field
  user: User
}

/**
 * 用户统计信息 - 字段命名与后端保持一致
 */
export interface UserStats {
  letters_sent: number       // 与后端 JSON 字段一致
  letters_received: number   // 与后端 JSON 字段一致
  drafts_count: number       // 与后端 JSON 字段一致
  delivered_count: number    // 与后端 JSON 字段一致（仅信使角色）
}

/**
 * 管理员更新用户请求 - 与后端API一致
 */
export interface AdminUpdateUserRequest {
  nickname?: string
  email?: string
  role?: UserRole
  school_code?: string       // Matches backend expected field
  is_active?: boolean        // Matches backend expected field
}

/**
 * 更新用户档案请求
 */
export interface UpdateProfileRequest {
  nickname?: string
  avatar?: string
  bio?: string
  address?: string
}

