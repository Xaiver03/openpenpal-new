'use client'

import React, { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { 
  CheckCircle, 
  Clock, 
  AlertCircle, 
  XCircle, 
  Play, 
  Pause,
  RefreshCw,
  Filter,
  Award
} from 'lucide-react'
import { useCreditTasks, useCreditStore } from '@/stores/credit-store'
import { formatPoints } from '@/lib/api/credit'
import { CreditTaskListParams, TASK_DESCRIPTIONS, TASK_STATUS_DESCRIPTIONS } from '@/types/credit'
import type { CreditTaskStatus, CreditTaskType } from '@/types/credit'

interface CreditTaskListProps {
  showFilters?: boolean
  pageSize?: number
  className?: string
}

export function CreditTaskList({ 
  showFilters = true, 
  pageSize = 20,
  className = '' 
}: CreditTaskListProps) {
  const { tasks, total, page, limit, loading, error } = useCreditTasks()
  const { fetchTasks, clearError } = useCreditStore()
  
  const [filters, setFilters] = useState<CreditTaskListParams>({
    page: 1,
    limit: pageSize,
  })

  useEffect(() => {
    fetchTasks(filters)
  }, [filters, fetchTasks])

  const handleFilterChange = (key: keyof CreditTaskListParams, value: any) => {
    setFilters(prev => ({
      ...prev,
      [key]: value,
      page: 1 // 重置页码
    }))
  }

  const handleLoadMore = () => {
    const nextPage = page + 1
    setFilters(prev => ({
      ...prev,
      page: nextPage
    }))
  }

  const handleRefresh = () => {
    clearError()
    fetchTasks({ ...filters, page: 1 })
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffHours = diffMs / (1000 * 60 * 60)
    const diffDays = diffMs / (1000 * 60 * 60 * 24)

    if (diffHours < 1) {
      return '刚刚'
    } else if (diffHours < 24) {
      return `${Math.floor(diffHours)}小时前`
    } else if (diffDays < 7) {
      return `${Math.floor(diffDays)}天前`
    } else {
      return date.toLocaleDateString('zh-CN')
    }
  }

  const getStatusIcon = (status: CreditTaskStatus) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'pending':
      case 'scheduled':
        return <Clock className="h-4 w-4 text-blue-500" />
      case 'executing':
        return <Play className="h-4 w-4 text-orange-500" />
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />
      case 'cancelled':
        return <Pause className="h-4 w-4 text-gray-500" />
      case 'skipped':
        return <AlertCircle className="h-4 w-4 text-yellow-500" />
      default:
        return <Clock className="h-4 w-4 text-gray-500" />
    }
  }

  const getStatusVariant = (status: CreditTaskStatus) => {
    switch (status) {
      case 'completed':
        return 'default'
      case 'pending':
      case 'scheduled':
        return 'secondary'
      case 'executing':
        return 'outline'
      case 'failed':
        return 'destructive'
      case 'cancelled':
      case 'skipped':
        return 'secondary'
      default:
        return 'secondary'
    }
  }

  if (loading && tasks.length === 0) {
    return (
      <Card className={`w-full ${className}`}>
        <CardHeader>
          <Skeleton className="h-6 w-32" />
        </CardHeader>
        <CardContent className="space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-center space-x-4 p-4 border rounded-lg">
              <Skeleton className="h-8 w-8 rounded-full" />
              <div className="flex-1 space-y-2">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/2" />
              </div>
              <Skeleton className="h-6 w-16" />
              <Skeleton className="h-6 w-12" />
            </div>
          ))}
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className={`w-full ${className}`}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Award className="h-5 w-5" />
            积分任务
          </CardTitle>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleRefresh}
            disabled={loading}
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* 筛选器 */}
        {showFilters && (
          <div className="flex items-center gap-4 p-4 bg-muted/50 rounded-lg flex-wrap">
            <Filter className="h-4 w-4 text-muted-foreground" />
            
            <Select
              value={filters.status || 'all'}
              onValueChange={(value) => handleFilterChange('status', value === 'all' ? undefined : value)}
            >
              <SelectTrigger className="w-32">
                <SelectValue placeholder="状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="pending">等待执行</SelectItem>
                <SelectItem value="executing">执行中</SelectItem>
                <SelectItem value="completed">已完成</SelectItem>
                <SelectItem value="failed">失败</SelectItem>
              </SelectContent>
            </Select>
            
            <Select
              value={filters.task_type || 'all'}
              onValueChange={(value) => handleFilterChange('task_type', value === 'all' ? undefined : value)}
            >
              <SelectTrigger className="w-40">
                <SelectValue placeholder="任务类型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部类型</SelectItem>
                <SelectItem value="letter_created">创建信件</SelectItem>
                <SelectItem value="ai_interaction">AI互动</SelectItem>
                <SelectItem value="courier_delivery">信使送达</SelectItem>
                <SelectItem value="museum_submit">博物馆提交</SelectItem>
                <SelectItem value="public_like">公开信点赞</SelectItem>
              </SelectContent>
            </Select>
            
            <div className="text-sm text-muted-foreground">
              共 {total} 个任务
            </div>
          </div>
        )}

        {/* 错误状态 */}
        {error && (
          <div className="text-center p-4 text-destructive text-sm">
            {error}
          </div>
        )}

        {/* 任务列表 */}
        <div className="space-y-3">
          {tasks.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              暂无积分任务
            </div>
          ) : (
            tasks.map((task) => (
              <div
                key={task.id}
                className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
              >
                <div className="flex items-center space-x-3 flex-1">
                  <div className="flex-shrink-0">
                    {getStatusIcon(task.status)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-sm font-medium">
                        {TASK_DESCRIPTIONS[task.task_type] || task.description}
                      </span>
                      <Badge variant="outline" className="text-xs">
                        {formatPoints(task.points)} 积分
                      </Badge>
                    </div>
                    
                    <div className="text-xs text-muted-foreground space-y-1">
                      <div>创建时间: {formatDate(task.created_at)}</div>
                      {task.executed_at && (
                        <div>执行时间: {formatDate(task.executed_at)}</div>
                      )}
                      {task.completed_at && (
                        <div>完成时间: {formatDate(task.completed_at)}</div>
                      )}
                      {task.error_message && (
                        <div className="text-red-500 text-xs">
                          错误: {task.error_message}
                        </div>
                      )}
                    </div>
                    
                    {task.reference && (
                      <div className="text-xs text-muted-foreground opacity-75 mt-1">
                        关联: {task.reference.slice(-8)}
                      </div>
                    )}
                  </div>
                </div>
                
                <div className="flex items-center space-x-2 flex-shrink-0">
                  <Badge
                    variant={getStatusVariant(task.status)}
                    className="text-xs"
                  >
                    {TASK_STATUS_DESCRIPTIONS[task.status]}
                  </Badge>
                  
                  {task.priority > 5 && (
                    <Badge variant="outline" className="text-xs bg-orange-50 text-orange-700">
                      高优先级
                    </Badge>
                  )}
                  
                  {task.attempts > 1 && (
                    <Badge variant="outline" className="text-xs">
                      重试 {task.attempts}
                    </Badge>
                  )}
                </div>
              </div>
            ))
          )}
        </div>

        {/* 加载更多 */}
        {tasks.length > 0 && tasks.length < total && (
          <div className="text-center pt-4">
            <Button
              variant="outline"
              onClick={handleLoadMore}
              disabled={loading}
            >
              {loading ? (
                <>
                  <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                  加载中...
                </>
              ) : (
                `加载更多 (${tasks.length}/${total})`
              )}
            </Button>
          </div>
        )}

        {/* 任务统计 */}
        {tasks.length > 0 && (
          <div className="pt-4 border-t">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div className="text-center">
                <div className="font-medium text-green-600">
                  {tasks.filter(t => t.status === 'completed').length}
                </div>
                <div className="text-muted-foreground">已完成</div>
              </div>
              <div className="text-center">
                <div className="font-medium text-blue-600">
                  {tasks.filter(t => t.status === 'pending' || t.status === 'scheduled').length}
                </div>
                <div className="text-muted-foreground">待执行</div>
              </div>
              <div className="text-center">
                <div className="font-medium text-orange-600">
                  {tasks.filter(t => t.status === 'executing').length}
                </div>
                <div className="text-muted-foreground">执行中</div>
              </div>
              <div className="text-center">
                <div className="font-medium text-red-600">
                  {tasks.filter(t => t.status === 'failed').length}
                </div>
                <div className="text-muted-foreground">失败</div>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default CreditTaskList