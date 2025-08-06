'use client'

import { useCartStore } from '@/stores/cart-store'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { 
  ShoppingCart, 
  Plus, 
  Minus, 
  Trash2, 
  ArrowLeft, 
  CreditCard,
  Package,
  Shield,
  Truck
} from 'lucide-react'
import { toast } from 'sonner'

export default function CartPage() {
  const router = useRouter()
  const { 
    items, 
    removeItem, 
    updateQuantity, 
    clearCart,
    calculateTotal,
    calculateItemCount 
  } = useCartStore()

  const total = calculateTotal()
  const itemCount = calculateItemCount()
  const shippingFee = 0 // 免费配送
  const grandTotal = total + shippingFee

  const handleQuantityChange = (id: number, change: number) => {
    const item = items.find(item => item.id === id)
    if (item) {
      const newQuantity = item.quantity + change
      if (newQuantity > 0) {
        updateQuantity(id, newQuantity)
      }
    }
  }

  const handleRemoveItem = (id: number, name: string) => {
    removeItem(id)
    toast.success(`${name} 已从购物车移除`)
  }

  const handleClearCart = () => {
    if (window.confirm('确定要清空购物车吗？')) {
      clearCart()
      toast.success('购物车已清空')
    }
  }

  const handleCheckout = () => {
    if (items.length === 0) {
      toast.error('购物车为空，请先添加商品')
      return
    }
    router.push('/checkout')
  }

  if (items.length === 0) {
    return (
      <div className="min-h-screen flex flex-col bg-letter-paper">
        <Header />
        
        <main className="flex-1 py-16">
          <div className="container px-4">
            <div className="text-center max-w-md mx-auto">
              <ShoppingCart className="w-16 h-16 text-gray-400 mx-auto mb-4" />
              <h1 className="font-serif text-2xl font-bold text-gray-900 mb-4">
                购物车是空的
              </h1>
              <p className="text-gray-600 mb-8">
                还没有添加任何商品到购物车
              </p>
              <Button asChild>
                <Link href="/shop">
                  <ArrowLeft className="mr-2 h-4 w-4" />
                  去购物
                </Link>
              </Button>
            </div>
          </div>
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
                购物车 ({itemCount} 件商品)
              </h1>
              <Button 
                variant="outline" 
                size="sm"
                onClick={handleClearCart}
                className="text-red-600 hover:text-red-700 border-red-300 hover:border-red-400"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                清空购物车
              </Button>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Cart Items */}
            <div className="lg:col-span-2 space-y-4">
              {items.map((item) => (
                <Card key={item.id} className="overflow-hidden">
                  <CardContent className="p-0">
                    <div className="flex">
                      <div className="w-32 h-32 bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center">
                        <Package className="w-12 h-12 text-amber-600" />
                      </div>
                      
                      <div className="flex-1 p-6">
                        <div className="flex justify-between">
                          <div className="flex-1">
                            <h3 className="font-serif text-lg font-semibold text-gray-900 mb-1">
                              {item.name}
                            </h3>
                            <p className="text-sm text-gray-600 mb-3">
                              {item.description}
                            </p>
                            
                            <div className="flex items-center gap-2 mb-3">
                              {item.tags.map(tag => (
                                <span 
                                  key={tag} 
                                  className="text-xs bg-amber-100 text-amber-800 px-2 py-1 rounded"
                                >
                                  {tag}
                                </span>
                              ))}
                            </div>
                            
                            <div className="flex items-center gap-4">
                              {/* Quantity Controls */}
                              <div className="flex items-center border rounded-lg">
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleQuantityChange(item.id, -1)}
                                  disabled={item.quantity <= 1}
                                  className="h-8 w-8 p-0"
                                >
                                  <Minus className="h-4 w-4" />
                                </Button>
                                <Input
                                  type="number"
                                  value={item.quantity}
                                  onChange={(e) => {
                                    const val = parseInt(e.target.value)
                                    if (!isNaN(val) && val > 0) {
                                      updateQuantity(item.id, val)
                                    }
                                  }}
                                  className="h-8 w-16 text-center border-0 focus:ring-0"
                                />
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleQuantityChange(item.id, 1)}
                                  className="h-8 w-8 p-0"
                                >
                                  <Plus className="h-4 w-4" />
                                </Button>
                              </div>
                              
                              {/* Remove Button */}
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleRemoveItem(item.id, item.name)}
                                className="text-red-600 hover:text-red-700"
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            </div>
                          </div>
                          
                          <div className="text-right ml-4">
                            <div className="text-2xl font-bold text-red-600">
                              ¥{(item.price * item.quantity).toFixed(2)}
                            </div>
                            {item.originalPrice > item.price && (
                              <div className="text-sm text-gray-500 line-through">
                                ¥{(item.originalPrice * item.quantity).toFixed(2)}
                              </div>
                            )}
                            <div className="text-sm text-gray-600 mt-1">
                              ¥{item.price} × {item.quantity}
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}

              {/* Continue Shopping */}
              <div className="pt-4">
                <Button asChild variant="outline">
                  <Link href="/shop">
                    <ArrowLeft className="mr-2 h-4 w-4" />
                    继续购物
                  </Link>
                </Button>
              </div>
            </div>

            {/* Order Summary */}
            <div className="lg:col-span-1">
              <Card className="sticky top-4">
                <CardHeader>
                  <CardTitle className="font-serif">订单摘要</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <div className="flex justify-between text-sm">
                      <span>商品小计</span>
                      <span>¥{total.toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between text-sm">
                      <span>运费</span>
                      <span className="text-green-600">免费</span>
                    </div>
                    <div className="border-t pt-2 flex justify-between font-semibold text-lg">
                      <span>总计</span>
                      <span className="text-red-600">¥{grandTotal.toFixed(2)}</span>
                    </div>
                  </div>

                  <Button 
                    className="w-full bg-amber-600 hover:bg-amber-700 text-white"
                    size="lg"
                    onClick={handleCheckout}
                  >
                    <CreditCard className="mr-2 h-5 w-5" />
                    去结算
                  </Button>

                  {/* Service Features */}
                  <div className="space-y-3 pt-4 border-t">
                    <div className="flex items-center gap-3 text-sm">
                      <Truck className="h-4 w-4 text-amber-600" />
                      <span>全国包邮，3-5天送达</span>
                    </div>
                    <div className="flex items-center gap-3 text-sm">
                      <Shield className="h-4 w-4 text-amber-600" />
                      <span>7天无理由退换货</span>
                    </div>
                    <div className="flex items-center gap-3 text-sm">
                      <CreditCard className="h-4 w-4 text-amber-600" />
                      <span>支持多种支付方式</span>
                    </div>
                  </div>
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