'use client'

import { useState } from 'react'
import { Key, Shield, Eye, EyeOff, Loader2, AlertTriangle, Check } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useToast } from '@/components/ui/use-toast'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import { securityApi } from '@/lib/api/security'
import { validatePassword, getPasswordStrength } from '@/utils/validation'

interface PasswordStrengthIndicatorProps {
  password: string
}

function PasswordStrengthIndicator({ password }: PasswordStrengthIndicatorProps) {
  const strength = getPasswordStrength(password)
  
  const getStrengthColor = () => {
    if (strength.score <= 1) return 'bg-red-500'
    if (strength.score <= 2) return 'bg-orange-500'
    if (strength.score <= 3) return 'bg-yellow-500'
    return 'bg-green-500'
  }
  
  const getStrengthText = () => {
    if (strength.score <= 1) return '弱'
    if (strength.score <= 2) return '一般'
    if (strength.score <= 3) return '强'
    return '很强'
  }

  if (!password) return null

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-sm">
        <span className="text-muted-foreground">密码强度</span>
        <span className={cn(
          'font-medium',
          strength.score <= 1 && 'text-red-500',
          strength.score === 2 && 'text-orange-500',
          strength.score === 3 && 'text-yellow-500',
          strength.score >= 4 && 'text-green-500'
        )}>
          {getStrengthText()}
        </span>
      </div>
      <div className="h-2 bg-muted rounded-full overflow-hidden">
        <div 
          className={cn('h-full transition-all duration-300', getStrengthColor())}
          style={{ width: `${(strength.score / 4) * 100}%` }}
        />
      </div>
      {strength.feedback.length > 0 && (
        <ul className="text-xs text-muted-foreground space-y-1">
          {strength.feedback.map((item, index) => (
            <li key={index} className="flex items-start">
              <span className="mr-1">•</span>
              <span>{item}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

export function SecuritySettings() {
  const { toast } = useToast()
  const [isChangingPassword, setIsChangingPassword] = useState(false)
  const [showCurrentPassword, setShowCurrentPassword] = useState(false)
  const [showNewPassword, setShowNewPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  
  const [passwordForm, setPasswordForm] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  })
  
  const [passwordErrors, setPasswordErrors] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  })

  const handlePasswordChange = (field: keyof typeof passwordForm, value: string) => {
    setPasswordForm(prev => ({ ...prev, [field]: value }))
    
    // Clear errors when user starts typing
    if (passwordErrors[field]) {
      setPasswordErrors(prev => ({ ...prev, [field]: '' }))
    }
    
    // Validate new password in real-time
    if (field === 'newPassword') {
      const validation = validatePassword(value)
      if (!validation.isValid && value.length > 0) {
        setPasswordErrors(prev => ({ 
          ...prev, 
          newPassword: validation.error || '密码不符合要求' 
        }))
      }
    }
    
    // Check password match in real-time
    if (field === 'confirmPassword' || (field === 'newPassword' && passwordForm.confirmPassword)) {
      const confirmValue = field === 'confirmPassword' ? value : passwordForm.confirmPassword
      const newValue = field === 'newPassword' ? value : passwordForm.newPassword
      
      if (confirmValue && newValue !== confirmValue) {
        setPasswordErrors(prev => ({ 
          ...prev, 
          confirmPassword: '两次输入的密码不一致' 
        }))
      } else {
        setPasswordErrors(prev => ({ 
          ...prev, 
          confirmPassword: '' 
        }))
      }
    }
  }

  const validatePasswordForm = (): boolean => {
    const errors = {
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    }
    
    if (!passwordForm.currentPassword) {
      errors.currentPassword = '请输入当前密码'
    }
    
    if (!passwordForm.newPassword) {
      errors.newPassword = '请输入新密码'
    } else {
      const validation = validatePassword(passwordForm.newPassword)
      if (!validation.isValid) {
        errors.newPassword = validation.error || '密码不符合要求'
      }
    }
    
    if (!passwordForm.confirmPassword) {
      errors.confirmPassword = '请确认新密码'
    } else if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      errors.confirmPassword = '两次输入的密码不一致'
    }
    
    if (passwordForm.currentPassword === passwordForm.newPassword) {
      errors.newPassword = '新密码不能与当前密码相同'
    }
    
    setPasswordErrors(errors)
    return !Object.values(errors).some(error => error !== '')
  }

  const handleSubmitPasswordChange = async () => {
    if (!validatePasswordForm()) {
      return
    }
    
    setIsChangingPassword(true)
    try {
      await securityApi.changePassword({
        current_password: passwordForm.currentPassword,
        new_password: passwordForm.newPassword
      })
      
      toast({
        title: '密码修改成功',
        description: '您的密码已成功更新，请使用新密码登录',
      })
      
      // Reset form
      setPasswordForm({
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
      })
      setPasswordErrors({
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
      })
    } catch (error: any) {
      console.error('Failed to change password:', error)
      
      // Handle specific error cases
      if (error.response?.status === 401) {
        setPasswordErrors(prev => ({ 
          ...prev, 
          currentPassword: '当前密码不正确' 
        }))
      } else {
        toast({
          title: '修改失败',
          description: error.response?.data?.message || '密码修改失败，请稍后重试',
          variant: 'destructive'
        })
      }
    } finally {
      setIsChangingPassword(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* 密码修改 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Key className="h-5 w-5" />
            修改密码
          </CardTitle>
          <CardDescription>
            定期更改密码可以提高账户安全性
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="current-password">当前密码</Label>
            <div className="relative">
              <Input
                id="current-password"
                type={showCurrentPassword ? 'text' : 'password'}
                value={passwordForm.currentPassword}
                onChange={(e) => handlePasswordChange('currentPassword', e.target.value)}
                placeholder="输入当前密码"
                className={cn(
                  passwordErrors.currentPassword && 'border-destructive'
                )}
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                onClick={() => setShowCurrentPassword(!showCurrentPassword)}
              >
                {showCurrentPassword ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </Button>
            </div>
            {passwordErrors.currentPassword && (
              <p className="text-sm text-destructive">{passwordErrors.currentPassword}</p>
            )}
          </div>

          <Separator />

          <div className="space-y-2">
            <Label htmlFor="new-password">新密码</Label>
            <div className="relative">
              <Input
                id="new-password"
                type={showNewPassword ? 'text' : 'password'}
                value={passwordForm.newPassword}
                onChange={(e) => handlePasswordChange('newPassword', e.target.value)}
                placeholder="输入新密码"
                className={cn(
                  passwordErrors.newPassword && 'border-destructive'
                )}
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                onClick={() => setShowNewPassword(!showNewPassword)}
              >
                {showNewPassword ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </Button>
            </div>
            {passwordErrors.newPassword && (
              <p className="text-sm text-destructive">{passwordErrors.newPassword}</p>
            )}
            <PasswordStrengthIndicator password={passwordForm.newPassword} />
          </div>

          <div className="space-y-2">
            <Label htmlFor="confirm-password">确认新密码</Label>
            <div className="relative">
              <Input
                id="confirm-password"
                type={showConfirmPassword ? 'text' : 'password'}
                value={passwordForm.confirmPassword}
                onChange={(e) => handlePasswordChange('confirmPassword', e.target.value)}
                placeholder="再次输入新密码"
                className={cn(
                  passwordErrors.confirmPassword && 'border-destructive'
                )}
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              >
                {showConfirmPassword ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </Button>
            </div>
            {passwordErrors.confirmPassword && (
              <p className="text-sm text-destructive">{passwordErrors.confirmPassword}</p>
            )}
          </div>

          <Alert>
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              密码要求：至少8个字符，包含大小写字母、数字和特殊字符
            </AlertDescription>
          </Alert>

          <div className="flex justify-end">
            <Button
              onClick={handleSubmitPasswordChange}
              disabled={isChangingPassword}
            >
              {isChangingPassword && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              修改密码
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* 两步验证 - 预留位置 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Shield className="h-5 w-5" />
            两步验证
          </CardTitle>
          <CardDescription>
            为您的账户添加额外的安全保护层
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            <Shield className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>两步验证功能即将推出</p>
            <p className="text-sm mt-2">通过手机验证码或认证器应用保护您的账户</p>
          </div>
        </CardContent>
      </Card>

      {/* 登录活动 - 预留位置 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">登录活动</CardTitle>
          <CardDescription>
            查看您的账户最近的登录记录
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            <p>登录活动记录功能即将推出</p>
            <p className="text-sm mt-2">监控异常登录活动，保护账户安全</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}