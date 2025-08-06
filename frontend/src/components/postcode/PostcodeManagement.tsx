'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  Building, 
  MapPin, 
  Search, 
  Plus, 
  Edit, 
  Trash2, 
  Eye,
  AlertCircle,
  CheckCircle,
  Clock,
  BarChart3
} from 'lucide-react'
import { useCourierPermission } from '@/hooks/use-courier-permission'
import PostcodeService from '@/lib/services/postcode-service'
import type { 
  SchoolSite, 
  SiteArea, 
  AreaBuilding, 
  BuildingRoom,
  AddressFeedback,
  PostcodeStats
} from '@/lib/types/postcode'

interface PostcodeManagementProps {
  className?: string
}

export function PostcodeManagement({ className = '' }: PostcodeManagementProps) {
  const { courierInfo, hasCourierPermission, COURIER_PERMISSIONS } = useCourierPermission()
  
  // 状态管理
  const [activeTab, setActiveTab] = useState('structure')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  
  // 地址结构数据
  const [schools, setSchools] = useState<SchoolSite[]>([])
  const [selectedSchool, setSelectedSchool] = useState<SchoolSite | null>(null)
  const [areas, setAreas] = useState<SiteArea[]>([])
  const [buildings, setBuildings] = useState<AreaBuilding[]>([])
  const [rooms, setRooms] = useState<BuildingRoom[]>([])
  
  // 地址反馈数据
  const [feedbacks, setFeedbacks] = useState<AddressFeedback[]>([])
  const [stats, setStats] = useState<PostcodeStats[]>([])
  
  // 搜索状态
  const [searchQuery, setSearchQuery] = useState('')

  // 权限检查
  const canManagePostcode = hasCourierPermission(COURIER_PERMISSIONS.MANAGE_CITY_OPERATIONS)
  const canCreateSchools = hasCourierPermission(COURIER_PERMISSIONS.CREATE_SCHOOL_LEVEL_COURIER)

  // 初始化数据加载
  useEffect(() => {
    if (canManagePostcode) {
      loadInitialData()
    }
  }, [canManagePostcode])

  const loadInitialData = async () => {
    try {
      setLoading(true)
      setError(null)
      
      // 并行加载数据
      const [schoolsRes, feedbacksRes, statsRes] = await Promise.all([
        PostcodeService.getSchoolSites(),
        PostcodeService.getPendingFeedbacks(),
        PostcodeService.getPostcodeStats()
      ])
      
      if (schoolsRes.success && schoolsRes.data) {
        setSchools(schoolsRes.data)
      }
      
      if (feedbacksRes.success && feedbacksRes.data) {
        setFeedbacks(feedbacksRes.data)
      }
      
      if (statsRes.success && statsRes.data) {
        setStats(statsRes.data)
      }
      
    } catch (error) {
      console.error('Failed to load postcode data:', error)
      setError('加载 Postcode 数据失败')
      
      // 使用模拟数据作为后备
      setSchools([
        {
          id: 'school_pk',
          code: 'PK',
          name: '北京大学',
          fullName: '北京大学',
          status: 'active',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
          managedBy: 'courier_level4_city'
        },
        {
          id: 'school_qh', 
          code: 'QH',
          name: '清华大学',
          fullName: '清华大学',
          status: 'active',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
          managedBy: 'courier_level4_city'
        }
      ])
    } finally {
      setLoading(false)
    }
  }

  // 选择学校时加载片区数据
  const handleSchoolSelect = async (school: SchoolSite) => {
    setSelectedSchool(school)
    setAreas([])
    setBuildings([])
    setRooms([])
    
    try {
      setLoading(true)
      const response = await PostcodeService.getSchoolAreas(school.code)
      if (response.success && response.data) {
        setAreas(response.data)
      }
    } catch (error) {
      console.error('Failed to load areas:', error)
      // 使用模拟数据
      setAreas([
        {
          id: 'area_pk5',
          schoolCode: school.code,
          code: '5',
          name: '第五片区',
          description: '主要宿舍区域',
          status: 'active',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
          managedBy: 'courier_level3_school'
        },
        {
          id: 'area_pk3',
          schoolCode: school.code,
          code: '3', 
          name: '第三片区',
          description: '教学区域',
          status: 'active',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
          managedBy: 'courier_level3_school'
        }
      ])
    } finally {
      setLoading(false)
    }
  }

  // 处理地址反馈审核
  const handleFeedbackReview = async (feedbackId: string, action: 'approve' | 'reject', notes?: string) => {
    try {
      setLoading(true)
      const response = await PostcodeService.reviewAddressFeedback(feedbackId, action, notes)
      
      if (response.success) {
        // 更新本地状态
        setFeedbacks(prev => prev.filter(f => f.id !== feedbackId))
        alert(`地址反馈已${action === 'approve' ? '批准' : '拒绝'}`)
      }
    } catch (error) {
      console.error('Failed to review feedback:', error)
      setError('处理地址反馈失败')
    } finally {
      setLoading(false)
    }
  }

  // 权限不足提示
  if (!canManagePostcode) {
    return (
      <Card className={className}>
        <CardContent className="pt-6 text-center">
          <AlertCircle className="w-12 h-12 text-amber-500 mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-amber-900 mb-2">权限不足</h3>
          <p className="text-amber-700">
            只有四级信使（城市总代）才能管理 Postcode 编码系统
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* 页面标题 */}
      <div>
        <h2 className="text-2xl font-bold text-amber-900 mb-2">Postcode 编码系统管理</h2>
        <p className="text-amber-700">管理城市地址编码结构和信使投递权限</p>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 快速统计 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center gap-2">
              <Building className="w-5 h-5 text-blue-600" />
              <div>
                <div className="text-2xl font-bold">{schools.length}</div>
                <div className="text-sm text-gray-600">管理学校</div>
              </div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center gap-2">
              <MapPin className="w-5 h-5 text-green-600" />
              <div>
                <div className="text-2xl font-bold">{areas.length}</div>
                <div className="text-sm text-gray-600">活跃片区</div>
              </div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center gap-2">
              <Clock className="w-5 h-5 text-orange-600" />
              <div>
                <div className="text-2xl font-bold">{feedbacks.length}</div>
                <div className="text-sm text-gray-600">待审核</div>
              </div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="pt-4">
            <div className="flex items-center gap-2">
              <BarChart3 className="w-5 h-5 text-purple-600" />
              <div>
                <div className="text-2xl font-bold">{stats.length}</div>
                <div className="text-sm text-gray-600">编码统计</div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 主要功能标签页 */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
        <TabsList className="bg-amber-100">
          <TabsTrigger value="structure" className="data-[state=active]:bg-amber-200">
            地址结构管理
          </TabsTrigger>
          <TabsTrigger value="feedback" className="data-[state=active]:bg-amber-200">
            地址反馈审核 {feedbacks.length > 0 && <Badge className="ml-2">{feedbacks.length}</Badge>}
          </TabsTrigger>
          <TabsTrigger value="analytics" className="data-[state=active]:bg-amber-200">
            使用分析
          </TabsTrigger>
        </TabsList>

        {/* 地址结构管理 */}
        <TabsContent value="structure" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* 学校列表 */}
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900 flex items-center justify-between">
                  学校管理
                  {canCreateSchools && (
                    <Button size="sm" variant="outline">
                      <Plus className="w-4 h-4 mr-2" />
                      新增
                    </Button>
                  )}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  {schools.map((school) => (
                    <div
                      key={school.id}
                      className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                        selectedSchool?.id === school.id
                          ? 'border-amber-300 bg-amber-50'
                          : 'hover:bg-gray-50'
                      }`}
                      onClick={() => handleSchoolSelect(school)}
                    >
                      <div className="flex items-center justify-between">
                        <div>
                          <div className="font-medium">{school.name}</div>
                          <div className="text-sm text-gray-600">
                            <Badge variant="outline">{school.code}</Badge>
                          </div>
                        </div>
                        <div className="flex gap-1">
                          <Button size="sm" variant="ghost">
                            <Edit className="w-4 h-4" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            {/* 片区列表 */}
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900 flex items-center justify-between">
                  片区管理
                  {selectedSchool && (
                    <Button size="sm" variant="outline">
                      <Plus className="w-4 h-4 mr-2" />
                      新增
                    </Button>
                  )}
                </CardTitle>
              </CardHeader>
              <CardContent>
                {!selectedSchool ? (
                  <div className="text-center text-gray-500 py-8">
                    请先选择一个学校
                  </div>
                ) : (
                  <div className="space-y-2">
                    {areas.map((area) => (
                      <div
                        key={area.id}
                        className="p-3 border rounded-lg hover:bg-gray-50 transition-colors"
                      >
                        <div className="flex items-center justify-between">
                          <div>
                            <div className="font-medium">{area.name}</div>
                            <div className="text-sm text-gray-600">
                              <Badge variant="outline">{selectedSchool.code}{area.code}</Badge>
                              {area.description && (
                                <span className="ml-2">{area.description}</span>
                              )}
                            </div>
                          </div>
                          <div className="flex gap-1">
                            <Button size="sm" variant="ghost">
                              <Eye className="w-4 h-4" />
                            </Button>
                            <Button size="sm" variant="ghost">
                              <Edit className="w-4 h-4" />
                            </Button>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* 编码使用概览 */}
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900">编码使用概览</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="text-sm">
                    <div className="flex justify-between mb-1">
                      <span>活跃编码</span>
                      <span className="font-medium">1,247</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div className="bg-green-600 h-2 rounded-full" style={{ width: '78%' }}></div>
                    </div>
                  </div>
                  
                  <div className="text-sm">
                    <div className="flex justify-between mb-1">
                      <span>投递成功率</span>
                      <span className="font-medium">94.2%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div className="bg-blue-600 h-2 rounded-full" style={{ width: '94%' }}></div>
                    </div>
                  </div>
                  
                  <div className="text-sm">
                    <div className="flex justify-between mb-1">
                      <span>错误率</span>
                      <span className="font-medium">2.1%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div className="bg-red-600 h-2 rounded-full" style={{ width: '2%' }}></div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* 地址反馈审核 */}
        <TabsContent value="feedback" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-amber-900">待审核地址反馈</CardTitle>
            </CardHeader>
            <CardContent>
              {feedbacks.length === 0 ? (
                <div className="text-center py-8">
                  <CheckCircle className="w-12 h-12 text-green-500 mx-auto mb-4" />
                  <h3 className="text-lg font-semibold text-gray-900 mb-2">暂无待审核反馈</h3>
                  <p className="text-gray-600">所有地址反馈都已处理完成</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {feedbacks.map((feedback) => (
                    <div key={feedback.id} className="border rounded-lg p-4">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-2">
                            <Badge variant={
                              feedback.type === 'new_address' ? 'default' :
                              feedback.type === 'error_report' ? 'destructive' : 'secondary'
                            }>
                              {feedback.type === 'new_address' ? '新增地址' :
                               feedback.type === 'error_report' ? '错误报告' : '投递失败'}
                            </Badge>
                            <span className="text-sm text-gray-500">
                              {feedback.submitterType === 'user' ? '用户' : '信使'}提交
                            </span>
                          </div>
                          
                          <div className="mb-2">
                            <strong>描述：</strong>
                            <p className="text-gray-700 mt-1">{feedback.description}</p>
                          </div>
                          
                          {feedback.postcode && (
                            <div className="mb-2">
                              <strong>相关编码：</strong>
                              <Badge variant="outline" className="ml-2">{feedback.postcode}</Badge>
                            </div>
                          )}
                          
                          <div className="text-sm text-gray-500">
                            提交时间：{new Date(feedback.createdAt).toLocaleString()}
                          </div>
                        </div>
                        
                        <div className="flex gap-2 ml-4">
                          <Button
                            size="sm"
                            variant="outline"
                            className="border-green-300 text-green-700 hover:bg-green-50"
                            onClick={() => handleFeedbackReview(feedback.id, 'approve', '已批准')}
                            disabled={loading}
                          >
                            <CheckCircle className="w-4 h-4 mr-2" />
                            批准
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            className="border-red-300 text-red-700 hover:bg-red-50"
                            onClick={() => handleFeedbackReview(feedback.id, 'reject', '不符合要求')}
                            disabled={loading}
                          >
                            <Trash2 className="w-4 h-4 mr-2" />
                            拒绝
                          </Button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* 使用分析 */}
        <TabsContent value="analytics" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900">热门地址 Top 10</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {['PK5F3D', 'PK5F2A', 'QH1C2E', 'PK3A1B'].map((code, index) => (
                    <div key={code} className="flex items-center justify-between p-2 border rounded">
                      <div className="flex items-center gap-3">
                        <div className="w-6 h-6 bg-amber-600 text-white rounded-full flex items-center justify-center text-sm font-bold">
                          {index + 1}
                        </div>
                        <Badge variant="outline">{code}</Badge>
                      </div>
                      <div className="text-sm text-gray-600">
                        {Math.floor(Math.random() * 50) + 10} 次使用
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900">问题地址监控</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {['PK7F1A', 'QH2B3C'].map((code, index) => (
                    <div key={code} className="flex items-center justify-between p-2 border border-red-200 rounded bg-red-50">
                      <div className="flex items-center gap-3">
                        <AlertCircle className="w-5 h-5 text-red-600" />
                        <Badge variant="outline">{code}</Badge>
                      </div>
                      <div className="text-sm text-red-600">
                        错误率 {Math.floor(Math.random() * 20) + 5}%
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}

export default PostcodeManagement