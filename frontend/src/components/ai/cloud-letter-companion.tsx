'use client'

import React, { useState, useEffect } from 'react'
import { Send, Bot, Heart, Calendar, Clock, Sparkles, MessageCircle, Plus, Users, UserPlus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { aiService, type AIPersona } from '@/lib/services/ai-service'
import { toast } from 'sonner'

interface CloudLetterCompanionProps {
  selectedPersonaId: string
  className?: string
  mode?: 'preset' | 'custom' // 新增模式选择
}

interface LetterExchange {
  id: string
  from: 'user' | 'ai'
  content: string
  timestamp: Date
  persona?: AIPersona
}

// 人设图标映射
const personaIcons: Record<string, string> = {
  poet: '🎭',
  philosopher: '🤔', 
  artist: '🎨',
  scientist: '🔬',
  traveler: '✈️',
  historian: '📚',
  mentor: '💝',
  friend: '👫',
}

// 自定义现实角色接口（从UnreachableCompanion复用）
interface CustomPersona {
  id: string
  name: string
  relationship: string 
  lastContactDate?: string
  memories: string[]
  personality: string
  writingStyle: string
  createdAt: Date
}

export function CloudLetterCompanion({
  selectedPersonaId,
  className = '',
  mode = 'preset',
}: CloudLetterCompanionProps) {
  const [selectedPersona, setSelectedPersona] = useState<AIPersona | null>(null)
  const [customPersonas, setCustomPersonas] = useState<CustomPersona[]>([]) // 自定义角色列表
  const [selectedCustomPersona, setSelectedCustomPersona] = useState<CustomPersona | null>(null)
  const [letterContent, setLetterContent] = useState('')
  const [letters, setLetters] = useState<LetterExchange[]>([])
  const [loading, setLoading] = useState(false)
  const [fetchingPersona, setFetchingPersona] = useState(true)

  useEffect(() => {
    if (mode === 'preset') {
      fetchPersona()
    } else {
      loadCustomPersonas()
    }
  }, [selectedPersonaId])

  // 加载自定义角色（复用UnreachableCompanion的逻辑）
  const loadCustomPersonas = () => {
    const saved = localStorage.getItem('unreachable_personas')
    if (saved) {
      const parsed = JSON.parse(saved)
      const personas = parsed.map((p: any) => ({
        ...p,
        createdAt: new Date(p.created_at)
      }))
      setCustomPersonas(personas)
      
      // 如果有选中的ID，设置选中的自定义角色
      if (selectedPersonaId) {
        const selected = personas.find((p: CustomPersona) => p.id === selectedPersonaId)
        setSelectedCustomPersona(selected || null)
      }
    }
    setFetchingPersona(false)
  }

  const fetchPersona = async () => {
    if (!selectedPersonaId) return
    
    setFetchingPersona(true)
    try {
      const result = await aiService.getAIPersonas()
      const persona = result.personas?.find((p: AIPersona) => p.id === selectedPersonaId)
      setSelectedPersona(persona || null)
    } catch (error) {
      console.error('Failed to fetch persona:', error)
    } finally {
      setFetchingPersona(false)
    }
  }

  const sendLetter = async () => {
    // 验证输入
    if (!letterContent.trim()) {
      toast.error('请输入信件内容')
      return
    }
    if (mode === 'preset' && !selectedPersona) {
      toast.error('请选择AI笔友')
      return
    }
    if (mode === 'custom' && !selectedCustomPersona) {
      toast.error('请选择要写信的人')
      return
    }

    // Add user letter to the list
    const userLetter: LetterExchange = {
      id: `user_${Date.now()}`,
      from: 'user',
      content: letterContent,
      timestamp: new Date(),
    }

    setLetters(prev => [...prev, userLetter])
    setLoading(true)
    const currentContent = letterContent
    setLetterContent('')

    try {
      if (mode === 'preset' && selectedPersona) {
        // 预设AI笔友模式
        const response = await aiService.scheduleDelayedReply({
          letterId: `user_letter_${Date.now()}`,
          persona: selectedPersona.id as any,
          delay_hours: 24 // 24小时延迟，符合PRD要求
        })
        
        // 显示调度成功消息
        toast.success(`AI笔友收到了你的信件！预计在${response.delay_hours}小时后回信`, {
          duration: 5000
        })
        
        // 添加一个系统提示消息
        const systemMessage: LetterExchange = {
          id: `system_${Date.now()}`,
          from: 'ai',
          content: `📬 你的信件已送达！${selectedPersona.name}会在24小时内给你回信，请耐心等待...\n\n（这就是手写信的魅力所在 - 等待与惊喜 ✨）`,
          timestamp: new Date(),
          persona: selectedPersona,
        }
        
        setLetters(prev => [...prev, systemMessage])
      } else if (mode === 'custom' && selectedCustomPersona) {
        // 自定义现实角色模式 - 需要信使审核（模拟API调用）
        const simulatedResponse = {
          id: `custom_reply_${Date.now()}`,
          status: 'pending_review',
          message: '信件已提交审核'
        }
        
        toast.success(`你的信件已送达！${selectedCustomPersona.name}会在审核后给你回信`, {
          duration: 5000
        })
        
        // 添加系统提示消息
        const systemMessage: LetterExchange = {
          id: `system_${Date.now()}`,
          from: 'ai',
          content: `📮 你的信件已送达${selectedCustomPersona.name}！\n\n由于这是给特殊的人写信，我们的信使会帮助润色回信内容，确保每一个字都充满温度...\n\n请耐心等待，好的回信值得等待 💝`,
          timestamp: new Date(),
        }
        
        setLetters(prev => [...prev, systemMessage])
      }
    } catch (error) {
      console.error('Failed to get AI reply:', error)
      toast.error('信件发送失败，请稍后再试')
    } finally {
      setLoading(false)
    }
  }

  // Mock response generator (would be replaced with actual AI service call)
  const generateMockResponse = (persona: AIPersona, userContent: string): string => {
    const responses: Record<string, string[]> = {
      poet: [
        "亲爱的朋友，\n\n你的文字如春风拂过心田，让我想起了那句诗：'山重水复疑无路，柳暗花明又一村。'人生的路虽然曲折，但总有美好在前方等待。\n\n愿你如诗一般，在平凡中发现不平凡的美。\n\n你的诗意伙伴 🎭",
        "看到你的信，我的心中涌起千言万语，如潮水般汹涌。你提到的感受让我想起了一首诗：'人生若只如初见，何事秋风悲画扇。'\n\n生活中的每一个瞬间都值得被记录，被珍藏。让我们一起用文字编织美好的回忆吧。\n\n以诗相伴 🌸"
      ],
      friend: [
        "嗨！收到你的信真开心～\n\n看到你分享的这些，我觉得我们真的很有共同话题呢！你知道吗，我昨天也遇到了类似的事情，当时我的感受和你描述的几乎一模一样。\n\n有时候生活就是这样，会给我们一些意想不到的小惊喜。希望我们可以一直这样分享彼此的生活点滴！\n\n你的朋友 👫",
        "看到你的信我笑了，你的表达方式总是那么有趣！\n\n你提到的那件事让我想起了我们之前聊过的话题。我觉得你真的很棒，总是能在平凡的事情中找到乐趣。这样的心态真的很难得！\n\n下次记得也要告诉我更多有趣的事情哦～\n\n温暖的陪伴 💕"
      ],
      philosopher: [
        "我的朋友，\n\n读你的信如同品一杯清茶，需要慢慢体味其中的深意。你提到的问题让我想起苏格拉底曾说：'认识你自己。'\n\n人生的意义往往不在于我们寻找什么答案，而在于我们提出什么样的问题。你的思考已经让你走在了正确的道路上。\n\n让我们继续探索这个充满奥秘的世界吧。\n\n你的思辨伙伴 🤔",
        "亲爱的求知者，\n\n你的话语中蕴含着深刻的思考，这让我想起了老子的话：'知者不言，言者不知。'有时候，最深刻的智慧往往隐藏在最简单的话语中。\n\n生活的智慧不在书本里，而在于我们如何理解和体验这个世界。你的感悟已经证明了这一点。\n\n愿智慧伴你前行 📚"
      ]
    }

    const personaResponses = responses[persona.id] || responses.friend
    return personaResponses[Math.floor(Math.random() * personaResponses.length)]
  }

  if (fetchingPersona) {
    return (
      <div className={`space-y-4 ${className}`}>
        <Skeleton className="h-20 w-full" />
        <Skeleton className="h-40 w-full" />
      </div>
    )
  }

  if (mode === 'preset' && !selectedPersona) {
    return (
      <Card className={className}>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Bot className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground text-center">
            请先选择一个AI笔友人设开始对话
          </p>
        </CardContent>
      </Card>
    )
  }

  if (mode === 'custom' && !selectedCustomPersona) {
    return (
      <Card className={className}>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Users className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground text-center">
            请先创建或选择一个现实角色开始对话
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Header - 支持两种模式 */}
      {mode === 'preset' && selectedPersona && (
        <Card className="bg-gradient-to-r from-blue-50 to-indigo-50 border-blue-200">
          <CardHeader>
            <div className="flex items-center gap-3">
              <Avatar className="h-12 w-12">
                <AvatarFallback className="text-lg bg-white">
                  {personaIcons[selectedPersona.id] || '🤖'}
                </AvatarFallback>
              </Avatar>
              <div className="flex-1">
                <CardTitle className="flex items-center gap-2">
                  {selectedPersona.name}
                  <Badge variant="secondary">AI笔友</Badge>
                </CardTitle>
                <CardDescription>
                  {selectedPersona.description}
                </CardDescription>
              </div>
              <Sparkles className="h-5 w-5 text-blue-600" />
            </div>
          </CardHeader>
        </Card>
      )}

      {mode === 'custom' && selectedCustomPersona && (
        <Card className="bg-gradient-to-r from-rose-50 to-pink-50 border-rose-200">
          <CardHeader>
            <div className="flex items-center gap-3">
              <Avatar className="h-12 w-12">
                <AvatarFallback className="text-lg bg-white">
                  ❤️
                </AvatarFallback>
              </Avatar>
              <div className="flex-1">
                <CardTitle className="flex items-center gap-2">
                  {selectedCustomPersona.name}
                  <Badge variant="outline" className="border-rose-300 text-rose-700">
                    {selectedCustomPersona.relationship}
                  </Badge>
                </CardTitle>
                <CardDescription>
                  {selectedCustomPersona.personality}
                  {selectedCustomPersona.lastContactDate && (
                    <div className="flex items-center gap-1 mt-1 text-xs">
                      <Calendar className="h-3 w-3" />
                      最后联系: {selectedCustomPersona.lastContactDate}
                    </div>
                  )}
                </CardDescription>
              </div>
              <Heart className="h-5 w-5 text-rose-600" />
            </div>
          </CardHeader>
        </Card>
      )}

      {/* Letter Exchange History */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MessageCircle className="h-5 w-5" />
            书信往来
          </CardTitle>
          <CardDescription>
            {mode === 'preset' && selectedPersona && `与 ${selectedPersona.name} 的长期笔友对话`}
            {mode === 'custom' && selectedCustomPersona && `与 ${selectedCustomPersona.name} 的特殊对话`}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {letters.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Heart className="h-8 w-8 mx-auto mb-3 text-pink-400" />
              <p>还没有开始对话</p>
              <p className="text-sm">写下你的第一封信，开始这段特别的笔友关系吧！</p>
            </div>
          ) : (
            <div className="space-y-4 max-h-96 overflow-y-auto">
              {letters.map((letter) => (
                <div
                  key={letter.id}
                  className={`flex ${
                    letter.from === 'user' ? 'justify-end' : 'justify-start'
                  }`}
                >
                  <div
                    className={`max-w-[80%] rounded-lg p-4 ${
                      letter.from === 'user'
                        ? 'bg-blue-500 text-white'
                        : 'bg-gray-100 text-gray-900'
                    }`}
                  >
                    <div className="flex items-center gap-2 mb-2">
                      {letter.from === 'ai' && (
                        <Avatar className="h-6 w-6">
                          <AvatarFallback className="text-xs">
                            {selectedPersona ? personaIcons[selectedPersona.id] || '🤖' : '🤖'}
                          </AvatarFallback>
                        </Avatar>
                      )}
                      <span className="text-xs opacity-75">
                        {letter.from === 'user' ? '你' : selectedPersona?.name || '未知'}
                      </span>
                      <span className="text-xs opacity-60">
                        {letter.timestamp.toLocaleTimeString()}
                      </span>
                    </div>
                    <div className="whitespace-pre-wrap text-sm leading-relaxed">
                      {letter.content}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Write New Letter */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Plus className="h-5 w-5" />
            写信给 {selectedPersona?.name || '未选择'}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <Textarea
            placeholder={`写下你想对${selectedPersona?.name || '对方'}说的话...`}
            value={letterContent}
            onChange={(e) => setLetterContent(e.target.value)}
            rows={6}
            className="resize-none"
          />
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <Clock className="h-3 w-3" />
              <span>AI笔友会在1-3天内回信</span>
            </div>
            <Button
              onClick={sendLetter}
              disabled={loading || !letterContent.trim()}
              className="gap-2"
            >
              {loading ? (
                <>
                  <Bot className="h-4 w-4 animate-spin" />
                  等待回信...
                </>
              ) : (
                <>
                  <Send className="h-4 w-4" />
                  发送信件
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Usage Tips */}
      <Alert>
        <Sparkles className="h-4 w-4" />
        <AlertDescription>
          <strong>提示：</strong> {selectedPersona?.name || '对方'} 会记住你们的对话历史，随着交流的深入，回信会越来越个性化和贴心。
        </AlertDescription>
      </Alert>
    </div>
  )
}