'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useAuth } from '@/contexts/auth-context-new'
import { useLetterStore } from '@/stores/letter-store'
import { museumService } from '@/lib/services/museum-service'
import { toast } from '@/components/ui/use-toast'
import Link from 'next/link'
import { 
  Sparkles,
  Send,
  Inbox,
  Archive,
  Search,
  FileText,
  Star,
  Clock,
  User,
  Eye,
  ChevronRight,
  Mail
} from 'lucide-react'

interface LetterItem {
  id: string
  title: string
  content: string
  created_at: string
  status: 'draft' | 'sent' | 'received'
  code?: string
  sender?: string
}

export default function MuseumContributePage() {
  const { user } = useAuth()
  const { savedDrafts, sentLetters, receivedLetters } = useLetterStore()
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedLetter, setSelectedLetter] = useState<LetterItem | null>(null)

  const allContributableLetters = [
    // 草稿信件
    ...savedDrafts.map(draft => ({
      id: draft.id,
      title: draft.title || '无标题草稿',
      content: draft.content.substring(0, 100) + '...',
      created_at: draft.created_at.toISOString(),
      status: 'draft' as const,
      code: undefined,
      sender: undefined,
    })),
    // 已发送信件
    ...sentLetters.map(letter => ({
      id: letter.id,
      title: letter.title || '我的信件',
      content: letter.content.substring(0, 100) + '...',
      created_at: letter.created_at.toISOString(),
      status: 'sent' as const,
      code: letter.code?.code,
      sender: user?.nickname || '匿名用户'
    })),
    // 收到的信件
    ...receivedLetters.map(letter => ({
      id: letter.id,
      title: letter.title || '收到的信件',
      content: letter.content.substring(0, 100) + '...',
      created_at: letter.created_at.toISOString(),
      status: 'received' as const,
      code: letter.code?.code,
      sender: letter.sender_nickname || '匿名用户'
    }))
  ].filter(letter => 
    letter.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
    letter.content.toLowerCase().includes(searchTerm.toLowerCase())
  ).sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())

  const handleContributeLetter = async (letter: LetterItem) => {
    try {
      await museumService.submitToMuseum({
        letter_id: letter.id,
        title: letter.title,
        author_name: letter.sender || user?.nickname || '匿名用户',
        tags: []
      })
      toast({
        title: '贡献成功',
        description: `信件「${letter.title}」已成功贡献到博物馆！感谢您的分享。`
      })
    } catch (error: any) {
      console.error('Contribute letter error:', error)
      toast({
        title: '贡献失败',
        description: error?.message || '请稍后重试',
        variant: 'destructive'
      })
    }
  }

  const handleContributeNewNote = async (data: {
    title: string
    content: string 
    tags: string
    isHandwritten: boolean
    imageFile: File | null
  }) => {
    try {
      // 准备标签数组
      const tagsArray = data.tags
        .split(',')
        .map(tag => tag.trim())
        .filter(tag => tag.length > 0)

      // 如果有手写图片，先上传图片
      let imageUrl = null
      if (data.isHandwritten && data.imageFile) {
        const { imageUploadService } = await import('@/lib/services/image-upload-service')
        const uploadResult = await imageUploadService.uploadSingleImage(data.imageFile, {
          category: 'museum',
          isPublic: true,
          relatedType: 'museum_item'
        })
        imageUrl = uploadResult.url
      }

      // 创建博物馆内容
      const response = await museumService.createMuseumItem({
        title: data.title,
        content: data.content,
        author_name: user?.nickname || '匿名用户',
        description: data.isHandwritten ? '手写信件' : '信件笔记',
        image_url: imageUrl || undefined,
        source_type: 'direct',
        tags: tagsArray,
        metadata: {
          is_handwritten: data.isHandwritten
        }
      })

      toast({
        title: '贡献成功',
        description: `您的作品「${data.title}」已成功提交到博物馆！`,
      })

      // 表单重置将在子组件中处理
      return response.data
    } catch (error: any) {
      console.error('Contribute new note error:', error)
      toast({
        title: '贡献失败',
        description: error?.response?.data?.message || error?.message || '请稍后重试',
        variant: 'destructive'
      })
    }
  }

  const getStatusInfo = (status: string) => {
    switch (status) {
      case 'draft':
        return { label: '草稿', color: 'bg-gray-100 text-gray-800', icon: FileText }
      case 'sent':
        return { label: '已发送', color: 'bg-blue-100 text-blue-800', icon: Send }
      case 'received':
        return { label: '已收到', color: 'bg-green-100 text-green-800', icon: Inbox }
      default:
        return { label: '未知', color: 'bg-gray-100 text-gray-800', icon: Mail }
    }
  }

  return (
    <div className="container max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="font-serif text-3xl font-bold text-amber-900 mb-2">
          贡献作品到博物馆
        </h1>
        <p className="text-amber-700">
          分享你的珍贵信件，让更多人感受到文字的温度与力量
        </p>
      </div>

      {/* Search */}
      <div className="mb-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-amber-500" />
          <Input
            placeholder="搜索你的信件标题或内容..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10 border-amber-300 focus:border-amber-500"
          />
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        <Card className="border-amber-200 bg-gradient-to-br from-gray-50 to-gray-100">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">草稿信件</p>
                <p className="text-2xl font-bold text-gray-900">{savedDrafts.length}</p>
              </div>
              <FileText className="h-8 w-8 text-gray-400" />
            </div>
          </CardContent>
        </Card>
        <Card className="border-amber-200 bg-gradient-to-br from-blue-50 to-blue-100">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-blue-600">已发送信件</p>
                <p className="text-2xl font-bold text-blue-900">{sentLetters.length}</p>
              </div>
              <Send className="h-8 w-8 text-blue-400" />
            </div>
          </CardContent>
        </Card>
        <Card className="border-amber-200 bg-gradient-to-br from-green-50 to-green-100">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-green-600">收到的信件</p>
                <p className="text-2xl font-bold text-green-900">{receivedLetters.length}</p>
              </div>
              <Inbox className="h-8 w-8 text-green-400" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Action Tabs */}
      <Tabs defaultValue="existing" className="mb-8">
        <TabsList className="grid w-full grid-cols-2 bg-amber-50 border border-amber-200">
          <TabsTrigger value="existing" className="text-amber-700">
            <Archive className="w-4 h-4 mr-2" />
            选择已有信件
          </TabsTrigger>
          <TabsTrigger value="new" className="text-amber-700">
            <FileText className="w-4 h-4 mr-2" />
            写新信件笔记
          </TabsTrigger>
        </TabsList>
        
        <TabsContent value="existing">
          {/* Letter List */}
          {allContributableLetters.length > 0 ? (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {allContributableLetters.map((letter) => {
                const statusInfo = getStatusInfo(letter.status)
                return (
                  <Card key={letter.id} className="border-amber-200 hover:shadow-lg transition-shadow">
                    <CardHeader>
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <CardTitle className="font-serif text-lg text-amber-900 line-clamp-2">
                            {letter.title}
                          </CardTitle>
                          <div className="flex items-center gap-2 mt-2">
                            <Badge className={statusInfo.color}>
                              <statusInfo.icon className="w-3 h-3 mr-1" />
                              {statusInfo.label}
                            </Badge>
                            {letter.code && (
                              <Badge variant="outline" className="text-xs font-mono border-amber-300 text-amber-700">
                                {letter.code}
                              </Badge>
                            )}
                          </div>
                        </div>
                        <Button
                          onClick={() => handleContributeLetter(letter)}
                          size="sm"
                          className="bg-amber-600 hover:bg-amber-700 text-white whitespace-nowrap"
                        >
                          <Sparkles className="w-3 h-3 mr-1" />
                          贡献
                        </Button>
                      </div>
                    </CardHeader>
                    <CardContent className="space-y-3">
                      <p className="text-sm text-amber-700 line-clamp-3">
                        {letter.content}
                      </p>
                      
                      <div className="flex items-center justify-between text-xs text-amber-600 pt-3 border-t border-amber-200">
                        <div className="flex items-center gap-1">
                          <Clock className="w-3 h-3" />
                          <span>{new Date(letter.created_at).toLocaleDateString('zh-CN')}</span>
                        </div>
                        {letter.sender && (
                          <div className="flex items-center gap-1">
                            <User className="w-3 h-3" />
                            <span>{letter.sender}</span>
                          </div>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                )
              })}
            </div>
          ) : (
            <Card className="text-center py-12 border-amber-200">
              <CardContent>
                <div className="w-16 h-16 mx-auto bg-amber-100 rounded-full flex items-center justify-center mb-4">
                  <Archive className="h-8 w-8 text-amber-600" />
                </div>
                <h3 className="text-lg font-semibold mb-2 text-amber-900">暂无信件</h3>
                <p className="text-amber-700 mb-4">
                  你还没有可以贡献的信件。开始写信或等待收到信件后再来分享吧！
                </p>
                <Button asChild className="bg-amber-600 hover:bg-amber-700 text-white">
                  <Link href="/letters/write">
                    <Send className="w-4 h-4 mr-2" />
                    开始写信
                  </Link>
                </Button>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="new" className="space-y-6">
          <CreateNewNoteForm onSubmit={handleContributeNewNote} />
        </TabsContent>
      </Tabs>
    </div>
  )
}

// 新增信件笔记创建表单组件
function CreateNewNoteForm({ onSubmit }: { 
  onSubmit: (data: {
    title: string
    content: string 
    tags: string
    isHandwritten: boolean
    imageFile: File | null
  }) => Promise<any>
}) {
  const [formData, setFormData] = useState({
    title: '',
    content: '',
    tags: '',
    isHandwritten: false
  })
  const [imageFile, setImageFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setImageFile(file)
      const reader = new FileReader()
      reader.onload = (e) => setPreview(e.target?.result as string)
      reader.readAsDataURL(file)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (submitting) return
    
    setSubmitting(true)
    try {
      await onSubmit({ ...formData, imageFile })
      // 成功后重置表单
      setFormData({ title: '', content: '', tags: '', isHandwritten: false })
      setImageFile(null)
      setPreview(null)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card className="border-amber-200">
      <CardHeader>
        <CardTitle className="font-serif text-xl text-amber-900">
          创作新的信件笔记
        </CardTitle>
        <CardDescription className="text-amber-700">
          分享你的手写信件或创作感想，为博物馆增添新的珍藏
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* 标题 */}
          <div>
            <label className="block text-sm font-medium text-amber-900 mb-2">
              标题 *
            </label>
            <Input
              placeholder="为你的信件笔记起一个标题..."
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              className="border-amber-300 focus:border-amber-500"
              required
            />
          </div>

          {/* 内容 */}
          <div>
            <label className="block text-sm font-medium text-amber-900 mb-2">
              内容 *
            </label>
            <textarea
              placeholder="在这里写下你的信件内容、创作感想或相关故事..."
              value={formData.content}
              onChange={(e) => setFormData({ ...formData, content: e.target.value })}
              className="min-h-[200px] w-full px-3 py-2 border border-amber-300 rounded-md focus:outline-none focus:ring-2 focus:ring-amber-500 focus:border-amber-500 resize-y"
              required
            />
          </div>

          {/* 标签 */}
          <div>
            <label className="block text-sm font-medium text-amber-900 mb-2">
              标签
            </label>
            <Input
              placeholder="例如: 友情, 思念, 青春 (用逗号分隔)"
              value={formData.tags}
              onChange={(e) => setFormData({ ...formData, tags: e.target.value })}
              className="border-amber-300 focus:border-amber-500"
            />
          </div>

          {/* 手写信件上传 */}
          <div>
            <div className="flex items-center gap-2 mb-3">
              <input
                type="checkbox"
                id="handwritten"
                checked={formData.isHandwritten}
                onChange={(e) => setFormData({ ...formData, isHandwritten: e.target.checked })}
                className="text-amber-600 focus:ring-amber-500"
              />
              <label htmlFor="handwritten" className="text-sm font-medium text-amber-900">
                这是一封手写信件
              </label>
            </div>
            
            {formData.isHandwritten && (
              <div className="space-y-3">
                <label className="block text-sm font-medium text-amber-900">
                  上传手写信件照片
                </label>
                <input
                  type="file"
                  accept="image/*"
                  onChange={handleImageUpload}
                  className="block w-full text-sm text-amber-700 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-amber-50 file:text-amber-700 hover:file:bg-amber-100"
                />
                {preview && (
                  <div className="mt-3">
                    <img
                      src={preview}
                      alt="手写信件预览"
                      className="max-h-64 rounded-lg border border-amber-300"
                    />
                  </div>
                )}
              </div>
            )}
          </div>

          {/* 提交按钮 */}
          <div className="flex gap-3">
            <Button
              type="submit"
              className="bg-amber-600 hover:bg-amber-700 text-white"
              disabled={!formData.title || !formData.content || submitting}
            >
              <Sparkles className="w-4 h-4 mr-2" />
              {submitting ? '提交中...' : '贡献到博物馆'}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setFormData({ title: '', content: '', tags: '', isHandwritten: false })
                setImageFile(null)
                setPreview(null)
              }}
              className="border-amber-300 text-amber-700 hover:bg-amber-50"
            >
              重置表单
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}