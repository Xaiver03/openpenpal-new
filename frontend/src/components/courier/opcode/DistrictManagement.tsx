'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
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
  Save,
  X,
  Building2,
  MapPin,
  AlertCircle,
  Check
} from 'lucide-react'
import { apiClient } from '@/lib/api-client-enhanced'
import { useToast } from '@/hooks/use-toast'

interface District {
  id: string
  school_code: string
  school_name: string
  area_code: string
  area_name: string
  description?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

interface DistrictManagementProps {
  courierInfo: any
}

export function DistrictManagement({ courierInfo }: DistrictManagementProps) {
  const { toast } = useToast()
  const [districts, setDistricts] = useState<District[]>([])
  const [schools, setSchools] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [showAddDialog, setShowAddDialog] = useState(false)
  
  // 表单状态
  const [formData, setFormData] = useState({
    school_code: '',
    area_code: '',
    area_name: '',
    description: ''
  })

  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    try {
      // 获取管理范围内的学校
      const schoolPrefix = courierInfo.level === 4 ? '' : (courierInfo.managedOPCodePrefix || '').slice(0, 2)
      
      // 获取学校列表
      const schoolsResponse = await apiClient.get('/api/v1/opcode/schools/managed')
      if ((schoolsResponse.data as any).success) {
        setSchools((schoolsResponse.data as any).data.schools || [])
      }

      // 获取片区列表
      const districtsResponse = await apiClient.get(`/api/v1/opcode/areas/managed`)
      if ((districtsResponse.data as any).success) {
        setDistricts((districtsResponse.data as any).data.areas || [])
      }
    } catch (error) {
      console.error('Failed to fetch data:', error)
      // 使用模拟数据
      setSchools([
        { school_code: 'CS', school_name: '中南大学' },
        { school_code: 'HU', school_name: '湖南大学' },
        { school_code: 'LC', school_name: '长沙理工大学' }
      ])
      setDistricts([
        {
          id: '1',
          school_code: 'CS',
          school_name: '中南大学',
          area_code: '01',
          area_name: '本部东区',
          description: '包含1-5栋宿舍楼',
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        },
        {
          id: '2',
          school_code: 'CS',
          school_name: '中南大学',
          area_code: '02',
          area_name: '本部西区',
          description: '包含6-10栋宿舍楼',
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
    if (!formData.school_code || !formData.area_code || !formData.area_name) {
      toast({
        title: '请填写必要字段',
        description: '学校、片区代码和名称为必填项',
        variant: 'destructive'
      })
      return
    }

    // 验证片区代码格式（2位字母或数字）
    if (!/^[A-Z0-9]{2}$/i.test(formData.area_code)) {
      toast({
        title: '片区代码格式错误',
        description: '片区代码必须是2位字母或数字',
        variant: 'destructive'
      })
      return
    }

    try {
      const response = await apiClient.post('/api/v1/opcode/areas', {
        ...formData,
        area_code: formData.area_code.toUpperCase()
      })

      if ((response.data as any).success) {
        toast({
          title: '添加成功',
          description: '片区已成功创建'
        })
        setShowAddDialog(false)
        setFormData({ school_code: '', area_code: '', area_name: '', description: '' })
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

  const handleUpdate = async (district: District) => {
    try {
      const response = await apiClient.put(`/api/v1/opcode/areas/${district.id}`, {
        area_name: district.area_name,
        description: district.description,
        is_active: district.is_active
      })

      if ((response.data as any).success) {
        toast({
          title: '更新成功',
          description: '片区信息已更新'
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
    if (!confirm('确定要删除这个片区吗？删除后该片区下的所有楼栋和投递点将失效。')) {
      return
    }

    try {
      const response = await apiClient.delete(`/api/v1/opcode/areas/${id}`)
      if ((response.data as any).success) {
        toast({
          title: '删除成功',
          description: '片区已删除'
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

  const handleEdit = (district: District) => {
    setEditingId(district.id)
  }

  const handleCancelEdit = () => {
    setEditingId(null)
    fetchData() // 重新加载数据以恢复原始值
  }

  const updateDistrictField = (id: string, field: string, value: any) => {
    setDistricts(districts.map(d => 
      d.id === id ? { ...d, [field]: value } : d
    ))
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
          <h3 className="text-lg font-semibold">片区管理</h3>
          <p className="text-sm text-muted-foreground">
            管理学校的片区划分，每个片区对应OP Code的第3-4位
          </p>
        </div>
        <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="w-4 h-4 mr-2" />
              添加片区
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>添加新片区</DialogTitle>
              <DialogDescription>
                为学校添加新的片区划分
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label>选择学校</Label>
                <select
                  className="w-full px-3 py-2 border rounded-md"
                  value={formData.school_code}
                  onChange={(e) => setFormData({ ...formData, school_code: e.target.value })}
                >
                  <option value="">请选择学校</option>
                  {schools.map(school => (
                    <option key={school.school_code} value={school.school_code}>
                      [{school.school_code}] {school.school_name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-2">
                <Label>片区代码（2位）</Label>
                <Input
                  placeholder="如：01, 0A, 1B"
                  value={formData.area_code}
                  onChange={(e) => setFormData({ ...formData, area_code: e.target.value.toUpperCase() })}
                  maxLength={2}
                />
              </div>
              <div className="space-y-2">
                <Label>片区名称</Label>
                <Input
                  placeholder="如：东区、西区、南区"
                  value={formData.area_name}
                  onChange={(e) => setFormData({ ...formData, area_name: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label>描述（选填）</Label>
                <Input
                  placeholder="如：包含1-5栋宿舍楼"
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

      {/* 片区列表 */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>学校</TableHead>
                <TableHead>片区代码</TableHead>
                <TableHead>片区名称</TableHead>
                <TableHead>描述</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {districts.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">
                    暂无片区数据
                  </TableCell>
                </TableRow>
              ) : (
                districts.map((district) => (
                  <TableRow key={district.id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Building2 className="w-4 h-4 text-muted-foreground" />
                        <span className="font-medium">{district.school_name}</span>
                        <Badge variant="outline" className="text-xs">{district.school_code}</Badge>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge className="font-mono">{district.area_code}</Badge>
                    </TableCell>
                    <TableCell>
                      {editingId === district.id ? (
                        <Input
                          value={district.area_name}
                          onChange={(e) => updateDistrictField(district.id, 'area_name', e.target.value)}
                          className="w-32"
                        />
                      ) : (
                        district.area_name
                      )}
                    </TableCell>
                    <TableCell>
                      {editingId === district.id ? (
                        <Input
                          value={district.description || ''}
                          onChange={(e) => updateDistrictField(district.id, 'description', e.target.value)}
                          className="w-48"
                        />
                      ) : (
                        <span className="text-sm text-muted-foreground">
                          {district.description || '-'}
                        </span>
                      )}
                    </TableCell>
                    <TableCell>
                      {editingId === district.id ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => updateDistrictField(district.id, 'is_active', !district.is_active)}
                        >
                          <Badge variant={district.is_active ? 'default' : 'secondary'}>
                            {district.is_active ? '启用' : '停用'}
                          </Badge>
                        </Button>
                      ) : (
                        <Badge variant={district.is_active ? 'default' : 'secondary'}>
                          {district.is_active ? '启用' : '停用'}
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {editingId === district.id ? (
                          <>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleUpdate(district)}
                            >
                              <Check className="h-4 w-4 text-green-600" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={handleCancelEdit}
                            >
                              <X className="h-4 w-4 text-red-600" />
                            </Button>
                          </>
                        ) : (
                          <>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleEdit(district)}
                            >
                              <Edit2 className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDelete(district.id)}
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
          片区代码将成为OP Code的第3-4位。例如：学校代码CS + 片区代码01 = CS01**（前4位确定）
        </AlertDescription>
      </Alert>
    </div>
  )
}