/**
 * 权限配置管理界面 - SOTA权限系统管理组件
 * 系统管理员可以动态配置各级信使和角色的权限
 */

'use client'

import React, { useState, useEffect, useMemo } from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Separator } from '@/components/ui/separator'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { 
  Settings, 
  Shield, 
  Users, 
  Download, 
  Upload, 
  RotateCcw, 
  Save,
  AlertTriangle,
  Info,
  CheckCircle2,
  XCircle,
  Crown,
  UserCog
} from 'lucide-react'
import { usePermissions } from '@/hooks/use-permissions'
import { permissionService } from '@/lib/permissions/permission-service'
import { 
  PERMISSION_MODULES, 
  PERMISSION_GROUPS,
  PermissionModule,
  PermissionCategory,
  RiskLevel
} from '@/lib/permissions/permission-modules'
import { UserRole, CourierLevel } from '@/constants/roles'

interface PermissionConfigManagerProps {
  className?: string
}

export function PermissionConfigManager({ className }: PermissionConfigManagerProps) {
  const { canManagePermissions, refreshPermissions } = usePermissions()
  const [activeTab, setActiveTab] = useState<'roles' | 'courier-levels' | 'overview'>('overview')
  const [selectedRole, setSelectedRole] = useState<UserRole>('user')
  const [selectedCourierLevel, setSelectedCourierLevel] = useState<CourierLevel>(1)
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error' | 'info'; text: string } | null>(null)

  // 权限状态
  const [rolePermissions, setRolePermissions] = useState<Record<UserRole, string[]>>({} as any)
  const [courierLevelPermissions, setCourierLevelPermissions] = useState<Record<CourierLevel, string[]>>({} as any)
  const [hasChanges, setHasChanges] = useState(false)

  // 检查管理权限
  if (!canManagePermissions()) {
    return (
      <Card className={className}>
        <CardContent className="pt-6">
          <Alert>
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              您没有权限访问权限配置管理功能。需要系统管理员或权限管理权限。
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    )
  }

  // 加载当前配置
  useEffect(() => {
    loadCurrentConfigs()
  }, [])

  const loadCurrentConfigs = async () => {
    setLoading(true)
    try {
      // 加载角色权限
      const roles: UserRole[] = ['user', 'courier_level1', 'courier_level2', 'courier_level3', 'courier_level4', 'platform_admin', 'super_admin']
      const rolePerms = {} as Record<UserRole, string[]>
      roles.forEach(role => {
        rolePerms[role] = permissionService.getRolePermissions(role)
      })
      setRolePermissions(rolePerms)

      // 加载信使等级权限
      const courierLevels: CourierLevel[] = [1, 2, 3, 4]
      const courierPerms = {} as Record<CourierLevel, string[]>
      courierLevels.forEach(level => {
        courierPerms[level] = permissionService.getCourierLevelPermissions(level)
      })
      setCourierLevelPermissions(courierPerms)

      setHasChanges(false)
    } catch (error) {
      setMessage({ type: 'error', text: '加载权限配置失败' })
    } finally {
      setLoading(false)
    }
  }

  // 权限模块按分类分组
  const permissionsByCategory = useMemo(() => {
    const categories: Record<PermissionCategory, PermissionModule[]> = {
      basic: [],
      courier: [],
      management: [],
      admin: [],
      system: []
    }

    Object.values(PERMISSION_MODULES).forEach(module => {
      categories[module.category].push(module)
    })

    return categories
  }, [])

  // 风险级别颜色映射
  const getRiskLevelColor = (risk: RiskLevel) => {
    switch (risk) {
      case 'low': return 'bg-green-100 text-green-800'
      case 'medium': return 'bg-yellow-100 text-yellow-800'
      case 'high': return 'bg-orange-100 text-orange-800'
      case 'critical': return 'bg-red-100 text-red-800'
    }
  }

  // 更新角色权限
  const updateRolePermission = (role: UserRole, permissionId: string, granted: boolean) => {
    setRolePermissions(prev => {
      const current = prev[role] || []
      const updated = granted 
        ? [...current, permissionId]
        : current.filter(p => p !== permissionId)
      
      return { ...prev, [role]: updated }
    })
    setHasChanges(true)
  }

  // 更新信使等级权限
  const updateCourierLevelPermission = (level: CourierLevel, permissionId: string, granted: boolean) => {
    setCourierLevelPermissions(prev => {
      const current = prev[level] || []
      const updated = granted 
        ? [...current, permissionId]
        : current.filter(p => p !== permissionId)
      
      return { ...prev, [level]: updated }
    })
    setHasChanges(true)
  }

  // 保存配置
  const saveConfigs = async () => {
    setLoading(true)
    try {
      // 保存角色权限
      for (const [role, permissions] of Object.entries(rolePermissions)) {
        await permissionService.updateRolePermissions(role as UserRole, permissions, 'admin')
      }

      // 保存信使等级权限
      for (const [level, permissions] of Object.entries(courierLevelPermissions)) {
        await permissionService.updateCourierLevelPermissions(Number(level) as CourierLevel, permissions, 'admin')
      }

      await refreshPermissions()
      setHasChanges(false)
      setMessage({ type: 'success', text: '权限配置已保存' })
    } catch (error) {
      setMessage({ type: 'error', text: '保存权限配置失败' })
    } finally {
      setLoading(false)
    }
  }

  // 重置配置
  const resetConfigs = async () => {
    if (!confirm('确定要重置所有权限配置为默认值吗？此操作不可恢复。')) return

    setLoading(true)
    try {
      const roles: UserRole[] = Object.keys(rolePermissions) as UserRole[]
      for (const role of roles) {
        await permissionService.resetRolePermissions(role)
      }

      const levels: CourierLevel[] = Object.keys(courierLevelPermissions).map(Number) as CourierLevel[]
      for (const level of levels) {
        await permissionService.resetCourierLevelPermissions(level)
      }

      await loadCurrentConfigs()
      setMessage({ type: 'success', text: '权限配置已重置为默认值' })
    } catch (error) {
      setMessage({ type: 'error', text: '重置权限配置失败' })
    } finally {
      setLoading(false)
    }
  }

  // 导出配置
  const exportConfig = () => {
    try {
      const config = permissionService.exportConfigs()
      const blob = new Blob([config], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `permission-config-${new Date().toISOString().slice(0, 10)}.json`
      a.click()
      URL.revokeObjectURL(url)
      setMessage({ type: 'success', text: '权限配置已导出' })
    } catch (error) {
      setMessage({ type: 'error', text: '导出权限配置失败' })
    }
  }

  // 导入配置
  const importConfig = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = async (e) => {
      try {
        const config = e.target?.result as string
        await permissionService.importConfigs(config, true)
        await loadCurrentConfigs()
        setMessage({ type: 'success', text: '权限配置已导入' })
      } catch (error) {
        setMessage({ type: 'error', text: '导入权限配置失败' })
      }
    }
    reader.readAsText(file)
  }

  // 渲染权限模块
  const renderPermissionModule = (
    module: PermissionModule, 
    isGranted: boolean, 
    onToggle: (granted: boolean) => void
  ) => (
    <div key={module.id} className="flex items-center justify-between p-3 border rounded-lg hover:bg-gray-50">
      <div className="flex items-center space-x-3">
        <Checkbox 
          checked={isGranted}
          onCheckedChange={onToggle}
          disabled={module.isSystemCore && module.id === 'SYSTEM_ADMIN'}
        />
        <div className="flex items-center space-x-2">
          <span className="text-lg">{module.icon}</span>
          <div>
            <div className="font-medium">{module.name}</div>
            <div className="text-sm text-gray-600">{module.description}</div>
          </div>
        </div>
      </div>
      <div className="flex items-center space-x-2">
        <Badge variant="outline" className={getRiskLevelColor(module.riskLevel)}>
          {module.riskLevel}
        </Badge>
        {module.isSystemCore && <Badge variant="secondary">核心</Badge>}
      </div>
    </div>
  )

  return (
    <div className={className}>
      {/* 消息提示 */}
      {message && (
        <Alert className={`mb-4 ${message.type === 'error' ? 'border-red-200' : message.type === 'success' ? 'border-green-200' : 'border-blue-200'}`}>
          {message.type === 'error' ? <XCircle className="h-4 w-4" /> : 
           message.type === 'success' ? <CheckCircle2 className="h-4 w-4" /> : 
           <Info className="h-4 w-4" />}
          <AlertDescription>{message.text}</AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Settings className="h-5 w-5" />
              <CardTitle>权限配置管理</CardTitle>
            </div>
            <div className="flex items-center space-x-2">
              <Button variant="outline" size="sm" onClick={exportConfig}>
                <Download className="h-4 w-4 mr-1" />
                导出
              </Button>
              <label className="cursor-pointer">
                <Button variant="outline" size="sm" asChild>
                  <span>
                    <Upload className="h-4 w-4 mr-1" />
                    导入
                  </span>
                </Button>
                <input
                  type="file"
                  accept=".json"
                  className="hidden"
                  onChange={importConfig}
                />
              </label>
              <Button variant="outline" size="sm" onClick={resetConfigs} disabled={loading}>
                <RotateCcw className="h-4 w-4 mr-1" />
                重置
              </Button>
              <Button onClick={saveConfigs} disabled={!hasChanges || loading}>
                <Save className="h-4 w-4 mr-1" />
                保存
              </Button>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <Tabs value={activeTab} onValueChange={setActiveTab as any}>
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="overview">概览</TabsTrigger>
              <TabsTrigger value="roles">角色权限</TabsTrigger>
              <TabsTrigger value="courier-levels">信使等级</TabsTrigger>
            </TabsList>

            {/* 概览页面 */}
            <TabsContent value="overview" className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card>
                  <CardContent className="pt-6">
                    <div className="flex items-center space-x-2">
                      <Shield className="h-5 w-5 text-blue-600" />
                      <div>
                        <p className="text-2xl font-bold">{Object.keys(PERMISSION_MODULES).length}</p>
                        <p className="text-sm text-gray-600">权限模块总数</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardContent className="pt-6">
                    <div className="flex items-center space-x-2">
                      <Users className="h-5 w-5 text-green-600" />
                      <div>
                        <p className="text-2xl font-bold">8</p>
                        <p className="text-sm text-gray-600">用户角色</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardContent className="pt-6">
                    <div className="flex items-center space-x-2">
                      <Crown className="h-5 w-5 text-purple-600" />
                      <div>
                        <p className="text-2xl font-bold">4</p>
                        <p className="text-sm text-gray-600">信使等级</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardContent className="pt-6">
                    <div className="flex items-center space-x-2">
                      <UserCog className="h-5 w-5 text-orange-600" />
                      <div>
                        <p className="text-2xl font-bold">{hasChanges ? '是' : '否'}</p>
                        <p className="text-sm text-gray-600">有未保存更改</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>

              {/* 权限分类统计 */}
              <Card>
                <CardHeader>
                  <CardTitle>权限模块分布</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                    {Object.entries(permissionsByCategory).map(([category, modules]) => (
                      <div key={category} className="text-center">
                        <div className="text-2xl font-bold text-blue-600">{modules.length}</div>
                        <div className="text-sm text-gray-600 capitalize">{category}</div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* 角色权限页面 */}
            <TabsContent value="roles" className="space-y-6">
              <div className="flex items-center space-x-4">
                <Label htmlFor="role-select">选择角色:</Label>
                <Select value={selectedRole} onValueChange={setSelectedRole as any}>
                  <SelectTrigger className="w-48">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="user">普通用户</SelectItem>
                    <SelectItem value="courier">信使</SelectItem>
                    <SelectItem value="senior_courier">高级信使</SelectItem>
                    <SelectItem value="courier_coordinator">信使协调员</SelectItem>
                    <SelectItem value="school_admin">学校管理员</SelectItem>
                    <SelectItem value="platform_admin">平台管理员</SelectItem>
                    <SelectItem value="admin">管理员</SelectItem>
                    <SelectItem value="super_admin">超级管理员</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {Object.entries(permissionsByCategory).map(([category, modules]) => (
                <Card key={category}>
                  <CardHeader>
                    <CardTitle className="capitalize">{category} 权限</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      {modules.map(module => 
                        renderPermissionModule(
                          module,
                          rolePermissions[selectedRole]?.includes(module.id) || false,
                          (granted) => updateRolePermission(selectedRole, module.id, granted)
                        )
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </TabsContent>

            {/* 信使等级权限页面 */}
            <TabsContent value="courier-levels" className="space-y-6">
              <div className="flex items-center space-x-4">
                <Label htmlFor="level-select">选择信使等级:</Label>
                <Select value={selectedCourierLevel.toString()} onValueChange={(value) => setSelectedCourierLevel(Number(value) as CourierLevel)}>
                  <SelectTrigger className="w-48">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1">一级信使</SelectItem>
                    <SelectItem value="2">二级信使</SelectItem>
                    <SelectItem value="3">三级信使</SelectItem>
                    <SelectItem value="4">四级信使</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {Object.entries(permissionsByCategory).map(([category, modules]) => (
                <Card key={category}>
                  <CardHeader>
                    <CardTitle className="capitalize">{category} 权限</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      {modules.map(module => 
                        renderPermissionModule(
                          module,
                          courierLevelPermissions[selectedCourierLevel]?.includes(module.id) || false,
                          (granted) => updateCourierLevelPermission(selectedCourierLevel, module.id, granted)
                        )
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}

export default PermissionConfigManager