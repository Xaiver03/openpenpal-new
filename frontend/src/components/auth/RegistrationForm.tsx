import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { SchoolSelector } from '@/components/ui/school-selector';
import { AuthService, type RegisterRequest } from '@/lib/services';
import { useAuth } from '@/contexts/auth-context-new';
import { Eye, EyeOff, Check, X, Loader2 } from 'lucide-react';

interface RegistrationFormProps {
  email: string;
  verificationCode: string;
  onSuccess: (userId: string) => void;
  onBack: () => void;
}

export const RegistrationForm: React.FC<RegistrationFormProps> = ({
  email,
  verificationCode,
  onSuccess,
  onBack,
}) => {
  const { register } = useAuth()

  const [formData, setFormData] = useState<RegisterRequest>({
    username: '',
    email,
    password: '',
    nickname: '',
    schoolCode: '',
    school_name: ''
  });

  const [confirmPassword, setConfirmPassword] = useState('')
  const [agreeToTerms, setAgreeToTerms] = useState(false)
  const [agreeToPrivacy, setAgreeToPrivacy] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)

  const [validation, setValidation] = useState({
    username: { available: false, checking: false, error: '' },
    password: { valid: false, error: '' },
    confirmPassword: { valid: false, error: '' },
    nickname: { valid: false, error: '' },
    email: { valid: false, error: '' }
  });

  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // 密码强度验证
  const validatePassword = (password: string) => {
    const requirements = [
      { test: password.length >= 8, message: '至少8位字符' },
      { test: /[A-Z]/.test(password), message: '包含大写字母' },
      { test: /[a-z]/.test(password), message: '包含小写字母' },
      { test: /\d/.test(password), message: '包含数字' }
    ]
    
    const failedRequirements = requirements.filter(req => !req.test)
    return {
      valid: failedRequirements.length === 0,
      error: failedRequirements.map(req => req.message).join(', ')
    }
  }

  // 用户名验证
  const validateUsername = async (username: string) => {
    if (!username || username.length < 3) {
      return { valid: false, error: '用户名至少3位字符' }
    }
    
    if (!/^[a-zA-Z0-9_]+$/.test(username)) {
      return { valid: false, error: '用户名只能包含字母、数字和下划线' }
    }
    
    try {
      const response = await AuthService.checkUsernameAvailability(username)
      if (response.success && response.data) {
        return {
          valid: response.data.available,
          error: response.data.available ? '' : '用户名已被占用'
        }
      }
    } catch (error) {
      console.error('Username validation failed:', error)
    }
    
    return { valid: false, error: '验证失败，请稍后重试' }
  }

  // 昵称验证
  const validateNickname = (nickname: string) => {
    if (!nickname || nickname.trim().length < 1) {
      return { valid: false, error: '请输入昵称' }
    }
    if (nickname.length > 20) {
      return { valid: false, error: '昵称不能超过20个字符' }
    }
    return { valid: true, error: '' }
  }

  // 表单字段更新处理
  const handleInputChange = (field: keyof RegisterRequest, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }))
    setError('')

    // 实时验证
    if (field === 'password') {
      const passwordValidation = validatePassword(value)
      setValidation(prev => ({
        ...prev,
        password: passwordValidation,
        confirmPassword: {
          valid: confirmPassword ? confirmPassword === value : false,
          error: confirmPassword && confirmPassword !== value ? '密码不匹配' : ''
        }
      }))
    }

    if (field === 'nickname') {
      const nicknameValidation = validateNickname(value)
      setValidation(prev => ({ ...prev, nickname: nicknameValidation }))
    }

    if (field === 'username') {
      // 防抖验证用户名
      const timeoutId = setTimeout(async () => {
        if (value.length >= 3) {
          setValidation(prev => ({ ...prev, username: { ...prev.username, checking: true } }))
          const usernameValidation = await validateUsername(value)
          setValidation(prev => ({
            ...prev,
            username: {
              available: usernameValidation.valid,
              checking: false,
              error: usernameValidation.error
            }
          }))
        }
      }, 500)

      return () => clearTimeout(timeoutId)
    }
  }

  // 确认密码处理
  const handleConfirmPasswordChange = (value: string) => {
    setConfirmPassword(value)
    const isValid = value === formData.password
    setValidation(prev => ({
      ...prev,
      confirmPassword: {
        valid: isValid,
        error: isValid ? '' : '密码不匹配'
      }
    }))
  }

  // 学校选择处理
  const handleSchoolSelect = (schoolCode: string, schoolName: string) => {
    setFormData(prev => ({
      ...prev,
      schoolCode: schoolCode,
      school_name: schoolName
    }))
  }

  // 表单验证
  const isFormValid = () => {
    return (
      validation.username.available &&
      validation.password.valid &&
      validation.confirmPassword.valid &&
      validation.nickname.valid &&
      formData.schoolCode &&
      agreeToTerms &&
      agreeToPrivacy
    )
  }

  // 提交表单
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!isFormValid()) {
      setError('请完善所有必填信息并同意用户协议')
      return
    }

    setLoading(true)
    setError('')

    try {
      const result = await register(formData)
      if (result.success) {
        onSuccess('registration_success')
      } else {
        setError(result.message || '注册失败，请稍后重试')
      }
    } catch (error: any) {
      console.error('Registration error:', error)
      setError(error.message || '注册过程中发生错误，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {error && (
        <Alert className="border-red-200 bg-red-50">
          <X className="h-4 w-4 text-red-600" />
          <AlertDescription className="text-red-800">{error}</AlertDescription>
        </Alert>
      )}

      {/* 用户名输入 */}
      <div className="space-y-2">
        <Label htmlFor="username" className="text-sm font-medium">
          用户名 <span className="text-red-500">*</span>
        </Label>
        <div className="relative">
          <Input
            id="username"
            type="text"
            value={formData.username}
            onChange={(e) => handleInputChange('username', e.target.value)}
            placeholder="请输入用户名"
            className={`pr-10 ${
              formData.username
                ? validation.username.available
                  ? 'border-green-500 focus:border-green-500'
                  : 'border-red-500 focus:border-red-500'
                : ''
            }`}
            disabled={loading}
          />
          <div className="absolute inset-y-0 right-0 flex items-center pr-3">
            {validation.username.checking ? (
              <Loader2 className="h-4 w-4 animate-spin text-gray-400" />
            ) : formData.username ? (
              validation.username.available ? (
                <Check className="h-4 w-4 text-green-500" />
              ) : (
                <X className="h-4 w-4 text-red-500" />
              )
            ) : null}
          </div>
        </div>
        {validation.username.error && (
          <p className="text-sm text-red-600">{validation.username.error}</p>
        )}
        <p className="text-xs text-gray-500">
          用户名用于登录，3-20位字符，只能包含字母、数字和下划线
        </p>
      </div>

      {/* 昵称输入 */}
      <div className="space-y-2">
        <Label htmlFor="nickname" className="text-sm font-medium">
          昵称 <span className="text-red-500">*</span>
        </Label>
        <Input
          id="nickname"
          type="text"
          value={formData.nickname}
          onChange={(e) => handleInputChange('nickname', e.target.value)}
          placeholder="请输入您的昵称"
          className={`${
            formData.nickname
              ? validation.nickname.valid
                ? 'border-green-500 focus:border-green-500'
                : 'border-red-500 focus:border-red-500'
              : ''
          }`}
          disabled={loading}
        />
        {validation.nickname.error && (
          <p className="text-sm text-red-600">{validation.nickname.error}</p>
        )}
      </div>

      {/* 学校选择 */}
      <div className="space-y-2">
        <Label className="text-sm font-medium">
          学校 <span className="text-red-500">*</span>
        </Label>
        <SchoolSelector
          value={formData.schoolCode}
          onChange={handleSchoolSelect}
          placeholder="请选择您的学校"
          disabled={loading}
        />
        {formData.school_name && (
          <p className="text-sm text-green-600">
            已选择: {formData.school_name} ({formData.schoolCode})
          </p>
        )}
      </div>

      {/* 密码输入 */}
      <div className="space-y-2">
        <Label htmlFor="password" className="text-sm font-medium">
          密码 <span className="text-red-500">*</span>
        </Label>
        <div className="relative">
          <Input
            id="password"
            type={showPassword ? 'text' : 'password'}
            value={formData.password}
            onChange={(e) => handleInputChange('password', e.target.value)}
            placeholder="请输入密码"
            className={`pr-10 ${
              formData.password
                ? validation.password.valid
                  ? 'border-green-500 focus:border-green-500'
                  : 'border-red-500 focus:border-red-500'
                : ''
            }`}
            disabled={loading}
          />
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="absolute inset-y-0 right-0 px-3 py-0 hover:bg-transparent"
            onClick={() => setShowPassword(!showPassword)}
            tabIndex={-1}
          >
            {showPassword ? (
              <EyeOff className="h-4 w-4 text-gray-400" />
            ) : (
              <Eye className="h-4 w-4 text-gray-400" />
            )}
          </Button>
        </div>
        {validation.password.error && (
          <p className="text-sm text-red-600">{validation.password.error}</p>
        )}
        <p className="text-xs text-gray-500">
          密码至少8位，包含大小写字母和数字
        </p>
      </div>

      {/* 确认密码输入 */}
      <div className="space-y-2">
        <Label htmlFor="confirmPassword" className="text-sm font-medium">
          确认密码 <span className="text-red-500">*</span>
        </Label>
        <div className="relative">
          <Input
            id="confirmPassword"
            type={showConfirmPassword ? 'text' : 'password'}
            value={confirmPassword}
            onChange={(e) => handleConfirmPasswordChange(e.target.value)}
            placeholder="请再次输入密码"
            className={`pr-10 ${
              confirmPassword
                ? validation.confirmPassword.valid
                  ? 'border-green-500 focus:border-green-500'
                  : 'border-red-500 focus:border-red-500'
                : ''
            }`}
            disabled={loading}
          />
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="absolute inset-y-0 right-0 px-3 py-0 hover:bg-transparent"
            onClick={() => setShowConfirmPassword(!showConfirmPassword)}
            tabIndex={-1}
          >
            {showConfirmPassword ? (
              <EyeOff className="h-4 w-4 text-gray-400" />
            ) : (
              <Eye className="h-4 w-4 text-gray-400" />
            )}
          </Button>
        </div>
        {validation.confirmPassword.error && (
          <p className="text-sm text-red-600">{validation.confirmPassword.error}</p>
        )}
      </div>

      {/* 协议同意 */}
      <div className="space-y-3">
        <div className="flex items-start space-x-2">
          <Checkbox
            id="agreeToTerms"
            checked={agreeToTerms}
            onCheckedChange={setAgreeToTerms}
          />
          <Label htmlFor="agreeToTerms" className="text-sm leading-5">
            我已阅读并同意{' '}
            <a href="/terms" target="_blank" className="text-blue-600 hover:underline">
              《用户服务协议》
            </a>
          </Label>
        </div>
        <div className="flex items-start space-x-2">
          <Checkbox
            id="agreeToPrivacy"
            checked={agreeToPrivacy}
            onCheckedChange={setAgreeToPrivacy}
          />
          <Label htmlFor="agreeToPrivacy" className="text-sm leading-5">
            我已阅读并同意{' '}
            <a href="/privacy" target="_blank" className="text-blue-600 hover:underline">
              《隐私政策》
            </a>
          </Label>
        </div>
      </div>

      {/* 操作按钮 */}
      <div className="flex space-x-4">
        <Button
          type="button"
          variant="outline"
          onClick={onBack}
          disabled={loading}
          className="flex-1"
        >
          返回
        </Button>
        <Button
          type="submit"
          disabled={!isFormValid() || loading}
          className="flex-1"
        >
          {loading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              注册中...
            </>
          ) : (
            '完成注册'
          )}
        </Button>
      </div>
    </form>
  )
}