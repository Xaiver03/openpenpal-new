'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { usePermission } from '@/hooks/use-permission'
import { 
  Package, 
  Truck, 
  MapPin, 
  Clock, 
  User, 
  Search,
  Filter,
  Calendar,
  Target,
  AlertTriangle,
  CheckCircle,
  RefreshCw,
  Phone,
  Navigation,
  Route
} from 'lucide-react'

interface CourierTask {
  id: string
  letterCode: string
  title: string
  senderName: string
  senderPhone?: string
  recipientHint: string
  targetLocation: string
  currentLocation?: string
  priority: 'normal' | 'urgent'
  status: 'pending' | 'collected' | 'in_transit' | 'delivered' | 'failed'
  estimatedTime: number // 预计投递时间（分钟）
  distance: number // 距离（公里）
  createdAt: string
  deadline?: string
  instructions?: string
  reward: number // 积分奖励
}

export default function CourierTasksPage() {
  const { user, isCourier } = usePermission()
  const [tasks, setTasks] = useState<CourierTask[]>([])
  const [filteredTasks, setFilteredTasks] = useState<CourierTask[]>([])
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | CourierTask['status']>('all')
  const [priorityFilter, setPriorityFilter] = useState<'all' | 'normal' | 'urgent'>('all')
  const [sortBy, setSortBy] = useState<'deadline' | 'distance' | 'reward' | 'created'>('deadline')
  const [isLoading, setIsLoading] = useState(true)

  // 权限检查
  if (!isCourier()) {
    return (
      <div className="container max-w-6xl mx-auto px-4 py-8">
        <div className="text-center py-16">
          <AlertTriangle className="w-16 h-16 text-amber-600 mx-auto mb-4" />
          <h2 className="text-2xl font-bold text-amber-900 mb-2">权限不足</h2>
          <p className="text-amber-700 mb-6">
            只有信使才能查看任务中心。如需申请成为信使，请前往信使中心。
          </p>
          <Button asChild className="bg-amber-600 hover:bg-amber-700 text-white">
            <Link href="/courier/apply">
              申请成为信使
            </Link>
          </Button>
        </div>
      </div>
    )
  }

  // 模拟获取任务数据
  useEffect(() => {
    const fetchTasks = async () => {
      setIsLoading(true)
      
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      const mockTasks: CourierTask[] = [
        {
          id: '1',
          letterCode: 'OP1K2L3M4N5O',
          title: '给朋友的问候信',
          senderName: '小明',
          senderPhone: '138****5678',
          recipientHint: '北大宿舍楼，李同学',
          targetLocation: '北京大学宿舍楼32栋',
          priority: 'normal',
          status: 'pending',
          estimatedTime: 25,
          distance: 1.2,
          createdAt: '2024-01-20T09:00:00Z',
          deadline: '2024-01-21T18:00:00Z',
          instructions: '请投递到宿舍管理员处',
          reward: 10
        },
        {
          id: '2',
          letterCode: 'OP2K3L4M5N6P',
          title: '紧急通知信件',
          senderName: '王老师',
          senderPhone: '139****1234',
          recipientHint: '计算机学院，张教授',
          targetLocation: '计算机学院办公楼203室',
          currentLocation: '计算机学院大门',
          priority: 'urgent',
          status: 'in_transit',
          estimatedTime: 15,
          distance: 0.8,
          createdAt: '2024-01-20T11:30:00Z',
          deadline: '2024-01-20T17:00:00Z',
          instructions: '请直接交给本人，如不在请联系',
          reward: 20
        },
        {
          id: '3',
          letterCode: 'OP3K4L5M6N7P',
          title: '生日祝福卡片',
          senderName: '小红',
          recipientHint: '图书馆，管理员阿姨',
          targetLocation: '图书馆服务台',
          priority: 'normal',
          status: 'collected',
          estimatedTime: 20,
          distance: 0.5,
          createdAt: '2024-01-19T16:00:00Z',
          deadline: '2024-01-22T12:00:00Z',
          instructions: '工作时间投递',
          reward: 8
        },
        {
          id: '4',
          letterCode: 'OP4K5L6M7N8P',
          title: '感谢信',
          senderName: '小华',
          recipientHint: '食堂，李师傅',
          targetLocation: '第一食堂二层',
          priority: 'normal',
          status: 'delivered',
          estimatedTime: 10,
          distance: 0.3,
          createdAt: '2024-01-19T14:00:00Z',
          deadline: '2024-01-21T20:00:00Z',
          instructions: '饭点时间投递',
          reward: 6
        }
      ]
      
      setTasks(mockTasks)
      setIsLoading(false)
    }

    fetchTasks()
  }, [])

  // 过滤和排序任务
  useEffect(() => {
    let filtered = tasks.filter(task => {
      const matchesSearch = task.letterCode.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           task.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           task.senderName.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           task.recipientHint.toLowerCase().includes(searchTerm.toLowerCase())
      
      const matchesStatus = statusFilter === 'all' || task.status === statusFilter
      const matchesPriority = priorityFilter === 'all' || task.priority === priorityFilter
      
      return matchesSearch && matchesStatus && matchesPriority
    })

    // 排序
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'deadline':
          if (!a.deadline) return 1
          if (!b.deadline) return -1
          return new Date(a.deadline).getTime() - new Date(b.deadline).getTime()
        case 'distance':
          return a.distance - b.distance
        case 'reward':
          return b.reward - a.reward
        case 'created':
          return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        default:
          return 0
      }
    })

    setFilteredTasks(filtered)
  }, [tasks, searchTerm, statusFilter, priorityFilter, sortBy])

  const getStatusInfo = (status: CourierTask['status']) => {
    switch (status) {
      case 'pending':
        return {
          label: '待收取',
          color: 'bg-gray-100 text-gray-800 border-gray-200',
          icon: Package
        }
      case 'collected':
        return {
          label: '已收取',
          color: 'bg-blue-100 text-blue-800 border-blue-200',
          icon: Package
        }
      case 'in_transit':
        return {
          label: '投递中',
          color: 'bg-orange-100 text-orange-800 border-orange-200',
          icon: Truck
        }
      case 'delivered':
        return {
          label: '已投递',
          color: 'bg-green-100 text-green-800 border-green-200',
          icon: CheckCircle
        }
      case 'failed':
        return {
          label: '投递失败',
          color: 'bg-red-100 text-red-800 border-red-200',
          icon: AlertTriangle
        }
    }
  }

  const getPriorityInfo = (priority: CourierTask['priority']) => {
    switch (priority) {
      case 'urgent':
        return {
          label: '紧急',
          color: 'bg-red-100 text-red-800 border-red-200'
        }
      case 'normal':
        return {
          label: '普通',
          color: 'bg-gray-100 text-gray-800 border-gray-200'
        }
    }
  }

  const handleAcceptTask = async (taskId: string) => {
    // 模拟接受任务
    setTasks(prev => prev.map(task => 
      task.id === taskId ? { ...task, status: 'collected' as const } : task
    ))
  }

  const taskStats = {
    total: tasks.length,
    pending: tasks.filter(t => t.status === 'pending').length,
    inProgress: tasks.filter(t => t.status === 'collected' || t.status === 'in_transit').length,
    completed: tasks.filter(t => t.status === 'delivered').length,
    totalReward: tasks.filter(t => t.status === 'delivered').reduce((sum, t) => sum + t.reward, 0)
  }

  return (
    <div className="container max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-amber-900 mb-2">
          信使任务中心
        </h1>
        <p className="text-amber-700">
          管理您的投递任务，获得积分奖励。欢迎您，{user?.nickname}！
        </p>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4 mb-8">
        <Card className="border-amber-200">
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-amber-900">{taskStats.total}</div>
            <p className="text-sm text-amber-600">总任务数</p>
          </CardContent>
        </Card>
        <Card className="border-amber-200">
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-gray-600">{taskStats.pending}</div>
            <p className="text-sm text-amber-600">待接取</p>
          </CardContent>
        </Card>
        <Card className="border-amber-200">
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-orange-600">{taskStats.inProgress}</div>
            <p className="text-sm text-amber-600">进行中</p>
          </CardContent>
        </Card>
        <Card className="border-amber-200">
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-green-600">{taskStats.completed}</div>
            <p className="text-sm text-amber-600">已完成</p>
          </CardContent>
        </Card>
        <Card className="border-amber-200">
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-amber-600">{taskStats.totalReward}</div>
            <p className="text-sm text-amber-600">今日积分</p>
          </CardContent>
        </Card>
      </div>

      {/* 筛选和搜索 */}
      <Card className="border-amber-200 mb-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-amber-900">
            <Filter className="h-5 w-5" />
            筛选条件
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label className="text-sm font-medium text-amber-900 mb-2 block">搜索</label>
              <div className="relative">
                <Search className="absolute left-3 top-3 h-4 w-4 text-amber-500" />
                <Input
                  placeholder="搜索编号、标题或收发人..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 border-amber-300 focus:border-amber-500"
                />
              </div>
            </div>
            
            <div>
              <label className="text-sm font-medium text-amber-900 mb-2 block">状态</label>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value as any)}
                className="w-full p-2 border border-amber-300 rounded-md bg-white focus:border-amber-500 focus:outline-none"
              >
                <option value="all">全部状态</option>
                <option value="pending">待收取</option>
                <option value="collected">已收取</option>
                <option value="in_transit">投递中</option>
                <option value="delivered">已投递</option>
                <option value="failed">投递失败</option>
              </select>
            </div>

            <div>
              <label className="text-sm font-medium text-amber-900 mb-2 block">优先级</label>
              <select
                value={priorityFilter}
                onChange={(e) => setPriorityFilter(e.target.value as any)}
                className="w-full p-2 border border-amber-300 rounded-md bg-white focus:border-amber-500 focus:outline-none"
              >
                <option value="all">全部优先级</option>
                <option value="urgent">紧急</option>
                <option value="normal">普通</option>
              </select>
            </div>

            <div>
              <label className="text-sm font-medium text-amber-900 mb-2 block">排序</label>
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as any)}
                className="w-full p-2 border border-amber-300 rounded-md bg-white focus:border-amber-500 focus:outline-none"
              >
                <option value="deadline">截止时间</option>
                <option value="distance">距离</option>
                <option value="reward">积分奖励</option>
                <option value="created">创建时间</option>
              </select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 任务列表 */}
      {isLoading ? (
        <div className="text-center py-16">
          <RefreshCw className="w-8 h-8 animate-spin text-amber-600 mx-auto mb-4" />
          <p className="text-amber-700">加载任务中...</p>
        </div>
      ) : filteredTasks.length > 0 ? (
        <div className="space-y-4">
          {filteredTasks.map(task => {
            const statusInfo = getStatusInfo(task.status)
            const priorityInfo = getPriorityInfo(task.priority)
            const StatusIcon = statusInfo.icon
            
            return (
              <Card key={task.id} className="border-amber-200 hover:border-amber-400 transition-colors">
                <CardContent className="p-6">
                  <div className="flex items-start justify-between mb-4">
                    <div className="flex items-start gap-4">
                      <div className="w-12 h-12 bg-amber-100 rounded-lg flex items-center justify-center">
                        <StatusIcon className="h-6 w-6 text-amber-600" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2 mb-1">
                          <h3 className="font-semibold text-amber-900">{task.title}</h3>
                          <Badge className={priorityInfo.color}>
                            {priorityInfo.label}
                          </Badge>
                          <Badge className={statusInfo.color}>
                            {statusInfo.label}
                          </Badge>
                        </div>
                        <p className="text-sm text-amber-600 font-mono mb-1">{task.letterCode}</p>
                        <div className="flex items-center gap-4 text-sm text-amber-700">
                          <span className="flex items-center gap-1">
                            <User className="h-3 w-3" />
                            {task.senderName}
                          </span>
                          {task.senderPhone && (
                            <span className="flex items-center gap-1">
                              <Phone className="h-3 w-3" />
                              {task.senderPhone}
                            </span>
                          )}
                          <span className="flex items-center gap-1">
                            <Target className="h-3 w-3" />
                            {task.recipientHint}
                          </span>
                        </div>
                      </div>
                    </div>
                    
                    <div className="text-right">
                      <div className="text-lg font-bold text-amber-600">+{task.reward}</div>
                      <div className="text-xs text-amber-500">积分</div>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                    <div className="flex items-center gap-2 text-sm text-amber-700">
                      <MapPin className="h-4 w-4" />
                      <span>{task.targetLocation}</span>
                    </div>
                    <div className="flex items-center gap-2 text-sm text-amber-700">
                      <Route className="h-4 w-4" />
                      <span>{task.distance}km · {task.estimatedTime}分钟</span>
                    </div>
                    {task.deadline && (
                      <div className="flex items-center gap-2 text-sm text-amber-700">
                        <Clock className="h-4 w-4" />
                        <span>截止：{new Date(task.deadline).toLocaleString('zh-CN', {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit'
                        })}</span>
                      </div>
                    )}
                  </div>

                  {task.instructions && (
                    <div className="mb-4 p-3 bg-amber-50 rounded-md border border-amber-200">
                      <p className="text-sm text-amber-700">
                        <strong>投递说明：</strong>{task.instructions}
                      </p>
                    </div>
                  )}

                  <div className="flex items-center justify-between">
                    <div className="text-xs text-amber-500">
                      创建时间：{new Date(task.createdAt).toLocaleString('zh-CN')}
                    </div>
                    
                    <div className="flex gap-2">
                      {task.status === 'pending' && (
                        <Button
                          onClick={() => handleAcceptTask(task.id)}
                          size="sm"
                          className="bg-amber-600 hover:bg-amber-700 text-white"
                        >
                          <Package className="h-3 w-3 mr-1" />
                          接受任务
                        </Button>
                      )}
                      
                      {(task.status === 'collected' || task.status === 'in_transit') && (
                        <Button asChild size="sm" variant="outline" className="border-amber-300 text-amber-700">
                          <Link href="/courier/scan">
                            <Navigation className="h-3 w-3 mr-1" />
                            扫码更新
                          </Link>
                        </Button>
                      )}

                      <Button size="sm" variant="outline" className="border-amber-300 text-amber-700">
                        查看详情
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      ) : (
        <div className="text-center py-16">
          <Package className="w-16 h-16 text-amber-400 mx-auto mb-4" />
          <h3 className="text-xl font-semibold text-amber-900 mb-2">暂无任务</h3>
          <p className="text-amber-700 mb-6">
            当前没有符合条件的任务，请稍后再来查看
          </p>
          <Button onClick={() => {
            setSearchTerm('')
            setStatusFilter('all')
            setPriorityFilter('all')
          }} variant="outline" className="border-amber-300 text-amber-700">
            <RefreshCw className="h-4 w-4 mr-2" />
            重置筛选
          </Button>
        </div>
      )}
    </div>
  )
}