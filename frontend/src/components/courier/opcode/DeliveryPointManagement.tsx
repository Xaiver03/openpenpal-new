'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { 
  Package, 
  Search,
  Filter,
  Download,
  Upload,
  CheckCircle,
  XCircle,
  AlertCircle,
  Home,
  Building,
  MapPin,
  User,
  Calendar,
  ToggleLeft,
  ToggleRight
} from 'lucide-react'
import { apiClient } from '@/lib/api-client-enhanced'
import { useToast } from '@/hooks/use-toast'

interface DeliveryPoint {
  id: string
  op_code: string
  school_code: string
  school_name: string
  area_code: string
  area_name: string
  building_code: string
  building_name: string
  point_code: string
  point_name: string
  point_type: string
  is_active: boolean
  is_occupied: boolean
  occupant_id?: string
  occupant_name?: string
  occupied_at?: string
  created_at: string
  updated_at: string
}

interface DeliveryPointManagementProps {
  courierInfo: any
}

export function DeliveryPointManagement({ courierInfo }: DeliveryPointManagementProps) {
  const { toast } = useToast()
  const [deliveryPoints, setDeliveryPoints] = useState<DeliveryPoint[]>([])
  const [buildings, setBuildings] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [filterBuilding, setFilterBuilding] = useState('')
  const [filterStatus, setFilterStatus] = useState<'all' | 'available' | 'occupied'>('all')
  const [selectedPoints, setSelectedPoints] = useState<string[]>([])
  const [showBatchDialog, setShowBatchDialog] = useState(false)
  const [batchAction, setBatchAction] = useState<'activate' | 'deactivate' | 'generate'>('activate')

  // 批量生成表单
  const [generateForm, setGenerateForm] = useState({
    building_prefix: '',
    start_floor: 1,
    end_floor: 6,
    rooms_per_floor: 20,
    room_format: 'floor_room' // floor_room: 101, 102... | sequential: 001, 002...
  })

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    try {
      // 获取管理范围内的楼栋
      const buildingsResponse = await apiClient.get('/api/v1/opcode/buildings/managed')
      if ((buildingsResponse.data as any).success) {
        setBuildings((buildingsResponse.data as any).data.buildings || [])
      }

      // 获取投递点列表
      const prefix = courierInfo.managedOPCodePrefix || ''
      const pointsResponse = await apiClient.get(`/api/v1/opcode/delivery-points/${prefix}`)
      if ((pointsResponse.data as any).success) {
        setDeliveryPoints((pointsResponse.data as any).data.points || [])
      }
    } catch (error) {
      console.error('Failed to fetch data:', error)
      // 使用模拟数据
      setBuildings([
        { 
          school_code: 'CS', 
          school_name: '中南大学', 
          area_code: '01', 
          area_name: '本部东区',
          building_code: 'A',
          building_name: 'A栋',
          prefix: 'CS01A'
        }
      ])
      
      // 生成模拟投递点数据
      const mockPoints: DeliveryPoint[] = []
      for (let floor = 1; floor <= 6; floor++) {
        for (let room = 1; room <= 10; room++) {
          const pointCode = `${floor}${room.toString().padStart(2, '0')}`
          mockPoints.push({
            id: `${floor}-${room}`,
            op_code: `CS01A${pointCode.slice(-2)}`,
            school_code: 'CS',
            school_name: '中南大学',
            area_code: '01',
            area_name: '本部东区',
            building_code: 'A',
            building_name: 'A栋',
            point_code: pointCode.slice(-2),
            point_name: `${floor}${room.toString().padStart(2, '0')}室`,
            point_type: 'dormitory',
            is_active: true,
            is_occupied: Math.random() > 0.4,
            occupant_name: Math.random() > 0.4 ? `用户${Math.floor(Math.random() * 1000)}` : undefined,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          })
        }
      }
      setDeliveryPoints(mockPoints)
    } finally {
      setLoading(false)
    }
  }

  // 过滤投递点
  const filteredPoints = deliveryPoints.filter(point => {
    // 搜索过滤
    if (searchTerm && !point.op_code.toLowerCase().includes(searchTerm.toLowerCase()) &&
        !point.point_name.toLowerCase().includes(searchTerm.toLowerCase())) {
      return false
    }
    
    // 楼栋过滤
    if (filterBuilding && !point.op_code.startsWith(filterBuilding)) {
      return false
    }
    
    // 状态过滤
    if (filterStatus === 'available' && point.is_occupied) return false
    if (filterStatus === 'occupied' && !point.is_occupied) return false
    
    return true
  })

  // 检查是否有编辑权限
  const canEdit = (point?: DeliveryPoint) => {
    // L2及以上信使可以编辑所有投递点
    if (courierInfo.level >= 2) {
      return true
    }
    
    // L1信使只能编辑自己管理范围内的投递点
    if (courierInfo.level === 1 && point) {
      const managedPrefix = courierInfo.managedOPCodePrefix || ''
      // 检查投递点是否在L1信使的管理范围内（前4位匹配）
      return point.op_code.startsWith(managedPrefix.substring(0, 4))
    }
    
    return false
  }
  
  // 检查是否可以批量操作
  const canBatchEdit = () => {
    // 只有L2及以上信使可以批量操作
    return courierInfo.level >= 2
  }

  // 切换投递点状态
  const togglePointStatus = async (point: DeliveryPoint) => {
    if (!canEdit(point)) {
      toast({
        title: '权限不足',
        description: '您没有修改此投递点的权限',
        variant: 'destructive'
      })
      return
    }

    try {
      const response = await apiClient.patch(`/api/v1/opcode/delivery-points/${point.id}/status`, {
        is_active: !point.is_active
      })

      if ((response.data as any).success) {
        toast({
          title: '状态更新成功',
          description: `投递点已${!point.is_active ? '激活' : '停用'}`
        })
        fetchData()
      }
    } catch (error) {
      toast({
        title: '更新失败',
        description: '请稍后重试',
        variant: 'destructive'
      })
    }
  }

  // 批量操作
  const handleBatchAction = async () => {
    if (batchAction === 'generate') {
      await handleBatchGenerate()
    } else {
      await handleBatchStatusChange()
    }
  }

  // 批量生成投递点
  const handleBatchGenerate = async () => {
    if (!generateForm.building_prefix) {
      toast({
        title: '请选择楼栋',
        variant: 'destructive'
      })
      return
    }

    try {
      const points = []
      for (let floor = generateForm.start_floor; floor <= generateForm.end_floor; floor++) {
        for (let room = 1; room <= generateForm.rooms_per_floor; room++) {
          let pointCode = ''
          if (generateForm.room_format === 'floor_room') {
            pointCode = `${floor}${room.toString().padStart(2, '0')}`.slice(-2)
          } else {
            const sequential = (floor - generateForm.start_floor) * generateForm.rooms_per_floor + room
            pointCode = sequential.toString().padStart(2, '0')
          }
          
          points.push({
            point_code: pointCode,
            point_name: `${floor}${room.toString().padStart(2, '0')}室`,
            point_type: 'dormitory'
          })
        }
      }

      const response = await apiClient.post('/api/v1/opcode/delivery-points/batch', {
        building_prefix: generateForm.building_prefix,
        points
      })

      if ((response.data as any).success) {
        toast({
          title: '批量生成成功',
          description: `成功生成 ${points.length} 个投递点`
        })
        setShowBatchDialog(false)
        fetchData()
      }
    } catch (error) {
      toast({
        title: '生成失败',
        description: '请稍后重试',
        variant: 'destructive'
      })
    }
  }

  // 批量状态更改
  const handleBatchStatusChange = async () => {
    if (selectedPoints.length === 0) {
      toast({
        title: '请选择投递点',
        variant: 'destructive'
      })
      return
    }

    try {
      const response = await apiClient.patch('/api/v1/opcode/delivery-points/batch/status', {
        ids: selectedPoints,
        is_active: batchAction === 'activate'
      })

      if ((response.data as any).success) {
        toast({
          title: '批量操作成功',
          description: `成功${batchAction === 'activate' ? '激活' : '停用'} ${selectedPoints.length} 个投递点`
        })
        setSelectedPoints([])
        setShowBatchDialog(false)
        fetchData()
      }
    } catch (error) {
      toast({
        title: '操作失败',
        description: '请稍后重试',
        variant: 'destructive'
      })
    }
  }

  // 导出数据
  const handleExport = () => {
    const data = filteredPoints.map(point => ({
      'OP Code': point.op_code,
      '学校': point.school_name,
      '片区': point.area_name,
      '楼栋': point.building_name,
      '投递点': point.point_name,
      '状态': point.is_active ? '启用' : '停用',
      '占用情况': point.is_occupied ? '已占用' : '可用',
      '占用人': point.occupant_name || '-'
    }))

    const csv = [
      Object.keys(data[0]).join(','),
      ...data.map(row => Object.values(row).join(','))
    ].join('\n')

    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = `delivery_points_${new Date().toISOString().split('T')[0]}.csv`
    link.click()
  }

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-amber-600"></div>
      </div>
    )
  }

  const stats = {
    total: filteredPoints.length,
    active: filteredPoints.filter(p => p.is_active).length,
    occupied: filteredPoints.filter(p => p.is_occupied).length,
    available: filteredPoints.filter(p => p.is_active && !p.is_occupied).length
  }

  return (
    <div className="space-y-6">
      {/* 统计信息 */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">总数</p>
                <p className="text-2xl font-bold">{stats.total}</p>
              </div>
              <Package className="h-8 w-8 text-gray-400" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">已激活</p>
                <p className="text-2xl font-bold text-green-600">{stats.active}</p>
              </div>
              <CheckCircle className="h-8 w-8 text-green-400" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">已占用</p>
                <p className="text-2xl font-bold text-amber-600">{stats.occupied}</p>
              </div>
              <User className="h-8 w-8 text-amber-400" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">可用</p>
                <p className="text-2xl font-bold text-blue-600">{stats.available}</p>
              </div>
              <Home className="h-8 w-8 text-blue-400" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 搜索和筛选 */}
      <Card>
        <CardHeader>
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <CardTitle>投递点列表</CardTitle>
            <div className="flex flex-wrap gap-2">
              {canBatchEdit() && (
                <>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setBatchAction('generate')
                      setShowBatchDialog(true)
                    }}
                  >
                    <Upload className="w-4 h-4 mr-2" />
                    批量生成
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setBatchAction('activate')
                      setShowBatchDialog(true)
                    }}
                    disabled={selectedPoints.length === 0}
                  >
                    批量激活
                  </Button>
                </>
              )}
              <Button
                variant="outline"
                size="sm"
                onClick={handleExport}
              >
                <Download className="w-4 h-4 mr-2" />
                导出
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col md:flex-row gap-4 mb-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                <Input
                  placeholder="搜索OP Code或投递点名称..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
            <Select value={filterBuilding} onValueChange={setFilterBuilding}>
              <SelectTrigger className="w-48">
                <SelectValue placeholder="选择楼栋" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">全部楼栋</SelectItem>
                {buildings.map(building => (
                  <SelectItem 
                    key={building.prefix} 
                    value={building.prefix}
                  >
                    {building.building_name} [{building.prefix}]
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={filterStatus} onValueChange={(value: any) => setFilterStatus(value)}>
              <SelectTrigger className="w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="available">可用</SelectItem>
                <SelectItem value="occupied">已占用</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* 投递点网格 */}
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-3">
            {filteredPoints.map((point) => (
              <Card 
                key={point.id}
                className={`relative ${!point.is_active ? 'opacity-60' : ''}`}
              >
                <CardContent className="p-3">
                  {canBatchEdit() && (
                    <Checkbox
                      checked={selectedPoints.includes(point.id)}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          setSelectedPoints([...selectedPoints, point.id])
                        } else {
                          setSelectedPoints(selectedPoints.filter(id => id !== point.id))
                        }
                      }}
                      className="absolute top-2 left-2"
                    />
                  )}
                  
                  <div className="text-center space-y-2">
                    <Badge className="font-mono text-xs">
                      {point.op_code}
                    </Badge>
                    <div className="font-medium">{point.point_name}</div>
                    <div className="flex items-center justify-center gap-2">
                      {point.is_occupied ? (
                        <Badge variant="secondary" className="text-xs">
                          <User className="w-3 h-3 mr-1" />
                          已占用
                        </Badge>
                      ) : (
                        <Badge variant="outline" className="text-xs text-green-600">
                          <CheckCircle className="w-3 h-3 mr-1" />
                          可用
                        </Badge>
                      )}
                    </div>
                    {canEdit(point) && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => togglePointStatus(point)}
                        className="w-full"
                      >
                        {point.is_active ? (
                          <>
                            <ToggleRight className="w-4 h-4 mr-1" />
                            停用
                          </>
                        ) : (
                          <>
                            <ToggleLeft className="w-4 h-4 mr-1" />
                            激活
                          </>
                        )}
                      </Button>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* 批量操作对话框 */}
      <Dialog open={showBatchDialog} onOpenChange={setShowBatchDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>
              {batchAction === 'generate' ? '批量生成投递点' : '批量状态更改'}
            </DialogTitle>
          </DialogHeader>
          
          {batchAction === 'generate' ? (
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label>选择楼栋</Label>
                <Select
                  value={generateForm.building_prefix}
                  onValueChange={(value) => setGenerateForm({ ...generateForm, building_prefix: value })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="请选择楼栋" />
                  </SelectTrigger>
                  <SelectContent>
                    {buildings.map(building => (
                      <SelectItem key={building.prefix} value={building.prefix}>
                        [{building.prefix}] {building.building_name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>起始楼层</Label>
                  <Input
                    type="number"
                    value={generateForm.start_floor}
                    onChange={(e) => setGenerateForm({ 
                      ...generateForm, 
                      start_floor: parseInt(e.target.value) || 1 
                    })}
                    min={1}
                  />
                </div>
                <div className="space-y-2">
                  <Label>结束楼层</Label>
                  <Input
                    type="number"
                    value={generateForm.end_floor}
                    onChange={(e) => setGenerateForm({ 
                      ...generateForm, 
                      end_floor: parseInt(e.target.value) || 1 
                    })}
                    min={generateForm.start_floor}
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label>每层房间数</Label>
                <Input
                  type="number"
                  value={generateForm.rooms_per_floor}
                  onChange={(e) => setGenerateForm({ 
                    ...generateForm, 
                    rooms_per_floor: parseInt(e.target.value) || 1 
                  })}
                  min={1}
                  max={99}
                />
              </div>
              <Alert>
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  将生成 {(generateForm.end_floor - generateForm.start_floor + 1) * generateForm.rooms_per_floor} 个投递点
                </AlertDescription>
              </Alert>
            </div>
          ) : (
            <div className="py-4">
              <Alert>
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  确定要{batchAction === 'activate' ? '激活' : '停用'}选中的 {selectedPoints.length} 个投递点吗？
                </AlertDescription>
              </Alert>
            </div>
          )}
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBatchDialog(false)}>
              取消
            </Button>
            <Button onClick={handleBatchAction}>
              确定
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 权限提示 */}
      {courierInfo.level === 1 && (
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            作为一级信使，您可以管理自己负责楼栋（{courierInfo.managedOPCodePrefix?.substring(0, 4) || '未设置'}**）内的投递点。
            如需批量操作或管理其他楼栋，请联系您的上级信使。
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}