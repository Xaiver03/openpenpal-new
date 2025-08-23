import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import Link from 'next/link'
import { 
  Mail, 
  Heart, 
  Users, 
  Clock,
  Shield,
  Sparkles,
  ArrowLeft,
  Target,
  BookOpen,
  Globe
} from 'lucide-react'

export default function AboutPage() {
  const teamMembers = [
    {
      name: "项目发起团队",
      role: "产品设计与开发", 
      description: "致力于重建校园社交的温度感知"
    },
  ]

  const values = [
    {
      icon: Heart,
      title: "有温度的连接",
      description: "通过手写信件传递真实情感，让每一个字都承载着人与人之间的温度"
    },
    {
      icon: Clock,
      title: "慢节奏社交",
      description: "告别即时反馈的焦虑，重拾等待与期待的美好，让社交回归本质"
    },
    {
      icon: Shield,
      title: "安全与隐私",
      description: "平衡匿名表达的自由与实名投递的安全，保护每一位用户的隐私"
    },
    {
      icon: Users,
      title: "校园社群",
      description: "专注校园场景，为学生群体打造专属的社交平台和文化体验"
    }
  ]

  return (
    <div className="min-h-screen bg-letter-paper">
      <div className="container max-w-4xl mx-auto px-4 py-8">
        {/* Back Button */}
        <div className="mb-8">
          <Button asChild variant="outline" size="sm">
            <Link href="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              返回首页
            </Link>
          </Button>
        </div>

        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="font-serif text-4xl font-bold text-letter-ink mb-4">
            关于 OpenPenPal
          </h1>
          <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
            重建校园社群的温度感知与精神连接
          </p>
        </div>

        {/* Mission */}
        <Card className="mb-12 border-letter-accent/20">
          <CardHeader className="text-center">
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-letter-accent/10 mb-4">
              <Target className="h-8 w-8 text-letter-accent" />
            </div>
            <CardTitle className="font-serif text-2xl">我们的使命</CardTitle>
          </CardHeader>
          <CardContent className="text-center">
            <p className="text-lg leading-relaxed text-muted-foreground">
              在数字化快节奏的今天，我们相信<strong className="text-letter-ink">手写信件</strong>仍然拥有独特的力量。
              OpenPenPal 致力于通过<strong className="text-letter-ink">实体手写信 + 数字跟踪平台</strong>的创新模式，
              为校园学生打造有温度、有深度的慢节奏社交体验，重新定义校园社交的意义。
            </p>
          </CardContent>
        </Card>

        {/* Values */}
        <div className="mb-12">
          <h2 className="font-serif text-3xl font-bold text-letter-ink text-center mb-8">
            我们的价值观
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {values.map((value) => {
              const Icon = value.icon
              return (
                <Card key={value.title} className="border-letter-accent/20 hover:shadow-lg transition-shadow">
                  <CardHeader>
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-full bg-letter-accent/10">
                        <Icon className="h-5 w-5 text-letter-accent" />
                      </div>
                      <CardTitle className="font-serif text-lg">{value.title}</CardTitle>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <CardDescription className="text-base">
                      {value.description}
                    </CardDescription>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        </div>

        {/* How It Started */}
        <Card className="mb-12 border-letter-accent/20">
          <CardHeader className="text-center">
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-letter-accent/10 mb-4">
              <BookOpen className="h-8 w-8 text-letter-accent" />
            </div>
            <CardTitle className="font-serif text-2xl">项目起源</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="prose prose-lg max-w-none text-muted-foreground">
              <p>
                在信息爆炸的时代，我们发现校园社交正在失去它原有的温度。即时消息、社交媒体虽然提供了便利，
                但也让人与人之间的交流变得浮躁和表面化。
              </p>
              <p>
                OpenPenPal 项目源于对<strong className="text-letter-ink">慢节奏社交</strong>的向往和对
                <strong className="text-letter-ink">真实连接</strong>的追求。我们相信，手写信件承载着数字文字无法传递的情感重量，
                每一个笔画都是写信人内心的真实表达。
              </p>
              <p>
                通过结合传统的手写信件和现代的数字跟踪技术，我们希望为校园学生创造一个全新的社交体验：
                既保持了手写信件的温度和深度，又提供了便捷的投递追踪服务。
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Platform Features */}
        <Card className="mb-12 border-letter-accent/20">
          <CardHeader className="text-center">
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-letter-accent/10 mb-4">
              <Globe className="h-8 w-8 text-letter-accent" />
            </div>
            <CardTitle className="font-serif text-2xl">平台特色</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-4">
                <h3 className="font-semibold text-letter-ink flex items-center">
                  <Mail className="h-5 w-5 mr-2 text-letter-accent" />
                  数字化草稿
                </h3>
                <p className="text-muted-foreground">
                  在线编辑信件内容，选择个性化信纸样式，为手写做好准备。
                </p>
                
                <h3 className="font-semibold text-letter-ink flex items-center">
                  <Sparkles className="h-5 w-5 mr-2 text-letter-accent" />
                  编号追踪
                </h3>
                <p className="text-muted-foreground">
                  每封信获得唯一编号和二维码，实现投递状态的实时追踪。
                </p>
              </div>
              
              <div className="space-y-4">
                <h3 className="font-semibold text-letter-ink flex items-center">
                  <Users className="h-5 w-5 mr-2 text-letter-accent" />
                  信使网络
                </h3>
                <p className="text-muted-foreground">
                  校园信使提供安全可靠的投递服务，连接写信人与收信人。
                </p>
                
                <h3 className="font-semibold text-letter-ink flex items-center">
                  <Shield className="h-5 w-5 mr-2 text-letter-accent" />
                  隐私保护
                </h3>
                <p className="text-muted-foreground">
                  匿名写信，实名投递，平衡表达自由与安全保障。
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Team */}
        <Card className="mb-12 border-letter-accent/20">
          <CardHeader className="text-center">
            <CardTitle className="font-serif text-2xl">团队介绍</CardTitle>
            <CardDescription>
              一群相信慢节奏社交力量的年轻人
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {teamMembers.map((member, index) => (
                <div key={index} className="text-center">
                  <div className="mx-auto w-20 h-20 bg-letter-accent/10 rounded-full flex items-center justify-center mb-4">
                    <Users className="h-8 w-8 text-letter-accent" />
                  </div>
                  <h3 className="font-semibold text-letter-ink">{member.name}</h3>
                  <p className="text-sm text-letter-accent">{member.role}</p>
                  <p className="text-sm text-muted-foreground mt-2">{member.description}</p>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Contact & CTA */}
        <Card className="border-letter-accent/20 bg-gradient-to-br from-letter-paper to-white">
          <CardHeader className="text-center">
            <CardTitle className="font-serif text-2xl">加入我们的旅程</CardTitle>
            <CardDescription>
              一起重新定义校园社交，让每一封信都成为连接心灵的桥梁
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center">
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Button asChild size="lg" className="font-serif">
                <Link href="/letters/write">
                  <Mail className="mr-2 h-5 w-5" />
                  写第一封信
                </Link>
              </Button>
              <Button asChild variant="outline" size="lg">
                <Link href="/register">
                  <Users className="mr-2 h-5 w-5" />
                  注册账户
                </Link>
              </Button>
            </div>
            
            <div className="mt-8 pt-8 border-t border-letter-accent/20">
              <p className="text-sm text-muted-foreground">
                OpenPenPal - 让文字重新拥有温度 ✨
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}