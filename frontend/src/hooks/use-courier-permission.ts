import { useCourier, usePermissions, useUser, type CourierInfo as APICourierInfo } from '@/stores/user-store'
import { CourierService } from '@/lib/services/courier-service'
import { useEffect, useState } from 'react'
import { 
  getCourierLevelName as getUnifiedCourierLevelName, 
  getCourierLevelManagementPath, 
  canManageSublevels,
  hasPermission as roleHasPermission,
  getRoleDisplayName,
  canAccessAdmin,
  type UserRole,
  type CourierLevel 
} from '@/constants/roles'
import { permissionService } from '@/lib/permissions/permission-service'

export const COURIER_LEVELS = {
  LEVEL_1: 1, // 一级信使（楼栋/班级）
  LEVEL_2: 2, // 二级信使（片区/年级）
  LEVEL_3: 3, // 三级信使（校级）
  LEVEL_4: 4, // 四级信使（城市总代）
} as const

const COURIER_LEVEL_NAMES: Record<number, string> = {
  1: getUnifiedCourierLevelName(1),
  2: getUnifiedCourierLevelName(2),
  3: getUnifiedCourierLevelName(3),
  4: getUnifiedCourierLevelName(4),
}

export const COURIER_PERMISSIONS = {
  SCAN_CODE: 'courier_scan_code',
  DELIVER_LETTER: 'courier_deliver_letter',
  VIEW_OWN_TASKS: 'courier_view_own_tasks',
  REPORT_EXCEPTION: 'courier_report_exception',
  
  MANAGE_SUBORDINATES: 'courier_manage_subordinates',
  ASSIGN_TASKS: 'courier_assign_tasks',
  VIEW_SUBORDINATE_REPORTS: 'courier_view_subordinate_reports',
  CREATE_LOWER_LEVEL_COURIER: 'courier_create_subordinate',
  
  MANAGE_SCHOOL_ZONE: 'courier_manage_school_zone',
  VIEW_SCHOOL_ANALYTICS: 'courier_view_school_analytics',
  COORDINATE_CROSS_ZONE: 'courier_coordinate_cross_zone',
  
  MANAGE_CITY_OPERATIONS: 'courier_manage_city_operations',
  CREATE_SCHOOL_LEVEL_COURIER: 'courier_create_school_courier',
  VIEW_CITY_ANALYTICS: 'courier_view_city_analytics',
} as const

const COURIER_LEVEL_PERMISSIONS: Record<number, string[]> = {
  1: [ // 一级信使
    COURIER_PERMISSIONS.SCAN_CODE,
    COURIER_PERMISSIONS.DELIVER_LETTER,
    COURIER_PERMISSIONS.VIEW_OWN_TASKS,
    COURIER_PERMISSIONS.REPORT_EXCEPTION,
  ],
  2: [ // 二级信使
    COURIER_PERMISSIONS.SCAN_CODE,
    COURIER_PERMISSIONS.DELIVER_LETTER,
    COURIER_PERMISSIONS.VIEW_OWN_TASKS,
    COURIER_PERMISSIONS.REPORT_EXCEPTION,
    COURIER_PERMISSIONS.MANAGE_SUBORDINATES,
    COURIER_PERMISSIONS.ASSIGN_TASKS,
    COURIER_PERMISSIONS.VIEW_SUBORDINATE_REPORTS,
    COURIER_PERMISSIONS.CREATE_LOWER_LEVEL_COURIER,
  ],
  3: [ // 三级信使
    COURIER_PERMISSIONS.SCAN_CODE,
    COURIER_PERMISSIONS.DELIVER_LETTER,
    COURIER_PERMISSIONS.VIEW_OWN_TASKS,
    COURIER_PERMISSIONS.REPORT_EXCEPTION,
    COURIER_PERMISSIONS.MANAGE_SUBORDINATES,
    COURIER_PERMISSIONS.ASSIGN_TASKS,
    COURIER_PERMISSIONS.VIEW_SUBORDINATE_REPORTS,
    COURIER_PERMISSIONS.CREATE_LOWER_LEVEL_COURIER,
    COURIER_PERMISSIONS.MANAGE_SCHOOL_ZONE,
    COURIER_PERMISSIONS.VIEW_SCHOOL_ANALYTICS,
    COURIER_PERMISSIONS.COORDINATE_CROSS_ZONE,
  ],
  4: [ // 四级信使
    COURIER_PERMISSIONS.SCAN_CODE,
    COURIER_PERMISSIONS.DELIVER_LETTER,
    COURIER_PERMISSIONS.VIEW_OWN_TASKS,
    COURIER_PERMISSIONS.REPORT_EXCEPTION,
    COURIER_PERMISSIONS.MANAGE_SUBORDINATES,
    COURIER_PERMISSIONS.ASSIGN_TASKS,
    COURIER_PERMISSIONS.VIEW_SUBORDINATE_REPORTS,
    COURIER_PERMISSIONS.CREATE_LOWER_LEVEL_COURIER,
    COURIER_PERMISSIONS.MANAGE_SCHOOL_ZONE,
    COURIER_PERMISSIONS.VIEW_SCHOOL_ANALYTICS,
    COURIER_PERMISSIONS.COORDINATE_CROSS_ZONE,
    COURIER_PERMISSIONS.MANAGE_CITY_OPERATIONS,
    COURIER_PERMISSIONS.CREATE_SCHOOL_LEVEL_COURIER,
    COURIER_PERMISSIONS.VIEW_CITY_ANALYTICS,
  ],
}

export interface CourierInfo {
  id: string
  userId: string
  level: number
  parentId?: string
  zoneCode: string
  zoneType: 'city' | 'school' | 'zone' | 'building'
  status: 'active' | 'pending' | 'frozen'
  points: number
  taskCount: number
  school_code: string
  username: string
  school_name: string
}

export function useCourierPermission() {
  const { user } = useUser()
  const { courierInfo, updateCourierInfo } = useCourier()
  const { isCourier } = usePermissions()
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const loadCourierInfo = async () => {
      if (!user || !isCourier) {
        return
      }
      
      if (courierInfo) {
        return
      }
      
      if (user.role === 'super_admin') {
        updateCourierInfo({
          level: 4,
          zoneCode: 'ADMIN_ALL',
          zoneType: 'city',
          status: 'active',
          points: 9999,
          taskCount: 0,
          completedTasks: 0,
          averageRating: 5.0,
          lastActiveAt: new Date().toISOString()
        })
        return
      }

      if (user.courierInfo) {
        updateCourierInfo(user.courierInfo)
        return
      }

      setLoading(true)
      try {
        const response = await CourierService.getCourierInfo()
        const responseData = response.data
        
        // 处理新的响应格式
        if (responseData && 'courier_info' in responseData && responseData.courier_info) {
          const apiCourierInfo = responseData.courier_info as any
          updateCourierInfo({
            level: (apiCourierInfo.level || 1) as 1 | 2 | 3 | 4,
            zoneCode: apiCourierInfo.region || apiCourierInfo.zone || 'DEFAULT',
            zoneType: getZoneTypeFromLevel(apiCourierInfo.level || 1),
            status: 'active',
            points: apiCourierInfo.total_points || apiCourierInfo.TotalPoints || 0,
            taskCount: apiCourierInfo.completed_tasks || apiCourierInfo.CompletedTasks || 0,
            completedTasks: apiCourierInfo.completed_tasks || apiCourierInfo.CompletedTasks || 0,
            averageRating: apiCourierInfo.rating || 4.0,
            lastActiveAt: new Date().toISOString()
          })
        } else if (responseData && 'is_courier' in responseData && responseData.is_courier === false) {
          // 用户不是信使但可能有管理权限
          console.log('User is not a courier but has role:', responseData && 'user_role' in responseData ? responseData.user_role : 'unknown')
        } else {
          // 旧格式兼容
          const typedResponseData = responseData as any
          updateCourierInfo({
            level: (typedResponseData?.level || 1) as 1 | 2 | 3 | 4,
            zoneCode: typedResponseData?.region || 'DEFAULT',
            zoneType: getZoneTypeFromLevel(typedResponseData?.level || 1),
            status: 'active',
            points: typedResponseData?.total_points || 0,
            taskCount: typedResponseData?.completed_tasks || 0,
            completedTasks: typedResponseData?.completed_tasks || 0,
            averageRating: typedResponseData?.rating || 4.0,
            lastActiveAt: new Date().toISOString()
          })
        }
      } catch (error) {
        console.error('Failed to load courier info:', error)
        // 如果API失败，根据角色名称设置默认级别
        if (user?.role.includes('courier')) {
          // 从角色名称中提取级别
          let defaultLevel = 1
          if (user.role === 'courier_level4') {
            defaultLevel = 4
          } else if (user.role === 'courier_level3') {
            defaultLevel = 3
          } else if (user.role === 'courier_level2') {
            defaultLevel = 2
          } else if (user.role === 'courier_level1') {
            defaultLevel = 1
          }
          
          updateCourierInfo({
            level: defaultLevel as 1 | 2 | 3 | 4,
            zoneCode: defaultLevel === 4 ? 'BEIJING' : 'DEFAULT',
            zoneType: getZoneTypeFromLevel(defaultLevel),
            status: 'active',
            points: 0,
            taskCount: 0,
            completedTasks: 0,
            averageRating: 0,
            lastActiveAt: new Date().toISOString()
          })
        }
      } finally {
        setLoading(false)
      }
    }

    loadCourierInfo()
  }, [user, isCourier, courierInfo, updateCourierInfo])

  // 根据级别确定区域类型
  const getZoneTypeFromLevel = (level: number): 'city' | 'school' | 'zone' | 'building' => {
    switch (level) {
      case 4:
        return 'city'
      case 3:
        return 'school'
      case 2:
        return 'zone'
      case 1:
      default:
        return 'building'
    }
  }

  // 基于SOTA动态权限系统的权限检查
  const hasCourierPermission = (permission: string): boolean => {
    if (!user) return false
    return permissionService.hasPermission(user, permission)
  }

  // 基于统一角色系统的级别检查
  const isCourierLevel = (requiredLevel: number): boolean => {
    if (!user || !courierInfo) return false
    return courierInfo.level >= requiredLevel
  }

  // 基于SOTA动态权限系统的管理权限检查
  const canManageSubordinates = (): boolean => {
    if (!user) return false
    const hasPermission = permissionService.hasAnyPermission(user, ['MANAGE_SUBORDINATES', 'MANAGE_COURIERS'])
    console.log('🔍 canManageSubordinates check:', {
      user: user.username,
      role: user.role,
      courierInfo,
      hasPermission,
      userPermissions: permissionService.getUserPermissions(user)
    })
    return hasPermission
  }

  const canCreateSubordinate = (): boolean => {
    if (!user || !courierInfo) return false
    const hasPermission = courierInfo.level > 1 && permissionService.hasAnyPermission(user, ['CREATE_SUBORDINATE', 'MANAGE_COURIERS'])
    console.log('🔍 canCreateSubordinate check:', {
      user: user.username,
      role: user.role,
      courierLevel: courierInfo.level,
      levelCheck: courierInfo.level > 1,
      hasPermission,
      userPermissions: permissionService.getUserPermissions(user)
    })
    return hasPermission
  }

  const canAssignTasks = (): boolean => {
    if (!user) return false
    const hasPermission = permissionService.hasAnyPermission(user, ['ASSIGN_TASKS', 'MANAGE_COURIERS'])
    console.log('🔍 canAssignTasks check:', {
      user: user.username,
      role: user.role,
      courierInfo,
      hasPermission,
      userPermissions: permissionService.getUserPermissions(user)
    })
    return hasPermission
  }

  // 基于统一角色系统的显示名称
  const getCourierLevelName = (): string => {
    if (!user) return ''
    if (courierInfo?.level) {
      return getUnifiedCourierLevelName(courierInfo.level as CourierLevel)
    }
    return getRoleDisplayName(user.role as UserRole)
  }

  // 基于SOTA动态权限系统的管理级别
  const getManageableLevels = (): number[] => {
    if (!user || !courierInfo) return []
    
    // 管理员可以管理所有级别
    if (permissionService.canAccessAdmin(user)) {
      return [4, 3, 2, 1]
    }
    
    // 信使只能管理比自己低的级别
    const maxLevel = courierInfo.level - 1
    return maxLevel > 0 ? Array.from({length: maxLevel}, (_, i) => maxLevel - i) : []
  }

  // 基于统一角色系统的管理路径
  const getManagementDashboardPath = (): string => {
    if (!user) return '/courier'
    
    if (courierInfo?.level) {
      return getCourierLevelManagementPath(courierInfo.level as CourierLevel)
    }
    
    // 其他角色使用默认路径
    return '/courier'
  }

  // 基于SOTA动态权限系统的管理后台显示
  const showManagementDashboard = (): boolean => {
    if (!user) return false
    return permissionService.canAccessAdmin(user) || Boolean(courierInfo?.level && courierInfo.level > 1)
  }

  // Convert courier info from store format to hook format for compatibility
  const legacyCourierInfo: CourierInfo | null = courierInfo ? {
    id: `courier_${user?.id}`,
    userId: user?.id || '',
    level: courierInfo.level,
    zoneCode: courierInfo.zoneCode,
    zoneType: courierInfo.zoneType,
    status: courierInfo.status,
    points: courierInfo.points,
    taskCount: courierInfo.taskCount,
    school_code: courierInfo.school_code,
    username: courierInfo.username,
    school_name: courierInfo.school_name,
  } : null

  return {
    courierInfo: legacyCourierInfo,
    loading,
    hasCourierPermission,
    isCourierLevel,
    canManageSubordinates,
    canCreateSubordinate,
    canAssignTasks,
    getCourierLevelName,
    getManageableLevels,
    getManagementDashboardPath,
    showManagementDashboard,
    COURIER_LEVELS,
    COURIER_PERMISSIONS,
    // 刷新信使信息的方法 (使用新的store系统)
    refreshCourierInfo: async () => {
      if (!user || !isCourier) return

      // 如果是super_admin，直接设置最高权限
      if (user.role === 'super_admin') {
        updateCourierInfo({
          level: 4,
          zoneCode: 'ADMIN_ALL',
          zoneType: 'city',
          status: 'active',
          points: 9999,
          taskCount: 0,
          completedTasks: 0,
          averageRating: 5.0,
          lastActiveAt: new Date().toISOString()
        })
        return
      }

      setLoading(true)
      try {
        const response = await CourierService.getCourierInfo()
        const responseData = response.data
        
        // 处理新的响应格式
        if (responseData && 'courier_info' in responseData && responseData.courier_info) {
          const apiCourierInfo = responseData.courier_info as any
          updateCourierInfo({
            level: (apiCourierInfo.level || 1) as 1 | 2 | 3 | 4,
            zoneCode: apiCourierInfo.region || apiCourierInfo.zone || 'DEFAULT',
            zoneType: getZoneTypeFromLevel(apiCourierInfo.level || 1),
            status: 'active',
            points: apiCourierInfo.total_points || apiCourierInfo.TotalPoints || 0,
            taskCount: apiCourierInfo.completed_tasks || apiCourierInfo.CompletedTasks || 0,
            completedTasks: apiCourierInfo.completed_tasks || apiCourierInfo.CompletedTasks || 0,
            averageRating: apiCourierInfo.rating || 4.0,
            lastActiveAt: new Date().toISOString()
          })
        } else {
          // 旧格式兼容
          const typedResponseData = responseData as any
          updateCourierInfo({
            level: (typedResponseData?.level || 1) as 1 | 2 | 3 | 4,
            zoneCode: typedResponseData?.region || 'DEFAULT',
            zoneType: getZoneTypeFromLevel(typedResponseData?.level || 1),
            status: 'active',
            points: typedResponseData?.total_points || 0,
            taskCount: typedResponseData?.completed_tasks || 0,
            completedTasks: typedResponseData?.completed_tasks || 0,
            averageRating: typedResponseData?.rating || 4.0,
            lastActiveAt: new Date().toISOString()
          })
        }
      } catch (error) {
        console.error('Failed to refresh courier info:', error)
        // 如果API失败，根据角色名称设置默认级别
        if (user?.role.includes('courier')) {
          // 从角色名称中提取级别
          let defaultLevel = 1
          if (user.role === 'courier_level4') {
            defaultLevel = 4
          } else if (user.role === 'courier_level3') {
            defaultLevel = 3
          } else if (user.role === 'courier_level2') {
            defaultLevel = 2
          } else if (user.role === 'courier_level1') {
            defaultLevel = 1
          }
          
          updateCourierInfo({
            level: defaultLevel as 1 | 2 | 3 | 4,
            zoneCode: defaultLevel === 4 ? 'BEIJING' : 'DEFAULT',
            zoneType: getZoneTypeFromLevel(defaultLevel),
            status: 'active',
            points: 0,
            taskCount: 0,
            completedTasks: 0,
            averageRating: 0,
            lastActiveAt: new Date().toISOString()
          })
        }
      } finally {
        setLoading(false)
      }
    }
  }
}