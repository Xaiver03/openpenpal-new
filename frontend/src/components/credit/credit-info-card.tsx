'use client'

import React, { useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'
import { Coins, TrendingUp, Star, RefreshCw } from 'lucide-react'
import { useCreditInfo, useCreditStore } from '@/stores/credit-store'
import { formatPoints, getCreditLevelName, getLevelProgress, getPointsToNextLevel } from '@/lib/api/credit'

interface CreditInfoCardProps {
  showActions?: boolean
  compact?: boolean
  className?: string
}

export function CreditInfoCard({ 
  showActions = true, 
  compact = false, 
  className = '' 
}: CreditInfoCardProps) {
  const { userCredit, creditSummary, loading, error } = useCreditInfo()
  const { fetchUserCredit, fetchCreditSummary, clearError } = useCreditStore()

  useEffect(() => {
    if (!userCredit) {
      fetchUserCredit()
    }
    if (!creditSummary) {
      fetchCreditSummary()
    }
  }, [userCredit, creditSummary, fetchUserCredit, fetchCreditSummary])

  const handleRefresh = async () => {
    await Promise.all([fetchUserCredit(), fetchCreditSummary()])
  }

  if (loading) {
    return (
      <Card className={`w-full ${className}`}>
        <CardHeader>
          <Skeleton className="h-6 w-32" />
        </CardHeader>
        <CardContent className="space-y-4">
          <Skeleton className="h-8 w-24" />
          <Skeleton className="h-4 w-full" />
          <div className="grid grid-cols-2 gap-4">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className={`w-full ${className}`}>
        <CardContent className="p-6">
          <div className="text-center space-y-4">
            <p className="text-destructive text-sm">{error}</p>
            <Button variant="outline" size="sm" onClick={() => {
              clearError()
              handleRefresh()
            }}>
              <RefreshCw className="h-4 w-4 mr-2" />
              重试
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (!userCredit || !creditSummary) {
    return (
      <Card className={`w-full ${className}`}>
        <CardContent className="p-6">
          <div className="text-center text-muted-foreground">
            <p>暂无积分信息</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  const levelProgress = getLevelProgress(userCredit.total, userCredit.level)
  const pointsToNext = getPointsToNextLevel(userCredit.total, userCredit.level)
  const levelName = getCreditLevelName(userCredit.level)

  return (
    <Card className={`w-full ${className}`}>
      <CardHeader className={compact ? 'pb-3' : ''}>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Coins className="h-5 w-5 text-amber-500" />
            我的积分
          </CardTitle>
          {showActions && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleRefresh}
              disabled={loading}
            >
              <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            </Button>
          )}
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* 总积分和等级 */}
        <div className="flex items-center justify-between">
          <div>
            <div className="text-2xl font-bold text-amber-600">
              {formatPoints(userCredit.total)}
            </div>
            <div className="text-sm text-muted-foreground">总积分</div>
          </div>
          <Badge variant="secondary" className="text-sm">
            <Star className="h-3 w-3 mr-1" />
            {levelName}
          </Badge>
        </div>

        {/* 等级进度条 */}
        {!compact && pointsToNext > 0 && (
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>距离下一等级</span>
              <span className="font-medium">{pointsToNext} 积分</span>
            </div>
            <Progress value={levelProgress} className="h-2" />
          </div>
        )}

        {/* 积分详情 */}
        <div className={`grid ${compact ? 'grid-cols-2' : 'grid-cols-3'} gap-4`}>
          <div className="text-center p-3 bg-green-50 rounded-lg dark:bg-green-950">
            <div className="text-lg font-semibold text-green-600">
              {formatPoints(userCredit.available)}
            </div>
            <div className="text-xs text-muted-foreground">可用积分</div>
          </div>
          
          <div className="text-center p-3 bg-blue-50 rounded-lg dark:bg-blue-950">
            <div className="text-lg font-semibold text-blue-600">
              {formatPoints(creditSummary.today_earned)}
            </div>
            <div className="text-xs text-muted-foreground">今日获得</div>
          </div>
          
          {!compact && (
            <div className="text-center p-3 bg-purple-50 rounded-lg dark:bg-purple-950">
              <div className="text-lg font-semibold text-purple-600">
                {creditSummary.pending_tasks}
              </div>
              <div className="text-xs text-muted-foreground">待处理任务</div>
            </div>
          )}
        </div>

        {/* 快速统计 */}
        {!compact && (
          <div className="pt-2 border-t">
            <div className="flex items-center justify-between text-sm text-muted-foreground">
              <span>本周获得: {formatPoints(creditSummary.week_earned)}</span>
              <span>今日任务: {creditSummary.completed_tasks_today}</span>
            </div>
          </div>
        )}

        {/* 趋势指示器 */}
        {showActions && creditSummary.today_earned > 0 && (
          <div className="flex items-center gap-2 text-sm text-green-600">
            <TrendingUp className="h-4 w-4" />
            <span>今日积分增长了 {creditSummary.today_earned} 分</span>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default CreditInfoCard