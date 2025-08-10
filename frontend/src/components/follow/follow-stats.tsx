/**
 * FollowStats Component - SOTA Implementation
 * 关注统计显示组件 - 支持详细统计、实时更新、交互式显示
 */

'use client'

import React from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { 
  Users, 
  UserPlus, 
  Heart,
  TrendingUp,
  Eye,
  Loader2,
  RefreshCw,
  MoreHorizontal
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useFollow } from '@/stores/follow-store'
import { followApi } from '@/lib/api/follow'
import { formatRelativeTime } from '@/lib/utils'
import type { FollowStatsProps, FollowStatsResponse } from '@/types/follow'

interface FollowStatsState {
  stats: FollowStatsResponse | null
  isLoading: boolean
  error: string | null
  lastUpdated: number | null
}

export function FollowStats({
  user_id,
  show_detailed = false,
  show_recent = false,
  compact = false,
  className,
}: FollowStatsProps) {
  const { followerCount, followingCount } = useFollow()
  const [localStats, setLocalStats] = React.useState<FollowStatsState>({
    stats: null,
    isLoading: false,
    error: null,
    lastUpdated: null,
  })
  
  // Load stats on mount
  React.useEffect(() => {
    if (show_detailed || show_recent) {
      loadStats()
    }
  }, [user_id, show_detailed, show_recent])
  
  const loadStats = async () => {
    setLocalStats(prev => ({ ...prev, isLoading: true, error: null }))
    
    try {
      const stats = await followApi.getFollowStats(user_id)
      setLocalStats(prev => ({
        ...prev,
        stats,
        isLoading: false,
        lastUpdated: Date.now(),
      }))
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '加载统计失败'
      setLocalStats(prev => ({
        ...prev,
        isLoading: false,
        error: errorMessage,
      }))
    }
  }
  
  const handleRefresh = () => {
    loadStats()
  }
  
  if (compact) {
    return (
      <div className={cn('flex items-center gap-4 text-sm text-muted-foreground', className)}>
        <div className="flex items-center gap-1">
          <Users className="h-4 w-4" />
          <span>{followerCount || 0} 关注者</span>
        </div>
        <div className="flex items-center gap-1">
          <UserPlus className="h-4 w-4" />
          <span>{followingCount || 0} 关注</span>
        </div>
      </div>
    )
  }
  
  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-medium">关注统计</CardTitle>
          {show_detailed && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleRefresh}
              disabled={localStats.isLoading}
              className="h-8 w-8 p-0"
            >
              {localStats.isLoading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <RefreshCw className="h-4 w-4" />
              )}
            </Button>
          )}
        </div>
        {localStats.lastUpdated && (
          <CardDescription className="text-xs">
            更新于 {formatRelativeTime(new Date(localStats.lastUpdated))}
          </CardDescription>
        )}
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Basic Stats */}
        <div className="grid grid-cols-2 gap-4">
          <div className="text-center p-3 bg-blue-50 rounded-lg">
            <div className="text-2xl font-bold text-blue-600">
              {localStats.stats?.followers_count ?? followerCount ?? 0}
            </div>
            <div className="text-sm text-blue-600">关注者</div>
          </div>
          <div className="text-center p-3 bg-green-50 rounded-lg">
            <div className="text-2xl font-bold text-green-600">
              {localStats.stats?.following_count ?? followingCount ?? 0}
            </div>
            <div className="text-sm text-green-600">关注中</div>
          </div>
        </div>
        
        {/* Detailed Stats */}
        {show_detailed && localStats.stats && (
          <>
            <Separator />
            <div className="space-y-3">
              {localStats.stats.mutual_followers_count > 0 && (
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-sm">
                    <Heart className="h-4 w-4 text-pink-500" />
                    <span>相互关注</span>
                  </div>
                  <Badge variant="secondary">
                    {localStats.stats.mutual_followers_count}
                  </Badge>
                </div>
              )}
              
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-sm">
                  <TrendingUp className="h-4 w-4 text-orange-500" />
                  <span>关注比率</span>
                </div>
                <Badge variant="outline">
                  {localStats.stats.followers_count > 0
                    ? Math.round((localStats.stats.following_count / localStats.stats.followers_count) * 100)
                    : 0}%
                </Badge>
              </div>
            </div>
          </>
        )}
        
        {/* Recent Followers */}
        {show_recent && localStats.stats?.recent_followers?.length > 0 && (
          <>
            <Separator />
            <div className="space-y-3">
              <h4 className="text-sm font-medium flex items-center gap-2">
                <Eye className="h-4 w-4" />
                最近关注者
              </h4>
              <div className="space-y-2">
                {localStats.stats.recent_followers.slice(0, 3).map((follower) => (
                  <div key={follower.id} className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-gray-200 rounded-full flex items-center justify-center">
                      {follower.avatar ? (
                        <img 
                          src={follower.avatar} 
                          alt={follower.nickname}
                          className="w-full h-full rounded-full object-cover"
                        />
                      ) : (
                        <span className="text-xs font-medium text-gray-600">
                          {follower.nickname.charAt(0)}
                        </span>
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-medium truncate">
                        {follower.nickname}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {follower.followed_at && formatRelativeTime(new Date(follower.followed_at))}
                      </div>
                    </div>
                    {follower.is_following && (
                      <Badge variant="secondary" className="text-xs">
                        互关
                      </Badge>
                    )}
                  </div>
                ))}
                
                {localStats.stats.recent_followers.length > 3 && (
                  <Button variant="ghost" size="sm" className="w-full">
                    <MoreHorizontal className="h-4 w-4 mr-2" />
                    查看更多 ({localStats.stats.recent_followers.length - 3})
                  </Button>
                )}
              </div>
            </div>
          </>
        )}
        
        {/* Popular Following */}
        {show_recent && localStats.stats?.popular_following?.length > 0 && (
          <>
            <Separator />
            <div className="space-y-3">
              <h4 className="text-sm font-medium flex items-center gap-2">
                <TrendingUp className="h-4 w-4" />
                热门关注
              </h4>
              <div className="space-y-2">
                {localStats.stats.popular_following.slice(0, 3).map((user) => (
                  <div key={user.id} className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-gray-200 rounded-full flex items-center justify-center">
                      {user.avatar ? (
                        <img 
                          src={user.avatar} 
                          alt={user.nickname}
                          className="w-full h-full rounded-full object-cover"
                        />
                      ) : (
                        <span className="text-xs font-medium text-gray-600">
                          {user.nickname.charAt(0)}
                        </span>
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-medium truncate">
                        {user.nickname}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {user.followers_count} 关注者
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </>
        )}
        
        {/* Loading State */}
        {localStats.isLoading && !localStats.stats && (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        )}
        
        {/* Error State */}
        {localStats.error && (
          <div className="text-center py-4 text-sm text-red-600">
            {localStats.error}
            <Button
              variant="ghost"
              size="sm"
              onClick={handleRefresh}
              className="ml-2"
            >
              重试
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

/**
 * Simple inline stats display
 */
export function InlineFollowStats({
  user_id,
  followerCount,
  followingCount,
  className,
}: {
  user_id?: string
  followerCount: number
  followingCount: number
  className?: string
}) {
  return (
    <div className={cn('flex items-center gap-4', className)}>
      <Button variant="ghost" size="sm" className="h-auto p-0 hover:bg-transparent">
        <div className="text-center">
          <div className="font-semibold">{followerCount}</div>
          <div className="text-xs text-muted-foreground">关注者</div>
        </div>
      </Button>
      <Button variant="ghost" size="sm" className="h-auto p-0 hover:bg-transparent">
        <div className="text-center">
          <div className="font-semibold">{followingCount}</div>
          <div className="text-xs text-muted-foreground">关注</div>
        </div>
      </Button>
    </div>
  )
}

/**
 * Badge-style stats display
 */
export function BadgeFollowStats({
  followerCount,
  followingCount,
  className,
}: {
  followerCount: number
  followingCount: number
  className?: string
}) {
  return (
    <div className={cn('flex items-center gap-2', className)}>
      <Badge variant="secondary" className="text-xs">
        <Users className="h-3 w-3 mr-1" />
        {followerCount} 关注者
      </Badge>
      <Badge variant="outline" className="text-xs">
        <UserPlus className="h-3 w-3 mr-1" />
        {followingCount} 关注
      </Badge>
    </div>
  )
}

export default FollowStats