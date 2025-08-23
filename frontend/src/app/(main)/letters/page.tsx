'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import {
  Mail,
  Send,
  Archive,
  Star,
  Search,
  Filter,
  Plus,
  Clock,
  CheckCircle,
  AlertCircle,
  Package,
  MoreVertical,
  Eye,
  Trash2
} from 'lucide-react'
import { BackButton } from '@/components/ui/back-button'
import { useAuth } from '@/contexts/auth-context-new'
import { apiClient } from '@/lib/api-client'
import { formatDate } from '@/lib/utils'

interface Letter {
  id: string
  title: string
  content: string
  sender_id: string
  sender_name: string
  recipient_id?: string
  recipient_name?: string
  status: 'draft' | 'sent' | 'delivered' | 'read'
  letter_code?: string
  recipient_op_code?: string
  created_at: string
  sent_at?: string
  delivered_at?: string
  read_at?: string
  is_starred: boolean
  is_archived: boolean
  style?: string
  type: 'regular' | 'drift' | 'anonymous'
}

export default function LettersPage() {
  const router = useRouter()
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState('inbox')
  const [letters, setLetters] = useState<Letter[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedStatus, setSelectedStatus] = useState('all')

  useEffect(() => {
    if (!user) {
      router.push('/login')
      return
    }
    loadLetters()
  }, [user, activeTab])

  const loadLetters = async () => {
    setLoading(true)
    try {
      let endpoint = ''
      switch (activeTab) {
        case 'inbox':
          endpoint = '/letters/inbox'
          break
        case 'sent':
          endpoint = '/letters/sent'
          break
        case 'drafts':
          endpoint = '/letters/drafts'
          break
        case 'starred':
          endpoint = '/letters/starred'
          break
        case 'archived':
          endpoint = '/letters/archived'
          break
      }

      const response = await apiClient.get<any>(endpoint)
      const data = response.data
      const letters = Array.isArray(data) ? data : (data && typeof data === 'object' && 'letters' in data ? data.letters : [])
      setLetters(letters || [])
    } catch (error) {
      console.error('Failed to load letters:', error)
      setLetters([])
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteLetter = async (letterId: string) => {
    if (!confirm('确定要删除这封信吗？')) return

    try {
      await apiClient.delete(`/letters/${letterId}`)
      setLetters(letters.filter(l => l.id !== letterId))
    } catch (error) {
      console.error('Failed to delete letter:', error)
      alert('删除失败')
    }
  }

  const handleStarLetter = async (letterId: string, isStarred: boolean) => {
    try {
      await apiClient.put(`/letters/${letterId}/star`, {
        is_starred: !isStarred
      })
      setLetters(letters.map(l => 
        l.id === letterId ? { ...l, is_starred: !isStarred } : l
      ))
    } catch (error) {
      console.error('Failed to star letter:', error)
    }
  }

  const handleArchiveLetter = async (letterId: string) => {
    try {
      await apiClient.put(`/letters/${letterId}/archive`)
      setLetters(letters.filter(l => l.id !== letterId))
    } catch (error) {
      console.error('Failed to archive letter:', error)
    }
  }

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      draft: { variant: 'secondary' as const, label: '草稿', icon: Clock },
      sent: { variant: 'default' as const, label: '已发送', icon: Send },
      delivered: { variant: 'outline' as const, label: '已送达', icon: Package },
      read: { variant: 'default' as const, label: '已读', icon: CheckCircle }
    }

    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.draft
    const Icon = config.icon

    return (
      <Badge variant={config.variant} className="flex items-center gap-1">
        <Icon className="h-3 w-3" />
        {config.label}
      </Badge>
    )
  }

  const filteredLetters = letters.filter(letter => {
    if (searchTerm && !letter.title.toLowerCase().includes(searchTerm.toLowerCase()) &&
        !letter.content.toLowerCase().includes(searchTerm.toLowerCase())) {
      return false
    }
    
    if (selectedStatus !== 'all' && letter.status !== selectedStatus) {
      return false
    }
    
    return true
  })

  const tabConfig = [
    { value: 'inbox', label: '收件箱', icon: Mail },
    { value: 'sent', label: '已发送', icon: Send },
    { value: 'drafts', label: '草稿箱', icon: Clock },
    { value: 'starred', label: '星标', icon: Star },
    { value: 'archived', label: '归档', icon: Archive }
  ]

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <BackButton />
        <div className="flex items-center justify-between mt-4">
          <div>
            <h1 className="text-3xl font-bold">我的信件</h1>
            <p className="text-gray-600 mt-2">
              管理您的所有信件
            </p>
          </div>
          
          <Button onClick={() => router.push('/letters/write')}>
            <Plus className="h-4 w-4 mr-2" />
            写信
          </Button>
        </div>
      </div>

      {/* 搜索和筛选 */}
      <div className="mb-6 flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
          <Input
            placeholder="搜索信件..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
        
        <select
          value={selectedStatus}
          onChange={(e) => setSelectedStatus(e.target.value)}
          className="px-4 py-2 border rounded-md"
        >
          <option value="all">全部状态</option>
          <option value="draft">草稿</option>
          <option value="sent">已发送</option>
          <option value="delivered">已送达</option>
          <option value="read">已读</option>
        </select>
      </div>

      {/* 标签页 */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-5">
          {tabConfig.map(tab => {
            const Icon = tab.icon
            return (
              <TabsTrigger key={tab.value} value={tab.value} className="flex items-center gap-2">
                <Icon className="h-4 w-4" />
                <span className="hidden sm:inline">{tab.label}</span>
              </TabsTrigger>
            )
          })}
        </TabsList>

        {tabConfig.map(tab => (
          <TabsContent key={tab.value} value={tab.value}>
            {loading ? (
              <div className="space-y-4">
                {[...Array(5)].map((_, i) => (
                  <Card key={i} className="animate-pulse">
                    <CardContent className="p-6">
                      <div className="h-4 bg-gray-200 rounded w-1/3 mb-3"></div>
                      <div className="h-3 bg-gray-200 rounded w-full mb-2"></div>
                      <div className="h-3 bg-gray-200 rounded w-2/3"></div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : filteredLetters.length === 0 ? (
              <Card>
                <CardContent className="text-center py-12">
                  <Mail className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-600">暂无信件</p>
                </CardContent>
              </Card>
            ) : (
              <div className="space-y-4">
                {filteredLetters.map(letter => (
                  <Card 
                    key={letter.id}
                    className="hover:shadow-md transition-shadow cursor-pointer"
                    onClick={() => router.push(`/letters/read/${letter.letter_code || letter.id}`)}
                  >
                    <CardContent className="p-6">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center gap-3 mb-2">
                            <h3 className="font-semibold text-lg">{letter.title}</h3>
                            {getStatusBadge(letter.status)}
                            {letter.is_starred && (
                              <Star className="h-4 w-4 text-yellow-500 fill-current" />
                            )}
                          </div>
                          
                          <p className="text-sm text-gray-600 mb-2 line-clamp-2">
                            {letter.content}
                          </p>
                          
                          <div className="flex items-center gap-4 text-sm text-gray-500">
                            {activeTab === 'inbox' ? (
                              <>
                                <span>来自: {letter.sender_name || '匿名'}</span>
                                {letter.recipient_op_code && (
                                  <span>• 投递至: {letter.recipient_op_code}</span>
                                )}
                              </>
                            ) : (
                              <>
                                <span>收件人: {letter.recipient_name || '漂流瓶'}</span>
                                {letter.recipient_op_code && (
                                  <span>• 投递至: {letter.recipient_op_code}</span>
                                )}
                              </>
                            )}
                            <span>• {formatDate(letter.created_at)}</span>
                          </div>
                        </div>

                        <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleStarLetter(letter.id, letter.is_starred)}
                          >
                            <Star className={`h-4 w-4 ${letter.is_starred ? 'text-yellow-500 fill-current' : ''}`} />
                          </Button>
                          
                          {activeTab !== 'archived' && (
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleArchiveLetter(letter.id)}
                            >
                              <Archive className="h-4 w-4" />
                            </Button>
                          )}
                          
                          {activeTab === 'drafts' && (
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDeleteLetter(letter.id)}
                            >
                              <Trash2 className="h-4 w-4 text-red-600" />
                            </Button>
                          )}
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </TabsContent>
        ))}
      </Tabs>
    </div>
  )
}