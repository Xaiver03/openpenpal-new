'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { 
  Plus, 
  Edit2, 
  Trash2, 
  Home,
  Building,
  MapPin,
  AlertCircle,
  Check,
  X,
  Package
} from 'lucide-react'
import { apiClient } from '@/lib/api-client-enhanced'
import { useToast } from '@/hooks/use-toast'

interface Building {
  id: string
  school_code: string
  school_name: string
  area_code: string
  area_name: string
  building_code: string
  building_name: string
  building_type: string
  floor_count?: number
  room_count?: number
  description?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

interface BuildingManagementProps {
  courierInfo: any
}

const buildingTypes = [
  { value: 'dormitory', label: '宿舍楼', icon: Home },
  { value: 'teaching', label: '教学楼', icon: Building },
  { value: 'dining', label: '食堂', icon: Package },
  { value: 'library', label: '图书馆', icon: Building },
  { value: 'office', label: '办公楼', icon: Building },
  { value: 'other', label: '其他', icon: MapPin }
]

export function BuildingManagement({ courierInfo }: BuildingManagementProps) {
  const { toast } = useToast()
  const [buildings, setBuildings] = useState<Building[]>([])
  const [districts, setDistricts] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [showAddDialog, setShowAddDialog] = useState(false)
  
  // 表单状态
  const [formData, setFormData] = useState({
    school_code: '',
    area_code: '',
    building_code: '',
    building_name: '',
    building_type: 'dormitory',
    floor_count: 6,
    room_count: 20,
    description: ''
  })

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    try {
      // 根据信使级别获取管理范围
      let prefix = ''
      if (courierInfo.level === 2) {
        // L2信使只能管理特定片区
        prefix = courierInfo.managedOPCodePrefix || ''
      } else if (courierInfo.level === 3) {
        // L3信使管理整个学校
        prefix = (courierInfo.managedOPCodePrefix || '').slice(0, 2)
      }
      // L4信使管理所有

      // 获取片区列表
      const districtsResponse = await apiClient.get(`/api/v1/opcode/areas/managed`)
      if ((districtsResponse.data as any).success) {
        setDistricts((districtsResponse.data as any).data.areas || [])
      }

      // 获取楼栋列表
      const buildingsResponse = await apiClient.get(`/api/v1/opcode/buildings/managed`)
      if ((buildingsResponse.data as any).success) {
        setBuildings((buildingsResponse.data as any).data.buildings || [])
      }
    } catch (error) {
      console.error('Failed to fetch data:', error)
      // 使用模拟数据
      setDistricts([
        { school_code: 'CS', school_name: '中南大学', area_code: '01', area_name: '本部东区' },
        { school_code: 'CS', school_name: '中南大学', area_code: '02', area_name: '本部西区' }
      ])
      setBuildings([
        {
          id: '1',
          school_code: 'CS',
          school_name: '中南大学',
          area_code: '01',
          area_name: '本部东区',
          building_code: 'A',
          building_name: 'A栋',
          building_type: 'dormitory',
          floor_count: 6,
          room_count: 120,
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        },
        {
          id: '2',
          school_code: 'CS',
          school_name: '中南大学',
          area_code: '01',
          area_name: '本部东区',
          building_code: 'B',
          building_name: 'B栋',
          building_type: 'dormitory',
          floor_count: 7,
          room_count: 140,
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }
      ])
    } finally {
      setLoading(false)
    }
  }

  const handleAdd = async () => {
    if (!formData.area_code || !formData.building_code || !formData.building_name) {
      toast({
        title: '请填写必要字段',
        description: '片区、楼栋代码和名称为必填项',
        variant: 'destructive'
      })
      return
    }

    // 验证楼栋代码格式（1位字母或数字）
    if (!/^[A-Z0-9]$/i.test(formData.building_code)) {
      toast({
        title: '楼栋代码格式错误',
        description: '楼栋代码必须是1位字母或数字',
        variant: 'destructive'
      })
      return
    }

    try {
      // 从选中的片区获取学校代码
      const selectedDistrict = districts.find(d => 
        `${d.school_code}${d.area_code}` === formData.area_code
      )
      
      const response = await apiClient.post('/api/v1/opcode/buildings', {
        school_code: selectedDistrict?.school_code,
        area_code: selectedDistrict?.area_code,
        building_code: formData.building_code.toUpperCase(),
        building_name: formData.building_name,
        building_type: formData.building_type,
        floor_count: formData.floor_count,
        room_count: formData.room_count,
        description: formData.description
      })

      if ((response.data as any).success) {
        toast({
          title: '添加成功',
          description: '楼栋已成功创建'
        })
        setShowAddDialog(false)
        setFormData({
          school_code: '',
          area_code: '',
          building_code: '',
          building_name: '',
          building_type: 'dormitory',
          floor_count: 6,
          room_count: 20,
          description: ''
        })
        fetchData()
      }
    } catch (error) {
      toast({
        title: '添加失败',
        description: '请稍后重试',
        variant: 'destructive'
      })
    }
  }

  const handleUpdate = async (building: Building) => {
    try {
      const response = await apiClient.put(`/api/v1/opcode/buildings/${building.id}`, {
        building_name: building.building_name,
        building_type: building.building_type,
        floor_count: building.floor_count,
        room_count: building.room_count,
        description: building.description,
        is_active: building.is_active
      })

      if ((response.data as any).success) {
        toast({
          title: '更新成功',
          description: '楼栋信息已更新'
        })
        setEditingId(null)
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

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除这个楼栋吗？删除后该楼栋下的所有投递点将失效。')) {
      return
    }

    try {
      const response = await apiClient.delete(`/api/v1/opcode/buildings/${id}`)
      if ((response.data as any).success) {
        toast({
          title: '删除成功',
          description: '楼栋已删除'
        })
        fetchData()
      }
    } catch (error) {
      toast({
        title: '删除失败',
        description: '请稍后重试',
        variant: 'destructive'
      })
    }
  }

  const updateBuildingField = (id: string, field: string, value: any) => {
    setBuildings(buildings.map(b => 
      b.id === id ? { ...b, [field]: value } : b
    ))
  }

  const getBuildingTypeIcon = (type: string) => {
    const typeConfig = buildingTypes.find(t => t.value === type)
    const Icon = typeConfig?.icon || MapPin
    return <Icon className="w-4 h-4" />
  }

  const getBuildingTypeLabel = (type: string) => {
    const typeConfig = buildingTypes.find(t => t.value === type)
    return typeConfig?.label || type
  }

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-amber-600"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 操作栏 */}
      <div className="flex justify-between items-center">
        <div>
          <h3 className="text-lg font-semibold">楼栋管理</h3>
          <p className="text-sm text-muted-foreground">
            管理片区内的楼栋，每个楼栋对应OP Code的第5位
          </p>
        </div>
        <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="w-4 h-4 mr-2" />
              添加楼栋
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>添加新楼栋</DialogTitle>
              <DialogDescription>
                为片区添加新的楼栋
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label>选择片区</Label>
                <Select
                  value={formData.area_code}
                  onValueChange={(value) => setFormData({ ...formData, area_code: value })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="请选择片区" />
                  </SelectTrigger>
                  <SelectContent>
                    {districts.map(district => (
                      <SelectItem 
                        key={`${district.school_code}${district.area_code}`} 
                        value={`${district.school_code}${district.area_code}`}
                      >
                        [{district.school_code}{district.area_code}] {district.school_name} - {district.area_name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>楼栋代码（1位）</Label>
                <Input
                  placeholder="如：A, B, 1, 2"
                  value={formData.building_code}
                  onChange={(e) => setFormData({ ...formData, building_code: e.target.value.toUpperCase() })}
                  maxLength={1}
                />
              </div>
              <div className="space-y-2">
                <Label>楼栋名称</Label>
                <Input
                  placeholder="如：A栋、1号楼"
                  value={formData.building_name}
                  onChange={(e) => setFormData({ ...formData, building_name: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label>楼栋类型</Label>
                <Select
                  value={formData.building_type}
                  onValueChange={(value) => setFormData({ ...formData, building_type: value })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {buildingTypes.map(type => (
                      <SelectItem key={type.value} value={type.value}>
                        <div className="flex items-center gap-2">
                          {getBuildingTypeIcon(type.value)}
                          <span>{type.label}</span>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>楼层数</Label>
                  <Input
                    type="number"
                    value={formData.floor_count}
                    onChange={(e) => setFormData({ ...formData, floor_count: parseInt(e.target.value) || 1 })}
                    min={1}
                    max={50}
                  />
                </div>
                <div className="space-y-2">
                  <Label>每层房间数</Label>
                  <Input
                    type="number"
                    value={formData.room_count}
                    onChange={(e) => setFormData({ ...formData, room_count: parseInt(e.target.value) || 1 })}
                    min={1}
                    max={100}
                  />
                </div>
              </div>
              <div className="space-y-2">
                <Label>描述（选填）</Label>
                <Input
                  placeholder="如：男生宿舍楼"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setShowAddDialog(false)}>
                取消
              </Button>
              <Button onClick={handleAdd}>
                确定添加
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* 楼栋列表 */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>学校/片区</TableHead>
                <TableHead>楼栋代码</TableHead>
                <TableHead>楼栋名称</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>规模</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {buildings.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                    暂无楼栋数据
                  </TableCell>
                </TableRow>
              ) : (
                buildings.map((building) => (
                  <TableRow key={building.id}>
                    <TableCell>
                      <div>
                        <div className="font-medium">{building.school_name}</div>
                        <div className="text-sm text-muted-foreground">{building.area_name}</div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge className="font-mono">
                        {building.school_code}{building.area_code}{building.building_code}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {editingId === building.id ? (
                        <Input
                          value={building.building_name}
                          onChange={(e) => updateBuildingField(building.id, 'building_name', e.target.value)}
                          className="w-24"
                        />
                      ) : (
                        building.building_name
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {getBuildingTypeIcon(building.building_type)}
                        <span className="text-sm">
                          {getBuildingTypeLabel(building.building_type)}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">
                        <div>{building.floor_count}层</div>
                        <div className="text-muted-foreground">
                          约{(building.floor_count || 0) * (building.room_count || 0)}个投递点
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      {editingId === building.id ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => updateBuildingField(building.id, 'is_active', !building.is_active)}
                        >
                          <Badge variant={building.is_active ? 'default' : 'secondary'}>
                            {building.is_active ? '启用' : '停用'}
                          </Badge>
                        </Button>
                      ) : (
                        <Badge variant={building.is_active ? 'default' : 'secondary'}>
                          {building.is_active ? '启用' : '停用'}
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {editingId === building.id ? (
                          <>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleUpdate(building)}
                            >
                              <Check className="h-4 w-4 text-green-600" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => setEditingId(null)}
                            >
                              <X className="h-4 w-4 text-red-600" />
                            </Button>
                          </>
                        ) : (
                          <>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => setEditingId(building.id)}
                            >
                              <Edit2 className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDelete(building.id)}
                            >
                              <Trash2 className="h-4 w-4 text-red-600" />
                            </Button>
                          </>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* 编码说明 */}
      <Alert>
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          楼栋代码将成为OP Code的第5位。例如：学校CS + 片区01 + 楼栋A = CS01A*（前5位确定）
        </AlertDescription>
      </Alert>
    </div>
  )
}