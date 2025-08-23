'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { WelcomeBanner } from '@/components/ui/welcome-banner'
import { 
  Sparkles, 
  Brain, 
  Bot, 
  Users, 
  BookOpen,
  Calendar,
  MessageCircle,
  Settings
} from 'lucide-react'
import { AIWritingInspiration } from '@/components/ai/ai-writing-inspiration'
import { AIDailyInspiration } from '@/components/ai/ai-daily-inspiration'
import { AIPenpalMatch } from '@/components/ai/ai-penpal-match'
import { UnreachableCompanion } from '@/components/ai/unreachable-companion'
import { CharacterStation } from '@/components/ai/character-station'
import { AuthFixBanner } from '@/components/ai/auth-fix-banner'
import { UsageStatsCard } from '@/components/ai/usage-stats-card'
import { useAuth } from '@/contexts/auth-context-new'
import { TokenManager } from '@/lib/auth/cookie-token-manager'

export default function AIPage() {
  // Use a real letter ID from the database for testing
  const [testLetterId] = useState<string>('24b6c37e-b2eb-4639-9bc8-8834cea914e2')
  const [showAuthFix, setShowAuthFix] = useState(false)
  
  const router = useRouter()
  const { isAuthenticated, user, refreshUser } = useAuth()
  
  // Check authentication and detect issues
  useEffect(() => {
    // Add a small delay to allow auth state to initialize
    const timer = setTimeout(() => {
      const checkAuth = () => {
        const token = TokenManager.get()
        const cookieUser = TokenManager.getUser()
        
        // Debug auth state
        console.log('🔧 AI Page Auth Debug:', {
          isAuthenticated,
          hasToken: !!token,
          user: !!user,
          cookieUser: !!cookieUser,
          tokenExpired: token ? TokenManager.isExpired(token) : 'no token'
        })
        
        // If we have a valid token and user in storage but context says not authenticated,
        // it's likely a timing issue - don't redirect
        if (!isAuthenticated && token && cookieUser && !TokenManager.isExpired(token)) {
          console.log('🔧 Token and user exist but auth context not ready - refreshing auth...')
          refreshUser().catch(console.error)
          return
        }
        
        // If not authenticated at all, show the page but with limited functionality
        // Don't redirect to login - AI page should be viewable without auth
        if (!isAuthenticated && !token) {
          console.log('🔧 Not authenticated, showing AI page in view-only mode')
          return
        }
      
      // If authenticated but no token, or token expired, show fix banner
      if (isAuthenticated && (!token || (token && TokenManager.isExpired(token)))) {
        console.log('🔧 Authentication issue detected, showing fix banner')
        setShowAuthFix(true)
        return
      }
      
      // If we have token but no user, show fix banner
      if (token && !user && !TokenManager.isExpired(token)) {
        console.log('🔧 Token exists but no user, showing fix banner')
        setShowAuthFix(true)
        return
      }
      
      // Everything looks good
      setShowAuthFix(false)
    }
    
    checkAuth()
    }, 100) // 100ms delay to allow auth initialization
    
    return () => {
      clearTimeout(timer)
    }
  }, [isAuthenticated, user, router, refreshUser])
  
  // Remove the loading spinner - show the page even when not authenticated

  return (
    <div className="container max-w-6xl mx-auto px-4 py-8">
      <WelcomeBanner />
      
      {/* Auth Fix Banner */}
      {showAuthFix && (
        <AuthFixBanner onFixed={() => setShowAuthFix(false)} />
      )}
      
      {/* Unauthenticated Notice */}
      {!isAuthenticated && (
        <Card className="mb-6 border-amber-200 bg-amber-50">
          <CardContent className="flex items-center justify-between p-4">
            <div className="flex items-center gap-3">
              <Settings className="h-5 w-5 text-amber-600" />
              <div>
                <p className="font-medium text-amber-900">登录后可使用所有AI功能</p>
                <p className="text-sm text-amber-700">
                  目前您正在以访客模式浏览，部分功能需要登录后才能使用
                </p>
              </div>
            </div>
            <Button 
              variant="default" 
              size="sm"
              onClick={() => router.push('/login')}
            >
              立即登录
            </Button>
          </CardContent>
        </Card>
      )}
      
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-4">
          <div className="p-2 bg-gradient-to-br from-purple-500 to-blue-500 rounded-lg">
            <Brain className="h-6 w-6 text-white" />
          </div>
          <div>
            <h1 className="font-serif text-3xl font-bold text-letter-ink">
              AI写信助手
            </h1>
            <p className="text-muted-foreground">
              体验最先进的AI技术，让写信变得更有趣、更有创意
            </p>
          </div>
        </div>
        
        {/* Feature Overview */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          {[
            { icon: Sparkles, label: '云锦传驿', desc: '创意提示', color: 'text-amber-600' },
            { icon: Users, label: '笔友匹配', desc: '智能推荐', color: 'text-purple-600' },
            { icon: Bot, label: 'AI人设', desc: '个性回信', color: 'text-blue-600' },
            { icon: MessageCircle, label: '回信建议', desc: '角度指导', color: 'text-green-600' },
          ].map((feature) => {
            const Icon = feature.icon
            return (
              <Card key={feature.label} className="text-center">
                <CardContent className="pt-6">
                  <Icon className={`h-8 w-8 mx-auto mb-2 ${feature.color}`} />
                  <h3 className="font-semibold">{feature.label}</h3>
                  <p className="text-sm text-muted-foreground">{feature.desc}</p>
                </CardContent>
              </Card>
            )
          })}
        </div>
      </div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* AI功能区域 */}
        <div className="lg:col-span-3">
          <Tabs defaultValue="inspiration" className="space-y-6">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="inspiration" className="gap-2">
            <Sparkles className="h-4 w-4" />
            云锦传驿
          </TabsTrigger>
          <TabsTrigger value="personas" className="gap-2">
            <Bot className="h-4 w-4" />
            云中锦书
          </TabsTrigger>
          <TabsTrigger value="matching" className="gap-2">
            <Users className="h-4 w-4" />
            笔友匹配
          </TabsTrigger>
          <TabsTrigger value="reply" className="gap-2">
            <MessageCircle className="h-4 w-4" />
            角色驿站
          </TabsTrigger>
        </TabsList>

        {/* Writing Inspiration Tab */}
        <TabsContent value="inspiration" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Calendar className="h-5 w-5 text-amber-600" />
                  今日写作主题
                </CardTitle>
                <CardDescription>
                  每日更新的写作主题，为你提供创作灵感
                </CardDescription>
              </CardHeader>
              <CardContent>
                <AIDailyInspiration
                  onSelectPrompt={(prompt) => {
                    console.log('选择了今日主题:', prompt)
                  }}
                />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <BookOpen className="h-5 w-5 text-amber-600" />
                  AI写作提示
                </CardTitle>
                <CardDescription>
                  根据不同主题生成个性化写作建议
                </CardDescription>
              </CardHeader>
              <CardContent>
                <AIWritingInspiration
                  theme="日常生活"
                  onSelectInspiration={(inspiration) => {
                    console.log('选择了写作灵感:', inspiration)
                  }}
                />
              </CardContent>
            </Card>
          </div>

          {/* Usage Tips */}
          <Card className="bg-gradient-to-r from-amber-50 to-orange-50 border-amber-200">
            <CardHeader>
              <CardTitle className="text-lg">💡 使用小贴士</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 text-sm">
              <p>• 每日主题会根据时间、节日和季节自动更新</p>
              <p>• AI写作提示支持多种主题：日常生活、感悟心得、校园生活等</p>
              <p>• 点击任何灵感卡片可直接应用到写信页面</p>
              <p>• 建议将AI提示作为创作起点，融入个人真实感受</p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Cloud Letter (Unreachable Companion) Tab */}
        <TabsContent value="personas" className="space-y-6">
          <UnreachableCompanion />
        </TabsContent>

        {/* Penpal Matching Tab */}
        <TabsContent value="matching" className="space-y-6">
          <div className="max-w-4xl mx-auto">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Users className="h-5 w-5 text-purple-600" />
                  AI智能笔友匹配
                </CardTitle>
                <CardDescription>
                  基于信件内容和兴趣爱好，AI为你推荐最合适的笔友
                </CardDescription>
              </CardHeader>
              <CardContent>
                <AIPenpalMatch
                  letterId={testLetterId}
                  onSelectMatch={(match) => {
                    console.log('选择了笔友:', match)
                  }}
                />
              </CardContent>
            </Card>

            {/* How it works */}
            <Card className="bg-gradient-to-r from-purple-50 to-blue-50 border-purple-200">
              <CardHeader>
                <CardTitle className="text-lg">🤖 AI匹配原理</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3 text-sm">
                <div className="flex items-start gap-3">
                  <Badge variant="secondary" className="mt-1">1</Badge>
                  <div>
                    <h4 className="font-medium">内容分析</h4>
                    <p className="text-muted-foreground">AI分析你的信件内容，提取情感倾向、兴趣点和话题偏好</p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <Badge variant="secondary" className="mt-1">2</Badge>
                  <div>
                    <h4 className="font-medium">智能匹配</h4>
                    <p className="text-muted-foreground">基于兴趣相似度、性格互补性和地理位置进行综合评分</p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <Badge variant="secondary" className="mt-1">3</Badge>
                  <div>
                    <h4 className="font-medium">推荐排序</h4>
                    <p className="text-muted-foreground">按匹配度从高到低排序，并提供详细的匹配理由</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Character Station Tab */}
        <TabsContent value="reply" className="space-y-6">
          <CharacterStation 
            letters={[
              {
                id: '1',
                content: '亲爱的朋友，最近生活怎么样？我这边期末考试快到了，压力有点大。你还记得我们上次聊的那本书吗？我终于看完了，感触很深...',
                senderName: '小明',
                receivedDate: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000)
              },
              {
                id: '2',
                content: '好久不见！听说你最近在学习新技能，进展如何？我最近也在尝试一些新事物，虽然有点困难，但感觉很充实...',
                senderName: '小红',
                receivedDate: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000)
              }
            ]}
          />
        </TabsContent>
      </Tabs>
        </div>

        {/* 右侧边栏 */}
        <div className="lg:col-span-1">
          <div className="sticky top-6 space-y-4">
            <UsageStatsCard />
          </div>
        </div>
      </div>
    </div>
  )
}