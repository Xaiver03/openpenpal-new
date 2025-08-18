/**
 * Sensitive Words Management - 敏感词管理界面
 * 仅供四级信使和平台管理员使用
 */

import React, { useState, useEffect, useCallback } from 'react'
import { format } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useToast } from "@/hooks/use-toast"
import { useAuth } from "@/contexts/auth-context-new"
import { enhancedApiClient as apiClient } from "@/lib/api-client-enhanced"
import {
  Plus,
  Upload,
  Download,
  RefreshCw,
  Trash2,
  Edit,
  Search,
  Shield,
  AlertTriangle,
  BarChart,
  Filter,
  FileSpreadsheet,
  Save
} from "lucide-react"

// 敏感词接口
interface SensitiveWord {
  id: string
  word: string
  category: string
  level: 'low' | 'medium' | 'high' | 'block'
  is_active: boolean
  reason?: string
  created_by: string
  created_at: string
  updated_at: string
}

// 统计信息接口
interface SensitiveWordStats {
  total_words: number
  active_words: number
  categories: Array<{ category: string; count: number }>
  levels: Array<{ level: string; count: number }>
  recent_words: SensitiveWord[]
  loaded_in_memory: number
  generated_at: string
}

// 敏感词分类
const CATEGORIES = [
  { value: 'spam', label: '垃圾信息' },
  { value: 'inappropriate', label: '不当内容' },
  { value: 'offensive', label: '冒犯性内容' },
  { value: 'political', label: '政治敏感' },
  { value: 'violence', label: '暴力内容' },
  { value: 'advertisement', label: '广告营销' },
  { value: 'other', label: '其他' }
]

// 风险等级
const LEVELS = [
  { value: 'low', label: '低风险', color: 'default' },
  { value: 'medium', label: '中风险', color: 'warning' },
  { value: 'high', label: '高风险', color: 'destructive' },
  { value: 'block', label: '屏蔽', color: 'destructive' }
]

export function SensitiveWordsManagement() {
  const { toast } = useToast()
  const { user } = useAuth()
  
  // 检查权限
  const hasPermission = user?.role === 'courier_level4' || 
                       user?.role === 'platform_admin' || 
                       user?.role === 'super_admin'
  
  // 数据状态
  const [words, setWords] = useState<SensitiveWord[]>([])
  const [stats, setStats] = useState<SensitiveWordStats | null>(null)
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  
  // 过滤和分页状态
  const [searchTerm, setSearchTerm] = useState('')
  const [categoryFilter, setCategoryFilter] = useState<string>('all')
  const [levelFilter, setLevelFilter] = useState<string>('all')
  const [activeFilter, setActiveFilter] = useState<string>('all')
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize] = useState(20)
  
  // 对话框状态
  const [showAddDialog, setShowAddDialog] = useState(false)
  const [showEditDialog, setShowEditDialog] = useState(false)
  const [showImportDialog, setShowImportDialog] = useState(false)
  const [selectedWord, setSelectedWord] = useState<SensitiveWord | null>(null)
  
  // 表单状态
  const [formData, setFormData] = useState({
    word: '',
    category: 'other',
    level: 'medium' as 'low' | 'medium' | 'high' | 'block'
  })
  
  const [importData, setImportData] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  // 获取敏感词列表
  const loadWords = useCallback(async () => {
    if (!hasPermission) return
    
    setLoading(true)
    try {
      const params = new URLSearchParams({
        page: currentPage.toString(),
        limit: pageSize.toString()
      })
      
      if (categoryFilter !== 'all') params.append('category', categoryFilter)
      if (activeFilter !== 'all') params.append('is_active', activeFilter)
      
      const response = await apiClient.get<{ data: any[]; total: number }>(`/admin/sensitive-words?${params}`)
      if (response.data) {
        setWords(response.data.data)
        setTotal(response.data.total)
      }
    } catch (error) {
      toast({
        title: "加载失败",
        description: "无法加载敏感词列表",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }, [hasPermission, currentPage, pageSize, categoryFilter, activeFilter, toast])

  // 获取统计信息
  const loadStats = useCallback(async () => {
    if (!hasPermission) return
    
    try {
      const response = await apiClient.get<{ data: any }>('/admin/sensitive-words/stats')
      if (response.data) {
        setStats(response.data.data)
      }
    } catch (error) {
      console.error('Failed to load stats:', error)
    }
  }, [hasPermission])

  // 初始化加载
  useEffect(() => {
    if (hasPermission) {
      loadWords()
      loadStats()
    }
  }, [hasPermission, loadWords, loadStats])

  // 添加敏感词
  const handleAdd = async () => {
    setIsSubmitting(true)
    try {
      await apiClient.post('/admin/sensitive-words', formData)
      toast({
        title: "添加成功",
        description: "敏感词已添加到列表",
      })
      setShowAddDialog(false)
      setFormData({ word: '', category: 'other', level: 'medium' })
      loadWords()
      loadStats()
    } catch (error: any) {
      toast({
        title: "添加失败",
        description: error.response?.data?.message || "无法添加敏感词",
        variant: "destructive",
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  // 编辑敏感词
  const handleEdit = async () => {
    if (!selectedWord) return
    
    setIsSubmitting(true)
    try {
      await apiClient.put(`/admin/sensitive-words/${selectedWord.id}`, formData)
      toast({
        title: "更新成功",
        description: "敏感词已更新",
      })
      setShowEditDialog(false)
      setSelectedWord(null)
      loadWords()
    } catch (error) {
      toast({
        title: "更新失败",
        description: "无法更新敏感词",
        variant: "destructive",
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  // 删除敏感词
  const handleDelete = async (word: SensitiveWord) => {
    if (!confirm(`确定要删除敏感词"${word.word}"吗？`)) return
    
    try {
      await apiClient.delete(`/admin/sensitive-words/${word.id}`)
      toast({
        title: "删除成功",
        description: "敏感词已删除",
      })
      loadWords()
      loadStats()
    } catch (error) {
      toast({
        title: "删除失败",
        description: "无法删除敏感词",
        variant: "destructive",
      })
    }
  }

  // 批量导入
  const handleImport = async () => {
    setIsSubmitting(true)
    try {
      const words = importData.split('\n')
        .map(line => line.trim())
        .filter(line => line)
        .map(word => ({
          word,
          category: 'other',
          level: 'medium'
        }))
      
      await apiClient.post('/admin/sensitive-words/batch-import', { words })
      
      toast({
        title: "导入成功",
        description: `成功导入${words.length}个敏感词`,
      })
      setShowImportDialog(false)
      setImportData('')
      loadWords()
      loadStats()
    } catch (error: any) {
      toast({
        title: "导入失败",
        description: error.response?.data?.message || "批量导入失败",
        variant: "destructive",
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  // 导出敏感词
  const handleExport = async () => {
    try {
      const response = await apiClient.get<{ data: any[] }>('/admin/sensitive-words/export')
      const words = response.data?.data || []
      
      // 生成CSV内容
      const csv = [
        ['词汇', '分类', '级别'].join(','),
        ...words.map((w: any) => [w.word, w.category, w.level].join(','))
      ].join('\n')
      
      // 下载文件
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
      const link = document.createElement('a')
      link.href = URL.createObjectURL(blob)
      link.download = `sensitive_words_${format(new Date(), 'yyyyMMdd_HHmmss')}.csv`
      link.click()
      
      toast({
        title: "导出成功",
        description: "敏感词列表已导出",
      })
    } catch (error) {
      toast({
        title: "导出失败",
        description: "无法导出敏感词列表",
        variant: "destructive",
      })
    }
  }

  // 刷新词库
  const handleRefresh = async () => {
    try {
      await apiClient.post('/admin/sensitive-words/refresh')
      toast({
        title: "刷新成功",
        description: "敏感词库已重新加载到内存",
      })
      loadStats()
    } catch (error) {
      toast({
        title: "刷新失败",
        description: "无法刷新敏感词库",
        variant: "destructive",
      })
    }
  }

  // 过滤显示的敏感词
  const filteredWords = words.filter(word => {
    if (searchTerm && !word.word.includes(searchTerm)) return false
    if (levelFilter !== 'all' && word.level !== levelFilter) return false
    return true
  })

  // 权限检查
  if (!hasPermission) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Card className="w-96">
          <CardContent className="pt-6">
            <div className="text-center">
              <Shield className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
              <h3 className="text-lg font-semibold mb-2">权限不足</h3>
              <p className="text-muted-foreground">
                只有四级信使和平台管理员可以管理敏感词
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">敏感词管理</h1>
          <p className="text-muted-foreground">
            管理平台敏感词库，保护用户内容安全
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={handleRefresh} variant="outline">
            <RefreshCw className="h-4 w-4 mr-2" />
            刷新词库
          </Button>
          <Button onClick={handleExport} variant="outline">
            <Download className="h-4 w-4 mr-2" />
            导出
          </Button>
          <Button onClick={() => setShowImportDialog(true)} variant="outline">
            <Upload className="h-4 w-4 mr-2" />
            批量导入
          </Button>
          <Button onClick={() => setShowAddDialog(true)}>
            <Plus className="h-4 w-4 mr-2" />
            添加敏感词
          </Button>
        </div>
      </div>

      {/* 统计卡片 */}
      {stats && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">总敏感词数</CardTitle>
              <FileSpreadsheet className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total_words}</div>
              <p className="text-xs text-muted-foreground">
                活跃: {stats.active_words}
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">内存加载</CardTitle>
              <BarChart className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.loaded_in_memory}</div>
              <p className="text-xs text-muted-foreground">
                实时生效中
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">分类统计</CardTitle>
              <Filter className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.categories.length}</div>
              <p className="text-xs text-muted-foreground">
                个分类
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">高风险词</CardTitle>
              <AlertTriangle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-red-600">
                {stats.levels.find(l => l.level === 'high')?.count || 0}
              </div>
              <p className="text-xs text-muted-foreground">
                需要重点关注
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 过滤器 */}
      <Card>
        <CardHeader>
          <CardTitle>过滤条件</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <div className="flex-1">
              <Label htmlFor="search">搜索敏感词</Label>
              <div className="relative">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  id="search"
                  placeholder="输入敏感词..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>
            
            <div className="w-48">
              <Label htmlFor="category-filter">分类</Label>
              <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                <SelectTrigger>
                  <SelectValue placeholder="选择分类" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部分类</SelectItem>
                  {CATEGORIES.map(cat => (
                    <SelectItem key={cat.value} value={cat.value}>
                      {cat.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            <div className="w-48">
              <Label htmlFor="level-filter">级别</Label>
              <Select value={levelFilter} onValueChange={setLevelFilter}>
                <SelectTrigger>
                  <SelectValue placeholder="选择级别" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部级别</SelectItem>
                  {LEVELS.map(level => (
                    <SelectItem key={level.value} value={level.value}>
                      {level.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            <div className="w-48">
              <Label htmlFor="active-filter">状态</Label>
              <Select value={activeFilter} onValueChange={setActiveFilter}>
                <SelectTrigger>
                  <SelectValue placeholder="选择状态" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部状态</SelectItem>
                  <SelectItem value="true">活跃</SelectItem>
                  <SelectItem value="false">已停用</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 敏感词列表 */}
      <Card>
        <CardHeader>
          <CardTitle>敏感词列表</CardTitle>
          <CardDescription>
            共 {total} 个敏感词
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center p-8">
              <RefreshCw className="h-8 w-8 animate-spin" />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>敏感词</TableHead>
                  <TableHead>分类</TableHead>
                  <TableHead>级别</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredWords.map((word) => (
                  <TableRow key={word.id}>
                    <TableCell className="font-medium">{word.word}</TableCell>
                    <TableCell>
                      {CATEGORIES.find(c => c.value === word.category)?.label || word.category}
                    </TableCell>
                    <TableCell>
                      <Badge variant={LEVELS.find(l => l.value === word.level)?.color as any}>
                        {LEVELS.find(l => l.value === word.level)?.label}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={word.is_active ? "default" : "secondary"}>
                        {word.is_active ? '活跃' : '已停用'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {format(new Date(word.created_at), 'yyyy-MM-dd HH:mm', { locale: zhCN })}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            setSelectedWord(word)
                            setFormData({
                              word: word.word,
                              category: word.category,
                              level: word.level
                            })
                            setShowEditDialog(true)
                          }}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(word)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                
                {filteredWords.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8">
                      <div className="text-muted-foreground">
                        暂无敏感词数据
                      </div>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* 添加敏感词对话框 */}
      <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>添加敏感词</DialogTitle>
            <DialogDescription>
              添加新的敏感词到过滤列表
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="add-word">敏感词</Label>
              <Input
                id="add-word"
                value={formData.word}
                onChange={(e) => setFormData({ ...formData, word: e.target.value })}
                placeholder="输入敏感词..."
              />
            </div>
            
            <div>
              <Label htmlFor="add-category">分类</Label>
              <Select 
                value={formData.category} 
                onValueChange={(value) => setFormData({ ...formData, category: value })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {CATEGORIES.map(cat => (
                    <SelectItem key={cat.value} value={cat.value}>
                      {cat.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <Label htmlFor="add-level">风险级别</Label>
              <Select 
                value={formData.level} 
                onValueChange={(value) => setFormData({ ...formData, level: value as any })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {LEVELS.map(level => (
                    <SelectItem key={level.value} value={level.value}>
                      {level.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAddDialog(false)} disabled={isSubmitting}>
              取消
            </Button>
            <Button onClick={handleAdd} disabled={isSubmitting || !formData.word}>
              {isSubmitting ? '添加中...' : '添加'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 编辑敏感词对话框 */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>编辑敏感词</DialogTitle>
            <DialogDescription>
              修改敏感词的属性
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="edit-word">敏感词</Label>
              <Input
                id="edit-word"
                value={formData.word}
                onChange={(e) => setFormData({ ...formData, word: e.target.value })}
                placeholder="输入敏感词..."
              />
            </div>
            
            <div>
              <Label htmlFor="edit-category">分类</Label>
              <Select 
                value={formData.category} 
                onValueChange={(value) => setFormData({ ...formData, category: value })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {CATEGORIES.map(cat => (
                    <SelectItem key={cat.value} value={cat.value}>
                      {cat.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            <div>
              <Label htmlFor="edit-level">风险级别</Label>
              <Select 
                value={formData.level} 
                onValueChange={(value) => setFormData({ ...formData, level: value as any })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {LEVELS.map(level => (
                    <SelectItem key={level.value} value={level.value}>
                      {level.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEditDialog(false)} disabled={isSubmitting}>
              取消
            </Button>
            <Button onClick={handleEdit} disabled={isSubmitting || !formData.word}>
              {isSubmitting ? '保存中...' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 批量导入对话框 */}
      <Dialog open={showImportDialog} onOpenChange={setShowImportDialog}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>批量导入敏感词</DialogTitle>
            <DialogDescription>
              每行输入一个敏感词，系统将自动处理重复项
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="import-data">敏感词列表</Label>
              <textarea
                id="import-data"
                className="w-full h-48 p-3 border rounded-md"
                placeholder="每行一个敏感词..."
                value={importData}
                onChange={(e) => setImportData(e.target.value)}
              />
            </div>
            
            <div className="text-sm text-muted-foreground">
              <p>提示：</p>
              <ul className="list-disc list-inside space-y-1">
                <li>每行输入一个敏感词</li>
                <li>系统会自动转换为小写</li>
                <li>重复的词汇将被忽略</li>
                <li>默认分类为"其他"，级别为"中风险"</li>
              </ul>
            </div>
          </div>
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowImportDialog(false)} disabled={isSubmitting}>
              取消
            </Button>
            <Button onClick={handleImport} disabled={isSubmitting || !importData.trim()}>
              {isSubmitting ? '导入中...' : '开始导入'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default SensitiveWordsManagement