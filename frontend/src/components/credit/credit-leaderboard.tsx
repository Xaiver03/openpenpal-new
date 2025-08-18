'use client'

import React, { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Skeleton } from '@/components/ui/skeleton'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { 
  Trophy, 
  Medal, 
  Award, 
  Crown,
  TrendingUp,
  RefreshCw,
  Users,
  Star,
  ChevronUp,
  ChevronDown,
  Minus
} from 'lucide-react'
import { useCreditLeaderboard, useCreditStore } from '@/stores/credit-store'
import { formatPoints, getCreditLevelName } from '@/lib/api/credit'
import { CreditLevelBadge } from './credit-level-badge'
import type { UserCreditLeaderboard } from '@/types/credit'

interface CreditLeaderboardProps {
  limit?: number
  showTimeFilter?: boolean
  showCurrentUser?: boolean
  className?: string
}

export function CreditLeaderboard({ 
  limit = 20,
  showTimeFilter = true,
  showCurrentUser = true,
  className = '' 
}: CreditLeaderboardProps) {
  const { leaderboard, loading, error } = useCreditLeaderboard()
  const { fetchLeaderboard, clearError } = useCreditStore()
  const [currentUserRank, setCurrentUserRank] = useState<UserCreditLeaderboard | null>(null)
  
  const [timeRange, setTimeRange] = useState<'all' | 'month' | 'week'>('all')
  const [category, setCategory] = useState<'total' | 'level' | 'tasks'>('total')

  useEffect(() => {
    fetchLeaderboard(limit)
  }, [timeRange, category, limit, fetchLeaderboard])

  const handleRefresh = () => {
    clearError()
    fetchLeaderboard(limit)
  }

  const getRankIcon = (rank: number) => {
    switch (rank) {
      case 1:
        return <Crown className="h-5 w-5 text-yellow-500" />
      case 2:
        return <Trophy className="h-5 w-5 text-gray-400" />
      case 3:
        return <Medal className="h-5 w-5 text-amber-600" />
      default:
        return (
          <div className="h-5 w-5 rounded-full bg-muted flex items-center justify-center text-xs font-medium">
            {rank}
          </div>
        )
    }
  }

  const getRankBackground = (rank: number) => {
    switch (rank) {
      case 1:
        return 'bg-gradient-to-r from-yellow-50 to-amber-50 dark:from-yellow-950 dark:to-amber-950 border-yellow-200 dark:border-yellow-800'
      case 2:
        return 'bg-gradient-to-r from-gray-50 to-slate-50 dark:from-gray-950 dark:to-slate-950 border-gray-200 dark:border-gray-800'
      case 3:
        return 'bg-gradient-to-r from-amber-50 to-orange-50 dark:from-amber-950 dark:to-orange-950 border-amber-200 dark:border-amber-800'
      default:
        return 'bg-card hover:bg-muted/50'
    }
  }

  const getRankChangeIcon = (change: number) => {
    if (change > 0) return <ChevronUp className="h-3 w-3 text-green-500" />
    if (change < 0) return <ChevronDown className="h-3 w-3 text-red-500" />
    return <Minus className="h-3 w-3 text-gray-400" />
  }

  const getTimeRangeLabel = (range: string) => {
    switch (range) {
      case 'all': return '总榜'
      case 'month': return '月榜'
      case 'week': return '周榜'
      default: return '总榜'
    }
  }

  const getCategoryLabel = (cat: string) => {
    switch (cat) {
      case 'total': return '总积分'
      case 'level': return '等级'
      case 'tasks': return '任务数'
      default: return '总积分'
    }
  }

  if (loading && !leaderboard) {
    return (
      <Card className={`w-full ${className}`}>
        <CardHeader>
          <Skeleton className="h-6 w-32" />
        </CardHeader>
        <CardContent className="space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-center space-x-4 p-4 border rounded-lg">
              <Skeleton className="h-8 w-8 rounded-full" />
              <Skeleton className="h-10 w-10 rounded-full" />
              <div className="flex-1 space-y-2">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-3 w-24" />
              </div>
              <Skeleton className="h-6 w-16" />
            </div>
          ))}
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className={`w-full ${className}`}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Trophy className="h-5 w-5" />
            积分排行榜
          </CardTitle>
          
          <div className="flex items-center gap-2">
            {showTimeFilter && (
              <>
                <Select
                  value={timeRange}
                  onValueChange={(value: 'all' | 'month' | 'week') => setTimeRange(value)}
                >
                  <SelectTrigger className="w-20">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">总榜</SelectItem>
                    <SelectItem value="month">月榜</SelectItem>
                    <SelectItem value="week">周榜</SelectItem>
                  </SelectContent>
                </Select>
                
                <Select
                  value={category}
                  onValueChange={(value: 'total' | 'level' | 'tasks') => setCategory(value)}
                >
                  <SelectTrigger className="w-24">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="total">总积分</SelectItem>
                    <SelectItem value="level">等级</SelectItem>
                    <SelectItem value="tasks">任务数</SelectItem>
                  </SelectContent>
                </Select>
              </>
            )}
            
            <Button
              onClick={handleRefresh}
              disabled={loading}
              variant="ghost"
              size="sm"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            </Button>
          </div>
        </div>
        
        <div className="text-sm text-muted-foreground">
          {getTimeRangeLabel(timeRange)} · {getCategoryLabel(category)}
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* 错误状态 */}
        {error && (
          <div className="text-center p-4 text-destructive text-sm bg-destructive/10 rounded-lg">
            {error}
          </div>
        )}

        {/* 当前用户排名 */}
        {showCurrentUser && currentUserRank && (
          <div className="p-4 bg-primary/5 border border-primary/20 rounded-lg">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="flex items-center gap-2">
                  {getRankIcon(currentUserRank.rank)}
                  <span className="font-medium">您的排名</span>
                </div>
                <Badge variant="outline" className="text-xs">
                  #{currentUserRank.rank}
                </Badge>
              </div>
              <div className="text-right">
                <div className="font-medium">
                  {formatPoints(currentUserRank.totalPoints)}
                </div>
                <div className="text-xs text-muted-foreground">
                  {getCreditLevelName(currentUserRank.level)}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* 排行榜列表 */}
        <div className="space-y-2">
          {!leaderboard || leaderboard.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Users className="h-12 w-12 mx-auto mb-2 opacity-50" />
              暂无排行榜数据
            </div>
          ) : (
            leaderboard.map((user, index) => (
              <div
                key={user.user_id}
                className={`flex items-center justify-between p-4 rounded-lg border transition-colors ${getRankBackground(user.rank)}`}
              >
                <div className="flex items-center space-x-4 flex-1">
                  {/* 排名 */}
                  <div className="flex items-center gap-2 w-12">
                    {getRankIcon(user.rank)}
                    {user.rankChange !== undefined && (
                      <div className="flex items-center">
                        {getRankChangeIcon(user.rankChange)}
                      </div>
                    )}
                  </div>

                  {/* 用户头像 */}
                  <Avatar className="h-10 w-10">
                    <AvatarImage src={user.avatarUrl} alt={user.username} />
                    <AvatarFallback>
                      {user.username.slice(0, 2).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>

                  {/* 用户信息 */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium truncate">
                        {user.username}
                      </span>
                      <CreditLevelBadge
                        level={user.level}
                        totalPoints={user.totalPoints}
                        showTooltip={false}
                        size="sm"
                      />
                    </div>
                    
                    <div className="flex items-center gap-4 text-xs text-muted-foreground mt-1">
                      <span>等级 {user.level}</span>
                      {user.completedTasks && (
                        <span>任务 {user.completedTasks}</span>
                      )}
                      {timeRange === 'week' && user.weekPoints && (
                        <span className="text-green-600">
                          本周 +{formatPoints(user.weekPoints)}
                        </span>
                      )}
                      {timeRange === 'month' && user.monthPoints && (
                        <span className="text-blue-600">
                          本月 +{formatPoints(user.monthPoints)}
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                {/* 积分显示 */}
                <div className="text-right">
                  <div className="font-bold text-lg">
                    {formatPoints(user.totalPoints)}
                  </div>
                  
                  {/* 特殊徽章 */}
                  <div className="flex items-center justify-end gap-1 mt-1">
                    {user.rank <= 3 && (
                      <Badge 
                        variant={user.rank === 1 ? 'default' : 'secondary'}
                        className="text-xs"
                      >
                        {user.rank === 1 ? '🥇 冠军' : user.rank === 2 ? '🥈 亚军' : '🥉 季军'}
                      </Badge>
                    )}
                    
                    {user.isRising && (
                      <Badge variant="outline" className="text-xs text-green-600">
                        <TrendingUp className="h-3 w-3 mr-1" />
                        上升
                      </Badge>
                    )}
                    
                    {user.achievements && user.achievements.length > 0 && (
                      <Badge variant="outline" className="text-xs">
                        <Star className="h-3 w-3 mr-1" />
                        {user.achievements.length}
                      </Badge>
                    )}
                  </div>
                </div>
              </div>
            ))
          )}
        </div>

        {/* 排行榜统计 */}
        {leaderboard && leaderboard.length > 0 && (
          <div className="pt-4 border-t">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div className="text-center">
                <div className="font-medium text-yellow-600">
                  {formatPoints(leaderboard[0]?.totalPoints || 0)}
                </div>
                <div className="text-muted-foreground">第一名积分</div>
              </div>
              <div className="text-center">
                <div className="font-medium text-blue-600">
                  {formatPoints(
                    Math.round(leaderboard.reduce((sum, user) => sum + user.totalPoints, 0) / leaderboard.length)
                  )}
                </div>
                <div className="text-muted-foreground">平均积分</div>
              </div>
              <div className="text-center">
                <div className="font-medium text-green-600">
                  {leaderboard.filter(user => user.isRising).length}
                </div>
                <div className="text-muted-foreground">上升人数</div>
              </div>
              <div className="text-center">
                <div className="font-medium text-purple-600">
                  {Math.max(...leaderboard.map(user => user.level))}
                </div>
                <div className="text-muted-foreground">最高等级</div>
              </div>
            </div>
          </div>
        )}

        {/* 刷新提示 */}
        <div className="text-center text-xs text-muted-foreground pt-2">
          数据每小时更新一次 · 最后更新: {new Date().toLocaleTimeString('zh-CN')}
        </div>
      </CardContent>
    </Card>
  )
}

export default CreditLeaderboard