'use client'

import React, { useState } from 'react'
import { Bot, Lightbulb, Heart, Clock, User, MessageCircle, Palette, Key } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { aiService, type AIReplyAdviceRequest, type AIReplyAdvice } from '@/lib/services/ai-service'
import { toast } from 'sonner'

interface AIReplyAdviceProps {
  letterId: string
  letterContent?: string
  onUseAdvice?: (advice: AIReplyAdvice) => void
  className?: string
}

const PERSONA_TYPES = {
  'family': {
    label: '亲人',
    description: '家庭成员的关爱视角',
    icon: Heart,
    examples: ['父母', '兄弟姐妹', '祖父母', '其他家人']
  },
  'friend': {
    label: '朋友',
    description: '朋友的支持与陪伴视角',
    icon: User,
    examples: ['同学', '闺蜜', '兄弟', '同事朋友']
  },
  'romantic': {
    label: 'TA',
    description: '恋人或心仪对象的视角',
    icon: MessageCircle,
    examples: ['恋人', '喜欢的人', '伴侣', '心仪对象']
  },
  'custom': {
    label: '自定义人设',
    description: '创建一个独特的回信角色',
    icon: Bot,
    examples: ['人生导师', '智慧长者', '温暖朋友', '理解者']
  }
}

export function AIReplyAdvice({
  letterId,
  letterContent,
  onUseAdvice,
  className = '',
}: AIReplyAdviceProps) {
  const [personaType, setPersonaType] = useState<string>('')
  const [personaName, setPersonaName] = useState<string>('')
  const [personaDesc, setPersonaDesc] = useState<string>('')
  const [relationship, setRelationship] = useState<string>('')
  const [deliveryDays, setDeliveryDays] = useState<number>(0)
  const [advice, setAdvice] = useState<AIReplyAdvice | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const generateAdvice = async () => {
    if (!letterId) {
      toast.error('请先保存信件草稿')
      return
    }

    if (!personaType || !personaName) {
      toast.error('请选择人设类型并填写姓名')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const request: AIReplyAdviceRequest = {
        letterId: letterId,
        persona_type: personaType as any,
        persona_name: personaName,
        persona_desc: personaDesc || undefined,
        relationship: relationship || undefined,
        delivery_days: deliveryDays
      }

      const result = await aiService.generateReplyAdvice(request)
      setAdvice(result)
      toast.success('回信角度建议生成成功！')
    } catch (error: any) {
      console.error('Failed to generate reply advice:', error)
      setError(error.message || '生成回信建议失败')
      toast.error('生成回信建议失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  const parseAdviceList = (adviceString: string): string[] => {
    if (!adviceString) return []
    try {
      // 尝试解析JSON数组
      if (adviceString.startsWith('[')) {
        return JSON.parse(adviceString)
      }
      // 如果不是JSON，按分隔符拆分
      return adviceString.split(/[，,、\n]/).filter(item => item.trim())
    } catch {
      return [adviceString]
    }
  }

  const handleUseAdvice = () => {
    if (advice && onUseAdvice) {
      onUseAdvice(advice)
      toast.success('已采用此回信建议')
    }
  }

  const selectedPersonaInfo = personaType ? PERSONA_TYPES[personaType as keyof typeof PERSONA_TYPES] : null

  return (
    <Card className={`w-full ${className}`}>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Lightbulb className="h-5 w-5 text-blue-600" />
          <CardTitle>AI回信角度建议</CardTitle>
        </div>
        <CardDescription>
          让AI以特定身份为你提供回信的角度建议，而不是直接生成回信内容
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {!advice ? (
          <>
            {/* 人设类型选择 */}
            <div className="space-y-3">
              <Label htmlFor="persona-type">选择人设类型</Label>
              <Select value={personaType} onValueChange={setPersonaType}>
                <SelectTrigger>
                  <SelectValue placeholder="请选择一个有意义的身份..." />
                </SelectTrigger>
                <SelectContent>
                  {Object.entries(PERSONA_TYPES).map(([key, info]) => (
                    <SelectItem key={key} value={key}>
                      <div className="flex items-center gap-2">
                        <info.icon className="h-4 w-4" />
                        <div>
                          <div className="font-medium">{info.label}</div>
                          <div className="text-xs text-muted-foreground">{info.description}</div>
                        </div>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* 显示选中的人设信息 */}
            {selectedPersonaInfo && (
              <div className="bg-blue-50 p-4 rounded-lg border border-blue-200">
                <div className="flex items-center gap-2 mb-2">
                  <selectedPersonaInfo.icon className="h-4 w-4 text-blue-600" />
                  <span className="font-medium text-blue-900">{selectedPersonaInfo.label}</span>
                </div>
                <p className="text-sm text-blue-700 mb-3">{selectedPersonaInfo.description}</p>
                <div className="flex flex-wrap gap-1">
                  {selectedPersonaInfo.examples.map((example, index) => (
                    <Badge
                      key={index}
                      variant="secondary"
                      className="cursor-pointer bg-blue-100 text-blue-800 hover:bg-blue-200"
                      onClick={() => setPersonaName(example)}
                    >
                      {example}
                    </Badge>
                  ))}
                </div>
              </div>
            )}

            {/* 人设详细信息 */}
            {personaType && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="persona-name">人设姓名/称呼 *</Label>
                  <Input
                    id="persona-name"
                    value={personaName}
                    onChange={(e) => setPersonaName(e.target.value)}
                    placeholder="例如：奶奶、小明、李雷..."
                  />
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="relationship">关系描述</Label>
                  <Input
                    id="relationship"
                    value={relationship}
                    onChange={(e) => setRelationship(e.target.value)}
                    placeholder={`例如：${selectedPersonaInfo?.examples[0]}...`}
                  />
                </div>
              </div>
            )}

            {/* 人设描述 */}
            {personaType && (
              <div className="space-y-2">
                <Label htmlFor="persona-desc">人设详细描述 (可选)</Label>
                <Textarea
                  id="persona-desc"
                  value={personaDesc}
                  onChange={(e) => setPersonaDesc(e.target.value)}
                  placeholder="描述这个人的性格特点、生活经历、说话风格等..."
                  rows={3}
                />
              </div>
            )}

            {/* 延迟投递 */}
            {personaType && (
              <div className="space-y-2">
                <Label htmlFor="delivery-days">延迟投递天数</Label>
                <Select value={deliveryDays.toString()} onValueChange={(v) => setDeliveryDays(parseInt(v))}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="0">立即获取</SelectItem>
                    <SelectItem value="1">1天后</SelectItem>
                    <SelectItem value="2">2天后</SelectItem>
                    <SelectItem value="3">3天后</SelectItem>
                    <SelectItem value="7">1周后</SelectItem>
                  </SelectContent>
                </Select>
                {deliveryDays > 0 && (
                  <p className="text-xs text-muted-foreground flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    建议将在 {deliveryDays} 天后送达，营造真实的通信体验
                  </p>
                )}
              </div>
            )}

            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <Button 
              onClick={generateAdvice} 
              disabled={loading || !personaType || !personaName}
              className="w-full"
            >
              {loading ? (
                <>
                  <Bot className="h-4 w-4 mr-2 animate-spin" />
                  AI正在思考中...
                </>
              ) : (
                <>
                  <Lightbulb className="h-4 w-4 mr-2" />
                  获取回信角度建议
                </>
              )}
            </Button>
          </>
        ) : (
          /* 显示生成的建议 */
          <div className="space-y-6">
            {/* 人设信息 */}
            <div className="bg-gradient-to-r from-blue-50 to-purple-50 p-4 rounded-lg border">
              <div className="flex items-center gap-2 mb-2">
                <Heart className="h-4 w-4 text-purple-600" />
                <span className="font-semibold text-purple-900">{advice.persona_name}</span>
                <Badge variant="outline">{PERSONA_TYPES[advice.persona_type as keyof typeof PERSONA_TYPES]?.label}</Badge>
              </div>
              {advice.persona_desc && (
                <p className="text-sm text-purple-700">{advice.persona_desc}</p>
              )}
            </div>

            {/* 回信建议内容 */}
            <Tabs defaultValue="perspectives" className="w-full">
              <TabsList className="grid w-full grid-cols-4">
                <TabsTrigger value="perspectives">角度观点</TabsTrigger>
                <TabsTrigger value="tone">情感基调</TabsTrigger>
                <TabsTrigger value="topics">话题方向</TabsTrigger>
                <TabsTrigger value="style">写作要点</TabsTrigger>
              </TabsList>

              <TabsContent value="perspectives" className="space-y-3">
                <div className="flex items-center gap-2 mb-3">
                  <MessageCircle className="h-4 w-4 text-blue-600" />
                  <span className="font-medium">回信角度建议</span>
                </div>
                <div className="space-y-2">
                  {advice.perspectives.map((perspective, index) => (
                    <div key={index} className="bg-blue-50 p-3 rounded-lg border-l-4 border-blue-400">
                      <p className="text-blue-900">{perspective}</p>
                    </div>
                  ))}
                </div>
              </TabsContent>

              <TabsContent value="tone" className="space-y-3">
                <div className="flex items-center gap-2 mb-3">
                  <Heart className="h-4 w-4 text-pink-600" />
                  <span className="font-medium">情感基调</span>
                </div>
                <div className="bg-pink-50 p-4 rounded-lg border border-pink-200">
                  <p className="text-pink-900 font-medium">{advice.emotional_tone}</p>
                </div>
              </TabsContent>

              <TabsContent value="topics" className="space-y-3">
                <div className="flex items-center gap-2 mb-3">
                  <Lightbulb className="h-4 w-4 text-yellow-600" />
                  <span className="font-medium">建议话题</span>
                </div>
                <div className="bg-yellow-50 p-4 rounded-lg border border-yellow-200">
                  <p className="text-yellow-900">{advice.suggested_topics}</p>
                </div>
              </TabsContent>

              <TabsContent value="style" className="space-y-3">
                <div className="flex items-center gap-2 mb-3">
                  <Palette className="h-4 w-4 text-green-600" />
                  <span className="font-medium">写作风格</span>
                </div>
                <div className="bg-green-50 p-4 rounded-lg border border-green-200 space-y-3">
                  <div>
                    <span className="font-medium text-green-900">写作风格：</span>
                    <span className="text-green-800 ml-2">{advice.writing_style}</span>
                  </div>
                  <div>
                    <span className="font-medium text-green-900">关键要点：</span>
                    <p className="text-green-800 mt-1">{advice.key_points}</p>
                  </div>
                </div>
              </TabsContent>
            </Tabs>

            {/* 操作按钮 */}
            <div className="flex flex-col sm:flex-row gap-3">
              <Button onClick={handleUseAdvice} className="flex-1">
                <Key className="h-4 w-4 mr-2" />
                使用此建议开始写信
              </Button>
              <Button 
                variant="outline" 
                onClick={() => setAdvice(null)}
                className="flex-1"
              >
                重新选择人设
              </Button>
            </div>

            {/* 延迟投递信息 */}
            {advice.delivery_delay > 0 && advice.scheduled_for && (
              <Alert>
                <Clock className="h-4 w-4" />
                <AlertDescription>
                  此建议将在 {new Date(advice.scheduled_for).toLocaleString()} 送达，
                  为你营造真实的书信往来体验。
                </AlertDescription>
              </Alert>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}