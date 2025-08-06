/**
 * 权限管理API - 支持动态权限配置的RESTful接口
 */

import { NextRequest, NextResponse } from 'next/server'
import { permissionService } from '@/lib/permissions/permission-service'
import { PERMISSION_MODULES } from '@/lib/permissions/permission-modules'
import type { UserRole, CourierLevel } from '@/constants/roles'
import { broadcastPermissionChange } from '@/lib/permissions/permission-notifications'
import { dispatchPermissionChangeEvent } from '@/lib/permissions/permission-enforcer'

// ================================
// 获取权限配置信息
// ================================

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const type = searchParams.get('type') // 'overview' | 'roles' | 'courier-levels' | 'modules'
    const target = searchParams.get('target') // 具体的角色或等级

    switch (type) {
      case 'overview':
        return NextResponse.json({
          success: true,
          data: {
            totalModules: Object.keys(PERMISSION_MODULES).length,
            totalRoles: 8,
            totalCourierLevels: 4,
            modulesByCategory: getModulesByCategory(),
            customConfigs: permissionService.getAllCustomConfigs()
          }
        })

      case 'modules':
        return NextResponse.json({
          success: true,
          data: {
            modules: PERMISSION_MODULES,
            categories: getModulesByCategory()
          }
        })

      case 'roles':
        if (target) {
          const rolePermissions = permissionService.getRolePermissions(target as UserRole)
          const roleConfig = permissionService.getRolePermissionConfig(target as UserRole)
          return NextResponse.json({
            success: true,
            data: {
              role: target,
              permissions: rolePermissions,
              config: roleConfig
            }
          })
        } else {
          const allRoles: UserRole[] = ['user', 'courier', 'senior_courier', 'courier_coordinator', 'school_admin', 'platform_admin', 'admin', 'super_admin']
          const rolesData = allRoles.map(role => ({
            role,
            permissions: permissionService.getRolePermissions(role),
            config: permissionService.getRolePermissionConfig(role)
          }))
          return NextResponse.json({
            success: true,
            data: rolesData
          })
        }

      case 'courier-levels':
        if (target) {
          const level = parseInt(target) as CourierLevel
          const levelPermissions = permissionService.getCourierLevelPermissions(level)
          const levelConfig = permissionService.getCourierLevelPermissionConfig(level)
          return NextResponse.json({
            success: true,
            data: {
              level,
              permissions: levelPermissions,
              config: levelConfig
            }
          })
        } else {
          const allLevels: CourierLevel[] = [1, 2, 3, 4]
          const levelsData = allLevels.map(level => ({
            level,
            permissions: permissionService.getCourierLevelPermissions(level),
            config: permissionService.getCourierLevelPermissionConfig(level)
          }))
          return NextResponse.json({
            success: true,
            data: levelsData
          })
        }

      default:
        return NextResponse.json({
          success: false,
          error: 'Invalid type parameter'
        }, { status: 400 })
    }
  } catch (error) {
    console.error('Permission API GET error:', error)
    return NextResponse.json({
      success: false,
      error: '获取权限配置失败'
    }, { status: 500 })
  }
}

// ================================
// 更新权限配置
// ================================

export async function PUT(request: NextRequest) {
  try {
    const body = await request.json()
    const { type, target, permissions, modifiedBy } = body

    if (!modifiedBy) {
      return NextResponse.json({
        success: false,
        error: '缺少修改者信息'
      }, { status: 400 })
    }

    switch (type) {
      case 'role':
        const originalPermissions = permissionService.getRolePermissions(target as UserRole)
        await permissionService.updateRolePermissions(target as UserRole, permissions, modifiedBy)
        
        // 计算权限变更
        const changes = calculatePermissionChanges(originalPermissions, permissions)
        
        // 记录变更日志
        await logPermissionChange({
          type: 'role_update',
          target,
          permissions,
          modifiedBy,
          timestamp: new Date().toISOString()
        })

        // 广播权限变更通知
        const changeEvent = {
          type: 'permission_updated' as const,
          data: {
            target,
            targetType: 'role' as const,
            modifiedBy,
            timestamp: new Date().toISOString(),
            changes
          }
        }
        broadcastPermissionChange(changeEvent)
        
        // 分发客户端权限变更事件
        dispatchPermissionChangeEvent({
          type: 'permission_updated',
          target,
          targetType: 'role',
          modifiedBy,
          timestamp: new Date().toISOString()
        })

        return NextResponse.json({
          success: true,
          message: `角色 ${target} 权限已更新`,
          data: {
            role: target,
            permissions: permissionService.getRolePermissions(target),
            config: permissionService.getRolePermissionConfig(target)
          }
        })

      case 'courier-level':
        const level = parseInt(target) as CourierLevel
        const originalCourierPermissions = permissionService.getCourierLevelPermissions(level)
        await permissionService.updateCourierLevelPermissions(level, permissions, modifiedBy)
        
        // 计算权限变更
        const courierChanges = calculatePermissionChanges(originalCourierPermissions, permissions)
        
        // 记录变更日志
        await logPermissionChange({
          type: 'courier_level_update',
          target: level,
          permissions,
          modifiedBy,
          timestamp: new Date().toISOString()
        })

        // 广播权限变更通知
        broadcastPermissionChange({
          type: 'permission_updated',
          data: {
            target: level.toString(),
            targetType: 'courier-level',
            modifiedBy,
            timestamp: new Date().toISOString(),
            changes: courierChanges
          }
        })

        return NextResponse.json({
          success: true,
          message: `${level}级信使权限已更新`,
          data: {
            level,
            permissions: permissionService.getCourierLevelPermissions(level),
            config: permissionService.getCourierLevelPermissionConfig(level)
          }
        })

      default:
        return NextResponse.json({
          success: false,
          error: 'Invalid type parameter'
        }, { status: 400 })
    }
  } catch (error) {
    console.error('Permission API PUT error:', error)
    return NextResponse.json({
      success: false,
      error: '更新权限配置失败'
    }, { status: 500 })
  }
}

// ================================
// 重置权限配置
// ================================

export async function DELETE(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const type = searchParams.get('type')
    const target = searchParams.get('target')
    const modifiedBy = searchParams.get('modifiedBy')

    if (!modifiedBy) {
      return NextResponse.json({
        success: false,
        error: '缺少修改者信息'
      }, { status: 400 })
    }

    switch (type) {
      case 'role':
        await permissionService.resetRolePermissions(target as UserRole)
        
        // 记录变更日志
        await logPermissionChange({
          type: 'role_reset',
          target,
          permissions: [],
          modifiedBy,
          timestamp: new Date().toISOString()
        })

        // 广播权限重置通知
        broadcastPermissionChange({
          type: 'permission_reset',
          data: {
            target: target || '',
            targetType: 'role',
            modifiedBy,
            timestamp: new Date().toISOString()
          }
        })

        return NextResponse.json({
          success: true,
          message: `角色 ${target} 权限已重置为默认值`,
          data: {
            role: target,
            permissions: permissionService.getRolePermissions(target as UserRole),
            config: permissionService.getRolePermissionConfig(target as UserRole)
          }
        })

      case 'courier-level':
        const level = parseInt(target!) as CourierLevel
        await permissionService.resetCourierLevelPermissions(level)
        
        // 记录变更日志
        await logPermissionChange({
          type: 'courier_level_reset',
          target: level,
          permissions: [],
          modifiedBy,
          timestamp: new Date().toISOString()
        })

        // 广播权限重置通知
        broadcastPermissionChange({
          type: 'permission_reset',
          data: {
            target: level.toString(),
            targetType: 'courier-level',
            modifiedBy,
            timestamp: new Date().toISOString()
          }
        })

        return NextResponse.json({
          success: true,
          message: `${level}级信使权限已重置为默认值`,
          data: {
            level,
            permissions: permissionService.getCourierLevelPermissions(level),
            config: permissionService.getCourierLevelPermissionConfig(level)
          }
        })

      case 'all':
        // 重置所有配置
        const roles: UserRole[] = ['user', 'courier', 'senior_courier', 'courier_coordinator', 'school_admin', 'platform_admin', 'admin', 'super_admin']
        const levels: CourierLevel[] = [1, 2, 3, 4]
        
        await Promise.all([
          ...roles.map(role => permissionService.resetRolePermissions(role)),
          ...levels.map(level => permissionService.resetCourierLevelPermissions(level))
        ])

        // 记录变更日志
        await logPermissionChange({
          type: 'system_reset',
          target: 'all',
          permissions: [],
          modifiedBy,
          timestamp: new Date().toISOString()
        })

        // 广播系统重置通知
        broadcastPermissionChange({
          type: 'permission_reset',
          data: {
            target: 'all',
            targetType: 'system',
            modifiedBy,
            timestamp: new Date().toISOString()
          }
        })

        return NextResponse.json({
          success: true,
          message: '所有权限配置已重置为默认值'
        })

      default:
        return NextResponse.json({
          success: false,
          error: 'Invalid type parameter'
        }, { status: 400 })
    }
  } catch (error) {
    console.error('Permission API DELETE error:', error)
    return NextResponse.json({
      success: false,
      error: '重置权限配置失败'
    }, { status: 500 })
  }
}

// ================================
// 权限配置导入导出
// ================================

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { action, data, modifiedBy } = body

    switch (action) {
      case 'export':
        const config = permissionService.exportConfigs()
        return NextResponse.json({
          success: true,
          data: {
            config,
            filename: `permission-config-${new Date().toISOString().slice(0, 10)}.json`,
            timestamp: new Date().toISOString()
          }
        })

      case 'import':
        if (!data || !modifiedBy) {
          return NextResponse.json({
            success: false,
            error: '缺少导入数据或修改者信息'
          }, { status: 400 })
        }

        await permissionService.importConfigs(data, true) // overwrite = true
        
        // 记录变更日志
        await logPermissionChange({
          type: 'config_import',
          target: 'system',
          permissions: [],
          modifiedBy,
          timestamp: new Date().toISOString()
        })

        // 广播配置导入通知
        broadcastPermissionChange({
          type: 'config_imported',
          data: {
            target: 'system',
            targetType: 'system',
            modifiedBy,
            timestamp: new Date().toISOString()
          }
        })

        return NextResponse.json({
          success: true,
          message: '权限配置导入成功'
        })

      case 'validate':
        // 验证配置数据格式
        try {
          const parsedData = typeof data === 'string' ? JSON.parse(data) : data
          const isValid = validateConfigData(parsedData)
          
          return NextResponse.json({
            success: true,
            data: {
              valid: isValid,
              preview: isValid ? generateConfigPreview(parsedData) : null
            }
          })
        } catch (error) {
          return NextResponse.json({
            success: false,
            error: '配置数据格式无效'
          }, { status: 400 })
        }

      default:
        return NextResponse.json({
          success: false,
          error: 'Invalid action parameter'
        }, { status: 400 })
    }
  } catch (error) {
    console.error('Permission API POST error:', error)
    return NextResponse.json({
      success: false,
      error: '权限配置操作失败'
    }, { status: 500 })
  }
}

// ================================
// 辅助函数
// ================================

function getModulesByCategory() {
  const categories = {
    basic: [],
    courier: [],
    management: [],
    admin: [],
    system: []
  } as any

  Object.values(PERMISSION_MODULES).forEach(module => {
    categories[module.category].push(module)
  })

  return categories
}

async function logPermissionChange(change: {
  type: string
  target: any
  permissions: string[]
  modifiedBy: string
  timestamp: string
}) {
  try {
    // 记录到控制台
    console.log('Permission Change Log:', change)
    
    // 发送到审计日志API
    const auditResponse = await fetch('/api/admin/permissions/audit', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        type: change.type,
        target: change.target,
        targetType: change.type.includes('role') ? 'role' : 
                   change.type.includes('courier') ? 'courier-level' : 'system',
        modifiedBy: change.modifiedBy,
        changes: change.type.includes('reset') ? undefined : {
          added: change.permissions,
          removed: []
        }
      })
    })
    
    if (!auditResponse.ok) {
      console.error('审计日志记录失败:', await auditResponse.text())
    }
  } catch (error) {
    console.error('记录权限变更日志失败:', error)
  }
}

function validateConfigData(data: any): boolean {
  try {
    if (!data || typeof data !== 'object') return false
    if (!data.version || !data.timestamp) return false
    
    // 验证角色配置
    if (data.roleConfigs) {
      for (const [roleId, config] of Object.entries(data.roleConfigs as any)) {
        if (!config || typeof config !== 'object') return false
        if (!Array.isArray((config as any).permissions)) return false
      }
    }

    // 验证信使等级配置
    if (data.courierLevelConfigs) {
      for (const [level, config] of Object.entries(data.courierLevelConfigs as any)) {
        if (!config || typeof config !== 'object') return false
        if (!Array.isArray((config as any).permissions)) return false
      }
    }

    return true
  } catch {
    return false
  }
}

function generateConfigPreview(data: any) {
  const preview = {
    roleChanges: 0,
    courierLevelChanges: 0,
    totalPermissions: 0,
    lastModified: data.timestamp
  }

  if (data.roleConfigs) {
    preview.roleChanges = Object.keys(data.roleConfigs).length
    Object.values(data.roleConfigs).forEach((config: any) => {
      preview.totalPermissions += (config.permissions || []).length
    })
  }

  if (data.courierLevelConfigs) {
    preview.courierLevelChanges = Object.keys(data.courierLevelConfigs).length
    Object.values(data.courierLevelConfigs).forEach((config: any) => {
      preview.totalPermissions += (config.permissions || []).length
    })
  }

  return preview
}

function calculatePermissionChanges(originalPermissions: string[], newPermissions: string[]) {
  const added = newPermissions.filter(p => !originalPermissions.includes(p))
  const removed = originalPermissions.filter(p => !newPermissions.includes(p))
  
  return { added, removed }
}