'use client'

import { useState, useRef } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  Printer, 
  Download, 
  Eye, 
  Grid3X3,
  FileImage,
  FileText,
  Settings,
  Zap,
  Package
} from 'lucide-react'
import { BatchRecord, BarcodeRecord } from '@/lib/api/batch-management'

// 条码预览配置
interface PreviewConfig {
  format: 'pdf' | 'png' | 'svg'
  layout: 'grid' | 'list' | 'labels'
  size: 'small' | 'medium' | 'large'
  includeText: boolean
  includeDesc: boolean
  codesPerPage: number
  margin: 'none' | 'small' | 'medium' | 'large'
}

interface BarcodePreviewProps {
  batch: BatchRecord
  codes: BarcodeRecord[]
  onDownload: (config: PreviewConfig) => void
  onPrint: (config: PreviewConfig) => void
}

// 生成模拟QR码SVG
const generateQRCodeSVG = (code: string, size: number = 120) => {
  // 简单的网格模式生成模拟QR码
  const modules = 21 // 标准QR码为21x21
  const moduleSize = size / modules
  const pattern = code.split('').reduce((acc, char, index) => {
    acc[Math.floor(index / modules)] = acc[Math.floor(index / modules)] || []
    acc[Math.floor(index / modules)][index % modules] = char.charCodeAt(0) % 2 === 0
    return acc
  }, [] as boolean[][])

  // 填充剩余格子
  for (let i = 0; i < modules; i++) {
    pattern[i] = pattern[i] || []
    for (let j = 0; j < modules; j++) {
      if (pattern[i][j] === undefined) {
        pattern[i][j] = (i + j) % 3 === 0
      }
    }
  }

  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} className="border">
      <rect width={size} height={size} fill="white" />
      {pattern.map((row, i) =>
        row.map((module, j) =>
          module ? (
            <rect
              key={`${i}-${j}`}
              x={j * moduleSize}
              y={i * moduleSize}
              width={moduleSize}
              height={moduleSize}
              fill="black"
            />
          ) : null
        )
      )}
    </svg>
  )
}

export default function BarcodePreview({ batch, codes, onDownload, onPrint }: BarcodePreviewProps) {
  const [config, setConfig] = useState<PreviewConfig>({
    format: 'pdf',
    layout: 'grid',
    size: 'medium',
    includeText: true,
    includeDesc: true,
    codesPerPage: 12,
    margin: 'medium'
  })

  const previewRef = useRef<HTMLDivElement>(null)

  const getQRSize = () => {
    switch (config.size) {
      case 'small': return 80
      case 'medium': return 120
      case 'large': return 160
      default: return 120
    }
  }

  const getGridCols = () => {
    if (config.layout === 'list') return 1
    switch (config.codesPerPage) {
      case 4: return 2
      case 6: return 3
      case 9: return 3
      case 12: return 4
      case 16: return 4
      case 20: return 5
      default: return 4
    }
  }

  const getMarginClass = () => {
    switch (config.margin) {
      case 'none': return 'p-0'
      case 'small': return 'p-2'
      case 'medium': return 'p-4'
      case 'large': return 'p-6'
      default: return 'p-4'
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('zh-CN')
  }

  const displayCodes = codes.slice(0, config.codesPerPage)

  return (
    <div className="space-y-6">
      {/* 预览配置面板 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings className="w-5 h-5" />
            打印预览设置
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <Label htmlFor="format">输出格式</Label>
              <Select 
                value={config.format} 
                onValueChange={(value: PreviewConfig['format']) => 
                  setConfig({...config, format: value})
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="pdf">
                    <div className="flex items-center gap-2">
                      <FileText className="w-4 h-4" />
                      PDF文档
                    </div>
                  </SelectItem>
                  <SelectItem value="png">
                    <div className="flex items-center gap-2">
                      <FileImage className="w-4 h-4" />
                      PNG图片
                    </div>
                  </SelectItem>
                  <SelectItem value="svg">
                    <div className="flex items-center gap-2">
                      <FileImage className="w-4 h-4" />
                      SVG矢量
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="layout">布局方式</Label>
              <Select 
                value={config.layout} 
                onValueChange={(value: PreviewConfig['layout']) => 
                  setConfig({...config, layout: value})
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="grid">网格布局</SelectItem>
                  <SelectItem value="list">列表布局</SelectItem>
                  <SelectItem value="labels">标签模式</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="size">条码大小</Label>
              <Select 
                value={config.size} 
                onValueChange={(value: PreviewConfig['size']) => 
                  setConfig({...config, size: value})
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="small">小 (80px)</SelectItem>
                  <SelectItem value="medium">中 (120px)</SelectItem>
                  <SelectItem value="large">大 (160px)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="perPage">每页数量</Label>
              <Select 
                value={config.codesPerPage.toString()} 
                onValueChange={(value) => 
                  setConfig({...config, codesPerPage: parseInt(value)})
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="4">4个</SelectItem>
                  <SelectItem value="6">6个</SelectItem>
                  <SelectItem value="9">9个</SelectItem>
                  <SelectItem value="12">12个</SelectItem>
                  <SelectItem value="16">16个</SelectItem>
                  <SelectItem value="20">20个</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="flex gap-4">
            <div>
              <Label htmlFor="margin">页面边距</Label>
              <Select 
                value={config.margin} 
                onValueChange={(value: PreviewConfig['margin']) => 
                  setConfig({...config, margin: value})
                }
              >
                <SelectTrigger className="w-32">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">无边距</SelectItem>
                  <SelectItem value="small">小边距</SelectItem>
                  <SelectItem value="medium">中边距</SelectItem>
                  <SelectItem value="large">大边距</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-end gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={config.includeText}
                  onChange={(e) => setConfig({...config, includeText: e.target.checked})}
                  className="rounded border-amber-300"
                />
                <span className="text-sm">显示条码文字</span>
              </label>
              
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={config.includeDesc}
                  onChange={(e) => setConfig({...config, includeDesc: e.target.checked})}
                  className="rounded border-amber-300"
                />
                <span className="text-sm">显示描述信息</span>
              </label>
            </div>
          </div>

          <div className="flex gap-2">
            <Button
              onClick={() => onDownload(config)}
              className="bg-amber-600 hover:bg-amber-700"
            >
              <Download className="w-4 h-4 mr-2" />
              下载 ({config.format.toUpperCase()})
            </Button>
            <Button
              onClick={() => onPrint(config)}
              variant="outline"
              className="border-amber-300 text-amber-700 hover:bg-amber-50"
            >
              <Printer className="w-4 h-4 mr-2" />
              打印预览
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* 批次信息 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Package className="w-5 h-5" />
            批次信息
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <Label>批次编号</Label>
              <p className="font-mono text-amber-900">{batch.batch_no}</p>
            </div>
            <div>
              <Label>学校代码</Label>
              <p>{batch.school_code}</p>
            </div>
            <div>
              <Label>条码类型</Label>
              <Badge variant={batch.code_type === 'drift' ? 'secondary' : 'default'}>
                {batch.code_type === 'drift' ? '漂流信' : '普通'}
              </Badge>
            </div>
            <div>
              <Label>创建日期</Label>
              <p>{formatDate(batch.created_at)}</p>
            </div>
          </div>
          {config.includeDesc && batch.description && (
            <div className="mt-4">
              <Label>描述</Label>
              <p>{batch.description}</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 条码预览 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Eye className="w-5 h-5" />
              条码预览 (显示前 {displayCodes.length} 个)
            </div>
            <div className="flex items-center gap-2 text-sm text-amber-600">
              <Grid3X3 className="w-4 h-4" />
              {getGridCols()} 列 × {Math.ceil(displayCodes.length / getGridCols())} 行
            </div>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Alert className="mb-6">
            <Zap className="h-4 w-4" />
            <AlertDescription>
              这是条码的预览效果。实际打印时将生成高质量的{config.format.toUpperCase()}文件。
              建议使用专业标签打印机以获得最佳效果。
            </AlertDescription>
          </Alert>

          <div 
            ref={previewRef}
            className={`bg-white border-2 border-dashed border-amber-300 ${getMarginClass()}`}
            style={{ minHeight: '400px' }}
          >
            <div 
              className={`grid gap-4`}
              style={{ 
                gridTemplateColumns: `repeat(${getGridCols()}, 1fr)`,
                justifyItems: 'center',
                alignItems: 'start'
              }}
            >
              {displayCodes.map((code, index) => (
                <div 
                  key={code.id} 
                  className={`flex flex-col items-center text-center ${
                    config.layout === 'labels' ? 'border border-gray-300 rounded p-2' : ''
                  }`}
                >
                  {/* QR码 */}
                  <div className="mb-2">
                    {generateQRCodeSVG(code.code, getQRSize())}
                  </div>
                  
                  {/* 条码文字 */}
                  {config.includeText && (
                    <div className="space-y-1">
                      <p className="font-mono text-xs font-semibold">
                        {code.code}
                      </p>
                      <p className="text-xs text-gray-600">
                        {batch.school_code} - {batch.code_type === 'drift' ? '漂流' : '普通'}
                      </p>
                      {code.status !== 'unactivated' && (
                        <Badge 
                          variant="secondary" 
                          className="text-xs"
                        >
                          {code.status === 'bound' ? '已绑定' : 
                           code.status === 'in_transit' ? '投递中' : 
                           code.status === 'delivered' ? '已送达' : '已过期'}
                        </Badge>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          {codes.length > displayCodes.length && (
            <div className="mt-4 text-center">
              <Alert>
                <AlertDescription>
                  当前预览显示前 {displayCodes.length} 个条码，
                  完整批次共 {codes.length} 个条码将在实际下载时全部包含。
                </AlertDescription>
              </Alert>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 打印建议 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Printer className="w-5 h-5" />
            打印建议
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <h4 className="font-semibold text-amber-900 mb-2">推荐设置</h4>
              <ul className="space-y-1 text-amber-700">
                <li>• 纸张：A4 不干胶标签纸</li>
                <li>• 打印质量：高质量/600DPI</li>
                <li>• 颜色：黑白打印即可</li>
                <li>• 边距：根据实际标签尺寸调整</li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold text-amber-900 mb-2">注意事项</h4>
              <ul className="space-y-1 text-amber-700">
                <li>• 确保QR码清晰可扫描</li>
                <li>• 避免污损和折叠</li>
                <li>• 标签牢固粘贴在信封上</li>
                <li>• 定期检查打印机状态</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}