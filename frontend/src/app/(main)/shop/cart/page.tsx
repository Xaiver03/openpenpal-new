'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  ShoppingCart, 
  Trash2,
  Plus,
  Minus,
  CreditCard,
  Coins,
  ArrowRight,
  Package,
  AlertCircle
} from 'lucide-react'
import { BackButton } from '@/components/ui/back-button'
import { useAuth } from '@/contexts/auth-context-new'
import { apiClient } from '@/lib/api-client'
import { formatPrice } from '@/lib/utils'

interface CartItem {
  id: string
  product_id: string
  product: {
    id: string
    name: string
    description: string
    price: number
    credit_price?: number
    stock: number
    image_url?: string
  }
  quantity: number
  payment_type: 'cash' | 'credit'
}

export default function ShoppingCartPage() {
  const router = useRouter()
  const { user } = useAuth()
  const [cart, setCart] = useState<CartItem[]>([])
  const [loading, setLoading] = useState(true)
  const [updating, setUpdating] = useState<string | null>(null)
  const [userCredits, setUserCredits] = useState(0)

  useEffect(() => {
    if (!user) {
      router.push('/login')
      return
    }
    loadCart()
    loadUserCredits()
  }, [user])

  const loadCart = async () => {
    setLoading(true)
    try {
      const response = await apiClient.get<any>('/shop/cart')
      const data = response.data
      const items = data && typeof data === 'object' ? (data.items ?? data.data?.items ?? []) : []
      setCart(items)
    } catch (error) {
      console.error('Failed to load cart:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadUserCredits = async () => {
    try {
      const response = await apiClient.get<any>('/credits/balance')
      const data = response.data
      const balance = data && typeof data === 'object' ? (data.balance ?? data.data?.balance ?? 0) : 0
      setUserCredits(balance)
    } catch (error) {
      console.error('Failed to load credits:', error)
    }
  }

  const updateQuantity = async (itemId: string, newQuantity: number) => {
    if (newQuantity < 1) return
    
    setUpdating(itemId)
    try {
      await apiClient.put(`/shop/cart/items/${itemId}`, {
        quantity: newQuantity
      })
      
      // 更新本地状态
      setCart(cart.map(item => 
        item.id === itemId ? { ...item, quantity: newQuantity } : item
      ))
    } catch (error) {
      console.error('Failed to update quantity:', error)
      alert('更新数量失败')
    } finally {
      setUpdating(null)
    }
  }

  const removeItem = async (itemId: string) => {
    if (!confirm('确定要删除这个商品吗？')) return
    
    setUpdating(itemId)
    try {
      await apiClient.delete(`/shop/cart/items/${itemId}`)
      
      // 更新本地状态
      setCart(cart.filter(item => item.id !== itemId))
    } catch (error) {
      console.error('Failed to remove item:', error)
      alert('删除商品失败')
    } finally {
      setUpdating(null)
    }
  }

  const clearCart = async () => {
    if (!confirm('确定要清空购物车吗？')) return
    
    setLoading(true)
    try {
      await apiClient.delete('/shop/cart')
      setCart([])
    } catch (error) {
      console.error('Failed to clear cart:', error)
      alert('清空购物车失败')
    } finally {
      setLoading(false)
    }
  }

  // 计算总价
  const cashTotal = cart
    .filter(item => item.payment_type === 'cash')
    .reduce((total, item) => total + item.product.price * item.quantity, 0)

  const creditTotal = cart
    .filter(item => item.payment_type === 'credit')
    .reduce((total, item) => total + (item.product.credit_price || 0) * item.quantity, 0)

  // 检查是否可以结算
  const canCheckout = cart.length > 0 && (creditTotal <= userCredits || creditTotal === 0)

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
        </div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <BackButton />
        <div className="flex items-center justify-between mt-4">
          <div>
            <h1 className="text-3xl font-bold">购物车</h1>
            <p className="text-gray-600 mt-2">
              共 {cart.length} 件商品
            </p>
          </div>
          
          {cart.length > 0 && (
            <Button variant="outline" onClick={clearCart}>
              清空购物车
            </Button>
          )}
        </div>
      </div>

      {cart.length === 0 ? (
        <Card>
          <CardContent className="text-center py-12">
            <ShoppingCart className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600 mb-6">购物车为空</p>
            <Button onClick={() => router.push('/shop')}>
              去购物
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* 商品列表 */}
          <div className="lg:col-span-2 space-y-4">
            {cart.map(item => (
              <Card key={item.id}>
                <CardContent className="p-6">
                  <div className="flex gap-4">
                    {/* 商品图片 */}
                    <div className="w-24 h-24 bg-gray-100 rounded-md flex items-center justify-center">
                      {item.product.image_url ? (
                        <img 
                          src={item.product.image_url} 
                          alt={item.product.name}
                          className="w-full h-full object-cover rounded-md"
                        />
                      ) : (
                        <Package className="h-8 w-8 text-gray-400" />
                      )}
                    </div>

                    {/* 商品信息 */}
                    <div className="flex-1">
                      <h3 className="font-semibold text-lg">{item.product.name}</h3>
                      <p className="text-sm text-gray-600 mt-1">
                        {item.product.description}
                      </p>
                      
                      {/* 价格 */}
                      <div className="mt-3 flex items-center gap-2">
                        {item.payment_type === 'credit' ? (
                          <>
                            <Coins className="h-4 w-4 text-yellow-600" />
                            <span className="font-semibold text-yellow-600">
                              {item.product.credit_price} 积分
                            </span>
                          </>
                        ) : (
                          <>
                            <CreditCard className="h-4 w-4 text-gray-600" />
                            <span className="font-semibold">
                              ¥{formatPrice(item.product.price)}
                            </span>
                          </>
                        )}
                      </div>
                    </div>

                    {/* 数量和操作 */}
                    <div className="flex flex-col items-end justify-between">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="text-red-600"
                        onClick={() => removeItem(item.id)}
                        disabled={updating === item.id}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                      
                      <div className="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="icon"
                          onClick={() => updateQuantity(item.id, item.quantity - 1)}
                          disabled={item.quantity <= 1 || updating === item.id}
                        >
                          <Minus className="h-4 w-4" />
                        </Button>
                        
                        <Input
                          type="number"
                          value={item.quantity}
                          onChange={(e) => {
                            const value = parseInt(e.target.value)
                            if (!isNaN(value) && value > 0) {
                              updateQuantity(item.id, value)
                            }
                          }}
                          className="w-16 text-center"
                          disabled={updating === item.id}
                        />
                        
                        <Button
                          variant="outline"
                          size="icon"
                          onClick={() => updateQuantity(item.id, item.quantity + 1)}
                          disabled={item.quantity >= item.product.stock || updating === item.id}
                        >
                          <Plus className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* 结算信息 */}
          <div className="lg:col-span-1">
            <Card className="sticky top-4">
              <CardHeader>
                <CardTitle>订单摘要</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {/* 积分余额 */}
                {creditTotal > 0 && (
                  <div className="flex items-center justify-between p-3 bg-yellow-50 rounded-md">
                    <div className="flex items-center gap-2">
                      <Coins className="h-4 w-4 text-yellow-600" />
                      <span className="text-sm">可用积分</span>
                    </div>
                    <span className="font-semibold">{userCredits}</span>
                  </div>
                )}

                {/* 价格明细 */}
                <div className="space-y-2">
                  {cashTotal > 0 && (
                    <div className="flex justify-between">
                      <span>商品总价</span>
                      <span>¥{formatPrice(cashTotal)}</span>
                    </div>
                  )}
                  
                  {creditTotal > 0 && (
                    <div className="flex justify-between text-yellow-600">
                      <span>积分总计</span>
                      <span>{creditTotal} 积分</span>
                    </div>
                  )}
                </div>

                <div className="border-t pt-4">
                  <div className="flex justify-between font-semibold text-lg">
                    <span>应付总额</span>
                    <div className="text-right">
                      {cashTotal > 0 && <div>¥{formatPrice(cashTotal)}</div>}
                      {creditTotal > 0 && <div className="text-yellow-600">{creditTotal} 积分</div>}
                    </div>
                  </div>
                </div>

                {/* 积分不足提示 */}
                {creditTotal > userCredits && (
                  <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>
                      积分不足，还需要 {creditTotal - userCredits} 积分
                    </AlertDescription>
                  </Alert>
                )}

                {/* 结算按钮 */}
                <Button 
                  className="w-full" 
                  size="lg"
                  onClick={() => router.push('/shop/checkout')}
                  disabled={!canCheckout}
                >
                  去结算
                  <ArrowRight className="h-4 w-4 ml-2" />
                </Button>

                <Button
                  variant="outline"
                  className="w-full"
                  onClick={() => router.push('/shop')}
                >
                  继续购物
                </Button>
              </CardContent>
            </Card>
          </div>
        </div>
      )}
    </div>
  )
}