'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  Mail, 
  Phone, 
  MapPin, 
  Clock,
  Send,
  ArrowLeft,
  CheckCircle,
  MessageSquare,
  Users,
  Bug,
  Lightbulb
} from 'lucide-react'

interface ContactForm {
  name: string
  email: string
  subject: string
  category: string
  message: string
}

export default function ContactPage() {
  const [form, setForm] = useState<ContactForm>({
    name: '',
    email: '',
    subject: '',
    category: 'general',
    message: ''
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [submitSuccess, setSubmitSuccess] = useState(false)

  const categories = [
    { value: 'general', label: '一般咨询', icon: MessageSquare },
    { value: 'technical', label: '技术问题', icon: Bug },
    { value: 'feature', label: '功能建议', icon: Lightbulb },
    { value: 'cooperation', label: '合作洽谈', icon: Users }
  ]

  const contactInfo = [
    {
      icon: Mail,
      title: '邮箱联系',
      content: 'support@openpenpal.cn',
      description: '发送邮件给我们，24小时内回复'
    },
    {
      icon: Phone,
      title: '电话咨询',
      content: '400-123-4567',
      description: '工作日 9:00-18:00'
    },
    {
      icon: MapPin,
      title: '办公地址',
      content: '北京市海淀区中关村大街1号',
      description: '欢迎预约参观'
    },
    {
      icon: Clock,
      title: '服务时间',
      content: '周一至周五 9:00-18:00',
      description: '节假日暂停服务'
    }
  ]

  const handleInputChange = (field: keyof ContactForm, value: string) => {
    setForm(prev => ({ ...prev, [field]: value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    
    // 模拟提交
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    setIsSubmitting(false)
    setSubmitSuccess(true)
    
    // 重置表单
    setForm({
      name: '',
      email: '',
      subject: '',
      category: 'general',
      message: ''
    })

    // 3秒后隐藏成功消息
    setTimeout(() => setSubmitSuccess(false), 3000)
  }

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
            <Mail className="w-10 h-10 text-amber-700" />
          </div>
          <h1 className="font-serif text-4xl font-bold text-amber-900 mb-4">
            联系我们
          </h1>
          <p className="text-xl text-amber-700 max-w-2xl mx-auto">
            有任何问题或建议？我们很乐意为您提供帮助
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
          {/* 联系表单 */}
          <div>
            <Card className="border-amber-200 shadow-lg">
              <CardHeader>
                <CardTitle className="text-amber-900">发送消息</CardTitle>
                <CardDescription>
                  填写下面的表单，我们会尽快回复您
                </CardDescription>
              </CardHeader>
              <CardContent>
                {submitSuccess && (
                  <Alert className="mb-6 border-green-200 bg-green-50">
                    <CheckCircle className="h-4 w-4 text-green-600" />
                    <AlertDescription className="text-green-700">
                      消息发送成功！我们会在24小时内回复您。
                    </AlertDescription>
                  </Alert>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="name">姓名 *</Label>
                      <Input
                        id="name"
                        value={form.name}
                        onChange={(e) => handleInputChange('name', e.target.value)}
                        required
                        className="border-amber-300 focus:border-amber-500"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="email">邮箱 *</Label>
                      <Input
                        id="email"
                        type="email"
                        value={form.email}
                        onChange={(e) => handleInputChange('email', e.target.value)}
                        required
                        className="border-amber-300 focus:border-amber-500"
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="category">咨询类型</Label>
                    <select
                      id="category"
                      value={form.category}
                      onChange={(e) => handleInputChange('category', e.target.value)}
                      className="w-full p-2 border border-amber-300 rounded-md focus:border-amber-500 focus:outline-none"
                    >
                      {categories.map(cat => (
                        <option key={cat.value} value={cat.value}>
                          {cat.label}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="subject">主题 *</Label>
                    <Input
                      id="subject"
                      value={form.subject}
                      onChange={(e) => handleInputChange('subject', e.target.value)}
                      required
                      className="border-amber-300 focus:border-amber-500"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="message">详细描述 *</Label>
                    <Textarea
                      id="message"
                      value={form.message}
                      onChange={(e) => handleInputChange('message', e.target.value)}
                      required
                      rows={6}
                      className="border-amber-300 focus:border-amber-500"
                      placeholder="请详细描述您的问题或建议..."
                    />
                  </div>

                  <Button 
                    type="submit" 
                    disabled={isSubmitting}
                    className="w-full bg-amber-600 hover:bg-amber-700 text-white"
                  >
                    {isSubmitting ? (
                      <>发送中...</>
                    ) : (
                      <>
                        <Send className="mr-2 h-4 w-4" />
                        发送消息
                      </>
                    )}
                  </Button>
                </form>
              </CardContent>
            </Card>
          </div>

          {/* 联系信息 */}
          <div className="space-y-6">
            <div>
              <h2 className="text-2xl font-bold text-amber-900 mb-6">联系方式</h2>
              <div className="space-y-4">
                {contactInfo.map((info, index) => {
                  const Icon = info.icon
                  return (
                    <Card key={index} className="border-amber-200">
                      <CardContent className="p-4">
                        <div className="flex items-start gap-4">
                          <div className="w-12 h-12 bg-amber-100 rounded-lg flex items-center justify-center flex-shrink-0">
                            <Icon className="w-6 h-6 text-amber-600" />
                          </div>
                          <div>
                            <h3 className="font-semibold text-amber-900 mb-1">
                              {info.title}
                            </h3>
                            <p className="text-lg text-amber-800 mb-1">
                              {info.content}
                            </p>
                            <p className="text-sm text-amber-600">
                              {info.description}
                            </p>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  )
                })}
              </div>
            </div>

            {/* 快速链接 */}
            <Card className="border-amber-200">
              <CardHeader>
                <CardTitle className="text-amber-900">快速链接</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <Link 
                  href="/help" 
                  className="flex items-center gap-3 p-3 rounded-md hover:bg-amber-50 transition-colors"
                >
                  <MessageSquare className="w-5 h-5 text-amber-600" />
                  <span className="text-amber-700">帮助中心</span>
                </Link>
                <Link 
                  href="/guide" 
                  className="flex items-center gap-3 p-3 rounded-md hover:bg-amber-50 transition-colors"
                >
                  <Lightbulb className="w-5 h-5 text-amber-600" />
                  <span className="text-amber-700">使用指南</span>
                </Link>
                <Link 
                  href="/faq" 
                  className="flex items-center gap-3 p-3 rounded-md hover:bg-amber-50 transition-colors"
                >
                  <Users className="w-5 h-5 text-amber-600" />
                  <span className="text-amber-700">常见问题</span>
                </Link>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  )
}