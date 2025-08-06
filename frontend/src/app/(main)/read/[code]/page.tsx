'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { 
  Mail,
  User,
  Calendar,
  MapPin,
  Heart,
  Reply,
  Share,
  CheckCircle,
  Clock,
  Package,
  AlertCircle
} from 'lucide-react'
import { formatRelativeTime } from '@/lib/utils'
import { LetterService, type Letter } from '@/lib/services/letter-service'

interface LetterData extends Letter {
  deliveryNote?: string
  isRead: boolean
  is_sender?: boolean
  can_reply?: boolean
  delivery_info?: {
    courier_name?: string
    delivery_time?: string
    delivery_location?: string
  }
}

export default function ReadLetterPage() {
  const params = useParams()
  const router = useRouter()
  const code = params?.code as string
  
  const [letter, setLetter] = useState<LetterData | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isMarkingAsRead, setIsMarkingAsRead] = useState(false)

  useEffect(() => {
    const fetchLetter = async () => {
      if (!code) return
      
      setIsLoading(true)
      setError(null)
      
      try {
        // 使用真实API调用
        const response = await LetterService.getLetterByCode(code)
        
        if (response.success && response.data) {
          const letterData: LetterData = {
            ...response.data,
            isRead: response.data.status === 'read',
            deliveryNote: response.data.delivery_info?.delivery_location,
          }
          setLetter(letterData)
        } else {
          setError('未找到对应的信件，请检查编号是否正确')
        }
      } catch (err) {
        console.error('Failed to load letter:', err)
        setError('加载信件失败，请稍后重试')
      } finally {
        setIsLoading(false)
      }
    }

    fetchLetter()
  }, [code])

  const handleMarkAsRead = async () => {
    if (!letter || letter.isRead) return
    
    setIsMarkingAsRead(true)
    
    try {
      // 使用真实API标记为已读
      const response = await LetterService.markAsRead(letter.code!)
      
      if (response.success) {
        setLetter(prev => prev ? { ...prev, isRead: true, status: 'read' } : null)
      }
    } catch (err) {
      console.error('标记已读失败:', err)
    } finally {
      setIsMarkingAsRead(false)
    }
  }

  const handleReply = () => {
    if (!letter) return
    
    // 跳转到写信页面，并传入回信参数
    const searchParams = new URLSearchParams({
      replyTo: letter.code!,
      reply_to_sender: letter.sender_name,
      reply_to_title: letter.title || ''
    })
    
    router.push(`/write?${searchParams.toString()}`)
  }

  const handleShare = async () => {
    if (navigator.share) {
      try {
        await navigator.share({
          title: letter?.title || '一封来自朋友的信',
          text: '我收到了一封温暖的信件',
          url: window.location.href,
        })
      } catch (err) {
        console.log('分享失败:', err)
      }
    } else {
      // 降级方案：复制链接
      await navigator.clipboard.writeText(window.location.href)
      alert('链接已复制到剪贴板')
    }
  }

  const getStatusInfo = (status: string) => {
    switch (status) {
      case 'in_transit':
        return {
          label: '投递中',
          color: 'bg-orange-100 text-orange-800 border-orange-200',
          icon: Package
        }
      case 'delivered':
        return {
          label: '已投递',
          color: 'bg-green-100 text-green-800 border-green-200',
          icon: CheckCircle
        }
      case 'read':
        return {
          label: '已查看',
          color: 'bg-blue-100 text-blue-800 border-blue-200',
          icon: Mail
        }
      default:
        return {
          label: '已生成',
          color: 'bg-gray-100 text-gray-800 border-gray-200',
          icon: Mail
        }
    }
  }

  if (isLoading) {
    return (
      <div className="container max-w-4xl mx-auto px-4 py-8">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-muted-foreground">正在加载信件...</p>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container max-w-4xl mx-auto px-4 py-8">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    )
  }

  if (!letter) {
    return (
      <div className="container max-w-4xl mx-auto px-4 py-8">
        <Card className="text-center py-12">
          <CardContent>
            <Mail className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-semibold mb-2">信件不存在</h3>
            <p className="text-muted-foreground">
              未找到对应的信件，请检查编号是否正确
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  const statusInfo = getStatusInfo(letter.status)
  const StatusIcon = statusInfo.icon

  return (
    <div className="container max-w-4xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between mb-4">
          <h1 className="font-serif text-3xl font-bold text-letter-ink">
            {letter.title || '无标题信件'}
          </h1>
          <Badge className={statusInfo.color}>
            <StatusIcon className="mr-1 h-3 w-3" />
            {statusInfo.label}
          </Badge>
        </div>
        
        <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
          <div className="flex items-center gap-1">
            <User className="h-4 w-4" />
            <span>来自：{letter.sender_name}</span>
          </div>
          <div className="flex items-center gap-1">
            <Calendar className="h-4 w-4" />
            <span>{formatRelativeTime(new Date(letter.createdAt))}</span>
          </div>
          <div className="flex items-center gap-1">
            <Package className="h-4 w-4" />
            <span className="font-mono">{letter.code}</span>
          </div>
        </div>
      </div>

      {/* 投递信息 */}
      {letter.deliveryNote && (
        <Alert className="mb-6">
          <MapPin className="h-4 w-4" />
          <AlertDescription>
            <strong>投递信息：</strong> {letter.deliveryNote}
          </AlertDescription>
        </Alert>
      )}

      {letter.delivery_info && (
        <Alert className="mb-6">
          <MapPin className="h-4 w-4" />
          <AlertDescription>
            <strong>投递信息：</strong>
            {letter.delivery_info.delivery_location && (
              <div>位置: {letter.delivery_info.delivery_location}</div>
            )}
            {letter.delivery_info.courier_name && (
              <div>信使: {letter.delivery_info.courier_name}</div>
            )}
            {letter.delivery_info.deliveryTime && (
              <div>时间: {formatRelativeTime(new Date(letter.delivery_info.deliveryTime))}</div>
            )}
          </AlertDescription>
        </Alert>
      )}

      {/* 信件内容 */}
      <Card className="mb-6 letter-paper">
        <CardContent className="p-8">
          <div 
            className="prose prose-lg max-w-none font-serif text-letter-ink leading-loose"
            style={{ 
              backgroundImage: 'linear-gradient(transparent 29px, #e5e5e5 1px)', 
              backgroundSize: '100% 30px',
              minHeight: '400px'
            }}
          >
            <div className="whitespace-pre-wrap">
              {letter.content}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 操作按钮 */}
      <div className="flex flex-wrap gap-4 justify-center">
        {!letter.is_sender && !letter.isRead && (
          <Button 
            onClick={handleMarkAsRead}
            disabled={isMarkingAsRead}
            size="lg"
          >
            {isMarkingAsRead ? (
              <>
                <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                标记中...
              </>
            ) : (
              <>
                <CheckCircle className="mr-2 h-5 w-5" />
                标记为已读
              </>
            )}
          </Button>
        )}
        
        {letter.can_reply !== false && (
          <Button 
            variant="outline" 
            size="lg"
            onClick={handleReply}
          >
            <Reply className="mr-2 h-5 w-5" />
            回信
          </Button>
        )}
        
        <Button 
          variant="outline" 
          size="lg"
          onClick={handleShare}
        >
          <Share className="mr-2 h-5 w-5" />
          分享
        </Button>
        
        <Button variant="outline" size="lg">
          <Heart className="mr-2 h-5 w-5" />
          收藏
        </Button>
      </div>

      {/* 提示信息 */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle className="text-base">📮 温馨提示</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-muted-foreground">
          <p>• 这是一封通过OpenPenPal信使计划传递的真实手写信件</p>
          <p>• 如果你喜欢这封信，可以给写信人回信表达感谢</p>
          {!letter.isRead && <p>• 点击"标记为已读"让写信人知道你已经收到了这份温暖</p>}
          <p>• 你也可以加入OpenPenPal，体验手写信的魅力</p>
        </CardContent>
      </Card>
    </div>
  )
}