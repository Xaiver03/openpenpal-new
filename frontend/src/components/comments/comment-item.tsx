/**
 * CommentItem - Individual comment display component with actions
 * 评论项组件 - 显示单个评论及其操作
 */

import * as React from 'react'
import { useState, useCallback } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { 
  Heart, 
  MessageSquare, 
  MoreHorizontal, 
  Edit, 
  Trash2, 
  Flag,
  ChevronDown,
  ChevronUp
} from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'
import CommentForm from './comment-form'
import ReportCommentDialog from './report-comment-dialog'
import type { CommentItemProps, CommentAction, CommentFormData } from '@/types/comment'
import type { CommentReportRequestSOTA } from '@/types/comment-sota'

export const CommentItem = React.forwardRef<HTMLDivElement, CommentItemProps>(
  ({
    comment,
    depth = 0,
    max_depth = 3,
    enable_reply = true,
    enable_like = true,
    enable_edit = true,
    enable_delete = true,
    show_replies = true,
    on_action,
    className,
    ...props
  }, ref) => {
    const [show_reply_form, setShowReplyForm] = useState(false)
    const [show_edit_form, setShowEditForm] = useState(false)
    const [replies_expanded, setRepliesExpanded] = useState(true)
    const [is_loading, setIsLoading] = useState(false)
    const [show_report_dialog, setShowReportDialog] = useState(false)
    const [is_reporting, setIsReporting] = useState(false)

    // Format creation time
    const created_time = formatDistanceToNow(new Date(comment.created_at), {
      addSuffix: true,
      locale: zhCN
    })

    // Calculate indentation for nested comments
    const indent_level = Math.min(depth, max_depth)
    const indent_class = indent_level > 0 ? `ml-${indent_level * 4}` : ''

    // Handle action dispatch
    const dispatch_action = useCallback((action: CommentAction) => {
      setIsLoading(true)
      on_action?.(action)
      // Reset loading state after a delay (will be managed by parent component)
      setTimeout(() => setIsLoading(false), 1000)
    }, [on_action])

    // Handle like action
    const handle_like = useCallback(() => {
      dispatch_action({
        type: 'like',
        comment_id: comment.id,
      })
    }, [dispatch_action, comment.id])

    // Handle reply action
    const handle_reply = useCallback(async (form_data: CommentFormData) => {
      dispatch_action({
        type: 'reply',
        comment_id: comment.id,
        data: form_data,
      })
      setShowReplyForm(false)
    }, [dispatch_action, comment.id])

    // Handle edit action
    const handle_edit = useCallback(async (form_data: CommentFormData) => {
      dispatch_action({
        type: 'edit',
        comment_id: comment.id,
        data: form_data,
      })
      setShowEditForm(false)
    }, [dispatch_action, comment.id])

    // Handle delete action
    const handle_delete = useCallback(() => {
      if (window.confirm('确定要删除这条评论吗？')) {
        dispatch_action({
          type: 'delete',
          comment_id: comment.id,
        })
      }
    }, [dispatch_action, comment.id])

    // Handle report action
    const handle_report = useCallback(async (commentId: string, report: CommentReportRequestSOTA) => {
      setIsReporting(true)
      try {
        dispatch_action({
          type: 'report',
          comment_id: commentId,
          data: { report },
        })
      } finally {
        setIsReporting(false)
      }
    }, [dispatch_action])

    // Check if user can edit/delete (basic client-side check)
    const can_edit = enable_edit && comment.user_id === 'current_user_id' // TODO: Get from auth context
    const can_delete = enable_delete && comment.user_id === 'current_user_id' // TODO: Get from auth context

    return (
      <div
        ref={ref}
        className={cn(
          'group relative',
          indent_class,
          depth > 0 && 'border-l border-muted',
          className
        )}
        {...props}
      >
        {/* Main comment content */}
        <div className={cn(
          'flex gap-3 p-4',
          depth > 0 && 'pl-6'
        )}>
          {/* Avatar */}
          <Avatar className="h-8 w-8 flex-shrink-0">
            <AvatarImage src={comment.user?.avatar} />
            <AvatarFallback className="text-sm">
              {comment.user?.nickname?.[0] || comment.user?.username?.[0] || 'A'}
            </AvatarFallback>
          </Avatar>

          {/* Comment content */}
          <div className="flex-1 min-w-0 space-y-2">
            {/* Header with user info and time */}
            <div className="flex items-center gap-2 text-sm">
              <span className="font-medium">
                {comment.user?.nickname || comment.user?.username || '匿名用户'}
              </span>
              {comment.is_top && (
                <Badge variant="secondary" className="text-xs">
                  置顶
                </Badge>
              )}
              <span className="text-muted-foreground">·</span>
              <time className="text-muted-foreground" dateTime={comment.created_at}>
                {created_time}
              </time>
              {comment.updated_at !== comment.created_at && (
                <span className="text-xs text-muted-foreground">（已编辑）</span>
              )}
            </div>

            {/* Comment content */}
            {show_edit_form ? (
              <CommentForm
                letter_id={comment.letter_id}
                placeholder="编辑你的评论..."
                submit_text="保存"
                cancel_text="取消"
                on_submit={handle_edit}
                on_cancel={() => setShowEditForm(false)}
                className="mt-2"
              />
            ) : (
              <div className="prose prose-sm max-w-none text-foreground">
                <p className="whitespace-pre-wrap break-words">
                  {comment.content}
                </p>
              </div>
            )}

            {/* Action buttons */}
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              {/* Like button */}
              {enable_like && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handle_like}
                  disabled={is_loading}
                  className={cn(
                    'h-auto p-1 gap-1',
                    comment.is_liked && 'text-red-500'
                  )}
                >
                  <Heart className={cn(
                    'h-4 w-4',
                    comment.is_liked && 'fill-current'
                  )} />
                  {comment.like_count > 0 && (
                    <span>{comment.like_count}</span>
                  )}
                </Button>
              )}

              {/* Reply button */}
              {enable_reply && depth < max_depth && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowReplyForm(!show_reply_form)}
                  className="h-auto p-1 gap-1"
                >
                  <MessageSquare className="h-4 w-4" />
                  <span>回复</span>
                </Button>
              )}

              {/* Show replies toggle */}
              {show_replies && comment.reply_count > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setRepliesExpanded(!replies_expanded)}
                  className="h-auto p-1 gap-1"
                >
                  {replies_expanded ? (
                    <ChevronUp className="h-4 w-4" />
                  ) : (
                    <ChevronDown className="h-4 w-4" />
                  )}
                  <span>{comment.reply_count} 条回复</span>
                </Button>
              )}

              {/* More actions menu */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto p-1 opacity-0 group-hover:opacity-100"
                  >
                    <MoreHorizontal className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start">
                  {can_edit && (
                    <DropdownMenuItem onClick={() => setShowEditForm(true)}>
                      <Edit className="h-4 w-4 mr-2" />
                      编辑
                    </DropdownMenuItem>
                  )}
                  {can_delete && (
                    <DropdownMenuItem onClick={handle_delete} className="text-destructive">
                      <Trash2 className="h-4 w-4 mr-2" />
                      删除
                    </DropdownMenuItem>
                  )}
                  {(can_edit || can_delete) && <DropdownMenuSeparator />}
                  <DropdownMenuItem onClick={() => setShowReportDialog(true)}>
                    <Flag className="h-4 w-4 mr-2" />
                    举报
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {/* Reply form */}
            {show_reply_form && (
              <CommentForm
                letter_id={comment.letter_id}
                parent_id={comment.id}
                parent_comment={comment}
                placeholder="写下你的回复..."
                submit_text="回复"
                cancel_text="取消"
                on_submit={handle_reply}
                on_cancel={() => setShowReplyForm(false)}
                className="mt-3"
                auto_focus
              />
            )}
          </div>
        </div>

        {/* Nested replies */}
        {show_replies && replies_expanded && comment.replies && comment.replies.length > 0 && (
          <div className="space-y-0">
            {comment.replies.map((reply) => (
              <CommentItem
                key={reply.id}
                comment={reply}
                depth={depth + 1}
                max_depth={max_depth}
                enable_reply={enable_reply}
                enable_like={enable_like}
                enable_edit={enable_edit}
                enable_delete={enable_delete}
                show_replies={show_replies}
                on_action={on_action}
              />
            ))}
          </div>
        )}

        {/* Loading overlay */}
        {is_loading && (
          <div className="absolute inset-0 bg-background/50 flex items-center justify-center">
            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary" />
          </div>
        )}

        {/* Report dialog */}
        <ReportCommentDialog
          open={show_report_dialog}
          onOpenChange={setShowReportDialog}
          commentId={comment.id}
          commentAuthor={{
            username: comment.user?.username || '',
            nickname: comment.user?.nickname || ''
          }}
          onSubmit={handle_report}
          isSubmitting={is_reporting}
        />
      </div>
    )
  }
)

CommentItem.displayName = 'CommentItem'

export default CommentItem