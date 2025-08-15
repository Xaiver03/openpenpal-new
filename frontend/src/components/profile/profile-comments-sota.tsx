/**
 * ProfileCommentsSOTA - SOTA Profile Guestbook/Comments Component
 * SOTA个人主页留言板组件 - 支持多目标评论系统的完整实现
 */

'use client'

import React, { useState, useCallback, useEffect } from 'react'
import { MessageSquare, Plus, RefreshCw, UserCircle, TrendingUp, Activity } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { cn } from '@/lib/utils'
import { useUser } from '@/stores/user-store'
import { useCommentsSOTA } from '@/hooks/useCommentsSOTA'
import { CommentItem as CommentItemSOTA } from '@/components/comments/comment-item'
import { CommentForm as CommentFormSOTA } from '@/components/comments/comment-form'
import type { 
  CommentSOTA, 
  CommentActionSOTA, 
  CommentFormDataSOTA,
  CommentSortModeSOTA,
  CommentDisplayModeSOTA 
} from '@/types/comment-sota'

interface ProfileCommentsSOTAProps {
  profile_id: string
  profile_username?: string
  profile_nickname?: string
  allow_comments?: boolean
  allow_anonymous?: boolean
  max_display?: number
  enable_real_time?: boolean
  className?: string
}

export function ProfileCommentsSOTA({ 
  profile_id, 
  profile_username,
  profile_nickname,
  allow_comments = true,
  allow_anonymous = false,
  max_display = 20,
  enable_real_time = true,
  className 
}: ProfileCommentsSOTAProps) {
  const { user: currentUser } = useUser()
  const [showCommentForm, setShowCommentForm] = useState(false)
  const [sortMode, setSortMode] = useState<CommentSortModeSOTA>('newest')
  const [displayMode, setDisplayMode] = useState<CommentDisplayModeSOTA>('threaded')
  const [activeTab, setActiveTab] = useState<'all' | 'popular' | 'recent'>('all')

  // Use SOTA comment hook
  const {
    comments,
    stats,
    loading,
    error,
    has_more,
    load_comments,
    create_comment,
    like_comment,
    report_comment,
    refresh,
    clear_error,
    get_replies,
    toggle_replies_expanded
  } = useCommentsSOTA({
    target_id: profile_id,
    target_type: 'profile',
    initial_query: {
      limit: max_display,
      sort_by: sortMode === 'newest' ? 'created_at' : 'like_count',
      order: sortMode === 'newest' ? 'desc' : 'desc',
      include_replies: true,
      max_level: 3
    },
    auto_load: true,
    enable_real_time: enable_real_time,
    enable_polling: !enable_real_time,
    polling_interval: 30000 // 30 seconds
  })

  // Handle sort change
  const handleSortChange = useCallback((newSort: CommentSortModeSOTA) => {
    setSortMode(newSort)
    const sortBy = newSort === 'popular' ? 'like_count' : 'created_at'
    const order = newSort === 'oldest' ? 'asc' : 'desc'
    
    load_comments({
      sort_by: sortBy,
      order: order,
      limit: max_display
    })
  }, [load_comments, max_display])

  // Handle comment actions
  const handleCommentAction = useCallback(async (action: CommentActionSOTA) => {
    try {
      switch (action.type) {
        case 'like':
          if (action.comment_id) {
            await like_comment(action.comment_id)
          }
          break
          
        case 'report':
          if (action.comment_id && action.data) {
            await report_comment(action.comment_id, action.data)
          }
          break
          
        case 'reply':
          // Handle reply through the reply form in CommentItemSOTA
          break
          
        default:
          console.warn('Unhandled comment action:', action)
      }
    } catch (err) {
      console.error('Comment action failed:', err)
    }
  }, [like_comment, report_comment])

  // Handle new comment submission
  const handleNewComment = useCallback(async (formData: CommentFormDataSOTA) => {
    try {
      await create_comment({
        target_id: profile_id,
        target_type: 'profile',
        content: formData.content,
        is_anonymous: formData.is_anonymous || false
      })
      setShowCommentForm(false)
    } catch (error) {
      console.error('Failed to create comment:', error)
    }
  }, [create_comment, profile_id])

  // Filter comments based on active tab
  const filteredComments = React.useMemo(() => {
    switch (activeTab) {
      case 'popular':
        return [...comments].sort((a, b) => b.like_count - a.like_count)
      case 'recent':
        return [...comments].sort((a, b) => 
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        )
      default:
        return comments
    }
  }, [comments, activeTab])

  // Display name for the profile owner
  const displayName = profile_nickname || profile_username || '这位用户'

  return (
    <Card className={cn('w-full', className)}>
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <MessageSquare className="h-5 w-5" />
              <h3 className="text-lg font-semibold">留言板</h3>
              {stats && (
                <Badge variant="secondary">
                  {stats.total_comments}
                </Badge>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            {/* Sort selector */}
            <Select value={sortMode} onValueChange={(value: CommentSortModeSOTA) => handleSortChange(value)}>
              <SelectTrigger className="w-24 h-8 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="newest">最新</SelectItem>
                <SelectItem value="oldest">最早</SelectItem>
                <SelectItem value="popular">热门</SelectItem>
              </SelectContent>
            </Select>

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
          </div>
        </div>

        {/* Stats summary */}
        {stats && (
          <div className="flex items-center gap-4 text-sm text-muted-foreground pt-2">
            <div className="flex items-center gap-1">
              <MessageSquare className="h-3 w-3" />
              <span>{stats.total_comments} 条留言</span>
            </div>
            <div className="flex items-center gap-1">
              <TrendingUp className="h-3 w-3" />
              <span>{stats.total_likes} 个赞</span>
            </div>
            {stats.pending_comments > 0 && (
              <div className="flex items-center gap-1">
                <Activity className="h-3 w-3" />
                <span>{stats.pending_comments} 条待审核</span>
              </div>
            )}
          </div>
        )}

        {/* Error display */}
        {error && (
          <Alert variant="destructive" className="mt-2">
            <AlertDescription className="flex items-center justify-between">
              {error}
              <Button variant="ghost" size="sm" onClick={clear_error}>
                关闭
              </Button>
            </AlertDescription>
          </Alert>
        )}

        {/* Add comment form */}
        {allow_comments && currentUser && (
          <div className="pt-2">
            {showCommentForm ? (
              <CommentFormSOTA
                letter_id={profile_id}
                placeholder={`给 ${displayName} 留言...`}
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
        {/* Content tabs */}
        <Tabs value={activeTab} onValueChange={(value: any) => setActiveTab(value)} className="px-4">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="all" className="text-xs">全部</TabsTrigger>
            <TabsTrigger value="popular" className="text-xs">热门</TabsTrigger>
            <TabsTrigger value="recent" className="text-xs">最新</TabsTrigger>
          </TabsList>
        </Tabs>

        {/* Loading state */}
        {loading && filteredComments.length === 0 && (
          <div className="p-8 text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4" />
            <p className="text-muted-foreground">加载留言中...</p>
          </div>
        )}

        {/* Empty state */}
        {!loading && filteredComments.length === 0 && (
          <div className="p-8 text-center">
            <MessageSquare className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h4 className="font-medium mb-2">暂无留言</h4>
            <p className="text-sm text-muted-foreground mb-4">
              成为第一个给 {displayName} 留言的人
            </p>
            {allow_comments && currentUser && !showCommentForm && (
              <Button onClick={() => setShowCommentForm(true)}>
                写第一条留言
              </Button>
            )}
          </div>
        )}

        {/* Comment list */}
        {filteredComments.length > 0 && (
          <div className="divide-y">
            {filteredComments.map((comment, index) => (
              <React.Fragment key={comment.id}>
                <CommentItemSOTA
                  comment={comment}
                  depth={0}
                  max_depth={3}
                  enable_reply={allow_comments && !!currentUser}
                  enable_like={true}
                  enable_edit={currentUser?.id === comment.user_id}
                  enable_delete={currentUser?.id === comment.user_id}
                  enable_report={currentUser && currentUser.id !== comment.user_id}
                  show_replies={true}
                  on_action={handleCommentAction}
                />
                {index < filteredComments.length - 1 && <Separator />}
              </React.Fragment>
            ))}
          </div>
        )}

        {/* Load more button */}
        {has_more && filteredComments.length > 0 && (
          <div className="p-4 text-center border-t">
            <Button 
              variant="ghost" 
              size="sm"
              onClick={() => load_comments({ 
                page: Math.floor(filteredComments.length / max_display) + 1 
              })}
              disabled={loading}
            >
              {loading ? (
                <>
                  <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                  加载中...
                </>
              ) : (
                '加载更多留言'
              )}
            </Button>
          </div>
        )}

        {/* Real-time indicator */}
        {enable_real_time && (
          <div className="p-2 text-center text-xs text-muted-foreground border-t">
            <div className="flex items-center justify-center gap-2">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              实时更新
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default ProfileCommentsSOTA