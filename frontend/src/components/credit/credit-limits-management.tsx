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
import { Switch } from '@/components/ui/switch'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  Shield, 
  ShieldAlert, 
  ShieldCheck,
  Eye,
  AlertTriangle,
  Settings,
  Download,
  Upload,
  Plus,
  Edit,
  Trash2,
  Search,
  Filter,
  RefreshCw,
  BarChart3,
  Activity,
  Users,
  Clock,
  Ban,
  CheckCircle,
  XCircle,
  AlertCircle,
  TrendingUp,
  TrendingDown,
  Zap
} from 'lucide-react'
import {
  getLimitRules,
  createLimitRule,
  updateLimitRule,
  deleteLimitRule,
  batchCreateRules,
  batchUpdateRules,
  exportRules,
  importRules,
  getRiskUsers,
  blockUser,
  unblockUser,
  getUserRiskAnalysis,
  getDashboardStats,
  getRealTimeAlerts,
  getSystemHealth,
  type CreditLimitRule,
  type CreditRiskUser,
  type DashboardStats
} from '@/lib/api/credit-limits'

// ==================== 主要组件 ====================

export function CreditLimitsManagement() {
  const [activeTab, setActiveTab] = useState('dashboard')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 状态管理
  const [dashboardStats, setDashboardStats] = useState<DashboardStats | null>(null)
  const [rules, setRules] = useState<CreditLimitRule[]>([])
  const [riskUsers, setRiskUsers] = useState<CreditRiskUser[]>([])
  const [alerts, setAlerts] = useState<any[]>([])
  const [systemHealth, setSystemHealth] = useState<any>(null)

  // 刷新数据
  const refreshData = async () => {
    setLoading(true)
    try {
      const [statsRes, rulesRes, riskUsersRes, alertsRes, healthRes] = await Promise.all([
        getDashboardStats('7d'),
        getLimitRules({ limit: 50 }),
        getRiskUsers({ limit: 20 }),
        getRealTimeAlerts({ limit: 10 }),
        getSystemHealth()
      ])

      setDashboardStats(statsRes.stats)
      setRules(rulesRes.rules)
      setRiskUsers(riskUsersRes.users)
      setAlerts(alertsRes.alerts)
      setSystemHealth(healthRes.health)
    } catch (err) {
      setError('加载数据失败')
      console.error('数据加载错误:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refreshData()
  }, [])

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Shield className="h-6 w-6 text-blue-600" />
            积分限制与防作弊管理
          </h2>
          <p className="text-gray-600 mt-1">管理积分限制规则、监控风险用户、防作弊检测</p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={refreshData}
            disabled={loading}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
          <SystemHealthIndicator health={systemHealth} />
        </div>
      </div>

      {error && (
        <Alert className="border-red-200 bg-red-50">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 主要内容标签页 */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
        <TabsList className="grid grid-cols-6 w-full max-w-4xl">
          <TabsTrigger value="dashboard">仪表板</TabsTrigger>
          <TabsTrigger value="rules">限制规则</TabsTrigger>
          <TabsTrigger value="fraud">防作弊</TabsTrigger>
          <TabsTrigger value="users">风险用户</TabsTrigger>
          <TabsTrigger value="monitoring">实时监控</TabsTrigger>
          <TabsTrigger value="reports">报表分析</TabsTrigger>
        </TabsList>

        {/* 仪表板 */}
        <TabsContent value="dashboard">
          <DashboardOverview stats={dashboardStats} />
        </TabsContent>

        {/* 限制规则管理 */}
        <TabsContent value="rules">
          <LimitRulesManagement
            rules={rules}
            onRulesChange={setRules}
            onError={setError}
          />
        </TabsContent>

        {/* 防作弊检测 */}
        <TabsContent value="fraud">
          <FraudDetectionPanel />
        </TabsContent>

        {/* 风险用户管理 */}
        <TabsContent value="users">
          <RiskUsersManagement
            users={riskUsers}
            onUsersChange={setRiskUsers}
            onError={setError}
          />
        </TabsContent>

        {/* 实时监控 */}
        <TabsContent value="monitoring">
          <RealTimeMonitoring
            alerts={alerts}
            health={systemHealth}
            onAlertsChange={setAlerts}
          />
        </TabsContent>

        {/* 报表分析 */}
        <TabsContent value="reports">
          <ReportsAnalysis />
        </TabsContent>
      </Tabs>
    </div>
  )
}

// ==================== 仪表板概览 ====================

function DashboardOverview({ stats }: { stats: DashboardStats | null }) {
  if (!stats) {
    return <div className="flex items-center justify-center h-64">加载中...</div>
  }

  return (
    <div className="space-y-6">
      {/* 基础统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="活动规则"
          value={stats.basic.active_rules}
          total={stats.basic.total_rules}
          icon={<Settings className="h-5 w-5" />}
          trend="stable"
          color="blue"
        />
        <StatCard
          title="风险用户"
          value={stats.basic.blocked_users}
          total={stats.basic.total_risk_users}
          icon={<ShieldAlert className="h-5 w-5" />}
          trend="down"
          color="red"
        />
        <StatCard
          title="今日检测"
          value={stats.detection.anomalous_count}
          total={stats.detection.total_detections}
          icon={<Eye className="h-5 w-5" />}
          trend="up"
          color="orange"
        />
        <StatCard
          title="高风险检测"
          value={stats.detection.high_risk_count}
          total={stats.detection.total_detections}
          icon={<AlertTriangle className="h-5 w-5" />}
          trend="stable"
          color="purple"
        />
      </div>

      {/* 趋势图表 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>每日活动趋势</CardTitle>
            <CardDescription>最近7天的用户行为和检测趋势</CardDescription>
          </CardHeader>
          <CardContent>
            <DailyTrendsChart trends={stats.daily_trends} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>风险等级分布</CardTitle>
            <CardDescription>用户风险等级统计</CardDescription>
          </CardHeader>
          <CardContent>
            <RiskDistributionChart distribution={stats.risk_distribution} />
          </CardContent>
        </Card>
      </div>

      {/* 热门行为类型 */}
      <Card>
        <CardHeader>
          <CardTitle>热门行为类型</CardTitle>
          <CardDescription>最活跃的积分行为统计</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {stats.actions.slice(0, 5).map((action, index) => (
              <div key={action.action_type} className="flex items-center justify-between p-3 border rounded-lg">
                <div className="flex items-center gap-3">
                  <Badge variant="outline" className="w-6 h-6 rounded-full p-0 flex items-center justify-center">
                    {index + 1}
                  </Badge>
                  <div>
                    <div className="font-medium">{action.action_type}</div>
                    <div className="text-sm text-gray-500">总积分: {action.total_points}</div>
                  </div>
                </div>
                <div className="text-right">
                  <div className="font-medium">{action.count}次</div>
                  <div className="text-sm text-gray-500">操作数</div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

// ==================== 统计卡片组件 ====================

function StatCard({
  title,
  value,
  total,
  icon,
  trend,
  color
}: {
  title: string
  value: number
  total?: number
  icon: React.ReactNode
  trend: 'up' | 'down' | 'stable'
  color: 'blue' | 'red' | 'orange' | 'purple' | 'green'
}) {
  const percentage = total ? Math.round((value / total) * 100) : 0

  const colorClasses = {
    blue: 'text-blue-600 bg-blue-50 border-blue-200',
    red: 'text-red-600 bg-red-50 border-red-200',
    orange: 'text-orange-600 bg-orange-50 border-orange-200',
    purple: 'text-purple-600 bg-purple-50 border-purple-200',
    green: 'text-green-600 bg-green-50 border-green-200'
  }

  const TrendIcon = trend === 'up' ? TrendingUp : trend === 'down' ? TrendingDown : Activity

  return (
    <Card className={colorClasses[color]}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium">{title}</CardTitle>
          {icon}
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div className="text-2xl font-bold">
            {value.toLocaleString()}
            {total && (
              <span className="text-base font-normal text-gray-600 ml-2">
                / {total.toLocaleString()}
              </span>
            )}
          </div>
          {total && (
            <div className="flex items-center gap-2 text-sm">
              <TrendIcon className="h-3 w-3" />
              <span>{percentage}% 比例</span>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

// ==================== 趋势图表组件 ====================

function DailyTrendsChart({ trends }: { trends: any[] }) {
  return (
    <div className="space-y-4">
      {trends.map((trend, index) => (
        <div key={trend.date} className="flex items-center justify-between p-3 border rounded">
          <div className="font-medium">{trend.date}</div>
          <div className="flex items-center gap-4 text-sm">
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 rounded-full bg-blue-500"></div>
              <span>行为: {trend.action_count}</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 rounded-full bg-orange-500"></div>
              <span>检测: {trend.detection_count}</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-3 h-3 rounded-full bg-red-500"></div>
              <span>异常: {trend.anomalous_count}</span>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}

// ==================== 风险分布图表 ====================

function RiskDistributionChart({ distribution }: { distribution: any[] }) {
  const total = distribution.reduce((sum, item) => sum + item.count, 0)
  
  const colors = {
    low: 'bg-green-500',
    medium: 'bg-yellow-500',
    high: 'bg-orange-500',
    blocked: 'bg-red-500'
  }

  return (
    <div className="space-y-4">
      {distribution.map((item) => (
        <div key={item.risk_level} className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="capitalize">{item.risk_level}</span>
            <span>{item.count} 用户</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className={`h-2 rounded-full ${colors[item.risk_level as keyof typeof colors]}`}
              style={{ width: `${(item.count / total) * 100}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  )
}

// ==================== 系统健康指示器 ====================

function SystemHealthIndicator({ health }: { health: any }) {
  if (!health) return null

  const statusColors = {
    healthy: 'bg-green-500',
    warning: 'bg-yellow-500',
    critical: 'bg-red-500'
  }

  const statusColor = statusColors[health.overall_status as keyof typeof statusColors] || 'bg-gray-500'

  return (
    <div className="flex items-center gap-2">
      <div className={`w-3 h-3 rounded-full ${statusColor}`}></div>
      <span className="text-sm text-gray-600">系统状态</span>
    </div>
  )
}

// ==================== 其他子组件占位符 ====================

function LimitRulesManagement({
  rules,
  onRulesChange,
  onError
}: {
  rules: CreditLimitRule[]
  onRulesChange: (rules: CreditLimitRule[]) => void
  onError: (error: string) => void
}) {
  return (
    <div className="text-center p-8 text-gray-500">
      限制规则管理界面
      <br />
      <small>待实现详细功能</small>
    </div>
  )
}

function FraudDetectionPanel() {
  return (
    <div className="text-center p-8 text-gray-500">
      防作弊检测面板
      <br />
      <small>待实现详细功能</small>
    </div>
  )
}

function RiskUsersManagement({
  users,
  onUsersChange,
  onError
}: {
  users: CreditRiskUser[]
  onUsersChange: (users: CreditRiskUser[]) => void
  onError: (error: string) => void
}) {
  return (
    <div className="text-center p-8 text-gray-500">
      风险用户管理界面
      <br />
      <small>待实现详细功能</small>
    </div>
  )
}

function RealTimeMonitoring({
  alerts,
  health,
  onAlertsChange
}: {
  alerts: any[]
  health: any
  onAlertsChange: (alerts: any[]) => void
}) {
  return (
    <div className="text-center p-8 text-gray-500">
      实时监控界面
      <br />
      <small>待实现详细功能</small>
    </div>
  )
}

function ReportsAnalysis() {
  return (
    <div className="text-center p-8 text-gray-500">
      报表分析界面
      <br />
      <small>待实现详细功能</small>
    </div>
  )
}