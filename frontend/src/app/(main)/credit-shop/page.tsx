'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  ShoppingCart, 
  Package, 
  QrCode, 
  Heart, 
  Target,
  Star,
  CheckCircle,
  AlertCircle,
  Truck,
  Clock,
  MapPin,
  CreditCard,
  History,
  Store,
  Plus,
  Minus,
  Coins,
  Gift,
  Award,
  Search,
  Filter,
  Eye,
  X
} from 'lucide-react'
import { BackButton } from '@/components/ui/back-button'
import { useAuth } from '@/contexts/auth-context-new'

// 积分商城商品类型
interface CreditShopProduct {
  id: string
  name: string
  description: string
  short_desc: string
  category: string
  product_type: 'physical' | 'virtual' | 'service' | 'voucher'
  credit_price: number
  original_price: number
  stock: number
  total_stock: number
  redeem_count: number
  image_url?: string
  images?: string[]
  tags?: string[]
  specifications?: Record<string, any>
  status: 'draft' | 'active' | 'inactive' | 'sold_out' | 'deleted'
  is_featured: boolean
  is_limited: boolean
  limit_per_user: number
  priority: number
  valid_from?: string
  valid_to?: string
  created_at: string
  updated_at: string
}

// 积分兑换订单类型
interface CreditRedemption {
  id: string
  redemption_no: string
  user_id: string
  product_id: string
  product?: CreditShopProduct
  quantity: number
  credit_price: number
  total_credits: number
  status: 'pending' | 'confirmed' | 'processing' | 'shipped' | 'delivered' | 'completed' | 'cancelled' | 'refunded'
  delivery_info?: Record<string, any>
  redemption_code?: string
  tracking_number?: string
  notes?: string
  processed_at?: string
  shipped_at?: string
  delivered_at?: string
  completed_at?: string
  cancelled_at?: string
  created_at: string
  updated_at: string
}

// 积分购物车项目
interface CreditCartItem {
  id: string
  cart_id: string
  product_id: string
  product?: CreditShopProduct
  quantity: number
  credit_price: number
  subtotal: number
  created_at: string
  updated_at: string
}

// 积分购物车
interface CreditCart {
  id: string
  user_id: string
  items: CreditCartItem[]
  total_items: number
  total_credits: number
  created_at: string
  updated_at: string
}

// 用户积分信息
interface UserCredit {
  available: number
  total: number
  used: number
  level: number
}

export default function CreditShopPage() {
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState('products')
  const [products, setProducts] = useState<CreditShopProduct[]>([])
  const [cart, setCart] = useState<CreditCart | null>(null)
  const [redemptions, setRedemptions] = useState<CreditRedemption[]>([])
  const [userCredit, setUserCredit] = useState<UserCredit | null>(null)
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCategory, setSelectedCategory] = useState<string>('all')
  const [selectedType, setSelectedType] = useState<string>('all')
  const [searchKeyword, setSearchKeyword] = useState('')
  const [sortBy, setSortBy] = useState('priority')
  const [showFeaturedOnly, setShowFeaturedOnly] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 获取用户积分余额
  useEffect(() => {
    fetchUserCredit()
  }, [])

  // 获取积分商品列表
  useEffect(() => {
    fetchProducts()
  }, [selectedCategory, selectedType, searchKeyword, sortBy, showFeaturedOnly])

  // 获取积分购物车
  useEffect(() => {
    fetchCart()
  }, [])

  // 获取兑换订单
  useEffect(() => {
    if (activeTab === 'orders') {
      fetchRedemptions()
    }
  }, [activeTab])

  const fetchUserCredit = async () => {
    if (!user) return
    
    try {
      const response = await fetch('/api/v1/credit-shop/balance', {
        headers: {
          'Authorization': `Bearer ${user.token}`
        }
      })
      
      if (response.ok) {
        const data = await response.json()
        setUserCredit(data)
      }
    } catch (error) {
      console.error('Failed to fetch user credit:', error)
    }
  }

  const fetchProducts = async () => {
    setIsLoading(true)
    try {
      const params = new URLSearchParams()
      params.append('page', '1')
      params.append('limit', '50')
      
      if (selectedCategory !== 'all') {
        params.append('category', selectedCategory)
      }
      if (selectedType !== 'all') {
        params.append('product_type', selectedType)
      }
      if (searchKeyword) {
        params.append('keyword', searchKeyword)
      }
      if (sortBy) {
        params.append('sort_by', sortBy)
      }
      if (showFeaturedOnly) {
        params.append('featured_only', 'true')
      }
      params.append('in_stock_only', 'true')

      const response = await fetch(`/api/v1/credit-shop/products?${params}`)
      
      if (response.ok) {
        const data = await response.json()
        setProducts(data.items || [])
        
        // 提取分类
        const uniqueCategories = [...new Set(data.items?.map((p: CreditShopProduct) => p.category).filter(Boolean))]
        setCategories(uniqueCategories)
      }
    } catch (error) {
      console.error('Failed to fetch products:', error)
      setError('获取商品列表失败')
    } finally {
      setIsLoading(false)
    }
  }

  const fetchCart = async () => {
    if (!user) return
    
    try {
      const response = await fetch('/api/v1/credit-shop/cart', {
        headers: {
          'Authorization': `Bearer ${user.token}`
        }
      })
      
      if (response.ok) {
        const data = await response.json()
        setCart(data)
      }
    } catch (error) {
      console.error('Failed to fetch cart:', error)
    }
  }

  const fetchRedemptions = async () => {
    if (!user) return
    
    try {
      const response = await fetch('/api/v1/credit-shop/redemptions?page=1&limit=20', {
        headers: {
          'Authorization': `Bearer ${user.token}`
        }
      })
      
      if (response.ok) {
        const data = await response.json()
        setRedemptions(data.items || [])
      }
    } catch (error) {
      console.error('Failed to fetch redemptions:', error)
    }
  }

  const addToCart = async (product: CreditShopProduct) => {
    if (!user) {
      setError('请先登录')
      return
    }

    setIsLoading(true)
    try {
      const response = await fetch('/api/v1/credit-shop/cart/items', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${user.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          product_id: product.id,
          quantity: 1
        })
      })

      if (response.ok) {
        await fetchCart()
        setError(null)
      } else {
        const errorData = await response.json()
        setError(errorData.message || '添加到购物车失败')
      }
    } catch (error) {
      setError('添加到购物车失败')
    } finally {
      setIsLoading(false)
    }
  }

  const updateCartItem = async (itemId: string, quantity: number) => {
    if (!user) return

    try {
      const response = await fetch(`/api/v1/credit-shop/cart/items/${itemId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${user.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ quantity })
      })

      if (response.ok) {
        await fetchCart()
      }
    } catch (error) {
      console.error('Failed to update cart item:', error)
    }
  }

  const removeCartItem = async (itemId: string) => {
    if (!user) return

    try {
      const response = await fetch(`/api/v1/credit-shop/cart/items/${itemId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${user.token}`
        }
      })

      if (response.ok) {
        await fetchCart()
      }
    } catch (error) {
      console.error('Failed to remove cart item:', error)
    }
  }

  const createRedemption = async (productId: string, quantity: number, deliveryInfo?: Record<string, any>) => {
    if (!user) return

    setIsLoading(true)
    try {
      const response = await fetch('/api/v1/credit-shop/redemptions', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${user.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          product_id: productId,
          quantity,
          delivery_info: deliveryInfo,
          notes: ''
        })
      })

      if (response.ok) {
        await fetchUserCredit()
        await fetchCart()
        setActiveTab('orders')
        setError(null)
      } else {
        const errorData = await response.json()
        setError(errorData.message || '兑换失败')
      }
    } catch (error) {
      setError('兑换失败')
    } finally {
      setIsLoading(false)
    }
  }

  const createBatchRedemption = async (deliveryInfo?: Record<string, any>) => {
    if (!user || !cart || cart.items.length === 0) return

    setIsLoading(true)
    try {
      const response = await fetch('/api/v1/credit-shop/redemptions/from-cart', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${user.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          delivery_info: deliveryInfo
        })
      })

      if (response.ok) {
        await fetchUserCredit()
        await fetchCart()
        setActiveTab('orders')
        setError(null)
      } else {
        const errorData = await response.json()
        setError(errorData.message || '批量兑换失败')
      }
    } catch (error) {
      setError('批量兑换失败')
    } finally {
      setIsLoading(false)
    }
  }

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'physical': return <Package className="w-4 h-4 text-blue-500" />
      case 'virtual': return <QrCode className="w-4 h-4 text-purple-500" />
      case 'service': return <Target className="w-4 h-4 text-green-500" />
      case 'voucher': return <Gift className="w-4 h-4 text-orange-500" />
      default: return <Package className="w-4 h-4" />
    }
  }

  const getTypeName = (type: string) => {
    switch (type) {
      case 'physical': return '实物商品'
      case 'virtual': return '虚拟商品'
      case 'service': return '服务类'
      case 'voucher': return '优惠券'
      default: return '未知'
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-yellow-100 text-yellow-800'
      case 'confirmed': return 'bg-blue-100 text-blue-800'
      case 'processing': return 'bg-orange-100 text-orange-800'
      case 'shipped': return 'bg-purple-100 text-purple-800'
      case 'delivered': return 'bg-green-100 text-green-800'
      case 'completed': return 'bg-gray-100 text-gray-800'
      case 'cancelled': return 'bg-red-100 text-red-800'
      case 'refunded': return 'bg-pink-100 text-pink-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending': return '待处理'
      case 'confirmed': return '已确认'
      case 'processing': return '处理中'
      case 'shipped': return '已发货'
      case 'delivered': return '已送达'
      case 'completed': return '已完成'
      case 'cancelled': return '已取消'
      case 'refunded': return '已退款'
      default: return '未知'
    }
  }

  const filteredProducts = products.filter(product => {
    if (selectedCategory !== 'all' && product.category !== selectedCategory) return false
    if (selectedType !== 'all' && product.product_type !== selectedType) return false
    if (showFeaturedOnly && !product.is_featured) return false
    if (searchKeyword && !product.name.toLowerCase().includes(searchKeyword.toLowerCase())) return false
    return true
  })

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50">
      <div className="container max-w-7xl mx-auto px-4 py-8">
        <div className="flex items-center gap-4 mb-6">
          <BackButton />
          <div className="flex-1">
            <h1 className="text-3xl font-bold text-blue-900 flex items-center gap-2">
              <Coins className="w-8 h-8 text-yellow-500" />
              积分商城
            </h1>
            <p className="text-blue-700">用积分兑换心仪商品</p>
          </div>
          
          {/* 用户积分显示 */}
          {userCredit && (
            <Card className="bg-gradient-to-r from-yellow-400 to-orange-400 text-white border-0">
              <CardContent className="p-4">
                <div className="flex items-center gap-2">
                  <Coins className="w-5 h-5" />
                  <div>
                    <div className="text-sm opacity-90">可用积分</div>
                    <div className="text-xl font-bold">{userCredit.available.toLocaleString()}</div>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
          
          {/* 购物车图标 */}
          <div className="relative">
            <Button
              variant="outline"
              className="border-blue-300 text-blue-700 hover:bg-blue-50"
              onClick={() => setActiveTab('cart')}
            >
              <ShoppingCart className="w-5 h-5 mr-2" />
              购物车
              {cart && cart.items.length > 0 && (
                <Badge className="ml-2 bg-red-500 text-white">
                  {cart.items.reduce((sum, item) => sum + item.quantity, 0)}
                </Badge>
              )}
            </Button>
          </div>
        </div>

        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-3 mb-6">
            <TabsTrigger value="products" className="data-[state=active]:bg-blue-600 data-[state=active]:text-white">
              <Store className="w-4 h-4 mr-2" />
              商品浏览
            </TabsTrigger>
            <TabsTrigger value="cart" className="data-[state=active]:bg-blue-600 data-[state=active]:text-white">
              <ShoppingCart className="w-4 h-4 mr-2" />
              购物车
              {cart && cart.items.length > 0 && (
                <Badge className="ml-2 bg-red-500 text-white">
                  {cart.items.reduce((sum, item) => sum + item.quantity, 0)}
                </Badge>
              )}
            </TabsTrigger>
            <TabsTrigger value="orders" className="data-[state=active]:bg-blue-600 data-[state=active]:text-white">
              <History className="w-4 h-4 mr-2" />
              兑换记录
            </TabsTrigger>
          </TabsList>

          {/* 商品浏览 */}
          <TabsContent value="products">
            <div className="space-y-6">
              {/* 筛选器 */}
              <Card>
                <CardContent className="p-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                    <div className="space-y-2">
                      <Label>搜索商品</Label>
                      <div className="relative">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                        <Input
                          placeholder="商品名称..."
                          value={searchKeyword}
                          onChange={(e) => setSearchKeyword(e.target.value)}
                          className="pl-10"
                        />
                      </div>
                    </div>

                    <div className="space-y-2">
                      <Label>商品分类</Label>
                      <Select value={selectedCategory} onValueChange={setSelectedCategory}>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="all">全部分类</SelectItem>
                          {categories.map((category) => (
                            <SelectItem key={category} value={category}>
                              {category}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-2">
                      <Label>商品类型</Label>
                      <Select value={selectedType} onValueChange={setSelectedType}>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="all">全部类型</SelectItem>
                          <SelectItem value="physical">实物商品</SelectItem>
                          <SelectItem value="virtual">虚拟商品</SelectItem>
                          <SelectItem value="service">服务类</SelectItem>
                          <SelectItem value="voucher">优惠券</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-2">
                      <Label>排序方式</Label>
                      <Select value={sortBy} onValueChange={setSortBy}>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="priority">推荐排序</SelectItem>
                          <SelectItem value="credit_price_asc">积分价格升序</SelectItem>
                          <SelectItem value="credit_price_desc">积分价格降序</SelectItem>
                          <SelectItem value="redeem_count_desc">兑换量降序</SelectItem>
                          <SelectItem value="created_at_desc">最新上架</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  <div className="flex items-center gap-4 mt-4">
                    <Button
                      variant={showFeaturedOnly ? "default" : "outline"}
                      size="sm"
                      onClick={() => setShowFeaturedOnly(!showFeaturedOnly)}
                    >
                      <Star className="w-4 h-4 mr-2" />
                      仅显示推荐
                    </Button>
                  </div>
                </CardContent>
              </Card>

              {/* 商品列表 */}
              {isLoading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                  {[...Array(8)].map((_, index) => (
                    <Card key={index} className="animate-pulse">
                      <CardHeader>
                        <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                        <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                      </CardHeader>
                      <CardContent>
                        <div className="space-y-3">
                          <div className="h-20 bg-gray-200 rounded"></div>
                          <div className="h-8 bg-gray-200 rounded w-1/3"></div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                  {filteredProducts.map((product) => (
                    <Card key={product.id} className="border-blue-200 hover:border-blue-400 transition-all hover:shadow-lg">
                      <CardHeader>
                        <div className="flex items-start justify-between">
                          <div className="flex items-center gap-2">
                            {getTypeIcon(product.product_type)}
                            <CardTitle className="text-lg">{product.name}</CardTitle>
                          </div>
                          {product.is_featured && (
                            <Badge className="bg-red-500 text-white">
                              <Star className="w-3 h-3 mr-1" />
                              推荐
                            </Badge>
                          )}
                        </div>
                        <CardDescription className="line-clamp-2">{product.short_desc || product.description}</CardDescription>
                      </CardHeader>
                      <CardContent className="space-y-4">
                        {/* 商品标签 */}
                        {product.tags && product.tags.length > 0 && (
                          <div className="flex flex-wrap gap-1">
                            {product.tags.slice(0, 3).map((tag, index) => (
                              <Badge key={index} variant="secondary" className="text-xs">
                                {tag}
                              </Badge>
                            ))}
                          </div>
                        )}

                        {/* 类型和分类 */}
                        <div className="flex items-center gap-2 text-sm text-blue-700">
                          <span>{getTypeName(product.product_type)}</span>
                          {product.category && (
                            <>
                              <span>•</span>
                              <span>{product.category}</span>
                            </>
                          )}
                        </div>

                        {/* 积分价格和库存 */}
                        <div className="space-y-2">
                          <div className="flex items-center gap-2">
                            <Coins className="w-5 h-5 text-yellow-500" />
                            <span className="text-2xl font-bold text-blue-900">
                              {product.credit_price.toLocaleString()}
                            </span>
                            <span className="text-sm text-gray-500">积分</span>
                          </div>
                          
                          {product.original_price > 0 && (
                            <div className="text-sm text-gray-500">
                              原价: ¥{product.original_price.toFixed(2)}
                            </div>
                          )}
                          
                          <div className="flex items-center justify-between text-sm">
                            <span className="text-blue-600">
                              库存: {product.stock} 件
                            </span>
                            <span className="text-gray-500">
                              已兑: {product.redeem_count} 次
                            </span>
                          </div>
                        </div>

                        {/* 限购提示 */}
                        {product.is_limited && product.limit_per_user > 0 && (
                          <div className="text-xs text-orange-600 bg-orange-50 p-2 rounded">
                            限购 {product.limit_per_user} 件
                          </div>
                        )}

                        {/* 操作按钮 */}
                        <div className="flex gap-2">
                          <Button
                            onClick={() => addToCart(product)}
                            disabled={product.stock === 0 || isLoading}
                            className="flex-1 bg-blue-600 hover:bg-blue-700 text-white"
                            size="sm"
                          >
                            {product.stock > 0 ? (
                              <>
                                <Plus className="w-4 h-4 mr-1" />
                                加入购物车
                              </>
                            ) : (
                              '暂时缺货'
                            )}
                          </Button>
                          
                          <Button
                            onClick={() => createRedemption(product.id, 1)}
                            disabled={product.stock === 0 || isLoading || !userCredit || userCredit.available < product.credit_price}
                            variant="outline"
                            size="sm"
                          >
                            <Award className="w-4 h-4" />
                          </Button>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              )}

              {filteredProducts.length === 0 && !isLoading && (
                <Card>
                  <CardContent className="py-12 text-center">
                    <Package className="w-16 h-16 text-blue-300 mx-auto mb-4" />
                    <p className="text-blue-600">没有找到符合条件的商品</p>
                  </CardContent>
                </Card>
              )}
            </div>
          </TabsContent>

          {/* 购物车 */}
          <TabsContent value="cart">
            <Card>
              <CardHeader>
                <CardTitle>积分购物车</CardTitle>
                <CardDescription>
                  {cart && cart.items.length > 0 ? 
                    `共 ${cart.total_items} 件商品，需要 ${cart.total_credits.toLocaleString()} 积分` : 
                    '购物车为空'
                  }
                </CardDescription>
              </CardHeader>
              <CardContent>
                {cart && cart.items.length > 0 ? (
                  <div className="space-y-6">
                    {/* 购物车商品列表 */}
                    <div className="space-y-4">
                      {cart.items.map((item) => (
                        <div key={item.id} className="flex items-center gap-4 p-4 border border-blue-200 rounded-lg">
                          <div className="flex items-center gap-2 flex-1">
                            {item.product && getTypeIcon(item.product.product_type)}
                            <div className="flex-1">
                              <h3 className="font-semibold text-blue-900">
                                {item.product?.name || 'Unknown Product'}
                              </h3>
                              <p className="text-sm text-blue-700">
                                {item.product && getTypeName(item.product.product_type)}
                              </p>
                            </div>
                          </div>

                          <div className="flex items-center gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => updateCartItem(item.id, item.quantity - 1)}
                              disabled={item.quantity <= 1}
                            >
                              <Minus className="w-3 h-3" />
                            </Button>
                            <span className="w-8 text-center">{item.quantity}</span>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => updateCartItem(item.id, item.quantity + 1)}
                            >
                              <Plus className="w-3 h-3" />
                            </Button>
                          </div>

                          <div className="text-right">
                            <div className="font-semibold text-blue-900 flex items-center gap-1">
                              <Coins className="w-4 h-4 text-yellow-500" />
                              {item.subtotal.toLocaleString()}
                            </div>
                            <div className="text-sm text-blue-600">
                              单价: {item.credit_price.toLocaleString()} 积分
                            </div>
                          </div>

                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => removeCartItem(item.id)}
                            className="text-red-500 hover:text-red-700"
                          >
                            <X className="w-4 h-4" />
                          </Button>
                        </div>
                      ))}
                    </div>

                    {/* 总计 */}
                    <div className="border-t border-blue-200 pt-4">
                      <div className="flex items-center justify-between text-lg font-semibold text-blue-900">
                        <span>总计：</span>
                        <div className="flex items-center gap-1">
                          <Coins className="w-5 h-5 text-yellow-500" />
                          <span>{cart.total_credits.toLocaleString()} 积分</span>
                        </div>
                      </div>
                      
                      {userCredit && (
                        <div className="flex items-center justify-between text-sm text-blue-600 mt-2">
                          <span>可用积分：</span>
                          <span>{userCredit.available.toLocaleString()} 积分</span>
                        </div>
                      )}
                    </div>

                    {/* 批量兑换按钮 */}
                    <Button
                      onClick={() => createBatchRedemption()}
                      disabled={isLoading || !userCredit || userCredit.available < cart.total_credits}
                      className="w-full bg-blue-600 hover:bg-blue-700 text-white"
                    >
                      {isLoading ? (
                        <>
                          <Clock className="w-4 h-4 mr-2 animate-spin" />
                          处理中...
                        </>
                      ) : (
                        <>
                          <Award className="w-4 h-4 mr-2" />
                          立即兑换
                        </>
                      )}
                    </Button>
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <ShoppingCart className="w-16 h-16 text-blue-300 mx-auto mb-4" />
                    <p className="text-blue-600 mb-4">购物车还是空的</p>
                    <Button
                      onClick={() => setActiveTab('products')}
                      className="bg-blue-600 hover:bg-blue-700 text-white"
                    >
                      去逛逛
                    </Button>
                  </div>
                )}

                {error && (
                  <Alert variant="destructive" className="mt-4">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          {/* 兑换记录 */}
          <TabsContent value="orders">
            <Card>
              <CardHeader>
                <CardTitle>兑换记录</CardTitle>
                <CardDescription>
                  查看兑换订单状态和详情
                </CardDescription>
              </CardHeader>
              <CardContent>
                {redemptions.length > 0 ? (
                  <div className="space-y-4">
                    {redemptions.map((redemption) => (
                      <div key={redemption.id} className="border border-blue-200 rounded-lg p-4">
                        <div className="flex items-start justify-between mb-4">
                          <div>
                            <h3 className="font-semibold text-blue-900">
                              订单 {redemption.redemption_no}
                            </h3>
                            <p className="text-sm text-blue-700">
                              {new Date(redemption.created_at).toLocaleString()}
                            </p>
                            {redemption.product && (
                              <p className="text-sm text-gray-600 mt-1">
                                {redemption.product.name} × {redemption.quantity}
                              </p>
                            )}
                          </div>
                          <Badge className={getStatusColor(redemption.status)}>
                            {getStatusText(redemption.status)}
                          </Badge>
                        </div>

                        <div className="flex items-center justify-between text-lg font-semibold text-blue-900 border-t border-blue-200 pt-2">
                          <span>消耗积分：</span>
                          <div className="flex items-center gap-1">
                            <Coins className="w-5 h-5 text-yellow-500" />
                            <span>{redemption.total_credits.toLocaleString()}</span>
                          </div>
                        </div>

                        {redemption.redemption_code && (
                          <div className="mt-4 p-3 bg-green-50 rounded-lg border border-green-200">
                            <div className="text-sm text-green-700 mb-1">兑换码</div>
                            <div className="text-lg font-mono font-bold text-green-900">
                              {redemption.redemption_code}
                            </div>
                          </div>
                        )}

                        {redemption.tracking_number && (
                          <div className="mt-4 flex items-center gap-2 text-sm text-blue-700">
                            <Truck className="w-4 h-4" />
                            物流单号：{redemption.tracking_number}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <History className="w-16 h-16 text-blue-300 mx-auto mb-4" />
                    <p className="text-blue-600">暂无兑换记录</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}