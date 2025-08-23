'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Heart,
  Eye,
  MessageSquare,
  Share2,
  Bookmark,
  Calendar,
  User,
  Clock,
  Flag,
  ArrowLeft
} from 'lucide-react'
import { museumService, type MuseumEntry } from '@/lib/services/museum-service'
import { formatDate, formatRelativeTime } from '@/lib/utils'
import { useAuth } from '@/contexts/auth-context-new'
import { toast } from '@/components/ui/use-toast'
import { SafeTimestamp } from '@/components/ui/safe-timestamp'
import { CommentSystemSOTA } from '@/components/comments/comment-system-sota'

interface ExtendedMuseumEntry extends MuseumEntry {
  views: number
  likes: number
  shares: number
  comments: number
  theme: string
  tags: string[]
  user_reactions?: {
    liked: boolean
    bookmarked: boolean
    shared: boolean
  }
}


export default function MuseumEntryDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { user } = useAuth()
  const [entry, setEntry] = useState<ExtendedMuseumEntry | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [liked, setLiked] = useState(false)
  const [bookmarked, setBookmarked] = useState(false)

  useEffect(() => {
    if (params?.id) {
      fetchEntry(params.id as string)
      // 记录浏览
      recordInteraction('view')
    }
  }, [params?.id])

  const fetchEntry = async (id: string) => {
    setLoading(true)
    setError(null)

    try {
      const response = await museumService.getEntry(id)
      
      if (!response.data) {
        throw new Error('未找到信件数据')
      }

      // 模拟数据增强
      const enhancedEntry: ExtendedMuseumEntry = {
        ...response.data,
        views: response.data.viewCount || Math.floor(Math.random() * 5000) + 1000,
        likes: response.data.likeCount || Math.floor(Math.random() * 500) + 100,
        shares: Math.floor(Math.random() * 200) + 50,
        comments: 0, // Will be updated by CommentStats component
        theme: response.data.theme || '青春回忆',
        tags: response.data.tags || ['感动', '青春', '回忆'],
        user_reactions: {
          liked: false,
          bookmarked: false,
          shared: false
        }
      }

      setEntry(enhancedEntry)
      setLiked(enhancedEntry.user_reactions?.liked || false)
      setBookmarked(enhancedEntry.user_reactions?.bookmarked || false)

    } catch (err) {
      console.error('获取条目失败:', err)
      setError('获取信件详情失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }


  const recordInteraction = async (type: 'view' | 'like' | 'share' | 'bookmark') => {
    if (!params?.id || !user) return

    try {
      await museumService.interactWithEntry(params.id as string, { type })
    } catch (err) {
      console.error('记录互动失败:', err)
    }
  }

  const handleLike = async () => {
    if (!user) {
      toast({
        title: '请先登录',
        description: '登录后才能点赞',
        variant: 'destructive'
      })
      return
    }

    setLiked(!liked)
    if (!liked) {
      recordInteraction('like')
      setEntry(prev => prev ? { ...prev, likes: prev.likes + 1 } : prev)
    } else {
      setEntry(prev => prev ? { ...prev, likes: prev.likes - 1 } : prev)
    }
  }

  const handleBookmark = async () => {
    if (!user) {
      toast({
        title: '请先登录',
        description: '登录后才能收藏',
        variant: 'destructive'
      })
      return
    }

    setBookmarked(!bookmarked)
    if (!bookmarked) {
      recordInteraction('bookmark')
      toast({
        title: '收藏成功',
        description: '已添加到您的收藏夹'
      })
    } else {
      toast({
        title: '取消收藏',
        description: '已从收藏夹移除'
      })
    }
  }

  const handleShare = async () => {
    recordInteraction('share')
    
    if (navigator.share) {
      try {
        await navigator.share({
          title: entry?.title,
          text: entry?.content.substring(0, 100) + '...',
          url: window.location.href
        })
        setEntry(prev => prev ? { ...prev, shares: prev.shares + 1 } : prev)
      } catch (err) {
        console.error('分享失败:', err)
      }
    } else {
      // 复制链接到剪贴板
      navigator.clipboard.writeText(window.location.href)
      toast({
        title: '链接已复制',
        description: '您可以将链接分享给朋友'
      })
      setEntry(prev => prev ? { ...prev, shares: prev.shares + 1 } : prev)
    }
  }


  if (loading) {
    return (
      <div className="container max-w-4xl mx-auto px-4 py-8">
        <Card className="animate-pulse">
          <CardHeader>
            <div className="h-8 bg-muted rounded w-3/4 mb-4"></div>
            <div className="h-4 bg-muted rounded w-1/2"></div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="h-4 bg-muted rounded"></div>
              <div className="h-4 bg-muted rounded"></div>
              <div className="h-4 bg-muted rounded w-4/5"></div>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (error || !entry) {
    return (
      <div className="container max-w-4xl mx-auto px-4 py-8">
        <Alert variant="destructive">
          <AlertDescription>{error || '信件不存在'}</AlertDescription>
        </Alert>
        <Button onClick={() => router.back()} className="mt-4">
          <ArrowLeft className="w-4 h-4 mr-2" />
          返回
        </Button>
      </div>
    )
  }

  return (
    <div className="container max-w-4xl mx-auto px-4 py-8">
      {/* Back Button */}
      <Button variant="ghost" onClick={() => router.back()} className="mb-6">
        <ArrowLeft className="w-4 h-4 mr-2" />
        返回博物馆
      </Button>

      {/* Main Content */}
      <Card className="mb-8">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <CardTitle className="font-serif text-2xl mb-2">
                {entry.title}
              </CardTitle>
              <CardDescription className="flex items-center gap-4">
                <span className="flex items-center gap-1">
                  <User className="w-4 h-4" />
                  {entry.author_name}
                </span>
                <span className="flex items-center gap-1">
                  <Calendar className="w-4 h-4" />
                  {formatDate(entry.createdAt)}
                </span>
                <span className="flex items-center gap-1">
                  <Clock className="w-4 h-4" />
                  {formatRelativeTime(entry.createdAt)}
                </span>
              </CardDescription>
            </div>
            {entry.is_featured && (
              <Badge variant="secondary" className="bg-yellow-100 text-yellow-800">
                精选
              </Badge>
            )}
          </div>

          {/* Tags */}
          <div className="flex flex-wrap gap-2 mt-4">
            {entry.tags.map(tag => (
              <Badge key={tag} variant="outline">
                #{tag}
              </Badge>
            ))}
          </div>
        </CardHeader>

        <CardContent>
          {/* Letter Content */}
          <div className="prose prose-gray max-w-none mb-8">
            <p className="whitespace-pre-wrap text-gray-700 leading-relaxed">
              {entry.content}
            </p>
          </div>

          {/* Stats Bar */}
          <div className="flex items-center justify-between py-4 border-t border-b">
            <div className="flex items-center gap-6 text-sm text-muted-foreground">
              <span className="flex items-center gap-1">
                <Eye className="w-4 h-4" />
                {entry.views.toLocaleString()} 浏览
              </span>
              <span className="flex items-center gap-1">
                <Heart className="w-4 h-4" />
                {entry.likes.toLocaleString()} 喜欢
              </span>
              {/* Comment count will be displayed by CommentStats component */}
              <span className="flex items-center gap-1">
                <Share2 className="w-4 h-4" />
                {entry.shares} 分享
              </span>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex items-center gap-2 mt-6">
            <Button
              variant={liked ? 'default' : 'outline'}
              size="sm"
              onClick={handleLike}
            >
              <Heart className={`w-4 h-4 mr-2 ${liked ? 'fill-current' : ''}`} />
              {liked ? '已喜欢' : '喜欢'}
            </Button>
            <Button
              variant={bookmarked ? 'default' : 'outline'}
              size="sm"
              onClick={handleBookmark}
            >
              <Bookmark className={`w-4 h-4 mr-2 ${bookmarked ? 'fill-current' : ''}`} />
              {bookmarked ? '已收藏' : '收藏'}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleShare}
            >
              <Share2 className="w-4 h-4 mr-2" />
              分享
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="ml-auto text-muted-foreground"
            >
              <Flag className="w-4 h-4 mr-2" />
              举报
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Comments Section */}
      <CommentSystemSOTA
        targetId={entry.id}
        targetType="museum"
        title="评论"
        placeholder="写下您的感想..."
        enableReplies={true}
        maxDepth={3}
        showStats={true}
        onCommentCreated={() => {
          // Optionally refresh entry stats
          setEntry(prev => prev ? { ...prev, comments: prev.comments + 1 } : prev)
        }}
      />
    </div>
  )
}