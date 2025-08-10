/**
 * CommentForm - Comment creation and editing form component
 * 评论表单组件 - 用于创建和编辑评论
 */

import * as React from 'react'
import { useState, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardFooter } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { cn } from '@/lib/utils'
import type { CommentFormProps, CommentFormData } from '@/types/comment'

export const CommentForm = React.forwardRef<HTMLFormElement, CommentFormProps>(
  ({
    letter_id,
    parent_id,
    parent_comment,
    placeholder = '写下你的评论...',
    max_length = 500,
    auto_focus = false,
    submit_text = '发表',
    cancel_text = '取消',
    on_submit,
    on_cancel,
    className,
    ...props
  }, ref) => {
    const [content, setContent] = useState('')
    const [is_submitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    // Handle form submission
    const handle_submit = useCallback(async (e: React.FormEvent) => {
      e.preventDefault()
      
      if (!content.trim()) {
        setError('评论内容不能为空')
        return
      }

      if (content.length > max_length) {
        setError(`评论内容不能超过 ${max_length} 个字符`)
        return
      }

      setIsSubmitting(true)
      setError(null)

      try {
        const form_data: CommentFormData = {
          content: content.trim(),
          parent_id,
        }

        await on_submit?.(form_data)
        
        // Clear form on successful submission
        setContent('')
      } catch (err) {
        const error_message = err instanceof Error ? err.message : '提交评论失败'
        setError(error_message)
      } finally {
        setIsSubmitting(false)
      }
    }, [content, max_length, parent_id, on_submit])

    // Handle cancel
    const handle_cancel = useCallback(() => {
      setContent('')
      setError(null)
      on_cancel?.()
    }, [on_cancel])

    // Character count with color coding
    const char_count = content.length
    const char_count_class = cn(
      'text-xs',
      char_count > max_length * 0.9 ? 'text-destructive' :
      char_count > max_length * 0.7 ? 'text-warning' :
      'text-muted-foreground'
    )

    return (
      <Card className={cn('border-none shadow-none', className)}>
        <form ref={ref} onSubmit={handle_submit} {...props}>
          <CardContent className="p-4 space-y-3">
            {/* Reply context for nested comments */}
            {parent_comment && (
              <div className="p-3 bg-muted/50 rounded-lg border-l-2 border-primary/20">
                <div className="flex items-start gap-2">
                  <Avatar className="h-6 w-6">
                    <AvatarImage src={parent_comment.user?.avatar} />
                    <AvatarFallback className="text-xs">
                      {parent_comment.user?.nickname?.[0] || 'A'}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium">
                      回复 @{parent_comment.user?.nickname || parent_comment.user?.username}
                    </p>
                    <p className="text-xs text-muted-foreground truncate">
                      {parent_comment.content}
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Comment input */}
            <div className="space-y-2">
              <Textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder={placeholder}
                autoFocus={auto_focus}
                disabled={is_submitting}
                rows={3}
                className={cn(
                  'resize-none',
                  error && 'border-destructive focus-visible:ring-destructive'
                )}
                maxLength={max_length + 50} // Allow slight overflow for better UX
              />
              
              {/* Character count */}
              <div className="flex justify-between items-center">
                <div>
                  {error && (
                    <p className="text-xs text-destructive">{error}</p>
                  )}
                </div>
                <span className={char_count_class}>
                  {char_count}/{max_length}
                </span>
              </div>
            </div>
          </CardContent>

          <CardFooter className="px-4 py-3 bg-muted/30 flex justify-end gap-2">
            {on_cancel && (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={handle_cancel}
                disabled={is_submitting}
              >
                {cancel_text}
              </Button>
            )}
            <Button
              type="submit"
              size="sm"
              disabled={is_submitting || !content.trim() || char_count > max_length}
              className="min-w-[60px]"
            >
              {is_submitting ? '提交中...' : submit_text}
            </Button>
          </CardFooter>
        </form>
      </Card>
    )
  }
)

CommentForm.displayName = 'CommentForm'

export default CommentForm