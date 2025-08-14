/**
 * UserSuggestions Component - SOTA Implementation
 * 用户推荐组件 - 智能推荐算法、多种推荐理由、批量操作
 */

'use client'

import React from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { 
  Sparkles,
  RefreshCw,
  Loader2,
  Users,
  GraduationCap,
  TrendingUp,
  Clock,
  Heart,
  X,
  CheckCheck,
  AlertCircle
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useFollow } from '@/stores/follow-store'
import { UserCard, UserCardSkeleton } from './user-card'
import { toast } from 'sonner'
import type { 
  UserSuggestionsProps, 
  FollowSuggestion, 
  SuggestionReason,
  UserSuggestionsQuery 
} from '@/types/follow'

interface SuggestionsState {
  dismissedIds: Set<string>
  followingIds: Set<string>
  isRefreshing: boolean
  lastRefresh: number | null
  bulkFollowLoading: boolean
}

export function UserSuggestions({
  limit = 6,
  show_reason = true,
  show_mutual = true,
  show_refresh = true,
  algorithm = 'school',
  className,
  onUserFollow,
}: UserSuggestionsProps) {
  const { 
    suggestions, 
    isLoading, 
    loadSuggestions, 
    refreshSuggestions,
    followUser, 
    followMultipleUsers 
  } = useFollow()
  
  const [localState, setLocalState] = React.useState<SuggestionsState>({
    dismissedIds: new Set(),
    followingIds: new Set(),
    isRefreshing: false,
    lastRefresh: null,
    bulkFollowLoading: false,
  })
  
  // Load suggestions on mount
  React.useEffect(() => {
    if (suggestions.length === 0 && !isLoading.suggestions) {
      handleLoadSuggestions()
    }
  }, [])
  
  const handleLoadSuggestions = async () => {
    const query: UserSuggestionsQuery = {
      limit,
      based_on: algorithm,
      exclude_followed: true,
    }
    
    try {
      await loadSuggestions(query)
    } catch (error) {
      toast.error('加载推荐失败')
    }
  }
  
  const handleRefreshSuggestions = async () => {
    setLocalState(prev => ({ ...prev, isRefreshing: true }))
    
    try {
      await refreshSuggestions()
      setLocalState(prev => ({ 
        ...prev, 
        isRefreshing: false, 
        lastRefresh: Date.now(),
        dismissedIds: new Set(), // Clear dismissed items on refresh
      }))
      toast.success('推荐已更新')
    } catch (error) {
      setLocalState(prev => ({ ...prev, isRefreshing: false }))
      toast.error('刷新失败，请稍后重试')
    }
  }
  
  const handleFollowUser = async (suggestion: FollowSuggestion) => {
    setLocalState(prev => ({
      ...prev,
      followingIds: new Set([...prev.followingIds, suggestion.user.id])
    }))
    
    try {
      await followUser(suggestion.user.id)
      onUserFollow?.(suggestion.user)
      toast.success(`已关注 ${suggestion.user.nickname}`)
    } catch (error) {
      setLocalState(prev => ({
        ...prev,
        followingIds: new Set([...prev.followingIds].filter(id => id !== suggestion.user.id))
      }))
      toast.error('关注失败')
    }
  }
  
  const handleDismissSuggestion = (suggestionId: string) => {
    setLocalState(prev => ({
      ...prev,
      dismissedIds: new Set([...prev.dismissedIds, suggestionId])
    }))
    toast.success('已隐藏此推荐')
  }
  
  const handleFollowAll = async () => {
    const visibleSuggestions = filteredSuggestions.slice(0, 3) // Limit to first 3 for safety
    const userIds = visibleSuggestions.map(s => s.user.id)
    
    setLocalState(prev => ({ ...prev, bulkFollowLoading: true }))
    
    try {
      const result = await followMultipleUsers(userIds)
      
      // Update following state
      setLocalState(prev => ({
        ...prev,
        followingIds: new Set([...prev.followingIds, ...result.success]),
        bulkFollowLoading: false,
      }))
      
      // Notify about results
      if (result.success.length > 0) {
        toast.success(`已关注 ${result.success.length} 位用户`)
        result.success.forEach(userId => {
          const user = visibleSuggestions.find(s => s.user.id === userId)?.user
          if (user) onUserFollow?.(user)
        })
      }
      
      if (result.failed.length > 0) {
        toast.error(`${result.failed.length} 位用户关注失败`)
      }
    } catch (error) {
      setLocalState(prev => ({ ...prev, bulkFollowLoading: false }))
      toast.error('批量关注失败')
    }
  }
  
  const getSuggestionReasonInfo = (reason: SuggestionReason) => {
    switch (reason) {
      case 'same_school':
        return {
          icon: GraduationCap,
          label: '同校用户',
          color: 'text-blue-600 bg-blue-50',
        }
      case 'mutual_followers':
        return {
          icon: Users,
          label: '共同关注',
          color: 'text-green-600 bg-green-50',
        }
      case 'similar_interests':
        return {
          icon: Heart,
          label: '兴趣相似',
          color: 'text-pink-600 bg-pink-50',
        }
      case 'active_user':
        return {
          icon: TrendingUp,
          label: '活跃用户',
          color: 'text-orange-600 bg-orange-50',
        }
      case 'new_user':
        return {
          icon: Sparkles,
          label: '新用户',
          color: 'text-purple-600 bg-purple-50',
        }
      case 'trending':
        return {
          icon: TrendingUp,
          label: '热门',
          color: 'text-red-600 bg-red-50',
        }
      default:
        return {
          icon: Users,
          label: '推荐',
          color: 'text-gray-600 bg-gray-50',
        }
    }
  }
  
  const filteredSuggestions = suggestions.filter(
    suggestion => !localState.dismissedIds.has(suggestion.user.id)
  ).slice(0, limit)
  
  const renderSuggestionCard = (suggestion: FollowSuggestion) => {
    const reasonInfo = getSuggestionReasonInfo(suggestion.reason)
    const Icon = reasonInfo.icon
    const isFollowing = localState.followingIds.has(suggestion.user.id)
    
    return (
      <Card key={suggestion.user.id} className="relative group hover:shadow-lg transition-all">
        {/* Dismiss button */}
        <Button
          variant="ghost"
          size="sm"
          onClick={() => handleDismissSuggestion(suggestion.user.id)}
          className="absolute top-2 right-2 h-6 w-6 p-0 opacity-0 group-hover:opacity-100 transition-opacity"
        >
          <X className="h-3 w-3" />
        </Button>
        
        <CardContent className="p-4">
          <UserCard
            user={suggestion.user}
            variant="compact"
            show_follow_button={false}
            show_stats={false}
            show_mutual={show_mutual}
            clickable={true}
          />
          
          {/* Recommendation reason */}
          {show_reason && (
            <div className="mt-3">
              <Badge variant="secondary" className={cn('text-xs', reasonInfo.color)}>
                <Icon className="h-3 w-3 mr-1" />
                {reasonInfo.label}
              </Badge>
              {suggestion.confidence_score && (
                <span className="text-xs text-muted-foreground ml-2">
                  匹配度 {Math.round(suggestion.confidence_score * 100)}%
                </span>
              )}
            </div>
          )}
          
          {/* Mutual followers */}
          {show_mutual && suggestion.mutual_followers && suggestion.mutual_followers.length > 0 && (
            <div className="mt-2">
              <p className="text-xs text-muted-foreground">
                与 {suggestion.mutual_followers.slice(0, 2).map(user => user.nickname).join('、')} 
                {suggestion.mutual_followers.length > 2 && ` 等${suggestion.mutual_followers.length}人`}
                共同关注
              </p>
            </div>
          )}
          
          {/* Common interests */}
          {suggestion.common_interests && suggestion.common_interests.length > 0 && (
            <div className="mt-2 flex flex-wrap gap-1">
              {suggestion.common_interests.slice(0, 3).map((interest) => (
                <Badge key={interest} variant="outline" className="text-xs">
                  {interest}
                </Badge>
              ))}
            </div>
          )}
          
          {/* Follow button */}
          <div className="mt-4">
            <Button
              onClick={() => handleFollowUser(suggestion)}
              disabled={isFollowing}
              className="w-full"
              size="sm"
            >
              {isFollowing ? (
                <>
                  <CheckCheck className="h-4 w-4 mr-2" />
                  已关注
                </>
              ) : (
                <>
                  <Users className="h-4 w-4 mr-2" />
                  关注
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }
  
  const renderHeader = () => (
    <div className="flex items-center justify-between mb-4">
      <div className="flex items-center gap-2">
        <Sparkles className="h-5 w-5 text-primary" />
        <h2 className="font-semibold text-lg">推荐关注</h2>
      </div>
      
      <div className="flex items-center gap-2">
        {filteredSuggestions.length > 2 && (
          <Button
            variant="outline"
            size="sm"
            onClick={handleFollowAll}
            disabled={localState.bulkFollowLoading}
          >
            {localState.bulkFollowLoading ? (
              <Loader2 className="h-4 w-4 animate-spin mr-2" />
            ) : (
              <CheckCheck className="h-4 w-4 mr-2" />
            )}
            全部关注
          </Button>
        )}
        
        {show_refresh && (
          <Button
            variant="ghost"
            size="sm"
            onClick={handleRefreshSuggestions}
            disabled={localState.isRefreshing}
          >
            {localState.isRefreshing ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="h-4 w-4" />
            )}
          </Button>
        )}
      </div>
    </div>
  )
  
  if (isLoading.suggestions && filteredSuggestions.length === 0) {
    return (
      <div className={className}>
        {renderHeader()}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardContent className="p-4">
                <UserCardSkeleton variant="compact" />
                <div className="mt-3 space-y-2">
                  <div className="h-4 bg-muted rounded w-20" />
                  <div className="h-8 bg-muted rounded" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    )
  }
  
  if (filteredSuggestions.length === 0) {
    return (
      <div className={className}>
        {renderHeader()}
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Sparkles className="h-16 w-16 text-muted-foreground/50 mb-4" />
            <h3 className="font-medium text-lg mb-2">暂无推荐</h3>
            <p className="text-muted-foreground text-center mb-4">
              我们正在为你寻找有趣的用户，稍后再来看看吧
            </p>
            <Button onClick={handleRefreshSuggestions} variant="outline">
              刷新推荐
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }
  
  return (
    <div className={className}>
      {renderHeader()}
      
      {/* Last refresh info */}
      {localState.lastRefresh && (
        <div className="flex items-center gap-2 text-xs text-muted-foreground mb-4">
          <Clock className="h-3 w-3" />
          <span>
            推荐更新于 {new Date(localState.lastRefresh).toLocaleTimeString()}
          </span>
        </div>
      )}
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredSuggestions.map(renderSuggestionCard)}
      </div>
    </div>
  )
}

/**
 * Compact UserSuggestions for sidebar or smaller spaces
 */
export function CompactUserSuggestions({
  limit = 3,
  className,
  onUserFollow,
}: Pick<UserSuggestionsProps, 'limit' | 'className' | 'onUserFollow'>) {
  return (
    <UserSuggestions
      limit={limit}
      show_reason={false}
      show_mutual={false}
      show_refresh={false}
      className={className}
      onUserFollow={onUserFollow}
    />
  )
}

export default UserSuggestions