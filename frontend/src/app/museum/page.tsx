'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { EnvelopeAnimation } from '@/components/ui/envelope-animation'
import { Skeleton } from '@/components/ui/skeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { useMuseumEntries, useMuseumExhibitions, useMuseumStats } from '@/hooks/use-museum'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { 
  Archive,
  Heart, 
  Eye, 
  Calendar,
  MapPin,
  Clock,
  Filter,
  Search,
  Star,
  BookOpen,
  Mail,
  Sparkles,
  Globe,
  Users,
  ChevronLeft,
  ChevronRight,
  Play,
  Pause,
  AlertCircle,
  Hash
} from 'lucide-react'

export default function MuseumPage() {
  const [selectedTheme, setSelectedTheme] = useState('all')
  const [selectedSort, setSelectedSort] = useState<'created_at' | 'view_count' | 'like_count'>('created_at')
  const [currentSlide, setCurrentSlide] = useState(0)
  const [isAutoPlaying, setIsAutoPlaying] = useState(true)
  const [page, setPage] = useState(1)
  const limit = 12

  // Fetch data using hooks
  const { data: entriesData, isLoading: entriesLoading, error: entriesError } = useMuseumEntries({
    page,
    limit,
    theme: selectedTheme === 'all' ? undefined : selectedTheme,
    sort_by: selectedSort,
    order: 'desc'
  })

  const { data: exhibitions, isLoading: exhibitionsLoading } = useMuseumExhibitions(true)
  const { data: stats } = useMuseumStats()

  const themes = [
    { id: 'all', label: '全部主题' },
    { id: 'future', label: '未来憧憬' },
    { id: 'memory', label: '青春记忆' },
    { id: 'warmth', label: '温暖治愈' },
    { id: 'story', label: '故事分享' },
    { id: 'friendship', label: '友谊永恒' },
  ]

  const sortOptions = [
    { value: 'created_at', label: '最新发布' },
    { value: 'view_count', label: '最多浏览' },
    { value: 'like_count', label: '最受欢迎' },
  ]

  // Auto-play carousel for exhibitions
  useEffect(() => {
    if (!isAutoPlaying || !exhibitions || exhibitions.length === 0) return
    const interval = setInterval(() => {
      setCurrentSlide(prev => (prev + 1) % exhibitions.length)
    }, 4000)
    return () => clearInterval(interval)
  }, [isAutoPlaying, exhibitions])

  const handlePrevSlide = () => {
    if (!exhibitions) return
    setCurrentSlide(prev => (prev - 1 + exhibitions.length) % exhibitions.length)
  }

  const handleNextSlide = () => {
    if (!exhibitions) return
    setCurrentSlide(prev => (prev + 1) % exhibitions.length)
  }

  return (
    <div className="min-h-screen flex flex-col bg-gradient-to-b from-amber-50 to-white">
      <Header />
      
      <main className="flex-1 container mx-auto px-4 py-8">
        {/* Hero Section */}
        <div className="relative mb-12 overflow-hidden rounded-2xl bg-gradient-to-r from-amber-600 to-orange-600 text-white">
          <div className="absolute inset-0 bg-black/20" />
          <div className="relative z-10 px-8 py-16 text-center">
            <h1 className="mb-4 text-5xl font-bold">信件博物馆</h1>
            <p className="mb-8 text-xl opacity-90">
              珍藏每一份真挚情感，让美好的文字永恒流传
            </p>
            <div className="flex flex-wrap items-center justify-center gap-8 text-sm">
              <div className="flex items-center gap-2">
                <Archive className="h-5 w-5" />
                <span>{stats?.total_entries || 0} 封珍藏信件</span>
              </div>
              <div className="flex items-center gap-2">
                <Eye className="h-5 w-5" />
                <span>{stats?.total_views || 0} 次阅读</span>
              </div>
              <div className="flex items-center gap-2">
                <Star className="h-5 w-5" />
                <span>{stats?.featured_count || 0} 精选作品</span>
              </div>
            </div>
          </div>
          <div className="absolute -bottom-10 -right-10 h-40 w-40 opacity-10">
            {/* @ts-ignore */}
            <EnvelopeAnimation letter={{
              id: 1,
              title: "Demo Letter",
              preview: "Demo content",
              author: "Demo Author",
              type: "letter",
              date: new Date().toISOString(),
              location: "Demo Location",
              likes: 0,
              views: 0,
              significance: "Demo",
              featured: false
            }} />
          </div>
        </div>

        {/* Current Exhibitions */}
        {exhibitions && exhibitions.length > 0 && (
          <section className="mb-12">
            <div className="mb-6 flex items-center justify-between">
              <h2 className="text-3xl font-bold text-gray-800">当前展览</h2>
              <div className="flex items-center gap-2">
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setIsAutoPlaying(!isAutoPlaying)}
                >
                  {isAutoPlaying ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                </Button>
                <Button variant="ghost" size="icon" onClick={handlePrevSlide}>
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon" onClick={handleNextSlide}>
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>

            <div className="relative overflow-hidden rounded-xl">
              <div
                className="flex transition-transform duration-500"
                style={{ transform: `translateX(-${currentSlide * 100}%)` }}
              >
                {exhibitions.map((exhibition, index) => (
                  <div key={exhibition.id} className="w-full flex-shrink-0">
                    <Card className="border-0 bg-gradient-to-r from-purple-500 to-pink-500 text-white">
                      <CardContent className="p-8">
                        <div className="flex flex-col md:flex-row items-center gap-8">
                          {exhibition.cover_image && (
                            <img
                              src={exhibition.cover_image}
                              alt={exhibition.title}
                              className="w-full md:w-1/3 rounded-lg object-cover"
                            />
                          )}
                          <div className="flex-1">
                            <h3 className="mb-4 text-3xl font-bold">{exhibition.title}</h3>
                            <p className="mb-4 text-lg opacity-90">{exhibition.description}</p>
                            <div className="mb-6 flex flex-wrap gap-2">
                              {exhibition.theme_keywords.map((keyword, i) => (
                                <span
                                  key={i}
                                  className="rounded-full bg-white/20 px-3 py-1 text-sm"
                                >
                                  {keyword}
                                </span>
                              ))}
                            </div>
                            <div className="flex items-center justify-between">
                              <div className="text-sm opacity-75">
                                <p>策展人：{exhibition.curator_name || '博物馆团队'}</p>
                                <p>展品数量：{exhibition.entry_count} 件</p>
                              </div>
                              <Link href={`/museum/exhibition/${exhibition.id}`}>
                                <Button variant="secondary">
                                  进入展览
                                </Button>
                              </Link>
                            </div>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                ))}
              </div>
            </div>

            {/* Carousel Indicators */}
            <div className="mt-4 flex justify-center gap-2">
              {exhibitions.map((_, index) => (
                <button
                  key={index}
                  className={`h-2 w-2 rounded-full transition-all ${
                    index === currentSlide
                      ? 'w-8 bg-amber-600'
                      : 'bg-gray-300 hover:bg-gray-400'
                  }`}
                  onClick={() => setCurrentSlide(index)}
                />
              ))}
            </div>
          </section>
        )}

        {/* Filters */}
        <section className="mb-8">
          <div className="flex flex-col md:flex-row gap-4 items-center justify-between">
            <div className="flex flex-wrap gap-2">
              {themes.map(theme => (
                <Button
                  key={theme.id}
                  variant={selectedTheme === theme.id ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => {
                    setSelectedTheme(theme.id)
                    setPage(1)
                  }}
                >
                  {theme.label}
                </Button>
              ))}
            </div>
            <div className="flex items-center gap-4">
              <select
                value={selectedSort}
                onChange={(e) => {
                  setSelectedSort(e.target.value as typeof selectedSort)
                  setPage(1)
                }}
                className="rounded-lg border px-3 py-2 text-sm"
              >
                {sortOptions.map(option => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </section>

        {/* Quick Actions */}
        <section className="mb-12">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Link href="/museum/popular">
              <Card className="cursor-pointer hover:shadow-lg transition-all">
                <CardContent className="p-6">
                  <div className="flex items-center gap-4">
                    <div className="p-3 rounded-full bg-red-100">
                      <Heart className="h-6 w-6 text-red-600" />
                    </div>
                    <div>
                      <h3 className="font-semibold">热门信件</h3>
                      <p className="text-sm text-muted-foreground">查看最受欢迎的信件</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </Link>
            
            <Link href="/museum/tags">
              <Card className="cursor-pointer hover:shadow-lg transition-all">
                <CardContent className="p-6">
                  <div className="flex items-center gap-4">
                    <div className="p-3 rounded-full bg-blue-100">
                      <Hash className="h-6 w-6 text-blue-600" />
                    </div>
                    <div>
                      <h3 className="font-semibold">标签云</h3>
                      <p className="text-sm text-muted-foreground">按标签浏览信件</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </Link>
            
            <Link href="/museum/my-submissions">
              <Card className="cursor-pointer hover:shadow-lg transition-all">
                <CardContent className="p-6">
                  <div className="flex items-center gap-4">
                    <div className="p-3 rounded-full bg-green-100">
                      <BookOpen className="h-6 w-6 text-green-600" />
                    </div>
                    <div>
                      <h3 className="font-semibold">我的提交</h3>
                      <p className="text-sm text-muted-foreground">查看您提交的信件</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </Link>
          </div>
        </section>

        {/* Museum Entries */}
        <section className="mb-12">
          <h2 className="mb-6 text-3xl font-bold text-gray-800">馆藏信件</h2>
          
          {entriesError && (
            <Alert variant="destructive" className="mb-6">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                加载失败：{entriesError.message}
              </AlertDescription>
            </Alert>
          )}

          {entriesLoading ? (
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              {[...Array(6)].map((_, i) => (
                <Card key={i}>
                  <CardHeader>
                    <Skeleton className="h-6 w-3/4" />
                    <Skeleton className="h-4 w-full mt-2" />
                  </CardHeader>
                  <CardContent>
                    <Skeleton className="h-20 w-full" />
                    <div className="mt-4 flex justify-between">
                      <Skeleton className="h-4 w-20" />
                      <Skeleton className="h-4 w-20" />
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : entriesData?.entries && entriesData.entries.length > 0 ? (
            <>
              <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                {entriesData.entries.map(entry => (
                  <Link key={entry.id} href={`/museum/letter/${entry.id}`}>
                    <Card className="h-full transition-all hover:shadow-lg hover:-translate-y-1">
                      {entry.is_featured && (
                        <div className="absolute -right-2 -top-2 z-10">
                          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-yellow-400 text-white shadow-lg">
                            <Star className="h-4 w-4 fill-current" />
                          </div>
                        </div>
                      )}
                      <CardHeader>
                        <CardTitle className="line-clamp-2">{entry.title}</CardTitle>
                        <CardDescription className="flex items-center gap-4 text-sm">
                          <span className="flex items-center gap-1">
                            <Calendar className="h-3 w-3" />
                            {formatDistanceToNow(new Date(entry.createdAt), {
                              addSuffix: true,
                              locale: zhCN
                            })}
                          </span>
                          {entry.theme && (
                            <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs text-amber-700">
                              {entry.theme}
                            </span>
                          )}
                        </CardDescription>
                      </CardHeader>
                      <CardContent>
                        <p className="mb-4 line-clamp-3 text-gray-600">
                          {entry.content}
                        </p>
                        <div className="flex items-center justify-between text-sm text-gray-500">
                          <span className="font-medium">{entry.author_name}</span>
                          <div className="flex items-center gap-3">
                            <span className="flex items-center gap-1">
                              <Eye className="h-3 w-3" />
                              {entry.viewCount}
                            </span>
                            <span className="flex items-center gap-1">
                              <Heart className="h-3 w-3" />
                              {entry.likeCount}
                            </span>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>

              {/* Pagination */}
              {entriesData.total > limit && (
                <div className="mt-8 flex justify-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={page === 1}
                  >
                    上一页
                  </Button>
                  <span className="flex items-center px-4 text-sm text-gray-600">
                    第 {page} 页，共 {Math.ceil(entriesData.total / limit)} 页
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage(p => p + 1)}
                    disabled={page >= Math.ceil(entriesData.total / limit)}
                  >
                    下一页
                  </Button>
                </div>
              )}
            </>
          ) : (
            <div className="text-center py-12">
              <Mail className="mx-auto h-12 w-12 text-gray-300 mb-4" />
              <p className="text-gray-500">暂无馆藏信件</p>
            </div>
          )}
        </section>

        {/* Museum Introduction */}
        <section className="rounded-2xl bg-gradient-to-r from-blue-50 to-purple-50 p-8">
          <div className="mx-auto max-w-3xl text-center">
            <h3 className="mb-4 text-2xl font-bold text-gray-800">关于信件博物馆</h3>
            <p className="mb-6 text-gray-600">
              OpenPenPal 信件博物馆致力于收藏和展示优秀的手写信件作品。每一封被收录的信件都经过精心挑选，
              它们或温暖人心，或发人深省，或记录时代，共同构成了一个充满人文关怀的数字展馆。
            </p>
            <div className="flex flex-wrap justify-center gap-8">
              <div className="text-center">
                <BookOpen className="mx-auto mb-2 h-8 w-8 text-blue-600" />
                <p className="text-sm font-medium">精选展品</p>
              </div>
              <div className="text-center">
                <Users className="mx-auto mb-2 h-8 w-8 text-purple-600" />
                <p className="text-sm font-medium">社区共建</p>
              </div>
              <div className="text-center">
                <Globe className="mx-auto mb-2 h-8 w-8 text-green-600" />
                <p className="text-sm font-medium">文化传承</p>
              </div>
              <div className="text-center">
                <Sparkles className="mx-auto mb-2 h-8 w-8 text-yellow-600" />
                <p className="text-sm font-medium">永恒珍藏</p>
              </div>
            </div>
          </div>
        </section>
      </main>

      <Footer />
    </div>
  )
}