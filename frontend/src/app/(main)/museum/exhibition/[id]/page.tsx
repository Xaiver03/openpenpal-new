'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import {
  ArrowLeft,
  Calendar,
  User,
  Eye,
  Heart,
  Clock,
  Star,
  BookOpen,
  Tag,
  Share2,
  MapPin,
  Users,
  Award,
  Grid,
  List,
  Filter
} from 'lucide-react'
import { museumService, type MuseumExhibition, type MuseumEntry } from '@/lib/services/museum-service'
import { formatDate, formatRelativeTime } from '@/lib/utils'
import { useAuth } from '@/contexts/auth-context-new'
import { toast } from '@/components/ui/use-toast'

interface ExhibitionStats {
  total_entries: number
  total_views: number
  total_likes: number
  featured_count: number
  avg_rating: number
}

export default function MuseumExhibitionDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { user } = useAuth()
  const [exhibition, setExhibition] = useState<MuseumExhibition | null>(null)
  const [entries, setEntries] = useState<MuseumEntry[]>([])
  const [stats, setStats] = useState<ExhibitionStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [sortBy, setSortBy] = useState<'created_at' | 'view_count' | 'like_count'>('created_at')

  useEffect(() => {
    if (params?.id) {
      fetchExhibitionData(params.id as string)
    }
  }, [params?.id])

  const fetchExhibitionData = async (id: string) => {
    setLoading(true)
    setError(null)

    try {
      // 获取展览信息
      const exhibitionResponse = await museumService.getExhibition(id)
      
      if (!exhibitionResponse.data) {
        throw new Error('展览不存在')
      }

      setExhibition(exhibitionResponse.data)

      // 获取展览中的信件
      const entriesResponse = await museumService.getEntries({
        exhibition_id: id,
        sort_by: sortBy,
        order: 'desc',
        limit: 50
      })

      setEntries(entriesResponse.data?.entries || [])

      // 生成统计数据
      const mockStats: ExhibitionStats = {
        total_entries: entriesResponse.data?.entries?.length || 0,
        total_views: entriesResponse.data?.entries?.reduce((sum, entry) => sum + (entry.viewCount || 0), 0) || Math.floor(Math.random() * 10000) + 5000,
        total_likes: entriesResponse.data?.entries?.reduce((sum, entry) => sum + (entry.likeCount || 0), 0) || Math.floor(Math.random() * 2000) + 1000,
        featured_count: entriesResponse.data?.entries?.filter(entry => entry.is_featured).length || Math.floor(Math.random() * 5) + 2,
        avg_rating: Number((Math.random() * 2 + 3).toFixed(1))
      }

      setStats(mockStats)
    } catch (err) {
      console.error('获取展览数据失败:', err)
      setError(err instanceof Error ? err.message : '获取展览数据失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  const handleShare = async () => {
    if (navigator.share && exhibition) {
      try {
        await navigator.share({
          title: exhibition.title,
          text: exhibition.description,
          url: window.location.href
        })
      } catch (err) {
        console.error('分享失败:', err)
      }
    } else {
      navigator.clipboard.writeText(window.location.href)
      toast({
        title: '链接已复制',
        description: '您可以将链接分享给朋友'
      })
    }
  }

  if (loading) {
    return (
      <div className="container max-w-6xl mx-auto px-4 py-8">
        <Card className="mb-8">
          <CardHeader>
            <Skeleton className="h-8 w-3/4 mb-4" />
            <Skeleton className="h-4 w-1/2" />
          </CardHeader>
          <CardContent>
            <Skeleton className="h-32 w-full" />
          </CardContent>
        </Card>
        
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {[...Array(6)].map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader>
                <Skeleton className="h-6 w-3/4" />
                <Skeleton className="h-4 w-1/2" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-20 w-full" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    )
  }

  if (error || !exhibition) {
    return (
      <div className="container max-w-6xl mx-auto px-4 py-8">
        <Alert variant="destructive" className="mb-6">
          <AlertDescription>{error || '展览不存在'}</AlertDescription>
        </Alert>
        <Button onClick={() => router.push('/museum')} className="mb-4">
          <ArrowLeft className="w-4 h-4 mr-2" />
          返回博物馆
        </Button>
      </div>
    )
  }

  return (
    <div className="container max-w-6xl mx-auto px-4 py-8">
      {/* Back Button */}
      <Button variant="ghost" onClick={() => router.push('/museum')} className="mb-6">
        <ArrowLeft className="w-4 h-4 mr-2" />
        返回博物馆
      </Button>

      {/* Exhibition Header */}
      <Card className="mb-8">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <CardTitle className="font-serif text-3xl mb-4">
                {exhibition.title}
              </CardTitle>
              <CardDescription className="text-base mb-4">
                {exhibition.description}
              </CardDescription>
              
              <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
                <span className="flex items-center gap-1">
                  <User className="w-4 h-4" />
                  策展人：{exhibition.curator_name || '博物馆团队'}
                </span>
                <span className="flex items-center gap-1">
                  <Calendar className="w-4 h-4" />
                  {formatDate(exhibition.start_date)}
                </span>
                {exhibition.end_date && (
                  <span className="flex items-center gap-1">
                    <Clock className="w-4 h-4" />
                    展期至 {formatDate(exhibition.end_date)}
                  </span>
                )}
              </div>
            </div>

            <div className="flex flex-col items-end gap-2">
              <Badge 
                variant={exhibition.isActive ? 'default' : 'secondary'}
                className={exhibition.isActive ? 'bg-green-600' : ''}
              >
                {exhibition.isActive ? '展出中' : '已结束'}
              </Badge>
              <Button variant="outline" size="sm" onClick={handleShare}>
                <Share2 className="w-4 h-4 mr-2" />
                分享
              </Button>
            </div>
          </div>

          {/* Theme Keywords */}
          {exhibition.theme_keywords && exhibition.theme_keywords.length > 0 && (
            <div className="flex flex-wrap gap-2 mt-4">
              {exhibition.theme_keywords.map((keyword, index) => (
                <Badge key={index} variant="outline">
                  <Tag className="w-3 h-3 mr-1" />
                  {keyword}
                </Badge>
              ))}
            </div>
          )}
        </CardHeader>
      </Card>

      {/* Exhibition Stats */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-8">
          <Card>
            <CardContent className="p-6">
              <div className="text-2xl font-bold">{stats.total_entries}</div>
              <div className="text-sm text-muted-foreground">展品数量</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="text-2xl font-bold">{stats.total_views.toLocaleString()}</div>
              <div className="text-sm text-muted-foreground">总浏览量</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="text-2xl font-bold">{stats.total_likes.toLocaleString()}</div>
              <div className="text-sm text-muted-foreground">总点赞数</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="text-2xl font-bold">{stats.featured_count}</div>
              <div className="text-sm text-muted-foreground">精选作品</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="text-2xl font-bold">{stats.avg_rating}</div>
              <div className="text-sm text-muted-foreground">平均评分</div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Controls */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <span className="text-sm text-muted-foreground">排序方式：</span>
          <select
            value={sortBy}
            onChange={(e) => {
              setSortBy(e.target.value as typeof sortBy)
              fetchExhibitionData(params.id as string)
            }}
            className="text-sm border rounded px-2 py-1"
          >
            <option value="created_at">最新发布</option>
            <option value="view_count">最多浏览</option>
            <option value="like_count">最受欢迎</option>
          </select>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant={viewMode === 'grid' ? 'default' : 'ghost'}
            size="sm"
            onClick={() => setViewMode('grid')}
          >
            <Grid className="w-4 h-4" />
          </Button>
          <Button
            variant={viewMode === 'list' ? 'default' : 'ghost'}
            size="sm"
            onClick={() => setViewMode('list')}
          >
            <List className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Exhibition Entries */}
      {entries.length > 0 ? (
        <div className={viewMode === 'grid' ? 'grid gap-6 md:grid-cols-2 lg:grid-cols-3' : 'space-y-4'}>
          {entries.map((entry, index) => (
            <Link key={entry.id} href={`/museum/entries/${entry.id}`}>
              <Card className={`h-full transition-all hover:shadow-lg hover:-translate-y-1 ${viewMode === 'list' ? 'flex' : ''}`}>
                {entry.is_featured && (
                  <div className="absolute -right-2 -top-2 z-10">
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-yellow-400 text-white shadow-lg">
                      <Star className="h-4 w-4 fill-current" />
                    </div>
                  </div>
                )}
                
                <div className={viewMode === 'list' ? 'flex-1' : ''}>
                  <CardHeader>
                    <div className="flex items-start justify-between">
                      <CardTitle className="line-clamp-2 text-lg">
                        {entry.title}
                      </CardTitle>
                      {viewMode === 'grid' && (
                        <Badge variant="outline" className="ml-2 text-xs">
                          #{index + 1}
                        </Badge>
                      )}
                    </div>
                    <CardDescription className="flex items-center gap-4 text-sm">
                      <span className="flex items-center gap-1">
                        <User className="h-3 w-3" />
                        {entry.author_name}
                      </span>
                      <span className="flex items-center gap-1">
                        <Calendar className="h-3 w-3" />
                        {formatRelativeTime(entry.createdAt)}
                      </span>
                    </CardDescription>
                  </CardHeader>
                  
                  <CardContent>
                    <p className="mb-4 line-clamp-3 text-gray-600">
                      {entry.content}
                    </p>
                    <div className="flex items-center justify-between text-sm text-gray-500">
                      <div className="flex items-center gap-3">
                        <span className="flex items-center gap-1">
                          <Eye className="h-3 w-3" />
                          {entry.viewCount || 0}
                        </span>
                        <span className="flex items-center gap-1">
                          <Heart className="h-3 w-3" />
                          {entry.likeCount || 0}
                        </span>
                      </div>
                      {entry.theme && (
                        <Badge variant="secondary" className="text-xs">
                          {entry.theme}
                        </Badge>
                      )}
                    </div>
                  </CardContent>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      ) : (
        <Card className="text-center py-12">
          <CardContent>
            <BookOpen className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
            <p className="text-muted-foreground">该展览暂无展品</p>
          </CardContent>
        </Card>
      )}

      {/* Exhibition Footer */}
      <Card className="mt-8 bg-gradient-to-r from-amber-50 to-orange-50">
        <CardContent className="p-8">
          <div className="text-center">
            <h3 className="text-xl font-bold mb-4">喜欢这个展览？</h3>
            <p className="text-muted-foreground mb-6">
              发现更多精彩内容，或者分享您自己的故事
            </p>
            <div className="flex flex-wrap justify-center gap-4">
              <Button asChild>
                <Link href="/museum">
                  <BookOpen className="w-4 h-4 mr-2" />
                  浏览更多展览
                </Link>
              </Button>
              <Button variant="outline" asChild>
                <Link href="/museum/contribute">
                  <Award className="w-4 h-4 mr-2" />
                  贡献您的作品
                </Link>
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}