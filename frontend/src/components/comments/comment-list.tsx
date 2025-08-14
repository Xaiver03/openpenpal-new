/**
 * CommentList - Main comment list component with sorting, pagination, and stats
 * 评论列表组件 - 包含排序、分页和统计功能
 */

import * as React from 'react'
import { useState, useCallback } from 'react'
import { ArrowUpDown, MessageSquare, Plus, RefreshCw } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { useComments } from '@/hooks/use-comments'
import CommentItem from './comment-item'
import CommentForm from './comment-form'
import type { CommentListProps, CommentAction, CommentFormData } from '@/types/comment'

export const CommentList = React.forwardRef<HTMLDivElement, CommentListProps>(
  ({
    letterId,
    maxDepth = 3,
    enableNested = true,
    showStats = true,
    allowComments = true,
    initialSort = 'createdAt',
    className,
    ...props
  }, ref) => {
    const [showCommentForm, setShowCommentForm] = useState(false)
    const [sortBy, setSortBy] = useState(initialSort)
    const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')

    // Use comment hook for state management
    const {
      comments,
      stats,
      loading,
      error,
      hasMore,
      loadComments,
      createComment,
      updateComment,
      deleteComment,
      likeComment,
      loadReplies,
      refresh,
      clearError,
    } = useComments({
      letterId,
      initialQuery: {
        sortBy: initialSort,
        order: 'desc',
        limit: 20,
      },
      auto_load: true,
      enable_real_time: false,
    })

    // Handle sorting change
    const handle_sort_change = useCallback((new_sort_by: string) => {
      const sort_field = new_sort_by as 'created_at' | 'like_count'
      setSortBy(sort_field)
      load_comments({ 
        page: 1, 
        sort_by: sort_field,
        order: sort_order 
      })
    }, [sort_order, load_comments])

    // Toggle sort order
    const toggle_sort_order = useCallback(() => {
      const new_order = sort_order === 'desc' ? 'asc' : 'desc'
      setSortOrder(new_order)
      load_comments({ 
        page: 1, 
        sort_by,
        order: new_order 
      })
    }, [sort_by, sort_order, load_comments])

    // Handle comment actions
    const handle_comment_action = useCallback(async (action: CommentAction) => {
      try {
        switch (action.type) {
          case 'like':
            await like_comment(action.comment_id)
            break
          
          case 'reply':
            if (action.data) {
              await create_comment({
                letter_id,
                parent_id: action.comment_id,
                content: action.data.content,
              })
            }
            break
          
          case 'edit':
            if (action.data) {
              await update_comment(action.comment_id, {
                content: action.data.content,
              })
            }
            break
          
          case 'delete':
            await delete_comment(action.comment_id)
            break
        }
      } catch (err) {
        console.error('Comment action failed:', err)
        // Error handling is managed by the hook
      }
    }, [letter_id, like_comment, create_comment, update_comment, delete_comment])

    // Handle new comment submission
    const handle_new_comment = useCallback(async (form_data: CommentFormData) => {
      await create_comment({
        letter_id,
        content: form_data.content,
      })
      setShowCommentForm(false)
    }, [letter_id, create_comment])

    // Load more comments
    const load_more = useCallback(() => {
      const current_page = Math.floor(comments.length / 20) + 1
      load_comments({ 
        page: current_page + 1,
        sort_by,
        order: sort_order 
      })
    }, [comments.length, sort_by, sort_order, load_comments])

    return (
      <Card ref={ref} className={cn('w-full', className)} {...props}>
        <CardHeader className="pb-4">
          {/* Header with stats and controls */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <MessageSquare className="h-5 w-5" />
                <h3 className="text-lg font-semibold">评论</h3>
                {show_stats && stats && (
                  <Badge variant="secondary">
                    {stats.comment_count}
                  </Badge>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2">
              {/* Refresh button */}
              <Button
                variant="ghost"
                size="sm"
                onClick={refresh}
                disabled={loading}
                className="h-8 w-8 p-0"
              >
                <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
              </Button>

              {/* Sort controls */}
              <div className="flex items-center gap-1">
                <Select value={sort_by} onValueChange={handle_sort_change}>
                  <SelectTrigger className="h-8 w-auto text-sm">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="created_at">时间</SelectItem>
                    <SelectItem value="like_count">热度</SelectItem>
                  </SelectContent>
                </Select>
                
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={toggle_sort_order}
                  className="h-8 w-8 p-0"
                >
                  <ArrowUpDown className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>

          {/* Add comment button */}
          {allow_comments && (
            <div className="pt-2">
              {show_comment_form ? (
                <CommentForm
                  letter_id={letter_id}
                  placeholder="分享你的想法..."
                  submit_text="发表评论"
                  cancel_text="取消"
                  on_submit={handle_new_comment}
                  on_cancel={() => setShowCommentForm(false)}
                  auto_focus
                />
              ) : (
                <Button
                  variant="outline"
                  onClick={() => setShowCommentForm(true)}
                  className="w-full justify-start text-muted-foreground"
                >
                  <Plus className="h-4 w-4 mr-2" />
                  写评论...
                </Button>
              )}
            </div>
          )}
        </CardHeader>

        <CardContent className="p-0">
          {/* Error display */}
          {error && (
            <div className="p-4 bg-destructive/10 text-destructive text-sm">
              <p>{error}</p>
              <Button
                variant="ghost"
                size="sm"
                onClick={clear_error}
                className="mt-2 h-auto p-0 text-destructive"
              >
                重试
              </Button>
            </div>
          )}

          {/* Loading state */}
          {loading && comments.length === 0 && (
            <div className="p-8 text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4" />
              <p className="text-muted-foreground">加载评论中...</p>
            </div>
          )}

          {/* Empty state */}
          {!loading && comments.length === 0 && !error && (
            <div className="p-8 text-center">
              <MessageSquare className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h4 className="font-medium mb-2">暂无评论</h4>
              <p className="text-sm text-muted-foreground mb-4">
                成为第一个发表评论的人
              </p>
              {allow_comments && !show_comment_form && (
                <Button onClick={() => setShowCommentForm(true)}>
                  写第一条评论
                </Button>
              )}
            </div>
          )}

          {/* Comment list */}
          {comments.length > 0 && (
            <div className="divide-y">
              {comments.map((comment, index) => (
                <React.Fragment key={comment.id}>
                  <CommentItem
                    comment={comment}
                    depth={0}
                    max_depth={max_depth}
                    enable_reply={allow_comments && enable_nested}
                    enable_like={true}
                    enable_edit={allow_comments}
                    enable_delete={allow_comments}
                    show_replies={enable_nested}
                    on_action={handle_comment_action}
                  />
                  {index < comments.length - 1 && <Separator />}
                </React.Fragment>
              ))}
            </div>
          )}

          {/* Load more button */}
          {has_more && !loading && (
            <div className="p-4 text-center">
              <Button
                variant="outline"
                onClick={load_more}
                disabled={loading}
              >
                {loading ? '加载中...' : '加载更多'}
              </Button>
            </div>
          )}

          {/* Bottom loading indicator */}
          {loading && comments.length > 0 && (
            <div className="p-4 text-center">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary mx-auto" />
            </div>
          )}
        </CardContent>
      </Card>
    )
  }
)

CommentList.displayName = 'CommentList'

export default CommentList