/**
 * Enhanced Authentication Context - Compatible with New User Store
 * 增强的认证上下文 - 兼容新的用户状态管理
 */

'use client'

import { createContext, useContext, useEffect, ReactNode, useCallback } from 'react'
import { useUserStore, useAuth as useAuthStore, usePermissions, useUser, type CourierInfo } from '@/stores/user-store'
import { TokenManager, wsManager } from '@/lib/api-client'
import { type UserRole, type Permission } from '@/constants/roles'
import { log } from '@/utils/logger'

// Legacy interfaces for backward compatibility
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
  status: 'active' | 'inactive' | 'banned'
  is_active?: boolean
  permissions: Permission[]
  courierInfo?: CourierInfo
}

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
  email: string
  nickname?: string
}

interface AuthContextType {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  permissions: Permission[]
  login: (data: LoginRequest) => Promise<void>
  register: (data: RegisterRequest) => Promise<{ success: boolean; message?: string }>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
  checkPermission: (permission: Permission) => boolean
  hasRole: (role: UserRole) => boolean
  updateProfile: (data: Partial<User>) => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  // Use the new user store
  const { 
    user: storeUser, 
    isAuthenticated, 
    isLoading, 
    error,
    login: storeLogin,
    logout: storeLogout,
    refreshUser: storeRefreshUser,
    clearError
  } = useAuthStore()
  
  const { hasPermission, hasRole } = usePermissions()
  const { updateUser } = useUserStore()

  // Convert store user to legacy format
  const user: User | null = storeUser ? {
    ...storeUser,
    permissions: storeUser.permissions
  } as User : null

  const permissions = storeUser?.permissions || []

  // Auto-initialize WebSocket connection based on auth state
  // 使用防抖机制，避免临时认证状态变化导致意外登出
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (isAuthenticated && user) {
        console.log('🔐 Auth state: User authenticated, connecting WebSocket')
        // Connect WebSocket
        wsManager.connect().catch(console.error)
        
        // Emit auth events for legacy components
        const authEvent = new CustomEvent('auth:login', { 
          detail: { user } 
        })
        window.dispatchEvent(authEvent)
      } else if (!isLoading) {
        // 只有在不处于加载状态时才处理登出
        // 这避免了初始化期间的临时未认证状态触发登出
        console.log('🔐 Auth state: User not authenticated, disconnecting WebSocket')
        // Disconnect WebSocket
        wsManager.disconnect()
        
        // 只有当确实存在token但认证失败时才触发logout事件
        // 这避免了页面刷新或初始化时的误触发
        const hasToken = TokenManager.get()
        if (hasToken && !isAuthenticated) {
          console.log('🔐 Token exists but user not authenticated, triggering logout')
          const logoutEvent = new CustomEvent('auth:logout')
          window.dispatchEvent(logoutEvent)
        }
      }
    }, 100) // 100ms防抖，给状态变化一些时间稳定

    return () => clearTimeout(timeoutId)
  }, [isAuthenticated, user, isLoading])

  // Enhanced login with optimistic updates
  const login = useCallback(async (data: LoginRequest) => {
    clearError()
    
    const result = await storeLogin(data)
    if (!result.success) {
      throw new Error(result.error || 'Login failed')
    }
  }, [storeLogin, clearError])

  // Enhanced register with error handling
  const register = useCallback(async (data: RegisterRequest): Promise<{ success: boolean; message?: string }> => {
    // TODO: Implement register in store
    return { success: false, message: 'Register not implemented yet' }
  }, [])

  // Enhanced logout
  const logout = useCallback(async () => {
    await storeLogout()
    TokenManager.clear()
  }, [storeLogout])

  // Enhanced refresh with optimistic updates
  const refreshUser = useCallback(async () => {
    await storeRefreshUser()
  }, [storeRefreshUser])

  // Legacy permission check
  const checkPermission = useCallback((permission: Permission): boolean => {
    return hasPermission(permission)
  }, [hasPermission])

  // Legacy role check
  const hasRoleCheck = useCallback((role: UserRole): boolean => {
    return hasRole(role)
  }, [hasRole])

  // Enhanced profile update with optimistic updates
  const updateProfile = useCallback(async (data: Partial<User>) => {
    if (!user) throw new Error('No user to update')

    // Use optimistic update from store
    const { optimisticUpdate } = useUserStore.getState()
    
    await optimisticUpdate(
      // Optimistic update function
      () => {
        updateUser(data)
      },
      // Async action
      async () => {
        // TODO: Implement actual API call
        // const response = await AuthService.updateProfile(data)
        // return response
        return data
      },
      // Rollback function (optional - will use automatic rollback)
      undefined
    )
  }, [user, updateUser])

  // Auto-refresh user data on mount
  useEffect(() => {
    const token = TokenManager.get()
    if (token && !TokenManager.isExpired(token) && !user) {
      refreshUser()
    }
  }, [refreshUser, user])

  // Error handling
  useEffect(() => {
    if (error) {
      log.error('Auth error', error, 'AuthProvider')
    }
  }, [error])

  const contextValue: AuthContextType = {
    user,
    isLoading,
    isAuthenticated,
    permissions,
    login,
    register,
    logout,
    refreshUser,
    checkPermission,
    hasRole: hasRoleCheck,
    updateProfile
  }

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

export default AuthContext