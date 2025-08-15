/**
 * Comment Reports Management - SOTA实现
 * 评论举报管理 - 管理员处理举报的完整界面
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
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import { useToast } from "@/hooks/use-toast"
import {
  AlertTriangle,
  Eye,
  CheckCircle,
  XCircle,
  Clock,
  Search,
  Filter,
  RefreshCw,
  MessageSquare,
  User,
  Calendar
} from "lucide-react"

// 举报状态类型
type ReportStatus = 'pending' | 'resolved' | 'dismissed'

// 举报记录接口
interface CommentReport {
  id: string
  comment_id: string
  reporter_id: string
  reason: string
  description?: string
  status: ReportStatus
  created_at: string
  handled_at?: string
  handled_by?: string
  handler_note?: string
  
  // 关联数据
  comment?: {
    id: string
    content: string
    user?: {
      id: string
      username: string
      nickname: string
    }
  }
  reporter?: {
    id: string
    username: string
    nickname: string
  }
  handler?: {
    id: string
    username: string
    nickname: string
  }
}

// 举报处理请求
interface HandleReportRequest {
  status: 'resolved' | 'dismissed'
  handler_note?: string
}

// 举报统计
interface ReportStats {
  total: number
  pending: number
  resolved: number
  dismissed: number
  recent_increase: number
}

// Mock API 函数（实际项目中应该连接到真实API）
const fetchReports = async (filters: {
  status?: ReportStatus
  page?: number
  limit?: number
}): Promise<{ reports: CommentReport[]; total: number }> => {
  // TODO: 实现真实API调用
  await new Promise(resolve => setTimeout(resolve, 500))
  
  const mockReports: CommentReport[] = [
    {
      id: '1',
      comment_id: 'comment-1',
      reporter_id: 'user-1',
      reason: 'spam',
      description: '这是垃圾广告内容',
      status: 'pending',
      created_at: new Date().toISOString(),
      comment: {
        id: 'comment-1',
        content: '这是一条需要审核的评论内容...',
        user: {
          id: 'commenter-1',
          username: 'user123',
          nickname: '用户123'
        }
      },
      reporter: {
        id: 'user-1',
        username: 'reporter1',
        nickname: '举报者1'
      }
    }
  ]
  
  return {
    reports: mockReports.filter(r => !filters.status || r.status === filters.status),
    total: mockReports.length
  }
}

const fetchReportStats = async (): Promise<ReportStats> => {
  await new Promise(resolve => setTimeout(resolve, 300))
  return {
    total: 25,
    pending: 8,
    resolved: 12,
    dismissed: 5,
    recent_increase: 3
  }
}

const handleReport = async (reportId: string, request: HandleReportRequest): Promise<void> => {
  await new Promise(resolve => setTimeout(resolve, 1000))
  // TODO: 实现真实API调用
}

export function CommentReportsManagement() {
  const { toast } = useToast()
  
  // 数据状态
  const [reports, setReports] = useState<CommentReport[]>([])
  const [stats, setStats] = useState<ReportStats | null>(null)
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  
  // 过滤和分页状态
  const [statusFilter, setStatusFilter] = useState<ReportStatus | 'all'>('all')
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize] = useState(20)
  
  // 处理对话框状态
  const [selectedReport, setSelectedReport] = useState<CommentReport | null>(null)
  const [showHandleDialog, setShowHandleDialog] = useState(false)
  const [handleAction, setHandleAction] = useState<'resolved' | 'dismissed'>('resolved')
  const [handlerNote, setHandlerNote] = useState('')
  const [isHandling, setIsHandling] = useState(false)

  // 获取举报列表
  const loadReports = useCallback(async () => {
    setLoading(true)
    try {
      const filters = {
        status: statusFilter === 'all' ? undefined : statusFilter,
        page: currentPage,
        limit: pageSize
      }
      
      const { reports: data, total: totalCount } = await fetchReports(filters)
      setReports(data)
      setTotal(totalCount)
    } catch (error) {
      toast({
        title: "加载失败",
        description: "无法加载举报列表",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }, [statusFilter, currentPage, pageSize, toast])

  // 获取统计信息
  const loadStats = useCallback(async () => {
    try {
      const statsData = await fetchReportStats()
      setStats(statsData)
    } catch (error) {
      console.error('Failed to load stats:', error)
    }
  }, [])

  // 初始化加载
  useEffect(() => {
    loadReports()
    loadStats()
  }, [loadReports, loadStats])

  // 处理举报
  const handleReportAction = async () => {
    if (!selectedReport) return
    
    setIsHandling(true)
    try {
      const request: HandleReportRequest = {
        status: handleAction,
        handler_note: handlerNote.trim() || undefined
      }
      
      await handleReport(selectedReport.id, request)
      
      toast({
        title: "处理成功",
        description: `举报已${handleAction === 'resolved' ? '解决' : '驳回'}`,
      })
      
      // 重新加载数据
      loadReports()
      loadStats()
      
      // 关闭对话框
      setShowHandleDialog(false)
      setSelectedReport(null)
      setHandlerNote('')
    } catch (error) {
      toast({
        title: "处理失败",
        description: "无法处理举报，请稍后重试",
        variant: "destructive",
      })
    } finally {
      setIsHandling(false)
    }
  }

  // 获取状态徽章
  const getStatusBadge = (status: ReportStatus) => {
    const variants = {
      pending: { variant: 'destructive' as const, label: '待处理', icon: Clock },
      resolved: { variant: 'default' as const, label: '已解决', icon: CheckCircle },
      dismissed: { variant: 'secondary' as const, label: '已驳回', icon: XCircle }
    }
    
    const config = variants[status]
    const Icon = config.icon
    
    return (
      <Badge variant={config.variant} className="gap-1">
        <Icon className="h-3 w-3" />
        {config.label}
      </Badge>
    )
  }

  // 获取举报原因显示文本
  const getReasonText = (reason: string) => {
    const reasons = {
      spam: '垃圾信息',
      inappropriate: '不当内容',
      offensive: '冒犯性内容',
      false_info: '虚假信息',
      other: '其他'
    }
    return reasons[reason as keyof typeof reasons] || reason
  }

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">举报管理</h1>
          <p className="text-muted-foreground">
            管理和处理用户举报的评论内容
          </p>
        </div>
        <Button onClick={loadReports} disabled={loading}>
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          刷新
        </Button>
      </div>

      {/* 统计卡片 */}
      {stats && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">总举报数</CardTitle>
              <AlertTriangle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total}</div>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">待处理</CardTitle>
              <Clock className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-destructive">{stats.pending}</div>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">已解决</CardTitle>
              <CheckCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-600">{stats.resolved}</div>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">已驳回</CardTitle>
              <XCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-orange-600">{stats.dismissed}</div>
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
            <div className="w-48">
              <Label htmlFor="status-filter">举报状态</Label>
              <Select value={statusFilter} onValueChange={(value) => setStatusFilter(value as ReportStatus | 'all')}>
                <SelectTrigger>
                  <SelectValue placeholder="选择状态" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部状态</SelectItem>
                  <SelectItem value="pending">待处理</SelectItem>
                  <SelectItem value="resolved">已解决</SelectItem>
                  <SelectItem value="dismissed">已驳回</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 举报列表 */}
      <Card>
        <CardHeader>
          <CardTitle>举报列表</CardTitle>
          <CardDescription>
            共 {total} 条举报记录
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
                  <TableHead>举报原因</TableHead>
                  <TableHead>评论内容</TableHead>
                  <TableHead>举报者</TableHead>
                  <TableHead>评论作者</TableHead>
                  <TableHead>举报时间</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {reports.map((report) => (
                  <TableRow key={report.id}>
                    <TableCell>
                      <div>
                        <div className="font-medium">{getReasonText(report.reason)}</div>
                        {report.description && (
                          <div className="text-sm text-muted-foreground mt-1">
                            {report.description}
                          </div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="max-w-xs">
                        <p className="truncate text-sm">
                          {report.comment?.content || '评论已删除'}
                        </p>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <User className="h-4 w-4" />
                        <span className="text-sm">
                          {report.reporter?.nickname || report.reporter?.username}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <MessageSquare className="h-4 w-4" />
                        <span className="text-sm">
                          {report.comment?.user?.nickname || report.comment?.user?.username}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Calendar className="h-4 w-4" />
                        <span className="text-sm">
                          {format(new Date(report.created_at), 'yyyy-MM-dd HH:mm', { locale: zhCN })}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      {getStatusBadge(report.status)}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => {
                            setSelectedReport(report)
                            setShowHandleDialog(true)
                          }}
                          disabled={report.status !== 'pending'}
                        >
                          <Eye className="h-4 w-4 mr-1" />
                          处理
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                
                {reports.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-8">
                      <div className="text-muted-foreground">
                        暂无举报记录
                      </div>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* 处理举报对话框 */}
      <Dialog open={showHandleDialog} onOpenChange={setShowHandleDialog}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>处理举报</DialogTitle>
            <DialogDescription>
              请选择处理方式并填写处理说明
            </DialogDescription>
          </DialogHeader>

          {selectedReport && (
            <div className="space-y-4">
              {/* 举报详情 */}
              <div className="p-4 rounded-lg bg-muted/50">
                <h4 className="font-medium mb-2">举报详情</h4>
                <div className="space-y-2 text-sm">
                  <div><strong>举报原因：</strong>{getReasonText(selectedReport.reason)}</div>
                  {selectedReport.description && (
                    <div><strong>详细说明：</strong>{selectedReport.description}</div>
                  )}
                  <div><strong>举报者：</strong>{selectedReport.reporter?.nickname}</div>
                  <div><strong>评论内容：</strong>{selectedReport.comment?.content}</div>
                </div>
              </div>

              {/* 处理选项 */}
              <div className="space-y-2">
                <Label>处理方式</Label>
                <Select value={handleAction} onValueChange={(value) => setHandleAction(value as 'resolved' | 'dismissed')}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="resolved">解决举报（隐藏/删除评论）</SelectItem>
                    <SelectItem value="dismissed">驳回举报（保留评论）</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* 处理说明 */}
              <div className="space-y-2">
                <Label htmlFor="handler-note">处理说明（可选）</Label>
                <Textarea
                  id="handler-note"
                  placeholder="请说明处理理由..."
                  value={handlerNote}
                  onChange={(e) => setHandlerNote(e.target.value)}
                  rows={3}
                />
              </div>
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowHandleDialog(false)}
              disabled={isHandling}
            >
              取消
            </Button>
            <Button onClick={handleReportAction} disabled={isHandling}>
              {isHandling ? '处理中...' : '确认处理'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default CommentReportsManagement