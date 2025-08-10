'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  TrendingUp,
  Heart,
  Eye,
  MessageSquare,
  Calendar,
  User,
  Clock,
  Star,
  Award,
  Flame,
  Filter,
  ArrowUp,
  Trophy
} from 'lucide-react'
import { museumService } from '@/lib/services/museum-service'
import { formatDate, formatRelativeTime } from '@/lib/utils'

interface PopularEntry {
  id: string
  letter_id: string
  title: string
  excerpt: string
  author: string
  exhibition_id: string
  exhibition_name: string
  created_at: string
  updated_at: string
  views: number
  likes: number
  shares: number
  comments: number
  trending_score: number
  category: string
  tags: string[]
  featured: boolean
}

interface TimeRange {
  id: string
  label: string
  value: string
  icon: React.ElementType
}

const timeRanges: TimeRange[] = [
  { id: 'today', label: '今日', value: '1d', icon: Flame },
  { id: 'week', label: '本周', value: '7d', icon: TrendingUp },
  { id: 'month', label: '本月', value: '30d', icon: Calendar },
  { id: 'all', label: '全部', value: 'all', icon: Star }
]

export default function MuseumPopularPage() {
  const router = useRouter()
  const [entries, setEntries] = useState<PopularEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedTimeRange, setSelectedTimeRange] = useState('week')
  const [sortBy, setSortBy] = useState<'trending' | 'likes' | 'views'>('trending')

  useEffect(() => {
    fetchPopularEntries()
  }, [selectedTimeRange, sortBy])

  const fetchPopularEntries = async () => {
    setLoading(true)
    setError(null)

    try {
      const params = {
        time_range: timeRanges.find(r => r.id === selectedTimeRange)?.value || '7d',
        sort_by: sortBy,
        limit: 20
      }

      const response = await museumService.getPopularMuseumEntries(params)
      
      // 模拟数据转换，实际应该从后端返回
      const formattedEntries: PopularEntry[] = (response.data || []).map((entry: any, index: number) => ({
        ...entry,
        views: Math.floor(Math.random() * 10000) + 1000,
        likes: Math.floor(Math.random() * 1000) + 100,
        shares: Math.floor(Math.random() * 500) + 50,
        comments: Math.floor(Math.random() * 200) + 20,
        trending_score: Math.floor(Math.random() * 100) + 50,
        category: entry.category || 'story',
        tags: entry.tags || ['感动', '回忆', '青春'],
        featured: index < 3
      }))

      setEntries(formattedEntries)
    } catch (err) {
      console.error('获取热门条目失败:', err)
      setError('获取热门条目失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  const getRankIcon = (rank: number) => {
    switch (rank) {
      case 1:
        return <Trophy className="w-5 h-5 text-yellow-500" />
      case 2:
        return <Award className="w-5 h-5 text-gray-400" />
      case 3:
        return <Award className="w-5 h-5 text-amber-600" />
      default:
        return <span className="w-5 h-5 text-center text-sm font-bold text-muted-foreground">{rank}</span>
    }
  }

  const getTrendingBadge = (score: number) => {
    if (score >= 90) return { label: '🔥 爆款', color: 'bg-red-500' }
    if (score >= 70) return { label: '📈 热门', color: 'bg-orange-500' }
    if (score >= 50) return { label: '✨ 上升', color: 'bg-yellow-500' }
    return null
  }

  return (
    <div className="container max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-letter-ink mb-2">
          热门信件榜
        </h1>
        <p className="text-muted-foreground">
          发现最受欢迎的信件，感受文字的力量
        </p>
      </div>

      {/* Time Range Tabs */}
      <Tabs value={selectedTimeRange} onValueChange={setSelectedTimeRange} className="mb-6">
        <TabsList className="grid grid-cols-4 w-full max-w-md">
          {timeRanges.map(range => {
            const Icon = range.icon
            return (
              <TabsTrigger key={range.id} value={range.id} className="flex items-center gap-1">
                <Icon className="w-4 h-4" />
                {range.label}
              </TabsTrigger>
            )
          })}
        </TabsList>
      </Tabs>

      {/* Sort Options */}
      <div className="flex items-center gap-4 mb-6">
        <span className="text-sm text-muted-foreground">排序方式：</span>
        <div className="flex gap-2">
          <Button
            variant={sortBy === 'trending' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setSortBy('trending')}
          >
            <TrendingUp className="w-4 h-4 mr-1" />
            热度
          </Button>
          <Button
            variant={sortBy === 'likes' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setSortBy('likes')}
          >
            <Heart className="w-4 h-4 mr-1" />
            点赞
          </Button>
          <Button
            variant={sortBy === 'views' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setSortBy('views')}
          >
            <Eye className="w-4 h-4 mr-1" />
            浏览
          </Button>
        </div>
      </div>

      {/* Error State */}
      {error && (
        <Alert variant="destructive" className="mb-6">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Loading State */}
      {loading && (
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader>
                <div className="h-6 bg-muted rounded w-3/4"></div>
                <div className="h-4 bg-muted rounded w-1/2 mt-2"></div>
              </CardHeader>
              <CardContent>
                <div className="h-20 bg-muted rounded"></div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Entries List */}
      {!loading && entries.length > 0 && (
        <div className="space-y-4">
          {entries.map((entry, index) => {
            const rank = index + 1
            const trendingBadge = getTrendingBadge(entry.trending_score)
            
            return (
              <Card 
                key={entry.id} 
                className={`cursor-pointer transition-all hover:shadow-lg ${
                  entry.featured ? 'border-primary ring-2 ring-primary/20' : ''
                }`}
                onClick={() => router.push(`/museum/entries/${entry.id}`)}
              >
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      {/* Rank */}
                      <div className="flex items-center justify-center w-10 h-10 rounded-full bg-muted">
                        {getRankIcon(rank)}
                      </div>
                      
                      {/* Title and Meta */}
                      <div className="flex-1">
                        <CardTitle className="font-serif text-lg line-clamp-1">
                          {entry.title}
                        </CardTitle>
                        <CardDescription className="flex items-center gap-2 mt-1">
                          <User className="w-3 h-3" />
                          <span>{entry.author}</span>
                          <span className="text-muted-foreground">·</span>
                          <Clock className="w-3 h-3" />
                          <span>{formatRelativeTime(entry.created_at)}</span>
                        </CardDescription>
                      </div>
                    </div>

                    {/* Trending Badge */}
                    {trendingBadge && (
                      <Badge className={`${trendingBadge.color} text-white`}>
                        {trendingBadge.label}
                      </Badge>
                    )}
                  </div>
                </CardHeader>

                <CardContent>
                  {/* Excerpt */}
                  <p className="text-sm text-muted-foreground line-clamp-2 mb-4">
                    {entry.excerpt}
                  </p>

                  {/* Stats */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4 text-sm text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <Eye className="w-4 h-4" />
                        {entry.views.toLocaleString()}
                      </span>
                      <span className="flex items-center gap-1">
                        <Heart className="w-4 h-4" />
                        {entry.likes.toLocaleString()}
                      </span>
                      <span className="flex items-center gap-1">
                        <MessageSquare className="w-4 h-4" />
                        {entry.comments}
                      </span>
                    </div>

                    {/* Tags */}
                    <div className="flex gap-1">
                      {entry.tags.slice(0, 2).map(tag => (
                        <Badge key={tag} variant="secondary" className="text-xs">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  </div>

                  {/* Trending Score Indicator */}
                  {rank <= 10 && (
                    <div className="mt-3 pt-3 border-t">
                      <div className="flex items-center justify-between text-xs text-muted-foreground">
                        <span>热度指数</span>
                        <span className="flex items-center gap-1">
                          <ArrowUp className="w-3 h-3 text-green-500" />
                          {entry.trending_score}%
                        </span>
                      </div>
                      <div className="mt-1 h-2 bg-muted rounded-full overflow-hidden">
                        <div 
                          className="h-full bg-gradient-to-r from-yellow-500 to-red-500 transition-all"
                          style={{ width: `${entry.trending_score}%` }}
                        />
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}

      {/* Empty State */}
      {!loading && entries.length === 0 && (
        <Card className="text-center py-12">
          <CardContent>
            <TrendingUp className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
            <p className="text-muted-foreground">暂无热门信件</p>
          </CardContent>
        </Card>
      )}

      {/* Load More */}
      {!loading && entries.length >= 20 && (
        <div className="text-center mt-8">
          <Button variant="outline" onClick={fetchPopularEntries}>
            加载更多
          </Button>
        </div>
      )}
    </div>
  )
}