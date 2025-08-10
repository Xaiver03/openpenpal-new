'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import {
  Tag,
  TrendingUp,
  Search,
  Hash,
  Eye,
  BookOpen,
  Heart,
  MessageCircle,
  Calendar,
  Filter,
  Grid,
  List,
  Star
} from 'lucide-react'
import { museumService } from '@/lib/services/museum-service'

interface MuseumTag {
  id: string
  name: string
  count: number
  trending: boolean
  category: string
  description: string
  related_tags: string[]
  popularity_score: number
  created_at: string
}

interface TagCategory {
  id: string
  name: string
  icon: React.ElementType
  color: string
}

const tagCategories: TagCategory[] = [
  { id: 'emotion', name: '情感', icon: Heart, color: 'text-red-500' },
  { id: 'memory', name: '回忆', icon: Calendar, color: 'text-blue-500' },
  { id: 'growth', name: '成长', icon: TrendingUp, color: 'text-green-500' },
  { id: 'story', name: '故事', icon: BookOpen, color: 'text-purple-500' },
  { id: 'other', name: '其他', icon: Tag, color: 'text-gray-500' }
]

export default function MuseumTagsPage() {
  const router = useRouter()
  const [tags, setTags] = useState<MuseumTag[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [sortBy, setSortBy] = useState<'popular' | 'trending' | 'alphabetical'>('popular')

  useEffect(() => {
    fetchTags()
  }, [])

  const fetchTags = async () => {
    setLoading(true)
    setError(null)

    try {
      const response = await museumService.getMuseumTags()
      
      // 模拟数据转换，实际应该从后端返回
      const formattedTags: MuseumTag[] = response.data?.map((tag: any) => ({
        id: tag.id,
        name: tag.name,
        count: Math.floor(Math.random() * 500) + 50,
        trending: Math.random() > 0.7,
        category: ['emotion', 'memory', 'growth', 'story', 'other'][Math.floor(Math.random() * 5)],
        description: tag.description || `探索关于"${tag.name}"的信件故事`,
        related_tags: generateRelatedTags(tag.name),
        popularity_score: Math.floor(Math.random() * 100),
        created_at: tag.created_at || new Date().toISOString()
      })) || []

      setTags(formattedTags)
    } catch (err) {
      console.error('获取标签失败:', err)
      setError('获取标签失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  const generateRelatedTags = (tagName: string): string[] => {
    const relatedPool = ['青春', '回忆', '梦想', '友谊', '爱情', '成长', '感动', '温暖', '思念', '希望']
    return relatedPool
      .filter(t => t !== tagName)
      .sort(() => Math.random() - 0.5)
      .slice(0, 3)
  }

  const filteredTags = tags
    .filter(tag => {
      const matchesSearch = tag.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          tag.description.toLowerCase().includes(searchQuery.toLowerCase())
      const matchesCategory = !selectedCategory || tag.category === selectedCategory
      return matchesSearch && matchesCategory
    })
    .sort((a, b) => {
      switch (sortBy) {
        case 'popular':
          return b.count - a.count
        case 'trending':
          return b.popularity_score - a.popularity_score
        case 'alphabetical':
          return a.name.localeCompare(b.name)
        default:
          return 0
      }
    })

  const getTagSize = (count: number): string => {
    if (count > 300) return 'text-2xl font-bold'
    if (count > 200) return 'text-xl font-semibold'
    if (count > 100) return 'text-lg font-medium'
    return 'text-base'
  }

  const getTagOpacity = (count: number): string => {
    const maxCount = Math.max(...tags.map(t => t.count))
    const opacity = 0.4 + (count / maxCount) * 0.6
    return `opacity-${Math.round(opacity * 100)}`
  }

  return (
    <div className="container max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-letter-ink mb-2">
          标签云
        </h1>
        <p className="text-muted-foreground">
          探索不同主题的信件，发现感兴趣的内容
        </p>
      </div>

      {/* Search and Filters */}
      <div className="mb-6 space-y-4">
        {/* Search Bar */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <Input
            placeholder="搜索标签..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>

        {/* Filter Controls */}
        <div className="flex flex-wrap items-center gap-4">
          {/* Category Filter */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">分类：</span>
            <Button
              variant={!selectedCategory ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSelectedCategory(null)}
            >
              全部
            </Button>
            {tagCategories.map(category => {
              const Icon = category.icon
              return (
                <Button
                  key={category.id}
                  variant={selectedCategory === category.id ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setSelectedCategory(category.id)}
                >
                  <Icon className={`w-4 h-4 mr-1 ${category.color}`} />
                  {category.name}
                </Button>
              )
            })}
          </div>

          {/* Sort Options */}
          <div className="flex items-center gap-2 ml-auto">
            <span className="text-sm text-muted-foreground">排序：</span>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as any)}
              className="text-sm border rounded px-2 py-1"
            >
              <option value="popular">最热门</option>
              <option value="trending">趋势</option>
              <option value="alphabetical">字母序</option>
            </select>
          </div>

          {/* View Mode Toggle */}
          <div className="flex items-center gap-1">
            <Button
              variant={viewMode === 'grid' ? 'default' : 'ghost'}
              size="icon"
              onClick={() => setViewMode('grid')}
            >
              <Grid className="h-4 w-4" />
            </Button>
            <Button
              variant={viewMode === 'list' ? 'default' : 'ghost'}
              size="icon"
              onClick={() => setViewMode('list')}
            >
              <List className="h-4 w-4" />
            </Button>
          </div>
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
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[...Array(9)].map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader>
                <div className="h-6 bg-muted rounded w-3/4"></div>
              </CardHeader>
              <CardContent>
                <div className="h-4 bg-muted rounded w-full mb-2"></div>
                <div className="h-4 bg-muted rounded w-2/3"></div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Tags Display */}
      {!loading && filteredTags.length > 0 && viewMode === 'grid' && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredTags.map(tag => {
            const category = tagCategories.find(c => c.id === tag.category)
            const Icon = category?.icon || Tag
            
            return (
              <Card 
                key={tag.id}
                className="cursor-pointer hover:shadow-lg transition-all"
                onClick={() => router.push(`/museum?tag=${encodeURIComponent(tag.name)}`)}
              >
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-2">
                      <Icon className={`w-5 h-5 ${category?.color || 'text-gray-500'}`} />
                      <CardTitle className="text-lg">
                        #{tag.name}
                      </CardTitle>
                    </div>
                    {tag.trending && (
                      <Badge variant="secondary" className="text-xs">
                        <TrendingUp className="w-3 h-3 mr-1" />
                        热门
                      </Badge>
                    )}
                  </div>
                  <CardDescription>
                    {tag.description}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center justify-between text-sm text-muted-foreground mb-3">
                    <span className="flex items-center gap-1">
                      <BookOpen className="w-4 h-4" />
                      {tag.count} 篇信件
                    </span>
                    <span className="flex items-center gap-1">
                      <Star className="w-4 h-4" />
                      {tag.popularity_score}% 热度
                    </span>
                  </div>
                  
                  {/* Related Tags */}
                  {tag.related_tags.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      <span className="text-xs text-muted-foreground">相关：</span>
                      {tag.related_tags.map(relatedTag => (
                        <Badge key={relatedTag} variant="outline" className="text-xs">
                          {relatedTag}
                        </Badge>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}

      {/* Tag Cloud View */}
      {!loading && filteredTags.length > 0 && viewMode === 'list' && (
        <Card className="p-8">
          <div className="flex flex-wrap gap-4 justify-center items-center">
            {filteredTags.map(tag => {
              const category = tagCategories.find(c => c.id === tag.category)
              
              return (
                <Button
                  key={tag.id}
                  variant="ghost"
                  className={`${getTagSize(tag.count)} ${category?.color || 'text-gray-700'} hover:scale-110 transition-transform`}
                  onClick={() => router.push(`/museum?tag=${encodeURIComponent(tag.name)}`)}
                >
                  <Hash className="w-4 h-4 mr-1" />
                  {tag.name}
                  <Badge variant="secondary" className="ml-2 text-xs">
                    {tag.count}
                  </Badge>
                </Button>
              )
            })}
          </div>
        </Card>
      )}

      {/* Empty State */}
      {!loading && filteredTags.length === 0 && (
        <Card className="text-center py-12">
          <CardContent>
            <Tag className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
            <p className="text-muted-foreground">
              {searchQuery ? `没有找到包含"${searchQuery}"的标签` : '暂无标签'}
            </p>
          </CardContent>
        </Card>
      )}

      {/* Stats Summary */}
      {!loading && tags.length > 0 && (
        <Card className="mt-8">
          <CardHeader>
            <CardTitle>标签统计</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
              <div>
                <div className="text-2xl font-bold">{tags.length}</div>
                <div className="text-sm text-muted-foreground">总标签数</div>
              </div>
              <div>
                <div className="text-2xl font-bold">
                  {tags.reduce((sum, tag) => sum + tag.count, 0).toLocaleString()}
                </div>
                <div className="text-sm text-muted-foreground">标记次数</div>
              </div>
              <div>
                <div className="text-2xl font-bold">
                  {tags.filter(tag => tag.trending).length}
                </div>
                <div className="text-sm text-muted-foreground">热门标签</div>
              </div>
              <div>
                <div className="text-2xl font-bold">
                  {Math.round(tags.reduce((sum, tag) => sum + tag.count, 0) / tags.length)}
                </div>
                <div className="text-sm text-muted-foreground">平均使用</div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}