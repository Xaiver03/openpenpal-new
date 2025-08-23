'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { 
  MapPin, 
  Search,
  Plus,
  Eye,
  CheckCircle,
  XCircle,
  AlertCircle,
  Building2,
  School,
  Users,
  BarChart3,
  Settings,
  Download,
  Upload,
  Filter
} from 'lucide-react'
import { BackButton } from '@/components/ui/back-button'
import { Breadcrumb, ADMIN_BREADCRUMBS } from '@/components/ui/breadcrumb'
import { usePermission, PERMISSIONS } from '@/hooks/use-permission'
import { apiClient } from '@/lib/api-client-enhanced'
import { useToast } from '@/hooks/use-toast'

// 类型定义
interface OPCodeApplication {
  id: string
  user_id: string
  user_name: string
  school_code: string
  school_name: string
  requested_op_code: string
  op_code_type: 'dormitory' | 'shop' | 'box' | 'club'
  location_description: string
  status: 'pending' | 'approved' | 'rejected'
  reason?: string
  applied_at: string
  reviewed_at?: string
  reviewed_by?: string
}

interface OPCodeStats {
  total_codes: number
  active_codes: number
  pending_applications: number
  schools_count: number
  usage_by_type: Record<string, number>
  usage_by_school: Record<string, number>
}

interface OPCodeAllocation {
  school_code: string
  school_name: string
  prefix: string
  allocated_count: number
  used_count: number
  usage_rate: number
}

const APPLICATION_STATUS_COLORS = {
  pending: 'bg-yellow-100 text-yellow-800',
  approved: 'bg-green-100 text-green-800',
  rejected: 'bg-red-100 text-red-800'
}

const APPLICATION_STATUS_NAMES = {
  pending: '待审核',
  approved: '已批准',
  rejected: '已拒绝'
}

const OP_CODE_TYPES = {
  dormitory: '宿舍',
  shop: '商店',
  box: '信箱',
  club: '社团'
}

export default function OPCodeManagementPage() {
  const { user, hasPermission } = usePermission()
  const { toast } = useToast()

  // 数据状态
  const [applications, setApplications] = useState<OPCodeApplication[]>([])
  const [stats, setStats] = useState<OPCodeStats | null>(null)
  const [allocations, setAllocations] = useState<OPCodeAllocation[]>([])
  const [loading, setLoading] = useState(true)

  // 筛选状态
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [schoolFilter, setSchoolFilter] = useState<string>('all')
  const [typeFilter, setTypeFilter] = useState<string>('all')

  // 对话框状态
  const [selectedApplication, setSelectedApplication] = useState<OPCodeApplication | null>(null)
  const [showReviewDialog, setShowReviewDialog] = useState(false)
  const [reviewAction, setReviewAction] = useState<'approve' | 'reject'>('approve')
  const [reviewReason, setReviewReason] = useState('')

  // 加载数据
  useEffect(() => {
    loadData()
  }, [])

  // 权限检查
  if (!user || !hasPermission(PERMISSIONS.SYSTEM_CONFIG)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <MapPin className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">访问权限不足</h2>
            <p className="text-gray-600 mb-4">
              您没有访问OP Code管理的权限
            </p>
            <Button asChild variant="outline">
              <a href="/admin">返回管理控制台</a>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  const loadData = async () => {
    setLoading(true)
    try {
      // 并行加载所有数据
      const [applicationsRes, statsRes, allocationsRes] = await Promise.all([
        apiClient.get('/admin/opcodes/applications'),
        apiClient.get('/admin/opcodes/stats'),
        apiClient.get('/admin/opcodes/allocations')
      ])

      if (applicationsRes.data && (applicationsRes.data as any).success) {
        setApplications((applicationsRes.data as any).data.applications || [])
      }

      if (statsRes.data && (statsRes.data as any).success) {
        setStats((statsRes.data as any).data.stats || {
          total_codes: 0,
          active_codes: 0,
          pending_applications: 0,
          schools_count: 0,
          usage_by_type: {},
          usage_by_school: {}
        })
      }

      if (allocationsRes.data && (allocationsRes.data as any).success) {
        setAllocations((allocationsRes.data as any).data.allocations || [])
      }
    } catch (error) {
      console.error('Failed to load OP Code data:', error)
      
      // 设置空数据而不是mock数据
      setApplications([])
      setStats({
        total_codes: 0,
        active_codes: 0,
        pending_applications: 0,
        schools_count: 0,
        usage_by_type: {},
        usage_by_school: {}
      })
      setAllocations([])
    } finally {
      setLoading(false)
    }
  }

  // 过滤申请
  const filteredApplications = applications.filter(app => {
    const matchesSearch = app.user_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         app.requested_op_code.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         app.school_name.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesStatus = statusFilter === 'all' || app.status === statusFilter
    const matchesSchool = schoolFilter === 'all' || app.school_code === schoolFilter
    const matchesType = typeFilter === 'all' || app.op_code_type === typeFilter
    return matchesSearch && matchesStatus && matchesSchool && matchesType
  })

  // 审核申请
  const handleReviewApplication = async () => {
    if (!selectedApplication) return

    try {
      const response = await apiClient.post(`/admin/opcodes/applications/${selectedApplication.id}/review`, {
        action: reviewAction,
        reason: reviewReason
      })

      if (response.data && (response.data as any).success) {
        toast({
          title: '审核完成',
          description: `申请已${reviewAction === 'approve' ? '批准' : '拒绝'}`
        })
        
        // 更新本地状态
        setApplications(prev => prev.map(app => 
          app.id === selectedApplication.id 
            ? { ...app, status: reviewAction === 'approve' ? 'approved' : 'rejected', reason: reviewReason }
            : app
        ))
        
        setShowReviewDialog(false)
        setSelectedApplication(null)
        setReviewReason('')
      }
    } catch (error) {
      toast({
        title: '审核失败',
        description: '请稍后重试',
        variant: 'destructive'
      })
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <Breadcrumb items={[...ADMIN_BREADCRUMBS.root, { label: 'OP Code管理', href: '/admin/opcodes' }]} />
      
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <BackButton href="/admin" />
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-2">
              <MapPin className="w-8 h-8" />
              OP Code管理系统
            </h1>
            <p className="text-muted-foreground mt-1">
              管理校园OP Code编码系统和申请审核
            </p>
          </div>
        </div>
        <Button onClick={() => window.open('/admin/opcodes/export', '_blank')}>
          <Download className="w-4 h-4 mr-2" />
          导出数据
        </Button>
      </div>

      {/* 统计卡片 */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">总OP Code数</CardTitle>
              <MapPin className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total_codes}</div>
              <p className="text-xs text-muted-foreground">
                活跃 {stats.active_codes} 个
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">待审申请</CardTitle>
              <AlertCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.pending_applications}</div>
              <p className="text-xs text-muted-foreground">
                需要处理
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">覆盖学校</CardTitle>
              <School className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.schools_count}</div>
              <p className="text-xs text-muted-foreground">
                已接入学校
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">使用率</CardTitle>
              <BarChart3 className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {Math.round((stats.active_codes / stats.total_codes) * 100)}%
              </div>
              <p className="text-xs text-muted-foreground">
                系统利用率
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      <Tabs defaultValue="applications" className="space-y-6">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="applications">申请管理</TabsTrigger>
          <TabsTrigger value="allocations">分配管理</TabsTrigger>
          <TabsTrigger value="analytics">使用分析</TabsTrigger>
        </TabsList>

        {/* 申请管理 */}
        <TabsContent value="applications" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>OP Code申请审核</CardTitle>
              <CardDescription>审核用户提交的OP Code申请</CardDescription>
            </CardHeader>
            <CardContent>
              {/* 搜索和筛选 */}
              <div className="flex flex-col md:flex-row gap-4 mb-6">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input
                    placeholder="搜索申请人、OP Code或学校..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                  />
                </div>
                <Select value={statusFilter} onValueChange={setStatusFilter}>
                  <SelectTrigger className="w-32">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部状态</SelectItem>
                    <SelectItem value="pending">待审核</SelectItem>
                    <SelectItem value="approved">已批准</SelectItem>
                    <SelectItem value="rejected">已拒绝</SelectItem>
                  </SelectContent>
                </Select>
                <Select value={typeFilter} onValueChange={setTypeFilter}>
                  <SelectTrigger className="w-32">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部类型</SelectItem>
                    <SelectItem value="dormitory">宿舍</SelectItem>
                    <SelectItem value="shop">商店</SelectItem>
                    <SelectItem value="box">信箱</SelectItem>
                    <SelectItem value="club">社团</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* 申请表格 */}
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>申请信息</TableHead>
                      <TableHead>申请OP Code</TableHead>
                      <TableHead>类型</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>申请时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredApplications.map((application) => (
                      <TableRow key={application.id}>
                        <TableCell>
                          <div>
                            <div className="font-medium">{application.user_name}</div>
                            <div className="text-sm text-muted-foreground">
                              {application.school_name}
                            </div>
                            <div className="text-xs text-muted-foreground mt-1">
                              {application.location_description}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <span className="font-mono bg-blue-50 px-2 py-1 rounded text-sm text-blue-700">
                            {application.requested_op_code}
                          </span>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {OP_CODE_TYPES[application.op_code_type]}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge className={APPLICATION_STATUS_COLORS[application.status]}>
                            {APPLICATION_STATUS_NAMES[application.status]}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            {new Date(application.applied_at).toLocaleString()}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-2">
                            <Button 
                              size="sm" 
                              variant="outline"
                              onClick={() => {
                                setSelectedApplication(application)
                                setShowReviewDialog(true)
                              }}
                              disabled={application.status !== 'pending'}
                            >
                              <Eye className="w-4 h-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 分配管理 */}
        <TabsContent value="allocations" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>学校OP Code分配</CardTitle>
              <CardDescription>管理各学校的OP Code前缀分配情况</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>学校</TableHead>
                      <TableHead>OP Code前缀</TableHead>
                      <TableHead>分配数量</TableHead>
                      <TableHead>使用数量</TableHead>
                      <TableHead>使用率</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {allocations.map((allocation) => (
                      <TableRow key={allocation.school_code}>
                        <TableCell>
                          <div>
                            <div className="font-medium">{allocation.school_name}</div>
                            <div className="text-sm text-muted-foreground">
                              {allocation.school_code}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <span className="font-mono bg-green-50 px-2 py-1 rounded text-sm text-green-700">
                            {allocation.prefix}
                          </span>
                        </TableCell>
                        <TableCell>
                          <div className="font-medium">{allocation.allocated_count}</div>
                        </TableCell>
                        <TableCell>
                          <div className="font-medium">{allocation.used_count}</div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <div className={`w-16 h-2 rounded-full overflow-hidden ${
                              allocation.usage_rate > 80 ? 'bg-red-200' :
                              allocation.usage_rate > 60 ? 'bg-yellow-200' : 'bg-green-200'
                            }`}>
                              <div 
                                className={`h-full ${
                                  allocation.usage_rate > 80 ? 'bg-red-500' :
                                  allocation.usage_rate > 60 ? 'bg-yellow-500' : 'bg-green-500'
                                }`}
                                style={{ width: `${allocation.usage_rate}%` }}
                              />
                            </div>
                            <span className="text-sm font-medium">
                              {allocation.usage_rate.toFixed(1)}%
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Button size="sm" variant="outline">
                            <Settings className="w-4 h-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* 使用分析 */}
        <TabsContent value="analytics" className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle>按类型统计</CardTitle>
              </CardHeader>
              <CardContent>
                {stats && (
                  <div className="space-y-3">
                    {Object.entries(stats.usage_by_type).map(([type, count]) => (
                      <div key={type} className="flex items-center justify-between">
                        <span className="text-sm">{OP_CODE_TYPES[type as keyof typeof OP_CODE_TYPES]}</span>
                        <div className="flex items-center gap-2">
                          <div className="w-24 h-2 bg-gray-200 rounded-full overflow-hidden">
                            <div 
                              className="h-full bg-blue-500"
                              style={{ width: `${(count / stats.total_codes) * 100}%` }}
                            />
                          </div>
                          <span className="text-sm font-medium w-8">{count}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>按学校统计</CardTitle>
              </CardHeader>
              <CardContent>
                {stats && (
                  <div className="space-y-3">
                    {Object.entries(stats.usage_by_school).map(([school, count]) => (
                      <div key={school} className="flex items-center justify-between">
                        <span className="text-sm font-mono">{school}</span>
                        <div className="flex items-center gap-2">
                          <div className="w-24 h-2 bg-gray-200 rounded-full overflow-hidden">
                            <div 
                              className="h-full bg-green-500"
                              style={{ width: `${(count / stats.total_codes) * 100}%` }}
                            />
                          </div>
                          <span className="text-sm font-medium w-8">{count}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>

      {/* 审核对话框 */}
      <Dialog open={showReviewDialog} onOpenChange={setShowReviewDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>审核OP Code申请</DialogTitle>
            <DialogDescription>
              请审核 {selectedApplication?.user_name} 的OP Code申请
            </DialogDescription>
          </DialogHeader>
          
          {selectedApplication && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>申请信息</Label>
                <div className="border rounded-lg p-3 space-y-2">
                  <div className="flex justify-between">
                    <span className="text-sm text-muted-foreground">申请人:</span>
                    <span className="font-medium">{selectedApplication.user_name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-muted-foreground">学校:</span>
                    <span>{selectedApplication.school_name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-muted-foreground">申请OP Code:</span>
                    <span className="font-mono bg-blue-50 px-2 py-1 rounded text-blue-700">
                      {selectedApplication.requested_op_code}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-muted-foreground">类型:</span>
                    <span>{OP_CODE_TYPES[selectedApplication.op_code_type]}</span>
                  </div>
                  <div className="pt-2 border-t">
                    <span className="text-sm text-muted-foreground">位置描述:</span>
                    <p className="mt-1 text-sm">{selectedApplication.location_description}</p>
                  </div>
                </div>
              </div>
              
              <div className="space-y-2">
                <Label>审核操作</Label>
                <div className="flex gap-2">
                  <Button 
                    variant={reviewAction === 'approve' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setReviewAction('approve')}
                  >
                    <CheckCircle className="w-4 h-4 mr-1" />
                    批准
                  </Button>
                  <Button 
                    variant={reviewAction === 'reject' ? 'destructive' : 'outline'}
                    size="sm"
                    onClick={() => setReviewAction('reject')}
                  >
                    <XCircle className="w-4 h-4 mr-1" />
                    拒绝
                  </Button>
                </div>
              </div>

              <div className="space-y-2">
                <Label>审核备注</Label>
                <Input
                  placeholder={reviewAction === 'approve' ? '批准原因（可选）' : '拒绝原因（必填）'}
                  value={reviewReason}
                  onChange={(e) => setReviewReason(e.target.value)}
                />
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowReviewDialog(false)}>
              取消
            </Button>
            <Button 
              onClick={handleReviewApplication}
              disabled={reviewAction === 'reject' && !reviewReason.trim()}
              variant={reviewAction === 'approve' ? 'default' : 'destructive'}
            >
              确定{reviewAction === 'approve' ? '批准' : '拒绝'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}