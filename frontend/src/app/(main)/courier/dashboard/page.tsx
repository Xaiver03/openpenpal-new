'use client'

import React from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Package, 
  TrendingUp, 
  Users, 
  Settings, 
  QrCode,
  CreditCard,
  MapPin,
  Award,
  ClipboardList,
  Building2,
  BarChart3,
  AlertCircle
} from 'lucide-react'
import { useUserStore } from '@/stores/user-store'
import { useEffect, useState } from 'react'
import { apiClient } from '@/lib/api-client'
import { formatDate } from '@/lib/utils'

interface DashboardStats {
  todayTasks: number
  completedTasks: number
  pendingTasks: number
  totalPoints: number
  currentLevel: number
  teamMembers: number
  managedArea: string
  monthlyGrowth: number
}

export default function CourierDashboardPage() {
  const router = useRouter()
  const { user } = useUserStore()
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)

  // 获取信使等级 - 后端角色格式是 courier_level4 而不是 courier_level_4
  const courierLevel = user?.role?.includes('courier_level') 
    ? parseInt(user.role.replace('courier_level', '')) 
    : 0

  useEffect(() => {
    if (!user || courierLevel === 0) {
      router.push('/courier')
      return
    }
    loadDashboardData()
  }, [user, courierLevel, router])

  const loadDashboardData = async () => {
    setLoading(true)
    try {
      const [statsRes, tasksRes] = await Promise.all([
        apiClient.get('/courier/stats'),
        apiClient.get('/courier/tasks?status=pending&limit=5')
      ])

      // 模拟数据，实际应从API获取
      setStats({
        todayTasks: 12,
        completedTasks: 8,
        pendingTasks: 4,
        totalPoints: 2450,
        currentLevel: courierLevel,
        teamMembers: courierLevel > 1 ? 15 : 0,
        managedArea: user?.managed_op_code_prefix || user?.zone_code || '未设置',
        monthlyGrowth: 23.5
      })
    } catch (error) {
      console.error('Failed to load dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  const getLevelName = (level: number) => {
    const names = {
      1: '楼栋投递员',
      2: '片区管理员',
      3: '学校协调员',
      4: '城市总监'
    }
    return names[level as keyof typeof names] || '未知'
  }

  const getLevelColor = (level: number) => {
    const colors = {
      1: 'bg-green-100 text-green-800',
      2: 'bg-blue-100 text-blue-800',
      3: 'bg-purple-100 text-purple-800',
      4: 'bg-orange-100 text-orange-800'
    }
    return colors[level as keyof typeof colors] || 'bg-gray-100 text-gray-800'
  }

  const quickActions = [
    {
      title: '任务管理',
      description: '查看和管理投递任务',
      icon: Package,
      href: '/courier/tasks',
      color: 'text-blue-600'
    },
    {
      title: '扫码投递',
      description: '快速扫码处理信件',
      icon: QrCode,
      href: '/courier/tasks?mode=scan',
      color: 'text-green-600'
    },
    {
      title: '团队管理',
      description: '管理下属信使',
      icon: Users,
      href: '/courier/management/hierarchy',
      color: 'text-purple-600',
      minLevel: 2
    },
    {
      title: 'OP Code管理',
      description: '管理投递点编码',
      icon: MapPin,
      href: '/courier/management/opcode',
      color: 'text-orange-600',
      minLevel: 2
    },
    {
      title: '批量操作',
      description: '批量生成和分配',
      icon: ClipboardList,
      href: '/courier/management/batch',
      color: 'text-indigo-600',
      minLevel: 3
    },
    {
      title: '个人成长',
      description: '查看成长进度',
      icon: TrendingUp,
      href: '/courier/profile',
      color: 'text-pink-600'
    }
  ]

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 页面标题 */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">信使工作台</h1>
            <p className="text-gray-600 mt-2">
              欢迎回来，{user?.username}！
            </p>
          </div>
          <Badge className={getLevelColor(courierLevel) + ' px-4 py-2'}>
            L{courierLevel} {getLevelName(courierLevel)}
          </Badge>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日任务</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.todayTasks || 0}</div>
            <p className="text-xs text-muted-foreground">
              已完成 {stats?.completedTasks || 0} / 待处理 {stats?.pendingTasks || 0}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">积分总数</CardTitle>
            <Award className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.totalPoints || 0}</div>
            <p className="text-xs text-muted-foreground">
              本月增长 {stats?.monthlyGrowth || 0}%
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">管理区域</CardTitle>
            <MapPin className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-lg font-bold font-mono">{stats?.managedArea}</div>
            <p className="text-xs text-muted-foreground">
              {courierLevel >= 4 ? '全城管理' : 
               courierLevel === 3 ? '学校级管理' :
               courierLevel === 2 ? '片区级管理' : '楼栋级管理'}
            </p>
          </CardContent>
        </Card>

        {courierLevel > 1 && (
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">团队成员</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats?.teamMembers || 0}</div>
              <p className="text-xs text-muted-foreground">
                下属信使数量
              </p>
            </CardContent>
          </Card>
        )}
      </div>

      {/* 快速操作 */}
      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-4">快速操作</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {quickActions
            .filter(action => !action.minLevel || courierLevel >= action.minLevel)
            .map((action, index) => (
              <Card 
                key={index} 
                className="cursor-pointer hover:shadow-lg transition-shadow"
                onClick={() => router.push(action.href)}
              >
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <action.icon className={`h-8 w-8 ${action.color}`} />
                    {action.minLevel && (
                      <Badge variant="outline" className="text-xs">
                        L{action.minLevel}+
                      </Badge>
                    )}
                  </div>
                  <CardTitle className="text-lg">{action.title}</CardTitle>
                  <CardDescription>{action.description}</CardDescription>
                </CardHeader>
              </Card>
            ))}
        </div>
      </div>

      {/* 待办提醒 */}
      {stats && stats.pendingTasks > 0 && (
        <Card className="border-orange-200 bg-orange-50">
          <CardHeader>
            <div className="flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-orange-600" />
              <CardTitle className="text-lg">待办提醒</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-gray-700">
              您有 <span className="font-bold text-orange-600">{stats.pendingTasks}</span> 个待处理任务
            </p>
            <Button 
              className="mt-3" 
              size="sm"
              onClick={() => router.push('/courier/tasks?status=pending')}
            >
              立即查看
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  )
}