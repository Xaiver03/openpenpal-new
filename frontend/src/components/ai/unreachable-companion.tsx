'use client'

import React, { useState, useEffect } from 'react'
import { Send, Heart, Plus, UserPlus, Users, Edit, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { aiService } from '@/lib/services/ai-service'
import { toast } from 'sonner'

interface UnreachablePersona {
  id: string
  name: string
  relationship: string // "祖母", "失联的朋友", etc.
  lastContactDate?: string
  memories: string[] // 共同回忆
  personality: string // 性格特征
  writingStyle: string // 写作风格
  createdAt: Date
}

interface LetterExchange {
  id: string
  from: 'user' | 'persona'
  content: string
  timestamp: Date
  personaId: string
}

export function UnreachableCompanion({ className = '' }: { className?: string }) {
  const [personas, setPersonas] = useState<UnreachablePersona[]>([])
  const [selectedPersona, setSelectedPersona] = useState<UnreachablePersona | null>(null)
  const [letterContent, setLetterContent] = useState('')
  const [letters, setLetters] = useState<LetterExchange[]>([])
  const [loading, setLoading] = useState(false)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  
  // 创建新人物的表单状态
  const [newPersona, setNewPersona] = useState({
    name: '',
    relationship: '',
    lastContactDate: '',
    memories: '',
    personality: '',
    writingStyle: '',
  })

  useEffect(() => {
    loadPersonas()
  }, [])

  const loadPersonas = () => {
    // 从localStorage加载已创建的虚拟人物
    const saved = localStorage.getItem('unreachable_personas')
    if (saved) {
      const parsed = JSON.parse(saved)
      setPersonas(parsed.map((p: any) => ({
        ...p,
        createdAt: new Date(p.created_at)
      })))
    }
  }

  const savePersonas = (updatedPersonas: UnreachablePersona[]) => {
    localStorage.setItem('unreachable_personas', JSON.stringify(updatedPersonas))
    setPersonas(updatedPersonas)
  }

  const createPersona = () => {
    if (!newPersona.name || !newPersona.relationship) {
      toast.error('请填写姓名和关系')
      return
    }

    const persona: UnreachablePersona = {
      id: `persona_${Date.now()}`,
      name: newPersona.name,
      relationship: newPersona.relationship,
      lastContactDate: newPersona.lastContactDate,
      memories: newPersona.memories.split('\n').filter(m => m.trim()),
      personality: newPersona.personality,
      writingStyle: newPersona.writingStyle,
      createdAt: new Date(),
    }

    const updated = [...personas, persona]
    savePersonas(updated)
    setNewPersona({
      name: '',
      relationship: '',
      lastContactDate: '',
      memories: '',
      personality: '',
      writingStyle: '',
    })
    setShowCreateDialog(false)
    toast.success(`已创建虚拟笔友：${persona.name}`)
  }

  const deletePersona = (id: string) => {
    const updated = personas.filter(p => p.id !== id)
    savePersonas(updated)
    if (selectedPersona?.id === id) {
      setSelectedPersona(null)
    }
    toast.success('已删除虚拟笔友')
  }

  const sendLetter = async () => {
    if (!letterContent.trim() || !selectedPersona) {
      toast.error('请输入信件内容')
      return
    }

    // 添加用户的信件
    const userLetter: LetterExchange = {
      id: `letter_${Date.now()}`,
      from: 'user',
      content: letterContent,
      timestamp: new Date(),
      personaId: selectedPersona.id,
    }

    setLetters(prev => [...prev, userLetter])
    setLoading(true)
    const currentContent = letterContent
    setLetterContent('')

    try {
      // 构建包含记忆和性格的提示词
      const context = {
        name: selectedPersona.name,
        relationship: selectedPersona.relationship,
        lastContact: selectedPersona.lastContactDate,
        memories: selectedPersona.memories.join('；'),
        personality: selectedPersona.personality,
        writingStyle: selectedPersona.writingStyle,
        letterContent: currentContent,
      }

      // 使用AI服务生成回信
      const prompt = `你现在扮演${context.name}，一个${context.relationship}。
      
背景信息：
- 最后联系时间：${context.lastContact || '很久以前'}
- 共同回忆：${context.memories}
- 性格特征：${context.personality}
- 写作风格：${context.writingStyle}

请以${context.name}的身份回复这封信，保持其性格和写作风格，适当提及共同回忆。回信要温暖、真实，让人感觉像是真的收到了来自${context.relationship}的信。

用户的信内容：
${context.letterContent}`

      // Use generateReply for AI letter responses
      const mockLetterId = Date.now().toString()
      const response = await aiService.generateReply({
        letterId: mockLetterId,
        persona: selectedPersona.name
      })

      // 添加AI回信
      const aiLetter: LetterExchange = {
        id: `ai_${Date.now()}`,
        from: 'persona',
        content: response.reply_content,
        timestamp: new Date(),
        personaId: selectedPersona.id,
      }

      setLetters(prev => [...prev, aiLetter])
      toast.success('收到回信了！')
    } catch (error) {
      console.error('Failed to get AI reply:', error)
      toast.error('生成回信失败，请稍后再试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* 人物选择区 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              云中锦书 - 虚拟笔友
            </span>
            <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
              <DialogTrigger asChild>
                <Button size="sm" variant="outline">
                  <UserPlus className="h-4 w-4 mr-2" />
                  创建虚拟笔友
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                  <DialogTitle>创建虚拟笔友</DialogTitle>
                  <DialogDescription>
                    为那些无法再联系的人创建一个虚拟形象，通过AI技术重温对话的温暖
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 mt-4">
                  <div>
                    <Label htmlFor="name">姓名</Label>
                    <Input
                      id="name"
                      value={newPersona.name}
                      onChange={(e) => setNewPersona({ ...newPersona, name: e.target.value })}
                      placeholder="例如：奶奶、小明"
                    />
                  </div>
                  <div>
                    <Label htmlFor="relationship">关系</Label>
                    <Input
                      id="relationship"
                      value={newPersona.relationship}
                      onChange={(e) => setNewPersona({ ...newPersona, relationship: e.target.value })}
                      placeholder="例如：已故的祖母、失联的朋友、分离的恋人"
                    />
                  </div>
                  <div>
                    <Label htmlFor="lastContact">最后联系时间（可选）</Label>
                    <Input
                      id="lastContact"
                      type="date"
                      value={newPersona.lastContactDate}
                      onChange={(e) => setNewPersona({ ...newPersona, lastContactDate: e.target.value })}
                    />
                  </div>
                  <div>
                    <Label htmlFor="memories">共同回忆（每行一个）</Label>
                    <Textarea
                      id="memories"
                      value={newPersona.memories}
                      onChange={(e) => setNewPersona({ ...newPersona, memories: e.target.value })}
                      placeholder="例如：
小时候一起去公园放风筝
每年春节包饺子
教我骑自行车"
                      rows={4}
                    />
                  </div>
                  <div>
                    <Label htmlFor="personality">性格特征</Label>
                    <Input
                      id="personality"
                      value={newPersona.personality}
                      onChange={(e) => setNewPersona({ ...newPersona, personality: e.target.value })}
                      placeholder="例如：温柔慈祥、幽默风趣、严肃认真"
                    />
                  </div>
                  <div>
                    <Label htmlFor="writingStyle">写作风格</Label>
                    <Input
                      id="writingStyle"
                      value={newPersona.writingStyle}
                      onChange={(e) => setNewPersona({ ...newPersona, writingStyle: e.target.value })}
                      placeholder="例如：亲切温暖、简洁有力、诗意浪漫"
                    />
                  </div>
                  <Button onClick={createPersona} className="w-full">
                    创建虚拟笔友
                  </Button>
                </div>
              </DialogContent>
            </Dialog>
          </CardTitle>
          <CardDescription>
            与那些再也无法相见的人进行虚拟对话，重温美好回忆
          </CardDescription>
        </CardHeader>
        <CardContent>
          {personas.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Heart className="h-8 w-8 mx-auto mb-3 text-pink-400 animate-pulse" />
              <p>还没有创建虚拟笔友</p>
              <p className="text-sm">点击上方按钮，为思念的人创建一个虚拟形象</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {personas.map((persona) => (
                <Card
                  key={persona.id}
                  className={`cursor-pointer transition-all ${
                    selectedPersona?.id === persona.id
                      ? 'border-blue-500 shadow-md'
                      : 'hover:shadow-sm'
                  }`}
                  onClick={() => setSelectedPersona(persona)}
                >
                  <CardContent className="p-4">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <Avatar className="h-8 w-8">
                          <AvatarFallback className="text-xs">
                            {persona.name.slice(0, 2)}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <p className="font-medium">{persona.name}</p>
                          <p className="text-xs text-muted-foreground">{persona.relationship}</p>
                        </div>
                      </div>
                      <Button
                        size="icon"
                        variant="ghost"
                        className="h-6 w-6"
                        onClick={(e) => {
                          e.stopPropagation()
                          deletePersona(persona.id)
                        }}
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {persona.memories.length} 个共同回忆
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 对话区域 */}
      {selectedPersona && (
        <>
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                与 {selectedPersona.name} 的书信往来
              </CardTitle>
              <CardDescription>
                {selectedPersona.relationship} · {selectedPersona.personality}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4 max-h-96 overflow-y-auto">
                {letters
                  .filter(l => l.personaId === selectedPersona.id)
                  .map((letter) => (
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
                          <span className="text-xs opacity-75">
                            {letter.from === 'user' ? '你' : selectedPersona.name}
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
                {letters.filter(l => l.personaId === selectedPersona.id).length === 0 && (
                  <div className="text-center py-8 text-muted-foreground">
                    <p>还没有开始对话</p>
                    <p className="text-sm">写下你想说的话，{selectedPersona.name}会以TA的方式回应你</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* 写信区域 */}
          <Card>
            <CardHeader>
              <CardTitle>写信给 {selectedPersona.name}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <Textarea
                placeholder={`写下你想对${selectedPersona.name}说的话...`}
                value={letterContent}
                onChange={(e) => setLetterContent(e.target.value)}
                rows={6}
                className="resize-none"
              />
              <div className="flex items-center justify-between">
                <div className="text-xs text-muted-foreground">
                  基于共同回忆和性格特征生成回信
                </div>
                <Button
                  onClick={sendLetter}
                  disabled={loading || !letterContent.trim()}
                  className="gap-2"
                >
                  {loading ? (
                    <>等待回信...</>
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
        </>
      )}

      {/* 使用提示 */}
      <Alert>
        <Heart className="h-4 w-4" />
        <AlertDescription>
          <strong>云中锦书：</strong>通过AI技术，让那些无法再联系的人以虚拟的方式陪伴在你身边。每一次对话都基于你提供的回忆和性格特征，让回信更加真实和温暖。
        </AlertDescription>
      </Alert>
    </div>
  )
}