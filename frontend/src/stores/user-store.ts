/**
 * Global User State Store - 全局用户状态管理
 * Unified state management for user data, authentication, permissions, and loading states
 */

import React from 'react'
import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { AuthService } from '@/lib/services/auth-service'
import { EnhancedAuthService } from '@/lib/services/auth-service-enhanced'
import { AuthOrchestrator } from '@/lib/auth/auth-orchestrator'
import { 
  hasPermission, 
  canAccessAdmin, 
  getCourierLevelName,
  type UserRole, 
  type Permission 
} from '@/constants/roles'
import { permissionService } from '@/lib/permissions/permission-service'
import { TokenManager } from '@/lib/auth/cookie-token-manager'
import { tokenRefreshInterceptor } from '@/lib/auth/token-refresh-interceptor'
import { isTestMode, createTestCourierUser, getTestCourierLevel } from '@/lib/auth/test-courier-mock'


export interface User {
  id: string
  username: string
  nickname: string
  email: string
  role: UserRole
  school_code: string
  school_name?: string
  avatar?: string
  bio?: string
  address?: string
  created_at: string
  updated_at: string
  last_login_at?: string
  status?: 'active' | 'inactive' | 'banned'
  is_active?: boolean
  permissions: Permission[]
  courierInfo?: CourierInfo
  token?: string // Temporary bridge for legacy code - use TokenManager.get() instead
}

export interface CourierInfo {
  level: 1 | 2 | 3 | 4
  zoneCode: string
  zoneType: 'city' | 'school' | 'zone' | 'building'
  status: 'active' | 'pending' | 'frozen'
  points: number
  taskCount: number
  completedTasks: number
  averageRating: number
  lastActiveAt: string
  school_code: string
  username: string
  school_name: string
}

export interface LoadingState {
  isLoading: boolean
  isRefreshing: boolean
  lastUpdated: number | null
  error: string | null
}

export interface UserStoreState {
  user: User | null
  isAuthenticated: boolean
  
  loading: LoadingState
  
  setUser: (user: User | null) => void
  updateUser: (updates: Partial<User>) => void
  updateCourierInfo: (courierInfo: Partial<CourierInfo>) => void
  
  login: (credentials: { username: string; password: string }) => Promise<{ success: boolean; error?: string }>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
  
  hasPermission: (permission: Permission) => boolean
  hasAnyPermission: (permissions: Permission[]) => boolean
  hasAllPermissions: (permissions: Permission[]) => boolean
  getUserPermissions: () => string[]
  getUserPermissionDetails: () => Array<{id: string, module: any, granted: boolean}>
  hasRole: (role: UserRole) => boolean
  canAccessAdmin: () => boolean
  isCourier: () => boolean
  getCourierLevel: () => number | null
  getCourierLevelName: () => string | null
  refreshPermissions: () => Promise<void>
  
  setLoading: (loading: Partial<LoadingState>) => void
  clearError: () => void
  reset: () => void
  
  // Optimistic updates
  optimisticUpdate: <T>(
    updateFn: () => void,
    asyncAction: () => Promise<T>,
    rollbackFn?: () => void
  ) => Promise<T>
}

// ================================
// Initial State
// ================================

const initialLoadingState: LoadingState = {
  isLoading: false,
  isRefreshing: false,
  lastUpdated: null,
  error: null
}

// ================================
// User Store Implementation
// ================================

export const useUserStore = create<UserStoreState>()(
  devtools(
    persist(
      (set, get) => ({
        // Initial state - consistent for server and client to avoid hydration errors
        user: null,
        isAuthenticated: false,
        loading: initialLoadingState,

        // ================================
        // Core State Actions
        // ================================

        setUser: (user: User | null) => {
          // 在开发模式下，如果启用了测试信使模式，修改用户数据
          if (user && isTestMode()) {
            const testLevel = getTestCourierLevel()
            user = createTestCourierUser(user, testLevel)
            console.log('🧪 应用测试信使模式:', user)
          }
          
          // Save user to cookie for persistence
          if (user) {
            TokenManager.setUser(user)
          } else {
            TokenManager.clear()
          }
          
          set(
            (state) => ({
              user,
              isAuthenticated: !!user,
              loading: {
                ...state.loading,
                lastUpdated: user ? Date.now() : null,
                error: null
              }
            }),
            false,
            'setUser'
          )
        },

        updateUser: (updates: Partial<User>) => {
          set(
            (state) => {
              if (!state.user) return state
              
              const updatedUser = { ...state.user, ...updates }
              
              // Update user in cookie
              TokenManager.setUser(updatedUser)
              
              return {
                user: updatedUser,
                loading: {
                  ...state.loading,
                  lastUpdated: Date.now(),
                  error: null
                }
              }
            },
            false,
            'updateUser'
          )
        },

        updateCourierInfo: (courierInfo: Partial<CourierInfo>) => {
          set(
            (state) => {
              if (!state.user) return state
              
              const updatedUser = {
                ...state.user,
                courierInfo: state.user.courierInfo 
                  ? { ...state.user.courierInfo, ...courierInfo }
                  : courierInfo as CourierInfo
              }
              
              return {
                user: updatedUser,
                loading: {
                  ...state.loading,
                  lastUpdated: Date.now(),
                  error: null
                }
              }
            },
            false,
            'updateCourierInfo'
          )
        },

        // ================================
        // Authentication Actions
        // ================================

        login: async (credentials: { username: string; password: string }) => {
          const { setLoading, setUser } = get()
          
          setLoading({ isLoading: true, error: null })
          
          try {
            // 使用增强版认证服务
        const response = await EnhancedAuthService.login(credentials)
            
            if (response.success && response.data?.user) {
              // Add token to user object for legacy compatibility
              const userWithToken = {
                ...response.data.user,
                token: response.data.token
              }
              setUser(userWithToken as any)
              setLoading({ isLoading: false })
              
              // 登录成功后启动token自动刷新
              if (typeof window !== 'undefined') {
                // Delay token refresh initialization to avoid conflicts
                setTimeout(() => {
                  tokenRefreshInterceptor.initialize()
                }, 1000)
                window.dispatchEvent(new CustomEvent('auth:login', {
                  detail: { user: response.data.user }
                }))
              }
              
              return { success: true }
            } else {
              setLoading({ isLoading: false, error: response.message || '登录失败' })
              return { success: false, error: response.message || '登录失败' }
            }
          } catch (error) {
            const errorMessage = error instanceof Error ? error.message : '登录失败'
            setLoading({ isLoading: false, error: errorMessage })
            return { success: false, error: errorMessage }
          }
        },

        logout: async () => {
          const { setUser, setLoading } = get()
          
          setLoading({ isLoading: true })
          
          try {
            // 使用增强版认证服务
        await EnhancedAuthService.logout()
          } catch (error) {
            console.error('Logout error:', error)
          } finally {
            setUser(null)
            setLoading({ isLoading: false, error: null })
            
            // 触发登出事件（标记为用户主动登出）
            if (typeof window !== 'undefined') {
              window.dispatchEvent(new CustomEvent('auth:logout', {
                detail: { source: 'user_action' }
              }))
            }
          }
        },

        refreshUser: async () => {
          const { setLoading, setUser } = get()
          
          setLoading({ isRefreshing: true, error: null })
          
          try {
            // 使用增强版认证服务
        const response = await EnhancedAuthService.getCurrentUser()
            
            if (response.success && response.data) {
              setUser(response.data as any)
            } else {
              // If refresh fails, clear the user state
              setUser(null)
            }
          } catch (error) {
            console.error('Refresh user error:', error)
            setUser(null)
            setLoading({ 
              isRefreshing: false, 
              error: error instanceof Error ? error.message : '刷新用户信息失败' 
            })
            return
          }
          
          setLoading({ isRefreshing: false })
        },

        // ================================
        // Permission Methods
        // ================================

        hasPermission: (permission: Permission) => {
          const { user } = get()
          
          if (!user) return false
          
          // Use the new dynamic permission service
          return permissionService.hasPermission(user, permission)
        },

        hasRole: (role: UserRole) => {
          const { user } = get()
          return user?.role === role
        },

        canAccessAdmin: () => {
          const { user } = get()
          if (!user) return false
          return permissionService.canAccessAdmin(user)
        },

        isCourier: () => {
          const { user } = get()
          if (!user) return false
          
          return permissionService.isCourier(user)
        },

        getCourierLevel: () => {
          const { user } = get()
          return user?.courierInfo?.level || null
        },

        getCourierLevelName: () => {
          const { user } = get()
          if (!user?.courierInfo?.level) return null
          return getCourierLevelName(user.courierInfo.level)
        },

        // ================================
        // Enhanced Permission Methods (SOTA)
        // ================================

        hasAnyPermission: (permissions: Permission[]) => {
          const { user } = get()
          if (!user) return false
          return permissionService.hasAnyPermission(user, permissions)
        },

        hasAllPermissions: (permissions: Permission[]) => {
          const { user } = get()
          if (!user) return false
          return permissionService.hasAllPermissions(user, permissions)
        },

        getUserPermissions: () => {
          const { user } = get()
          if (!user) return []
          return permissionService.getUserPermissions(user)
        },

        getUserPermissionDetails: () => {
          const { user } = get()
          if (!user) return []
          return permissionService.getUserPermissionDetails(user)
        },

        refreshPermissions: async () => {
          await permissionService.refreshPermissions()
        },

        // ================================
        // State Management
        // ================================

        setLoading: (updates: Partial<LoadingState>) => {
          set(
            (state) => ({
              loading: { ...state.loading, ...updates }
            }),
            false,
            'setLoading'
          )
        },

        clearError: () => {
          set(
            (state) => ({
              loading: { ...state.loading, error: null }
            }),
            false,
            'clearError'
          )
        },

        reset: () => {
          set(
            {
              user: null,
              isAuthenticated: false,
              loading: initialLoadingState
            },
            false,
            'reset'
          )
        },

        // ================================
        // Optimistic Updates
        // ================================

        optimisticUpdate: async <T>(
          updateFn: () => void,
          asyncAction: () => Promise<T>,
          rollbackFn?: () => void
        ): Promise<T> => {
          // Store current state for rollback
          const currentState = get()
          const rollback = rollbackFn || (() => {
            set(currentState, false, 'rollback')
          })
          
          try {
            // Apply optimistic update
            updateFn()
            
            // Perform async action
            const result = await asyncAction()
            
            return result
          } catch (error) {
            // Rollback on error
            rollback()
            throw error
          }
        }
      }),
      {
        name: 'openpenpal-user-store',
        partialize: (state) => ({
          user: state.user,
          isAuthenticated: state.isAuthenticated
        }),
        // Skip hydration to prevent mismatch errors
        skipHydration: true
      }
    ),
    {
      name: 'user-store'
    }
  )
)

// ================================
// Client-side Initialization Hook
// ================================

/**
 * Hook to initialize user state from persisted storage after hydration
 * This prevents hydration mismatch by only loading user data on the client side
 */
export const useClientUserInitialization = () => {
  const setUser = useUserStore((state) => state.setUser)
  const [isInitialized, setIsInitialized] = React.useState(false)
  
  React.useEffect(() => {
    // Only run on client side after hydration
    if (typeof window !== 'undefined' && !isInitialized) {
      // Try to restore user from TokenManager
      const storedUser = TokenManager.getUser()
      if (storedUser) {
        setUser(storedUser)
      }
      setIsInitialized(true)
    }
  }, [isInitialized, setUser])
  
  return isInitialized
}

// ================================
// Convenience Hooks
// ================================

/**
 * Hook for accessing user data with automatic refresh
 */
export const useUser = () => {
  const store = useUserStore()
  
  // Auto-refresh user data if stale (older than 5 minutes)
  React.useEffect(() => {
    const { user, loading, refreshUser } = store
    
    if (user && loading.lastUpdated) {
      const fiveMinutes = 5 * 60 * 1000
      const isStale = Date.now() - loading.lastUpdated > fiveMinutes
      
      if (isStale && !loading.isRefreshing) {
        refreshUser()
      }
    }
  }, [store.user, store.loading.lastUpdated])
  
  return {
    user: store.user,
    isAuthenticated: store.isAuthenticated,
    isLoading: store.loading.isLoading,
    isRefreshing: store.loading.isRefreshing,
    error: store.loading.error
  }
}

/**
 * Hook for accessing permission methods
 */
export const usePermissions = () => {
  const store = useUserStore()
  
  return {
    hasPermission: store.hasPermission,
    hasRole: store.hasRole,
    canAccessAdmin: store.canAccessAdmin,
    isCourier: store.isCourier,
    getCourierLevel: store.getCourierLevel,
    getCourierLevelName: store.getCourierLevelName
  }
}

/**
 * Hook for courier-specific data and methods
 */
export const useCourier = () => {
  const store = useUserStore()
  
  return {
    courierInfo: store.user?.courierInfo,
    isCourier: store.isCourier(),
    level: store.getCourierLevel(),
    levelName: store.getCourierLevelName(),
    updateCourierInfo: store.updateCourierInfo
  }
}

/**
 * Hook for authentication actions
 */
export const useAuth = () => {
  const store = useUserStore()
  
  return {
    user: store.user,
    isAuthenticated: store.isAuthenticated,
    isLoading: store.loading.isLoading,
    error: store.loading.error,
    login: store.login,
    logout: store.logout,
    refreshUser: store.refreshUser,
    clearError: store.clearError
  }
}

export default useUserStore