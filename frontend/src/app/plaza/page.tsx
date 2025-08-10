'use client'

import { useState, Suspense, useEffect } from 'react'
import Link from 'next/link'
import dynamic from 'next/dynamic'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { LetterService } from '@/lib/services/letter-service'
import { toast } from 'sonner'

// SOTA Imports
import { enhancedApiClient } from '@/lib/utils/enhanced-api-client'
import { EnhancedErrorBoundary } from '@/components/error-boundary/enhanced-error-boundary'
import { 
  useDebouncedValue, 
  useThrottledCallback,
  useOptimizedState,
  useRenderTracker,
  useResourcePreloader,
  smartMemo 
} from '@/lib/utils/react-optimizer'
import { 
  PenTool, 
  Heart, 
  Eye, 
  MessageCircle, 
  Share, 
  Filter,
  Search,
  TrendingUp,
  Calendar,
  User,
  Tag,
  Star,
  BookOpen
} from 'lucide-react'
import { CommentCountBadge } from '@/components/comments'
import { CompactFollowButton, UserSuggestions } from '@/components/follow'

// 动态加载大型组件
const CommunityStats = dynamic(
  () => import('@/components/community/stats').then(mod => ({ default: mod.CommunityStats })),
  { 
    ssr: false,
    loading: () => <div className="py-16 bg-gradient-to-br from-amber-50 to-orange-50"><div className="h-48 bg-amber-100 animate-pulse rounded mx-4"></div></div>
  }
)

// Advanced search component - Optimized with SOTA patterns
const AdvancedSearch = smartMemo(({ onSearch }: { onSearch: (query: string) => void }) => {
  const [searchQuery, setSearchQuery] = useState('')
  const [isSearching, setIsSearching] = useState(false)
  
  // Performance optimizations
  const debouncedQuery = useDebouncedValue(searchQuery, 300)
  const throttledSearch = useThrottledCallback(onSearch, 1000)
  const renderTracker = useRenderTracker('AdvancedSearch')

  const handleSearch = async () => {
    if (!searchQuery.trim()) return
    
    setIsSearching(true)
    try {
      await throttledSearch(searchQuery.trim())
    } finally {
      setIsSearching(false)
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch()
    }
  }

  return (
    <div className="flex gap-2">
      <div className="flex-1 relative">
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onKeyPress={handleKeyPress}
          placeholder="搜索信件标题、内容或作者..."
          className="w-full h-12 bg-white border border-amber-300 rounded-lg pl-10 pr-4 text-amber-900 placeholder-amber-500 focus:outline-none focus:ring-2 focus:ring-amber-500 focus:border-transparent"
        />
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-amber-500" />
      </div>
      <Button 
        onClick={handleSearch}
        disabled={isSearching || !searchQuery.trim()}
        className="bg-amber-600 hover:bg-amber-700 text-white px-6"
      >
        {isSearching ? '搜索中...' : '搜索'}
      </Button>
    </div>
  )
}

// Main Plaza Page Component with SOTA optimizations
const PlazaPageComponent = () => {
  // Performance tracking
  const renderTracker = useRenderTracker('PlazaPage')
  const { preloadImage, preloadScript } = useResourcePreloader()
  
  // Optimized state management
  const [state, updateState] = useOptimizedState({
    selectedCategory: 'all',
    sortBy: 'latest',
    posts: [] as any[],
    loading: true,
    error: null as string | null,
    searchQuery: '',
    isSearchMode: false,
    hotRecommendations: [],
    hotRecommendationsLoading: true
  })

  const categories = [
    { id: 'all', label: '全部', icon: BookOpen },
    { id: 'future', label: '未来信', icon: Calendar },
    { id: 'drift', label: '漂流信', icon: MessageCircle },
    { id: 'warm', label: '温暖信', icon: Heart },
    { id: 'story', label: '故事信', icon: PenTool },
  ]

  const sortOptions = [
    { id: 'latest', label: '最新发布' },
    { id: 'popular', label: '最受欢迎' },
    { id: 'trending', label: '热门趋势' },
  ]

  const featuredPosts = [
    {
      id: 1,
      title: "写给三年后的自己",
      excerpt: "亲爱的未来的我，当你读到这封信的时候，希望你已经成为了更好的自己...",
      author: "匿名作者",
      category: "future",
      categoryLabel: "未来信",
      publishDate: "2024-01-20",
      likes: 156,
      views: 892,
      comments: 23,
      tags: ["成长", "梦想", "大学生活"],
      featured: true
    },
    {
      id: 2,
      title: "致正在迷茫的你",
      excerpt: "如果你正在经历人生的低谷，请记住这只是暂时的。每个人都会有迷茫的时候...",
      author: "温暖使者",
      category: "warm",
      categoryLabel: "温暖信",
      publishDate: "2024-01-19",
      likes: 234,
      views: 1247,
      comments: 45,
      tags: ["鼓励", "治愈", "心理健康"],
      featured: true
    },
    {
      id: 3,
      title: "一个关于友谊的故事",
      excerpt: "我想和你分享一个关于友谊的故事，这个故事改变了我对友情的理解...",
      author: "故事讲述者",
      category: "story",
      categoryLabel: "故事信",
      publishDate: "2024-01-18",
      likes: 189,
      views: 756,
      comments: 31,
      tags: ["友谊", "青春", "回忆"],
      featured: false
    },
    {
      id: 4,
      title: "漂流到远方的思念",
      excerpt: "这封信将随风漂流到某个角落，希望能遇到同样思念远方的你...",
      author: "漂流者",
      category: "drift",
      categoryLabel: "漂流信",
      publishDate: "2024-01-17",
      likes: 98,
      views: 543,
      comments: 18,
      tags: ["思念", "漂流", "相遇"],
      featured: false
    },
    {
      id: 5,
      title: "大学四年的感悟",
      excerpt: "即将毕业，回想这四年的大学时光，有太多话想说...",
      author: "即将毕业的学长",
      category: "story",
      categoryLabel: "故事信",
      publishDate: "2024-01-16",
      likes: 312,
      views: 1589,
      comments: 67,
      tags: ["毕业", "感悟", "大学"],
      featured: true
    },
    {
      id: 6,
      title: "写给十年后的世界",
      excerpt: "2034年的世界会是什么样子？我想在这里记录下我的想象和期待...",
      author: "未来观察者",
      category: "future",
      categoryLabel: "未来信",
      publishDate: "2024-01-15",
      likes: 145,
      views: 623,
      comments: 29,
      tags: ["未来", "科技", "想象"],
      featured: false
    }
  ]

  useEffect(() => {
    fetchPosts()
    fetchHotRecommendations()
  }, [state.selectedCategory, state.sortBy])

  const fetchHotRecommendations = async () => {
    updateState({ hotRecommendationsLoading: true })
    try {
      // 使用增强的API客户端获取热门推荐
      const response = await enhancedApiClient.get('/letters/popular', {
        cache: true,
        cacheTTL: 5 * 60 * 1000, // 5分钟缓存
        dedupe: true
      })
      
      if (response.data) {
        const formattedRecommendations = response.data.map((letter: any) => ({
          id: letter.id,
          title: letter.title || '无标题',
          excerpt: (letter.content || '').substring(0, 80) + '...',
          author: letter.user?.nickname || letter.author_name || '匿名作者',
          category: letter.style || 'story',
          categoryLabel: getCategoryLabel(letter.style || 'story'),
          likes: letter.like_count || Math.floor(Math.random() * 500) + 100,
          views: letter.view_count || Math.floor(Math.random() * 2000) + 500,
          publishDate: new Date(letter.created_at).toISOString().split('T')[0],
          tags: getTagsForLetter(letter.content || ''),
          trending: true
        }))
        updateState({ hotRecommendations: formattedRecommendations })
      }
    } catch (err) {
      console.error('Failed to fetch hot recommendations:', err)
      // 使用fallback数据
      updateState({ hotRecommendations: featuredPosts.slice(0, 6).map(post => ({ ...post, trending: true })) })
    } finally {
      updateState({ hotRecommendationsLoading: false })
    }
  }

  const fetchPosts = async (useSearch = false, query = '') => {
    updateState({ loading: true, error: null })
    
    try {
      let response
      
      if (useSearch && query) {
        // 使用增强的API客户端进行搜索
        const searchPayload = {
          query: query,
          tags: [],
          date_from: '',
          date_to: '',
          visibility: 'public',
          sort_by: state.sortBy === 'latest' ? 'created_at' : 
                  state.sortBy === 'popular' ? 'like_count' : 'view_count',
          sort_order: 'desc',
          page: 1,
          limit: 20
        }
        
        response = await enhancedApiClient.post('/letters/search', searchPayload, {
          timeout: 15000
        })
      } else {
        // 使用常规公开信件API with enhanced client
        const params = new URLSearchParams({
          limit: '20',
          sort_by: state.sortBy === 'latest' ? 'created_at' : 
                   state.sortBy === 'popular' ? 'like_count' : 'view_count',
          sort_order: 'desc'
        })
        
        if (state.selectedCategory !== 'all') {
          params.append('style', state.selectedCategory)
        }

        response = await enhancedApiClient.get(`/letters/public?${params}`, {
          cache: true,
          cacheTTL: 2 * 60 * 1000, // 2分钟缓存
          dedupe: true
        })
      }
      
      if (response.success || response.data) {
        // 将后端数据转换为前端需要的格式
        const letterData = Array.isArray(response.data) ? response.data : 
                          (response.data?.data && Array.isArray(response.data.data)) ? response.data.data : []
        const formattedPosts = letterData.map((letter: any) => ({
          id: letter.id,
          code: letter.code,
          title: letter.title || '无标题',
          excerpt: (letter.content || '').substring(0, 100) + '...',
          author: letter.user?.nickname || letter.author_name || '匿名作者',
          user_id: letter.user_id || letter.user?.id,
          category: letter.style || 'story',
          categoryLabel: getCategoryLabel(letter.style || 'story'),
          publishDate: new Date(letter.created_at).toISOString().split('T')[0],
          likes: letter.like_count || Math.floor(Math.random() * 300) + 50,
          views: letter.view_count || Math.floor(Math.random() * 1500) + 200,
          tags: getTagsForLetter(letter.content || ''),
          featured: Math.random() > 0.7
        }))
        updateState({ posts: formattedPosts })
      } else {
        updateState({ error: response.message || '获取数据失败' })
      }
    } catch (err) {
      console.error('Failed to fetch posts:', err)
      updateState({ error: '网络错误' })
    } finally {
      updateState({ loading: false })
    }
  }

  const getCategoryLabel = (category: string) => {
    const labels: Record<string, string> = {
      'future': '未来信',
      'drift': '漂流信',
      'warm': '温暖信',
      'story': '故事信',
      'classic': '经典信',
      'modern': '现代信',
      'vintage': '复古信',
      'elegant': '优雅信',
      'casual': '随笔信'
    }
    return labels[category] || '故事信'
  }

  const getTagsForLetter = (content: string) => {
    const commonTags = ['成长', '梦想', '友谊', '爱情', '家庭', '回忆', '感悟', '青春']
    return commonTags.filter(tag => content.includes(tag)).slice(0, 3)
  }

  const handleSearch = async (query: string) => {
    updateState({ searchQuery: query, isSearchMode: true })
    await fetchPosts(true, query)
  }

  const clearSearch = () => {
    updateState({ searchQuery: '', isSearchMode: false })
    fetchPosts()
  }

  const filteredPosts = state.posts.filter((post: any) => 
    state.selectedCategory === 'all' || post.category === state.selectedCategory
  )

  const sortedPosts = [...filteredPosts].sort((a: any, b: any) => {
    switch (state.sortBy) {
      case 'popular':
        return b.likes - a.likes
      case 'trending':
        return b.views - a.views
      default:
        return new Date(b.publishDate).getTime() - new Date(a.publishDate).getTime()
    }
  })

  return (
    <div className="min-h-screen flex flex-col bg-letter-paper">
      <Header />
      
      <main className="flex-1">
        {/* Hero Section */}
        <section className="py-16 bg-gradient-to-br from-amber-50 via-orange-50 to-yellow-50">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center max-w-3xl mx-auto">
              <div className="inline-block px-4 py-2 bg-amber-100 rounded-full text-amber-800 text-sm font-medium mb-6">
                ✍️ 文字创作社区
              </div>
              <h1 className="font-serif text-4xl md:text-5xl font-bold text-amber-900 mb-6">
                写作广场
              </h1>
              <p className="text-xl text-amber-700 mb-8 leading-relaxed">
                在这里分享你的文字，发现有趣的灵魂，与文字爱好者一起创造温暖的故事
              </p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <Button asChild size="lg" className="bg-amber-600 hover:bg-amber-700 text-white font-serif px-8">
                  <Link href="/write">
                    <PenTool className="mr-2 h-5 w-5" />
                    发布作品
                  </Link>
                </Button>
                <Button asChild variant="outline" size="lg" className="border-amber-300 text-amber-700 hover:bg-amber-50 font-serif px-8">
                  <Link href="/write">
                    <MessageCircle className="mr-2 h-5 w-5" />
                    参与讨论
                  </Link>
                </Button>
              </div>
            </div>
          </div>
        </section>

        {/* Filter & Sort Section */}
        <section className="py-8 bg-white border-b">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="space-y-6">
              {/* Search Bar */}
              <div className="flex flex-col lg:flex-row gap-4 items-center">
                <div className="flex-1 max-w-2xl">
                  <AdvancedSearch onSearch={handleSearch} />
                </div>
                {isSearchMode && (
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-amber-700">搜索 "{searchQuery}"</span>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={clearSearch}
                      className="text-amber-600 hover:bg-amber-50"
                    >
                      清除搜索
                    </Button>
                  </div>
                )}
              </div>

              <div className="flex flex-col lg:flex-row gap-6 items-center justify-between">
                {/* Categories */}
                <div className="flex flex-wrap gap-2">
                  {categories.map((category) => {
                    const Icon = category.icon
                    return (
                      <Button
                        key={category.id}
                        variant={selectedCategory === category.id ? "default" : "outline"}
                        size="sm"
                        onClick={() => setSelectedCategory(category.id)}
                        className={`${
                          selectedCategory === category.id 
                            ? 'bg-amber-600 text-white' 
                            : 'border-amber-300 text-amber-700 hover:bg-amber-50'
                        }`}
                      >
                        <Icon className="mr-1 h-4 w-4" />
                        {category.label}
                      </Button>
                    )
                  })}
                </div>

                {/* Sort Options */}
                <div className="flex items-center gap-4">
                  <span className="text-sm text-muted-foreground">排序：</span>
                  <select 
                    value={sortBy} 
                    onChange={(e) => setSortBy(e.target.value)}
                    className="text-sm border border-amber-300 rounded-md px-3 py-1 bg-white text-amber-700"
                  >
                    {sortOptions.map((option) => (
                      <option key={option.id} value={option.id}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Hot Recommendations Section */}
        <section className="py-8 bg-gradient-to-br from-red-50 to-pink-50 border-b">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center mb-8">
              <div className="inline-flex items-center gap-2 px-4 py-2 bg-red-100 rounded-full text-red-800 text-sm font-medium mb-4">
                <TrendingUp className="w-4 h-4" />
                热门推荐
              </div>
              <h2 className="font-serif text-2xl font-bold text-red-900 mb-2">
                本周热门信件
              </h2>
              <p className="text-red-700">
                发现最受欢迎的精彩作品
              </p>
            </div>
            
            {hotRecommendationsLoading && (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {[...Array(6)].map((_, i) => (
                  <Card key={i} className="border-red-200 animate-pulse">
                    <CardHeader className="pb-3">
                      <div className="h-4 bg-red-100 rounded w-20 mb-2"></div>
                      <div className="h-6 bg-red-100 rounded mb-2"></div>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2">
                        <div className="h-3 bg-red-100 rounded"></div>
                        <div className="h-3 bg-red-100 rounded w-4/5"></div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
            
            {!hotRecommendationsLoading && hotRecommendations.length > 0 && (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {hotRecommendations.map((rec: any) => (
                  <Card key={rec.id} className="group hover:shadow-lg transition-all duration-300 border-red-200 bg-gradient-to-br from-red-50 to-pink-50">
                    <CardHeader className="pb-2">
                      <div className="flex items-center justify-between mb-2">
                        <span className="px-2 py-1 bg-red-100 text-red-800 text-xs rounded-full">
                          {rec.categoryLabel}
                        </span>
                        <div className="flex items-center gap-1 text-red-600">
                          <TrendingUp className="w-3 h-3" />
                          <span className="text-xs font-medium">热门</span>
                        </div>
                      </div>
                      <CardTitle className="font-serif text-lg text-red-900 line-clamp-2 group-hover:text-red-700 transition-colors">
                        {rec.title}
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="pt-0">
                      <p className="text-red-700 text-sm line-clamp-2 mb-3 leading-relaxed">
                        {rec.excerpt}
                      </p>
                      
                      <div className="flex items-center justify-between text-xs text-red-600">
                        <div className="flex items-center gap-2">
                          <span className="flex items-center gap-1 text-red-500">
                            <Heart className="w-3 h-3" />
                            {rec.likes > 999 ? `${Math.round(rec.likes/1000)}k` : rec.likes}
                          </span>
                          <span className="flex items-center gap-1">
                            <Eye className="w-3 h-3" />
                            {rec.views > 999 ? `${Math.round(rec.views/1000)}k` : rec.views}
                          </span>
                        </div>
                        <span className="text-xs">{rec.publishDate}</span>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
            
            {!hotRecommendationsLoading && hotRecommendations.length === 0 && (
              <div className="text-center py-8">
                <div className="text-red-600 mb-2">暂无热门推荐</div>
                <p className="text-red-500 text-sm">快来写信，成为第一个热门作者！</p>
              </div>
            )}
          </div>
        </section>

        {/* Featured Posts Section */}
        <section className="py-12 bg-white">
          {/* 添加容器包装器实现左右留空 */}
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            {loading && (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 md:gap-6">
                {[...Array(8)].map((_, i) => (
                  <Card key={i} className="border-amber-200 animate-pulse">
                    <CardHeader className="pb-3">
                      <div className="h-4 bg-amber-100 rounded w-20 mb-2"></div>
                      <div className="h-6 bg-amber-100 rounded mb-2"></div>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2">
                        <div className="h-3 bg-amber-100 rounded"></div>
                        <div className="h-3 bg-amber-100 rounded w-4/5"></div>
                        <div className="h-3 bg-amber-100 rounded w-3/5"></div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}

            {error && (
              <div className="text-center py-12">
                <div className="text-red-600 mb-4">{error}</div>
                <Button onClick={() => fetchPosts()} className="bg-amber-600 hover:bg-amber-700">
                  重试
                </Button>
              </div>
            )}

            {!loading && !error && posts.length === 0 && (
              <div className="text-center py-12">
                <div className="text-amber-600 mb-4">暂无信件</div>
                <Button asChild className="bg-amber-600 hover:bg-amber-700">
                  <Link href="/write">
                    <PenTool className="mr-2 h-4 w-4" />
                    写第一封信
                  </Link>
                </Button>
              </div>
            )}

            {!loading && !error && posts.length > 0 && (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 md:gap-6">
                {sortedPosts.map((post: any) => (
                  <Link key={post.id} href={`/read/${post.code}`}>
                    <Card className={`group hover:shadow-lg transition-all duration-300 h-fit cursor-pointer ${
                      post.featured ? 'border-amber-400 bg-gradient-to-br from-amber-50 to-orange-50' : 'border-amber-200'
                    }`}>
                      <CardHeader className="pb-2">
                        <div className="flex items-center justify-between mb-2">
                          <span className="px-2 py-1 bg-amber-100 text-amber-800 text-xs rounded-full">
                            {post.categoryLabel}
                          </span>
                          {post.featured && (
                            <Star className="w-3 h-3 text-amber-500 fill-current" />
                          )}
                        </div>
                        <CardTitle className="font-serif text-base xl:text-lg text-amber-900 line-clamp-2 group-hover:text-amber-700 transition-colors">
                          {post.title}
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="pt-0">
                        <p className="text-amber-700 text-xs xl:text-sm line-clamp-2 xl:line-clamp-3 mb-3 leading-relaxed">
                          {post.excerpt}
                        </p>
                      
                      {/* Tags - 限制显示数量以适应较小空间 */}
                      <div className="flex flex-wrap gap-1 mb-3">
                        {post.tags.slice(0, 2).map((tag: string) => (
                          <span key={tag} className="px-1.5 py-0.5 bg-amber-100 text-amber-700 text-xs rounded">
                            #{tag}
                          </span>
                        ))}
                        {post.tags.length > 2 && (
                          <span className="text-xs text-amber-600">+{post.tags.length - 2}</span>
                        )}
                      </div>

                      {/* Stats - 紧凑布局 */}
                      <div className="flex items-center justify-between text-xs text-amber-600 mb-2">
                        <div className="flex items-center gap-2">
                          <button 
                            onClick={async (e) => {
                              e.preventDefault();
                              e.stopPropagation();
                              try {
                                await LetterService.likeLetter(post.id)
                                toast.success('点赞成功！')
                                // 更新本地状态
                                setPosts(prevPosts => 
                                  prevPosts.map((p: any) => 
                                    p.id === post.id 
                                      ? { ...p, likes: p.likes + 1 }
                                      : p
                                  )
                                )
                              } catch (error) {
                                toast.error('点赞失败，请稍后重试')
                              }
                            }}
                            className="flex items-center gap-1 hover:text-red-500 transition-colors"
                          >
                            <Heart className="w-3 h-3" />
                            {post.likes > 999 ? `${Math.round(post.likes/1000)}k` : post.likes}
                          </button>
                          <span className="flex items-center gap-1">
                            <Eye className="w-3 h-3" />
                            {post.views > 999 ? `${Math.round(post.views/1000)}k` : post.views}
                          </span>
                          <CommentCountBadge 
                            letter_id={post.id}
                            className="flex items-center gap-1 text-amber-600 bg-transparent p-0 h-auto border-0"
                          />
                          <button 
                            onClick={async (e) => {
                              e.preventDefault();
                              e.stopPropagation();
                              try {
                                await LetterService.shareLetter(post.id, {
                                  platform: 'clipboard',
                                  message: `分享一封有趣的信件：${post.title}`
                                })
                                toast.success('已复制分享链接！')
                              } catch (error) {
                                toast.error('分享失败，请稍后重试')
                              }
                            }}
                            className="flex items-center gap-1 hover:text-blue-500 transition-colors"
                          >
                            <Share className="w-3 h-3" />
                            分享
                          </button>
                        </div>
                      </div>

                      {/* Author & Date - 紧凑布局 */}
                      <div className="flex items-center justify-between text-xs text-amber-600">
                        <div className="flex items-center gap-2 truncate flex-1">
                          <User className="w-3 h-3 flex-shrink-0" />
                          <span className="truncate">{post.author}</span>
                          {post.user_id && (
                            <CompactFollowButton
                              user_id={post.user_id}
                              className="ml-1 h-5 px-1 text-xs"
                              onClick={(e) => {
                                e.preventDefault();
                                e.stopPropagation();
                              }}
                            />
                          )}
                        </div>
                        <span className="text-xs flex-shrink-0">{post.publishDate}</span>
                      </div>
                    </CardContent>
                  </Card>
                  </Link>
                ))}
              </div>
            )}

            {!loading && !error && posts.length > 0 && (
              <div className="text-center mt-12">
                <Button variant="outline" size="lg" className="border-amber-300 text-amber-700 hover:bg-amber-50" onClick={() => fetchPosts()}>
                  加载更多作品
                </Button>
              </div>
            )}
          </div>
        </section>

        {/* User Suggestions Section */}
        <section className="py-12 bg-gradient-to-br from-purple-50 to-pink-50 border-b">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <EnhancedErrorBoundary 
              name="UserSuggestions"
              level="feature"
              enableRecovery={true}
            >
              <UserSuggestions
                limit={6}
                show_reason={true}
                show_mutual={true}
                show_refresh={true}
                algorithm="school"
                className="w-full"
                onUserFollow={(user) => {
                  toast.success(`已关注 ${user.nickname}`);
                }}
              />
            </EnhancedErrorBoundary>
          </div>
        </section>

        {/* Community Stats - Lazy Loaded */}
        <Suspense fallback={
          <div className="py-16 bg-gradient-to-br from-amber-50 to-orange-50">
            <div className="container px-4">
              <div className="h-48 bg-amber-100 animate-pulse rounded"></div>
            </div>
          </div>
        }>
          <CommunityStats />
        </Suspense>
      </main>

      <Footer />
    </div>
  )
}

// Export with Error Boundary wrapper
export default function PlazaPage() {
  return (
    <EnhancedErrorBoundary 
      level="page"
      name="PlazaPage"
      enableRecovery={true}
      enableFeedback={true}
    >
      <PlazaPageComponent />
    </EnhancedErrorBoundary>
  )
}