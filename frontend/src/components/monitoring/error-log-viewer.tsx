'use client'

import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { 
  AlertTriangle,
  AlertCircle,
  Info,
  Bug,
  Search,
  Filter,
  Download,
  RefreshCw,
  Copy,
  ExternalLink,
  Clock,
  FileText,
  Code,
  XCircle
} from 'lucide-react'

interface LogEntry {
  id: string
  timestamp: Date
  level: 'error' | 'warning' | 'info' | 'debug'
  service: string
  message: string
  details?: {
    stack?: string
    request?: {
      method: string
      url: string
      headers: Record<string, string>
      body?: any
    }
    user?: {
      id: string
      email: string
      role: string
    }
    context?: Record<string, any>
  }
  count: number // Number of occurrences
  firstSeen: Date
  lastSeen: Date
}

// Mock data generator
function generateMockLogs(): LogEntry[] {
  const services = ['main-api', 'write-service', 'courier-service', 'admin-service']
  const errors = [
    { message: 'Database connection timeout', level: 'error' as const },
    { message: 'Invalid JWT token', level: 'warning' as const },
    { message: 'Rate limit exceeded', level: 'warning' as const },
    { message: 'Failed to send email notification', level: 'error' as const },
    { message: 'Memory usage above threshold', level: 'warning' as const },
    { message: 'API endpoint deprecated', level: 'info' as const },
    { message: 'Slow query detected', level: 'warning' as const },
    { message: 'Authentication failed', level: 'error' as const }
  ]

  return Array.from({ length: 50 }, (_, i) => {
    const error = errors[Math.floor(Math.random() * errors.length)]
    const now = new Date()
    const timestamp = new Date(now.getTime() - Math.random() * 24 * 60 * 60 * 1000)
    
    return {
      id: `log-${i}`,
      timestamp,
      level: error.level,
      service: services[Math.floor(Math.random() * services.length)],
      message: error.message,
      details: Math.random() > 0.5 ? {
        stack: `Error: ${error.message}\n    at function() (app.js:123:45)\n    at process() (handler.js:67:12)`,
        request: {
          method: 'POST',
          url: '/api/v1/letters',
          headers: { 'Content-Type': 'application/json' },
          body: { letter_id: '123' }
        },
        user: {
          id: `user-${Math.floor(Math.random() * 1000)}`,
          email: 'user@example.com',
          role: 'student'
        }
      } : undefined,
      count: Math.floor(Math.random() * 100) + 1,
      firstSeen: new Date(timestamp.getTime() - Math.random() * 60 * 60 * 1000),
      lastSeen: timestamp
    }
  })
}

export function ErrorLogViewer() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [filteredLogs, setFilteredLogs] = useState<LogEntry[]>([])
  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [levelFilter, setLevelFilter] = useState<string>('all')
  const [serviceFilter, setServiceFilter] = useState<string>('all')
  const [timeRange, setTimeRange] = useState('24h')
  const [autoRefresh, setAutoRefresh] = useState(false)

  // Load initial logs
  useEffect(() => {
    loadLogs()
  }, [])

  // Auto refresh
  useEffect(() => {
    if (!autoRefresh) return

    const interval = setInterval(loadLogs, 30000) // Every 30 seconds
    return () => clearInterval(interval)
  }, [autoRefresh])

  // Filter logs
  useEffect(() => {
    let filtered = [...logs]

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(log =>
        log.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
        log.service.toLowerCase().includes(searchTerm.toLowerCase()) ||
        JSON.stringify(log.details).toLowerCase().includes(searchTerm.toLowerCase())
      )
    }

    // Level filter
    if (levelFilter !== 'all') {
      filtered = filtered.filter(log => log.level === levelFilter)
    }

    // Service filter
    if (serviceFilter !== 'all') {
      filtered = filtered.filter(log => log.service === serviceFilter)
    }

    // Time range filter
    const now = new Date()
    const timeRanges: Record<string, number> = {
      '1h': 60 * 60 * 1000,
      '6h': 6 * 60 * 60 * 1000,
      '24h': 24 * 60 * 60 * 1000,
      '7d': 7 * 24 * 60 * 60 * 1000,
      '30d': 30 * 24 * 60 * 60 * 1000
    }
    
    if (timeRange in timeRanges) {
      const cutoff = new Date(now.getTime() - timeRanges[timeRange])
      filtered = filtered.filter(log => log.timestamp > cutoff)
    }

    // Sort by timestamp (newest first)
    filtered.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime())

    setFilteredLogs(filtered)
  }, [logs, searchTerm, levelFilter, serviceFilter, timeRange])

  const loadLogs = useCallback(async () => {
    setIsLoading(true)
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 500))
      const mockLogs = generateMockLogs()
      setLogs(mockLogs)
    } catch (error) {
      console.error('Failed to load logs:', error)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const exportLogs = () => {
    const data = filteredLogs.map(log => ({
      timestamp: log.timestamp.toISOString(),
      level: log.level,
      service: log.service,
      message: log.message,
      count: log.count,
      details: log.details
    }))

    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `error-logs-${new Date().toISOString().split('T')[0]}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    // TODO: Show toast notification
  }

  // Statistics
  const stats = useMemo(() => {
    const errorCount = filteredLogs.filter(l => l.level === 'error').length
    const warningCount = filteredLogs.filter(l => l.level === 'warning').length
    const totalOccurrences = filteredLogs.reduce((sum, log) => sum + log.count, 0)
    const services = [...new Set(filteredLogs.map(l => l.service))]

    return { errorCount, warningCount, totalOccurrences, services }
  }, [filteredLogs])

  const getLevelIcon = (level: string) => {
    switch (level) {
      case 'error': return <XCircle className="w-4 h-4 text-red-500" />
      case 'warning': return <AlertTriangle className="w-4 h-4 text-yellow-500" />
      case 'info': return <Info className="w-4 h-4 text-blue-500" />
      case 'debug': return <Bug className="w-4 h-4 text-gray-500" />
      default: return <AlertCircle className="w-4 h-4" />
    }
  }

  const getLevelBadge = (level: string) => {
    const variants: Record<string, 'destructive' | 'secondary' | 'default' | 'outline'> = {
      error: 'destructive',
      warning: 'secondary',
      info: 'default',
      debug: 'outline'
    }

    return (
      <Badge variant={variants[level] || 'outline'}>
        {level.toUpperCase()}
      </Badge>
    )
  }

  return (
    <div className="space-y-6">
      {/* Statistics Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">错误总数</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold text-red-600">{stats.errorCount}</span>
              <XCircle className="w-5 h-5 text-red-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">警告数量</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold text-yellow-600">{stats.warningCount}</span>
              <AlertTriangle className="w-5 h-5 text-yellow-500" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">总出现次数</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold">{stats.totalOccurrences}</span>
              <AlertCircle className="w-5 h-5 text-muted-foreground" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">受影响服务</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <span className="text-2xl font-bold">{stats.services.length}</span>
              <FileText className="w-5 h-5 text-muted-foreground" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filters and Controls */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>错误日志</CardTitle>
              <CardDescription>系统错误和异常记录</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant={autoRefresh ? 'default' : 'outline'}
                size="sm"
                onClick={() => setAutoRefresh(!autoRefresh)}
              >
                <RefreshCw className={`w-4 h-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
                自动刷新
              </Button>
              <Button variant="outline" size="sm" onClick={exportLogs}>
                <Download className="w-4 h-4 mr-2" />
                导出
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {/* Filters */}
          <div className="flex flex-col md:flex-row gap-4 mb-6">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder="搜索错误信息..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>

            <Select value={levelFilter} onValueChange={setLevelFilter}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="日志级别" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部级别</SelectItem>
                <SelectItem value="error">错误</SelectItem>
                <SelectItem value="warning">警告</SelectItem>
                <SelectItem value="info">信息</SelectItem>
                <SelectItem value="debug">调试</SelectItem>
              </SelectContent>
            </Select>

            <Select value={serviceFilter} onValueChange={setServiceFilter}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="服务" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部服务</SelectItem>
                <SelectItem value="main-api">主API</SelectItem>
                <SelectItem value="write-service">Write服务</SelectItem>
                <SelectItem value="courier-service">Courier服务</SelectItem>
                <SelectItem value="admin-service">Admin服务</SelectItem>
              </SelectContent>
            </Select>

            <Select value={timeRange} onValueChange={setTimeRange}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="时间范围" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1h">最近1小时</SelectItem>
                <SelectItem value="6h">最近6小时</SelectItem>
                <SelectItem value="24h">最近24小时</SelectItem>
                <SelectItem value="7d">最近7天</SelectItem>
                <SelectItem value="30d">最近30天</SelectItem>
                <SelectItem value="all">全部</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Logs Table */}
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <RefreshCw className="w-6 h-6 animate-spin text-muted-foreground" />
            </div>
          ) : filteredLogs.length === 0 ? (
            <div className="text-center py-12">
              <AlertCircle className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">没有找到匹配的日志记录</p>
            </div>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>时间</TableHead>
                    <TableHead>级别</TableHead>
                    <TableHead>服务</TableHead>
                    <TableHead>错误信息</TableHead>
                    <TableHead>出现次数</TableHead>
                    <TableHead>操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredLogs.map((log) => (
                    <TableRow key={log.id} className="cursor-pointer hover:bg-gray-50">
                      <TableCell className="font-mono text-sm">
                        {log.timestamp.toLocaleString()}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {getLevelIcon(log.level)}
                          {getLevelBadge(log.level)}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{log.service}</Badge>
                      </TableCell>
                      <TableCell className="max-w-md">
                        <p className="truncate">{log.message}</p>
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary">{log.count}</Badge>
                      </TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setSelectedLog(log)}
                        >
                          查看详情
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Log Detail Dialog */}
      {selectedLog && (
        <Dialog open={!!selectedLog} onOpenChange={() => setSelectedLog(null)}>
          <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                {getLevelIcon(selectedLog.level)}
                错误详情
              </DialogTitle>
              <DialogDescription>
                {selectedLog.timestamp.toLocaleString()} - {selectedLog.service}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              <div>
                <h4 className="font-medium mb-2">错误信息</h4>
                <Alert>
                  <AlertDescription>{selectedLog.message}</AlertDescription>
                </Alert>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-muted-foreground">级别</p>
                  <div className="mt-1">{getLevelBadge(selectedLog.level)}</div>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">出现次数</p>
                  <p className="mt-1 font-medium">{selectedLog.count} 次</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">首次出现</p>
                  <p className="mt-1 text-sm">{selectedLog.firstSeen.toLocaleString()}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">最后出现</p>
                  <p className="mt-1 text-sm">{selectedLog.lastSeen.toLocaleString()}</p>
                </div>
              </div>

              {selectedLog.details && (
                <Tabs defaultValue="stack" className="mt-4">
                  <TabsList className="grid w-full grid-cols-3">
                    <TabsTrigger value="stack">堆栈跟踪</TabsTrigger>
                    <TabsTrigger value="request">请求信息</TabsTrigger>
                    <TabsTrigger value="context">上下文</TabsTrigger>
                  </TabsList>

                  <TabsContent value="stack" className="mt-4">
                    {selectedLog.details.stack ? (
                      <div className="relative">
                        <pre className="bg-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
                          <code>{selectedLog.details.stack}</code>
                        </pre>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="absolute top-2 right-2"
                          onClick={() => copyToClipboard(selectedLog.details!.stack!)}
                        >
                          <Copy className="w-4 h-4" />
                        </Button>
                      </div>
                    ) : (
                      <p className="text-muted-foreground">无堆栈跟踪信息</p>
                    )}
                  </TabsContent>

                  <TabsContent value="request" className="mt-4">
                    {selectedLog.details.request ? (
                      <div className="space-y-3">
                        <div>
                          <p className="text-sm font-medium">请求方法</p>
                          <Badge>{selectedLog.details.request.method}</Badge>
                        </div>
                        <div>
                          <p className="text-sm font-medium">URL</p>
                          <code className="text-sm bg-gray-100 px-2 py-1 rounded">
                            {selectedLog.details.request.url}
                          </code>
                        </div>
                        {selectedLog.details.request.body && (
                          <div>
                            <p className="text-sm font-medium">请求体</p>
                            <pre className="bg-gray-100 p-3 rounded text-sm mt-1">
                              <code>{JSON.stringify(selectedLog.details.request.body, null, 2)}</code>
                            </pre>
                          </div>
                        )}
                      </div>
                    ) : (
                      <p className="text-muted-foreground">无请求信息</p>
                    )}
                  </TabsContent>

                  <TabsContent value="context" className="mt-4">
                    {selectedLog.details.user ? (
                      <div className="space-y-3">
                        <h5 className="font-medium">用户信息</h5>
                        <div className="grid grid-cols-2 gap-2 text-sm">
                          <div>
                            <span className="text-muted-foreground">ID:</span>
                            <span className="ml-2">{selectedLog.details.user.id}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">邮箱:</span>
                            <span className="ml-2">{selectedLog.details.user.email}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">角色:</span>
                            <Badge variant="outline" className="ml-2">
                              {selectedLog.details.user.role}
                            </Badge>
                          </div>
                        </div>
                      </div>
                    ) : (
                      <p className="text-muted-foreground">无上下文信息</p>
                    )}
                  </TabsContent>
                </Tabs>
              )}
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}