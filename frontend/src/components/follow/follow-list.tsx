/**
 * FollowList Component - SOTA Implementation
 * 关注列表组件 - 支持分页、搜索、过滤、无限滚动
 */

'use client'

import React from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { 
  Search,
  Filter,
  Users,
  UserPlus,
  Loader2,
  RefreshCw,
  AlertCircle,
  ChevronDown,
  SortAsc,
  SortDesc
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useFollow } from '@/stores/follow-store'
import { UserCard, UserCardSkeleton } from './user-card'
import type { FollowListProps, FollowListQuery, FollowUser } from '@/types/follow'

interface FollowListState {
  searchQuery: string
  sortBy: FollowListQuery['sort_by']
  sortOrder: FollowListQuery['order']
  currentPage: number
  hasNextPage: boolean
  isLoadingMore: boolean
}

export function FollowList({
  user_id,
  type,
  initial_data = [],
  show_stats = true,
  enable_search = true,
  enable_filters = true,
  max_height,
  className,
}: FollowListProps) {
  const { 
    followers, 
    following, 
    isLoading, 
    errors,
    followerCount,
    followingCount,
    loadFollowers,
    loadFollowing
  } = useFollow()
  
  const [localState, setLocalState] = React.useState<FollowListState>({
    searchQuery: '',
    sortBy: 'created_at',
    sortOrder: 'desc',
    currentPage: 1,
    hasNextPage: true,
    isLoadingMore: false,
  })
  
  const [displayUsers, setDisplayUsers] = React.useState<FollowUser[]>(initial_data)
  
  // Get data based on type
  const users = type === 'followers' ? followers : following
  const loading = type === 'followers' ? isLoading.followers : isLoading.following
  const error = type === 'followers' ? errors.followers : errors.following
  const totalCount = type === 'followers' ? followerCount : followingCount
  
  // Update display users when store data changes
  React.useEffect(() => {
    if (users.length > 0) {
      setDisplayUsers(users)
    }
  }, [users])
  
  // Load initial data
  React.useEffect(() => {
    if (users.length === 0 && !loading) {
      handleLoadData()
    }
  }, [type, user_id])
  
  const handleLoadData = async (page = 1, append = false) => {
    const query: FollowListQuery = {
      page,
      limit: 20,
      sort_by: localState.sortBy,
      order: localState.sortOrder,
    }
    
    if (localState.searchQuery) {
      query.search = localState.searchQuery
    }
    
    try {
      if (type === 'followers') {
        await loadFollowers(user_id, query)
      } else {
        await loadFollowing(user_id, query)
      }
      
      setLocalState(prev => ({
        ...prev,
        currentPage: page,
        hasNextPage: users.length >= (query.limit || 20),
      }))
    } catch (error) {
      console.error('Failed to load data:', error)
    }
  }
  
  const handleSearch = (query: string) => {
    setLocalState(prev => ({ ...prev, searchQuery: query, currentPage: 1 }))
    
    // Debounced search
    const timeoutId = setTimeout(() => {
      handleLoadData(1)
    }, 500)
    
    return () => clearTimeout(timeoutId)
  }
  
  const handleSort = (sortBy: FollowListQuery['sort_by']) => {
    const newOrder = localState.sortBy === sortBy && localState.sortOrder === 'desc' ? 'asc' : 'desc'
    setLocalState(prev => ({
      ...prev,
      sortBy,
      sortOrder: newOrder,
      currentPage: 1,
    }))
    handleLoadData(1)
  }
  
  const handleLoadMore = () => {
    if (!localState.hasNextPage || localState.isLoadingMore) return
    
    setLocalState(prev => ({ ...prev, isLoadingMore: true }))
    handleLoadData(localState.currentPage + 1, true).finally(() => {
      setLocalState(prev => ({ ...prev, isLoadingMore: false }))
    })
  }
  
  const handleRefresh = () => {
    setLocalState(prev => ({ ...prev, currentPage: 1 }))
    handleLoadData(1)
  }
  
  const filteredUsers = displayUsers.filter(user => {
    if (!localState.searchQuery) return true
    
    const query = localState.searchQuery.toLowerCase()
    return (
      user.nickname.toLowerCase().includes(query) ||
      user.username.toLowerCase().includes(query) ||
      user.school_name?.toLowerCase().includes(query)
    )
  })
  
  const renderHeader = () => (
    <CardHeader className="pb-3">
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <CardTitle className="text-lg font-medium flex items-center gap-2">
            {type === 'followers' ? (
              <>
                <Users className="h-5 w-5" />
                关注者
              </>
            ) : (
              <>
                <UserPlus className="h-5 w-5" />
                关注中
              </>
            )}
            {show_stats && (
              <Badge variant="secondary" className="ml-2">
                {totalCount}
              </Badge>
            )}
          </CardTitle>
          {show_stats && (
            <CardDescription>
              {type === 'followers' ? '关注你的用户' : '你关注的用户'}
            </CardDescription>
          )}
        </div>
        
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={loading}
          >
            {loading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="h-4 w-4" />
            )}
          </Button>
        </div>
      </div>
    </CardHeader>
  )
  
  const renderSearchAndFilters = () => (
    <div className="px-6 pb-4 space-y-3">
      {/* Search */}
      {enable_search && (
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="搜索用户..."
            value={localState.searchQuery}
            onChange={(e) => handleSearch(e.target.value)}
            className="pl-10"
          />
        </div>
      )}
      
      {/* Filters and Sort */}
      {enable_filters && (
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => handleSort('created_at')}
            className="gap-1"
          >
            按时间
            {localState.sortBy === 'created_at' && (
              localState.sortOrder === 'desc' ? (
                <SortDesc className="h-3 w-3" />
              ) : (
                <SortAsc className="h-3 w-3" />
              )
            )}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => handleSort('nickname')}
            className="gap-1"
          >
            按姓名
            {localState.sortBy === 'nickname' && (
              localState.sortOrder === 'desc' ? (
                <SortDesc className="h-3 w-3" />
              ) : (
                <SortAsc className="h-3 w-3" />
              )
            )}
          </Button>
          {type === 'following' && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleSort('letters_count')}
              className="gap-1"
            >
              按活跃度
              {localState.sortBy === 'letters_count' && (
                localState.sortOrder === 'desc' ? (
                  <SortDesc className="h-3 w-3" />
                ) : (
                  <SortAsc className="h-3 w-3" />
                )
              )}
            </Button>
          )}
        </div>
      )}
    </div>
  )
  
  const renderUserList = () => {
    if (loading && displayUsers.length === 0) {
      return (
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <UserCardSkeleton key={i} variant="compact" />
          ))}
        </div>
      )
    }
    
    if (error) {
      return (
        <div className="flex flex-col items-center justify-center py-8 text-center">
          <AlertCircle className="h-12 w-12 text-red-500 mb-4" />
          <h3 className="font-medium text-lg mb-2">加载失败</h3>
          <p className="text-muted-foreground mb-4">{error}</p>
          <Button onClick={handleRefresh} variant="outline">
            重试
          </Button>
        </div>
      )
    }
    
    if (filteredUsers.length === 0) {
      const emptyMessage = localState.searchQuery 
        ? '没有找到匹配的用户'
        : type === 'followers' 
          ? '还没有关注者' 
          : '还没有关注任何人'
      
      const emptyDescription = localState.searchQuery
        ? '尝试使用不同的搜索词'
        : type === 'followers'
          ? '当有人关注你时，他们会出现在这里'
          : '关注其他用户来建立连接'
      
      return (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Users className="h-16 w-16 text-muted-foreground/50 mb-4" />
          <h3 className="font-medium text-lg mb-2">{emptyMessage}</h3>
          <p className="text-muted-foreground text-sm">{emptyDescription}</p>
        </div>
      )
    }
    
    return (
      <div className="space-y-3">
        {filteredUsers.map((user) => (
          <UserCard
            key={user.id}
            user={user}
            variant="compact"
            show_follow_button={type === 'followers' || user.id !== user_id}
            show_stats={true}
            show_mutual={true}
            onFollowChange={(isFollowing, followerCount) => {
              // Update local state optimistically
              setDisplayUsers(prev => 
                prev.map(u => 
                  u.id === user.id 
                    ? { ...u, is_following: isFollowing, followers_count: followerCount }
                    : u
                )
              )
            }}
          />
        ))}
        
        {/* Load More */}
        {localState.hasNextPage && (
          <div className="flex justify-center pt-4">
            <Button
              variant="outline"
              onClick={handleLoadMore}
              disabled={localState.isLoadingMore}
            >
              {localState.isLoadingMore ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  加载中...
                </>
              ) : (
                <>
                  <ChevronDown className="h-4 w-4 mr-2" />
                  加载更多
                </>
              )}
            </Button>
          </div>
        )}
      </div>
    )
  }
  
  return (
    <Card className={cn('w-full', className)}>
      {renderHeader()}
      {renderSearchAndFilters()}
      
      <CardContent 
        className="pt-0"
        style={{ maxHeight: max_height }}
      >
        <div className={cn(
          'overflow-y-auto',
          max_height && 'pr-2'
        )}>
          {renderUserList()}
        </div>
      </CardContent>
    </Card>
  )
}

/**
 * FollowListTabs - Tabbed view for both followers and following
 */
export function FollowListTabs({
  user_id,
  initial_followers_data = [],
  initial_following_data = [],
  show_stats = true,
  enable_search = true,
  enable_filters = true,
  className,
}: {
  user_id?: string
  initial_followers_data?: FollowUser[]
  initial_following_data?: FollowUser[]
  show_stats?: boolean
  enable_search?: boolean
  enable_filters?: boolean
  className?: string
}) {
  const { followerCount, followingCount } = useFollow()
  
  return (
    <Tabs defaultValue="followers" className={className}>
      <TabsList className="grid w-full grid-cols-2">
        <TabsTrigger value="followers" className="flex items-center gap-2">
          <Users className="h-4 w-4" />
          关注者
          {show_stats && (
            <Badge variant="secondary" className="ml-1">
              {followerCount}
            </Badge>
          )}
        </TabsTrigger>
        <TabsTrigger value="following" className="flex items-center gap-2">
          <UserPlus className="h-4 w-4" />
          关注中
          {show_stats && (
            <Badge variant="secondary" className="ml-1">
              {followingCount}
            </Badge>
          )}
        </TabsTrigger>
      </TabsList>
      
      <TabsContent value="followers">
        <FollowList
          user_id={user_id}
          type="followers"
          initial_data={initial_followers_data}
          show_stats={show_stats}
          enable_search={enable_search}
          enable_filters={enable_filters}
        />
      </TabsContent>
      
      <TabsContent value="following">
        <FollowList
          user_id={user_id}
          type="following"
          initial_data={initial_following_data}
          show_stats={show_stats}
          enable_search={enable_search}
          enable_filters={enable_filters}
        />
      </TabsContent>
    </Tabs>
  )
}

export default FollowList