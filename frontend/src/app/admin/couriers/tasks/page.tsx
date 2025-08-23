'use client'

import React, { useState, useEffect } from 'react'
import { 
  Package, 
  Search, 
  Filter, 
  Plus,
  Eye,
  Edit,
  MapPin,
  Clock,
  User,
  CheckCircle,
  XCircle,
  ArrowRight,
  Truck,
  AlertCircle
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
import { courierApi, type CourierTask, type Courier } from '@/lib/api/courier'

// 常量定义
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

const PRIORITY_COLORS = {
  low: 'bg-gray-100 text-gray-800',
  normal: 'bg-blue-100 text-blue-800',
  high: 'bg-red-100 text-red-800'
}

const PRIORITY_NAMES = {
  low: '低优先级',
  normal: '普通',
  high: '高优先级'
}

export default function CourierTasksPage() {
  const { user, hasPermission } = usePermission()
  const [tasks, setTasks] = useState<CourierTask[]>([])
  const [couriers, setCouriers] = useState<Courier[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [priorityFilter, setPriorityFilter] = useState<string>('all')
  const [courierFilter, setCourierFilter] = useState<string>('all')
  
  // 对话框状态
  const [selectedTask, setSelectedTask] = useState<CourierTask | null>(null)
  const [showDetailDialog, setShowDetailDialog] = useState(false)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showAssignDialog, setShowAssignDialog] = useState(false)
  
  // 创建任务表单
  const [createForm, setCreateForm] = useState({
    letter_code: '',
    priority: 'normal' as const,
    pickup_address: '',
    delivery_address: '',
    reward: 15,
    description: '',
    courier_id: ''
  })

  // 分配任务表单
  const [assignForm, setAssignForm] = useState({
    courier_id: '',
    reason: ''
  })

  // 加载数据
  const loadData = async () => {
    setLoading(true)
    try {
      // 并行调用所有API
      const [tasksRes, couriersRes] = await Promise.all([
        courierApi.getAllTasks({ limit: 100 }),
        courierApi.getAllCouriers({ status: 'active' })
      ])
      
      // 处理API响应
      if (tasksRes.data && typeof tasksRes.data === 'object' && 'tasks' in tasksRes.data) {
        setTasks((tasksRes.data as any).tasks)
      }
      
      if (couriersRes.data && typeof couriersRes.data === 'object' && 'couriers' in couriersRes.data) {
        setCouriers((couriersRes.data as any).couriers)
      }
    } catch (error) {
      console.error('Failed to load tasks data:', error)
      
      // 设置空数据而不是mock数据
      setTasks([])
      setCouriers([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  // 权限检查
  if (!user || !hasPermission(PERMISSIONS.SYSTEM_CONFIG)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Package className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">访问权限不足</h2>
            <p className="text-gray-600 mb-4">
              您没有访问任务管理的权限
            </p>
            <Button asChild variant="outline">
              <a href="/admin">返回管理控制台</a>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  // 创建任务
  const handleCreateTask = async () => {
    try {
      await courierApi.createTask({
        ...createForm,
        courierId: createForm.courier_id
      })
      
      // 模拟添加到本地状态
      const newTask: CourierTask = {
        id: Date.now().toString(),
        ...createForm,
        status: 'pending',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }

      setTasks(prev => [newTask, ...prev])
      setShowCreateDialog(false)
      setCreateForm({
        letter_code: '',
        priority: 'normal',
        pickup_address: '',
        delivery_address: '',
        reward: 15,
        description: '',
        courier_id: ''
      })
    } catch (error) {
      console.error('Failed to create task:', error)
      alert('创建任务失败，请重试')
    }
  }

  // 分配任务
  const handleAssignTask = async () => {
    if (!selectedTask) return
    
    try {
      // await courierApi.assignTask(selectedTask.id, assignForm)
      
      // 模拟更新本地状态
      setTasks(prev => prev.map(task => 
        task.id === selectedTask.id 
          ? { ...task, courier_id: assignForm.courier_id, status: 'accepted' as const }
          : task
      ))

      setShowAssignDialog(false)
      setSelectedTask(null)
      setAssignForm({ courier_id: '', reason: '' })
    } catch (error) {
      console.error('Failed to assign task:', error)
      alert('分配任务失败，请重试')
    }
  }

  // 过滤任务
  const filteredTasks = tasks.filter(task => {
    const matchesSearch = task.letter_code.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         task.pickup_address.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         task.delivery_address.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesStatus = statusFilter === 'all' || task.status === statusFilter
    const matchesPriority = priorityFilter === 'all' || task.priority === priorityFilter
    const matchesCourier = courierFilter === 'all' || task.courier_id === courierFilter
    return matchesSearch && matchesStatus && matchesPriority && matchesCourier
  })

  // 获取信使姓名
  const getCourierName = (courierId: string) => {
    const courier = couriers.find(c => c.id === courierId)
    return courier ? courier.nickname : '未分配'
  }

  // 统计数据
  const stats = {
    total: tasks.length,
    pending: tasks.filter(t => t.status === 'pending').length,
    in_progress: tasks.filter(t => ['accepted', 'collected', 'in_transit'].includes(t.status)).length,
    completed: tasks.filter(t => t.status === 'delivered').length,
    failed: tasks.filter(t => t.status === 'failed').length
  }

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
          <BackButton href="/admin/couriers" />
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-2">
              <Package className="w-8 h-8" />
              信使任务管理
            </h1>
            <p className="text-muted-foreground mt-1">
              管理信使配送任务和状态跟踪
            </p>
          </div>
        </div>
        <Button onClick={() => setShowCreateDialog(true)}>
          <Plus className="w-4 h-4 mr-2" />
          创建任务
        </Button>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总任务数</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total}</div>
            <p className="text-xs text-muted-foreground">
              全部任务
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">待分配</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.pending}</div>
            <p className="text-xs text-muted-foreground">
              等待分配
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">进行中</CardTitle>
            <Truck className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.in_progress}</div>
            <p className="text-xs text-muted-foreground">
              配送中
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">已完成</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.completed}</div>
            <p className="text-xs text-muted-foreground">
              成功投递
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">失败任务</CardTitle>
            <XCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.failed}</div>
            <p className="text-xs text-muted-foreground">
              需要处理
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 主要内容 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>任务列表</CardTitle>
              <CardDescription>查看和管理所有配送任务</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {/* 搜索和筛选 */}
          <div className="flex flex-col sm:flex-row gap-4 mb-6">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="搜索信件编号或地址..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-full sm:w-32">
                <SelectValue placeholder="状态筛选" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="pending">待分配</SelectItem>
                <SelectItem value="accepted">已接受</SelectItem>
                <SelectItem value="collected">已收件</SelectItem>
                <SelectItem value="in_transit">运输中</SelectItem>
                <SelectItem value="delivered">已投递</SelectItem>
                <SelectItem value="failed">失败</SelectItem>
              </SelectContent>
            </Select>
            <Select value={priorityFilter} onValueChange={setPriorityFilter}>
              <SelectTrigger className="w-full sm:w-32">
                <SelectValue placeholder="优先级" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部优先级</SelectItem>
                <SelectItem value="high">高优先级</SelectItem>
                <SelectItem value="normal">普通</SelectItem>
                <SelectItem value="low">低优先级</SelectItem>
              </SelectContent>
            </Select>
            <Select value={courierFilter} onValueChange={setCourierFilter}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue placeholder="信使筛选" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部信使</SelectItem>
                <SelectItem value="">未分配</SelectItem>
                {couriers.map(courier => (
                  <SelectItem key={courier.id} value={courier.id}>
                    {courier.nickname}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* 任务表格 */}
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>信件编号</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>优先级</TableHead>
                  <TableHead>信使</TableHead>
                  <TableHead>取件地址</TableHead>
                  <TableHead>投递地址</TableHead>
                  <TableHead>奖励</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredTasks.map((task) => (
                  <TableRow key={task.id}>
                    <TableCell className="font-medium">
                      {task.letter_code}
                    </TableCell>
                    <TableCell>
                      <Badge className={TASK_STATUS_COLORS[task.status]}>
                        {TASK_STATUS_NAMES[task.status]}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge className={PRIORITY_COLORS[task.priority]}>
                        {PRIORITY_NAMES[task.priority]}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <User className="w-4 h-4 text-muted-foreground" />
                        <span>{getCourierName(task.courier_id)}</span>
                      </div>
                    </TableCell>
                    <TableCell className="max-w-xs">
                      <div className="truncate" title={task.pickup_address}>
                        <MapPin className="w-4 h-4 inline mr-1 text-muted-foreground" />
                        {task.pickup_address}
                      </div>
                    </TableCell>
                    <TableCell className="max-w-xs">
                      <div className="truncate" title={task.delivery_address}>
                        <MapPin className="w-4 h-4 inline mr-1 text-muted-foreground" />
                        {task.delivery_address}
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className="font-medium text-green-600">
                        +{task.reward}积分
                      </span>
                    </TableCell>
                    <TableCell>
                      {new Date(task.created_at).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button 
                          size="sm" 
                          variant="outline"
                          onClick={() => {
                            setSelectedTask(task)
                            setShowDetailDialog(true)
                          }}
                        >
                          <Eye className="w-4 h-4" />
                        </Button>
                        {task.status === 'pending' && (
                          <Button 
                            size="sm" 
                            onClick={() => {
                              setSelectedTask(task)
                              setShowAssignDialog(true)
                            }}
                          >
                            <User className="w-4 h-4 mr-1" />
                            分配
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            
            {filteredTasks.length === 0 && (
              <div className="text-center py-8 text-muted-foreground">
                没有找到匹配的任务
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* 任务详情对话框 */}
      <Dialog open={showDetailDialog} onOpenChange={setShowDetailDialog}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>任务详情</DialogTitle>
            <DialogDescription>
              查看任务的详细信息和状态
            </DialogDescription>
          </DialogHeader>
          
          {selectedTask && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label>信件编号</Label>
                  <div className="mt-1 font-medium">{selectedTask.letter_code}</div>
                </div>
                <div>
                  <Label>任务状态</Label>
                  <div className="mt-1">
                    <Badge className={TASK_STATUS_COLORS[selectedTask.status]}>
                      {TASK_STATUS_NAMES[selectedTask.status]}
                    </Badge>
                  </div>
                </div>
                <div>
                  <Label>优先级</Label>
                  <div className="mt-1">
                    <Badge className={PRIORITY_COLORS[selectedTask.priority]}>
                      {PRIORITY_NAMES[selectedTask.priority]}
                    </Badge>
                  </div>
                </div>
                <div>
                  <Label>奖励积分</Label>
                  <div className="mt-1 font-medium text-green-600">
                    +{selectedTask.reward}积分
                  </div>
                </div>
              </div>

              <div>
                <Label>负责信使</Label>
                <div className="mt-1 font-medium">
                  {getCourierName(selectedTask.courier_id)}
                </div>
              </div>

              <div>
                <Label>取件地址</Label>
                <div className="mt-1 p-2 bg-gray-50 rounded text-sm">
                  {selectedTask.pickup_address}
                </div>
              </div>

              <div>
                <Label>投递地址</Label>
                <div className="mt-1 p-2 bg-gray-50 rounded text-sm">
                  {selectedTask.delivery_address}
                </div>
              </div>

              {selectedTask.description && (
                <div>
                  <Label>任务描述</Label>
                  <div className="mt-1 p-2 bg-gray-50 rounded text-sm">
                    {selectedTask.description}
                  </div>
                </div>
              )}

              <div className="grid grid-cols-2 gap-4 pt-4 border-t">
                <div>
                  <Label>创建时间</Label>
                  <div className="mt-1 text-sm">
                    {new Date(selectedTask.created_at).toLocaleString()}
                  </div>
                </div>
                <div>
                  <Label>最后更新</Label>
                  <div className="mt-1 text-sm">
                    {new Date(selectedTask.updated_at).toLocaleString()}
                  </div>
                </div>
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

      {/* 创建任务对话框 */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>创建配送任务</DialogTitle>
            <DialogDescription>
              为信件创建新的配送任务
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="letter_code">信件编号</Label>
              <Input
                id="letter_code"
                value={createForm.letter_code}
                onChange={(e) => setCreateForm(prev => ({ ...prev, letter_code: e.target.value }))}
                placeholder="如：LTR-20240121-001"
              />
            </div>
            
            <div>
              <Label htmlFor="priority">优先级</Label>
              <Select value={createForm.priority} onValueChange={(value: any) => setCreateForm(prev => ({ ...prev, priority: value }))}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">低优先级</SelectItem>
                  <SelectItem value="normal">普通</SelectItem>
                  <SelectItem value="high">高优先级</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <Label htmlFor="pickup_address">取件地址</Label>
              <Input
                id="pickup_address"
                value={createForm.pickup_address}
                onChange={(e) => setCreateForm(prev => ({ ...prev, pickup_address: e.target.value }))}
                placeholder="输入取件地址..."
              />
            </div>
            
            <div>
              <Label htmlFor="delivery_address">投递地址</Label>
              <Input
                id="delivery_address"
                value={createForm.delivery_address}
                onChange={(e) => setCreateForm(prev => ({ ...prev, delivery_address: e.target.value }))}
                placeholder="输入投递地址..."
              />
            </div>
            
            <div>
              <Label htmlFor="reward">奖励积分</Label>
              <Input
                id="reward"
                type="number"
                value={createForm.reward}
                onChange={(e) => setCreateForm(prev => ({ ...prev, reward: parseInt(e.target.value) || 15 }))}
                min="5"
                max="100"
              />
            </div>

            <div>
              <Label htmlFor="courier_id">指定信使（可选）</Label>
              <Select value={createForm.courier_id} onValueChange={(value) => setCreateForm(prev => ({ ...prev, courier_id: value }))}>
                <SelectTrigger>
                  <SelectValue placeholder="自动分配" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">自动分配</SelectItem>
                  {couriers.map(courier => (
                    <SelectItem key={courier.id} value={courier.id}>
                      {courier.nickname} - {courier.zone_name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <Label htmlFor="description">任务描述</Label>
              <Textarea
                id="description"
                value={createForm.description}
                onChange={(e) => setCreateForm(prev => ({ ...prev, description: e.target.value }))}
                placeholder="添加任务相关说明..."
                rows={2}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
              取消
            </Button>
            <Button 
              onClick={handleCreateTask} 
              disabled={!createForm.letter_code.trim() || !createForm.pickup_address.trim() || !createForm.delivery_address.trim()}
            >
              创建任务
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 分配任务对话框 */}
      <Dialog open={showAssignDialog} onOpenChange={setShowAssignDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>分配任务</DialogTitle>
            <DialogDescription>
              为任务 {selectedTask?.letter_code} 分配信使
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="assign_courier">选择信使</Label>
              <Select value={assignForm.courier_id} onValueChange={(value) => setAssignForm(prev => ({ ...prev, courier_id: value }))}>
                <SelectTrigger>
                  <SelectValue placeholder="选择信使..." />
                </SelectTrigger>
                <SelectContent>
                  {couriers.map(courier => (
                    <SelectItem key={courier.id} value={courier.id}>
                      <div className="flex items-center justify-between w-full">
                        <span>{courier.nickname}</span>
                        <span className="text-sm text-muted-foreground ml-2">
                          {courier.zone_name} | 成功率: {courier.success_rate.toFixed(1)}%
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <Label htmlFor="assign_reason">分配原因</Label>
              <Textarea
                id="assign_reason"
                value={assignForm.reason}
                onChange={(e) => setAssignForm(prev => ({ ...prev, reason: e.target.value }))}
                placeholder="请说明分配原因..."
                rows={2}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAssignDialog(false)}>
              取消
            </Button>
            <Button 
              onClick={handleAssignTask} 
              disabled={!assignForm.courier_id}
            >
              确认分配
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}