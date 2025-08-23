'use client'

import { useState, useEffect, useMemo } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { DataTable } from '@/components/ui/data-table'
import { Breadcrumb, ADMIN_BREADCRUMBS } from '@/components/ui/breadcrumb'
import { 
  QrCode,
  BarChart3,
  Search,
  Filter,
  Download,
  AlertTriangle,
  CheckCircle,
  Clock,
  Truck,
  Package,
  Ban,
  RefreshCw,
  Eye,
  MoreVertical,
  Trash2
} from 'lucide-react'
import { 
  BarcodeService, 
  type Barcode, 
  type BarcodeStats, 
  type BarcodeStatus,
  type BarcodeSource,
  getBarcodeStatusInfo 
} from '@/lib/services/barcode-service'
import { usePermission } from '@/hooks/use-permission'
import { ColumnDef } from '@tanstack/react-table'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { toast } from '@/components/ui/use-toast'

export default function AdminBarcodesPage() {
  const { hasPermission } = usePermission()
  const [barcodes, setBarcodes] = useState<Barcode[]>([])
  const [stats, setStats] = useState<BarcodeStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  
  // 筛选和分页状态
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<BarcodeStatus | 'all'>('all')
  const [sourceFilter, setSourceFilter] = useState<BarcodeSource | 'all'>('all')
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [totalCount, setTotalCount] = useState(0)

  // 加载数据
  const loadBarcodes = async (page = 1, limit = 20) => {
    try {
      setLoading(true)
      const response = await BarcodeService.listBarcodes({
        page,
        limit,
        status: statusFilter === 'all' ? undefined : statusFilter,
        source: sourceFilter === 'all' ? undefined : sourceFilter,
        search: searchTerm || undefined,
        sort_by: 'createdAt',
        sort_order: 'desc'
      })
      
      setBarcodes(((response as any)?.data?.data || (response as any)?.data).data)
      setTotalCount(((response as any)?.data?.data || (response as any)?.data).meta.total)
      setCurrentPage(page)
      setPageSize(limit)
    } catch (err: any) {
      setError(err.message || '加载条码数据失败')
    } finally {
      setLoading(false)
    }
  }

  // 加载统计数据
  const loadStats = async () => {
    try {
      const response = await BarcodeService.getBarcodeStats()
      setStats(((response as any)?.data?.data || (response as any)?.data))
    } catch (err) {
      console.error('Failed to load barcode stats:', err)
    }
  }

  // 作废条码
  const handleVoidBarcode = async (code: string) => {
    if (!confirm('确定要作废此条码吗？此操作不可撤销。')) {
      return
    }

    try {
      await BarcodeService.voidBarcode(code)
      toast({
        title: '操作成功',
        description: `条码 ${code} 已被作废`
      })
      loadBarcodes(currentPage, pageSize)
      loadStats()
    } catch (err: any) {
      toast({
        title: '操作失败',
        description: err.message || '作废条码失败',
        variant: 'destructive'
      })
    }
  }

  useEffect(() => {
    loadBarcodes()
    loadStats()
  }, [])

  // 筛选变化时重新加载
  useEffect(() => {
    loadBarcodes(1, pageSize)
  }, [searchTerm, statusFilter, sourceFilter])

  // 表格列定义
  const columns: ColumnDef<Barcode>[] = useMemo(() => [
    {
      accessorKey: 'code',
      header: '条码编号',
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <QrCode className="h-4 w-4 text-gray-400" />
          <span className="font-mono text-sm">{row.getValue('code')}</span>
        </div>
      ),
    },
    {
      accessorKey: 'status',
      header: '状态',
      cell: ({ row }) => {
        const status = row.getValue('status') as BarcodeStatus
        const statusInfo = getBarcodeStatusInfo(status)
        return (
          <Badge variant="outline" className={statusInfo.color}>
            {statusInfo.label}
          </Badge>
        )
      },
    },
    {
      accessorKey: 'source',
      header: '来源',
      cell: ({ row }) => {
        const sourceMap: Record<BarcodeSource, string> = {
          'write-page': '写信页面',
          'admin': '管理员生成',
          'batch-request': '批量生成',
          'store': '信封商店'
        }
        const source = row.getValue('source') as BarcodeSource
        return <span className="text-sm">{sourceMap[source] || source}</span>
      },
    },
    {
      accessorKey: 'letter_id',
      header: '绑定信件',
      cell: ({ row }) => {
        const letterId = row.getValue('letter_id') as string
        return letterId ? (
          <span className="text-sm text-blue-600">{letterId}</span>
        ) : (
          <span className="text-sm text-gray-400">未绑定</span>
        )
      },
    },
    {
      accessorKey: 'createdAt',
      header: '创建时间',
      cell: ({ row }) => {
        const date = new Date(row.getValue('createdAt'))
        return (
          <span className="text-sm text-gray-600">
            {formatDistanceToNow(date, { addSuffix: true, locale: zhCN })}
          </span>
        )
      },
    },
    {
      accessorKey: 'bound_at',
      header: '绑定时间',
      cell: ({ row }) => {
        const boundAt = row.getValue('bound_at') as string
        return boundAt ? (
          <span className="text-sm text-gray-600">
            {formatDistanceToNow(new Date(boundAt), { addSuffix: true, locale: zhCN })}
          </span>
        ) : (
          <span className="text-sm text-gray-400">未绑定</span>
        )
      },
    },
    {
      id: 'actions',
      header: '操作',
      cell: ({ row }) => {
        const barcode = row.original
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={() => {
                  // 查看详情逻辑
                  window.open(`/admin/barcodes/${barcode.code}`, '_blank')
                }}
              >
                <Eye className="mr-2 h-4 w-4" />
                查看详情
              </DropdownMenuItem>
              {hasPermission('admin.barcodes.void') && barcode.status !== 'voided' && (
                <DropdownMenuItem
                  onClick={() => handleVoidBarcode(barcode.code)}
                  className="text-red-600"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  作废条码
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ], [hasPermission])

  // 检查权限
  if (!hasPermission('admin.barcodes.read')) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            您没有权限访问条码管理功能
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  // 统计卡片数据
  const statsCards = [
    {
      title: '总条码数',
      value: stats?.total_generated || 0,
      icon: QrCode,
      color: 'text-blue-600',
      bgColor: 'bg-blue-50'
    },
    {
      title: '未激活',
      value: stats?.unactivated || 0,
      icon: Clock,
      color: 'text-gray-600',
      bgColor: 'bg-gray-50'
    },
    {
      title: '已绑定',
      value: stats?.bound || 0,
      icon: CheckCircle,
      color: 'text-blue-600',
      bgColor: 'bg-blue-50'
    },
    {
      title: '投递中',
      value: stats?.in_transit || 0,
      icon: Truck,
      color: 'text-yellow-600',
      bgColor: 'bg-yellow-50'
    },
    {
      title: '已送达',
      value: stats?.delivered || 0,
      icon: Package,
      color: 'text-green-600',
      bgColor: 'bg-green-50'
    },
    {
      title: '已作废',
      value: stats?.voided || 0,
      icon: Ban,
      color: 'text-red-600',
      bgColor: 'bg-red-50'
    }
  ]

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 面包屑导航 */}
      <Breadcrumb items={[...ADMIN_BREADCRUMBS.root, { label: '条码管理', href: '/admin/barcodes' }]} />
      
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">条码管理</h1>
          <p className="text-gray-600 mt-2">管理和监控系统中的所有条码</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => {
            loadBarcodes(currentPage, pageSize)
            loadStats()
          }}>
            <RefreshCw className="w-4 h-4 mr-2" />
            刷新
          </Button>
          {hasPermission('admin.barcodes.export') && (
            <Button variant="outline">
              <Download className="w-4 h-4 mr-2" />
              导出数据
            </Button>
          )}
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-6 gap-4 mb-8">
        {statsCards.map((card, index) => (
          <Card key={index}>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">{card.title}</p>
                  <p className="text-2xl font-bold">{card.value.toLocaleString()}</p>
                </div>
                <div className={`p-2 rounded-lg ${card.bgColor}`}>
                  <card.icon className={`h-5 w-5 ${card.color}`} />
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* 错误提示 */}
      {error && (
        <Alert variant="destructive" className="mb-6">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 筛选区域 */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Filter className="h-5 w-5" />
            筛选和搜索
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">搜索条码</label>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  placeholder="输入条码编号或信件ID..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
            
            <div className="space-y-2">
              <label className="text-sm font-medium">状态筛选</label>
              <Select value={statusFilter} onValueChange={(value) => setStatusFilter(value as any)}>
                <SelectTrigger>
                  <SelectValue placeholder="选择状态" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部状态</SelectItem>
                  <SelectItem value="unactivated">未激活</SelectItem>
                  <SelectItem value="bound">已绑定</SelectItem>
                  <SelectItem value="in_transit">投递中</SelectItem>
                  <SelectItem value="delivered">已送达</SelectItem>
                  <SelectItem value="expired">已过期</SelectItem>
                  <SelectItem value="voided">已作废</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">来源筛选</label>
              <Select value={sourceFilter} onValueChange={(value) => setSourceFilter(value as any)}>
                <SelectTrigger>
                  <SelectValue placeholder="选择来源" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部来源</SelectItem>
                  <SelectItem value="write-page">写信页面</SelectItem>
                  <SelectItem value="admin">管理员生成</SelectItem>
                  <SelectItem value="batch-request">批量生成</SelectItem>
                  <SelectItem value="store">信封商店</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-end">
              <Button 
                variant="outline" 
                onClick={() => {
                  setSearchTerm('')
                  setStatusFilter('all')
                  setSourceFilter('all')
                }}
                className="w-full"
              >
                重置筛选
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 条码数据表 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <QrCode className="h-5 w-5" />
            条码列表
          </CardTitle>
          <CardDescription>
            共找到 {totalCount} 个条码
          </CardDescription>
        </CardHeader>
        <CardContent>
          <DataTable
            columns={columns}
            data={barcodes}
            loading={loading}
            pagination={{
              pageSize: pageSize,
              pageIndex: currentPage - 1,
              total: totalCount
            }}
            onPaginationChange={(pagination) => {
              loadBarcodes(pagination.pageIndex + 1, pagination.pageSize)
            }}
            showSearch={false}
            className="border-none"
          />
        </CardContent>
      </Card>
    </div>
  )
}