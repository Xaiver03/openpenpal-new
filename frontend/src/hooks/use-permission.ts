import { useAuth } from '@/contexts/auth-context-new'

// 角色层级定义（与后端保持一致）
const ROLE_HIERARCHY: Record<string, number> = {
  'user': 1,
  'courier': 2,
  'courier_level1': 2,
  'courier_level2': 3,
  'senior_courier': 3,
  'courier_level3': 4,
  'courier_coordinator': 4,
  'courier_level4': 5,
  'school_admin': 5,
  'platform_admin': 6,
  'super_admin': 7,
}

// 权限定义
export const PERMISSIONS = {
  // 用户权限
  WRITE_LETTER: 'write_letter',
  READ_LETTER: 'read_letter',
  MANAGE_PROFILE: 'manage_profile',
  
  // 信使权限
  DELIVER_LETTER: 'deliver_letter',
  SCAN_CODE: 'scan_code',
  VIEW_TASKS: 'view_tasks',
  
  // 协调员权限
  MANAGE_COURIERS: 'manage_couriers',
  ASSIGN_TASKS: 'assign_tasks',
  VIEW_REPORTS: 'view_reports',
  
  // 管理员权限
  MANAGE_USERS: 'manage_users',
  MANAGE_SCHOOL: 'manage_school',
  VIEW_ANALYTICS: 'view_analytics',
  MANAGE_SYSTEM: 'manage_system',
  
  // 超级管理员权限
  MANAGE_PLATFORM: 'manage_platform',
  MANAGE_ADMINS: 'manage_admins',
  SYSTEM_CONFIG: 'system_config',
} as const

// 角色权限映射
const ROLE_PERMISSIONS: Record<string, string[]> = {
  'user': [
    PERMISSIONS.WRITE_LETTER,
    PERMISSIONS.READ_LETTER,
    PERMISSIONS.MANAGE_PROFILE,
  ],
  'courier': [
    PERMISSIONS.WRITE_LETTER,
    PERMISSIONS.READ_LETTER,
    PERMISSIONS.MANAGE_PROFILE,
    PERMISSIONS.DELIVER_LETTER,
    PERMISSIONS.SCAN_CODE,
    PERMISSIONS.VIEW_TASKS,
  ],
  'courier_level1': [
    PERMISSIONS.WRITE_LETTER,
    PERMISSIONS.READ_LETTER,
    PERMISSIONS.MANAGE_PROFILE,
    PERMISSIONS.DELIVER_LETTER,
    PERMISSIONS.SCAN_CODE,
    PERMISSIONS.VIEW_TASKS,
  ],
  'courier_level2': [
    PERMISSIONS.WRITE_LETTER,
    PERMISSIONS.READ_LETTER,
    PERMISSIONS.MANAGE_PROFILE,
    PERMISSIONS.DELIVER_LETTER,
    PERMISSIONS.SCAN_CODE,
    PERMISSIONS.VIEW_TASKS,
    PERMISSIONS.VIEW_REPORTS,
  ],
  'senior_courier': [
    PERMISSIONS.WRITE_LETTER,
    PERMISSIONS.READ_LETTER,
    PERMISSIONS.MANAGE_PROFILE,
    PERMISSIONS.DELIVER_LETTER,
    PERMISSIONS.SCAN_CODE,
    PERMISSIONS.VIEW_TASKS,
    PERMISSIONS.VIEW_REPORTS,
  ],
  'courier_level3': [
    PERMISSIONS.WRITE_LETTER,
    PERMISSIONS.READ_LETTER,
    PERMISSIONS.MANAGE_PROFILE,
    PERMISSIONS.DELIVER_LETTER,
    PERMISSIONS.SCAN_CODE,
    PERMISSIONS.VIEW_TASKS,
    PERMISSIONS.MANAGE_COURIERS,
    PERMISSIONS.ASSIGN_TASKS,
    PERMISSIONS.VIEW_REPORTS,
  ],
  'courier_coordinator': [
    PERMISSIONS.WRITE_LETTER,
    PERMISSIONS.READ_LETTER,
    PERMISSIONS.MANAGE_PROFILE,
    PERMISSIONS.DELIVER_LETTER,
    PERMISSIONS.SCAN_CODE,
    PERMISSIONS.VIEW_TASKS,
    PERMISSIONS.MANAGE_COURIERS,
    PERMISSIONS.ASSIGN_TASKS,
    PERMISSIONS.VIEW_REPORTS,
  ],
  'courier_level4': [
    PERMISSIONS.WRITE_LETTER,
    PERMISSIONS.READ_LETTER,
    PERMISSIONS.MANAGE_PROFILE,
    PERMISSIONS.DELIVER_LETTER,
    PERMISSIONS.SCAN_CODE,
    PERMISSIONS.VIEW_TASKS,
    PERMISSIONS.MANAGE_COURIERS,
    PERMISSIONS.ASSIGN_TASKS,
    PERMISSIONS.VIEW_REPORTS,
    PERMISSIONS.VIEW_ANALYTICS,
  ],
  'school_admin': [
    PERMISSIONS.WRITE_LETTER,
    PERMISSIONS.READ_LETTER,
    PERMISSIONS.MANAGE_PROFILE,
    PERMISSIONS.MANAGE_USERS,
    PERMISSIONS.MANAGE_COURIERS,
    PERMISSIONS.ASSIGN_TASKS,
    PERMISSIONS.VIEW_REPORTS,
    PERMISSIONS.MANAGE_SCHOOL,
    PERMISSIONS.VIEW_ANALYTICS,
  ],
  'platform_admin': [
    PERMISSIONS.WRITE_LETTER,
    PERMISSIONS.READ_LETTER,
    PERMISSIONS.MANAGE_PROFILE,
    PERMISSIONS.MANAGE_USERS,
    PERMISSIONS.MANAGE_COURIERS,
    PERMISSIONS.ASSIGN_TASKS,
    PERMISSIONS.VIEW_REPORTS,
    PERMISSIONS.MANAGE_SCHOOL,
    PERMISSIONS.VIEW_ANALYTICS,
    PERMISSIONS.MANAGE_SYSTEM,
  ],
  'super_admin': [
    PERMISSIONS.WRITE_LETTER,
    PERMISSIONS.READ_LETTER,
    PERMISSIONS.MANAGE_PROFILE,
    PERMISSIONS.MANAGE_USERS,
    PERMISSIONS.MANAGE_COURIERS,
    PERMISSIONS.ASSIGN_TASKS,
    PERMISSIONS.VIEW_REPORTS,
    PERMISSIONS.MANAGE_SCHOOL,
    PERMISSIONS.VIEW_ANALYTICS,
    PERMISSIONS.MANAGE_SYSTEM,
    PERMISSIONS.MANAGE_PLATFORM,
    PERMISSIONS.MANAGE_ADMINS,
    PERMISSIONS.SYSTEM_CONFIG,
  ],
}

export function usePermission() {
  const { user } = useAuth()

  // 检查用户是否有特定权限
  const hasPermission = (permission: string): boolean => {
    if (!user) return false
    
    const userPermissions = ROLE_PERMISSIONS[user.role] || []
    return userPermissions.includes(permission)
  }

  // 检查用户是否有特定角色或更高权限
  const hasRole = (requiredRole: string): boolean => {
    if (!user) {
      console.log('🐛 hasRole: No user found')
      return false
    }
    
    const userLevel = ROLE_HIERARCHY[user.role] || 0
    const requiredLevel = ROLE_HIERARCHY[requiredRole] || 0
    
    console.log(`🐛 hasRole check:`, {
      requiredRole,
      userRole: user.role,
      userLevel,
      requiredLevel,
      result: userLevel >= requiredLevel,
      roleHierarchy: ROLE_HIERARCHY
    })
    
    return userLevel >= requiredLevel
  }

  // 检查是否是信使（任何级别）
  const isCourier = (): boolean => {
    if (!user) return false
    // Check if user has any courier-related role
    const courierRoles = ['courier', 'courier_level1', 'courier_level2', 'courier_level3', 'courier_level4', 'senior_courier', 'courier_coordinator']
    return courierRoles.includes(user.role)
  }

  // 检查是否是管理员（任何级别）
  const isAdmin = (): boolean => {
    return hasRole('school_admin')
  }

  // 检查是否是超级管理员
  const isSuperAdmin = (): boolean => {
    return user?.role === 'super_admin'
  }

  // 获取用户角色显示名称
  const getRoleDisplayName = (): string => {
    const roleNames: Record<string, string> = {
      'user': '普通用户',
      'courier': '信使',
      'courier_level1': '一级信使',
      'courier_level2': '二级信使',
      'courier_level3': '三级信使',
      'courier_level4': '四级信使',
      'senior_courier': '高级信使',
      'courier_coordinator': '信使协调员',
      'school_admin': '学校管理员',
      'platform_admin': '平台管理员',
      'super_admin': '超级管理员',
    }
    
    return user ? (roleNames[user.role] || '未知角色') : ''
  }

  return {
    user,
    hasPermission,
    hasRole,
    isCourier,
    isAdmin,
    isSuperAdmin,
    getRoleDisplayName,
  }
}