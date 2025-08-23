import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { FutureLetterResponse } from '@/types/future-letter'
import { format, formatDistanceToNow, isAfter, isBefore } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { Calendar, Clock, Send, User, AlertCircle } from 'lucide-react'

interface FutureLetterCardProps {
  letter: FutureLetterResponse
  onView?: (letterId: string) => void
  onEdit?: (letterId: string) => void
  onCancel?: (letterId: string) => void
  showActions?: boolean
  variant?: 'scheduled' | 'incoming' | 'sent'
}

export function FutureLetterCard({
  letter,
  onView,
  onEdit,
  onCancel,
  showActions = true,
  variant = 'scheduled'
}: FutureLetterCardProps) {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'scheduled':
        return 'bg-blue-500'
      case 'sent':
        return 'bg-green-500'
      case 'cancelled':
        return 'bg-gray-500'
      default:
        return 'bg-blue-500'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'scheduled':
        return '已安排'
      case 'sent':
        return '已发送'
      case 'cancelled':
        return '已取消'
      default:
        return '未知'
    }
  }

  const getDeliveryMethodText = (method: string) => {
    switch (method) {
      case 'system':
        return '系统投递'
      case 'courier':
        return '信使投递'
      default:
        return '未知方式'
    }
  }

  const formatDate = (date: Date) => {
    return formatDistanceToNow(new Date(date), {
      addSuffix: true,
      locale: zhCN
    })
  }

  const formatScheduledDate = (date: Date) => {
    return format(new Date(date), 'yyyy年MM月dd日 HH:mm', { locale: zhCN })
  }

  const isOverdue = () => {
    return letter.status === 'scheduled' && isBefore(new Date(letter.scheduled_date), new Date())
  }

  const isUpcoming = () => {
    const scheduledDate = new Date(letter.scheduled_date)
    const now = new Date()
    const oneDayFromNow = new Date(now.getTime() + 24 * 60 * 60 * 1000)
    
    return letter.status === 'scheduled' && 
           isAfter(scheduledDate, now) && 
           isBefore(scheduledDate, oneDayFromNow)
  }

  return (
    <Card className="hover:shadow-md transition-shadow duration-200 group">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            <Calendar className="w-5 h-5 text-primary" />
            <span className="truncate">
              {letter.letter.title || '无题未来信'}
            </span>
            {isOverdue() && (
              <AlertCircle className="w-4 h-4 text-red-500" />
            )}
          </CardTitle>
          <div className="flex items-center gap-2">
            {isUpcoming() && (
              <Badge variant="outline" className="text-orange-600 border-orange-600">
                即将发送
              </Badge>
            )}
            <Badge className={`text-white ${getStatusColor(letter.status)}`}>
              {getStatusText(letter.status)}
            </Badge>
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* 信件内容预览 */}
        <div className="text-sm text-muted-foreground line-clamp-3">
          {variant === 'incoming' && letter.status === 'scheduled' ? (
            <span className="italic">内容将在预定时间揭晓...</span>
          ) : (
            letter.letter.content.length > 100 
              ? `${letter.letter.content.substring(0, 100)}...`
              : letter.letter.content
          )}
        </div>

        {/* 安排信息 */}
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-sm">
            <Clock className="w-4 h-4 text-primary" />
            <span className="font-medium">安排时间:</span>
            <span className={isOverdue() ? 'text-red-600' : 'text-foreground'}>
              {formatScheduledDate(letter.scheduled_date)}
            </span>
            <span className="text-muted-foreground">
              ({formatDate(letter.scheduled_date)})
            </span>
          </div>

          <div className="flex items-center gap-2 text-sm">
            <Send className="w-4 h-4 text-primary" />
            <span className="font-medium">投递方式:</span>
            <span>{getDeliveryMethodText(letter.delivery_method)}</span>
          </div>

          {letter.recipient && (
            <div className="flex items-center gap-2 text-sm">
              <User className="w-4 h-4 text-primary" />
              <span className="font-medium">收件人:</span>
              <span>
                {letter.recipient.nickname || letter.recipient.op_code || '匿名用户'}
              </span>
            </div>
          )}
        </div>

        {/* 提醒设置 */}
        {letter.reminder_enabled && (
          <div className="text-xs text-muted-foreground bg-muted/50 p-2 rounded">
            提醒已开启：提前 {letter.reminder_days} 天通知
            {letter.last_reminder_sent && (
              <span className="ml-2">
                (上次提醒: {formatDate(letter.last_reminder_sent)})
              </span>
            )}
          </div>
        )}

        {/* 发送信息 */}
        {letter.status === 'sent' && letter.sent_at && (
          <div className="text-xs text-green-600 bg-green-50 p-2 rounded">
            已于 {formatScheduledDate(letter.sent_at)} 发送
          </div>
        )}

        {/* 操作按钮 */}
        {showActions && (
          <div className="flex gap-2 pt-2">
            {onView && (
              <Button 
                size="sm" 
                variant="outline" 
                onClick={() => onView(letter.id)}
                className="flex-1"
              >
                查看详情
              </Button>
            )}
            
            {letter.status === 'scheduled' && variant === 'scheduled' && (
              <>
                {onEdit && (
                  <Button 
                    size="sm" 
                    variant="outline" 
                    onClick={() => onEdit(letter.id)}
                  >
                    编辑
                  </Button>
                )}
                
                {onCancel && (
                  <Button 
                    size="sm" 
                    variant="destructive" 
                    onClick={() => onCancel(letter.id)}
                  >
                    取消
                  </Button>
                )}
              </>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default FutureLetterCard