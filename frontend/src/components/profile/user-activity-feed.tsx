/**
 * UserActivityFeed - User Activity Timeline Component
 * 用户活动时间线组件 - 展示用户的近期活动
 */

'use client'

import React, { useState, useEffect } from 'react'
import { 
  Clock, 
  Send, 
  MessageSquare, 
  Award, 
  Users, 
  Package,
  Heart,
  ChevronDown,
  Calendar,
  MapPin,
  BookOpen
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import { formatDistanceToNow } from '@/lib/utils/date'

interface Activity {
  id: string
  type: 'letter_sent' | 'letter_received' | 'comment' | 'achievement' | 'follow' | 'museum_contribution' | 'courier_task'
  title: string
  description?: string
  metadata?: {
    letter_id?: string
    letter_title?: string
    user_id?: string
    username?: string
    achievement_name?: string
    museum_item_id?: string
    task_id?: string
    op_code?: string
  }
  created_at: string
  icon_color?: string
}

interface UserActivityFeedProps {
  user_id: string
  max_items?: number
  show_load_more?: boolean
  className?: string
}

// Mock activities - will be replaced with real API
const mockActivities: Activity[] = [
  {
    id: '1',
    type: 'letter_sent',
    title: '发送了一封信',
    description: '给远方的朋友写了一封关于夏天的信',
    metadata: {
      letter_id: '123',
      letter_title: '夏日的回忆',
      op_code: 'PK5F3D'
    },
    created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
    icon_color: 'text-blue-600'
  },
  {
    id: '2',
    type: 'achievement',
    title: '获得新成就',
    description: '解锁了"笔友达人"成就',
    metadata: {
      achievement_name: 'pen_pal_master'
    },
    created_at: new Date(Date.now() - 5 * 60 * 60 * 1000).toISOString(),
    icon_color: 'text-yellow-600'
  },
  {
    id: '3',
    type: 'follow',
    title: '关注了新朋友',
    metadata: {
      user_id: '456',
      username: 'alice'
    },
    created_at: new Date(Date.now() - 12 * 60 * 60 * 1000).toISOString(),
    icon_color: 'text-pink-600'
  },
  {
    id: '4',
    type: 'museum_contribution',
    title: '贡献了博物馆藏品',
    description: '分享了一封特别的信到信件博物馆',
    metadata: {
      museum_item_id: '789'
    },
    created_at: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
    icon_color: 'text-purple-600'
  },
  {
    id: '5',
    type: 'courier_task',
    title: '完成了信使任务',
    description: '成功投递了3封信件',
    metadata: {
      task_id: '101',
      op_code: 'QH2G1A'
    },
    created_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
    icon_color: 'text-green-600'
  },
  {
    id: '6',
    type: 'comment',
    title: '发表了评论',
    description: '在博物馆展品下留下了精彩评论',
    metadata: {
      museum_item_id: '202'
    },
    created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
    icon_color: 'text-orange-600'
  }
]

export function UserActivityFeed({ 
  user_id, 
  max_items = 5,
  show_load_more = true,
  className 
}: UserActivityFeedProps) {
  const [activities, setActivities] = useState<Activity[]>([])
  const [loading, setLoading] = useState(true)
  const [hasMore, setHasMore] = useState(true)
  const [displayCount, setDisplayCount] = useState(max_items)

  useEffect(() => {
    loadActivities()
  }, [user_id])

  const loadActivities = async () => {
    try {
      setLoading(true)
      // TODO: Replace with real API call
      // const response = await fetch(`/api/users/${user_id}/activities`)
      // const data = await response.json()
      
      // Mock implementation
      setTimeout(() => {
        setActivities(mockActivities)
        setHasMore(mockActivities.length > max_items)
        setLoading(false)
      }, 500)
    } catch (error) {
      console.error('Failed to load activities:', error)
      setLoading(false)
    }
  }

  const getActivityIcon = (type: Activity['type']) => {
    switch (type) {
      case 'letter_sent':
        return Send
      case 'letter_received':
        return Package
      case 'comment':
        return MessageSquare
      case 'achievement':
        return Award
      case 'follow':
        return Users
      case 'museum_contribution':
        return BookOpen
      case 'courier_task':
        return MapPin
      default:
        return Clock
    }
  }

  const formatActivityTime = (timestamp: string) => {
    try {
      return formatDistanceToNow(new Date(timestamp), { addSuffix: true })
    } catch {
      return '刚刚'
    }
  }

  const loadMore = () => {
    setDisplayCount(prev => prev + max_items)
  }

  if (loading) {
    return (
      <Card className={cn('w-full', className)}>
        <CardContent className="p-6">
          <div className="space-y-4">
            {[1, 2, 3].map(i => (
              <div key={i} className="flex gap-3 animate-pulse">
                <div className="w-10 h-10 bg-muted rounded-full" />
                <div className="flex-1 space-y-2">
                  <div className="h-4 bg-muted rounded w-3/4" />
                  <div className="h-3 bg-muted rounded w-1/2" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  const displayedActivities = activities.slice(0, displayCount)

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            <Calendar className="h-5 w-5" />
            最近动态
          </CardTitle>
          <Badge variant="secondary">
            {activities.length} 条动态
          </Badge>
        </div>
      </CardHeader>
      
      <CardContent className="p-0">
        {activities.length === 0 ? (
          <div className="p-8 text-center">
            <Clock className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <p className="text-muted-foreground">暂无动态</p>
          </div>
        ) : (
          <>
            <div className="px-6 pb-4">
              <div className="relative">
                {/* Timeline line */}
                <div className="absolute left-5 top-8 bottom-0 w-0.5 bg-border" />
                
                {/* Activities */}
                <div className="space-y-6">
                  {displayedActivities.map((activity, index) => {
                    const Icon = getActivityIcon(activity.type)
                    const isLast = index === displayedActivities.length - 1
                    
                    return (
                      <div key={activity.id} className="relative flex gap-3">
                        {/* Icon */}
                        <div className={cn(
                          "relative z-10 flex h-10 w-10 items-center justify-center rounded-full border-2 bg-background",
                          activity.icon_color || 'text-muted-foreground'
                        )}>
                          <Icon className="h-5 w-5" />
                          {!isLast && (
                            <div className="absolute top-10 left-1/2 -translate-x-1/2 w-0.5 h-6 bg-border" />
                          )}
                        </div>
                        
                        {/* Content */}
                        <div className="flex-1 space-y-1">
                          <div className="flex items-start justify-between gap-2">
                            <div>
                              <p className="font-medium text-sm">
                                {activity.title}
                                {activity.metadata?.username && (
                                  <a 
                                    href={`/u/${activity.metadata.username}`}
                                    className="ml-1 text-primary hover:underline"
                                  >
                                    @{activity.metadata.username}
                                  </a>
                                )}
                              </p>
                              {activity.description && (
                                <p className="text-sm text-muted-foreground mt-0.5">
                                  {activity.description}
                                </p>
                              )}
                              {activity.metadata?.op_code && (
                                <div className="flex items-center gap-1 mt-1">
                                  <MapPin className="h-3 w-3 text-muted-foreground" />
                                  <span className="text-xs text-muted-foreground">
                                    {activity.metadata.op_code}
                                  </span>
                                </div>
                              )}
                            </div>
                          </div>
                          <p className="text-xs text-muted-foreground">
                            {formatActivityTime(activity.created_at)}
                          </p>
                        </div>
                      </div>
                    )
                  })}
                </div>
              </div>
            </div>
            
            {/* Load more */}
            {show_load_more && hasMore && displayCount < activities.length && (
              <>
                <Separator />
                <div className="p-4 text-center">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={loadMore}
                    className="gap-2"
                  >
                    <ChevronDown className="h-4 w-4" />
                    加载更多
                  </Button>
                </div>
              </>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}

export default UserActivityFeed