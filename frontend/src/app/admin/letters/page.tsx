'use client'

import React, { useState, useEffect } from 'react'
import { 
  Mail, 
  Search, 
  Filter, 
  Eye, 
  Download,
  Flag,
  Truck,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Clock,
  MoreVertical,
  MapPin,
  User,
  Calendar
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { usePermission, PERMISSIONS } from '@/hooks/use-permission'

interface Letter {
  id: string
  title: string
  sender: {
    id: string
    username: string
    nickname: string
    avatar?: string
    school_name: string
  }
  recipient?: {
    id: string
    username: string
    nickname: string
    school_name: string
  }
  status: 'draft' | 'generated' | 'collected' | 'in_transit' | 'delivered' | 'failed'
  priority: 'normal' | 'high' | 'urgent'
  content_preview: string
  word_count: number
  created_at: string
  updated_at: string
  delivered_at?: string
  courier?: {
    id: string
    name: string
  }
  tracking_code?: string
  delivery_address?: string
  flags: string[]
}

interface LetterStats {
  total_letters: number
  pending_letters: number
  in_transit_letters: number
  delivered_letters: number
  failed_letters: number
  today_letters: number
  this_month_letters: number
}

const STATUS_COLORS: Record<string, string> = {
  'draft': 'bg-gray-100 text-gray-800',
  'generated': 'bg-blue-100 text-blue-800',
  'collected': 'bg-yellow-100 text-yellow-800',
  'in_transit': 'bg-orange-100 text-orange-800',
  'delivered': 'bg-green-100 text-green-800',
  'failed': 'bg-red-100 text-red-800'
}

const STATUS_NAMES: Record<string, string> = {
  'draft': '草稿',
  'generated': '已生成',
  'collected': '已收集',
  'in_transit': '运输中',
  'delivered': '已送达',
  'failed': '失败'
}

const STATUS_ICONS: Record<string, React.ReactNode> = {
  'draft': <Clock className="w-3 h-3" />,
  'generated': <Mail className="w-3 h-3" />,
  'collected': <CheckCircle className="w-3 h-3" />,
  'in_transit': <Truck className="w-3 h-3" />,
  'delivered': <CheckCircle className="w-3 h-3" />,
  'failed': <XCircle className="w-3 h-3" />
}

const PRIORITY_COLORS: Record<string, string> = {
  'normal': 'bg-gray-100 text-gray-800',
  'high': 'bg-yellow-100 text-yellow-800',
  'urgent': 'bg-red-100 text-red-800'
}

export default function LettersManagePage() {
  const { user, hasPermission } = usePermission()
  const [letters, setLetters] = useState<Letter[]>([])
  const [stats, setStats] = useState<LetterStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [priorityFilter, setPriorityFilter] = useState<string>('all')
  const [schoolFilter, setSchoolFilter] = useState<string>('all')
  const [selectedLetter, setSelectedLetter] = useState<Letter | null>(null)
  const [showLetterDetail, setShowLetterDetail] = useState(false)
  const [currentTab, setCurrentTab] = useState('all')

  // 权限检查
  if (!user || !hasPermission(PERMISSIONS.VIEW_REPORTS)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Mail className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">访问权限不足</h2>
            <p className="text-gray-600 mb-4">
              您没有访问信件管理功能的权限
            </p>
            <Button asChild variant="outline">
              <a href="/admin">返回管理控制台</a>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  // 加载数据
  useEffect(() => {
    loadLetters()
    loadStats()
  }, [])

  const loadLetters = async () => {
    setLoading(true)
    try {
      // TODO: 替换为实际API调用
      const mockLetters: Letter[] = [
        {
          id: 'L001',
          title: '给远方朋友的信',
          sender: {
            id: 'U001',
            username: 'student001',
            nickname: '北大小明',
            school_name: '北京大学'
          },
          recipient: {
            id: 'U002',
            username: 'student002',
            nickname: '清华小红',
            school_name: '清华大学'
          },
          status: 'in_transit',
          priority: 'normal',
          content_preview: '亲爱的朋友，好久不见，最近过得怎么样？我在北大的学习生活很充实...',
          word_count: 856,
          created_at: '2024-01-20T10:30:00Z',
          updated_at: '2024-01-21T14:20:00Z',
          courier: {
            id: 'C001',
            name: '快递小王'
          },
          tracking_code: 'OP20240121001',
          delivery_address: '清华大学紫荆公寓',
          flags: []
        },
        {
          id: 'L002',
          title: '新年祝福',
          sender: {
            id: 'U003',
            username: 'student003',
            nickname: '复旦小李',
            school_name: '复旦大学'
          },
          status: 'delivered',
          priority: 'normal',
          content_preview: '新年快乐！愿你在新的一年里身体健康，学业进步，心想事成...',
          word_count: 432,
          created_at: '2024-01-15T08:00:00Z',
          updated_at: '2024-01-18T16:45:00Z',
          delivered_at: '2024-01-18T16:45:00Z',
          courier: {
            id: 'C002',
            name: '快递小张'
          },
          tracking_code: 'OP20240115001',
          flags: []
        },
        {
          id: 'L003',
          title: '紧急通知',
          sender: {
            id: 'U004',
            username: 'admin001',
            nickname: '管理员',
            school_name: '北京大学'
          },
          status: 'failed',
          priority: 'urgent',
          content_preview: '关于学期末考试安排的重要通知，请各位同学务必注意时间安排...',
          word_count: 234,
          created_at: '2024-01-19T14:00:00Z',
          updated_at: '2024-01-20T09:30:00Z',
          tracking_code: 'OP20240119001',
          flags: ['urgent', 'admin']
        }
      ]
      setLetters(mockLetters)
    } catch (error) {
      console.error('Failed to load letters:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadStats = async () => {
    try {
      // TODO: 替换为实际API调用
      const mockStats: LetterStats = {
        total_letters: 5678,
        pending_letters: 234,
        in_transit_letters: 156,
        delivered_letters: 5234,
        failed_letters: 54,
        today_letters: 89,
        this_month_letters: 1456
      }
      setStats(mockStats)
    } catch (error) {
      console.error('Failed to load stats:', error)
    }
  }

  // 根据当前选项卡过滤信件
  const getFilteredLetters = () => {
    let filtered = letters

    // 按选项卡过滤
    if (currentTab !== 'all') {
      filtered = filtered.filter(letter => letter.status === currentTab)
    }

    // 按搜索词过滤
    if (searchTerm) {
      filtered = filtered.filter(letter =>
        letter.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
        letter.sender.nickname.toLowerCase().includes(searchTerm.toLowerCase()) ||
        letter.tracking_code?.toLowerCase().includes(searchTerm.toLowerCase())
      )
    }

    // 按状态过滤
    if (statusFilter !== 'all') {
      filtered = filtered.filter(letter => letter.status === statusFilter)
    }

    // 按优先级过滤
    if (priorityFilter !== 'all') {
      filtered = filtered.filter(letter => letter.priority === priorityFilter)
    }

    // 按学校过滤
    if (schoolFilter !== 'all') {
      filtered = filtered.filter(letter => 
        letter.sender.school_name.includes(schoolFilter)
      )
    }

    return filtered
  }

  const filteredLetters = getFilteredLetters()

  // 信件操作
  const handleViewLetter = (letter: Letter) => {
    setSelectedLetter(letter)
    setShowLetterDetail(true)
  }

  const handleDownloadLetter = (letterId: string) => {
    // TODO: 实现信件下载功能
    console.log('Download letter:', letterId)
  }

  const handleFlagLetter = (letterId: string) => {
    // TODO: 实现信件标记功能
    console.log('Flag letter:', letterId)
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Mail className="w-8 h-8" />
            信件管理
          </h1>
          <p className="text-muted-foreground mt-1">
            监控和管理平台上的所有信件投递状态
          </p>
        </div>
        <Button>
          <Download className="w-4 h-4 mr-2" />
          导出报告
        </Button>
      </div>

      {/* 统计卡片 */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">总信件数</CardTitle>
              <Mail className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total_letters}</div>
              <p className="text-xs text-muted-foreground">
                本月新增 {stats.this_month_letters} 封
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">运输中</CardTitle>
              <Truck className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.in_transit_letters}</div>
              <p className="text-xs text-muted-foreground">
                待处理 {stats.pending_letters} 封
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">已送达</CardTitle>
              <CheckCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.delivered_letters}</div>
              <p className="text-xs text-muted-foreground">
                成功率 {Math.round((stats.delivered_letters / stats.total_letters) * 100)}%
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">今日新增</CardTitle>
              <Calendar className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.today_letters}</div>
              <p className="text-xs text-muted-foreground">
                失败 {stats.failed_letters} 封
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 信件列表 */}
      <Card>
        <CardHeader>
          <CardTitle>信件列表</CardTitle>
          <CardDescription>
            查看和管理所有信件的投递状态
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs value={currentTab} onValueChange={setCurrentTab} className="space-y-4">
            <TabsList>
              <TabsTrigger value="all">全部信件</TabsTrigger>
              <TabsTrigger value="draft">草稿</TabsTrigger>
              <TabsTrigger value="generated">已生成</TabsTrigger>
              <TabsTrigger value="in_transit">运输中</TabsTrigger>
              <TabsTrigger value="delivered">已送达</TabsTrigger>
              <TabsTrigger value="failed">失败</TabsTrigger>
            </TabsList>

            <TabsContent value={currentTab} className="space-y-4">
              {/* 搜索和筛选 */}
              <div className="flex flex-col sm:flex-row gap-4">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input
                    placeholder="搜索信件标题、发送者或追踪码..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                  />
                </div>

                <Select value={statusFilter} onValueChange={setStatusFilter}>
                  <SelectTrigger className="w-full sm:w-40">
                    <SelectValue placeholder="状态筛选" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部状态</SelectItem>
                    {Object.entries(STATUS_NAMES).map(([status, name]) => (
                      <SelectItem key={status} value={status}>{name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select value={priorityFilter} onValueChange={setPriorityFilter}>
                  <SelectTrigger className="w-full sm:w-40">
                    <SelectValue placeholder="优先级" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部优先级</SelectItem>
                    <SelectItem value="normal">普通</SelectItem>
                    <SelectItem value="high">高优先级</SelectItem>
                    <SelectItem value="urgent">紧急</SelectItem>
                  </SelectContent>
                </Select>

                <Select value={schoolFilter} onValueChange={setSchoolFilter}>
                  <SelectTrigger className="w-full sm:w-40">
                    <SelectValue placeholder="学校筛选" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部学校</SelectItem>
                    <SelectItem value="北京大学">北京大学</SelectItem>
                    <SelectItem value="清华大学">清华大学</SelectItem>
                    <SelectItem value="复旦大学">复旦大学</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* 信件表格 */}
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>信件信息</TableHead>
                      <TableHead>发送者</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>优先级</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead>追踪码</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredLetters.map((letter) => (
                      <TableRow key={letter.id}>
                        <TableCell>
                          <div>
                            <div className="font-medium">{letter.title}</div>
                            <div className="text-sm text-muted-foreground">
                              {letter.content_preview.substring(0, 50)}...
                            </div>
                            <div className="text-xs text-muted-foreground mt-1">
                              {letter.word_count} 字
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Avatar className="h-6 w-6">
                              <AvatarImage src={letter.sender.avatar} />
                              <AvatarFallback>{letter.sender.nickname.charAt(0)}</AvatarFallback>
                            </Avatar>
                            <div>
                              <div className="text-sm font-medium">{letter.sender.nickname}</div>
                              <div className="text-xs text-muted-foreground">
                                {letter.sender.school_name}
                              </div>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge className={STATUS_COLORS[letter.status]}>
                            {STATUS_ICONS[letter.status]}
                            <span className="ml-1">{STATUS_NAMES[letter.status]}</span>
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge className={PRIORITY_COLORS[letter.priority]}>
                            {letter.priority === 'urgent' ? '紧急' : 
                             letter.priority === 'high' ? '高' : '普通'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            {new Date(letter.created_at).toLocaleString()}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm font-mono">
                            {letter.tracking_code || '-'}
                          </div>
                        </TableCell>
                        <TableCell>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" className="h-8 w-8 p-0">
                                <MoreVertical className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuLabel>操作</DropdownMenuLabel>
                              <DropdownMenuItem onClick={() => handleViewLetter(letter)}>
                                <Eye className="mr-2 h-4 w-4" />
                                查看详情
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => handleDownloadLetter(letter.id)}>
                                <Download className="mr-2 h-4 w-4" />
                                下载信件
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem onClick={() => handleFlagLetter(letter.id)}>
                                <Flag className="mr-2 h-4 w-4" />
                                标记信件
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {filteredLetters.length === 0 && (
                <div className="text-center py-12">
                  <Mail className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-lg font-semibold mb-2">没有找到信件</h3>
                  <p className="text-muted-foreground">请尝试调整筛选条件</p>
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* 信件详情对话框 */}
      <Dialog open={showLetterDetail} onOpenChange={setShowLetterDetail}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>信件详情</DialogTitle>
            <DialogDescription>
              查看信件的详细信息和投递状态
            </DialogDescription>
          </DialogHeader>
          
          {selectedLetter && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h4 className="font-medium mb-2">基本信息</h4>
                  <div className="space-y-2 text-sm">
                    <div><span className="font-medium">标题:</span> {selectedLetter.title}</div>
                    <div><span className="font-medium">字数:</span> {selectedLetter.word_count}</div>
                    <div>
                      <span className="font-medium">状态:</span> 
                      <Badge className={`ml-2 ${STATUS_COLORS[selectedLetter.status]}`}>
                        {STATUS_NAMES[selectedLetter.status]}
                      </Badge>
                    </div>
                    <div>
                      <span className="font-medium">优先级:</span> 
                      <Badge className={`ml-2 ${PRIORITY_COLORS[selectedLetter.priority]}`}>
                        {selectedLetter.priority === 'urgent' ? '紧急' : 
                         selectedLetter.priority === 'high' ? '高' : '普通'}
                      </Badge>
                    </div>
                  </div>
                </div>
                <div>
                  <h4 className="font-medium mb-2">发送者信息</h4>
                  <div className="space-y-2 text-sm">
                    <div className="flex items-center gap-2">
                      <Avatar className="h-6 w-6">
                        <AvatarImage src={selectedLetter.sender.avatar} />
                        <AvatarFallback>{selectedLetter.sender.nickname.charAt(0)}</AvatarFallback>
                      </Avatar>
                      <span>{selectedLetter.sender.nickname}</span>
                    </div>
                    <div><span className="font-medium">学校:</span> {selectedLetter.sender.school_name}</div>
                    <div><span className="font-medium">用户名:</span> @{selectedLetter.sender.username}</div>
                  </div>
                </div>
              </div>

              {selectedLetter.tracking_code && (
                <div>
                  <h4 className="font-medium mb-2">投递信息</h4>
                  <div className="space-y-2 text-sm">
                    <div><span className="font-medium">追踪码:</span> {selectedLetter.tracking_code}</div>
                    {selectedLetter.courier && (
                      <div><span className="font-medium">信使:</span> {selectedLetter.courier.name}</div>
                    )}
                    {selectedLetter.delivery_address && (
                      <div><span className="font-medium">投递地址:</span> {selectedLetter.delivery_address}</div>
                    )}
                    {selectedLetter.delivered_at && (
                      <div><span className="font-medium">送达时间:</span> {new Date(selectedLetter.delivered_at).toLocaleString()}</div>
                    )}
                  </div>
                </div>
              )}

              <div>
                <h4 className="font-medium mb-2">内容预览</h4>
                <div className="bg-gray-50 p-3 rounded-md text-sm">
                  {selectedLetter.content_preview}
                </div>
              </div>

              <div>
                <h4 className="font-medium mb-2">时间线</h4>
                <div className="space-y-2 text-sm">
                  <div><span className="font-medium">创建时间:</span> {new Date(selectedLetter.created_at).toLocaleString()}</div>
                  <div><span className="font-medium">更新时间:</span> {new Date(selectedLetter.updated_at).toLocaleString()}</div>
                </div>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowLetterDetail(false)}>
              关闭
            </Button>
            <Button onClick={() => selectedLetter && handleDownloadLetter(selectedLetter.id)}>
              下载信件
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}