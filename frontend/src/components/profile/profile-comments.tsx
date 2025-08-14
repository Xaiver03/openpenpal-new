/**
 * ProfileComments - Profile Guestbook/Comments Component
 * 个人主页留言板组件 - 适配现有评论系统用于个人主页
 */

'use client'

import React, { useState, useCallback, useEffect } from 'react'
import { MessageSquare, Plus, RefreshCw, UserCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import CommentItem from '@/components/comments/comment-item'
import CommentForm from '@/components/comments/comment-form'
import { cn } from '@/lib/utils'
import { useUser } from '@/stores/user-store'
import type { Comment, CommentAction, CommentFormData } from '@/types/comment'

interface ProfileCommentsProps {
  profile_id: string
  profile_username?: string
  allow_comments?: boolean
  max_display?: number
  className?: string
}

// Mock data - will be replaced with real API calls
const mockComments: Comment[] = [
  {
    id: '1',
    letter_id: 'profile', // Using 'profile' as a special identifier
    user_id: '2',
    content: '很高兴认识你！你的信写得真好，希望能和你多交流 😊',
    status: 'active',
    like_count: 3,
    reply_count: 1,
    is_top: false,
    created_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
    user: {
      id: '2',
      username: 'alice',
      nickname: 'Alice',
      avatar: undefined
    },
    is_liked: false,
    replies: []
  },
  {
    id: '2',
    letter_id: 'profile',
    user_id: '3',
    content: '看了你的作品集，文笔很棒！期待更多精彩的信件 📝',
    status: 'active',
    like_count: 5,
    reply_count: 0,
    is_top: false,
    created_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
    user: {
      id: '3',
      username: 'bob',
      nickname: 'Bob',
      avatar: undefined
    },
    is_liked: true,
    replies: []
  }
]

export function ProfileComments({ 
  profile_id, 
  profile_username,
  allow_comments = true,
  max_display = 10,
  className 
}: ProfileCommentsProps) {
  const { user: currentUser } = useUser()
  const [comments, setComments] = useState<Comment[]>([])
  const [loading, setLoading] = useState(true)
  const [showCommentForm, setShowCommentForm] = useState(false)
  const [commentCount, setCommentCount] = useState(0)

  // Load comments
  useEffect(() => {
    loadComments()
  }, [profile_id])

  const loadComments = async () => {
    try {
      setLoading(true)
      // TODO: Replace with real API call
      // const response = await fetch(`/api/users/${profile_id}/comments`)
      // const data = await response.json()
      
      // Mock implementation
      setTimeout(() => {
        setComments(mockComments)
        setCommentCount(mockComments.length)
        setLoading(false)
      }, 500)
    } catch (error) {
      console.error('Failed to load profile comments:', error)
      setLoading(false)
    }
  }

  // Handle comment actions
  const handleCommentAction = useCallback(async (action: CommentAction) => {
    try {
      switch (action.type) {
        case 'like':
          // TODO: Implement like API
          setComments(prev => prev.map(comment => 
            comment.id === action.comment_id
              ? { 
                  ...comment, 
                  is_liked: !comment.is_liked,
                  like_count: comment.is_liked ? comment.like_count - 1 : comment.like_count + 1
                }
              : comment
          ))
          break
          
        case 'reply':
          if (action.data && currentUser) {
            // TODO: Implement reply API
            const newReply: Comment = {
              id: `reply-${Date.now()}`,
              letter_id: 'profile',
              user_id: currentUser.id,
              parent_id: action.comment_id,
              content: action.data.content,
              status: 'active',
              like_count: 0,
              reply_count: 0,
              is_top: false,
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
              user: {
                id: currentUser.id,
                username: currentUser.username,
                nickname: currentUser.nickname || currentUser.username,
                avatar: (currentUser as any).avatar_url
              },
              is_liked: false
            }
            
            setComments(prev => prev.map(comment => 
              comment.id === action.comment_id
                ? { 
                    ...comment, 
                    replies: [...(comment.replies || []), newReply],
                    reply_count: comment.reply_count + 1
                  }
                : comment
            ))
          }
          break
          
        case 'delete':
          // TODO: Implement delete API
          setComments(prev => prev.filter(comment => comment.id !== action.comment_id))
          setCommentCount(prev => prev - 1)
          break
      }
    } catch (err) {
      console.error('Comment action failed:', err)
    }
  }, [currentUser])

  // Handle new comment submission
  const handleNewComment = useCallback(async (formData: CommentFormData) => {
    if (!currentUser) return

    try {
      // TODO: Replace with real API call
      // const response = await fetch(`/api/users/${profile_id}/comments`, {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ content: formData.content })
      // })
      
      // Mock implementation
      const newComment: Comment = {
        id: `comment-${Date.now()}`,
        letter_id: 'profile',
        user_id: currentUser.id,
        content: formData.content,
        status: 'active',
        like_count: 0,
        reply_count: 0,
        is_top: false,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        user: {
          id: currentUser.id,
          username: currentUser.username,
          nickname: currentUser.nickname || currentUser.username,
          avatar: (currentUser as any).avatar_url
        },
        is_liked: false,
        replies: []
      }
      
      setComments(prev => [newComment, ...prev])
      setCommentCount(prev => prev + 1)
      setShowCommentForm(false)
    } catch (error) {
      console.error('Failed to create comment:', error)
    }
  }, [currentUser, profile_id])

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <MessageSquare className="h-5 w-5" />
              <h3 className="text-lg font-semibold">留言板</h3>
              <Badge variant="secondary">
                {commentCount}
              </Badge>
            </div>
          </div>

          <Button
            variant="ghost"
            size="sm"
            onClick={loadComments}
            disabled={loading}
            className="h-8 w-8 p-0"
          >
            <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
          </Button>
        </div>

        {/* Add comment button */}
        {allow_comments && currentUser && (
          <div className="pt-2">
            {showCommentForm ? (
              <CommentForm
                letter_id={`profile-${profile_id}`} // Using profile prefix
                placeholder={`给 ${profile_username || '这位用户'} 留言...`}
                submit_text="发表留言"
                cancel_text="取消"
                on_submit={handleNewComment}
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
                写留言...
              </Button>
            )}
          </div>
        )}
        
        {/* Login prompt for non-logged in users */}
        {allow_comments && !currentUser && (
          <div className="pt-2">
            <div className="text-center p-4 bg-muted/50 rounded-lg">
              <UserCircle className="h-8 w-8 text-muted-foreground mx-auto mb-2" />
              <p className="text-sm text-muted-foreground">
                <a href="/login" className="text-primary hover:underline">登录</a>
                {' '}后即可留言
              </p>
            </div>
          </div>
        )}
      </CardHeader>

      <CardContent className="p-0">
        {/* Loading state */}
        {loading && comments.length === 0 && (
          <div className="p-8 text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4" />
            <p className="text-muted-foreground">加载留言中...</p>
          </div>
        )}

        {/* Empty state */}
        {!loading && comments.length === 0 && (
          <div className="p-8 text-center">
            <MessageSquare className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h4 className="font-medium mb-2">暂无留言</h4>
            <p className="text-sm text-muted-foreground mb-4">
              成为第一个留言的人
            </p>
            {allow_comments && currentUser && !showCommentForm && (
              <Button onClick={() => setShowCommentForm(true)}>
                写第一条留言
              </Button>
            )}
          </div>
        )}

        {/* Comment list */}
        {comments.length > 0 && (
          <div className="divide-y">
            {comments.slice(0, max_display).map((comment, index) => (
              <React.Fragment key={comment.id}>
                <CommentItem
                  comment={comment}
                  depth={0}
                  max_depth={1} // Limit depth for profile comments
                  enable_reply={allow_comments && !!currentUser}
                  enable_like={true}
                  enable_edit={false} // Disable edit for now
                  enable_delete={currentUser?.id === comment.user_id}
                  show_replies={true}
                  on_action={handleCommentAction}
                />
                {index < comments.length - 1 && <Separator />}
              </React.Fragment>
            ))}
          </div>
        )}

        {/* Show more link */}
        {comments.length > max_display && (
          <div className="p-4 text-center border-t">
            <Button variant="ghost" size="sm">
              查看全部 {commentCount} 条留言
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default ProfileComments