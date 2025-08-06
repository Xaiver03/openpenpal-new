'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { LetterService } from '@/lib/services/letter-service'
import type { Letter } from '@/lib/services/letter-service'
import { useAuth, usePermissions, useCourier } from '@/stores/user-store'
import { getCourierLevelManagementPath } from '@/constants/roles'
import { 
  Mail, 
  Send, 
  Inbox, 
  Heart, 
  Users, 
  Clock,
  Shield,
  Crown,
  Sparkles,
  ArrowRight,
  PenTool,
  BookOpen,
  Globe,
  Star,
  MessageCircle,
  MapPin,
  ChevronRight
} from 'lucide-react'

export default function HomePage() {
  const [currentStory, setCurrentStory] = useState(0)
  const [publicLetters, setPublicLetters] = useState<Letter[]>([])
  const [isLoadingLetters, setIsLoadingLetters] = useState(true)
  
  // User state
  const { isAuthenticated } = useAuth()
  const { canAccessAdmin } = usePermissions()
  const { courierInfo, isCourier, levelName } = useCourier()

  const features = [
    {
      icon: PenTool,
      title: '写一封信给过去或未来',
      description: '用墨水记录此刻的心情，寄给未来的自己或陌生的朋友',
      href: '/write',
      buttonText: '立即写信',
      color: 'amber'
    },
    {
      icon: Mail,
      title: '打开别人的来信',
      description: '每一封信都是一个故事，每一个故事都值得被倾听',
      href: '/mailbox',
      buttonText: '去查收',
      color: 'orange'
    },
    {
      icon: Send,
      title: '成为投递的连接者',
      description: '加入信使队伍，成为连接心灵的桥梁',
      href: '/courier',
      buttonText: '加入信使',
      color: 'yellow'
    },
    {
      icon: BookOpen,
      title: '与文字爱好者共创',
      description: '在写作广场分享你的文字，发现更多有趣的灵魂',
      href: '/plaza',
      buttonText: '浏览作品',
      color: 'lime'
    }
  ]

  const stories = [
    {
      content: "收到一封来自陌生人的信，里面写着'你并不孤单'，那一刻眼泪就掉下来了。",
      author: "来自北京的小雨",
      location: "北京大学"
    },
    {
      content: "给三年后的自己写了一封信，希望到时候能成为更好的人。",
      author: "来自上海的阿明",
      location: "复旦大学"
    },
    {
      content: "当信使三个月了，每次看到收信人开心的表情，都觉得很有意义。",
      author: "来自广州的小李",
      location: "中山大学"
    }
  ]

  // 加载公开信件
  useEffect(() => {
    const loadPublicLetters = async () => {
      setIsLoadingLetters(true)
      try {
        // 使用公开信件API
        const response = await LetterService.getPublicLetters({ limit: 3 })
        if (response.data?.data) {
          // 转换数据格式
          const letters = response.data.data.map((letter: any) => ({
            ...letter,
            sender_name: letter.user?.nickname || '匿名用户',
            likeCount: 0, // API暂时没有返回点赞数
          }))
          setPublicLetters(letters)
        }
      } catch (error) {
        console.error('Failed to load public letters:', error)
        // 使用默认数据作为回退
        setPublicLetters([
          {
            id: '1',
            title: "写给2027年的自己",
            content: "亲爱的未来的我，现在是2024年，我刚开始大学生活...",
            sender_name: "匿名用户",
            likeCount: 42,
            createdAt: "2024-01-15",
          } as any,
          {
            id: '2',
            title: "给远方朋友的一封信",
            content: "很久没有联系了，想起我们一起度过的那些日子...",
            sender_name: "匿名用户",
            likeCount: 28,
            createdAt: "2024-01-14",
          } as any,
          {
            id: '3',
            title: "致迷茫的你",
            content: "如果你正在经历困难，请相信一切都会过去的...",
            sender_name: "匿名用户",
            likeCount: 67,
            createdAt: "2024-01-13",
          } as any,
        ])
      } finally {
        setIsLoadingLetters(false)
      }
    }
    loadPublicLetters()
  }, [])

  // 自动轮播故事
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentStory((prev) => (prev + 1) % stories.length)
    }, 4000)
    return () => clearInterval(timer)
  }, [stories.length])

  return (
    <div className="min-h-screen flex flex-col bg-letter-paper">
      <Header />
      
      <main className="flex-1">
        {/* Hero Section - 全屏 Banner & 价值主张 */}
        <section className="relative min-h-screen flex items-center overflow-hidden bg-gradient-to-br from-amber-50 via-orange-50 to-yellow-50">
          {/* 背景装饰 */}
          <div className="absolute inset-0 bg-[url('/paper-texture.svg')] opacity-5" />
          <div className="absolute top-20 left-20 w-32 h-32 bg-amber-200/20 rounded-full blur-3xl" />
          <div className="absolute bottom-20 right-20 w-48 h-48 bg-orange-200/20 rounded-full blur-3xl" />
          
          <div className="container relative px-4 grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
            {/* 左侧：主标语 */}
            <div className="text-center lg:text-left">
              <div className="inline-block px-4 py-2 bg-amber-100 rounded-full text-amber-800 text-sm font-medium mb-6">
                ✨ 重新定义校园社交
              </div>
              <h1 className="font-serif text-4xl md:text-5xl lg:text-6xl font-bold text-amber-900 mb-6 leading-tight">
                一封手写信，
                <br />
                <span className="text-amber-600">慢下来连结世界</span>
              </h1>
              <p className="text-xl text-amber-700 mb-8 leading-relaxed max-w-xl">
                在快节奏的数字时代，让我们重新拾起笔墨，
                用最真挚的文字传递温暖，用最慢的方式建立最深的连接。
              </p>
              <div className="flex flex-col sm:flex-row gap-4">
                <Button asChild size="lg" className="bg-amber-600 hover:bg-amber-700 text-white font-serif text-lg px-8 py-6">
                  <Link href="/write">
                    <PenTool className="mr-2 h-6 w-6" />
                    写信去
                  </Link>
                </Button>
                <Button asChild variant="outline" size="lg" className="border-amber-300 text-amber-700 hover:bg-amber-50 font-serif text-lg px-8 py-6">
                  <Link href="/courier">
                    <Send className="mr-2 h-6 w-6" />
                    加入信使
                  </Link>
                </Button>
              </div>
            </div>

            {/* 右侧：动态动效 */}
            <div className="relative">
              <div className="relative w-full max-w-md mx-auto">
                {/* 主信封 */}
                <div className="relative bg-white rounded-lg shadow-2xl p-8 transform rotate-3 hover:rotate-0 transition-transform duration-500">
                  <div className="absolute top-0 left-0 w-full h-2 bg-amber-400 rounded-t-lg"></div>
                  <div className="space-y-4">
                    <div className="h-4 bg-amber-100 rounded w-3/4"></div>
                    <div className="h-4 bg-amber-100 rounded w-full"></div>
                    <div className="h-4 bg-amber-100 rounded w-2/3"></div>
                  </div>
                  <div className="mt-6 flex justify-end">
                    <div className="w-8 h-8 bg-amber-200 rounded-full"></div>
                  </div>
                </div>
                
                {/* 飞舞的小信件 */}
                <div className="absolute -top-10 -right-10 w-16 h-16 bg-white rounded-lg shadow-lg transform rotate-12 animate-pulse">
                  <Mail className="w-8 h-8 text-amber-600 m-4" />
                </div>
                <div className="absolute -bottom-10 -left-10 w-12 h-12 bg-orange-100 rounded-lg shadow-lg transform -rotate-12 animate-bounce">
                  <Heart className="w-6 h-6 text-orange-600 m-3" />
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Feature Highlights - 核心功能引导区 */}
        <section className="py-20 bg-white">
          <div className="container px-4">
            <div className="text-center mb-16">
              <h2 className="font-serif text-3xl md:text-4xl font-bold text-amber-900 mb-4">
                四种方式，连接心灵
              </h2>
              <p className="text-xl text-amber-700 max-w-2xl mx-auto">
                每一种体验都是一次心灵的旅程
              </p>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
              {features.map((feature, index) => {
                const Icon = feature.icon
                return (
                  <Card key={feature.title} className="group border-amber-200 hover:border-amber-400 hover:shadow-xl transition-all duration-300 hover:-translate-y-2">
                    <CardHeader className="text-center pb-4">
                      <div className={`mx-auto w-16 h-16 bg-${feature.color}-100 rounded-2xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform`}>
                        <Icon className={`w-8 h-8 text-${feature.color}-600`} />
                      </div>
                      <CardTitle className="font-serif text-xl text-amber-900">{feature.title}</CardTitle>
                    </CardHeader>
                    <CardContent className="text-center">
                      <CardDescription className="text-amber-700 mb-6 text-base leading-relaxed">
                        {feature.description}
                      </CardDescription>
                      <Button asChild className="w-full bg-amber-600 hover:bg-amber-700 text-white">
                        <Link href={feature.href}>
                          {feature.buttonText}
                          <ChevronRight className="ml-2 h-4 w-4" />
                        </Link>
                      </Button>
                    </CardContent>
                  </Card>
                )
              })}
            </div>
          </div>
        </section>

        {/* Story & Vision - 慢社交故事区 */}
        <section className="py-20 bg-gradient-to-br from-amber-50 to-orange-50">
          <div className="container px-4">
            <div className="text-center mb-16">
              <h2 className="font-serif text-3xl md:text-4xl font-bold text-amber-900 mb-4">
                来自各地的温暖故事
              </h2>
              <p className="text-xl text-amber-700 max-w-2xl mx-auto">
                每一个故事都见证着真实连接的力量
              </p>
            </div>

            <div className="max-w-4xl mx-auto">
              {/* 故事轮播 */}
              <Card className="border-amber-200 bg-white/80 backdrop-blur-sm shadow-xl">
                <CardContent className="p-8 text-center">
                  <div className="mb-6">
                    <MessageCircle className="w-12 h-12 text-amber-600 mx-auto mb-4" />
                    <blockquote className="text-2xl font-serif text-amber-900 leading-relaxed">
                      "{stories[currentStory].content}"
                    </blockquote>
                  </div>
                  <div className="flex items-center justify-center gap-2 text-amber-700">
                    <span className="font-medium">{stories[currentStory].author}</span>
                    <span>·</span>
                    <MapPin className="w-4 h-4" />
                    <span>{stories[currentStory].location}</span>
                  </div>
                </CardContent>
              </Card>

              {/* 轮播指示器 */}
              <div className="flex justify-center mt-6 gap-2">
                {stories.map((_, index) => (
                  <button
                    key={index}
                    onClick={() => setCurrentStory(index)}
                    className={`w-3 h-3 rounded-full transition-colors ${
                      index === currentStory ? 'bg-amber-600' : 'bg-amber-300'
                    }`}
                  />
                ))}
              </div>

              {/* 核心理念 */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mt-16">
                <div className="text-center">
                  <Clock className="w-12 h-12 text-amber-600 mx-auto mb-4" />
                  <h3 className="font-serif text-xl font-bold text-amber-900 mb-2">慢节奏</h3>
                  <p className="text-amber-700">告别即时反馈的焦虑，重拾等待的美好</p>
                </div>
                <div className="text-center">
                  <Heart className="w-12 h-12 text-amber-600 mx-auto mb-4" />
                  <h3 className="font-serif text-xl font-bold text-amber-900 mb-2">真实感</h3>
                  <p className="text-amber-700">手写的温度，墨水的香气，最真挚的表达</p>
                </div>
                <div className="text-center">
                  <Users className="w-12 h-12 text-amber-600 mx-auto mb-4" />
                  <h3 className="font-serif text-xl font-bold text-amber-900 mb-2">深连接</h3>
                  <p className="text-amber-700">跨越时空的心灵对话，建立持久的情感纽带</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Public Letter Wall - 精选公开信件墙 */}
        <section className="py-20 bg-white">
          <div className="container px-4">
            <div className="text-center mb-16">
              <h2 className="font-serif text-3xl md:text-4xl font-bold text-amber-900 mb-4">
                信件博物馆
              </h2>
              <p className="text-xl text-amber-700 max-w-2xl mx-auto">
                那些被时光记录的美好文字
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
              {isLoadingLetters ? (
                // 加载骨架屏
                Array.from({ length: 3 }).map((_, i) => (
                  <Card key={i} className="border-amber-200">
                    <CardHeader className="pb-3">
                      <div className="flex items-center justify-between">
                        <div className="w-16 h-6 bg-amber-100 rounded-full animate-pulse" />
                        <div className="w-12 h-4 bg-amber-100 rounded animate-pulse" />
                      </div>
                      <div className="w-3/4 h-6 bg-amber-100 rounded mt-3 animate-pulse" />
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2 mb-4">
                        <div className="w-full h-4 bg-amber-100 rounded animate-pulse" />
                        <div className="w-5/6 h-4 bg-amber-100 rounded animate-pulse" />
                        <div className="w-4/6 h-4 bg-amber-100 rounded animate-pulse" />
                      </div>
                      <div className="flex items-center justify-between">
                        <div className="w-20 h-4 bg-amber-100 rounded animate-pulse" />
                        <div className="w-24 h-4 bg-amber-100 rounded animate-pulse" />
                      </div>
                    </CardContent>
                  </Card>
                ))
              ) : publicLetters.length > 0 ? (
                publicLetters.map((letter) => (
                  <Card key={letter.id} className="group border-amber-200 hover:border-amber-400 hover:shadow-lg transition-all duration-300">
                    <CardHeader className="pb-3">
                      <div className="flex items-center justify-between">
                        <span className="px-3 py-1 bg-amber-100 text-amber-800 text-sm rounded-full">
                          {letter.tags?.[0] || '公开信'}
                        </span>
                        <div className="flex items-center gap-1 text-amber-600">
                          <Heart className="w-4 h-4" />
                          <span className="text-sm">{letter.likeCount || 0}</span>
                        </div>
                      </div>
                      <CardTitle className="font-serif text-lg text-amber-900 line-clamp-2">
                        {letter.title || '无标题'}
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <p className="text-amber-700 line-clamp-3 mb-4">
                        {letter.content}
                      </p>
                      <div className="flex items-center justify-between text-sm text-amber-600">
                        <span>{letter.sender_name || '匿名用户'}</span>
                        <span>{new Date(letter.createdAt).toLocaleDateString('zh-CN')}</span>
                      </div>
                    </CardContent>
                  </Card>
                ))
              ) : (
                <div className="col-span-3 text-center py-8 text-amber-700">
                  暂无公开信件
                </div>
              )}
            </div>

            <div className="text-center mt-12">
              <Button asChild variant="outline" size="lg" className="border-amber-300 text-amber-700 hover:bg-amber-50">
                <Link href="/museum">
                  查看更多信件
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Link>
              </Button>
            </div>
          </div>
        </section>

        {/* User Dashboard - 用户专属面板 */}
        {isAuthenticated && (isCourier || canAccessAdmin()) && (
          <section className="py-20 bg-gradient-to-br from-blue-50 to-indigo-50">
            <div className="container px-4">
              <div className="max-w-4xl mx-auto">
                <div className="text-center mb-12">
                  <h2 className="font-serif text-3xl md:text-4xl font-bold text-blue-900 mb-4">
                    您的管理中心
                  </h2>
                  <p className="text-xl text-blue-700">
                    快速访问您的专属功能
                  </p>
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                  {/* 信使管理入口 */}
                  {isCourier && (
                    <Card className="border-blue-200 hover:border-blue-400 hover:shadow-xl transition-all duration-300">
                      <CardHeader className="text-center pb-4">
                        <div className="mx-auto w-16 h-16 bg-blue-100 rounded-2xl flex items-center justify-center mb-4">
                          <Send className="w-8 h-8 text-blue-600" />
                        </div>
                        <CardTitle className="font-serif text-xl text-blue-900">信使中心</CardTitle>
                        <CardDescription className="text-blue-700">
                          查看任务、管理投递、更新状态
                        </CardDescription>
                      </CardHeader>
                      <CardContent className="text-center">
                        <Button asChild className="w-full bg-blue-600 hover:bg-blue-700 text-white mb-3">
                          <Link href="/courier">
                            进入信使中心
                            <ArrowRight className="ml-2 h-4 w-4" />
                          </Link>
                        </Button>
                        {courierInfo && (
                          <div className="text-sm text-blue-600">
                            当前等级: {levelName || `${courierInfo.level}级信使`}
                          </div>
                        )}
                      </CardContent>
                    </Card>
                  )}
                  
                  {/* 各级信使管理入口 */}
                  {courierInfo && courierInfo.level > 1 && (
                    <Card className="border-purple-200 hover:border-purple-400 hover:shadow-xl transition-all duration-300">
                      <CardHeader className="text-center pb-4">
                        <div className="mx-auto w-16 h-16 bg-purple-100 rounded-2xl flex items-center justify-center mb-4">
                          {courierInfo.level === 4 ? (
                            <Crown className="w-8 h-8 text-purple-600" />
                          ) : (
                            <Users className="w-8 h-8 text-purple-600" />
                          )}
                        </div>
                        <CardTitle className="font-serif text-xl text-purple-900">
                          {levelName?.replace(/（.*?）/, '')}管理
                        </CardTitle>
                        <CardDescription className="text-purple-700">
                          管理下级信使、分配任务、查看数据
                        </CardDescription>
                      </CardHeader>
                      <CardContent className="text-center">
                        <Button asChild className="w-full bg-purple-600 hover:bg-purple-700 text-white">
                          <Link href={getCourierLevelManagementPath(courierInfo.level)}>
                            进入管理面板
                            <ArrowRight className="ml-2 h-4 w-4" />
                          </Link>
                        </Button>
                      </CardContent>
                    </Card>
                  )}
                  
                  {/* 管理员入口 */}
                  {canAccessAdmin() && (
                    <Card className="border-red-200 hover:border-red-400 hover:shadow-xl transition-all duration-300">
                      <CardHeader className="text-center pb-4">
                        <div className="mx-auto w-16 h-16 bg-red-100 rounded-2xl flex items-center justify-center mb-4">
                          <Shield className="w-8 h-8 text-red-600" />
                        </div>
                        <CardTitle className="font-serif text-xl text-red-900">管理控制台</CardTitle>
                        <CardDescription className="text-red-700">
                          系统管理、用户管理、数据分析
                        </CardDescription>
                      </CardHeader>
                      <CardContent className="text-center">
                        <Button asChild className="w-full bg-red-600 hover:bg-red-700 text-white">
                          <Link href="/admin">
                            进入控制台
                            <ArrowRight className="ml-2 h-4 w-4" />
                          </Link>
                        </Button>
                      </CardContent>
                    </Card>
                  )}
                </div>
              </div>
            </div>
          </section>
        )}

        {/* Join Us - 加入我们 & 信使入口 */}
        <section className="py-20 bg-gradient-to-br from-amber-100 to-orange-100">
          <div className="container px-4">
            <div className="max-w-4xl mx-auto text-center">
              <h2 className="font-serif text-3xl md:text-4xl font-bold text-amber-900 mb-6">
                {isAuthenticated && isCourier ? '信使成长之路' : '成为连接世界的信使'}
              </h2>
              <p className="text-xl text-amber-700 mb-12 max-w-2xl mx-auto">
                {isAuthenticated && isCourier 
                  ? '继续您的信使旅程，帮助更多人传递温暖'
                  : '加入我们的信使网络，成为传递温暖的使者，在帮助他人的同时收获成长与友谊'
                }
              </p>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
                <div className="text-center">
                  <div className="w-16 h-16 bg-amber-200 rounded-full flex items-center justify-center mx-auto mb-4">
                    <Star className="w-8 h-8 text-amber-700" />
                  </div>
                  <h3 className="font-semibold text-amber-900 mb-2">成长体系</h3>
                  <p className="text-amber-700 text-sm">从新手信使到资深导师，见证自己的成长</p>
                </div>
                <div className="text-center">
                  <div className="w-16 h-16 bg-orange-200 rounded-full flex items-center justify-center mx-auto mb-4">
                    <Heart className="w-8 h-8 text-orange-700" />
                  </div>
                  <h3 className="font-semibold text-amber-900 mb-2">温暖奖励</h3>
                  <p className="text-amber-700 text-sm">每一次投递都有意义，收获感谢与友谊</p>
                </div>
                <div className="text-center">
                  <div className="w-16 h-16 bg-yellow-200 rounded-full flex items-center justify-center mx-auto mb-4">
                    <Users className="w-8 h-8 text-yellow-700" />
                  </div>
                  <h3 className="font-semibold text-amber-900 mb-2">社区归属</h3>
                  <p className="text-amber-700 text-sm">加入温暖的信使大家庭，结识志同道合的朋友</p>
                </div>
              </div>

              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                {!isAuthenticated || !isCourier ? (
                  <Button asChild size="lg" className="bg-amber-600 hover:bg-amber-700 text-white font-serif px-8">
                    <Link href="/courier">
                      <Send className="mr-2 h-5 w-5" />
                      申请成为信使
                    </Link>
                  </Button>
                ) : (
                  <Button asChild size="lg" className="bg-amber-600 hover:bg-amber-700 text-white font-serif px-8">
                    <Link href="/courier">
                      <Send className="mr-2 h-5 w-5" />
                      继续信使之路
                    </Link>
                  </Button>
                )}
                <Button asChild variant="outline" size="lg" className="border-amber-300 text-amber-700 hover:bg-amber-50 font-serif px-8">
                  <Link href="/about">
                    <Globe className="mr-2 h-5 w-5" />
                    了解合作方式
                  </Link>
                </Button>
              </div>
            </div>
          </div>
        </section>
      </main>

      <Footer />
    </div>
  )
}