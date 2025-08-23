'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { BackButton } from '@/components/ui/back-button'
import { 
  Package,
  Clock,
  CheckCircle,
  Truck,
  X,
  MapPin,
  Phone,
  CreditCard,
  Calendar,
  ArrowRight,
  Copy,
  MessageSquare,
  RefreshCw,
  AlertTriangle
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

// 订单详情接口
interface OrderDetail {
  id: string
  order_no: string
  total_amount: number
  payment_amount: number
  discount_amount: number
  shipping_fee: number
  status: OrderStatus
  payment_status: PaymentStatus
  payment_method?: string
  payment_time?: string
  created_at: string
  updated_at: string
  note?: string
  
  // 配送信息
  delivery_info: {
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
    product_description?: string
    quantity: number
    unit_price: number
    total_price: number
    attributes?: Record<string, string>
  }>
  
  // 物流信息
  tracking_number?: string
  carrier?: string
  estimated_delivery?: string
  delivered_at?: string
  
  // 状态历史
  status_history?: Array<{
    status: string
    note?: string
    created_at: string
  }>
}

export default function OrderDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { user } = useAuth()
  const [order, setOrder] = useState<OrderDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const orderId = params?.id as string

  // 加载订单详情
  const loadOrderDetail = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await apiClient.get(`/api/v1/shop/orders/${orderId}`)
      setOrder(((response as any)?.data?.data || (response as any)?.data)?.data)
    } catch (err: any) {
      if (err.message?.includes('404')) {
        setError('订单不存在')
      } else {
        setError(err.message || '加载订单详情失败')
      }
    } finally {
      setLoading(false)
    }
  }

  // 取消订单
  const cancelOrder = async () => {
    if (!order || !confirm('确定要取消此订单吗？')) return

    try {
      await apiClient.patch(`/api/v1/shop/orders/${orderId}/cancel`)
      toast({
        title: '订单已取消',
        description: '您的订单已成功取消'
      })
      loadOrderDetail()
    } catch (err: any) {
      toast({
        title: '取消失败',
        description: err.message || '取消订单失败',
        variant: 'destructive'
      })
    }
  }

  // 确认收货
  const confirmDelivery = async () => {
    if (!order || !confirm('确认已收到商品？')) return

    try {
      await apiClient.patch(`/api/v1/shop/orders/${orderId}/confirm-delivery`)
      toast({
        title: '确认收货成功',
        description: '订单已完成'
      })
      loadOrderDetail()
    } catch (err: any) {
      toast({
        title: '确认失败',
        description: err.message || '确认收货失败',
        variant: 'destructive'
      })
    }
  }

  // 复制订单号
  const copyOrderNumber = () => {
    if (order) {
      navigator.clipboard.writeText(order.order_no)
      toast({
        title: '已复制',
        description: '订单号已复制到剪贴板'
      })
    }
  }

  useEffect(() => {
    if (user && orderId) {
      loadOrderDetail()
    }
  }, [user, orderId])

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

  if (!user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertDescription>
            请先登录以查看订单详情
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/4"></div>
          <div className="h-32 bg-gray-200 rounded"></div>
          <div className="h-64 bg-gray-200 rounded"></div>
        </div>
      </div>
    )
  }

  if (error || !order) {
    return (
      <div className="container mx-auto px-4 py-8">
        <BackButton href="/orders" />
        <Alert variant="destructive" className="mt-4">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>{error || '订单不存在'}</AlertDescription>
        </Alert>
      </div>
    )
  }

  const statusInfo = getOrderStatusInfo(order.status)
  const paymentInfo = getPaymentStatusInfo(order.payment_status)

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <BackButton href="/orders" />
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
              <Package className="h-8 w-8" />
              订单详情
            </h1>
            <p className="text-gray-600 mt-1">订单号: {order.order_no}</p>
          </div>
        </div>
        
        <div className="flex gap-2">
          <Button variant="outline" onClick={loadOrderDetail}>
            <RefreshCw className="w-4 h-4 mr-2" />
            刷新
          </Button>
          
          <Button variant="outline" onClick={copyOrderNumber}>
            <Copy className="w-4 h-4 mr-2" />
            复制订单号
          </Button>
        </div>
      </div>

      {/* 订单状态卡片 */}
      <Card className="mb-8">
        <CardContent className="p-6">
          <div className="flex justify-between items-start">
            <div className="space-y-3">
              <div className="flex items-center gap-4">
                <Badge variant="outline" className={`${statusInfo.color} px-3 py-1`}>
                  <statusInfo.icon className="w-4 h-4 mr-2" />
                  {statusInfo.label}
                </Badge>
                <Badge variant="outline" className={paymentInfo.color}>
                  {paymentInfo.label}
                </Badge>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                <div>
                  <p className="text-gray-600">下单时间</p>
                  <p className="font-medium">{format(new Date(order.created_at), 'yyyy-MM-dd HH:mm:ss')}</p>
                </div>
                {order.payment_time && (
                  <div>
                    <p className="text-gray-600">支付时间</p>
                    <p className="font-medium">{format(new Date(order.payment_time), 'yyyy-MM-dd HH:mm:ss')}</p>
                  </div>
                )}
                {order.delivered_at && (
                  <div>
                    <p className="text-gray-600">送达时间</p>
                    <p className="font-medium">{format(new Date(order.delivered_at), 'yyyy-MM-dd HH:mm:ss')}</p>
                  </div>
                )}
              </div>
            </div>
            
            <div className="text-right space-y-2">
              <p className="text-2xl font-bold">¥{order.total_amount.toFixed(2)}</p>
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
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 左侧主要内容 */}
        <div className="lg:col-span-2 space-y-6">
          {/* 商品信息 */}
          <Card>
            <CardHeader>
              <CardTitle>商品信息</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {order.items.map((item) => (
                  <div key={item.id} className="flex items-start gap-4 p-4 border rounded-lg">
                    {item.product_image && (
                      <img 
                        src={item.product_image} 
                        alt={item.product_name}
                        className="w-20 h-20 object-cover rounded"
                      />
                    )}
                    <div className="flex-1">
                      <h4 className="font-medium mb-1">{item.product_name}</h4>
                      {item.product_description && (
                        <p className="text-sm text-gray-600 mb-2">{item.product_description}</p>
                      )}
                      {item.attributes && Object.entries(item.attributes).length > 0 && (
                        <div className="flex flex-wrap gap-2 mb-2">
                          {Object.entries(item.attributes).map(([key, value]) => (
                            <span key={key} className="text-xs bg-gray-100 px-2 py-1 rounded">
                              {key}: {value}
                            </span>
                          ))}
                        </div>
                      )}
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">
                          ¥{item.unit_price.toFixed(2)} × {item.quantity}
                        </span>
                        <span className="font-bold">¥{item.total_price.toFixed(2)}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* 物流信息 */}
          {order.tracking_number && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Truck className="h-5 w-5" />
                  物流信息
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span className="text-gray-600">快递单号</span>
                    <span className="font-mono">{order.tracking_number}</span>
                  </div>
                  {order.carrier && (
                    <div className="flex justify-between items-center">
                      <span className="text-gray-600">承运公司</span>
                      <span>{order.carrier}</span>
                    </div>
                  )}
                  {order.estimated_delivery && (
                    <div className="flex justify-between items-center">
                      <span className="text-gray-600">预计送达</span>
                      <span>{format(new Date(order.estimated_delivery), 'yyyy-MM-dd HH:mm')}</span>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          )}

          {/* 状态历史 */}
          {order.status_history && order.status_history.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Clock className="h-5 w-5" />
                  状态历史
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {order.status_history.map((history, index) => (
                    <div key={index} className="flex items-start gap-3">
                      <div className="w-2 h-2 bg-blue-600 rounded-full mt-2"></div>
                      <div className="flex-1">
                        <div className="flex items-center justify-between">
                          <span className="font-medium">{history.status}</span>
                          <span className="text-sm text-gray-500">
                            {format(new Date(history.created_at), 'MM-dd HH:mm')}
                          </span>
                        </div>
                        {history.note && (
                          <p className="text-sm text-gray-600 mt-1">{history.note}</p>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        {/* 右侧边栏 */}
        <div className="space-y-6">
          {/* 配送信息 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <MapPin className="h-5 w-5" />
                {order.delivery_info.method === 'delivery' ? '配送信息' : '自提信息'}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div>
                <p className="text-gray-600">收货人</p>
                <p className="font-medium">{order.delivery_info.name}</p>
              </div>
              <div>
                <p className="text-gray-600">联系电话</p>
                <p className="font-medium">{order.delivery_info.phone}</p>
              </div>
              {order.delivery_info.address && (
                <div>
                  <p className="text-gray-600">配送地址</p>
                  <p className="font-medium">{order.delivery_info.address}</p>
                </div>
              )}
              {order.delivery_info.pickup_location && (
                <div>
                  <p className="text-gray-600">自提地点</p>
                  <p className="font-medium">{order.delivery_info.pickup_location}</p>
                </div>
              )}
              {order.delivery_info.pickup_time && (
                <div>
                  <p className="text-gray-600">自提时间</p>
                  <p className="font-medium">{order.delivery_info.pickup_time}</p>
                </div>
              )}
            </CardContent>
          </Card>

          {/* 费用明细 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <CreditCard className="h-5 w-5" />
                费用明细
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between">
                <span className="text-gray-600">商品金额</span>
                <span>¥{(order.total_amount - order.shipping_fee + order.discount_amount).toFixed(2)}</span>
              </div>
              {order.discount_amount > 0 && (
                <div className="flex justify-between text-red-600">
                  <span>优惠金额</span>
                  <span>-¥{order.discount_amount.toFixed(2)}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-gray-600">配送费用</span>
                <span>{order.shipping_fee === 0 ? '免费' : `¥${order.shipping_fee.toFixed(2)}`}</span>
              </div>
              <div className="border-t pt-3 flex justify-between font-bold text-lg">
                <span>实付金额</span>
                <span>¥{order.payment_amount.toFixed(2)}</span>
              </div>
            </CardContent>
          </Card>

          {/* 操作按钮 */}
          <Card>
            <CardContent className="p-4 space-y-3">
              {order.status === 'pending' && (
                <>
                  <Link href={`/shop/checkout?order=${order.id}`} className="w-full">
                    <Button className="w-full">
                      <CreditCard className="w-4 h-4 mr-2" />
                      立即支付
                    </Button>
                  </Link>
                  <Button variant="outline" className="w-full" onClick={cancelOrder}>
                    取消订单
                  </Button>
                </>
              )}
              
              {order.status === 'delivered' && (
                <Button className="w-full" onClick={confirmDelivery}>
                  <CheckCircle className="w-4 h-4 mr-2" />
                  确认收货
                </Button>
              )}
              
              {order.status === 'completed' && (
                <Link href={`/orders/${order.id}/review`} className="w-full">
                  <Button variant="outline" className="w-full">
                    <MessageSquare className="w-4 h-4 mr-2" />
                    评价商品
                  </Button>
                </Link>
              )}
              
              <Link href="/orders" className="w-full">
                <Button variant="outline" className="w-full">
                  返回订单列表
                </Button>
              </Link>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}