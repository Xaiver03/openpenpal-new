'use client'

import { createContext, useContext, useEffect, useState, ReactNode, useCallback } from 'react'
import { AuthService as SimpleAuthService } from '@/lib/services/auth-service-simple'
import AuthService, { 
  type LoginRequest, 
  type RegisterRequest, 
  type User,
  type AuthContextData
} from '@/lib/services/auth-service'
import { type Permission } from '@/constants/roles'
import { TokenManager, wsManager } from '@/lib/api-client'

interface AuthContextType extends AuthContextData {
  login: (data: LoginRequest) => Promise<void>
  register: (data: RegisterRequest) => Promise<{ success: boolean; message?: string }>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
  checkPermission: (permission: string) => boolean
  hasRole: (role: string) => boolean
  updateProfile: (data: Partial<User>) => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [permissions, setPermissions] = useState<Permission[]>([])

  const isAuthenticated = !!user

  // 检查权限
  const checkPermission = useCallback((permission: string): boolean => {
    return permissions.includes(permission as Permission)
  }, [permissions])

  // 检查角色
  const hasRole = useCallback((role: string): boolean => {
    return user?.role === role
  }, [user?.role])

  // 更新用户信息
  const updateProfile = useCallback(async (data: Partial<User>) => {
    try {
      const response = await AuthService.updateProfile(data)
      if (response.success && response.data) {
        setUser(response.data)
      }
    } catch (error) {
      console.error('Failed to update profile:', error)
      throw error
    }
  }, [])

  // 刷新用户信息
  const refreshUser = useCallback(async () => {
    try {
      const [userResponse, permissionsResponse] = await Promise.all([
        AuthService.getCurrentUser(),
        AuthService.getUserPermissions()
      ])

      if (userResponse.success && userResponse.data) {
        setUser(userResponse.data)
      }

      if (permissionsResponse.success && permissionsResponse.data) {
        setPermissions(permissionsResponse.data.permissions as Permission[])
      }
    } catch (error) {
      console.error('Failed to refresh user:', error)
      throw error
    }
  }, [])

  // 初始化认证状态
  useEffect(() => {
    let isMounted = true

    const initAuth = async () => {
      try {
        console.log('🐛 AuthContext: Starting initAuth')
        const currentUser = await AuthService.autoLogin()
        console.log('🐛 AuthContext: autoLogin result:', currentUser)
        
        if (isMounted && currentUser) {
          console.log('🐛 AuthContext: Setting user:', currentUser)
          setUser(currentUser)
          // 获取权限
          try {
            const permissionsResponse = await AuthService.getUserPermissions()
            console.log('🐛 AuthContext: Permissions response:', permissionsResponse)
            if (permissionsResponse.success && permissionsResponse.data) {
              console.log('🐛 AuthContext: Setting permissions:', permissionsResponse.data.permissions)
              setPermissions(permissionsResponse.data.permissions as Permission[])
            }
          } catch (error) {
            console.error('Failed to load permissions:', error)
          }
          // 初始化WebSocket连接
          try {
            await wsManager.connect()
          } catch (error) {
            console.error('Failed to connect WebSocket:', error)
          }
        } else {
          console.log('🐛 AuthContext: No user found or component unmounted')
        }
      } catch (error) {
        console.error('Failed to initialize auth:', error)
      } finally {
        if (isMounted) {
          setIsLoading(false)
        }
      }
    }

    initAuth()

    return () => {
      isMounted = false
    }
  }, [])

  // 监听认证事件
  useEffect(() => {
    const handleLogin = (event: CustomEvent) => {
      setUser(event.detail.user)
      refreshUser().catch(console.error)
    }

    const handleLogout = () => {
      setUser(null)
      setPermissions([])
      wsManager.disconnect()
    }

    const handleProfileUpdate = (event: CustomEvent) => {
      setUser(event.detail.user)
    }

    if (typeof window !== 'undefined') {
      window.addEventListener('auth:login', handleLogin as EventListener)
      window.addEventListener('auth:logout', handleLogout)
      window.addEventListener('auth:profile-updated', handleProfileUpdate as EventListener)
    }

    return () => {
      if (typeof window !== 'undefined') {
        window.removeEventListener('auth:login', handleLogin as EventListener)
        window.removeEventListener('auth:logout', handleLogout)
        window.removeEventListener('auth:profile-updated', handleProfileUpdate as EventListener)
      }
    }
  }, [refreshUser])

  const handleLogin = async (credentials: LoginRequest): Promise<void> => {
    setIsLoading(true)
    try {
      console.log('🔐 AuthContext: 使用完整CSRF + JWT认证...')
      
      // 使用完整的CSRF + JWT认证流程
      const response = await AuthService.login(credentials)
      console.log('🔐 Full CSRF + JWT auth successful')
      
      if (response.success && response.data) {
        console.log('🐛 AuthContext.handleLogin: Login response user:', response.data.user)
        setUser(response.data.user)
        setPermissions(response.data.user.permissions || [] as Permission[])
        
        // 连接WebSocket
        try {
          await wsManager.connect()
        } catch (error) {
          console.error('Failed to connect WebSocket after login:', error)
        }
        
        // Login successful, no need to return anything
      } else {
        throw new Error(response.message || 'Login failed')
      }
    } catch (error) {
      console.error('Login error:', error)
      throw error
    } finally {
      setIsLoading(false)
    }
  }

  const handleRegister = async (userData: RegisterRequest): Promise<{ success: boolean; message?: string }> => {
    try {
      const response = await AuthService.register(userData)
      if (response.success) {
        return { success: true, message: '注册成功！' }
      } else {
        return { success: false, message: response.message || '注册失败' }
      }
    } catch (error: any) {
      console.error('Registration error:', error)
      return { 
        success: false, 
        message: error.message || '注册过程中发生错误，请稍后重试' 
      }
    }
  }

  const handleLogout = async () => {
    try {
      await AuthService.logout()
      setUser(null)
      setPermissions([])
      wsManager.disconnect()
    } catch (error) {
      console.error('Logout error:', error)
      // 即使后端登出失败，也要清除前端状态
      setUser(null)
      setPermissions([])
      wsManager.disconnect()
    }
  }

  const value: AuthContextType = {
    user,
    isLoading,
    isAuthenticated,
    permissions,
    login: handleLogin,
    register: handleRegister,
    logout: handleLogout,
    refreshUser,
    checkPermission,
    hasRole,
    updateProfile
  }

  return (
    <AuthContext.Provider value={value}>
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