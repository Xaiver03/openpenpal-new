'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { WelcomeBanner } from '@/components/ui/welcome-banner'
import { 
  Mail, 
  Save, 
  Send, 
  FileText, 
  Palette,
  Download,
  QrCode,
  CheckCircle,
  AlertCircle,
  Sparkles
} from 'lucide-react'
import { useLetterStore } from '@/stores/letter-store'
import { createLetterDraft, generateLetterCode } from '@/lib/api'
import { LetterService } from '@/lib/services/letter-service'
import type { LetterStyle, LetterTemplate } from '@/types/letter'
import { useUnsavedChanges } from '@/hooks/use-unsaved-changes'
import { SafeBackButton } from '@/components/ui/safe-back-button'
import { AIWritingInspiration } from '@/components/ai/ai-writing-inspiration'
import { AIDailyInspiration } from '@/components/ai/ai-daily-inspiration'
import { AIPenpalMatch } from '@/components/ai/ai-penpal-match'
import { AIReplyGenerator } from '@/components/ai/ai-reply-generator'
import { RichTextEditor } from '@/components/editor/rich-text-editor'
import { stripHtml, getPlainTextLength } from '@/lib/utils/html'

export default function WritePage() {
  const searchParams = useSearchParams()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [selectedStyle, setSelectedStyle] = useState<LetterStyle>('classic')
  const [isGeneratingCode, setIsGeneratingCode] = useState(false)
  const [generatedCode, setGeneratedCode] = useState<string | null>(null)
  const [qrCodeImage, setQrCodeImage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isReplyMode, setIsReplyMode] = useState(false)
  const [replyToInfo, setReplyToInfo] = useState<{
    code: string
    sender: string
    title: string
  } | null>(null)
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false)
  const [templates, setTemplates] = useState<LetterTemplate[]>([])
  const [isLoadingTemplates, setIsLoadingTemplates] = useState(false)
  const [selectedTemplate, setSelectedTemplate] = useState<LetterTemplate | null>(null)
  const [showAIInspiration, setShowAIInspiration] = useState(false)
  const [currentLetterId, setCurrentLetterId] = useState<string | null>(null)
  const [showAIPenpalMatch, setShowAIPenpalMatch] = useState(false)
  const [showAIReplyGenerator, setShowAIReplyGenerator] = useState(false)
  
  const { createDraft, saveDraft, currentDraft } = useLetterStore()

  const letterStyles: { id: LetterStyle; name: string; description: string; preview: string }[] = [
    {
      id: 'classic',
      name: '经典',
      description: '传统信纸样式，简洁优雅',
      preview: '#fdfcf9'
    },
    {
      id: 'modern',
      name: '现代',
      description: '现代简约风格，清新明快',
      preview: '#ffffff'
    },
    {
      id: 'vintage',
      name: '复古',
      description: '复古怀旧风格，温馨怀旧',
      preview: '#f4f1e8'
    },
    {
      id: 'elegant',
      name: '优雅',
      description: '优雅精致，适合正式信件',
      preview: '#f8f7f4'
    },
  ]

  // 根据选中的样式返回textarea的样式
  const getTextareaStyle = (style: LetterStyle): React.CSSProperties => {
    switch (style) {
      case 'classic':
        return {
          backgroundImage: 'linear-gradient(transparent 29px, rgba(139, 69, 19, 0.2) 1px)',
          backgroundSize: '100% 30px'
        }
      case 'modern':
        return {
          backgroundImage: 'linear-gradient(transparent 29px, rgba(0, 0, 0, 0.1) 1px)',
          backgroundSize: '100% 30px'
        }
      case 'vintage':
        return {
          backgroundImage: 'linear-gradient(transparent 31px, rgba(120, 45, 18, 0.25) 1px)',
          backgroundSize: '100% 32px'
        }
      case 'elegant':
        return {
          backgroundImage: 'linear-gradient(transparent 27px, rgba(139, 69, 19, 0.15) 1px)',
          backgroundSize: '100% 28px'
        }
      default:
        return {
          backgroundImage: 'linear-gradient(transparent 29px, rgba(139, 69, 19, 0.2) 1px)',
          backgroundSize: '100% 30px'
        }
    }
  }

  // 加载信件模板
  useEffect(() => {
    const loadTemplates = async () => {
      setIsLoadingTemplates(true)
      try {
        const response = await LetterService.getTemplates({ limit: 10 })
        if (response.data?.templates) {
          setTemplates(response.data.templates)
        }
      } catch (error) {
        console.error('Failed to load templates:', error)
      } finally {
        setIsLoadingTemplates(false)
      }
    }
    loadTemplates()
  }, [])

  // 检查是否是回信模式
  useEffect(() => {
    const replyTo = searchParams?.get('reply_to')
    const replyToSender = searchParams?.get('reply_to_sender')
    const replyToTitle = searchParams?.get('reply_to_title')
    
    if (replyTo && replyToSender && replyToTitle) {
      setIsReplyMode(true)
      setReplyToInfo({
        code: replyTo,
        sender: replyToSender,
        title: replyToTitle
      })
      
      // 设置回信标题和初始内容
      setTitle(`回信：${replyToTitle}`)
      setContent(`亲爱的${replyToSender}，

感谢你的来信，我很高兴收到你的信件。

`)
    }
  }, [searchParams])
  
  // 检测内容变化
  useEffect(() => {
    if (title.trim() || content.trim()) {
      setHasUnsavedChanges(true)
    }
  }, [title, content])

  const handleSaveDraft = () => {
    const plainContent = stripHtml(content)
    if (!plainContent.trim()) return
    
    if (currentDraft) {
      saveDraft({
        ...currentDraft,
        title,
        content: plainContent,
        style: selectedStyle,
        updatedAt: new Date()
      })
    } else {
      createDraft(plainContent, selectedStyle)
    }
    
    setHasUnsavedChanges(false)
  }
  
  // 未保存更改检测
  useUnsavedChanges({
    hasUnsavedChanges,
    message: '您有未保存的信件草稿。是否要在离开前保存？',
    onSave: handleSaveDraft,
    onDiscard: () => {
      setTitle('')
      setContent('')
      setHasUnsavedChanges(false)
    }
  })

  const handleGenerateCode = async () => {
    const plainContent = stripHtml(content)
    if (!plainContent.trim()) return
    
    setIsGeneratingCode(true)
    setError(null)
    
    try {
      // 先创建草稿
      const draftResult = await createLetterDraft({
        title,
        content: plainContent,
        style: selectedStyle,
      })
      
      if (!draftResult.data) {
        throw new Error('创建草稿失败')
      }
      
      const letterData = draftResult.data as any
      if (!letterData.id) {
        throw new Error('创建草稿失败：无效的信件ID')
      }
      
      // 保存当前信件ID供AI功能使用
      setCurrentLetterId(letterData.id)
      
      // 生成编号和二维码
      const result = await generateLetterCode(letterData.id)
      
      if (result.data) {
        const codeData = result.data as any
        setGeneratedCode(codeData.letter_code)
        setQrCodeImage(codeData.qrCodeUrl)
      }
      setError(null)
    } catch (error) {
      console.error('生成编号失败:', error)
      setError('生成编号失败，请稍后重试')
    } finally {
      setIsGeneratingCode(false)
    }
  }

  // 应用模板
  const applyTemplate = (template: LetterTemplate) => {
    setSelectedTemplate(template)
    setContent(template.content_template)
    setHasUnsavedChanges(true)
  }

  return (
    <div className="container max-w-7xl mx-auto px-4 py-8">
      <WelcomeBanner />
      
      {/* Header */}
      <div className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-letter-ink mb-2">
          {isReplyMode ? '回信' : '写信'}
        </h1>
        <p className="text-muted-foreground">
          {isReplyMode 
            ? '回复一封温暖的信件，延续这份美好的连接' 
            : '用文字传递真实情感，每一个字都承载着温度'
          }
        </p>
        
        {/* 回信提示信息 */}
        {isReplyMode && replyToInfo && (
          <Alert className="mt-4">
            <Mail className="h-4 w-4" />
            <AlertDescription>
              你正在回复 <strong>{replyToInfo.sender}</strong> 的信件《{replyToInfo.title}》
              （编号：{replyToInfo.code}）
            </AlertDescription>
          </Alert>
        )}
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-4 gap-8">
        {/* 写信区域 */}
        <div className="xl:col-span-3 space-y-6">
          {/* 信件标题 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileText className="h-5 w-5" />
                信件标题（可选）
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Input
                placeholder="给这封信起个标题吧..."
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="text-lg"
              />
            </CardContent>
          </Card>

          {/* 信件内容 */}
          <Card className={`letter-paper ${selectedStyle}`}>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Mail className="h-5 w-5" />
                信件内容
              </CardTitle>
              <CardDescription>
                在这里编写你的信件草稿，稍后需要手写到实体信纸上
              </CardDescription>
            </CardHeader>
            <CardContent>
              <RichTextEditor
                content={content}
                onChange={setContent}
                placeholder="亲爱的朋友，\n\n见字如面...\n\n此时此刻，我想对你说..."
                className="font-serif text-base leading-loose"
                style={getTextareaStyle(selectedStyle)}
                maxLength={2000}
              />
            </CardContent>
          </Card>

          {/* 错误信息 */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* 成功生成编号 */}
          {generatedCode && qrCodeImage && (
            <Alert>
              <CheckCircle className="h-4 w-4" />
              <AlertDescription>
                编号生成成功！信件编号：<span className="font-mono font-semibold">{generatedCode}</span>
              </AlertDescription>
            </Alert>
          )}

          {/* 操作按钮 */}
          <div className="flex flex-wrap gap-4">
            <Button 
              onClick={handleSaveDraft}
              variant="outline"
              disabled={!stripHtml(content).trim()}
            >
              <Save className="mr-2 h-4 w-4" />
              保存草稿
            </Button>
            <Button 
              onClick={handleGenerateCode}
              disabled={!stripHtml(content).trim() || isGeneratingCode || !!generatedCode}
              className="font-serif"
            >
              {isGeneratingCode ? (
                <>
                  <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                  生成中...
                </>
              ) : generatedCode ? (
                <>
                  <CheckCircle className="mr-2 h-4 w-4" />
                  已生成编号
                </>
              ) : (
                <>
                  <QrCode className="mr-2 h-4 w-4" />
                  生成编号贴纸
                </>
              )}
            </Button>
          </div>
        </div>

        {/* 侧边栏 */}
        <div className="space-y-6 xl:min-w-[320px]">
          {/* AI写作灵感 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span className="flex items-center gap-2">
                  <Sparkles className="h-5 w-5 text-amber-600" />
                  云锦传驿
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowAIInspiration(!showAIInspiration)}
                >
                  {showAIInspiration ? '收起' : '展开'}
                </Button>
              </CardTitle>
            </CardHeader>
            {showAIInspiration && (
              <CardContent className="space-y-4">
                <AIDailyInspiration
                  onSelectPrompt={(prompt) => {
                    setContent(prompt)
                    setHasUnsavedChanges(true)
                  }}
                />
                <AIWritingInspiration
                  theme={isReplyMode ? '回信' : '日常生活'}
                  onSelectInspiration={(inspiration) => {
                    setContent(content ? `${content}\n\n${inspiration.prompt}` : inspiration.prompt)
                    setHasUnsavedChanges(true)
                  }}
                />
              </CardContent>
            )}
          </Card>

          {/* AI笔友匹配 */}
          {!isReplyMode && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <span className="flex items-center gap-2">
                    <Sparkles className="h-5 w-5 text-purple-600" />
                    AI笔友匹配
                  </span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowAIPenpalMatch(!showAIPenpalMatch)}
                  >
                    {showAIPenpalMatch ? '收起' : '展开'}
                  </Button>
                </CardTitle>
                <CardDescription>
                  AI智能推荐合适的笔友
                </CardDescription>
              </CardHeader>
              {showAIPenpalMatch && (
                <CardContent>
                  <AIPenpalMatch
                    letterId={currentLetterId || ''}
                    onSelectMatch={(match) => {
                      console.log('选择了笔友:', match)
                      // 这里可以实现选择笔友后的逻辑
                    }}
                  />
                </CardContent>
              )}
            </Card>
          )}

          {/* AI回信生成 */}
          {isReplyMode && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <span className="flex items-center gap-2">
                    <Sparkles className="h-5 w-5 text-blue-600" />
                    云锦传驿
                  </span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowAIReplyGenerator(!showAIReplyGenerator)}
                  >
                    {showAIReplyGenerator ? '收起' : '展开'}
                  </Button>
                </CardTitle>
                <CardDescription>
                  AI智能生成回信内容
                </CardDescription>
              </CardHeader>
              {showAIReplyGenerator && (
                <CardContent>
                  <AIReplyGenerator
                    letterId={currentLetterId || ''}
                    letterContent={content}
                    onUseReply={(reply) => {
                      setContent(reply)
                      setHasUnsavedChanges(true)
                    }}
                  />
                </CardContent>
              )}
            </Card>
          )}

          {/* 二维码预览 */}
          {generatedCode && qrCodeImage && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <QrCode className="h-5 w-5" />
                  二维码贴纸
                </CardTitle>
                <CardDescription>
                  打印并贴在信封上
                </CardDescription>
              </CardHeader>
              <CardContent className="text-center space-y-4">
                <div className="mx-auto w-48 h-48 bg-white rounded-lg border border-border p-4 flex items-center justify-center">
                  <img 
                    src={qrCodeImage} 
                    alt="信件二维码" 
                    className="w-full h-full object-contain"
                  />
                </div>
                <div className="space-y-2">
                  <div className="text-sm font-mono bg-muted p-2 rounded text-center">
                    {generatedCode}
                  </div>
                  <Button 
                    variant="outline" 
                    size="sm" 
                    className="w-full"
                    onClick={() => window.print()}
                  >
                    <Download className="mr-2 h-4 w-4" />
                    打印贴纸
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}

          {/* 信纸样式选择 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Palette className="h-5 w-5" />
                信纸样式
              </CardTitle>
              <CardDescription>
                选择你喜欢的信纸风格
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {letterStyles.map((style) => (
                <div 
                  key={style.id}
                  className={`p-3 rounded-lg border cursor-pointer transition-all ${
                    selectedStyle === style.id 
                      ? 'border-primary bg-primary/5' 
                      : 'border-border hover:border-muted-foreground'
                  }`}
                  onClick={() => setSelectedStyle(style.id)}
                >
                  <div className="flex items-center gap-3">
                    <div 
                      className="w-8 h-8 rounded border"
                      style={{ backgroundColor: style.preview }}
                    />
                    <div className="flex-1">
                      <div className="font-medium text-sm">{style.name}</div>
                      <div className="text-xs text-muted-foreground">{style.description}</div>
                    </div>
                    {selectedStyle === style.id && (
                      <Badge variant="default" className="text-xs">已选</Badge>
                    )}
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>

          {/* 信件模板 */}
          {!isReplyMode && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <FileText className="h-5 w-5" />
                  信件模板
                </CardTitle>
                <CardDescription>
                  选择一个模板快速开始
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-2">
                {isLoadingTemplates ? (
                  <div className="text-center py-4 text-sm text-muted-foreground">
                    加载模板中...
                  </div>
                ) : templates.length > 0 ? (
                  templates.map((template) => (
                    <div
                      key={template.id}
                      className={`p-3 rounded-lg border cursor-pointer transition-all hover:shadow-sm ${
                        selectedTemplate?.id === template.id
                          ? 'border-primary bg-primary/5'
                          : 'border-border hover:border-muted-foreground'
                      }`}
                      onClick={() => applyTemplate(template)}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex-1">
                          <div className="font-medium text-sm">{template.name}</div>
                          <div className="text-xs text-muted-foreground mt-0.5">
                            {template.description}
                          </div>
                        </div>
                        {template.is_premium && (
                          <Badge variant="secondary" className="text-xs">高级</Badge>
                        )}
                      </div>
                      {selectedTemplate?.id === template.id && (
                        <Badge variant="default" className="text-xs mt-2">使用中</Badge>
                      )}
                    </div>
                  ))
                ) : (
                  <div className="text-center py-4 text-sm text-muted-foreground">
                    暂无可用模板
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* 写信提示 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">
                {isReplyMode ? '💌 回信提示' : '📝 写信提示'}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm text-muted-foreground">
              {isReplyMode ? (
                <>
                  <div className="flex items-start gap-2">
                    <span className="text-primary">•</span>
                    <span>回信会关联到原信件编号，对方能看到你的回复</span>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="text-primary">•</span>
                    <span>表达你的真实感受，分享你的想法和故事</span>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="text-primary">•</span>
                    <span>回信同样需要手写，完成后投递给信使</span>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="text-primary">•</span>
                    <span>温暖的回复是最好的鼓励，让连接延续下去</span>
                  </div>
                </>
              ) : (
                <>
                  <div className="flex items-start gap-2">
                    <span className="text-primary">•</span>
                    <span>平台内容仅作草稿参考，实际需要手写到信纸上</span>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="text-primary">•</span>
                    <span>生成编号后会得到二维码贴纸，请贴在信封上</span>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="text-primary">•</span>
                    <span>建议使用钢笔或中性笔，字迹清晰易读</span>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="text-primary">•</span>
                    <span>信件内容积极向上，传递正能量</span>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}