'use client'

import React, { useEffect, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { 
  Activity, 
  CheckCircle, 
  AlertTriangle, 
  XCircle,
  Wifi,
  WifiOff
} from 'lucide-react'
import Link from 'next/link'

interface SystemStatus {
  overall: 'healthy' | 'degraded' | 'critical'
  services: {
    online: number
    total: number
  }
  performance: {
    cpu: number
    memory: number
    responseTime: number
  }
  errors: number
}

export function SystemHealthIndicator() {
  const [status, setStatus] = useState<SystemStatus>({
    overall: 'healthy',
    services: { online: 6, total: 6 },
    performance: { cpu: 45, memory: 67, responseTime: 87 },
    errors: 0
  })

  // Simulate real-time updates
  useEffect(() => {
    const interval = setInterval(() => {
      setStatus(prev => ({
        overall: Math.random() > 0.9 ? 'degraded' : 'healthy',
        services: {
          online: Math.random() > 0.95 ? 5 : 6,
          total: 6
        },
        performance: {
          cpu: 30 + Math.random() * 40,
          memory: 50 + Math.random() * 30,
          responseTime: 50 + Math.random() * 100
        },
        errors: Math.random() > 0.8 ? Math.floor(Math.random() * 10) : 0
      }))
    }, 10000)

    return () => clearInterval(interval)
  }, [])

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy': return 'bg-green-500'
      case 'degraded': return 'bg-yellow-500'
      case 'critical': return 'bg-red-500'
      default: return 'bg-gray-500'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy': return <CheckCircle className="w-4 h-4" />
      case 'degraded': return <AlertTriangle className="w-4 h-4" />
      case 'critical': return <XCircle className="w-4 h-4" />
      default: return <Activity className="w-4 h-4" />
    }
  }

  const isAllServicesOnline = status.services.online === status.services.total
  const hasPerformanceIssues = status.performance.cpu > 80 || status.performance.memory > 80 || status.performance.responseTime > 300

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button 
          variant="ghost" 
          size="sm" 
          className="relative"
        >
          <div className={`absolute top-1 right-1 w-2 h-2 rounded-full ${getStatusColor(status.overall)} animate-pulse`} />
          <Activity className="w-4 h-4 mr-2" />
          系统状态
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80" align="end">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h4 className="font-semibold">系统健康状态</h4>
            <Badge 
              variant={status.overall === 'healthy' ? 'default' : status.overall === 'degraded' ? 'secondary' : 'destructive'}
              className="gap-1"
            >
              {getStatusIcon(status.overall)}
              {status.overall === 'healthy' ? '正常' : status.overall === 'degraded' ? '降级' : '异常'}
            </Badge>
          </div>

          <div className="space-y-3">
            {/* Services Status */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {isAllServicesOnline ? 
                  <Wifi className="w-4 h-4 text-green-600" /> : 
                  <WifiOff className="w-4 h-4 text-red-600" />
                }
                <span className="text-sm">服务状态</span>
              </div>
              <span className="text-sm font-medium">
                {status.services.online}/{status.services.total} 在线
              </span>
            </div>

            {/* Performance Metrics */}
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span>CPU使用率</span>
                <span className={status.performance.cpu > 80 ? 'text-red-600 font-medium' : ''}>
                  {status.performance.cpu.toFixed(0)}%
                </span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span>内存使用率</span>
                <span className={status.performance.memory > 80 ? 'text-red-600 font-medium' : ''}>
                  {status.performance.memory.toFixed(0)}%
                </span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span>响应时间</span>
                <span className={status.performance.responseTime > 300 ? 'text-red-600 font-medium' : ''}>
                  {status.performance.responseTime.toFixed(0)}ms
                </span>
              </div>
            </div>

            {/* Errors */}
            {status.errors > 0 && (
              <div className="flex items-center justify-between p-2 bg-red-50 rounded">
                <div className="flex items-center gap-2">
                  <AlertTriangle className="w-4 h-4 text-red-600" />
                  <span className="text-sm text-red-900">检测到错误</span>
                </div>
                <Badge variant="destructive">{status.errors}</Badge>
              </div>
            )}
          </div>

          <div className="pt-2 border-t">
            <Button asChild variant="outline" size="sm" className="w-full">
              <Link href="/admin/monitoring">
                查看详细监控
              </Link>
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}

// Mini version for mobile or compact views
export function SystemHealthBadge() {
  const [status, setStatus] = useState<'healthy' | 'degraded' | 'critical'>('healthy')

  useEffect(() => {
    const interval = setInterval(() => {
      setStatus(Math.random() > 0.9 ? 'degraded' : 'healthy')
    }, 10000)

    return () => clearInterval(interval)
  }, [])

  return (
    <Link href="/admin/monitoring">
      <Badge 
        variant={status === 'healthy' ? 'default' : status === 'degraded' ? 'secondary' : 'destructive'}
        className="gap-1"
      >
        <div className={`w-2 h-2 rounded-full ${
          status === 'healthy' ? 'bg-green-400' : 
          status === 'degraded' ? 'bg-yellow-400' : 
          'bg-red-400'
        } animate-pulse`} />
        {status === 'healthy' ? '系统正常' : status === 'degraded' ? '性能降级' : '系统异常'}
      </Badge>
    </Link>
  )
}