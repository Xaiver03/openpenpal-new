'use client'

import React, { useState, useEffect, useMemo, useCallback, memo } from 'react'
import { 
  Users, 
  Search, 
  Filter, 
  UserPlus, 
  UserX, 
  Edit, 
  Shield,
  Mail,
  Phone,
  Calendar,
  MapPin,
  Activity,
  Settings,
  MoreVertical,
  Eye,
  Ban,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Truck
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { usePermission, PERMISSIONS } from '@/hooks/use-permission'
import { BackButton } from '@/components/ui/back-button'
import { Breadcrumb, ADMIN_BREADCRUMBS } from '@/components/ui/breadcrumb'
import { 
  getRoleDisplayName, 
  getRoleColors, 
  getAllRoleOptions,
  type UserRole 
} from '@/constants/roles'
import { BaseUser } from '@/types'
import { FeatureErrorBoundary, ComponentErrorBoundary } from '@/components/error-boundary'
import AdminService from '@/lib/services/admin-service'

interface AdminUser extends BaseUser {
  role: UserRole
  school_code: string
  school_name: string
  is_verified: boolean
  stats: {
    letters_sent: number
    letters_received: number
    courier_tasks?: number
    rating?: number
  }
}

interface UserStats {
  total_users: number
  active_users: number
  new_users_this_month: number
  by_role: Record<string, number>
  by_school: Record<string, number>
}

const roleOptions = getAllRoleOptions()

export default function UsersManagePage() {
  const { user, hasPermission, hasRole } = usePermission()
  const [users, setUsers] = useState<AdminUser[]>([])
  const [stats, setStats] = useState<UserStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [roleFilter, setRoleFilter] = useState<string>('all')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [schoolFilter, setSchoolFilter] = useState<string>('all')
  const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null)
  const [showUserDetail, setShowUserDetail] = useState(false)
  const [showBanDialog, setShowBanDialog] = useState(false)
  const [showEditDialog, setShowEditDialog] = useState(false)
  const [editFormData, setEditFormData] = useState<{
    nickname: string
    email: string
    role: UserRole | ''
    school_code: string
    is_active: boolean
  }>({
    nickname: '',
    email: '',
    role: '',
    school_code: '',
    is_active: true
  })
  
  // 批量选择状态
  const [selectedUsers, setSelectedUsers] = useState<Set<string>>(new Set())
  const [showBulkActions, setShowBulkActions] = useState(false)

  if (!user || !hasPermission(PERMISSIONS.MANAGE_USERS)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Shield className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">访问权限不足</h2>
            <p className="text-gray-600 mb-4">
              您没有访问用户管理功能的权限
            </p>
            <Button asChild variant="outline">
              <a href="/admin">返回管理控制台</a>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  useEffect(() => {
    loadUsers()
    loadStats()
  }, [])

  const loadUsers = useCallback(async () => {
    setLoading(true)
    try {
      const response = await AdminService.getUsers({
        page: 1,
        limit: 50,
        sort_by: 'created_at',
        sort_order: 'desc'
      })
      
      if (response.success && response.data?.users) {
        // Map API response to local AdminUser type
        const mappedUsers: AdminUser[] = response.data.users.map((user: any) => ({
          ...user,
          school_name: user.school_name || '未知学校',
          stats: {
            letters_sent: user.activity_summary?.letters_sent || 0,
            letters_received: user.activity_summary?.letters_received || 0,
            courier_tasks: user.activity_summary?.courier_tasks,
            rating: user.activity_summary?.rating
          }
        }))
        setUsers(mappedUsers)
      } else {
        // Fallback to empty array if API fails
        setUsers([])
        console.error('Failed to load users: Invalid response format')
      }
    } catch (error) {
      console.error('Failed to load users', error)
      setUsers([])
      // TODO: 显示错误提示给用户
    } finally {
      setLoading(false)
    }
  }, [])

  const loadStats = useCallback(async () => {
    try {
      const response = await AdminService.getDashboardStats()
      
      if (response.success && response.data) {
        const systemStats = response.data
        // Map SystemStats to UserStats format
        const userStats: UserStats = {
          total_users: systemStats.users.total,
          active_users: systemStats.users.active,
          new_users_this_month: systemStats.users.new_this_week * 4, // 估算月度新增
          by_role: systemStats.users.by_role || {},
          by_school: systemStats.users.by_school || {}
        }
        setStats(userStats)
      } else {
        console.error('Failed to load stats: Invalid response format')
      }
    } catch (error) {
      console.error('Failed to load stats', error)
      // 设置默认值以避免UI错误
      setStats({
        total_users: 0,
        active_users: 0,
        new_users_this_month: 0,
        by_role: {},
        by_school: {}
      })
    }
  }, [])

  const filteredUsers = useMemo(() => {
    return users.filter(u => {
      const matchesSearch = 
        u.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
        u.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
        u.nickname.toLowerCase().includes(searchTerm.toLowerCase())
      
      const matchesRole = roleFilter === 'all' || u.role === roleFilter
      const matchesStatus = statusFilter === 'all' || 
        (statusFilter === 'active' && u.is_active) ||
        (statusFilter === 'inactive' && !u.is_active) ||
        (statusFilter === 'verified' && u.is_verified) ||
        (statusFilter === 'unverified' && !u.is_verified)
      
      const matchesSchool = schoolFilter === 'all' || u.school_code === schoolFilter

      return matchesSearch && matchesRole && matchesStatus && matchesSchool
    })
  }, [users, searchTerm, roleFilter, statusFilter, schoolFilter])

  const handleViewUser = useCallback((user: AdminUser) => {
    setSelectedUser(user)
    setShowUserDetail(true)
  }, [])

  const handleBanUser = useCallback((user: AdminUser) => {
    setSelectedUser(user)
    setShowBanDialog(true)
  }, [])

  const confirmBanUser = useCallback(async () => {
    if (!selectedUser) return
    
    try {
      const response = await AdminService.updateUserStatus(
        selectedUser.id, 
        'inactive',
        '管理员禁用'
      )
      
      if (response.success) {
        setUsers(prev => prev.map(u => 
          u.id === selectedUser.id ? { ...u, is_active: false, status: 'inactive' } : u
        ))
        setShowBanDialog(false)
        setSelectedUser(null)
      }
    } catch (error) {
      console.error('Failed to ban user', error)
      alert('禁用用户失败')
    }
  }, [selectedUser])

  const handleUnbanUser = useCallback(async (userId: string) => {
    try {
      const response = await AdminService.updateUserStatus(
        userId,
        'active',
        '管理员解除禁用'
      )
      
      if (response.success) {
        setUsers(prev => prev.map(u => 
          u.id === userId ? { ...u, is_active: true, status: 'active' } : u
        ))
      }
    } catch (error) {
      console.error('Failed to unban user', error)
      alert('解除禁用失败')
    }
  }, [])

  const handleEditUser = useCallback((user: AdminUser) => {
    setSelectedUser(user)
    setEditFormData({
      nickname: user.nickname,
      email: user.email,
      role: user.role,
      school_code: user.school_code,
      is_active: user.is_active ?? true
    })
    setShowEditDialog(true)
  }, [])

  const handleSaveEdit = useCallback(async () => {
    if (!selectedUser) return
    
    try {
      const response = await AdminService.updateUser(selectedUser.id, {
        nickname: editFormData.nickname,
        email: editFormData.email,
        role: editFormData.role === '' ? undefined : editFormData.role,
        school_code: editFormData.school_code,
        is_active: editFormData.is_active
      })
      
      if (response.success && response.data) {
        // Update the user in the local state
        setUsers(prev => prev.map(u => 
          u.id === selectedUser.id 
            ? { 
                ...u, 
                ...response.data,
                stats: u.stats // Preserve stats
              } 
            : u
        ))
        
        setShowEditDialog(false)
        setSelectedUser(null)
        
        // TODO: 显示成功提示
        console.log('用户信息更新成功')
      } else {
        throw new Error(response.message || '更新失败')
      }
    } catch (error) {
      console.error('Failed to update user', error)
      // TODO: 显示错误提示
      alert(error instanceof Error ? error.message : '更新用户信息失败')
    }
  }, [selectedUser, editFormData])

  // 批量选择功能
  const toggleUserSelection = (userId: string) => {
    setSelectedUsers(prev => {
      const newSelected = new Set(prev)
      if (newSelected.has(userId)) {
        newSelected.delete(userId)
      } else {
        newSelected.add(userId)
      }
      setShowBulkActions(newSelected.size > 0)
      return newSelected
    })
  }

  const selectAllUsers = () => {
    const allUserIds = new Set(filteredUsers.map(u => u.id))
    setSelectedUsers(allUserIds)
    setShowBulkActions(allUserIds.size > 0)
  }

  const clearSelection = () => {
    setSelectedUsers(new Set())
    setShowBulkActions(false)
  }

  const handleBulkActivate = async () => {
    try {
      const response = await AdminService.batchOperateUsers({
        user_ids: Array.from(selectedUsers),
        action: 'activate'
      })
      
      if (response.success) {
        // Update local state for successful operations
        setUsers(prev => prev.map(u => 
          selectedUsers.has(u.id) ? { ...u, is_active: true } : u
        ))
        clearSelection()
        console.log(`成功激活 ${response.data?.success_count} 个用户`)
      }
    } catch (error) {
      console.error('Failed to bulk activate users:', error)
      alert('批量激活用户失败')
    }
  }

  const handleBulkDeactivate = async () => {
    try {
      const response = await AdminService.batchOperateUsers({
        user_ids: Array.from(selectedUsers),
        action: 'deactivate'
      })
      
      if (response.success) {
        // Update local state for successful operations
        setUsers(prev => prev.map(u => 
          selectedUsers.has(u.id) ? { ...u, is_active: false } : u
        ))
        clearSelection()
        console.log(`成功禁用 ${response.data?.success_count} 个用户`)
      }
    } catch (error) {
      console.error('Failed to bulk deactivate users:', error)
      alert('批量禁用用户失败')
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  return (
    <FeatureErrorBoundary>
      <div className="container mx-auto p-6 space-y-6">
      
      <Breadcrumb items={ADMIN_BREADCRUMBS.users} />
      
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <BackButton href="/admin" />
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Users className="w-8 h-8" />
            用户管理
          </h1>
          <p className="text-muted-foreground mt-1">
            管理平台用户信息、权限和状态
          </p>
        </div>
        <Button>
          <UserPlus className="w-4 h-4 mr-2" />
          添加用户
        </Button>
      </div>

      {/* 统计卡片 */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">总用户数</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total_users}</div>
              <p className="text-xs text-muted-foreground">
                较上月 +{stats.new_users_this_month} 人
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">活跃用户</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.active_users}</div>
              <p className="text-xs text-muted-foreground">
                活跃率 {Math.round((stats.active_users / stats.total_users) * 100)}%
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">信使用户</CardTitle>
              <Shield className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {(stats.by_role.courier_level1 || 0) + 
                 (stats.by_role.courier_level2 || 0) + 
                 (stats.by_role.courier_level3 || 0) + 
                 (stats.by_role.courier_level4 || 0)}
              </div>
              <p className="text-xs text-muted-foreground">
                含各级信使
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">本月新增</CardTitle>
              <UserPlus className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.new_users_this_month}</div>
              <p className="text-xs text-muted-foreground">
                日均 {Math.round(stats.new_users_this_month / 30)} 人
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 搜索和筛选 */}
      <Card>
        <CardHeader>
          <CardTitle>用户列表</CardTitle>
          <CardDescription>
            查看和管理所有平台用户
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* 快速筛选按钮 */}
          <div className="flex flex-wrap gap-2 mb-4">
            <Button 
              variant={statusFilter === 'all' && roleFilter === 'all' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => {
                setStatusFilter('all')
                setRoleFilter('all')
                setSchoolFilter('all')
              }}
            >
              全部用户
            </Button>
            <Button 
              variant={statusFilter === 'active' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setStatusFilter('active')}
            >
              活跃用户
            </Button>
            <Button 
              variant={roleFilter.includes('courier') ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setRoleFilter(roleFilter.includes('courier') ? 'all' : 'courier_level1')}
            >
              信使用户
            </Button>
            <Button 
              variant={roleFilter === 'platform_admin' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setRoleFilter('platform_admin')}
            >
              管理员
            </Button>
            <Button 
              variant={statusFilter === 'unverified' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setStatusFilter('unverified')}
            >
              未验证
            </Button>
          </div>
          
          {/* 批量操作栏 */}
          {showBulkActions && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">
                    已选择 {selectedUsers.size} 个用户
                  </span>
                  <Button variant="ghost" size="sm" onClick={clearSelection}>
                    取消选择
                  </Button>
                </div>
                <div className="flex gap-2">
                  <Button size="sm" onClick={handleBulkActivate}>
                    <CheckCircle className="w-4 h-4 mr-1" />
                    批量激活
                  </Button>
                  <Button size="sm" variant="destructive" onClick={handleBulkDeactivate}>
                    <Ban className="w-4 h-4 mr-1" />
                    批量禁用
                  </Button>
                </div>
              </div>
            </div>
          )}
          
          <div className="flex flex-col sm:flex-row gap-4 mb-6">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="搜索用户名、邮箱或昵称..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>

            <Select value={roleFilter} onValueChange={setRoleFilter}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue placeholder="角色筛选" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部角色</SelectItem>
                {roleOptions.map(({ value, label }) => (
                  <SelectItem key={value} value={value}>{label}</SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue placeholder="状态筛选" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="active">活跃</SelectItem>
                <SelectItem value="inactive">禁用</SelectItem>
                <SelectItem value="verified">已验证</SelectItem>
                <SelectItem value="unverified">未验证</SelectItem>
              </SelectContent>
            </Select>

            <Select value={schoolFilter} onValueChange={setSchoolFilter}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue placeholder="学校筛选" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部学校</SelectItem>
                <SelectItem value="BJDX01">北京大学</SelectItem>
                <SelectItem value="QHDX01">清华大学</SelectItem>
                <SelectItem value="FDDX01">复旦大学</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* 用户表格 */}
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">
                    <Checkbox
                      checked={selectedUsers.size === filteredUsers.length && filteredUsers.length > 0}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          selectAllUsers()
                        } else {
                          clearSelection()
                        }
                      }}
                    />
                  </TableHead>
                  <TableHead>用户</TableHead>
                  <TableHead>角色</TableHead>
                  <TableHead>学校</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>统计</TableHead>
                  <TableHead>最后登录</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredUsers.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>
                      <Checkbox
                        checked={selectedUsers.has(user.id)}
                        onCheckedChange={() => toggleUserSelection(user.id)}
                      />
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <Avatar className="h-8 w-8">
                          <AvatarImage src={user.avatar} />
                          <AvatarFallback>{user.nickname.charAt(0)}</AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="font-medium">{user.nickname}</div>
                          <div className="text-sm text-muted-foreground">@{user.username}</div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge className={getRoleColors(user.role as UserRole).badge}>
                        {getRoleDisplayName(user.role as UserRole)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <MapPin className="w-3 h-3" />
                        <span className="text-sm">{user.school_name}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {user.is_active ? (
                          <CheckCircle className="w-4 h-4 text-green-500" />
                        ) : (
                          <XCircle className="w-4 h-4 text-red-500" />
                        )}
                        <span className="text-sm">
                          {user.is_active ? '活跃' : '禁用'}
                        </span>
                        {user.is_verified && (
                          <Badge variant="outline" className="text-xs">已验证</Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-xs space-y-1">
                        <div>发信 {user.stats.letters_sent}</div>
                        <div>收信 {user.stats.letters_received}</div>
                        {user.stats.courier_tasks && (
                          <div>投递 {user.stats.courier_tasks}</div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">
                        {user.last_login_at ? 
                          new Date(user.last_login_at).toLocaleString() : 
                          '从未登录'
                        }
                      </div>
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" className="h-8 w-8 p-0">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuLabel>操作</DropdownMenuLabel>
                          <DropdownMenuItem onClick={() => handleViewUser(user)}>
                            <Eye className="mr-2 h-4 w-4" />
                            查看详情
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleEditUser(user)}>
                            <Edit className="mr-2 h-4 w-4" />
                            编辑信息
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem asChild>
                            <a href={`/admin/letters?user=${user.id}`} className="flex items-center">
                              <Mail className="mr-2 h-4 w-4" />
                              查看用户信件
                            </a>
                          </DropdownMenuItem>
                          {user.role.includes('courier') && (
                            <DropdownMenuItem asChild>
                              <a href={`/admin/couriers?user=${user.id}`} className="flex items-center">
                                <Truck className="mr-2 h-4 w-4" />
                                查看信使任务
                              </a>
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuSeparator />
                          {user.is_active ? (
                            <DropdownMenuItem 
                              onClick={() => handleBanUser(user)}
                              className="text-red-600"
                            >
                              <Ban className="mr-2 h-4 w-4" />
                              禁用用户
                            </DropdownMenuItem>
                          ) : (
                            <DropdownMenuItem 
                              onClick={() => handleUnbanUser(user.id)}
                              className="text-green-600"
                            >
                              <CheckCircle className="mr-2 h-4 w-4" />
                              解除禁用
                            </DropdownMenuItem>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {filteredUsers.length === 0 && (
            <div className="text-center py-12">
              <Users className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-semibold mb-2">没有找到用户</h3>
              <p className="text-muted-foreground">请尝试调整筛选条件</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 用户详情对话框 */}
      <ComponentErrorBoundary>
        <Dialog open={showUserDetail} onOpenChange={setShowUserDetail}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>用户详情</DialogTitle>
            <DialogDescription>
              查看用户 {selectedUser?.nickname} 的详细信息
            </DialogDescription>
          </DialogHeader>
          
          {selectedUser && (
            <div className="space-y-4">
              <div className="flex items-center gap-4">
                <Avatar className="h-16 w-16">
                  <AvatarImage src={selectedUser.avatar} />
                  <AvatarFallback>{selectedUser.nickname.charAt(0)}</AvatarFallback>
                </Avatar>
                <div>
                  <h3 className="text-lg font-semibold">{selectedUser.nickname}</h3>
                  <p className="text-sm text-muted-foreground">@{selectedUser.username}</p>
                  <Badge className={getRoleColors(selectedUser.role as UserRole).badge}>
                    {getRoleDisplayName(selectedUser.role as UserRole)}
                  </Badge>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <div className="flex items-center gap-1 mb-1">
                    <Mail className="w-3 h-3" />
                    <span className="font-medium">邮箱</span>
                  </div>
                  <p>{selectedUser.email}</p>
                </div>
                <div>
                  <div className="flex items-center gap-1 mb-1">
                    <MapPin className="w-3 h-3" />
                    <span className="font-medium">学校</span>
                  </div>
                  <p>{selectedUser.school_name}</p>
                </div>
                <div>
                  <div className="flex items-center gap-1 mb-1">
                    <Calendar className="w-3 h-3" />
                    <span className="font-medium">注册时间</span>
                  </div>
                  <p>{new Date(selectedUser.created_at).toLocaleDateString()}</p>
                </div>
                <div>
                  <div className="flex items-center gap-1 mb-1">
                    <Activity className="w-3 h-3" />
                    <span className="font-medium">最后登录</span>
                  </div>
                  <p>
                    {selectedUser.last_login_at ? 
                      new Date(selectedUser.last_login_at).toLocaleDateString() : 
                      '从未登录'
                    }
                  </p>
                </div>
              </div>

              <div>
                <h4 className="font-medium mb-2">活动统计</h4>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>发送信件: {selectedUser.stats.letters_sent}</div>
                  <div>接收信件: {selectedUser.stats.letters_received}</div>
                  {selectedUser.stats.courier_tasks && (
                    <>
                      <div>投递任务: {selectedUser.stats.courier_tasks}</div>
                      <div>评分: {selectedUser.stats.rating}/5.0</div>
                    </>
                  )}
                </div>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowUserDetail(false)}>
              关闭
            </Button>
            <Button onClick={() => {
              if (selectedUser) {
                handleEditUser(selectedUser)
                setShowUserDetail(false)
              }
            }}>编辑用户</Button>
          </DialogFooter>
        </DialogContent>
        </Dialog>
      </ComponentErrorBoundary>

      {/* 禁用用户确认对话框 */}
      <ComponentErrorBoundary>
        <Dialog open={showBanDialog} onOpenChange={setShowBanDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-500" />
              确认禁用用户
            </DialogTitle>
            <DialogDescription>
              您确定要禁用用户 "{selectedUser?.nickname}" 吗？此操作会阻止用户登录和使用平台功能。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBanDialog(false)}>
              取消
            </Button>
            <Button variant="destructive" onClick={confirmBanUser}>
              确认禁用
            </Button>
          </DialogFooter>
        </DialogContent>
        </Dialog>
      </ComponentErrorBoundary>

      {/* 编辑用户对话框 */}
      <ComponentErrorBoundary>
        <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
          <DialogContent className="sm:max-w-md">
            <DialogHeader>
              <DialogTitle>编辑用户信息</DialogTitle>
              <DialogDescription>
                修改用户 {selectedUser?.nickname} 的信息
              </DialogDescription>
            </DialogHeader>
            
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="edit-nickname">昵称</Label>
                <Input
                  id="edit-nickname"
                  value={editFormData.nickname}
                  onChange={(e) => setEditFormData({ ...editFormData, nickname: e.target.value })}
                  placeholder="请输入昵称"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-email">邮箱</Label>
                <Input
                  id="edit-email"
                  type="email"
                  value={editFormData.email}
                  onChange={(e) => setEditFormData({ ...editFormData, email: e.target.value })}
                  placeholder="请输入邮箱"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-role">角色</Label>
                <Select 
                  value={editFormData.role} 
                  onValueChange={(value) => setEditFormData({ ...editFormData, role: value as UserRole | '' })}
                >
                  <SelectTrigger id="edit-role">
                    <SelectValue placeholder="选择角色" />
                  </SelectTrigger>
                  <SelectContent>
                    {roleOptions.map(({ value, label }) => (
                      <SelectItem key={value} value={value}>{label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-school">学校代码</Label>
                <Select 
                  value={editFormData.school_code} 
                  onValueChange={(value) => setEditFormData({ ...editFormData, school_code: value })}
                >
                  <SelectTrigger id="edit-school">
                    <SelectValue placeholder="选择学校" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="BJDX01">北京大学</SelectItem>
                    <SelectItem value="QHDX01">清华大学</SelectItem>
                    <SelectItem value="FDDX01">复旦大学</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-center justify-between">
                <Label htmlFor="edit-active">账号状态</Label>
                <div className="flex items-center space-x-2">
                  <Switch
                    id="edit-active"
                    checked={editFormData.is_active}
                    onCheckedChange={(checked) => setEditFormData({ ...editFormData, is_active: checked })}
                  />
                  <Label htmlFor="edit-active" className="text-sm">
                    {editFormData.is_active ? '活跃' : '禁用'}
                  </Label>
                </div>
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setShowEditDialog(false)}>
                取消
              </Button>
              <Button onClick={handleSaveEdit}>
                保存修改
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </ComponentErrorBoundary>
      </div>
    </FeatureErrorBoundary>
  )
}