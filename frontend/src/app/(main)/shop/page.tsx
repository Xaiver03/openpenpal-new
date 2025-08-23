'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  ShoppingCart, 
  Package, 
  Gift,
  Search,
  Filter,
  CreditCard,
  Coins,
  Store,
  Star,
  TrendingUp,
  AlertCircle,
  X
} from 'lucide-react'
import { BackButton } from '@/components/ui/back-button'
import { useAuth } from '@/contexts/auth-context-new'
import { apiClient } from '@/lib/api-client'
import { formatPrice } from '@/lib/utils'

// 统一商品类型
interface Product {
  id: string
  name: string
  description: string
  price: number          // 现金价格
  credit_price?: number  // 积分价格
  stock: number
  category: string
  type: 'regular' | 'credit' | 'both'  // 支持的购买方式
  features: string[]
  image_url?: string
  rating?: number
  sales_count?: number
  is_popular?: boolean
  discount?: number
}

// 购物车项目
interface CartItem {
  id: string
  product_id: string
  product: Product
  quantity: number
  payment_type: 'cash' | 'credit'
}

export default function UnifiedShopPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { user } = useAuth()
  
  // 从URL参数获取默认类型
  const defaultType = searchParams.get('type') || 'all'
  
  const [activeTab, setActiveTab] = useState(defaultType)
  const [products, setProducts] = useState<Product[]>([])
  const [cart, setCart] = useState<CartItem[]>([])
  const [loading, setLoading] = useState(true)
  const [userCredits, setUserCredits] = useState(0)
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [sortBy, setSortBy] = useState('popular')
  const [showCart, setShowCart] = useState(false)

  // 商品分类
  const categories = [
    { value: 'all', label: '全部商品' },
    { value: 'envelope', label: '信封' },
    { value: 'stamp', label: '邮票' },
    { value: 'stationery', label: '文具' },
    { value: 'gift', label: '礼品' },
    { value: 'virtual', label: '虚拟物品' }
  ]

  useEffect(() => {
    loadProducts()
    loadUserData()
  }, [activeTab, selectedCategory, sortBy])

  const loadProducts = async () => {
    setLoading(true)
    try {
      // 构建查询参数
      const params = new URLSearchParams({
        type: activeTab === 'all' ? '' : activeTab,
        category: selectedCategory === 'all' ? '' : selectedCategory,
        sort: sortBy
      })

      const response = await apiClient.get<any>(`/shop/products?${params}`)
      if (response.data) {
        const data = response.data
        const products = Array.isArray(data) ? data : (data && typeof data === 'object' && 'data' in data ? data.data : [])
        setProducts(products || [])
      }
    } catch (error) {
      console.error('Failed to load products:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadUserData = async () => {
    if (!user) return
    
    try {
      // 加载积分余额
      const creditsRes = await apiClient.get<any>('/credits/balance')
      const creditsData = creditsRes.data
      const balance = creditsData && typeof creditsData === 'object' ? (creditsData.balance ?? creditsData.data?.balance ?? 0) : 0
      setUserCredits(balance)
      
      // 加载购物车
      const cartRes = await apiClient.get<any>('/shop/cart')
      const cartData = cartRes.data
      const items = cartData && typeof cartData === 'object' ? (cartData.items ?? cartData.data?.items ?? []) : []
      setCart(items)
    } catch (error) {
      console.error('Failed to load user data:', error)
    }
  }

  // 添加到购物车
  const addToCart = async (product: Product, paymentType: 'cash' | 'credit') => {
    if (!user) {
      router.push('/login')
      return
    }

    // 检查是否支持该支付方式
    if (paymentType === 'credit' && !product.credit_price) {
      alert('该商品不支持积分购买')
      return
    }

    if (paymentType === 'cash' && product.type === 'credit') {
      alert('该商品仅支持积分购买')
      return
    }

    try {
      await apiClient.post('/shop/cart/add', {
        product_id: product.id,
        quantity: 1,
        payment_type: paymentType
      })
      
      // 重新加载购物车
      await loadUserData()
      setShowCart(true)
    } catch (error) {
      console.error('Failed to add to cart:', error)
      alert('添加到购物车失败')
    }
  }

  // 过滤和搜索商品
  const filteredProducts = products.filter(product => {
    // 按标签页过滤
    if (activeTab === 'cash' && product.type === 'credit') return false
    if (activeTab === 'credit' && product.type === 'regular') return false
    
    // 按搜索词过滤
    if (searchTerm && !product.name.toLowerCase().includes(searchTerm.toLowerCase())) {
      return false
    }
    
    return true
  })

  // 计算购物车总价
  const cartTotal = cart.reduce((total, item) => {
    const price = item.payment_type === 'credit' 
      ? (item.product.credit_price || 0) 
      : item.product.price
    return total + price * item.quantity
  }, 0)

  const cartTotalCredits = cart
    .filter(item => item.payment_type === 'credit')
    .reduce((total, item) => total + (item.product.credit_price || 0) * item.quantity, 0)

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <BackButton />
        <div className="flex items-center justify-between mt-4">
          <div>
            <h1 className="text-3xl font-bold">商城</h1>
            <p className="text-gray-600 mt-2">
              购买信封、邮票和精美文具
            </p>
          </div>
          
          {/* 用户信息 */}
          <div className="flex items-center gap-4">
            <Card className="px-4 py-2">
              <div className="flex items-center gap-2">
                <Coins className="h-4 w-4 text-yellow-600" />
                <span className="font-semibold">{userCredits}</span>
                <span className="text-sm text-gray-600">积分</span>
              </div>
            </Card>
            
            <Button
              onClick={() => setShowCart(!showCart)}
              variant="outline"
              className="relative"
            >
              <ShoppingCart className="h-5 w-5" />
              {cart.length > 0 && (
                <Badge className="absolute -top-2 -right-2 h-6 w-6 rounded-full p-0 flex items-center justify-center">
                  {cart.length}
                </Badge>
              )}
            </Button>
          </div>
        </div>
      </div>

      {/* 搜索和筛选栏 */}
      <div className="mb-6 space-y-4">
        <div className="flex gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
            <Input
              placeholder="搜索商品..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
          
          <Select value={selectedCategory} onValueChange={setSelectedCategory}>
            <SelectTrigger className="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {categories.map(cat => (
                <SelectItem key={cat.value} value={cat.value}>
                  {cat.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          
          <Select value={sortBy} onValueChange={setSortBy}>
            <SelectTrigger className="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="popular">热门商品</SelectItem>
              <SelectItem value="price_asc">价格从低到高</SelectItem>
              <SelectItem value="price_desc">价格从高到低</SelectItem>
              <SelectItem value="newest">最新上架</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* 商品展示 */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="all" className="flex items-center gap-2">
            <Store className="h-4 w-4" />
            全部商品
          </TabsTrigger>
          <TabsTrigger value="cash" className="flex items-center gap-2">
            <CreditCard className="h-4 w-4" />
            现金商品
          </TabsTrigger>
          <TabsTrigger value="credit" className="flex items-center gap-2">
            <Coins className="h-4 w-4" />
            积分商品
          </TabsTrigger>
        </TabsList>

        <TabsContent value={activeTab} className="mt-6">
          {loading ? (
            <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-6">
              {[...Array(8)].map((_, i) => (
                <Card key={i} className="animate-pulse">
                  <div className="h-48 bg-gray-200"></div>
                  <CardContent className="p-4 space-y-3">
                    <div className="h-4 bg-gray-200 rounded"></div>
                    <div className="h-3 bg-gray-200 rounded w-2/3"></div>
                    <div className="h-8 bg-gray-200 rounded"></div>
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-6">
              {filteredProducts.map(product => (
                <Card key={product.id} className="overflow-hidden hover:shadow-lg transition-shadow">
                  {/* 商品图片 */}
                  <div className="h-48 bg-gray-100 relative">
                    {product.image_url ? (
                      <img 
                        src={product.image_url} 
                        alt={product.name}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <Package className="h-16 w-16 text-gray-400" />
                      </div>
                    )}
                    
                    {/* 标签 */}
                    {product.is_popular && (
                      <Badge className="absolute top-2 left-2">热门</Badge>
                    )}
                    {product.discount && (
                      <Badge variant="destructive" className="absolute top-2 right-2">
                        -{product.discount}%
                      </Badge>
                    )}
                  </div>

                  <CardContent className="p-4">
                    <h3 className="font-semibold text-lg mb-2">{product.name}</h3>
                    <p className="text-sm text-gray-600 mb-4 line-clamp-2">
                      {product.description}
                    </p>

                    {/* 价格和购买按钮 */}
                    <div className="space-y-3">
                      {/* 现金价格 */}
                      {product.type !== 'credit' && (
                        <div className="flex items-center justify-between">
                          <div>
                            <span className="text-lg font-bold text-primary">
                              ¥{formatPrice(product.price)}
                            </span>
                            {product.discount && (
                              <span className="text-sm text-gray-400 line-through ml-2">
                                ¥{formatPrice(product.price * (1 + product.discount / 100))}
                              </span>
                            )}
                          </div>
                          <Button
                            size="sm"
                            onClick={() => addToCart(product, 'cash')}
                            disabled={product.stock === 0}
                          >
                            <CreditCard className="h-4 w-4 mr-1" />
                            购买
                          </Button>
                        </div>
                      )}

                      {/* 积分价格 */}
                      {product.credit_price && (
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <span className="text-lg font-bold text-yellow-600">
                              {product.credit_price}
                            </span>
                            <span className="text-sm text-gray-600">积分</span>
                          </div>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => addToCart(product, 'credit')}
                            disabled={product.stock === 0 || userCredits < product.credit_price}
                          >
                            <Coins className="h-4 w-4 mr-1" />
                            兑换
                          </Button>
                        </div>
                      )}
                    </div>

                    {/* 库存状态 */}
                    <div className="mt-3 flex items-center justify-between text-sm">
                      <span className="text-gray-500">
                        库存: {product.stock > 0 ? product.stock : '售罄'}
                      </span>
                      {product.sales_count && (
                        <span className="text-gray-500">
                          已售 {product.sales_count}
                        </span>
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}

          {filteredProducts.length === 0 && !loading && (
            <div className="text-center py-12">
              <Package className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-600">暂无相关商品</p>
            </div>
          )}
        </TabsContent>
      </Tabs>

      {/* 购物车侧边栏 */}
      {showCart && (
        <div className="fixed inset-0 z-50 flex justify-end">
          <div 
            className="absolute inset-0 bg-black/50" 
            onClick={() => setShowCart(false)}
          />
          <div className="relative bg-white w-full max-w-md h-full overflow-y-auto shadow-xl">
            <div className="sticky top-0 bg-white border-b p-4">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-semibold">购物车</h2>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setShowCart(false)}
                >
                  <X className="h-5 w-5" />
                </Button>
              </div>
            </div>

            <div className="p-4">
              {cart.length === 0 ? (
                <div className="text-center py-8">
                  <ShoppingCart className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-600">购物车为空</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {cart.map(item => (
                    <Card key={item.id}>
                      <CardContent className="p-4">
                        <div className="flex justify-between items-start">
                          <div className="flex-1">
                            <h4 className="font-semibold">{item.product.name}</h4>
                            <p className="text-sm text-gray-600 mt-1">
                              {item.payment_type === 'credit' ? (
                                <span className="text-yellow-600">
                                  {item.product.credit_price} 积分
                                </span>
                              ) : (
                                <span>¥{formatPrice(item.product.price)}</span>
                              )}
                              {' × '}{item.quantity}
                            </p>
                          </div>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-red-600"
                            onClick={() => {/* 实现删除功能 */}}
                          >
                            删除
                          </Button>
                        </div>
                      </CardContent>
                    </Card>
                  ))}

                  {/* 总计 */}
                  <div className="border-t pt-4 space-y-2">
                    {cartTotal > 0 && (
                      <div className="flex justify-between text-lg font-semibold">
                        <span>现金总计</span>
                        <span>¥{formatPrice(cartTotal)}</span>
                      </div>
                    )}
                    {cartTotalCredits > 0 && (
                      <div className="flex justify-between text-lg font-semibold text-yellow-600">
                        <span>积分总计</span>
                        <span>{cartTotalCredits} 积分</span>
                      </div>
                    )}
                  </div>

                  {/* 结算按钮 */}
                  <Button 
                    className="w-full" 
                    size="lg"
                    onClick={() => router.push('/shop/checkout')}
                  >
                    去结算
                  </Button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}