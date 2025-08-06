'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  BookOpen,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  Eye,
  Heart,
  MessageSquare,
  Calendar,
  Edit,
  Trash2,
  Send,
  Archive
} from 'lucide-react'
import { museumService } from '@/lib/services/museum-service'
import { formatDate, formatRelativeTime } from '@/lib/utils'
import { useAuth } from '@/contexts/auth-context-new'
import { toast } from '@/components/ui/use-toast'

interface Submission {
  id: string
  letterId: string
  title: string
  excerpt: string
  status: 'pending' | 'approved' | 'rejected' | 'withdrawn'
  submitted_at: string
  reviewed_at?: string
  rejection_reason?: string
  moderation_notes?: string
  views: number
  likes: number
  comments: number
  is_featured: boolean
  exhibition_id?: string
  exhibition_name?: string
}

const statusConfig = {
  pending: {
    label: '审核中',
    icon: Clock,
    color: 'bg-yellow-100 text-yellow-800'
  },
  approved: {
    label: '已通过',
    icon: CheckCircle,
    color: 'bg-green-100 text-green-800'
  },
  rejected: {
    label: '未通过',
    icon: XCircle,
    color: 'bg-red-100 text-red-800'
  },
  withdrawn: {
    label: '已撤回',
    icon: Archive,
    color: 'bg-gray-100 text-gray-800'
  }
}

export default function MySubmissionsPage() {
  const router = useRouter()
  const { user } = useAuth()
  const [submissions, setSubmissions] = useState<Submission[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'all' | 'pending' | 'approved' | 'rejected'>('all')

  useEffect(() => {
    if (!user) {
      router.push('/auth/login?redirect=/museum/my-submissions')
      return
    }
    fetchSubmissions()
  }, [user])

  const fetchSubmissions = async () => {
    setLoading(true)
    setError(null)

    try {
      const response = await museumService.getMySubmissions()
      
      if (!response.data) {
        throw new Error('未找到提交数据')
      }

      // 模拟数据转换
      const formattedSubmissions: Submission[] = response.data.map((sub: any) => ({
        id: sub.id,
        letterId: sub.letterId,
        title: sub.title,
        excerpt: sub.content?.substring(0, 150) + '...' || '',
        status: sub.status || 'pending',
        submitted_at: sub.createdAt,
        reviewed_at: sub.reviewed_at,
        rejection_reason: sub.rejection_reason,
        moderation_notes: sub.moderation_notes,
        views: Math.floor(Math.random() * 1000),
        likes: Math.floor(Math.random() * 100),
        comments: Math.floor(Math.random() * 50),
        is_featured: sub.is_featured || false,
        exhibition_id: sub.exhibition_id,
        exhibition_name: sub.exhibition_name
      }))

      setSubmissions(formattedSubmissions)
    } catch (err) {
      console.error('获取提交记录失败:', err)
      setError('获取提交记录失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  const handleWithdraw = async (id: string) => {
    if (!confirm('确定要撤回这个提交吗？撤回后将无法恢复。')) {
      return
    }

    try {
      await museumService.withdrawMuseumEntry(id)
      
      // 更新本地状态
      setSubmissions(prev => 
        prev.map(sub => 
          sub.id === id ? { ...sub, status: 'withdrawn' as const } : sub
        )
      )
      
      toast({
        title: '撤回成功',
        description: '您的提交已被撤回'
      })
    } catch (err) {
      toast({
        title: '撤回失败',
        description: '请稍后重试',
        variant: 'destructive'
      })
    }
  }

  const filteredSubmissions = submissions.filter(sub => {
    if (activeTab === 'all') return true
    return sub.status === activeTab
  })

  const stats = {
    total: submissions.length,
    pending: submissions.filter(s => s.status === 'pending').length,
    approved: submissions.filter(s => s.status === 'approved').length,
    rejected: submissions.filter(s => s.status === 'rejected').length,
    featured: submissions.filter(s => s.is_featured).length
  }

  if (!user) {
    return null
  }

  return (
    <div className="container max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-letter-ink mb-2">
          我的博物馆提交
        </h1>
        <p className="text-muted-foreground">
          管理您提交到信件博物馆的作品
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">总提交</p>
                <p className="text-2xl font-bold">{stats.total}</p>
              </div>
              <BookOpen className="w-8 h-8 text-muted-foreground opacity-20" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">审核中</p>
                <p className="text-2xl font-bold text-yellow-600">{stats.pending}</p>
              </div>
              <Clock className="w-8 h-8 text-yellow-600 opacity-20" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">已通过</p>
                <p className="text-2xl font-bold text-green-600">{stats.approved}</p>
              </div>
              <CheckCircle className="w-8 h-8 text-green-600 opacity-20" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">精选作品</p>
                <p className="text-2xl font-bold text-amber-600">{stats.featured}</p>
              </div>
              <Star className="w-8 h-8 text-amber-600 opacity-20" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as any)} className="mb-6">
        <TabsList className="grid grid-cols-4 w-full max-w-md">
          <TabsTrigger value="all">
            全部 ({stats.total})
          </TabsTrigger>
          <TabsTrigger value="pending">
            审核中 ({stats.pending})
          </TabsTrigger>
          <TabsTrigger value="approved">
            已通过 ({stats.approved})
          </TabsTrigger>
          <TabsTrigger value="rejected">
            未通过 ({stats.rejected})
          </TabsTrigger>
        </TabsList>
      </Tabs>

      {/* Error State */}
      {error && (
        <Alert variant="destructive" className="mb-6">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Loading State */}
      {loading && (
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader>
                <div className="h-6 bg-muted rounded w-3/4"></div>
                <div className="h-4 bg-muted rounded w-1/2 mt-2"></div>
              </CardHeader>
              <CardContent>
                <div className="h-16 bg-muted rounded"></div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Submissions List */}
      {!loading && filteredSubmissions.length > 0 && (
        <div className="space-y-4">
          {filteredSubmissions.map(submission => {
            const statusInfo = statusConfig[submission.status]
            const StatusIcon = statusInfo.icon
            
            return (
              <Card key={submission.id}>
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <CardTitle className="text-lg line-clamp-1">
                        {submission.title}
                      </CardTitle>
                      <CardDescription className="flex items-center gap-3 mt-1">
                        <span className="flex items-center gap-1">
                          <Calendar className="w-3 h-3" />
                          提交于 {formatDate(submission.submitted_at)}
                        </span>
                        {submission.reviewed_at && (
                          <span className="text-xs">
                            审核于 {formatDate(submission.reviewed_at)}
                          </span>
                        )}
                      </CardDescription>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge className={statusInfo.color}>
                        <StatusIcon className="w-3 h-3 mr-1" />
                        {statusInfo.label}
                      </Badge>
                      {submission.is_featured && (
                        <Badge variant="secondary" className="bg-yellow-100 text-yellow-800">
                          精选
                        </Badge>
                      )}
                    </div>
                  </div>
                </CardHeader>
                
                <CardContent>
                  <p className="text-sm text-muted-foreground line-clamp-2 mb-4">
                    {submission.excerpt}
                  </p>
                  
                  {/* Rejection Reason */}
                  {submission.status === 'rejected' && submission.rejection_reason && (
                    <Alert variant="destructive" className="mb-4">
                      <AlertCircle className="h-4 w-4" />
                      <AlertDescription>
                        <strong>未通过原因：</strong> {submission.rejection_reason}
                      </AlertDescription>
                    </Alert>
                  )}
                  
                  {/* Exhibition Info */}
                  {submission.exhibition_name && (
                    <div className="mb-4 p-3 bg-muted rounded-lg">
                      <p className="text-sm">
                        收录于展览：<strong>{submission.exhibition_name}</strong>
                      </p>
                    </div>
                  )}
                  
                  {/* Stats (for approved submissions) */}
                  {submission.status === 'approved' && (
                    <div className="flex items-center gap-4 text-sm text-muted-foreground mb-4">
                      <span className="flex items-center gap-1">
                        <Eye className="w-4 h-4" />
                        {submission.views} 浏览
                      </span>
                      <span className="flex items-center gap-1">
                        <Heart className="w-4 h-4" />
                        {submission.likes} 喜欢
                      </span>
                      <span className="flex items-center gap-1">
                        <MessageSquare className="w-4 h-4" />
                        {submission.comments} 评论
                      </span>
                    </div>
                  )}
                  
                  {/* Actions */}
                  <div className="flex items-center gap-2">
                    {submission.status === 'approved' && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => router.push(`/museum/entries/${submission.id}`)}
                      >
                        <Eye className="w-4 h-4 mr-2" />
                        查看
                      </Button>
                    )}
                    
                    {submission.status === 'pending' && (
                      <>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleWithdraw(submission.id)}
                        >
                          <Archive className="w-4 h-4 mr-2" />
                          撤回
                        </Button>
                      </>
                    )}
                    
                    {submission.status === 'rejected' && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => router.push(`/write?edit=${submission.letterId}`)}
                      >
                        <Edit className="w-4 h-4 mr-2" />
                        修改后重新提交
                      </Button>
                    )}
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}

      {/* Empty State */}
      {!loading && filteredSubmissions.length === 0 && (
        <Card className="text-center py-12">
          <CardContent>
            <BookOpen className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
            <p className="text-muted-foreground mb-4">
              {activeTab === 'all' 
                ? '您还没有提交任何信件到博物馆'
                : `没有${statusConfig[activeTab as keyof typeof statusConfig]?.label || activeTab}的提交`
              }
            </p>
            {activeTab === 'all' && (
              <Button onClick={() => router.push('/letters')}>
                <Send className="w-4 h-4 mr-2" />
                去提交信件
              </Button>
            )}
          </CardContent>
        </Card>
      )}

      {/* Tips */}
      <Card className="mt-8 bg-amber-50 border-amber-200">
        <CardHeader>
          <CardTitle className="text-base">提交小贴士</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground space-y-2">
          <p>• 提交的信件需要经过审核，通常在 1-3 个工作日内完成</p>
          <p>• 内容积极向上、文字优美的信件更容易通过审核</p>
          <p>• 被选为精选作品的信件会在首页展示</p>
          <p>• 审核未通过的信件可以修改后重新提交</p>
        </CardContent>
      </Card>
    </div>
  )
}

// 添加 Star import
import { Star } from 'lucide-react'