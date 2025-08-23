'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { BackButton } from '@/components/ui/back-button'
import { 
  Building,
  MapPin,
  Navigation,
  Search,
  Map,
  Compass,
  ArrowUp,
  ArrowDown,
  ArrowLeft,
  ArrowRight,
  Clock,
  Phone,
  Info,
  Star,
  Users,
  Car,
  Wifi,
  Coffee,
  ShoppingCart,
  Book,
  Heart,
  AlertCircle,
  CheckCircle,
  Eye,
  Camera,
  Route,
  Target,
  Zap
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { apiClient } from '@/lib/api-client'
import { toast } from '@/components/ui/use-toast'

interface BuildingInfo {
  id: string
  name: string
  code: string
  op_code_prefix: string
  address: string
  floors: number
  rooms: Room[]
  facilities: Facility[]
  entrance_info: EntranceInfo[]
  floor_plan_url?: string
  operating_hours?: string
  contact?: string
}

interface Room {
  number: string
  name: string
  type: 'office' | 'classroom' | 'lab' | 'dormitory' | 'common'
  floor: number
  op_code: string
  coordinates?: {
    x: number
    y: number
  }
}

interface Facility {
  name: string
  icon: string
  floor: number
  description: string
  coordinates?: {
    x: number
    y: number
  }
}

interface EntranceInfo {
  name: string
  direction: 'north' | 'south' | 'east' | 'west'
  accessibility: boolean
  notes?: string
}

export default function BuildingNavigationPage() {
  const { user } = useAuth()
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedBuilding, setSelectedBuilding] = useState<BuildingInfo | null>(null)
  const [selectedFloor, setSelectedFloor] = useState(1)
  const [searchResults, setSearchResults] = useState<Room[]>([])
  const [loading, setLoading] = useState(false)
  const [currentLocation, setCurrentLocation] = useState<{lat: number, lng: number} | null>(null)

  // 获取当前位置
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

  // 常用建筑列表
  const popularBuildings = [
    {
      id: 'pk-5f',
      name: '北大第五教学楼',
      code: 'PK5F',
      op_code_prefix: 'PK5F',
      address: '北京大学燕园校区',
      floors: 5,
      type: 'academic',
      icon: Book
    },
    {
      id: 'pk-3d',
      name: '北大第三食堂',
      code: 'PK3D',
      op_code_prefix: 'PK3D',
      address: '北京大学燕园校区',
      floors: 3,
      type: 'dining',
      icon: Coffee
    },
    {
      id: 'qh-fx',
      name: '清华FIT楼',
      code: 'QHFX',
      op_code_prefix: 'QHFX',
      address: '清华大学紫荆校区',
      floors: 8,
      type: 'research',
      icon: Building
    },
    {
      id: 'bd-xy',
      name: '北交大学苑',
      code: 'BDXY',
      op_code_prefix: 'BDXY',
      address: '北京交通大学本部',
      floors: 12,
      type: 'dormitory',
      icon: Users
    }
  ]

  // 搜索房间
  const searchRooms = async (query: string, buildingCode?: string) => {
    if (!query.trim()) {
      setSearchResults([])
      return
    }

    try {
      setLoading(true)
      const params = new URLSearchParams({
        q: query.trim(),
        ...(buildingCode && { building: buildingCode })
      })
      
      const response = await apiClient.get(`/api/v1/buildings/search/rooms?${params}`)
      const results = (response as any)?.data?.data || (response as any)?.data || []
      setSearchResults(results)
      
      if (results.length === 0) {
        toast({
          title: '未找到房间',
          description: '请尝试使用其他关键词搜索',
          variant: 'destructive'
        })
      }
    } catch (err: any) {
      // 使用模拟数据作为备选方案
      const mockResults: Room[] = [
        {
          number: '303',
          name: '计算机实验室',
          type: 'lab' as const,
          floor: 3,
          op_code: 'PK5F3D',
          coordinates: { x: 50, y: 30 }
        },
        {
          number: '201',
          name: '多媒体教室',
          type: 'classroom' as const,
          floor: 2,
          op_code: 'PK5F2A',
          coordinates: { x: 20, y: 45 }
        }
      ].filter(room => 
        room.name.includes(query) || 
        room.number.includes(query) ||
        room.op_code.includes(query.toUpperCase())
      )
      
      setSearchResults(mockResults)
      
      toast({
        title: '使用本地搜索',
        description: `找到 ${mockResults.length} 个相关房间`
      })
    } finally {
      setLoading(false)
    }
  }

  // 获取建筑详情
  const loadBuildingDetails = async (buildingId: string) => {
    try {
      const response = await apiClient.get(`/api/v1/buildings/${buildingId}`)
      setSelectedBuilding((response as any)?.data?.data || (response as any)?.data)
    } catch (err: any) {
      // 使用模拟数据
      const mockBuilding: BuildingInfo = {
        id: buildingId,
        name: popularBuildings.find(b => b.id === buildingId)?.name || '未知建筑',
        code: popularBuildings.find(b => b.id === buildingId)?.code || 'UNKN',
        op_code_prefix: popularBuildings.find(b => b.id === buildingId)?.op_code_prefix || 'UNKN',
        address: popularBuildings.find(b => b.id === buildingId)?.address || '地址未知',
        floors: popularBuildings.find(b => b.id === buildingId)?.floors || 5,
        rooms: [
          { number: '101', name: '大厅', type: 'common', floor: 1, op_code: 'PK5F01' },
          { number: '201', name: '教室A', type: 'classroom', floor: 2, op_code: 'PK5F2A' },
          { number: '202', name: '教室B', type: 'classroom', floor: 2, op_code: 'PK5F2B' },
          { number: '301', name: '实验室', type: 'lab', floor: 3, op_code: 'PK5F3A' },
          { number: '303', name: '计算机房', type: 'lab', floor: 3, op_code: 'PK5F3D' }
        ],
        facilities: [
          { name: '电梯', icon: 'elevator', floor: 0, description: '主电梯' },
          { name: '楼梯', icon: 'stairs', floor: 0, description: '安全出口' },
          { name: 'WiFi', icon: 'wifi', floor: 0, description: '全楼覆盖' },
          { name: '停车场', icon: 'parking', floor: 0, description: '地下停车' }
        ],
        entrance_info: [
          { name: '正门', direction: 'south', accessibility: true, notes: '无障碍通道' },
          { name: '侧门', direction: 'east', accessibility: false, notes: '员工专用' }
        ],
        operating_hours: '06:00 - 22:00',
        contact: '010-62751234'
      }
      
      setSelectedBuilding(mockBuilding)
      toast({
        title: '加载建筑信息',
        description: '显示示例数据'
      })
    }
  }

  // 获取设施图标
  const getFacilityIcon = (iconName: string) => {
    const iconMap: Record<string, any> = {
      elevator: ArrowUp,
      stairs: ArrowUp,
      wifi: Wifi,
      parking: Car,
      coffee: Coffee,
      shop: ShoppingCart,
      restroom: Users
    }
    return iconMap[iconName] || Info
  }

  // 获取房间类型标签
  const getRoomTypeInfo = (type: string) => {
    const typeMap = {
      office: { label: '办公室', color: 'bg-blue-100 text-blue-800' },
      classroom: { label: '教室', color: 'bg-green-100 text-green-800' },
      lab: { label: '实验室', color: 'bg-purple-100 text-purple-800' },
      dormitory: { label: '宿舍', color: 'bg-orange-100 text-orange-800' },
      common: { label: '公共区域', color: 'bg-gray-100 text-gray-800' }
    }
    return typeMap[type as keyof typeof typeMap] || typeMap.common
  }

  // 处理搜索
  const handleSearch = () => {
    searchRooms(searchQuery, selectedBuilding?.code)
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <BackButton href="/delivery-guide" />
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
              <Building className="h-8 w-8" />
              建筑导航系统
            </h1>
            <p className="text-gray-600 mt-2">校园建筑内部导航，精确定位房间位置</p>
          </div>
        </div>
        
        {currentLocation && (
          <div className="text-right text-sm text-gray-600">
            <p>位置已定位</p>
            <p className="font-mono">
              {currentLocation.lat.toFixed(4)}, {currentLocation.lng.toFixed(4)}
            </p>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* 左侧：建筑选择和搜索 */}
        <div className="lg:col-span-1 space-y-6">
          {/* 建筑搜索 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">查找房间</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex gap-2">
                <Input
                  placeholder="房间号或名称"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                  className="text-sm"
                />
                <Button 
                  size="sm" 
                  onClick={handleSearch}
                  disabled={loading}
                >
                  <Search className="w-4 h-4" />
                </Button>
              </div>
              
              {searchResults.length > 0 && (
                <div className="space-y-2 max-h-48 overflow-y-auto">
                  {searchResults.map((room, index) => (
                    <div key={index} className="p-2 border rounded text-xs hover:bg-gray-50">
                      <div className="flex items-center justify-between mb-1">
                        <span className="font-mono">{room.op_code}</span>
                        <Badge variant="outline" className={`text-xs ${getRoomTypeInfo(room.type).color}`}>
                          {getRoomTypeInfo(room.type).label}
                        </Badge>
                      </div>
                      <p className="font-medium">{room.number} - {room.name}</p>
                      <p className="text-gray-500">{room.floor}楼</p>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {/* 常用建筑 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">常用建筑</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {popularBuildings.map((building) => (
                <button
                  key={building.id}
                  onClick={() => loadBuildingDetails(building.id)}
                  className={`w-full text-left p-3 border rounded hover:bg-gray-50 transition-colors ${
                    selectedBuilding?.id === building.id ? 'border-blue-200 bg-blue-50' : ''
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <div className="p-1 bg-blue-100 rounded">
                      <building.icon className="h-4 w-4 text-blue-600" />
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center justify-between">
                        <p className="font-medium text-sm">{building.name}</p>
                        <Badge variant="outline" className="text-xs">
                          {building.code}
                        </Badge>
                      </div>
                      <p className="text-xs text-gray-600">{building.floors}层楼</p>
                    </div>
                  </div>
                </button>
              ))}
            </CardContent>
          </Card>

          {/* 导航说明 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">使用说明</CardTitle>
            </CardHeader>
            <CardContent className="text-xs text-gray-600 space-y-2">
              <p>• 选择建筑查看详细平面图</p>
              <p>• 搜索房间号或名称快速定位</p>
              <p>• 点击房间获取导航指引</p>
              <p>• 查看设施分布和入口信息</p>
            </CardContent>
          </Card>
        </div>

        {/* 右侧：建筑详情和导航 */}
        <div className="lg:col-span-3 space-y-6">
          {!selectedBuilding ? (
            <Card>
              <CardContent className="p-12 text-center">
                <Building className="h-16 w-16 mx-auto mb-4 text-gray-400" />
                <h3 className="text-lg font-medium mb-2">选择建筑开始导航</h3>
                <p className="text-gray-600">从左侧列表选择建筑，或搜索特定房间</p>
              </CardContent>
            </Card>
          ) : (
            <>
              {/* 建筑信息 */}
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="flex items-center gap-2">
                      <Building className="h-5 w-5" />
                      {selectedBuilding.name}
                    </CardTitle>
                    <Badge variant="outline" className="font-mono">
                      {selectedBuilding.op_code_prefix}**
                    </Badge>
                  </div>
                  <CardDescription>{selectedBuilding.address}</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="flex items-center gap-2">
                      <ArrowUp className="h-4 w-4 text-gray-600" />
                      <span className="text-sm">{selectedBuilding.floors} 层楼</span>
                    </div>
                    {selectedBuilding.operating_hours && (
                      <div className="flex items-center gap-2">
                        <Clock className="h-4 w-4 text-gray-600" />
                        <span className="text-sm">{selectedBuilding.operating_hours}</span>
                      </div>
                    )}
                    {selectedBuilding.contact && (
                      <div className="flex items-center gap-2">
                        <Phone className="h-4 w-4 text-gray-600" />
                        <span className="text-sm">{selectedBuilding.contact}</span>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>

              {/* 楼层导航 */}
              <Tabs value={selectedFloor.toString()} onValueChange={(value) => setSelectedFloor(parseInt(value))} className="w-full">
                <div className="flex items-center justify-between mb-4">
                  <TabsList className="grid grid-cols-auto gap-1" style={{ gridTemplateColumns: `repeat(${selectedBuilding.floors}, 1fr)` }}>
                    {Array.from({ length: selectedBuilding.floors }, (_, i) => i + 1).map((floor) => (
                      <TabsTrigger key={floor} value={floor.toString()} className="px-3">
                        {floor}F
                      </TabsTrigger>
                    ))}
                  </TabsList>
                  
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm">
                      <Navigation className="w-4 h-4 mr-2" />
                      开始导航
                    </Button>
                    <Button variant="outline" size="sm">
                      <Map className="w-4 h-4 mr-2" />
                      查看地图
                    </Button>
                  </div>
                </div>

                {Array.from({ length: selectedBuilding.floors }, (_, i) => i + 1).map((floor) => (
                  <TabsContent key={floor} value={floor.toString()}>
                    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                      {/* 楼层平面图 */}
                      <div className="lg:col-span-2">
                        <Card>
                          <CardHeader>
                            <CardTitle className="text-lg">
                              {floor}楼 平面图
                            </CardTitle>
                          </CardHeader>
                          <CardContent>
                            {selectedBuilding.floor_plan_url ? (
                              <div className="aspect-video bg-gray-100 rounded-lg flex items-center justify-center">
                                <img 
                                  src={selectedBuilding.floor_plan_url} 
                                  alt={`${floor}楼平面图`}
                                  className="max-w-full max-h-full object-contain"
                                />
                              </div>
                            ) : (
                              <div className="aspect-video bg-gray-100 rounded-lg flex items-center justify-center">
                                <div className="text-center">
                                  <Map className="h-12 w-12 mx-auto mb-2 text-gray-400" />
                                  <p className="text-gray-600">平面图加载中...</p>
                                  <p className="text-xs text-gray-500 mt-1">
                                    显示 {floor} 楼房间分布
                                  </p>
                                </div>
                              </div>
                            )}
                          </CardContent>
                        </Card>
                      </div>

                      {/* 楼层房间列表 */}
                      <div className="space-y-6">
                        <Card>
                          <CardHeader>
                            <CardTitle className="text-sm">房间列表</CardTitle>
                          </CardHeader>
                          <CardContent className="space-y-2">
                            {selectedBuilding.rooms
                              .filter(room => room.floor === floor)
                              .map((room, index) => {
                                const typeInfo = getRoomTypeInfo(room.type)
                                return (
                                  <div key={index} className="p-3 border rounded hover:bg-gray-50">
                                    <div className="flex items-center justify-between mb-2">
                                      <span className="font-mono text-sm">{room.op_code}</span>
                                      <Badge variant="outline" className={`text-xs ${typeInfo.color}`}>
                                        {typeInfo.label}
                                      </Badge>
                                    </div>
                                    <div className="flex justify-between items-center">
                                      <div>
                                        <p className="font-medium">{room.number}</p>
                                        <p className="text-sm text-gray-600">{room.name}</p>
                                      </div>
                                      <Button variant="outline" size="sm">
                                        <Navigation className="w-3 h-3 mr-1" />
                                        导航
                                      </Button>
                                    </div>
                                  </div>
                                )
                              })}
                            {selectedBuilding.rooms.filter(room => room.floor === floor).length === 0 && (
                              <div className="text-center py-4 text-gray-500">
                                <p>该楼层暂无房间信息</p>
                              </div>
                            )}
                          </CardContent>
                        </Card>

                        {/* 楼层设施 */}
                        <Card>
                          <CardHeader>
                            <CardTitle className="text-sm">楼层设施</CardTitle>
                          </CardHeader>
                          <CardContent className="space-y-2">
                            {selectedBuilding.facilities
                              .filter(facility => facility.floor === floor || facility.floor === 0)
                              .map((facility, index) => {
                                const IconComponent = getFacilityIcon(facility.icon)
                                return (
                                  <div key={index} className="flex items-center gap-3">
                                    <div className="p-1 bg-blue-100 rounded">
                                      <IconComponent className="h-4 w-4 text-blue-600" />
                                    </div>
                                    <div>
                                      <p className="font-medium text-sm">{facility.name}</p>
                                      <p className="text-xs text-gray-600">{facility.description}</p>
                                    </div>
                                  </div>
                                )
                              })}
                          </CardContent>
                        </Card>

                        {/* 楼层入口（仅1楼显示） */}
                        {floor === 1 && (
                          <Card>
                            <CardHeader>
                              <CardTitle className="text-sm">出入口信息</CardTitle>
                            </CardHeader>
                            <CardContent className="space-y-2">
                              {selectedBuilding.entrance_info.map((entrance, index) => (
                                <div key={index} className="flex items-center justify-between">
                                  <div className="flex items-center gap-2">
                                    <Compass className="h-4 w-4 text-gray-600" />
                                    <span className="text-sm font-medium">{entrance.name}</span>
                                  </div>
                                  <div className="flex items-center gap-2">
                                    {entrance.accessibility && (
                                      <Badge variant="outline" className="text-xs bg-green-100 text-green-800">
                                        无障碍
                                      </Badge>
                                    )}
                                    <span className="text-xs text-gray-600 capitalize">
                                      {entrance.direction}
                                    </span>
                                  </div>
                                </div>
                              ))}
                            </CardContent>
                          </Card>
                        )}
                      </div>
                    </div>
                  </TabsContent>
                ))}
              </Tabs>
            </>
          )}
        </div>
      </div>
    </div>
  )
}