'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import Link from 'next/link'
import { 
  ArrowLeft, 
  Mail, 
  MapPin,
  Clock,
  Users,
  Check,
  AlertCircle,
  Heart,
  Send
} from 'lucide-react'

interface ApplicationForm {
  name: string
  contact: string
  school: string
  zone: string
  hasPrinter: string
  selfIntro: string
  canMentor: string
  weeklyHours: number
  maxDailyTasks: number
  transportMethod: string
  timeSlots: string[]
}

export default function CourierApplyPage() {
  const router = useRouter()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null)
  
  const [form, setForm] = useState<ApplicationForm>({
    name: '',
    contact: '',
    school: '',
    zone: '',
    hasPrinter: '',
    selfIntro: '',
    canMentor: '',
    weeklyHours: 5,
    maxDailyTasks: 10,
    transportMethod: '',
    timeSlots: []
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    
    try {
      // TODO: 连接后端API
      const response = await fetch('/api/v1/courier/apply', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(form),
      })
      
      if (response.ok) {
        setMessage({ 
          type: 'success', 
          text: '申请提交成功！我们会在24小时内审核并通知您。' 
        })
        // 延迟跳转到信使中心
        setTimeout(() => {
          router.push('/courier')
        }, 3000)
      } else {
        const error = await response.json()
        setMessage({ 
          type: 'error', 
          text: error.message || '提交失败，请重试' 
        })
      }
    } catch (error) {
      setMessage({ 
        type: 'error', 
        text: '网络错误，请稍后重试' 
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  const updateForm = (field: keyof ApplicationForm, value: any) => {
    setForm(prev => ({ ...prev, [field]: value }))
  }

  const timeSlotOptions = [
    '08:00-12:00 上午',
    '12:00-14:00 午休',
    '14:00-18:00 下午',
    '18:00-22:00 晚上'
  ]

  const toggleTimeSlot = (slot: string) => {
    const newSlots = form.timeSlots.includes(slot)
      ? form.timeSlots.filter(s => s !== slot)
      : [...form.timeSlots, slot]
    updateForm('timeSlots', newSlots)
  }

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-2xl mx-auto px-4 py-8">
        {/* Back Button */}
        <div className="mb-6">
          <Button asChild variant="outline" size="sm" className="border-amber-300 text-amber-700 hover:bg-amber-50">
            <Link href="/courier">
              <ArrowLeft className="mr-2 h-4 w-4" />
              返回信使中心
            </Link>
          </Button>
        </div>

        {/* Header */}
        <div className="text-center mb-8">
          <div className="w-16 h-16 bg-amber-200 rounded-full flex items-center justify-center mx-auto mb-4">
            <Send className="w-8 h-8 text-amber-700" />
          </div>
          <h1 className="font-serif text-3xl font-bold text-amber-900 mb-2">
            申请成为 OpenPenPal 信使
          </h1>
          <p className="text-amber-700 text-lg">
            成为信使，传递每一封有温度的信件
          </p>
          <div className="flex items-center justify-center gap-2 mt-4 text-sm text-amber-600">
            <Check className="w-4 h-4" />
            <span>完成首次任务后即可正式上线</span>
          </div>
        </div>

        {/* Message */}
        {message && (
          <Alert className={`mb-6 ${message.type === 'error' ? 'border-red-300 bg-red-50' : 'border-green-300 bg-green-50'}`}>
            {message.type === 'success' ? (
              <Check className="h-4 w-4 text-green-600" />
            ) : (
              <AlertCircle className="h-4 w-4 text-red-600" />
            )}
            <AlertDescription className={message.type === 'error' ? 'text-red-700' : 'text-green-700'}>
              {message.text}
            </AlertDescription>
          </Alert>
        )}

        <form onSubmit={handleSubmit}>
          <div className="space-y-6">
            {/* 基本信息 */}
            <Card className="border-amber-200 bg-white shadow-lg">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Users className="h-5 w-5" />
                  基本信息
                </CardTitle>
                <CardDescription>
                  请填写您的基本信息和联系方式
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">姓名 *</Label>
                  <Input
                    id="name"
                    value={form.name}
                    onChange={(e) => updateForm('name', e.target.value)}
                    placeholder="请输入您的真实姓名"
                    required
                  />
                  <p className="text-xs text-amber-600">平台内部使用，用于身份验证</p>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="contact">微信号 / 联系方式 *</Label>
                  <Input
                    id="contact"
                    value={form.contact}
                    onChange={(e) => updateForm('contact', e.target.value)}
                    placeholder="微信号或手机号"
                    required
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="school">学校名称 *</Label>
                  <Input
                    id="school"
                    value={form.school}
                    onChange={(e) => updateForm('school', e.target.value)}
                    placeholder="如：北京大学"
                    required
                  />
                </div>
              </CardContent>
            </Card>

            {/* 服务区域 */}
            <Card className="border-amber-200 bg-white shadow-lg">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <MapPin className="h-5 w-5" />
                  服务区域
                </CardTitle>
                <CardDescription>
                  设置您希望覆盖的投递区域
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="zone">投递片区编码 *</Label>
                  <Input
                    id="zone"
                    value={form.zone}
                    onChange={(e) => updateForm('zone', e.target.value)}
                    placeholder="如：PK5F（北大5楼）或 PK5F* （整个5楼）"
                    required
                  />
                  <p className="text-xs text-amber-600">
                    支持通配符 * 表示覆盖整个区域。建议新手信使先申请较小区域。
                  </p>
                </div>

                <div className="space-y-3">
                  <Label>是否有打印二维码贴纸的条件？ *</Label>
                  <RadioGroup
                    value={form.hasPrinter}
                    onValueChange={(value) => updateForm('hasPrinter', value)}
                  >
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="yes" id="printer-yes" />
                      <Label htmlFor="printer-yes">是，我可以打印贴纸</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="no" id="printer-no" />
                      <Label htmlFor="printer-no">否，需要平台提供</Label>
                    </div>
                  </RadioGroup>
                </div>
              </CardContent>
            </Card>

            {/* 投递能力 */}
            <Card className="border-amber-200 bg-white shadow-lg">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Clock className="h-5 w-5" />
                  投递能力评估
                </CardTitle>
                <CardDescription>
                  帮助我们合理分配任务
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="weeklyHours">每周可投递小时数</Label>
                    <Input
                      id="weeklyHours"
                      type="number"
                      min="1"
                      max="40"
                      value={form.weeklyHours}
                      onChange={(e) => updateForm('weeklyHours', parseInt(e.target.value))}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="maxDailyTasks">每日最大任务数</Label>
                    <Input
                      id="maxDailyTasks"
                      type="number"
                      min="1"
                      max="50"
                      value={form.maxDailyTasks}
                      onChange={(e) => updateForm('maxDailyTasks', parseInt(e.target.value))}
                    />
                  </div>
                </div>

                <div className="space-y-3">
                  <Label>主要交通方式 *</Label>
                  <RadioGroup
                    value={form.transportMethod}
                    onValueChange={(value) => updateForm('transportMethod', value)}
                  >
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="walk" id="transport-walk" />
                      <Label htmlFor="transport-walk">步行</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="bike" id="transport-bike" />
                      <Label htmlFor="transport-bike">自行车</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="ebike" id="transport-ebike" />
                      <Label htmlFor="transport-ebike">电动车</Label>
                    </div>
                  </RadioGroup>
                </div>

                <div className="space-y-3">
                  <Label>可投递时间段（可多选）</Label>
                  <div className="grid grid-cols-2 gap-2">
                    {timeSlotOptions.map((slot) => (
                      <div key={slot} className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          id={`slot-${slot}`}
                          checked={form.timeSlots.includes(slot)}
                          onChange={() => toggleTimeSlot(slot)}
                          className="rounded border-amber-300"
                        />
                        <Label htmlFor={`slot-${slot}`} className="text-sm">{slot}</Label>
                      </div>
                    ))}
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* 社区参与 */}
            <Card className="border-amber-200 bg-white shadow-lg">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Heart className="h-5 w-5" />
                  社区参与
                </CardTitle>
                <CardDescription>
                  让我们了解您的想法和意愿
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="selfIntro">为什么想成为信使？（选填）</Label>
                  <Textarea
                    id="selfIntro"
                    value={form.selfIntro}
                    onChange={(e) => updateForm('selfIntro', e.target.value)}
                    placeholder="分享您的想法和动机..."
                    rows={3}
                  />
                </div>

                <div className="space-y-3">
                  <Label>是否愿意帮助新信使？</Label>
                  <RadioGroup
                    value={form.canMentor}
                    onValueChange={(value) => updateForm('canMentor', value)}
                  >
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="yes" id="mentor-yes" />
                      <Label htmlFor="mentor-yes">愿意，我可以分享经验</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="maybe" id="mentor-maybe" />
                      <Label htmlFor="mentor-maybe">看情况，有时间的话</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="no" id="mentor-no" />
                      <Label htmlFor="mentor-no">暂时不考虑</Label>
                    </div>
                  </RadioGroup>
                </div>
              </CardContent>
            </Card>

            {/* 提交按钮 */}
            <div className="text-center space-y-4">
              <div className="p-4 bg-amber-100 rounded-lg border border-amber-200">
                <p className="text-amber-800 text-sm">
                  <Mail className="inline w-4 h-4 mr-1" />
                  提交申请后，我们将在24小时内审核并通过微信/短信通知您结果。
                  审核通过后，您将收到新手任务指引。
                </p>
              </div>
              
              <Button 
                type="submit" 
                size="lg"
                disabled={isSubmitting || !form.name || !form.contact || !form.school || !form.zone || !form.hasPrinter || !form.transportMethod}
                className="w-full bg-amber-600 hover:bg-amber-700 text-white font-serif text-lg px-8 py-6"
              >
                {isSubmitting ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                    提交中...
                  </>
                ) : (
                  <>
                    <Send className="mr-2 h-5 w-5" />
                    提交申请
                  </>
                )}
              </Button>
            </div>
          </div>
        </form>
      </div>
    </div>
  )
}