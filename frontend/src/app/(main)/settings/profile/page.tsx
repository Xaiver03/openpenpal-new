'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  User, 
  Mail, 
  School,
  Edit,
  Save,
  AlertCircle,
  CheckCircle
} from 'lucide-react'
import { useAuth } from '@/stores/user-store'
import { useUserProfile, useUserRoleInfo, useCourierInfo } from '@/hooks/use-optimized-subscriptions'
import { 
  getRoleDisplayName, 
  getRoleColors, 
  getCourierLevelName,
  type UserRole,
  type CourierLevel 
} from '@/constants/roles'
import { AvatarUpload } from '@/components/profile/avatar-upload'

export default function ProfileSettings() {
  // Optimized state subscriptions
  const { user, isAuthenticated, isLoading: authLoading, refreshUser } = useAuth()
  const { username, nickname, email, avatar, updateProfile } = useUserProfile()
  const { role } = useUserRoleInfo()
  const { courierInfo } = useCourierInfo()
  
  const [isEditing, setIsEditing] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null)
  
  const [formData, setFormData] = useState({
    nickname: '',
    email: '',
    school_code: '',
  })

  useEffect(() => {
    if (user) {
      setFormData({
        nickname: nickname || '',
        email: email || '',
        school_code: user.school_code || '',
      })
    }
  }, [user, nickname, email])

  const handleSave = async () => {
    if (!isAuthenticated) {
      setMessage({ type: 'error', text: '请先登录' })
      return
    }

    setIsLoading(true)
    setMessage(null)

    try {
      await updateProfile({
        nickname: formData.nickname,
        email: formData.email,
        school_code: formData.school_code,
      })
      
      await refreshUser()
      setIsEditing(false)
      setMessage({ type: 'success', text: '个人资料更新成功' })
    } catch (error) {
      console.error('Profile update failed:', error)
      setMessage({ 
        type: 'error', 
        text: error instanceof Error ? error.message : '更新失败，请稍后重试' 
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleCancel = () => {
    if (user) {
      setFormData({
        nickname: nickname || '',
        email: email || '',
        school_code: user.school_code || '',
      })
    }
    setIsEditing(false)
    setMessage(null)
  }

  if (authLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (!isAuthenticated || !user) {
    return (
      <Alert>
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          请先登录以访问个人资料设置
        </AlertDescription>
      </Alert>
    )
  }

  const roleColors = getRoleColors(role as UserRole)
  const roleDisplayName = getRoleDisplayName(role as UserRole)
  const courierLevel = courierInfo?.level ? getCourierLevelName(courierInfo.level as CourierLevel) : null

  return (
    <div className="space-y-6">
      {/* 消息提示 */}
      {message && (
        <Alert variant={message.type === 'error' ? 'destructive' : 'default'}>
          {message.type === 'error' ? (
            <AlertCircle className="h-4 w-4" />
          ) : (
            <CheckCircle className="h-4 w-4" />
          )}
          <AlertDescription>{message.text}</AlertDescription>
        </Alert>
      )}

      {/* 头像设置 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <User className="h-5 w-5" />
            头像设置
          </CardTitle>
          <CardDescription>
            上传您的个人头像照片
          </CardDescription>
        </CardHeader>
        <CardContent>
          <AvatarUpload currentAvatar={avatar} onAvatarChange={() => refreshUser()} />
        </CardContent>
      </Card>

      {/* 基本信息 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <User className="h-5 w-5" />
                基本信息
              </CardTitle>
              <CardDescription>
                管理您的基本个人信息
              </CardDescription>
            </div>
            {!isEditing && (
              <Button variant="outline" onClick={() => setIsEditing(true)}>
                <Edit className="h-4 w-4 mr-2" />
                编辑
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <Input
                id="username"
                value={username || ''}
                disabled
                className="bg-muted"
              />
              <p className="text-sm text-muted-foreground">用户名不可修改</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="nickname">昵称</Label>
              <Input
                id="nickname"
                value={formData.nickname}
                onChange={(e) => setFormData(prev => ({ ...prev, nickname: e.target.value }))}
                disabled={!isEditing}
                placeholder="请输入昵称"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="email">邮箱</Label>
              <Input
                id="email"
                type="email"
                value={formData.email}
                onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
                disabled={!isEditing}
                placeholder="请输入邮箱地址"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="school_code">学校代码</Label>
              <Input
                id="school_code"
                value={formData.school_code}
                onChange={(e) => setFormData(prev => ({ ...prev, school_code: e.target.value }))}
                disabled={!isEditing}
                placeholder="请输入学校代码"
              />
            </div>
          </div>

          {isEditing && (
            <div className="flex gap-2 pt-4">
              <Button onClick={handleSave} disabled={isLoading}>
                <Save className="h-4 w-4 mr-2" />
                {isLoading ? '保存中...' : '保存'}
              </Button>
              <Button variant="outline" onClick={handleCancel} disabled={isLoading}>
                取消
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 角色信息 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <School className="h-5 w-5" />
            角色信息
          </CardTitle>
          <CardDescription>
            您在系统中的角色和权限信息
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center gap-3 p-4 rounded-lg border bg-card">
              <div className={`w-3 h-3 rounded-full ${roleColors.bg}`}></div>
              <div>
                <p className="font-medium">{roleDisplayName}</p>
                {courierLevel && (
                  <p className="text-sm text-muted-foreground">
                    信使等级：{courierLevel}
                  </p>
                )}
                {courierInfo?.zoneCode && (
                  <p className="text-sm text-muted-foreground">
                    管理区域：{courierInfo.zoneCode}
                  </p>
                )}
              </div>
            </div>

            {courierInfo && (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
                <div className="p-3 rounded-lg bg-muted/50 text-center">
                  <p className="text-2xl font-bold text-primary">{courierInfo.points || 0}</p>
                  <p className="text-sm text-muted-foreground">总积分</p>
                </div>
                <div className="p-3 rounded-lg bg-muted/50 text-center">
                  <p className="text-2xl font-bold text-green-600">{courierInfo.completedTasks || 0}</p>
                  <p className="text-sm text-muted-foreground">完成任务</p>
                </div>
                <div className="p-3 rounded-lg bg-muted/50 text-center">
                  <p className="text-2xl font-bold text-blue-600">{courierInfo.averageRating || 0}</p>
                  <p className="text-sm text-muted-foreground">平均评分</p>
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}