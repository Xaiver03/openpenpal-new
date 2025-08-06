/**
 * API Integration Tests - Authentication
 * API集成测试 - 认证相关
 */

import { NextRequest } from 'next/server'
import { POST as loginHandler } from '../auth/login/route'
import { GET as csrfHandler } from '../auth/csrf/route'

// Mock environment variables
Object.defineProperty(process.env, 'NODE_ENV', { value: 'test', writable: true })
process.env.JWT_SECRET = 'test-secret-key'

// Mock external dependencies
jest.mock('@/lib/redis/redis-client', () => ({
  RedisClient: {
    get: jest.fn(),
    set: jest.fn(),
    del: jest.fn(),
    incr: jest.fn(),
    expire: jest.fn()
  },
  SessionManager: {
    createSession: jest.fn(),
    getSession: jest.fn(),
    deleteSession: jest.fn()
  }
}))

jest.mock('@/lib/auth/test-data-manager', () => ({
  TestDataManager: {
    initializeTestUsers: jest.fn().mockResolvedValue({
      'testuser': {
        id: 'test-user-id',
        username: 'testuser',
        email: 'test@example.com',
        passwordHash: '$2b$10$mock.hash.value',
        role: 'user',
        permissions: ['user_write_letter'],
        school_code: 'TEST01',
        school_name: 'Test School',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z'
      }
    })
  }
}))

jest.mock('@/lib/auth/jwt-utils', () => ({
  JWTUtils: {
    generateTokenPair: jest.fn().mockReturnValue({
      accessToken: 'mock-access-token',
      refreshToken: 'mock-refresh-token',
      expiresIn: 3600
    })
  },
  PasswordUtils: {
    comparePassword: jest.fn()
  },
  SecurityUtils: {
    validateUsername: jest.fn().mockReturnValue({ isValid: true }),
    generateSessionId: jest.fn().mockReturnValue('mock-session-id')
  }
}))

jest.mock('@/lib/security/monitoring', () => ({
  securityMonitor: {
    logEvent: jest.fn()
  },
  SecurityEventType: {
    INVALID_INPUT: 'invalid_input',
    LOGIN_FAILED: 'login_failed',
    LOGIN_SUCCESS: 'login_success'
  },
  SecurityEventSeverity: {
    LOW: 'low',
    MEDIUM: 'medium'
  }
}))

// Global variable for mock requests
let mockRequest: NextRequest

describe('Authentication API Integration Tests', () => {
  describe('POST /api/auth/login', () => {

    beforeEach(() => {
      jest.clearAllMocks()
      
      // Mock CSRF validation to pass
      jest.doMock('@/lib/security/csrf', () => ({
        CSRFServer: {
          validate: jest.fn().mockReturnValue(true)
        }
      }))

      // Mock rate limiting to pass
      jest.doMock('@/lib/security/rate-limit', () => ({
        rateLimiters: {
          auth: {
            middleware: () => (req: any, handler: any) => handler(req)
          }
        }
      }))
    })

    test('successful login with valid credentials', async () => {
      const { PasswordUtils } = require('@/lib/auth/jwt-utils')
      PasswordUtils.comparePassword.mockResolvedValue(true)

      mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({
          username: 'testuser',
          password: 'validpassword'
        }),
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'valid-token',
          'Cookie': 'csrf-token=valid-token'
        }
      })

      const response = await loginHandler(mockRequest)
      const data = await response.json()

      expect(response.status).toBe(200)
      expect(data.code).toBe(0)
      expect(data.message).toBe('登录成功')
      expect(data.data).toHaveProperty('accessToken')
      expect(data.data).toHaveProperty('refreshToken')
      expect(data.data).toHaveProperty('user')
      expect(data.data.user).not.toHaveProperty('passwordHash')
    })

    test('login fails with invalid credentials', async () => {
      const { PasswordUtils } = require('@/lib/auth/jwt-utils')
      PasswordUtils.comparePassword.mockResolvedValue(false)

      mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({
          username: 'testuser',
          password: 'wrongpassword'
        }),
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'valid-token',
          'Cookie': 'csrf-token=valid-token'
        }
      })

      const response = await loginHandler(mockRequest)
      const data = await response.json()

      expect(response.status).toBe(401)
      expect(data.code).toBe(401)
      expect(data.message).toBe('用户名或密码错误')
    })

    test('login fails with missing credentials', async () => {
      mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({
          username: '',
          password: ''
        }),
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'valid-token',
          'Cookie': 'csrf-token=valid-token'
        }
      })

      const response = await loginHandler(mockRequest)
      const data = await response.json()

      expect(response.status).toBe(400)
      expect(data.code).toBe(400)
      expect(data.message).toBe('用户名和密码不能为空')
    })

    test('login fails with invalid CSRF token', async () => {
      // Mock CSRF validation to fail
      jest.doMock('@/lib/security/csrf', () => ({
        CSRFServer: {
          validate: jest.fn().mockReturnValue(false)
        }
      }))

      mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({
          username: 'testuser',
          password: 'validpassword'
        }),
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF 토큰': 'invalid-token'
        }
      })

      const response = await loginHandler(mockRequest)
      const data = await response.json()

      expect(response.status).toBe(403)
      expect(data.code).toBe('CSRF_VALIDATION_FAILED')
    })

    test('login fails for non-existent user', async () => {
      mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({
          username: 'nonexistentuser',
          password: 'password'
        }),
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'valid-token',
          'Cookie': 'csrf-token=valid-token'
        }
      })

      const response = await loginHandler(mockRequest)
      const data = await response.json()

      expect(response.status).toBe(401)
      expect(data.code).toBe(401)
      expect(data.message).toBe('用户名或密码错误')
    })

    test('login handles server errors gracefully', async () => {
      const { PasswordUtils } = require('@/lib/auth/jwt-utils')
      PasswordUtils.comparePassword.mockRejectedValue(new Error('Database error'))

      mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({
          username: 'testuser',
          password: 'validpassword'
        }),
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'valid-token',
          'Cookie': 'csrf-token=valid-token'
        }
      })

      const response = await loginHandler(mockRequest)
      const data = await response.json()

      expect(response.status).toBe(500)
      expect(data.code).toBe(500)
      expect(data.message).toBe('登录失败，请稍后重试')
    })

    test('security monitoring logs events correctly', async () => {
      const { securityMonitor } = require('@/lib/security/monitoring')
      const { PasswordUtils } = require('@/lib/auth/jwt-utils')
      PasswordUtils.comparePassword.mockResolvedValue(true)

      mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({
          username: 'testuser',
          password: 'validpassword'
        }),
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'valid-token',
          'Cookie': 'csrf-token=valid-token'
        }
      })

      await loginHandler(mockRequest)

      expect(securityMonitor.logEvent).toHaveBeenCalledWith(
        'login_success',
        'low',
        { username: 'testuser', role: 'user' },
        expect.objectContaining({
          request: mockRequest,
          username: 'testuser',
          userId: 'test-user-id',
          action: 'login'
        })
      )
    })
  })

  describe('GET /api/auth/csrf', () => {
    test('generates and returns CSRF token', async () => {
      const mockRequest = new NextRequest('http://localhost:3000/api/auth/csrf', {
        method: 'GET'
      })

      const response = await csrfHandler()
      const data = await response.json()

      expect(response.status).toBe(200)
      expect(data.success).toBe(true)
      expect(data.data).toHaveProperty('token')
      expect(data.data).toHaveProperty('expiresIn', 86400)
      expect(typeof data.data.token).toBe('string')
      expect(data.data.token.length).toBeGreaterThan(0)
    })

    test('returns existing CSRF token if available', async () => {
      // This test would require mocking the cookie functionality
      // For now, we'll test that the endpoint responds correctly
      const mockRequest = new NextRequest('http://localhost:3000/api/auth/csrf', {
        method: 'GET'
      })

      const response = await csrfHandler()
      const data = await response.json()

      expect(response.status).toBe(200)
      expect(data.success).toBe(true)
    })
  })

  describe('Rate Limiting Integration', () => {
    test('login endpoint applies rate limiting', async () => {
      // Mock rate limiter to simulate limit exceeded
      jest.doMock('@/lib/security/rate-limit', () => ({
        rateLimiters: {
          auth: {
            middleware: () => (req: any, handler: any) => 
              Promise.resolve(new Response(
                JSON.stringify({
                  success: false,
                  error: 'Rate limit exceeded',
                  code: 'RATE_LIMIT_EXCEEDED'
                }),
                { 
                  status: 429,
                  headers: { 'Content-Type': 'application/json' }
                }
              ))
          }
        }
      }))

      const mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({
          username: 'testuser',
          password: 'validpassword'
        }),
        headers: {
          'Content-Type': 'application/json'
        }
      })

      const response = await loginHandler(mockRequest)
      const data = await response.json()

      expect(response.status).toBe(429)
      expect(data.code).toBe('RATE_LIMIT_EXCEEDED')
    })
  })

  describe('Error Handling', () => {
    test('handles malformed JSON gracefully', async () => {
      mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
        method: 'POST',
        body: 'invalid json',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'valid-token',
          'Cookie': 'csrf-token=valid-token'
        }
      })

      const response = await loginHandler(mockRequest)
      
      expect(response.status).toBe(500)
    })

    test('handles missing request body', async () => {
      mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'valid-token',
          'Cookie': 'csrf-token=valid-token'
        }
      })

      const response = await loginHandler(mockRequest)
      
      expect(response.status).toBe(500)
    })
  })

  describe('Gateway Integration', () => {
    test('fallback authentication when gateway unavailable', async () => {
      // Mock fetch to simulate gateway unavailability
      global.fetch = jest.fn().mockRejectedValue(new Error('Gateway unavailable'))

      const { PasswordUtils } = require('@/lib/auth/jwt-utils')
      PasswordUtils.comparePassword.mockResolvedValue(true)

      mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({
          username: 'testuser',
          password: 'validpassword'
        }),
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'valid-token',
          'Cookie': 'csrf-token=valid-token'
        }
      })

      const response = await loginHandler(mockRequest)
      const data = await response.json()

      // Should still succeed using fallback authentication
      expect(response.status).toBe(200)
      expect(data.code).toBe(0)
    })
  })
})

describe('API Response Format', () => {
  test('all API responses follow standard format', async () => {
    const mockRequest = new NextRequest('http://localhost:3000/api/auth/csrf', {
      method: 'GET'
    })

    const response = await csrfHandler()
    const data = await response.json()

    // Should follow StandardApiResponse format
    expect(data).toHaveProperty('success')
    expect(data).toHaveProperty('data')
    expect(typeof data.success).toBe('boolean')
  })

  test('error responses include proper error codes', async () => {
    mockRequest = new NextRequest('http://localhost:3000/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({}),
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': 'valid-token',
        'Cookie': 'csrf-token=valid-token'
      }
    })

    const response = await loginHandler(mockRequest)
    const data = await response.json()

    expect(data).toHaveProperty('code')
    expect(data).toHaveProperty('message')
    expect(typeof data.code).toBe('number')
    expect(typeof data.message).toBe('string')
  })
})