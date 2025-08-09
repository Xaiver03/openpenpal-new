/**
 * Unit tests for role system
 * 角色系统单元测试
 */

import {
  ROLE_CONFIGS,
  COURIER_LEVEL_CONFIGS,
  getRoleDisplayName,
  getRoleColors,
  getCourierLevelName,
  getCourierLevelManagementPath,
  hasPermission,
  canManageSublevels,
  getAllRoleOptions,
  getAllCourierLevelOptions,
  type UserRole,
  type Permission,
  type CourierLevel
} from '../roles'

describe('Role System', () => {
  describe('Role Constants', () => {
    test('ROLE_CONFIGS contains all expected roles (PRD compliant)', () => {
      expect(ROLE_CONFIGS.user).toBeDefined()
      expect(ROLE_CONFIGS.courier_level1).toBeDefined()
      expect(ROLE_CONFIGS.courier_level2).toBeDefined()
      expect(ROLE_CONFIGS.courier_level3).toBeDefined()
      expect(ROLE_CONFIGS.courier_level4).toBeDefined()
      expect(ROLE_CONFIGS.platform_admin).toBeDefined()
      expect(ROLE_CONFIGS.super_admin).toBeDefined()
      
      // Verify old roles are not present
      expect(ROLE_CONFIGS['courier' as keyof typeof ROLE_CONFIGS]).toBeUndefined()
      expect(ROLE_CONFIGS['senior_courier' as keyof typeof ROLE_CONFIGS]).toBeUndefined()
      expect(ROLE_CONFIGS['courier_coordinator' as keyof typeof ROLE_CONFIGS]).toBeUndefined()
      expect(ROLE_CONFIGS['school_admin' as keyof typeof ROLE_CONFIGS]).toBeUndefined()
    })

    test('COURIER_LEVEL_CONFIGS contains all levels', () => {
      expect(COURIER_LEVEL_CONFIGS[1]).toBeDefined()
      expect(COURIER_LEVEL_CONFIGS[2]).toBeDefined()
      expect(COURIER_LEVEL_CONFIGS[3]).toBeDefined()
      expect(COURIER_LEVEL_CONFIGS[4]).toBeDefined()
    })

    test('All roles have permissions defined', () => {
      Object.values(ROLE_CONFIGS).forEach(roleConfig => {
        expect(Array.isArray(roleConfig.permissions)).toBe(true)
        expect(roleConfig.permissions.length).toBeGreaterThan(0)
      })
    })
  })

  describe('getRoleDisplayName', () => {
    test('returns correct display names for all roles (PRD compliant)', () => {
      expect(getRoleDisplayName('user')).toBe('普通用户')
      expect(getRoleDisplayName('courier_level1')).toBe('一级信使（基础投递）')
      expect(getRoleDisplayName('courier_level2')).toBe('二级信使（片区协调员）')
      expect(getRoleDisplayName('courier_level3')).toBe('三级信使（校区负责人）')
      expect(getRoleDisplayName('courier_level4')).toBe('四级信使（城市负责人）')
      expect(getRoleDisplayName('platform_admin')).toBe('平台管理员')
      expect(getRoleDisplayName('super_admin')).toBe('超级管理员')
    })

    test('returns fallback for unknown role', () => {
      expect(getRoleDisplayName('unknown' as UserRole)).toBe('unknown')
    })
  })

  describe('getRoleColors', () => {
    test('returns color configuration for all roles', () => {
      const roles = Object.keys(ROLE_CONFIGS) as UserRole[]
      roles.forEach(role => {
        const colors = getRoleColors(role)
        expect(colors).toHaveProperty('badge')
        expect(colors).toHaveProperty('bg')
        expect(colors).toHaveProperty('text')
        expect(typeof colors.badge).toBe('string')
        expect(typeof colors.bg).toBe('string')
        expect(typeof colors.text).toBe('string')
      })
    })

    test('returns specific colors for user role', () => {
      const colors = getRoleColors('user')
      expect(colors.badge).toContain('gray')
      expect(colors.bg).toContain('gray')
      expect(colors.text).toContain('white')
    })
  })

  describe('getCourierLevelName', () => {
    test('returns correct names for all courier levels', () => {
      expect(getCourierLevelName(1)).toBe('一级信使')
      expect(getCourierLevelName(2)).toBe('二级信使')
      expect(getCourierLevelName(3)).toBe('三级信使')
      expect(getCourierLevelName(4)).toBe('四级信使')
    })

    test('returns fallback for invalid level', () => {
      expect(getCourierLevelName(0 as any)).toBe('0级信使')
      expect(getCourierLevelName(5 as any)).toBe('5级信使')
    })
  })

  describe('getCourierLevelManagementPath', () => {
    test('returns correct paths for all levels', () => {
      expect(getCourierLevelManagementPath(1)).toBe('/courier/building-manage')
      expect(getCourierLevelManagementPath(2)).toBe('/courier/zone-manage')
      expect(getCourierLevelManagementPath(3)).toBe('/courier/school-manage')
      expect(getCourierLevelManagementPath(4)).toBe('/courier/city-manage')
    })

    test('returns default path for invalid level', () => {
      expect(getCourierLevelManagementPath(0 as any)).toBe('/courier')
      expect(getCourierLevelManagementPath(5 as any)).toBe('/courier')
    })
  })

  describe('hasPermission', () => {
    test('returns true for courier role with specific permission', () => {
      expect(hasPermission('courier_level1', 'COURIER_SCAN_CODE')).toBe(true)
      expect(hasPermission('courier_level1', 'COURIER_DELIVER_LETTER')).toBe(true)
    })

    test('returns false for role without specific permission', () => {
      expect(hasPermission('user', 'MANAGE_USERS')).toBe(false)
    })

    test('returns true for super_admin with any permission', () => {
      expect(hasPermission('super_admin', 'MANAGE_USERS')).toBe(true)
      expect(hasPermission('super_admin', 'COURIER_SCAN_CODE')).toBe(true)
    })

    test('returns false for invalid role', () => {
      expect(hasPermission('invalid_role' as UserRole, 'READ_LETTER')).toBe(false)
    })
  })

  describe('canManageSublevels', () => {
    test('returns correct manageable status for courier levels', () => {
      expect(canManageSublevels(4)).toBe(true)
      expect(canManageSublevels(3)).toBe(true)
      expect(canManageSublevels(2)).toBe(true)
      expect(canManageSublevels(1)).toBe(false)
    })

    test('handles invalid levels', () => {
      expect(canManageSublevels(0 as CourierLevel)).toBe(false)
      expect(canManageSublevels(5 as CourierLevel)).toBe(false)
    })
  })

  describe('getAllRoleOptions', () => {
    test('returns all role options with correct format', () => {
      const options = getAllRoleOptions()
      expect(Array.isArray(options)).toBe(true)
      expect(options.length).toBeGreaterThan(0)
      
      options.forEach(option => {
        expect(option).toHaveProperty('value')
        expect(option).toHaveProperty('label')
        expect(typeof option.value).toBe('string')
        expect(typeof option.label).toBe('string')
      })
    })

    test('includes all user roles', () => {
      const options = getAllRoleOptions()
      const values = options.map(opt => opt.value)
      const expectedRoles = Object.keys(ROLE_CONFIGS)
      
      expectedRoles.forEach(role => {
        expect(values).toContain(role)
      })
    })
  })

  describe('Role Hierarchies', () => {
    test('roles have correct hierarchy levels (PRD compliant)', () => {
      expect(ROLE_CONFIGS.user.hierarchy).toBe(1)
      expect(ROLE_CONFIGS.courier_level1.hierarchy).toBe(2)
      expect(ROLE_CONFIGS.courier_level2.hierarchy).toBe(3)
      expect(ROLE_CONFIGS.courier_level3.hierarchy).toBe(4)
      expect(ROLE_CONFIGS.courier_level4.hierarchy).toBe(5)
      expect(ROLE_CONFIGS.platform_admin.hierarchy).toBe(6)
      expect(ROLE_CONFIGS.super_admin.hierarchy).toBe(7)
    })

    test('admin roles have access to admin panel', () => {
      expect(ROLE_CONFIGS.super_admin.canAccessAdmin).toBe(true)
      expect(ROLE_CONFIGS.platform_admin.canAccessAdmin).toBe(true)
    })

    test('courier roles have appropriate admin access based on level', () => {
      expect(ROLE_CONFIGS.user.canAccessAdmin).toBe(false)
      expect(ROLE_CONFIGS.courier_level1.canAccessAdmin).toBe(false)
      expect(ROLE_CONFIGS.courier_level2.canAccessAdmin).toBe(true)
      expect(ROLE_CONFIGS.courier_level3.canAccessAdmin).toBe(true)
      expect(ROLE_CONFIGS.courier_level4.canAccessAdmin).toBe(true)
    })
  })

  describe('Permission Validation', () => {
    test('all roles have valid permission arrays', () => {
      Object.values(ROLE_CONFIGS).forEach(roleConfig => {
        expect(Array.isArray(roleConfig.permissions)).toBe(true)
        expect(roleConfig.permissions.length).toBeGreaterThan(0)
        
        roleConfig.permissions.forEach(permission => {
          expect(typeof permission).toBe('string')
          expect(permission.length).toBeGreaterThan(0)
        })
      })
    })

    test('super_admin has most permissions', () => {
      const superAdminPermissions = ROLE_CONFIGS.super_admin.permissions
      const otherRolePermissions = Object.values(ROLE_CONFIGS)
        .filter(config => config.id !== 'super_admin')
        .map(config => config.permissions)
      
      otherRolePermissions.forEach(permissions => {
        expect(superAdminPermissions.length).toBeGreaterThanOrEqual(permissions.length)
      })
    })

    test('user role has minimal permissions', () => {
      const userPermissions = ROLE_CONFIGS.user.permissions
      const otherRolePermissions = Object.values(ROLE_CONFIGS)
        .filter(config => config.id !== 'user')
        .map(config => config.permissions)
      
      otherRolePermissions.forEach(permissions => {
        expect(userPermissions.length).toBeLessThanOrEqual(permissions.length)
      })
    })
  })
})