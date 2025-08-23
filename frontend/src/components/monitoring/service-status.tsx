'use client'

import React, { useEffect, useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { 
  Server,
  Database,
  Wifi,
  WifiOff,
  CheckCircle,
  XCircle,
  AlertCircle,
  RefreshCw,
  Activity,
  Clock,
  TrendingUp,
  TrendingDown,
  Globe,
  Shield,
  Zap,
  HardDrive
} from 'lucide-react'

interface ServiceInfo {
  id: string
  name: string
  description: string
  endpoint: string
  status: 'online' | 'offline' | 'degraded' | 'maintenance'
  health: number // 0-100
  uptime: number // percentage
  responseTime: number // ms
  lastCheck: Date
  metrics: {
    requests: number
    errors: number
    avgResponseTime: number
  }
  dependencies: string[]
  region: string
}

const SERVICES: ServiceInfo[] = [
  {
    id: 'main-api',
    name: '主API服务',
    description: 'OpenPenPal核心API服务',
    endpoint: 'https://api.openpenpal.com',
    status: 'online',
    health: 98,
    uptime: 99.9,
    responseTime: 45,
    lastCheck: new Date(),
    metrics: { requests: 125000, errors: 23, avgResponseTime: 52 },
    dependencies: ['database', 'redis'],
    region: '北京'
  },
  {
    id: 'write-service',
    name: 'Write服务',
    description: '信件写作和处理服务',
    endpoint: 'https://write.openpenpal.com',
    status: 'online',
    health: 95,
    uptime: 99.7,
    responseTime: 120,
    lastCheck: new Date(),
    metrics: { requests: 45000, errors: 12, avgResponseTime: 98 },
    dependencies: ['main-api', 'database'],
    region: '北京'
  },
  {
    id: 'courier-service',
    name: 'Courier服务',
    description: '信使管理和投递服务',
    endpoint: 'https://courier.openpenpal.com',
    status: 'online',
    health: 96,
    uptime: 99.8,
    responseTime: 89,
    lastCheck: new Date(),
    metrics: { requests: 67000, errors: 34, avgResponseTime: 76 },
    dependencies: ['main-api', 'database', 'redis'],
    region: '上海'
  },
  {
    id: 'admin-service',
    name: 'Admin服务',
    description: '管理后台服务',
    endpoint: 'https://admin.openpenpal.com',
    status: 'online',
    health: 100,
    uptime: 99.95,
    responseTime: 67,
    lastCheck: new Date(),
    metrics: { requests: 12000, errors: 2, avgResponseTime: 61 },
    dependencies: ['main-api'],
    region: '北京'
  },
  {
    id: 'database',
    name: 'PostgreSQL数据库',
    description: '主数据库服务',
    endpoint: 'postgres://db.openpenpal.com',
    status: 'online',
    health: 100,
    uptime: 100,
    responseTime: 12,
    lastCheck: new Date(),
    metrics: { requests: 890000, errors: 0, avgResponseTime: 8 },
    dependencies: [],
    region: '北京'
  },
  {
    id: 'redis',
    name: 'Redis缓存',
    description: '缓存和会话存储',
    endpoint: 'redis://cache.openpenpal.com',
    status: 'online',
    health: 99,
    uptime: 99.99,
    responseTime: 3,
    lastCheck: new Date(),
    metrics: { requests: 2340000, errors: 12, avgResponseTime: 2 },
    dependencies: [],
    region: '北京'
  }
]

// Helper functions
const getStatusColor = (status: string) => {
  switch (status) {
    case 'online': return 'bg-green-500'
    case 'offline': return 'bg-red-500'
    case 'degraded': return 'bg-yellow-500'
    case 'maintenance': return 'bg-blue-500'
    default: return 'bg-gray-500'
  }
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'online': return <CheckCircle className="w-4 h-4 text-green-600" />
    case 'offline': return <XCircle className="w-4 h-4 text-red-600" />
    case 'degraded': return <AlertCircle className="w-4 h-4 text-yellow-600" />
    case 'maintenance': return <AlertCircle className="w-4 h-4 text-blue-600" />
    default: return <AlertCircle className="w-4 h-4 text-gray-600" />
  }
}

export function ServiceStatusDashboard() {
  const [services, setServices] = useState<ServiceInfo[]>(SERVICES)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [selectedService, setSelectedService] = useState<ServiceInfo | null>(null)
  const [showDependencyMap, setShowDependencyMap] = useState(false)

  // Simulate real-time updates
  useEffect(() => {
    const interval = setInterval(() => {
      setServices(prev => prev.map(service => ({
        ...service,
        responseTime: Math.max(1, service.responseTime + (Math.random() - 0.5) * 10),
        lastCheck: new Date(),
        metrics: {
          ...service.metrics,
          requests: service.metrics.requests + Math.floor(Math.random() * 100),
          errors: service.metrics.errors + (Math.random() > 0.95 ? 1 : 0)
        }
      })))
    }, 5000)

    return () => clearInterval(interval)
  }, [])

  const refreshStatus = async () => {
    setIsRefreshing(true)
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1000))
    setIsRefreshing(false)
  }

  const getHealthColor = (health: number) => {
    if (health >= 95) return 'text-green-600'
    if (health >= 80) return 'text-yellow-600'
    return 'text-red-600'
  }

  const overallHealth = Math.round(
    services.reduce((sum, s) => sum + s.health, 0) / services.length
  )

  const onlineServices = services.filter(s => s.status === 'online').length
  const totalRequests = services.reduce((sum, s) => sum + s.metrics.requests, 0)
  const totalErrors = services.reduce((sum, s) => sum + s.metrics.errors, 0)
  const errorRate = ((totalErrors / totalRequests) * 100).toFixed(3)

  return (
    <div className="space-y-6">
      {/* Overall Status Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">整体健康度</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className={`text-2xl font-bold ${getHealthColor(overallHealth)}`}>
                {overallHealth}%
              </span>
              <Activity className="w-5 h-5 text-muted-foreground" />
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              系统整体运行状况
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">在线服务</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold">
                {onlineServices}/{services.length}
              </span>
              <Server className="w-5 h-5 text-muted-foreground" />
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              正常运行的服务数
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">总请求数</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold">
                {(totalRequests / 1000000).toFixed(1)}M
              </span>
              <Globe className="w-5 h-5 text-muted-foreground" />
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              今日API调用次数
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">错误率</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className={`text-2xl font-bold ${parseFloat(errorRate) < 0.1 ? 'text-green-600' : 'text-red-600'}`}>
                {errorRate}%
              </span>
              <Shield className="w-5 h-5 text-muted-foreground" />
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              请求错误比例
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Service Status Grid */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>服务状态监控</CardTitle>
              <CardDescription>各个微服务的实时运行状态</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowDependencyMap(!showDependencyMap)}
              >
                {showDependencyMap ? '列表视图' : '依赖关系'}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={refreshStatus}
                disabled={isRefreshing}
              >
                <RefreshCw className={`w-4 h-4 mr-2 ${isRefreshing ? 'animate-spin' : ''}`} />
                刷新
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {showDependencyMap ? (
            <DependencyMap services={services} />
          ) : (
            <div className="space-y-4">
              {services.map((service) => (
                <ServiceCard
                  key={service.id}
                  service={service}
                  onClick={() => setSelectedService(service)}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Service Details Modal */}
      {selectedService && (
        <ServiceDetailsModal
          service={selectedService}
          onClose={() => setSelectedService(null)}
        />
      )}
    </div>
  )
}

function ServiceCard({ 
  service, 
  onClick 
}: { 
  service: ServiceInfo
  onClick: () => void 
}) {
  const isHealthy = service.health >= 95
  const hasErrors = service.metrics.errors > 0

  return (
    <div
      className="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50 cursor-pointer transition-colors"
      onClick={onClick}
    >
      <div className="flex items-center gap-4">
        <div className={`w-3 h-3 rounded-full ${getStatusColor(service.status)} animate-pulse`} />
        <div>
          <h4 className="font-medium flex items-center gap-2">
            {service.name}
            {service.status === 'online' ? (
              <Wifi className="w-4 h-4 text-green-600" />
            ) : (
              <WifiOff className="w-4 h-4 text-red-600" />
            )}
          </h4>
          <p className="text-sm text-muted-foreground">{service.description}</p>
          <div className="flex items-center gap-4 mt-1 text-xs text-muted-foreground">
            <span>Region: {service.region}</span>
            <span>Uptime: {service.uptime}%</span>
            <span>最后检查: {service.lastCheck.toLocaleTimeString()}</span>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-6">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger>
              <div className="text-center">
                <p className={`text-lg font-semibold ${service.health >= 95 ? 'text-green-600' : service.health >= 80 ? 'text-yellow-600' : 'text-red-600'}`}>
                  {service.health}%
                </p>
                <p className="text-xs text-muted-foreground">健康度</p>
              </div>
            </TooltipTrigger>
            <TooltipContent>
              <p>服务健康评分</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>

        <div className="text-center">
          <p className="text-lg font-semibold flex items-center gap-1">
            <Zap className="w-3 h-3" />
            {service.responseTime}ms
          </p>
          <p className="text-xs text-muted-foreground">响应时间</p>
        </div>

        <div className="text-center">
          <Badge variant={hasErrors ? 'destructive' : 'secondary'}>
            {service.metrics.errors} 错误
          </Badge>
        </div>
      </div>
    </div>
  )
}

function DependencyMap({ services }: { services: ServiceInfo[] }) {
  return (
    <div className="py-8">
      <Alert>
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          服务依赖关系图可视化功能正在开发中...
        </AlertDescription>
      </Alert>
    </div>
  )
}

function ServiceDetailsModal({ 
  service, 
  onClose 
}: { 
  service: ServiceInfo
  onClose: () => void 
}) {
  return (
    <Card className="fixed inset-4 md:inset-auto md:left-1/2 md:top-1/2 md:-translate-x-1/2 md:-translate-y-1/2 md:w-[600px] z-50">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>{service.name} 详细信息</CardTitle>
          <Button variant="ghost" size="sm" onClick={onClose}>
            ✕
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>服务端点</Label>
              <p className="text-sm font-mono">{service.endpoint}</p>
            </div>
            <div>
              <Label>部署区域</Label>
              <p className="text-sm">{service.region}</p>
            </div>
          </div>

          <Separator />

          <div className="grid grid-cols-3 gap-4">
            <div>
              <Label>总请求数</Label>
              <p className="text-lg font-semibold">{service.metrics.requests.toLocaleString()}</p>
            </div>
            <div>
              <Label>错误数</Label>
              <p className="text-lg font-semibold text-red-600">{service.metrics.errors}</p>
            </div>
            <div>
              <Label>平均响应时间</Label>
              <p className="text-lg font-semibold">{service.metrics.avgResponseTime}ms</p>
            </div>
          </div>

          {service.dependencies.length > 0 && (
            <>
              <Separator />
              <div>
                <Label>服务依赖</Label>
                <div className="flex flex-wrap gap-2 mt-2">
                  {service.dependencies.map(dep => (
                    <Badge key={dep} variant="outline">{dep}</Badge>
                  ))}
                </div>
              </div>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

// Helper components
function Label({ children }: { children: React.ReactNode }) {
  return <p className="text-sm font-medium text-muted-foreground mb-1">{children}</p>
}

function Separator() {
  return <div className="border-t" />
}