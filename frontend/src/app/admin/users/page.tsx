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
  AlertTriangle
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
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
import { 
  getRoleDisplayName, 
  getRoleColors, 
  getAllRoleOptions,
  type UserRole 
} from '@/constants/roles'
import { BaseUser } from '@/types'
import { FeatureErrorBoundary, ComponentErrorBoundary } from '@/components/error-boundary'

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
      const mockUsers: AdminUser[] = [
        {
          id: '1',
          username: 'student001',
          email: 'student001@pku.edu.cn',
          nickname: '北大小明',
          role: 'user',
          school_code: 'BJDX01',
          school_name: '北京大学',
          is_active: true,
          is_verified: true,
          last_login_at: '2024-01-20T10:30:00Z',
          created_at: '2024-01-15T08:00:00Z',
          updated_at: '2024-01-20T10:30:00Z',
          status: 'active',
          stats: {
            letters_sent: 25,
            letters_received: 18
          }
        },
        {
          id: '2',
          username: 'courier002',
          email: 'courier002@pku.edu.cn',
          nickname: '快递小王',
          role: 'courier',
          school_code: 'BJDX01',
          school_name: '北京大学',
          is_active: true,
          is_verified: true,
          last_login_at: '2024-01-21T14:20:00Z',
          created_at: '2024-01-10T08:00:00Z',
          updated_at: '2024-01-21T14:20:00Z',
          status: 'active',
          stats: {
            letters_sent: 45,
            letters_received: 32,
            courier_tasks: 156,
            rating: 4.8
          }
        },
        {
          id: '3',
          username: 'admin003',
          email: 'admin003@pku.edu.cn',
          nickname: '管理员李',
          role: 'school_admin',
          school_code: 'BJDX01',
          school_name: '北京大学',
          is_active: true,
          is_verified: true,
          last_login_at: '2024-01-21T16:45:00Z',
          created_at: '2024-01-01T08:00:00Z',
          updated_at: '2024-01-21T16:45:00Z',
          status: 'active',
          stats: {
            letters_sent: 12,
            letters_received: 8
          }
        }
      ]
      setUsers(mockUsers)
    } catch (error) {
      console.error('Failed to load users', error)
    } finally {
      setLoading(false)
    }
  }, [])

  const loadStats = useCallback(async () => {
    try {
      const mockStats: UserStats = {
        total_users: 1234,
        active_users: 1156,
        new_users_this_month: 89,
        by_role: {
          'user': 1000,
          'courier': 180,
          'senior_courier': 35,
          'courier_coordinator': 12,
          'school_admin': 6,
          'platform_admin': 1
        },
        by_school: {
          'BJDX01': 450,
          'QHDX01': 420,
          'FDDX01': 364
        }
      }
      setStats(mockStats)
    } catch (error) {
      console.error('Failed to load stats', error)
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
      setUsers(prev => prev.map(u => 
        u.id === selectedUser.id ? { ...u, is_active: false } : u
      ))
      setShowBanDialog(false)
      setSelectedUser(null)
    } catch (error) {
      console.error('Failed to ban user', error)
    }
  }, [selectedUser])

  const handleUnbanUser = useCallback(async (userId: string) => {
    try {
      setUsers(prev => prev.map(u => 
        u.id === userId ? { ...u, is_active: true } : u
      ))
    } catch (error) {
      console.error('Failed to unban user', error)
    }
  }, [])

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
                {(stats.by_role.courier || 0) + (stats.by_role.senior_courier || 0)}
              </div>
              <p className="text-xs text-muted-foreground">
                含高级信使 {stats.by_role.senior_courier || 0} 人
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
                          <DropdownMenuItem>
                            <Edit className="mr-2 h-4 w-4" />
                            编辑信息
                          </DropdownMenuItem>
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
            <Button>编辑用户</Button>
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
      </div>
    </FeatureErrorBoundary>
  )
}