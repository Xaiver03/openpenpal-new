'use client'

import { useState, useRef } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Progress } from '@/components/ui/progress'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { 
  QrCode, 
  Heart, 
  Target, 
  Camera, 
  Type, 
  Scan,
  CheckCircle, 
  AlertCircle, 
  Wand2,
  Send,
  Users,
  Sparkles,
  Clock
} from 'lucide-react'
import { BackButton } from '@/components/ui/back-button'
import { 
  barcodeBindingAPI, 
  BindingType, 
  BarcodeBindingRequest,
  AIMatchRequest,
  AIMatchResult,
  BarcodeValidation
} from '@/lib/api/barcode-binding'

// AI匹配状态
type MatchStatus = 'idle' | 'analyzing' | 'matching' | 'completed' | 'failed'

// 绑定表单数据
interface BindingFormData {
  barcode: string
  type: BindingType
  // 定向信字段
  recipientName?: string
  recipientOPCode?: string
  // 漂流信字段
  letterContent?: string
  writerAge?: number
  writerGender?: 'male' | 'female' | 'other'
  interests?: string[]
  mood?: string
  themes?: string[]
  matchDelay?: number // 用户可选择的延迟时间（分钟）
}

export default function BindBarcodePage() {
  const [activeTab, setActiveTab] = useState<'scan' | 'manual'>('scan')
  const [formData, setFormData] = useState<BindingFormData>({
    barcode: '',
    type: 'directed',
    matchDelay: 30 // 默认30分钟延迟
  })
  const [matchStatus, setMatchStatus] = useState<MatchStatus>('idle')
  const [matchResult, setMatchResult] = useState<AIMatchResult | null>(null)
  const [barcodeValidation, setBarcodeValidation] = useState<BarcodeValidation | null>(null)
  const [isScanning, setIsScanning] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isBinding, setIsBinding] = useState(false)
  const [isValidating, setIsValidating] = useState(false)
  const [progress, setProgress] = useState(0)
  
  const videoRef = useRef<HTMLVideoElement>(null)
  const streamRef = useRef<MediaStream | null>(null)

  // 可选的兴趣标签
  const availableInterests = [
    '音乐', '电影', '阅读', '旅行', '运动', '美食', 
    '摄影', '绘画', '写作', '科技', '游戏', '动漫'
  ]

  // 可选的心情标签
  const availableMoods = [
    '开心', '思念', '孤独', '兴奋', '平静', '焦虑', 
    '感恩', '怀念', '期待', '困惑', '温暖', '伤感'
  ]

  // 可选的主题标签
  const availableThemes = [
    '友情', '爱情', '家庭', '成长', '梦想', '困惑',
    '感谢', '道歉', '鼓励', '分享', '回忆', '未来'
  ]

  // 开始扫码
  const startScanning = async () => {
    setIsScanning(true)
    setError(null)
    
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { 
          facingMode: 'environment',
          width: { ideal: 1280 },
          height: { ideal: 720 }
        }
      })
      streamRef.current = stream
      
      if (videoRef.current) {
        videoRef.current.srcObject = stream
        videoRef.current.play()
      }
    } catch (error) {
      setError('无法访问摄像头，请检查权限或使用手动输入')
      setIsScanning(false)
    }
  }

  const stopScanning = () => {
    setIsScanning(false)
    if (streamRef.current) {
      streamRef.current.getTracks().forEach(track => track.stop())
      streamRef.current = null
    }
  }

  // 处理条码输入
  const handleBarcodeDetected = async (code: string) => {
    setFormData(prev => ({ ...prev, barcode: code }))
    stopScanning()
    setActiveTab('manual') // 切换到配置页面
    
    // 验证条码
    await validateBarcode(code)
  }

  // 验证条码
  const validateBarcode = async (code: string) => {
    if (!code.trim()) return

    setIsValidating(true)
    setError(null)

    try {
      const response = await barcodeBindingAPI.validateBarcode(code)
      
      if (response.success && response.data) {
        setBarcodeValidation(response.data)
        
        if (!response.data.valid) {
          setError(response.data.error_message || '条码无效或已被使用')
        }
      } else {
        setError('条码验证失败，请检查编号是否正确')
      }
    } catch (error) {
      console.error('条码验证失败:', error)
      setError('条码验证失败，请检查网络连接')
    } finally {
      setIsValidating(false)
    }
  }

  // 模拟扫码检测（实际应用中需要集成QR码检测库）
  const simulateScan = () => {
    const mockCodes = ['OP7X1F2K', 'OP8Y3M5N', 'OP9Z4P6Q']
    const randomCode = mockCodes[Math.floor(Math.random() * mockCodes.length)]
    handleBarcodeDetected(randomCode)
  }

  // AI匹配处理
  const handleAIMatching = async () => {
    if (!formData.letterContent?.trim()) {
      setError('请输入信件内容以便AI匹配')
      return
    }

    setMatchStatus('analyzing')
    setProgress(10)
    setError(null)

    try {
      // 构建AI匹配请求
      const matchRequest: AIMatchRequest = {
        letter_content: formData.letterContent,
        writer_profile: {
          age: formData.writerAge,
          gender: formData.writerGender,
          interests: formData.interests || [],
          mood: formData.mood,
          themes: formData.themes || []
        }
      }

      setMatchStatus('matching')
      setProgress(50)

      // 调用AI匹配API
      const response = await barcodeBindingAPI.matchRecipient(matchRequest)
      
      setProgress(90)

      if (response.success && response.data) {
        setProgress(100)
        setMatchResult(response.data)
        setMatchStatus('completed')
        
        if (response.data?.matched && response.data?.recipient_op_code) {
          setFormData(prev => ({
            ...prev,
            recipientOPCode: response.data?.recipient_op_code || ''
          }))
        } else {
          setMatchStatus('failed')
          setError('很抱歉，暂时没有找到合适的匹配对象，请稍后重试')
        }
      } else {
        setMatchStatus('failed')
        setError(response.message || 'AI匹配失败，请重试')
        setProgress(0)
      }
    } catch (error) {
      console.error('AI匹配失败:', error)
      setMatchStatus('failed')
      setError('AI匹配过程中出现错误，请检查网络连接后重试')
      setProgress(0)
    }
  }

  // 提交绑定
  const handleSubmit = async () => {
    setIsBinding(true)
    setError(null)

    try {
      // 验证表单
      if (!formData.barcode.trim()) {
        throw new Error('请输入或扫描条码')
      }

      // 验证条码状态
      if (!barcodeValidation?.valid) {
        throw new Error('条码无效或已被使用，无法绑定')
      }

      if (formData.type === 'directed') {
        if (!formData.recipientOPCode?.trim()) {
          throw new Error('请输入收件人OP Code')
        }
        if (!formData.recipientName?.trim()) {
          throw new Error('请输入收件人姓名')
        }
      } else if (formData.type === 'drift') {
        if (matchStatus !== 'completed' || !matchResult?.matched) {
          throw new Error('请先完成AI匹配')
        }
      }

      // 构建绑定请求
      const bindingRequest: BarcodeBindingRequest = {
        barcode: formData.barcode,
        type: formData.type,
        recipient_name: formData.recipientName,
        recipient_op_code: formData.recipientOPCode,
        letter_content: formData.letterContent,
        writer_profile: formData.type === 'drift' ? {
          age: formData.writerAge,
          gender: formData.writerGender,
          interests: formData.interests,
          mood: formData.mood,
          themes: formData.themes
        } : undefined,
        match_delay: formData.matchDelay
      }

      // 调用绑定API
      const response = await barcodeBindingAPI.bindBarcode(bindingRequest)

      if (response.success && response.data) {
        // 成功提示
        alert(`条码绑定成功！\n` +
              `条码：${response.data.barcode}\n` +
              `类型：${response.data.type === 'directed' ? '定向信' : '漂流信'}\n` +
              `收件人：${response.data.recipient_op_code}\n` +
              `${response.data.estimated_delivery ? `预计送达：${response.data.estimated_delivery}` : ''}`)
        
        // 重置表单
        setFormData({
          barcode: '',
          type: 'directed',
          matchDelay: 30
        })
        setBarcodeValidation(null)
        setMatchStatus('idle')
        setMatchResult(null)
        setProgress(0)
        
      } else {
        throw new Error(response.message || '绑定失败，请重试')
      }
      
    } catch (error: any) {
      console.error('绑定失败:', error)
      setError(error.message || '绑定失败，请检查网络连接后重试')
    } finally {
      setIsBinding(false)
    }
  }

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-4xl mx-auto px-4 py-8">
        <div className="flex items-center gap-4 mb-6">
          <BackButton />
          <div>
            <h1 className="text-3xl font-bold text-amber-900">条码绑定</h1>
            <p className="text-amber-700">将条码绑定到信件，选择投递方式</p>
          </div>
        </div>

        <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as any)}>
          <TabsList className="grid w-full grid-cols-2 mb-6">
            <TabsTrigger value="scan" className="data-[state=active]:bg-amber-600 data-[state=active]:text-white">
              <Camera className="w-4 h-4 mr-2" />
              扫码绑定
            </TabsTrigger>
            <TabsTrigger value="manual" className="data-[state=active]:bg-amber-600 data-[state=active]:text-white">
              <Type className="w-4 h-4 mr-2" />
              手动输入
            </TabsTrigger>
          </TabsList>

          {/* 扫码模式 */}
          <TabsContent value="scan">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <QrCode className="w-5 h-5" />
                  扫描条码
                </CardTitle>
                <CardDescription>
                  将摄像头对准条码进行扫描
                </CardDescription>
              </CardHeader>
              <CardContent>
                {!isScanning ? (
                  <div className="text-center py-8">
                    <div className="w-24 h-24 bg-amber-200 rounded-full flex items-center justify-center mx-auto mb-4">
                      <QrCode className="w-12 h-12 text-amber-700" />
                    </div>
                    <Button
                      onClick={startScanning}
                      className="bg-amber-600 hover:bg-amber-700 text-white"
                    >
                      <Camera className="w-4 h-4 mr-2" />
                      开始扫描
                    </Button>
                    <div className="mt-4">
                      <Button
                        variant="outline"
                        onClick={simulateScan}
                        className="border-amber-300 text-amber-700 hover:bg-amber-50"
                      >
                        <Scan className="w-4 h-4 mr-2" />
                        模拟扫描（演示）
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div className="relative bg-black rounded-lg overflow-hidden">
                      <video
                        ref={videoRef}
                        className="w-full h-64 object-cover"
                        autoPlay
                        playsInline
                        muted
                      />
                      <div className="absolute inset-0 flex items-center justify-center">
                        <div className="w-48 h-48 border-2 border-white rounded-lg"></div>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        onClick={stopScanning}
                        variant="outline"
                        className="flex-1"
                      >
                        停止扫描
                      </Button>
                      <Button
                        onClick={simulateScan}
                        className="bg-amber-600 hover:bg-amber-700"
                      >
                        模拟检测
                      </Button>
                    </div>
                  </div>
                )}

                {error && (
                  <Alert variant="destructive" className="mt-4">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          {/* 手动输入模式 */}
          <TabsContent value="manual">
            <div className="space-y-6">
              {/* 条码输入 */}
              <Card>
                <CardHeader>
                  <CardTitle>条码信息</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label htmlFor="barcode">条码编号</Label>
                    <div className="relative">
                      <Input
                        id="barcode"
                        value={formData.barcode}
                        onChange={(e) => {
                          const value = e.target.value.toUpperCase()
                          setFormData(prev => ({ ...prev, barcode: value }))
                          if (value.length === 8) {
                            validateBarcode(value)
                          } else {
                            setBarcodeValidation(null)
                            setError(null)
                          }
                        }}
                        placeholder="输入8位条码编号，如：OP7X1F2K"
                        maxLength={8}
                        className={
                          barcodeValidation 
                            ? barcodeValidation.valid 
                              ? 'border-green-500 bg-green-50' 
                              : 'border-red-500 bg-red-50'
                            : ''
                        }
                      />
                      {isValidating && (
                        <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                          <Clock className="w-4 h-4 animate-spin text-amber-600" />
                        </div>
                      )}
                      {barcodeValidation && (
                        <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
                          {barcodeValidation.valid ? (
                            <CheckCircle className="w-4 h-4 text-green-600" />
                          ) : (
                            <AlertCircle className="w-4 h-4 text-red-600" />
                          )}
                        </div>
                      )}
                    </div>
                    
                    {barcodeValidation && (
                      <div className="mt-2">
                        {barcodeValidation.valid ? (
                          <Alert>
                            <CheckCircle className="h-4 w-4" />
                            <AlertDescription className="text-green-800">
                              条码验证成功！状态：{barcodeValidation.barcode_info?.status === 'unactivated' ? '未绑定' : '已绑定'}
                              {barcodeValidation.barcode_info?.created_at && (
                                <span className="block text-xs mt-1">
                                  创建时间：{new Date(barcodeValidation.barcode_info.created_at).toLocaleString()}
                                </span>
                              )}
                            </AlertDescription>
                          </Alert>
                        ) : (
                          <Alert variant="destructive">
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>
                              {barcodeValidation.error_message || '条码无效或已被使用'}
                            </AlertDescription>
                          </Alert>
                        )}
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>

              {/* 信件类型选择 */}
              <Card>
                <CardHeader>
                  <CardTitle>选择信件类型</CardTitle>
                  <CardDescription>
                    选择这封信的投递方式
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <RadioGroup 
                    value={formData.type} 
                    onValueChange={(value: BindingType) => {
                      setFormData(prev => ({ ...prev, type: value }))
                      setMatchStatus('idle')
                      setMatchResult(null)
                      setError(null)
                    }}
                    className="space-y-4"
                  >
                    <div className="flex items-start space-x-3 p-4 border border-amber-200 rounded-lg hover:bg-amber-50 transition-colors">
                      <RadioGroupItem value="directed" id="directed" className="mt-1" />
                      <div className="flex-1">
                        <Label htmlFor="directed" className="flex items-center gap-2 text-base font-semibold cursor-pointer">
                          <Target className="w-5 h-5 text-amber-600" />
                          定向信
                        </Label>
                        <p className="text-sm text-amber-700 mt-1">
                          发送给指定的收件人，需要输入对方的OP Code地址
                        </p>
                      </div>
                    </div>

                    <div className="flex items-start space-x-3 p-4 border border-amber-200 rounded-lg hover:bg-amber-50 transition-colors">
                      <RadioGroupItem value="drift" id="drift" className="mt-1" />
                      <div className="flex-1">
                        <Label htmlFor="drift" className="flex items-center gap-2 text-base font-semibold cursor-pointer">
                          <Heart className="w-5 h-5 text-red-500" />
                          漂流信
                        </Label>
                        <p className="text-sm text-amber-700 mt-1">
                          让AI为你匹配一个陌生的朋友，开启温暖的相遇
                        </p>
                      </div>
                    </div>
                  </RadioGroup>
                </CardContent>
              </Card>

              {/* 定向信配置 */}
              {formData.type === 'directed' && (
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <Target className="w-5 h-5" />
                      定向信配置
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div>
                      <Label htmlFor="recipientName">收件人姓名</Label>
                      <Input
                        id="recipientName"
                        value={formData.recipientName || ''}
                        onChange={(e) => setFormData(prev => ({ ...prev, recipientName: e.target.value }))}
                        placeholder="收件人的姓名或昵称"
                      />
                    </div>

                    <div>
                      <Label htmlFor="recipientOPCode">收件人OP Code</Label>
                      <Input
                        id="recipientOPCode"
                        value={formData.recipientOPCode || ''}
                        onChange={(e) => setFormData(prev => ({ ...prev, recipientOPCode: e.target.value.toUpperCase() }))}
                        placeholder="6位OP Code地址，如：PK5F3D"
                        maxLength={6}
                      />
                    </div>

                    <Alert>
                      <Target className="h-4 w-4" />
                      <AlertDescription>
                        OP Code是收件人的6位地址编码，格式为{'"学校代码+区域代码+位置代码"'}，
                        如PK5F3D代表北大5号楼303室。
                      </AlertDescription>
                    </Alert>
                  </CardContent>
                </Card>
              )}

              {/* 漂流信配置 */}
              {formData.type === 'drift' && (
                <div className="space-y-6">
                  <Card>
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2">
                        <Heart className="w-5 h-5 text-red-500" />
                        漂流信配置
                      </CardTitle>
                      <CardDescription>
                        AI将基于以下信息为你匹配最合适的收件人
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div>
                        <Label htmlFor="letterContent">信件内容预览</Label>
                        <Textarea
                          id="letterContent"
                          value={formData.letterContent || ''}
                          onChange={(e) => setFormData(prev => ({ ...prev, letterContent: e.target.value }))}
                          placeholder="请输入信件的主要内容，AI将基于内容进行匹配..."
                          rows={4}
                        />
                      </div>

                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <Label htmlFor="writerAge">你的年龄</Label>
                          <Input
                            id="writerAge"
                            type="number"
                            value={formData.writerAge || ''}
                            onChange={(e) => setFormData(prev => ({ ...prev, writerAge: parseInt(e.target.value) }))}
                            placeholder="如：20"
                            min="10"
                            max="100"
                          />
                        </div>

                        <div>
                          <Label htmlFor="writerGender">性别</Label>
                          <Select value={formData.writerGender} onValueChange={(value: any) => setFormData(prev => ({ ...prev, writerGender: value }))}>
                            <SelectTrigger>
                              <SelectValue placeholder="选择性别" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="male">男</SelectItem>
                              <SelectItem value="female">女</SelectItem>
                              <SelectItem value="other">其他</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </div>

                      <div>
                        <Label>兴趣爱好</Label>
                        <div className="flex flex-wrap gap-2 mt-2">
                          {availableInterests.map((interest) => (
                            <Button
                              key={interest}
                              variant={formData.interests?.includes(interest) ? "default" : "outline"}
                              size="sm"
                              onClick={() => {
                                const newInterests = formData.interests?.includes(interest)
                                  ? formData.interests.filter(i => i !== interest)
                                  : [...(formData.interests || []), interest]
                                setFormData(prev => ({ ...prev, interests: newInterests }))
                              }}
                              className={formData.interests?.includes(interest) 
                                ? "bg-amber-600 hover:bg-amber-700" 
                                : "border-amber-300 text-amber-700 hover:bg-amber-50"
                              }
                            >
                              {interest}
                            </Button>
                          ))}
                        </div>
                      </div>

                      <div>
                        <Label>当前心情</Label>
                        <Select value={formData.mood} onValueChange={(value) => setFormData(prev => ({ ...prev, mood: value }))}>
                          <SelectTrigger>
                            <SelectValue placeholder="选择当前的心情" />
                          </SelectTrigger>
                          <SelectContent>
                            {availableMoods.map((mood) => (
                              <SelectItem key={mood} value={mood}>{mood}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>

                      <div>
                        <Label>漂流延迟</Label>
                        <Select 
                          value={formData.matchDelay?.toString()} 
                          onValueChange={(value) => setFormData(prev => ({ ...prev, matchDelay: parseInt(value) }))}
                        >
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="5">5分钟 - 快速匹配</SelectItem>
                            <SelectItem value="30">30分钟 - 享受期待</SelectItem>
                            <SelectItem value="60">1小时 - 慢慢漂流</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </CardContent>
                  </Card>

                  {/* AI匹配执行 */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2">
                        <Wand2 className="w-5 h-5 text-purple-600" />
                        AI智能匹配
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      {matchStatus === 'idle' && (
                        <div className="text-center py-6">
                          <div className="w-16 h-16 bg-purple-200 rounded-full flex items-center justify-center mx-auto mb-4">
                            <Users className="w-8 h-8 text-purple-600" />
                          </div>
                          <Button
                            onClick={handleAIMatching}
                            disabled={!formData.letterContent?.trim()}
                            className="bg-purple-600 hover:bg-purple-700 text-white"
                          >
                            <Sparkles className="w-4 h-4 mr-2" />
                            开始AI匹配
                          </Button>
                          <p className="text-sm text-amber-600 mt-2">
                            请先填写信件内容以便AI分析
                          </p>
                        </div>
                      )}

                      {(matchStatus === 'analyzing' || matchStatus === 'matching') && (
                        <div className="space-y-4">
                          <div className="text-center">
                            <div className="w-16 h-16 bg-purple-200 rounded-full flex items-center justify-center mx-auto mb-4">
                              <Clock className="w-8 h-8 text-purple-600 animate-spin" />
                            </div>
                            <h3 className="font-semibold text-amber-900 mb-2">
                              {matchStatus === 'analyzing' ? '分析信件内容中...' : '匹配收件人中...'}
                            </h3>
                            <Progress value={progress} className="w-full mb-2" />
                            <p className="text-sm text-amber-600">
                              {matchStatus === 'analyzing' 
                                ? 'AI正在分析你的文字情感和主题' 
                                : `正在寻找最契合的笔友，预计等待 ${formData.matchDelay} 分钟`
                              }
                            </p>
                          </div>
                        </div>
                      )}

                      {matchStatus === 'completed' && matchResult && (
                        <div className="space-y-4">
                          <Alert>
                            <CheckCircle className="h-4 w-4" />
                            <AlertDescription className="text-green-800">
                              <strong>匹配成功！</strong> AI为你找到了一位合适的收件人
                            </AlertDescription>
                          </Alert>

                          <div className="bg-gradient-to-r from-purple-50 to-pink-50 rounded-lg p-4">
                            <div className="grid grid-cols-2 gap-4 text-sm">
                              <div>
                                <Label>收件人编码</Label>
                                <p className="font-mono text-purple-900">{matchResult.recipient_op_code}</p>
                              </div>
                              <div>
                                <Label>匹配度</Label>
                                <div className="flex items-center gap-2">
                                  <Progress value={matchResult.recipient_profile?.compatibility_score} className="flex-1" />
                                  <span className="text-purple-900 font-semibold">
                                    {matchResult.recipient_profile?.compatibility_score}%
                                  </span>
                                </div>
                              </div>
                              <div className="col-span-2">
                                <Label>匹配原因</Label>
                                <p className="text-purple-700">{matchResult.recipient_profile?.match_reason}</p>
                              </div>
                              {matchResult.estimated_delivery && (
                                <div className="col-span-2">
                                  <Label>预计送达</Label>
                                  <p className="text-purple-700">{matchResult.estimated_delivery}</p>
                                </div>
                              )}
                            </div>
                          </div>
                        </div>
                      )}

                      {matchStatus === 'failed' && (
                        <Alert variant="destructive">
                          <AlertCircle className="h-4 w-4" />
                          <AlertDescription>
                            匹配失败，请调整信件内容或稍后重试
                          </AlertDescription>
                        </Alert>
                      )}
                    </CardContent>
                  </Card>
                </div>
              )}

              {/* 提交按钮 */}
              <div className="flex justify-end">
                <Button
                  onClick={handleSubmit}
                  disabled={isBinding || !formData.barcode || 
                    (formData.type === 'directed' && (!formData.recipientName || !formData.recipientOPCode)) ||
                    (formData.type === 'drift' && matchStatus !== 'completed')
                  }
                  className="bg-amber-600 hover:bg-amber-700 text-white"
                >
                  {isBinding ? (
                    <>
                      <Clock className="w-4 h-4 mr-2 animate-spin" />
                      绑定中...
                    </>
                  ) : (
                    <>
                      <Send className="w-4 h-4 mr-2" />
                      完成绑定
                    </>
                  )}
                </Button>
              </div>

              {error && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}