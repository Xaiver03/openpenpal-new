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
  Minus
} from 'lucide-react'
import { BackButton } from '@/components/ui/back-button'
import { useAuth } from '@/contexts/auth-context-new'

// 商店商品类型
interface ShopItem {
  id: string
  name: string
  description: string
  price: number
  stock: number
  type: 'normal' | 'drift' | 'bundle'
  school_code?: string
  batch_id?: string
  features: string[]
  image_url?: string
  discount?: number
  popular: boolean
}

// 订单类型
interface Order {
  id: string
  items: Array<{
    item_id: string
    name: string
    quantity: number
    price: number
    type: 'normal' | 'drift' | 'bundle'
  }>
  total_amount: number
  status: 'pending' | 'paid' | 'preparing' | 'ready' | 'completed'
  created_at: string
  pickup_location?: string
  tracking_code?: string
}

// 购物车项目
interface CartItem {
  item_id: string
  name: string
  price: number
  quantity: number
  type: 'normal' | 'drift' | 'bundle'
}

export default function ShopPage() {
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState('products')
  const [products, setProducts] = useState<ShopItem[]>([])
  const [cart, setCart] = useState<CartItem[]>([])
  const [orders, setOrders] = useState<Order[]>([])
  const [selectedSchool, setSelectedSchool] = useState<string>('all')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 模拟商品数据
  useEffect(() => {
    const mockProducts: ShopItem[] = [
      {
        id: '1',
        name: '普通条码包 (10个)',
        description: '适用于定向信件，每个条码都有唯一编号',
        price: 15.00,
        stock: 50,
        type: 'normal',
        school_code: 'BJDX',
        features: ['8位唯一编号', '高质量二维码', '防水标签', '快速绑定'],
        popular: true
      },
      {
        id: '2',
        name: '漂流信条码包 (5个)',
        description: '专为漂流信设计，支持AI智能匹配',
        price: 20.00,
        stock: 30,
        type: 'drift',
        school_code: 'BJDX',
        features: ['AI匹配优化', '情感分析', '特殊标识', '神秘包装'],
        popular: false
      },
      {
        id: '3',
        name: '混合套装 (15个)',
        description: '包含10个普通条码 + 5个漂流信条码',
        price: 30.00,
        stock: 25,
        type: 'bundle',
        school_code: 'BJDX',
        features: ['混合搭配', '超值优惠', '灵活使用', '完整体验'],
        discount: 15,
        popular: true
      },
      {
        id: '4',
        name: '清华大学专版 (10个)',
        description: '清华大学专属条码，校内快速投递',
        price: 18.00,
        stock: 35,
        type: 'normal',
        school_code: 'QHDX',
        features: ['校内专用', '快速投递', '学校定制', '品质保证'],
        popular: false
      },
      {
        id: '5',
        name: '限量版漂流信包 (3个)',
        description: '限量版设计，特殊工艺制作',
        price: 25.00,
        stock: 10,
        type: 'drift',
        features: ['限量发售', '特殊工艺', '收藏价值', '独特体验'],
        popular: false
      }
    ]
    setProducts(mockProducts)
  }, [])

  // 获取学校列表
  const schools = [
    { code: 'all', name: '全部学校' },
    { code: 'BJDX', name: '北京大学' },
    { code: 'QHDX', name: '清华大学' },
    { code: 'BJHK', name: '北京航空航天大学' }
  ]

  // 过滤商品
  const filteredProducts = products.filter(product => 
    selectedSchool === 'all' || product.school_code === selectedSchool
  )

  // 添加到购物车
  const addToCart = (product: ShopItem) => {
    const existingItem = cart.find(item => item.item_id === product.id)
    
    if (existingItem) {
      setCart(cart.map(item => 
        item.item_id === product.id 
          ? { ...item, quantity: item.quantity + 1 }
          : item
      ))
    } else {
      setCart([...cart, {
        item_id: product.id,
        name: product.name,
        price: product.discount ? product.price * (1 - product.discount / 100) : product.price,
        quantity: 1,
        type: product.type
      }])
    }
  }

  // 更新购物车数量
  const updateCartQuantity = (itemId: string, quantity: number) => {
    if (quantity <= 0) {
      setCart(cart.filter(item => item.item_id !== itemId))
    } else {
      setCart(cart.map(item => 
        item.item_id === itemId 
          ? { ...item, quantity }
          : item
      ))
    }
  }

  // 计算总金额
  const getTotalAmount = () => {
    return cart.reduce((total, item) => total + (item.price * item.quantity), 0)
  }

  // 提交订单
  const handleCheckout = async () => {
    if (cart.length === 0) {
      setError('购物车为空')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 2000))

      const newOrder: Order = {
        id: `ORDER${Date.now()}`,
        items: cart,
        total_amount: getTotalAmount(),
        status: 'pending',
        created_at: new Date().toISOString(),
        pickup_location: '学校信使服务点'
      }

      setOrders([newOrder, ...orders])
      setCart([]) // 清空购物车
      setActiveTab('orders') // 跳转到订单页面
      
      alert('订单创建成功！请前往指定地点取货')
    } catch (error) {
      setError('订单创建失败，请重试')
    } finally {
      setIsLoading(false)
    }
  }

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'normal': return <Target className="w-4 h-4" />
      case 'drift': return <Heart className="w-4 h-4 text-red-500" />
      case 'bundle': return <Package className="w-4 h-4 text-purple-500" />
      default: return <QrCode className="w-4 h-4" />
    }
  }

  const getTypeName = (type: string) => {
    switch (type) {
      case 'normal': return '定向信'
      case 'drift': return '漂流信'
      case 'bundle': return '套装'
      default: return '未知'
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-yellow-100 text-yellow-800'
      case 'paid': return 'bg-blue-100 text-blue-800'
      case 'preparing': return 'bg-orange-100 text-orange-800'
      case 'ready': return 'bg-green-100 text-green-800'
      case 'completed': return 'bg-gray-100 text-gray-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending': return '待支付'
      case 'paid': return '已支付'
      case 'preparing': return '准备中'
      case 'ready': return '可取货'
      case 'completed': return '已完成'
      default: return '未知'
    }
  }

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-6xl mx-auto px-4 py-8">
        <div className="flex items-center gap-4 mb-6">
          <BackButton />
          <div>
            <h1 className="text-3xl font-bold text-amber-900 flex items-center gap-2">
              <Store className="w-8 h-8" />
              条码商店
            </h1>
            <p className="text-amber-700">购买条码贴纸，开启写信旅程</p>
          </div>
          
          {/* 购物车图标 */}
          <div className="ml-auto relative">
            <Button
              variant="outline"
              className="border-amber-300 text-amber-700 hover:bg-amber-50"
              onClick={() => setActiveTab('cart')}
            >
              <ShoppingCart className="w-5 h-5 mr-2" />
              购物车
              {cart.length > 0 && (
                <Badge className="ml-2 bg-red-500 text-white">
                  {cart.reduce((sum, item) => sum + item.quantity, 0)}
                </Badge>
              )}
            </Button>
          </div>
        </div>

        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-3 mb-6">
            <TabsTrigger value="products" className="data-[state=active]:bg-amber-600 data-[state=active]:text-white">
              <Package className="w-4 h-4 mr-2" />
              商品浏览
            </TabsTrigger>
            <TabsTrigger value="cart" className="data-[state=active]:bg-amber-600 data-[state=active]:text-white">
              <ShoppingCart className="w-4 h-4 mr-2" />
              购物车
              {cart.length > 0 && (
                <Badge className="ml-2 bg-red-500 text-white">
                  {cart.reduce((sum, item) => sum + item.quantity, 0)}
                </Badge>
              )}
            </TabsTrigger>
            <TabsTrigger value="orders" className="data-[state=active]:bg-amber-600 data-[state=active]:text-white">
              <History className="w-4 h-4 mr-2" />
              我的订单
            </TabsTrigger>
          </TabsList>

          {/* 商品浏览 */}
          <TabsContent value="products">
            <div className="space-y-6">
              {/* 筛选器 */}
              <Card>
                <CardContent className="p-4">
                  <div className="flex items-center gap-4">
                    <Label>选择学校：</Label>
                    <Select value={selectedSchool} onValueChange={setSelectedSchool}>
                      <SelectTrigger className="w-48">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {schools.map((school) => (
                          <SelectItem key={school.code} value={school.code}>
                            {school.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </CardContent>
              </Card>

              {/* 商品列表 */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {filteredProducts.map((product) => (
                  <Card key={product.id} className="border-amber-200 hover:border-amber-400 transition-all">
                    <CardHeader>
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-2">
                          {getTypeIcon(product.type)}
                          <CardTitle className="text-lg">{product.name}</CardTitle>
                        </div>
                        {product.popular && (
                          <Badge className="bg-red-500 text-white">
                            <Star className="w-3 h-3 mr-1" />
                            热门
                          </Badge>
                        )}
                      </div>
                      <CardDescription>{product.description}</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      {/* 特性列表 */}
                      <div className="flex flex-wrap gap-1">
                        {product.features.map((feature, index) => (
                          <Badge key={index} variant="secondary" className="text-xs">
                            {feature}
                          </Badge>
                        ))}
                      </div>

                      {/* 类型和学校 */}
                      <div className="flex items-center gap-2 text-sm text-amber-700">
                        <span>类型：{getTypeName(product.type)}</span>
                        {product.school_code && (
                          <>
                            <span>•</span>
                            <span>{schools.find(s => s.code === product.school_code)?.name}</span>
                          </>
                        )}
                      </div>

                      {/* 价格和库存 */}
                      <div className="flex items-center justify-between">
                        <div className="space-y-1">
                          <div className="flex items-center gap-2">
                            <span className="text-2xl font-bold text-amber-900">
                              ¥{product.discount ? (product.price * (1 - product.discount / 100)).toFixed(2) : product.price.toFixed(2)}
                            </span>
                            {product.discount && (
                              <>
                                <span className="text-sm line-through text-gray-500">
                                  ¥{product.price.toFixed(2)}
                                </span>
                                <Badge className="bg-red-500 text-white text-xs">
                                  -{product.discount}%
                                </Badge>
                              </>
                            )}
                          </div>
                          <div className="text-sm text-amber-600">
                            库存：{product.stock} 件
                          </div>
                        </div>
                      </div>

                      <Button
                        onClick={() => addToCart(product)}
                        disabled={product.stock === 0}
                        className="w-full bg-amber-600 hover:bg-amber-700 text-white"
                      >
                        {product.stock > 0 ? (
                          <>
                            <Plus className="w-4 h-4 mr-2" />
                            加入购物车
                          </>
                        ) : (
                          '暂时缺货'
                        )}
                      </Button>
                    </CardContent>
                  </Card>
                ))}
              </div>

              {filteredProducts.length === 0 && (
                <Card>
                  <CardContent className="py-8 text-center">
                    <Package className="w-16 h-16 text-amber-300 mx-auto mb-4" />
                    <p className="text-amber-600">该学校暂无可用商品</p>
                  </CardContent>
                </Card>
              )}
            </div>
          </TabsContent>

          {/* 购物车 */}
          <TabsContent value="cart">
            <Card>
              <CardHeader>
                <CardTitle>购物车</CardTitle>
                <CardDescription>
                  {cart.length > 0 ? `共 ${cart.reduce((sum, item) => sum + item.quantity, 0)} 件商品` : '购物车为空'}
                </CardDescription>
              </CardHeader>
              <CardContent>
                {cart.length > 0 ? (
                  <div className="space-y-6">
                    {/* 购物车商品列表 */}
                    <div className="space-y-4">
                      {cart.map((item) => (
                        <div key={item.item_id} className="flex items-center gap-4 p-4 border border-amber-200 rounded-lg">
                          <div className="flex items-center gap-2">
                            {getTypeIcon(item.type)}
                            <div className="flex-1">
                              <h3 className="font-semibold text-amber-900">{item.name}</h3>
                              <p className="text-sm text-amber-700">类型：{getTypeName(item.type)}</p>
                            </div>
                          </div>

                          <div className="flex items-center gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => updateCartQuantity(item.item_id, item.quantity - 1)}
                            >
                              <Minus className="w-3 h-3" />
                            </Button>
                            <span className="w-8 text-center">{item.quantity}</span>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => updateCartQuantity(item.item_id, item.quantity + 1)}
                            >
                              <Plus className="w-3 h-3" />
                            </Button>
                          </div>

                          <div className="text-right">
                            <div className="font-semibold text-amber-900">
                              ¥{(item.price * item.quantity).toFixed(2)}
                            </div>
                            <div className="text-sm text-amber-600">
                              单价: ¥{item.price.toFixed(2)}
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>

                    {/* 总计 */}
                    <div className="border-t border-amber-200 pt-4">
                      <div className="flex items-center justify-between text-lg font-semibold text-amber-900">
                        <span>总计：</span>
                        <span>¥{getTotalAmount().toFixed(2)}</span>
                      </div>
                    </div>

                    {/* 结账按钮 */}
                    <Button
                      onClick={handleCheckout}
                      disabled={isLoading}
                      className="w-full bg-amber-600 hover:bg-amber-700 text-white"
                    >
                      {isLoading ? (
                        <>
                          <Clock className="w-4 h-4 mr-2 animate-spin" />
                          处理中...
                        </>
                      ) : (
                        <>
                          <CreditCard className="w-4 h-4 mr-2" />
                          立即结账
                        </>
                      )}
                    </Button>
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <ShoppingCart className="w-16 h-16 text-amber-300 mx-auto mb-4" />
                    <p className="text-amber-600 mb-4">购物车还是空的</p>
                    <Button
                      onClick={() => setActiveTab('products')}
                      className="bg-amber-600 hover:bg-amber-700 text-white"
                    >
                      去购物
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

          {/* 我的订单 */}
          <TabsContent value="orders">
            <Card>
              <CardHeader>
                <CardTitle>我的订单</CardTitle>
                <CardDescription>
                  查看订单状态和取货信息
                </CardDescription>
              </CardHeader>
              <CardContent>
                {orders.length > 0 ? (
                  <div className="space-y-4">
                    {orders.map((order) => (
                      <div key={order.id} className="border border-amber-200 rounded-lg p-4">
                        <div className="flex items-start justify-between mb-4">
                          <div>
                            <h3 className="font-semibold text-amber-900">订单 {order.id}</h3>
                            <p className="text-sm text-amber-700">
                              {new Date(order.created_at).toLocaleString()}
                            </p>
                          </div>
                          <Badge className={getStatusColor(order.status)}>
                            {getStatusText(order.status)}
                          </Badge>
                        </div>

                        <div className="space-y-2 mb-4">
                          {order.items.map((item, index) => (
                            <div key={index} className="flex items-center justify-between text-sm">
                              <span>{item.name} × {item.quantity}</span>
                              <span>¥{(item.price * item.quantity).toFixed(2)}</span>
                            </div>
                          ))}
                        </div>

                        <div className="flex items-center justify-between text-lg font-semibold text-amber-900 border-t border-amber-200 pt-2">
                          <span>总计：</span>
                          <span>¥{order.total_amount.toFixed(2)}</span>
                        </div>

                        {order.pickup_location && (
                          <div className="mt-4 flex items-center gap-2 text-sm text-amber-700">
                            <MapPin className="w-4 h-4" />
                            取货地点：{order.pickup_location}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <History className="w-16 h-16 text-amber-300 mx-auto mb-4" />
                    <p className="text-amber-600">暂无订单记录</p>
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