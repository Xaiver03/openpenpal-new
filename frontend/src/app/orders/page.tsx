'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useUserStore } from '@/stores/user-store'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import Link from 'next/link'
import { apiClient } from '@/lib/api-client'
import { toast } from 'sonner'
import { 
  Package, 
  Truck, 
  CheckCircle, 
  Clock,
  AlertCircle,
  ShoppingBag,
  ArrowLeft,
  Eye,
  Loader2
} from 'lucide-react'

interface EnvelopeOrder {
  id: string
  design: {
    id: string
    theme: string
    image_url: string
    thumbnail_url: string
  }
  quantity: number
  total_price: number
  status: string
  payment_method: string
  payment_id: string
  delivery_method: string
  delivery_info: string
  created_at: string
  updated_at: string
}

export default function OrdersPage() {
  const router = useRouter()
  const { user } = useUserStore()
  const [orders, setOrders] = useState<EnvelopeOrder[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('all')

  useEffect(() => {
    if (!user) {
      router.push('/login?redirect=/orders')
      return
    }

    fetchOrders()
  }, [user, router])

  const fetchOrders = async () => {
    setLoading(true)
    try {
      const response = await apiClient.get('/api/v1/envelopes/orders')
      
      if (response.success && response.data) {
        setOrders((response.data as any) || [])
      } else {
        throw new Error(response.message || '获取订单失败')
      }
    } catch (error: any) {
      toast.error(error.message || '获取订单失败')
    } finally {
      setLoading(false)
    }
  }

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      pending: { label: '待支付', variant: 'secondary' as const, icon: Clock },
      processing: { label: '处理中', variant: 'default' as const, icon: Package },
      shipped: { label: '已发货', variant: 'outline' as const, icon: Truck },
      completed: { label: '已完成', variant: 'success' as const, icon: CheckCircle },
      cancelled: { label: '已取消', variant: 'destructive' as const, icon: AlertCircle }
    }

    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.pending
    const Icon = config.icon

    return (
      <Badge variant={config.variant} className="flex items-center gap-1">
        <Icon className="h-3 w-3" />
        {config.label}
      </Badge>
    )
  }

  const filteredOrders = orders.filter(order => {
    if (activeTab === 'all') return true
    return order.status === activeTab
  })

  if (!user) {
    return null
  }

  if (loading) {
    return (
      <div className="min-h-screen flex flex-col bg-letter-paper">
        <Header />
        <main className="flex-1 flex items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-amber-600" />
        </main>
        <Footer />
      </div>
    )
  }

  return (
    <div className="min-h-screen flex flex-col bg-letter-paper">
      <Header />
      
      <main className="flex-1 py-8">
        <div className="container px-4">
          {/* Page Header */}
          <div className="mb-8">
            <div className="flex items-center justify-between">
              <h1 className="font-serif text-3xl font-bold text-gray-900">
                我的订单
              </h1>
              <Button asChild variant="outline">
                <Link href="/shop">
                  <ShoppingBag className="mr-2 h-4 w-4" />
                  继续购物
                </Link>
              </Button>
            </div>
          </div>

          {/* Order Tabs */}
          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList className="grid w-full max-w-md grid-cols-5">
              <TabsTrigger value="all">全部</TabsTrigger>
              <TabsTrigger value="pending">待支付</TabsTrigger>
              <TabsTrigger value="processing">处理中</TabsTrigger>
              <TabsTrigger value="shipped">已发货</TabsTrigger>
              <TabsTrigger value="completed">已完成</TabsTrigger>
            </TabsList>

            <TabsContent value={activeTab} className="mt-6">
              {filteredOrders.length === 0 ? (
                <Card>
                  <CardContent className="py-12 text-center">
                    <Package className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">
                      暂无订单
                    </h3>
                    <p className="text-gray-600 mb-6">
                      {activeTab === 'all' ? '您还没有任何订单' : `没有${getStatusBadge(activeTab).props.children[1]}的订单`}
                    </p>
                    <Button asChild>
                      <Link href="/shop">
                        去购物
                      </Link>
                    </Button>
                  </CardContent>
                </Card>
              ) : (
                <div className="space-y-4">
                  {filteredOrders.map((order) => {
                    const deliveryInfo = JSON.parse(order.delivery_info || '{}')
                    
                    return (
                      <Card key={order.id}>
                        <CardHeader>
                          <div className="flex items-center justify-between">
                            <div>
                              <p className="text-sm text-gray-500 mb-1">
                                订单号：{order.id}
                              </p>
                              <p className="text-sm text-gray-500">
                                下单时间：{new Date(order.created_at).toLocaleString()}
                              </p>
                            </div>
                            {getStatusBadge(order.status)}
                          </div>
                        </CardHeader>
                        <CardContent>
                          <div className="flex gap-4">
                            {/* Product Image */}
                            <div className="w-24 h-24 bg-gradient-to-br from-amber-100 to-orange-100 rounded-lg flex items-center justify-center">
                              <Package className="w-10 h-10 text-amber-600" />
                            </div>
                            
                            {/* Order Details */}
                            <div className="flex-1">
                              <h3 className="font-semibold text-gray-900 mb-1">
                                {order.design.theme}
                              </h3>
                              <p className="text-sm text-gray-600 mb-2">
                                数量：{order.quantity} 件
                              </p>
                              
                              <div className="flex items-center gap-4 text-sm text-gray-600">
                                <span>支付方式：{
                                  order.payment_method === 'alipay' ? '支付宝' :
                                  order.payment_method === 'wechat' ? '微信支付' : '银行卡'
                                }</span>
                                <span>配送方式：{
                                  deliveryInfo.method === 'delivery' ? '送货上门' : '到店自提'
                                }</span>
                              </div>
                              
                              {deliveryInfo.method === 'delivery' ? (
                                <p className="text-sm text-gray-600 mt-1">
                                  收货地址：{deliveryInfo.address}
                                </p>
                              ) : (
                                <p className="text-sm text-gray-600 mt-1">
                                  自提地点：{deliveryInfo.pickupLocation} ({deliveryInfo.pickupTime})
                                </p>
                              )}
                            </div>
                            
                            {/* Price and Actions */}
                            <div className="text-right">
                              <p className="text-xl font-bold text-red-600 mb-4">
                                ¥{order.total_price.toFixed(2)}
                              </p>
                              
                              <div className="space-y-2">
                                <Button variant="outline" size="sm" className="w-full">
                                  <Eye className="mr-2 h-4 w-4" />
                                  查看详情
                                </Button>
                                
                                {order.status === 'pending' && (
                                  <Button size="sm" className="w-full bg-amber-600 hover:bg-amber-700 text-white">
                                    立即支付
                                  </Button>
                                )}
                                
                                {order.status === 'completed' && (
                                  <Button variant="outline" size="sm" className="w-full">
                                    再次购买
                                  </Button>
                                )}
                              </div>
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    )
                  })}
                </div>
              )}
            </TabsContent>
          </Tabs>
        </div>
      </main>

      <Footer />
    </div>
  )
}