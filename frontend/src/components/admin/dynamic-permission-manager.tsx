/**
 * 动态权限管理界面 - 响应式、实时交互的权限配置系统
 */

'use client'

import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Progress } from '@/components/ui/progress'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { ScrollArea } from '@/components/ui/scroll-area'
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
  Wifi,
  WifiOff,
  CheckCircle2,
  XCircle,
  Crown,
  UserCog,
  Bell,
  Activity,
  Search,
  Filter,
  MoreHorizontal,
  Eye,
  EyeOff,
  Zap,
  Clock,
  Database
} from 'lucide-react'
import { usePermissions } from '@/hooks/use-permissions'
import { usePermissionNotifications } from '@/hooks/use-permission-notifications'
import { useUser } from '@/stores/user-store'
import { UserRole, CourierLevel } from '@/constants/roles'

interface PermissionData {
  roles: any[]
  courierLevels: any[]
  modules: any
  overview: any
}

interface ChangeLog {
  id: string
  type: string
  target: string
  targetType: 'role' | 'courier-level' | 'system'
  modifiedBy: string
  timestamp: string
  changes?: {
    added: string[]
    removed: string[]
  }
}

export function DynamicPermissionManager() {
  const { user } = useUser()
  const { canManagePermissions } = usePermissions()
  const { connected, lastEvent, connectionError, eventCount, refreshUserPermissions } = usePermissionNotifications()
  
  // 数据状态
  const [data, setData] = useState<PermissionData | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  
  // UI状态
  const [activeTab, setActiveTab] = useState<'overview' | 'roles' | 'courier-levels' | 'audit'>('overview')
  const [selectedRole, setSelectedRole] = useState<UserRole>('user')
  const [selectedCourierLevel, setSelectedCourierLevel] = useState<CourierLevel>(1)
  const [searchTerm, setSearchTerm] = useState('')
  const [filterCategory, setFilterCategory] = useState<string>('all')
  const [showAdvanced, setShowAdvanced] = useState(false)
  
  // 实时通知状态
  const [notifications, setNotifications] = useState<any[]>([])
  const [changeLogs, setChangeLogs] = useState<ChangeLog[]>([])
  
  // 权限变更状态
  const [pendingChanges, setPendingChanges] = useState<Map<string, any>>(new Map())
  const [lastSync, setLastSync] = useState<Date | null>(null)

  // 检查管理权限
  if (!canManagePermissions()) {
    return (
      <Card>
        <CardContent className="pt-6">
          <Alert>
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              您没有权限访问动态权限管理功能。
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    )
  }

  // 初始化数据加载
  useEffect(() => {
    loadInitialData()
  }, [])

  // 监听权限变更事件
  useEffect(() => {
    if (lastEvent && lastEvent.type !== 'connected' && lastEvent.type !== 'heartbeat') {
      console.log('收到权限变更事件:', lastEvent)
      
      // 添加到变更日志
      if (lastEvent.data) {
        const changeLog: ChangeLog = {
          id: `${Date.now()}-${Math.random()}`,
          type: lastEvent.type,
          target: lastEvent.data.target,
          targetType: lastEvent.data.targetType,
          modifiedBy: lastEvent.data.modifiedBy,
          timestamp: lastEvent.data.timestamp,
          changes: lastEvent.data.changes
        }
        setChangeLogs(prev => [changeLog, ...prev.slice(0, 49)]) // 保留最近50条
      }

      // 刷新相关数据
      if (lastEvent.type === 'permission_updated' || lastEvent.type === 'permission_reset' || lastEvent.type === 'config_imported') {
        loadInitialData() // 重新加载数据以反映变更
      }

      // 显示通知
      setSuccess(`权限配置已更新: ${lastEvent.data?.target || '未知目标'}`)
      setTimeout(() => setSuccess(null), 5000)
    }
  }, [lastEvent])

  // 加载初始数据
  const loadInitialData = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const [overviewRes, rolesRes, courierLevelsRes, modulesRes] = await Promise.all([
        fetch('/api/admin/permissions?type=overview'),
        fetch('/api/admin/permissions?type=roles'),
        fetch('/api/admin/permissions?type=courier-levels'),
        fetch('/api/admin/permissions?type=modules')
      ])

      const [overview, roles, courierLevels, modules] = await Promise.all([
        overviewRes.json(),
        rolesRes.json(),
        courierLevelsRes.json(),
        modulesRes.json()
      ])

      if (overview.success && roles.success && courierLevels.success && modules.success) {
        setData({
          overview: overview.data,
          roles: roles.data,
          courierLevels: courierLevels.data,
          modules: modules.data
        })
        setLastSync(new Date())
      } else {
        throw new Error('加载数据失败')
      }
    } catch (error) {
      setError('加载权限数据失败，请刷新重试')
      console.error('加载数据失败:', error)
    } finally {
      setLoading(false)
    }
  }

  // 显示通知消息
  const showNotification = (notification: any) => {
    const { type, data } = notification
    let message = ''
    
    switch (type) {
      case 'permission_updated':
        message = `${data.target} 的权限已被 ${data.modifiedBy} 更新`
        break
      case 'permission_reset':
        message = `${data.target} 的权限已被 ${data.modifiedBy} 重置`
        break
      case 'config_imported':
        message = `权限配置已被 ${data.modifiedBy} 导入`
        break
    }
    
    if (message) {
      setSuccess(message)
      setTimeout(() => setSuccess(null), 5000)
    }
  }

  // 更新角色权限
  const updateRolePermissions = async (role: UserRole, permissions: string[]) => {
    setSaving(true)
    setError(null)
    
    try {
      const response = await fetch('/api/admin/permissions', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: 'role',
          target: role,
          permissions,
          modifiedBy: user?.username || 'unknown'
        })
      })

      const result = await response.json()
      
      if (result.success) {
        setSuccess(result.message)
        
        // 更新本地数据
        setData(prev => prev ? {
          ...prev,
          roles: prev.roles.map(r => 
            r.role === role ? { ...r, ...result.data } : r
          )
        } : null)

        // 清除待保存的更改
        setPendingChanges(prev => {
          const newChanges = new Map(prev)
          newChanges.delete(`role-${role}`)
          return newChanges
        })
      } else {
        setError(result.error || '更新失败')
      }
    } catch (error) {
      setError('网络错误，请重试')
      console.error('更新角色权限失败:', error)
    } finally {
      setSaving(false)
    }
  }

  // 更新信使等级权限
  const updateCourierLevelPermissions = async (level: CourierLevel, permissions: string[]) => {
    setSaving(true)
    setError(null)
    
    try {
      const response = await fetch('/api/admin/permissions', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: 'courier-level',
          target: level,
          permissions,
          modifiedBy: user?.username || 'unknown'
        })
      })

      const result = await response.json()
      
      if (result.success) {
        setSuccess(result.message)
        
        // 更新本地数据
        setData(prev => prev ? {
          ...prev,
          courierLevels: prev.courierLevels.map(l => 
            l.level === level ? { ...l, ...result.data } : l
          )
        } : null)

        // 清除待保存的更改
        setPendingChanges(prev => {
          const newChanges = new Map(prev)
          newChanges.delete(`courier-${level}`)
          return newChanges
        })
      } else {
        setError(result.error || '更新失败')
      }
    } catch (error) {
      setError('网络错误，请重试')
      console.error('更新信使等级权限失败:', error)
    } finally {
      setSaving(false)
    }
  }

  // 切换权限状态
  const togglePermission = (type: 'role' | 'courier-level', target: string | number, permissionId: string, granted: boolean) => {
    const key = `${type}-${target}`
    
    setPendingChanges(prev => {
      const newChanges = new Map(prev)
      const currentChanges = newChanges.get(key) || { permissions: [] }
      
      if (granted) {
        currentChanges.permissions = [...currentChanges.permissions.filter((p: string) => p !== permissionId), permissionId]
      } else {
        currentChanges.permissions = currentChanges.permissions.filter((p: string) => p !== permissionId)
      }
      
      newChanges.set(key, currentChanges)
      return newChanges
    })
  }

  // 批量保存更改
  const saveAllPendingChanges = async () => {
    const changes = Array.from(pendingChanges.entries())
    
    for (const [key, change] of changes) {
      const [type, target] = key.split('-')
      
      if (type === 'role') {
        await updateRolePermissions(target as UserRole, change.permissions)
      } else if (type === 'courier') {
        await updateCourierLevelPermissions(parseInt(target) as CourierLevel, change.permissions)
      }
    }
  }

  // 过滤权限模块
  const filteredModules = useMemo(() => {
    if (!data?.modules?.modules) return []
    
    let modules = Object.values(data.modules.modules)
    
    // 搜索过滤
    if (searchTerm) {
      modules = modules.filter((module: any) => 
        module.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        module.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        module.id.toLowerCase().includes(searchTerm.toLowerCase())
      )
    }
    
    // 分类过滤
    if (filterCategory !== 'all') {
      modules = modules.filter((module: any) => module.category === filterCategory)
    }
    
    return modules
  }, [data?.modules?.modules, searchTerm, filterCategory])

  // 获取当前选中对象的权限
  const getCurrentPermissions = () => {
    if (!data) return []
    
    if (activeTab === 'roles') {
      const roleData = data.roles.find(r => r.role === selectedRole)
      const pendingKey = `role-${selectedRole}`
      return pendingChanges.get(pendingKey)?.permissions || roleData?.permissions || []
    } else if (activeTab === 'courier-levels') {
      const levelData = data.courierLevels.find(l => l.level === selectedCourierLevel)
      const pendingKey = `courier-level-${selectedCourierLevel}`
      return pendingChanges.get(pendingKey)?.permissions || levelData?.permissions || []
    }
    
    return []
  }

  // 检查权限是否被授予
  const isPermissionGranted = (permissionId: string) => {
    const currentPermissions = getCurrentPermissions()
    return currentPermissions.includes(permissionId)
  }

  // 重置权限配置
  const resetPermissions = async (type: 'role' | 'courier-level', target: string | number) => {
    if (!confirm(`确定要重置 ${target} 的权限配置吗？此操作不可恢复。`)) return
    
    setSaving(true)
    setError(null)
    
    try {
      const response = await fetch(`/api/admin/permissions?type=${type}&target=${target}&modifiedBy=${user?.username || 'unknown'}`, {
        method: 'DELETE'
      })

      const result = await response.json()
      
      if (result.success) {
        setSuccess(result.message)
        await loadInitialData() // 重新加载数据
      } else {
        setError(result.error || '重置失败')
      }
    } catch (error) {
      setError('网络错误，请重试')
      console.error('重置权限失败:', error)
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Settings className="h-5 w-5 animate-spin" />
            <span>动态权限管理</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-32 w-full" />
            <Skeleton className="h-64 w-full" />
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error && !data) {
    return (
      <Card>
        <CardContent className="pt-6">
          <Alert>
            <XCircle className="h-4 w-4" />
            <AlertDescription>
              {error}
              <Button variant="outline" size="sm" className="ml-2" onClick={loadInitialData}>
                重试
              </Button>
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* 状态栏 */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                {connected ? (
                  <Wifi className="h-4 w-4 text-green-600" />
                ) : (
                  <WifiOff className="h-4 w-4 text-red-600" />
                )}
                <span className="text-sm text-gray-600">
                  {connected ? '实时连接正常' : connectionError || '实时连接中断'}
                </span>
                {eventCount > 0 && (
                  <Badge variant="secondary" className="text-xs">
                    {eventCount} 事件
                  </Badge>
                )}
              </div>
              
              {lastSync && (
                <div className="flex items-center space-x-2">
                  <Clock className="h-4 w-4 text-gray-400" />
                  <span className="text-sm text-gray-600">
                    最后同步: {lastSync.toLocaleTimeString()}
                  </span>
                </div>
              )}
              
              {pendingChanges.size > 0 && (
                <div className="flex items-center space-x-2">
                  <AlertTriangle className="h-4 w-4 text-orange-600" />
                  <span className="text-sm text-orange-600">
                    {pendingChanges.size} 个待保存的更改
                  </span>
                </div>
              )}
            </div>
            
            <div className="flex items-center space-x-2">
              {pendingChanges.size > 0 && (
                <Button 
                  onClick={saveAllPendingChanges} 
                  disabled={saving}
                  size="sm"
                >
                  <Save className="h-4 w-4 mr-1" />
                  保存全部更改
                </Button>
              )}
              
              <Button variant="outline" size="sm" onClick={loadInitialData} disabled={loading}>
                <Database className="h-4 w-4 mr-1" />
                刷新数据
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 消息提示 */}
      {(success || error) && (
        <Alert className={error ? 'border-red-200' : 'border-green-200'}>
          {error ? <XCircle className="h-4 w-4" /> : <CheckCircle2 className="h-4 w-4" />}
          <AlertDescription>
            {error || success}
            {error && (
              <Button variant="outline" size="sm" className="ml-2" onClick={() => setError(null)}>
                关闭
              </Button>
            )}
          </AlertDescription>
        </Alert>
      )}

      {/* 主界面 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center space-x-2">
              <Zap className="h-5 w-5" />
              <span>动态权限管理</span>
            </CardTitle>
            
            <div className="flex items-center space-x-2">
              <Badge variant="outline" className="flex items-center space-x-1">
                <Activity className="h-3 w-3" />
                <span>{notifications.length} 个通知</span>
              </Badge>
              
              <Switch
                checked={showAdvanced}
                onCheckedChange={setShowAdvanced}
                id="advanced-mode"
              />
              <Label htmlFor="advanced-mode" className="text-sm">高级模式</Label>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <Tabs value={activeTab} onValueChange={setActiveTab as any}>
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="overview">概览</TabsTrigger>
              <TabsTrigger value="roles">角色权限</TabsTrigger>
              <TabsTrigger value="courier-levels">信使等级</TabsTrigger>
              <TabsTrigger value="audit">审计日志</TabsTrigger>
            </TabsList>

            {/* 概览标签页 */}
            <TabsContent value="overview" className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card>
                  <CardContent className="pt-6">
                    <div className="flex items-center space-x-2">
                      <Shield className="h-5 w-5 text-blue-600" />
                      <div>
                        <p className="text-2xl font-bold">{data?.overview?.totalModules || 0}</p>
                        <p className="text-sm text-gray-600">权限模块</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardContent className="pt-6">
                    <div className="flex items-center space-x-2">
                      <Users className="h-5 w-5 text-green-600" />
                      <div>
                        <p className="text-2xl font-bold">{data?.overview?.totalRoles || 0}</p>
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
                        <p className="text-2xl font-bold">{data?.overview?.totalCourierLevels || 0}</p>
                        <p className="text-sm text-gray-600">信使等级</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardContent className="pt-6">
                    <div className="flex items-center space-x-2">
                      <Bell className="h-5 w-5 text-orange-600" />
                      <div>
                        <p className="text-2xl font-bold">{notifications.length}</p>
                        <p className="text-sm text-gray-600">实时通知</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>

              {/* 实时通知列表 */}
              {notifications.length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle>最近的权限变更</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <ScrollArea className="h-64">
                      <div className="space-y-2">
                        {notifications.map((notification, index) => (
                          <div key={index} className="flex items-center justify-between p-2 border rounded">
                            <div>
                              <p className="font-medium">{notification.data?.target}</p>
                              <p className="text-sm text-gray-600">
                                被 {notification.data?.modifiedBy} {notification.type === 'permission_updated' ? '更新' : notification.type === 'permission_reset' ? '重置' : '导入'}
                              </p>
                            </div>
                            <div className="text-xs text-gray-500">
                              {new Date(notification.data?.timestamp).toLocaleTimeString()}
                            </div>
                          </div>
                        ))}
                      </div>
                    </ScrollArea>
                  </CardContent>
                </Card>
              )}
            </TabsContent>

            {/* 角色权限标签页 */}
            <TabsContent value="roles" className="space-y-6">
              <div className="flex items-center justify-between">
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
                
                <div className="flex items-center space-x-2">
                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={() => updateRolePermissions(selectedRole, getCurrentPermissions())}
                    disabled={saving || !pendingChanges.has(`role-${selectedRole}`)}
                  >
                    <Save className="h-4 w-4 mr-1" />
                    保存更改
                  </Button>
                  
                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={() => resetPermissions('role', selectedRole)}
                    disabled={saving}
                  >
                    <RotateCcw className="h-4 w-4 mr-1" />
                    重置
                  </Button>
                </div>
              </div>

              {/* 权限模块搜索和筛选 */}
              <div className="flex items-center space-x-4">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                  <Input
                    placeholder="搜索权限模块..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                  />
                </div>
                
                <Select value={filterCategory} onValueChange={setFilterCategory}>
                  <SelectTrigger className="w-32">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部</SelectItem>
                    <SelectItem value="basic">基础</SelectItem>
                    <SelectItem value="courier">信使</SelectItem>
                    <SelectItem value="management">管理</SelectItem>
                    <SelectItem value="admin">管理员</SelectItem>
                    <SelectItem value="system">系统</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* 权限模块列表 */}
              <ScrollArea className="h-96">
                <div className="space-y-2">
                  {filteredModules.map((module: any) => (
                    <div key={module.id} className="flex items-center justify-between p-3 border rounded-lg hover:bg-gray-50">
                      <div className="flex items-center space-x-3">
                        <Checkbox 
                          checked={isPermissionGranted(module.id)}
                          onCheckedChange={(checked) => 
                            togglePermission('role', selectedRole, module.id, !!checked)
                          }
                          disabled={saving}
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
                        <Badge variant="secondary" className="capitalize">
                          {module.category}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </TabsContent>

            {/* 信使等级权限标签页 */}
            <TabsContent value="courier-levels" className="space-y-6">
              <div className="flex items-center justify-between">
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
                
                <div className="flex items-center space-x-2">
                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={() => updateCourierLevelPermissions(selectedCourierLevel, getCurrentPermissions())}
                    disabled={saving || !pendingChanges.has(`courier-level-${selectedCourierLevel}`)}
                  >
                    <Save className="h-4 w-4 mr-1" />
                    保存更改
                  </Button>
                  
                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={() => resetPermissions('courier-level', selectedCourierLevel)}
                    disabled={saving}
                  >
                    <RotateCcw className="h-4 w-4 mr-1" />
                    重置
                  </Button>
                </div>
              </div>

              {/* 权限模块搜索和筛选 */}
              <div className="flex items-center space-x-4">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                  <Input
                    placeholder="搜索权限模块..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                  />
                </div>
                
                <Select value={filterCategory} onValueChange={setFilterCategory}>
                  <SelectTrigger className="w-32">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部</SelectItem>
                    <SelectItem value="basic">基础</SelectItem>
                    <SelectItem value="courier">信使</SelectItem>
                    <SelectItem value="management">管理</SelectItem>
                    <SelectItem value="admin">管理员</SelectItem>
                    <SelectItem value="system">系统</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* 权限模块列表 */}
              <ScrollArea className="h-96">
                <div className="space-y-2">
                  {filteredModules.map((module: any) => (
                    <div key={module.id} className="flex items-center justify-between p-3 border rounded-lg hover:bg-gray-50">
                      <div className="flex items-center space-x-3">
                        <Checkbox 
                          checked={isPermissionGranted(module.id)}
                          onCheckedChange={(checked) => 
                            togglePermission('courier-level', selectedCourierLevel, module.id, !!checked)
                          }
                          disabled={saving}
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
                        <Badge variant="secondary" className="capitalize">
                          {module.category}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </TabsContent>

            {/* 审计日志标签页 */}
            <TabsContent value="audit" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>权限变更历史</CardTitle>
                </CardHeader>
                <CardContent>
                  <ScrollArea className="h-96">
                    <div className="space-y-3">
                      {changeLogs.map((log) => (
                        <div key={log.id} className="flex items-center justify-between p-3 border rounded-lg">
                          <div>
                            <div className="font-medium">
                              {log.targetType === 'role' ? '角色' : log.targetType === 'courier-level' ? '信使等级' : '系统'}: {log.target}
                            </div>
                            <div className="text-sm text-gray-600">
                              {log.type === 'permission_updated' ? '权限更新' : 
                               log.type === 'permission_reset' ? '权限重置' : 
                               log.type === 'config_imported' ? '配置导入' : log.type}
                              {' by '}{log.modifiedBy}
                            </div>
                            {log.changes && (
                              <div className="text-xs text-gray-500 mt-1">
                                {log.changes.added?.length > 0 && (
                                  <span className="text-green-600">+{log.changes.added.length} </span>
                                )}
                                {log.changes.removed?.length > 0 && (
                                  <span className="text-red-600">-{log.changes.removed.length}</span>
                                )}
                              </div>
                            )}
                          </div>
                          <div className="text-xs text-gray-500">
                            {new Date(log.timestamp).toLocaleString()}
                          </div>
                        </div>
                      ))}
                      
                      {changeLogs.length === 0 && (
                        <div className="text-center text-gray-500 py-8">
                          暂无权限变更记录
                        </div>
                      )}
                    </div>
                  </ScrollArea>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )

  // 获取风险级别颜色
  function getRiskLevelColor(riskLevel: string) {
    switch (riskLevel) {
      case 'low': return 'bg-green-100 text-green-800'
      case 'medium': return 'bg-yellow-100 text-yellow-800'
      case 'high': return 'bg-orange-100 text-orange-800'
      case 'critical': return 'bg-red-100 text-red-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }
}

export default DynamicPermissionManager