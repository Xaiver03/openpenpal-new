'use client'

import React, { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { 
  BarChart3, 
  TrendingUp, 
  Calendar, 
  RefreshCw,
  Award,
  Users,
  Target,
  Activity
} from 'lucide-react'
import { useCreditStore, useCreditStatistics } from '@/stores/credit-store'
import { formatPoints } from '@/lib/api/credit'

interface CreditStatisticsProps {
  className?: string
}

export function CreditStatistics({ className = '' }: CreditStatisticsProps) {
  const { statistics, loading, error } = useCreditStatistics()
  const { fetchStatistics, clearError } = useCreditStore()
  
  const [timeRange, setTimeRange] = useState<'week' | 'month' | 'year'>('month')

  useEffect(() => {
    fetchStatistics(timeRange)
  }, [timeRange, fetchStatistics])

  const handleRefresh = () => {
    clearError()
    fetchStatistics(timeRange)
  }

  const handleTimeRangeChange = (range: 'week' | 'month' | 'year') => {
    setTimeRange(range)
  }

  if (loading && !statistics) {
    return (
      <div className={`space-y-6 ${className}`}>
        {/* 统计卡片骨架 */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i}>
              <CardHeader className="pb-2">
                <Skeleton className="h-4 w-20" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-16" />
                <Skeleton className="h-3 w-24 mt-2" />
              </CardContent>
            </Card>
          ))}
        </div>
        
        {/* 图表骨架 */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <Skeleton className="h-6 w-32" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-64 w-full" />
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <Skeleton className="h-6 w-32" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-64 w-full" />
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  const getTimeRangeLabel = (range: string) => {
    switch (range) {
      case 'week': return '本周'
      case 'month': return '本月'
      case 'year': return '本年'
      default: return '本月'
    }
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* 页面标题和控件 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <BarChart3 className="h-6 w-6" />
          <h2 className="text-2xl font-bold">积分统计</h2>
        </div>
        
        <div className="flex items-center gap-2">
          {/* 时间范围选择 */}
          <div className="flex bg-muted rounded-lg p-1">
            {(['week', 'month', 'year'] as const).map((range) => (
              <Button
                key={range}
                variant={timeRange === range ? 'default' : 'ghost'}
                size="sm"
                onClick={() => handleTimeRangeChange(range)}
                className="text-xs"
              >
                {getTimeRangeLabel(range)}
              </Button>
            ))}
          </div>
          
          <Button
            onClick={handleRefresh}
            disabled={loading}
            variant="outline"
            size="sm"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </div>

      {/* 错误状态 */}
      {error && (
        <div className="text-center p-4 text-destructive text-sm bg-destructive/10 rounded-lg">
          {error}
        </div>
      )}

      {/* 统计概览卡片 */}
      {statistics && (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* 总获得积分 */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  {getTimeRangeLabel(timeRange)}获得
                </CardTitle>
                <TrendingUp className="h-4 w-4 text-green-600" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">
                  +{formatPoints(statistics.total_earned || 0)}
                </div>
                <p className="text-xs text-muted-foreground">
                  较上期 {statistics.earn_growth >= 0 ? '+' : ''}{statistics.earn_growth?.toFixed(1) || 0}%
                </p>
              </CardContent>
            </Card>

            {/* 任务完成数 */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">完成任务</CardTitle>
                <Award className="h-4 w-4 text-blue-600" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-blue-600">
                  {statistics.tasks_completed || 0}
                </div>
                <p className="text-xs text-muted-foreground">
                  成功率 {((statistics.tasks_completed || 0) / Math.max(1, (statistics.tasks_total || 1)) * 100).toFixed(1)}%
                </p>
              </CardContent>
            </Card>

            {/* 平均每日积分 */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">日均积分</CardTitle>
                <Calendar className="h-4 w-4 text-orange-600" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-orange-600">
                  {formatPoints(statistics.daily_average || 0)}
                </div>
                <p className="text-xs text-muted-foreground">
                  最高单日 {formatPoints(statistics.max_daily || 0)}
                </p>
              </CardContent>
            </Card>

            {/* 当前排名 */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">当前排名</CardTitle>
                <Users className="h-4 w-4 text-purple-600" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-purple-600">
                  #{statistics.current_rank || 'N/A'}
                </div>
                <p className="text-xs text-muted-foreground">
                  总用户 {statistics.total_users || 0} 人
                </p>
              </CardContent>
            </Card>
          </div>

          {/* 图表区域 */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* 积分趋势图 */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">积分趋势</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {/* 简化的趋势展示 */}
                  {statistics.daily_breakdown && statistics.daily_breakdown.length > 0 ? (
                    <div className="space-y-2">
                      {statistics.daily_breakdown.slice(-7).map((day, index) => (
                        <div key={index} className="flex items-center justify-between text-sm">
                          <span className="text-muted-foreground">
                            {new Date(day.date).toLocaleDateString('zh-CN', { 
                              month: 'short', 
                              day: 'numeric' 
                            })}
                          </span>
                          <div className="flex items-center gap-2">
                            <div 
                              className="bg-blue-500 rounded-sm"
                              style={{ 
                                width: `${Math.max(4, (day.points / Math.max(...statistics.daily_breakdown.map(d => d.points))) * 60)}px`,
                                height: '8px'
                              }}
                            />
                            <span className="font-medium w-12 text-right">
                              +{formatPoints(day.points)}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-center py-8 text-muted-foreground">
                      暂无数据
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* 任务类型分布 */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">任务类型分布</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {statistics.task_type_breakdown && statistics.task_type_breakdown.length > 0 ? (
                    statistics.task_type_breakdown.map((item, index) => {
                      const percentage = ((item.count / statistics.tasks_total!) * 100).toFixed(1)
                      return (
                        <div key={index} className="space-y-2">
                          <div className="flex justify-between text-sm">
                            <span>{item.task_type}</span>
                            <span className="font-medium">{item.count} ({percentage}%)</span>
                          </div>
                          <div className="bg-muted rounded-full h-2">
                            <div 
                              className="bg-blue-500 rounded-full h-2 transition-all duration-500"
                              style={{ width: `${percentage}%` }}
                            />
                          </div>
                        </div>
                      )
                    })
                  ) : (
                    <div className="text-center py-8 text-muted-foreground">
                      暂无数据
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* 详细统计表格 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">详细统计</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <div className="space-y-3">
                  <h4 className="font-medium flex items-center gap-2">
                    <Activity className="h-4 w-4" />
                    任务统计
                  </h4>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span>总任务数</span>
                      <span className="font-medium">{statistics.tasks_total || 0}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>已完成</span>
                      <span className="font-medium text-green-600">{statistics.tasks_completed || 0}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>执行中</span>
                      <span className="font-medium text-blue-600">{statistics.tasks_executing || 0}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>失败</span>
                      <span className="font-medium text-red-600">{statistics.tasks_failed || 0}</span>
                    </div>
                  </div>
                </div>

                <div className="space-y-3">
                  <h4 className="font-medium flex items-center gap-2">
                    <Target className="h-4 w-4" />
                    积分明细
                  </h4>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span>历史总计</span>
                      <span className="font-medium">{formatPoints(statistics.total_earned || 0)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>本月获得</span>
                      <span className="font-medium text-green-600">+{formatPoints(statistics.month_earned || 0)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>本周获得</span>
                      <span className="font-medium text-blue-600">+{formatPoints(statistics.week_earned || 0)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>今日获得</span>
                      <span className="font-medium text-orange-600">+{formatPoints(statistics.today_earned || 0)}</span>
                    </div>
                  </div>
                </div>

                <div className="space-y-3">
                  <h4 className="font-medium flex items-center gap-2">
                    <TrendingUp className="h-4 w-4" />
                    性能指标
                  </h4>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span>任务成功率</span>
                      <span className="font-medium">
                        {((statistics.tasks_completed || 0) / Math.max(1, statistics.tasks_total || 1) * 100).toFixed(1)}%
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span>平均响应时间</span>
                      <span className="font-medium">{statistics.avg_response_time || 'N/A'}</span>
                    </div>
                    <div className="flex justify-between">
                      <span>活跃天数</span>
                      <span className="font-medium">{statistics.active_days || 0} 天</span>
                    </div>
                    <div className="flex justify-between">
                      <span>连续签到</span>
                      <span className="font-medium">{statistics.streak_days || 0} 天</span>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}

export default CreditStatistics