'use client'

import { useState, useEffect } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { WelcomeBanner } from '@/components/ui/welcome-banner'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
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
  Sparkles,
  Upload,
  Edit3,
  ArrowLeft,
  Heart,
  Calendar,
  Clock,
  MapPin,
  Tag,
  Waves
} from 'lucide-react'
import { useLetterStore } from '@/stores/letter-store'
import { createLetterDraft, generateLetterCode } from '@/lib/api'
import { LetterService } from '@/lib/services/letter-service'
import { driftBottleApi } from '@/lib/api/drift-bottle'
import { futureLetterApi } from '@/lib/api/future-letter'
import { driftBottleAIApi } from '@/lib/api/drift-bottle-ai'
import type { LetterStyle } from '@/types/letter'
import { useUnsavedChanges } from '@/hooks/use-unsaved-changes'
import { SafeBackButton } from '@/components/ui/safe-back-button'
import { AIWritingInspiration } from '@/components/ai/ai-writing-inspiration'
import { AIDailyInspiration } from '@/components/ai/ai-daily-inspiration'
import { AIPenpalMatch } from '@/components/ai/ai-penpal-match'
import { AIReplyGenerator } from '@/components/ai/ai-reply-generator'
import { RichTextEditor } from '@/components/editor/rich-text-editor'
import { stripHtml, getPlainTextLength } from '@/lib/utils/html'
import { HandwrittenUpload } from '@/components/write/handwritten-upload'
import { toast } from 'sonner'

// 信件类型定义
type LetterType = 'normal' | 'drift' | 'future'

export default function WriteLetterPage() {
  const router = useRouter()
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
  const [showAIInspiration, setShowAIInspiration] = useState(false)
  const [currentLetterId, setCurrentLetterId] = useState<string | null>(null)
  const [showAIPenpalMatch, setShowAIPenpalMatch] = useState(false)
  const [showAIReplyGenerator, setShowAIReplyGenerator] = useState(false)
  const [activeTab, setActiveTab] = useState<'compose' | 'upload'>('compose')
  const [uploadedImages, setUploadedImages] = useState<any[]>([])
  const [extractedText, setExtractedText] = useState<string>('')
  
  // 新增: 信件类型选择
  const [letterType, setLetterType] = useState<LetterType>('normal')
  
  // 新增: AI匹配结果
  const [aiMatchInfo, setAiMatchInfo] = useState<{
    matched: boolean
    compatibility_score?: number
    match_reason?: string
  } | null>(null)
  
  // 新增: 漂流瓶配置
  const [driftConfig, setDriftConfig] = useState({
    theme: '',
    region: '',
    aiMatch: true,
    delayMinutes: 30
  })
  
  // 新增: 未来信配置
  const [futureConfig, setFutureConfig] = useState({
    deliveryDate: '',
    deliveryTime: '',
    reminderEnabled: false,
    reminderDays: 7
  })
  
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
  
  // 检查信件类型参数
  useEffect(() => {
    const type = searchParams?.get('type')
    if (type === 'drift' || type === 'future') {
      setLetterType(type)
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
        updated_at: new Date()
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
      
      // 根据信件类型创建相应的记录
      if (letterType === 'drift') {
        // 创建漂流瓶
        try {
          if (driftConfig.aiMatch) {
            // 使用AI匹配创建漂流瓶
            const aiResult = await driftBottleAIApi.createWithAIMatch({
              letter_id: letterData.id,
              theme: driftConfig.theme,
              region: driftConfig.region,
              days: Math.ceil(driftConfig.delayMinutes / 1440) || 7,
              letter_content: plainContent,
              letter_title: title || '无标题',
              writer_profile: {
                interests: driftConfig.theme ? [driftConfig.theme] : undefined,
                mood: selectedStyle
              },
              match_preferences: {
                same_school_preferred: driftConfig.region === 'same-city',
                compatibility_threshold: 0.7
              }
            })
            
            if (!aiResult.drift_bottle || !aiResult.drift_bottle.id) {
              throw new Error('创建漂流瓶失败')
            }
            
            // 如果有AI匹配信息，保存到状态中
            if (aiResult.match_info?.matched) {
              setAiMatchInfo({
                matched: true,
                compatibility_score: aiResult.match_info.recipient_profile?.compatibility_score,
                match_reason: aiResult.match_info.recipient_profile?.match_reason
              })
            }
          } else {
            // 不使用AI匹配，使用普通方式创建
            const driftResult = await driftBottleApi.create({
              letter_id: letterData.id,
              theme: driftConfig.theme,
              region: driftConfig.region,
              days: Math.ceil(driftConfig.delayMinutes / 1440) || 7
            })
            
            if (!driftResult || !driftResult.id) {
              throw new Error('创建漂流瓶失败')
            }
          }
        } catch (error: any) {
          throw new Error('创建漂流瓶失败：' + (error.message || '未知错误'))
        }
      } else if (letterType === 'future') {
        // 创建未来信
        const deliveryDateTime = new Date(`${futureConfig.deliveryDate}T${
          futureConfig.deliveryTime === 'morning' ? '09:00' :
          futureConfig.deliveryTime === 'noon' ? '13:00' :
          futureConfig.deliveryTime === 'afternoon' ? '16:00' :
          futureConfig.deliveryTime === 'evening' ? '19:00' : '22:00'
        }:00`)
        
        try {
          const futureResult = await futureLetterApi.schedule({
            letter_id: letterData.id,
            scheduled_date: deliveryDateTime,
            recipient_id: '', // 如果是写给自己，后端会处理
            reminder_enabled: futureConfig.reminderEnabled,
            reminder_days: futureConfig.reminderDays
          })
          
          if (!futureResult || !futureResult.id) {
            throw new Error('创建未来信失败')
          }
        } catch (error: any) {
          throw new Error('创建未来信失败：' + (error.message || '未知错误'))
        }
      }
      
      // 生成编号和二维码
      const result = await generateLetterCode(letterData.id)
      
      if (result.data) {
        const codeData = result.data as any
        setGeneratedCode(codeData.letter_code)
        setQrCodeImage(codeData.qr_code_url || codeData.qrCodeUrl)
      }
      setError(null)
    } catch (error) {
      console.error('生成编号失败:', error)
      setError('生成编号失败，请稍后重试')
    } finally {
      setIsGeneratingCode(false)
    }
  }


  return (
    <div className="container max-w-7xl mx-auto px-4 py-8">
      <WelcomeBanner />
      
      {/* Header with back button */}
      <div className="mb-8">
        <Button
          variant="ghost"
          onClick={() => router.push('/letters')}
          className="mb-4"
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          返回信件列表
        </Button>
        
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
          <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as 'compose' | 'upload')} className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="compose" className="flex items-center gap-2">
                <Edit3 className="h-4 w-4" />
                在线编写
              </TabsTrigger>
              <TabsTrigger value="upload" className="flex items-center gap-2">
                <Upload className="h-4 w-4" />
                上传手写信
              </TabsTrigger>
            </TabsList>
            
            {/* 在线编写标签页 */}
            <TabsContent value="compose" className="space-y-6">
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

              {/* 信件类型选择 */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Send className="h-5 w-5" />
                    选择信件类型
                  </CardTitle>
                  <CardDescription>
                    选择这封信的投递方式
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <RadioGroup 
                    value={letterType} 
                    onValueChange={(value: LetterType) => setLetterType(value)}
                    className="space-y-4"
                  >
                    <div className="flex items-start space-x-3 p-4 border border-amber-200 rounded-lg hover:bg-amber-50 transition-colors">
                      <RadioGroupItem value="normal" id="normal" className="mt-1" />
                      <div className="flex-1">
                        <Label htmlFor="normal" className="flex items-center gap-2 text-base font-semibold cursor-pointer">
                          <Mail className="w-5 h-5 text-amber-600" />
                          普通信件
                        </Label>
                        <p className="text-sm text-amber-700 mt-1">
                          传统的点对点投递，需要知道收件人的OP Code
                        </p>
                      </div>
                    </div>

                    <div className="flex items-start space-x-3 p-4 border border-amber-200 rounded-lg hover:bg-amber-50 transition-colors">
                      <RadioGroupItem value="drift" id="drift" className="mt-1" />
                      <div className="flex-1">
                        <Label htmlFor="drift" className="flex items-center gap-2 text-base font-semibold cursor-pointer">
                          <Waves className="w-5 h-5 text-blue-600" />
                          漂流瓶
                        </Label>
                        <p className="text-sm text-amber-700 mt-1">
                          让AI为你匹配一个陌生的朋友，开启温暖的相遇
                        </p>
                      </div>
                    </div>

                    <div className="flex items-start space-x-3 p-4 border border-amber-200 rounded-lg hover:bg-amber-50 transition-colors">
                      <RadioGroupItem value="future" id="future" className="mt-1" />
                      <div className="flex-1">
                        <Label htmlFor="future" className="flex items-center gap-2 text-base font-semibold cursor-pointer">
                          <Calendar className="w-5 h-5 text-purple-600" />
                          未来信
                        </Label>
                        <p className="text-sm text-amber-700 mt-1">
                          写给未来的自己或他人，在指定时间送达
                        </p>
                      </div>
                    </div>
                  </RadioGroup>
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

              {/* 漂流瓶配置 */}
              {letterType === 'drift' && (
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <Waves className="h-5 w-5 text-blue-600" />
                      漂流瓶设置
                    </CardTitle>
                    <CardDescription>
                      设置你的漂流瓶偏好，AI将基于这些信息进行匹配
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div>
                      <Label htmlFor="drift-theme">漂流主题</Label>
                      <Select 
                        value={driftConfig.theme} 
                        onValueChange={(value) => setDriftConfig(prev => ({ ...prev, theme: value }))}
                      >
                        <SelectTrigger id="drift-theme">
                          <SelectValue placeholder="选择一个主题" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="friendship">友情</SelectItem>
                          <SelectItem value="growth">成长</SelectItem>
                          <SelectItem value="dream">梦想</SelectItem>
                          <SelectItem value="emotion">情感</SelectItem>
                          <SelectItem value="story">故事</SelectItem>
                          <SelectItem value="random">随机</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div>
                      <Label htmlFor="drift-region">地区偏好</Label>
                      <Select 
                        value={driftConfig.region} 
                        onValueChange={(value) => setDriftConfig(prev => ({ ...prev, region: value }))}
                      >
                        <SelectTrigger id="drift-region">
                          <SelectValue placeholder="选择地区偏好" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="same-city">同城优先</SelectItem>
                          <SelectItem value="same-province">同省优先</SelectItem>
                          <SelectItem value="nationwide">全国随机</SelectItem>
                          <SelectItem value="no-preference">无偏好</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div>
                      <Label htmlFor="drift-delay">漂流延迟时间</Label>
                      <Select 
                        value={driftConfig.delayMinutes.toString()} 
                        onValueChange={(value) => setDriftConfig(prev => ({ ...prev, delayMinutes: parseInt(value) }))}
                      >
                        <SelectTrigger id="drift-delay">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="5">5分钟 - 快速漂流</SelectItem>
                          <SelectItem value="30">30分钟 - 标准漂流</SelectItem>
                          <SelectItem value="60">1小时 - 慢速漂流</SelectItem>
                          <SelectItem value="180">3小时 - 深度漂流</SelectItem>
                          <SelectItem value="1440">1天 - 长途漂流</SelectItem>
                          <SelectItem value="10080">1周 - 远航漂流</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="flex items-center justify-between">
                      <div className="space-y-0.5">
                        <Label htmlFor="ai-match">AI智能匹配</Label>
                        <p className="text-sm text-muted-foreground">
                          使用AI分析信件内容，找到最合适的收件人
                        </p>
                      </div>
                      <Switch
                        id="ai-match"
                        checked={driftConfig.aiMatch}
                        onCheckedChange={(checked) => setDriftConfig(prev => ({ ...prev, aiMatch: checked }))}
                      />
                    </div>

                    <Alert>
                      <Heart className="h-4 w-4" />
                      <AlertDescription>
                        漂流瓶会根据你的设置和信件内容，通过AI匹配到合适的陌生朋友。请真诚表达，让美好的相遇发生。
                      </AlertDescription>
                    </Alert>
                  </CardContent>
                </Card>
              )}

              {/* 未来信配置 */}
              {letterType === 'future' && (
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <Calendar className="h-5 w-5 text-purple-600" />
                      未来信设置
                    </CardTitle>
                    <CardDescription>
                      设置这封信的送达时间
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div>
                      <Label htmlFor="future-date">送达日期</Label>
                      <Input
                        id="future-date"
                        type="date"
                        value={futureConfig.deliveryDate}
                        onChange={(e) => setFutureConfig(prev => ({ ...prev, deliveryDate: e.target.value }))}
                        min={new Date(Date.now() + 86400000).toISOString().split('T')[0]} // 最少1天后
                      />
                    </div>

                    <div>
                      <Label htmlFor="future-time">送达时间</Label>
                      <Select 
                        value={futureConfig.deliveryTime} 
                        onValueChange={(value) => setFutureConfig(prev => ({ ...prev, deliveryTime: value }))}
                      >
                        <SelectTrigger id="future-time">
                          <SelectValue placeholder="选择送达时间" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="morning">早晨 (8:00-10:00)</SelectItem>
                          <SelectItem value="noon">中午 (12:00-14:00)</SelectItem>
                          <SelectItem value="afternoon">下午 (15:00-17:00)</SelectItem>
                          <SelectItem value="evening">傍晚 (18:00-20:00)</SelectItem>
                          <SelectItem value="night">晚上 (21:00-23:00)</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="flex items-center justify-between">
                      <div className="space-y-0.5">
                        <Label htmlFor="reminder">提前提醒</Label>
                        <p className="text-sm text-muted-foreground">
                          在信件送达前{futureConfig.reminderDays}天发送提醒
                        </p>
                      </div>
                      <Switch
                        id="reminder"
                        checked={futureConfig.reminderEnabled}
                        onCheckedChange={(checked) => setFutureConfig(prev => ({ ...prev, reminderEnabled: checked }))}
                      />
                    </div>

                    {futureConfig.reminderEnabled && (
                      <div>
                        <Label htmlFor="reminder-days">提前天数</Label>
                        <Select 
                          value={futureConfig.reminderDays.toString()} 
                          onValueChange={(value) => setFutureConfig(prev => ({ ...prev, reminderDays: parseInt(value) }))}
                        >
                          <SelectTrigger id="reminder-days">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="1">1天前</SelectItem>
                            <SelectItem value="3">3天前</SelectItem>
                            <SelectItem value="7">7天前</SelectItem>
                            <SelectItem value="14">14天前</SelectItem>
                            <SelectItem value="30">30天前</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    )}

                    <Alert>
                      <Clock className="h-4 w-4" />
                      <AlertDescription>
                        未来信将在指定时间送达。你可以写给未来的自己，也可以写给未来的朋友。时间会让这封信变得更有意义。
                      </AlertDescription>
                    </Alert>
                  </CardContent>
                </Card>
              )}

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
                    {letterType === 'drift' ? (
                      <>
                        漂流瓶编号生成成功！编号：<span className="font-mono font-semibold">{generatedCode}</span>
                        <br />
                        {aiMatchInfo?.matched ? (
                          <>
                            AI已为你找到合适的收件人！
                            {aiMatchInfo.compatibility_score && (
                              <>匹配度：{Math.round(aiMatchInfo.compatibility_score * 100)}%</>
                            )}
                            {aiMatchInfo.match_reason && (
                              <>，{aiMatchInfo.match_reason}</>
                            )}
                            <br />
                          </>
                        ) : null}
                        你的漂流瓶将在{driftConfig.delayMinutes}分钟后开始漂流。
                      </>
                    ) : letterType === 'future' ? (
                      <>
                        未来信编号生成成功！编号：<span className="font-mono font-semibold">{generatedCode}</span>
                        <br />
                        这封信将在 {futureConfig.deliveryDate} {
                          futureConfig.deliveryTime === 'morning' ? '早晨' :
                          futureConfig.deliveryTime === 'noon' ? '中午' :
                          futureConfig.deliveryTime === 'afternoon' ? '下午' :
                          futureConfig.deliveryTime === 'evening' ? '傍晚' : '晚上'
                        } 送达。
                      </>
                    ) : (
                      <>编号生成成功！信件编号：<span className="font-mono font-semibold">{generatedCode}</span></>
                    )}
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
                  disabled={
                    !stripHtml(content).trim() || 
                    isGeneratingCode || 
                    !!generatedCode ||
                    (letterType === 'drift' && !driftConfig.theme) ||
                    (letterType === 'future' && (!futureConfig.deliveryDate || !futureConfig.deliveryTime))
                  }
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
                      {letterType === 'drift' ? '生成漂流瓶编号' : 
                       letterType === 'future' ? '生成未来信编号' : 
                       '生成编号贴纸'}
                    </>
                  )}
                </Button>
                {generatedCode && (
                  <Button 
                    onClick={() => {
                      if (letterType === 'drift') {
                        toast.success('漂流瓶已准备就绪！去投递吧')
                      } else if (letterType === 'future') {
                        toast.success('未来信已设定！去完成投递')
                      }
                      router.push('/letters/send')
                    }}
                    className="font-serif"
                  >
                    <Send className="mr-2 h-4 w-4" />
                    {letterType === 'drift' ? '投递漂流瓶' : 
                     letterType === 'future' ? '投递未来信' : 
                     '去投递'}
                  </Button>
                )}
              </div>
            </TabsContent>
            
            {/* 上传手写信标签页 */}
            <TabsContent value="upload" className="space-y-6">
              <HandwrittenUpload
                onImagesUploaded={(images) => {
                  setUploadedImages(images)
                  // TODO: 后续调用OCR服务
                }}
                onTextExtracted={(text) => {
                  setExtractedText(text)
                  // 可以将提取的文字填充到content中
                  setContent(text)
                }}
                maxImages={5}
                maxFileSize={10}
              />
              
              {/* 如果有提取的文字，显示编辑器 */}
              {extractedText && (
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <Edit3 className="h-5 w-5" />
                      识别结果编辑
                    </CardTitle>
                    <CardDescription>
                      请检查并编辑OCR识别的文字内容
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <RichTextEditor
                      content={extractedText}
                      onChange={setExtractedText}
                      placeholder="识别的文字将显示在这里..."
                      className="font-serif text-base leading-loose"
                      maxLength={2000}
                    />
                  </CardContent>
                </Card>
              )}
              
              {/* 操作按钮 */}
              <div className="flex flex-wrap gap-4">
                <Button 
                  onClick={handleSaveDraft}
                  variant="outline"
                  disabled={!uploadedImages.length && !extractedText}
                >
                  <Save className="mr-2 h-4 w-4" />
                  保存手写信
                </Button>
                <Button 
                  onClick={handleGenerateCode}
                  disabled={(!uploadedImages.length && !extractedText) || isGeneratingCode || !!generatedCode}
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
            </TabsContent>
          </Tabs>
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