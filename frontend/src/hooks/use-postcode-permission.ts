'use client'

import { useState, useEffect, useCallback } from 'react'
import { useCourierPermission } from './use-courier-permission'
import PostcodeService from '@/lib/services/postcode-service'
import type { CourierPostcodePermission } from '@/lib/types/postcode'

/**
 * Postcode 权限管理 Hook
 * 基于现有的信使权限系统，扩展 Postcode 地址编码权限
 */
export function usePostcodePermission() {
  const { courierInfo, hasCourierPermission, COURIER_PERMISSIONS, loading: courierLoading } = useCourierPermission()
  
  const [postcodePermissions, setPostcodePermissions] = useState<CourierPostcodePermission | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // 加载 Postcode 权限
  useEffect(() => {
    const loadPostcodePermissions = async () => {
      if (!courierInfo || courierLoading) return

      try {
        setLoading(true)
        setError(null)

        // 根据信使等级生成默认的 Postcode 权限
        const defaultPermissions: CourierPostcodePermission = {
          courierId: courierInfo.id || 'unknown',
          level: courierInfo.level as 1 | 2 | 3 | 4,
          prefixPatterns: generateDefaultPrefixPatterns(courierInfo.level, courierInfo.zoneCode),
          canManage: courierInfo.level > 1,
          canCreate: courierInfo.level > 1,
          canReview: courierInfo.level > 2
        }

        // 尝试从服务器获取权限，如果失败则使用默认权限
        try {
          const response = await PostcodeService.getCourierPostcodePermissions(courierInfo.id || 'unknown')
          if (response.success && response.data) {
            setPostcodePermissions(response.data)
          } else {
            setPostcodePermissions(defaultPermissions)
          }
        } catch (serverError) {
          console.log('使用默认 Postcode 权限配置')
          setPostcodePermissions(defaultPermissions)
        }

      } catch (err) {
        console.error('Failed to load postcode permissions:', err)
        setError('加载 Postcode 权限失败')
      } finally {
        setLoading(false)
      }
    }

    loadPostcodePermissions()
  }, [courierInfo, courierLoading])

  // 生成默认的前缀权限模式
  const generateDefaultPrefixPatterns = useCallback((level: number, zoneCode: string): string[] => {
    // 基于现有的 zoneCode 生成权限前缀
    // 例如：zoneCode = "BEIJING" 对应学校编码 "PK"
    const schoolCodeMap: Record<string, string> = {
      'BEIJING': 'PK',    // 北京大学
      'TSINGHUA': 'QH',   // 清华大学
      'SYSTEM': 'SY'      // 系统默认
    }

    const schoolCode = schoolCodeMap[zoneCode] || 'SY'
    
    switch (level) {
      case 4: // 四级信使：学校级别权限
        return [schoolCode]
      case 3: // 三级信使：片区级别权限（假设管理多个片区）
        return [`${schoolCode}1`, `${schoolCode}2`, `${schoolCode}3`, `${schoolCode}4`, `${schoolCode}5`]
      case 2: // 二级信使：楼栋级别权限（假设管理某个片区的多个楼栋）
        return [`${schoolCode}5A`, `${schoolCode}5B`, `${schoolCode}5C`, `${schoolCode}5F`]
      case 1: // 一级信使：具体房间权限
        return [`${schoolCode}5F1A`, `${schoolCode}5F2A`, `${schoolCode}5F3D`]
      default:
        return []
    }
  }, [])

  // 检查是否有指定 Postcode 的权限
  const hasPostcodePermission = useCallback((postcode: string): boolean => {
    if (!postcodePermissions) return false
    
    return postcodePermissions.prefixPatterns.some(pattern => 
      postcode.startsWith(pattern)
    )
  }, [postcodePermissions])

  // 检查是否可以管理指定层级的地址结构
  const canManageAddressLevel = useCallback((level: 'school' | 'area' | 'building' | 'room'): boolean => {
    if (!postcodePermissions || !postcodePermissions.canManage) return false

    const requiredLevel = {
      'school': 4,    // 四级信使管理学校
      'area': 3,      // 三级信使管理片区
      'building': 2,  // 二级信使管理楼栋
      'room': 1       // 一级信使管理房间
    }

    return postcodePermissions.level >= requiredLevel[level]
  }, [postcodePermissions])

  // 检查是否可以创建下级地址结构
  const canCreateSubAddressLevel = useCallback((level: 'area' | 'building' | 'room'): boolean => {
    if (!postcodePermissions || !postcodePermissions.canCreate) return false

    const requiredLevel = {
      'area': 4,      // 四级信使创建片区
      'building': 3,  // 三级信使创建楼栋
      'room': 2       // 二级信使创建房间
    }

    return postcodePermissions.level >= requiredLevel[level]
  }, [postcodePermissions])

  // 检查是否可以审核地址反馈
  const canReviewAddressFeedback = useCallback((): boolean => {
    return postcodePermissions?.canReview || false
  }, [postcodePermissions])

  // 获取可管理的 Postcode 前缀列表
  const getManagedPrefixes = useCallback((): string[] => {
    return postcodePermissions?.prefixPatterns || []
  }, [postcodePermissions])

  // 根据信使等级获取管理范围描述
  const getManagementScope = useCallback((): string => {
    if (!postcodePermissions) return '无权限'

    switch (postcodePermissions.level) {
      case 4:
        return '城市级：管理整个学校的地址结构'
      case 3:
        return '学校级：管理指定片区的地址结构'
      case 2:
        return '片区级：管理指定楼栋的地址结构'
      case 1:
        return '楼栋级：负责具体房间的投递'
      default:
        return '未知权限级别'
    }
  }, [postcodePermissions])

  // 检查是否为 Postcode 系统管理员（四级信使且有城市管理权限）
  const isPostcodeAdmin = useCallback((): boolean => {
    return (
      postcodePermissions?.level === 4 &&
      hasCourierPermission('MANAGE_COURIERS')
    )
  }, [postcodePermissions, hasCourierPermission])

  // 获取权限状态摘要
  const getPermissionSummary = useCallback(() => {
    if (!postcodePermissions) {
      return {
        level: 0,
        scope: '无权限',
        canManage: false,
        canCreate: false,
        canReview: false,
        prefixCount: 0
      }
    }

    return {
      level: postcodePermissions.level,
      scope: getManagementScope(),
      canManage: postcodePermissions.canManage,
      canCreate: postcodePermissions.canCreate,
      canReview: postcodePermissions.canReview,
      prefixCount: postcodePermissions.prefixPatterns.length
    }
  }, [postcodePermissions, getManagementScope])

  return {
    // 数据状态
    postcodePermissions,
    loading: loading || courierLoading,
    error,

    // 权限检查方法
    hasPostcodePermission,
    canManageAddressLevel,
    canCreateSubAddressLevel,
    canReviewAddressFeedback,
    isPostcodeAdmin,

    // 数据获取方法
    getManagedPrefixes,
    getManagementScope,
    getPermissionSummary,

    // 继承的信使权限
    courierInfo,
    hasCourierPermission
  }
}

export default usePostcodePermission