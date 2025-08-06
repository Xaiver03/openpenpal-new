/**
 * Optimized State Subscriptions Hook - 优化的状态订阅钩子
 * Performance-optimized selectors and subscriptions for user store
 */

import { useMemo, useCallback } from 'react'
import { shallow } from 'zustand/shallow'
import { useUserStore, type User } from '@/stores/user-store'
import { type UserRole, type Permission } from '@/constants/roles'

/**
 * Optimized selector for user basic info
 */
export function useUserBasicInfo() {
  return useUserStore(
    useCallback((state) => ({
      id: state.user?.id,
      username: state.user?.username,
      nickname: state.user?.nickname,
      email: state.user?.email,
      avatar: state.user?.avatar,
      isAuthenticated: state.isAuthenticated
    }), []),
    shallow
  )
}

/**
 * Optimized selector for user role and permissions
 */
export function useUserRoleInfo() {
  return useUserStore(
    useCallback((state) => ({
      role: state.user?.role,
      permissions: state.user?.permissions,
      canAccessAdmin: state.canAccessAdmin(),
      isCourier: state.isCourier()
    }), []),
    shallow
  )
}

/**
 * Optimized selector for courier information
 */
export function useCourierInfo() {
  return useUserStore(
    useCallback((state) => ({
      courierInfo: state.user?.courierInfo,
      level: state.getCourierLevel(),
      levelName: state.getCourierLevelName(),
      isCourier: state.isCourier()
    }), []),
    shallow
  )
}

/**
 * Optimized selector for loading states
 */
export function useLoadingStates() {
  return useUserStore(
    useCallback((state) => ({
      isLoading: state.loading.isLoading,
      isRefreshing: state.loading.isRefreshing,
      error: state.loading.error,
      lastUpdated: state.loading.lastUpdated
    }), []),
    shallow
  )
}

/**
 * Memoized permission checker
 */
export function usePermissionChecker() {
  const hasPermission = useUserStore(state => state.hasPermission)
  const hasRole = useUserStore(state => state.hasRole)
  
  // Memoize permission checkers to prevent unnecessary re-renders
  const memoizedPermissionChecker = useMemo(() => ({
    hasPermission: (permission: Permission) => hasPermission(permission),
    hasRole: (role: UserRole) => hasRole(role),
    hasAnyPermission: (permissions: Permission[]) => 
      permissions.some(permission => hasPermission(permission)),
    hasAllPermissions: (permissions: Permission[]) =>
      permissions.every(permission => hasPermission(permission))
  }), [hasPermission, hasRole])

  return memoizedPermissionChecker
}

/**
 * Optimized selector for authentication actions
 */
export function useAuthActions() {
  return useUserStore(
    useCallback((state) => ({
      login: state.login,
      logout: state.logout,
      refreshUser: state.refreshUser,
      clearError: state.clearError
    }), []),
    shallow
  )
}

/**
 * Selective user data hook with field-level optimization
 */
export function useSelectiveUserData<T>(
  selector: (user: User | null) => T,
  dependencies: any[] = []
) {
  return useUserStore(
    useCallback((state) => selector(state.user), dependencies),
    shallow
  )
}

/**
 * Optimized hook for user profile management
 */
export function useUserProfile() {
  const basicInfo = useUserBasicInfo()
  const updateUser = useUserStore(state => state.updateUser)
  const optimisticUpdate = useUserStore(state => state.optimisticUpdate)

  const updateProfile = useCallback(async (updates: Partial<User>) => {
    return optimisticUpdate(
      () => updateUser(updates),
      async () => {
        // TODO: Implement actual API call
        // const response = await AuthService.updateProfile(updates)
        // return response
        return updates
      }
    )
  }, [updateUser, optimisticUpdate])

  return {
    ...basicInfo,
    updateProfile
  }
}

/**
 * Debounced state selector for performance
 */
export function useDebouncedUserState(delay: number = 300) {
  const user = useUserStore(state => state.user)
  
  // Simple debounce implementation
  const [debouncedUser, setDebouncedUser] = useState(user)
  
  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedUser(user)
    }, delay)
    
    return () => {
      clearTimeout(handler)
    }
  }, [user, delay])
  
  return debouncedUser
}

/**
 * Performance monitoring hook for store subscriptions
 */
export function useStorePerformanceMonitor() {
  const renderCount = useRef(0)
  const subscriptions = useRef<string[]>([])
  
  useEffect(() => {
    renderCount.current += 1
    
    // Log performance metrics in development
    if (process.env.NODE_ENV === 'development') {
      console.log(`[Performance] Store render count: ${renderCount.current}`)
      console.log(`[Performance] Active subscriptions: ${subscriptions.current.length}`)
    }
  })
  
  const trackSubscription = useCallback((subscriptionName: string) => {
    subscriptions.current.push(subscriptionName)
    
    return () => {
      subscriptions.current = subscriptions.current.filter(
        name => name !== subscriptionName
      )
    }
  }, [])
  
  return {
    renderCount: renderCount.current,
    subscriptionCount: subscriptions.current.length,
    trackSubscription
  }
}

/**
 * Batch selector for multiple user properties
 */
export function useBatchUserData() {
  return useUserStore(
    useCallback((state) => {
      const user = state.user
      if (!user) return null
      
      return {
        // Basic info
        basic: {
          id: user.id,
          username: user.username,
          nickname: user.nickname,
          email: user.email,
          avatar: user.avatar
        },
        
        // Role and permissions
        auth: {
          role: user.role,
          permissions: user.permissions,
          isAuthenticated: state.isAuthenticated
        },
        
        // School info
        school: {
          schoolCode: user.schoolCode,
          school_name: user.school_name
        },
        
        // Courier info
        courier: user.courierInfo ? {
          level: user.courierInfo.level,
          zoneCode: user.courierInfo.zoneCode,
          zoneType: user.courierInfo.zoneType,
          status: user.courierInfo.status,
          points: user.courierInfo.points
        } : null,
        
        // Status
        status: {
          status: user.status,
          isActive: user.isActive,
          lastLoginAt: user.lastLoginAt
        }
      }
    }, []),
    shallow
  )
}

// Re-export useState and useEffect for the debounced hook
import { useState, useEffect, useRef } from 'react'

export default {
  useUserBasicInfo,
  useUserRoleInfo,
  useCourierInfo,
  useLoadingStates,
  usePermissionChecker,
  useAuthActions,
  useSelectiveUserData,
  useUserProfile,
  useDebouncedUserState,
  useStorePerformanceMonitor,
  useBatchUserData
}