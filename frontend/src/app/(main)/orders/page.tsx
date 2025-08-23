'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { 
  Package,
  Clock,
  CheckCircle,
  Truck,
  X,
  Search,
  Filter,
  RefreshCw,
  Eye,
  MessageSquare,
  ArrowRight,
  ShoppingBag,
  CreditCard,
  MapPin,
  Calendar
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { apiClient } from '@/lib/api-client'
import { formatDistanceToNow, format } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import Link from 'next/link'
import { toast } from '@/components/ui/use-toast'

// 订单状态类型
type OrderStatus = 'pending' | 'paid' | 'processing' | 'shipped' | 'delivered' | 'completed' | 'cancelled' | 'refunded'
type PaymentStatus = 'pending' | 'paid' | 'failed' | 'refunded'

// 订单接口
interface Order {
  id: string
  order_no: string
  total_amount: number
  payment_amount: number
  status: OrderStatus
  payment_status: PaymentStatus
  payment_method?: string
  created_at: string
  updated_at: string
  
  // 配送信息
  delivery_info?: {
    name: string
    phone: string
    address?: string
    method: 'delivery' | 'pickup'
    pickup_location?: string
    pickup_time?: string
  }
  
  // 订单项
  items: Array<{
    id: string
    product_id: string
    product_name: string
    product_image?: string
    quantity: number
    unit_price: number
    total_price: number
  }>
  
  // 物流信息
  tracking_number?: string
  estimated_delivery?: string
  delivered_at?: string
}

export default function OrdersPage() {
  const { user } = useAuth()
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  
  // 筛选状态
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<OrderStatus | 'all'>('all')
  const [dateRange, setDateRange] = useState<'all' | '7d' | '30d' | '90d'>('all')

  // 加载订单列表
  const loadOrders = async () => {
    try {
      setLoading(true)
      setError(null)
      
      const params = new URLSearchParams()
      if (statusFilter !== 'all') params.append('status', statusFilter)
      if (searchTerm) params.append('search', searchTerm)
      if (dateRange !== 'all') {
        const days = { '7d': 7, '30d': 30, '90d': 90 }[dateRange]
        const since = new Date()
        since.setDate(since.getDate() - days!)
        params.append('since', since.toISOString())
      }
      
      const response = await apiClient.get(`/api/v1/shop/orders?${params.toString()}`)
      setOrders(((response as any)?.data?.data || (response as any)?.data)?.data || [])
    } catch (err: any) {
      setError(err.message || '加载订单失败')
    } finally {
      setLoading(false)
    }
  }

  // 取消订单
  const cancelOrder = async (orderId: string) => {
    if (!confirm('确定要取消此订单吗？')) return

    try {
      await apiClient.patch(`/api/v1/shop/orders/${orderId}/cancel`)
      toast({
        title: '订单已取消',
        description: '您的订单已成功取消'
      })
      loadOrders()
    } catch (err: any) {
      toast({
        title: '取消失败',
        description: err.message || '取消订单失败',
        variant: 'destructive'
      })
    }
  }

  // 确认收货
  const confirmDelivery = async (orderId: string) => {
    if (!confirm('确认已收到商品？')) return

    try {
      await apiClient.patch(`/api/v1/shop/orders/${orderId}/confirm-delivery`)
      toast({
        title: '确认收货成功',
        description: '订单已完成'
      })
      loadOrders()
    } catch (err: any) {
      toast({
        title: '确认失败',
        description: err.message || '确认收货失败',
        variant: 'destructive'
      })
    }
  }

  useEffect(() => {
    if (user) {
      loadOrders()
    }
  }, [user])

  useEffect(() => {
    const timer = setTimeout(() => {
      loadOrders()
    }, 500) // 防抖搜索
    return () => clearTimeout(timer)
  }, [searchTerm, statusFilter, dateRange])

  // 获取状态信息
  const getOrderStatusInfo = (status: OrderStatus) => {
    const statusMap = {
      pending: { label: '待支付', color: 'bg-yellow-100 text-yellow-800', icon: Clock },
      paid: { label: '已支付', color: 'bg-blue-100 text-blue-800', icon: CreditCard },
      processing: { label: '处理中', color: 'bg-blue-100 text-blue-800', icon: Package },
      shipped: { label: '已发货', color: 'bg-purple-100 text-purple-800', icon: Truck },
      delivered: { label: '已送达', color: 'bg-green-100 text-green-800', icon: CheckCircle },
      completed: { label: '已完成', color: 'bg-green-100 text-green-800', icon: CheckCircle },
      cancelled: { label: '已取消', color: 'bg-gray-100 text-gray-800', icon: X },
      refunded: { label: '已退款', color: 'bg-red-100 text-red-800', icon: X }
    }
    return statusMap[status] || statusMap.pending
  }

  const getPaymentStatusInfo = (status: PaymentStatus) => {
    const statusMap = {
      pending: { label: '待支付', color: 'bg-yellow-100 text-yellow-800' },
      paid: { label: '已支付', color: 'bg-green-100 text-green-800' },
      failed: { label: '支付失败', color: 'bg-red-100 text-red-800' },
      refunded: { label: '已退款', color: 'bg-red-100 text-red-800' }
    }
    return statusMap[status] || statusMap.pending
  }

  // 统计数据
  const statusCounts = orders.reduce((acc, order) => {
    acc[order.status] = (acc[order.status] || 0) + 1
    return acc
  }, {} as Record<string, number>)

  if (!user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertDescription>
            请先登录以查看您的订单
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
            <Package className="h-8 w-8" />
            我的订单
          </h1>
          <p className="text-gray-600 mt-2">查看和管理您的购买订单</p>
        </div>
        
        <div className="flex gap-2">
          <Button variant="outline" onClick={loadOrders}>
            <RefreshCw className="w-4 h-4 mr-2" />
            刷新
          </Button>
          <Link href="/shop">
            <Button>
              <ShoppingBag className="w-4 h-4 mr-2" />
              继续购物
            </Button>
          </Link>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">全部订单</p>
                <p className="text-2xl font-bold">{orders.length}</p>
              </div>
              <Package className="h-5 w-5 text-gray-400" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">待支付</p>
                <p className="text-2xl font-bold">{statusCounts.pending || 0}</p>
              </div>
              <Clock className="h-5 w-5 text-yellow-500" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">待收货</p>
                <p className="text-2xl font-bold">{(statusCounts.shipped || 0) + (statusCounts.delivered || 0)}</p>
              </div>
              <Truck className="h-5 w-5 text-blue-500" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">已完成</p>
                <p className="text-2xl font-bold">{statusCounts.completed || 0}</p>
              </div>
              <CheckCircle className="h-5 w-5 text-green-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 筛选区域 */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Filter className="h-5 w-5" />
            筛选订单
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">搜索订单</label>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  placeholder="订单号或商品名称..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
            
            <div className="space-y-2">
              <label className="text-sm font-medium">订单状态</label>
              <Select value={statusFilter} onValueChange={(value) => setStatusFilter(value as any)}>
                <SelectTrigger>
                  <SelectValue placeholder="选择状态" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部状态</SelectItem>
                  <SelectItem value="pending">待支付</SelectItem>
                  <SelectItem value="paid">已支付</SelectItem>
                  <SelectItem value="processing">处理中</SelectItem>
                  <SelectItem value="shipped">已发货</SelectItem>
                  <SelectItem value="delivered">已送达</SelectItem>
                  <SelectItem value="completed">已完成</SelectItem>
                  <SelectItem value="cancelled">已取消</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">时间范围</label>
              <Select value={dateRange} onValueChange={(value) => setDateRange(value as any)}>
                <SelectTrigger>
                  <SelectValue placeholder="选择时间" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部时间</SelectItem>
                  <SelectItem value="7d">最近7天</SelectItem>
                  <SelectItem value="30d">最近30天</SelectItem>
                  <SelectItem value="90d">最近90天</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 订单列表 */}
      <div className="space-y-4">
        {loading ? (
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <Card key={i} className="animate-pulse">
                <CardContent className="p-6">
                  <div className="space-y-4">
                    <div className="flex justify-between items-start">
                      <div className="space-y-2">
                        <div className="h-4 bg-gray-200 rounded w-32"></div>
                        <div className="h-3 bg-gray-200 rounded w-24"></div>
                      </div>
                      <div className="h-6 bg-gray-200 rounded w-16"></div>
                    </div>
                    <div className="h-20 bg-gray-200 rounded"></div>
                    <div className="flex justify-between">
                      <div className="h-4 bg-gray-200 rounded w-20"></div>
                      <div className="h-8 bg-gray-200 rounded w-16"></div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        ) : error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : orders.length > 0 ? (
          orders.map((order) => {
            const statusInfo = getOrderStatusInfo(order.status)
            const paymentInfo = getPaymentStatusInfo(order.payment_status)
            
            return (
              <Card key={order.id} className="hover:shadow-md transition-shadow">
                <CardContent className="p-6">
                  {/* 订单头部 */}
                  <div className="flex justify-between items-start mb-4">
                    <div className="space-y-1">
                      <div className="flex items-center gap-3">
                        <span className="font-mono font-semibold">#{order.order_no}</span>
                        <Badge variant="outline" className={statusInfo.color}>
                          {statusInfo.label}
                        </Badge>
                        <Badge variant="outline" className={paymentInfo.color}>
                          {paymentInfo.label}
                        </Badge>
                      </div>
                      <p className="text-sm text-gray-600">
                        下单时间: {format(new Date(order.created_at), 'yyyy-MM-dd HH:mm')}
                      </p>
                    </div>
                    
                    <div className="text-right">
                      <p className="text-lg font-bold">¥{order.total_amount.toFixed(2)}</p>
                      {order.payment_method && (
                        <p className="text-sm text-gray-600">
                          {{
                            'alipay': '支付宝',
                            'wechat': '微信支付',
                            'card': '银行卡',
                            'credit': '积分支付'
                          }[order.payment_method] || order.payment_method}
                        </p>
                      )}
                    </div>
                  </div>

                  {/* 商品列表 */}
                  <div className="space-y-3 mb-4">
                    {order.items.map((item) => (
                      <div key={item.id} className="flex items-center gap-4 p-3 bg-gray-50 rounded-lg">
                        {item.product_image && (
                          <img 
                            src={item.product_image} 
                            alt={item.product_name}
                            className="w-12 h-12 object-cover rounded"
                          />
                        )}
                        <div className="flex-1">
                          <p className="font-medium">{item.product_name}</p>
                          <p className="text-sm text-gray-600">
                            ¥{item.unit_price.toFixed(2)} × {item.quantity}
                          </p>
                        </div>
                        <div className="text-right">
                          <p className="font-medium">¥{item.total_price.toFixed(2)}</p>
                        </div>
                      </div>
                    ))}
                  </div>

                  {/* 配送信息 */}
                  {order.delivery_info && (
                    <div className="p-3 bg-blue-50 rounded-lg mb-4">
                      <div className="flex items-start gap-2">
                        <MapPin className="h-4 w-4 text-blue-600 mt-1" />
                        <div className="flex-1">
                          <p className="font-medium text-blue-900">
                            {order.delivery_info.method === 'delivery' ? '配送地址' : '自提信息'}
                          </p>
                          <p className="text-sm text-blue-700">
                            收货人: {order.delivery_info.name} {order.delivery_info.phone}
                          </p>
                          {order.delivery_info.address && (
                            <p className="text-sm text-blue-700">
                              地址: {order.delivery_info.address}
                            </p>
                          )}
                          {order.delivery_info.pickup_location && (
                            <p className="text-sm text-blue-700">
                              自提点: {order.delivery_info.pickup_location}
                            </p>
                          )}
                        </div>
                      </div>
                    </div>
                  )}

                  {/* 物流信息 */}
                  {order.tracking_number && (
                    <div className="p-3 bg-purple-50 rounded-lg mb-4">
                      <div className="flex items-center gap-2">
                        <Truck className="h-4 w-4 text-purple-600" />
                        <span className="font-medium text-purple-900">物流追踪</span>
                        <span className="font-mono text-purple-700">{order.tracking_number}</span>
                      </div>
                      {order.estimated_delivery && (
                        <p className="text-sm text-purple-700 mt-1">
                          预计送达: {format(new Date(order.estimated_delivery), 'MM-dd HH:mm')}
                        </p>
                      )}
                    </div>
                  )}

                  {/* 操作按钮 */}
                  <div className="flex justify-between items-center pt-4 border-t">
                    <div className="text-sm text-gray-500">
                      更新时间: {formatDistanceToNow(new Date(order.updated_at), { addSuffix: true, locale: zhCN })}
                    </div>
                    
                    <div className="flex gap-2">
                      <Link href={`/orders/${order.id}`}>
                        <Button variant="outline" size="sm">
                          <Eye className="w-4 h-4 mr-1" />
                          详情
                        </Button>
                      </Link>
                      
                      {order.status === 'pending' && (
                        <>
                          <Link href={`/shop/checkout?order=${order.id}`}>
                            <Button size="sm">
                              <CreditCard className="w-4 h-4 mr-1" />
                              去支付
                            </Button>
                          </Link>
                          <Button 
                            variant="outline" 
                            size="sm"
                            onClick={() => cancelOrder(order.id)}
                          >
                            取消订单
                          </Button>
                        </>
                      )}
                      
                      {order.status === 'delivered' && (
                        <Button 
                          size="sm"
                          onClick={() => confirmDelivery(order.id)}
                        >
                          确认收货
                        </Button>
                      )}
                      
                      {order.status === 'completed' && (
                        <Link href={`/orders/${order.id}/review`}>
                          <Button variant="outline" size="sm">
                            <MessageSquare className="w-4 h-4 mr-1" />
                            评价
                          </Button>
                        </Link>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            )
          })
        ) : (
          <Card>
            <CardContent className="text-center py-12">
              <Package className="h-16 w-16 mx-auto mb-4 text-gray-300" />
              <h3 className="text-lg font-semibold mb-2">还没有订单</h3>
              <p className="text-gray-600 mb-6">去商店看看有什么好东西吧</p>
              <Link href="/shop">
                <Button>
                  <ShoppingBag className="w-4 h-4 mr-2" />
                  去购物
                </Button>
              </Link>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}