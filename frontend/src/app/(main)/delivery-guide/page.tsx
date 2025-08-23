'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  MapPin,
  Navigation,
  Search,
  Route,
  Clock,
  Building,
  Users,
  BookOpen,
  Star,
  Target,
  Compass,
  Phone,
  Map,
  Truck,
  CheckCircle,
  AlertCircle,
  Info,
  Lightbulb,
  Play,
  ArrowRight,
  MessageSquare
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { usePermission } from '@/hooks/use-permission'
import Link from 'next/link'

export default function DeliveryGuidePage() {
  const { user } = useAuth()
  const { hasPermission, getRoleDisplayName } = usePermission()
  const [currentLocation, setCurrentLocation] = useState<{lat: number, lng: number} | null>(null)
  const [locationPermission, setLocationPermission] = useState<'granted' | 'denied' | 'prompt'>('prompt')

  // 获取当前位置
  useEffect(() => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          setCurrentLocation({
            lat: position.coords.latitude,
            lng: position.coords.longitude
          })
          setLocationPermission('granted')
        },
        (error) => {
          console.warn('Location access denied:', error)
          setLocationPermission('denied')
        }
      )
    }
  }, [])

  // 检查是否是信使
  const isCourier = hasPermission('courier.basic') || 
                   hasPermission('courier.intermediate') || 
                   hasPermission('courier.advanced') || 
                   hasPermission('courier.management')

  const courierLevel = user?.role === 'courier_level1' ? 1 :
                      user?.role === 'courier_level2' ? 2 :
                      user?.role === 'courier_level3' ? 3 :
                      user?.role === 'courier_level4' ? 4 : 0

  // 快速导航卡片
  const quickActions = [
    {
      title: 'OP Code查询',
      description: '查询和验证OP Code地址',
      icon: Search,
      href: '/delivery-guide/opcode-search',
      color: 'bg-blue-50 text-blue-700 border-blue-200',
      available: true
    },
    {
      title: '路线规划',
      description: '优化多点投递路线',
      icon: Route,
      href: '/delivery-guide/route-planner',
      color: 'bg-green-50 text-green-700 border-green-200',
      available: isCourier
    },
    {
      title: '建筑导航',
      description: '校园建筑内部导航',
      icon: Building,
      href: '/delivery-guide/building-nav',
      color: 'bg-purple-50 text-purple-700 border-purple-200',
      available: true
    },
    {
      title: '信使培训',
      description: '投递技巧和最佳实践',
      icon: BookOpen,
      href: '/delivery-guide/training',
      color: 'bg-orange-50 text-orange-700 border-orange-200',
      available: true
    }
  ]

  // 投递统计数据（模拟）
  const deliveryStats = [
    {
      label: '今日投递',
      value: 12,
      icon: Truck,
      color: 'text-blue-600'
    },
    {
      label: '成功率',
      value: '95%',
      icon: CheckCircle,
      color: 'text-green-600'
    },
    {
      label: '平均用时',
      value: '8min',
      icon: Clock,
      color: 'text-purple-600'
    },
    {
      label: '信使等级',
      value: courierLevel ? `L${courierLevel}` : '用户',
      icon: Star,
      color: 'text-yellow-600'
    }
  ]

  // 实用技巧
  const deliveryTips = [
    {
      icon: MapPin,
      title: 'OP Code识别技巧',
      content: 'PK5F3D = 北大(PK) + 5号楼(5F) + 303室(3D)，快速记忆建筑编码规律'
    },
    {
      icon: Clock,
      title: '时间优化',
      content: '按地理位置聚合任务，避免往返，提高投递效率'
    },
    {
      icon: Phone,
      title: '沟通技巧',
      content: '投递前电话确认，避免空跑，提升用户满意度'
    },
    {
      icon: Target,
      title: '准确投递',
      content: '仔细核对OP Code，确认收件人身份，避免投递错误'
    }
  ]

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
            <Navigation className="h-8 w-8" />
            投递指导中心
          </h1>
          <p className="text-gray-600 mt-2">
            提供高效投递指导，优化信使工作流程
            {isCourier && ` • 当前等级: ${getRoleDisplayName()}`}
          </p>
        </div>
        
        {currentLocation && (
          <div className="text-right">
            <p className="text-sm text-gray-600">当前位置</p>
            <p className="text-xs font-mono">
              {currentLocation.lat.toFixed(6)}, {currentLocation.lng.toFixed(6)}
            </p>
          </div>
        )}
      </div>

      {/* 位置权限提示 */}
      {locationPermission === 'denied' && (
        <Alert className="mb-6">
          <Info className="h-4 w-4" />
          <AlertDescription>
            启用位置权限可获得更精确的导航和路线规划服务。
            <Button variant="link" className="p-0 ml-2" onClick={() => window.location.reload()}>
              重新获取位置
            </Button>
          </AlertDescription>
        </Alert>
      )}

      {/* 信使统计（仅信使可见） */}
      {isCourier && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
          {deliveryStats.map((stat, index) => (
            <Card key={index}>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-gray-600">{stat.label}</p>
                    <p className="text-2xl font-bold">{stat.value}</p>
                  </div>
                  <stat.icon className={`h-5 w-5 ${stat.color}`} />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* 快速操作 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {quickActions.map((action, index) => (
          <Link 
            key={index} 
            href={action.available ? action.href : '#'}
            className={action.available ? '' : 'pointer-events-none opacity-60'}
          >
            <Card className={`cursor-pointer hover:shadow-lg transition-all border ${action.color}`}>
              <CardContent className="p-6">
                <div className="flex flex-col items-center text-center space-y-4">
                  <div className={`p-3 rounded-full ${action.color.replace('border-', 'bg-').replace('text-', 'text-')}`}>
                    <action.icon className="h-6 w-6" />
                  </div>
                  <div>
                    <h3 className="font-semibold mb-1">{action.title}</h3>
                    <p className="text-sm text-gray-600">{action.description}</p>
                  </div>
                  {!action.available && (
                    <Badge variant="outline" className="text-xs">
                      需要信使权限
                    </Badge>
                  )}
                </div>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 主要内容区域 */}
        <div className="lg:col-span-2 space-y-6">
          {/* 功能导航标签 */}
          <Tabs defaultValue="overview" className="w-full">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="overview" className="flex items-center gap-2">
                <Map className="h-4 w-4" />
                概览
              </TabsTrigger>
              <TabsTrigger value="navigation" className="flex items-center gap-2">
                <Compass className="h-4 w-4" />
                导航工具
              </TabsTrigger>
              <TabsTrigger value="training" className="flex items-center gap-2">
                <BookOpen className="h-4 w-4" />
                学习中心
              </TabsTrigger>
            </TabsList>

            {/* 概览标签页 */}
            <TabsContent value="overview" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <MapPin className="h-5 w-5" />
                    OP Code地址系统
                  </CardTitle>
                  <CardDescription>
                    OpenPenPal使用6位OP Code进行精确地址定位
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="p-4 bg-blue-50 rounded-lg">
                      <h4 className="font-semibold mb-2">学校代码 (AA)</h4>
                      <p className="text-sm text-gray-600">PK=北京大学, QH=清华大学</p>
                      <p className="text-xs text-blue-600 mt-1">前两位标识学校</p>
                    </div>
                    <div className="p-4 bg-green-50 rounded-lg">
                      <h4 className="font-semibold mb-2">区域代码 (BB)</h4>
                      <p className="text-sm text-gray-600">5F=5号楼, 3D=第3食堂</p>
                      <p className="text-xs text-green-600 mt-1">中间两位标识建筑</p>
                    </div>
                    <div className="p-4 bg-purple-50 rounded-lg">
                      <h4 className="font-semibold mb-2">位置代码 (CC)</h4>
                      <p className="text-sm text-gray-600">3D=303室, 12=12号桌</p>
                      <p className="text-xs text-purple-600 mt-1">后两位精确定位</p>
                    </div>
                  </div>
                  
                  <div className="p-4 bg-yellow-50 rounded-lg border border-yellow-200">
                    <div className="flex items-start gap-2">
                      <Lightbulb className="h-5 w-5 text-yellow-600 mt-0.5" />
                      <div>
                        <h4 className="font-semibold text-yellow-900 mb-2">示例解析</h4>
                        <p className="text-sm text-yellow-800 mb-2">
                          <span className="font-mono bg-white px-2 py-1 rounded">PK5F3D</span> = 
                          北京大学 + 5号楼 + 303室
                        </p>
                        <p className="text-xs text-yellow-700">
                          掌握编码规律，快速定位投递地址
                        </p>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* 权限说明 */}
              {isCourier && (
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                      <Users className="h-5 w-5" />
                      信使权限范围
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-3">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 bg-yellow-100 rounded-full flex items-center justify-center">
                          <span className="text-sm font-bold text-yellow-600">L1</span>
                        </div>
                        <div>
                          <p className="font-medium">楼宇信使 - 同4位前缀</p>
                          <p className="text-sm text-gray-600">如 PK5F01 可投递 PK5F** 区域</p>
                        </div>
                      </div>
                      
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                          <span className="text-sm font-bold text-blue-600">L2</span>
                        </div>
                        <div>
                          <p className="font-medium">区域信使 - 同校区</p>
                          <p className="text-sm text-gray-600">可投递整个校区的任意地址</p>
                        </div>
                      </div>

                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
                          <span className="text-sm font-bold text-green-600">L3</span>
                        </div>
                        <div>
                          <p className="font-medium">学校信使 - 跨校区</p>
                          <p className="text-sm text-gray-600">可管理整个学校的投递任务</p>
                        </div>
                      </div>

                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 bg-purple-100 rounded-full flex items-center justify-center">
                          <span className="text-sm font-bold text-purple-600">L4</span>
                        </div>
                        <div>
                          <p className="font-medium">城市信使 - 无限制</p>
                          <p className="text-sm text-gray-600">可投递任意地址，管理下级信使</p>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}
            </TabsContent>

            {/* 导航工具标签页 */}
            <TabsContent value="navigation" className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Link href="/delivery-guide/opcode-search">
                  <Card className="cursor-pointer hover:shadow-md transition-shadow h-full">
                    <CardContent className="p-6">
                      <div className="flex items-start gap-4">
                        <div className="p-2 bg-blue-100 rounded-lg">
                          <Search className="h-6 w-6 text-blue-600" />
                        </div>
                        <div className="flex-1">
                          <h3 className="font-semibold mb-2">地址查询验证</h3>
                          <p className="text-sm text-gray-600 mb-3">
                            输入OP Code快速查询地址信息，验证投递权限
                          </p>
                          <div className="flex items-center gap-1 text-sm text-blue-600">
                            <Play className="w-3 h-3" />
                            立即使用
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </Link>

                <Link href="/delivery-guide/route-planner">
                  <Card className="cursor-pointer hover:shadow-md transition-shadow h-full">
                    <CardContent className="p-6">
                      <div className="flex items-start gap-4">
                        <div className="p-2 bg-green-100 rounded-lg">
                          <Route className="h-6 w-6 text-green-600" />
                        </div>
                        <div className="flex-1">
                          <h3 className="font-semibold mb-2">智能路线规划</h3>
                          <p className="text-sm text-gray-600 mb-3">
                            优化多点投递路线，提高配送效率
                          </p>
                          <div className="flex items-center gap-1 text-sm text-green-600">
                            <Play className="w-3 h-3" />
                            {isCourier ? '立即使用' : '需要信使权限'}
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </Link>

                <Link href="/delivery-guide/building-nav">
                  <Card className="cursor-pointer hover:shadow-md transition-shadow h-full">
                    <CardContent className="p-6">
                      <div className="flex items-start gap-4">
                        <div className="p-2 bg-purple-100 rounded-lg">
                          <Building className="h-6 w-6 text-purple-600" />
                        </div>
                        <div className="flex-1">
                          <h3 className="font-semibold mb-2">建筑内部导航</h3>
                          <p className="text-sm text-gray-600 mb-3">
                            校园建筑平面图，精确室内定位
                          </p>
                          <div className="flex items-center gap-1 text-sm text-purple-600">
                            <Play className="w-3 h-3" />
                            立即使用
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </Link>

                <Card className="opacity-60">
                  <CardContent className="p-6">
                    <div className="flex items-start gap-4">
                      <div className="p-2 bg-gray-100 rounded-lg">
                        <Compass className="h-6 w-6 text-gray-600" />
                      </div>
                      <div className="flex-1">
                        <h3 className="font-semibold mb-2">AR导航助手</h3>
                        <p className="text-sm text-gray-600 mb-3">
                          增强现实导航，即将上线
                        </p>
                        <Badge variant="outline" className="text-xs">
                          开发中
                        </Badge>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </TabsContent>

            {/* 学习中心标签页 */}
            <TabsContent value="training" className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Link href="/delivery-guide/training">
                  <Card className="cursor-pointer hover:shadow-md transition-shadow">
                    <CardContent className="p-6">
                      <div className="flex items-start gap-4">
                        <div className="p-2 bg-orange-100 rounded-lg">
                          <BookOpen className="h-6 w-6 text-orange-600" />
                        </div>
                        <div className="flex-1">
                          <h3 className="font-semibold mb-2">信使培训教程</h3>
                          <p className="text-sm text-gray-600 mb-3">
                            从入门到精通的完整培训体系
                          </p>
                          <div className="flex items-center gap-1 text-sm text-orange-600">
                            <ArrowRight className="w-3 h-3" />
                            开始学习
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </Link>

                <Card>
                  <CardContent className="p-6">
                    <div className="flex items-start gap-4">
                      <div className="p-2 bg-red-100 rounded-lg">
                        <AlertCircle className="h-6 w-6 text-red-600" />
                      </div>
                      <div className="flex-1">
                        <h3 className="font-semibold mb-2">安全规范</h3>
                        <p className="text-sm text-gray-600 mb-3">
                          投递安全注意事项和应急处理
                        </p>
                        <div className="flex items-center gap-1 text-sm text-red-600">
                          <ArrowRight className="w-3 h-3" />
                          查看规范
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </TabsContent>
          </Tabs>
        </div>

        {/* 右侧边栏 */}
        <div className="space-y-6">
          {/* 实用技巧 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Lightbulb className="h-5 w-5" />
                实用技巧
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {deliveryTips.map((tip, index) => (
                <div key={index} className="flex items-start gap-3">
                  <div className="p-1.5 bg-blue-100 rounded">
                    <tip.icon className="h-4 w-4 text-blue-600" />
                  </div>
                  <div className="flex-1">
                    <h4 className="font-medium text-sm mb-1">{tip.title}</h4>
                    <p className="text-xs text-gray-600">{tip.content}</p>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>

          {/* 快速联系 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Phone className="h-5 w-5" />
                需要帮助？
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="text-sm">
                <p className="font-medium mb-1">技术支持</p>
                <p className="text-gray-600">support@openpenpal.com</p>
              </div>
              <div className="text-sm">
                <p className="font-medium mb-1">紧急联系</p>
                <p className="text-gray-600">400-000-0000</p>
              </div>
              <Button variant="outline" className="w-full" size="sm">
                <MessageSquare className="w-4 h-4 mr-2" />
                在线客服
              </Button>
            </CardContent>
          </Card>

          {/* 系统状态 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">系统状态</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="flex justify-between items-center text-sm">
                <span>服务状态</span>
                <div className="flex items-center gap-1">
                  <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                  <span className="text-green-600">正常</span>
                </div>
              </div>
              <div className="flex justify-between items-center text-sm">
                <span>API响应</span>
                <span className="text-gray-600">&lt; 100ms</span>
              </div>
              <div className="flex justify-between items-center text-sm">
                <span>数据更新</span>
                <span className="text-gray-600">实时同步</span>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}