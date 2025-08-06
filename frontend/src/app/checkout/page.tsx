'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useCartStore } from '@/stores/cart-store'
import { useUserStore } from '@/stores/user-store'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Textarea } from '@/components/ui/textarea'
import { toast } from 'sonner'
import { apiClient } from '@/lib/api-client'
import Link from 'next/link'
import { 
  CreditCard, 
  Truck, 
  MapPin, 
  Phone, 
  Mail,
  Package,
  ArrowLeft,
  Loader2,
  Building
} from 'lucide-react'

interface DeliveryInfo {
  method: 'delivery' | 'pickup'
  name: string
  phone: string
  address?: string
  pickupLocation?: string
  pickupTime?: string
}

export default function CheckoutPage() {
  const router = useRouter()
  const { user } = useUserStore()
  const { items, calculateTotal, clearCart } = useCartStore()
  const [loading, setLoading] = useState(false)
  
  const [paymentMethod, setPaymentMethod] = useState<'alipay' | 'wechat' | 'card'>('alipay')
  const [deliveryInfo, setDeliveryInfo] = useState<DeliveryInfo>({
    method: 'delivery',
    name: user?.nickname || '',
    phone: '',
    address: ''
  })

  const total = calculateTotal()
  const shippingFee = 0 // 免费配送
  const grandTotal = total + shippingFee

  if (!user) {
    router.push('/login?redirect=/checkout')
    return null
  }

  if (items.length === 0) {
    router.push('/cart')
    return null
  }

  const handleSubmitOrder = async () => {
    // Validate form
    if (!deliveryInfo.name || !deliveryInfo.phone) {
      toast.error('请填写收货人姓名和电话')
      return
    }

    if (deliveryInfo.method === 'delivery' && !deliveryInfo.address) {
      toast.error('请填写收货地址')
      return
    }

    if (deliveryInfo.method === 'pickup' && (!deliveryInfo.pickupLocation || !deliveryInfo.pickupTime)) {
      toast.error('请选择自提地点和时间')
      return
    }

    setLoading(true)

    try {
      // Create order for each design (since backend expects one design per order)
      // In a real implementation, you might want to modify backend to support multiple items
      for (const item of items) {
        const orderData = {
          design_id: item.id.toString(), // Assuming item.id corresponds to design_id
          quantity: item.quantity,
          payment_method: paymentMethod,
          delivery_method: deliveryInfo.method,
          delivery_info: JSON.stringify(deliveryInfo)
        }

        const response = await apiClient.post('/api/v1/envelopes/orders', orderData)
        
        if (!response.success) {
          throw new Error(response.message || '创建订单失败')
        }

        // Process payment (mock for now)
        const paymentResponse = await apiClient.post(`/api/v1/envelopes/orders/${(response.data as any)?.id}/pay`, {
          payment_id: `PAY_${Date.now()}`,
          payment_method: paymentMethod
        })

        if (!paymentResponse.success) {
          throw new Error('支付处理失败')
        }
      }

      // Clear cart after successful order
      clearCart()
      toast.success('订单创建成功！')
      
      // Redirect to order success page or order list
      router.push('/orders')
    } catch (error: any) {
      toast.error(error.message || '订单创建失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex flex-col bg-letter-paper">
      <Header />
      
      <main className="flex-1 py-8">
        <div className="container px-4">
          {/* Page Header */}
          <div className="mb-8">
            <h1 className="font-serif text-3xl font-bold text-gray-900">
              订单结算
            </h1>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Checkout Form */}
            <div className="lg:col-span-2 space-y-6">
              {/* Delivery Information */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Truck className="h-5 w-5" />
                    配送信息
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <RadioGroup 
                    value={deliveryInfo.method} 
                    onValueChange={(value) => setDeliveryInfo({...deliveryInfo, method: value as 'delivery' | 'pickup'})}
                  >
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="delivery" id="delivery" />
                      <Label htmlFor="delivery">送货上门</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="pickup" id="pickup" />
                      <Label htmlFor="pickup">到店自提</Label>
                    </div>
                  </RadioGroup>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="name">收货人姓名</Label>
                      <Input
                        id="name"
                        value={deliveryInfo.name}
                        onChange={(e) => setDeliveryInfo({...deliveryInfo, name: e.target.value})}
                        placeholder="请输入姓名"
                      />
                    </div>
                    <div>
                      <Label htmlFor="phone">联系电话</Label>
                      <Input
                        id="phone"
                        value={deliveryInfo.phone}
                        onChange={(e) => setDeliveryInfo({...deliveryInfo, phone: e.target.value})}
                        placeholder="请输入手机号"
                      />
                    </div>
                  </div>

                  {deliveryInfo.method === 'delivery' ? (
                    <div>
                      <Label htmlFor="address">收货地址</Label>
                      <Textarea
                        id="address"
                        value={deliveryInfo.address}
                        onChange={(e) => setDeliveryInfo({...deliveryInfo, address: e.target.value})}
                        placeholder="请输入详细地址"
                        rows={3}
                      />
                    </div>
                  ) : (
                    <div className="space-y-4">
                      <div>
                        <Label>自提地点</Label>
                        <RadioGroup
                          value={deliveryInfo.pickupLocation}
                          onValueChange={(value) => setDeliveryInfo({...deliveryInfo, pickupLocation: value})}
                        >
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="学生活动中心" id="location1" />
                            <Label htmlFor="location1">学生活动中心</Label>
                          </div>
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="图书馆一楼" id="location2" />
                            <Label htmlFor="location2">图书馆一楼</Label>
                          </div>
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="教学楼A座" id="location3" />
                            <Label htmlFor="location3">教学楼A座</Label>
                          </div>
                        </RadioGroup>
                      </div>
                      <div>
                        <Label>自提时间</Label>
                        <RadioGroup
                          value={deliveryInfo.pickupTime}
                          onValueChange={(value) => setDeliveryInfo({...deliveryInfo, pickupTime: value})}
                        >
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="09:00-12:00" id="time1" />
                            <Label htmlFor="time1">上午 9:00-12:00</Label>
                          </div>
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="14:00-17:00" id="time2" />
                            <Label htmlFor="time2">下午 14:00-17:00</Label>
                          </div>
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="18:00-21:00" id="time3" />
                            <Label htmlFor="time3">晚上 18:00-21:00</Label>
                          </div>
                        </RadioGroup>
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>

              {/* Payment Method */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <CreditCard className="h-5 w-5" />
                    支付方式
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <RadioGroup 
                    value={paymentMethod} 
                    onValueChange={(value) => setPaymentMethod(value as any)}
                  >
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="alipay" id="alipay" />
                      <Label htmlFor="alipay" className="flex items-center gap-2 cursor-pointer">
                        支付宝
                      </Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="wechat" id="wechat" />
                      <Label htmlFor="wechat" className="flex items-center gap-2 cursor-pointer">
                        微信支付
                      </Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="card" id="card" />
                      <Label htmlFor="card" className="flex items-center gap-2 cursor-pointer">
                        银行卡
                      </Label>
                    </div>
                  </RadioGroup>
                </CardContent>
              </Card>

              {/* Back to Cart */}
              <div>
                <Button asChild variant="outline">
                  <Link href="/cart">
                    <ArrowLeft className="mr-2 h-4 w-4" />
                    返回购物车
                  </Link>
                </Button>
              </div>
            </div>

            {/* Order Summary */}
            <div className="lg:col-span-1">
              <Card className="sticky top-4">
                <CardHeader>
                  <CardTitle className="font-serif">订单详情</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Order Items */}
                  <div className="space-y-3 max-h-64 overflow-y-auto">
                    {items.map((item) => (
                      <div key={item.id} className="flex justify-between text-sm">
                        <div className="flex-1">
                          <p className="font-medium">{item.name}</p>
                          <p className="text-gray-500">¥{item.price} × {item.quantity}</p>
                        </div>
                        <div className="text-right">
                          <p className="font-medium">¥{(item.price * item.quantity).toFixed(2)}</p>
                        </div>
                      </div>
                    ))}
                  </div>

                  {/* Totals */}
                  <div className="border-t pt-4 space-y-2">
                    <div className="flex justify-between text-sm">
                      <span>商品小计</span>
                      <span>¥{total.toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span>运费</span>
                      <span className="text-green-600">免费</span>
                    </div>
                    <div className="border-t pt-2 flex justify-between font-semibold text-lg">
                      <span>应付总额</span>
                      <span className="text-red-600">¥{grandTotal.toFixed(2)}</span>
                    </div>
                  </div>

                  {/* Submit Button */}
                  <Button 
                    className="w-full bg-amber-600 hover:bg-amber-700 text-white"
                    size="lg"
                    onClick={handleSubmitOrder}
                    disabled={loading}
                  >
                    {loading ? (
                      <>
                        <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                        处理中...
                      </>
                    ) : (
                      <>
                        <CreditCard className="mr-2 h-5 w-5" />
                        立即支付
                      </>
                    )}
                  </Button>

                  {/* Security Notice */}
                  <p className="text-xs text-gray-500 text-center">
                    您的支付信息将被安全加密处理
                  </p>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </main>

      <Footer />
    </div>
  )
}