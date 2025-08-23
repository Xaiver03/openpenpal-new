import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { DriftBottleResponse } from '@/types/drift-bottle'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { Waves, Clock, User } from 'lucide-react'

interface DriftBottleCardProps {
  bottle: DriftBottleResponse
  onCatch?: (bottleId: string) => void
  onView?: (bottleId: string) => void
  showActions?: boolean
  variant?: 'floating' | 'caught' | 'sent'
}

export function DriftBottleCard({
  bottle,
  onCatch,
  onView,
  showActions = true,
  variant = 'floating'
}: DriftBottleCardProps) {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'floating':
        return 'bg-blue-500'
      case 'collected':
        return 'bg-green-500'
      case 'expired':
        return 'bg-gray-500'
      default:
        return 'bg-blue-500'
    }
  }

  const getThemeEmoji = (theme?: string) => {
    switch (theme) {
      case 'friendship':
        return '🤝'
      case 'love':
        return '💕'
      case 'confession':
        return '💌'
      case 'wish':
        return '🌟'
      case 'gratitude':
        return '🙏'
      case 'memory':
        return '📷'
      default:
        return '🌊'
    }
  }

  const formatDate = (date: Date) => {
    return formatDistanceToNow(new Date(date), {
      addSuffix: true,
      locale: zhCN
    })
  }

  return (
    <Card className="hover:shadow-md transition-shadow duration-200 group">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            <span className="text-2xl">{getThemeEmoji(bottle.theme)}</span>
            <span className="truncate">
              {bottle.letter.title || '无题信件'}
            </span>
          </CardTitle>
          <Badge className={`text-white ${getStatusColor(bottle.status)}`}>
            {bottle.status === 'floating' ? '漂流中' : 
             bottle.status === 'collected' ? '已捞取' : '已过期'}
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* 信件内容预览 */}
        <div className="text-sm text-muted-foreground line-clamp-3">
          {bottle.letter.content.length > 100 
            ? `${bottle.letter.content.substring(0, 100)}...`
            : bottle.letter.content
          }
        </div>

        {/* 信息行 */}
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-1">
              <Waves className="w-4 h-4" />
              <span>{bottle.region || '未知海域'}</span>
            </div>
            
            {bottle.status === 'floating' && (
              <div className="flex items-center gap-1">
                <Clock className="w-4 h-4" />
                <span>
                  {formatDate(bottle.expires_at)} 过期
                </span>
              </div>
            )}
            
            {bottle.status === 'collected' && bottle.collector && (
              <div className="flex items-center gap-1">
                <User className="w-4 h-4" />
                <span>被 {bottle.collector.nickname || '匿名用户'} 捞取</span>
              </div>
            )}
          </div>
          
          <div className="text-xs">
            {formatDate(bottle.created_at)}
          </div>
        </div>

        {/* 操作按钮 */}
        {showActions && (
          <div className="flex gap-2 pt-2">
            {variant === 'floating' && bottle.status === 'floating' && onCatch && (
              <Button 
                size="sm" 
                onClick={() => onCatch(bottle.id)}
                className="flex-1 group-hover:shadow-sm transition-shadow"
              >
                <Waves className="w-4 h-4 mr-2" />
                捞取这个瓶子
              </Button>
            )}
            
            {onView && (
              <Button 
                size="sm" 
                variant="outline" 
                onClick={() => onView(bottle.id)}
                className="flex-1"
              >
                查看详情
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default DriftBottleCard