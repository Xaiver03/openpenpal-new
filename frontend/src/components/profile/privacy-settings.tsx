'use client'

/**
 * Privacy Settings Component - Comprehensive profile privacy controls
 * Based on SOTA patterns with modern UX design
 */

import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { 
  Shield, 
  Eye, 
  EyeOff, 
  Bell, 
  BellOff, 
  UserMinus, 
  Volume2, 
  VolumeX,
  KeyRound,
  Trash2,
  Plus,
  AlertTriangle,
  Info,
  Settings
} from 'lucide-react'

import { privacyApi } from '../../lib/api/privacy'
import type {
  PrivacySettings,
  PrivacyLevel,
  ProfileVisibility,
  SocialPrivacy,
  NotificationPrivacy,
  BlockingSettings
} from '../../types/privacy'
import {
  PRIVACY_LEVEL_LABELS,
  PRIVACY_LEVEL_DESCRIPTIONS,
  PROFILE_FIELD_LABELS
} from '../../types/privacy'
import { toast } from 'sonner'

interface PrivacySettingsProps {
  className?: string
}

export function PrivacySettings({ className }: PrivacySettingsProps) {
  const [settings, setSettings] = useState<PrivacySettings | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [activeTab, setActiveTab] = useState('profile')
  const [newKeyword, setNewKeyword] = useState('')

  // Load privacy settings on component mount
  useEffect(() => {
    loadPrivacySettings()
  }, [])

  const loadPrivacySettings = async () => {
    try {
      setLoading(true)
      const privacySettings = await privacyApi.getPrivacySettings()
      setSettings(privacySettings)
    } catch (error) {
      console.error('Failed to load privacy settings:', error)
      toast.error('加载隐私设置失败')
    } finally {
      setLoading(false)
    }
  }

  const saveSettings = async (updates: Partial<PrivacySettings>) => {
    if (!settings) return

    try {
      setSaving(true)
      const updatedSettings = await privacyApi.updatePrivacySettings(updates)
      setSettings(updatedSettings)
      toast.success('隐私设置已更新')
    } catch (error) {
      console.error('Failed to update privacy settings:', error)
      toast.error('更新隐私设置失败')
    } finally {
      setSaving(false)
    }
  }

  const updateProfileVisibility = (field: keyof ProfileVisibility, level: PrivacyLevel) => {
    if (!settings) return
    
    const updates = {
      profile_visibility: {
        ...settings.profile_visibility,
        [field]: level
      }
    }
    saveSettings(updates)
  }

  const updateSocialPrivacy = (field: keyof SocialPrivacy, value: boolean) => {
    if (!settings) return
    
    const updates = {
      social_privacy: {
        ...settings.social_privacy,
        [field]: value
      }
    }
    saveSettings(updates)
  }

  const updateNotificationPrivacy = (field: keyof NotificationPrivacy, value: boolean) => {
    if (!settings) return
    
    const updates = {
      notification_privacy: {
        ...settings.notification_privacy,
        [field]: value
      }
    }
    saveSettings(updates)
  }

  const addBlockedKeyword = async () => {
    if (!newKeyword.trim()) return

    try {
      await privacyApi.addBlockedKeyword(newKeyword.trim())
      setNewKeyword('')
      await loadPrivacySettings() // Reload to get updated keywords
      toast.success('已添加屏蔽关键词')
    } catch (error) {
      console.error('Failed to add blocked keyword:', error)
      toast.error('添加屏蔽关键词失败')
    }
  }

  const removeBlockedKeyword = async (keyword: string) => {
    try {
      await privacyApi.removeBlockedKeyword(keyword)
      await loadPrivacySettings() // Reload to get updated keywords
      toast.success('已移除屏蔽关键词')
    } catch (error) {
      console.error('Failed to remove blocked keyword:', error)
      toast.error('移除屏蔽关键词失败')
    }
  }

  const handleBlockUser = async (userId: string) => {
    try {
      await privacyApi.blockUser(userId)
      await loadPrivacySettings()
      toast.success('已屏蔽用户')
    } catch (error) {
      console.error('Failed to block user:', error)
      toast.error('屏蔽用户失败')
    }
  }

  const handleUnblockUser = async (userId: string) => {
    try {
      await privacyApi.unblockUser(userId)
      await loadPrivacySettings()
      toast.success('已取消屏蔽用户')
    } catch (error) {
      console.error('Failed to unblock user:', error)
      toast.error('取消屏蔽用户失败')
    }
  }

  const resetToDefaults = async () => {
    try {
      setSaving(true)
      const defaultSettings = await privacyApi.resetPrivacySettings()
      setSettings(defaultSettings)
      toast.success('隐私设置已重置为默认值')
    } catch (error) {
      console.error('Failed to reset privacy settings:', error)
      toast.error('重置隐私设置失败')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <Card className={className}>
        <CardContent className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </CardContent>
      </Card>
    )
  }

  if (!settings) {
    return (
      <Card className={className}>
        <CardContent className="flex items-center justify-center py-12">
          <p className="text-gray-500">无法加载隐私设置</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          隐私设置
        </CardTitle>
        <CardDescription>
          管理您的隐私偏好和内容可见性设置
        </CardDescription>
      </CardHeader>
      
      <CardContent>
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="profile" className="flex items-center gap-2">
              <Eye className="h-4 w-4" />
              档案可见性
            </TabsTrigger>
            <TabsTrigger value="social" className="flex items-center gap-2">
              <Settings className="h-4 w-4" />
              社交设置
            </TabsTrigger>
            <TabsTrigger value="notifications" className="flex items-center gap-2">
              <Bell className="h-4 w-4" />
              通知设置
            </TabsTrigger>
            <TabsTrigger value="blocking" className="flex items-center gap-2">
              <UserMinus className="h-4 w-4" />
              屏蔽管理
            </TabsTrigger>
          </TabsList>

          {/* Profile Visibility Tab */}
          <TabsContent value="profile" className="space-y-6">
            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription>
                控制谁可以看到您的个人资料信息。更严格的设置可以更好地保护您的隐私。
              </AlertDescription>
            </Alert>

            <div className="space-y-4">
              {Object.entries(settings.profile_visibility).map(([field, currentLevel]) => (
                <div key={field} className="flex items-center justify-between">
                  <div>
                    <Label className="text-sm font-medium">
                      {PROFILE_FIELD_LABELS[field as keyof ProfileVisibility]}
                    </Label>
                    <p className="text-sm text-gray-500 mt-1">
                      {PRIVACY_LEVEL_DESCRIPTIONS[currentLevel as PrivacyLevel]}
                    </p>
                  </div>
                  <Select
                    value={currentLevel}
                    onValueChange={(level: PrivacyLevel) => 
                      updateProfileVisibility(field as keyof ProfileVisibility, level)
                    }
                  >
                    <SelectTrigger className="w-32">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {Object.entries(PRIVACY_LEVEL_LABELS).map(([level, label]) => (
                        <SelectItem key={level} value={level}>
                          {label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              ))}
            </div>
          </TabsContent>

          {/* Social Privacy Tab */}
          <TabsContent value="social" className="space-y-6">
            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription>
                管理其他用户如何与您互动，以及您在平台上的可发现性。
              </AlertDescription>
            </Alert>

            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label>允许关注请求</Label>
                  <p className="text-sm text-gray-500">其他用户可以向您发送关注请求</p>
                </div>
                <Switch
                  checked={settings.social_privacy.allow_follow_requests}
                  onCheckedChange={(checked) => 
                    updateSocialPrivacy('allow_follow_requests', checked)
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label>允许评论</Label>
                  <p className="text-sm text-gray-500">其他用户可以在您的个人资料上留言</p>
                </div>
                <Switch
                  checked={settings.social_privacy.allow_comments}
                  onCheckedChange={(checked) => 
                    updateSocialPrivacy('allow_comments', checked)
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label>允许私信</Label>
                  <p className="text-sm text-gray-500">其他用户可以向您发送私人消息</p>
                </div>
                <Switch
                  checked={settings.social_privacy.allow_direct_messages}
                  onCheckedChange={(checked) => 
                    updateSocialPrivacy('allow_direct_messages', checked)
                  }
                />
              </div>

              <Separator />

              <div className="flex items-center justify-between">
                <div>
                  <Label>出现在发现页面</Label>
                  <p className="text-sm text-gray-500">您的资料会出现在用户发现推荐中</p>
                </div>
                <Switch
                  checked={settings.social_privacy.show_in_discovery}
                  onCheckedChange={(checked) => 
                    updateSocialPrivacy('show_in_discovery', checked)
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label>出现在推荐中</Label>
                  <p className="text-sm text-gray-500">系统会将您推荐给其他用户</p>
                </div>
                <Switch
                  checked={settings.social_privacy.show_in_suggestions}
                  onCheckedChange={(checked) => 
                    updateSocialPrivacy('show_in_suggestions', checked)
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label>允许同校搜索</Label>
                  <p className="text-sm text-gray-500">同校用户可以通过用户名搜索到您</p>
                </div>
                <Switch
                  checked={settings.social_privacy.allow_school_search}
                  onCheckedChange={(checked) => 
                    updateSocialPrivacy('allow_school_search', checked)
                  }
                />
              </div>
            </div>
          </TabsContent>

          {/* Notification Privacy Tab */}
          <TabsContent value="notifications" className="space-y-6">
            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription>
                选择您希望接收哪些类型的通知。
              </AlertDescription>
            </Alert>

            <div className="space-y-4">
              {Object.entries({
                new_followers: '新粉丝',
                follow_requests: '关注请求',
                comments: '评论',
                mentions: '提及',
                direct_messages: '私信',
                system_updates: '系统更新',
                email_notifications: '邮件通知'
              }).map(([field, label]) => (
                <div key={field} className="flex items-center justify-between">
                  <div>
                    <Label>{label}</Label>
                    <p className="text-sm text-gray-500">
                      接收{label.toLowerCase()}的推送通知
                    </p>
                  </div>
                  <Switch
                    checked={settings.notification_privacy[field as keyof NotificationPrivacy]}
                    onCheckedChange={(checked) => 
                      updateNotificationPrivacy(field as keyof NotificationPrivacy, checked)
                    }
                  />
                </div>
              ))}
            </div>
          </TabsContent>

          {/* Blocking Management Tab */}
          <TabsContent value="blocking" className="space-y-6">
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>
                管理屏蔽的用户和内容过滤规则。被屏蔽的用户将无法看到您的资料或与您互动。
              </AlertDescription>
            </Alert>

            {/* Blocked Keywords Section */}
            <div className="space-y-4">
              <h4 className="text-lg font-medium">屏蔽关键词</h4>
              
              <div className="flex gap-2">
                <Input
                  placeholder="添加要屏蔽的关键词..."
                  value={newKeyword}
                  onChange={(e) => setNewKeyword(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && addBlockedKeyword()}
                />
                <Button onClick={addBlockedKeyword} size="sm">
                  <Plus className="h-4 w-4" />
                </Button>
              </div>

              <div className="flex flex-wrap gap-2">
                {settings.blocking_settings.blocked_keywords.map((keyword) => (
                  <Badge key={keyword} variant="secondary" className="flex items-center gap-1">
                    {keyword}
                    <button
                      onClick={() => removeBlockedKeyword(keyword)}
                      className="ml-1 hover:text-red-500"
                    >
                      <Trash2 className="h-3 w-3" />
                    </button>
                  </Badge>
                ))}
              </div>
            </div>

            <Separator />

            {/* Auto-block Settings */}
            <div className="space-y-4">
              <h4 className="text-lg font-medium">自动屏蔽设置</h4>
              
              <div className="flex items-center justify-between">
                <div>
                  <Label>自动屏蔽新账户</Label>
                  <p className="text-sm text-gray-500">
                    自动屏蔽创建时间少于7天的账户
                  </p>
                </div>
                <Switch
                  checked={settings.blocking_settings.auto_block_new_accounts}
                  onCheckedChange={(checked) =>
                    saveSettings({
                      blocking_settings: {
                        ...settings.blocking_settings,
                        auto_block_new_accounts: checked
                      }
                    })
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label>屏蔽非同校用户</Label>
                  <p className="text-sm text-gray-500">
                    自动屏蔽不是同校的用户
                  </p>
                </div>
                <Switch
                  checked={settings.blocking_settings.block_non_school_users}
                  onCheckedChange={(checked) =>
                    saveSettings({
                      blocking_settings: {
                        ...settings.blocking_settings,
                        block_non_school_users: checked
                      }
                    })
                  }
                />
              </div>
            </div>

            <Separator />

            {/* Blocked Users List */}
            <div className="space-y-4">
              <h4 className="text-lg font-medium">已屏蔽的用户</h4>
              {settings.blocking_settings.blocked_users.length === 0 ? (
                <p className="text-gray-500 text-sm">暂无屏蔽用户</p>
              ) : (
                <div className="space-y-2">
                  {settings.blocking_settings.blocked_users.map((userId) => (
                    <div key={userId} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                      <span className="text-sm">{userId}</span>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleUnblockUser(userId)}
                      >
                        取消屏蔽
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>

        {/* Action Buttons */}
        <div className="flex justify-between pt-6 border-t">
          <Button
            variant="outline"
            onClick={resetToDefaults}
            disabled={saving}
          >
            重置为默认
          </Button>
          <div className="text-sm text-gray-500">
            {saving ? '保存中...' : '设置会自动保存'}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export default PrivacySettings