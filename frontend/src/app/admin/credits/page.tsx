'use client'

import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { 
  Award, 
  Users, 
  TrendingUp, 
  Settings,
  Shield,
  BarChart3,
  Gift,
  AlertCircle,
  Plus,
  Search,
  Edit,
  Trash2,
  CheckCircle,
  XCircle,
  Clock,
  DollarSign
} from 'lucide-react'
import { Breadcrumb, ADMIN_BREADCRUMBS } from '@/components/ui/breadcrumb'
import { WelcomeBanner } from '@/components/ui/welcome-banner'
import { CreditStatistics } from '@/components/credit/credit-statistics'
import { CreditLeaderboard } from '@/components/credit/credit-leaderboard'
import { CreditManagementPage } from '@/components/credit/credit-management-page'
import { CreditLimitsManagement } from '@/components/credit/credit-limits-management'

export default function AdminCreditsPage() {
  const [activeTab, setActiveTab] = useState('overview')
  const [searchUser, setSearchUser] = useState('')
  const [adjustAmount, setAdjustAmount] = useState('')
  const [adjustReason, setAdjustReason] = useState('')
  
  // 状态数据
  const [creditSummary, setCreditSummary] = useState({
    totalPoints: 0,
    activeUsers: 0,
    todayEarned: 0,
    pendingTasks: 0,
    weekGrowth: 0,
    monthGrowth: 0
  })

  const [creditRules, setCreditRules] = useState<Array<{
    id: number
    action: string
    points: number
    enabled: boolean
  }>>([])

  const [recentAdjustments, setRecentAdjustments] = useState<Array<{
    id: number
    user: string
    amount: number
    type: string
    reason: string
    operator: string
    time: string
  }>>([])

  // 加载数据
  useEffect(() => {
    const loadData = async () => {
      try {
        // TODO: 调用真实API获取数据
        // const response = await apiClient.get('/admin/credits/summary')
        // setCreditSummary(response.data)
        
        // 暂时保持空数据
      } catch (error) {
        console.error('Failed to load credit data:', error)
      }
    }
    
    loadData()
  }, [])

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <WelcomeBanner />
        
        <Breadcrumb 
          items={[
            ...ADMIN_BREADCRUMBS.dashboard,
            { label: '积分管理', href: '/admin/credits' }
          ]} 
        />
        
        {/* 页面标题 */}
        <div className="mb-8">
          <div className="flex items-center gap-3">
            <Award className="h-8 w-8 text-purple-600" />
            <div>
              <h1 className="text-3xl font-bold text-gray-900">积分管理中心</h1>
              <p className="text-gray-600 mt-1">管理平台积分系统、规则和用户积分</p>
            </div>
          </div>
        </div>

        {/* 统计卡片 */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-gray-600">总积分</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-gray-900">
                {creditSummary.totalPoints.toLocaleString()}
              </div>
              <div className="flex items-center gap-1 text-sm text-green-600 mt-1">
                <TrendingUp className="h-3 w-3" />
                <span>月增长 {creditSummary.monthGrowth}%</span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-gray-600">活跃用户</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-gray-900">
                {creditSummary.activeUsers.toLocaleString()}
              </div>
              <div className="text-sm text-gray-500 mt-1">
                有积分记录的用户
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-gray-600">今日产生</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-gray-900">
                +{creditSummary.todayEarned.toLocaleString()}
              </div>
              <div className="flex items-center gap-1 text-sm text-blue-600 mt-1">
                <TrendingUp className="h-3 w-3" />
                <span>周增长 {creditSummary.weekGrowth}%</span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-gray-600">待处理任务</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-gray-900">
                {creditSummary.pendingTasks}
              </div>
              <Badge variant="outline" className="mt-1 text-orange-600 border-orange-300">
                需要审核
              </Badge>
            </CardContent>
          </Card>
        </div>

        {/* 主要内容 */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="grid grid-cols-6 w-full max-w-4xl">
            <TabsTrigger value="overview">总览</TabsTrigger>
            <TabsTrigger value="rules">积分规则</TabsTrigger>
            <TabsTrigger value="adjustments">积分调整</TabsTrigger>
            <TabsTrigger value="activities">积分活动</TabsTrigger>
            <TabsTrigger value="limits">限制管理</TabsTrigger>
            <TabsTrigger value="analytics">数据分析</TabsTrigger>
          </TabsList>

          {/* 总览 */}
          <TabsContent value="overview">
            <CreditManagementPage className="shadow-none" />
          </TabsContent>

          {/* 积分规则 */}
          <TabsContent value="rules" className="space-y-6">
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>积分规则配置</CardTitle>
                    <CardDescription>设置各种行为的积分奖励值</CardDescription>
                  </div>
                  <Button className="bg-purple-600 hover:bg-purple-700">
                    <Plus className="h-4 w-4 mr-2" />
                    添加规则
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {creditRules.map((rule) => (
                    <div key={rule.id} className="flex items-center justify-between p-4 border rounded-lg">
                      <div className="flex-1">
                        <div className="font-medium">{rule.action}</div>
                        <div className="text-sm text-gray-500">
                          奖励积分: {rule.points}
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <Badge variant={rule.enabled ? 'default' : 'secondary'}>
                          {rule.enabled ? '已启用' : '已禁用'}
                        </Badge>
                        <div className="flex gap-2">
                          <Button variant="ghost" size="sm">
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="sm" className="text-red-600">
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* 积分调整 */}
          <TabsContent value="adjustments" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>手动调整积分</CardTitle>
                <CardDescription>为用户增加或扣除积分</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-4">
                    <div>
                      <Label>搜索用户</Label>
                      <div className="flex gap-2 mt-2">
                        <Input 
                          placeholder="输入用户名或ID" 
                          value={searchUser}
                          onChange={(e) => setSearchUser(e.target.value)}
                        />
                        <Button variant="outline">
                          <Search className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                    
                    <div>
                      <Label>调整类型</Label>
                      <Select defaultValue="add">
                        <SelectTrigger className="mt-2">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="add">增加积分</SelectItem>
                          <SelectItem value="deduct">扣除积分</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    
                    <div>
                      <Label>调整数量</Label>
                      <Input 
                        type="number" 
                        placeholder="输入积分数量" 
                        className="mt-2"
                        value={adjustAmount}
                        onChange={(e) => setAdjustAmount(e.target.value)}
                      />
                    </div>
                    
                    <div>
                      <Label>调整原因</Label>
                      <Textarea 
                        placeholder="请输入调整原因（必填）" 
                        className="mt-2"
                        rows={4}
                        value={adjustReason}
                        onChange={(e) => setAdjustReason(e.target.value)}
                      />
                    </div>
                    
                    <Button className="w-full bg-purple-600 hover:bg-purple-700">
                      确认调整
                    </Button>
                  </div>
                  
                  <div>
                    <h4 className="font-medium mb-4">最近调整记录</h4>
                    <div className="space-y-3">
                      {recentAdjustments.map((record) => (
                        <div key={record.id} className="p-3 border rounded-lg">
                          <div className="flex items-center justify-between mb-2">
                            <span className="font-medium">{record.user}</span>
                            <Badge variant={record.type === 'add' ? 'default' : 'destructive'}>
                              {record.type === 'add' ? '+' : '-'}{record.amount}
                            </Badge>
                          </div>
                          <div className="text-sm text-gray-600">
                            <div>原因: {record.reason}</div>
                            <div className="flex justify-between mt-1">
                              <span>操作人: {record.operator}</span>
                              <span>{record.time}</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* 积分活动 */}
          <TabsContent value="activities" className="space-y-6">
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>积分活动管理</CardTitle>
                    <CardDescription>创建和管理限时积分活动</CardDescription>
                  </div>
                  <Button className="bg-purple-600 hover:bg-purple-700">
                    <Plus className="h-4 w-4 mr-2" />
                    创建活动
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {/* 进行中的活动 */}
                  <Card className="border-green-200 bg-green-50">
                    <CardHeader>
                      <CardTitle className="text-lg flex items-center gap-2">
                        <Gift className="h-5 w-5 text-green-600" />
                        双倍积分周末
                      </CardTitle>
                      <Badge className="w-fit bg-green-600">进行中</Badge>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2 text-sm">
                        <div>开始时间: 2024-01-20 00:00</div>
                        <div>结束时间: 2024-01-21 23:59</div>
                        <div>活动内容: 所有积分奖励翻倍</div>
                        <div className="pt-2">
                          <Button variant="outline" size="sm" className="mr-2">编辑</Button>
                          <Button variant="outline" size="sm" className="text-red-600">结束活动</Button>
                        </div>
                      </div>
                    </CardContent>
                  </Card>

                  {/* 计划中的活动 */}
                  <Card className="border-blue-200 bg-blue-50">
                    <CardHeader>
                      <CardTitle className="text-lg flex items-center gap-2">
                        <Clock className="h-5 w-5 text-blue-600" />
                        新春积分狂欢
                      </CardTitle>
                      <Badge className="w-fit bg-blue-600">计划中</Badge>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2 text-sm">
                        <div>开始时间: 2024-02-10 00:00</div>
                        <div>结束时间: 2024-02-17 23:59</div>
                        <div>活动内容: 写信积分×3，首投积分×5</div>
                        <div className="pt-2">
                          <Button variant="outline" size="sm" className="mr-2">编辑</Button>
                          <Button variant="outline" size="sm" className="text-red-600">取消</Button>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* 限制管理 */}
          <TabsContent value="limits">
            <CreditLimitsManagement />
          </TabsContent>

          {/* 数据分析 */}
          <TabsContent value="analytics">
            <CreditStatistics />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}