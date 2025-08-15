'use client'

import React, { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Calendar, TrendingUp, TrendingDown, Filter, RefreshCw } from 'lucide-react'
import { useCreditTransactions, useCreditStore } from '@/stores/credit-store'
import { formatPoints } from '@/lib/api/credit'
import { CreditHistoryParams } from '@/types/credit'

interface CreditHistoryListProps {
  showFilters?: boolean
  pageSize?: number
  className?: string
}

export function CreditHistoryList({ 
  showFilters = true, 
  pageSize = 20,
  className = '' 
}: CreditHistoryListProps) {
  const { transactions, total, page, limit, loading, error } = useCreditTransactions()
  const { fetchTransactions, clearError } = useCreditStore()
  
  const [filters, setFilters] = useState<CreditHistoryParams>({
    page: 1,
    limit: pageSize,
    type: 'all'
  })

  useEffect(() => {
    fetchTransactions(filters)
  }, [filters, fetchTransactions])

  const handleFilterChange = (key: keyof CreditHistoryParams, value: any) => {
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
    fetchTransactions({ ...filters, page: 1 })
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

  const getTransactionIcon = (type: 'earn' | 'spend') => {
    return type === 'earn' ? (
      <TrendingUp className="h-4 w-4 text-green-500" />
    ) : (
      <TrendingDown className="h-4 w-4 text-red-500" />
    )
  }

  if (loading && transactions.length === 0) {
    return (
      <Card className={`w-full ${className}`}>
        <CardHeader>
          <Skeleton className="h-6 w-32" />
        </CardHeader>
        <CardContent className="space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-center space-x-4">
              <Skeleton className="h-10 w-10 rounded-full" />
              <div className="flex-1 space-y-2">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/2" />
              </div>
              <Skeleton className="h-6 w-16" />
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
            <Calendar className="h-5 w-5" />
            积分历史
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
          <div className="flex items-center gap-4 p-4 bg-muted/50 rounded-lg">
            <Filter className="h-4 w-4 text-muted-foreground" />
            <Select
              value={filters.type || 'all'}
              onValueChange={(value) => handleFilterChange('type', value === 'all' ? undefined : value)}
            >
              <SelectTrigger className="w-32">
                <SelectValue placeholder="类型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部</SelectItem>
                <SelectItem value="earn">获得</SelectItem>
                <SelectItem value="spend">消费</SelectItem>
              </SelectContent>
            </Select>
            
            <div className="text-sm text-muted-foreground">
              共 {total} 条记录
            </div>
          </div>
        )}

        {/* 错误状态 */}
        {error && (
          <div className="text-center p-4 text-destructive text-sm">
            {error}
          </div>
        )}

        {/* 交易列表 */}
        <div className="space-y-3">
          {transactions.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              暂无积分记录
            </div>
          ) : (
            transactions.map((transaction) => (
              <div
                key={transaction.id}
                className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
              >
                <div className="flex items-center space-x-3">
                  <div className="flex-shrink-0">
                    {getTransactionIcon(transaction.type)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium">
                      {transaction.description}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {formatDate(transaction.created_at)}
                    </div>
                    {transaction.reference && (
                      <div className="text-xs text-muted-foreground opacity-75">
                        ID: {transaction.reference.slice(-8)}
                      </div>
                    )}
                  </div>
                </div>
                
                <div className="flex items-center space-x-2">
                  <Badge
                    variant={transaction.type === 'earn' ? 'default' : 'destructive'}
                    className="text-sm font-medium"
                  >
                    {transaction.type === 'earn' ? '+' : '-'}
                    {formatPoints(Math.abs(transaction.amount))}
                  </Badge>
                </div>
              </div>
            ))
          )}
        </div>

        {/* 加载更多 */}
        {transactions.length > 0 && transactions.length < total && (
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
                `加载更多 (${transactions.length}/${total})`
              )}
            </Button>
          </div>
        )}

        {/* 统计信息 */}
        {transactions.length > 0 && (
          <div className="pt-4 border-t">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div className="text-center">
                <div className="font-medium text-green-600">
                  +{formatPoints(
                    transactions
                      .filter(t => t.type === 'earn')
                      .reduce((sum, t) => sum + t.amount, 0)
                  )}
                </div>
                <div className="text-muted-foreground">累计获得</div>
              </div>
              <div className="text-center">
                <div className="font-medium text-red-600">
                  -{formatPoints(
                    Math.abs(transactions
                      .filter(t => t.type === 'spend')
                      .reduce((sum, t) => sum + t.amount, 0))
                  )}
                </div>
                <div className="text-muted-foreground">累计消费</div>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default CreditHistoryList