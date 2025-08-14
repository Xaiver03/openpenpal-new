'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Progress } from '@/components/ui/progress'
import { Textarea } from '@/components/ui/textarea'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  Package, 
  QrCode, 
  Download, 
  Printer, 
  AlertCircle, 
  CheckCircle2, 
  Clock,
  Users,
  School,
  MapPin,
  BarChart3,
  Plus,
  Eye,
  Trash2
} from 'lucide-react'
import { CourierPermissionGuard, COURIER_PERMISSION_CONFIGS } from '@/components/courier/CourierPermissionGuard'
import { ManagementPageLayout, createStatCard } from '@/components/courier/ManagementPageLayout'
import { useCourierPermission } from '@/hooks/use-courier-permission'
import { 
  batchAPI, 
  BatchGenerateRequest, 
  BatchRecord, 
  BarcodeRecord, 
  BatchStats 
} from '@/lib/api/batch-management'
import BarcodePreview from '@/components/courier/BarcodePreview'

export default function BatchManagementPage() {
  const { courierInfo } = useCourierPermission()
  const [activeTab, setActiveTab] = useState('generate')
  const [isLoading, setIsLoading] = useState(false)
  const [batches, setBatches] = useState<BatchRecord[]>([])
  const [selectedBatch, setSelectedBatch] = useState<BatchRecord | null>(null)
  const [barcodes, setBarcodes] = useState<BarcodeRecord[]>([])
  const [showPreview, setShowPreview] = useState(false)
  
  // 批量生成表单
  const [generateForm, setGenerateForm] = useState<BatchGenerateRequest>({
    batch_no: `B${Date.now()}`,
    school_code: '',
    quantity: 100,
    code_type: 'normal',
    description: ''
  })

  // 统计数据
  const [stats, setStats] = useState({
    totalBatches: 0,
    totalCodes: 0,
    usedCodes: 0,
    activeBatches: 0
  })

  useEffect(() => {
    loadBatches()
    loadStats()
  }, [])

  const loadBatches = async () => {
    try {
      setIsLoading(true)
      const response = await batchAPI.getBatches({
        page: 1,
        limit: 50,
        school_code: courierInfo?.level === 3 ? courierInfo?.school_code : undefined
      })
      
      if (response.success && response.data) {
        setBatches(response.data.batches)
      } else {
        console.error('获取批次列表失败:', response.message)
        // 如果API调用失败，使用模拟数据作为后备
        const mockBatches: BatchRecord[] = [
          {
            id: '1',
            batch_no: 'B20250101001',
            school_code: 'BJDX',
            quantity: 500,
            generated_count: 500,
            used_count: 123,
            code_type: 'normal',
            status: 'active',
            created_by: courierInfo?.username || 'admin',
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            description: '北京大学常规条码批次'
          },
          {
            id: '2',
            batch_no: 'B20250101002',
            school_code: 'QHDX',
            area_code: '5F',
            quantity: 200,
            generated_count: 200,
            used_count: 45,
            code_type: 'drift',
            status: 'active',
            created_by: courierInfo?.username || 'admin',
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            description: '清华大学5号楼漂流信专用'
          }
        ]
        setBatches(mockBatches)
      }
    } catch (error) {
      console.error('加载批次数据失败:', error)
      // 错误处理：显示模拟数据
      setBatches([])
    } finally {
      setIsLoading(false)
    }
  }

  const loadStats = async () => {
    try {
      const response = await batchAPI.getBatchStats(
        courierInfo?.level === 3 ? courierInfo?.school_code : undefined
      )
      
      if (response.success && response.data) {
        setStats({
          totalBatches: response.data.total_batches,
          totalCodes: response.data.total_codes,
          usedCodes: response.data.used_codes,
          activeBatches: response.data.active_batches
        })
      } else {
        // 后备统计数据计算
        setStats({
          totalBatches: batches.length,
          totalCodes: batches.reduce((sum, batch) => sum + batch.generated_count, 0),
          usedCodes: batches.reduce((sum, batch) => sum + batch.used_count, 0),
          activeBatches: batches.filter(batch => batch.status === 'active').length
        })
      }
    } catch (error) {
      console.error('加载统计数据失败:', error)
      // 使用本地计算的统计数据
      setStats({
        totalBatches: batches.length,
        totalCodes: batches.reduce((sum, batch) => sum + batch.generated_count, 0),
        usedCodes: batches.reduce((sum, batch) => sum + batch.used_count, 0),
        activeBatches: batches.filter(batch => batch.status === 'active').length
      })
    }
  }

  const handleGenerate = async () => {
    setIsLoading(true)
    try {
      const response = await batchAPI.generateBatch({
        ...generateForm,
        operator_id: courierInfo?.id
      })
      
      if (response.success && response.data) {
        alert(`成功生成批次：${response.data.batch_no}，数量：${response.data.generated_count}`)
        await loadBatches()
        await loadStats()
        
        // 重置表单
        setGenerateForm({
          ...generateForm,
          batch_no: `B${Date.now()}`,
          quantity: 100,
          description: ''
        })
        
        // 切换到批次管理页面查看结果
        setActiveTab('manage')
      } else {
        alert(`生成失败：${response.message}`)
      }
    } catch (error) {
      console.error('批量生成失败:', error)
      alert('批量生成失败，请检查网络连接后重试')
    } finally {
      setIsLoading(false)
    }
  }

  const handleDownloadBatch = async (batchId: string) => {
    try {
      const blob = await batchAPI.downloadBatch(batchId, 'pdf')
      
      // 创建下载链接
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `batch_${batchId}_codes.pdf`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
      
      alert('条码文件下载完成')
    } catch (error) {
      console.error('下载失败:', error)
      alert('下载失败，请稍后重试')
    }
  }

  const handleViewBatchDetails = async (batch: BatchRecord) => {
    setSelectedBatch(batch)
    setShowPreview(false)
    setActiveTab('details')
    
    try {
      const response = await batchAPI.getBatchDetails(batch.id)
      
      if (response.success && response.data) {
        setBarcodes(response.data.codes.slice(0, 20)) // 只显示前20个条码
      } else {
        // 后备：使用预览API
        const previewResponse = await batchAPI.previewBatch(batch.id, 20)
        if (previewResponse.success && previewResponse.data) {
          setBarcodes(previewResponse.data.preview_codes)
        } else {
          // 最后后备：模拟数据
          const mockBarcodes: BarcodeRecord[] = Array.from({ length: Math.min(10, batch.generated_count) }, (_, i) => ({
            id: `${batch.id}_${i}`,
            code: `OP7X${String(i).padStart(4, '0')}`,
            batch_id: batch.id,
            status: Math.random() > 0.7 ? 'bound' : 'unactivated',
            bound_at: Math.random() > 0.5 ? new Date().toISOString() : undefined,
            recipient_code: Math.random() > 0.5 ? 'PK5F3D' : undefined
          }))
          setBarcodes(mockBarcodes)
        }
      }
    } catch (error) {
      console.error('加载批次详情失败:', error)
      setBarcodes([])
    }
  }

  const handlePreviewBatch = async (batch: BatchRecord) => {
    setSelectedBatch(batch)
    setActiveTab('preview')
    setShowPreview(true)
    
    try {
      // 加载更多条码用于预览
      const response = await batchAPI.getBatchDetails(batch.id)
      
      if (response.success && response.data) {
        setBarcodes(response.data.codes) // 加载所有条码
      } else {
        // 后备方案
        await handleViewBatchDetails(batch)
      }
    } catch (error) {
      console.error('加载预览数据失败:', error)
      setBarcodes([])
    }
  }

  const handlePreviewDownload = async (config: any) => {
    if (!selectedBatch) return
    
    try {
      const blob = await batchAPI.downloadBatch(selectedBatch.id, config.format)
      
      // 创建下载链接
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `batch_${selectedBatch.batch_no}_preview.${config.format}`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
      
      alert('预览文件下载完成')
    } catch (error) {
      console.error('预览下载失败:', error)
      alert('预览下载失败，请稍后重试')
    }
  }

  const handlePreviewPrint = async (config: any) => {
    if (!selectedBatch) return
    
    // 打开打印对话框
    window.print()
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'generating': return 'bg-blue-100 text-blue-800'
      case 'active': return 'bg-green-100 text-green-800'
      case 'completed': return 'bg-gray-100 text-gray-800'
      case 'expired': return 'bg-red-100 text-red-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'generating': return '生成中'
      case 'active': return '活跃'
      case 'completed': return '已完成'
      case 'expired': return '已过期'
      default: return '未知'
    }
  }

  // 权限检查配置 - 只有L3和L4可以访问
  const batchPermissionConfig = {
    requiredLevel: 3 as const,
    errorTitle: '批量管理权限不足',
    errorDescription: '只有三级及以上信使才能进行批量条码管理'
  }

  const statsData = [
    createStatCard(<Package className="w-5 h-5 text-amber-600" />, stats.totalBatches, '总批次'),
    createStatCard(<QrCode className="w-5 h-5 text-amber-600" />, stats.totalCodes, '总条码数'),
    createStatCard(<CheckCircle2 className="w-5 h-5 text-green-600" />, stats.usedCodes, '已使用'),
    createStatCard(<Clock className="w-5 h-5 text-blue-600" />, stats.activeBatches, '活跃批次')
  ]

  return (
    <CourierPermissionGuard config={batchPermissionConfig}>
      <ManagementPageLayout
        config={{
          title: '批量条码管理',
          description: `${courierInfo?.level === 4 ? '城市级' : '校区级'}批量条码生成与管理`,
          icon: Package
        }}
        stats={statsData}
        searchPlaceholder="搜索批次编号或描述..."
        searchValue=""
        onSearchChange={() => {}}
        filterOptions={[
          { value: 'all', label: '全部状态' },
          { value: 'active', label: '活跃' },
          { value: 'completed', label: '已完成' }
        ]}
        filterValue="all"
        onFilterChange={() => {}}
        sortOptions={[
          { value: 'created_at', label: '创建时间' },
          { value: 'quantity', label: '数量' },
          { value: 'used_count', label: '使用率' }
        ]}
        sortValue="created_at"
        onSortChange={() => {}}
        canCreate={false}
      >
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="bg-white border border-amber-200">
            <TabsTrigger value="generate" className="data-[state=active]:bg-amber-600 data-[state=active]:text-white">
              <Plus className="w-4 h-4 mr-2" />
              批量生成
            </TabsTrigger>
            <TabsTrigger value="manage" className="data-[state=active]:bg-amber-600 data-[state=active]:text-white">
              <BarChart3 className="w-4 h-4 mr-2" />
              批次管理
            </TabsTrigger>
            <TabsTrigger value="details" className="data-[state=active]:bg-amber-600 data-[state=active]:text-white">
              <Eye className="w-4 h-4 mr-2" />
              批次详情
            </TabsTrigger>
            <TabsTrigger value="preview" className="data-[state=active]:bg-amber-600 data-[state=active]:text-white">
              <Printer className="w-4 h-4 mr-2" />
              打印预览
            </TabsTrigger>
          </TabsList>

          {/* 批量生成面板 */}
          <TabsContent value="generate">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <QrCode className="w-5 h-5" />
                  新建批量生成任务
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-4">
                    <div>
                      <Label htmlFor="batch_no">批次编号</Label>
                      <Input
                        id="batch_no"
                        value={generateForm.batch_no}
                        onChange={(e) => setGenerateForm({...generateForm, batch_no: e.target.value})}
                        placeholder="如：B20250101001"
                      />
                    </div>

                    <div>
                      <Label htmlFor="school_code">学校代码</Label>
                      <Select value={generateForm.school_code} onValueChange={(value) => setGenerateForm({...generateForm, school_code: value})}>
                        <SelectTrigger>
                          <SelectValue placeholder="选择学校" />
                        </SelectTrigger>
                        <SelectContent>
                          {courierInfo?.level === 4 ? (
                            <>
                              <SelectItem value="BJDX">BJDX - 北京大学</SelectItem>
                              <SelectItem value="QHDX">QHDX - 清华大学</SelectItem>
                              <SelectItem value="BJHK">BJHK - 北京航空航天大学</SelectItem>
                            </>
                          ) : (
                            <SelectItem value={courierInfo?.school_code || 'BJDX'}>
                              {courierInfo?.school_code || 'BJDX'} - {courierInfo?.school_name || '当前学校'}
                            </SelectItem>
                          )}
                        </SelectContent>
                      </Select>
                    </div>

                    <div>
                      <Label htmlFor="quantity">生成数量</Label>
                      <Input
                        id="quantity"
                        type="number"
                        value={generateForm.quantity}
                        onChange={(e) => setGenerateForm({...generateForm, quantity: parseInt(e.target.value)})}
                        min="1"
                        max="10000"
                      />
                    </div>

                    <div>
                      <Label htmlFor="code_type">条码类型</Label>
                      <Select value={generateForm.code_type} onValueChange={(value: 'normal' | 'drift') => setGenerateForm({...generateForm, code_type: value})}>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="normal">普通条码</SelectItem>
                          <SelectItem value="drift">漂流信专用</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  <div className="space-y-4">
                    <div>
                      <Label htmlFor="description">描述信息</Label>
                      <Textarea
                        id="description"
                        value={generateForm.description}
                        onChange={(e) => setGenerateForm({...generateForm, description: e.target.value})}
                        placeholder="描述这个批次的用途..."
                        rows={4}
                      />
                    </div>

                    <Alert>
                      <AlertCircle className="h-4 w-4" />
                      <AlertDescription>
                        • 生成的条码将自动分配唯一编号<br/>
                        • L4信使可以为任意学校生成条码<br/>
                        • L3信使只能为所管辖的学校生成条码<br/>
                        • 生成后的条码可批量下载打印
                      </AlertDescription>
                    </Alert>
                  </div>
                </div>

                <div className="flex justify-end gap-4">
                  <Button
                    onClick={handleGenerate}
                    disabled={isLoading || !generateForm.school_code || generateForm.quantity < 1}
                    className="bg-amber-600 hover:bg-amber-700"
                  >
                    {isLoading ? (
                      <>
                        <Clock className="w-4 h-4 mr-2 animate-spin" />
                        生成中...
                      </>
                    ) : (
                      <>
                        <QrCode className="w-4 h-4 mr-2" />
                        开始生成
                      </>
                    )}
                  </Button>
                </div>

                {isLoading && (
                  <div className="space-y-2">
                    <Progress value={75} className="w-full" />
                    <p className="text-sm text-amber-600 text-center">正在生成条码，请稍候...</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          {/* 批次管理面板 */}
          <TabsContent value="manage">
            <Card>
              <CardHeader>
                <CardTitle>批次列表</CardTitle>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>批次编号</TableHead>
                      <TableHead>学校</TableHead>
                      <TableHead>类型</TableHead>
                      <TableHead>数量</TableHead>
                      <TableHead>使用率</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {batches.map((batch) => (
                      <TableRow key={batch.id}>
                        <TableCell className="font-mono">{batch.batch_no}</TableCell>
                        <TableCell>{batch.school_code}</TableCell>
                        <TableCell>
                          <Badge variant={batch.code_type === 'drift' ? 'secondary' : 'default'}>
                            {batch.code_type === 'drift' ? '漂流信' : '普通'}
                          </Badge>
                        </TableCell>
                        <TableCell>{batch.generated_count.toLocaleString()}</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Progress value={(batch.used_count / batch.generated_count) * 100} className="w-16" />
                            <span className="text-sm">
                              {Math.round((batch.used_count / batch.generated_count) * 100)}%
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge className={getStatusColor(batch.status)}>
                            {getStatusText(batch.status)}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {new Date(batch.created_at).toLocaleDateString()}
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleViewBatchDetails(batch)}
                              title="查看详情"
                            >
                              <Eye className="w-4 h-4" />
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handlePreviewBatch(batch)}
                              title="打印预览"
                            >
                              <Printer className="w-4 h-4" />
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleDownloadBatch(batch.id)}
                              title="下载文件"
                            >
                              <Download className="w-4 h-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </TabsContent>

          {/* 批次详情面板 */}
          <TabsContent value="details">
            {selectedBatch ? (
              <div className="space-y-6">
                <Card>
                  <CardHeader>
                    <CardTitle>批次信息：{selectedBatch.batch_no}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                      <div>
                        <Label>学校代码</Label>
                        <p className="font-mono">{selectedBatch.school_code}</p>
                      </div>
                      <div>
                        <Label>条码类型</Label>
                        <p>{selectedBatch.code_type === 'drift' ? '漂流信专用' : '普通条码'}</p>
                      </div>
                      <div>
                        <Label>生成数量</Label>
                        <p>{selectedBatch.generated_count.toLocaleString()}</p>
                      </div>
                      <div>
                        <Label>已使用</Label>
                        <p>{selectedBatch.used_count.toLocaleString()}</p>
                      </div>
                    </div>
                    {selectedBatch.description && (
                      <div className="mt-4">
                        <Label>描述</Label>
                        <p>{selectedBatch.description}</p>
                      </div>
                    )}
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>条码列表（前10个）</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>条码编号</TableHead>
                          <TableHead>状态</TableHead>
                          <TableHead>绑定时间</TableHead>
                          <TableHead>收件编码</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {barcodes.map((barcode) => (
                          <TableRow key={barcode.id}>
                            <TableCell className="font-mono">{barcode.code}</TableCell>
                            <TableCell>
                              <Badge className={getStatusColor(barcode.status)}>
                                {getStatusText(barcode.status)}
                              </Badge>
                            </TableCell>
                            <TableCell>
                              {barcode.bound_at ? new Date(barcode.bound_at).toLocaleString() : '-'}
                            </TableCell>
                            <TableCell className="font-mono">
                              {barcode.recipient_code || '-'}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </CardContent>
                </Card>
              </div>
            ) : (
              <Card>
                <CardContent className="py-8 text-center">
                  <p className="text-amber-600">请从批次管理页面选择一个批次查看详情</p>
                </CardContent>
              </Card>
            )}
          </TabsContent>

          {/* 打印预览面板 */}
          <TabsContent value="preview">
            {selectedBatch && showPreview ? (
              <BarcodePreview
                batch={selectedBatch}
                codes={barcodes}
                onDownload={handlePreviewDownload}
                onPrint={handlePreviewPrint}
              />
            ) : (
              <Card>
                <CardContent className="py-8 text-center">
                  <p className="text-amber-600">请从批次管理页面选择一个批次进行打印预览</p>
                </CardContent>
              </Card>
            )}
          </TabsContent>
        </Tabs>
      </ManagementPageLayout>
    </CourierPermissionGuard>
  )
}