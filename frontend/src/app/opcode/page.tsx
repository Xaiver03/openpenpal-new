'use client'

import { useState, useEffect, useCallback } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { 
  MapPin, 
  Building, 
  Users, 
  Crown, 
  CheckCircle,
  AlertCircle,
  BookOpen,
  Settings,
  Search,
  School,
  ArrowRight,
  Home,
  Loader2,
  Package
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { apiClient } from '@/lib/api-client-enhanced'
import { useToast } from '@/hooks/use-toast'

// 类型定义
interface School {
  school_code: string
  school_name: string
  city: string
  province: string
  full_name?: string
}

interface District {
  code: string
  name: string
  description?: string
}

interface Building {
  code: string
  name: string
  type: string // dormitory, teaching, dining, etc
}

interface DeliveryPoint {
  code: string
  name: string
  available: boolean
  type: string
}

interface OPCodeSelection {
  school?: School
  district?: District
  building?: Building
  deliveryPoint?: DeliveryPoint
  finalCode?: string
}

export default function OPCodePage() {
  const { user, isAuthenticated } = useAuth()
  const { toast } = useToast()
  const [activeTab, setActiveTab] = useState('apply')
  const [loading, setLoading] = useState(false)
  
  // 层级选择状态
  const [selection, setSelection] = useState<OPCodeSelection>({})
  const [searchKeyword, setSearchKeyword] = useState('')
  const [schools, setSchools] = useState<School[]>([])
  const [districts, setDistricts] = useState<District[]>([])
  const [buildings, setBuildings] = useState<Building[]>([])
  const [deliveryPoints, setDeliveryPoints] = useState<DeliveryPoint[]>([])
  const [recommendedCodes, setRecommendedCodes] = useState<string[]>([])
  const [isSearching, setIsSearching] = useState(false)
  const [searchTimer, setSearchTimer] = useState<NodeJS.Timeout | null>(null)

  // 带防抖的搜索处理
  useEffect(() => {
    if (searchKeyword.trim().length > 0) {
      const timer = setTimeout(() => {
        searchSchoolsByKeyword(searchKeyword)
      }, 300) // 300ms 防抖延迟
      setSearchTimer(timer)
      
      return () => {
        clearTimeout(timer)
      }
    } else {
      setSchools([])
    }
  }, [searchKeyword])

  // 模糊搜索学校
  const searchSchoolsByKeyword = async (keyword: string) => {
    if (!keyword || keyword.trim().length === 0) {
      setSchools([])
      return
    }
    
    setIsSearching(true)
    try {
      // Use fetch directly to avoid any base URL issues
      const response = await fetch(`/api/schools/fuzzy-search?keyword=${encodeURIComponent(keyword)}`)
      const data = await response.json()
      
      if (data.code === 0) {
        const schoolData = data.data.schools || []
        // Transform data to match the expected format
        const transformedSchools = schoolData.map((school: any) => ({
          school_code: school.school_code,
          school_name: school.school_name,
          city: school.city,
          province: school.province,
          full_name: school.full_name || school.school_name
        }))
        setSchools(transformedSchools)
      } else {
        setSchools([])
        console.error('Search error:', data)
      }
    } catch (error) {
      console.error('School search error:', error)
      // Try backend API as fallback
      try {
        const backendResponse = await apiClient.get(`/api/v1/opcode/schools/search?keyword=${encodeURIComponent(keyword)}`)
        if ((backendResponse.data as any).success) {
          setSchools((backendResponse.data as any).data.schools || [])
        }
      } catch (backendError) {
        toast({
          title: '搜索失败',
          description: '无法获取学校列表',
          variant: 'destructive'
        })
        setSchools([])
      }
    } finally {
      setIsSearching(false)
    }
  }

  // 选择学校后加载片区
  const selectSchool = async (school: School) => {
    setSelection({ ...selection, school, district: undefined, building: undefined, deliveryPoint: undefined })
    setLoading(true)
    try {
      const response = await fetch(`/api/schools/districts?school_code=${school.school_code}`)
      const data = await response.json()
      
      if (data.success || data.code === 0) {
        const districtsData = data.data?.districts || data.data?.areas || []
        // 转换数据格式
        const formattedDistricts = districtsData.map((d: any) => ({
          code: d.area_code || d.code,
          name: d.area_name || d.name,
          description: d.description || ''
        }))
        setDistricts(formattedDistricts)
      } else {
        throw new Error('Failed to fetch districts')
      }
    } catch (error) {
      console.error('Failed to fetch districts:', error)
      // 使用模拟数据
      setDistricts([
        { code: '01', name: '东区', description: '宿舍楼1-5栋' },
        { code: '02', name: '西区', description: '宿舍楼6-10栋' },
        { code: '03', name: '南区', description: '宿舍楼11-15栋' },
        { code: '04', name: '北区', description: '宿舍楼16-20栋' },
        { code: '05', name: '中心区', description: '教学楼、图书馆' }
      ])
    } finally {
      setLoading(false)
    }
  }

  // 选择片区后加载楼栋
  const selectDistrict = async (district: District) => {
    setSelection({ ...selection, district, building: undefined, deliveryPoint: undefined })
    setLoading(true)
    try {
      const schoolCode = selection.school?.school_code
      const response = await fetch(`/api/schools/buildings?school_code=${schoolCode}&district_code=${district.code}`)
      const data = await response.json()
      
      if (data.success || data.code === 0) {
        setBuildings(data.data?.buildings || [])
      } else {
        throw new Error('Failed to fetch buildings')
      }
    } catch (error) {
      console.error('Failed to fetch buildings:', error)
      // 使用模拟数据
      setBuildings([
        { code: 'A', name: 'A栋', type: 'dormitory' },
        { code: 'B', name: 'B栋', type: 'dormitory' },
        { code: 'C', name: 'C栋', type: 'dormitory' },
        { code: 'D', name: 'D栋', type: 'teaching' },
        { code: 'E', name: 'E栋', type: 'dining' },
        { code: 'F', name: 'F栋', type: 'dormitory' }
      ])
    } finally {
      setLoading(false)
    }
  }

  // 选择楼栋后加载投递点
  const selectBuilding = async (building: Building) => {
    setSelection({ ...selection, building, deliveryPoint: undefined })
    setLoading(true)
    try {
      const prefix = `${selection.school?.school_code}${selection.district?.code}${building.code}`
      const response = await fetch(`/api/schools/delivery-points?prefix=${prefix}`)
      const data = await response.json()
      
      if (data.success || data.code === 0) {
        setDeliveryPoints(data.data?.points || [])
        // 获取推荐的未占用编码
        const available = (data.data?.points || []).filter((p: DeliveryPoint) => p.available)
        setRecommendedCodes(available.slice(0, 5).map((p: DeliveryPoint) => `${prefix}${p.code}`))
      } else {
        throw new Error('Failed to fetch delivery points')
      }
    } catch (error) {
      console.error('Failed to fetch delivery points:', error)
      // 使用模拟数据
      const points: DeliveryPoint[] = []
      for (let floor = 1; floor <= 6; floor++) {
        for (let room = 1; room <= 10; room++) {
          const code = `${floor}${room.toString().padStart(2, '0')}`
          points.push({
            code: code.slice(-2),
            name: `${floor}${room.toString().padStart(2, '0')}室`,
            available: Math.random() > 0.3,
            type: 'room'
          })
        }
      }
      setDeliveryPoints(points)
      const prefix = `${selection.school?.school_code}${selection.district?.code}${building.code}`
      const available = points.filter(p => p.available)
      setRecommendedCodes(available.slice(0, 5).map(p => `${prefix}${p.code}`))
    } finally {
      setLoading(false)
    }
  }

  // 选择投递点
  const selectDeliveryPoint = (point: DeliveryPoint) => {
    const finalCode = `${selection.school?.school_code}${selection.district?.code}${selection.building?.code}${point.code}`
    setSelection({ ...selection, deliveryPoint: point, finalCode })
  }

  // 申请选定的OP Code
  const applyForCode = async () => {
    if (!selection.finalCode) return
    
    setLoading(true)
    try {
      const response = await apiClient.post('/api/v1/opcode/apply', {
        code: selection.finalCode,
        type: 'dormitory',
        description: `${selection.school?.school_name} ${selection.district?.name} ${selection.building?.name} ${selection.deliveryPoint?.name}`
      })
      if ((response.data as any).success) {
        toast({
          title: '申请成功',
          description: `OP Code ${selection.finalCode} 申请已提交`,
        })
        // 重置选择
        setSelection({})
      }
    } catch (error) {
      toast({
        title: '申请失败',
        description: '请稍后再试',
        variant: 'destructive'
      })
    } finally {
      setLoading(false)
    }
  }

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-amber-50 flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <AlertCircle className="w-12 h-12 text-amber-600 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-amber-900 mb-2">需要登录</h2>
            <p className="text-amber-700 mb-4">
              请先登录以访问 OP Code 系统
            </p>
            <Button asChild>
              <a href="/login">前往登录</a>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-7xl mx-auto px-4 py-8">
        {/* 页面标题 */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <MapPin className="w-8 h-8 text-amber-600" />
            <h1 className="text-3xl font-bold text-amber-900">OP Code 地理编码系统</h1>
          </div>
          <p className="text-amber-700">
            通过层级选择精准定位投递地址 - 学校 → 片区 → 楼栋 → 投递点
          </p>
        </div>

        {/* 主要功能标签页 */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="bg-amber-100">
            <TabsTrigger value="apply" className="data-[state=active]:bg-amber-200">
              申请编码
            </TabsTrigger>
            <TabsTrigger value="search" className="data-[state=active]:bg-amber-200">
              查询编码
            </TabsTrigger>
            <TabsTrigger value="manage" className="data-[state=active]:bg-amber-200">
              我的编码
            </TabsTrigger>
            <TabsTrigger value="help" className="data-[state=active]:bg-amber-200">
              使用帮助
            </TabsTrigger>
          </TabsList>

          {/* 申请编码 - 层级选择 */}
          <TabsContent value="apply" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900">申请新的 OP Code</CardTitle>
              </CardHeader>
              <CardContent>
                {/* 进度指示器 */}
                <div className="mb-8">
                  <div className="flex items-center justify-between mb-2">
                    <div className={`flex items-center gap-2 ${selection.school ? 'text-green-600' : 'text-gray-400'}`}>
                      <div className={`w-8 h-8 rounded-full flex items-center justify-center ${selection.school ? 'bg-green-100' : 'bg-gray-100'}`}>
                        {selection.school ? <CheckCircle className="w-5 h-5" /> : '1'}
                      </div>
                      <span className="text-sm font-medium">选择学校</span>
                    </div>
                    <ArrowRight className="w-4 h-4 text-gray-400" />
                    <div className={`flex items-center gap-2 ${selection.district ? 'text-green-600' : 'text-gray-400'}`}>
                      <div className={`w-8 h-8 rounded-full flex items-center justify-center ${selection.district ? 'bg-green-100' : 'bg-gray-100'}`}>
                        {selection.district ? <CheckCircle className="w-5 h-5" /> : '2'}
                      </div>
                      <span className="text-sm font-medium">选择片区</span>
                    </div>
                    <ArrowRight className="w-4 h-4 text-gray-400" />
                    <div className={`flex items-center gap-2 ${selection.building ? 'text-green-600' : 'text-gray-400'}`}>
                      <div className={`w-8 h-8 rounded-full flex items-center justify-center ${selection.building ? 'bg-green-100' : 'bg-gray-100'}`}>
                        {selection.building ? <CheckCircle className="w-5 h-5" /> : '3'}
                      </div>
                      <span className="text-sm font-medium">选择楼栋</span>
                    </div>
                    <ArrowRight className="w-4 h-4 text-gray-400" />
                    <div className={`flex items-center gap-2 ${selection.deliveryPoint ? 'text-green-600' : 'text-gray-400'}`}>
                      <div className={`w-8 h-8 rounded-full flex items-center justify-center ${selection.deliveryPoint ? 'bg-green-100' : 'bg-gray-100'}`}>
                        {selection.deliveryPoint ? <CheckCircle className="w-5 h-5" /> : '4'}
                      </div>
                      <span className="text-sm font-medium">选择投递点</span>
                    </div>
                  </div>
                </div>

                {/* 当前选择显示 */}
                {selection.finalCode && (
                  <Alert className="mb-6 bg-green-50 border-green-200">
                    <CheckCircle className="h-4 w-4 text-green-600" />
                    <AlertDescription className="text-green-800">
                      <div className="font-semibold mb-1">您选择的 OP Code：{selection.finalCode}</div>
                      <div className="text-sm">
                        {selection.school?.school_name} - {selection.district?.name} - {selection.building?.name} - {selection.deliveryPoint?.name}
                      </div>
                    </AlertDescription>
                  </Alert>
                )}

                {/* 步骤1：搜索学校 */}
                <div className="space-y-6">
                  <div>
                    <Label className="text-base font-semibold mb-3 block">步骤1：搜索学校</Label>
                    <div className="relative">
                      <Input
                        placeholder="输入关键词搜索学校（如：长沙、北京、复旦等）"
                        value={searchKeyword}
                        onChange={(e) => setSearchKeyword(e.target.value)}
                        className="pr-10"
                      />
                      {isSearching && (
                        <div className="absolute right-3 top-1/2 -translate-y-1/2">
                          <Loader2 className="w-4 h-4 animate-spin text-gray-400" />
                        </div>
                      )}
                      {!isSearching && searchKeyword && (
                        <div className="absolute right-3 top-1/2 -translate-y-1/2">
                          <Search className="w-4 h-4 text-gray-400" />
                        </div>
                      )}
                    </div>
                    {searchKeyword && schools.length === 0 && !isSearching && (
                      <Alert className="bg-amber-50 border-amber-200">
                        <AlertCircle className="h-4 w-4 text-amber-600" />
                        <AlertDescription className="text-amber-800">
                          没有找到相关学校，请尝试其他关键词
                        </AlertDescription>
                      </Alert>
                    )}
                  </div>

                  {/* 学校列表 */}
                  {schools.length > 0 && (
                    <div>
                      <Label className="text-base font-semibold mb-3 block">
                        选择学校（确定前2位编码）
                      </Label>
                      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 max-h-96 overflow-y-auto">
                        {schools.map((school) => (
                          <Card 
                            key={school.school_code}
                            className={`cursor-pointer transition-all ${
                              selection.school?.school_code === school.school_code 
                                ? 'ring-2 ring-amber-500 bg-amber-50' 
                                : 'hover:shadow-md'
                            }`}
                            onClick={() => selectSchool(school)}
                          >
                            <CardContent className="p-4">
                              <div className="flex items-start gap-3">
                                <School className="w-5 h-5 text-blue-600 mt-1 flex-shrink-0" />
                                <div className="flex-1 min-w-0">
                                  <h4 className="font-semibold text-sm mb-1 truncate">{school.school_name}</h4>
                                  <Badge variant="outline" className="text-xs mb-1">
                                    {school.school_code}
                                  </Badge>
                                  <div className="text-xs text-gray-600">
                                    {school.city} · {school.province}
                                  </div>
                                </div>
                              </div>
                            </CardContent>
                          </Card>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* 步骤2：选择片区 */}
                  {districts.length > 0 && selection.school && (
                    <div>
                      <Label className="text-base font-semibold mb-3 block">
                        步骤2：选择片区（确定第3位编码）
                      </Label>
                      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
                        {districts.map((district) => (
                          <Card 
                            key={district.code}
                            className={`cursor-pointer transition-all ${
                              selection.district?.code === district.code 
                                ? 'ring-2 ring-amber-500 bg-amber-50' 
                                : 'hover:shadow-md'
                            }`}
                            onClick={() => selectDistrict(district)}
                          >
                            <CardContent className="p-4 text-center">
                              <MapPin className="w-6 h-6 text-green-600 mx-auto mb-2" />
                              <h4 className="font-semibold text-sm">{district.name}</h4>
                              <div className="text-xs text-gray-500 mt-1">{district.description}</div>
                              <Badge variant="outline" className="mt-2 text-xs">{district.code}</Badge>
                            </CardContent>
                          </Card>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* 步骤3：选择楼栋 */}
                  {buildings.length > 0 && selection.district && (
                    <div>
                      <Label className="text-base font-semibold mb-3 block">
                        步骤3：选择楼栋（确定第4位编码）
                      </Label>
                      <div className="grid grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
                        {buildings.map((building) => (
                          <Card 
                            key={building.code}
                            className={`cursor-pointer transition-all ${
                              selection.building?.code === building.code 
                                ? 'ring-2 ring-amber-500 bg-amber-50' 
                                : 'hover:shadow-md'
                            }`}
                            onClick={() => selectBuilding(building)}
                          >
                            <CardContent className="p-4 text-center">
                              {building.type === 'dormitory' && <Home className="w-6 h-6 text-blue-600 mx-auto mb-2" />}
                              {building.type === 'teaching' && <Building className="w-6 h-6 text-purple-600 mx-auto mb-2" />}
                              {building.type === 'dining' && <Package className="w-6 h-6 text-orange-600 mx-auto mb-2" />}
                              <h4 className="font-semibold text-sm">{building.name}</h4>
                              <Badge variant="outline" className="mt-2 text-xs">{building.code}</Badge>
                            </CardContent>
                          </Card>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* 步骤4：选择投递点 */}
                  {deliveryPoints.length > 0 && selection.building && (
                    <div>
                      <Label className="text-base font-semibold mb-3 block">
                        步骤4：选择投递点（确定第5-6位编码）
                      </Label>
                      
                      {/* 推荐的可用编码 */}
                      {recommendedCodes.length > 0 && (
                        <Alert className="mb-4 bg-blue-50 border-blue-200">
                          <AlertCircle className="h-4 w-4 text-blue-600" />
                          <AlertDescription>
                            <div className="font-medium text-blue-900 mb-2">推荐可用编码：</div>
                            <div className="flex flex-wrap gap-2">
                              {recommendedCodes.map((code) => (
                                <Badge key={code} variant="secondary" className="bg-blue-100 text-blue-800">
                                  {code}
                                </Badge>
                              ))}
                            </div>
                          </AlertDescription>
                        </Alert>
                      )}

                      <div className="grid grid-cols-4 md:grid-cols-6 lg:grid-cols-10 gap-2 max-h-64 overflow-y-auto">
                        {deliveryPoints.map((point) => (
                          <Button
                            key={point.code}
                            variant={selection.deliveryPoint?.code === point.code ? "default" : point.available ? "outline" : "ghost"}
                            size="sm"
                            onClick={() => point.available && selectDeliveryPoint(point)}
                            disabled={!point.available}
                            className={`relative ${!point.available && 'opacity-50'}`}
                          >
                            {point.name}
                            {!point.available && (
                              <span className="absolute -top-1 -right-1 w-2 h-2 bg-red-500 rounded-full" />
                            )}
                          </Button>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* 提交申请按钮 */}
                  {selection.finalCode && (
                    <div className="flex justify-center pt-4">
                      <Button 
                        size="lg" 
                        onClick={applyForCode}
                        disabled={loading}
                        className="min-w-[200px]"
                      >
                        {loading ? (
                          <>
                            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                            申请中...
                          </>
                        ) : (
                          <>
                            <CheckCircle className="w-4 h-4 mr-2" />
                            申请 {selection.finalCode}
                          </>
                        )}
                      </Button>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* 查询编码 */}
          <TabsContent value="search" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900">查询 OP Code</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <Label>输入 OP Code 查询</Label>
                    <div className="flex gap-2 mt-2">
                      <Input placeholder="例如：PK5F3D" className="max-w-xs" />
                      <Button>查询</Button>
                    </div>
                  </div>
                  
                  <Alert>
                    <MapPin className="h-4 w-4" />
                    <AlertDescription>
                      支持模糊查询：PK**** 查询北京大学所有位置
                    </AlertDescription>
                  </Alert>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* 我的编码 */}
          <TabsContent value="manage" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900">我的 OP Code</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-center py-8 text-gray-500">
                  <MapPin className="w-12 h-12 mx-auto mb-4 opacity-50" />
                  <p>您还没有申请任何 OP Code</p>
                  <Button className="mt-4" onClick={() => setActiveTab('apply')}>
                    申请新编码
                  </Button>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* 使用帮助 */}
          <TabsContent value="help" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900">使用帮助</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="prose prose-sm max-w-none">
                  <h3>什么是 OP Code？</h3>
                  <p>
                    OP Code 是 OpenPenPal 的 6 位地理编码系统，用于精准定位校园内的投递地址。
                  </p>
                  
                  <h3>编码结构</h3>
                  <ul>
                    <li><strong>第1-2位</strong>：学校代码（如 PK = 北京大学）</li>
                    <li><strong>第3位</strong>：片区代码（如 5 = 第五片区）</li>
                    <li><strong>第4位</strong>：楼栋代码（如 F = F栋）</li>
                    <li><strong>第5-6位</strong>：具体投递点（如 3D = 303室）</li>
                  </ul>
                  
                  <h3>信使权限对应</h3>
                  <ul>
                    <li><strong>四级信使</strong>：管理整个城市（所有学校）</li>
                    <li><strong>三级信使</strong>：管理整个学校（PK****）</li>
                    <li><strong>二级信使</strong>：管理学校片区（PK5***）</li>
                    <li><strong>一级信使</strong>：管理具体楼栋（PK5F**）</li>
                  </ul>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}