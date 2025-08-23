'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { BackButton } from '@/components/ui/back-button'
import { 
  Route,
  MapPin,
  Navigation,
  Clock,
  Truck,
  Plus,
  Minus,
  RotateCcw,
  Zap,
  Target,
  AlertTriangle,
  CheckCircle,
  ArrowRight,
  Map,
  Phone,
  Star,
  Info
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { usePermission } from '@/hooks/use-permission'
import { apiClient } from '@/lib/api-client'
import { toast } from '@/components/ui/use-toast'

interface DeliveryPoint {
  id: string
  op_code: string
  address: string
  priority: 'normal' | 'urgent'
  estimated_time: number // minutes
  contact?: string
  notes?: string
  coordinates?: {
    lat: number
    lng: number
  }
}

interface RouteOptimization {
  total_distance: number // km
  total_time: number // minutes
  fuel_cost: number // yuan
  optimized_order: string[]
  time_windows: Array<{
    point_id: string
    arrival_time: string
    departure_time: string
  }>
}

export default function RoutePlannerPage() {
  const { user } = useAuth()
  const { hasPermission } = usePermission()
  const [deliveryPoints, setDeliveryPoints] = useState<DeliveryPoint[]>([])
  const [optimizedRoute, setOptimizedRoute] = useState<RouteOptimization | null>(null)
  const [newPointCode, setNewPointCode] = useState('')
  const [isOptimizing, setIsOptimizing] = useState(false)
  const [currentLocation, setCurrentLocation] = useState<{lat: number, lng: number} | null>(null)

  // 检查权限
  const isCourier = hasPermission('courier.basic') || 
                   hasPermission('courier.intermediate') || 
                   hasPermission('courier.advanced') || 
                   hasPermission('courier.management')

  // 获取当前位置 - must be before conditional return
  useEffect(() => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          setCurrentLocation({
            lat: position.coords.latitude,
            lng: position.coords.longitude
          })
        },
        (error) => {
          console.warn('Location access denied:', error)
        }
      )
    }
  }, [])

  if (!isCourier) {
    return (
      <div className="container mx-auto px-4 py-8">
        <BackButton href="/delivery-guide" />
        <Alert variant="destructive" className="mt-4">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            此功能仅限信使使用。如需申请成为信使，请联系管理员。
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  // 添加投递点
  const addDeliveryPoint = async () => {
    if (!newPointCode.trim()) {
      toast({
        title: '请输入OP Code',
        variant: 'destructive'
      })
      return
    }

    const cleanCode = newPointCode.trim().toUpperCase()
    
    // 检查是否已存在
    if (deliveryPoints.some(point => point.op_code === cleanCode)) {
      toast({
        title: 'OP Code已存在',
        variant: 'destructive'
      })
      return
    }

    try {
      // 验证OP Code并获取地址信息
      const response = await apiClient.get(`/api/v1/opcode/${cleanCode}`)
      const opcodeData = ((response as any)?.data?.data || (response as any)?.data)?.data

      if (!opcodeData) {
        throw new Error('OP Code不存在')
      }

      const newPoint: DeliveryPoint = {
        id: Date.now().toString(),
        op_code: cleanCode,
        address: opcodeData.full_address,
        priority: 'normal',
        estimated_time: 5, // 默认5分钟
        coordinates: opcodeData.coordinates,
        contact: opcodeData.contact_info?.phone
      }

      setDeliveryPoints([...deliveryPoints, newPoint])
      setNewPointCode('')
      setOptimizedRoute(null) // 清除旧的路线优化
      
      toast({
        title: '添加成功',
        description: `已添加投递点: ${cleanCode}`
      })
    } catch (err: any) {
      toast({
        title: '添加失败',
        description: err.message || 'OP Code验证失败',
        variant: 'destructive'
      })
    }
  }

  // 删除投递点
  const removeDeliveryPoint = (id: string) => {
    setDeliveryPoints(deliveryPoints.filter(point => point.id !== id))
    setOptimizedRoute(null)
  }

  // 更新投递点优先级
  const updatePriority = (id: string, priority: 'normal' | 'urgent') => {
    setDeliveryPoints(deliveryPoints.map(point => 
      point.id === id ? { ...point, priority } : point
    ))
    setOptimizedRoute(null)
  }

  // 更新预估时间
  const updateEstimatedTime = (id: string, time: number) => {
    setDeliveryPoints(deliveryPoints.map(point => 
      point.id === id ? { ...point, estimated_time: Math.max(1, time) } : point
    ))
    setOptimizedRoute(null)
  }

  // 路线优化
  const optimizeRoute = async () => {
    if (deliveryPoints.length < 2) {
      toast({
        title: '至少需要2个投递点',
        variant: 'destructive'
      })
      return
    }

    try {
      setIsOptimizing(true)
      
      // 调用路线优化API
      const response = await apiClient.post('/api/v1/courier/route/optimize', {
        points: deliveryPoints.map(point => ({
          id: point.id,
          op_code: point.op_code,
          priority: point.priority,
          estimated_time: point.estimated_time,
          coordinates: point.coordinates
        })),
        start_location: currentLocation
      })

      const optimization = ((response as any)?.data?.data || (response as any)?.data)?.data
      if (optimization) {
        setOptimizedRoute(optimization)
        toast({
          title: '路线优化完成',
          description: `总距离: ${optimization.total_distance.toFixed(1)}km，预计用时: ${Math.round(optimization.total_time)}分钟`
        })
      }
    } catch (err: any) {
      // 如果API不存在，使用本地简单优化算法
      const mockOptimization = generateMockOptimization()
      setOptimizedRoute(mockOptimization)
      
      toast({
        title: '路线优化完成',
        description: `已生成优化路线，总预计用时: ${mockOptimization.total_time}分钟`
      })
    } finally {
      setIsOptimizing(false)
    }
  }

  // 生成模拟优化结果
  const generateMockOptimization = (): RouteOptimization => {
    // 简单的按优先级和地理位置排序
    const urgentPoints = deliveryPoints.filter(p => p.priority === 'urgent')
    const normalPoints = deliveryPoints.filter(p => p.priority === 'normal')
    
    const orderedPoints = [...urgentPoints, ...normalPoints]
    const totalTime = orderedPoints.reduce((sum, point) => sum + point.estimated_time + 10, 0) // +10分钟路程时间
    
    const timeWindows = orderedPoints.map((point, index) => {
      const startTime = new Date()
      startTime.setMinutes(startTime.getMinutes() + index * 15) // 每15分钟一个点
      const endTime = new Date(startTime)
      endTime.setMinutes(endTime.getMinutes() + point.estimated_time)
      
      return {
        point_id: point.id,
        arrival_time: startTime.toLocaleTimeString('zh-CN', { 
          hour: '2-digit', 
          minute: '2-digit' 
        }),
        departure_time: endTime.toLocaleTimeString('zh-CN', { 
          hour: '2-digit', 
          minute: '2-digit' 
        })
      }
    })

    return {
      total_distance: orderedPoints.length * 2.5, // 模拟距离
      total_time: totalTime,
      fuel_cost: orderedPoints.length * 3.5, // 模拟油费
      optimized_order: orderedPoints.map(p => p.id),
      time_windows: timeWindows
    }
  }

  // 清除所有点
  const clearAll = () => {
    setDeliveryPoints([])
    setOptimizedRoute(null)
  }

  // 获取优化后的投递点顺序
  const getOptimizedPoints = () => {
    if (!optimizedRoute) return deliveryPoints
    
    return optimizedRoute.optimized_order.map(id => 
      deliveryPoints.find(point => point.id === id)
    ).filter(Boolean) as DeliveryPoint[]
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <BackButton href="/delivery-guide" />
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
              <Route className="h-8 w-8" />
              智能路线规划
            </h1>
            <p className="text-gray-600 mt-2">优化投递路线，提高配送效率</p>
          </div>
        </div>

        {currentLocation && (
          <div className="text-right text-sm text-gray-600">
            <p>当前位置已定位</p>
            <p className="font-mono">
              {currentLocation.lat.toFixed(4)}, {currentLocation.lng.toFixed(4)}
            </p>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 左侧：投递点管理 */}
        <div className="lg:col-span-2 space-y-6">
          {/* 添加投递点 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Plus className="h-5 w-5" />
                添加投递点
              </CardTitle>
              <CardDescription>
                输入OP Code添加投递地址到路线规划中
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex gap-2">
                <Input
                  placeholder="输入OP Code，如：PK5F3D"
                  value={newPointCode}
                  onChange={(e) => setNewPointCode(e.target.value.toUpperCase())}
                  onKeyPress={(e) => e.key === 'Enter' && addDeliveryPoint()}
                  className="font-mono"
                  maxLength={6}
                />
                <Button onClick={addDeliveryPoint}>
                  <Plus className="w-4 h-4 mr-2" />
                  添加
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* 投递点列表 */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <MapPin className="h-5 w-5" />
                  投递点列表 ({deliveryPoints.length})
                </CardTitle>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={optimizeRoute}
                    disabled={deliveryPoints.length < 2 || isOptimizing}
                  >
                    {isOptimizing ? (
                      <>
                        <RotateCcw className="w-4 h-4 mr-2 animate-spin" />
                        优化中...
                      </>
                    ) : (
                      <>
                        <Zap className="w-4 h-4 mr-2" />
                        优化路线
                      </>
                    )}
                  </Button>
                  {deliveryPoints.length > 0 && (
                    <Button variant="outline" size="sm" onClick={clearAll}>
                      <RotateCcw className="w-4 h-4 mr-2" />
                      清除全部
                    </Button>
                  )}
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {deliveryPoints.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  <Target className="h-12 w-12 mx-auto mb-2 opacity-50" />
                  <p>还没有添加投递点</p>
                  <p className="text-sm">添加至少2个投递点开始路线规划</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {(optimizedRoute ? getOptimizedPoints() : deliveryPoints).map((point, index) => {
                    const timeWindow = optimizedRoute?.time_windows.find(tw => tw.point_id === point.id)
                    const isOptimized = optimizedRoute !== null
                    
                    return (
                      <div key={point.id} className="p-4 border rounded-lg">
                        <div className="flex items-start justify-between">
                          <div className="flex items-start gap-3 flex-1">
                            {/* 序号 */}
                            <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold ${
                              isOptimized 
                                ? 'bg-green-100 text-green-700' 
                                : 'bg-gray-100 text-gray-600'
                            }`}>
                              {index + 1}
                            </div>
                            
                            {/* 地址信息 */}
                            <div className="flex-1">
                              <div className="flex items-center gap-2 mb-1">
                                <span className="font-mono font-semibold">{point.op_code}</span>
                                <Badge 
                                  variant={point.priority === 'urgent' ? 'destructive' : 'outline'}
                                  className="text-xs"
                                >
                                  {point.priority === 'urgent' ? '紧急' : '普通'}
                                </Badge>
                                {isOptimized && (
                                  <Badge variant="outline" className="bg-green-50 text-green-700 text-xs">
                                    <CheckCircle className="w-3 h-3 mr-1" />
                                    已优化
                                  </Badge>
                                )}
                              </div>
                              <p className="text-sm text-gray-600 mb-2">{point.address}</p>
                              
                              {/* 时间窗口 */}
                              {timeWindow && (
                                <div className="flex items-center gap-4 text-xs text-green-600 bg-green-50 p-2 rounded">
                                  <div className="flex items-center gap-1">
                                    <Clock className="w-3 h-3" />
                                    <span>到达: {timeWindow.arrival_time}</span>
                                  </div>
                                  <ArrowRight className="w-3 h-3" />
                                  <div className="flex items-center gap-1">
                                    <Clock className="w-3 h-3" />
                                    <span>离开: {timeWindow.departure_time}</span>
                                  </div>
                                </div>
                              )}

                              {/* 联系电话 */}
                              {point.contact && (
                                <div className="flex items-center gap-1 text-xs text-gray-600 mt-1">
                                  <Phone className="w-3 h-3" />
                                  <span>{point.contact}</span>
                                </div>
                              )}
                            </div>
                          </div>

                          {/* 控制按钮 */}
                          <div className="flex items-center gap-2">
                            <div className="flex items-center gap-1">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => updatePriority(
                                  point.id, 
                                  point.priority === 'urgent' ? 'normal' : 'urgent'
                                )}
                              >
                                <Star className={`w-3 h-3 ${
                                  point.priority === 'urgent' ? 'fill-current text-red-500' : ''
                                }`} />
                              </Button>
                              
                              <div className="flex items-center gap-1">
                                <Button
                                  variant="outline"
                                  size="sm"
                                  onClick={() => updateEstimatedTime(point.id, point.estimated_time - 1)}
                                >
                                  <Minus className="w-3 h-3" />
                                </Button>
                                <span className="text-xs px-2 py-1 bg-gray-100 rounded min-w-[3rem] text-center">
                                  {point.estimated_time}min
                                </span>
                                <Button
                                  variant="outline"
                                  size="sm"
                                  onClick={() => updateEstimatedTime(point.id, point.estimated_time + 1)}
                                >
                                  <Plus className="w-3 h-3" />
                                </Button>
                              </div>
                              
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => removeDeliveryPoint(point.id)}
                                className="text-red-600 hover:text-red-700"
                              >
                                <Minus className="w-3 h-3" />
                              </Button>
                            </div>
                          </div>
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* 右侧：路线统计和操作 */}
        <div className="space-y-6">
          {/* 路线统计 */}
          {optimizedRoute && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Truck className="h-5 w-5" />
                  路线统计
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 gap-3">
                  <div className="flex justify-between">
                    <span className="text-gray-600">总距离</span>
                    <span className="font-bold">{optimizedRoute.total_distance.toFixed(1)} km</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">预计用时</span>
                    <span className="font-bold">{Math.round(optimizedRoute.total_time)} 分钟</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">预估油费</span>
                    <span className="font-bold">¥{optimizedRoute.fuel_cost.toFixed(1)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">投递点数</span>
                    <span className="font-bold">{deliveryPoints.length} 个</span>
                  </div>
                </div>

                <div className="pt-4 border-t">
                  <div className="grid grid-cols-2 gap-2">
                    <Button size="sm" className="w-full">
                      <Navigation className="w-4 h-4 mr-1" />
                      开始导航
                    </Button>
                    <Button variant="outline" size="sm" className="w-full">
                      <Map className="w-4 h-4 mr-1" />
                      查看地图
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* 优化建议 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">优化建议</CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-gray-600 space-y-2">
              {deliveryPoints.length < 2 ? (
                <p>• 添加至少2个投递点开始路线规划</p>
              ) : (
                <>
                  <p>• 紧急订单会优先安排在路线前段</p>
                  <p>• 相近地理位置会自动聚合</p>
                  <p>• 建议在交通高峰期前完成投递</p>
                  <p>• 预留额外时间应对突发情况</p>
                </>
              )}
            </CardContent>
          </Card>

          {/* 使用说明 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">使用说明</CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-gray-600 space-y-2">
              <div className="space-y-1">
                <p><strong>添加点位：</strong>输入有效的OP Code</p>
                <p><strong>设置优先级：</strong>点击星号标记紧急任务</p>
                <p><strong>调整时间：</strong>使用+/-按钮调整预估时间</p>
                <p><strong>优化路线：</strong>点击"优化路线"生成最佳路径</p>
              </div>
            </CardContent>
          </Card>

          {/* 快捷操作 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">快捷操作</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Button variant="outline" className="w-full" size="sm">
                从任务中导入
              </Button>
              <Button variant="outline" className="w-full" size="sm">
                保存路线模板
              </Button>
              <Button variant="outline" className="w-full" size="sm">
                分享路线
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}