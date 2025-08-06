'use client'

import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Sparkles, Bot, Users, Palette, AlertCircle, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { aiService } from '@/lib/services/ai-service'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

interface UsageStats {
  userId: number
  usage: {
    matches_created: number
    replies_generated: number
    inspirations_used: number
    letters_curated: number
  }
  limits: {
    daily_matches: number
    daily_replies: number
    daily_inspirations: number
    daily_curations: number
  }
  remaining: {
    matches: number
    replies: number
    inspirations: number
    curations: number
  }
}

interface UsageStatsCardProps {
  className?: string
}

export function UsageStatsCard({ className = '' }: UsageStatsCardProps) {
  const [stats, setStats] = useState<UsageStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchStats = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await aiService.getAIStats()
      setStats(response)
    } catch (err: any) {
      setError(err.message || '获取使用统计失败')
      toast.error('获取使用统计失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchStats()
  }, [])

  if (loading) {
    return (
      <Card className={className}>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <RefreshCw className="h-5 w-5 animate-spin" />
            使用统计
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="space-y-2">
                <div className="h-4 bg-gray-200 rounded animate-pulse"></div>
                <div className="h-2 bg-gray-100 rounded animate-pulse"></div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className={className}>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-red-500" />
            使用统计
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error}
              <Button
                variant="ghost"
                size="sm"
                onClick={fetchStats}
                className="ml-2"
              >
                重试
              </Button>
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    )
  }

  if (!stats) {
    return null
  }

  const usageItems = [
    {
      icon: Sparkles,
      label: '写作灵感',
      used: stats.usage.inspirations_used,
      limit: stats.limits.daily_inspirations,
      remaining: stats.remaining.inspirations,
      color: 'text-amber-600',
      bgColor: 'bg-amber-100',
    },
    {
      icon: Bot,
      label: 'AI回信',
      used: stats.usage.replies_generated,
      limit: stats.limits.daily_replies,
      remaining: stats.remaining.replies,
      color: 'text-blue-600',
      bgColor: 'bg-blue-100',
    },
    {
      icon: Users,
      label: '笔友匹配',
      used: stats.usage.matches_created,
      limit: stats.limits.daily_matches,
      remaining: stats.remaining.matches,
      color: 'text-purple-600',
      bgColor: 'bg-purple-100',
    },
    {
      icon: Palette,
      label: '信件策展',
      used: stats.usage.letters_curated,
      limit: stats.limits.daily_curations,
      remaining: stats.remaining.curations,
      color: 'text-green-600',
      bgColor: 'bg-green-100',
    },
  ]

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <RefreshCw className="h-5 w-5" />
          今日使用量
        </CardTitle>
        <CardDescription>
          每日使用限制帮助维持平台的慢节奏体验
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {usageItems.map((item) => {
          const Icon = item.icon
          const percentage = item.limit > 0 ? (item.used / item.limit) * 100 : 0
          const isNearLimit = percentage >= 80
          const isAtLimit = item.remaining === 0

          return (
            <div key={item.label} className="space-y-2">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className={`p-1.5 rounded ${item.bgColor}`}>
                    <Icon className={`h-4 w-4 ${item.color}`} />
                  </div>
                  <span className="font-medium text-sm">{item.label}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-muted-foreground">
                    {item.used}/{item.limit}
                  </span>
                  {isAtLimit ? (
                    <Badge variant="secondary" className="text-xs">
                      已用尽
                    </Badge>
                  ) : isNearLimit ? (
                    <Badge variant="outline" className="text-xs">
                      接近上限
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="text-xs text-green-600">
                      剩余 {item.remaining}
                    </Badge>
                  )}
                </div>
              </div>
              <div className="relative h-2 w-full overflow-hidden rounded-full bg-gray-200">
                <div
                  className={cn(
                    "h-full transition-all",
                    isAtLimit ? 'bg-red-500' : 
                    isNearLimit ? 'bg-yellow-500' : 
                    'bg-green-500'
                  )}
                  style={{ width: `${percentage}%` }}
                />
              </div>
            </div>
          )
        })}

        {/* 提示信息 */}
        <div className="mt-4 p-3 bg-blue-50 rounded-lg">
          <div className="text-sm text-blue-800">
            <div className="font-medium mb-1">💡 使用建议</div>
            <ul className="text-xs space-y-1 text-blue-700">
              <li>• 每日灵感限制2条，珍惜每次创作机会</li>
              <li>• AI回信延迟24小时，体验慢节奏魅力</li>
              <li>• 笔友匹配每日3次，用心选择对象</li>
              <li>• 使用量每日24:00重置</li>
            </ul>
          </div>
        </div>

        <div className="flex justify-end">
          <Button variant="ghost" size="sm" onClick={fetchStats}>
            <RefreshCw className="h-4 w-4 mr-1" />
            刷新统计
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}