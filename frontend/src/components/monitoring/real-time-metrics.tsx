'use client'

import React, { useEffect, useState, useRef } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { 
  Activity, 
  TrendingUp, 
  TrendingDown,
  Minus,
  Zap,
  Users,
  Database,
  HardDrive
} from 'lucide-react'
import { performanceMonitor } from '@/lib/utils/performance-monitor'

interface MetricData {
  timestamp: number
  value: number
}

interface RealtimeMetrics {
  cpu: MetricData[]
  memory: MetricData[]
  requests: MetricData[]
  activeUsers: MetricData[]
  responseTime: MetricData[]
  errorRate: MetricData[]
}

const MAX_DATA_POINTS = 30

export function RealtimeMetricsPanel() {
  const [metrics, setMetrics] = useState<RealtimeMetrics>({
    cpu: [],
    memory: [],
    requests: [],
    activeUsers: [],
    responseTime: [],
    errorRate: []
  })

  const animationFrameRef = useRef<number>()
  const lastUpdateRef = useRef<number>(Date.now())

  useEffect(() => {
    const updateMetrics = () => {
      const now = Date.now()
      
      // Update every second
      if (now - lastUpdateRef.current < 1000) {
        animationFrameRef.current = requestAnimationFrame(updateMetrics)
        return
      }

      lastUpdateRef.current = now

      // Generate simulated real-time data (replace with actual API calls)
      setMetrics(prev => ({
        cpu: [...prev.cpu.slice(-MAX_DATA_POINTS + 1), {
          timestamp: now,
          value: 30 + Math.random() * 40 + Math.sin(now / 10000) * 10
        }],
        memory: [...prev.memory.slice(-MAX_DATA_POINTS + 1), {
          timestamp: now,
          value: 50 + Math.random() * 30 + Math.cos(now / 15000) * 10
        }],
        requests: [...prev.requests.slice(-MAX_DATA_POINTS + 1), {
          timestamp: now,
          value: 300 + Math.random() * 200 + Math.sin(now / 5000) * 50
        }],
        activeUsers: [...prev.activeUsers.slice(-MAX_DATA_POINTS + 1), {
          timestamp: now,
          value: 1000 + Math.random() * 500 + Math.sin(now / 20000) * 200
        }],
        responseTime: [...prev.responseTime.slice(-MAX_DATA_POINTS + 1), {
          timestamp: now,
          value: 50 + Math.random() * 100 + Math.sin(now / 8000) * 30
        }],
        errorRate: [...prev.errorRate.slice(-MAX_DATA_POINTS + 1), {
          timestamp: now,
          value: Math.max(0, 0.5 + Math.random() * 2 + Math.sin(now / 12000) * 0.5)
        }]
      }))

      // Also record to performance monitor
      performanceMonitor.recordMetric('system_cpu', metrics.cpu[metrics.cpu.length - 1]?.value || 0, '%', 'memory')
      performanceMonitor.recordMetric('active_users', metrics.activeUsers[metrics.activeUsers.length - 1]?.value || 0, 'count', 'user_interaction')

      animationFrameRef.current = requestAnimationFrame(updateMetrics)
    }

    animationFrameRef.current = requestAnimationFrame(updateMetrics)

    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current)
      }
    }
  }, [])

  const getTrendIcon = (data: MetricData[]) => {
    if (data.length < 2) return <Minus className="w-4 h-4 text-gray-400" />
    
    const current = data[data.length - 1].value
    const previous = data[data.length - 2].value
    const change = ((current - previous) / previous) * 100

    if (Math.abs(change) < 1) {
      return <Minus className="w-4 h-4 text-gray-400" />
    } else if (change > 0) {
      return <TrendingUp className="w-4 h-4 text-green-600" />
    } else {
      return <TrendingDown className="w-4 h-4 text-red-600" />
    }
  }

  const renderSparkline = (data: MetricData[], color: string) => {
    if (data.length < 2) return null

    const min = Math.min(...data.map(d => d.value))
    const max = Math.max(...data.map(d => d.value))
    const range = max - min || 1
    const width = 100
    const height = 30

    const points = data.map((d, i) => {
      const x = (i / (MAX_DATA_POINTS - 1)) * width
      const y = height - ((d.value - min) / range) * height
      return `${x},${y}`
    }).join(' ')

    return (
      <svg width={width} height={height} className="mt-2">
        <polyline
          points={points}
          fill="none"
          stroke={color}
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    )
  }

  const currentCPU = metrics.cpu[metrics.cpu.length - 1]?.value || 0
  const currentMemory = metrics.memory[metrics.memory.length - 1]?.value || 0
  const currentRequests = metrics.requests[metrics.requests.length - 1]?.value || 0
  const currentUsers = metrics.activeUsers[metrics.activeUsers.length - 1]?.value || 0
  const currentResponseTime = metrics.responseTime[metrics.responseTime.length - 1]?.value || 0
  const currentErrorRate = metrics.errorRate[metrics.errorRate.length - 1]?.value || 0

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {/* CPU Usage */}
      <Card>
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium">CPU 使用率</CardTitle>
            <Activity className="w-4 h-4 text-muted-foreground" />
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <span className="text-2xl font-bold">{currentCPU.toFixed(1)}%</span>
            {getTrendIcon(metrics.cpu)}
          </div>
          {renderSparkline(metrics.cpu, '#3b82f6')}
        </CardContent>
      </Card>

      {/* Memory Usage */}
      <Card>
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium">内存使用率</CardTitle>
            <Database className="w-4 h-4 text-muted-foreground" />
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <span className="text-2xl font-bold">{currentMemory.toFixed(1)}%</span>
            {getTrendIcon(metrics.memory)}
          </div>
          {renderSparkline(metrics.memory, '#10b981')}
        </CardContent>
      </Card>

      {/* Request Rate */}
      <Card>
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium">请求速率</CardTitle>
            <Zap className="w-4 h-4 text-muted-foreground" />
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <span className="text-2xl font-bold">{Math.round(currentRequests)}</span>
            {getTrendIcon(metrics.requests)}
          </div>
          <p className="text-xs text-muted-foreground">请求/分钟</p>
          {renderSparkline(metrics.requests, '#f59e0b')}
        </CardContent>
      </Card>

      {/* Active Users */}
      <Card>
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium">活跃用户</CardTitle>
            <Users className="w-4 h-4 text-muted-foreground" />
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <span className="text-2xl font-bold">{Math.round(currentUsers)}</span>
            {getTrendIcon(metrics.activeUsers)}
          </div>
          {renderSparkline(metrics.activeUsers, '#8b5cf6')}
        </CardContent>
      </Card>

      {/* Response Time */}
      <Card>
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium">响应时间</CardTitle>
            <Zap className="w-4 h-4 text-muted-foreground" />
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <span className="text-2xl font-bold">{currentResponseTime.toFixed(0)}ms</span>
            {getTrendIcon(metrics.responseTime)}
          </div>
          <Badge variant={currentResponseTime < 100 ? 'default' : currentResponseTime < 300 ? 'secondary' : 'destructive'}>
            {currentResponseTime < 100 ? '优秀' : currentResponseTime < 300 ? '正常' : '缓慢'}
          </Badge>
          {renderSparkline(metrics.responseTime, '#ec4899')}
        </CardContent>
      </Card>

      {/* Error Rate */}
      <Card>
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-medium">错误率</CardTitle>
            <Activity className="w-4 h-4 text-muted-foreground" />
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <span className="text-2xl font-bold">{currentErrorRate.toFixed(2)}%</span>
            {getTrendIcon(metrics.errorRate)}
          </div>
          <Badge variant={currentErrorRate < 1 ? 'default' : currentErrorRate < 5 ? 'secondary' : 'destructive'}>
            {currentErrorRate < 1 ? '正常' : currentErrorRate < 5 ? '警告' : '异常'}
          </Badge>
          {renderSparkline(metrics.errorRate, '#ef4444')}
        </CardContent>
      </Card>
    </div>
  )
}

// Export a simpler version for embedding
export function MiniMetricsPanel() {
  const [metrics, setMetrics] = useState({
    cpu: 0,
    memory: 0,
    responseTime: 0,
    errorRate: 0
  })

  useEffect(() => {
    const interval = setInterval(() => {
      setMetrics({
        cpu: 30 + Math.random() * 40,
        memory: 50 + Math.random() * 30,
        responseTime: 50 + Math.random() * 100,
        errorRate: Math.max(0, 0.5 + Math.random() * 2)
      })
    }, 2000)

    return () => clearInterval(interval)
  }, [])

  return (
    <div className="flex items-center gap-4 text-sm">
      <div className="flex items-center gap-1">
        <Activity className="w-3 h-3" />
        <span>CPU: {metrics.cpu.toFixed(0)}%</span>
      </div>
      <div className="flex items-center gap-1">
        <Database className="w-3 h-3" />
        <span>内存: {metrics.memory.toFixed(0)}%</span>
      </div>
      <div className="flex items-center gap-1">
        <Zap className="w-3 h-3" />
        <span>{metrics.responseTime.toFixed(0)}ms</span>
      </div>
      <Badge variant={metrics.errorRate < 1 ? 'default' : 'destructive'} className="text-xs">
        {metrics.errorRate.toFixed(1)}% 错误
      </Badge>
    </div>
  )
}