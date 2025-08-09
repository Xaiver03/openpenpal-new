/**
 * 统一角色配置系统 - OpenPenPal (符合PRD简化版本)
 * Unified Role Configuration System (PRD Compliant Simplified Version)
 * 
 * 这个文件是所有角色相关配置的唯一来源
 * This file is the single source of truth for all role-related configurations
 */

import { 
  Mail, 
  Users, 
  Shield, 
  Crown, 
  Home, 
  MapPin, 
  School, 
  Building 
} from 'lucide-react'

// ================================
// 基础类型定义 - Basic Type Definitions
// ================================

/**
 * 系统中所有可能的用户角色 (只保留PRD要求的7种)
 * All possible user roles in the system (Only 7 roles per PRD)
 */
export type UserRole = 
  | 'user'                  // 普通用户
  | 'courier_level1'        // 一级信使（基础投递信使）
  | 'courier_level2'        // 二级信使（片区协调员）
  | 'courier_level3'        // 三级信使（校区负责人）
  | 'courier_level4'        // 四级信使（城市负责人）
  | 'platform_admin'        // 平台管理员
  | 'super_admin'           // 超级管理员

/**
 * 信使等级 (1-4级)
 * Courier levels (1-4)
 */
export type CourierLevel = 1 | 2 | 3 | 4

/**
 * 系统中所有权限
 * All permissions in the system
 */
export type Permission = 
  // 基础权限 - Basic Permissions
  | 'READ_LETTER'           // 阅读信件
  | 'WRITE_LETTER'          // 写信
  | 'MANAGE_PROFILE'        // 管理个人资料
  | 'VIEW_LETTER_SQUARE'    // 查看信件广场
  | 'VIEW_MUSEUM'           // 参观博物馆
  
  // 信使权限 - Courier Permissions
  | 'COURIER_SCAN_CODE'     // 扫码投递
  | 'COURIER_DELIVER_LETTER' // 投递信件
  | 'COURIER_VIEW_TASKS'    // 查看任务
  | 'COURIER_MANAGE_PROFILE' // 管理信使资料
  | 'COURIER_VIEW_STATISTICS' // 查看统计数据
  | 'COURIER_VIEW_POINTS'   // 查看积分
  | 'COURIER_EXCHANGE_REWARDS' // 兑换奖励
  | 'COURIER_VIEW_LEADERBOARD' // 查看排行榜
  | 'COURIER_MANAGE_SCHEDULE' // 管理投递时间表
  | 'COURIER_VIEW_DELIVERY_AREA' // 查看投递区域
  | 'COURIER_REPORT_ISSUES'  // 报告问题
  | 'COURIER_VIEW_FEEDBACK' // 查看反馈
  | 'COURIER_PARTICIPATE_ACTIVITIES' // 参与活动
  
  // 管理权限 - Management Permissions
  | 'MANAGE_USERS'          // 管理用户
  | 'MANAGE_LETTERS'        // 管理信件
  | 'MANAGE_COURIERS'       // 管理信使
  | 'MANAGE_SCHOOLS'        // 管理学校
  | 'MANAGE_SYSTEM_SETTINGS' // 管理系统设置
  | 'VIEW_ANALYTICS'        // 查看数据分析
  | 'MANAGE_CONTENT'        // 管理内容
  | 'MODERATE_CONTENT'      // 内容审核
  | 'MANAGE_ANNOUNCEMENTS'  // 管理公告
  | 'AUDIT_LOGS'            // 审计日志

// ================================
// 角色配置 - Role Configurations
// ================================

/**
 * 角色配置接口
 * Role configuration interface
 */
export interface RoleConfig {
  id: UserRole
  name: string              // 中文显示名称
  englishName: string       // 英文名称
  description: string       // 角色描述
  hierarchy: number         // 权限层级 (1-7, 数字越大权限越高)
  color: {
    bg: string              // 背景色 (Tailwind CSS类名)
    text: string            // 文字色 (Tailwind CSS类名)
    badge: string           // 徽章样式
    hover: string           // 悬停样式
  }
  icon: string              // 图标字符串
  iconComponent?: any       // Lucide图标组件
  permissions: Permission[] // 拥有的权限列表
  defaultHomePage: string   // 默认首页路径
  canAccessAdmin: boolean   // 是否可以访问管理后台
  isSystemRole: boolean     // 是否为系统角色
}

/**
 * 信使等级配置接口
 * Courier level configuration interface
 */
export interface CourierLevelConfig {
  level: CourierLevel
  name: string              // 中文名称
  englishName: string       // 英文名称
  description: string       // 等级描述
  managementArea: string    // 管理范围
  color: {
    bg: string
    text: string
    badge: string
    hover: string
  }
  icon: string
  iconComponent?: any
  permissions: Permission[]
  managementPath: string    // 管理后台路径
  canManageSublevels: boolean // 是否可以管理下级
}

// ================================
// 角色配置数据 - Role Configuration Data
// ================================

/**
 * 统一角色配置 (符合PRD的7种角色)
 * Unified role configuration (7 roles per PRD)
 */
export const ROLE_CONFIGS: Record<UserRole, RoleConfig> = {
  // 普通用户
  user: {
    id: 'user',
    name: '普通用户',
    englishName: 'User',
    description: '平台的普通用户，可以写信、阅读、参观博物馆',
    hierarchy: 1,
    color: {
      bg: 'bg-gray-600',
      text: 'text-white',
      badge: 'bg-gray-100 text-gray-800',
      hover: 'hover:bg-gray-700'
    },
    icon: '👤',
    iconComponent: Users,
    permissions: [
      'READ_LETTER',
      'WRITE_LETTER',
      'MANAGE_PROFILE',
      'VIEW_LETTER_SQUARE',
      'VIEW_MUSEUM'
    ],
    defaultHomePage: '/write',
    canAccessAdmin: false,
    isSystemRole: false
  },

  // 一级信使：基础投递信使
  courier_level1: {
    id: 'courier_level1',
    name: '一级信使（基础投递）',
    englishName: 'Level 1 Courier',
    description: '基础投递信使，负责宿舍楼栋、商店路径等具体投递任务',
    hierarchy: 2,
    color: {
      bg: 'bg-amber-600',
      text: 'text-white',
      badge: 'bg-amber-100 text-amber-800',
      hover: 'hover:bg-amber-700'
    },
    icon: '📮',
    iconComponent: Mail,
    permissions: [
      'READ_LETTER',
      'WRITE_LETTER',
      'MANAGE_PROFILE',
      'VIEW_LETTER_SQUARE',
      'VIEW_MUSEUM',
      'COURIER_SCAN_CODE',
      'COURIER_DELIVER_LETTER',
      'COURIER_VIEW_TASKS',
      'COURIER_MANAGE_PROFILE',
      'COURIER_VIEW_STATISTICS',
      'COURIER_VIEW_POINTS',
      'COURIER_MANAGE_SCHEDULE',
      'COURIER_VIEW_DELIVERY_AREA',
      'COURIER_REPORT_ISSUES'
    ],
    defaultHomePage: '/courier',
    canAccessAdmin: false,
    isSystemRole: false
  },

  // 二级信使：片区协调员
  courier_level2: {
    id: 'courier_level2',
    name: '二级信使（片区协调员）',
    englishName: 'Level 2 Courier',
    description: '片区协调员，管理宿舍区/楼栋组/商业片区，分发任务给一级信使',
    hierarchy: 3,
    color: {
      bg: 'bg-green-600',
      text: 'text-white',
      badge: 'bg-green-100 text-green-800',
      hover: 'hover:bg-green-700'
    },
    icon: '📍',
    iconComponent: MapPin,
    permissions: [
      'READ_LETTER',
      'WRITE_LETTER',
      'MANAGE_PROFILE',
      'VIEW_LETTER_SQUARE',
      'VIEW_MUSEUM',
      'COURIER_SCAN_CODE',
      'COURIER_DELIVER_LETTER',
      'COURIER_VIEW_TASKS',
      'COURIER_MANAGE_PROFILE',
      'COURIER_VIEW_STATISTICS',
      'COURIER_VIEW_POINTS',
      'COURIER_EXCHANGE_REWARDS',
      'COURIER_VIEW_LEADERBOARD',
      'COURIER_MANAGE_SCHEDULE',
      'COURIER_VIEW_DELIVERY_AREA',
      'COURIER_REPORT_ISSUES',
      'COURIER_VIEW_FEEDBACK'
    ],
    defaultHomePage: '/courier',
    canAccessAdmin: true,
    isSystemRole: false
  },

  // 三级信使：校区负责人
  courier_level3: {
    id: 'courier_level3',
    name: '三级信使（校区负责人）',
    englishName: 'Level 3 Courier',
    description: '校区负责人，管理所在学校的信使网络，任命二级信使',
    hierarchy: 4,
    color: {
      bg: 'bg-blue-600',
      text: 'text-white',
      badge: 'bg-blue-100 text-blue-800',
      hover: 'hover:bg-blue-700'
    },
    icon: '🏫',
    iconComponent: School,
    permissions: [
      'READ_LETTER',
      'WRITE_LETTER',
      'MANAGE_PROFILE',
      'VIEW_LETTER_SQUARE',
      'VIEW_MUSEUM',
      'COURIER_SCAN_CODE',
      'COURIER_DELIVER_LETTER',
      'COURIER_VIEW_TASKS',
      'COURIER_MANAGE_PROFILE',
      'COURIER_VIEW_STATISTICS',
      'COURIER_VIEW_POINTS',
      'COURIER_EXCHANGE_REWARDS',
      'COURIER_VIEW_LEADERBOARD',
      'COURIER_MANAGE_SCHEDULE',
      'COURIER_VIEW_DELIVERY_AREA',
      'COURIER_REPORT_ISSUES',
      'COURIER_VIEW_FEEDBACK',
      'COURIER_PARTICIPATE_ACTIVITIES',
      'MANAGE_COURIERS',
      'VIEW_ANALYTICS'
    ],
    defaultHomePage: '/courier',
    canAccessAdmin: true,
    isSystemRole: false
  },

  // 四级信使：城市负责人
  courier_level4: {
    id: 'courier_level4',
    name: '四级信使（城市负责人）',
    englishName: 'Level 4 Courier',
    description: '城市负责人，管理所在城市所有学校的信使网络，开通新学校',
    hierarchy: 5,
    color: {
      bg: 'bg-purple-600',
      text: 'text-white',
      badge: 'bg-purple-100 text-purple-800',
      hover: 'hover:bg-purple-700'
    },
    icon: '👑',
    iconComponent: Crown,
    permissions: [
      'READ_LETTER',
      'WRITE_LETTER',
      'MANAGE_PROFILE',
      'VIEW_LETTER_SQUARE',
      'VIEW_MUSEUM',
      'COURIER_SCAN_CODE',
      'COURIER_DELIVER_LETTER',
      'COURIER_VIEW_TASKS',
      'COURIER_MANAGE_PROFILE',
      'COURIER_VIEW_STATISTICS',
      'COURIER_VIEW_POINTS',
      'COURIER_EXCHANGE_REWARDS',
      'COURIER_VIEW_LEADERBOARD',
      'COURIER_MANAGE_SCHEDULE',
      'COURIER_VIEW_DELIVERY_AREA',
      'COURIER_REPORT_ISSUES',
      'COURIER_VIEW_FEEDBACK',
      'COURIER_PARTICIPATE_ACTIVITIES',
      'MANAGE_COURIERS',
      'MANAGE_SCHOOLS',
      'VIEW_ANALYTICS'
    ],
    defaultHomePage: '/courier',
    canAccessAdmin: true,
    isSystemRole: false
  },

  // 平台管理员
  platform_admin: {
    id: 'platform_admin',
    name: '平台管理员',
    englishName: 'Platform Admin',
    description: '平台管理员，具有平台级别的管理权限',
    hierarchy: 6,
    color: {
      bg: 'bg-blue-600',
      text: 'text-white',
      badge: 'bg-blue-100 text-blue-800',
      hover: 'hover:bg-blue-700'
    },
    icon: '🛡️',
    iconComponent: Shield,
    permissions: [
      'READ_LETTER',
      'WRITE_LETTER',
      'MANAGE_PROFILE',
      'VIEW_LETTER_SQUARE',
      'VIEW_MUSEUM',
      'MANAGE_USERS',
      'MANAGE_LETTERS',
      'MANAGE_COURIERS',
      'MANAGE_SCHOOLS',
      'MANAGE_SYSTEM_SETTINGS',
      'VIEW_ANALYTICS',
      'MANAGE_CONTENT',
      'MODERATE_CONTENT',
      'MANAGE_ANNOUNCEMENTS',
      'AUDIT_LOGS'
    ],
    defaultHomePage: '/admin/dashboard',
    canAccessAdmin: true,
    isSystemRole: true
  },

  // 超级管理员
  super_admin: {
    id: 'super_admin',
    name: '超级管理员',
    englishName: 'Super Admin',
    description: '系统超级管理员，拥有所有权限',
    hierarchy: 7,
    color: {
      bg: 'bg-purple-600',
      text: 'text-white',
      badge: 'bg-red-100 text-red-800',
      hover: 'hover:bg-purple-700'
    },
    icon: '👑',
    iconComponent: Crown,
    permissions: [
      'READ_LETTER',
      'WRITE_LETTER',
      'MANAGE_PROFILE',
      'VIEW_LETTER_SQUARE',
      'VIEW_MUSEUM',
      'COURIER_SCAN_CODE',
      'COURIER_DELIVER_LETTER',
      'COURIER_VIEW_TASKS',
      'COURIER_MANAGE_PROFILE',
      'COURIER_VIEW_STATISTICS',
      'COURIER_VIEW_POINTS',
      'COURIER_EXCHANGE_REWARDS',
      'COURIER_VIEW_LEADERBOARD',
      'COURIER_MANAGE_SCHEDULE',
      'COURIER_VIEW_DELIVERY_AREA',
      'COURIER_REPORT_ISSUES',
      'COURIER_VIEW_FEEDBACK',
      'COURIER_PARTICIPATE_ACTIVITIES',
      'MANAGE_USERS',
      'MANAGE_LETTERS',
      'MANAGE_COURIERS',
      'MANAGE_SCHOOLS',
      'MANAGE_SYSTEM_SETTINGS',
      'VIEW_ANALYTICS',
      'MANAGE_CONTENT',
      'MODERATE_CONTENT',
      'MANAGE_ANNOUNCEMENTS',
      'AUDIT_LOGS'
    ],
    defaultHomePage: '/admin/dashboard',
    canAccessAdmin: true,
    isSystemRole: true
  }
}

/**
 * 信使等级配置 (PRD中的四级信使体系)
 * Courier level configuration (Four-level courier system per PRD)
 */
export const COURIER_LEVEL_CONFIGS: Record<CourierLevel, CourierLevelConfig> = {
  // 一级信使 (基础投递信使)
  1: {
    level: 1,
    name: '一级信使',
    englishName: 'Level 1 Courier',
    description: '基础投递信使，负责宿舍楼栋、商店路径等具体投递任务',
    managementArea: '楼栋/商店',
    color: {
      bg: 'bg-amber-600',
      text: 'text-white',
      badge: 'bg-amber-100 text-amber-800',
      hover: 'hover:bg-amber-700'
    },
    icon: '🏠',
    iconComponent: Home,
    permissions: [
      'COURIER_SCAN_CODE',
      'COURIER_DELIVER_LETTER',
      'COURIER_VIEW_TASKS',
      'COURIER_MANAGE_PROFILE',
      'COURIER_VIEW_STATISTICS',
      'COURIER_VIEW_POINTS',
      'COURIER_MANAGE_SCHEDULE',
      'COURIER_VIEW_DELIVERY_AREA',
      'COURIER_REPORT_ISSUES'
    ],
    managementPath: '/courier',
    canManageSublevels: false
  },

  // 二级信使 (片区协调员)
  2: {
    level: 2,
    name: '二级信使',
    englishName: 'Level 2 Courier',
    description: '片区协调员，管理宿舍区/楼栋组/商业片区',
    managementArea: '片区',
    color: {
      bg: 'bg-green-600',
      text: 'text-white',
      badge: 'bg-green-100 text-green-800',
      hover: 'hover:bg-green-700'
    },
    icon: '📍',
    iconComponent: MapPin,
    permissions: [
      'COURIER_SCAN_CODE',
      'COURIER_DELIVER_LETTER',
      'COURIER_VIEW_TASKS',
      'COURIER_MANAGE_PROFILE',
      'COURIER_VIEW_STATISTICS',
      'COURIER_VIEW_POINTS',
      'COURIER_EXCHANGE_REWARDS',
      'COURIER_VIEW_LEADERBOARD',
      'COURIER_MANAGE_SCHEDULE',
      'COURIER_VIEW_DELIVERY_AREA',
      'COURIER_REPORT_ISSUES',
      'COURIER_VIEW_FEEDBACK'
    ],
    managementPath: '/courier/zone-manage',
    canManageSublevels: true
  },

  // 三级信使 (校区负责人)
  3: {
    level: 3,
    name: '三级信使',
    englishName: 'Level 3 Courier',
    description: '校区负责人，管理所在学校的信使网络',
    managementArea: '学校',
    color: {
      bg: 'bg-blue-600',
      text: 'text-white',
      badge: 'bg-blue-100 text-blue-800',
      hover: 'hover:bg-blue-700'
    },
    icon: '🏫',
    iconComponent: School,
    permissions: [
      'COURIER_SCAN_CODE',
      'COURIER_DELIVER_LETTER',
      'COURIER_VIEW_TASKS',
      'COURIER_MANAGE_PROFILE',
      'COURIER_VIEW_STATISTICS',
      'COURIER_VIEW_POINTS',
      'COURIER_EXCHANGE_REWARDS',
      'COURIER_VIEW_LEADERBOARD',
      'COURIER_MANAGE_SCHEDULE',
      'COURIER_VIEW_DELIVERY_AREA',
      'COURIER_REPORT_ISSUES',
      'COURIER_VIEW_FEEDBACK',
      'COURIER_PARTICIPATE_ACTIVITIES',
      'MANAGE_COURIERS'
    ],
    managementPath: '/courier/school-manage',
    canManageSublevels: true
  },

  // 四级信使 (城市负责人)
  4: {
    level: 4,
    name: '四级信使',
    englishName: 'Level 4 Courier',
    description: '城市负责人，管理所在城市所有学校的信使网络',
    managementArea: '城市',
    color: {
      bg: 'bg-purple-600',
      text: 'text-white',
      badge: 'bg-purple-100 text-purple-800',
      hover: 'hover:bg-purple-700'
    },
    icon: '👑',
    iconComponent: Crown,
    permissions: [
      'COURIER_SCAN_CODE',
      'COURIER_DELIVER_LETTER',
      'COURIER_VIEW_TASKS',
      'COURIER_MANAGE_PROFILE',
      'COURIER_VIEW_STATISTICS',
      'COURIER_VIEW_POINTS',
      'COURIER_EXCHANGE_REWARDS',
      'COURIER_VIEW_LEADERBOARD',
      'COURIER_MANAGE_SCHEDULE',
      'COURIER_VIEW_DELIVERY_AREA',
      'COURIER_REPORT_ISSUES',
      'COURIER_VIEW_FEEDBACK',
      'COURIER_PARTICIPATE_ACTIVITIES',
      'MANAGE_COURIERS',
      'VIEW_ANALYTICS'
    ],
    managementPath: '/courier/city-manage',
    canManageSublevels: true
  }
}

// ================================
// 工具函数 - Utility Functions
// ================================

/**
 * 获取角色配置
 * Get role configuration
 */
export function getRoleConfig(role: UserRole): RoleConfig {
  return ROLE_CONFIGS[role]
}

/**
 * 获取信使等级配置
 * Get courier level configuration
 */
export function getCourierLevelConfig(level: CourierLevel): CourierLevelConfig {
  return COURIER_LEVEL_CONFIGS[level]
}

/**
 * 获取角色显示名称
 * Get role display name
 */
export function getRoleDisplayName(role: UserRole): string {
  return ROLE_CONFIGS[role]?.name || role
}

/**
 * 获取角色英文名称
 * Get role English name
 */
export function getRoleEnglishName(role: UserRole): string {
  return ROLE_CONFIGS[role]?.englishName || role
}

/**
 * 获取角色颜色配置
 * Get role color configuration
 */
export function getRoleColors(role: UserRole) {
  return ROLE_CONFIGS[role]?.color || ROLE_CONFIGS.user.color
}

/**
 * 获取角色图标
 * Get role icon
 */
export function getRoleIcon(role: UserRole): string {
  return ROLE_CONFIGS[role]?.icon || '👤'
}

/**
 * 获取角色权限列表
 * Get role permissions
 */
export function getRolePermissions(role: UserRole): Permission[] {
  return ROLE_CONFIGS[role]?.permissions || []
}

/**
 * 检查角色是否拥有特定权限
 * Check if role has specific permission
 */
export function hasPermission(role: UserRole, permission: Permission): boolean {
  return getRolePermissions(role).includes(permission)
}

/**
 * 获取角色的默认首页
 * Get role's default homepage
 */
export function getRoleDefaultHomePage(role: UserRole): string {
  return ROLE_CONFIGS[role]?.defaultHomePage || '/write'
}

/**
 * 检查角色是否可以访问管理后台
 * Check if role can access admin panel
 */
export function canAccessAdmin(role: UserRole): boolean {
  return ROLE_CONFIGS[role]?.canAccessAdmin || false
}

/**
 * 检查是否为系统角色
 * Check if it's a system role
 */
export function isSystemRole(role: UserRole): boolean {
  return ROLE_CONFIGS[role]?.isSystemRole || false
}

/**
 * 根据权限层级排序角色
 * Sort roles by hierarchy level
 */
export function sortRolesByHierarchy(roles: UserRole[]): UserRole[] {
  return roles.sort((a, b) => ROLE_CONFIGS[b].hierarchy - ROLE_CONFIGS[a].hierarchy)
}

/**
 * 获取比当前角色权限低的所有角色
 * Get all roles with lower hierarchy than current role
 */
export function getLowerHierarchyRoles(role: UserRole): UserRole[] {
  const currentHierarchy = ROLE_CONFIGS[role].hierarchy
  return Object.keys(ROLE_CONFIGS)
    .filter(r => ROLE_CONFIGS[r as UserRole].hierarchy < currentHierarchy) as UserRole[]
}

/**
 * 获取信使等级名称
 * Get courier level name
 */
export function getCourierLevelName(level: CourierLevel): string {
  return COURIER_LEVEL_CONFIGS[level]?.name || `${level}级信使`
}

/**
 * 获取信使等级管理路径
 * Get courier level management path
 */
export function getCourierLevelManagementPath(level: CourierLevel): string {
  return COURIER_LEVEL_CONFIGS[level]?.managementPath || '/courier'
}

/**
 * 检查信使等级是否可以管理下级
 * Check if courier level can manage sublevels
 */
export function canManageSublevels(level: CourierLevel): boolean {
  return COURIER_LEVEL_CONFIGS[level]?.canManageSublevels || false
}

/**
 * 检查是否为信使角色
 * Check if it's a courier role
 */
export function isCourierRole(role: UserRole): boolean {
  return role.startsWith('courier_level')
}

/**
 * 从角色获取信使等级
 * Get courier level from role
 */
export function getCourierLevelFromRole(role: UserRole): CourierLevel | null {
  if (!isCourierRole(role)) return null
  const level = parseInt(role.split('courier_level')[1])
  return (level >= 1 && level <= 4) ? level as CourierLevel : null
}

/**
 * 获取所有角色选项 (用于下拉选择等)
 * Get all role options (for dropdowns, etc.)
 */
export function getAllRoleOptions() {
  return Object.values(ROLE_CONFIGS).map(config => ({
    value: config.id,
    label: config.name,
    description: config.description,
    hierarchy: config.hierarchy
  }))
}

/**
 * 获取所有信使等级选项
 * Get all courier level options
 */
export function getAllCourierLevelOptions() {
  return Object.values(COURIER_LEVEL_CONFIGS).map(config => ({
    value: config.level,
    label: config.name,
    description: config.description,
    managementArea: config.managementArea
  }))
}

// ================================
// 导出所有配置 - Export All Configurations
// ================================

export default {
  ROLE_CONFIGS,
  COURIER_LEVEL_CONFIGS,
  getRoleConfig,
  getCourierLevelConfig,
  getRoleDisplayName,
  getRoleEnglishName,
  getRoleColors,
  getRoleIcon,
  getRolePermissions,
  hasPermission,
  getRoleDefaultHomePage,
  canAccessAdmin,
  isSystemRole,
  sortRolesByHierarchy,
  getLowerHierarchyRoles,
  getCourierLevelName,
  getCourierLevelManagementPath,
  canManageSublevels,
  isCourierRole,
  getCourierLevelFromRole,
  getAllRoleOptions,
  getAllCourierLevelOptions
}