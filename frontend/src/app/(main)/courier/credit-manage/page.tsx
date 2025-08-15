'use client'

import React from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  Award, 
  Users, 
  TrendingUp, 
  Settings,
  Crown,
  Shield,
  BarChart3,
  UserCheck,
  Gift,
  AlertCircle
} from 'lucide-react'
import Link from 'next/link'
import { useCourierPermission } from '@/hooks/use-courier-permission'
import { CourierPermissionGuard } from '@/components/courier/CourierPermissionGuard'
import { CreditLeaderboard } from '@/components/credit/credit-leaderboard'
import { CreditStatistics } from '@/components/credit/credit-statistics'

export default function CourierCreditManagePage() {
  const { courierInfo, getCourierLevelName } = useCourierPermission()
  
  if (!courierInfo || courierInfo.level < 2) {
    return (
      <div className="min-h-screen bg-amber-50">
        <div className="container max-w-6xl mx-auto px-4 py-8">
          <Card className="border-amber-200">
            <CardContent className="p-12 text-center">
              <AlertCircle className="h-12 w-12 text-amber-600 mx-auto mb-4" />
              <h2 className="text-2xl font-bold text-amber-900 mb-2">权限不足</h2>
              <p className="text-amber-700 mb-6">
                只有2级及以上信使才能访问积分管理功能
              </p>
              <Link href="/courier">
                <Button className="bg-amber-600 hover:bg-amber-700 text-white">
                  返回信使中心
                </Button>
              </Link>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-6xl mx-auto px-4 py-8">
        {/* 页面标题 */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <Crown className="h-8 w-8 text-amber-600" />
            <h1 className="text-3xl font-bold text-amber-900">积分管理中心</h1>
          </div>
          <p className="text-amber-700">
            {getCourierLevelName()}专属积分管理功能
          </p>
        </div>

        {/* 管理功能卡片 */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          {/* L2+ 可见：团队积分概览 */}
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-amber-900">
                <Users className="h-5 w-5" />
                团队积分概览
              </CardTitle>
              <CardDescription>查看下级信使积分情况</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="flex justify-between items-center">
                  <span className="text-sm text-amber-700">团队总积分</span>
                  <span className="font-bold text-amber-900">12,850</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-amber-700">本月新增</span>
                  <span className="font-bold text-green-600">+2,340</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-amber-700">活跃信使</span>
                  <span className="font-bold text-blue-600">15</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* L3+ 可见：积分审批 */}
          {courierInfo.level >= 3 && (
            <Card className="border-amber-200">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-amber-900">
                  <UserCheck className="h-5 w-5" />
                  积分审批
                </CardTitle>
                <CardDescription>审核特殊积分申请</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <Badge variant="outline" className="w-full justify-center py-2 bg-orange-50 text-orange-700 border-orange-300">
                    待审批: 3
                  </Badge>
                  <Button className="w-full bg-amber-600 hover:bg-amber-700 text-white">
                    查看申请
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}

          {/* L4 可见：积分政策 */}
          {courierInfo.level >= 4 && (
            <Card className="border-amber-200">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-amber-900">
                  <Shield className="h-5 w-5" />
                  积分政策管理
                </CardTitle>
                <CardDescription>调整积分规则和奖励</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <Button variant="outline" className="w-full justify-start border-amber-300 text-amber-700 hover:bg-amber-50">
                    <Settings className="h-4 w-4 mr-2" />
                    规则设置
                  </Button>
                  <Button variant="outline" className="w-full justify-start border-amber-300 text-amber-700 hover:bg-amber-50">
                    <Gift className="h-4 w-4 mr-2" />
                    奖励配置
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        {/* 主要内容区域 */}
        <Tabs defaultValue="overview" className="space-y-6">
          <TabsList className="bg-amber-100">
            <TabsTrigger value="overview" className="data-[state=active]:bg-amber-200">
              总览
            </TabsTrigger>
            <TabsTrigger value="leaderboard" className="data-[state=active]:bg-amber-200">
              排行榜
            </TabsTrigger>
            <TabsTrigger value="statistics" className="data-[state=active]:bg-amber-200">
              统计分析
            </TabsTrigger>
            {courierInfo.level >= 3 && (
              <TabsTrigger value="management" className="data-[state=active]:bg-amber-200">
                高级管理
              </TabsTrigger>
            )}
          </TabsList>

          {/* 总览标签页 */}
          <TabsContent value="overview">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* 团队积分排行 */}
              <Card className="border-amber-200">
                <CardHeader>
                  <CardTitle className="text-amber-900">团队积分排行</CardTitle>
                  <CardDescription>您管理区域内的信使积分表现</CardDescription>
                </CardHeader>
                <CardContent>
                  <CreditLeaderboard
                    limit={5}
                    showTimeFilter={false}
                    showCurrentUser={false}
                    className="border-0 shadow-none"
                  />
                </CardContent>
              </Card>

              {/* 积分趋势 */}
              <Card className="border-amber-200">
                <CardHeader>
                  <CardTitle className="text-amber-900">积分趋势</CardTitle>
                  <CardDescription>团队积分增长情况</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="h-64 flex items-center justify-center text-amber-600">
                    <BarChart3 className="h-12 w-12 opacity-50" />
                    <span className="ml-2">图表开发中...</span>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* 排行榜标签页 */}
          <TabsContent value="leaderboard">
            <CreditLeaderboard
              limit={20}
              showTimeFilter={true}
              className="border-amber-200"
            />
          </TabsContent>

          {/* 统计分析标签页 */}
          <TabsContent value="statistics">
            <CreditStatistics />
          </TabsContent>

          {/* 高级管理标签页 (L3+) */}
          {courierInfo.level >= 3 && (
            <TabsContent value="management">
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* 积分调整 */}
                <Card className="border-amber-200">
                  <CardHeader>
                    <CardTitle className="text-amber-900">积分调整</CardTitle>
                    <CardDescription>手动调整信使积分</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <p className="text-sm text-amber-700">
                        可以为特殊贡献的信使手动增加积分，或因违规扣除积分
                      </p>
                      <Button className="w-full bg-amber-600 hover:bg-amber-700 text-white">
                        <Award className="h-4 w-4 mr-2" />
                        积分调整
                      </Button>
                    </div>
                  </CardContent>
                </Card>

                {/* 活动管理 (L4) */}
                {courierInfo.level >= 4 && (
                  <Card className="border-amber-200">
                    <CardHeader>
                      <CardTitle className="text-amber-900">积分活动</CardTitle>
                      <CardDescription>创建和管理积分活动</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
                          <p className="text-sm font-medium text-green-900">当前活动</p>
                          <p className="text-xs text-green-700 mt-1">
                            双倍积分周末 (进行中)
                          </p>
                        </div>
                        <Button className="w-full bg-amber-600 hover:bg-amber-700 text-white">
                          <TrendingUp className="h-4 w-4 mr-2" />
                          创建新活动
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                )}
              </div>
            </TabsContent>
          )}
        </Tabs>
      </div>
    </div>
  )
}