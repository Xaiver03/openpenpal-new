'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { 
  Package,
  Truck,
  MapPin,
  Clock,
  CheckCircle,
  AlertCircle,
  QrCode,
  Printer,
  Send,
  RefreshCw,
  Eye
} from 'lucide-react'
import { LetterService, type Letter } from '@/lib/services/letter-service'
import { useAuth } from '@/contexts/auth-context'
import { formatRelativeTime } from '@/lib/utils'

export default function DeliverPage() {
  const { user } = useAuth()
  const [letters, setLetters] = useState<Letter[]>([])
  const [selectedLetter, setSelectedLetter] = useState<Letter | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // 加载信件数据
  const loadLetters = async () => {
    if (!user) return
    
    setLoading(true)
    setError(null)
    
    try {
      // 获取已发送的信件（有二维码的信件）
      const response = await LetterService.getUserLetters({
        type: 'sent',
        page: 1,
        limit: 100,
        sort_by: 'created_at',
        sort_order: 'desc'
      })
      
      if (response.success && response.data) {
        // 过滤出有二维码的信件（已生成编号的信件）
        const lettersWithCodes = response.data.letters.filter(letter => 
          letter.code && ['generated', 'collected', 'in_transit', 'delivered'].includes(letter.status)
        )
        setLetters(lettersWithCodes)
      }
    } catch (err) {
      console.error('Failed to load letters:', err)
      setError('加载信件失败，请刷新重试')
      setLetters([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadLetters()
  }, [user])

  const getStatusInfo = (status: string) => {
    switch (status) {
      case 'generated':
        return {
          label: '待投递',
          color: 'bg-blue-100 text-blue-800 border-blue-200',
          icon: Package
        }
      case 'collected':
        return {
          label: '已收取',
          color: 'bg-yellow-100 text-yellow-800 border-yellow-200',
          icon: Printer
        }
      case 'in_transit':
        return {
          label: '投递中',
          color: 'bg-orange-100 text-orange-800 border-orange-200',
          icon: Truck
        }
      case 'delivered':
        return {
          label: '已投递',
          color: 'bg-green-100 text-green-800 border-green-200',
          icon: CheckCircle
        }
      default:
        return {
          label: '未知状态',
          color: 'bg-gray-100 text-gray-800 border-gray-200',
          icon: AlertCircle
        }
    }
  }

  const handlePrintSticker = (letter: Letter) => {
    // 打印贴纸
    const printWindow = window.open('', '_blank')
    if (printWindow) {
      printWindow.document.write(`
        <html>
          <head>
            <title>信件贴纸 - ${letter.code}</title>
            <style>
              body { 
                font-family: Arial, sans-serif; 
                margin: 20px; 
                text-align: center; 
              }
              .sticker {
                border: 2px solid #000;
                padding: 20px;
                width: 200px;
                margin: 0 auto;
                background: white;
              }
              .qr-code {
                width: 150px;
                height: 150px;
                margin: 10px auto;
                background: #f0f0f0;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 12px;
                color: #666;
              }
              .code {
                font-family: monospace;
                font-size: 14px;
                font-weight: bold;
                margin-top: 10px;
              }
            </style>
          </head>
          <body>
            <div class="sticker">
              <h3>OpenPenPal 信使计划</h3>
              <div class="qr-code">
                ${letter.qrCodeUrl ? 
                  `<img src="${letter.qrCodeUrl}" alt="QR Code" style="max-width: 100%; max-height: 100%;" />` :
                  '二维码'
                }
              </div>
              <div class="code">${letter.code}</div>
              <p style="font-size: 12px; margin-top: 15px;">
                扫码查看信件内容
              </p>
              ${letter.delivery_address ? `<p style="font-size: 10px; margin-top: 10px;">投递地址: ${letter.delivery_address}</p>` : ''}
            </div>
          </body>
        </html>
      `)
      printWindow.document.close()
      printWindow.print()
    }
  }

  const handleViewLetter = (letter: Letter) => {
    if (letter.read_url) {
      window.open(letter.read_url, '_blank')
    } else if (letter.code) {
      window.open(`/read/${letter.code}`, '_blank')
    }
  }

  const generateLetterCode = async (letter: Letter) => {
    try {
      const response = await LetterService.generateLetterCode(letter.id)
      if (response.success && response.data) {
        // 重新加载信件列表以获取更新的数据
        await loadLetters()
      }
    } catch (err) {
      console.error('Failed to generate letter code:', err)
    }
  }

  if (loading) {
    return (
      <div className="container max-w-6xl mx-auto px-4 py-8">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          <span className="ml-2 text-muted-foreground">加载中...</span>
        </div>
      </div>
    )
  }

  return (
    <div className="container max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="font-serif text-3xl font-bold text-letter-ink mb-2">
              信件投递
            </h1>
            <p className="text-muted-foreground">
              管理你的信件投递流程，跟踪投递状态
            </p>
          </div>
          <Button variant="outline" onClick={loadLetters} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
        </div>
      </div>

      {/* Error State */}
      {error && (
        <Card className="border-destructive mb-6">
          <CardContent className="pt-6">
            <div className="text-center">
              <p className="text-destructive mb-4">{error}</p>
              <Button onClick={loadLetters} variant="outline">
                重新加载
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 统计信息 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        {[
          { label: '待投递', value: letters.filter(l => l.status === 'pending').length, icon: Package, color: 'text-blue-600' },
          { label: '已收取', value: letters.filter(l => l.delivery_status === 'assigned').length, icon: Printer, color: 'text-yellow-600' },
          { label: '投递中', value: letters.filter(l => l.delivery_status === 'in_transit').length, icon: Truck, color: 'text-orange-600' },
          { label: '已投递', value: letters.filter(l => l.status === 'delivered').length, icon: CheckCircle, color: 'text-green-600' },
        ].map((stat) => {
          const Icon = stat.icon
          return (
            <Card key={stat.label}>
              <CardContent className="flex items-center p-6">
                <Icon className={`h-8 w-8 ${stat.color} mr-4`} />
                <div>
                  <div className="text-2xl font-bold">{stat.value}</div>
                  <div className="text-sm text-muted-foreground">{stat.label}</div>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* 信件列表 */}
      {letters.length === 0 ? (
        <Card className="text-center py-12">
          <CardContent>
            <Package className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-semibold mb-2">暂无投递信件</h3>
            <p className="text-muted-foreground mb-4">
              还没有已生成编号的信件需要投递
            </p>
            <Button asChild>
              <a href="/write">
                <Send className="mr-2 h-4 w-4" />
                写第一封信
              </a>
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-6">
          {letters.map((letter) => {
            const statusInfo = getStatusInfo(letter.status)
            const StatusIcon = statusInfo.icon
            
            return (
              <Card key={letter.id} className="overflow-hidden">
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle className="text-lg">{letter.title || '无标题信件'}</CardTitle>
                      <CardDescription className="flex items-center gap-2 mt-1">
                        <span className="font-mono text-sm">{letter.code}</span>
                        <span>•</span>
                        <span>{formatRelativeTime(new Date(letter.createdAt))}</span>
                        {letter.recipient_info?.name && (
                          <>
                            <span>•</span>
                            <span>收件人: {letter.recipient_info.name}</span>
                          </>
                        )}
                      </CardDescription>
                    </div>
                    <Badge className={`${statusInfo.color} border`}>
                      <StatusIcon className="mr-1 h-3 w-3" />
                      {statusInfo.label}
                    </Badge>
                  </div>
                </CardHeader>
                
                <CardContent className="pt-0">
                  <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    {/* 信件信息 */}
                    <div className="lg:col-span-2 space-y-4">
                      {(letter.delivery_address || letter.delivery_notes) && (
                        <Alert>
                          <MapPin className="h-4 w-4" />
                          <AlertDescription>
                            <strong>投递信息：</strong>
                            {letter.delivery_address && <div>地址: {letter.delivery_address}</div>}
                            {letter.delivery_notes && <div>备注: {letter.delivery_notes}</div>}
                          </AlertDescription>
                        </Alert>
                      )}
                      
                      <div className="flex flex-wrap gap-3">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handlePrintSticker(letter)}
                        >
                          <Printer className="mr-2 h-4 w-4" />
                          打印贴纸
                        </Button>
                        
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleViewLetter(letter)}
                        >
                          <Eye className="mr-2 h-4 w-4" />
                          查看信件
                        </Button>
                        
                        {letter.qrCodeUrl && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setSelectedLetter(letter)}
                          >
                            <QrCode className="mr-2 h-4 w-4" />
                            查看二维码
                          </Button>
                        )}
                        
                        {!letter.code && (
                          <Button
                            size="sm"
                            onClick={() => generateLetterCode(letter)}
                          >
                            <QrCode className="mr-2 h-4 w-4" />
                            生成编号
                          </Button>
                        )}
                      </div>
                    </div>
                    
                    {/* 二维码预览 */}
                    <div className="flex justify-center">
                      <div className="w-24 h-24 bg-white border border-border rounded-lg p-2 flex items-center justify-center">
                        {letter.qrCodeUrl ? (
                          <img 
                            src={letter.qrCodeUrl} 
                            alt="QR Code" 
                            className="w-full h-full object-contain"
                          />
                        ) : (
                          <QrCode className="h-full w-full text-muted-foreground" />
                        )}
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}

      {/* 投递指南 */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertCircle className="h-5 w-5" />
            投递指南
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm text-muted-foreground">
          <div className="flex items-start gap-2">
            <span className="text-primary">1.</span>
            <span>生成信件编号和二维码</span>
          </div>
          <div className="flex items-start gap-2">
            <span className="text-primary">2.</span>
            <span>打印二维码贴纸，贴在信封明显位置</span>
          </div>
          <div className="flex items-start gap-2">
            <span className="text-primary">3.</span>
            <span>手写信件内容到信纸上，装入信封</span>
          </div>
          <div className="flex items-start gap-2">
            <span className="text-primary">4.</span>
            <span>根据收信人信息投递到指定位置</span>
          </div>
          <div className="flex items-start gap-2">
            <span className="text-primary">5.</span>
            <span>信使会扫码更新投递状态</span>
          </div>
        </CardContent>
      </Card>

      {/* 二维码查看弹窗 */}
      {selectedLetter && (
        <div 
          className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
          onClick={() => setSelectedLetter(null)}
        >
          <Card className="max-w-md mx-4" onClick={(e) => e.stopPropagation()}>
            <CardHeader>
              <CardTitle>二维码 - {selectedLetter.code}</CardTitle>
            </CardHeader>
            <CardContent className="text-center">
              <div className="w-64 h-64 bg-white border rounded-lg p-4 mx-auto mb-4">
                {selectedLetter.qrCodeUrl ? (
                  <img 
                    src={selectedLetter.qrCodeUrl} 
                    alt="QR Code" 
                    className="w-full h-full object-contain"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <QrCode className="h-24 w-24 text-muted-foreground" />
                  </div>
                )}
              </div>
              <p className="text-sm text-muted-foreground mb-4">
                扫描此二维码可查看信件内容
              </p>
              <Button variant="outline" onClick={() => setSelectedLetter(null)}>
                关闭
              </Button>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}