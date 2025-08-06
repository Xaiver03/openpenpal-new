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

export default function ProfilePage() {
  // Optimized state subscriptions
  const { user, isAuthenticated, isLoading: authLoading, refreshUser } = useAuth()
  const { username, nickname, email, updateProfile } = useUserProfile()
  const { role } = useUserRoleInfo()
  const { courierInfo } = useCourierInfo()
  
  const [isEditing, setIsEditing] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null)
  
  const [formData, setFormData] = useState({
    nickname: '',
    email: '',
    schoolCode: '',
  })

  useEffect(() => {
    if (user) {
      setFormData({
        nickname: nickname || '',
        email: email || '',
        schoolCode: user.schoolCode || '',
      })
    }
  }, [user, nickname, email])

  const handleEdit = () => {
    setIsEditing(true)
    setMessage(null)
  }

  const handleCancel = () => {
    setIsEditing(false)
    if (user) {
      setFormData({
        nickname: nickname || '',
        email: email || '',
        schoolCode: user.schoolCode || '',
      })
    }
    setMessage(null)
  }

  const handleSave = async () => {
    setIsLoading(true)
    setMessage(null)
    
    try {
      // Use optimistic update from store
      await updateProfile({
        nickname: formData.nickname,
        email: formData.email,
        schoolCode: formData.schoolCode
      })
      
      setIsEditing(false)
      setMessage({ type: 'success', text: '个人信息更新成功！' })
    } catch (error) {
      setMessage({ type: 'error', text: '更新失败，请稍后重试' })
    } finally {
      setIsLoading(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({
      ...prev,
      [e.target.name]: e.target.value
    }))
  }

  // 处理认证加载状态
  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-amber-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-amber-600 mx-auto mb-4"></div>
          <p className="text-amber-700">加载中...</p>
        </div>
      </div>
    )
  }
  
  // 处理未登录状态
  if (!isAuthenticated || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-amber-50">
        <Card className="w-full max-w-md mx-4 border-amber-200 bg-white">
          <CardContent className="text-center py-8">
            <User className="h-12 w-12 text-amber-600 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-amber-900 mb-2">请先登录</h2>
            <p className="text-amber-700 mb-4">您需要登录后才能查看个人资料</p>
            <Button 
              onClick={() => window.location.href = '/login'}
              className="bg-amber-600 hover:bg-amber-700 text-white"
            >
              去登录
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-2xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-amber-900 mb-2">
          个人资料
        </h1>
        <p className="text-amber-700">
          管理你的个人信息和账户设置
        </p>
      </div>

      {/* Profile Card */}
      <Card className="border-amber-200 bg-white shadow-lg">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <User className="h-5 w-5" />
                个人信息
              </CardTitle>
              <CardDescription>
                你的基本信息和联系方式
              </CardDescription>
            </div>
            {!isEditing && (
              <Button variant="outline" onClick={handleEdit}>
                <Edit className="mr-2 h-4 w-4" />
                编辑
              </Button>
            )}
          </div>
        </CardHeader>
        
        <CardContent className="space-y-6">
          {message && (
            <Alert variant={message.type === 'error' ? 'destructive' : 'default'}>
              {message.type === 'success' ? (
                <CheckCircle className="h-4 w-4" />
              ) : (
                <AlertCircle className="h-4 w-4" />
              )}
              <AlertDescription>{message.text}</AlertDescription>
            </Alert>
          )}

          {/* 用户名 (只读) */}
          <div className="space-y-2">
            <Label>用户名</Label>
            <div className="flex items-center gap-2 p-3 bg-amber-50 border border-amber-200 rounded-md">
              <User className="h-4 w-4 text-amber-700" />
              <span className="font-mono">{user.username}</span>
            </div>
            <p className="text-sm text-amber-700">用户名无法修改</p>
          </div>

          {/* 昵称 */}
          <div className="space-y-2">
            <Label htmlFor="nickname">昵称</Label>
            {isEditing ? (
              <Input
                id="nickname"
                name="nickname"
                placeholder="请输入昵称"
                value={formData.nickname}
                onChange={handleChange}
              />
            ) : (
              <div className="p-3 bg-amber-50 border border-amber-200 rounded-md">
                {user.nickname || '未设置昵称'}
              </div>
            )}
          </div>

          {/* 邮箱 */}
          <div className="space-y-2">
            <Label htmlFor="email">邮箱</Label>
            {isEditing ? (
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-amber-700" />
                <Input
                  id="email"
                  name="email"
                  type="email"
                  placeholder="请输入邮箱"
                  value={formData.email}
                  onChange={handleChange}
                  className="pl-10"
                />
              </div>
            ) : (
              <div className="flex items-center gap-2 p-3 bg-amber-50 border border-amber-200 rounded-md">
                <Mail className="h-4 w-4 text-amber-700" />
                <span>{user.email || '未设置邮箱'}</span>
              </div>
            )}
          </div>

          {/* 学校代码 */}
          <div className="space-y-2">
            <Label htmlFor="school_code">学校代码</Label>
            {isEditing ? (
              <div className="relative">
                <School className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-amber-700" />
                <Input
                  id="school_code"
                  name="school_code"
                  placeholder="请输入学校代码"
                  value={formData.schoolCode}
                  onChange={handleChange}
                  className="pl-10"
                />
              </div>
            ) : (
              <div className="flex items-center gap-2 p-3 bg-amber-50 border border-amber-200 rounded-md">
                <School className="h-4 w-4 text-amber-700" />
                <span className="font-mono">{user.schoolCode || '未设置学校代码'}</span>
              </div>
            )}
          </div>

          {/* 角色 (只读) */}
          <div className="space-y-2">
            <Label>角色</Label>
            <div className="p-3 bg-amber-50 border border-amber-200 rounded-md">
              <span className={`px-2 py-1 rounded text-sm ${getRoleColors(user.role as UserRole).badge}`}>
                {(user.role.includes('courier') || user.courierInfo?.level) ? 
                  getCourierLevelName(user.courierInfo?.level as CourierLevel) || getRoleDisplayName(user.role as UserRole) :
                  getRoleDisplayName(user.role as UserRole)
                }
              </span>
            </div>
          </div>

          {/* 信使层级信息 (仅信使账号显示) */}
          {(user.role.includes('courier') || user.courierInfo) && user.courierInfo && (
            <div className="space-y-2">
              <Label>信使层级详情</Label>
              <div className="p-3 bg-amber-50 border border-amber-200 rounded-md space-y-2">
                <div className="flex justify-between items-center">
                  <span className="text-sm text-amber-700">层级等级:</span>
                  <span className="font-mono text-sm">{user.courierInfo.level}级</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-amber-700">管辖区域:</span>
                  <span className="font-mono text-sm">{user.courierInfo.zoneCode}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-amber-700">区域类型:</span>
                  <span className="text-sm">
                    {user.courierInfo.zoneType === 'city' ? '城市' :
                     user.courierInfo.zoneType === 'school' ? '学校' :
                     user.courierInfo.zoneType === 'zone' ? '片区' :
                     user.courierInfo.zoneType === 'building' ? '楼栋' : '未知'}
                  </span>
                </div>
                {user.courierInfo.points && (
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-amber-700">信使积分:</span>
                    <span className="font-mono text-sm text-green-600">{user.courierInfo.points}</span>
                  </div>
                )}
                {user.courierInfo.taskCount && (
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-amber-700">完成任务:</span>
                    <span className="font-mono text-sm text-blue-600">{user.courierInfo.taskCount}</span>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* 账户状态 (只读) */}
          <div className="space-y-2">
            <Label>账户状态</Label>
            <div className="p-3 bg-amber-50 border border-amber-200 rounded-md">
              <span className={`px-2 py-1 rounded text-sm ${
                user.isActive ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
              }`}>
                {user.isActive ? '已激活' : '未激活'}
              </span>
            </div>
          </div>

          {/* 注册时间 (只读) */}
          <div className="space-y-2">
            <Label>注册时间</Label>
            <div className="p-3 bg-amber-50 border border-amber-200 rounded-md">
              {new Date(user.createdAt).toLocaleString('zh-CN')}
            </div>
          </div>

          {/* 操作按钮 */}
          {isEditing && (
            <div className="flex gap-4 pt-4">
              <Button 
                onClick={handleSave} 
                disabled={isLoading}
                className="flex-1 bg-amber-600 hover:bg-amber-700 text-white"
              >
                {isLoading ? (
                  <>
                    <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                    保存中...
                  </>
                ) : (
                  <>
                    <Save className="mr-2 h-4 w-4" />
                    保存
                  </>
                )}
              </Button>
              <Button 
                variant="outline" 
                onClick={handleCancel}
                disabled={isLoading}
                className="flex-1 border-amber-300 text-amber-700 hover:bg-amber-50"
              >
                取消
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 账户统计 */}
      <Card className="mt-6 border-amber-200 bg-white shadow-lg">
        <CardHeader>
          <CardTitle>账户统计</CardTitle>
          <CardDescription>你在OpenPenPal的活动统计</CardDescription>
        </CardHeader>
        <CardContent className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="text-center p-4 bg-amber-50 border border-amber-200 rounded-lg">
            <div className="text-2xl font-bold text-amber-600">0</div>
            <div className="text-sm text-amber-700">已发送</div>
          </div>
          <div className="text-center p-4 bg-amber-50 border border-amber-200 rounded-lg">
            <div className="text-2xl font-bold text-amber-600">0</div>
            <div className="text-sm text-amber-700">已接收</div>
          </div>
          <div className="text-center p-4 bg-amber-50 border border-amber-200 rounded-lg">
            <div className="text-2xl font-bold text-amber-600">0</div>
            <div className="text-sm text-amber-700">草稿</div>
          </div>
          <div className="text-center p-4 bg-amber-50 border border-amber-200 rounded-lg">
            <div className="text-2xl font-bold text-amber-600">0</div>
            <div className="text-sm text-amber-700">已读</div>
          </div>
        </CardContent>
      </Card>
      </div>
    </div>
  )
}