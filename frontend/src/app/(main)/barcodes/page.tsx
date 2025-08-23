'use client'

import { useState, useEffect, useCallback } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  QrCode,
  Search,
  Clock,
  CheckCircle,
  Truck,
  Package,
  MapPin,
  Eye,
  RefreshCw,
  Plus,
  AlertCircle
} from 'lucide-react'
import { 
  BarcodeService, 
  type Barcode,
  getBarcodeStatusInfo,
  formatBarcodeCode 
} from '@/lib/services/barcode-service'
import { useAuth } from '@/contexts/auth-context-new'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import Link from 'next/link'

export default function MyBarcodesPage() {
  const { user } = useAuth()
  const [myBarcodes, setMyBarcodes] = useState<Barcode[]>([])
  const [searchCode, setSearchCode] = useState('')
  const [searchResult, setSearchResult] = useState<Barcode | null>(null)
  const [loading, setLoading] = useState(true)
  const [searching, setSearching] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [searchError, setSearchError] = useState<string | null>(null)

  // 加载我的条码
  const loadMyBarcodes = useCallback(async () => {
    if (!user) return
    
    try {
      setLoading(true)
      setError(null)
      const response = await BarcodeService.getMyBarcodes({
        limit: 50,
        sort_by: 'createdAt',
        sort_order: 'desc'
      })
      setMyBarcodes((response as any).data?.data || [])
    } catch (err: any) {
      setError(err.message || '加载条码列表失败')
    } finally {
      setLoading(false)
    }
  }, [user])

  // 搜索条码
  const searchBarcode = async () => {
    if (!searchCode.trim()) {
      setSearchError('请输入条码编号')
      return
    }

    try {
      setSearching(true)
      setSearchError(null)
      setSearchResult(null)
      
      const response = await BarcodeService.getBarcodeByCode(searchCode.trim().toUpperCase())
      setSearchResult(response.data)
    } catch (err: any) {
      if (err.message?.includes('404') || err.message?.includes('not found')) {
        setSearchError('条码不存在')
      } else {
        setSearchError(err.message || '搜索失败')
      }
    } finally {
      setSearching(false)
    }
  }

  // 处理搜索输入回车
  const handleSearchKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      searchBarcode()
    }
  }

  useEffect(() => {
    if (user) {
      loadMyBarcodes()
    }
  }, [user, loadMyBarcodes])

  // 获取状态统计
  const statusCounts = myBarcodes.reduce((acc, barcode) => {
    acc[barcode.status] = (acc[barcode.status] || 0) + 1
    return acc
  }, {} as Record<string, number>)

  const statusStats = [
    {
      status: 'unactivated',
      label: '未激活',
      count: statusCounts.unactivated || 0,
      icon: Clock,
      color: 'text-gray-600',
      bgColor: 'bg-gray-50'
    },
    {
      status: 'bound',
      label: '已绑定',
      count: statusCounts.bound || 0,
      icon: CheckCircle,
      color: 'text-blue-600',
      bgColor: 'bg-blue-50'
    },
    {
      status: 'in_transit',
      label: '投递中',
      count: statusCounts.in_transit || 0,
      icon: Truck,
      color: 'text-yellow-600',
      bgColor: 'bg-yellow-50'
    },
    {
      status: 'delivered',
      label: '已送达',
      count: statusCounts.delivered || 0,
      icon: Package,
      color: 'text-green-600',
      bgColor: 'bg-green-50'
    }
  ]

  if (!user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            请先登录以查看您的条码
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
            <QrCode className="h-8 w-8" />
            我的条码
          </h1>
          <p className="text-gray-600 mt-2">查看和跟踪您的信件条码状态</p>
        </div>
        
        <div className="flex gap-2">
          <Button variant="outline" onClick={loadMyBarcodes}>
            <RefreshCw className="w-4 h-4 mr-2" />
            刷新
          </Button>
          <Link href="/letters/write">
            <Button>
              <Plus className="w-4 h-4 mr-2" />
              写新信件
            </Button>
          </Link>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {statusStats.map((stat, index) => (
          <Card key={index}>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">{stat.label}</p>
                  <p className="text-2xl font-bold">{stat.count}</p>
                </div>
                <div className={`p-2 rounded-lg ${stat.bgColor}`}>
                  <stat.icon className={`h-5 w-5 ${stat.color}`} />
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* 主要内容标签页 */}
      <Tabs defaultValue="my-barcodes" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="my-barcodes" className="flex items-center gap-2">
            <QrCode className="h-4 w-4" />
            我的条码 ({myBarcodes.length})
          </TabsTrigger>
          <TabsTrigger value="search" className="flex items-center gap-2">
            <Search className="h-4 w-4" />
            查询条码
          </TabsTrigger>
        </TabsList>

        {/* 我的条码列表 */}
        <TabsContent value="my-barcodes">
          <Card>
            <CardHeader>
              <CardTitle>我的条码列表</CardTitle>
              <CardDescription>
                您创建的所有条码及其状态
              </CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="space-y-4">
                  {[...Array(3)].map((_, i) => (
                    <div key={i} className="animate-pulse p-4 border rounded-lg">
                      <div className="flex justify-between items-start">
                        <div className="space-y-2">
                          <div className="h-4 bg-gray-200 rounded w-32"></div>
                          <div className="h-3 bg-gray-200 rounded w-48"></div>
                        </div>
                        <div className="h-6 bg-gray-200 rounded w-16"></div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : error ? (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              ) : myBarcodes.length > 0 ? (
                <div className="space-y-4">
                  {myBarcodes.map((barcode) => {
                    const statusInfo = getBarcodeStatusInfo(barcode.status)
                    return (
                      <div key={barcode.id} className="p-4 border rounded-lg hover:shadow-md transition-shadow">
                        <div className="flex justify-between items-start mb-3">
                          <div className="flex-1">
                            <div className="flex items-center gap-3 mb-2">
                              <span className="font-mono font-bold text-lg">
                                {formatBarcodeCode(barcode.code)}
                              </span>
                              <Badge variant="outline" className={statusInfo.color}>
                                {statusInfo.label}
                              </Badge>
                            </div>
                            <p className="text-sm text-gray-600">{statusInfo.description}</p>
                          </div>
                          
                          <div className="flex gap-2">
                            {barcode.png_url && (
                              <Button 
                                variant="outline" 
                                size="sm"
                                onClick={() => window.open(barcode.png_url, '_blank')}
                              >
                                <Eye className="w-4 h-4" />
                              </Button>
                            )}
                          </div>
                        </div>
                        
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                          <div className="space-y-1">
                            <p className="font-medium text-gray-700">基本信息</p>
                            <p className="text-gray-600">
                              创建时间: {formatDistanceToNow(new Date(barcode.createdAt), { 
                                addSuffix: true, 
                                locale: zhCN 
                              })}
                            </p>
                            <p className="text-gray-600">
                              来源: {{
                                'write-page': '写信页面',
                                'admin': '管理员生成',
                                'batch-request': '批量生成',
                                'store': '信封商店'
                              }[barcode.source] || barcode.source}
                            </p>
                          </div>
                          
                          <div className="space-y-1">
                            <p className="font-medium text-gray-700">绑定状态</p>
                            {barcode.letter_id ? (
                              <>
                                <p className="text-gray-600">信件ID: {barcode.letter_id}</p>
                                {barcode.bound_at && (
                                  <p className="text-gray-600">
                                    绑定时间: {formatDistanceToNow(new Date(barcode.bound_at), { 
                                      addSuffix: true, 
                                      locale: zhCN 
                                    })}
                                  </p>
                                )}
                              </>
                            ) : (
                              <p className="text-gray-600">尚未绑定信件</p>
                            )}
                          </div>
                          
                          <div className="space-y-1">
                            <p className="font-medium text-gray-700">收件信息</p>
                            {barcode.recipient_code ? (
                              <>
                                <p className="text-gray-600">收件编码: {barcode.recipient_code}</p>
                                <div className="flex items-center gap-1 text-gray-600">
                                  <MapPin className="w-3 h-3" />
                                  <span>位置已设定</span>
                                </div>
                              </>
                            ) : (
                              <p className="text-gray-600">收件信息待设定</p>
                            )}
                          </div>
                        </div>
                      </div>
                    )
                  })}
                </div>
              ) : (
                <div className="text-center py-12">
                  <QrCode className="h-16 w-16 mx-auto mb-4 text-gray-300" />
                  <h3 className="text-lg font-semibold mb-2">还没有条码</h3>
                  <p className="text-gray-600 mb-6">写一封新信件来获取您的第一个条码</p>
                  <Link href="/letters/write">
                    <Button>
                      <Plus className="w-4 h-4 mr-2" />
                      写新信件
                    </Button>
                  </Link>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* 查询条码 */}
        <TabsContent value="search">
          <Card>
            <CardHeader>
              <CardTitle>查询条码状态</CardTitle>
              <CardDescription>
                输入条码编号查询任意条码的投递状态
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* 搜索输入 */}
              <div className="flex gap-2">
                <div className="flex-1">
                  <Input
                    placeholder="输入条码编号，如：OPP-BJFU-5F3D-01"
                    value={searchCode}
                    onChange={(e) => setSearchCode(e.target.value)}
                    onKeyPress={handleSearchKeyPress}
                    className="font-mono"
                  />
                </div>
                <Button onClick={searchBarcode} disabled={searching}>
                  {searching ? (
                    <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                  ) : (
                    <Search className="w-4 h-4 mr-2" />
                  )}
                  查询
                </Button>
              </div>

              {/* 搜索错误 */}
              {searchError && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>{searchError}</AlertDescription>
                </Alert>
              )}

              {/* 搜索结果 */}
              {searchResult && (
                <div className="p-6 border rounded-lg bg-gray-50">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold">条码信息</h3>
                    <Badge 
                      variant="outline" 
                      className={getBarcodeStatusInfo(searchResult.status).color}
                    >
                      {getBarcodeStatusInfo(searchResult.status).label}
                    </Badge>
                  </div>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-3">
                      <div>
                        <p className="font-medium text-gray-700">条码编号</p>
                        <p className="font-mono">{formatBarcodeCode(searchResult.code)}</p>
                      </div>
                      <div>
                        <p className="font-medium text-gray-700">当前状态</p>
                        <p className="text-gray-600">
                          {getBarcodeStatusInfo(searchResult.status).description}
                        </p>
                      </div>
                      <div>
                        <p className="font-medium text-gray-700">创建时间</p>
                        <p className="text-gray-600">
                          {formatDistanceToNow(new Date(searchResult.createdAt), { 
                            addSuffix: true, 
                            locale: zhCN 
                          })}
                        </p>
                      </div>
                    </div>
                    
                    <div className="space-y-3">
                      {searchResult.letter_id && (
                        <div>
                          <p className="font-medium text-gray-700">关联信件</p>
                          <p className="text-gray-600">{searchResult.letter_id}</p>
                        </div>
                      )}
                      {searchResult.bound_at && (
                        <div>
                          <p className="font-medium text-gray-700">绑定时间</p>
                          <p className="text-gray-600">
                            {formatDistanceToNow(new Date(searchResult.bound_at), { 
                              addSuffix: true, 
                              locale: zhCN 
                            })}
                          </p>
                        </div>
                      )}
                      {searchResult.recipient_code && (
                        <div>
                          <p className="font-medium text-gray-700">收件编码</p>
                          <p className="font-mono text-gray-600">{searchResult.recipient_code}</p>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              )}

              {/* 使用说明 */}
              <div className="p-4 bg-blue-50 rounded-lg">
                <h4 className="font-medium text-blue-900 mb-2">使用说明</h4>
                <ul className="text-sm text-blue-700 space-y-1">
                  <li>• 支持新格式条码：OPP-XXXX-XXXX-XX</li>
                  <li>• 支持旧格式条码：OP + 6-10位字母数字</li>
                  <li>• 查询结果显示条码的实时投递状态</li>
                  <li>• 您可以查询任意条码，不仅限于自己的</li>
                </ul>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}