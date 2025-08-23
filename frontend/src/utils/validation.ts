/**
 * Comprehensive validation utilities for OpenPenPal
 * Implements client-side and server-side validation rules
 */

// Basic validation rules
export const ValidationRules = {
  // Required field
  required: (value: any) => {
    if (value === null || value === undefined || value === '') {
      return '此字段为必填项'
    }
    return null
  },

  // Email validation
  email: (value: string) => {
    if (!value) return null
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(value)) {
      return '请输入有效的邮箱地址'
    }
    return null
  },

  // Username validation
  username: (value: string) => {
    if (!value) return null
    if (value.length < 3) {
      return '用户名至少需要3个字符'
    }
    if (value.length > 20) {
      return '用户名不能超过20个字符'
    }
    const usernameRegex = /^[a-zA-Z0-9_\u4e00-\u9fa5]+$/
    if (!usernameRegex.test(value)) {
      return '用户名只能包含字母、数字、下划线和中文字符'
    }
    return null
  },

  // Password validation
  password: (value: string) => {
    if (!value) return null
    if (value.length < 6) {
      return '密码至少需要6个字符'
    }
    if (value.length > 50) {
      return '密码不能超过50个字符'
    }
    return null
  },

  // Strong password validation
  strongPassword: (value: string) => {
    if (!value) return null
    if (value.length < 8) {
      return '强密码至少需要8个字符'
    }
    const hasLower = /[a-z]/.test(value)
    const hasUpper = /[A-Z]/.test(value)
    const hasNumber = /\d/.test(value)
    const hasSpecial = /[!@#$%^&*(),.?":{}|<>]/.test(value)
    
    if (!hasLower || !hasUpper || !hasNumber || !hasSpecial) {
      return '密码必须包含大小写字母、数字和特殊字符'
    }
    return null
  },

  // Phone number validation
  phone: (value: string) => {
    if (!value) return null
    const phoneRegex = /^1[3-9]\d{9}$/
    if (!phoneRegex.test(value)) {
      return '请输入有效的手机号码'
    }
    return null
  },

  // Length validation
  minLength: (min: number) => (value: string) => {
    if (!value) return null
    if (value.length < min) {
      return `至少需要${min}个字符`
    }
    return null
  },

  maxLength: (max: number) => (value: string) => {
    if (!value) return null
    if (value.length > max) {
      return `不能超过${max}个字符`
    }
    return null
  },

  // Number validation
  number: (value: any) => {
    if (value === null || value === undefined || value === '') return null
    if (isNaN(Number(value))) {
      return '请输入有效的数字'
    }
    return null
  },

  // Range validation
  range: (min: number, max: number) => (value: number) => {
    if (value === null || value === undefined) return null
    if (value < min || value > max) {
      return `值必须在${min}和${max}之间`
    }
    return null
  },

  // School code validation
  schoolCode: (value: string) => {
    if (!value) return null
    const schoolCodeRegex = /^[A-Z0-9]{4,10}$/
    if (!schoolCodeRegex.test(value)) {
      return '学校代码必须是4-10位大写字母和数字组合'
    }
    return null
  },

  // Letter content validation
  letterContent: (value: string) => {
    if (!value) return '信件内容不能为空'
    if (value.length < 10) {
      return '信件内容至少需要10个字符'
    }
    if (value.length > 10000) {
      return '信件内容不能超过10000个字符'
    }
    return null
  },

  // Letter title validation
  letterTitle: (value: string) => {
    if (!value) return '信件标题不能为空'
    if (value.length < 1) {
      return '标题至少需要1个字符'
    }
    if (value.length > 200) {
      return '标题不能超过200个字符'
    }
    return null
  },

  // URL validation
  url: (value: string) => {
    if (!value) return null
    try {
      new URL(value)
      return null
    } catch {
      return '请输入有效的URL地址'
    }
  },

  // Date validation
  date: (value: string) => {
    if (!value) return null
    const date = new Date(value)
    if (isNaN(date.getTime())) {
      return '请输入有效的日期'
    }
    return null
  },

  // Future date validation
  futureDate: (value: string) => {
    if (!value) return null
    const date = new Date(value)
    if (isNaN(date.getTime())) {
      return '请输入有效的日期'
    }
    if (date <= new Date()) {
      return '日期必须是未来时间'
    }
    return null
  }
}

// Validation composer
export class ValidationComposer {
  private rules: Array<(value: any) => string | null> = []

  constructor(private fieldName: string = 'field') {}

  required() {
    this.rules.push(ValidationRules.required)
    return this
  }

  email() {
    this.rules.push(ValidationRules.email)
    return this
  }

  username() {
    this.rules.push(ValidationRules.username)
    return this
  }

  password() {
    this.rules.push(ValidationRules.password)
    return this
  }

  strongPassword() {
    this.rules.push(ValidationRules.strongPassword)
    return this
  }

  phone() {
    this.rules.push(ValidationRules.phone)
    return this
  }

  minLength(min: number) {
    this.rules.push(ValidationRules.minLength(min))
    return this
  }

  maxLength(max: number) {
    this.rules.push(ValidationRules.maxLength(max))
    return this
  }

  number() {
    this.rules.push(ValidationRules.number)
    return this
  }

  range(min: number, max: number) {
    this.rules.push(ValidationRules.range(min, max))
    return this
  }

  custom(rule: (value: any) => string | null) {
    this.rules.push(rule)
    return this
  }

  validate(value: any): string | null {
    for (const rule of this.rules) {
      const error = rule(value)
      if (error) return error
    }
    return null
  }
}

// Form validation schemas
export const FormSchemas = {
  // Login form
  login: {
    username: new ValidationComposer('用户名').required().username(),
    password: new ValidationComposer('密码').required().password()
  },

  // Registration form
  register: {
    username: new ValidationComposer('用户名').required().username(),
    email: new ValidationComposer('邮箱').required().email(),
    password: new ValidationComposer('密码').required().strongPassword(),
    nickname: new ValidationComposer('昵称').required().minLength(1).maxLength(50),
    schoolCode: new ValidationComposer('学校代码').required().custom(ValidationRules.schoolCode),
    phone: new ValidationComposer('手机号').custom(ValidationRules.phone)
  },

  // Letter form
  letter: {
    title: new ValidationComposer('标题').custom(ValidationRules.letterTitle),
    content: new ValidationComposer('内容').custom(ValidationRules.letterContent)
  },

  // Courier application form
  courierApplication: {
    coverageArea: new ValidationComposer('服务区域').required().minLength(2).maxLength(100),
    experience: new ValidationComposer('经验描述').maxLength(1000),
    reason: new ValidationComposer('申请理由').required().minLength(10).maxLength(500)
  },

  // Profile update form
  profileUpdate: {
    nickname: new ValidationComposer('昵称').required().minLength(1).maxLength(50),
    phone: new ValidationComposer('手机号').custom(ValidationRules.phone)
  }
}

// Validation utilities
export class FormValidator {
  static validateForm<T extends Record<string, any>>(
    data: T,
    schema: Record<keyof T, ValidationComposer>
  ): {
    isValid: boolean
    errors: Partial<Record<keyof T, string>>
  } {
    const errors: Partial<Record<keyof T, string>> = {}
    
    for (const [field, validator] of Object.entries(schema)) {
      const error = validator.validate(data[field])
      if (error) {
        errors[field as keyof T] = error
      }
    }
    
    return {
      isValid: Object.keys(errors).length === 0,
      errors
    }
  }

  static validateField<T>(
    value: T,
    validator: ValidationComposer
  ): string | null {
    return validator.validate(value)
  }

  static sanitizeInput(input: string): string {
    // Basic XSS prevention
    return input
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#x27;')
      .replace(/\//g, '&#x2F;')
      .trim()
  }

  static normalizeEmail(email: string): string {
    return email.toLowerCase().trim()
  }

  static normalizeUsername(username: string): string {
    return username.toLowerCase().trim()
  }
}

// File validation
export class FileValidator {
  static validateImage(file: File): string | null {
    const allowedTypes = ['image/jpeg', 'image/png', 'image/webp', 'image/gif']
    const maxSize = 5 * 1024 * 1024 // 5MB
    
    if (!allowedTypes.includes(file.type)) {
      return '只支持 JPEG、PNG、WebP 和 GIF 格式的图片'
    }
    
    if (file.size > maxSize) {
      return '图片大小不能超过5MB'
    }
    
    return null
  }

  static validateDocument(file: File): string | null {
    const allowedTypes = [
      'application/pdf',
      'application/msword',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      'text/plain'
    ]
    const maxSize = 10 * 1024 * 1024 // 10MB
    
    if (!allowedTypes.includes(file.type)) {
      return '只支持 PDF、Word 和文本文档'
    }
    
    if (file.size > maxSize) {
      return '文档大小不能超过10MB'
    }
    
    return null
  }
}

// Security validation
export class SecurityValidator {
  static checkPasswordStrength(password: string): {
    score: number
    feedback: string[]
  } {
    const feedback: string[] = []
    let score = 0
    
    if (password.length >= 8) {
      score += 1
    } else {
      feedback.push('密码至少需要8个字符')
    }
    
    if (/[a-z]/.test(password)) {
      score += 1
    } else {
      feedback.push('需要包含小写字母')
    }
    
    if (/[A-Z]/.test(password)) {
      score += 1
    } else {
      feedback.push('需要包含大写字母')
    }
    
    if (/\d/.test(password)) {
      score += 1
    } else {
      feedback.push('需要包含数字')
    }
    
    if (/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
      score += 1
    } else {
      feedback.push('需要包含特殊字符')
    }
    
    // Check for common patterns
    if (/(.)\1{2,}/.test(password)) {
      score -= 1
      feedback.push('避免使用重复字符')
    }
    
    if (/123|abc|qwe/i.test(password)) {
      score -= 1
      feedback.push('避免使用连续字符')
    }
    
    return { score: Math.max(0, score), feedback }
  }

  static detectSuspiciousContent(content: string): string[] {
    const flags: string[] = []
    
    // Check for potential XSS
    if (/<script|javascript:|on\w+=/i.test(content)) {
      flags.push('检测到可疑脚本内容')
    }
    
    // Check for SQL injection patterns
    if (/(union|select|insert|delete|drop|update)\s+/i.test(content)) {
      flags.push('检测到可疑SQL语句')
    }
    
    // Check for excessive special characters
    const specialCharRatio = (content.match(/[^a-zA-Z0-9\u4e00-\u9fa5\s]/g) || []).length / content.length
    if (specialCharRatio > 0.3) {
      flags.push('特殊字符比例过高')
    }
    
    return flags
  }
}

// Async validation
export class AsyncValidator {
  static async checkUsernameAvailability(username: string): Promise<boolean> {
    try {
      const response = await fetch('/api/auth/check-username', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username })
      })
      const data = await response.json()
      return data.available
    } catch {
      return false
    }
  }

  static async checkEmailAvailability(email: string): Promise<boolean> {
    try {
      const response = await fetch('/api/auth/check-email', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email })
      })
      const data = await response.json()
      return data.available
    } catch {
      return false
    }
  }

  static async validateSchoolCode(schoolCode: string): Promise<boolean> {
    try {
      const response = await fetch('/api/schools/validate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ schoolCode })
      })
      const data = await response.json()
      return data.valid
    } catch {
      return false
    }
  }
}

// Export commonly used validators
export const validate = FormValidator
export const createValidator = (fieldName: string) => new ValidationComposer(fieldName)

// Pre-configured validators for common use cases
export const validators = {
  required: ValidationRules.required,
  email: ValidationRules.email,
  username: ValidationRules.username,
  password: ValidationRules.password,
  strongPassword: ValidationRules.strongPassword,
  phone: ValidationRules.phone,
  letterTitle: ValidationRules.letterTitle,
  letterContent: ValidationRules.letterContent,
  schoolCode: ValidationRules.schoolCode
}

// Password validation helpers
export const validatePassword = (password: string): { isValid: boolean; error?: string } => {
  const strongPasswordError = ValidationRules.strongPassword(password)
  if (strongPasswordError) {
    return { isValid: false, error: strongPasswordError }
  }
  return { isValid: true }
}

export const getPasswordStrength = (password: string): { score: number; feedback: string[] } => {
  return SecurityValidator.checkPasswordStrength(password)
}