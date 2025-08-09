'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { 
  Inbox, 
  Send, 
  FileText, 
  Search, 
  Mail,
  Calendar,
  Eye,
  Reply,
  RefreshCw
} from 'lucide-react'
import { formatRelativeTime, getLetterStatusText, getLetterStatusColor } from '@/lib/utils'
import type { LetterStatus } from '@/types/letter'
import { LetterService, type Letter } from '@/lib/services/letter-service'
import { useAuth } from '@/contexts/auth-context'

export default function MailboxPage() {
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState<'sent' | 'received' | 'drafts'>('sent')
  const [searchQuery, setSearchQuery] = useState('')
  const [filterStatus, setFilterStatus] = useState<LetterStatus | 'all'>('all')
  const [letters, setLetters] = useState<Letter[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [total, setTotal] = useState(0)

  // 加载信件数据
  const loadLetters = async () => {
    if (!user) return
    
    setLoading(true)
    setError(null)
    
    try {
      if (activeTab === 'drafts') {
        // 获取草稿
        const response = await LetterService.getDrafts({
          page: 1,
          limit: 50,
          sort_by: 'updated_at',
          sort_order: 'desc'
        })
        if (response.success && response.data) {
          setLetters(response.data.drafts)
          setTotal(response.data.total)
        }
      } else {
        // 获取发送或接收的信件
        const response = await LetterService.getUserLetters({
          type: activeTab,
          page: 1,
          limit: 50,
          search: searchQuery || undefined,
          status: filterStatus !== 'all' ? filterStatus : undefined,
          sort_by: 'created_at',
          sort_order: 'desc'
        })
        if (response.success && response.data) {
          setLetters(response.data.letters)
          setTotal(response.data.total)
        }
      }
    } catch (err) {
      console.error('Failed to load letters:', err)
      setError('加载信件失败，请刷新重试')
      setLetters([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }

  // 当用户登录状态或筛选条件改变时重新加载
  useEffect(() => {
    loadLetters()
  }, [user, activeTab, searchQuery, filterStatus])

  const tabs = [
    { id: 'sent', label: '已发送', icon: Send, count: activeTab === 'sent' ? total : 0 },
    { id: 'received', label: '已收到', icon: Inbox, count: activeTab === 'received' ? total : 0 },
    { id: 'drafts', label: '草稿箱', icon: FileText, count: activeTab === 'drafts' ? total : 0 },
  ]

  const statusOptions = [
    { value: 'all', label: '全部状态' },
    { value: 'draft', label: '草稿' },
    { value: 'generated', label: '已生成编号' },
    { value: 'collected', label: '已收取' },
    { value: 'in_transit', label: '在途中' },
    { value: 'delivered', label: '已送达' },
    { value: 'read', label: '已查看' },
  ]

  const renderLetterCard = (letter: Letter) => (
    <Card key={letter.id} className="hover:shadow-md transition-shadow">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-lg font-serif flex items-center gap-2">
              {letter.title || '无标题信件'}
              {activeTab === 'received' && letter.status !== 'read' && (
                <div className="h-2 w-2 rounded-full bg-primary" />
              )}
            </CardTitle>
            <CardDescription className="mt-1">
              {activeTab === 'sent' && (
                <>
                  {letter.recipient_info?.name ? `收件人: ${letter.recipient_info.name}` : '收件人: 未设置'} • 
                  {letter.code ? `编号: ${letter.code}` : '未生成编号'}
                </>
              )}
              {activeTab === 'received' && (
                <>来自: {letter.sender_name || '匿名'}</>
              )}
              {activeTab === 'drafts' && (
                <>最后修改: {formatRelativeTime(new Date(letter.updated_at))}</>
              )}
            </CardDescription>
          </div>
          {activeTab === 'sent' && (
            <Badge variant={getLetterStatusColor(letter.status as LetterStatus) as any}>
              {getLetterStatusText(letter.status as LetterStatus)}
            </Badge>
          )}
          {activeTab === 'drafts' && (
            <Badge variant="secondary">草稿</Badge>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
          {letter.content}
        </p>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4 text-xs text-muted-foreground">
            <span className="flex items-center gap-1">
              <Calendar className="h-3 w-3" />
              {activeTab === 'received' 
                ? formatRelativeTime(new Date(letter.delivered_at || letter.created_at))
                : formatRelativeTime(new Date(letter.created_at))
              }
            </span>
          </div>
          <div className="flex items-center gap-2">
            {activeTab === 'sent' && (
              <Button variant="ghost" size="sm">
                <Eye className="h-4 w-4 mr-2" />
                查看详情
              </Button>
            )}
            {activeTab === 'received' && (
              <>
                <Button variant="ghost" size="sm">
                  <Eye className="h-4 w-4 mr-2" />
                  阅读
                </Button>
                <Button variant="outline" size="sm">
                  <Reply className="h-4 w-4 mr-2" />
                  回信
                </Button>
              </>
            )}
            {activeTab === 'drafts' && (
              <>
                <Button variant="outline" size="sm">
                  继续编辑
                </Button>
                <Button variant="ghost" size="sm" className="text-destructive">
                  删除
                </Button>
              </>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )

  return (
    <div className="container max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-letter-ink mb-2">
          我的信箱
        </h1>
        <p className="text-muted-foreground">
          管理你的信件，查看发送和接收记录
        </p>
      </div>

      {/* Tabs */}
      <div className="flex flex-wrap gap-2 mb-6 border-b">
        {tabs.map((tab) => {
          const Icon = tab.icon
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`flex items-center gap-2 px-4 py-2 rounded-t-lg font-medium transition-colors ${
                activeTab === tab.id
                  ? 'bg-background border-b-2 border-primary text-primary'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              <Icon className="h-4 w-4" />
              {tab.label}
              {tab.count > 0 && (
                <Badge variant="secondary" className="text-xs">
                  {tab.count}
                </Badge>
              )}
            </button>
          )
        })}
      </div>

      {/* Search and Filter */}
      <div className="flex flex-col sm:flex-row gap-4 mb-6">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="搜索信件标题或内容..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        {activeTab === 'sent' && (
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value as any)}
            className="px-3 py-2 border border-input rounded-md bg-background text-sm"
          >
            {statusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        )}
        <Button variant="outline" onClick={loadLetters} disabled={loading}>
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          刷新
        </Button>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          <span className="ml-2 text-muted-foreground">加载中...</span>
        </div>
      )}

      {/* Error State */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <div className="text-center">
              <p className="text-destructive mb-4">{error}</p>
              <Button onClick={loadLetters} variant="outline">
                重新加载
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Content */}
      {!loading && !error && (
        <div className="space-y-4">
          {letters.length > 0 ? (
            letters.map(renderLetterCard)
          ) : (
            /* 空状态 */
            <Card className="text-center py-12">
              <CardContent>
                <div className="flex flex-col items-center gap-4">
                  <div className="h-16 w-16 rounded-full bg-muted flex items-center justify-center">
                    <Mail className="h-8 w-8 text-muted-foreground" />
                  </div>
                  <div>
                    <h3 className="font-semibold mb-2">
                      {activeTab === 'sent' && '还没有发送过信件'}
                      {activeTab === 'received' && '还没有收到信件'}
                      {activeTab === 'drafts' && '没有保存的草稿'}
                    </h3>
                    <p className="text-muted-foreground mb-4">
                      {activeTab === 'sent' && '写下你的第一封信，开始温暖的交流吧'}
                      {activeTab === 'received' && '当有人给你写信时，它们会出现在这里'}
                      {activeTab === 'drafts' && '在写信页面保存的草稿会出现在这里'}
                    </p>
                    {activeTab !== 'received' && (
                      <Button>
                        <Mail className="h-4 w-4 mr-2" />
                        开始写信
                      </Button>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}
    </div>
  )
}