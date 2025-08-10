'use client'

import React, { useState, useEffect } from 'react'
import { 
  Shield, 
  Search, 
  Filter, 
  Eye, 
  CheckCircle,
  XCircle,
  Clock,
  AlertTriangle,
  Settings,
  Plus,
  Edit,
  Trash2,
  Flag,
  FileText,
  Users,
  BarChart3
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { usePermission, PERMISSIONS } from '@/hooks/use-permission'
import { BackButton } from '@/components/ui/back-button'
import { moderationApi, type ModerationRecord, type SensitiveWord, type ModerationRule } from '@/lib/api/moderation'

// 常量定义
const STATUS_COLORS = {
  pending: 'bg-yellow-100 text-yellow-800',
  approved: 'bg-green-100 text-green-800',
  rejected: 'bg-red-100 text-red-800',
  review: 'bg-blue-100 text-blue-800'
}

const STATUS_NAMES = {
  pending: '待审核',
  approved: '已通过',
  rejected: '已拒绝',
  review: '需复审'
}

const LEVEL_COLORS = {
  low: 'bg-gray-100 text-gray-800',
  medium: 'bg-yellow-100 text-yellow-800',
  high: 'bg-orange-100 text-orange-800',
  block: 'bg-red-100 text-red-800'
}

const LEVEL_NAMES = {
  low: '低风险',
  medium: '中风险',
  high: '高风险',
  block: '需屏蔽'
}

export default function ModerationPage() {
  const { user, hasPermission } = usePermission()
  const [queue, setQueue] = useState<ModerationRecord[]>([])
  const [sensitiveWords, setSensitiveWords] = useState<SensitiveWord[]>([])
  const [rules, setRules] = useState<ModerationRule[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  
  // 对话框状态
  const [selectedRecord, setSelectedRecord] = useState<ModerationRecord | null>(null)
  const [showReviewDialog, setShowReviewDialog] = useState(false)
  const [showWordDialog, setShowWordDialog] = useState(false)
  const [showRuleDialog, setShowRuleDialog] = useState(false)
  const [reviewNote, setReviewNote] = useState('')
  const [reviewStatus, setReviewStatus] = useState<'approved' | 'rejected'>('approved')

  // 敏感词表单
  const [wordForm, setWordForm] = useState({
    word: '',
    category: '',
    level: 'medium' as const
  })

  // 规则表单
  const [ruleForm, setRuleForm] = useState({
    name: '',
    description: '',
    content_type: 'letter',
    rule_type: 'keyword',
    pattern: '',
    action: 'review',
    priority: 50
  })

  // 权限检查
  if (!user || !hasPermission(PERMISSIONS.SYSTEM_CONFIG)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Shield className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">访问权限不足</h2>
            <p className="text-gray-600 mb-4">
              您没有访问内容审核管理的权限
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
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      // 并行调用所有API
      const [queueRes, wordsRes, rulesRes] = await Promise.all([
        moderationApi.getModerationQueue({ limit: 50}),
        moderationApi.getSensitiveWords(),
        moderationApi.getModerationRules()
      ])
      
      // 处理API响应
      if (queueRes.data && typeof queueRes.data === 'object' && 'queue' in queueRes.data) {
        setQueue((queueRes.data as any).queue)
      }
      
      if (wordsRes.data && typeof wordsRes.data === 'object' && 'words' in wordsRes.data) {
        setSensitiveWords((wordsRes.data as any).words)
      }
      
      if (rulesRes.data && typeof rulesRes.data === 'object' && 'rules' in rulesRes.data) {
        setRules((rulesRes.data as any).rules)
      }
    } catch (error) {
      console.error('Failed to load moderation data:', error)
      
      // 如果API调用失败，使用模拟数据作为后备
      const mockQueue: ModerationRecord[] = [
        {
          id: '1',
          content_type: 'letter',
          content_id: 'letter_001',
          userId: 'user_001',
          content: '这是一封需要审核的信件内容，可能包含一些敏感信息...',
          status: 'pending',
          level: 'medium',
          score: 0.6,
          reasons: ['包含敏感词: 测试', '内容长度异常'],
          categories: ['spam', 'inappropriate'],
          created_at: '2024-01-21T10:30:00Z'
        },
        {
          id: '2',
          content_type: 'letter',
          content_id: 'letter_002',
          userId: 'user_002',
          content: '另一封待审核的信件内容...',
          status: 'review',
          level: 'high',
          score: 0.8,
          reasons: ['触发规则: 违规关键词检测'],
          categories: ['inappropriate'],
          created_at: '2024-01-21T09:15:00Z'
        }
      ]

      const mockWords: SensitiveWord[] = [
        {
          id: '1',
          word: '测试敏感词',
          category: '不当内容',
          level: 'medium',
          is_active: true,
          created_at: '2024-01-20T14:20:00Z'
        },
        {
          id: '2',
          word: '违规词汇',
          category: '违法违规',
          level: 'high',
          is_active: true,
          created_at: '2024-01-19T16:45:00Z'
        }
      ]

      const mockRules: ModerationRule[] = [
        {
          id: '1',
          name: '违规关键词检测',
          description: '检测信件中的违规关键词',
          content_type: 'letter',
          rule_type: 'keyword',
          pattern: '违规|非法|不当',
          action: 'review',
          priority: 80,
          is_active: true,
          created_at: '2024-01-18T10:00:00Z'
        },
        {
          id: '2',
          name: '内容长度检查',
          description: '检查信件内容长度是否合理',
          content_type: 'letter',
          rule_type: 'length',
          pattern: '5000',
          action: 'flag',
          priority: 50,
          is_active: true,
          created_at: '2024-01-17T15:30:00Z'
        }
      ]

      setQueue(mockQueue)
      setSensitiveWords(mockWords)
      setRules(mockRules)
    } finally {
      setLoading(false)
    }
  }

  // 审核内容
  const handleReview = async (recordId: string, status: 'approved' | 'rejected', note: string) => {
    try {
      await moderationApi.reviewContent({
        record_id: recordId,
        status,
        review_note: note
      })

      // 更新本地状态
      setQueue(prev => prev.map(item => 
        item.id === recordId 
          ? { ...item, status, reviewer_id: user?.id, reviewed_at: new Date().toISOString() }
          : item
      ))

      setShowReviewDialog(false)
      setSelectedRecord(null)
      setReviewNote('')
    } catch (error) {
      console.error('Failed to review content:', error)
      alert('审核失败，请重试')
    }
  }

  // 添加敏感词
  const handleAddWord = async () => {
    try {
      await moderationApi.addSensitiveWord(wordForm)

      // 模拟添加到本地状态
      const newWord: SensitiveWord = {
        id: Date.now().toString(),
        ...wordForm,
        is_active: true,
        created_at: new Date().toISOString()
      }

      setSensitiveWords(prev => [newWord, ...prev])
      setShowWordDialog(false)
      setWordForm({ word: '', category: '', level: 'medium' })
    } catch (error) {
      console.error('Failed to add sensitive word:', error)
      alert('添加敏感词失败，请重试')
    }
  }

  // 添加规则
  const handleAddRule = async () => {
    try {
      await moderationApi.addModerationRule(ruleForm)

      // 模拟添加到本地状态
      const newRule: ModerationRule = {
        id: Date.now().toString(),
        ...ruleForm,
        is_active: true,
        created_at: new Date().toISOString()
      }

      setRules(prev => [newRule, ...prev])
      setShowRuleDialog(false)
      setRuleForm({
        name: '',
        description: '',
        content_type: 'letter',
        rule_type: 'keyword',
        pattern: '',
        action: 'review',
        priority: 50
      })
    } catch (error) {
      console.error('Failed to add rule:', error)
      alert('添加规则失败，请重试')
    }
  }

  // 过滤队列
  const filteredQueue = queue.filter(item => {
    const matchesSearch = item.content.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         item.reasons.some(reason => reason.toLowerCase().includes(searchTerm.toLowerCase()))
    const matchesStatus = statusFilter === 'all' || item.status === statusFilter
    return matchesSearch && matchesStatus
  })

  const stats = {
    total: queue.length,
    pending: queue.filter(q => q.status === 'pending').length,
    approved: queue.filter(q => q.status === 'approved').length,
    rejected: queue.filter(q => q.status === 'rejected').length,
    totalWords: sensitiveWords.length,
    totalRules: rules.length
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
        <div className="flex items-center gap-4">
          <BackButton href="/admin" />
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-2">
              <Shield className="w-8 h-8" />
              内容审核管理
            </h1>
            <p className="text-muted-foreground mt-1">
              管理平台内容审核、敏感词库和审核规则
            </p>
          </div>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">待审核内容</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.pending}</div>
            <p className="text-xs text-muted-foreground">
              总计 {stats.total} 条记录
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">已通过</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.approved}</div>
            <p className="text-xs text-muted-foreground">
              通过率 {stats.total > 0 ? Math.round((stats.approved / stats.total) * 100) : 0}%
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">敏感词库</CardTitle>
            <Flag className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalWords}</div>
            <p className="text-xs text-muted-foreground">
              活跃词汇
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">审核规则</CardTitle>
            <Settings className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalRules}</div>
            <p className="text-xs text-muted-foreground">
              配置规则
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 主要内容 */}
      <Tabs defaultValue="queue" className="space-y-6">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="queue">审核队列</TabsTrigger>
          <TabsTrigger value="words">敏感词库</TabsTrigger>
          <TabsTrigger value="rules">审核规则</TabsTrigger>
        </TabsList>

        {/* 审核队列 */}
        <TabsContent value="queue" className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>待审核队列</CardTitle>
                  <CardDescription>需要人工审核的内容列表</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {/* 搜索和筛选 */}
              <div className="flex flex-col sm:flex-row gap-4 mb-6">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input
                    placeholder="搜索内容或原因..."
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
                    <SelectItem value="pending">待审核</SelectItem>
                    <SelectItem value="review">需复审</SelectItem>
                    <SelectItem value="approved">已通过</SelectItem>
                    <SelectItem value="rejected">已拒绝</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* 审核队列表格 */}
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>内容</TableHead>
                      <TableHead>类型</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>风险等级</TableHead>
                      <TableHead>审核原因</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredQueue.map((record) => (
                      <TableRow key={record.id}>
                        <TableCell>
                          <div className="max-w-xs truncate">
                            {record.content}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {record.content_type === 'letter' ? '信件' : '其他'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge className={STATUS_COLORS[record.status]}>
                            {STATUS_NAMES[record.status]}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge className={LEVEL_COLORS[record.level]}>
                            {LEVEL_NAMES[record.level]}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="space-y-1">
                            {record.reasons.slice(0, 2).map((reason, index) => (
                              <div key={index} className="text-xs text-muted-foreground">
                                {reason}
                              </div>
                            ))}
                            {record.reasons.length > 2 && (
                              <div className="text-xs text-muted-foreground">
                                +{record.reasons.length - 2} more
                              </div>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            {new Date(record.created_at).toLocaleString()}
                          </div>
                        </TableCell>
                        <TableCell>
                          {record.status === 'pending' || record.status === 'review' ? (
                            <Button
                              size="sm"
                              onClick={() => {
                                setSelectedRecord(record)
                                setShowReviewDialog(true)
                              }}
                            >
                              <Eye className="w-4 h-4 mr-1" />
                              审核
                            </Button>
                          ) : (
                            <Button size="sm" variant="outline">
                              <Eye className="w-4 h-4 mr-1" />
                              查看
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 敏感词库 */}
        <TabsContent value="words" className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>敏感词库管理</CardTitle>
                  <CardDescription>管理系统敏感词汇库</CardDescription>
                </div>
                <Button onClick={() => setShowWordDialog(true)}>
                  <Plus className="w-4 h-4 mr-2" />
                  添加敏感词
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>敏感词</TableHead>
                      <TableHead>分类</TableHead>
                      <TableHead>风险等级</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {sensitiveWords.map((word) => (
                      <TableRow key={word.id}>
                        <TableCell className="font-medium">{word.word}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{word.category}</Badge>
                        </TableCell>
                        <TableCell>
                          <Badge className={LEVEL_COLORS[word.level]}>
                            {LEVEL_NAMES[word.level]}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant={word.is_active ? "default" : "secondary"}>
                            {word.is_active ? '启用' : '禁用'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {new Date(word.created_at).toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-2">
                            <Button size="sm" variant="outline">
                              <Edit className="w-4 h-4" />
                            </Button>
                            <Button size="sm" variant="outline">
                              <Trash2 className="w-4 h-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 审核规则 */}
        <TabsContent value="rules" className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>审核规则管理</CardTitle>
                  <CardDescription>配置内容自动审核规则</CardDescription>
                </div>
                <Button onClick={() => setShowRuleDialog(true)}>
                  <Plus className="w-4 h-4 mr-2" />
                  添加规则
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>规则名称</TableHead>
                      <TableHead>内容类型</TableHead>
                      <TableHead>规则类型</TableHead>
                      <TableHead>动作</TableHead>
                      <TableHead>优先级</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {rules.map((rule) => (
                      <TableRow key={rule.id}>
                        <TableCell>
                          <div>
                            <div className="font-medium">{rule.name}</div>
                            <div className="text-sm text-muted-foreground">
                              {rule.description}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {rule.content_type === 'letter' ? '信件' : '其他'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{rule.rule_type}</Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{rule.action}</Badge>
                        </TableCell>
                        <TableCell>{rule.priority}</TableCell>
                        <TableCell>
                          <Badge variant={rule.is_active ? "default" : "secondary"}>
                            {rule.is_active ? '启用' : '禁用'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-2">
                            <Button size="sm" variant="outline">
                              <Edit className="w-4 h-4" />
                            </Button>
                            <Button size="sm" variant="outline">
                              <Trash2 className="w-4 h-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* 审核对话框 */}
      <Dialog open={showReviewDialog} onOpenChange={setShowReviewDialog}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>内容审核</DialogTitle>
            <DialogDescription>
              请仔细审核以下内容并做出决定
            </DialogDescription>
          </DialogHeader>
          
          {selectedRecord && (
            <div className="space-y-4">
              <div>
                <Label>内容</Label>
                <div className="mt-1 p-3 bg-gray-50 rounded-md text-sm">
                  {selectedRecord.content}
                </div>
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label>内容类型</Label>
                  <div className="mt-1">
                    <Badge variant="outline">
                      {selectedRecord.content_type === 'letter' ? '信件' : '其他'}
                    </Badge>
                  </div>
                </div>
                <div>
                  <Label>风险分数</Label>
                  <div className="mt-1">
                    <Badge className={LEVEL_COLORS[selectedRecord.level]}>
                      {selectedRecord.score.toFixed(2)} - {LEVEL_NAMES[selectedRecord.level]}
                    </Badge>
                  </div>
                </div>
              </div>

              <div>
                <Label>检测原因</Label>
                <div className="mt-1 space-y-1">
                  {selectedRecord.reasons.map((reason, index) => (
                    <div key={index} className="text-sm text-muted-foreground">
                      • {reason}
                    </div>
                  ))}
                </div>
              </div>

              <div>
                <Label htmlFor="review-status">审核决定</Label>
                <Select value={reviewStatus} onValueChange={(value: 'approved' | 'rejected') => setReviewStatus(value)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="approved">通过</SelectItem>
                    <SelectItem value="rejected">拒绝</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div>
                <Label htmlFor="review-note">审核备注</Label>
                <Textarea
                  id="review-note"
                  placeholder="请填写审核意见..."
                  value={reviewNote}
                  onChange={(e) => setReviewNote(e.target.value)}
                  rows={3}
                />
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowReviewDialog(false)}>
              取消
            </Button>
            <Button 
              onClick={() => selectedRecord && handleReview(selectedRecord.id, reviewStatus, reviewNote)}
              className={reviewStatus === 'approved' ? 'bg-green-600 hover:bg-green-700' : 'bg-red-600 hover:bg-red-700'}
            >
              {reviewStatus === 'approved' ? (
                <>
                  <CheckCircle className="w-4 h-4 mr-2" />
                  通过审核
                </>
              ) : (
                <>
                  <XCircle className="w-4 h-4 mr-2" />
                  拒绝内容
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 添加敏感词对话框 */}
      <Dialog open={showWordDialog} onOpenChange={setShowWordDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>添加敏感词</DialogTitle>
            <DialogDescription>
              向敏感词库中添加新的词汇
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="word">敏感词</Label>
              <Input
                id="word"
                value={wordForm.word}
                onChange={(e) => setWordForm(prev => ({ ...prev, word: e.target.value }))}
                placeholder="输入敏感词..."
              />
            </div>
            
            <div>
              <Label htmlFor="category">分类</Label>
              <Input
                id="category"
                value={wordForm.category}
                onChange={(e) => setWordForm(prev => ({ ...prev, category: e.target.value }))}
                placeholder="如：不当内容、违法违规等"
              />
            </div>
            
            <div>
              <Label htmlFor="level">风险等级</Label>
              <Select value={wordForm.level} onValueChange={(value: any) => setWordForm(prev => ({ ...prev, level: value }))}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">低风险</SelectItem>
                  <SelectItem value="medium">中风险</SelectItem>
                  <SelectItem value="high">高风险</SelectItem>
                  <SelectItem value="block">需屏蔽</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowWordDialog(false)}>
              取消
            </Button>
            <Button onClick={handleAddWord} disabled={!wordForm.word.trim()}>
              添加
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 添加规则对话框 */}
      <Dialog open={showRuleDialog} onOpenChange={setShowRuleDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>添加审核规则</DialogTitle>
            <DialogDescription>
              创建新的内容自动审核规则
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="rule-name">规则名称</Label>
              <Input
                id="rule-name"
                value={ruleForm.name}
                onChange={(e) => setRuleForm(prev => ({ ...prev, name: e.target.value }))}
                placeholder="输入规则名称..."
              />
            </div>
            
            <div>
              <Label htmlFor="rule-description">规则描述</Label>
              <Textarea
                id="rule-description"
                value={ruleForm.description}
                onChange={(e) => setRuleForm(prev => ({ ...prev, description: e.target.value }))}
                placeholder="描述规则的作用..."
                rows={2}
              />
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="content-type">内容类型</Label>
                <Select value={ruleForm.content_type} onValueChange={(value) => setRuleForm(prev => ({ ...prev, content_type: value }))}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="letter">信件</SelectItem>
                    <SelectItem value="profile">个人资料</SelectItem>
                    <SelectItem value="museum">博物馆</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              
              <div>
                <Label htmlFor="rule-type">规则类型</Label>
                <Select value={ruleForm.rule_type} onValueChange={(value) => setRuleForm(prev => ({ ...prev, rule_type: value }))}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="keyword">关键词</SelectItem>
                    <SelectItem value="regex">正则表达式</SelectItem>
                    <SelectItem value="length">长度检查</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            
            <div>
              <Label htmlFor="pattern">规则模式</Label>
              <Input
                id="pattern"
                value={ruleForm.pattern}
                onChange={(e) => setRuleForm(prev => ({ ...prev, pattern: e.target.value }))}
                placeholder="输入检测模式..."
              />
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="action">触发动作</Label>
                <Select value={ruleForm.action} onValueChange={(value) => setRuleForm(prev => ({ ...prev, action: value }))}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="pass">放行</SelectItem>
                    <SelectItem value="flag">标记</SelectItem>
                    <SelectItem value="review">人工审核</SelectItem>
                    <SelectItem value="block">直接拒绝</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              
              <div>
                <Label htmlFor="priority">优先级</Label>
                <Input
                  id="priority"
                  type="number"
                  value={ruleForm.priority}
                  onChange={(e) => setRuleForm(prev => ({ ...prev, priority: parseInt(e.target.value) || 50 }))}
                  min="1"
                  max="100"
                />
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowRuleDialog(false)}>
              取消
            </Button>
            <Button onClick={handleAddRule} disabled={!ruleForm.name.trim() || !ruleForm.pattern.trim()}>
              添加规则
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}