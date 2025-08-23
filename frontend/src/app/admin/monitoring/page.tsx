'use client'

import React, { useState, useEffect, useCallback } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Progress } from '@/components/ui/progress'
import { 
  Activity, 
  AlertTriangle, 
  CheckCircle, 
  Clock, 
  Database, 
  HardDrive, 
  Monitor,
  RefreshCw,
  Server,
  Shield,
  TrendingUp,
  Users,
  Wifi,
  WifiOff,
  XCircle,
  Zap,
  BarChart3,
  LineChart,
  AlertCircle,
  ArrowUp,
  ArrowDown
} from 'lucide-react'
import { BackButton } from '@/components/ui/back-button'
import { Breadcrumb, ADMIN_BREADCRUMBS } from '@/components/ui/breadcrumb'
import { usePermission, PERMISSIONS } from '@/hooks/use-permission'
import dynamic from 'next/dynamic'

// Lazy load monitoring components to avoid circular dependencies
const RealtimeMetricsPanel = dynamic(
  () => import('@/components/monitoring/real-time-metrics').then(mod => ({ default: mod.RealtimeMetricsPanel })),
  { ssr: false }
)
const ServiceStatusDashboard = dynamic(
  () => import('@/components/monitoring/service-status').then(mod => ({ default: mod.ServiceStatusDashboard })),
  { ssr: false }
)
const ErrorLogViewer = dynamic(
  () => import('@/components/monitoring/error-log-viewer').then(mod => ({ default: mod.ErrorLogViewer })),
  { ssr: false }
)

interface SystemHealth {
  status: 'healthy' | 'degraded' | 'critical'
  uptime: number
  services: ServiceStatus[]
  resources: ResourceUsage
  metrics: PerformanceMetrics
}

interface ServiceStatus {
  name: string
  status: 'online' | 'offline' | 'degraded'
  uptime: number
  lastCheck: string
  responseTime: number
}

interface ResourceUsage {
  cpu: number
  memory: number
  disk: number
  network: {
    in: number
    out: number
  }
}

interface PerformanceMetrics {
  avgResponseTime: number
  errorRate: number
  requestsPerMinute: number
  activeUsers: number
  cacheHitRate: number
}

export default function MonitoringDashboard() {
  const { user, hasPermission } = usePermission()
  const [timeRange, setTimeRange] = useState('1h')
  const [refreshInterval, setRefreshInterval] = useState(30000)
  const [isAutoRefresh, setIsAutoRefresh] = useState(true)
  const [lastRefresh, setLastRefresh] = useState(new Date())
  
  // System health state
  const [systemHealth, setSystemHealth] = useState<SystemHealth | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Performance metrics
  const [performanceReport, setPerformanceReport] = useState<any>(null)
  
  // Security events
  const [securityStats, setSecurityStats] = useState<any>(null)
  const [recentSecurityEvents, setRecentSecurityEvents] = useState<any[]>([])

  // Load system health data
  const loadSystemHealth = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)

      // Simulate system health check (replace with real API)
      const health: SystemHealth = {
        status: 'healthy',
        uptime: Date.now() - (7 * 24 * 60 * 60 * 1000), // 7 days
        services: [
          { name: '主服务器', status: 'online', uptime: 99.9, lastCheck: new Date().toISOString(), responseTime: 45 },
          { name: 'Write服务', status: 'online', uptime: 99.8, lastCheck: new Date().toISOString(), responseTime: 120 },
          { name: 'Courier服务', status: 'online', uptime: 99.7, lastCheck: new Date().toISOString(), responseTime: 89 },
          { name: 'Admin服务', status: 'online', uptime: 99.9, lastCheck: new Date().toISOString(), responseTime: 67 },
          { name: '数据库', status: 'online', uptime: 100, lastCheck: new Date().toISOString(), responseTime: 12 },
          { name: 'Redis缓存', status: 'online', uptime: 99.99, lastCheck: new Date().toISOString(), responseTime: 3 }
        ],
        resources: {
          cpu: 45,
          memory: 67,
          disk: 23,
          network: {
            in: 1240,
            out: 3456
          }
        },
        metrics: {
          avgResponseTime: 87,
          errorRate: 0.02,
          requestsPerMinute: 450,
          activeUsers: 1234,
          cacheHitRate: 92.5
        }
      }

      setSystemHealth(health)

      // Load performance metrics (simplified mock)
      setPerformanceReport({
        summary: {
          first_input_delay: 12,
          cumulative_layout_shift: 0.05,
          largest_contentful_paint: 1250,
          time_to_first_byte: 120,
          cache_hit_rate: 85.5,
          memory_usage_mb: 45
        }
      })

      // Load security stats (mock)
      setSecurityStats({
        totalEvents: 127,
        eventsBySeverity: {
          critical: 2,
          high: 8,
          medium: 45,
          low: 72
        },
        topIpAddresses: [
          { ip: '192.168.1.100', count: 15 },
          { ip: '10.0.0.1', count: 12 }
        ]
      })

      // Load recent security events (mock)
      setRecentSecurityEvents([
        {
          id: '1',
          type: 'login_failed',
          severity: 'medium',
          timestamp: new Date().toISOString(),
          ipAddress: '192.168.1.100'
        },
        {
          id: '2', 
          type: 'rate_limit_exceeded',
          severity: 'high',
          timestamp: new Date(Date.now() - 300000).toISOString(),
          ipAddress: '10.0.0.1'
        }
      ])

      setLastRefresh(new Date())
    } catch (err) {
      console.error('Failed to load monitoring data:', err)
      setError('加载监控数据失败')
    } finally {
      setIsLoading(false)
    }
  }, [])

  // Auto refresh
  useEffect(() => {
    loadSystemHealth()

    if (isAutoRefresh) {
      const interval = setInterval(loadSystemHealth, refreshInterval)
      return () => clearInterval(interval)
    }
  }, [isAutoRefresh, refreshInterval, loadSystemHealth])

  // Permission check
  if (!user || !hasPermission(PERMISSIONS.SYSTEM_CONFIG)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Shield className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">访问权限不足</h2>
            <p className="text-gray-600 mb-4">
              您没有访问系统监控的权限
            </p>
            <Button asChild variant="outline">
              <a href="/admin">返回管理控制台</a>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'online':
        return 'text-green-600'
      case 'degraded':
        return 'text-yellow-600'
      case 'critical':
      case 'offline':
        return 'text-red-600'
      default:
        return 'text-gray-600'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'online':
        return <CheckCircle className="w-5 h-5 text-green-600" />
      case 'degraded':
        return <AlertTriangle className="w-5 h-5 text-yellow-600" />
      case 'critical':
      case 'offline':
        return <XCircle className="w-5 h-5 text-red-600" />
      default:
        return <AlertCircle className="w-5 h-5 text-gray-600" />
    }
  }

  const formatUptime = (startTime: number) => {
    const uptime = Date.now() - startTime
    const days = Math.floor(uptime / (24 * 60 * 60 * 1000))
    const hours = Math.floor((uptime % (24 * 60 * 60 * 1000)) / (60 * 60 * 1000))
    return `${days}天 ${hours}小时`
  }

  const getSeverityBadge = (severity: string) => {
    const variants: Record<string, 'secondary' | 'default' | 'destructive'> = {
      low: 'secondary',
      medium: 'default',
      high: 'destructive',
      critical: 'destructive'
    }

    return (
      <Badge variant={variants[severity] || 'default'}>
        {severity.toUpperCase()}
      </Badge>
    )
  }

  if (isLoading && !systemHealth) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <Breadcrumb items={[...ADMIN_BREADCRUMBS.dashboard, { label: '系统监控', href: '/admin/monitoring' }]} />

      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <BackButton href="/admin" />
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-2">
              <Monitor className="w-8 h-8" />
              系统监控
            </h1>
            <p className="text-muted-foreground mt-1">
              实时监控系统健康状态和性能指标
            </p>
          </div>
        </div>
        
        <div className="flex items-center gap-4">
          <Select value={timeRange} onValueChange={setTimeRange}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="1h">最近1小时</SelectItem>
              <SelectItem value="6h">最近6小时</SelectItem>
              <SelectItem value="24h">最近24小时</SelectItem>
              <SelectItem value="7d">最近7天</SelectItem>
            </SelectContent>
          </Select>
          
          <Button
            variant="outline"
            size="sm"
            onClick={loadSystemHealth}
            disabled={isLoading}
          >
            <RefreshCw className={`w-4 h-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
          
          <div className="text-sm text-muted-foreground">
            最后更新: {lastRefresh.toLocaleTimeString()}
          </div>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 系统状态概览 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">系统状态</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {getStatusIcon(systemHealth?.status || 'unknown')}
                <span className={`text-2xl font-bold ${getStatusColor(systemHealth?.status || 'unknown')}`}>
                  {systemHealth?.status === 'healthy' ? '正常' : 
                   systemHealth?.status === 'degraded' ? '降级' : '异常'}
                </span>
              </div>
              <Badge variant="outline">
                运行 {formatUptime(systemHealth?.uptime || Date.now())}
              </Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">活跃用户</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Users className="w-5 h-5 text-blue-600" />
                <span className="text-2xl font-bold">{systemHealth?.metrics.activeUsers || 0}</span>
              </div>
              <TrendingUp className="w-4 h-4 text-green-600" />
            </div>
            <p className="text-xs text-muted-foreground mt-1">在线用户数</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">平均响应时间</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Zap className="w-5 h-5 text-yellow-600" />
                <span className="text-2xl font-bold">{systemHealth?.metrics.avgResponseTime || 0}ms</span>
              </div>
              {(systemHealth?.metrics.avgResponseTime || 0) < 100 ? 
                <ArrowDown className="w-4 h-4 text-green-600" /> :
                <ArrowUp className="w-4 h-4 text-red-600" />
              }
            </div>
            <p className="text-xs text-muted-foreground mt-1">API响应速度</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">错误率</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <AlertTriangle className="w-5 h-5 text-red-600" />
                <span className="text-2xl font-bold">{((systemHealth?.metrics.errorRate || 0) * 100).toFixed(2)}%</span>
              </div>
              {(systemHealth?.metrics.errorRate || 0) < 0.01 ?
                <CheckCircle className="w-4 h-4 text-green-600" /> :
                <AlertCircle className="w-4 h-4 text-yellow-600" />
              }
            </div>
            <p className="text-xs text-muted-foreground mt-1">请求错误比例</p>
          </CardContent>
        </Card>
      </div>

      {/* 详细监控面板 */}
      <Tabs defaultValue="services" className="space-y-4">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="services">服务状态</TabsTrigger>
          <TabsTrigger value="resources">资源使用</TabsTrigger>
          <TabsTrigger value="performance">性能指标</TabsTrigger>
          <TabsTrigger value="security">安全事件</TabsTrigger>
          <TabsTrigger value="logs">错误日志</TabsTrigger>
        </TabsList>

        {/* 服务状态 */}
        <TabsContent value="services" className="space-y-4">
          <ServiceStatusDashboard />
        </TabsContent>

        {/* 资源使用 */}
        <TabsContent value="resources" className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-medium">CPU 使用率</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-2xl font-bold">{systemHealth?.resources.cpu}%</span>
                    <Activity className="w-5 h-5 text-blue-600" />
                  </div>
                  <Progress value={systemHealth?.resources.cpu} className="h-2" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-medium">内存使用率</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-2xl font-bold">{systemHealth?.resources.memory}%</span>
                    <Database className="w-5 h-5 text-green-600" />
                  </div>
                  <Progress value={systemHealth?.resources.memory} className="h-2" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-medium">磁盘使用率</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-2xl font-bold">{systemHealth?.resources.disk}%</span>
                    <HardDrive className="w-5 h-5 text-purple-600" />
                  </div>
                  <Progress value={systemHealth?.resources.disk} className="h-2" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-medium">网络流量</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>上行: {systemHealth?.resources.network.out} KB/s</span>
                    <span>下行: {systemHealth?.resources.network.in} KB/s</span>
                  </div>
                  <div className="flex gap-1">
                    <Progress value={30} className="h-2 flex-1" />
                    <Progress value={70} className="h-2 flex-1" />
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* 性能指标 */}
        <TabsContent value="performance" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <LineChart className="w-5 h-5" />
                实时性能监控
              </CardTitle>
              <CardDescription>
                系统性能的实时监控数据
              </CardDescription>
            </CardHeader>
            <CardContent>
              <RealtimeMetricsPanel />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <BarChart3 className="w-5 h-5" />
                Web Vitals 性能指标
              </CardTitle>
              <CardDescription>
                核心网页性能指标和用户体验数据
              </CardDescription>
            </CardHeader>
            <CardContent>
              {performanceReport && (
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div>
                    <h4 className="text-sm font-medium mb-2">最大内容绘制 (LCP)</h4>
                    <p className="text-2xl font-bold">{performanceReport.summary.largest_contentful_paint.toFixed(0)}ms</p>
                    <p className="text-xs text-muted-foreground">目标: &lt; 2500ms</p>
                  </div>
                  <div>
                    <h4 className="text-sm font-medium mb-2">首次输入延迟 (FID)</h4>
                    <p className="text-2xl font-bold">{performanceReport.summary.first_input_delay.toFixed(0)}ms</p>
                    <p className="text-xs text-muted-foreground">目标: &lt; 100ms</p>
                  </div>
                  <div>
                    <h4 className="text-sm font-medium mb-2">累积布局偏移 (CLS)</h4>
                    <p className="text-2xl font-bold">{performanceReport.summary.cumulative_layout_shift.toFixed(3)}</p>
                    <p className="text-xs text-muted-foreground">目标: &lt; 0.1</p>
                  </div>
                  <div>
                    <h4 className="text-sm font-medium mb-2">首字节时间 (TTFB)</h4>
                    <p className="text-2xl font-bold">{performanceReport.summary.time_to_first_byte.toFixed(0)}ms</p>
                    <p className="text-xs text-muted-foreground">目标: &lt; 600ms</p>
                  </div>
                  <div>
                    <h4 className="text-sm font-medium mb-2">缓存命中率</h4>
                    <p className="text-2xl font-bold">{performanceReport.summary.cache_hit_rate.toFixed(1)}%</p>
                    <p className="text-xs text-muted-foreground">目标: &gt; 80%</p>
                  </div>
                  <div>
                    <h4 className="text-sm font-medium mb-2">内存使用</h4>
                    <p className="text-2xl font-bold">{performanceReport.summary.memory_usage_mb.toFixed(0)}MB</p>
                    <p className="text-xs text-muted-foreground">当前使用量</p>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* 安全事件 */}
        <TabsContent value="security" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Shield className="w-5 h-5" />
                安全事件监控
              </CardTitle>
              <CardDescription>
                最近的安全事件和威胁检测
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {/* 安全统计 */}
                {securityStats && (
                  <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
                    <div className="text-center">
                      <p className="text-2xl font-bold">{securityStats.totalEvents}</p>
                      <p className="text-sm text-muted-foreground">总事件数</p>
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-red-600">
                        {securityStats.eventsBySeverity?.critical || 0}
                      </p>
                      <p className="text-sm text-muted-foreground">严重事件</p>
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-yellow-600">
                        {securityStats.eventsBySeverity?.high || 0}
                      </p>
                      <p className="text-sm text-muted-foreground">高危事件</p>
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-blue-600">
                        {securityStats.topIpAddresses?.length || 0}
                      </p>
                      <p className="text-sm text-muted-foreground">活跃IP</p>
                    </div>
                  </div>
                )}

                {/* 最近事件列表 */}
                <div className="space-y-2">
                  <h4 className="font-medium mb-2">最近安全事件</h4>
                  {recentSecurityEvents.length > 0 ? (
                    recentSecurityEvents.map((event) => (
                      <div key={event.id} className="flex items-center justify-between p-3 border rounded-lg">
                        <div className="flex items-center gap-3">
                          <AlertTriangle className="w-4 h-4 text-yellow-600" />
                          <div>
                            <p className="font-medium text-sm">{event.type}</p>
                            <p className="text-xs text-muted-foreground">
                              {new Date(event.timestamp).toLocaleString()}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          {getSeverityBadge(event.severity)}
                          <span className="text-xs text-muted-foreground">
                            {event.ipAddress || '未知IP'}
                          </span>
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="text-center py-8 text-muted-foreground">
                      暂无安全事件记录
                    </div>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 错误日志 */}
        <TabsContent value="logs" className="space-y-4">
          <ErrorLogViewer />
        </TabsContent>
      </Tabs>
    </div>
  )
}