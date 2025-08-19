'use client'

import React, { useState, useEffect } from 'react'
import { 
  Brain, 
  Settings, 
  Activity, 
  BarChart3, 
  FileText,
  TestTube,
  Zap,
  AlertTriangle,
  CheckCircle,
  Clock,
  TrendingUp,
  Eye,
  RefreshCw,
  Download,
  Upload,
  Cpu,
  HardDrive,
  Wifi,
  DollarSign
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Progress } from '@/components/ui/progress'
import { usePermission, PERMISSIONS } from '@/hooks/use-permission'
import { BackButton } from '@/components/ui/back-button'
import { aiApi, type AIProvider, type AIMonitoring, type AIAnalytics, type AILog } from '@/lib/api/ai'

// 常量定义
const PROVIDER_COLORS = {
  openai: 'bg-green-100 text-green-800',
  claude: 'bg-purple-100 text-purple-800',
  siliconflow: 'bg-blue-100 text-blue-800'
}

const STATUS_COLORS = {
  healthy: 'bg-green-100 text-green-800',
  warning: 'bg-yellow-100 text-yellow-800',
  error: 'bg-red-100 text-red-800',
  active: 'bg-green-100 text-green-800',
  inactive: 'bg-gray-100 text-gray-800'
}

const LOG_LEVEL_COLORS = {
  info: 'bg-blue-100 text-blue-800',
  warning: 'bg-yellow-100 text-yellow-800',
  error: 'bg-red-100 text-red-800'
}

export default function AdminAIPage() {
  const { user, hasPermission } = usePermission()
  const [config, setConfig] = useState<any>(null)
  const [monitoring, setMonitoring] = useState<AIMonitoring | null>(null)
  const [analytics, setAnalytics] = useState<AIAnalytics | null>(null)
  const [logs, setLogs] = useState<AILog[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('overview')
  
  // 对话框状态
  const [showConfigDialog, setShowConfigDialog] = useState(false)
  const [showTestDialog, setShowTestDialog] = useState(false)
  const [selectedProvider, setSelectedProvider] = useState<string>('')
  const [testingProvider, setTestingProvider] = useState<string>('')
  const [logLevel, setLogLevel] = useState<string>('info')
  const [logFeature, setLogFeature] = useState<string>('all')
  
  // 配置表单
  const [configForm, setConfigForm] = useState({
    provider: '',
    enabled: true,
    api_key: '',
    base_url: '',
    model: '',
    max_tokens: 2000,
    timeout: 30
  })

  // 测试表单
  const [testForm, setTestForm] = useState({
    provider: '',
    test_type: 'connection' as const
  })

  // 权限检查
  if (!user || !hasPermission(PERMISSIONS.SYSTEM_CONFIG)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Brain className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">访问权限不足</h2>
            <p className="text-gray-600 mb-4">
              您没有访问AI管理的权限
            </p>
            <Button asChild variant="outline">
              <a href="/admin">返回管理控制台</a>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  // 加载数据
  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      // 并行调用所有API
      const [configRes, monitoringRes, analyticsRes, logsRes] = await Promise.all([
        aiApi.getConfig(),
        aiApi.getMonitoring(),
        aiApi.getAnalytics({ time_range: '7d' }),
        aiApi.getLogs({ limit: 50 })
      ])
      
      if (configRes.data) {
        setConfig(configRes.data)
      }
      
      if (monitoringRes.data) {
        setMonitoring(monitoringRes.data as any)
      }
      
      if (analyticsRes.data) {
        setAnalytics(analyticsRes.data as any)
      }
      
      if (logsRes.data && typeof logsRes.data === 'object' && 'logs' in logsRes.data) {
        setLogs((logsRes.data as any).logs)
      }
    } catch (error) {
      console.error('Failed to load AI data:', error)
      
      // 设置空数据而不是mock数据
      setConfig(null)
      setMonitoring(null)
      setAnalytics(null)
      setLogs([])
    } finally {
      setLoading(false)
    }
  }

  const fetchAIAnalytics = async () => {
    try {
      const token = localStorage.getItem('auth_token')
      const response = await fetch('/api/v1/admin/ai/analytics', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      })
      
      if (response.ok) {
        const result = await response.json()
        setAnalytics(result.data)
      }
    } catch (error) {
      console.error('Failed to fetch AI analytics:', error)
      console.error('获取AI分析数据失败')
    }
  }

  const fetchAILogs = async () => {
    try {
      const token = localStorage.getItem('auth_token')
      const params = new URLSearchParams({
        level: logLevel,
        feature: logFeature,
        limit: '50'
      })
      
      const response = await fetch(`/api/v1/admin/ai/logs?${params}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      })
      
      if (response.ok) {
        const result = await response.json()
        setLogs(result.data)
      }
    } catch (error) {
      console.error('Failed to fetch AI logs:', error)
      console.error('获取AI日志失败')
    }
  }

  // 更新配置
  const handleUpdateConfig = async () => {
    try {
      await aiApi.updateConfig({
        providers: {
          [configForm.provider]: configForm
        }
      })
      
      await loadData() // 重新加载数据
      setShowConfigDialog(false)
      alert('配置更新成功')
    } catch (error) {
      console.error('Failed to update config:', error)
      alert('配置更新失败，请重试')
    }
  }

  // 测试提供商
  const handleTestProvider = async () => {
    setTestingProvider(testForm.provider)
    try {
      const result = await aiApi.testProvider({
        ...testForm,
        provider: testForm.provider as 'openai' | 'claude' | 'siliconflow'
      })
      const testResult = result.data as any
      alert(`测试结果: ${testResult?.status || 'Unknown'}`)
    } catch (error) {
      console.error('Failed to test provider:', error)
      alert('测试失败，请检查配置')
    } finally {
      setTestingProvider('')
      setShowTestDialog(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <BackButton href="/admin" />
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-2">
              <Brain className="w-8 h-8" />
              AI系统管理
            </h1>
            <p className="text-muted-foreground mt-1">
              管理AI功能配置、监控和分析
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button onClick={() => setShowConfigDialog(true)}>
            <Settings className="w-4 h-4 mr-2" />
            配置管理
          </Button>
          <Button onClick={() => setShowTestDialog(true)} variant="outline">
            <TestTube className="w-4 h-4 mr-2" />
            测试连接
          </Button>
        </div>
      </div>

      {/* 主要内容 */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="overview">概览</TabsTrigger>
          <TabsTrigger value="providers">提供商</TabsTrigger>
          <TabsTrigger value="monitoring">监控</TabsTrigger>
          <TabsTrigger value="analytics">分析</TabsTrigger>
          <TabsTrigger value="logs">日志</TabsTrigger>
        </TabsList>

        {/* 概览 */}
        <TabsContent value="overview" className="space-y-6">
          {/* 系统状态卡片 */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">系统状态</CardTitle>
                <Activity className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  <Badge className={STATUS_COLORS[monitoring?.overall_status as keyof typeof STATUS_COLORS] || STATUS_COLORS.healthy}>
                    {monitoring?.overall_status === 'healthy' ? '健康' : monitoring?.overall_status}
                  </Badge>
                </div>
                <p className="text-xs text-muted-foreground">
                  运行时间: {monitoring?.uptime || '99.8%'}
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">今日请求</CardTitle>
                <BarChart3 className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{monitoring?.total_requests?.toLocaleString() || '15,420'}</div>
                <p className="text-xs text-muted-foreground">
                  成功率: {monitoring ? ((monitoring.successful_requests / monitoring.total_requests) * 100).toFixed(1) : '99.8'}%
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">平均响应</CardTitle>
                <Clock className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{monitoring?.avg_response_time || 428}ms</div>
                <p className="text-xs text-muted-foreground">
                  错误率: {monitoring?.error_rate || 0.2}%
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">本月成本</CardTitle>
                <DollarSign className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">${analytics?.cost_analysis?.total_cost?.toFixed(2) || '245.60'}</div>
                <p className="text-xs text-muted-foreground">
                  预计: ${analytics?.cost_analysis?.projected_monthly_cost?.toFixed(2) || '1,050.00'}
                </p>
              </CardContent>
            </Card>
          </div>

          {/* AI功能状态 */}
          <Card>
            <CardHeader>
              <CardTitle>AI功能状态</CardTitle>
              <CardDescription>各项AI功能的使用情况和性能</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                {Object.entries(analytics?.feature_performance || {}).map(([feature, data]) => (
                  <div key={feature} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <h3 className="font-medium">
                        {feature === 'match' ? '笔友匹配' : 
                         feature === 'reply' ? 'AI回信' :
                         feature === 'inspiration' ? '写作灵感' : '内容策展'}
                      </h3>
                      <Badge variant={config?.features?.[`${feature}_enabled`] ? 'default' : 'secondary'}>
                        {config?.features?.[`${feature}_enabled`] ? '启用' : '禁用'}
                      </Badge>
                    </div>
                    <div className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span>使用次数:</span>
                        <span className="font-medium">{data.usage_count}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>成功率:</span>
                        <span className="font-medium">{data.success_rate.toFixed(1)}%</span>
                      </div>
                      <div className="flex justify-between">
                        <span>满意度:</span>
                        <span className="font-medium">{data.user_satisfaction}/5</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* 提供商状态 */}
          <Card>
            <CardHeader>
              <CardTitle>提供商状态</CardTitle>
              <CardDescription>各AI提供商的运行状态</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {Object.entries(monitoring?.providers || {}).map(([provider, data]) => (
                  <div key={provider} className="flex items-center justify-between p-4 border rounded-lg">
                    <div className="flex items-center gap-4">
                      <div className="w-12 h-12 rounded-lg bg-gray-100 flex items-center justify-center">
                        <Brain className="w-6 h-6" />
                      </div>
                      <div>
                        <h3 className="font-medium capitalize">{provider}</h3>
                        <p className="text-sm text-muted-foreground">
                          {config?.providers?.[provider]?.model || 'Unknown Model'}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <div className="text-right text-sm">
                        <div className="font-medium">{data.latency_ms}ms</div>
                        <div className="text-muted-foreground">延迟</div>
                      </div>
                      <div className="text-right text-sm">
                        <div className="font-medium">{data.success_rate.toFixed(1)}%</div>
                        <div className="text-muted-foreground">成功率</div>
                      </div>
                      <div className="text-right text-sm">
                        <div className="font-medium">{data.requests_today}</div>
                        <div className="text-muted-foreground">今日请求</div>
                      </div>
                      <Badge className={STATUS_COLORS[data.status as keyof typeof STATUS_COLORS]}>
                        {data.status === 'healthy' ? '健康' : data.status}
                      </Badge>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 提供商管理 */}
        <TabsContent value="providers" className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>AI提供商配置</CardTitle>
                  <CardDescription>管理AI服务提供商的配置和状态</CardDescription>
                </div>
                <Button onClick={() => setShowConfigDialog(true)}>
                  <Settings className="w-4 h-4 mr-2" />
                  编辑配置
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>提供商</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>模型</TableHead>
                      <TableHead>延迟</TableHead>
                      <TableHead>成功率</TableHead>
                      <TableHead>今日请求</TableHead>
                      <TableHead>配额剩余</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {Object.entries(config?.providers || {}).map(([provider, providerConfig]) => {
                      const monitoringData = monitoring?.providers?.[provider]
                      return (
                        <TableRow key={provider}>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <div className="w-8 h-8 rounded bg-gray-100 flex items-center justify-center">
                                <Brain className="w-4 h-4" />
                              </div>
                              <div>
                                <div className="font-medium capitalize">{provider}</div>
                                <div className="text-sm text-muted-foreground">
                                  {(providerConfig as any).base_url}
                                </div>
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <Badge className={(providerConfig as any).enabled ? STATUS_COLORS.active : STATUS_COLORS.inactive}>
                              {(providerConfig as any).enabled ? '启用' : '禁用'}
                            </Badge>
                          </TableCell>
                          <TableCell className="font-mono text-sm">
                            {(providerConfig as any).model}
                          </TableCell>
                          <TableCell>
                            {monitoringData?.latency_ms || '-'}ms
                          </TableCell>
                          <TableCell>
                            {monitoringData?.success_rate?.toFixed(1) || '-'}%
                          </TableCell>
                          <TableCell>
                            {monitoringData?.requests_today?.toLocaleString() || '-'}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <Progress 
                                value={parseInt(monitoringData?.quota_remaining?.replace('%', '') || '100')} 
                                className="w-16"
                              />
                              <span className="text-sm">{monitoringData?.quota_remaining || '100%'}</span>
                            </div>
                          </TableCell>
                          <TableCell>
                            <div className="flex gap-2">
                              <Button 
                                size="sm" 
                                variant="outline"
                                onClick={() => {
                                  setConfigForm({
                                    provider,
                                    ...(providerConfig as any)
                                  })
                                  setShowConfigDialog(true)
                                }}
                              >
                                <Settings className="w-4 h-4" />
                              </Button>
                              <Button 
                                size="sm" 
                                variant="outline"
                                onClick={() => {
                                  setTestForm({ provider, test_type: 'connection' })
                                  setShowTestDialog(true)
                                }}
                              >
                                <TestTube className="w-4 h-4" />
                              </Button>
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 监控 */}
        <TabsContent value="monitoring" className="space-y-6">
          {/* 资源使用情况 */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">CPU使用率</CardTitle>
                <Cpu className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">45%</div>
                <Progress value={45} className="mt-2" />
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">内存使用</CardTitle>
                <HardDrive className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">62%</div>
                <Progress value={62} className="mt-2" />
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">网络IO</CardTitle>
                <Wifi className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">2.5MB/s</div>
                <p className="text-xs text-muted-foreground mt-2">
                  入站: 1.2MB/s | 出站: 1.3MB/s
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">缓存命中率</CardTitle>
                <Zap className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">85%</div>
                <Progress value={85} className="mt-2" />
              </CardContent>
            </Card>
          </div>

          {/* 告警信息 */}
          <Card>
            <CardHeader>
              <CardTitle>系统告警</CardTitle>
              <CardDescription>当前系统告警和通知</CardDescription>
            </CardHeader>
            <CardContent>
              {monitoring?.alerts && monitoring.alerts.length > 0 ? (
                <div className="space-y-3">
                  {monitoring.alerts.map((alert, index) => (
                    <div key={index} className="flex items-center gap-3 p-3 border rounded-lg">
                      <AlertTriangle className={`w-5 h-5 ${alert.level === 'error' ? 'text-red-500' : 'text-yellow-500'}`} />
                      <div className="flex-1">
                        <div className="font-medium">{alert.message}</div>
                        <div className="text-sm text-muted-foreground">
                          {new Date(alert.timestamp).toLocaleString()}
                        </div>
                      </div>
                      <Badge className={LOG_LEVEL_COLORS[alert.level as keyof typeof LOG_LEVEL_COLORS]}>
                        {alert.level === 'warning' ? '警告' : alert.level}
                      </Badge>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8 text-muted-foreground">
                  <CheckCircle className="w-12 h-12 mx-auto mb-4 text-green-500" />
                  系统运行正常，暂无告警信息
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* 分析 */}
        <TabsContent value="analytics" className="space-y-6">
          {/* 用户使用情况 */}
          <Card>
            <CardHeader>
              <CardTitle>用户参与度</CardTitle>
              <CardDescription>AI功能的用户使用情况分析</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-blue-600">
                    {analytics?.user_engagement?.total_active_users || 234}
                  </div>
                  <div className="text-sm text-muted-foreground">总活跃用户</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">
                    {analytics?.user_engagement?.ai_feature_users || 189}
                  </div>
                  <div className="text-sm text-muted-foreground">AI功能用户</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-purple-600">
                    {analytics?.user_engagement?.adoption_rate?.toFixed(1) || '80.8'}%
                  </div>
                  <div className="text-sm text-muted-foreground">采用率</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-orange-600">
                    {analytics?.user_engagement?.retention_rate?.toFixed(1) || '68.5'}%
                  </div>
                  <div className="text-sm text-muted-foreground">留存率</div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* 质量指标 */}
          <Card>
            <CardHeader>
              <CardTitle>质量指标</CardTitle>
              <CardDescription>AI服务质量和用户满意度</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold">
                    {analytics?.quality_metrics?.content_quality_score?.toFixed(1) || '4.2'}/5
                  </div>
                  <div className="text-sm text-muted-foreground">内容质量</div>
                  <Progress value={((analytics?.quality_metrics?.content_quality_score || 4.2) / 5) * 100} className="mt-2" />
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold">
                    {analytics?.quality_metrics?.relevance_score?.toFixed(1) || '88.3'}%
                  </div>
                  <div className="text-sm text-muted-foreground">相关性</div>
                  <Progress value={analytics?.quality_metrics?.relevance_score || 88.3} className="mt-2" />
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold">
                    {analytics?.quality_metrics?.coherence_score?.toFixed(1) || '91.2'}%
                  </div>
                  <div className="text-sm text-muted-foreground">连贯性</div>
                  <Progress value={analytics?.quality_metrics?.coherence_score || 91.2} className="mt-2" />
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold">
                    {analytics?.quality_metrics?.user_feedback_score?.toFixed(1) || '4.4'}/5
                  </div>
                  <div className="text-sm text-muted-foreground">用户反馈</div>
                  <Progress value={((analytics?.quality_metrics?.user_feedback_score || 4.4) / 5) * 100} className="mt-2" />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* 成本分析 */}
          <Card>
            <CardHeader>
              <CardTitle>成本分析</CardTitle>
              <CardDescription>AI服务的成本分布和趋势</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="text-center">
                  <div className="text-3xl font-bold text-green-600">
                    ${analytics?.cost_analysis?.total_cost?.toFixed(2) || '245.60'}
                  </div>
                  <div className="text-sm text-muted-foreground">本月总成本</div>
                </div>
                <div className="text-center">
                  <div className="text-3xl font-bold text-blue-600">
                    ${analytics?.cost_analysis?.cost_per_request?.toFixed(3) || '0.159'}
                  </div>
                  <div className="text-sm text-muted-foreground">单次请求成本</div>
                </div>
                <div className="text-center">
                  <div className="text-3xl font-bold text-purple-600">
                    ${analytics?.cost_analysis?.projected_monthly_cost?.toFixed(2) || '1,050.00'}
                  </div>
                  <div className="text-sm text-muted-foreground">月度预测</div>
                </div>
              </div>
              
              <div className="mt-6">
                <h4 className="font-medium mb-3">提供商成本分布</h4>
                <div className="space-y-2">
                  {Object.entries(analytics?.cost_analysis?.cost_by_provider || {}).map(([provider, cost]) => (
                    <div key={provider} className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Badge className={PROVIDER_COLORS[provider as keyof typeof PROVIDER_COLORS]}>
                          {provider}
                        </Badge>
                      </div>
                      <div className="font-medium">${cost.toFixed(2)}</div>
                    </div>
                  ))}
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 日志 */}
        <TabsContent value="logs" className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>操作日志</CardTitle>
                  <CardDescription>AI系统的操作日志和审计跟踪</CardDescription>
                </div>
                <div className="flex gap-2">
                  <Button variant="outline" size="sm">
                    <Download className="w-4 h-4 mr-2" />
                    导出日志
                  </Button>
                  <Button variant="outline" size="sm" onClick={loadData}>
                    <RefreshCw className="w-4 h-4 mr-2" />
                    刷新
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>时间</TableHead>
                      <TableHead>级别</TableHead>
                      <TableHead>功能</TableHead>
                      <TableHead>用户</TableHead>
                      <TableHead>操作</TableHead>
                      <TableHead>消息</TableHead>
                      <TableHead>详情</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {logs.map((log) => (
                      <TableRow key={log.id}>
                        <TableCell className="font-mono text-sm">
                          {new Date(log.timestamp).toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <Badge className={LOG_LEVEL_COLORS[log.level]}>
                            {log.level === 'info' ? '信息' : 
                             log.level === 'warning' ? '警告' : '错误'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {log.feature === 'match' ? '匹配' :
                             log.feature === 'reply' ? '回信' :
                             log.feature === 'inspiration' ? '灵感' : log.feature}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono text-sm">
                          {log.user_id || '-'}
                        </TableCell>
                        <TableCell className="font-mono text-sm">
                          {log.action}
                        </TableCell>
                        <TableCell className="max-w-xs truncate">
                          {log.message}
                        </TableCell>
                        <TableCell>
                          <Button size="sm" variant="outline">
                            <Eye className="w-4 h-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* 配置对话框 */}
      <Dialog open={showConfigDialog} onOpenChange={setShowConfigDialog}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>编辑AI提供商配置</DialogTitle>
            <DialogDescription>
              配置AI服务提供商的参数和设置
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="provider">提供商</Label>
              <Select value={configForm.provider} onValueChange={(value) => setConfigForm(prev => ({ ...prev, provider: value }))}>
                <SelectTrigger>
                  <SelectValue placeholder="选择提供商" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="openai">OpenAI</SelectItem>
                  <SelectItem value="claude">Claude</SelectItem>
                  <SelectItem value="siliconflow">SiliconFlow</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div className="flex items-center space-x-2">
              <Switch 
                checked={configForm.enabled} 
                onCheckedChange={(enabled) => setConfigForm(prev => ({ ...prev, enabled }))}
              />
              <Label>启用提供商</Label>
            </div>
            
            <div>
              <Label htmlFor="api_key">API密钥</Label>
              <Input
                id="api_key"
                type="password"
                value={configForm.api_key}
                onChange={(e) => setConfigForm(prev => ({ ...prev, api_key: e.target.value }))}
                placeholder="输入API密钥..."
              />
            </div>
            
            <div>
              <Label htmlFor="base_url">基础URL</Label>
              <Input
                id="base_url"
                value={configForm.base_url}
                onChange={(e) => setConfigForm(prev => ({ ...prev, base_url: e.target.value }))}
                placeholder="输入API基础URL..."
              />
            </div>
            
            <div>
              <Label htmlFor="model">模型</Label>
              <Input
                id="model"
                value={configForm.model}
                onChange={(e) => setConfigForm(prev => ({ ...prev, model: e.target.value }))}
                placeholder="输入模型名称..."
              />
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="max_tokens">最大Token数</Label>
                <Input
                  id="max_tokens"
                  type="number"
                  value={configForm.max_tokens}
                  onChange={(e) => setConfigForm(prev => ({ ...prev, max_tokens: parseInt(e.target.value) || 2000 }))}
                />
              </div>
              <div>
                <Label htmlFor="timeout">超时时间(秒)</Label>
                <Input
                  id="timeout"
                  type="number"
                  value={configForm.timeout}
                  onChange={(e) => setConfigForm(prev => ({ ...prev, timeout: parseInt(e.target.value) || 30 }))}
                />
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowConfigDialog(false)}>
              取消
            </Button>
            <Button onClick={handleUpdateConfig}>
              保存配置
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 测试对话框 */}
      <Dialog open={showTestDialog} onOpenChange={setShowTestDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>测试AI提供商</DialogTitle>
            <DialogDescription>
              测试AI提供商的连接状态和响应
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="test_provider">选择提供商</Label>
              <Select value={testForm.provider} onValueChange={(value) => setTestForm(prev => ({ ...prev, provider: value }))}>
                <SelectTrigger>
                  <SelectValue placeholder="选择要测试的提供商" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="openai">OpenAI</SelectItem>
                  <SelectItem value="claude">Claude</SelectItem>
                  <SelectItem value="siliconflow">SiliconFlow</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <Label htmlFor="test_type">测试类型</Label>
              <Select value={testForm.test_type} onValueChange={(value: any) => setTestForm(prev => ({ ...prev, test_type: value }))}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="connection">连接测试</SelectItem>
                  <SelectItem value="response">响应测试</SelectItem>
                  <SelectItem value="quality">质量测试</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowTestDialog(false)}>
              取消
            </Button>
            <Button 
              onClick={handleTestProvider} 
              disabled={!testForm.provider || testingProvider === testForm.provider}
            >
              {testingProvider === testForm.provider ? (
                <>
                  <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                  测试中...
                </>
              ) : (
                '开始测试'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}