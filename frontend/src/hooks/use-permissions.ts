/**
 * 高级权限管理Hook - SOTA权限系统
 * 提供完整的动态权限检查、管理和监控功能
 */

import { useState, useEffect, useCallback, useMemo } from 'react'
import { useUser } from '@/stores/user-store'
import { permissionService } from '@/lib/permissions/permission-service'
import { 
  PERMISSION_MODULES, 
  PermissionModule, 
  PermissionCategory, 
  RiskLevel,
  getPermissionsByCategory,
  getPermissionsByRiskLevel 
} from '@/lib/permissions/permission-modules'

export interface PermissionStatus {
  id: string
  name: string
  description: string
  category: PermissionCategory
  riskLevel: RiskLevel
  granted: boolean
  source: 'role' | 'courier_level' | 'custom'
  dependencies?: string[]
  conflicts?: string[]
}

export interface PermissionSummary {
  total: number
  granted: number
  denied: number
  byCategory: Record<PermissionCategory, { granted: number; total: number }>
  byRiskLevel: Record<RiskLevel, { granted: number; total: number }>
  highRiskPermissions: string[]
  missingDependencies: string[]
}

/**
 * 高级权限管理Hook
 */
export function usePermissions() {
  const { user, isAuthenticated } = useUser()
  const [loading, setLoading] = useState(false)
  const [lastRefresh, setLastRefresh] = useState<Date | null>(null)

  // ================================
  // 基础权限检查
  // ================================

  const hasPermission = useCallback((permission: string): boolean => {
    if (!user) return false
    return permissionService.hasPermission(user, permission)
  }, [user])

  const hasAnyPermission = useCallback((permissions: string[]): boolean => {
    if (!user) return false
    return permissionService.hasAnyPermission(user, permissions)
  }, [user])

  const hasAllPermissions = useCallback((permissions: string[]): boolean => {
    if (!user) return false
    return permissionService.hasAllPermissions(user, permissions)
  }, [user])

  // ================================
  // 权限详情和状态
  // ================================

  const userPermissions = useMemo(() => {
    if (!user) return []
    return permissionService.getUserPermissions(user)
  }, [user])

  const permissionStatuses = useMemo((): PermissionStatus[] => {
    if (!user) return []

    return Object.values(PERMISSION_MODULES).map(module => {
      const granted = userPermissions.includes(module.id)
      
      // 确定权限来源
      let source: 'role' | 'courier_level' | 'custom' = 'role'
      const rolePermissions = permissionService.getRolePermissions(user.role)
      if (user.courierInfo?.level) {
        const courierPermissions = permissionService.getCourierLevelPermissions(user.courierInfo.level)
        if (courierPermissions.includes(module.id) && !rolePermissions.includes(module.id)) {
          source = 'courier_level'
        }
      }

      return {
        id: module.id,
        name: module.name,
        description: module.description,
        category: module.category,
        riskLevel: module.riskLevel,
        granted,
        source,
        dependencies: module.dependencies,
        conflicts: module.conflicts
      }
    })
  }, [user, userPermissions])

  const permissionSummary = useMemo((): PermissionSummary => {
    const total = permissionStatuses.length
    const granted = permissionStatuses.filter(p => p.granted).length
    const denied = total - granted

    // 按分类统计
    const byCategory = {} as Record<PermissionCategory, { granted: number; total: number }>
    const categories: PermissionCategory[] = ['basic', 'courier', 'management', 'admin', 'system']
    
    categories.forEach(category => {
      const categoryPermissions = permissionStatuses.filter(p => p.category === category)
      byCategory[category] = {
        granted: categoryPermissions.filter(p => p.granted).length,
        total: categoryPermissions.length
      }
    })

    // 按风险级别统计
    const byRiskLevel = {} as Record<RiskLevel, { granted: number; total: number }>
    const riskLevels: RiskLevel[] = ['low', 'medium', 'high', 'critical']
    
    riskLevels.forEach(risk => {
      const riskPermissions = permissionStatuses.filter(p => p.riskLevel === risk)
      byRiskLevel[risk] = {
        granted: riskPermissions.filter(p => p.granted).length,
        total: riskPermissions.length
      }
    })

    // 高风险权限
    const highRiskPermissions = permissionStatuses
      .filter(p => p.granted && (p.riskLevel === 'high' || p.riskLevel === 'critical'))
      .map(p => p.id)

    // 缺失的依赖
    const missingDependencies: string[] = []
    permissionStatuses.forEach(permission => {
      if (permission.granted && permission.dependencies) {
        permission.dependencies.forEach(dep => {
          if (!userPermissions.includes(dep)) {
            missingDependencies.push(dep)
          }
        })
      }
    })

    return {
      total,
      granted,
      denied,
      byCategory,
      byRiskLevel,
      highRiskPermissions,
      missingDependencies: [...new Set(missingDependencies)]
    }
  }, [permissionStatuses, userPermissions])

  // ================================
  // 权限分类查询
  // ================================

  const getPermissionsByCategory = useCallback((category: PermissionCategory) => {
    return permissionStatuses.filter(p => p.category === category)
  }, [permissionStatuses])

  const getPermissionsByRiskLevel = useCallback((riskLevel: RiskLevel) => {
    return permissionStatuses.filter(p => p.riskLevel === riskLevel)
  }, [permissionStatuses])

  const getGrantedPermissions = useCallback(() => {
    return permissionStatuses.filter(p => p.granted)
  }, [permissionStatuses])

  const getDeniedPermissions = useCallback(() => {
    return permissionStatuses.filter(p => !p.granted)
  }, [permissionStatuses])

  // ================================
  // 权限管理功能
  // ================================

  const refreshPermissions = useCallback(async () => {
    if (!isAuthenticated) return

    setLoading(true)
    try {
      await permissionService.refreshPermissions()
      setLastRefresh(new Date())
    } catch (error) {
      console.error('Failed to refresh permissions:', error)
    } finally {
      setLoading(false)
    }
  }, [isAuthenticated])

  // 自动刷新权限（可选）
  useEffect(() => {
    if (isAuthenticated && !lastRefresh) {
      refreshPermissions()
    }
  }, [isAuthenticated, lastRefresh, refreshPermissions])

  // ================================
  // 便利方法
  // ================================

  const canAccessFeature = useCallback((feature: {
    requiredPermissions?: string[]
    anyOfPermissions?: string[]
    forbiddenPermissions?: string[]
  }): boolean => {
    if (!user) return false

    // 检查禁止权限
    if (feature.forbiddenPermissions?.some(p => hasPermission(p))) {
      return false
    }

    // 检查必需权限
    if (feature.requiredPermissions && !hasAllPermissions(feature.requiredPermissions)) {
      return false
    }

    // 检查任一权限
    if (feature.anyOfPermissions && !hasAnyPermission(feature.anyOfPermissions)) {
      return false
    }

    return true
  }, [user, hasPermission, hasAllPermissions, hasAnyPermission])

  const isAdmin = useCallback(() => {
    if (!user) return false
    return permissionService.canAccessAdmin(user)
  }, [user])

  const isCourier = useCallback(() => {
    if (!user) return false
    return permissionService.isCourier(user)
  }, [user])

  const hasHighRiskPermissions = useCallback(() => {
    return permissionSummary.highRiskPermissions.length > 0
  }, [permissionSummary])

  const hasMissingDependencies = useCallback(() => {
    return permissionSummary.missingDependencies.length > 0
  }, [permissionSummary])

  // ================================
  // 权限配置管理（管理员专用）
  // ================================

  const canManagePermissions = useCallback(() => {
    return hasAnyPermission(['MANAGE_SYSTEM_SETTINGS', 'MANAGE_COURIERS'])
  }, [hasAnyPermission])

  const exportPermissionConfig = useCallback(() => {
    if (!canManagePermissions()) {
      throw new Error('Insufficient permissions to export configuration')
    }
    return permissionService.exportConfigs()
  }, [canManagePermissions])

  const importPermissionConfig = useCallback(async (configData: string, overwrite = false) => {
    if (!canManagePermissions()) {
      throw new Error('Insufficient permissions to import configuration')
    }
    await permissionService.importConfigs(configData, overwrite)
    await refreshPermissions()
  }, [canManagePermissions, refreshPermissions])

  // ================================
  // 返回接口
  // ================================

  return {
    // 基础权限检查
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    canAccessFeature,

    // 权限状态
    userPermissions,
    permissionStatuses,
    permissionSummary,
    loading,
    lastRefresh,

    // 权限分类查询
    getPermissionsByCategory,
    getPermissionsByRiskLevel,
    getGrantedPermissions,
    getDeniedPermissions,

    // 便利方法
    isAdmin,
    isCourier,
    hasHighRiskPermissions,
    hasMissingDependencies,

    // 管理功能
    refreshPermissions,
    canManagePermissions,
    exportPermissionConfig,
    importPermissionConfig,

    // 权限模块信息
    getAllPermissionModules: () => PERMISSION_MODULES,
    getPermissionModule: (id: string) => PERMISSION_MODULES[id]
  }
}

/**
 * 简化版权限Hook，仅提供基础权限检查
 */
export function useBasicPermissions() {
  const { user } = useUser()

  return {
    hasPermission: useCallback((permission: string) => {
      if (!user) return false
      return permissionService.hasPermission(user, permission)
    }, [user]),

    hasAnyPermission: useCallback((permissions: string[]) => {
      if (!user) return false
      return permissionService.hasAnyPermission(user, permissions)
    }, [user]),

    hasAllPermissions: useCallback((permissions: string[]) => {
      if (!user) return false
      return permissionService.hasAllPermissions(user, permissions)
    }, [user]),

    isAdmin: useCallback(() => {
      if (!user) return false
      return permissionService.canAccessAdmin(user)
    }, [user]),

    isCourier: useCallback(() => {
      if (!user) return false
      return permissionService.isCourier(user)
    }, [user])
  }
}

export default usePermissions