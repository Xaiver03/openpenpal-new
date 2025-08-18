'use client'

import React, { useState, useEffect } from 'react'
import { 
  Settings, 
  Save, 
  RotateCcw, 
  Database, 
  Mail, 
  Shield, 
  Bell, 
  Palette,
  Globe,
  Server,
  Key,
  AlertTriangle,
  CheckCircle,
  Info,
  Upload,
  ArrowLeft
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Separator } from '@/components/ui/separator'
import { usePermission, PERMISSIONS } from '@/hooks/use-permission'
import { useUnsavedChanges } from '@/hooks/use-unsaved-changes'
import { SafeBackButton } from '@/components/ui/safe-back-button'

interface SystemConfig {
  // 基本设置
  site_name: string
  site_description: string
  site_logo: string
  maintenance_mode: boolean
  
  // 邮件设置
  smtp_host: string
  smtp_port: number
  smtp_username: string
  smtp_password: string
  smtp_encryption: 'tls' | 'ssl' | 'none'
  email_from_name: string
  email_from_address: string
  
  // 信件设置
  max_letter_length: number
  allowed_file_types: string[]
  max_file_size: number
  letter_review_required: boolean
  auto_delivery_enabled: boolean
  
  // 用户设置
  user_registration_enabled: boolean
  email_verification_required: boolean
  max_users_per_school: number
  user_inactive_days: number
  
  // 信使设置
  courier_application_enabled: boolean
  courier_auto_approval: boolean
  max_delivery_distance: number
  courier_rating_required: boolean
  
  // 安全设置
  password_min_length: number
  password_require_symbols: boolean
  password_require_numbers: boolean
  session_timeout: number
  max_login_attempts: number
  jwt_expiry_hours: number
  refresh_token_days: number
  enable_token_refresh: boolean
  
  // 通知设置
  email_notifications: boolean
  sms_notifications: boolean
  push_notifications: boolean
  admin_notifications: boolean
}

const DEFAULT_CONFIG: SystemConfig = {
  site_name: 'OpenPenPal',
  site_description: '温暖的校园信件投递平台',
  site_logo: '',
  maintenance_mode: false,
  
  smtp_host: 'smtp.example.com',
  smtp_port: 587,
  smtp_username: '',
  smtp_password: '',
  smtp_encryption: 'tls',
  email_from_name: 'OpenPenPal',
  email_from_address: 'noreply@openpenpal.com',
  
  max_letter_length: 5000,
  allowed_file_types: ['jpg', 'png', 'pdf'],
  max_file_size: 10,
  letter_review_required: false,
  auto_delivery_enabled: true,
  
  user_registration_enabled: true,
  email_verification_required: true,
  max_users_per_school: 10000,
  user_inactive_days: 90,
  
  courier_application_enabled: true,
  courier_auto_approval: false,
  max_delivery_distance: 10,
  courier_rating_required: true,
  
  password_min_length: 6,
  password_require_symbols: false,
  password_require_numbers: true,
  session_timeout: 3600,
  max_login_attempts: 5,
  jwt_expiry_hours: 24,
  refresh_token_days: 7,
  enable_token_refresh: true,
  
  email_notifications: true,
  sms_notifications: false,
  push_notifications: true,
  admin_notifications: true
}

export default function SystemSettingsPage() {
  const { user, hasPermission } = usePermission()
  const [config, setConfig] = useState<SystemConfig>(DEFAULT_CONFIG)
  const [originalConfig, setOriginalConfig] = useState<SystemConfig>(DEFAULT_CONFIG)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [testingEmail, setTestingEmail] = useState(false)
  const [hasChanges, setHasChanges] = useState(false)

  // 权限检查
  if (!user || !hasPermission(PERMISSIONS.SYSTEM_CONFIG)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Shield className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">访问权限不足</h2>
            <p className="text-gray-600 mb-4">
              您没有访问系统设置的权限
            </p>
            <Button asChild variant="outline">
              <a href="/admin">返回管理控制台</a>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  // 加载配置
  useEffect(() => {
    loadConfig()
  }, [])

  // 检查是否有更改
  useEffect(() => {
    const hasChanges = JSON.stringify(config) !== JSON.stringify(originalConfig)
    setHasChanges(hasChanges)
  }, [config, originalConfig])

  const loadConfig = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/admin/settings')
      const result = await response.json()
      
      if (result.success && result.data?.code === 0) {
        setConfig(result.data.data)
        setOriginalConfig(result.data.data)
      } else {
        console.error('加载配置失败:', result.message)
        alert('加载配置失败: ' + result.message)
      }
    } catch (error) {
      console.error('Failed to load config:', error)
      alert('加载配置失败，请刷新页面重试')
    } finally {
      setLoading(false)
    }
  }

  const saveConfig = async () => {
    setSaving(true)
    try {
      const response = await fetch('/api/v1/admin/settings', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(config),
      })
      
      const result = await response.json()
      
      if (result.success && result.data?.code === 0) {
        setOriginalConfig(result.data.data)
        alert(result.data.message || '配置保存成功！')
      } else {
        alert('配置保存失败: ' + result.message)
      }
    } catch (error) {
      console.error('Failed to save config:', error)
      alert('配置保存失败，请重试')
    } finally {
      setSaving(false)
    }
  }

  const resetConfig = async () => {
    if (confirm('确定要重置所有设置到默认值吗？此操作不可撤销。')) {
      try {
        const response = await fetch('/api/v1/admin/settings', {
          method: 'POST',
        })
        
        const result = await response.json()
        
        if (result.success && result.data?.code === 0) {
          setConfig(result.data.data)
          setOriginalConfig(result.data.data)
          alert(result.data.message || '配置已重置为默认值！')
        } else {
          alert('重置失败: ' + result.message)
        }
      } catch (error) {
        console.error('Failed to reset config:', error)
        alert('重置失败，请重试')
      }
    }
  }

  const testEmailConfig = async () => {
    const testEmail = prompt('请输入用于测试的邮箱地址:', user?.email || '')
    if (!testEmail) return
    
    setTestingEmail(true)
    try {
      const response = await fetch('/api/v1/admin/settings/test-email', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...config,
          test_email: testEmail
        }),
      })
      
      const result = await response.json()
      
      if (result.success && result.data?.code === 0) {
        alert(result.data.message || '测试邮件发送成功！请检查邮箱: ' + testEmail)
      } else {
        alert('测试邮件发送失败: ' + result.message)
      }
    } catch (error) {
      console.error('Email test failed:', error)
      alert('测试邮件发送失败，请检查邮件配置。')
    } finally {
      setTestingEmail(false)
    }
  }

  const handleConfigChange = (key: keyof SystemConfig, value: any) => {
    setConfig(prev => ({ ...prev, [key]: value }))
  }
  
  // 未保存更改检测
  const { confirmLeave, safeNavigate } = useUnsavedChanges({
    hasUnsavedChanges: hasChanges,
    message: '您有未保存的系统设置更改。是否要在离开前保存？',
    onSave: saveConfig,
    onDiscard: () => {
      setConfig(originalConfig)
      setHasChanges(false)
    }
  })

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <SafeBackButton 
            href="/admin" 
            label="返回管理控制台"
            hasUnsavedChanges={hasChanges}
            confirmDescription="您有未保存的系统设置更改。确定要离开吗？"
          />
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-2">
              <Settings className="w-8 h-8" />
              系统设置
            </h1>
            <p className="text-muted-foreground mt-1">
              配置系统参数和功能开关
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={resetConfig}>
            <RotateCcw className="w-4 h-4 mr-2" />
            重置默认
          </Button>
          <Button 
            onClick={saveConfig} 
            disabled={!hasChanges || saving}
          >
            <Save className="w-4 h-4 mr-2" />
            {saving ? '保存中...' : '保存设置'}
          </Button>
        </div>
      </div>

      {/* 更改提醒 */}
      {hasChanges && (
        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            您有未保存的更改，请记得点击"保存设置"按钮。
          </AlertDescription>
        </Alert>
      )}

      {/* 设置选项卡 */}
      <Card>
        <CardContent className="p-6">
          <Tabs defaultValue="general" className="space-y-6">
            <TabsList className="grid w-full grid-cols-6">
              <TabsTrigger value="general">基本设置</TabsTrigger>
              <TabsTrigger value="email">邮件配置</TabsTrigger>
              <TabsTrigger value="letters">信件设置</TabsTrigger>
              <TabsTrigger value="users">用户管理</TabsTrigger>
              <TabsTrigger value="security">安全设置</TabsTrigger>
              <TabsTrigger value="notifications">通知设置</TabsTrigger>
            </TabsList>

            {/* 基本设置 */}
            <TabsContent value="general" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Globe className="w-5 h-5" />
                    站点信息
                  </CardTitle>
                  <CardDescription>配置网站基本信息和外观</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label htmlFor="site_name">网站名称</Label>
                    <Input
                      id="site_name"
                      value={config.site_name}
                      onChange={(e) => handleConfigChange('site_name', e.target.value)}
                    />
                  </div>

                  <div>
                    <Label htmlFor="site_description">网站描述</Label>
                    <Textarea
                      id="site_description"
                      value={config.site_description}
                      onChange={(e) => handleConfigChange('site_description', e.target.value)}
                      rows={3}
                    />
                  </div>

                  <div>
                    <Label htmlFor="site_logo">网站Logo URL</Label>
                    <div className="flex gap-2">
                      <Input
                        id="site_logo"
                        value={config.site_logo}
                        onChange={(e) => handleConfigChange('site_logo', e.target.value)}
                        placeholder="https://example.com/logo.png"
                      />
                      <Button variant="outline">
                        <Upload className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>

                  <Separator />

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="maintenance_mode">维护模式</Label>
                      <p className="text-sm text-muted-foreground">
                        开启后，普通用户将无法访问网站
                      </p>
                    </div>
                    <Switch
                      id="maintenance_mode"
                      checked={config.maintenance_mode}
                      onCheckedChange={(checked) => handleConfigChange('maintenance_mode', checked)}
                    />
                  </div>
                  
                  {config.maintenance_mode && (
                    <Alert>
                      <AlertTriangle className="h-4 w-4" />
                      <AlertDescription>
                        维护模式已开启，普通用户将看到维护页面。
                      </AlertDescription>
                    </Alert>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            {/* 邮件配置 */}
            <TabsContent value="email" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Mail className="w-5 h-5" />
                    SMTP 配置
                  </CardTitle>
                  <CardDescription>配置邮件发送服务器</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="smtp_host">SMTP 服务器</Label>
                      <Input
                        id="smtp_host"
                        value={config.smtp_host}
                        onChange={(e) => handleConfigChange('smtp_host', e.target.value)}
                        placeholder="smtp.example.com"
                      />
                    </div>
                    <div>
                      <Label htmlFor="smtp_port">端口</Label>
                      <Input
                        id="smtp_port"
                        type="number"
                        value={config.smtp_port}
                        onChange={(e) => handleConfigChange('smtp_port', parseInt(e.target.value))}
                      />
                    </div>
                  </div>

                  <div>
                    <Label htmlFor="smtp_username">用户名</Label>
                    <Input
                      id="smtp_username"
                      value={config.smtp_username}
                      onChange={(e) => handleConfigChange('smtp_username', e.target.value)}
                    />
                  </div>

                  <div>
                    <Label htmlFor="smtp_password">密码</Label>
                    <Input
                      id="smtp_password"
                      type="password"
                      value={config.smtp_password}
                      onChange={(e) => handleConfigChange('smtp_password', e.target.value)}
                    />
                  </div>

                  <div>
                    <Label htmlFor="smtp_encryption">加密方式</Label>
                    <Select
                      value={config.smtp_encryption}
                      onValueChange={(value: 'tls' | 'ssl' | 'none') => handleConfigChange('smtp_encryption', value)}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="none">无加密</SelectItem>
                        <SelectItem value="tls">TLS</SelectItem>
                        <SelectItem value="ssl">SSL</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <Separator />

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="email_from_name">发送者名称</Label>
                      <Input
                        id="email_from_name"
                        value={config.email_from_name}
                        onChange={(e) => handleConfigChange('email_from_name', e.target.value)}
                      />
                    </div>
                    <div>
                      <Label htmlFor="email_from_address">发送者邮箱</Label>
                      <Input
                        id="email_from_address"
                        type="email"
                        value={config.email_from_address}
                        onChange={(e) => handleConfigChange('email_from_address', e.target.value)}
                      />
                    </div>
                  </div>

                  <div className="flex justify-end">
                    <Button 
                      variant="outline" 
                      onClick={testEmailConfig}
                      disabled={testingEmail}
                    >
                      <Mail className="w-4 h-4 mr-2" />
                      {testingEmail ? '测试中...' : '测试邮件配置'}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* 信件设置 */}
            <TabsContent value="letters" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Mail className="w-5 h-5" />
                    信件限制
                  </CardTitle>
                  <CardDescription>设置信件内容和附件限制</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label htmlFor="max_letter_length">最大信件长度（字符）</Label>
                    <Input
                      id="max_letter_length"
                      type="number"
                      value={config.max_letter_length}
                      onChange={(e) => handleConfigChange('max_letter_length', parseInt(e.target.value))}
                    />
                  </div>

                  <div>
                    <Label htmlFor="max_file_size">最大文件大小（MB）</Label>
                    <Input
                      id="max_file_size"
                      type="number"
                      value={config.max_file_size}
                      onChange={(e) => handleConfigChange('max_file_size', parseInt(e.target.value))}
                    />
                  </div>

                  <div>
                    <Label htmlFor="allowed_file_types">允许的文件类型</Label>
                    <Input
                      id="allowed_file_types"
                      value={config.allowed_file_types.join(', ')}
                      onChange={(e) => handleConfigChange('allowed_file_types', e.target.value.split(', '))}
                      placeholder="jpg, png, pdf"
                    />
                  </div>

                  <Separator />

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="letter_review_required">信件审核</Label>
                      <p className="text-sm text-muted-foreground">
                        是否需要管理员审核信件后才能发送
                      </p>
                    </div>
                    <Switch
                      id="letter_review_required"
                      checked={config.letter_review_required}
                      onCheckedChange={(checked) => handleConfigChange('letter_review_required', checked)}
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="auto_delivery_enabled">自动投递</Label>
                      <p className="text-sm text-muted-foreground">
                        是否自动分配信使进行投递
                      </p>
                    </div>
                    <Switch
                      id="auto_delivery_enabled"
                      checked={config.auto_delivery_enabled}
                      onCheckedChange={(checked) => handleConfigChange('auto_delivery_enabled', checked)}
                    />
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* 用户管理 */}
            <TabsContent value="users" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Shield className="w-5 h-5" />
                    用户注册
                  </CardTitle>
                  <CardDescription>管理用户注册和验证设置</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="user_registration_enabled">用户注册</Label>
                      <p className="text-sm text-muted-foreground">
                        是否允许新用户注册
                      </p>
                    </div>
                    <Switch
                      id="user_registration_enabled"
                      checked={config.user_registration_enabled}
                      onCheckedChange={(checked) => handleConfigChange('user_registration_enabled', checked)}
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="email_verification_required">邮箱验证</Label>
                      <p className="text-sm text-muted-foreground">
                        新用户是否需要验证邮箱
                      </p>
                    </div>
                    <Switch
                      id="email_verification_required"
                      checked={config.email_verification_required}
                      onCheckedChange={(checked) => handleConfigChange('email_verification_required', checked)}
                    />
                  </div>

                  <div>
                    <Label htmlFor="max_users_per_school">每校最大用户数</Label>
                    <Input
                      id="max_users_per_school"
                      type="number"
                      value={config.max_users_per_school}
                      onChange={(e) => handleConfigChange('max_users_per_school', parseInt(e.target.value))}
                    />
                  </div>

                  <div>
                    <Label htmlFor="user_inactive_days">用户非活跃天数</Label>
                    <Input
                      id="user_inactive_days"
                      type="number"
                      value={config.user_inactive_days}
                      onChange={(e) => handleConfigChange('user_inactive_days', parseInt(e.target.value))}
                    />
                    <p className="text-sm text-muted-foreground mt-1">
                      超过此天数未登录的用户将被标记为非活跃
                    </p>
                  </div>

                  <Separator />

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="courier_application_enabled">信使申请</Label>
                      <p className="text-sm text-muted-foreground">
                        是否允许用户申请成为信使
                      </p>
                    </div>
                    <Switch
                      id="courier_application_enabled"
                      checked={config.courier_application_enabled}
                      onCheckedChange={(checked) => handleConfigChange('courier_application_enabled', checked)}
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="courier_auto_approval">信使自动审批</Label>
                      <p className="text-sm text-muted-foreground">
                        是否自动通过信使申请
                      </p>
                    </div>
                    <Switch
                      id="courier_auto_approval"
                      checked={config.courier_auto_approval}
                      onCheckedChange={(checked) => handleConfigChange('courier_auto_approval', checked)}
                    />
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* 安全设置 */}
            <TabsContent value="security" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Key className="w-5 h-5" />
                    密码策略
                  </CardTitle>
                  <CardDescription>设置用户密码安全要求</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label htmlFor="password_min_length">最小密码长度</Label>
                    <Input
                      id="password_min_length"
                      type="number"
                      value={config.password_min_length}
                      onChange={(e) => handleConfigChange('password_min_length', parseInt(e.target.value))}
                      min="6"
                      max="50"
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="password_require_numbers">需要数字</Label>
                      <p className="text-sm text-muted-foreground">
                        密码必须包含数字
                      </p>
                    </div>
                    <Switch
                      id="password_require_numbers"
                      checked={config.password_require_numbers}
                      onCheckedChange={(checked) => handleConfigChange('password_require_numbers', checked)}
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="password_require_symbols">需要特殊字符</Label>
                      <p className="text-sm text-muted-foreground">
                        密码必须包含特殊字符
                      </p>
                    </div>
                    <Switch
                      id="password_require_symbols"
                      checked={config.password_require_symbols}
                      onCheckedChange={(checked) => handleConfigChange('password_require_symbols', checked)}
                    />
                  </div>

                  <Separator />

                  <div>
                    <Label htmlFor="session_timeout">会话超时（秒）</Label>
                    <Input
                      id="session_timeout"
                      type="number"
                      value={config.session_timeout}
                      onChange={(e) => handleConfigChange('session_timeout', parseInt(e.target.value))}
                    />
                  </div>

                  <div>
                    <Label htmlFor="max_login_attempts">最大登录尝试次数</Label>
                    <Input
                      id="max_login_attempts"
                      type="number"
                      value={config.max_login_attempts}
                      onChange={(e) => handleConfigChange('max_login_attempts', parseInt(e.target.value))}
                    />
                  </div>
                </CardContent>
              </Card>

              {/* JWT Token设置 */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Key className="w-5 h-5" />
                    JWT Token配置
                  </CardTitle>
                  <CardDescription>配置用户认证Token的过期时间和刷新策略</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label htmlFor="jwt_expiry_hours">JWT过期时间（小时）</Label>
                    <Input
                      id="jwt_expiry_hours"
                      type="number"
                      value={config.jwt_expiry_hours}
                      onChange={(e) => handleConfigChange('jwt_expiry_hours', parseInt(e.target.value))}
                      min="1"
                      max="720"
                    />
                    <p className="text-sm text-muted-foreground mt-1">
                      用户登录后Token的有效时长
                    </p>
                  </div>

                  <div>
                    <Label htmlFor="refresh_token_days">刷新Token有效期（天）</Label>
                    <Input
                      id="refresh_token_days"
                      type="number"
                      value={config.refresh_token_days}
                      onChange={(e) => handleConfigChange('refresh_token_days', parseInt(e.target.value))}
                      min="1"
                      max="365"
                    />
                    <p className="text-sm text-muted-foreground mt-1">
                      刷新Token的最长有效期
                    </p>
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="enable_token_refresh">启用Token自动刷新</Label>
                      <p className="text-sm text-muted-foreground">
                        允许前端在Token即将过期时自动刷新
                      </p>
                    </div>
                    <Switch
                      id="enable_token_refresh"
                      checked={config.enable_token_refresh}
                      onCheckedChange={(checked) => handleConfigChange('enable_token_refresh', checked)}
                    />
                  </div>

                  {config.enable_token_refresh && (
                    <Alert>
                      <CheckCircle className="h-4 w-4" />
                      <AlertDescription>
                        Token自动刷新已启用。Token将在过期前5分钟自动刷新，用户无需重新登录。
                      </AlertDescription>
                    </Alert>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            {/* 通知设置 */}
            <TabsContent value="notifications" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Bell className="w-5 h-5" />
                    通知渠道
                  </CardTitle>
                  <CardDescription>配置系统通知方式</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="email_notifications">邮件通知</Label>
                      <p className="text-sm text-muted-foreground">
                        通过邮件发送系统通知
                      </p>
                    </div>
                    <Switch
                      id="email_notifications"
                      checked={config.email_notifications}
                      onCheckedChange={(checked) => handleConfigChange('email_notifications', checked)}
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="sms_notifications">短信通知</Label>
                      <p className="text-sm text-muted-foreground">
                        通过短信发送重要通知
                      </p>
                    </div>
                    <Switch
                      id="sms_notifications"
                      checked={config.sms_notifications}
                      onCheckedChange={(checked) => handleConfigChange('sms_notifications', checked)}
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="push_notifications">推送通知</Label>
                      <p className="text-sm text-muted-foreground">
                        通过浏览器推送通知
                      </p>
                    </div>
                    <Switch
                      id="push_notifications"
                      checked={config.push_notifications}
                      onCheckedChange={(checked) => handleConfigChange('push_notifications', checked)}
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <Label htmlFor="admin_notifications">管理员通知</Label>
                      <p className="text-sm text-muted-foreground">
                        向管理员发送系统事件通知
                      </p>
                    </div>
                    <Switch
                      id="admin_notifications"
                      checked={config.admin_notifications}
                      onCheckedChange={(checked) => handleConfigChange('admin_notifications', checked)}
                    />
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* 页面底部保存提示 */}
      {hasChanges && (
        <div className="fixed bottom-4 right-4 bg-white border rounded-lg shadow-lg p-4">
          <div className="flex items-center gap-3">
            <AlertTriangle className="w-5 h-5 text-yellow-500" />
            <span className="text-sm">您有未保存的更改</span>
            <Button size="sm" onClick={saveConfig} disabled={saving}>
              {saving ? '保存中...' : '保存设置'}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}