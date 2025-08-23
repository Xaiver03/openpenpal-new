'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Breadcrumb, ADMIN_BREADCRUMBS } from '@/components/ui/breadcrumb'
import { BackButton } from '@/components/ui/back-button'
import { 
  QrCode,
  MapPin,
  Clock,
  User,
  Mail,
  Truck,
  History,
  AlertTriangle,
  CheckCircle,
  ArrowLeft,
  Download,
  Eye,
  Trash2,
  RefreshCw,
  Copy
} from 'lucide-react'
import { 
  BarcodeService, 
  type Barcode, 
  type BarcodeScanLog,
  getBarcodeStatusInfo 
} from '@/lib/services/barcode-service'
import { usePermission } from '@/hooks/use-permission'
import { formatDistanceToNow, format } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { toast } from '@/components/ui/use-toast'

export default function AdminBarcodeDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { hasPermission } = usePermission()
  const [barcode, setBarcode] = useState<Barcode | null>(null)
  const [scanLogs, setScanLogs] = useState<BarcodeScanLog[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const code = params?.code as string

  // 加载条码详情
  const loadBarcodeDetail = async () => {
    if (!code) return
    
    try {
      setLoading(true)
      setError(null)
      
      const [barcodeResponse, logsResponse] = await Promise.all([
        BarcodeService.getBarcodeByCode(code),
        BarcodeService.getBarcodeScanLogs(code)
      ])
      
      setBarcode(barcodeResponse.data)
      setScanLogs(((logsResponse as any)?.data?.data || (logsResponse as any)?.data).logs)
    } catch (err: any) {
      setError(err.message || '加载条码详情失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadBarcodeDetail()
  }, [code])

  // 检查权限
  if (!hasPermission('admin.barcodes.read')) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            您没有权限访问条码详情
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  // 作废条码
  const handleVoidBarcode = async () => {
    if (!barcode || !confirm('确定要作废此条码吗？此操作不可撤销。')) {
      return
    }

    try {
      await BarcodeService.voidBarcode(barcode.code)
      toast({
        title: '操作成功',
        description: `条码 ${barcode.code} 已被作废`
      })
      loadBarcodeDetail()
    } catch (err: any) {
      toast({
        title: '操作失败',
        description: err.message || '作废条码失败',
        variant: 'destructive'
      })
    }
  }

  // 复制条码编号
  const copyBarcodeCode = () => {
    if (barcode) {
      navigator.clipboard.writeText(barcode.code)
      toast({
        title: '已复制',
        description: '条码编号已复制到剪贴板'
      })
    }
  }

  // 下载条码图片
  const downloadBarcodeImage = () => {
    if (barcode?.png_url) {
      window.open(barcode.png_url, '_blank')
    }
  }

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/4"></div>
          <div className="h-32 bg-gray-200 rounded"></div>
          <div className="h-64 bg-gray-200 rounded"></div>
        </div>
      </div>
    )
  }

  if (error || !barcode) {
    return (
      <div className="container mx-auto px-4 py-8">
        <BackButton href="/admin/barcodes" />
        <Alert variant="destructive" className="mt-4">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>{error || '条码不存在'}</AlertDescription>
        </Alert>
      </div>
    )
  }

  const statusInfo = getBarcodeStatusInfo(barcode.status)

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 面包屑导航 */}
      <Breadcrumb items={[
        ...ADMIN_BREADCRUMBS.root, 
        { label: '条码管理', href: '/admin/barcodes' },
        { label: barcode.code, href: `/admin/barcodes/${barcode.code}` }
      ]} />
      
      {/* 页面标题和操作 */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <BackButton href="/admin/barcodes" />
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
              <QrCode className="h-8 w-8" />
              条码详情
            </h1>
            <p className="text-gray-600 mt-1 font-mono">{barcode.code}</p>
          </div>
        </div>
        
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => loadBarcodeDetail()}>
            <RefreshCw className="w-4 h-4 mr-2" />
            刷新
          </Button>
          
          <Button variant="outline" onClick={copyBarcodeCode}>
            <Copy className="w-4 h-4 mr-2" />
            复制编号
          </Button>
          
          {barcode.png_url && (
            <Button variant="outline" onClick={downloadBarcodeImage}>
              <Download className="w-4 h-4 mr-2" />
              下载图片
            </Button>
          )}
          
          {hasPermission('admin.barcodes.void') && barcode.status !== 'voided' && (
            <Button variant="destructive" onClick={handleVoidBarcode}>
              <Trash2 className="w-4 h-4 mr-2" />
              作废条码
            </Button>
          )}
        </div>
      </div>

      {/* 基本信息卡片 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        {/* 条码预览 */}
        <Card>
          <CardHeader>
            <CardTitle>条码预览</CardTitle>
          </CardHeader>
          <CardContent className="text-center">
            {barcode.png_url ? (
              <div className="space-y-4">
                <img 
                  src={barcode.png_url} 
                  alt={`条码 ${barcode.code}`}
                  className="mx-auto max-w-full h-32 object-contain border rounded"
                />
                <p className="text-sm text-gray-600 font-mono">{barcode.code}</p>
              </div>
            ) : (
              <div className="py-8 text-gray-400">
                <QrCode className="h-16 w-16 mx-auto mb-2" />
                <p>条码图片不可用</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* 状态信息 */}
        <Card>
          <CardHeader>
            <CardTitle>状态信息</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">当前状态</span>
              <Badge variant="outline" className={statusInfo.color}>
                {statusInfo.label}
              </Badge>
            </div>
            <div>
              <p className="text-sm text-gray-600">{statusInfo.description}</p>
            </div>
            
            <div className="space-y-2 pt-2 border-t">
              <div className="flex justify-between text-sm">
                <span>创建时间</span>
                <span className="text-gray-600">
                  {format(new Date(barcode.createdAt), 'yyyy-MM-dd HH:mm:ss')}
                </span>
              </div>
              
              {barcode.bound_at && (
                <div className="flex justify-between text-sm">
                  <span>绑定时间</span>
                  <span className="text-gray-600">
                    {format(new Date(barcode.bound_at), 'yyyy-MM-dd HH:mm:ss')}
                  </span>
                </div>
              )}
              
              {barcode.expires_at && (
                <div className="flex justify-between text-sm">
                  <span>过期时间</span>
                  <span className="text-gray-600">
                    {format(new Date(barcode.expires_at), 'yyyy-MM-dd HH:mm:ss')}
                  </span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* 关联信息 */}
        <Card>
          <CardHeader>
            <CardTitle>关联信息</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <Mail className="h-4 w-4 text-gray-400" />
                <span className="text-sm font-medium">信件ID:</span>
                {barcode.letter_id ? (
                  <span className="text-sm text-blue-600">{barcode.letter_id}</span>
                ) : (
                  <span className="text-sm text-gray-400">未绑定</span>
                )}
              </div>
              
              <div className="flex items-center gap-2">
                <MapPin className="h-4 w-4 text-gray-400" />
                <span className="text-sm font-medium">收件编码:</span>
                {barcode.recipient_code ? (
                  <span className="text-sm font-mono">{barcode.recipient_code}</span>
                ) : (
                  <span className="text-sm text-gray-400">未设置</span>
                )}
              </div>
              
              <div className="flex items-center gap-2">
                <User className="h-4 w-4 text-gray-400" />
                <span className="text-sm font-medium">创建者:</span>
                <span className="text-sm">{barcode.created_by}</span>
              </div>
              
              {barcode.batch_id && (
                <div className="flex items-center gap-2">
                  <QrCode className="h-4 w-4 text-gray-400" />
                  <span className="text-sm font-medium">批次ID:</span>
                  <span className="text-sm font-mono">{barcode.batch_id}</span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 详细标签页 */}
      <Tabs defaultValue="logs" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="logs" className="flex items-center gap-2">
            <History className="h-4 w-4" />
            扫描记录
          </TabsTrigger>
          <TabsTrigger value="security" className="flex items-center gap-2">
            <CheckCircle className="h-4 w-4" />
            安全信息
          </TabsTrigger>
        </TabsList>

        {/* 扫描记录 */}
        <TabsContent value="logs">
          <Card>
            <CardHeader>
              <CardTitle>扫描记录</CardTitle>
              <CardDescription>
                共 {scanLogs.length} 条扫描记录
              </CardDescription>
            </CardHeader>
            <CardContent>
              {scanLogs.length > 0 ? (
                <div className="space-y-4">
                  {scanLogs.map((log, index) => (
                    <div 
                      key={log.id} 
                      className="flex items-start gap-4 p-4 border rounded-lg"
                    >
                      <div className="flex-shrink-0">
                        <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                          <span className="text-xs font-bold text-blue-600">
                            {scanLogs.length - index}
                          </span>
                        </div>
                      </div>
                      
                      <div className="flex-1 space-y-2">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <span className="font-medium">{log.action}</span>
                            <span className="text-sm text-gray-500">
                              {log.old_status} → {log.new_status}
                            </span>
                          </div>
                          <span className="text-sm text-gray-500">
                            {formatDistanceToNow(new Date(log.createdAt), { 
                              addSuffix: true, 
                              locale: zhCN 
                            })}
                          </span>
                        </div>
                        
                        <div className="text-sm text-gray-600">
                          <p>操作者: {log.scanner_id}</p>
                          {log.location && <p>位置: {log.location}</p>}
                          {log.note && <p>备注: {log.note}</p>}
                        </div>
                        
                        <div className="text-xs text-gray-400">
                          <p>时间: {format(new Date(log.createdAt), 'yyyy-MM-dd HH:mm:ss')}</p>
                          {log.ip_address && <p>IP: {log.ip_address}</p>}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8 text-gray-400">
                  <History className="h-12 w-12 mx-auto mb-2" />
                  <p>暂无扫描记录</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* 安全信息 */}
        <TabsContent value="security">
          <Card>
            <CardHeader>
              <CardTitle>安全信息</CardTitle>
              <CardDescription>
                条码的安全哈希和签名信息
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-3">
                <div>
                  <label className="text-sm font-medium">安全哈希</label>
                  <div className="mt-1 p-3 bg-gray-50 rounded font-mono text-sm break-all">
                    {barcode.security_hash}
                  </div>
                </div>
                
                <div>
                  <label className="text-sm font-medium">签名密钥</label>
                  <div className="mt-1 p-3 bg-gray-50 rounded font-mono text-sm break-all">
                    {barcode.signature_key}
                  </div>
                </div>
                
                <div>
                  <label className="text-sm font-medium">条码类型</label>
                  <div className="mt-1 p-3 bg-gray-50 rounded text-sm">
                    {barcode.type === 'qr' ? 'QR码' : '数字码'}
                  </div>
                </div>
                
                <div>
                  <label className="text-sm font-medium">来源</label>
                  <div className="mt-1 p-3 bg-gray-50 rounded text-sm">
                    {{
                      'write-page': '写信页面',
                      'admin': '管理员生成',
                      'batch-request': '批量生成',
                      'store': '信封商店'
                    }[barcode.source] || barcode.source}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}