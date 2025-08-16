'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
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
  School
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { OPCodeDisplay } from '@/components/user/opcode-display'

export default function OPCodePage() {
  const { user, isAuthenticated } = useAuth()
  const [activeTab, setActiveTab] = useState('overview')
  const [searchCity, setSearchCity] = useState('')
  const [searchResults, setSearchResults] = useState<any[]>([])
  const [loading, setLoading] = useState(false)

  // 城市搜索功能
  const handleCitySearch = async () => {
    if (!searchCity.trim()) return
    
    setLoading(true)
    try {
      const response = await fetch(`/api/v1/opcode/search/schools/by-city?city=${encodeURIComponent(searchCity)}&limit=50`)
      if (response.ok) {
        const data = await response.json()
        if (data.success) {
          setSearchResults(data.data.schools || [])
        }
      }
    } catch (error) {
      console.error('城市搜索失败:', error)
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
            <h1 className="text-3xl font-bold text-amber-900">OpenPenPal OP Code 系统</h1>
          </div>
          <p className="text-amber-700">
            基于四级信使体系的统一地理编码管理平台 - 6位精准定位系统
          </p>
        </div>

        {/* 主要功能标签页 */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="bg-amber-100">
            <TabsTrigger value="overview" className="data-[state=active]:bg-amber-200">
              系统概览
            </TabsTrigger>
            <TabsTrigger value="search" className="data-[state=active]:bg-amber-200">
              城市搜索
            </TabsTrigger>
            <TabsTrigger value="segmented" className="data-[state=active]:bg-amber-200">
              分段查询
            </TabsTrigger>
            <TabsTrigger value="documentation" className="data-[state=active]:bg-amber-200">
              使用文档
            </TabsTrigger>
          </TabsList>

          {/* 系统概览 */}
          <TabsContent value="overview" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* OP Code规则说明 */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-amber-900">OP Code 编码规则</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="text-center">
                      <div className="text-3xl font-mono font-bold text-amber-800 mb-2 p-4 bg-amber-100 rounded-lg">
                        PK 5F 3D
                      </div>
                      <div className="text-sm text-gray-600">示例：北京大学第五片区F栋3D宿舍</div>
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center justify-between p-2 border rounded">
                        <div className="flex items-center gap-2">
                          <Badge variant="outline">PK</Badge>
                          <span className="text-sm">学校代码</span>
                        </div>
                        <span className="text-sm text-gray-600">第1-2位</span>
                      </div>

                      <div className="flex items-center justify-between p-2 border rounded">
                        <div className="flex items-center gap-2">
                          <Badge variant="outline">5F</Badge>
                          <span className="text-sm">片区代码</span>
                        </div>
                        <span className="text-sm text-gray-600">第3-4位</span>
                      </div>

                      <div className="flex items-center justify-between p-2 border rounded">
                        <div className="flex items-center gap-2">
                          <Badge variant="outline">3D</Badge>
                          <span className="text-sm">位置代码</span>
                        </div>
                        <span className="text-sm text-gray-600">第5-6位</span>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* 权限层级说明 */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-amber-900">四级信使权限对应</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between p-3 border rounded-lg bg-purple-50 border-purple-200">
                      <div className="flex items-center gap-3">
                        <Crown className="w-5 h-5 text-purple-600" />
                        <div>
                          <div className="font-semibold">四级信使</div>
                          <div className="text-sm text-gray-600">城市总代</div>
                        </div>
                      </div>
                      <Badge variant="outline">PK**** (全城)</Badge>
                    </div>

                    <div className="flex items-center justify-between p-3 border rounded-lg bg-blue-50 border-blue-200">
                      <div className="flex items-center gap-3">
                        <Building className="w-5 h-5 text-blue-600" />
                        <div>
                          <div className="font-semibold">三级信使</div>
                          <div className="text-sm text-gray-600">校级负责人</div>
                        </div>
                      </div>
                      <Badge variant="outline">PK** (学校级)</Badge>
                    </div>

                    <div className="flex items-center justify-between p-3 border rounded-lg bg-green-50 border-green-200">
                      <div className="flex items-center gap-3">
                        <MapPin className="w-5 h-5 text-green-600" />
                        <div>
                          <div className="font-semibold">二级信使</div>
                          <div className="text-sm text-gray-600">片区管理员</div>
                        </div>
                      </div>
                      <Badge variant="outline">PK5F** (片区级)</Badge>
                    </div>

                    <div className="flex items-center justify-between p-3 border rounded-lg bg-amber-50 border-amber-200">
                      <div className="flex items-center gap-3">
                        <Users className="w-5 h-5 text-amber-600" />
                        <div>
                          <div className="font-semibold">一级信使</div>
                          <div className="text-sm text-gray-600">楼栋投递员</div>
                        </div>
                      </div>
                      <Badge variant="outline">PK5F3D (精确位置)</Badge>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* 城市搜索 */}
          <TabsContent value="search" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900 flex items-center gap-2">
                  <Search className="w-5 h-5" />
                  城市级学校搜索
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex gap-4 mb-6">
                  <Input
                    placeholder="请输入城市名称（如：北京、上海、广州）"
                    value={searchCity}
                    onChange={(e) => setSearchCity(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && handleCitySearch()}
                    className="flex-1"
                  />
                  <Button onClick={handleCitySearch} disabled={loading}>
                    {loading ? '搜索中...' : '搜索'}
                  </Button>
                </div>

                {searchResults.length > 0 && (
                  <div className="space-y-4">
                    <div className="text-sm text-gray-600">
                      找到 {searchResults.length} 所学校在 "{searchCity}" 地区
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                      {searchResults.map((school) => (
                        <Card key={school.school_code} className="hover:shadow-md transition-shadow">
                          <CardContent className="p-4">
                            <div className="flex items-start gap-3">
                              <School className="w-5 h-5 text-blue-600 mt-1" />
                              <div className="flex-1">
                                <h4 className="font-semibold text-sm mb-1">{school.school_name}</h4>
                                <div className="space-y-1">
                                  <Badge variant="outline" className="text-xs">
                                    {school.school_code}
                                  </Badge>
                                  <div className="text-xs text-gray-600">
                                    {school.city} · {school.province}
                                  </div>
                                  {school.full_name && school.full_name !== school.school_name && (
                                    <div className="text-xs text-gray-500 truncate">
                                      {school.full_name}
                                    </div>
                                  )}
                                </div>
                              </div>
                            </div>
                          </CardContent>
                        </Card>
                      ))}
                    </div>
                  </div>
                )}

                {searchCity && searchResults.length === 0 && !loading && (
                  <div className="text-center py-8 text-gray-500">
                    <School className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>未找到"{searchCity}"地区的学校</p>
                    <p className="text-sm">请尝试调整搜索关键词</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          {/* 分段查询 */}
          <TabsContent value="segmented" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900">OP Code 分段查询</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <Alert>
                    <MapPin className="h-4 w-4" />
                    <AlertDescription>
                      支持按学校、片区、具体位置进行分层查询，实现精准定位
                    </AlertDescription>
                  </Alert>
                  
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="text-center p-4 border rounded-lg">
                      <div className="text-2xl font-mono font-bold text-blue-600">PK****</div>
                      <div className="text-sm text-gray-600 mt-2">学校级查询</div>
                      <div className="text-xs text-gray-500">查询北京大学所有位置</div>
                    </div>
                    
                    <div className="text-center p-4 border rounded-lg">
                      <div className="text-2xl font-mono font-bold text-green-600">PK5F**</div>
                      <div className="text-sm text-gray-600 mt-2">片区级查询</div>
                      <div className="text-xs text-gray-500">查询5号楼F区所有位置</div>
                    </div>
                    
                    <div className="text-center p-4 border rounded-lg">
                      <div className="text-2xl font-mono font-bold text-amber-600">PK5F3D</div>
                      <div className="text-sm text-gray-600 mt-2">精确查询</div>
                      <div className="text-xs text-gray-500">查询具体宿舍位置</div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* 使用文档 */}
          <TabsContent value="documentation" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900 flex items-center gap-2">
                  <BookOpen className="w-5 h-5" />
                  OP Code 使用文档
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="prose prose-sm max-w-none">
                  <h3>系统概述</h3>
                  <p>
                    OpenPenPal OP Code 系统是基于四级信使体系的统一地理编码管理平台，
                    为校园信件投递提供6位精准定位和权限管理功能。
                  </p>

                  <h3>核心特性</h3>
                  <ul>
                    <li><strong>6位编码</strong>：AABBCC格式，支持学校-片区-位置三级结构</li>
                    <li><strong>城市搜索</strong>：支持按城市名称快速查找所有学校</li>
                    <li><strong>分段查询</strong>：支持通配符查询和层级权限控制</li>
                    <li><strong>权限管理</strong>：基于四级信使体系的分层权限控制</li>
                  </ul>

                  <h3>编码规则</h3>
                  <ul>
                    <li><strong>AA</strong>：学校代码（如PK=北大，QH=清华）</li>
                    <li><strong>BB</strong>：片区代码（如5F=5号楼F区）</li>
                    <li><strong>CC</strong>：位置代码（如3D=303宿舍）</li>
                  </ul>

                  <h3>使用场景</h3>
                  <ul>
                    <li><strong>写信用户</strong>：通过OP Code精准指定收件地址</li>
                    <li><strong>投递信使</strong>：根据权限等级接收和执行投递任务</li>
                    <li><strong>管理人员</strong>：维护编码体系，处理申请和分配</li>
                  </ul>

                  <h3>新功能亮点</h3>
                  <ul>
                    <li><strong>城市级搜索</strong>：输入"北京"即可查看所有北京地区学校</li>
                    <li><strong>智能匹配</strong>：支持模糊搜索和智能提示</li>
                    <li><strong>实时同步</strong>：与信使系统实时同步，确保数据一致</li>
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