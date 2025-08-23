'use client'

import React, { useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  Search, 
  Plus, 
  Edit, 
  Trash2, 
  CheckCircle,
  XCircle,
  Clock,
  MapPin,
  Building,
  Home,
  School,
  AlertCircle,
  RefreshCw
} from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useUserStore } from '@/stores/user-store'
import { apiClient } from '@/lib/api-client'
import { formatDate } from '@/lib/utils'
import { validateOPCodeAccess, CourierInfo } from '@/lib/utils/courier-permission-utils'
import { CourierPermissionDebug } from '@/components/debug/courier-permission-debug'

// 类型定义
interface OPCodeApplication {
  id: string
  user_id: string
  user_name: string
  school_code: string
  area_code: string
  point_type: string
  point_name: string
  full_address: string
  reason: string
  status: 'pending' | 'approved' | 'rejected'
  created_at: string
  reviewed_at?: string
  reviewer_id?: string
  assigned_code?: string
}

interface OPCode {
  id: string
  code: string
  school_code: string
  area_code: string
  point_code: string
  type: string
  name: string
  description: string
  is_active: boolean
  is_public: boolean
  created_at: string
  updated_at: string
}

export default function CourierOPCodeManagePage() {
  const router = useRouter()
  const { user } = useUserStore()
  const [activeTab, setActiveTab] = useState('applications')
  const [applications, setApplications] = useState<OPCodeApplication[]>([])
  const [opcodes, setOpcodes] = useState<OPCode[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedApplication, setSelectedApplication] = useState<OPCodeApplication | null>(null)
  const [showReviewDialog, setShowReviewDialog] = useState(false)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [reviewForm, setReviewForm] = useState({
    status: 'approved' as 'approved' | 'rejected',
    point_code: '',
    reason: ''
  })
  const [createForm, setCreateForm] = useState({
    code: '',
    type: 'dormitory',
    name: '',
    description: '',
    is_public: false
  })

  // 检查权限 - 后端角色格式是 courier_level4 而不是 courier_level_4
  const courierLevel = user?.role?.includes('courier_level') 
    ? parseInt(user.role.replace('courier_level', '')) 
    : 0


  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      // 加载申请列表
      const applicationsRes = await apiClient.get('/courier/opcode/applications')
      if (applicationsRes.data && Array.isArray(applicationsRes.data)) {
        // 根据信使级别过滤申请
        const filtered = filterApplicationsByLevel(applicationsRes.data)
        setApplications(filtered)
      }

      // 加载 OP Code 列表
      const opcodesRes = await apiClient.get('/courier/opcode/managed')
      if (opcodesRes.data && Array.isArray(opcodesRes.data)) {
        setOpcodes(opcodesRes.data)
      }
    } catch (error) {
      console.error('Failed to load data:', error)
      setApplications([])
      setOpcodes([])
    } finally {
      setLoading(false)
    }
  }, [courierLevel, user])

  // 根据信使级别过滤申请
  const filterApplicationsByLevel = (apps: OPCodeApplication[]) => {
    if (!user) return []
    
    const courierInfo: CourierInfo = {
      id: user.id,
      level: courierLevel,
      managedOPCodePrefix: user.managed_op_code_prefix,
      zoneCode: user.zone_code
    }
    
    return apps.filter(app => {
      const fullCode = app.school_code + app.area_code + '00'
      const permissions = validateOPCodeAccess(courierInfo, fullCode)
      return permissions.canView
    })
  }

  // 根据信使级别过滤 OP Code
  const filterOPCodesByLevel = (codes: OPCode[]) => {
    if (!user) return []
    
    const courierInfo: CourierInfo = {
      id: user.id,
      level: courierLevel,
      managedOPCodePrefix: user.managed_op_code_prefix,
      zoneCode: user.zone_code
    }
    
    return codes.filter(code => {
      const permissions = validateOPCodeAccess(courierInfo, code.code)
      return permissions.canView
    })
  }

  // 审核申请
  const handleReview = async () => {
    if (!selectedApplication) return

    try {
      await apiClient.post(`/courier/opcode/applications/${selectedApplication.id}/review`, {
        status: reviewForm.status,
        point_code: reviewForm.point_code,
        reason: reviewForm.reason
      })

      setShowReviewDialog(false)
      setSelectedApplication(null)
      setReviewForm({ status: 'approved', point_code: '', reason: '' })
      loadData()
    } catch (error) {
      console.error('Failed to review application:', error)
      alert('审核失败，请重试')
    }
  }

  // 创建 OP Code
  const handleCreate = async () => {
    try {
      await apiClient.post('/courier/opcode/create', createForm)
      
      setShowCreateDialog(false)
      setCreateForm({
        code: '',
        type: 'dormitory',
        name: '',
        description: '',
        is_public: false
      })
      loadData()
    } catch (error) {
      console.error('Failed to create OP Code:', error)
      alert('创建失败，请重试')
    }
  }

  // 获取权限说明
  const getPermissionDescription = () => {
    switch (courierLevel) {
      case 1:
        return '您可以查看和编辑投递任务中的投递点编码'
      case 2:
        return '您可以审核和管理投递点（后两位编码）'
      case 3:
        return '您可以管理片区和楼栋（中间两位编码）'
      case 4:
        return '您拥有全部OP Code的管理权限'
      default:
        return '权限不足'
    }
  }

  // 获取级别图标
  const getLevelIcon = () => {
    switch (courierLevel) {
      case 1: return <Home className="h-5 w-5" />
      case 2: return <Building className="h-5 w-5" />
      case 3: return <School className="h-5 w-5" />
      case 4: return <MapPin className="h-5 w-5" />
      default: return null
    }
  }

  // 过滤搜索结果
  const filteredApplications = applications.filter(app =>
    app.user_name.includes(searchTerm) ||
    app.point_name.includes(searchTerm) ||
    app.full_address.includes(searchTerm)
  )

  const filteredOPCodes = opcodes.filter(code =>
    code.code.includes(searchTerm) ||
    code.name.includes(searchTerm) ||
    code.description.includes(searchTerm)
  )

  // 调试信息
  useEffect(() => {
    if (user && courierLevel > 0) {
      console.log('Courier Debug Info:', {
        username: user.username,
        role: user.role,
        courierLevel,
        managed_op_code_prefix: user.managed_op_code_prefix,
        zone_code: user.zone_code
      })
    }
  }, [user, courierLevel])

  useEffect(() => {
    if (!user || courierLevel === 0) {
      router.push('/courier')
      return
    }
    loadData()
  }, [user, courierLevel, router, loadData])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 页面标题 */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            {getLevelIcon()}
            <div>
              <h1 className="text-3xl font-bold">OP Code 管理</h1>
              <p className="text-gray-600 mt-1">
                L{courierLevel}信使 - {getPermissionDescription()}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="px-3 py-1">
              管理范围: {courierLevel === 4 ? `${user?.managed_op_code_prefix?.substring(0,2) || 'BJ'}**** (城市级)` : 
                        courierLevel === 3 ? `${user?.managed_op_code_prefix || user?.zone_code?.substring(0,2) || 'PK'}**** (学校级)` :
                        courierLevel === 2 ? `${user?.managed_op_code_prefix || user?.zone_code?.substring(0,4) || 'PK5F'}** (片区级)` :
                        courierLevel === 1 ? `${user?.managed_op_code_prefix || user?.zone_code || 'PK5F3D'} (楼栋级)` :
                        '未设置'}
            </Badge>
            <Button onClick={loadData} variant="outline" size="sm">
              <RefreshCw className="h-4 w-4 mr-2" />
              刷新
            </Button>
          </div>
        </div>
      </div>

      {/* 搜索栏 */}
      <div className="mb-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
          <Input
            placeholder="搜索申请人、地址或编码..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
      </div>

      {/* 标签页 */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="applications">申请审核</TabsTrigger>
          <TabsTrigger value="opcodes">编码管理</TabsTrigger>
          <TabsTrigger value="create">创建编码</TabsTrigger>
        </TabsList>

        {/* 申请审核 */}
        <TabsContent value="applications">
          <Card>
            <CardHeader>
              <CardTitle>待审核申请</CardTitle>
              <CardDescription>
                审核用户提交的 OP Code 申请
              </CardDescription>
            </CardHeader>
            <CardContent>
              {filteredApplications.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  暂无待审核的申请
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>申请人</TableHead>
                      <TableHead>申请编码</TableHead>
                      <TableHead>类型</TableHead>
                      <TableHead>地址</TableHead>
                      <TableHead>申请时间</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredApplications.map((app) => (
                      <TableRow key={app.id}>
                        <TableCell>{app.user_name}</TableCell>
                        <TableCell>
                          <code className="text-sm bg-gray-100 px-2 py-1 rounded">
                            {app.school_code}{app.area_code}**
                          </code>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{app.point_type}</Badge>
                        </TableCell>
                        <TableCell>{app.point_name}</TableCell>
                        <TableCell>{formatDate(app.created_at)}</TableCell>
                        <TableCell>
                          <Badge 
                            variant={
                              app.status === 'approved' ? 'default' :
                              app.status === 'rejected' ? 'destructive' :
                              'secondary'
                            }
                          >
                            {app.status === 'approved' ? '已通过' :
                             app.status === 'rejected' ? '已拒绝' :
                             '待审核'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {app.status === 'pending' && (
                            <Button
                              size="sm"
                              onClick={() => {
                                setSelectedApplication(app)
                                setShowReviewDialog(true)
                              }}
                              disabled={courierLevel < 2}
                            >
                              {courierLevel < 2 ? '权限不足' : '审核'}
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* 编码管理 */}
        <TabsContent value="opcodes">
          <Card>
            <CardHeader>
              <CardTitle>已分配编码</CardTitle>
              <CardDescription>
                管理您负责区域的 OP Code
              </CardDescription>
            </CardHeader>
            <CardContent>
              {filteredOPCodes.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  暂无已分配的编码
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>编码</TableHead>
                      <TableHead>名称</TableHead>
                      <TableHead>类型</TableHead>
                      <TableHead>描述</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredOPCodes.map((code) => {
                      const courierInfo: CourierInfo = {
                        id: user?.id || '',
                        level: courierLevel,
                        managedOPCodePrefix: user?.managed_op_code_prefix,
                        zoneCode: user?.zone_code
                      }
                      const permissions = validateOPCodeAccess(courierInfo, code.code)
                      return (
                        <TableRow key={code.id}>
                          <TableCell>
                            <code className="text-sm bg-gray-100 px-2 py-1 rounded">
                              {code.code}
                            </code>
                          </TableCell>
                          <TableCell>{code.name}</TableCell>
                          <TableCell>
                            <Badge variant="outline">{code.type}</Badge>
                          </TableCell>
                          <TableCell>{code.description}</TableCell>
                          <TableCell>
                            <Badge variant={code.is_active ? 'default' : 'secondary'}>
                              {code.is_active ? '启用' : '停用'}
                            </Badge>
                          </TableCell>
                          <TableCell>{formatDate(code.created_at)}</TableCell>
                          <TableCell>
                            <div className="flex gap-2">
                              {permissions.canEdit && (
                                <Button size="sm" variant="outline">
                                  <Edit className="h-4 w-4" />
                                </Button>
                              )}
                              {permissions.canDelete && (
                                <Button size="sm" variant="outline" className="text-red-600">
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              )}
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* 创建编码 */}
        <TabsContent value="create">
          <Card>
            <CardHeader>
              <CardTitle>创建新编码</CardTitle>
              <CardDescription>
                {courierLevel < 2 ? '您没有创建编码的权限' :
                 courierLevel === 2 ? '创建新的投递点编码 (需遵循片区前缀)' :
                 courierLevel === 3 ? '创建新的OP Code (需遵循学校前缀)' :
                 '创建新的OP Code (可创建城市内任意编码)'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {courierLevel < 2 ? (
                <Alert>
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>
                    一级信使没有创建编码的权限，但可以编辑现有投递点信息
                  </AlertDescription>
                </Alert>
              ) : (
                <div className="space-y-4 max-w-md">
                  <div>
                    <Label>编码</Label>
                    <Input
                      value={createForm.code}
                      onChange={(e) => {
                        const value = e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, '')
                        setCreateForm(prev => ({ ...prev, code: value }))
                      }}
                      placeholder={
                        courierLevel === 4 ? '输入完整6位编码 (前2位为城市代码)' :
                        courierLevel === 3 ? '输入完整6位编码 (前2位固定为学校)' :
                        '输入完整6位编码 (前4位固定为片区)'
                      }
                      maxLength={6}
                    />
                    <p className="text-sm text-gray-500 mt-1">
                      {courierLevel === 4 ? `AA=城市码(${user?.managed_op_code_prefix?.substring(0,2) || 'BJ'}), BB=学校码, CC=区域码` :
                       courierLevel === 3 ? `您可以在${user?.managed_op_code_prefix || user?.zone_code?.substring(0,2) || 'PK'}****范围内创建` :
                       `您可以在${user?.managed_op_code_prefix || user?.zone_code?.substring(0,4) || 'PK5F'}**范围内创建`}
                    </p>
                    {createForm.code && createForm.code.length === 6 && (
                      <div className="text-sm text-green-600 mt-1">
                        ✓ 编码格式正确: {createForm.code.substring(0,2)} ({courierLevel === 4 ? '城市' : '学校'}) + {createForm.code.substring(2,4)} ({courierLevel === 4 ? '学校' : '区域'}) + {createForm.code.substring(4,6)} ({courierLevel === 4 ? '区域' : '位置'})
                      </div>
                    )}
                    {createForm.code && createForm.code.length > 0 && createForm.code.length < 6 && (
                      <div className="text-sm text-amber-600 mt-1">
                        请输入完整的6位编码 (当前: {createForm.code.length}/6)
                      </div>
                    )}
                  </div>

                  <div>
                    <Label>类型</Label>
                    <Select
                      value={createForm.type}
                      onValueChange={(value) => setCreateForm(prev => ({ ...prev, type: value }))}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="dormitory">宿舍</SelectItem>
                        <SelectItem value="shop">商店</SelectItem>
                        <SelectItem value="box">投递箱</SelectItem>
                        <SelectItem value="club">社团</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div>
                    <Label>名称</Label>
                    <Input
                      value={createForm.name}
                      onChange={(e) => setCreateForm(prev => ({ ...prev, name: e.target.value }))}
                      placeholder="输入名称"
                    />
                  </div>

                  <div>
                    <Label>描述</Label>
                    <Textarea
                      value={createForm.description}
                      onChange={(e) => setCreateForm(prev => ({ ...prev, description: e.target.value }))}
                      placeholder="输入描述信息"
                      rows={3}
                    />
                  </div>

                  <div className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      id="is_public"
                      checked={createForm.is_public}
                      onChange={(e) => setCreateForm(prev => ({ ...prev, is_public: e.target.checked }))}
                    />
                    <Label htmlFor="is_public">设为公开投递点</Label>
                  </div>

                  <Button
                    onClick={() => setShowCreateDialog(true)}
                    className="w-full"
                    disabled={!createForm.code || !createForm.name || createForm.code.length !== 6}
                  >
                    <Plus className="h-4 w-4 mr-2" />
                    创建编码
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* 审核对话框 */}
      <Dialog open={showReviewDialog} onOpenChange={setShowReviewDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>审核申请</DialogTitle>
            <DialogDescription>
              审核 {selectedApplication?.user_name} 的 OP Code 申请
            </DialogDescription>
          </DialogHeader>

          {selectedApplication && (
            <div className="space-y-4">
              <div>
                <Label>申请信息</Label>
                <div className="mt-2 p-3 bg-gray-50 rounded-md text-sm space-y-1">
                  <div>申请人：{selectedApplication.user_name}</div>
                  <div>申请编码前缀：{selectedApplication.school_code}{selectedApplication.area_code}</div>
                  <div>类型：{selectedApplication.point_type}</div>
                  <div>地址：{selectedApplication.full_address}</div>
                  <div>理由：{selectedApplication.reason}</div>
                </div>
              </div>

              <div>
                <Label>审核决定</Label>
                <Select
                  value={reviewForm.status}
                  onValueChange={(value: 'approved' | 'rejected') => 
                    setReviewForm(prev => ({ ...prev, status: value }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="approved">通过</SelectItem>
                    <SelectItem value="rejected">拒绝</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {reviewForm.status === 'approved' && (
                <div>
                  <Label>分配后两位编码</Label>
                  <Input
                    value={reviewForm.point_code}
                    onChange={(e) => setReviewForm(prev => ({ ...prev, point_code: e.target.value }))}
                    placeholder="输入两位编码，如: 01"
                    maxLength={2}
                  />
                </div>
              )}

              <div>
                <Label>审核备注</Label>
                <Textarea
                  value={reviewForm.reason}
                  onChange={(e) => setReviewForm(prev => ({ ...prev, reason: e.target.value }))}
                  placeholder="请输入审核意见..."
                  rows={3}
                />
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowReviewDialog(false)}>
              取消
            </Button>
            <Button
              onClick={handleReview}
              className={reviewForm.status === 'approved' ? 'bg-green-600 hover:bg-green-700' : 'bg-red-600 hover:bg-red-700'}
            >
              {reviewForm.status === 'approved' ? (
                <>
                  <CheckCircle className="h-4 w-4 mr-2" />
                  通过申请
                </>
              ) : (
                <>
                  <XCircle className="h-4 w-4 mr-2" />
                  拒绝申请
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 创建确认对话框 */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认创建</DialogTitle>
            <DialogDescription>
              确认创建新的 OP Code
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-2 text-sm">
            <div>编码：{createForm.code}</div>
            <div>类型：{createForm.type}</div>
            <div>名称：{createForm.name}</div>
            <div>描述：{createForm.description}</div>
            <div>公开状态：{createForm.is_public ? '公开' : '私有'}</div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
              取消
            </Button>
            <Button onClick={handleCreate}>
              确认创建
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      
      {/* 开发环境调试组件 */}
      <CourierPermissionDebug />
    </div>
  )
}