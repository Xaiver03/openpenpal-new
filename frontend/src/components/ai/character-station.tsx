'use client'

import React, { useState, useEffect } from 'react'
import { Bot, MessageCircle, Heart, User, Palette, Key, FileText, ThumbsUp, ThumbsDown, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
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
import { Slider } from '@/components/ui/slider'
import { aiService } from '@/lib/services/ai-service'
import { toast } from 'sonner'

interface Letter {
  id: string
  content: string
  senderName: string
  receivedDate: Date
}

interface CharacterStationProps {
  letters?: Letter[]
  onSelectLetter?: (letterId: string) => void
  className?: string
}

interface LetterSummary {
  mainTheme: string
  emotionalTone: string
  keyPoints: string[]
  senderIntent: string
  urgencyLevel: number // 1-5
}

interface ReplyAdvice {
  role: string
  perspectives: string[]
  suggestedTone: string
  keyPointsToAddress: string[]
  openingLines: string[]
  closingLines: string[]
}

const REPLY_ROLES = {
  'warm_friend': {
    label: '温暖的朋友',
    description: '以朋友的身份，给予温暖和支持',
    icon: Heart,
  },
  'wise_elder': {
    label: '智慧长者',
    description: '以长辈的智慧，给予指导和建议',
    icon: User,
  },
  'professional': {
    label: '专业人士',
    description: '以专业的角度，提供客观建议',
    icon: Bot,
  },
  'romantic': {
    label: '浪漫伴侣',
    description: '以恋人的角度，表达关心和爱意',
    icon: Heart,
  },
  'humorous': {
    label: '幽默达人',
    description: '用幽默化解尴尬，带来欢乐',
    icon: Sparkles,
  }
}

export function CharacterStation({
  letters = [],
  onSelectLetter,
  className = ''
}: CharacterStationProps) {
  const [selectedLetter, setSelectedLetter] = useState<Letter | null>(null)
  const [letterContent, setLetterContent] = useState('')
  const [summary, setSummary] = useState<LetterSummary | null>(null)
  const [replyRole, setReplyRole] = useState('')
  const [willingnessLevel, setWillingnessLevel] = useState(3) // 1-5 回信意愿
  const [advice, setAdvice] = useState<ReplyAdvice | null>(null)
  const [loading, setLoading] = useState(false)
  const [loadingSummary, setLoadingSummary] = useState(false)

  // 选择信件
  const handleSelectLetter = (letter: Letter) => {
    setSelectedLetter(letter)
    setLetterContent(letter.content)
    setSummary(null)
    setAdvice(null)
    if (onSelectLetter) {
      onSelectLetter(letter.id)
    }
    // 自动生成摘要
    generateSummary(letter.content)
  }

  // 生成信件摘要
  const generateSummary = async (content: string) => {
    setLoadingSummary(true)
    try {
      const prompt = `请分析这封信件并提供以下信息的JSON格式：
1. mainTheme: 信件的主要主题（一句话概括）
2. emotionalTone: 情感基调（积极/消极/中性/复杂）
3. keyPoints: 关键要点列表（3-5个要点）
4. senderIntent: 发信人的意图（询问/分享/请求/表达情感等）
5. urgencyLevel: 紧急程度（1-5，5最紧急）

信件内容：
${content}`

      // Use a simple mock for now since the AI service doesn't have this functionality
      const mockResponse = {
        content: JSON.stringify({
          mainTheme: '日常交流',
          emotionalTone: '积极',
          keyPoints: ['分享近况', '表达关心', '保持联系'],
          senderIntent: '维系关系',
          urgencyLevel: 3
        })
      }

      try {
        const summaryData = JSON.parse(mockResponse.content)
        setSummary(summaryData)
      } catch (e) {
        // 如果不是JSON格式，尝试解析文本
        setSummary({
          mainTheme: '日常交流',
          emotionalTone: '中性',
          keyPoints: ['分享近况', '表达关心', '保持联系'],
          senderIntent: '维系关系',
          urgencyLevel: 3
        })
      }
    } catch (error) {
      console.error('Failed to generate summary:', error)
      toast.error('生成信件摘要失败')
    } finally {
      setLoadingSummary(false)
    }
  }

  // 生成回信建议
  const generateAdvice = async () => {
    if (!replyRole || !selectedLetter) {
      toast.error('请选择回信角色')
      return
    }

    setLoading(true)
    try {
      const roleInfo = REPLY_ROLES[replyRole as keyof typeof REPLY_ROLES]
      const prompt = `基于以下信息生成回信建议：

信件摘要：
- 主题：${summary?.mainTheme}
- 情感基调：${summary?.emotionalTone}
- 关键要点：${summary?.keyPoints.join('、')}
- 发信人意图：${summary?.senderIntent}

回信角色：${roleInfo.label} - ${roleInfo.description}
回信意愿度：${willingnessLevel}/5

请提供JSON格式的回信建议：
1. role: 角色名称
2. perspectives: 3-4个回信角度建议
3. suggestedTone: 建议的语气基调
4. keyPointsToAddress: 需要回应的要点
5. openingLines: 2-3个开头建议
6. closingLines: 2-3个结尾建议

注意：回信意愿度低时，建议应该更加委婉和简短。`

      // Use a simple mock for now since the AI service doesn't have this functionality
      const mockResponse = {
        content: JSON.stringify({
          willingness: 4,
          tone: '友好',
          suggestions: ['表达理解', '分享经历', '提供帮助'],
          responseLength: 'medium',
          openingLines: ['很高兴收到你的来信...', '感谢你和我分享...'],
          closingLines: ['期待你的回复', '愿一切都好']
        })
      }

      try {
        const adviceData = JSON.parse(mockResponse.content)
        setAdvice({
          ...adviceData,
          role: roleInfo.label
        })
        toast.success('回信建议已生成')
      } catch (e) {
        // 提供默认建议
        setAdvice({
          role: roleInfo.label,
          perspectives: [
            '表达收到信件的喜悦',
            '回应对方关心的问题',
            '分享自己的近况',
            '表达对未来的期待'
          ],
          suggestedTone: '温暖友好',
          keyPointsToAddress: summary?.keyPoints || [],
          openingLines: [
            `亲爱的${selectedLetter.senderName}，收到你的来信真的很开心！`,
            `${selectedLetter.senderName}，好久不见，你的信带来了温暖。`,
            `看到你的信，仿佛听到了你的声音。`
          ],
          closingLines: [
            '期待你的回信，保重身体！',
            '愿一切安好，常联系。',
            '祝好，期待下次见面！'
          ]
        })
      }
    } catch (error) {
      console.error('Failed to generate advice:', error)
      toast.error('生成回信建议失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* 信件选择区 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            选择要回复的信件
          </CardTitle>
          <CardDescription>
            角色驿站帮您总结信件内容，并根据您的回信角色提供建议
          </CardDescription>
        </CardHeader>
        <CardContent>
          {letters.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <MessageCircle className="h-8 w-8 mx-auto mb-3" />
              <p>暂无收到的信件</p>
            </div>
          ) : (
            <div className="space-y-2">
              {letters.map((letter) => (
                <Card
                  key={letter.id}
                  className={`cursor-pointer transition-all ${
                    selectedLetter?.id === letter.id
                      ? 'border-blue-500 shadow-md'
                      : 'hover:shadow-sm'
                  }`}
                  onClick={() => handleSelectLetter(letter)}
                >
                  <CardContent className="p-4">
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium">{letter.senderName}</span>
                      <span className="text-xs text-muted-foreground">
                        {letter.receivedDate.toLocaleDateString()}
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {letter.content}
                    </p>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
          
          {/* 手动输入信件内容 */}
          <div className="mt-4 space-y-2">
            <Label>或直接输入信件内容</Label>
            <Textarea
              placeholder="粘贴或输入收到的信件内容..."
              value={letterContent}
              onChange={(e) => {
                setLetterContent(e.target.value)
                setSelectedLetter(null)
                setSummary(null)
                setAdvice(null)
              }}
              rows={4}
            />
            {letterContent && !selectedLetter && (
              <Button 
                onClick={() => generateSummary(letterContent)}
                disabled={loadingSummary}
                size="sm"
              >
                生成摘要
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* 信件摘要 */}
      {(selectedLetter || letterContent) && summary && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              信件摘要
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <span className="font-medium">主题：</span>
              <span className="ml-2">{summary.mainTheme}</span>
            </div>
            <div className="flex items-center gap-4">
              <div>
                <span className="font-medium">情感基调：</span>
                <Badge variant="outline" className="ml-2">
                  {summary.emotionalTone}
                </Badge>
              </div>
              <div>
                <span className="font-medium">紧急程度：</span>
                <span className="ml-2">
                  {'⭐'.repeat(summary.urgencyLevel)}
                </span>
              </div>
            </div>
            <div>
              <span className="font-medium">发信人意图：</span>
              <span className="ml-2">{summary.senderIntent}</span>
            </div>
            <div>
              <span className="font-medium block mb-2">关键要点：</span>
              <div className="space-y-1">
                {summary.keyPoints.map((point, index) => (
                  <div key={index} className="flex items-start gap-2">
                    <span className="text-blue-500">•</span>
                    <span className="text-sm">{point}</span>
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 回信设置 */}
      {summary && (
        <Card>
          <CardHeader>
            <CardTitle>设置回信角色和意愿</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* 回信意愿度 */}
            <div className="space-y-2">
              <Label>回信意愿度</Label>
              <div className="flex items-center gap-4">
                <ThumbsDown className="h-4 w-4 text-muted-foreground" />
                <Slider
                  value={[willingnessLevel]}
                  onValueChange={(value) => setWillingnessLevel(value[0])}
                  min={1}
                  max={5}
                  step={1}
                  className="flex-1"
                />
                <ThumbsUp className="h-4 w-4 text-muted-foreground" />
                <span className="w-12 text-center font-medium">
                  {willingnessLevel}/5
                </span>
              </div>
              <p className="text-xs text-muted-foreground">
                {willingnessLevel <= 2 && '不太想回复，建议简短礼貌'}
                {willingnessLevel === 3 && '一般意愿，正常回复即可'}
                {willingnessLevel >= 4 && '很想回复，可以详细热情'}
              </p>
            </div>

            {/* 角色选择 */}
            <div className="space-y-2">
              <Label>选择回信角色</Label>
              <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
                {Object.entries(REPLY_ROLES).map(([key, role]) => (
                  <Card
                    key={key}
                    className={`cursor-pointer transition-all ${
                      replyRole === key
                        ? 'border-blue-500 shadow-md'
                        : 'hover:shadow-sm'
                    }`}
                    onClick={() => setReplyRole(key)}
                  >
                    <CardContent className="p-3">
                      <div className="flex items-center gap-2 mb-1">
                        <role.icon className="h-4 w-4" />
                        <span className="font-medium text-sm">{role.label}</span>
                      </div>
                      <p className="text-xs text-muted-foreground">
                        {role.description}
                      </p>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </div>

            <Button
              onClick={generateAdvice}
              disabled={loading || !replyRole}
              className="w-full"
            >
              {loading ? (
                <>生成建议中...</>
              ) : (
                <>
                  <Sparkles className="h-4 w-4 mr-2" />
                  生成回信建议
                </>
              )}
            </Button>
          </CardContent>
        </Card>
      )}

      {/* 回信建议 */}
      {advice && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <MessageCircle className="h-5 w-5" />
              {advice.role}的回信建议
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="perspectives" className="w-full">
              <TabsList className="grid w-full grid-cols-4">
                <TabsTrigger value="perspectives">回信角度</TabsTrigger>
                <TabsTrigger value="tone">语气建议</TabsTrigger>
                <TabsTrigger value="opening">开头建议</TabsTrigger>
                <TabsTrigger value="closing">结尾建议</TabsTrigger>
              </TabsList>

              <TabsContent value="perspectives" className="space-y-2">
                {advice.perspectives.map((perspective, index) => (
                  <div key={index} className="bg-blue-50 p-3 rounded-lg">
                    <p className="text-sm">{perspective}</p>
                  </div>
                ))}
              </TabsContent>

              <TabsContent value="tone" className="space-y-2">
                <div className="bg-purple-50 p-4 rounded-lg">
                  <p className="font-medium mb-2">建议语气：{advice.suggestedTone}</p>
                  <p className="text-sm text-muted-foreground">
                    根据您的回信意愿度和选择的角色，建议采用这种语气来回信
                  </p>
                </div>
                <div>
                  <p className="font-medium mb-2">需要回应的要点：</p>
                  <div className="space-y-1">
                    {advice.keyPointsToAddress.map((point, index) => (
                      <div key={index} className="flex items-start gap-2">
                        <Key className="h-3 w-3 text-blue-500 mt-0.5" />
                        <span className="text-sm">{point}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="opening" className="space-y-2">
                {advice.openingLines.map((line, index) => (
                  <div key={index} className="bg-green-50 p-3 rounded-lg">
                    <p className="text-sm italic">"{line}"</p>
                  </div>
                ))}
              </TabsContent>

              <TabsContent value="closing" className="space-y-2">
                {advice.closingLines.map((line, index) => (
                  <div key={index} className="bg-orange-50 p-3 rounded-lg">
                    <p className="text-sm italic">"{line}"</p>
                  </div>
                ))}
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      )}

      {/* 使用提示 */}
      <Alert>
        <Sparkles className="h-4 w-4" />
        <AlertDescription>
          <strong>角色驿站：</strong>帮您快速理解信件内容，根据您的回信意愿和选择的角色，提供个性化的回信建议。让每一封回信都恰到好处。
        </AlertDescription>
      </Alert>
    </div>
  )
}