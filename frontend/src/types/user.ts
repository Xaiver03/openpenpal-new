/**
 * 用户角色枚举
 */
export type UserRole = 
  | 'user'                // 普通用户
  | 'courier'             // 普通信使
  | 'senior_courier'      // 高级信使
  | 'courier_coordinator' // 信使协调员
  | 'school_admin'        // 学校管理员
  | 'platform_admin'      // 平台管理员
  | 'super_admin'         // 超级管理员
  // 分级信使系统 (兼容性)
  | 'courier_level1'      // 一级信使
  | 'courier_level2'      // 二级信使
  | 'courier_level3'      // 三级信使
  | 'courier_level4'      // 四级信使

/**
 * 用户状态枚举
 */
export type UserStatus = 
  | 'active'       // 活跃
  | 'inactive'     // 非活跃
  | 'banned'       // 被禁用

/**
 * 用户信息
 */
export interface User {
  id: string
  username: string
  email?: string
  nickname: string
  avatar?: string          // 对应backend的avatar字段
  role: UserRole
  schoolCode?: string      // 映射到backend的school_code
  isActive: boolean        // 映射到backend的is_active
  lastLoginAt?: Date       // 映射到backend的last_login_at
  createdAt: Date
  updatedAt: Date
  // 关联数据
  sentLetters?: any[]
  authoredLetters?: any[]
}

/**
 * 用户档案
 */
export interface UserProfile extends User {
  email?: string
  phone?: string
  realName?: string
  studentId?: string
  grade?: string
  major?: string
  bio?: string
}

/**
 * 登录请求
 */
export interface LoginRequest {
  code: string // 微信授权码
}

/**
 * 登录响应
 */
export interface LoginResponse {
  user: User
  token: string
  refreshToken: string
}

/**
 * 更新用户档案请求
 */
export interface UpdateProfileRequest {
  nickname?: string
  avatar?: string          // 对应backend的avatar字段
  schoolCode?: string      // 将映射为school_code
  email?: string
  phone?: string
  bio?: string
}

/**
 * 用户统计
 */
export interface UserStats {
  totalUsers: number
  activeUsers: number
  newUsersToday: number
  usersByRole: Record<UserRole, number>
  usersBySchool: Record<string, number>
}