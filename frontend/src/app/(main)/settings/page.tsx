'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  User, 
  Shield, 
  Bell, 
  Key, 
  Settings as SettingsIcon,
  ArrowLeft
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useRouter } from 'next/navigation'
import { Container } from '@/components/ui/container'
import { PageHeader } from '@/components/ui/page-header'
import { PrivacySettings } from '@/components/profile/privacy-settings'
import { NotificationChannelSettings } from '@/components/settings/notification-settings'
import { SecuritySettings } from '@/components/settings/security-settings'
import ProfileSettings from './profile/page'
import { DisableTestMode } from '@/components/debug/disable-test-mode'

export default function SettingsPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [activeTab, setActiveTab] = useState('profile')
  
  // 支持URL参数切换标签页
  useEffect(() => {
    const tabParam = searchParams?.get('tab')
    if (tabParam && ['profile', 'privacy', 'notifications', 'security'].includes(tabParam)) {
      setActiveTab(tabParam)
    }
  }, [searchParams])

  const settingsTabs = [
    {
      id: 'profile',
      label: '个人资料',
      description: '管理您的基本信息、头像和联系方式',
      icon: User
    },
    {
      id: 'privacy',
      label: '隐私设置',
      description: '控制您的隐私偏好和可见性设置',
      icon: Shield
    },
    {
      id: 'notifications',
      label: '通知方式',
      description: '设置通过邮件或应用推送接收通知',
      icon: Bell
    },
    {
      id: 'security',
      label: '安全设置',
      description: '密码、两步验证和登录安全管理',
      icon: Key
    }
  ]

  return (
    <Container className="py-8">
      <DisableTestMode />
      <div className="mb-8">
        <Button
          variant="ghost"
          onClick={() => router.back()}
          className="mb-4"
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          返回
        </Button>
        
        <PageHeader
          title="设置中心"
          description="管理您的账户设置、隐私偏好和个人资料"
        />
      </div>

      <div className="max-w-6xl mx-auto">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="grid w-full grid-cols-4 lg:w-auto lg:inline-flex">
            {settingsTabs.map(tab => {
              const Icon = tab.icon
              return (
                <TabsTrigger 
                  key={tab.id} 
                  value={tab.id}
                  className="flex items-center gap-2 px-4 py-2"
                >
                  <Icon className="h-4 w-4" />
                  <span className="hidden sm:inline">{tab.label}</span>
                </TabsTrigger>
              )
            })}
          </TabsList>

          {/* 概览卡片 */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
            {settingsTabs.map(tab => {
              const Icon = tab.icon
              const isActive = activeTab === tab.id
              
              return (
                <Card 
                  key={tab.id}
                  className={`cursor-pointer transition-all hover:shadow-md ${
                    isActive ? 'ring-2 ring-primary border-primary' : ''
                  }`}
                  onClick={() => setActiveTab(tab.id)}
                >
                  <CardHeader className="pb-3">
                    <div className="flex items-center gap-3">
                      <div className={`p-2 rounded-lg ${
                        isActive ? 'bg-primary text-primary-foreground' : 'bg-muted'
                      }`}>
                        <Icon className="h-5 w-5" />
                      </div>
                      <CardTitle className="text-base">{tab.label}</CardTitle>
                    </div>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <CardDescription className="text-sm">
                      {tab.description}
                    </CardDescription>
                  </CardContent>
                </Card>
              )
            })}
          </div>

          {/* 设置内容 */}
          <TabsContent value="profile" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <User className="h-5 w-5" />
                  个人资料设置
                </CardTitle>
                <CardDescription>
                  管理您的基本信息、头像和联系方式
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ProfileSettings />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="privacy" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Shield className="h-5 w-5" />
                  隐私设置
                </CardTitle>
                <CardDescription>
                  控制您的隐私偏好和可见性设置
                </CardDescription>
              </CardHeader>
              <CardContent>
                <PrivacySettings />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="notifications" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Bell className="h-5 w-5" />
                  通知方式设置
                </CardTitle>
                <CardDescription>
                  选择您希望通过哪些渠道接收不同类型的通知
                </CardDescription>
              </CardHeader>
              <CardContent>
                <NotificationChannelSettings />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="security" className="space-y-6">
            <SecuritySettings />
          </TabsContent>
        </Tabs>
      </div>
    </Container>
  )
}