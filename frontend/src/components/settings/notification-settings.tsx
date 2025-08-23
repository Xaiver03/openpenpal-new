'use client'

import { useState, useEffect } from 'react'
import { Bell, Mail, MessageSquare, UserPlus, Award, Package, AlertCircle, Loader2 } from 'lucide-react'
import { Switch } from '@/components/ui/switch'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { useToast } from '@/components/ui/use-toast'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import { notificationApi } from '@/lib/api/notification'
import type { NotificationType } from '@/types/notification'
import type { NotificationChannelPreferences } from '@/types/notification-channel'

interface NotificationCategory {
  id: NotificationType
  name: string
  description: string
  icon: React.ComponentType<{ className?: string }>
  color: string
}

const notificationCategories: NotificationCategory[] = [
  {
    id: 'follow',
    name: '关注通知',
    description: '当有人关注您时收到通知',
    icon: UserPlus,
    color: 'text-blue-600'
  },
  {
    id: 'comment',
    name: '评论通知',
    description: '当有人评论您的内容时收到通知',
    icon: MessageSquare,
    color: 'text-green-600'
  },
  {
    id: 'comment_reply',
    name: '回复通知',
    description: '当有人回复您的评论时收到通知',
    icon: MessageSquare,
    color: 'text-green-600'
  },
  {
    id: 'letter_received',
    name: '信件通知',
    description: '当您收到新信件时收到通知',
    icon: Mail,
    color: 'text-purple-600'
  },
  {
    id: 'achievement',
    name: '成就通知',
    description: '当您获得新成就时收到通知',
    icon: Award,
    color: 'text-yellow-600'
  },
  {
    id: 'courier_task',
    name: '信使任务',
    description: '信使任务相关通知',
    icon: Package,
    color: 'text-orange-600'
  },
  {
    id: 'system',
    name: '系统通知',
    description: '重要的系统消息和更新',
    icon: AlertCircle,
    color: 'text-red-600'
  }
]

export function NotificationChannelSettings() {
  const { toast } = useToast()
  const [localPreferences, setLocalPreferences] = useState<NotificationChannelPreferences | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)

  // 加载通知渠道偏好
  useEffect(() => {
    loadPreferences()
  }, [])

  const loadPreferences = async () => {
    setIsLoading(true)
    try {
      const response = await notificationApi.getPreferences()
      // 转换后端格式到前端格式
      const channelPrefs: NotificationChannelPreferences = {
        email: {} as Record<NotificationType, boolean>,
        push: {} as Record<NotificationType, boolean>
      }
      
      // 根据后端的types字段初始化
      const notificationTypes: NotificationType[] = [
        'follow', 'comment', 'comment_reply', 'like',
        'letter_received', 'achievement', 'courier_task', 'system'
      ]
      
      notificationTypes.forEach(type => {
        channelPrefs.email[type] = response.types?.[type] ?? true
        channelPrefs.push[type] = response.types?.[type] ?? true
      })
      
      setLocalPreferences(channelPrefs)
    } catch (error) {
      console.error('Failed to load notification preferences:', error)
      // 使用默认值
      const defaultPrefs: NotificationChannelPreferences = {
        email: {} as Record<NotificationType, boolean>,
        push: {} as Record<NotificationType, boolean>
      }
      
      const notificationTypes: NotificationType[] = [
        'follow', 'comment', 'comment_reply', 'like',
        'letter_received', 'achievement', 'courier_task', 'system'
      ]
      
      notificationTypes.forEach(type => {
        defaultPrefs.email[type] = true
        defaultPrefs.push[type] = true
      })
      
      setLocalPreferences(defaultPrefs)
      toast({
        title: '加载失败',
        description: '无法加载通知设置，使用默认设置',
        variant: 'destructive'
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleToggle = (category: NotificationType, field: 'email' | 'push') => {
    if (!localPreferences) return

    const newPreferences = {
      ...localPreferences,
      [field]: {
        ...localPreferences[field],
        [category]: !localPreferences[field][category]
      }
    }
    setLocalPreferences(newPreferences)
  }

  const handleSave = async () => {
    if (!localPreferences) return

    setIsSaving(true)
    try {
      // 转换为后端期望的格式
      const types: Record<string, boolean> = {}
      Object.keys(localPreferences.email).forEach(key => {
        types[key] = localPreferences.email[key as NotificationType] && localPreferences.push[key as NotificationType]
      })
      
      await notificationApi.updatePreferences({
        email_enabled: Object.values(localPreferences.email).some(v => v),
        push_enabled: Object.values(localPreferences.push).some(v => v),
        types: types as {
          follow: boolean;
          comment: boolean;
          comment_reply: boolean;
          like: boolean;
          letter_received: boolean;
          achievement: boolean;
          courier_task: boolean;
          system: boolean;
        }
      })
      
      toast({
        title: '保存成功',
        description: '您的通知设置已更新',
      })
    } catch (error) {
      console.error('Failed to save notification preferences:', error)
      toast({
        title: '保存失败',
        description: '无法保存通知设置，请稍后重试',
        variant: 'destructive'
      })
    } finally {
      setIsSaving(false)
    }
  }

  const handleReset = () => {
    loadPreferences()
    toast({
      title: '已重置',
      description: '通知设置已恢复到上次保存的状态',
    })
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!localPreferences) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">无法加载通知设置</p>
        <Button onClick={loadPreferences} className="mt-4">
          重新加载
        </Button>
      </div>
    )
  }

  const [originalPreferences, setOriginalPreferences] = useState<NotificationChannelPreferences | null>(null)
  
  useEffect(() => {
    if (localPreferences && !originalPreferences) {
      setOriginalPreferences(JSON.parse(JSON.stringify(localPreferences)))
    }
  }, [localPreferences, originalPreferences])
  
  const hasChanges = originalPreferences ? JSON.stringify(localPreferences) !== JSON.stringify(originalPreferences) : false

  return (
    <div className="space-y-6">
      {/* 全局设置 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">通知接收方式</CardTitle>
          <CardDescription>
            选择您希望通过哪些渠道接收通知
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="all-email">邮件通知</Label>
              <p className="text-sm text-muted-foreground">
                通过邮件接收所有通知
              </p>
            </div>
            <Switch
              id="all-email"
              checked={Object.values(localPreferences.email).some(v => v)}
              onCheckedChange={(checked) => {
                const newPreferences = { ...localPreferences }
                Object.keys(newPreferences.email).forEach(key => {
                  newPreferences.email[key as NotificationType] = checked
                })
                setLocalPreferences(newPreferences)
              }}
            />
          </div>
          <Separator />
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="all-push">推送通知</Label>
              <p className="text-sm text-muted-foreground">
                在应用内接收实时通知
              </p>
            </div>
            <Switch
              id="all-push"
              checked={Object.values(localPreferences.push).some(v => v)}
              onCheckedChange={(checked) => {
                const newPreferences = { ...localPreferences }
                Object.keys(newPreferences.push).forEach(key => {
                  newPreferences.push[key as NotificationType] = checked
                })
                setLocalPreferences(newPreferences)
              }}
            />
          </div>
        </CardContent>
      </Card>

      {/* 分类设置 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">通知类别</CardTitle>
          <CardDescription>
            为每种通知类型设置接收方式
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-6">
            {notificationCategories.map((category, index) => {
              const Icon = category.icon
              return (
                <div key={category.id}>
                  {index > 0 && <Separator className="mb-6" />}
                  <div className="space-y-4">
                    <div className="flex items-start space-x-3">
                      <Icon className={cn('h-5 w-5 mt-0.5', category.color)} />
                      <div className="flex-1 space-y-1">
                        <p className="font-medium">{category.name}</p>
                        <p className="text-sm text-muted-foreground">
                          {category.description}
                        </p>
                      </div>
                    </div>
                    <div className="ml-8 space-y-3">
                      <div className="flex items-center justify-between">
                        <Label 
                          htmlFor={`${category.id}-email`}
                          className="text-sm font-normal cursor-pointer"
                        >
                          邮件通知
                        </Label>
                        <Switch
                          id={`${category.id}-email`}
                          checked={localPreferences.email[category.id]}
                          onCheckedChange={() => handleToggle(category.id, 'email')}
                        />
                      </div>
                      <div className="flex items-center justify-between">
                        <Label 
                          htmlFor={`${category.id}-push`}
                          className="text-sm font-normal cursor-pointer"
                        >
                          推送通知
                        </Label>
                        <Switch
                          id={`${category.id}-push`}
                          checked={localPreferences.push[category.id]}
                          onCheckedChange={() => handleToggle(category.id, 'push')}
                        />
                      </div>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </CardContent>
      </Card>

      {/* 操作按钮 */}
      <div className="flex items-center justify-between">
        <Button
          variant="outline"
          onClick={handleReset}
          disabled={!hasChanges || isSaving}
        >
          重置更改
        </Button>
        <Button
          onClick={handleSave}
          disabled={!hasChanges || isSaving}
        >
          {isSaving && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
          保存设置
        </Button>
      </div>
    </div>
  )
}