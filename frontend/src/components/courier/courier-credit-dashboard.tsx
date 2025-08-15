'use client'

import React, { useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Trophy,
  Award,
  Package,
  TrendingUp,
  Star,
  Target,
  Users,
  Calendar,
  Zap,
  Truck,
  CheckCircle2,
  Clock,
  ArrowRight
} from 'lucide-react'
import Link from 'next/link'
import { useCourierPermission } from '@/hooks/use-courier-permission'
import { useCreditStore, useCreditInfo } from '@/stores/credit-store'
import { CreditInfoCard } from '@/components/credit/credit-info-card'
import { CreditHistoryList } from '@/components/credit/credit-history-list'
import { CreditTaskList } from '@/components/credit/credit-task-list'
import { CreditLeaderboard } from '@/components/credit/credit-leaderboard'
import { CreditProgressBar } from '@/components/credit/credit-progress-bar'
import { formatPoints } from '@/lib/api/credit'

interface CourierCreditDashboardProps {
  className?: string
}

export function CourierCreditDashboard({ className = '' }: CourierCreditDashboardProps) {
  const { courierInfo, getCourierLevelName } = useCourierPermission()
  const { userCredit, creditSummary } = useCreditInfo()
  const { refreshAll, loading } = useCreditStore()

  useEffect(() => {
    refreshAll()
  }, [refreshAll])

  // 信使特定的积分统计
  const courierStats = [
    {
      title: '今日送达',
      value: creditSummary?.today_deliveries || 0,
      icon: Package,
      color: 'text-green-600',
      description: '成功投递'
    },
    {
      title: '本周积分',
      value: formatPoints(creditSummary?.week_earned || 0),
      icon: TrendingUp,
      color: 'text-blue-600',
      description: '累计获得'
    },
    {
      title: '首次送达',
      value: creditSummary?.first_deliveries || 0,
      icon: Star,
      color: 'text-yellow-600',
      description: '新用户首投'
    },
    {
      title: '准时率',
      value: `${creditSummary?.on_time_rate || 98}%`,
      icon: Clock,
      color: 'text-purple-600',
      description: '按时送达'
    }
  ]

  // 信使专属任务类型
  const courierTaskTypes = [
    'courier_delivery',
    'courier_first_task',
    'courier_weekly_bonus',
    'courier_monthly_achievement'
  ]

  return (
    <div className={`space-y-6 ${className}`}>
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Trophy className="h-8 w-8 text-amber-600" />
          <div>
            <h1 className="text-3xl font-bold text-amber-900">信使积分中心</h1>
            <p className="text-amber-700">追踪您的积分成长和晋升进度</p>
          </div>
        </div>
        
        <Button
          onClick={() => refreshAll()}
          disabled={loading.credit || loading.summary}
          variant="outline"
          className="border-amber-300 text-amber-700 hover:bg-amber-50"
        >
          刷新数据
        </Button>
      </div>

      {/* 信使身份卡片 */}
      <Card className="border-amber-200 bg-gradient-to-r from-amber-50 to-orange-50">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 bg-amber-600 text-white rounded-full flex items-center justify-center">
                <Truck className="h-6 w-6" />
              </div>
              <div>
                <CardTitle className="flex items-center gap-2">
                  <span>{getCourierLevelName()}</span>
                  <Badge variant="secondary" className="bg-amber-600 text-white">
                    L{courierInfo?.level || 1}
                  </Badge>
                </CardTitle>
                <p className="text-sm text-amber-700">
                  服务区域: {courierInfo?.zoneCode || '未设置'}
                </p>
              </div>
            </div>
            
            {userCredit && (
              <div className="text-right">
                <div className="text-3xl font-bold text-amber-900">
                  {formatPoints(userCredit.total)}
                </div>
                <div className="text-sm text-amber-600">总积分</div>
              </div>
            )}
          </div>
        </CardHeader>
      </Card>

      {/* 信使统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {courierStats.map((stat, index) => (
          <Card key={index} className="border-amber-200 hover:border-amber-300 transition-colors">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-amber-600 mb-1">{stat.description}</p>
                  <p className="text-2xl font-bold text-amber-900">{stat.value}</p>
                  <p className="text-sm font-medium text-amber-700">{stat.title}</p>
                </div>
                <stat.icon className={`h-8 w-8 ${stat.color}`} />
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* 等级进度 */}
      {userCredit && (
        <Card className="border-amber-200">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-amber-900">
              <Award className="h-5 w-5" />
              积分等级进度
            </CardTitle>
          </CardHeader>
          <CardContent>
            <CreditProgressBar
              currentLevel={userCredit.level}
              totalPoints={userCredit.total}
              showLabels={true}
              showNextLevel={true}
              animated={true}
            />
            
            {/* 信使晋升提示 */}
            {courierInfo && userCredit.level >= (courierInfo.level + 1) * 2 && (
              <div className="mt-4 p-4 bg-green-50 border border-green-200 rounded-lg">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-green-900">
                      🎉 您的积分等级已满足信使晋升要求！
                    </p>
                    <p className="text-sm text-green-700 mt-1">
                      可以申请晋升到 {courierInfo.level + 1} 级信使
                    </p>
                  </div>
                  <Link href="/courier/growth">
                    <Button size="sm" className="bg-green-600 hover:bg-green-700">
                      申请晋升
                      <ArrowRight className="ml-1 h-4 w-4" />
                    </Button>
                  </Link>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* 主要内容标签页 */}
      <Tabs defaultValue="tasks" className="space-y-6">
        <TabsList className="bg-amber-100 grid w-full grid-cols-4">
          <TabsTrigger value="tasks" className="data-[state=active]:bg-amber-200">
            <Package className="h-4 w-4 mr-2" />
            任务奖励
          </TabsTrigger>
          <TabsTrigger value="history" className="data-[state=active]:bg-amber-200">
            <Calendar className="h-4 w-4 mr-2" />
            积分历史
          </TabsTrigger>
          <TabsTrigger value="ranking" className="data-[state=active]:bg-amber-200">
            <Trophy className="h-4 w-4 mr-2" />
            信使排行
          </TabsTrigger>
          <TabsTrigger value="achievements" className="data-[state=active]:bg-amber-200">
            <Award className="h-4 w-4 mr-2" />
            成就徽章
          </TabsTrigger>
        </TabsList>

        {/* 任务奖励 */}
        <TabsContent value="tasks" className="space-y-6">
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">信使专属任务</CardTitle>
              <p className="text-sm text-amber-700">
                完成投递任务获得积分奖励
              </p>
            </CardHeader>
            <CardContent>
              <CreditTaskList
                showFilters={true}
                pageSize={10}
                className="border-0 shadow-none"
              />
            </CardContent>
          </Card>
        </TabsContent>

        {/* 积分历史 */}
        <TabsContent value="history" className="space-y-6">
          <CreditHistoryList
            showFilters={true}
            pageSize={20}
            className="border-amber-200"
          />
        </TabsContent>

        {/* 信使排行 */}
        <TabsContent value="ranking" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* 总排行榜 */}
            <CreditLeaderboard
              limit={10}
              showTimeFilter={true}
              className="border-amber-200"
            />
            
            {/* 信使专属排行 */}
            <Card className="border-amber-200">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-amber-900">
                  <Truck className="h-5 w-5" />
                  信使投递榜
                </CardTitle>
                <p className="text-sm text-amber-700">本月投递数量排行</p>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {/* TODO: 实现信使专属排行榜 */}
                  <div className="text-center py-8 text-amber-600">
                    <Users className="h-12 w-12 mx-auto mb-2 opacity-50" />
                    <p>信使投递排行榜开发中...</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* 成就徽章 */}
        <TabsContent value="achievements" className="space-y-6">
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="text-amber-900">成就徽章</CardTitle>
              <p className="text-sm text-amber-700">
                解锁成就获得专属徽章和额外积分
              </p>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {/* 成就徽章示例 */}
                <div className="text-center p-4 border border-amber-200 rounded-lg">
                  <div className="w-16 h-16 bg-yellow-100 rounded-full flex items-center justify-center mx-auto mb-2">
                    <Star className="h-8 w-8 text-yellow-600" />
                  </div>
                  <p className="font-medium text-amber-900">首次投递</p>
                  <p className="text-xs text-amber-600">完成第一次投递</p>
                  <Badge className="mt-2 bg-green-100 text-green-700">已获得</Badge>
                </div>
                
                <div className="text-center p-4 border border-amber-200 rounded-lg opacity-50">
                  <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-2">
                    <Target className="h-8 w-8 text-gray-400" />
                  </div>
                  <p className="font-medium text-gray-600">百投达人</p>
                  <p className="text-xs text-gray-500">完成100次投递</p>
                  <p className="text-xs text-amber-600 mt-2">进度: 45/100</p>
                </div>
                
                <div className="text-center p-4 border border-amber-200 rounded-lg opacity-50">
                  <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-2">
                    <Zap className="h-8 w-8 text-gray-400" />
                  </div>
                  <p className="font-medium text-gray-600">极速信使</p>
                  <p className="text-xs text-gray-500">连续7天准时送达</p>
                  <p className="text-xs text-amber-600 mt-2">进度: 3/7</p>
                </div>
                
                <div className="text-center p-4 border border-amber-200 rounded-lg opacity-50">
                  <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-2">
                    <Trophy className="h-8 w-8 text-gray-400" />
                  </div>
                  <p className="font-medium text-gray-600">金牌信使</p>
                  <p className="text-xs text-gray-500">月度排行榜第一</p>
                  <Badge className="mt-2 bg-gray-100 text-gray-600">未解锁</Badge>
                </div>
              </div>
              
              <div className="mt-6 p-4 bg-amber-50 rounded-lg">
                <p className="text-sm text-amber-700 text-center">
                  更多成就徽章即将推出，敬请期待！
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* 快捷操作 */}
      <Card className="border-amber-200">
        <CardHeader>
          <CardTitle className="text-amber-900">快捷操作</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <Link href="/courier/tasks">
              <Button variant="outline" className="w-full justify-start border-amber-200 text-amber-700 hover:bg-amber-50">
                <CheckCircle2 className="mr-2 h-4 w-4" />
                查看任务
              </Button>
            </Link>
            <Link href="/courier/scan">
              <Button variant="outline" className="w-full justify-start border-amber-200 text-amber-700 hover:bg-amber-50">
                <Package className="mr-2 h-4 w-4" />
                扫码投递
              </Button>
            </Link>
            <Link href="/courier/growth">
              <Button variant="outline" className="w-full justify-start border-amber-200 text-amber-700 hover:bg-amber-50">
                <TrendingUp className="mr-2 h-4 w-4" />
                晋升路径
              </Button>
            </Link>
            <Link href="/credit">
              <Button variant="outline" className="w-full justify-start border-amber-200 text-amber-700 hover:bg-amber-50">
                <Award className="mr-2 h-4 w-4" />
                积分详情
              </Button>
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default CourierCreditDashboard