'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Store,
  Coins,
  Package,
  Star,
  ArrowRight,
  ShoppingCart,
  Award,
  TrendingUp
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { useRouter } from 'next/navigation'
import { getCreditShopProducts, getUserCreditBalance, type CreditShopProduct, type UserCredit } from '@/lib/api/credit-shop'

interface CreditShopCardProps {
  variant?: 'default' | 'compact' | 'featured'
  showBalance?: boolean
  className?: string
}

export function CreditShopCard({ 
  variant = 'default', 
  showBalance = true,
  className = '' 
}: CreditShopCardProps) {
  const { user } = useAuth()
  const router = useRouter()
  const [featuredProducts, setFeaturedProducts] = useState<CreditShopProduct[]>([])
  const [userCredit, setUserCredit] = useState<UserCredit | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    if (user) {
      loadData()
    }
  }, [user])

  const loadData = async () => {
    setIsLoading(true)
    try {
      // 获取推荐商品
      const productsResponse = await getCreditShopProducts({
        featured_only: true,
        limit: 3,
        in_stock_only: true
      })
      setFeaturedProducts((productsResponse as any).items || [])

      // 获取用户积分（如果需要显示）
      if (showBalance && user) {
        const creditResponse = await getUserCreditBalance()
        setUserCredit(creditResponse)
      }
    } catch (error) {
      console.error('Failed to load credit shop data:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleNavigateToShop = () => {
    router.push('/credit-shop')
  }

  const handleNavigateToProduct = (productId: string) => {
    router.push(`/credit-shop?product=${productId}`)
  }

  const formatCredits = (credits: number) => {
    return credits.toLocaleString()
  }

  if (variant === 'compact') {
    return (
      <Card className={`hover:shadow-lg transition-shadow cursor-pointer ${className}`}>
        <CardContent className="p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 rounded-lg">
                <Store className="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <h3 className="font-semibold text-gray-900">积分商城</h3>
                <p className="text-sm text-gray-500">用积分兑换好礼</p>
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleNavigateToShop}
              className="text-blue-600 hover:text-blue-700"
            >
              <ArrowRight className="w-4 h-4" />
            </Button>
          </div>
          
          {showBalance && userCredit && (
            <div className="mt-3 p-2 bg-gradient-to-r from-yellow-50 to-orange-50 rounded-lg">
              <div className="flex items-center gap-2">
                <Coins className="w-4 h-4 text-yellow-600" />
                <span className="text-sm text-gray-700">可用积分:</span>
                <span className="font-semibold text-gray-900">
                  {formatCredits(userCredit.available)}
                </span>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    )
  }

  if (variant === 'featured') {
    return (
      <Card className={`bg-gradient-to-br from-blue-50 to-purple-50 border-blue-200 ${className}`}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Store className="w-6 h-6 text-blue-600" />
              <CardTitle className="text-blue-900">积分商城</CardTitle>
            </div>
            <Badge className="bg-red-500 text-white">
              <Star className="w-3 h-3 mr-1" />
              热门
            </Badge>
          </div>
          <CardDescription className="text-blue-700">
            用积分兑换心仪商品，更多惊喜等你发现
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {showBalance && userCredit && (
            <div className="p-3 bg-white/70 rounded-lg border border-blue-200">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Coins className="w-5 h-5 text-yellow-600" />
                  <span className="text-sm text-gray-700">可用积分</span>
                </div>
                <span className="text-xl font-bold text-blue-900">
                  {formatCredits(userCredit.available)}
                </span>
              </div>
            </div>
          )}

          {featuredProducts.length > 0 && (
            <div className="space-y-2">
              <h4 className="text-sm font-medium text-blue-800">热门商品</h4>
              <div className="space-y-2">
                {featuredProducts.map((product) => (
                  <div
                    key={product.id}
                    className="flex items-center justify-between p-2 bg-white/50 rounded-lg hover:bg-white/70 transition-colors cursor-pointer"
                    onClick={() => handleNavigateToProduct(product.id)}
                  >
                    <div className="flex-1">
                      <div className="font-medium text-sm text-gray-900 truncate">
                        {product.name}
                      </div>
                      <div className="flex items-center gap-1 text-xs text-gray-600">
                        <Coins className="w-3 h-3 text-yellow-500" />
                        {formatCredits(product.credit_price)} 积分
                      </div>
                    </div>
                    <div className="text-xs text-gray-500">
                      库存 {product.stock}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          <Button
            onClick={handleNavigateToShop}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white"
          >
            <ShoppingCart className="w-4 h-4 mr-2" />
            立即逛逛
          </Button>
        </CardContent>
      </Card>
    )
  }

  // Default variant
  return (
    <Card className={`hover:shadow-lg transition-shadow ${className}`}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Store className="w-6 h-6 text-blue-600" />
            <CardTitle>积分商城</CardTitle>
          </div>
          {featuredProducts.length > 0 && (
            <Badge variant="secondary">
              {featuredProducts.length} 款热门
            </Badge>
          )}
        </div>
        <CardDescription>
          用积分兑换各种商品和服务
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* 用户积分余额 */}
        {showBalance && userCredit && (
          <div className="flex items-center justify-between p-3 bg-gradient-to-r from-yellow-50 to-orange-50 rounded-lg">
            <div className="flex items-center gap-2">
              <Coins className="w-5 h-5 text-yellow-600" />
              <span className="text-sm text-gray-700">我的积分</span>
            </div>
            <div className="text-right">
              <div className="text-lg font-bold text-gray-900">
                {formatCredits(userCredit.available)}
              </div>
              <div className="text-xs text-gray-500">
                总积分: {formatCredits(userCredit.total)}
              </div>
            </div>
          </div>
        )}

        {/* 推荐商品 */}
        {featuredProducts.length > 0 && (
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Star className="w-4 h-4 text-orange-500" />
              <span className="text-sm font-medium text-gray-700">推荐商品</span>
            </div>
            <div className="space-y-2">
              {featuredProducts.map((product) => (
                <div
                  key={product.id}
                  className="flex items-center justify-between p-3 border border-gray-200 rounded-lg hover:border-blue-300 transition-colors cursor-pointer"
                  onClick={() => handleNavigateToProduct(product.id)}
                >
                  <div className="flex-1">
                    <div className="font-medium text-gray-900 mb-1 truncate">
                      {product.name}
                    </div>
                    <div className="text-xs text-gray-500 truncate">
                      {product.short_desc || product.description}
                    </div>
                  </div>
                  <div className="text-right ml-3">
                    <div className="flex items-center gap-1 text-blue-600 font-medium">
                      <Coins className="w-4 h-4 text-yellow-500" />
                      {formatCredits(product.credit_price)}
                    </div>
                    <div className="text-xs text-gray-500">
                      库存 {product.stock}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* 快捷操作 */}
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleNavigateToShop}
            className="flex-1"
          >
            <Package className="w-4 h-4 mr-2" />
            浏览商品
          </Button>
          <Button
            variant="outline"
            onClick={() => router.push('/credit-shop?tab=orders')}
            className="flex-1"
          >
            <Award className="w-4 h-4 mr-2" />
            兑换记录
          </Button>
        </div>

        {/* 统计信息 */}
        {featuredProducts.length > 0 && (
          <div className="flex items-center justify-center gap-4 text-xs text-gray-500 pt-2 border-t">
            <div className="flex items-center gap-1">
              <TrendingUp className="w-3 h-3" />
              {featuredProducts.reduce((total, p) => total + p.redeem_count, 0)} 次兑换
            </div>
            <div className="flex items-center gap-1">
              <Package className="w-3 h-3" />
              {featuredProducts.length} 款推荐
            </div>
          </div>
        )}

        {isLoading && (
          <div className="text-center py-4">
            <div className="text-sm text-gray-500">加载中...</div>
          </div>
        )}

        {!isLoading && featuredProducts.length === 0 && (
          <div className="text-center py-4">
            <Package className="w-8 h-8 text-gray-300 mx-auto mb-2" />
            <div className="text-sm text-gray-500">暂无推荐商品</div>
            <Button
              variant="link"
              onClick={handleNavigateToShop}
              className="text-blue-600 p-0 h-auto"
            >
              查看所有商品
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default CreditShopCard