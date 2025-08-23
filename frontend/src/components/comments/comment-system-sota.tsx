/**
 * CommentSystemSOTA - Unified comment system supporting multiple target types
 * 通用评论系统组件 - 支持多种目标类型
 */

import * as React from 'react'
import { useState, useCallback } from 'react'
import { MessageSquare, ChevronDown, ChevronUp } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Separator } from '@/components/ui/separator'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useAuth } from '@/contexts/auth-context-new'
import { toast } from '@/components/ui/use-toast'
import { cn } from '@/lib/utils'
import { useCommentsSOTA } from '@/hooks/useCommentsSOTA'
import type { CommentTargetType } from '@/types/comment-sota'
import CommentItem from './comment-item'
import CommentForm from './comment-form'

interface CommentSystemSOTAProps {
  targetId: string
  targetType: CommentTargetType
  title?: string
  placeholder?: string
  enableReplies?: boolean
  maxDepth?: number
  showStats?: boolean
  className?: string
  onCommentCreated?: () => void
}

export const CommentSystemSOTA: React.FC<CommentSystemSOTAProps> = ({
  targetId,
  targetType,
  title = '评论',
  placeholder = '写下您的想法...',
  enableReplies = true,
  maxDepth = 3,
  showStats = true,
  className,
  onCommentCreated
}) => {
  const { user } = useAuth()
  const [showCommentForm, setShowCommentForm] = useState(false)
  const [sortBy, setSortBy] = useState<'created_at' | 'like_count'>('created_at')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [expandedComments, setExpandedComments] = useState<Set<string>>(new Set())

  // Use SOTA comment hook
  const {
    comments,
    stats,
    loading,
    error,
    has_more,
    load_comments,
    create_comment,
    update_comment,
    delete_comment,
    like_comment,
    refresh
  } = useCommentsSOTA({
    target_id: targetId,
    target_type: targetType,
    auto_load: true,
    enable_real_time: true,
    initial_query: {
      sort_by: sortBy,
      order: sortOrder,
      limit: 20
    }
  })

  // Handle comment creation
  const handleCreateComment = useCallback(async (data: { content: string; parent_id?: string }) => {
    try {
      await create_comment({
        target_id: targetId,
        target_type: targetType,
        content: data.content,
        parent_id: data.parent_id
      })
      toast({
        title: '评论成功',
        description: '您的评论已发布'
      })
      setShowCommentForm(false)
      onCommentCreated?.()
    } catch (err) {
      toast({
        title: '评论失败',
        description: '请稍后重试',
        variant: 'destructive'
      })
    }
  }, [create_comment, targetId, targetType, onCommentCreated])

  // Handle sort change
  const handleSortChange = useCallback((newSortBy: string) => {
    const [field, order] = newSortBy.split('-') as ['created_at' | 'like_count', 'asc' | 'desc']
    setSortBy(field)
    setSortOrder(order)
    load_comments({
      sort_by: field,
      order: order,
      page: 1
    })
  }, [load_comments])

  // Toggle comment expansion
  const toggleCommentExpanded = useCallback((commentId: string) => {
    setExpandedComments(prev => {
      const next = new Set(prev)
      if (next.has(commentId)) {
        next.delete(commentId)
      } else {
        next.add(commentId)
      }
      return next
    })
  }, [])

  // Render comment tree recursively
  const renderCommentTree = useCallback((parentId: string | null = null, depth: number = 0) => {
    const childComments = comments.filter(c => c.parent_id === parentId)
    
    if (childComments.length === 0) return null

    return (
      <div className={cn(
        'space-y-4',
        depth > 0 && 'ml-12 mt-4'
      )}>
        {childComments.map(comment => {
          const hasReplies = comments.some(c => c.parent_id === comment.id)
          const isExpanded = expandedComments.has(comment.id)

          return (
            <div key={comment.id}>
              <CommentItem
                comment={{
                  id: comment.id,
                  user_id: comment.user_id,
                  user: {
                    id: comment.user?.id || comment.user_id,
                    username: comment.user?.username || '匿名用户',
                    nickname: comment.user?.nickname || comment.user?.username || '匿名用户',
                    avatar: comment.user?.avatar
                  },
                  content: comment.content,
                  created_at: comment.created_at,
                  updated_at: comment.updated_at,
                  likes: comment.like_count || 0,
                  is_liked: comment.is_liked || false,
                  status: comment.status as any, // Type compatibility between SOTA and regular comment
                  parent_id: comment.parent_id || undefined,
                  replies: []
                } as any}
                enable_reply={enableReplies && depth < maxDepth - 1}
                enable_like={true}
                enable_edit={user?.id === comment.user_id}
                enable_delete={user?.id === comment.user_id}
                on_action={async (action) => {
                  switch (action.type) {
                    case 'reply':
                      if (action.data?.content && user) {
                        await handleCreateComment({ content: action.data.content, parent_id: comment.id })
                      }
                      break
                    case 'like':
                      await like_comment(comment.id)
                      break
                    case 'edit':
                      if (action.data?.content) {
                        await update_comment(comment.id, { content: action.data.content })
                      }
                      break
                    case 'delete':
                      await delete_comment(comment.id)
                      break
                  }
                }}
              />
              
              {hasReplies && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="ml-12 mt-2 text-xs"
                  onClick={() => toggleCommentExpanded(comment.id)}
                >
                  {isExpanded ? <ChevronUp className="w-3 h-3 mr-1" /> : <ChevronDown className="w-3 h-3 mr-1" />}
                  {isExpanded ? '收起回复' : '查看回复'}
                </Button>
              )}

              {isExpanded && renderCommentTree(comment.id, depth + 1)}
            </div>
          )
        })}
      </div>
    )
  }, [comments, expandedComments, enableReplies, maxDepth, user, handleCreateComment, like_comment, update_comment, delete_comment, toggleCommentExpanded])

  if (loading && comments.length === 0) {
    return (
      <Card className={className}>
        <CardContent className="py-8">
          <div className="animate-pulse space-y-4">
            <div className="h-4 bg-muted rounded w-3/4"></div>
            <div className="h-4 bg-muted rounded w-1/2"></div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className={className}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <MessageSquare className="w-5 h-5" />
            {title}
            {showStats && stats && (
              <span className="text-sm font-normal text-muted-foreground">
                ({stats.total_comments || 0})
              </span>
            )}
          </CardTitle>
          
          {comments.length > 0 && (
            <Select value={`${sortBy}-${sortOrder}`} onValueChange={handleSortChange}>
              <SelectTrigger className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="created_at-desc">最新</SelectItem>
                <SelectItem value="created_at-asc">最早</SelectItem>
                <SelectItem value="like_count-desc">最热</SelectItem>
              </SelectContent>
            </Select>
          )}
        </div>
      </CardHeader>

      <CardContent>
        {/* Comment form */}
        {user ? (
          showCommentForm ? (
            <div className="mb-6">
              <CommentForm
                letter_id={targetId} // For backward compatibility
                placeholder={placeholder}
                on_submit={(data: { content: string }) => handleCreateComment({ content: data.content })}
                on_cancel={() => setShowCommentForm(false)}
                auto_focus
              />
            </div>
          ) : (
            <Button
              variant="outline"
              className="w-full mb-6 justify-start text-muted-foreground"
              onClick={() => setShowCommentForm(true)}
            >
              {placeholder}
            </Button>
          )
        ) : (
          <Alert className="mb-6">
            <AlertDescription>
              登录后即可发表评论
            </AlertDescription>
          </Alert>
        )}

        {/* Comments list */}
        {error && (
          <Alert variant="destructive" className="mb-4">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {comments.length > 0 ? (
          <>
            {renderCommentTree()}
            
            {has_more && (
              <div className="mt-6 text-center">
                <Button
                  variant="outline"
                  onClick={() => load_comments({ page: (comments.length / 20) + 1 })}
                  disabled={loading}
                >
                  {loading ? '加载中...' : '加载更多'}
                </Button>
              </div>
            )}
          </>
        ) : (
          <div className="text-center py-8 text-muted-foreground">
            <MessageSquare className="w-12 h-12 mx-auto mb-2 opacity-20" />
            <p>还没有评论，来做第一个评论的人吧！</p>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default CommentSystemSOTA