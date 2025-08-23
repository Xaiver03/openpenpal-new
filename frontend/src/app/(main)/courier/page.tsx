'use client'

import React, { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  MapPin, 
  Package, 
  BarChart, 
  Settings,
  Users,
  Truck,
  Building,
  Home,
  School,
  FileText,
  ShieldCheck,
  RefreshCw,
  AlertCircle,
  TrendingUp
} from 'lucide-react'
import Link from 'next/link'
import { useUserStore } from '@/stores/user-store'
import { CourierCenterNavigation } from '@/components/courier/CourierCenterNavigation'
import { CourierService, type CourierStats } from '@/lib/api/courier-service'

export default function CourierDashboard() {
  const router = useRouter()
  const { user } = useUserStore()
  const [stats, setStats] = useState<CourierStats | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // 获取信使级别 - 后端角色格式是 courier_level4 而不是 courier_level_4
  const courierLevel = user?.role?.includes('courier_level') 
    ? parseInt(user.role.replace('courier_level', '')) 
    : 0

  // 加载信使统计数据
  useEffect(() => {
    const loadCourierStats = async () => {
      if (!user || courierLevel === 0) {
        setIsLoading(false)
        return
      }

      try {
        setIsLoading(true)
        setError(null)
        
        const courierStats = await CourierService.getCourierStats()
        setStats(courierStats)
      } catch (err) {
        console.error('Failed to load courier stats:', err)
        setError(err instanceof Error ? err.message : '加载统计数据失败')
      } finally {
        setIsLoading(false)
      }
    }

    loadCourierStats()
  }, [user, courierLevel])

  // 刷新数据
  const handleRefresh = () => {
    if (user && courierLevel > 0) {
      const loadStats = async () => {
        try {
          setError(null)
          const courierStats = await CourierService.getCourierStats()
          setStats(courierStats)
        } catch (err) {
          setError(err instanceof Error ? err.message : '刷新失败')
        }
      }
      loadStats()
    }
  }

  // 根据级别定义可用功能
  const getAvailableFeatures = () => {
    const baseFeatures = [
      {
        title: '任务管理',
        description: '查看和管理投递任务',
        icon: Package,
        href: '/courier/tasks',
        color: 'bg-blue-600'
      },
      {
        title: '数据统计',
        description: '查看投递数据和绩效',
        icon: BarChart,
        href: '/courier/analytics',
        color: 'bg-purple-600'
      }
    ]

    // 所有信使都可以访问 OP Code 管理（权限不同）
    baseFeatures.push({
      title: 'OP Code 管理',
      description: courierLevel === 1 ? '查看和编辑投递点编码' : '管理负责区域的编码',
      icon: MapPin,
      href: '/courier/opcode-manage',
      color: 'bg-green-600'
    })

    // 一级信使 - 楼栋管理
    if (courierLevel === 1) {
      baseFeatures.push({
        title: '楼栋管理',
        description: '管理负责楼栋的投递点',
        icon: Home,
        href: '/courier/building-manage',
        color: 'bg-blue-500'
      })
    }

    // 二级信使 - 片区管理
    if (courierLevel === 2) {
      baseFeatures.push({
        title: '片区管理',
        description: '管理片区内的一级信使和投递点',
        icon: Truck,
        href: '/courier/zone-manage',
        color: 'bg-green-500'
      })
    }

    // 三级信使 - 学校管理
    if (courierLevel === 3) {
      baseFeatures.push({
        title: '学校管理',
        description: '管理校内片区和二级信使',
        icon: School,
        href: '/courier/school-manage',
        color: 'bg-purple-500'
      })
    }

    // 四级信使 - 城市管理
    if (courierLevel === 4) {
      baseFeatures.push({
        title: '城市管理',
        description: '管理城市内所有学校和三级信使',
        icon: Building,
        href: '/courier/city-manage',
        color: 'bg-red-500'
      })
    }


    // L3/L4信使有批量管理功能
    if (courierLevel >= 3) {
      baseFeatures.push({
        title: '批量管理',
        description: '批量生成和管理条码',
        icon: Settings,
        href: '/courier/batch',
        color: 'bg-yellow-600'
      })
    }

    // 信使成长系统
    baseFeatures.push({
      title: '成长中心',
      description: '查看成长进度和申请晋升',
      icon: TrendingUp,
      href: '/courier/growth',
      color: 'bg-indigo-600'
    })

    return baseFeatures
  }

  // 获取级别信息
  const getLevelInfo = () => {
    switch (courierLevel) {
      case 1:
        return {
          icon: <Home className="h-6 w-6" />,
          title: '一级信使（楼栋投递员）',
          description: '负责具体楼栋的信件投递',
          color: 'bg-blue-500'
        }
      case 2:
        return {
          icon: <Truck className="h-6 w-6" />,
          title: '二级信使（片区管理员）',
          description: '管理片区投递点和投递任务分配',
          color: 'bg-green-500'
        }
      case 3:
        return {
          icon: <School className="h-6 w-6" />,
          title: '三级信使（学校协调员）',
          description: '管理学校级信使团队和区域编码',
          color: 'bg-purple-500'
        }
      case 4:
        return {
          icon: <Building className="h-6 w-6" />,
          title: '四级信使（城市总监）',
          description: '管理城市级信使网络和学校编码',
          color: 'bg-red-500'
        }
      default:
        return null
    }
  }

  const levelInfo = getLevelInfo()
  const features = getAvailableFeatures()

  if (!user || courierLevel === 0) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <ShieldCheck className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">需要信使权限</h2>
            <p className="text-gray-600 mb-4">
              您需要成为信使才能访问此页面
            </p>
            <Button onClick={() => router.push('/')} variant="outline">
              返回首页
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 信使中心导航 */}
      <CourierCenterNavigation currentPage="home" className="mb-6" />
      
      {/* 欢迎横幅 */}
      <div className="mb-8">
        <Card className={`${levelInfo?.color} text-white`}>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                {levelInfo?.icon}
                <div>
                  <CardTitle className="text-2xl">欢迎回来，{user.nickname}</CardTitle>
                  <CardDescription className="text-white/80">
                    {levelInfo?.title} - {levelInfo?.description}
                  </CardDescription>
                </div>
              </div>
              <Badge variant="secondary" className="text-lg px-4 py-2">
                管理范围: {
                  isLoading ? '...' : 
                  (stats?.courierInfo?.managedOPCodePrefix || 
                   stats?.courierInfo?.zoneCode || 
                   user.managed_op_code_prefix || 
                   user.courierInfo?.managed_op_code_prefix || 
                   user.courierInfo?.zoneCode || 
                   '未设置')
                }
              </Badge>
            </div>
          </CardHeader>
        </Card>
      </div>

      {/* 错误提示 */}
      {error && (
        <Alert variant="destructive" className="mb-6">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription className="flex items-center justify-between">
            {error}
            <Button variant="outline" size="sm" onClick={handleRefresh} className="ml-4">
              <RefreshCw className="h-3 w-3 mr-1" />
              重试
            </Button>
          </AlertDescription>
        </Alert>
      )}

      {/* 快速统计 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <Card>
          <CardContent className="pt-6">
            <div className="text-2xl font-bold text-blue-600">
              {isLoading ? '...' : (stats?.dailyStats?.todayDeliveries || 0)}
            </div>
            <p className="text-gray-600">今日投递</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-2xl font-bold text-orange-600">
              {isLoading ? '...' : (stats?.dailyStats?.pendingTasks || 0)}
            </div>
            <p className="text-gray-600">待处理任务</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-2xl font-bold text-green-600">
              {isLoading ? '...' : (stats?.courierInfo?.successRate ? `${(stats.courierInfo.successRate * 100).toFixed(1)}%` : '0%')}
            </div>
            <p className="text-gray-600">投递成功率</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-2xl font-bold text-purple-600">
              {isLoading ? '...' : (stats?.teamStats?.totalMembers || 0)}
            </div>
            <p className="text-gray-600">团队成员</p>
          </CardContent>
        </Card>
      </div>
      
      {/* 详细统计卡片 */}
      {stats && !isLoading && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">总任务数</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.courierInfo.totalTasks}</div>
              <div className="text-xs text-muted-foreground">
                已完成 {stats.courierInfo.completedTasks} 个
              </div>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">平均评分</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.courierInfo.avgRating.toFixed(1)}</div>
              <div className="text-xs text-muted-foreground">
                满分 5.0 分
              </div>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">积分总数</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-amber-600">{stats.courierInfo.points}</div>
              <div className="text-xs text-muted-foreground">
                今日获得 {stats.dailyStats.todayPoints}
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 管理功能快捷入口 */}
      {courierLevel >= 2 && (
        <Card className="mb-8">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              管理中心
            </CardTitle>
            <CardDescription>
              {courierLevel === 2 && '管理片区内的一级信使和投递点'}
              {courierLevel === 3 && '管理学校内的片区和下级信使'}
              {courierLevel === 4 && '管理城市内的学校和信使网络'}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="text-center p-4 bg-blue-50 rounded-lg">
                <div className="text-2xl font-bold text-blue-600">
                  {isLoading ? '...' : (stats?.courierInfo?.totalTasks || 0)}
                </div>
                <div className="text-sm text-gray-600">管理任务</div>
              </div>
              <div className="text-center p-4 bg-green-50 rounded-lg">
                <div className="text-2xl font-bold text-green-600">
                  {isLoading ? '...' : (stats?.teamStats?.totalMembers || 0)}
                </div>
                <div className="text-sm text-gray-600">下级信使</div>
              </div>
              <div className="text-center p-4 bg-purple-50 rounded-lg">
                <div className="text-2xl font-bold text-purple-600">
                  {isLoading ? '...' : (stats?.teamStats?.totalDeliveries || 0)}
                </div>
                <div className="text-sm text-gray-600">团队配送</div>
              </div>
              <div className="text-center p-4 bg-amber-50 rounded-lg">
                <div className="text-2xl font-bold text-amber-600">
                  {isLoading ? '...' : (stats?.teamStats?.teamSuccessRate ? `${(stats.teamStats.teamSuccessRate * 100).toFixed(1)}%` : '0%')}
                </div>
                <div className="text-sm text-gray-600">团队成功率</div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 功能入口 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {features.map((feature) => {
          const Icon = feature.icon
          return (
            <Link key={feature.href} href={feature.href}>
              <Card className="hover:shadow-lg transition-shadow cursor-pointer h-full">
                <CardHeader>
                  <div className={`w-12 h-12 rounded-lg ${feature.color} flex items-center justify-center mb-4`}>
                    <Icon className="w-6 h-6 text-white" />
                  </div>
                  <CardTitle>{feature.title}</CardTitle>
                  <CardDescription>{feature.description}</CardDescription>
                </CardHeader>
              </Card>
            </Link>
          )
        })}
      </div>

      {/* 权限说明 */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            权限说明
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 text-sm text-gray-600">
            {courierLevel === 1 && (
              <>
                <p>• 🏠 管理负责楼栋的投递点和信件收发</p>
                <p>• 📦 查看和执行分配给您的投递任务</p>
                <p>• 🔍 扫描信件条码更新投递状态</p>
                <p>• 📍 查看和编辑负责区域的OP Code（后两位）</p>
                <p>• 📈 查看个人投递数据和成长进度</p>
              </>
            )}
            {courierLevel === 2 && (
              <>
                <p>• 🚛 管理片区内的投递点和一级信使</p>
                <p>• 👥 审核和分配投递任务给下级信使</p>
                <p>• 📍 管理片区OP Code编码（中间两位）</p>
                <p>• 📊 查看片区投递数据统计和分析</p>
                <p>• 🎯 审核新投递点申请和编码分配</p>
              </>
            )}
            {courierLevel === 3 && (
              <>
                <p>• 🏫 管理整个学校的信使网络和片区</p>
                <p>• 👑 创建和管理二级、一级信使账号</p>
                <p>• 📍 管理学校OP Code编码（前四位）</p>
                <p>• 📦 批量生成和管理条码系统</p>
                <p>• 🎨 设计学校专属信封和组织活动</p>
                <p>• 📈 制定学校投递策略和考核标准</p>
              </>
            )}
            {courierLevel === 4 && (
              <>
                <p>• 🌆 管理整个城市的信使网络和学校</p>
                <p>• 🏛️ 开通新学校和管理城市级配置</p>
                <p>• 📍 管理城市OP Code编码（前两位）</p>
                <p>• 👑 任命和管理三级信使（学校负责人）</p>
                <p>• 🎨 设计城市级信封和跨校活动</p>
                <p>• 📦 批量管理条码系统（跨学校操作）</p>
                <p>• 🚛 统筹城市级物流调度和优化</p>
              </>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}