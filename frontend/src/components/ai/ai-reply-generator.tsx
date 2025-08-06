'use client'

import React, { useState } from 'react'
import { Bot, Send, Loader2, Copy, RefreshCw, Wand2 } from 'lucide-react'
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
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { AIPersonaSelector, AIPersonaPreview } from './ai-persona-selector'
import { aiService } from '@/lib/services/ai-service'
import { toast } from 'sonner'

interface AIReplyGeneratorProps {
  letterId: string
  letterContent?: string
  onUseReply?: (reply: string) => void
  className?: string
}

export function AIReplyGenerator({
  letterId,
  letterContent,
  onUseReply,
  className = '',
}: AIReplyGeneratorProps) {
  const [selectedPersona, setSelectedPersona] = useState<string>('friend')
  const [generatedReply, setGeneratedReply] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [editedReply, setEditedReply] = useState<string>('')
  const [isEditing, setIsEditing] = useState(false)

  const generateReply = async () => {
    if (!letterId) {
      toast.error('请先保存信件草稿')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const result = await aiService.generateReply({
        letterId: letterId,
        persona: selectedPersona,
      })

      if (result.reply_content) {
        setGeneratedReply(result.reply_content)
        setEditedReply(result.reply_content)
        toast.success('回信生成成功')
      } else {
        throw new Error('未获取到回信内容')
      }
    } catch (err) {
      console.error('Failed to generate reply:', err)
      setError('生成回信失败，请稍后重试')
      toast.error('生成失败')
    } finally {
      setLoading(false)
    }
  }

  const handleUseReply = () => {
    const replyToUse = isEditing ? editedReply : generatedReply
    if (onUseReply && replyToUse) {
      onUseReply(replyToUse)
      toast.success('已使用AI生成的回信')
    }
  }

  const handleCopyReply = () => {
    const replyToCopy = isEditing ? editedReply : generatedReply
    navigator.clipboard.writeText(replyToCopy)
    toast.success('已复制到剪贴板')
  }

  const regenerateReply = () => {
    setGeneratedReply('')
    setEditedReply('')
    setIsEditing(false)
    generateReply()
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* 人设选择 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bot className="h-5 w-5" />
            选择AI回信风格
          </CardTitle>
          <CardDescription>
            选择一个AI人设，生成符合该风格的回信
          </CardDescription>
        </CardHeader>
        <CardContent>
          <AIPersonaSelector
            value={selectedPersona}
            onChange={setSelectedPersona}
          />
        </CardContent>
      </Card>

      {/* 当前选中的人设预览 */}
      {selectedPersona && (
        <AIPersonaPreview personaId={selectedPersona} />
      )}

      {/* 生成按钮 */}
      <div className="flex justify-center">
        <Button
          onClick={generateReply}
          disabled={loading || !letterId}
          size="lg"
          className="gap-2"
        >
          {loading ? (
            <>
              <Loader2 className="h-5 w-5 animate-spin" />
              生成中...
            </>
          ) : (
            <>
              <Wand2 className="h-5 w-5" />
              生成AI回信
            </>
          )}
        </Button>
      </div>

      {/* 错误提示 */}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 生成的回信 */}
      {generatedReply && (
        <Card className="border-2 border-primary/20">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg">AI生成的回信</CardTitle>
              <div className="flex items-center gap-2">
                <Badge variant="secondary">
                  {selectedPersona === 'poet' ? '诗人' :
                   selectedPersona === 'philosopher' ? '哲学家' :
                   selectedPersona === 'artist' ? '艺术家' :
                   selectedPersona === 'scientist' ? '科学家' :
                   selectedPersona === 'traveler' ? '旅行者' :
                   selectedPersona === 'historian' ? '历史学家' :
                   selectedPersona === 'mentor' ? '导师' : '朋友'}风格
                </Badge>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <Tabs value={isEditing ? 'edit' : 'preview'} onValueChange={(v) => setIsEditing(v === 'edit')}>
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="preview">预览</TabsTrigger>
                <TabsTrigger value="edit">编辑</TabsTrigger>
              </TabsList>
              <TabsContent value="preview" className="mt-4">
                <div className="whitespace-pre-wrap bg-muted/50 rounded-lg p-4 min-h-[200px] font-serif text-base leading-relaxed">
                  {generatedReply}
                </div>
              </TabsContent>
              <TabsContent value="edit" className="mt-4">
                <Textarea
                  value={editedReply}
                  onChange={(e) => setEditedReply(e.target.value)}
                  className="min-h-[200px] font-serif text-base leading-relaxed"
                  placeholder="编辑AI生成的回信..."
                />
              </TabsContent>
            </Tabs>

            {/* 操作按钮 */}
            <div className="flex flex-wrap gap-3 pt-2">
              <Button onClick={handleUseReply} className="gap-2">
                <Send className="h-4 w-4" />
                使用此回信
              </Button>
              <Button onClick={handleCopyReply} variant="outline" className="gap-2">
                <Copy className="h-4 w-4" />
                复制
              </Button>
              <Button onClick={regenerateReply} variant="outline" className="gap-2">
                <RefreshCw className="h-4 w-4" />
                重新生成
              </Button>
            </div>

            <Alert>
              <AlertDescription className="text-sm">
                AI生成的回信仅供参考，建议根据实际情况进行修改，保持真实和个人风格
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      )}
    </div>
  )
}