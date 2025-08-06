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
  Wand2,
  Lightbulb,
  BookOpen,
  Calendar,
  MessageCircle,
  Settings
} from 'lucide-react'
import { AIWritingInspiration } from '@/components/ai/ai-writing-inspiration'
import { AIDailyInspiration } from '@/components/ai/ai-daily-inspiration'
import { AIPenpalMatch } from '@/components/ai/ai-penpal-match'
import { AIPersonaSelector, AIPersonaPreview } from '@/components/ai/ai-persona-selector'
import { AIReplyGenerator } from '@/components/ai/ai-reply-generator'
import { AIReplyAdvice } from '@/components/ai/ai-reply-advice'
import { CloudLetterCompanion } from '@/components/ai/cloud-letter-companion'
import { AuthFixBanner } from '@/components/ai/auth-fix-banner'
import { UsageStatsCard } from '@/components/ai/usage-stats-card'
import { useAuth } from '@/contexts/auth-context-new'
import { TokenManager } from '@/lib/auth/cookie-token-manager'

export default function AIPage() {
  const [selectedPersona, setSelectedPersona] = useState<string>('friend')
  // Use a real letter ID from the database for testing
  const [testLetterId, setTestLetterId] = useState<string>('24b6c37e-b2eb-4639-9bc8-8834cea914e2')
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
  }, [isAuthenticated, user, router])
  
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

        {/* AI Personas Tab */}
        <TabsContent value="personas" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            {/* Persona Selector */}
            <div className="lg:col-span-4">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Settings className="h-5 w-5" />
                    选择AI笔友
                  </CardTitle>
                  <CardDescription>
                    选择一个长期陪伴你的AI笔友人设
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <AIPersonaSelector
                    value={selectedPersona}
                    onChange={setSelectedPersona}
                  />
                </CardContent>
              </Card>
            </div>

            {/* Cloud Letter Companion */}
            <div className="lg:col-span-8">
              <CloudLetterCompanion selectedPersonaId={selectedPersona} />
            </div>
          </div>
          
          {/* Usage Instructions for Cloud Letter Companion */}
          <Card className="bg-gradient-to-r from-blue-50 to-indigo-50 border-blue-200">
            <CardHeader>
              <CardTitle className="text-lg">☁️ 云中锦书使用说明</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 text-sm">
              <p>• <strong>长期关系：</strong>选择的AI笔友将成为你的长期写信伙伴，保持一致的性格和记忆</p>
              <p>• <strong>个性化交流：</strong>AI会根据你们的对话历史，逐渐了解你的兴趣和写作风格</p>
              <p>• <strong>情感陪伴：</strong>不只是工具，更是一个有温度的写信伙伴，陪伴你的成长历程</p>
              <p>• <strong>多样选择：</strong>诗人、朋友、哲学家等不同类型，总有一个适合你的交流方式</p>
              <p>• <strong>持续互动：</strong>支持长期书信往来，建立深厚的"笔友"情感纽带</p>
            </CardContent>
          </Card>
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

        {/* Reply Generator Tab */}
        <TabsContent value="reply" className="space-y-6">
          <div className="max-w-4xl mx-auto">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Lightbulb className="h-5 w-5 text-green-600" />
                  角色驿站 - 回信角度建议
                </CardTitle>
                <CardDescription>
                  基于不同角色视角，为你的回信提供多样化的思路和建议。让AI帮你从不同角度思考如何回应，而非直接生成内容。
                </CardDescription>
              </CardHeader>
              <CardContent>
                <AIReplyAdvice
                  letterId={testLetterId}
                  letterContent="这里是一封测试信件的内容..."
                  onUseAdvice={(advice) => {
                    console.log('使用了AI回信建议:', advice)
                  }}
                />
              </CardContent>
            </Card>

            {/* Usage Guide for Character Station */}
            <Card className="bg-gradient-to-r from-green-50 to-emerald-50 border-green-200">
              <CardHeader>
                <CardTitle className="text-lg">🏤 角色驿站使用指南</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <p>• <strong>角色视角：</strong>从不同角色（朋友、长辈、同学等）的视角获取回信思路和建议</p>
                <p>• <strong>思路启发：</strong>AI提供回信角度和要点，而非直接生成完整内容，保持回信的原创性</p>
                <p>• <strong>自定义角色：</strong>支持创建个性化角色，根据特定关系和场景定制回信建议</p>
                <p>• <strong>情感引导：</strong>帮助理解来信的情感需求，提供合适的回应策略和语气建议</p>
                <p>• <strong>真实表达：</strong>以建议为基础，融入个人真实感受和具体经历，让回信更有温度</p>
              </CardContent>
            </Card>
          </div>
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