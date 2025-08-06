'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { useCartStore } from '@/stores/cart-store'
import { toast } from 'sonner'
import { shopApi, type Product } from '@/lib/api/shop'
import { 
  Mail,
  Heart, 
  Star,
  ShoppingCart,
  Filter,
  Search,
  Package,
  Truck,
  Shield,
  Sparkles,
  Palette,
  Scissors,
  Stamp,
  Gift,
  Award,
  Eye,
  Plus,
  Loader2
} from 'lucide-react'

export default function ShopPage() {
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [selectedPrice, setSelectedPrice] = useState('all')
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(false)
  const { addItem, itemCount } = useCartStore()

  const categories = [
    { id: 'all', label: '全部商品', icon: Package },
    { id: 'envelope', label: '信封', icon: Mail },
    { id: 'paper', label: '信纸', icon: Scissors },
    { id: 'stamp', label: '邮票', icon: Stamp },
    { id: 'gift', label: '礼品套装', icon: Gift },
  ]

  const priceRanges = [
    { id: 'all', label: '全部价格' },
    { id: 'low', label: '0-50元' },
    { id: 'mid', label: '50-150元' },
    { id: 'high', label: '150元以上' },
  ]

  // Fetch products on mount and when filters change
  useEffect(() => {
    fetchProducts()
  }, [selectedCategory, selectedPrice, page])

  const fetchProducts = async () => {
    setLoading(true)
    try {
      const params: any = {
        page,
        limit: 12,
        in_stock_only: true
      }

      // Apply category filter
      if (selectedCategory !== 'all') {
        params.category = selectedCategory
      }

      // Apply price filter
      if (selectedPrice === 'low') {
        params.max_price = 50
      } else if (selectedPrice === 'mid') {
        params.min_price = 50
        params.max_price = 150
      } else if (selectedPrice === 'high') {
        params.min_price = 150
      }

      const response = await shopApi.getProducts(params)
      
      if (page === 1) {
        setProducts(response.items)
      } else {
        setProducts(prev => [...prev, ...response.items])
      }
      setHasMore(response.has_next)
    } catch (error: any) {
      console.error('Failed to fetch products:', error)
      // Use mock data as fallback
      setProducts(mockProducts)
    } finally {
      setLoading(false)
    }
  }

  const handleAddToCart = async (product: Product) => {
    try {
      // Try to add to cart via API first
      await shopApi.addToCart(product.id, 1)
      
      // Then update local store
      addItem({
        id: parseInt(product.id),
        name: product.name,
        description: product.description,
        price: product.price,
        originalPrice: product.original_price || product.price,
        image: product.image_url || '/api/placeholder/300/300',
        category: product.category,
        tags: product.tags || []
      }, 1)
      
      toast.success(`${product.name} 已添加到购物车`)
    } catch (error) {
      // Fallback to local cart only
      addItem({
        id: parseInt(product.id),
        name: product.name,
        description: product.description,
        price: product.price,
        originalPrice: product.original_price || product.price,
        image: product.image_url || '/api/placeholder/300/300',
        category: product.category,
        tags: product.tags || []
      }, 1)
      
      toast.success(`${product.name} 已添加到购物车`)
    }
  }

  // Mock products as fallback
  const mockProducts: Product[] = [
    {
      id: "1",
      name: "复古牛皮纸信封套装",
      description: "优质牛皮纸制作，复古文艺风格，适合各种主题书信",
      price: 28,
      original_price: 35,
      image_url: "/api/placeholder/300/300",
      thumbnail_url: "/api/placeholder/300/300",
      category: "envelope",
      product_type: "envelope",
      stock: 100,
      rating: 4.8,
      review_count: 234,
      sold: 1456,
      tags: ["热销", "复古"],
      is_featured: true,
      is_active: true,
      discount: 20,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    } as Product,
    {
      id: "2",
      name: "樱花主题信纸礼盒",
      description: "精美樱花图案信纸，配套信封，春日限定款",
      price: 68,
      original_price: 88,
      image_url: "/api/placeholder/300/300",
      thumbnail_url: "/api/placeholder/300/300",
      category: "paper",
      product_type: "paper",
      stock: 100,
      rating: 4.9,
      review_count: 189,
      sold: 892,
      tags: ["限定", "精美"],
      is_featured: true,
      is_active: true,
      discount: 23,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    } as Product,
    {
      id: "3",
      name: "手绘校园风景邮票",
      description: "手绘各大高校标志性建筑，收藏与使用兼备",
      price: 45,
      original_price: 45,
      image_url: "/api/placeholder/300/300",
      thumbnail_url: "/api/placeholder/300/300",
      category: "stamp",
      product_type: "stamp",
      stock: 100,
      rating: 4.7,
      review_count: 156,
      sold: 567,
      tags: ["手绘", "校园"],
      is_featured: false,
      is_active: true,
      discount: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    } as Product,
    {
      id: "4",
      name: "OpenPenPal定制礼品套装",
      description: "品牌定制信纸、信封、封蜡、钢笔，完整书信体验",
      price: 168,
      original_price: 220,
      image_url: "/api/placeholder/300/300",
      thumbnail_url: "/api/placeholder/300/300",
      category: "gift",
      product_type: "gift",
      stock: 100,
      rating: 5.0,
      review_count: 89,
      sold: 234,
      tags: ["定制", "高端"],
      is_featured: true,
      is_active: true,
      discount: 24,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    } as Product,
    {
      id: "5",
      name: "简约白色信封 (50枚装)",
      description: "经典白色信封，适合日常书信往来",
      price: 15,
      original_price: 15,
      image_url: "/api/placeholder/300/300",
      thumbnail_url: "/api/placeholder/300/300",
      category: "envelope",
      product_type: "envelope",
      stock: 100,
      rating: 4.5,
      review_count: 456,
      sold: 2890,
      tags: ["实用", "经典"],
      is_featured: false,
      is_active: true,
      discount: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    } as Product,
    {
      id: "6",
      name: "文艺青年专用信纸",
      description: "淡雅设计，有横线和方格两种规格可选",
      price: 38,
      original_price: 45,
      image_url: "/api/placeholder/300/300",
      thumbnail_url: "/api/placeholder/300/300",
      category: "paper",
      product_type: "paper",
      stock: 100,
      rating: 4.6,
      review_count: 267,
      sold: 1123,
      tags: ["文艺", "实用"],
      is_featured: false,
      is_active: true,
      discount: 16,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    } as Product,
    {
      id: "7",
      name: "节日主题邮票套装",
      description: "春节、中秋、国庆等传统节日主题邮票",
      price: 88,
      original_price: 110,
      image_url: "/api/placeholder/300/300",
      thumbnail_url: "/api/placeholder/300/300",
      category: "stamp",
      product_type: "stamp",
      stock: 100,
      rating: 4.8,
      review_count: 98,
      sold: 345,
      tags: ["节日", "套装"],
      is_featured: true,
      is_active: true,
      discount: 20,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    } as Product,
    {
      id: "8",
      name: "新用户入门套装",
      description: "为初次体验手写信的用户精心准备的入门套装",
      price: 58,
      original_price: 78,
      image_url: "/api/placeholder/300/300",
      thumbnail_url: "/api/placeholder/300/300",
      category: "gift",
      product_type: "gift",
      stock: 100,
      rating: 4.7,
      review_count: 234,
      sold: 789,
      tags: ["新人", "优惠"],
      is_featured: false,
      is_active: true,
      discount: 26,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    } as Product
  ]

  const filteredProducts = products.filter(product => {
    const matchesCategory = selectedCategory === 'all' || product.category === selectedCategory
    const matchesPrice = selectedPrice === 'all' || 
      (selectedPrice === 'low' && product.price <= 50) ||
      (selectedPrice === 'mid' && product.price > 50 && product.price <= 150) ||
      (selectedPrice === 'high' && product.price > 150)
    return matchesCategory && matchesPrice
  })


  const services = [
    {
      icon: Truck,
      title: "免费配送",
      description: "全国包邮，3-5天送达"
    },
    {
      icon: Shield,
      title: "品质保证",
      description: "7天无理由退换货"
    },
    {
      icon: Award,
      title: "正品保障",
      description: "100%正品，假一赔十"
    }
  ]

  return (
    <div className="min-h-screen flex flex-col bg-letter-paper">
      <Header />
      
      <main className="flex-1">
        {/* Cart Floating Button */}
        <Link 
          href="/cart"
          className="fixed top-20 right-4 z-50 bg-amber-600 text-white rounded-full p-3 shadow-lg hover:bg-amber-700 transition-colors"
        >
          <div className="relative">
            <ShoppingCart className="h-6 w-6" />
            {itemCount > 0 && (
              <span className="absolute -top-2 -right-2 bg-red-500 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center">
                {itemCount}
              </span>
            )}
          </div>
        </Link>

        {/* Hero Section */}
        <section className="py-16 bg-gradient-to-br from-amber-50 via-orange-50 to-yellow-50">
          <div className="container px-4">
            <div className="text-center max-w-3xl mx-auto">
              <div className="inline-block px-4 py-2 bg-amber-100 rounded-full text-amber-800 text-sm font-medium mb-6">
                🛍️ 精选文具商城
              </div>
              <h1 className="font-serif text-4xl md:text-5xl font-bold text-amber-900 mb-6">
                信封商城
              </h1>
              <p className="text-xl text-amber-700 mb-8 leading-relaxed">
                精选优质文具用品，为您的每一封信增添独特的温度与美感
              </p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <Button asChild size="lg" className="bg-amber-600 hover:bg-amber-700 text-white font-serif px-8">
                  <Link href="#products">
                    <ShoppingCart className="mr-2 h-5 w-5" />
                    开始购物
                  </Link>
                </Button>
                <Button asChild variant="outline" size="lg" className="border-amber-300 text-amber-700 hover:bg-amber-50 font-serif px-8">
                  <Link href="#featured">
                    <Sparkles className="mr-2 h-5 w-5" />
                    精选推荐
                  </Link>
                </Button>
              </div>
            </div>
          </div>
        </section>

        {/* Services */}
        <section className="py-12 bg-white border-b">
          <div className="container px-4">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
              {services.map((service, index) => {
                const Icon = service.icon
                return (
                  <div key={index} className="text-center">
                    <div className="w-16 h-16 bg-amber-100 rounded-full flex items-center justify-center mx-auto mb-4">
                      <Icon className="w-8 h-8 text-amber-600" />
                    </div>
                    <h3 className="font-semibold text-amber-900 mb-2">{service.title}</h3>
                    <p className="text-amber-700 text-sm">{service.description}</p>
                  </div>
                )
              })}
            </div>
          </div>
        </section>

        {/* Featured Products */}
        <section id="featured" className="py-16 bg-gradient-to-br from-amber-50 to-orange-50">
          <div className="container px-4">
            <div className="text-center mb-12">
              <h2 className="font-serif text-3xl font-bold text-amber-900 mb-4">
                精选推荐
              </h2>
              <p className="text-amber-700 max-w-2xl mx-auto">
                为您精心挑选的热门商品，品质与美感兼具
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
              {products.filter(p => p.is_featured).slice(0, 4).map((product) => (
                <Card key={product.id} className="group hover:shadow-lg transition-all duration-300 border-amber-200 hover:border-amber-400">
                  <div className="relative">
                    <div className="aspect-square bg-gradient-to-br from-amber-100 to-orange-100 rounded-t-lg flex items-center justify-center">
                      <Package className="w-16 h-16 text-amber-600" />
                    </div>
                    {product.discount > 0 && (
                      <div className="absolute top-2 left-2 bg-red-500 text-white text-xs px-2 py-1 rounded">
                        -{product.discount}%
                      </div>
                    )}
                    <div className="absolute top-2 right-2 flex gap-1">
                      {product.tags.map(tag => (
                        <span key={tag} className="bg-amber-100 text-amber-800 text-xs px-2 py-1 rounded">
                          {tag}
                        </span>
                      ))}
                    </div>
                  </div>
                  
                  <CardHeader className="pb-2">
                    <CardTitle className="font-serif text-lg text-amber-900 line-clamp-2">
                      {product.name}
                    </CardTitle>
                    <CardDescription className="text-amber-700 text-sm line-clamp-2">
                      {product.description}
                    </CardDescription>
                  </CardHeader>
                  
                  <CardContent>
                    <div className="flex items-center gap-1 mb-2">
                      {[...Array(5)].map((_, i) => (
                        <Star 
                          key={i} 
                          className={`w-4 h-4 ${i < Math.floor(product.rating) ? 'text-yellow-400 fill-current' : 'text-gray-300'}`} 
                        />
                      ))}
                      <span className="text-sm text-amber-600 ml-1">
                        {product.rating} ({product.review_count})
                      </span>
                    </div>
                    
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <span className="text-xl font-bold text-red-600">¥{product.price}</span>
                        {product.original_price > product.price && (
                          <span className="text-sm text-gray-500 line-through">¥{product.original_price}</span>
                        )}
                      </div>
                      <span className="text-xs text-amber-600">已售 {product.sold}</span>
                    </div>
                    
                    <Button 
                      className="w-full bg-amber-600 hover:bg-amber-700 text-white"
                      onClick={() => handleAddToCart(product)}
                    >
                      <Plus className="mr-2 h-4 w-4" />
                      加入购物车
                    </Button>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </section>

        {/* Filter Section */}
        <section id="products" className="py-8 bg-white border-b">
          <div className="container px-4">
            <div className="flex flex-col lg:flex-row gap-6 items-center justify-between">
              <h3 className="font-serif text-2xl font-bold text-amber-900">全部商品</h3>
              
              <div className="flex flex-wrap gap-4">
                <div className="flex flex-wrap gap-2">
                  {categories.map((category) => {
                    const Icon = category.icon
                    return (
                      <Button
                        key={category.id}
                        variant={selectedCategory === category.id ? "default" : "outline"}
                        size="sm"
                        onClick={() => setSelectedCategory(category.id)}
                        className={`${
                          selectedCategory === category.id 
                            ? 'bg-amber-600 text-white' 
                            : 'border-amber-300 text-amber-700 hover:bg-amber-50'
                        }`}
                      >
                        <Icon className="mr-1 h-4 w-4" />
                        {category.label}
                      </Button>
                    )
                  })}
                </div>
                
                <div className="flex items-center gap-2">
                  <span className="text-sm text-muted-foreground">价格：</span>
                  <select 
                    value={selectedPrice} 
                    onChange={(e) => setSelectedPrice(e.target.value)}
                    className="text-sm border border-amber-300 rounded-md px-3 py-1 bg-white text-amber-700"
                  >
                    {priceRanges.map((range) => (
                      <option key={range.id} value={range.id}>
                        {range.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* All Products */}
        <section className="py-12 bg-white">
          <div className="container px-4">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
              {filteredProducts.map((product) => (
                <Card key={product.id} className="group hover:shadow-lg transition-all duration-300 border-amber-200 hover:border-amber-400">
                  <div className="relative">
                    <div className="aspect-square bg-gradient-to-br from-amber-100 to-orange-100 rounded-t-lg flex items-center justify-center">
                      <Package className="w-16 h-16 text-amber-600" />
                    </div>
                    {product.discount > 0 && (
                      <div className="absolute top-2 left-2 bg-red-500 text-white text-xs px-2 py-1 rounded">
                        -{product.discount}%
                      </div>
                    )}
                    <div className="absolute top-2 right-2 flex flex-col gap-1">
                      {product.tags.map(tag => (
                        <span key={tag} className="bg-amber-100 text-amber-800 text-xs px-2 py-1 rounded">
                          {tag}
                        </span>
                      ))}
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="absolute bottom-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                      <Eye className="h-4 w-4" />
                    </Button>
                  </div>
                  
                  <CardHeader className="pb-2">
                    <CardTitle className="font-serif text-lg text-amber-900 line-clamp-2">
                      {product.name}
                    </CardTitle>
                    <CardDescription className="text-amber-700 text-sm line-clamp-2">
                      {product.description}
                    </CardDescription>
                  </CardHeader>
                  
                  <CardContent>
                    <div className="flex items-center gap-1 mb-2">
                      {[...Array(5)].map((_, i) => (
                        <Star 
                          key={i} 
                          className={`w-4 h-4 ${i < Math.floor(product.rating) ? 'text-yellow-400 fill-current' : 'text-gray-300'}`} 
                        />
                      ))}
                      <span className="text-sm text-amber-600 ml-1">
                        {product.rating} ({product.review_count})
                      </span>
                    </div>
                    
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <span className="text-xl font-bold text-red-600">¥{product.price}</span>
                        {product.original_price > product.price && (
                          <span className="text-sm text-gray-500 line-through">¥{product.original_price}</span>
                        )}
                      </div>
                      <span className="text-xs text-amber-600">已售 {product.sold}</span>
                    </div>
                    
                    <Button 
                      className="w-full bg-amber-600 hover:bg-amber-700 text-white"
                      onClick={() => handleAddToCart(product)}
                    >
                      <Plus className="mr-2 h-4 w-4" />
                      加入购物车
                    </Button>
                  </CardContent>
                </Card>
              ))}
            </div>

            {/* Load More */}
            <div className="text-center mt-12">
              <Button variant="outline" size="lg" className="border-amber-300 text-amber-700 hover:bg-amber-50">
                查看更多商品
              </Button>
            </div>
          </div>
        </section>

        {/* Shopping Guide */}
        <section className="py-16 bg-gradient-to-br from-amber-50 to-orange-50">
          <div className="container px-4">
            <div className="text-center max-w-3xl mx-auto">
              <h2 className="font-serif text-3xl font-bold text-amber-900 mb-6">
                购物指南
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mt-8">
                <div className="text-center">
                  <ShoppingCart className="w-12 h-12 text-amber-600 mx-auto mb-4" />
                  <h3 className="font-semibold text-amber-900 mb-2">选择商品</h3>
                  <p className="text-amber-700 text-sm">浏览商品，选择心仪的文具<br/>加入购物车</p>
                </div>
                <div className="text-center">
                  <Shield className="w-12 h-12 text-amber-600 mx-auto mb-4" />
                  <h3 className="font-semibold text-amber-900 mb-2">安全支付</h3>
                  <p className="text-amber-700 text-sm">支持多种支付方式<br/>交易安全有保障</p>
                </div>
                <div className="text-center">
                  <Truck className="w-12 h-12 text-amber-600 mx-auto mb-4" />
                  <h3 className="font-semibold text-amber-900 mb-2">快速配送</h3>
                  <p className="text-amber-700 text-sm">全国包邮配送<br/>3-5个工作日送达</p>
                </div>
              </div>
            </div>
          </div>
        </section>
      </main>

      <Footer />
    </div>
  )
}