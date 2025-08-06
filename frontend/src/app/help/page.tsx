'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { 
  HelpCircle, 
  Mail, 
  Send, 
  QrCode,
  Users,
  Clock,
  Shield,
  ChevronDown,
  ChevronUp,
  ArrowLeft,
  MessageSquare
} from 'lucide-react'

interface FAQItem {
  question: string
  answer: string
  category: 'general' | 'letter' | 'courier' | 'privacy'
}

export default function HelpPage() {
  const [expandedItems, setExpandedItems] = useState<Set<number>>(new Set())

  const toggleItem = (index: number) => {
    const newExpanded = new Set(expandedItems)
    if (newExpanded.has(index)) {
      newExpanded.delete(index)
    } else {
      newExpanded.add(index)
    }
    setExpandedItems(newExpanded)
  }

  const faqItems: FAQItem[] = [
    {
      question: 'OpenPenPal 是什么？',
      answer: 'OpenPenPal 是一个结合传统手写信和数字技术的校园社交平台。我们提供数字化的信件编辑和跟踪功能，同时保留手写信的温度和仪式感。',
      category: 'general'
    },
    {
      question: '如何开始写第一封信？',
      answer: '1. 注册并登录账号\n2. 点击"写信去"按钮\n3. 在线编辑信件内容\n4. 生成唯一编号和二维码\n5. 手写到信纸上\n6. 贴上编号贴纸\n7. 投入校园信筒',
      category: 'letter'
    },
    {
      question: '信件编号有什么用？',
      answer: '每封信都有唯一的编号，用于：\n- 追踪投递状态\n- 保护隐私（收件人可以通过编号查看数字版本）\n- 验证信件真实性\n- 统计投递数据',
      category: 'letter'
    },
    {
      question: '如何成为信使？',
      answer: '1. 在信使中心点击"申请成为信使"\n2. 填写申请表，选择服务区域\n3. 等待审核（通常24小时内）\n4. 审核通过后完成新手任务\n5. 开始接收投递任务',
      category: 'courier'
    },
    {
      question: '信使有什么福利？',
      answer: '- 积分奖励（可兑换礼品）\n- 等级成长体系\n- 社区认可和荣誉\n- 优先体验新功能\n- 定期信使聚会活动',
      category: 'courier'
    },
    {
      question: '我的个人信息安全吗？',
      answer: '我们非常重视用户隐私：\n- 所有数据加密存储\n- 不会分享给第三方\n- 可以选择匿名投递\n- 支持账号注销和数据删除\n- 符合相关隐私法规',
      category: 'privacy'
    },
    {
      question: '收到信件后如何查看数字版本？',
      answer: '1. 扫描信件上的二维码\n2. 或在平台输入信件编号\n3. 即可查看信件的数字版本\n4. 支持标记已读和回复',
      category: 'letter'
    },
    {
      question: '信件投递需要多长时间？',
      answer: '通常情况下：\n- 同校区：1-2天\n- 跨校区：2-3天\n- 特殊情况可能延长\n- 可实时查看投递进度',
      category: 'letter'
    }
  ]

  const categories = [
    { id: 'general', name: '基础问题', icon: HelpCircle, color: 'text-blue-600' },
    { id: 'letter', name: '写信相关', icon: Mail, color: 'text-green-600' },
    { id: 'courier', name: '信使系统', icon: Send, color: 'text-orange-600' },
    { id: 'privacy', name: '隐私安全', icon: Shield, color: 'text-red-600' }
  ]

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-4xl mx-auto px-4 py-8">
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
            <HelpCircle className="w-10 h-10 text-amber-700" />
          </div>
          <h1 className="font-serif text-4xl font-bold text-amber-900 mb-4">
            帮助中心
          </h1>
          <p className="text-xl text-amber-700 max-w-2xl mx-auto">
            在这里找到关于 OpenPenPal 的所有答案
          </p>
        </div>

        {/* 快速入口 */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-12">
          <Card className="border-amber-200 hover:border-amber-400 hover:shadow-lg transition-all cursor-pointer">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center">
                  <Mail className="w-6 h-6 text-blue-600" />
                </div>
                <div>
                  <CardTitle className="text-lg">写信指南</CardTitle>
                  <CardDescription>了解如何写你的第一封信</CardDescription>
                </div>
              </div>
            </CardHeader>
          </Card>

          <Card className="border-amber-200 hover:border-amber-400 hover:shadow-lg transition-all cursor-pointer">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center">
                  <Users className="w-6 h-6 text-orange-600" />
                </div>
                <div>
                  <CardTitle className="text-lg">成为信使</CardTitle>
                  <CardDescription>加入我们的投递团队</CardDescription>
                </div>
              </div>
            </CardHeader>
          </Card>
        </div>

        {/* 分类标签 */}
        <div className="flex flex-wrap gap-2 mb-8 justify-center">
          {categories.map(category => {
            const Icon = category.icon
            return (
              <Button
                key={category.id}
                variant="outline"
                size="sm"
                className="border-amber-300 hover:bg-amber-100"
              >
                <Icon className={`mr-2 h-4 w-4 ${category.color}`} />
                {category.name}
              </Button>
            )
          })}
        </div>

        {/* FAQ 列表 */}
        <div className="space-y-4">
          {faqItems.map((item, index) => {
            const isExpanded = expandedItems.has(index)
            const categoryInfo = categories.find(c => c.id === item.category)
            const Icon = categoryInfo?.icon || HelpCircle
            
            return (
              <Card 
                key={index}
                className="border-amber-200 hover:border-amber-300 transition-all"
              >
                <CardHeader 
                  className="cursor-pointer"
                  onClick={() => toggleItem(index)}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3 flex-1">
                      <div className={`w-8 h-8 rounded-lg bg-amber-100 flex items-center justify-center flex-shrink-0`}>
                        <Icon className={`w-4 h-4 ${categoryInfo?.color}`} />
                      </div>
                      <h3 className="font-medium text-amber-900">{item.question}</h3>
                    </div>
                    {isExpanded ? (
                      <ChevronUp className="h-5 w-5 text-amber-600" />
                    ) : (
                      <ChevronDown className="h-5 w-5 text-amber-600" />
                    )}
                  </div>
                </CardHeader>
                {isExpanded && (
                  <CardContent className="pt-0">
                    <p className="text-amber-700 whitespace-pre-line pl-11">
                      {item.answer}
                    </p>
                  </CardContent>
                )}
              </Card>
            )
          })}
        </div>

        {/* 联系我们 */}
        <Card className="mt-12 border-amber-200 bg-gradient-to-r from-amber-50 to-orange-50">
          <CardContent className="p-8 text-center">
            <MessageSquare className="w-12 h-12 text-amber-600 mx-auto mb-4" />
            <h3 className="text-xl font-bold text-amber-900 mb-2">还有其他问题？</h3>
            <p className="text-amber-700 mb-6">
              我们的支持团队随时为您提供帮助
            </p>
            <div className="flex gap-4 justify-center">
              <Button asChild className="bg-amber-600 hover:bg-amber-700 text-white">
                <Link href="/contact">
                  联系我们
                </Link>
              </Button>
              <Button asChild variant="outline" className="border-amber-300 text-amber-700 hover:bg-amber-50">
                <Link href="/guide">
                  查看指南
                </Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}