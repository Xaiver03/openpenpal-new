'use client'

import React, { useEffect } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { RefreshCw, TrendingUp, Award, History, BarChart3 } from 'lucide-react'
import { CreditInfoCard } from './credit-info-card'
import { CreditHistoryList } from './credit-history-list'
import { CreditTaskList } from './credit-task-list'
import { CreditStatistics } from './credit-statistics'
import { CreditLeaderboard } from './credit-leaderboard'
import { CreditProgressBar } from './credit-progress-bar'
import { useCreditStore, useCreditInfo } from '@/stores/credit-store'

interface CreditManagementPageProps {
  className?: string
}

export function CreditManagementPage({ className = '' }: CreditManagementPageProps) {
  const { userCredit, creditSummary } = useCreditInfo()
  const { refreshAll, loading } = useCreditStore()

  useEffect(() => {
    // 初始化加载数据
    refreshAll()
  }, [refreshAll])

  const handleRefreshAll = async () => {
    await refreshAll()
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* 页面标题和操作 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">积分管理</h1>
          <p className="text-muted-foreground mt-1">
            查看您的积分状态、历史记录和任务进度
          </p>
        </div>
        <Button
          onClick={handleRefreshAll}
          disabled={loading.credit || loading.summary}
          variant="outline"
        >
          <RefreshCw className={`h-4 w-4 mr-2 ${loading.credit || loading.summary ? 'animate-spin' : ''}`} />
          刷新数据
        </Button>
      </div>

      {/* 积分概览卡片 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <CreditInfoCard showActions={false} />
        </div>
        
        <div className="space-y-4">
          {userCredit && (
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">等级进度</CardTitle>
              </CardHeader>
              <CardContent>
                <CreditProgressBar
                  currentLevel={userCredit.level}
                  totalPoints={userCredit.total}
                  showLabels={false}
                  showNextLevel={true}
                />
              </CardContent>
            </Card>
          )}
          
          {/* 快速统计 */}
          {creditSummary && (
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">本周概览</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex justify-between text-sm">
                  <span>本周获得</span>
                  <span className="font-medium text-green-600">
                    +{creditSummary.week_earned}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span>本月获得</span>
                  <span className="font-medium text-blue-600">
                    +{creditSummary.month_earned}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span>待处理任务</span>
                  <span className="font-medium text-orange-600">
                    {creditSummary.pending_tasks}
                  </span>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>

      {/* 主要内容区域 */}
      <Tabs defaultValue="overview" className="space-y-6">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview" className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4" />
            概览
          </TabsTrigger>
          <TabsTrigger value="tasks" className="flex items-center gap-2">
            <Award className="h-4 w-4" />
            任务
          </TabsTrigger>
          <TabsTrigger value="history" className="flex items-center gap-2">
            <History className="h-4 w-4" />
            历史
          </TabsTrigger>
          <TabsTrigger value="statistics" className="flex items-center gap-2">
            <BarChart3 className="h-4 w-4" />
            统计
          </TabsTrigger>
        </TabsList>

        {/* 概览页面 */}
        <TabsContent value="overview" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* 最近任务 */}
            <CreditTaskList
              showFilters={false}
              pageSize={5}
            />
            
            {/* 最近历史 */}
            <CreditHistoryList
              showFilters={false}
              pageSize={5}
            />
          </div>
          
          {/* 排行榜 */}
          <CreditLeaderboard />
        </TabsContent>

        {/* 任务页面 */}
        <TabsContent value="tasks" className="space-y-6">
          <CreditTaskList />
        </TabsContent>

        {/* 历史页面 */}
        <TabsContent value="history" className="space-y-6">
          <CreditHistoryList />
        </TabsContent>

        {/* 统计页面 */}
        <TabsContent value="statistics" className="space-y-6">
          <CreditStatistics />
        </TabsContent>
      </Tabs>
    </div>
  )
}

export default CreditManagementPage