/**
 * Dependency Injection Auth Context - Circular Dependency Free
 * 依赖注入认证上下文 - 无循环依赖版本
 * 
 * Purpose: Replace direct imports with dependency injection to break circular dependencies
 * 目的: 用依赖注入替代直接导入以打破循环依赖
 */

'use client'

import { createContext, useContext, useEffect, useState, useCallback, ReactNode } from 'react'
import { 
  ServiceRegistry,
  getAuthService,
  getUserStateService,
  getPermissionService,
  getTokenService
} from '@/lib/di/service-registry'
import { 
  IAuthService,
  IUserStateService,
  IPermissionService,
  ITokenService,
  User,
  LoginCredentials,
  RegisterRequest
} from '@/lib/di/service-interfaces'
import { type UserRole, type Permission } from '@/constants/roles'
import { log } from '@/utils/logger'

// ================================
// Legacy Interface Compatibility
// ================================

export interface AuthUser extends User {
  // Extend if needed for backward compatibility
}

export interface AuthContextType {
  // User state
  user: AuthUser | null
  isLoading: boolean
  isAuthenticated: boolean
  error: string | null
  
  // Permissions
  permissions: Permission[]
  checkPermission: (permission: Permission) => boolean
  hasRole: (role: UserRole) => boolean
  
  // Actions
  login: (data: LoginCredentials) => Promise<void>
  register: (data: RegisterRequest) => Promise<{ success: boolean; message?: string }>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
  updateProfile: (data: Partial<AuthUser>) => Promise<void>
  clearError: () => void
}

// ================================
// Context Creation
// ================================

const AuthContext = createContext<AuthContextType | undefined>(undefined)

// ================================
// Provider Implementation
// ================================

export function AuthProviderDI({ children }: { children: ReactNode }) {
  // ================================
  // State Management
  // ================================
  
  const [user, setUser] = useState<AuthUser | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)

  // ================================
  // Service Dependencies (via DI)
  // ================================
  
  const [services, setServices] = useState<{
    auth?: IAuthService
    userState?: IUserStateService  
    permission?: IPermissionService
    token?: ITokenService
  }>({})

  const [servicesReady, setServicesReady] = useState<boolean>(false)

  // ================================
  // Service Initialization
  // ================================

  useEffect(() => {
    const initializeServices = async () => {
      try {
        // Initialize service registry if not already done
        if (!ServiceRegistry.getStatus().initialized) {
          ServiceRegistry.initialize({
            enableDevtools: process.env.NODE_ENV === 'development',
            apiBaseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
          })
        }

        // Get services from DI container
        const authService = getAuthService()
        const userStateService = getUserStateService()
        const permissionService = getPermissionService()
        const tokenService = getTokenService()

        setServices({
          auth: authService,
          userState: userStateService,
          permission: permissionService,
          token: tokenService
        })

        setServicesReady(true)
        console.debug('🔧 AuthProviderDI: Services initialized via DI')

      } catch (error) {
        console.error('❌ AuthProviderDI: Service initialization failed:', error)
        setError('Service initialization failed')
        setServicesReady(false)
      }
    }

    initializeServices()
  }, [])

  // ================================
  // User State Synchronization
  // ================================

  useEffect(() => {
    if (!servicesReady || !services.userState) return

    // Subscribe to user state changes from DI service
    const unsubscribe = services.userState.subscribe((newUser) => {
      setUser(newUser as AuthUser)
      setIsLoading(false)
      
      if (newUser) {
        setError(null)
      }
    })

    return unsubscribe
  }, [servicesReady, services.userState])

  // ================================
  // Auto-restore User on Mount
  // ================================

  useEffect(() => {
    if (!servicesReady || !services.auth || !services.token) return

    const restoreUser = async () => {
      try {
        const token = services.token!.get()
        if (token && !services.token!.isExpired(token)) {
          console.debug('🔧 AuthProviderDI: Restoring user from valid token')
          await services.auth!.getCurrentUser()
        } else {
          setIsLoading(false)
        }
      } catch (error) {
        console.error('Failed to restore user:', error)
        setIsLoading(false)
      }
    }

    restoreUser()
  }, [servicesReady, services.auth, services.token])

  // ================================
  // Authentication Actions
  // ================================

  const login = useCallback(async (credentials: LoginCredentials) => {
    if (!services.auth) {
      throw new Error('Auth service not available')
    }

    setIsLoading(true)
    setError(null)

    try {
      const response = await services.auth.login(credentials)
      
      if (response.success) {
        console.debug('✅ AuthProviderDI: Login successful')
        // User state will be updated via subscription
      } else {
        throw new Error(response.message || 'Login failed')
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Login failed'
      setError(errorMessage)
      log.error('Login error', error, 'AuthProviderDI')
      throw error
    } finally {
      setIsLoading(false)
    }
  }, [services.auth])

  const logout = useCallback(async () => {
    if (!services.auth) {
      console.warn('Auth service not available for logout')
      return
    }

    setIsLoading(true)

    try {
      await services.auth.logout()
      console.debug('✅ AuthProviderDI: Logout successful')
      // User state will be cleared via subscription
    } catch (error) {
      console.error('Logout error:', error)
      // Clear local state even if API call fails
      if (services.userState) {
        services.userState.setUser(null)
      }
    } finally {
      setIsLoading(false)
    }
  }, [services.auth, services.userState])

  const register = useCallback(async (data: RegisterRequest): Promise<{ success: boolean; message?: string }> => {
    if (!services.auth?.register) {
      return { success: false, message: 'Register service not available' }
    }

    setIsLoading(true)
    setError(null)

    try {
      const response = await services.auth.register(data)
      return { 
        success: response.success, 
        message: response.message 
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Registration failed'
      setError(errorMessage)
      return { success: false, message: errorMessage }
    } finally {
      setIsLoading(false)
    }
  }, [services.auth])

  const refreshUser = useCallback(async () => {
    if (!services.auth) {
      console.warn('Auth service not available for refresh')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      await services.auth.refreshAuth()
      console.debug('✅ AuthProviderDI: User refresh successful')
    } catch (error) {
      console.error('Refresh user error:', error)
      setError('Failed to refresh user data')
    } finally {
      setIsLoading(false)
    }
  }, [services.auth])

  const updateProfile = useCallback(async (data: Partial<AuthUser>) => {
    if (!user || !services.userState) {
      throw new Error('No user to update or user state service unavailable')
    }

    try {
      // Optimistic update
      services.userState.updateUser(data)
      
      // TODO: Make actual API call to persist changes
      // const response = await services.api.updateProfile(data)
      
      console.debug('✅ AuthProviderDI: Profile update successful (optimistic)')
    } catch (error) {
      console.error('Update profile error:', error)
      // TODO: Rollback optimistic update on error
      throw error
    }
  }, [user, services.userState])

  const clearError = useCallback(() => {
    setError(null)
  }, [])

  // ================================
  // Permission Methods
  // ================================

  const checkPermission = useCallback((permission: Permission): boolean => {
    if (!user || !services.permission) return false
    return services.permission.hasPermission(user, permission)
  }, [user, services.permission])

  const hasRole = useCallback((role: UserRole): boolean => {
    return user?.role === role
  }, [user])

  // ================================
  // Computed Properties
  // ================================

  const isAuthenticated = !!user
  const permissions: Permission[] = user?.permissions || []

  // ================================
  // Context Value
  // ================================

  const contextValue: AuthContextType = {
    // User state
    user,
    isLoading,
    isAuthenticated,
    error,
    
    // Permissions
    permissions,
    checkPermission,
    hasRole,
    
    // Actions
    login,
    register,
    logout,
    refreshUser,
    updateProfile,
    clearError
  }

  // Don't render until services are ready
  if (!servicesReady) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"></div>
          <p className="text-sm text-muted-foreground">Initializing services...</p>
        </div>
      </div>
    )
  }

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  )
}

// ================================
// Hook for Using Auth Context
// ================================

export function useAuthDI(): AuthContextType {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuthDI must be used within an AuthProviderDI')
  }
  return context
}

// ================================
// Backward Compatibility Aliases
// ================================

export const AuthProvider = AuthProviderDI
export const useAuth = useAuthDI

// ================================
// Default Export
// ================================

export default AuthContext