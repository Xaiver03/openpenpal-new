'use client'

import { useState, useRef, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { usePermission } from '@/hooks/use-permission'
import { BackButton } from '@/components/ui/back-button'
import { SafeTimestamp } from '@/components/ui/safe-timestamp'
import { 
  Scan,
  Camera,
  Package,
  CheckCircle,
  AlertCircle,
  Truck,
  Upload,
  MapPin,
  Clock,
  User,
  QrCode,
  History,
  Phone,
  RefreshCw,
  Target,
  Image as ImageIcon,
  FlashlightIcon as Flashlight
} from 'lucide-react'

// 模拟信件信息
interface ScannedLetter {
  id: string
  code: string
  title: string
  content: string
  status: 'draft' | 'generated' | 'collected' | 'in_transit' | 'delivered' | 'failed'
  senderNickname: string
  senderPhone?: string
  recipientHint?: string
  createdAt: string
  deliveryLocation?: string
  targetLocation?: string
  lastUpdate?: string
  priority: 'normal' | 'urgent'
  deliveryInstructions?: string
}

interface ScanHistory {
  id: string
  code: string
  action: string
  timestamp: string
  location?: string
}

export default function CourierScanPage() {
  const { user, isCourier } = usePermission()
  const [isScanning, setIsScanning] = useState(false)
  const [scannedCode, setScannedCode] = useState('')
  const [letterInfo, setLetterInfo] = useState<ScannedLetter | null>(null)
  const [scanError, setScanError] = useState<string | null>(null)
  const [updateStatus, setUpdateStatus] = useState<'collected' | 'in_transit' | 'delivered' | 'failed'>('collected')
  const [location, setLocation] = useState('')
  const [note, setNote] = useState('')
  const [isUpdating, setIsUpdating] = useState(false)
  const [scanHistory, setScanHistory] = useState<ScanHistory[]>([])
  const [activeTab, setActiveTab] = useState<'scan' | 'history'>('scan')
  const [flashEnabled, setFlashEnabled] = useState(false)
  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const streamRef = useRef<MediaStream | null>(null)

  // Cleanup effect - must be before conditional returns
  useEffect(() => {
    return () => {
      stopScanning()
    }
  }, [])

  // 权限检查
  if (!isCourier()) {
    return (
      <div className="container max-w-4xl mx-auto px-4 py-8">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            只有信使才能使用扫码功能。如需申请成为信使，请前往信使中心。
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  // 获取用户位置
  const getCurrentLocation = (): Promise<string> => {
    return new Promise((resolve) => {
      if (!navigator.geolocation) {
        resolve('位置获取不可用')
        return
      }

      navigator.geolocation.getCurrentPosition(
        (position) => {
          const { latitude, longitude } = position.coords
          resolve(`${latitude.toFixed(6)}, ${longitude.toFixed(6)}`)
        },
        () => {
          resolve('位置获取失败')
        },
        { timeout: 10000 }
      )
    })
  }

  // 扫码功能
  const startScanning = async () => {
    setIsScanning(true)
    setScanError(null)
    
    try {
      const constraints = {
        video: { 
          facingMode: 'environment',
          width: { ideal: 1280 },
          height: { ideal: 720 }
        }
      }
      
      const stream = await navigator.mediaDevices.getUserMedia(constraints)
      streamRef.current = stream
      
      if (videoRef.current) {
        videoRef.current.srcObject = stream
        videoRef.current.play()
      }
    } catch (error) {
      setScanError('无法访问摄像头，请检查权限设置或尝试使用手动输入')
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

  const toggleFlash = async () => {
    if (streamRef.current) {
      const videoTrack = streamRef.current.getVideoTracks()[0]
      if (videoTrack && 'torch' in videoTrack.getCapabilities()) {
        try {
          await videoTrack.applyConstraints({
            advanced: [{ torch: !flashEnabled } as any]
          })
          setFlashEnabled(!flashEnabled)
        } catch (error) {
          console.log('Flash not supported')
        }
      }
    }
  }

  // 手动输入编号
  const handleManualInput = async () => {
    if (!scannedCode.trim()) return
    
    setScanError(null)
    
    // 模拟API调用获取信件信息
    const mockLetters: Record<string, ScannedLetter> = {
      'OP1K2L3M4N5O': {
        id: '1',
        code: 'OP1K2L3M4N5O',
        title: '给朋友的问候信',
        content: '最近怎么样？很想念和你一起度过的美好时光。希望这封信能够带给你温暖和快乐。',
        status: 'generated',
        senderNickname: '小明',
        senderPhone: '138****5678',
        recipientHint: '北大宿舍楼，李同学',
        createdAt: '2024-01-15T10:30:00Z',
        targetLocation: '北京大学宿舍楼32栋',
        priority: 'normal',
        deliveryInstructions: '请投递到宿舍管理员处'
      },
      'OP2K3L4M5N6P': {
        id: '2',
        code: 'OP2K3L4M5N6P',
        title: '紧急通知信件',
        content: '这是一封重要的通知信件，请尽快投递。',
        status: 'collected',
        senderNickname: '王老师',
        senderPhone: '139****1234',
        recipientHint: '计算机学院，张教授',
        createdAt: '2024-01-16T14:20:00Z',
        targetLocation: '计算机学院办公楼203室',
        priority: 'urgent',
        deliveryInstructions: '请直接交给本人，如不在请联系'
      }
    }
    
    // 模拟网络延迟
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    const foundLetter = mockLetters[scannedCode]
    if (foundLetter) {
      setLetterInfo(foundLetter)
      // 自动填入当前位置
      const currentLocation = await getCurrentLocation()
      setLocation(currentLocation)
    } else {
      setScanError('未找到对应的信件，请检查编号是否正确')
    }
  }

  // 更新信件状态
  const handleUpdateStatus = async () => {
    if (!letterInfo) return
    
    setIsUpdating(true)
    
    try {
      // 模拟API调用更新状态
      await new Promise(resolve => setTimeout(resolve, 1500))
      
      // 添加到历史记录
      const newHistoryItem: ScanHistory = {
        id: Date.now().toString(),
        code: letterInfo.code,
        action: getStatusInfo(updateStatus).label,
        timestamp: new Date().toISOString(),
        location: location || undefined
      }
      setScanHistory(prev => [newHistoryItem, ...prev])
      
      // 更新本地状态
      setLetterInfo(prev => prev ? {
        ...prev,
        status: updateStatus,
        lastUpdate: new Date().toISOString(),
        deliveryLocation: location || prev.deliveryLocation
      } : null)
      
      // 重置表单
      setLocation('')
      setNote('')
      setScannedCode('')
      
      // 显示成功消息
      setScanError(null)
      alert(`状态更新成功！信件 ${letterInfo.code} 已标记为${getStatusInfo(updateStatus).label}`)
    } catch (error) {
      setScanError('状态更新失败，请稍后重试')
    } finally {
      setIsUpdating(false)
    }
  }

  // 快速状态更新
  const quickStatusUpdate = async (status: ScannedLetter['status']) => {
    if (!letterInfo) return
    
    // 只允许信使可以设置的状态
    const allowedStatuses = ['collected', 'in_transit', 'delivered', 'failed'] as const
    if (allowedStatuses.includes(status as any)) {
      setUpdateStatus(status as 'collected' | 'in_transit' | 'delivered' | 'failed')
    }
    const currentLocation = await getCurrentLocation()
    setLocation(currentLocation)
    
    // 如果是已投递状态，自动提交
    if (status === 'delivered') {
      setTimeout(() => {
        handleUpdateStatus()
      }, 500)
    }
  }

  const getStatusInfo = (status: ScannedLetter['status']) => {
    switch (status) {
      case 'draft':
        return {
          label: '草稿',
          color: 'bg-gray-100 text-gray-800 border-gray-200',
          icon: Package
        }
      case 'generated':
        return {
          label: '已生成',
          color: 'bg-amber-100 text-amber-800 border-amber-200',
          icon: QrCode
        }
      case 'collected':
        return {
          label: '已收取',
          color: 'bg-blue-100 text-blue-800 border-blue-200',
          icon: Package
        }
      case 'in_transit':
        return {
          label: '投递中',
          color: 'bg-orange-100 text-orange-800 border-orange-200',
          icon: Truck
        }
      case 'delivered':
        return {
          label: '已投递',
          color: 'bg-green-100 text-green-800 border-green-200',
          icon: CheckCircle
        }
      case 'failed':
        return {
          label: '投递失败',
          color: 'bg-red-100 text-red-800 border-red-200',
          icon: AlertCircle
        }
    }
  }

  const getPriorityInfo = (priority: ScannedLetter['priority']) => {
    switch (priority) {
      case 'urgent':
        return {
          label: '紧急',
          color: 'bg-red-100 text-red-800 border-red-200'
        }
      case 'normal':
        return {
          label: '普通',
          color: 'bg-gray-100 text-gray-800 border-gray-200'
        }
    }
  }

  return (
    <div className="container max-w-6xl mx-auto px-4 py-4 md:py-8">
      {/* Header - Mobile Optimized */}
      <div className="mb-6 md:mb-8">
        <div className="flex items-center gap-2 md:gap-4 mb-3 md:mb-4">
          <BackButton href="/courier" className="md:hidden" />
          <div className="flex-1">
            <h1 className="font-serif text-xl md:text-3xl font-bold text-amber-900 mb-1 md:mb-2">
              信使扫码
            </h1>
            <p className="text-amber-700 text-sm md:text-base hidden md:block">
              扫描信件二维码，更新投递状态。欢迎您，{user?.nickname}！
            </p>
            <p className="text-amber-700 text-xs md:hidden">
              扫码更新投递状态
            </p>
          </div>
        </div>
      </div>

      {/* Tab Navigation - Mobile Optimized */}
      <div className="mb-4 md:mb-8">
        <div className="flex space-x-1 bg-amber-100 p-1 rounded-lg">
          <button
            onClick={() => setActiveTab('scan')}
            className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === 'scan'
                ? 'bg-white text-amber-900 shadow-sm'
                : 'text-amber-700 hover:text-amber-900'
            }`}
          >
            <Scan className="w-4 h-4" />
            扫码投递
          </button>
          <button
            onClick={() => setActiveTab('history')}
            className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === 'history'
                ? 'bg-white text-amber-900 shadow-sm'
                : 'text-amber-700 hover:text-amber-900'
            }`}
          >
            <History className="w-4 h-4" />
            扫码历史
            {scanHistory.length > 0 && (
              <Badge className="bg-amber-600 text-white text-xs">
                {scanHistory.length}
              </Badge>
            )}
          </button>
        </div>
      </div>

      {activeTab === 'scan' ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-8">
          {/* 扫码区域 */}
          <div className="space-y-6">
            {/* 摄像头扫描 */}
            <Card className="border-amber-200">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-amber-900">
                  <Camera className="h-5 w-5" />
                  摄像头扫描
                </CardTitle>
                <CardDescription>
                  使用摄像头扫描信件上的二维码
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {!isScanning ? (
                  <div className="text-center py-8">
                    <div className="w-24 h-24 mx-auto bg-amber-100 rounded-full flex items-center justify-center mb-4">
                      <QrCode className="h-12 w-12 text-amber-600" />
                    </div>
                    <Button 
                      onClick={startScanning} 
                      className="w-full bg-amber-600 hover:bg-amber-700 text-white"
                    >
                      <Scan className="mr-2 h-4 w-4" />
                      开始扫描
                    </Button>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div className="relative">
                      <video
                        ref={videoRef}
                        className="w-full aspect-square rounded-lg bg-black"
                        autoPlay
                        muted
                        playsInline
                      />
                      <div className="absolute inset-0 border-2 border-amber-500 rounded-lg pointer-events-none">
                        <div className="absolute top-4 left-4 w-6 h-6 border-l-2 border-t-2 border-amber-500"></div>
                        <div className="absolute top-4 right-4 w-6 h-6 border-r-2 border-t-2 border-amber-500"></div>
                        <div className="absolute bottom-4 left-4 w-6 h-6 border-l-2 border-b-2 border-amber-500"></div>
                        <div className="absolute bottom-4 right-4 w-6 h-6 border-r-2 border-b-2 border-amber-500"></div>
                      </div>
                      <div className="absolute top-4 right-4 p-2">
                        <Button
                          onClick={toggleFlash}
                          size="sm"
                          variant="secondary"
                          className={`${flashEnabled ? 'bg-yellow-400 text-black' : 'bg-gray-800 text-white'}`}
                        >
                          <Flashlight className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button onClick={stopScanning} variant="outline" className="flex-1 border-amber-300 text-amber-700">
                        停止扫描
                      </Button>
                      <Button 
                        onClick={toggleFlash}
                        variant="outline" 
                        size="sm"
                        className={`border-amber-300 ${flashEnabled ? 'bg-yellow-100 text-yellow-800' : 'text-amber-700'}`}
                      >
                        <Flashlight className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* 手动输入 */}
            <Card className="border-amber-200">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-amber-900">
                  <Upload className="h-5 w-5" />
                  手动输入编号
                </CardTitle>
                <CardDescription>
                  直接输入信件编号进行查询
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <Input
                  placeholder="请输入信件编号，如：OP1K2L3M4N5O"
                  value={scannedCode}
                  onChange={(e) => setScannedCode(e.target.value.toUpperCase())}
                  className="font-mono border-amber-300 focus:border-amber-500"
                />
                <div className="flex gap-2">
                  <Button 
                    onClick={handleManualInput} 
                    disabled={!scannedCode.trim()}
                    className="flex-1 bg-amber-600 hover:bg-amber-700 text-white"
                  >
                    查询信件
                  </Button>
                  <Button
                    onClick={() => setScannedCode('')}
                    variant="outline"
                    size="sm"
                    className="border-amber-300 text-amber-700"
                  >
                    <RefreshCw className="h-4 w-4" />
                  </Button>
                </div>
                
                {/* 快速测试按钮 */}
                <div className="pt-2 border-t border-amber-200">
                  <p className="text-sm text-amber-600 mb-2">测试编号：</p>
                  <div className="flex gap-2">
                    <Button
                      onClick={() => setScannedCode('OP1K2L3M4N5O')}
                      variant="outline"
                      size="sm"
                      className="text-xs border-amber-300 text-amber-700"
                    >
                      普通信件
                    </Button>
                    <Button
                      onClick={() => setScannedCode('OP2K3L4M5N6P')}
                      variant="outline"
                      size="sm"
                      className="text-xs border-amber-300 text-amber-700"
                    >
                      紧急信件
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>

          {/* 错误信息 */}
          {scanError && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{scanError}</AlertDescription>
            </Alert>
          )}
        </div>

        {/* 信件信息和状态更新 */}
        <div className="space-y-6">
          {letterInfo ? (
            <>
              {/* 信件信息 */}
              <Card className="border-amber-200">
                <CardHeader>
                  <CardTitle className="flex items-center justify-between text-amber-900">
                    <span>信件信息</span>
                    <div className="flex gap-2">
                      <Badge className={getPriorityInfo(letterInfo.priority).color}>
                        {getPriorityInfo(letterInfo.priority).label}
                      </Badge>
                      <Badge className={getStatusInfo(letterInfo.status).color}>
                        {getStatusInfo(letterInfo.status).label}
                      </Badge>
                    </div>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-3">
                    <div className="flex items-center gap-2 text-sm">
                      <QrCode className="h-4 w-4 text-amber-600" />
                      <span className="font-mono font-semibold">{letterInfo.code}</span>
                    </div>
                    <div className="flex items-center gap-2 text-sm">
                      <User className="h-4 w-4 text-amber-600" />
                      <span>发件人：{letterInfo.senderNickname}</span>
                      {letterInfo.senderPhone && (
                        <>
                          <Phone className="h-3 w-3 text-amber-500 ml-2" />
                          <span className="text-amber-600">{letterInfo.senderPhone}</span>
                        </>
                      )}
                    </div>
                    {letterInfo.recipientHint && (
                      <div className="flex items-center gap-2 text-sm">
                        <Target className="h-4 w-4 text-amber-600" />
                        <span>收件人：{letterInfo.recipientHint}</span>
                      </div>
                    )}
                    <div className="flex items-center gap-2 text-sm">
                      <Clock className="h-4 w-4 text-amber-600" />
                      <span>创建：</span>
                      <SafeTimestamp 
                        date={letterInfo.createdAt} 
                        format="locale" 
                        fallback="--"
                        className="inline"
                      />
                    </div>
                    {letterInfo.targetLocation && (
                      <div className="flex items-center gap-2 text-sm">
                        <MapPin className="h-4 w-4 text-amber-600" />
                        <span>目标位置：{letterInfo.targetLocation}</span>
                      </div>
                    )}
                    {letterInfo.deliveryLocation && (
                      <div className="flex items-center gap-2 text-sm">
                        <MapPin className="h-4 w-4 text-green-600" />
                        <span>当前位置：{letterInfo.deliveryLocation}</span>
                      </div>
                    )}
                  </div>
                  
                  <div className="pt-3 border-t border-amber-200">
                    <h4 className="font-medium mb-2 text-amber-900">{letterInfo.title}</h4>
                    <p className="text-sm text-amber-700 line-clamp-3">
                      {letterInfo.content}
                    </p>
                  </div>

                  {letterInfo.deliveryInstructions && (
                    <div className="pt-3 border-t border-amber-200">
                      <h5 className="text-sm font-medium text-amber-900 mb-1">投递说明：</h5>
                      <p className="text-sm text-amber-700">{letterInfo.deliveryInstructions}</p>
                    </div>
                  )}

                  {/* 快速操作 */}
                  <div className="pt-3 border-t border-amber-200">
                    <p className="text-sm font-medium text-amber-900 mb-2">快速操作：</p>
                    <div className="flex gap-2">
                      <Button
                        onClick={() => quickStatusUpdate('collected')}
                        size="sm"
                        variant="outline"
                        className="border-blue-300 text-blue-700 hover:bg-blue-50"
                      >
                        <Package className="h-3 w-3 mr-1" />
                        已收取
                      </Button>
                      <Button
                        onClick={() => quickStatusUpdate('in_transit')}
                        size="sm"
                        variant="outline"
                        className="border-orange-300 text-orange-700 hover:bg-orange-50"
                      >
                        <Truck className="h-3 w-3 mr-1" />
                        投递中
                      </Button>
                      <Button
                        onClick={() => quickStatusUpdate('delivered')}
                        size="sm"
                        variant="outline"
                        className="border-green-300 text-green-700 hover:bg-green-50"
                      >
                        <CheckCircle className="h-3 w-3 mr-1" />
                        已投递
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* 状态更新 */}
              <Card className="border-amber-200">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-amber-900">
                    <Truck className="h-5 w-5" />
                    更新状态
                  </CardTitle>
                  <CardDescription>
                    更新信件的投递状态
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <label className="text-sm font-medium mb-2 block text-amber-900">新状态</label>
                    <select
                      value={updateStatus}
                      onChange={(e) => setUpdateStatus(e.target.value as any)}
                      className="w-full p-2 border border-amber-300 rounded-md bg-white focus:border-amber-500 focus:outline-none"
                    >
                      <option value="collected">已收取</option>
                      <option value="in_transit">投递中</option>
                      <option value="delivered">已投递</option>
                      <option value="failed">投递失败</option>
                    </select>
                  </div>
                  
                  <div>
                    <label className="text-sm font-medium mb-2 block text-amber-900">位置信息</label>
                    <Input
                      placeholder="如：北京大学宿舍楼下信箱"
                      value={location}
                      onChange={(e) => setLocation(e.target.value)}
                      className="border-amber-300 focus:border-amber-500"
                    />
                    <Button
                      onClick={async () => {
                        const loc = await getCurrentLocation()
                        setLocation(loc)
                      }}
                      variant="outline"
                      size="sm"
                      className="mt-2 border-amber-300 text-amber-700"
                    >
                      <MapPin className="h-3 w-3 mr-1" />
                      获取当前位置
                    </Button>
                  </div>
                  
                  <div>
                    <label className="text-sm font-medium mb-2 block text-amber-900">备注信息</label>
                    <Textarea
                      placeholder="投递备注信息..."
                      value={note}
                      onChange={(e) => setNote(e.target.value)}
                      rows={3}
                      className="border-amber-300 focus:border-amber-500"
                    />
                  </div>
                  
                  <Button 
                    onClick={handleUpdateStatus}
                    disabled={isUpdating}
                    className="w-full bg-amber-600 hover:bg-amber-700 text-white"
                  >
                    {isUpdating ? (
                      <>
                        <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                        更新中...
                      </>
                    ) : (
                      <>
                        <CheckCircle className="mr-2 h-4 w-4" />
                        确认更新状态
                      </>
                    )}
                  </Button>
                </CardContent>
              </Card>
            </>
            ) : (
              <Card className="text-center py-12 border-amber-200">
                <CardContent>
                  <div className="w-16 h-16 mx-auto bg-amber-100 rounded-full flex items-center justify-center mb-4">
                    <Package className="h-8 w-8 text-amber-600" />
                  </div>
                  <h3 className="text-lg font-semibold mb-2 text-amber-900">等待扫描</h3>
                  <p className="text-amber-700">
                    请扫描二维码或手动输入编号查看信件信息
                  </p>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      ) : (
        /* 历史记录标签页 */
        <div className="space-y-6">
          <Card className="border-amber-200">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-amber-900">
                <History className="h-5 w-5" />
                扫码历史记录
              </CardTitle>
              <CardDescription>
                查看您的扫码和状态更新历史
              </CardDescription>
            </CardHeader>
            <CardContent>
              {scanHistory.length > 0 ? (
                <div className="space-y-3">
                  {scanHistory.map((item, index) => (
                    <div 
                      key={item.id}
                      className="flex items-center justify-between p-3 bg-amber-50 rounded-lg border border-amber-200"
                    >
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 bg-amber-200 rounded-full flex items-center justify-center text-xs font-mono">
                          {index + 1}
                        </div>
                        <div>
                          <p className="font-mono text-sm font-medium text-amber-900">
                            {item.code}
                          </p>
                          <p className="text-xs text-amber-600">
                            {item.action}
                          </p>
                        </div>
                      </div>
                      <div className="text-right">
                        <SafeTimestamp 
                          date={item.timestamp} 
                          format="locale" 
                          fallback="--"
                          className="text-xs text-amber-600"
                        />
                        {item.location && (
                          <p className="text-xs text-amber-500 mt-1">
                            📍 {item.location}
                          </p>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8">
                  <div className="w-16 h-16 mx-auto bg-amber-100 rounded-full flex items-center justify-center mb-4">
                    <History className="h-8 w-8 text-amber-600" />
                  </div>
                  <h3 className="text-lg font-semibold mb-2 text-amber-900">暂无历史记录</h3>
                  <p className="text-amber-700">
                    开始扫码投递信件后，历史记录会显示在这里
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      <canvas ref={canvasRef} className="hidden" />

      {/* 使用说明 */}
      <Card className="mt-8 border-amber-200">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-amber-900">
            <AlertCircle className="h-5 w-5" />
            使用说明
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm text-amber-700">
          <div className="flex items-start gap-2">
            <span className="text-amber-600 font-semibold">1.</span>
            <span>使用摄像头扫描信件上的二维码，或手动输入信件编号</span>
          </div>
          <div className="flex items-start gap-2">
            <span className="text-amber-600 font-semibold">2.</span>
            <span>确认信件信息无误后，选择要更新的状态</span>
          </div>
          <div className="flex items-start gap-2">
            <span className="text-amber-600 font-semibold">3.</span>
            <span>填写位置信息和备注（可选），点击更新状态</span>
          </div>
          <div className="flex items-start gap-2">
            <span className="text-amber-600 font-semibold">4.</span>
            <span>状态更新成功后，收发双方都能看到最新的投递进度</span>
          </div>
          <div className="flex items-start gap-2">
            <span className="text-amber-600 font-semibold">5.</span>
            <span>可以使用快速操作按钮或详细表单来更新状态</span>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}