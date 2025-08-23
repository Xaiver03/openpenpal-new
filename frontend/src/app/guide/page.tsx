'use client'

import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { 
  BookOpen, 
  PenTool, 
  Send, 
  QrCode,
  Users,
  Trophy,
  ArrowRight,
  ArrowLeft,
  CheckCircle
} from 'lucide-react'

interface GuideStep {
  title: string
  description: string
  icon: React.ElementType
  tips?: string[]
}

export default function GuidePage() {
  const writerSteps: GuideStep[] = [
    {
      title: '注册账号',
      description: '使用学校邮箱注册，验证身份',
      icon: Users,
      tips: [
        '使用真实姓名有助于建立信任',
        '学校代码确保同校交流',
        '一个邮箱只能注册一个账号'
      ]
    },
    {
      title: '撰写信件',
      description: '在线编辑器写下你的心意',
      icon: PenTool,
      tips: [
        '支持多种信纸样式',
        '可以保存草稿随时修改',
        '建议字数300-800字'
      ]
    },
    {
      title: '生成编码',
      description: '获取唯一编号和二维码贴纸',
      icon: QrCode,
      tips: [
        '编号用于追踪和验证',
        '二维码方便收件人查看',
        '请妥善保管编号信息'
      ]
    },
    {
      title: '手写投递',
      description: '抄写到信纸，贴上编号，投入信筒',
      icon: Send,
      tips: [
        '使用钢笔或中性笔',
        '字迹工整便于阅读',
        '记得贴上编号贴纸'
      ]
    }
  ]

  const courierSteps: GuideStep[] = [
    {
      title: '申请加入',
      description: '填写申请表，选择服务区域',
      icon: Users,
      tips: [
        '选择熟悉的区域',
        '诚实填写个人信息',
        '等待24小时审核'
      ]
    },
    {
      title: '接收任务',
      description: '查看待投递任务，规划路线',
      icon: Send,
      tips: [
        '优先处理紧急信件',
        '合理规划投递路线',
        '注意保护信件安全'
      ]
    },
    {
      title: '扫码更新',
      description: '投递时扫码更新状态',
      icon: QrCode,
      tips: [
        '确保扫码成功',
        '选择正确的状态',
        '可以添加投递备注'
      ]
    },
    {
      title: '获得奖励',
      description: '积累积分，提升等级',
      icon: Trophy,
      tips: [
        '完成任务获得积分',
        '保持高完成率',
        '参与特殊活动'
      ]
    }
  ]

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-6xl mx-auto px-4 py-8">
        {/* 返回按钮 */}
        <div className="mb-8">
          <Button asChild variant="outline" size="sm" className="border-amber-300 text-amber-700 hover:bg-amber-50">
            <Link href="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              返回首页
            </Link>
          </Button>
        </div>

        {/* 页面标题 */}
        <div className="text-center mb-12">
          <div className="w-20 h-20 bg-amber-200 rounded-full flex items-center justify-center mx-auto mb-6">
            <BookOpen className="w-10 h-10 text-amber-700" />
          </div>
          <h1 className="font-serif text-4xl font-bold text-amber-900 mb-4">
            使用指南
          </h1>
          <p className="text-xl text-amber-700 max-w-2xl mx-auto">
            详细了解如何使用 OpenPenPal 的各项功能
          </p>
        </div>

        {/* 写信指南 */}
        <section className="mb-16">
          <div className="text-center mb-8">
            <h2 className="font-serif text-2xl font-bold text-amber-900 mb-2">
              写信指南
            </h2>
            <p className="text-amber-700">
              四个简单步骤，开始你的手写信之旅
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {writerSteps.map((step, index) => {
              const Icon = step.icon
              return (
                <Card key={step.title} className="border-amber-200 hover:border-amber-400 hover:shadow-lg transition-all">
                  <CardHeader>
                    <div className="flex items-center justify-between mb-4">
                      <div className="w-12 h-12 bg-amber-600 text-white rounded-full flex items-center justify-center font-bold text-lg">
                        {index + 1}
                      </div>
                      <Icon className="w-8 h-8 text-amber-600" />
                    </div>
                    <CardTitle className="text-lg text-amber-900">{step.title}</CardTitle>
                    <CardDescription className="text-amber-700">
                      {step.description}
                    </CardDescription>
                  </CardHeader>
                  {step.tips && (
                    <CardContent>
                      <ul className="space-y-1">
                        {step.tips.map((tip, i) => (
                          <li key={i} className="flex items-start gap-2 text-sm text-amber-600">
                            <CheckCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
                            <span>{tip}</span>
                          </li>
                        ))}
                      </ul>
                    </CardContent>
                  )}
                </Card>
              )
            })}
          </div>

          <div className="text-center">
            <Button asChild size="lg" className="bg-amber-600 hover:bg-amber-700 text-white">
              <Link href="/letters/write">
                开始写信
                <ArrowRight className="ml-2 h-5 w-5" />
              </Link>
            </Button>
          </div>
        </section>

        {/* 信使指南 */}
        <section>
          <div className="text-center mb-8">
            <h2 className="font-serif text-2xl font-bold text-amber-900 mb-2">
              信使指南
            </h2>
            <p className="text-amber-700">
              成为信使，传递温暖与感动
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {courierSteps.map((step, index) => {
              const Icon = step.icon
              return (
                <Card key={step.title} className="border-amber-200 hover:border-amber-400 hover:shadow-lg transition-all">
                  <CardHeader>
                    <div className="flex items-center justify-between mb-4">
                      <div className="w-12 h-12 bg-orange-600 text-white rounded-full flex items-center justify-center font-bold text-lg">
                        {index + 1}
                      </div>
                      <Icon className="w-8 h-8 text-orange-600" />
                    </div>
                    <CardTitle className="text-lg text-amber-900">{step.title}</CardTitle>
                    <CardDescription className="text-amber-700">
                      {step.description}
                    </CardDescription>
                  </CardHeader>
                  {step.tips && (
                    <CardContent>
                      <ul className="space-y-1">
                        {step.tips.map((tip, i) => (
                          <li key={i} className="flex items-start gap-2 text-sm text-amber-600">
                            <CheckCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
                            <span>{tip}</span>
                          </li>
                        ))}
                      </ul>
                    </CardContent>
                  )}
                </Card>
              )
            })}
          </div>

          <div className="text-center">
            <Button asChild size="lg" className="bg-orange-600 hover:bg-orange-700 text-white">
              <Link href="/courier/apply">
                申请成为信使
                <ArrowRight className="ml-2 h-5 w-5" />
              </Link>
            </Button>
          </div>
        </section>

        {/* 额外提示 */}
        <Card className="mt-16 border-amber-200 bg-gradient-to-r from-amber-50 to-orange-50">
          <CardContent className="p-8">
            <h3 className="text-xl font-bold text-amber-900 mb-4 text-center">
              小贴士
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 text-center">
              <div>
                <div className="w-12 h-12 bg-amber-200 rounded-full flex items-center justify-center mx-auto mb-3">
                  <PenTool className="w-6 h-6 text-amber-700" />
                </div>
                <h4 className="font-semibold text-amber-900 mb-2">用心书写</h4>
                <p className="text-sm text-amber-700">
                  每一个字都承载着你的心意
                </p>
              </div>
              <div>
                <div className="w-12 h-12 bg-amber-200 rounded-full flex items-center justify-center mx-auto mb-3">
                  <Send className="w-6 h-6 text-amber-700" />
                </div>
                <h4 className="font-semibold text-amber-900 mb-2">耐心等待</h4>
                <p className="text-sm text-amber-700">
                  慢递的魅力在于期待
                </p>
              </div>
              <div>
                <div className="w-12 h-12 bg-amber-200 rounded-full flex items-center justify-center mx-auto mb-3">
                  <Trophy className="w-6 h-6 text-amber-700" />
                </div>
                <h4 className="font-semibold text-amber-900 mb-2">珍惜回忆</h4>
                <p className="text-sm text-amber-700">
                  每封信都是独特的记忆
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}