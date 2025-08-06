'use client'

import React, { useState, useEffect } from 'react'
import { 
  Users, 
  Search, 
  Filter, 
  Plus,
  Eye,
  Edit,
  Trash2,
  Crown,
  Building,
  MapPin,
  Award,
  TrendingUp,
  Activity,
  UserCheck,
  UserX,
  Truck,
  Package
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { usePermission, PERMISSIONS } from '@/hooks/use-permission'
import { BackButton } from '@/components/ui/back-button'
import { courierApi, type Courier, type CourierTask, type CourierStats } from '@/lib/api/courier'

// 常量定义
const COURIER_LEVELS = {
  1: { name: '楼栋信使', color: 'bg-green-100 text-green-800', icon: Building },
  2: { name: '片区信使', color: 'bg-blue-100 text-blue-800', icon: MapPin },
  3: { name: '校级信使', color: 'bg-purple-100 text-purple-800', icon: Award },
  4: { name: '城市总代', color: 'bg-orange-100 text-orange-800', icon: Crown }
}

const STATUS_COLORS = {
  pending: 'bg-yellow-100 text-yellow-800',
  active: 'bg-green-100 text-green-800',
  inactive: 'bg-gray-100 text-gray-800',
  suspended: 'bg-red-100 text-red-800'
}

const STATUS_NAMES = {
  pending: '待审核',
  active: '活跃',
  inactive: '非活跃',
  suspended: '暂停'
}

const TASK_STATUS_COLORS = {
  pending: 'bg-yellow-100 text-yellow-800',
  accepted: 'bg-blue-100 text-blue-800',
  collected: 'bg-purple-100 text-purple-800',
  in_transit: 'bg-indigo-100 text-indigo-800',
  delivered: 'bg-green-100 text-green-800',
  failed: 'bg-red-100 text-red-800'
}

const TASK_STATUS_NAMES = {
  pending: '待接受',
  accepted: '已接受',
  collected: '已收件',
  in_transit: '运输中',
  delivered: '已投递',
  failed: '失败'
}

export default function CouriersPage() {
  const { user, hasPermission } = usePermission()
  const [couriers, setCouriers] = useState<Courier[]>([])
  const [tasks, setTasks] = useState<CourierTask[]>([])
  const [stats, setStats] = useState<CourierStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [levelFilter, setLevelFilter] = useState<string>('all')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  
  // 对话框状态
  const [selectedCourier, setSelectedCourier] = useState<Courier | null>(null)
  const [showDetailDialog, setShowDetailDialog] = useState(false)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showTaskDialog, setShowTaskDialog] = useState(false)
  
  // 创建信使表单
  const [createForm, setCreateForm] = useState({
    username: '',
    email: '',
    nickname: '',
    level: 1,
    zone: '',
    description: ''
  })

  // 权限检查
  if (!user || !hasPermission(PERMISSIONS.SYSTEM_CONFIG)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Users className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">访问权限不足</h2>
            <p className="text-gray-600 mb-4">
              您没有访问信使管理的权限
            </p>
            <Button asChild variant="outline">
              <a href="/admin">返回管理控制台</a>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  // 加载数据
  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      // 并行调用所有API
      const [couriersRes, tasksRes, statsRes] = await Promise.all([
        courierApi.getAllCouriers(),
        courierApi.getAllTasks(),
        courierApi.getStats()
      ])
      
      // 处理API响应
      if (couriersRes.data && typeof couriersRes.data === 'object' && 'couriers' in couriersRes.data) {
        setCouriers((couriersRes.data as any).couriers)
      }
      
      if (tasksRes.data && typeof tasksRes.data === 'object' && 'tasks' in tasksRes.data) {
        setTasks((tasksRes.data as any).tasks)
      }
      
      if (statsRes.data) {
        setStats(statsRes.data as any)
      }
    } catch (error) {
      console.error('Failed to load courier data:', error)
      
      // 如果API调用失败，使用模拟数据作为后备
      const mockCouriers: Courier[] = [
        {
          id: '1',
          user_id: 'user_001',
          username: 'courier_level4_city',
          nickname: '城市总代-张三',
          email: 'zhang.san@example.com',
          level: 4,
          status: 'active',
          zone: 'BEIJING',
          zone_name: '北京市',
          points: 2850,
          task_count: 125,
          success_rate: 98.5,
          created_at: '2024-01-15T10:30:00Z',
          last_active_at: '2024-01-21T09:45:00Z'
        },
        {
          id: '2',
          user_id: 'user_002',
          username: 'courier_level3_school',
          nickname: '校级信使-李四',
          email: 'li.si@bjdx.edu.cn',
          level: 3,
          status: 'active',
          zone: 'BJDX',
          zone_name: '北京大学',
          points: 1820,
          task_count: 89,
          success_rate: 96.2,
          created_at: '2024-01-16T14:20:00Z',
          last_active_at: '2024-01-21T08:30:00Z'
        },
        {
          id: '3',
          user_id: 'user_003',
          username: 'courier_level2_zone',
          nickname: '片区信使-王五',
          email: 'wang.wu@bjdx.edu.cn',
          level: 2,
          status: 'active',
          zone: 'BJDX-A',
          zone_name: '北京大学A区',
          points: 1240,
          task_count: 67,
          success_rate: 94.8,
          created_at: '2024-01-17T16:45:00Z',
          last_active_at: '2024-01-21T07:15:00Z'
        },
        {
          id: '4',
          user_id: 'user_004',
          username: 'courier_level1_building',
          nickname: '楼栋信使-赵六',
          email: 'zhao.liu@bjdx.edu.cn',
          level: 1,
          status: 'pending',
          zone: 'BJDX-A-101',
          zone_name: '北京大学A区101楼',
          points: 350,
          task_count: 23,
          success_rate: 91.3,
          created_at: '2024-01-20T11:00:00Z',
          last_active_at: '2024-01-20T18:20:00Z'
        }
      ]

      const mockTasks: CourierTask[] = [
        {
          id: 'task_001',
          courier_id: '1',
          letter_code: 'LTR-20240121-001',
          status: 'delivered',
          priority: 'high',
          reward: 25,
          pickup_address: '北京大学图书馆',
          delivery_address: '北京大学宿舍A区',
          created_at: '2024-01-21T08:00:00Z',
          updated_at: '2024-01-21T09:30:00Z'
        },
        {
          id: 'task_002',
          courier_id: '2',
          letter_code: 'LTR-20240121-002',
          status: 'in_transit',
          priority: 'normal',
          reward: 15,
          pickup_address: '北京大学教学楼',
          delivery_address: '北京大学宿舍B区',
          created_at: '2024-01-21T09:00:00Z',
          updated_at: '2024-01-21T09:45:00Z'
        }
      ]

      const mockStats: CourierStats = {
        total_couriers: 4,
        active_couriers: 3,
        pending_couriers: 1,
        level_distribution: {
          '1': 1,
          '2': 1,
          '3': 1,
          '4': 1
        },
        total_tasks: 248,
        completed_tasks: 231,
        pending_tasks: 17,
        average_success_rate: 95.2
      }

      setCouriers(mockCouriers)
      setTasks(mockTasks)
      setStats(mockStats)
    } finally {
      setLoading(false)
    }
  }

  // 创建信使
  const handleCreateCourier = async () => {
    try {
      await courierApi.createCourier(createForm)
      
      // 模拟添加到本地状态
      const newCourier: Courier = {
        id: Date.now().toString(),
        user_id: `user_${Date.now()}`,
        ...createForm,
        status: 'pending',
        zone_name: createForm.zone,
        points: 0,
        task_count: 0,
        success_rate: 0,
        created_at: new Date().toISOString(),
        last_active_at: new Date().toISOString()
      }

      setCouriers(prev => [newCourier, ...prev])
      setShowCreateDialog(false)
      setCreateForm({
        username: '',
        email: '',
        nickname: '',
        level: 1,
        zone: '',
        description: ''
      })
    } catch (error) {
      console.error('Failed to create courier:', error)
      alert('创建信使失败，请重试')
    }
  }

  // 过滤信使
  const filteredCouriers = couriers.filter(courier => {
    const matchesSearch = courier.nickname.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         courier.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         courier.zone_name?.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesLevel = levelFilter === 'all' || courier.level.toString() === levelFilter
    const matchesStatus = statusFilter === 'all' || courier.status === statusFilter
    return matchesSearch && matchesLevel && matchesStatus
  })

  // 过滤任务
  const filteredTasks = tasks.filter(task => {
    if (!selectedCourier) return true
    return task.courier_id === selectedCourier.id
  })

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <BackButton href="/admin" />
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-2">
              <Users className="w-8 h-8" />
              信使管理系统
            </h1>
            <p className="text-muted-foreground mt-1">
              管理四级信使层级系统和任务分配
            </p>
          </div>
        </div>
        <Button onClick={() => setShowCreateDialog(true)}>
          <Plus className="w-4 h-4 mr-2" />
          创建信使
        </Button>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总信使数</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_couriers || 0}</div>
            <p className="text-xs text-muted-foreground">
              活跃 {stats?.active_couriers || 0} 人
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">待审核</CardTitle>
            <UserCheck className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.pending_couriers || 0}</div>
            <p className="text-xs text-muted-foreground">
              需要审核的申请
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总任务数</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_tasks || 0}</div>
            <p className="text-xs text-muted-foreground">
              已完成 {stats?.completed_tasks || 0} 个
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">平均成功率</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats?.average_success_rate?.toFixed(1) || 0}%
            </div>
            <p className="text-xs text-muted-foreground">
              系统整体效率
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 主要内容 */}
      <Tabs defaultValue="couriers" className="space-y-6">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="couriers">信使列表</TabsTrigger>
          <TabsTrigger value="hierarchy">层级结构</TabsTrigger>
        </TabsList>

        {/* 信使列表 */}
        <TabsContent value="couriers" className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>信使管理</CardTitle>
                  <CardDescription>查看和管理所有级别的信使</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {/* 搜索和筛选 */}
              <div className="flex flex-col sm:flex-row gap-4 mb-6">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input
                    placeholder="搜索信使姓名、用户名或区域..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                  />
                </div>
                <Select value={levelFilter} onValueChange={setLevelFilter}>
                  <SelectTrigger className="w-full sm:w-32">
                    <SelectValue placeholder="级别筛选" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部级别</SelectItem>
                    <SelectItem value="4">四级(城市)</SelectItem>
                    <SelectItem value="3">三级(校级)</SelectItem>
                    <SelectItem value="2">二级(片区)</SelectItem>
                    <SelectItem value="1">一级(楼栋)</SelectItem>
                  </SelectContent>
                </Select>
                <Select value={statusFilter} onValueChange={setStatusFilter}>
                  <SelectTrigger className="w-full sm:w-32">
                    <SelectValue placeholder="状态筛选" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部状态</SelectItem>
                    <SelectItem value="active">活跃</SelectItem>
                    <SelectItem value="pending">待审核</SelectItem>
                    <SelectItem value="inactive">非活跃</SelectItem>
                    <SelectItem value="suspended">暂停</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* 信使表格 */}
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>信使信息</TableHead>
                      <TableHead>级别</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>管理区域</TableHead>
                      <TableHead>积分</TableHead>
                      <TableHead>任务统计</TableHead>
                      <TableHead>成功率</TableHead>
                      <TableHead>最后活跃</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredCouriers.map((courier) => {
                      const levelInfo = COURIER_LEVELS[courier.level as keyof typeof COURIER_LEVELS] || COURIER_LEVELS[1]
                      const LevelIcon = levelInfo.icon
                      
                      return (
                        <TableRow key={courier.id}>
                          <TableCell>
                            <div>
                              <div className="font-medium">{courier.nickname}</div>
                              <div className="text-sm text-muted-foreground">
                                @{courier.username}
                              </div>
                              <div className="text-xs text-muted-foreground">
                                {courier.email}
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <Badge className={levelInfo.color}>
                              <LevelIcon className="w-3 h-3 mr-1" />
                              {levelInfo.name}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            <Badge className={STATUS_COLORS[courier.status as keyof typeof STATUS_COLORS]}>
                              {STATUS_NAMES[courier.status as keyof typeof STATUS_NAMES]}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            <div className="text-sm">
                              <div>{courier.zone_name}</div>
                              <div className="text-muted-foreground">{courier.zone}</div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <div className="font-medium">{courier.points}</div>
                          </TableCell>
                          <TableCell>
                            <div className="text-sm">
                              <div className="font-medium">{courier.task_count} 个</div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <div className="font-medium">
                              {courier.success_rate.toFixed(1)}%
                            </div>
                          </TableCell>
                          <TableCell>
                            <div className="text-sm">
                              {new Date(courier.last_active_at).toLocaleString()}
                            </div>
                          </TableCell>
                          <TableCell>
                            <div className="flex gap-2">
                              <Button 
                                size="sm" 
                                variant="outline"
                                onClick={() => {
                                  setSelectedCourier(courier)
                                  setShowDetailDialog(true)
                                }}
                              >
                                <Eye className="w-4 h-4" />
                              </Button>
                              <Button 
                                size="sm" 
                                variant="outline"
                                onClick={() => {
                                  setSelectedCourier(courier)
                                  setShowTaskDialog(true)
                                }}
                              >
                                <Truck className="w-4 h-4" />
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 层级结构 */}
        <TabsContent value="hierarchy" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>信使层级结构</CardTitle>
              <CardDescription>四级信使管理体系可视化</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-6">
                {[4, 3, 2, 1].map(level => {
                  const levelInfo = COURIER_LEVELS[level as keyof typeof COURIER_LEVELS]
                  const LevelIcon = levelInfo.icon
                  const levelCouriers = couriers.filter(c => c.level === level)
                  
                  return (
                    <div key={level} className="border rounded-lg p-4">
                      <div className="flex items-center gap-3 mb-4">
                        <div className={`w-10 h-10 rounded-lg ${levelInfo.color} flex items-center justify-center`}>
                          <LevelIcon className="w-5 h-5" />
                        </div>
                        <div>
                          <h3 className="font-semibold text-lg">
                            {level}级 - {levelInfo.name}
                          </h3>
                          <p className="text-sm text-muted-foreground">
                            {levelCouriers.length} 名信使
                          </p>
                        </div>
                        <div className="ml-auto">
                          <Badge variant="outline">
                            {stats?.level_distribution?.[level.toString()] || 0} 人
                          </Badge>
                        </div>
                      </div>
                      
                      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                        {levelCouriers.map(courier => (
                          <div key={courier.id} className="border rounded p-3 hover:bg-gray-50">
                            <div className="flex items-center justify-between">
                              <div>
                                <div className="font-medium">{courier.nickname}</div>
                                <div className="text-sm text-muted-foreground">
                                  {courier.zone_name}
                                </div>
                              </div>
                              <Badge className={STATUS_COLORS[courier.status as keyof typeof STATUS_COLORS]}>
                                {STATUS_NAMES[courier.status as keyof typeof STATUS_NAMES]}
                              </Badge>
                            </div>
                            <div className="mt-2 flex gap-4 text-xs text-muted-foreground">
                              <span>积分: {courier.points}</span>
                              <span>任务: {courier.task_count}</span>
                              <span>成功率: {courier.success_rate.toFixed(1)}%</span>
                            </div>
                          </div>
                        ))}
                        
                        {levelCouriers.length === 0 && (
                          <div className="col-span-full text-center py-4 text-muted-foreground">
                            暂无此级别的信使
                          </div>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* 信使详情对话框 */}
      <Dialog open={showDetailDialog} onOpenChange={setShowDetailDialog}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>信使详情</DialogTitle>
            <DialogDescription>
              查看信使的基本信息和绩效数据
            </DialogDescription>
          </DialogHeader>
          
          {selectedCourier && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label>信使姓名</Label>
                  <div className="mt-1 font-medium">{selectedCourier.nickname}</div>
                </div>
                <div>
                  <Label>用户名</Label>
                  <div className="mt-1 font-medium">@{selectedCourier.username}</div>
                </div>
                <div>
                  <Label>邮箱</Label>
                  <div className="mt-1">{selectedCourier.email}</div>
                </div>
                <div>
                  <Label>级别</Label>
                  <div className="mt-1">
                    <Badge className={COURIER_LEVELS[selectedCourier.level as keyof typeof COURIER_LEVELS].color}>
                      {selectedCourier.level}级 - {COURIER_LEVELS[selectedCourier.level as keyof typeof COURIER_LEVELS].name}
                    </Badge>
                  </div>
                </div>
                <div>
                  <Label>管理区域</Label>
                  <div className="mt-1">
                    <div className="font-medium">{selectedCourier.zone_name}</div>
                    <div className="text-sm text-muted-foreground">{selectedCourier.zone}</div>
                  </div>
                </div>
                <div>
                  <Label>状态</Label>
                  <div className="mt-1">
                    <Badge className={STATUS_COLORS[selectedCourier.status as keyof typeof STATUS_COLORS]}>
                      {STATUS_NAMES[selectedCourier.status as keyof typeof STATUS_NAMES]}
                    </Badge>
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4 pt-4 border-t">
                <div className="text-center">
                  <div className="text-2xl font-bold text-blue-600">{selectedCourier.points}</div>
                  <div className="text-sm text-muted-foreground">总积分</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">{selectedCourier.task_count}</div>
                  <div className="text-sm text-muted-foreground">完成任务</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-purple-600">
                    {selectedCourier.success_rate.toFixed(1)}%
                  </div>
                  <div className="text-sm text-muted-foreground">成功率</div>
                </div>
              </div>

              <div className="pt-4 border-t">
                <Label>注册时间</Label>
                <div className="mt-1">{new Date(selectedCourier.created_at).toLocaleString()}</div>
              </div>

              <div>  
                <Label>最后活跃</Label>
                <div className="mt-1">{new Date(selectedCourier.last_active_at).toLocaleString()}</div>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDetailDialog(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 创建信使对话框 */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>创建信使</DialogTitle>
            <DialogDescription>
              添加新的信使到管理系统
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="username">用户名</Label>
              <Input
                id="username"
                value={createForm.username}
                onChange={(e) => setCreateForm(prev => ({ ...prev, username: e.target.value }))}
                placeholder="输入用户名..."
              />
            </div>
            
            <div>
              <Label htmlFor="email">邮箱</Label>
              <Input
                id="email"
                type="email"
                value={createForm.email}
                onChange={(e) => setCreateForm(prev => ({ ...prev, email: e.target.value }))}
                placeholder="输入邮箱地址..."
              />
            </div>
            
            <div>
              <Label htmlFor="nickname">姓名</Label>
              <Input
                id="nickname"
                value={createForm.nickname}
                onChange={(e) => setCreateForm(prev => ({ ...prev, nickname: e.target.value }))}
                placeholder="输入真实姓名..."
              />
            </div>
            
            <div>
              <Label htmlFor="level">信使级别</Label>
              <Select value={createForm.level.toString()} onValueChange={(value) => setCreateForm(prev => ({ ...prev, level: parseInt(value) }))}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">1级 - 楼栋信使</SelectItem>
                  <SelectItem value="2">2级 - 片区信使</SelectItem>
                  <SelectItem value="3">3级 - 校级信使</SelectItem>
                  <SelectItem value="4">4级 - 城市总代</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <Label htmlFor="zone">管理区域</Label>
              <Input
                id="zone"
                value={createForm.zone}
                onChange={(e) => setCreateForm(prev => ({ ...prev, zone: e.target.value }))}
                placeholder="如：BJDX-A-101"
              />
            </div>
            
            <div>
              <Label htmlFor="description">备注</Label>
              <Textarea
                id="description"
                value={createForm.description}
                onChange={(e) => setCreateForm(prev => ({ ...prev, description: e.target.value }))}
                placeholder="添加备注信息..."
                rows={2}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
              取消
            </Button>
            <Button 
              onClick={handleCreateCourier} 
              disabled={!createForm.username.trim() || !createForm.email.trim() || !createForm.nickname.trim()}
            >
              创建信使
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 任务列表对话框 */}
      <Dialog open={showTaskDialog} onOpenChange={setShowTaskDialog}>
        <DialogContent className="sm:max-w-4xl">
          <DialogHeader>
            <DialogTitle>
              {selectedCourier?.nickname} - 任务列表
            </DialogTitle>
            <DialogDescription>
              查看信使的所有任务记录
            </DialogDescription>
          </DialogHeader>
          
          <div className="max-h-96 overflow-y-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>信件编号</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>优先级</TableHead>
                  <TableHead>取件地址</TableHead>
                  <TableHead>送达地址</TableHead>
                  <TableHead>奖励</TableHead>
                  <TableHead>创建时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredTasks.map((task) => (
                  <TableRow key={task.id}>
                    <TableCell className="font-medium">
                      {task.letter_code}
                    </TableCell>
                    <TableCell>
                      <Badge className={TASK_STATUS_COLORS[task.status as keyof typeof TASK_STATUS_COLORS]}>
                        {TASK_STATUS_NAMES[task.status as keyof typeof TASK_STATUS_NAMES]}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={task.priority === 'high' ? 'destructive' : task.priority === 'normal' ? 'default' : 'secondary'}>
                        {task.priority === 'high' ? '高' : task.priority === 'normal' ? '普通' : '低'}
                      </Badge>
                    </TableCell>
                    <TableCell className="max-w-xs truncate">
                      {task.pickup_address}
                    </TableCell>
                    <TableCell className="max-w-xs truncate">
                      {task.delivery_address}
                    </TableCell>
                    <TableCell>
                      <span className="font-medium text-green-600">
                        +{task.reward}积分
                      </span>
                    </TableCell>
                    <TableCell>
                      {new Date(task.created_at).toLocaleString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            
            {filteredTasks.length === 0 && (
              <div className="text-center py-8 text-muted-foreground">
                该信使暂无任务记录
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowTaskDialog(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}