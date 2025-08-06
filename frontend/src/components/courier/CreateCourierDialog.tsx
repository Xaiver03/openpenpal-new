'use client'

import { useState } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { UserPlus, Building, School, Crown, Home } from 'lucide-react'
import { useCourierPermission } from '@/hooks/use-courier-permission'

interface CreateCourierDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
  targetLevel: 1 | 2 | 3 // 要创建的信使级别
}

interface CourierFormData {
  username: string
  email: string
  password: string
  realName: string
  phone: string
  zoneCode: string
  zoneName: string
  description: string
}

const LEVEL_CONFIG = {
  1: {
    icon: Home,
    name: '一级信使',
    title: '楼栋/班级信使',
    description: '负责具体楼栋或班级的信件收发',
    zoneLabel: '楼栋/班级',
    zonePlaceholder: '如：1号楼、计算机201班',
    color: 'bg-yellow-600'
  },
  2: {
    icon: Building,
    name: '二级信使', 
    title: '片区/年级信使',
    description: '管理片区内的一级信使，负责区域协调',
    zoneLabel: '片区/年级',
    zonePlaceholder: '如：A区、计算机学院、2024级',
    color: 'bg-orange-600'
  },
  3: {
    icon: School,
    name: '三级信使',
    title: '校级信使',
    description: '管理学校内的二级信使，负责校园协调',
    zoneLabel: '学校/校区',
    zonePlaceholder: '如：北京大学、清华大学',
    color: 'bg-amber-600'
  }
}

export function CreateCourierDialog({ 
  open, 
  onOpenChange,
  onSuccess,
  targetLevel 
}: CreateCourierDialogProps) {
  const { courierInfo, hasCourierPermission, COURIER_PERMISSIONS } = useCourierPermission()
  
  const [formData, setFormData] = useState<CourierFormData>({
    username: '',
    email: '',
    password: '',
    realName: '',
    phone: '',
    zoneCode: '',
    zoneName: '',
    description: ''
  })
  
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errors, setErrors] = useState<Partial<CourierFormData>>({})

  const config = LEVEL_CONFIG[targetLevel]
  const IconComponent = config.icon

  // 权限检查
  const canCreateCourier = hasCourierPermission('CREATE_SUBORDINATE')
  
  if (!canCreateCourier || !courierInfo) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="text-red-600">权限不足</DialogTitle>
            <DialogDescription>
              您没有创建下级信使的权限
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    )
  }

  // 级别权限检查
  if (courierInfo.level <= targetLevel) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="text-red-600">级别权限不足</DialogTitle>
            <DialogDescription>
              您只能创建比自己级别更低的信使。当前您是{courierInfo.level}级信使，无法创建{targetLevel}级信使。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    )
  }

  const validateForm = (): boolean => {
    const newErrors: Partial<CourierFormData> = {}
    
    if (!formData.username.trim()) {
      newErrors.username = '用户名不能为空'
    } else if (formData.username.length < 3) {
      newErrors.username = '用户名至少3个字符'
    }
    
    if (!formData.email.trim()) {
      newErrors.email = '邮箱不能为空'
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = '邮箱格式不正确'
    }
    
    if (!formData.password.trim()) {
      newErrors.password = '密码不能为空'
    } else if (formData.password.length < 6) {
      newErrors.password = '密码至少6个字符'
    }
    
    if (!formData.realName.trim()) {
      newErrors.realName = '真实姓名不能为空'
    }
    
    if (!formData.phone.trim()) {
      newErrors.phone = '联系电话不能为空'
    } else if (!/^1[3-9]\d{9}$/.test(formData.phone)) {
      newErrors.phone = '请输入正确的手机号码'
    }
    
    if (!formData.zoneCode.trim()) {
      newErrors.zoneCode = '区域代码不能为空'
    }
    
    if (!formData.zoneName.trim()) {
      newErrors.zoneName = '区域名称不能为空'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async () => {
    if (!validateForm()) {
      return
    }

    setIsSubmitting(true)
    
    try {
      // 调用真实API创建信使
      const { courierApi } = await import('@/lib/api/courier')
      
      const result = await courierApi.createCourier({
        username: formData.username,
        email: formData.email,
        nickname: formData.realName,
        level: targetLevel,
        zone: formData.zoneCode,
        description: formData.description
      })

      console.log('创建信使成功:', result)
      
      // 重置表单
      setFormData({
        username: '',
        email: '',
        password: '',
        realName: '',
        phone: '',
        zoneCode: '',
        zoneName: '',
        description: ''
      })
      
      // 关闭对话框并触发成功回调
      onOpenChange(false)
      onSuccess()
      
    } catch (error) {
      console.error('创建信使失败:', error)
      setErrors({
        username: '创建失败，请稍后重试'
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleInputChange = (field: keyof CourierFormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }))
    // 清除相关错误
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }))
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className={`w-12 h-12 ${config.color} text-white rounded-full flex items-center justify-center`}>
              <IconComponent className="w-6 h-6" />
            </div>
            <div>
              <DialogTitle className="text-xl text-amber-900">
                创建{config.name}
              </DialogTitle>
              <DialogDescription className="text-amber-700">
                {config.description}
              </DialogDescription>
            </div>
          </div>
          
          <div className="flex items-center gap-2 mt-4">
            <Badge className={`${config.color} text-white`}>
              {config.name}
            </Badge>
            <Badge variant="outline" className="border-amber-300 text-amber-700">
              由 {courierInfo.level}级信使 创建
            </Badge>
          </div>
        </DialogHeader>

        <div className="grid gap-6 py-4">
          {/* 基础信息 */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-amber-900 border-b border-amber-200 pb-2">
              基础信息
            </h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="username" className="text-amber-800">
                  用户名 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="username"
                  value={formData.username}
                  onChange={(e) => handleInputChange('username', e.target.value)}
                  placeholder="设置登录用户名"
                  className={`border-amber-200 focus:border-amber-400 ${
                    errors.username ? 'border-red-300 focus:border-red-400' : ''
                  }`}
                />
                {errors.username && (
                  <p className="text-sm text-red-600">{errors.username}</p>
                )}
              </div>
              
              <div className="space-y-2">
                <Label htmlFor="realName" className="text-amber-800">
                  真实姓名 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="realName"
                  value={formData.realName}
                  onChange={(e) => handleInputChange('realName', e.target.value)}
                  placeholder="请输入真实姓名"
                  className={`border-amber-200 focus:border-amber-400 ${
                    errors.realName ? 'border-red-300 focus:border-red-400' : ''
                  }`}
                />
                {errors.realName && (
                  <p className="text-sm text-red-600">{errors.realName}</p>
                )}
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="email" className="text-amber-800">
                  邮箱地址 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="email"
                  type="email"
                  value={formData.email}
                  onChange={(e) => handleInputChange('email', e.target.value)}
                  placeholder="请输入邮箱地址"
                  className={`border-amber-200 focus:border-amber-400 ${
                    errors.email ? 'border-red-300 focus:border-red-400' : ''
                  }`}
                />
                {errors.email && (
                  <p className="text-sm text-red-600">{errors.email}</p>
                )}
              </div>
              
              <div className="space-y-2">
                <Label htmlFor="phone" className="text-amber-800">
                  联系电话 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="phone"
                  value={formData.phone}
                  onChange={(e) => handleInputChange('phone', e.target.value)}
                  placeholder="请输入手机号码"
                  className={`border-amber-200 focus:border-amber-400 ${
                    errors.phone ? 'border-red-300 focus:border-red-400' : ''
                  }`}
                />
                {errors.phone && (
                  <p className="text-sm text-red-600">{errors.phone}</p>
                )}
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="password" className="text-amber-800">
                初始密码 <span className="text-red-500">*</span>
              </Label>
              <Input
                id="password"
                type="password"
                value={formData.password}
                onChange={(e) => handleInputChange('password', e.target.value)}
                placeholder="设置初始登录密码"
                className={`border-amber-200 focus:border-amber-400 ${
                  errors.password ? 'border-red-300 focus:border-red-400' : ''
                }`}
              />
              {errors.password && (
                <p className="text-sm text-red-600">{errors.password}</p>
              )}
              <p className="text-xs text-amber-600">
                信使首次登录后可自行修改密码
              </p>
            </div>
          </div>

          {/* 区域信息 */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-amber-900 border-b border-amber-200 pb-2">
              区域分配
            </h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="zoneCode" className="text-amber-800">
                  {config.zoneLabel}代码 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="zoneCode"
                  value={formData.zoneCode}
                  onChange={(e) => handleInputChange('zoneCode', e.target.value.toUpperCase())}
                  placeholder="如：DORM001、CS_ZONE_A"
                  className={`border-amber-200 focus:border-amber-400 ${
                    errors.zoneCode ? 'border-red-300 focus:border-red-400' : ''
                  }`}
                />
                {errors.zoneCode && (
                  <p className="text-sm text-red-600">{errors.zoneCode}</p>
                )}
              </div>
              
              <div className="space-y-2">
                <Label htmlFor="zoneName" className="text-amber-800">
                  {config.zoneLabel}名称 <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="zoneName"
                  value={formData.zoneName}
                  onChange={(e) => handleInputChange('zoneName', e.target.value)}
                  placeholder={config.zonePlaceholder}
                  className={`border-amber-200 focus:border-amber-400 ${
                    errors.zoneName ? 'border-red-300 focus:border-red-400' : ''
                  }`}
                />
                {errors.zoneName && (
                  <p className="text-sm text-red-600">{errors.zoneName}</p>
                )}
              </div>
            </div>
          </div>

          {/* 备注信息 */}
          <div className="space-y-2">
            <Label htmlFor="description" className="text-amber-800">
              备注说明
            </Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => handleInputChange('description', e.target.value)}
              placeholder="可填写特殊说明、工作要求等信息"
              className="border-amber-200 focus:border-amber-400 min-h-[80px]"
              rows={3}
            />
          </div>
        </div>

        <DialogFooter className="flex gap-3">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isSubmitting}
            className="border-amber-300 text-amber-700 hover:bg-amber-50"
          >
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isSubmitting}
            className="bg-amber-600 hover:bg-amber-700 text-white"
          >
            {isSubmitting ? (
              <>
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
                创建中...
              </>
            ) : (
              <>
                <UserPlus className="w-4 h-4 mr-2" />
                创建{config.name}
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}