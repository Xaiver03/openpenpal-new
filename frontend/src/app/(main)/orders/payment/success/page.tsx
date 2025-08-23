'use client'

import { useState, useEffect } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  CheckCircle,
  Package,
  CreditCard,
  Calendar,
  MapPin,
  ArrowRight,
  Download,
  Share2,
  Home,
  ShoppingBag,
  Clock,
  X
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { apiClient } from '@/lib/api-client'
import { format } from 'date-fns'
import Link from 'next/link'
import { toast } from '@/components/ui/use-toast'

interface PaymentResult {
  order_id: string
  order_no: string
  payment_amount: number
  payment_method: string
  payment_time: string
  status: 'success' | 'failed' | 'pending'
  transaction_id?: string
  
  order_summary: {
    items: Array<{
      product_name: string
      quantity: number
      unit_price: number
    }>
    delivery_info: {
      name: string
      phone: string
      address?: string
      method: 'delivery' | 'pickup'
    }
  }
}

export default function PaymentSuccessPage() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const { user } = useAuth()
  const [paymentResult, setPaymentResult] = useState<PaymentResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const orderId = searchParams?.get('order_id')
  const transactionId = searchParams?.get('transaction_id')

  // 加载支付结果
  const loadPaymentResult = async () => {
    if (!orderId) {
      setError('缺少订单信息')
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      
      // 验证支付结果
      const response = await apiClient.get(`/api/v1/shop/orders/${orderId}/payment-result?transaction_id=${transactionId}`)
      
      setPaymentResult(((response as any)?.data?.data || (response as any)?.data)?.data)
    } catch (err: any) {
      setError(err.message || '获取支付结果失败')
    } finally {
      setLoading(false)
    }
  }

  // 下载订单收据
  const downloadReceipt = async () => {
    if (!paymentResult) return

    try {
      // 使用fetch直接下载blob文件
      const response = await fetch(`/api/v1/shop/orders/${paymentResult.order_id}/receipt`, {
        headers: { 
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}` 
        }
      })
      
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `订单收据_${paymentResult.order_no}.pdf`
      link.click()
      window.URL.revokeObjectURL(url)
    } catch (err: any) {
      toast({
        title: '下载失败',
        description: err.message || '下载收据失败',
        variant: 'destructive'
      })
    }
  }

  // 分享订单
  const shareOrder = async () => {
    if (!paymentResult) return

    try {
      if (navigator.share) {
        await navigator.share({
          title: '我的订单',
          text: `我在OpenPenPal购买了商品，订单号：${paymentResult.order_no}`,
          url: window.location.origin + `/orders/${paymentResult.order_id}`
        })
      } else {
        // 复制链接到剪贴板
        const orderUrl = `${window.location.origin}/orders/${paymentResult.order_id}`
        await navigator.clipboard.writeText(orderUrl)
        toast({
          title: '已复制链接',
          description: '订单链接已复制到剪贴板'
        })
      }
    } catch (err) {
      console.error('分享失败:', err)
    }
  }

  useEffect(() => {
    if (user && orderId) {
      loadPaymentResult()
    }
  }, [user, orderId])

  // 自动跳转倒计时
  useEffect(() => {
    if (paymentResult?.status === 'success') {
      const timer = setTimeout(() => {
        router.push(`/orders/${paymentResult.order_id}`)
      }, 10000) // 10秒后自动跳转

      return () => clearTimeout(timer)
    }
  }, [paymentResult, router])

  if (!user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertDescription>
            请先登录以查看支付结果
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-2xl mx-auto">
          <div className="animate-pulse space-y-4">
            <div className="h-32 bg-gray-200 rounded-lg"></div>
            <div className="h-64 bg-gray-200 rounded-lg"></div>
          </div>
        </div>
      </div>
    )
  }

  if (error || !paymentResult) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-2xl mx-auto">
          <Alert variant="destructive">
            <AlertDescription>{error || '获取支付结果失败'}</AlertDescription>
          </Alert>
          <div className="text-center mt-6">
            <Link href="/orders">
              <Button>返回订单列表</Button>
            </Link>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="container mx-auto px-4">
        <div className="max-w-2xl mx-auto">
          {/* 支付状态卡片 */}
          <Card className="mb-8">
            <CardContent className="text-center py-12">
              {paymentResult.status === 'success' ? (
                <>
                  <div className="w-20 h-20 mx-auto mb-6 bg-green-100 rounded-full flex items-center justify-center">
                    <CheckCircle className="w-10 h-10 text-green-600" />
                  </div>
                  <h1 className="text-3xl font-bold text-green-900 mb-2">支付成功！</h1>
                  <p className="text-gray-600 mb-6">
                    您的订单已支付完成，我们将尽快为您处理
                  </p>
                </>
              ) : paymentResult.status === 'failed' ? (
                <>
                  <div className="w-20 h-20 mx-auto mb-6 bg-red-100 rounded-full flex items-center justify-center">
                    <X className="w-10 h-10 text-red-600" />
                  </div>
                  <h1 className="text-3xl font-bold text-red-900 mb-2">支付失败</h1>
                  <p className="text-gray-600 mb-6">
                    支付过程中出现问题，请重试或联系客服
                  </p>
                </>
              ) : (
                <>
                  <div className="w-20 h-20 mx-auto mb-6 bg-yellow-100 rounded-full flex items-center justify-center">
                    <Clock className="w-10 h-10 text-yellow-600" />
                  </div>
                  <h1 className="text-3xl font-bold text-yellow-900 mb-2">支付处理中</h1>
                  <p className="text-gray-600 mb-6">
                    正在确认您的支付，请稍候...
                  </p>
                </>
              )}
              
              <div className="flex items-center justify-center gap-4 text-sm text-gray-600">
                <span>订单号: {paymentResult.order_no}</span>
                <span>•</span>
                <span>支付金额: ¥{paymentResult.payment_amount.toFixed(2)}</span>
              </div>
            </CardContent>
          </Card>

          {/* 订单详情 */}
          <Card className="mb-8">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Package className="h-5 w-5" />
                订单详情
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* 支付信息 */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 p-4 bg-gray-50 rounded-lg">
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <CreditCard className="h-4 w-4 text-gray-500" />
                    <span className="text-sm font-medium">支付方式</span>
                  </div>
                  <p className="text-sm">
                    {{
                      'alipay': '支付宝',
                      'wechat': '微信支付',
                      'card': '银行卡',
                      'credit': '积分支付'
                    }[paymentResult.payment_method] || paymentResult.payment_method}
                  </p>
                  {paymentResult.transaction_id && (
                    <p className="text-xs text-gray-500 font-mono">
                      交易号: {paymentResult.transaction_id}
                    </p>
                  )}
                </div>
                
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-gray-500" />
                    <span className="text-sm font-medium">支付时间</span>
                  </div>
                  <p className="text-sm">
                    {format(new Date(paymentResult.payment_time), 'yyyy年MM月dd日 HH:mm:ss')}
                  </p>
                </div>
              </div>

              {/* 商品列表 */}
              <div>
                <h3 className="font-medium mb-3">购买商品</h3>
                <div className="space-y-2">
                  {paymentResult.order_summary.items.map((item, index) => (
                    <div key={index} className="flex justify-between items-center p-3 border rounded">
                      <div className="flex-1">
                        <p className="font-medium">{item.product_name}</p>
                        <p className="text-sm text-gray-600">
                          ¥{item.unit_price.toFixed(2)} × {item.quantity}
                        </p>
                      </div>
                      <div className="font-medium">
                        ¥{(item.unit_price * item.quantity).toFixed(2)}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* 配送信息 */}
              <div className="p-4 bg-blue-50 rounded-lg">
                <div className="flex items-start gap-2 mb-2">
                  <MapPin className="h-4 w-4 text-blue-600 mt-1" />
                  <span className="text-sm font-medium text-blue-900">
                    {paymentResult.order_summary.delivery_info.method === 'delivery' ? '配送信息' : '自提信息'}
                  </span>
                </div>
                <div className="text-sm text-blue-800">
                  <p>收货人: {paymentResult.order_summary.delivery_info.name}</p>
                  <p>联系电话: {paymentResult.order_summary.delivery_info.phone}</p>
                  {paymentResult.order_summary.delivery_info.address && (
                    <p>配送地址: {paymentResult.order_summary.delivery_info.address}</p>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* 操作按钮 */}
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Link href={`/orders/${paymentResult.order_id}`} className="w-full">
                <Button className="w-full">
                  <Package className="w-4 h-4 mr-2" />
                  查看订单详情
                </Button>
              </Link>
              
              <Button variant="outline" className="w-full" onClick={downloadReceipt}>
                <Download className="w-4 h-4 mr-2" />
                下载订单收据
              </Button>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Button variant="outline" onClick={shareOrder}>
                <Share2 className="w-4 h-4 mr-2" />
                分享订单
              </Button>
              
              <Link href="/shop" className="w-full">
                <Button variant="outline" className="w-full">
                  <ShoppingBag className="w-4 h-4 mr-2" />
                  继续购物
                </Button>
              </Link>
              
              <Link href="/" className="w-full">
                <Button variant="outline" className="w-full">
                  <Home className="w-4 h-4 mr-2" />
                  返回首页
                </Button>
              </Link>
            </div>
          </div>

          {/* 温馨提示 */}
          {paymentResult.status === 'success' && (
            <Alert className="mt-6">
              <AlertDescription>
                <div className="space-y-2">
                  <p>• 订单支付成功后，我们将在1-2个工作日内安排发货</p>
                  <p>• 您可以在"我的订单"中随时查看订单状态和物流信息</p>
                  <p>• 如有任何问题，请联系客服：service@openpenpal.com</p>
                  <p className="text-xs text-gray-500">页面将在10秒后自动跳转到订单详情</p>
                </div>
              </AlertDescription>
            </Alert>
          )}
        </div>
      </div>
    </div>
  )
}