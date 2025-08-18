'use client'

import { useState, useEffect } from 'react'
import { apiClient } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { 
  Package, 
  Plus,
  Edit,
  Trash2,
  Eye,
  BarChart3,
  Users,
  DollarSign,
  TrendingUp,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  Package2,
  Store,
  Settings,
  FileText,
  Search,
  Filter,
  Download,
  Upload,
  ShoppingCart,
  Award,
  Calendar,
  MapPin,
  Truck
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
  user?: {
    username: string
    email: string
  }
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

// 统计数据类型
interface CreditShopStats {
  total_products: number
  active_products: number
  total_redemptions: number
  pending_redemptions: number
  total_credits_consumed: number
  total_revenue: number
  popular_products: Array<{
    product_id: string
    product_name: string
    redeem_count: number
  }>
  recent_redemptions: CreditRedemption[]
}

export default function AdminCreditShopPage() {
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState('overview')
  const [products, setProducts] = useState<CreditShopProduct[]>([])
  const [redemptions, setRedemptions] = useState<CreditRedemption[]>([])
  const [stats, setStats] = useState<CreditShopStats | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 商品管理状态
  const [editingProduct, setEditingProduct] = useState<CreditShopProduct | null>(null)
  const [showProductDialog, setShowProductDialog] = useState(false)
  const [productFilter, setProductFilter] = useState({
    status: 'all',
    category: 'all',
    type: 'all',
    keyword: ''
  })

  // 订单管理状态
  const [selectedRedemption, setSelectedRedemption] = useState<CreditRedemption | null>(null)
  const [showRedemptionDialog, setShowRedemptionDialog] = useState(false)
  const [redemptionFilter, setRedemptionFilter] = useState({
    status: 'all',
    user_id: '',
    product_type: 'all',
    date_range: 'all'
  })

  useEffect(() => {
    if (activeTab === 'overview') {
      fetchStats()
    } else if (activeTab === 'products') {
      fetchProducts()
    } else if (activeTab === 'orders') {
      fetchRedemptions()
    }
  }, [activeTab])

  const fetchStats = async () => {
    if (!user) return
    
    setIsLoading(true)
    try {
      const response = await apiClient.get('/admin/credit-shop/stats')
      
      if (response.success) {
        setStats(response.data as CreditShopStats)
      } else {
        setError(response.message || '获取统计数据失败')
      }
    } catch (error) {
      console.error('Failed to fetch stats:', error)
      setError('获取统计数据失败')
    } finally {
      setIsLoading(false)
    }
  }

  const fetchProducts = async () => {
    if (!user) return
    
    setIsLoading(true)
    try {
      const params = new URLSearchParams()
      params.append('page', '1')
      params.append('limit', '100')
      
      if (productFilter.status !== 'all') {
        params.append('status', productFilter.status)
      }
      if (productFilter.category !== 'all') {
        params.append('category', productFilter.category)
      }
      if (productFilter.type !== 'all') {
        params.append('product_type', productFilter.type)
      }
      if (productFilter.keyword) {
        params.append('keyword', productFilter.keyword)
      }

      const response = await apiClient.get(`/admin/credit-shop/products?${params}`)
      
      if (response.success) {
        const data = response.data as any
        setProducts(data.items || [])
      }
    } catch (error) {
      console.error('Failed to fetch products:', error)
      setError('获取商品列表失败')
    } finally {
      setIsLoading(false)
    }
  }

  const fetchRedemptions = async () => {
    if (!user) return
    
    setIsLoading(true)
    try {
      const params = new URLSearchParams()
      params.append('page', '1')
      params.append('limit', '100')
      
      if (redemptionFilter.status !== 'all') {
        params.append('status', redemptionFilter.status)
      }
      if (redemptionFilter.user_id) {
        params.append('user_id', redemptionFilter.user_id)
      }
      if (redemptionFilter.product_type !== 'all') {
        params.append('product_type', redemptionFilter.product_type)
      }

      const response = await apiClient.get(`/admin/credit-shop/redemptions?${params}`)
      
      if (response.success) {
        const data = response.data as any
        setRedemptions(data.items || [])
      }
    } catch (error) {
      console.error('Failed to fetch redemptions:', error)
      setError('获取兑换订单失败')
    } finally {
      setIsLoading(false)
    }
  }

  const createOrUpdateProduct = async (productData: Partial<CreditShopProduct>) => {
    if (!user) return
    
    setIsLoading(true)
    try {
      const url = editingProduct 
        ? `/admin/credit-shop/products/${editingProduct.id}`
        : '/admin/credit-shop/products'
      
      const method = editingProduct ? 'PUT' : 'POST'
      
      const response = method === 'POST' 
        ? await apiClient.post(url, productData)
        : await apiClient.put(url, productData)

      if (response.success) {
        await fetchProducts()
        setShowProductDialog(false)
        setEditingProduct(null)
        setError(null)
      } else {
        const errorData = response.data as any
        setError(errorData.message || '操作失败')
      }
    } catch (error) {
      setError('操作失败')
    } finally {
      setIsLoading(false)
    }
  }

  const deleteProduct = async (productId: string) => {
    if (!user || !confirm('确定要删除这个商品吗？')) return
    
    setIsLoading(true)
    try {
      const response = await apiClient.delete(`/admin/credit-shop/products/${productId}`)

      if (response.success) {
        await fetchProducts()
        setError(null)
      } else {
        const errorData = response.data as any
        setError(errorData.message || '删除失败')
      }
    } catch (error) {
      setError('删除失败')
    } finally {
      setIsLoading(false)
    }
  }

  const updateRedemptionStatus = async (redemptionId: string, status: string, adminNote?: string) => {
    if (!user) return
    
    setIsLoading(true)
    try {
      const response = await apiClient.put(`/admin/credit-shop/redemptions/${redemptionId}/status`, {
        status,
        admin_note: adminNote || ''
      })

      if (response.success) {
        await fetchRedemptions()
        setShowRedemptionDialog(false)
        setSelectedRedemption(null)
        setError(null)
      } else {
        const errorData = response.data as any
        setError(errorData.message || '更新失败')
      }
    } catch (error) {
      setError('更新失败')
    } finally {
      setIsLoading(false)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-100 text-green-800'
      case 'inactive': return 'bg-gray-100 text-gray-800'
      case 'draft': return 'bg-yellow-100 text-yellow-800'
      case 'sold_out': return 'bg-red-100 text-red-800'
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
      case 'active': return '上架中'
      case 'inactive': return '已下架'
      case 'draft': return '草稿'
      case 'sold_out': return '售罄'
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

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'physical': return <Package className="w-4 h-4 text-blue-500" />
      case 'virtual': return <Award className="w-4 h-4 text-purple-500" />
      case 'service': return <Settings className="w-4 h-4 text-green-500" />
      case 'voucher': return <FileText className="w-4 h-4 text-orange-500" />
      default: return <Package className="w-4 h-4" />
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container max-w-7xl mx-auto px-4 py-8">
        <div className="flex items-center gap-4 mb-6">
          <BackButton />
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-2">
              <Store className="w-8 h-8 text-blue-600" />
              积分商城管理
            </h1>
            <p className="text-gray-600">管理积分商城商品和兑换订单</p>
          </div>
        </div>

        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-4 mb-6">
            <TabsTrigger value="overview">
              <BarChart3 className="w-4 h-4 mr-2" />
              数据概览
            </TabsTrigger>
            <TabsTrigger value="products">
              <Package className="w-4 h-4 mr-2" />
              商品管理
            </TabsTrigger>
            <TabsTrigger value="orders">
              <ShoppingCart className="w-4 h-4 mr-2" />
              订单管理
            </TabsTrigger>
            <TabsTrigger value="settings">
              <Settings className="w-4 h-4 mr-2" />
              配置管理
            </TabsTrigger>
          </TabsList>

          {/* 数据概览 */}
          <TabsContent value="overview">
            <div className="space-y-6">
              {/* 统计卡片 */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">商品总数</CardTitle>
                    <Package2 className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{stats?.total_products || 0}</div>
                    <p className="text-xs text-muted-foreground">
                      其中 {stats?.active_products || 0} 个上架中
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">兑换订单</CardTitle>
                    <ShoppingCart className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{stats?.total_redemptions || 0}</div>
                    <p className="text-xs text-muted-foreground">
                      其中 {stats?.pending_redemptions || 0} 个待处理
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">积分消耗</CardTitle>
                    <TrendingUp className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {stats?.total_credits_consumed?.toLocaleString() || 0}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      累计积分消耗
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">等值金额</CardTitle>
                    <DollarSign className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      ¥{stats?.total_revenue?.toFixed(2) || '0.00'}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      累计等值金额
                    </p>
                  </CardContent>
                </Card>
              </div>

              {/* 热门商品和最近订单 */}
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <Card>
                  <CardHeader>
                    <CardTitle>热门商品</CardTitle>
                    <CardDescription>兑换次数最多的商品</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      {stats?.popular_products?.slice(0, 5).map((item, index) => (
                        <div key={item.product_id} className="flex items-center gap-4">
                          <Badge variant="secondary">{index + 1}</Badge>
                          <div className="flex-1">
                            <div className="font-medium">{item.product_name}</div>
                            <div className="text-sm text-gray-500">
                              兑换 {item.redeem_count} 次
                            </div>
                          </div>
                        </div>
                      )) || (
                        <p className="text-gray-500 text-center py-8">暂无数据</p>
                      )}
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>最近订单</CardTitle>
                    <CardDescription>最新的兑换订单</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      {stats?.recent_redemptions?.slice(0, 5).map((redemption) => (
                        <div key={redemption.id} className="flex items-center gap-4">
                          <Badge className={getStatusColor(redemption.status)}>
                            {getStatusText(redemption.status)}
                          </Badge>
                          <div className="flex-1">
                            <div className="font-medium">{redemption.redemption_no}</div>
                            <div className="text-sm text-gray-500">
                              {redemption.user?.username || 'Unknown User'}
                            </div>
                          </div>
                          <div className="text-right">
                            <div className="text-sm font-medium">
                              {redemption.total_credits.toLocaleString()} 积分
                            </div>
                            <div className="text-xs text-gray-500">
                              {new Date(redemption.created_at).toLocaleDateString()}
                            </div>
                          </div>
                        </div>
                      )) || (
                        <p className="text-gray-500 text-center py-8">暂无数据</p>
                      )}
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          </TabsContent>

          {/* 商品管理 */}
          <TabsContent value="products">
            <div className="space-y-6">
              {/* 操作栏 */}
              <Card>
                <CardContent className="p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      {/* 搜索 */}
                      <div className="relative">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                        <Input
                          placeholder="搜索商品..."
                          value={productFilter.keyword}
                          onChange={(e) => setProductFilter(prev => ({ ...prev, keyword: e.target.value }))}
                          className="pl-10 w-64"
                        />
                      </div>

                      {/* 筛选 */}
                      <Select 
                        value={productFilter.status} 
                        onValueChange={(value) => setProductFilter(prev => ({ ...prev, status: value }))}
                      >
                        <SelectTrigger className="w-32">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="all">全部状态</SelectItem>
                          <SelectItem value="active">上架中</SelectItem>
                          <SelectItem value="inactive">已下架</SelectItem>
                          <SelectItem value="draft">草稿</SelectItem>
                          <SelectItem value="sold_out">已售罄</SelectItem>
                        </SelectContent>
                      </Select>

                      <Select 
                        value={productFilter.type} 
                        onValueChange={(value) => setProductFilter(prev => ({ ...prev, type: value }))}
                      >
                        <SelectTrigger className="w-32">
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

                    <Button
                      onClick={() => {
                        setEditingProduct(null)
                        setShowProductDialog(true)
                      }}
                      className="bg-blue-600 hover:bg-blue-700"
                    >
                      <Plus className="w-4 h-4 mr-2" />
                      添加商品
                    </Button>
                  </div>
                </CardContent>
              </Card>

              {/* 商品列表 */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {products.map((product) => (
                  <Card key={product.id} className="hover:shadow-lg transition-shadow">
                    <CardHeader>
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-2">
                          {getTypeIcon(product.product_type)}
                          <CardTitle className="text-lg">{product.name}</CardTitle>
                        </div>
                        <Badge className={getStatusColor(product.status)}>
                          {getStatusText(product.status)}
                        </Badge>
                      </div>
                      <CardDescription className="line-clamp-2">
                        {product.short_desc || product.description}
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                          <span className="text-gray-500">积分价格:</span>
                          <div className="font-semibold">{product.credit_price.toLocaleString()}</div>
                        </div>
                        <div>
                          <span className="text-gray-500">库存:</span>
                          <div className="font-semibold">{product.stock}</div>
                        </div>
                        <div>
                          <span className="text-gray-500">已兑换:</span>
                          <div className="font-semibold">{product.redeem_count}</div>
                        </div>
                        <div>
                          <span className="text-gray-500">分类:</span>
                          <div className="font-semibold">{product.category || '-'}</div>
                        </div>
                      </div>

                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => {
                            setEditingProduct(product)
                            setShowProductDialog(true)
                          }}
                        >
                          <Edit className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => deleteProduct(product.id)}
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>

              {products.length === 0 && !isLoading && (
                <Card>
                  <CardContent className="py-12 text-center">
                    <Package className="w-16 h-16 text-gray-300 mx-auto mb-4" />
                    <p className="text-gray-500">暂无商品</p>
                  </CardContent>
                </Card>
              )}
            </div>
          </TabsContent>

          {/* 订单管理 */}
          <TabsContent value="orders">
            <div className="space-y-6">
              {/* 筛选栏 */}
              <Card>
                <CardContent className="p-4">
                  <div className="flex items-center gap-4">
                    <Select 
                      value={redemptionFilter.status} 
                      onValueChange={(value) => setRedemptionFilter(prev => ({ ...prev, status: value }))}
                    >
                      <SelectTrigger className="w-40">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">全部状态</SelectItem>
                        <SelectItem value="pending">待处理</SelectItem>
                        <SelectItem value="confirmed">已确认</SelectItem>
                        <SelectItem value="processing">处理中</SelectItem>
                        <SelectItem value="shipped">已发货</SelectItem>
                        <SelectItem value="delivered">已送达</SelectItem>
                        <SelectItem value="completed">已完成</SelectItem>
                        <SelectItem value="cancelled">已取消</SelectItem>
                        <SelectItem value="refunded">已退款</SelectItem>
                      </SelectContent>
                    </Select>

                    <Input
                      placeholder="用户ID"
                      value={redemptionFilter.user_id}
                      onChange={(e) => setRedemptionFilter(prev => ({ ...prev, user_id: e.target.value }))}
                      className="w-40"
                    />

                    <Select 
                      value={redemptionFilter.product_type} 
                      onValueChange={(value) => setRedemptionFilter(prev => ({ ...prev, product_type: value }))}
                    >
                      <SelectTrigger className="w-40">
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

                    <Button variant="outline" onClick={fetchRedemptions}>
                      <Search className="w-4 h-4 mr-2" />
                      搜索
                    </Button>
                  </div>
                </CardContent>
              </Card>

              {/* 订单列表 */}
              <div className="space-y-4">
                {redemptions.map((redemption) => (
                  <Card key={redemption.id}>
                    <CardContent className="p-6">
                      <div className="flex items-start justify-between mb-4">
                        <div>
                          <h3 className="font-semibold text-lg">
                            订单 {redemption.redemption_no}
                          </h3>
                          <p className="text-sm text-gray-500">
                            {new Date(redemption.created_at).toLocaleString()}
                          </p>
                          <p className="text-sm text-gray-600 mt-1">
                            用户: {redemption.user?.username || redemption.user_id}
                          </p>
                        </div>
                        <Badge className={getStatusColor(redemption.status)}>
                          {getStatusText(redemption.status)}
                        </Badge>
                      </div>

                      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                        <div>
                          <span className="text-sm text-gray-500">商品信息</span>
                          <div className="font-medium">
                            {redemption.product?.name || 'Unknown Product'}
                          </div>
                          <div className="text-sm text-gray-600">
                            数量: {redemption.quantity}
                          </div>
                        </div>
                        <div>
                          <span className="text-sm text-gray-500">积分消耗</span>
                          <div className="font-medium">
                            {redemption.total_credits.toLocaleString()} 积分
                          </div>
                          <div className="text-sm text-gray-600">
                            单价: {redemption.credit_price.toLocaleString()} 积分
                          </div>
                        </div>
                        <div>
                          <span className="text-sm text-gray-500">兑换码</span>
                          <div className="font-medium font-mono">
                            {redemption.redemption_code || '-'}
                          </div>
                          {redemption.tracking_number && (
                            <div className="text-sm text-gray-600">
                              物流: {redemption.tracking_number}
                            </div>
                          )}
                        </div>
                      </div>

                      <div className="flex items-center justify-between border-t pt-4">
                        <div className="flex items-center gap-2 text-sm text-gray-500">
                          <Calendar className="w-4 h-4" />
                          创建时间: {new Date(redemption.created_at).toLocaleString()}
                        </div>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => {
                            setSelectedRedemption(redemption)
                            setShowRedemptionDialog(true)
                          }}
                        >
                          <Eye className="w-4 h-4 mr-2" />
                          查看详情
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>

              {redemptions.length === 0 && !isLoading && (
                <Card>
                  <CardContent className="py-12 text-center">
                    <ShoppingCart className="w-16 h-16 text-gray-300 mx-auto mb-4" />
                    <p className="text-gray-500">暂无兑换订单</p>
                  </CardContent>
                </Card>
              )}
            </div>
          </TabsContent>

          {/* 配置管理 */}
          <TabsContent value="settings">
            <Card>
              <CardHeader>
                <CardTitle>积分商城配置</CardTitle>
                <CardDescription>管理积分商城的系统配置</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-6">
                  <div className="text-center py-12 text-gray-500">
                    配置管理功能开发中...
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        {error && (
          <Alert variant="destructive" className="mt-4">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
      </div>

      {/* 商品编辑对话框 */}
      <Dialog open={showProductDialog} onOpenChange={setShowProductDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {editingProduct ? '编辑商品' : '添加商品'}
            </DialogTitle>
            <DialogDescription>
              填写商品信息并设置积分价格
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="text-center py-8 text-gray-500">
              商品编辑表单待实现...
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* 订单详情对话框 */}
      <Dialog open={showRedemptionDialog} onOpenChange={setShowRedemptionDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              订单详情 - {selectedRedemption?.redemption_no}
            </DialogTitle>
            <DialogDescription>
              查看和管理兑换订单状态
            </DialogDescription>
          </DialogHeader>
          {selectedRedemption && (
            <div className="space-y-4">
              <div className="text-center py-8 text-gray-500">
                订单详情界面待实现...
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}