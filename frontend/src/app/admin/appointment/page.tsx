'use client'

import { useState, useEffect } from 'react'
import { getUsers, appointUser, getAppointmentRecords } from '@/lib/api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { 
  Users, 
  UserPlus, 
  Search, 
  Eye, 
  Crown, 
  Award,
  AlertTriangle,
  CheckCircle,
  Clock,
  History
} from 'lucide-react'
import { usePermission } from '@/hooks/use-permission'

interface User {
  id: string
  username: string
  email: string
  currentRole: string
  schoolCode?: string
  schoolName?: string
  join_date: string
  last_active: string
  letters_sent: number
  letters_received: number
  courier_tasks?: number
  average_rating?: number
}

interface AppointmentRecord {
  id: string
  appointer_id: string
  appointer_name: string
  target_user_id: string
  target_user_name: string
  from_role: string
  to_role: string
  reason: string
  status: 'pending' | 'approved' | 'rejected'
  created_at: string
  approved_at?: string
  approved_by?: string
}

const ROLE_HIERARCHY = {
  'user': { level: 1, name: '普通用户' },
  'courier': { level: 2, name: '信使' },
  'senior_courier': { level: 3, name: '高级信使' },
  'courier_coordinator': { level: 4, name: '信使协调员' },
  'school_admin': { level: 5, name: '学校管理员' },
  'platform_admin': { level: 6, name: '平台管理员' },
  'super_admin': { level: 7, name: '超级管理员' }
}

export default function AppointmentPage() {
  const { user, hasRole, getRoleDisplayName } = usePermission()
  
  const [searchTerm, setSearchTerm] = useState('')
  const [roleFilter, setRoleFilter] = useState<string>('all')
  const [users, setUsers] = useState<User[]>([])
  const [appointmentRecords, setAppointmentRecords] = useState<AppointmentRecord[]>([])
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [appointmentDialog, setAppointmentDialog] = useState(false)
  const [appointmentForm, setAppointmentForm] = useState({
    newRole: '',
    reason: ''
  })

  // 从API获取数据 - must be before conditional return
  useEffect(() => {
    const fetchData = async () => {
      try {
        // 获取用户列表
        const usersResponse = await getUsers()
        const usersData = (usersResponse.data as any)?.users || []
        
        // 转换用户数据格式
        const transformedUsers: User[] = usersData.map((apiUser: any) => ({
          id: apiUser.id,
          username: apiUser.username,
          email: apiUser.email,
          currentRole: apiUser.role,
          schoolCode: apiUser.school_code,
          schoolName: '学校名称', // TODO: 从API获取完整学校信息
          join_date: apiUser.created_at,
          last_active: apiUser.created_at, // TODO: 从API获取最后活跃时间
          letters_sent: 0, // TODO: 从API获取信件统计
          letters_received: 0, // TODO: 从API获取信件统计
          courier_tasks: 0, // TODO: 从API获取信使任务数
          average_rating: 4.8 // TODO: 从API获取评分
        }))
        
        setUsers(transformedUsers)

        // 可任命角色列表在组件中定义

        // 获取任命记录
        const recordsResponse = await getAppointmentRecords()
        const recordsData = (recordsResponse.data as any)?.records || []
        
        // 转换任命记录数据格式
        const transformedRecords: AppointmentRecord[] = recordsData.map((record: any) => ({
          id: record.id,
          appointer_id: record.appointed_by,
          appointer_name: '任命者', // TODO: 从API获取任命者姓名
          target_user_id: record.user_id,
          target_user_name: '目标用户', // TODO: 从API获取目标用户姓名
          from_role: record.old_role,
          to_role: record.new_role,
          reason: record.reason,
          status: record.status,
          created_at: record.appointed_at,
          approved_at: record.appointed_at,
          approved_by: record.appointed_by
        }))
        
        setAppointmentRecords(transformedRecords)
        
      } catch (error) {
        console.error('Failed to load appointment data:', error)
        // 设置空数据而不是mock数据
        setUsers([])
        setAppointmentRecords([])
      }
    }

    fetchData()
  }, [user])

  // 权限检查 - 只有管理员以上才能访问任命功能
  if (!user || !hasRole('school_admin')) {
    return (
      <div className="min-h-screen bg-amber-50 flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Crown className="w-12 h-12 text-amber-600 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-amber-900 mb-2">访问权限不足</h2>
            <p className="text-amber-700 mb-4">
              只有管理员以上角色才能访问任命系统
            </p>
            <Button asChild variant="outline" className="border-amber-300 text-amber-700">
              <a href="/settings">返回设置中心</a>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  const filteredUsers = users.filter(u => {
    const matchesSearch = u.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         u.email.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesRole = roleFilter === 'all' || u.currentRole === roleFilter
    return matchesSearch && matchesRole
  })

  const getAppointableRoles = (currentRole: string) => {
    const currentLevel = ROLE_HIERARCHY[currentRole as keyof typeof ROLE_HIERARCHY]?.level || 1
    const userLevel = ROLE_HIERARCHY[user.role as keyof typeof ROLE_HIERARCHY]?.level || 1
    
    // 只能任命比当前用户级别低的角色，且比目标用户当前级别高的角色
    return Object.entries(ROLE_HIERARCHY).filter(([role, info]) => 
      info.level > currentLevel && info.level < userLevel
    ).map(([role, info]) => ({ value: role, label: info.name }))
  }

  const handleAppointUser = (user: User) => {
    setSelectedUser(user)
    setAppointmentForm({ newRole: '', reason: '' })
    setAppointmentDialog(true)
  }

  const handleSubmitAppointment = async () => {
    if (!selectedUser || !appointmentForm.newRole || !appointmentForm.reason.trim()) {
      return
    }

    try {
      // 调用API提交任命申请
      const result = await appointUser({
        user_id: selectedUser.id,
        new_role: appointmentForm.newRole,
        reason: appointmentForm.reason
      })

      // 创建新的任命记录
      const newRecord: AppointmentRecord = {
        id: `a${Date.now()}`,
        appointer_id: user.id,
        appointer_name: user.username,
        target_user_id: selectedUser.id,
        target_user_name: selectedUser.username,
        from_role: selectedUser.currentRole,
        to_role: appointmentForm.newRole,
        reason: appointmentForm.reason,
        status: 'pending',
        created_at: new Date().toISOString()
      }

      setAppointmentRecords(prev => [newRecord, ...prev])
      setAppointmentDialog(false)
      alert('任命申请已提交成功！')
      setSelectedUser(null)
    } catch (error) {
      console.error('任命申请失败:', error)
    }
  }

  const getRoleColor = (role: string) => {
    const colors: Record<string, string> = {
      'user': 'bg-gray-100 text-gray-800',
      'courier': 'bg-yellow-100 text-yellow-800',
      'senior_courier': 'bg-orange-100 text-orange-800',
      'courier_coordinator': 'bg-amber-100 text-amber-800',
      'school_admin': 'bg-blue-100 text-blue-800',
      'platform_admin': 'bg-purple-100 text-purple-800',
      'super_admin': 'bg-red-100 text-red-800'
    }
    return colors[role] || 'bg-gray-100 text-gray-800'
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'approved': return 'bg-green-100 text-green-800'
      case 'pending': return 'bg-yellow-100 text-yellow-800'
      case 'rejected': return 'bg-red-100 text-red-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'approved': return <CheckCircle className="w-4 h-4" />
      case 'pending': return <Clock className="w-4 h-4" />
      case 'rejected': return <AlertTriangle className="w-4 h-4" />
      default: return null
    }
  }

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-7xl mx-auto px-4 py-8">
        {/* 页面标题 */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <Crown className="w-8 h-8 text-amber-600" />
            <h1 className="text-3xl font-bold text-amber-900">用户任命系统</h1>
          </div>
          <p className="text-amber-700">管理用户角色提升和权限分配</p>
        </div>

        <Tabs defaultValue="users" className="space-y-6">
          <TabsList className="bg-amber-100">
            <TabsTrigger value="users" className="data-[state=active]:bg-amber-200">待任命用户</TabsTrigger>
            <TabsTrigger value="records" className="data-[state=active]:bg-amber-200">任命记录</TabsTrigger>
          </TabsList>

          <TabsContent value="users" className="space-y-6">
            <Card className="border-amber-200">
              <CardHeader>
                <CardTitle className="text-amber-900">用户管理</CardTitle>
                <CardDescription>查看和任命平台用户到更高级别角色</CardDescription>
              </CardHeader>
              <CardContent>
                {/* 搜索和筛选 */}
                <div className="flex flex-col md:flex-row gap-4 mb-6">
                  <div className="flex-1">
                    <div className="relative">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-amber-500 w-4 h-4" />
                      <Input
                        placeholder="搜索用户名或邮箱..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="pl-10 border-amber-200 focus:border-amber-400"
                      />
                    </div>
                  </div>
                  <Select value={roleFilter} onValueChange={setRoleFilter}>
                    <SelectTrigger className="w-full md:w-48 border-amber-200">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">全部角色</SelectItem>
                      <SelectItem value="user">普通用户</SelectItem>
                      <SelectItem value="courier">信使</SelectItem>
                      <SelectItem value="senior_courier">高级信使</SelectItem>
                      <SelectItem value="courier_coordinator">信使协调员</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {/* 用户列表 */}
                <div className="space-y-4">
                  {filteredUsers.map((u) => (
                    <Card key={u.id} className="border-amber-200 hover:border-amber-400 transition-all">
                      <CardContent className="p-6">
                        <div className="flex items-start justify-between">
                          <div className="flex items-start space-x-4">
                            <div className="w-12 h-12 bg-amber-600 text-white rounded-full flex items-center justify-center font-bold">
                              {u.username.charAt(0)}
                            </div>
                            <div className="flex-1">
                              <div className="flex items-center gap-2 mb-2">
                                <h3 className="font-semibold text-amber-900">{u.username}</h3>
                                <Badge className={getRoleColor(u.currentRole)}>
                                  {ROLE_HIERARCHY[u.currentRole as keyof typeof ROLE_HIERARCHY]?.name || u.currentRole}
                                </Badge>
                              </div>
                              <div className="text-sm text-amber-700 space-y-1">
                                <div className="flex items-center gap-4">
                                  <span>{u.email}</span>
                                  {u.schoolName && <span>{u.schoolName}</span>}
                                </div>
                                <div className="flex items-center gap-4">
                                  <span>发信 {u.letters_sent} 封</span>
                                  <span>收信 {u.letters_received} 封</span>
                                  {u.courier_tasks && <span>投递 {u.courier_tasks} 次</span>}
                                  {u.average_rating && <span>评分 {u.average_rating}/5.0</span>}
                                </div>
                                <div className="text-xs text-amber-600">
                                  注册: {new Date(u.join_date).toLocaleDateString()} | 
                                  最后活跃: {new Date(u.last_active).toLocaleString()}
                                </div>
                              </div>
                            </div>
                          </div>
                          <div className="flex gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              className="border-amber-300 text-amber-700 hover:bg-amber-50"
                            >
                              <Eye className="w-4 h-4 mr-1" />
                              详情
                            </Button>
                            {getAppointableRoles(u.currentRole).length > 0 && (
                              <Button
                                onClick={() => handleAppointUser(u)}
                                className="bg-amber-600 hover:bg-amber-700 text-white"
                                size="sm"
                              >
                                <UserPlus className="w-4 h-4 mr-1" />
                                任命
                              </Button>
                            )}
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>

                {filteredUsers.length === 0 && (
                  <div className="text-center py-12">
                    <Users className="w-12 h-12 text-amber-400 mx-auto mb-4" />
                    <h3 className="text-lg font-semibold text-amber-900 mb-2">暂无用户数据</h3>
                    <p className="text-amber-700">请尝试调整筛选条件</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="records" className="space-y-6">
            <Card className="border-amber-200">
              <CardHeader>
                <CardTitle className="text-amber-900">任命记录</CardTitle>
                <CardDescription>查看历史任命申请和审核状态</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {appointmentRecords.map((record) => (
                    <Card key={record.id} className="border-amber-200">
                      <CardContent className="p-6">
                        <div className="flex items-start justify-between mb-4">
                          <div className="flex items-center gap-2">
                            <Badge className={getStatusColor(record.status)}>
                              {getStatusIcon(record.status)}
                              <span className="ml-1">
                                {record.status === 'approved' ? '已通过' :
                                 record.status === 'pending' ? '待审核' : '已拒绝'}
                              </span>
                            </Badge>
                          </div>
                          <div className="text-sm text-amber-600">
                            {new Date(record.created_at).toLocaleString()}
                          </div>
                        </div>

                        <div className="space-y-3">
                          <div className="flex items-center gap-4">
                            <span className="font-semibold text-amber-900">目标用户:</span>
                            <span>{record.target_user_name}</span>
                            <div className="flex items-center gap-2">
                              <Badge className={getRoleColor(record.from_role)}>
                                {ROLE_HIERARCHY[record.from_role as keyof typeof ROLE_HIERARCHY]?.name}
                              </Badge>
                              <span>→</span>
                              <Badge className={getRoleColor(record.to_role)}>
                                {ROLE_HIERARCHY[record.to_role as keyof typeof ROLE_HIERARCHY]?.name}
                              </Badge>
                            </div>
                          </div>

                          <div>
                            <span className="font-semibold text-amber-900">任命理由:</span>
                            <p className="text-amber-700 mt-1">{record.reason}</p>
                          </div>

                          <div className="flex items-center gap-4 text-sm text-amber-600">
                            <span>申请人: {record.appointer_name}</span>
                            {record.approved_at && record.approved_by && (
                              <span>审批人: {record.approved_by} ({new Date(record.approved_at).toLocaleString()})</span>
                            )}
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>

                {appointmentRecords.length === 0 && (
                  <div className="text-center py-12">
                    <History className="w-12 h-12 text-amber-400 mx-auto mb-4" />
                    <h3 className="text-lg font-semibold text-amber-900 mb-2">暂无任命记录</h3>
                    <p className="text-amber-700">还没有进行过任命操作</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        {/* 任命对话框 */}
        <Dialog open={appointmentDialog} onOpenChange={setAppointmentDialog}>
          <DialogContent className="sm:max-w-md">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <UserPlus className="w-5 h-5 text-amber-600" />
                任命用户
              </DialogTitle>
              <DialogDescription>
                为用户 "{selectedUser?.username}" 分配新的角色权限
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              <div>
                <Label htmlFor="current-role">当前角色</Label>
                <div className="mt-1">
                  <Badge className={getRoleColor(selectedUser?.currentRole || '')}>
                    {ROLE_HIERARCHY[selectedUser?.currentRole as keyof typeof ROLE_HIERARCHY]?.name || selectedUser?.currentRole}
                  </Badge>
                </div>
              </div>

              <div>
                <Label htmlFor="new-role">新角色 *</Label>
                <Select value={appointmentForm.newRole} onValueChange={(value) => 
                  setAppointmentForm(prev => ({ ...prev, newRole: value }))
                }>
                  <SelectTrigger className="mt-1 border-amber-200">
                    <SelectValue placeholder="选择新角色" />
                  </SelectTrigger>
                  <SelectContent>
                    {selectedUser && getAppointableRoles(selectedUser.currentRole).map(role => (
                      <SelectItem key={role.value} value={role.value}>
                        {role.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div>
                <Label htmlFor="reason">任命理由 *</Label>
                <Textarea
                  id="reason"
                  placeholder="请详细说明任命理由..."
                  value={appointmentForm.reason}
                  onChange={(e) => setAppointmentForm(prev => ({ ...prev, reason: e.target.value }))}
                  className="mt-1 border-amber-200 focus:border-amber-400"
                  rows={4}
                />
              </div>
            </div>

            <DialogFooter>
              <Button 
                variant="outline" 
                onClick={() => setAppointmentDialog(false)}
                className="border-amber-300 text-amber-700"
              >
                取消
              </Button>
              <Button 
                onClick={handleSubmitAppointment}
                disabled={!appointmentForm.newRole || !appointmentForm.reason.trim()}
                className="bg-amber-600 hover:bg-amber-700 text-white"
              >
                提交任命
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  )
}