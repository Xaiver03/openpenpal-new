/**
 * Test User Fixtures
 * 测试用户数据
 */

export interface TestUser {
  username: string
  password: string
  email: string
  role: 'user' | 'courier' | 'admin'
  nickname: string
  school_code: string
}

export const TEST_USERS: Record<string, TestUser> = {
  regularUser: {
    username: 'testuser',
    password: 'TestPass123!',
    email: 'testuser@example.com',
    role: 'user',
    nickname: '测试用户',
    school_code: 'BJDX01'
  },
  courier: {
    username: 'testcourier',
    password: 'CourierPass123!',
    email: 'courier@example.com',
    role: 'courier',
    nickname: '测试信使',
    school_code: 'BJDX01'
  },
  admin: {
    username: 'testadmin',
    password: 'AdminPass123!',
    email: 'admin@example.com',
    role: 'admin',
    nickname: '测试管理员',
    school_code: 'ADMIN'
  }
}

export const TEST_CREDENTIALS = {
  valid: {
    username: 'testuser',
    password: 'TestPass123!'
  },
  invalid: {
    username: 'invaliduser',
    password: 'wrongpassword'
  },
  empty: {
    username: '',
    password: ''
  }
}