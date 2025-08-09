/**
 * SOTA Type Guards and Validation Utilities
 * 
 * Provides comprehensive type-safe runtime validation for API responses,
 * user input, and internal data structures.
 */

import { UserRole, Permission, CourierLevel } from '@/constants/roles'
import type { User } from '@/types/user'
import type { ApiResponse } from '../api-client'

// ================================
// Core Type Guards
// ================================

/**
 * Type guard for checking if value is a valid UserRole
 */
export function isValidUserRole(role: any): role is UserRole {
  const validRoles: UserRole[] = [
    'user', 'courier_level1', 'courier_level2', 'courier_level3', 
    'courier_level4', 'platform_admin', 'super_admin'
  ]
  return typeof role === 'string' && validRoles.includes(role as UserRole)
}

/**
 * Type guard for checking if value is a valid CourierLevel
 */
export function isValidCourierLevel(level: any): level is CourierLevel {
  return typeof level === 'number' && level >= 1 && level <= 4 && Number.isInteger(level)
}

/**
 * Type guard for checking if value is a valid Permission
 */
export function isValidPermission(permission: any): permission is Permission {
  return typeof permission === 'string' && permission.length > 0
}

/**
 * Type guard for User object validation
 */
export function isValidUser(user: any): user is User {
  if (!user || typeof user !== 'object') return false
  
  const requiredFields = ['id', 'username', 'nickname', 'email', 'role']
  const hasRequiredFields = requiredFields.every(field => 
    field in user && user[field] !== null && user[field] !== undefined
  )
  
  return hasRequiredFields && 
    isValidUserRole(user.role) &&
    typeof user.id === 'string' &&
    typeof user.username === 'string' &&
    typeof user.email === 'string' &&
    typeof user.nickname === 'string'
}

/**
 * Type guard for API Response structure
 */
export function isValidApiResponse<T>(response: any): response is ApiResponse<T> {
  if (!response || typeof response !== 'object') return false
  
  return 'code' in response && 
    'message' in response && 
    'data' in response &&
    'timestamp' in response &&
    typeof response.code === 'number' &&
    typeof response.message === 'string' &&
    typeof response.timestamp === 'string'
}

// ================================
// Advanced Validation Functions
// ================================

/**
 * Validate and sanitize user input with comprehensive error reporting
 */
export interface ValidationResult<T> {
  isValid: boolean
  data?: T
  errors: ValidationError[]
  warnings: string[]
}

export interface ValidationError {
  field: string
  code: string
  message: string
  severity: 'error' | 'warning'
}

/**
 * Comprehensive user object validation with detailed error reporting
 */
export function validateUser(userData: any): ValidationResult<User> {
  const errors: ValidationError[] = []
  const warnings: string[] = []
  
  // Check required fields
  const requiredFields: (keyof User)[] = ['id', 'username', 'email', 'nickname', 'role']
  
  for (const field of requiredFields) {
    if (!userData[field]) {
      errors.push({
        field,
        code: 'REQUIRED_FIELD_MISSING',
        message: `Required field '${field}' is missing or empty`,
        severity: 'error'
      })
    }
  }
  
  // Validate specific field formats
  if (userData.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(userData.email)) {
    errors.push({
      field: 'email',
      code: 'INVALID_EMAIL_FORMAT',
      message: 'Invalid email format',
      severity: 'error'
    })
  }
  
  if (userData.role && !isValidUserRole(userData.role)) {
    errors.push({
      field: 'role',
      code: 'INVALID_ROLE',
      message: `Invalid role: ${userData.role}`,
      severity: 'error'
    })
  }
  
  // Validate optional courier info
  if (userData.courierInfo) {
    if (userData.courierInfo.level && !isValidCourierLevel(userData.courierInfo.level)) {
      errors.push({
        field: 'courierInfo.level',
        code: 'INVALID_COURIER_LEVEL',
        message: `Invalid courier level: ${userData.courierInfo.level}`,
        severity: 'error'
      })
    }
    
    // Validate role-level consistency
    const role = userData.role as UserRole
    const level = userData.courierInfo.level as CourierLevel
    
    if (role.startsWith('courier_level') && level) {
      const expectedLevel = parseInt(role.slice(-1))
      if (expectedLevel !== level) {
        warnings.push(`Role '${role}' doesn't match courier level '${level}'`)
      }
    }
  }
  
  // Validate timestamps
  const dateFields = ['created_at', 'updated_at', 'last_login_at']
  for (const field of dateFields) {
    if (userData[field] && !isValidISOString(userData[field])) {
      warnings.push(`Invalid date format for field '${field}': ${userData[field]}`)
    }
  }
  
  const isValid = errors.length === 0
  
  return {
    isValid,
    data: isValid ? userData as User : undefined,
    errors,
    warnings
  }
}

/**
 * Validate ISO date string
 */
export function isValidISOString(dateString: any): boolean {
  if (typeof dateString !== 'string') return false
  const date = new Date(dateString)
  return date instanceof Date && !isNaN(date.getTime()) && date.toISOString() === dateString
}

/**
 * Validate API response structure with detailed error reporting
 */
export function validateApiResponse<T>(
  response: any, 
  dataValidator?: (data: any) => boolean
): ValidationResult<ApiResponse<T>> {
  const errors: ValidationError[] = []
  const warnings: string[] = []
  
  // Check required response fields
  const requiredFields = ['code', 'message', 'data', 'timestamp']
  for (const field of requiredFields) {
    if (!(field in response)) {
      errors.push({
        field,
        code: 'REQUIRED_FIELD_MISSING',
        message: `Required API response field '${field}' is missing`,
        severity: 'error'
      })
    }
  }
  
  // Validate field types
  if (response.code !== undefined && typeof response.code !== 'number') {
    errors.push({
      field: 'code',
      code: 'INVALID_TYPE',
      message: 'API response code must be a number',
      severity: 'error'
    })
  }
  
  if (response.message !== undefined && typeof response.message !== 'string') {
    errors.push({
      field: 'message',
      code: 'INVALID_TYPE', 
      message: 'API response message must be a string',
      severity: 'error'
    })
  }
  
  // Validate success indicators
  if (response.code !== undefined) {
    const isSuccess = response.code === 0
    const hasSuccessField = response.success === true
    
    if (isSuccess !== hasSuccessField && response.success !== undefined) {
      warnings.push('Inconsistent success indicators between code and success field')
    }
  }
  
  // Validate data payload if validator provided
  if (dataValidator && response.data && !dataValidator(response.data)) {
    errors.push({
      field: 'data',
      code: 'INVALID_DATA_STRUCTURE',
      message: 'API response data failed validation',
      severity: 'error'
    })
  }
  
  // Validate timestamp
  if (response.timestamp && !isValidISOString(response.timestamp)) {
    warnings.push(`Invalid timestamp format: ${response.timestamp}`)
  }
  
  const isValid = errors.length === 0
  
  return {
    isValid,
    data: isValid ? response as ApiResponse<T> : undefined,
    errors,
    warnings
  }
}

// ================================
// Utility Functions
// ================================

/**
 * Safe JSON parsing with validation
 */
export function safeJsonParse<T>(
  jsonString: string,
  validator?: (data: any) => data is T
): { success: true; data: T } | { success: false; error: string } {
  try {
    const parsed = JSON.parse(jsonString)
    
    if (validator && !validator(parsed)) {
      return { success: false, error: 'Data failed validation after parsing' }
    }
    
    return { success: true, data: parsed }
  } catch (error) {
    return { 
      success: false, 
      error: error instanceof Error ? error.message : 'Unknown JSON parsing error' 
    }
  }
}

/**
 * Deep clone object with type safety
 */
export function safeClone<T>(obj: T): T {
  if (obj === null || typeof obj !== 'object') {
    return obj
  }
  
  if (obj instanceof Date) {
    return new Date(obj.getTime()) as unknown as T
  }
  
  if (Array.isArray(obj)) {
    return obj.map(item => safeClone(item)) as unknown as T
  }
  
  const cloned = {} as T
  for (const key in obj) {
    if (obj.hasOwnProperty(key)) {
      cloned[key] = safeClone(obj[key])
    }
  }
  
  return cloned
}

/**
 * Check if value is empty (null, undefined, empty string, empty array, empty object)
 */
export function isEmpty(value: any): boolean {
  if (value === null || value === undefined) return true
  if (typeof value === 'string') return value.trim().length === 0
  if (Array.isArray(value)) return value.length === 0
  if (typeof value === 'object') return Object.keys(value).length === 0
  return false
}

/**
 * Sanitize string input by removing dangerous characters
 */
export function sanitizeString(input: string, options: {
  allowHtml?: boolean
  maxLength?: number
  trimWhitespace?: boolean
} = {}): string {
  const { allowHtml = false, maxLength = 1000, trimWhitespace = true } = options
  
  let sanitized = input
  
  if (trimWhitespace) {
    sanitized = sanitized.trim()
  }
  
  if (!allowHtml) {
    sanitized = sanitized.replace(/<[^>]*>/g, '') // Remove HTML tags
  }
  
  if (maxLength && sanitized.length > maxLength) {
    sanitized = sanitized.substring(0, maxLength)
  }
  
  return sanitized
}

export default {
  isValidUserRole,
  isValidCourierLevel,
  isValidPermission,
  isValidUser,
  isValidApiResponse,
  validateUser,
  validateApiResponse,
  isValidISOString,
  safeJsonParse,
  safeClone,
  isEmpty,
  sanitizeString
}