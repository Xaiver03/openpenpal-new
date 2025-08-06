import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/utils/api'
import type { User, LoginForm } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const token = ref<string | null>(localStorage.getItem('admin_token'))
  const user = ref<User | null>(null)
  const permissions = ref<string[]>([])

  // 计算属性
  const isAuthenticated = computed(() => !!token.value && !!user.value)

  // 方法
  const login = async (loginForm: LoginForm) => {
    try {
      const response = await api.post('/auth/login', loginForm)
      
      if (response.data.code === 0) {
        const { token: newToken, user: userData } = response.data.data
        
        token.value = newToken
        user.value = userData
        permissions.value = userData.permissions || []
        
        localStorage.setItem('admin_token', newToken)
        localStorage.setItem('admin_user', JSON.stringify(userData))
        
        return { success: true }
      } else {
        return { success: false, message: response.data.msg }
      }
    } catch (error: any) {
      return { 
        success: false, 
        message: error.response?.data?.msg || '登录失败，请稍后重试'
      }
    }
  }

  const logout = () => {
    token.value = null
    user.value = null
    permissions.value = []
    
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
  }

  const initAuth = () => {
    const savedToken = localStorage.getItem('admin_token')
    const savedUser = localStorage.getItem('admin_user')
    
    if (savedToken && savedUser) {
      try {
        token.value = savedToken
        user.value = JSON.parse(savedUser)
        permissions.value = user.value?.permissions || []
      } catch (error) {
        logout()
      }
    }
  }

  const hasPermission = (permission: string): boolean => {
    if (!user.value) return false
    if (user.value.role === 'super_admin') return true
    return permissions.value.includes(permission)
  }

  const refreshUserInfo = async () => {
    try {
      const response = await api.get('/users/me')
      if (response.data.code === 0) {
        user.value = response.data.data
        permissions.value = response.data.data.permissions || []
        localStorage.setItem('admin_user', JSON.stringify(response.data.data))
      }
    } catch (error) {
      console.error('刷新用户信息失败:', error)
    }
  }

  // 初始化认证状态
  initAuth()

  return {
    token,
    user,
    permissions,
    isAuthenticated,
    login,
    logout,
    hasPermission,
    refreshUserInfo
  }
})