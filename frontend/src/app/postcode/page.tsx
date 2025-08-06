'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  MapPin, 
  Building, 
  Users, 
  Crown, 
  CheckCircle,
  AlertCircle,
  BookOpen,
  Settings
} from 'lucide-react'
import { AddressSelector } from '@/components/postcode/AddressSelector'
import { PostcodeManagement } from '@/components/postcode/PostcodeManagement'
import { usePostcodePermission } from '@/hooks/use-postcode-permission'
import { useAuth } from '@/contexts/auth-context-new'

export default function PostcodePage() {
  const { user, isAuthenticated } = useAuth()
  const { 
    postcodePermissions, 
    loading, 
    hasPostcodePermission,
    canManageAddressLevel,
    canCreateSubAddressLevel,
    getManagementScope,
    getPermissionSummary,
    isPostcodeAdmin
  } = usePostcodePermission()

  const [selectedPostcode, setSelectedPostcode] = useState('')
  const [selectedAddress, setSelectedAddress] = useState('')
  const [activeTab, setActiveTab] = useState('overview')

  const handleAddressChange = (postcode: string, fullAddress: string) => {
    setSelectedPostcode(postcode)
    setSelectedAddress(fullAddress)
  }

  const permissionSummary = getPermissionSummary()

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-amber-50 flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <AlertCircle className="w-12 h-12 text-amber-600 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-amber-900 mb-2">需要登录</h2>
            <p className="text-amber-700 mb-4">
              请先登录以访问 Postcode 编码系统
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
            <h1 className="text-3xl font-bold text-amber-900">OpenPenPal Postcode 编码系统</h1>
          </div>
          <p className="text-amber-700">
            基于四级信使体系的统一地址编码管理平台
          </p>
        </div>

        {/* 用户权限概览 */}
        <Card className="mb-8 border-amber-200">
          <CardHeader>
            <CardTitle className="text-amber-900 flex items-center gap-2">
              <Users className="w-5 h-5" />
              权限概览
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <div className="flex items-center gap-3">
                <Crown className="w-8 h-8 text-amber-600" />
                <div>
                  <div className="font-semibold">信使等级</div>
                  <div className="text-sm text-gray-600">
                    Level {permissionSummary.level} 
                    {permissionSummary.level === 4 && ' (城市总代)'}
                    {permissionSummary.level === 3 && ' (校级信使)'}
                    {permissionSummary.level === 2 && ' (片区信使)'}
                    {permissionSummary.level === 1 && ' (楼栋信使)'}
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <Building className="w-8 h-8 text-blue-600" />
                <div>
                  <div className="font-semibold">管理权限</div>
                  <div className="text-sm text-gray-600">
                    {permissionSummary.canManage ? (
                      <Badge variant="default">可管理</Badge>
                    ) : (
                      <Badge variant="secondary">仅投递</Badge>
                    )}
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <CheckCircle className={`w-8 h-8 ${permissionSummary.canCreate ? 'text-green-600' : 'text-gray-400'}`} />
                <div>
                  <div className="font-semibold">创建权限</div>
                  <div className="text-sm text-gray-600">
                    {permissionSummary.canCreate ? '可创建下级' : '无创建权限'}
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <Settings className="w-8 h-8 text-purple-600" />
                <div>
                  <div className="font-semibold">管理范围</div>
                  <div className="text-sm text-gray-600">
                    {permissionSummary.prefixCount} 个前缀
                  </div>
                </div>
              </div>
            </div>

            <div className="mt-4 p-3 bg-amber-50 border border-amber-200 rounded-lg">
              <div className="text-sm">
                <strong>权限范围：</strong>
                {getManagementScope()}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 主要功能标签页 */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="bg-amber-100">
            <TabsTrigger value="overview" className="data-[state=active]:bg-amber-200">
              系统概览
            </TabsTrigger>
            <TabsTrigger value="selector" className="data-[state=active]:bg-amber-200">
              地址选择器
            </TabsTrigger>
            {isPostcodeAdmin() && (
              <TabsTrigger value="management" className="data-[state=active]:bg-amber-200">
                系统管理
              </TabsTrigger>
            )}
            <TabsTrigger value="documentation" className="data-[state=active]:bg-amber-200">
              使用文档
            </TabsTrigger>
          </TabsList>

          {/* 系统概览 */}
          <TabsContent value="overview" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* 编码规则说明 */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-amber-900">Postcode 编码规则</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="text-center">
                      <div className="text-3xl font-mono font-bold text-amber-800 mb-2 p-4 bg-amber-100 rounded-lg">
                        PK 5 F 3D
                      </div>
                      <div className="text-sm text-gray-600">示例：北京大学第五片区F栋3D宿舍</div>
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center justify-between p-2 border rounded">
                        <div className="flex items-center gap-2">
                          <Badge variant="outline">PK</Badge>
                          <span className="text-sm">学校编码</span>
                        </div>
                        <span className="text-sm text-gray-600">第1-2位</span>
                      </div>

                      <div className="flex items-center justify-between p-2 border rounded">
                        <div className="flex items-center gap-2">
                          <Badge variant="outline">5</Badge>
                          <span className="text-sm">片区编码</span>
                        </div>
                        <span className="text-sm text-gray-600">第3位</span>
                      </div>

                      <div className="flex items-center justify-between p-2 border rounded">
                        <div className="flex items-center gap-2">
                          <Badge variant="outline">F</Badge>
                          <span className="text-sm">楼栋编码</span>
                        </div>
                        <span className="text-sm text-gray-600">第4位</span>
                      </div>

                      <div className="flex items-center justify-between p-2 border rounded">
                        <div className="flex items-center gap-2">
                          <Badge variant="outline">3D</Badge>
                          <span className="text-sm">房间编码</span>
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
                  <CardTitle className="text-amber-900">权限层级对应</CardTitle>
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
                      <Badge variant="outline">PK (学校级)</Badge>
                    </div>

                    <div className="flex items-center justify-between p-3 border rounded-lg bg-blue-50 border-blue-200">
                      <div className="flex items-center gap-3">
                        <Building className="w-5 h-5 text-blue-600" />
                        <div>
                          <div className="font-semibold">三级信使</div>
                          <div className="text-sm text-gray-600">校级负责人</div>
                        </div>
                      </div>
                      <Badge variant="outline">PK5 (片区级)</Badge>
                    </div>

                    <div className="flex items-center justify-between p-3 border rounded-lg bg-green-50 border-green-200">
                      <div className="flex items-center gap-3">
                        <MapPin className="w-5 h-5 text-green-600" />
                        <div>
                          <div className="font-semibold">二级信使</div>
                          <div className="text-sm text-gray-600">片区管理员</div>
                        </div>
                      </div>
                      <Badge variant="outline">PK5F (楼栋级)</Badge>
                    </div>

                    <div className="flex items-center justify-between p-3 border rounded-lg bg-amber-50 border-amber-200">
                      <div className="flex items-center gap-3">
                        <Users className="w-5 h-5 text-amber-600" />
                        <div>
                          <div className="font-semibold">一级信使</div>
                          <div className="text-sm text-gray-600">楼栋投递员</div>
                        </div>
                      </div>
                      <Badge variant="outline">PK5F3D (房间级)</Badge>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* 权限测试 */}
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900">权限测试</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                  {['PK5F3D', 'QH1C2E', 'PK3A1B', 'SY9Z9Z'].map((testCode) => (
                    <div key={testCode} className="p-3 border rounded-lg">
                      <div className="flex items-center justify-between mb-2">
                        <Badge variant="outline">{testCode}</Badge>
                        {hasPostcodePermission(testCode) ? (
                          <CheckCircle className="w-4 h-4 text-green-600" />
                        ) : (
                          <AlertCircle className="w-4 h-4 text-red-600" />
                        )}
                      </div>
                      <div className="text-sm text-gray-600">
                        {hasPostcodePermission(testCode) ? '有权限' : '无权限'}
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* 地址选择器测试 */}
          <TabsContent value="selector" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle className="text-amber-900">地址选择器测试</CardTitle>
                </CardHeader>
                <CardContent>
                  <AddressSelector
                    value={selectedPostcode}
                    onChange={handleAddressChange}
                    placeholder="请选择收件地址..."
                  />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-amber-900">选择结果</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium mb-2">Postcode 编码</label>
                      {selectedPostcode ? (
                        <Badge className="text-lg px-4 py-2">{selectedPostcode}</Badge>
                      ) : (
                        <div className="text-gray-500">未选择</div>
                      )}
                    </div>

                    <div>
                      <label className="block text-sm font-medium mb-2">完整地址</label>
                      {selectedAddress ? (
                        <div className="p-3 bg-gray-100 rounded-lg">{selectedAddress}</div>
                      ) : (
                        <div className="text-gray-500">未选择</div>
                      )}
                    </div>

                    {selectedPostcode && (
                      <div>
                        <label className="block text-sm font-medium mb-2">权限检查</label>
                        <div className="p-3 border rounded-lg">
                          {hasPostcodePermission(selectedPostcode) ? (
                            <div className="flex items-center gap-2 text-green-700">
                              <CheckCircle className="w-4 h-4" />
                              您有权限管理此地址
                            </div>
                          ) : (
                            <div className="flex items-center gap-2 text-red-700">
                              <AlertCircle className="w-4 h-4" />
                              您无权限管理此地址
                            </div>
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* 系统管理 (仅四级信使可见) */}
          {isPostcodeAdmin() && (
            <TabsContent value="management" className="space-y-6">
              <Alert>
                <Crown className="h-4 w-4" />
                <AlertDescription>
                  您具有 Postcode 系统管理员权限，可以管理整个地址编码体系。
                </AlertDescription>
              </Alert>
              <PostcodeManagement />
            </TabsContent>
          )}

          {/* 使用文档 */}
          <TabsContent value="documentation" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-amber-900 flex items-center gap-2">
                  <BookOpen className="w-5 h-5" />
                  使用文档
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="prose prose-sm max-w-none">
                  <h3>功能概述</h3>
                  <p>
                    OpenPenPal Postcode 编码系统是基于四级信使体系的统一地址管理平台，
                    为校园信件投递提供标准化的地址编码和权限管理功能。
                  </p>

                  <h3>主要功能</h3>
                  <ul>
                    <li><strong>地址编码</strong>：6位标准化编码，支持学校-片区-楼栋-房间四级结构</li>
                    <li><strong>权限管理</strong>：基于信使等级的分层权限控制</li>
                    <li><strong>地址选择</strong>：支持分级选择和智能搜索两种模式</li>
                    <li><strong>反馈机制</strong>：用户可申请新增地址或报告错误</li>
                  </ul>

                  <h3>使用场景</h3>
                  <ul>
                    <li><strong>写信用户</strong>：通过地址选择器选择准确的收件地址</li>
                    <li><strong>投递信使</strong>：根据编码权限接收和执行投递任务</li>
                    <li><strong>管理人员</strong>：维护地址结构，处理异常反馈</li>
                  </ul>

                  <h3>权限说明</h3>
                  <ul>
                    <li><strong>四级信使</strong>：管理整个学校的地址结构，审核地址反馈</li>
                    <li><strong>三级信使</strong>：管理指定片区，创建楼栋结构</li>
                    <li><strong>二级信使</strong>：管理指定楼栋，创建房间信息</li>
                    <li><strong>一级信使</strong>：负责具体投递，可反馈地址问题</li>
                  </ul>

                  <h3>技术支持</h3>
                  <p>
                    如遇到技术问题，请联系系统管理员或访问帮助文档获取更多信息。
                  </p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}