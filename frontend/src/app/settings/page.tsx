'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import Link from 'next/link'
import { 
  Settings, 
  Bell, 
  Shield, 
  Palette,
  Mail,
  User,
  ArrowLeft,
  Check,
  AlertCircle
} from 'lucide-react'

export default function SettingsPage() {
  const [activeSection, setActiveSection] = useState('notifications')
  const [notifications, setNotifications] = useState({
    email: true,
    letterReceived: true,
    letterDelivered: true,
    systemUpdates: false
  })

  const [preferences, setPreferences] = useState({
    theme: 'light',
    language: 'zh-CN',
    letterStyle: 'classic'
  })

  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null)

  const handleNotificationChange = (key: keyof typeof notifications) => {
    setNotifications(prev => ({
      ...prev,
      [key]: !prev[key]
    }))
    setMessage({ type: 'success', text: '通知设置已更新' })
    setTimeout(() => setMessage(null), 3000)
  }

  const handlePreferenceChange = (key: keyof typeof preferences, value: string) => {
    setPreferences(prev => ({
      ...prev,
      [key]: value
    }))
    setMessage({ type: 'success', text: '偏好设置已更新' })
    setTimeout(() => setMessage(null), 3000)
  }

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-4xl mx-auto px-4 py-8">
        {/* Back Button */}
        <div className="mb-8">
          <Button asChild variant="outline" size="sm" className="border-amber-300 text-amber-700 hover:bg-amber-50">
            <Link href="/profile">
              <ArrowLeft className="mr-2 h-4 w-4" />
              返回个人资料
            </Link>
          </Button>
        </div>

        {/* Header */}
        <div className="mb-8">
          <h1 className="font-serif text-3xl font-bold text-amber-900 mb-2">
            设置
          </h1>
          <p className="text-amber-700">
            管理你的账户设置和偏好
          </p>
        </div>

        {/* Message */}
        {message && (
          <Alert className="mb-6" variant={message.type === 'error' ? 'destructive' : 'default'}>
            {message.type === 'success' ? (
              <Check className="h-4 w-4" />
            ) : (
              <AlertCircle className="h-4 w-4" />
            )}
            <AlertDescription>{message.text}</AlertDescription>
          </Alert>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Settings Menu */}
          <div className="lg:col-span-1">
            <Card className="border-amber-200 bg-white shadow-lg">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Settings className="h-5 w-5" />
                  设置菜单
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <div 
                  className={`p-3 rounded-md cursor-pointer transition-all duration-200 ${
                    activeSection === 'notifications' 
                      ? 'bg-amber-100 border border-amber-300' 
                      : 'hover:bg-amber-50'
                  }`}
                  onClick={() => setActiveSection('notifications')}
                >
                  <div className="flex items-center gap-2">
                    <Bell className="h-4 w-4 text-amber-600" />
                    <span className="font-medium text-amber-900">通知设置</span>
                  </div>
                </div>
                <div 
                  className={`p-3 rounded-md cursor-pointer transition-all duration-200 ${
                    activeSection === 'preferences' 
                      ? 'bg-amber-100 border border-amber-300' 
                      : 'hover:bg-amber-50'
                  }`}
                  onClick={() => setActiveSection('preferences')}
                >
                  <div className="flex items-center gap-2">
                    <Palette className="h-4 w-4 text-amber-700" />
                    <span className="text-amber-700">偏好设置</span>
                  </div>
                </div>
                <div 
                  className={`p-3 rounded-md cursor-pointer transition-all duration-200 ${
                    activeSection === 'privacy' 
                      ? 'bg-amber-100 border border-amber-300' 
                      : 'hover:bg-amber-50'
                  }`}
                  onClick={() => setActiveSection('privacy')}
                >
                  <div className="flex items-center gap-2">
                    <Shield className="h-4 w-4 text-amber-700" />
                    <span className="text-amber-700">隐私安全</span>
                  </div>
                </div>
                <div 
                  className={`p-3 rounded-md cursor-pointer transition-all duration-200 ${
                    activeSection === 'account' 
                      ? 'bg-amber-100 border border-amber-300' 
                      : 'hover:bg-amber-50'
                  }`}
                  onClick={() => setActiveSection('account')}
                >
                  <div className="flex items-center gap-2">
                    <User className="h-4 w-4 text-amber-700" />
                    <span className="text-amber-700">账户管理</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Settings Content */}
          <div className="lg:col-span-2 space-y-6">
            {activeSection === 'notifications' && (
              <Card className="border-amber-200 bg-white shadow-lg">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Bell className="h-5 w-5" />
                    通知设置
                  </CardTitle>
                  <CardDescription>
                    管理你希望接收的通知类型
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center justify-between p-3 bg-amber-50 rounded-md border border-amber-200">
                    <div>
                      <Label className="font-medium">邮件通知</Label>
                      <p className="text-sm text-amber-700">接收重要的邮件通知</p>
                    </div>
                    <Button
                      variant={notifications.email ? "default" : "outline"}
                      size="sm"
                      onClick={() => handleNotificationChange('email')}
                      className={`transition-all duration-200 ${notifications.email 
                        ? 'bg-amber-600 hover:bg-amber-700 text-white border-amber-600' 
                        : 'border-amber-300 text-amber-700 hover:bg-amber-100'}
                      `}
                    >
                      {notifications.email ? '✓ 已开启' : '○ 已关闭'}
                    </Button>
                  </div>

                  <div className="flex items-center justify-between p-3 bg-amber-50 rounded-md border border-amber-200">
                    <div>
                      <Label className="font-medium">信件接收通知</Label>
                      <p className="text-sm text-amber-700">有新信件时通知我</p>
                    </div>
                    <Button
                      variant={notifications.letterReceived ? "default" : "outline"}
                      size="sm"
                      onClick={() => handleNotificationChange('letterReceived')}
                      className={`transition-all duration-200 ${notifications.letterReceived 
                        ? 'bg-amber-600 hover:bg-amber-700 text-white border-amber-600' 
                        : 'border-amber-300 text-amber-700 hover:bg-amber-100'}
                      `}
                    >
                      {notifications.letterReceived ? '✓ 已开启' : '○ 已关闭'}
                    </Button>
                  </div>

                  <div className="flex items-center justify-between p-3 bg-amber-50 rounded-md border border-amber-200">
                    <div>
                      <Label className="font-medium">投递状态通知</Label>
                      <p className="text-sm text-amber-700">信件投递状态更新时通知我</p>
                    </div>
                    <Button
                      variant={notifications.letterDelivered ? "default" : "outline"}
                      size="sm"
                      onClick={() => handleNotificationChange('letterDelivered')}
                      className={`transition-all duration-200 ${notifications.letterDelivered 
                        ? 'bg-amber-600 hover:bg-amber-700 text-white border-amber-600' 
                        : 'border-amber-300 text-amber-700 hover:bg-amber-100'}
                      `}
                    >
                      {notifications.letterDelivered ? '✓ 已开启' : '○ 已关闭'}
                    </Button>
                  </div>

                  <div className="flex items-center justify-between p-3 bg-amber-50 rounded-md border border-amber-200">
                    <div>
                      <Label className="font-medium">系统更新通知</Label>
                      <p className="text-sm text-amber-700">接收系统功能更新通知</p>
                    </div>
                    <Button
                      variant={notifications.systemUpdates ? "default" : "outline"}
                      size="sm"
                      onClick={() => handleNotificationChange('systemUpdates')}
                      className={`transition-all duration-200 ${notifications.systemUpdates 
                        ? 'bg-amber-600 hover:bg-amber-700 text-white border-amber-600' 
                        : 'border-amber-300 text-amber-700 hover:bg-amber-100'}
                      `}
                    >
                      {notifications.systemUpdates ? '✓ 已开启' : '○ 已关闭'}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            )}

            {activeSection === 'preferences' && (
              <Card className="border-amber-200 bg-white shadow-lg">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Palette className="h-5 w-5" />
                    偏好设置
                  </CardTitle>
                  <CardDescription>
                    自定义你的使用体验
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <Label>主题外观</Label>
                    <div className="grid grid-cols-2 gap-2">
                      <Button
                        variant="outline"
                        onClick={() => handlePreferenceChange('theme', 'light')}
                        className={preferences.theme === 'light' ? 'bg-amber-600 text-white border-amber-600' : 'border-amber-300'}
                      >
                        浅色主题
                      </Button>
                      <Button
                        variant="outline"
                        onClick={() => handlePreferenceChange('theme', 'warm')}
                        className={preferences.theme === 'warm' ? 'bg-amber-600 text-white border-amber-600' : 'border-amber-300'}
                      >
                        温暖主题
                      </Button>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label>默认信纸样式</Label>
                    <div className="grid grid-cols-2 gap-2">
                      <Button
                        variant="outline"
                        onClick={() => handlePreferenceChange('letterStyle', 'classic')}
                        className={preferences.letterStyle === 'classic' ? 'bg-amber-600 text-white border-amber-600' : 'border-amber-300'}
                      >
                        经典样式
                      </Button>
                      <Button
                        variant="outline"
                        onClick={() => handlePreferenceChange('letterStyle', 'modern')}
                        className={preferences.letterStyle === 'modern' ? 'bg-amber-600 text-white border-amber-600' : 'border-amber-300'}
                      >
                        现代样式
                      </Button>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="language">语言设置</Label>
                    <select 
                      id="language"
                      value={preferences.language}
                      onChange={(e) => handlePreferenceChange('language', e.target.value)}
                      className="w-full p-2 border border-amber-300 rounded-md bg-white"
                    >
                      <option value="zh-CN">简体中文</option>
                      <option value="zh-TW">繁體中文</option>
                      <option value="en-US">English</option>
                    </select>
                  </div>
                </CardContent>
              </Card>
            )}

            {activeSection === 'privacy' && (
              <Card className="border-amber-200 bg-white shadow-lg">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Shield className="h-5 w-5" />
                    隐私与安全
                  </CardTitle>
                  <CardDescription>
                    保护你的账户安全
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="p-4 bg-amber-50 rounded-md border border-amber-200">
                    <h3 className="font-medium text-amber-900 mb-2">密码安全</h3>
                    <p className="text-sm text-amber-700 mb-3">定期更换密码以保护账户安全</p>
                    <Button size="sm" className="bg-amber-600 hover:bg-amber-700 text-white">
                      修改密码
                    </Button>
                  </div>

                  <div className="p-4 bg-amber-50 rounded-md border border-amber-200">
                    <h3 className="font-medium text-amber-900 mb-2">隐私设置</h3>
                    <p className="text-sm text-amber-700 mb-3">管理你的个人信息显示方式</p>
                    <Button size="sm" variant="outline" className="border-amber-300 text-amber-700 hover:bg-amber-50">
                      查看隐私设置
                    </Button>
                  </div>

                  <div className="p-4 bg-red-50 rounded-md border border-red-200">
                    <h3 className="font-medium text-red-900 mb-2">危险操作</h3>
                    <p className="text-sm text-red-700 mb-3">注销账户将永久删除你的所有数据</p>
                    <Button size="sm" variant="destructive">
                      注销账户
                    </Button>
                  </div>
                </CardContent>
              </Card>
            )}

            {activeSection === 'account' && (
              <Card className="border-amber-200 bg-white shadow-lg">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <User className="h-5 w-5" />
                    账户管理
                  </CardTitle>
                  <CardDescription>
                    管理你的账户信息和设置
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="username">用户名</Label>
                    <Input 
                      id="username" 
                      type="text" 
                      placeholder="输入用户名"
                      className="border-amber-300"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="email">邮箱地址</Label>
                    <Input 
                      id="email" 
                      type="email" 
                      placeholder="输入邮箱地址"
                      className="border-amber-300"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="phone">手机号码</Label>
                    <Input 
                      id="phone" 
                      type="tel" 
                      placeholder="输入手机号码"
                      className="border-amber-300"
                    />
                  </div>

                  <div className="pt-4">
                    <Button className="bg-amber-600 hover:bg-amber-700 text-white">
                      保存更改
                    </Button>
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}