/**
 * CommentStats - Simple comment statistics display component
 * 评论统计组件 - 显示评论数量等统计信息
 */

import * as React from 'react'
import { MessageSquare } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { useCommentStats } from '@/hooks/use-comments'
import type { CommentStatsProps } from '@/types/comment'

export const CommentStats: React.FC<CommentStatsProps> = ({
  letter_id,
  show_icon = true,
  format = 'compact',
  className,
  ...props
}) => {
  const { stats, loading, error } = useCommentStats(letter_id)

  if (loading) {
    return (
      <div className={cn('flex items-center gap-2', className)} {...props}>
        {show_icon && <MessageSquare className="h-4 w-4 text-muted-foreground" />}
        <Skeleton className="h-4 w-8" />
      </div>
    )
  }

  if (error || !stats) {
    return null
  }

  const comment_count = stats.comment_count

  if (format === 'full') {
    return (
      <div className={cn('flex items-center gap-2 text-sm text-muted-foreground', className)} {...props}>
        {show_icon && <MessageSquare className="h-4 w-4" />}
        <span>
          {comment_count === 0 ? '暂无评论' : 
           comment_count === 1 ? '1 条评论' : 
           `${comment_count} 条评论`}
        </span>
      </div>
    )
  }

  // Compact format - just icon and number
  return (
    <Button
      variant="ghost"
      size="sm"
      className={cn(
        'h-auto p-1 gap-1 text-muted-foreground hover:text-foreground',
        className
      )}
      {...props}
    >
      {show_icon && <MessageSquare className="h-4 w-4" />}
      {comment_count > 0 && (
        <span className="text-sm">{comment_count}</span>
      )}
    </Button>
  )
}

CommentStats.displayName = 'CommentStats'

// Simple badge version for inline use
export const CommentCountBadge: React.FC<CommentStatsProps> = ({ 
  letter_id, 
  className, 
  ...props 
}) => {
  const { stats, loading } = useCommentStats(letter_id)

  if (loading || !stats || stats.comment_count === 0) {
    return null
  }

  return (
    <Badge
      variant="secondary"
      className={cn('text-xs', className)}
      {...props}
    >
      {stats.comment_count}
    </Badge>
  )
}

CommentCountBadge.displayName = 'CommentCountBadge'

export default CommentStats