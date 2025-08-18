'use client'

import { CreditInfoCard } from '@/components/credit/credit-info-card'
import { CreditStatistics } from '@/components/credit/credit-statistics'
import { CreditLeaderboard } from '@/components/credit/credit-leaderboard'
import { CreditHistoryList } from '@/components/credit/credit-history-list'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

export default function CreditsPage() {
  return (
    <div className="container mx-auto py-8 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">积分中心</h1>
      </div>

      <CreditInfoCard className="mb-6" />

      <Tabs defaultValue="statistics" className="space-y-4">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="statistics">统计分析</TabsTrigger>
          <TabsTrigger value="leaderboard">排行榜</TabsTrigger>
          <TabsTrigger value="history">历史记录</TabsTrigger>
        </TabsList>

        <TabsContent value="statistics" className="space-y-4">
          <CreditStatistics />
        </TabsContent>

        <TabsContent value="leaderboard" className="space-y-4">
          <CreditLeaderboard limit={20} />
        </TabsContent>

        <TabsContent value="history" className="space-y-4">
          <CreditHistoryList />
        </TabsContent>
      </Tabs>
    </div>
  )
}