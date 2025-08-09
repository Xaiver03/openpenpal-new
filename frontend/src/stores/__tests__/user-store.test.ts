/**
 * Unit tests for user store
 * 用户状态管理单元测试
 */

import { renderHook, act } from '@testing-library/react'
import { useUserStore, useAuth, usePermissions, useUser } from '../user-store'
import { AuthService } from '@/lib/services/auth-service'

// Mock dependencies
jest.mock('@/lib/services/auth-service', () => ({
  AuthService: {
    login: jest.fn(),
    logout: jest.fn(),
    refreshToken: jest.fn(),
    getCurrentUser: jest.fn()
  }
}))

jest.mock('@/constants/roles', () => ({
  hasPermission: jest.fn(),
  canAccessAdmin: jest.fn(),
  getCourierLevelName: jest.fn().mockReturnValue('一级信使'),
  type: {
    UserRole: {},
    Permission: {}
  }
}))

describe('useUserStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useUserStore.getState().reset()
    jest.clearAllMocks()
  })

  test('initial state is correct', () => {
    const { result } = renderHook(() => useUserStore())

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.loading.isLoading).toBe(false)
    expect(result.current.loading.isRefreshing).toBe(false)
    expect(result.current.loading.error).toBeNull()
    // permissionsCache removed - no longer exists in store
  })

  test('setUser updates user state correctly', () => {
    const { result } = renderHook(() => useUserStore())
    
    const mockUser = {
      id: '1',
      username: 'testuser',
      nickname: 'Test User',
      email: 'test@example.com',
      role: 'user' as const,
      school_code: 'TEST01',
      school_name: 'Test School',
      avatar: '',
      bio: '',
      address: '',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      last_login_at: '2024-01-01T12:00:00Z',
      status: 'active' as const,
      is_active: true,
      permissions: ['WRITE_LETTER' as const]
    }

    act(() => {
      result.current.setUser(mockUser)
    })

    expect(result.current.user).toEqual(mockUser)
    expect(result.current.isAuthenticated).toBe(true)
  })

  // updateUser test removed - method no longer exists in store

  test('updateCourierInfo updates courier information', () => {
    const { result } = renderHook(() => useUserStore())
    
    const mockUser = {
      id: '1',
      username: 'courier_level1',
      nickname: 'Level 1 Courier',
      email: 'courier1@example.com',
      role: 'courier_level1' as const,
      school_code: 'TEST01',
      school_name: 'Test School',
      avatar: '',
      bio: '',
      address: '',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      last_login_at: '2024-01-01T12:00:00Z',
      status: 'active' as const,
      is_active: true,
      permissions: ['COURIER_SCAN_CODE' as const]
    }

    act(() => {
      result.current.setUser(mockUser)
    })

    const courierInfo = {
      level: 2 as const,
      zoneCode: 'ZONE_A',
      zoneType: 'school' as const,
      status: 'active' as const,
      points: 100,
      taskCount: 5,
      completedTasks: 4,
      averageRating: 4.5,
      lastActiveAt: '2024-01-01T15:00:00Z'
    }

    act(() => {
      result.current.updateCourierInfo(courierInfo)
    })

    expect(result.current.user?.courierInfo).toEqual(courierInfo)
  })

  test('login handles successful authentication', async () => {
    const mockAuthService = AuthService as jest.Mocked<typeof AuthService>
    mockAuthService.login.mockResolvedValue({
      success: true,
      code: 0,
      message: 'Login successful',
      timestamp: new Date().toISOString(),
      data: {
        user: {
          id: '1',
          username: 'testuser',
          nickname: 'Test User',
          email: 'test@example.com',
          role: 'user',
          school_code: 'TEST01',
          school_name: 'Test School',
          is_active: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
          status: 'active' as const,
          permissions: ['WRITE_LETTER']
        },
        token: 'mock-token',
        refreshToken: 'mock-refresh-token',
        expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
      }
    })

    const { result } = renderHook(() => useUserStore())

    let loginResult: any
    await act(async () => {
      loginResult = await result.current.login({
        username: 'testuser',
        password: 'password'
      })
    })

    expect(loginResult.success).toBe(true)
    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.user?.username).toBe('testuser')
  })

  test('login handles authentication failure', async () => {
    const mockAuthService = AuthService as jest.Mocked<typeof AuthService>
    mockAuthService.login.mockResolvedValue({
      success: false,
      code: 401,
      message: 'Invalid credentials',
      timestamp: new Date().toISOString(),
      data: null
    })

    const { result } = renderHook(() => useUserStore())

    let loginResult: any
    await act(async () => {
      loginResult = await result.current.login({
        username: 'testuser',
        password: 'wrongpassword'
      })
    })

    expect(loginResult.success).toBe(false)
    expect(loginResult.error).toBe('Invalid credentials')
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.user).toBeNull()
  })

  test('logout clears user state', async () => {
    const mockAuthService = AuthService as jest.Mocked<typeof AuthService>
    mockAuthService.logout.mockResolvedValue(undefined)

    const { result } = renderHook(() => useUserStore())

    // Set initial user
    act(() => {
      result.current.setUser({
        id: '1',
        username: 'testuser',
        nickname: 'Test User',
        email: 'test@example.com',
        role: 'user',
        school_code: 'TEST01',
        school_name: 'Test School',
        avatar: '',
        bio: '',
        address: '',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        status: 'active',
        permissions: ['WRITE_LETTER']
      })
    })

    expect(result.current.isAuthenticated).toBe(true)

    await act(async () => {
      await result.current.logout()
    })

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    // permissionsCache removed - no longer exists in store
  })

  test('refreshUser updates user data', async () => {
    const mockAuthService = AuthService as jest.Mocked<typeof AuthService>
    mockAuthService.getCurrentUser.mockResolvedValue({
      success: true,
      data: {
        id: '1',
        username: 'testuser',
        nickname: 'Updated Nickname',
        email: 'updated@example.com',
        role: 'user',
        school_code: 'TEST01',
        school_name: 'Test School',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        status: 'active' as const,
        permissions: ['WRITE_LETTER', 'READ_LETTER']
      }
    } as any)

    const { result } = renderHook(() => useUserStore())

    // Set initial user
    act(() => {
      result.current.setUser({
        id: '1',
        username: 'testuser',
        nickname: 'Old Nickname',
        email: 'old@example.com',
        role: 'user',
        school_code: 'TEST01',
        school_name: 'Test School',
        avatar: '',
        bio: '',
        address: '',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        status: 'active',
        permissions: ['WRITE_LETTER']
      })
    })

    await act(async () => {
      await result.current.refreshUser()
    })

    expect(result.current.user?.nickname).toBe('Updated Nickname')
    expect(result.current.user?.email).toBe('updated@example.com')
    expect(result.current.user?.permissions).toEqual(['WRITE_LETTER', 'READ_LETTER'])
  })

  test('setLoading updates loading state', () => {
    const { result } = renderHook(() => useUserStore())

    act(() => {
      result.current.setLoading({
        isLoading: true,
        error: 'Test error'
      })
    })

    expect(result.current.loading.isLoading).toBe(true)
    expect(result.current.loading.error).toBe('Test error')
  })

  test('clearError clears error state', () => {
    const { result } = renderHook(() => useUserStore())

    act(() => {
      result.current.setLoading({ error: 'Test error' })
    })

    expect(result.current.loading.error).toBe('Test error')

    act(() => {
      result.current.clearError()
    })

    expect(result.current.loading.error).toBeNull()
  })

  test('reset clears all state', () => {
    const { result } = renderHook(() => useUserStore())

    // Set some state
    act(() => {
      result.current.setUser({
        id: '1',
        username: 'testuser',
        nickname: 'Test User',
        email: 'test@example.com',
        role: 'user',
        school_code: 'TEST01',
        school_name: 'Test School',
        avatar: '',
        bio: '',
        address: '',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        status: 'active',
        permissions: ['WRITE_LETTER']
      })
      result.current.setLoading({ isLoading: true, error: 'Test error' })
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.loading.isLoading).toBe(true)

    act(() => {
      result.current.reset()
    })

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.loading.isLoading).toBe(false)
    expect(result.current.loading.error).toBeNull()
  })
})

describe('useAuth hook', () => {
  beforeEach(() => {
    useUserStore.getState().reset()
    jest.clearAllMocks()
  })

  test('returns auth-related state and methods', () => {
    const { result } = renderHook(() => useAuth())

    expect(result.current).toHaveProperty('user')
    expect(result.current).toHaveProperty('isAuthenticated')
    expect(result.current).toHaveProperty('isLoading')
    expect(result.current).toHaveProperty('error')
    expect(result.current).toHaveProperty('login')
    expect(result.current).toHaveProperty('logout')
    expect(result.current).toHaveProperty('refreshUser')
    expect(result.current).toHaveProperty('clearError')
  })

  test('reflects store state changes', () => {
    const { result } = renderHook(() => useAuth())
    const { result: storeResult } = renderHook(() => useUserStore())

    expect(result.current.isAuthenticated).toBe(false)

    act(() => {
      storeResult.current.setUser({
        id: '1',
        username: 'testuser',
        nickname: 'Test User',
        email: 'test@example.com',
        role: 'user',
        school_code: 'TEST01',
        school_name: 'Test School',
        avatar: '',
        bio: '',
        address: '',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        status: 'active',
        permissions: ['WRITE_LETTER']
      })
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.user?.username).toBe('testuser')
  })
})

describe('usePermissions hook', () => {
  beforeEach(() => {
    useUserStore.getState().reset()
    jest.clearAllMocks()
  })

  test('returns permission-related methods', () => {
    const { result } = renderHook(() => usePermissions())

    expect(result.current).toHaveProperty('hasPermission')
    expect(result.current).toHaveProperty('hasRole')
    expect(result.current).toHaveProperty('canAccessAdmin')
    expect(result.current).toHaveProperty('isCourier')
    expect(result.current).toHaveProperty('getCourierLevel')
    expect(result.current).toHaveProperty('getCourierLevelName')
  })

  test('hasPermission works correctly', () => {
    const { hasPermission: mockHasPermission } = require('@/constants/roles')
    mockHasPermission.mockReturnValue(true)

    const { result } = renderHook(() => usePermissions())
    const { result: storeResult } = renderHook(() => useUserStore())

    act(() => {
      storeResult.current.setUser({
        id: '1',
        username: 'testuser',
        nickname: 'Test User',
        email: 'test@example.com',
        role: 'user',
        school_code: 'TEST01',
        school_name: 'Test School',
        avatar: '',
        bio: '',
        address: '',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        status: 'active',
        permissions: ['WRITE_LETTER']
      })
    })

    const hasPermissionResult = result.current.hasPermission('WRITE_LETTER' as any)
    expect(hasPermissionResult).toBe(true)
  })

  test('isCourier returns correct value', () => {
    const { result } = renderHook(() => usePermissions())
    const { result: storeResult } = renderHook(() => useUserStore())

    // Test with non-courier user
    act(() => {
      storeResult.current.setUser({
        id: '1',
        username: 'testuser',
        role: 'user',
        permissions: []
      } as any)
    })

    expect(result.current.isCourier()).toBe(false)

    // Test with courier user
    act(() => {
      storeResult.current.setUser({
        id: '1',
        username: 'courier',
        role: 'courier',
        permissions: []
      } as any)
    })

    expect(result.current.isCourier()).toBe(true)
  })

  test('getCourierLevel returns correct level', () => {
    const { result } = renderHook(() => usePermissions())
    const { result: storeResult } = renderHook(() => useUserStore())

    act(() => {
      storeResult.current.setUser({
        id: '1',
        username: 'courier',
        role: 'courier',
        permissions: [],
        courierInfo: {
          level: 2,
          zoneCode: 'ZONE_A',
          zoneType: 'school',
          status: 'active',
          points: 100,
          taskCount: 5,
          completedTasks: 4,
          averageRating: 4.5,
          lastActiveAt: '2024-01-01T15:00:00Z'
        }
      } as any)
    })

    expect(result.current.getCourierLevel()).toBe(2)
  })
})

describe('useUser hook', () => {
  beforeEach(() => {
    useUserStore.getState().reset()
    jest.clearAllMocks()
  })

  test('returns user-related state and methods', () => {
    const { result } = renderHook(() => useUser())

    expect(result.current).toHaveProperty('user')
    // updateUser method removed from store interface
    expect(result.current).toHaveProperty('updateCourierInfo')
  })

  // updateUser test removed - method no longer exists in store
})