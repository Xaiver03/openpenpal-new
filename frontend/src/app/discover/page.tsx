/**
 * User Discovery Page - Find and Connect with Other Users
 * 用户发现页面 - 寻找并连接其他用户
 */

'use client'

import React, { useState, useEffect, useCallback } from 'react'
import { Search, Filter, Users, TrendingUp, Award, MapPin, Loader2 } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { UserLevelDisplay } from '@/components/user/level-badge'
import { FollowButton } from '@/components/follow/follow-button'
import { cn } from '@/lib/utils'
import { useDebounce } from '@/hooks/use-debounce'
import Link from 'next/link'

interface UserSearchResult {
  id: string
  username: string
  nickname?: string
  avatar_url?: string
  bio?: string
  school?: string
  op_code?: string
  writing_level?: number
  courier_level?: number
  follower_count: number
  following_count: number
  is_following: boolean
  match_score?: number
  common_interests?: string[]
}

interface UserRecommendation extends UserSearchResult {
  reason: string
}

// Mock data
const mockRecommendations: UserRecommendation[] = [
  {
    id: '1',
    username: 'alice',
    nickname: 'Alice',
    bio: '喜欢写信，热爱生活',
    school: '北京大学',
    writing_level: 3,
    follower_count: 156,
    following_count: 89,
    is_following: false,
    reason: '同校推荐',
    common_interests: ['写作', '摄影']
  },
  {
    id: '2',
    username: 'bob',
    nickname: 'Bob',
    bio: '用文字记录生活的点滴',
    school: '清华大学',
    courier_level: 2,
    follower_count: 234,
    following_count: 123,
    is_following: false,
    reason: '活跃信使',
    common_interests: ['旅行', '音乐']
  },
  {
    id: '3',
    username: 'charlie',
    nickname: 'Charlie',
    bio: '每一封信都是一次心灵的旅行',
    writing_level: 4,
    follower_count: 567,
    following_count: 234,
    is_following: true,
    reason: '优秀写作者',
    common_interests: ['文学', '历史']
  }
]

const mockTrendingUsers: UserSearchResult[] = [
  {
    id: '4',
    username: 'david',
    nickname: 'David',
    bio: '分享日常，传递温暖',
    follower_count: 890,
    following_count: 345,
    is_following: false,
    match_score: 95
  },
  {
    id: '5',
    username: 'emma',
    nickname: 'Emma',
    bio: '用心写每一个字',
    writing_level: 5,
    follower_count: 1234,
    following_count: 456,
    is_following: false,
    match_score: 92
  }
]

export default function DiscoverPage() {
  const [activeTab, setActiveTab] = useState<'recommended' | 'trending' | 'search'>('recommended')
  const [searchQuery, setSearchQuery] = useState('')
  const [sortBy, setSortBy] = useState<'relevance' | 'followers' | 'activity'>('relevance')
  const [filterSchool, setFilterSchool] = useState<string>('all')
  const [recommendations, setRecommendations] = useState<UserRecommendation[]>([])
  const [trendingUsers, setTrendingUsers] = useState<UserSearchResult[]>([])
  const [searchResults, setSearchResults] = useState<UserSearchResult[]>([])
  const [loading, setLoading] = useState(false)
  
  const debouncedSearchQuery = useDebounce(searchQuery, 500)

  // Load initial data
  useEffect(() => {
    loadRecommendations()
    loadTrendingUsers()
  }, [])

  // Search when query changes
  useEffect(() => {
    if (debouncedSearchQuery) {
      searchUsers(debouncedSearchQuery)
    } else {
      setSearchResults([])
    }
  }, [debouncedSearchQuery])

  const loadRecommendations = async () => {
    setLoading(true)
    try {
      // TODO: Replace with real API
      await new Promise(resolve => setTimeout(resolve, 500))
      setRecommendations(mockRecommendations)
    } catch (error) {
      console.error('Failed to load recommendations:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadTrendingUsers = async () => {
    try {
      // TODO: Replace with real API
      await new Promise(resolve => setTimeout(resolve, 500))
      setTrendingUsers(mockTrendingUsers)
    } catch (error) {
      console.error('Failed to load trending users:', error)
    }
  }

  const searchUsers = async (query: string) => {
    setLoading(true)
    try {
      // TODO: Replace with real API
      await new Promise(resolve => setTimeout(resolve, 500))
      // Mock search - filter recommendations
      const results = [...mockRecommendations, ...mockTrendingUsers].filter(user =>
        user.username.toLowerCase().includes(query.toLowerCase()) ||
        user.nickname?.toLowerCase().includes(query.toLowerCase()) ||
        user.bio?.toLowerCase().includes(query.toLowerCase())
      )
      setSearchResults(results)
      if (results.length > 0) {
        setActiveTab('search')
      }
    } catch (error) {
      console.error('Failed to search users:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleFollowChange = useCallback((userId: string, isFollowing: boolean) => {
    // Update local state for all user lists
    const updateRecommendation = (user: UserRecommendation) => 
      user.id === userId 
        ? { ...user, is_following: isFollowing, follower_count: user.follower_count + (isFollowing ? 1 : -1) }
        : user
        
    const updateUser = (user: UserSearchResult) => 
      user.id === userId 
        ? { ...user, is_following: isFollowing, follower_count: user.follower_count + (isFollowing ? 1 : -1) }
        : user

    setRecommendations(prev => prev.map(updateRecommendation))
    setTrendingUsers(prev => prev.map(updateUser))
    setSearchResults(prev => prev.map(updateUser))
  }, [])

  const UserCard = ({ user, showReason = false }: { user: UserSearchResult | UserRecommendation, showReason?: boolean }) => (
    <Card className="hover:shadow-md transition-shadow">
      <CardContent className="p-6">
        <div className="flex items-start gap-4">
          <Link href={`/u/${user.username}`}>
            <Avatar className="h-12 w-12 cursor-pointer">
              <AvatarImage src={user.avatar_url} />
              <AvatarFallback>
                {user.nickname?.charAt(0) || user.username.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
          </Link>
          
          <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between gap-2 mb-1">
              <Link href={`/u/${user.username}`} className="hover:underline">
                <h3 className="font-semibold truncate">
                  {user.nickname || user.username}
                </h3>
              </Link>
              <FollowButton
                user_id={user.id}
                initial_is_following={user.is_following}
                initial_follower_count={user.follower_count}
                size="sm"
                onFollowChange={(isFollowing) => handleFollowChange(user.id, isFollowing)}
              />
            </div>
            
            <p className="text-sm text-muted-foreground mb-2">@{user.username}</p>
            
            {user.bio && (
              <p className="text-sm text-gray-600 mb-3 line-clamp-2">{user.bio}</p>
            )}
            
            <div className="flex items-center gap-3 text-xs text-muted-foreground">
              {(user.writing_level || user.courier_level) && (
                <UserLevelDisplay 
                  writingLevel={user.writing_level}
                  courierLevel={user.courier_level}
                  compact
                />
              )}
              
              {user.school && (
                <div className="flex items-center gap-1">
                  <MapPin className="h-3 w-3" />
                  <span>{user.school}</span>
                </div>
              )}
              
              <span>{user.follower_count} 粉丝</span>
              <span>{user.following_count} 关注</span>
            </div>
            
            {showReason && 'reason' in user && (
              <div className="mt-3">
                <Badge variant="secondary" className="text-xs">
                  {user.reason}
                </Badge>
              </div>
            )}
            
            {'common_interests' in user && user.common_interests && user.common_interests.length > 0 && (
              <div className="flex gap-1 mt-2">
                {user.common_interests.map(interest => (
                  <Badge key={interest} variant="outline" className="text-xs">
                    {interest}
                  </Badge>
                ))}
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )

  return (
    <div className="container max-w-6xl mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">发现用户</h1>
        <p className="text-muted-foreground">寻找志同道合的笔友，开启新的书信之旅</p>
      </div>

      {/* Search Bar */}
      <div className="mb-6">
        <div className="flex gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="搜索用户名、昵称或简介..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
            {loading && (
              <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin" />
            )}
          </div>
          
          <Select value={sortBy} onValueChange={(value: any) => setSortBy(value)}>
            <SelectTrigger className="w-[140px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="relevance">相关度</SelectItem>
              <SelectItem value="followers">粉丝数</SelectItem>
              <SelectItem value="activity">活跃度</SelectItem>
            </SelectContent>
          </Select>
          
          <Select value={filterSchool} onValueChange={setFilterSchool}>
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder="所有学校" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">所有学校</SelectItem>
              <SelectItem value="pku">北京大学</SelectItem>
              <SelectItem value="thu">清华大学</SelectItem>
              <SelectItem value="bjtu">北京交通大学</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={(value: any) => setActiveTab(value)}>
        <TabsList className="mb-6">
          <TabsTrigger value="recommended" className="gap-2">
            <Users className="h-4 w-4" />
            为你推荐
          </TabsTrigger>
          <TabsTrigger value="trending" className="gap-2">
            <TrendingUp className="h-4 w-4" />
            热门用户
          </TabsTrigger>
          {searchResults.length > 0 && (
            <TabsTrigger value="search" className="gap-2">
              <Search className="h-4 w-4" />
              搜索结果 ({searchResults.length})
            </TabsTrigger>
          )}
        </TabsList>

        <TabsContent value="recommended">
          <div className="grid gap-4 md:grid-cols-2">
            {recommendations.map(user => (
              <UserCard key={user.id} user={user} showReason />
            ))}
          </div>
          
          {recommendations.length === 0 && !loading && (
            <Card>
              <CardContent className="p-12 text-center">
                <Users className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <p className="text-muted-foreground">暂无推荐用户</p>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="trending">
          <div className="grid gap-4 md:grid-cols-2">
            {trendingUsers.map((user, index) => (
              <div key={user.id} className="relative">
                <div className="absolute -left-3 -top-3 z-10">
                  <Badge className="bg-gradient-to-r from-yellow-400 to-orange-400 text-white">
                    #{index + 1}
                  </Badge>
                </div>
                <UserCard user={user} />
              </div>
            ))}
          </div>
          
          {trendingUsers.length === 0 && !loading && (
            <Card>
              <CardContent className="p-12 text-center">
                <TrendingUp className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <p className="text-muted-foreground">暂无热门用户</p>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="search">
          <div className="grid gap-4 md:grid-cols-2">
            {searchResults.map(user => (
              <UserCard key={user.id} user={user} />
            ))}
          </div>
          
          {searchResults.length === 0 && searchQuery && !loading && (
            <Card>
              <CardContent className="p-12 text-center">
                <Search className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <p className="text-muted-foreground">未找到匹配的用户</p>
                <p className="text-sm text-muted-foreground mt-2">
                  试试其他关键词或筛选条件
                </p>
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}