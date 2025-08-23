'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { BackButton } from '@/components/ui/back-button'
import { 
  Search,
  MapPin,
  Building,
  Navigation,
  Copy,
  CheckCircle,
  AlertCircle,
  Info,
  Clock,
  Shield,
  Map,
  Phone,
  Users,
  Star,
  QrCode,
  Eye,
  EyeOff
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { usePermission } from '@/hooks/use-permission'
import { apiClient } from '@/lib/api-client'
import { toast } from '@/components/ui/use-toast'

interface OPCodeResult {
  code: string
  school_name: string
  school_code: string
  area_name: string
  area_code: string
  location_name: string
  location_code: string
  full_address: string
  coordinates?: {
    lat: number
    lng: number
  }
  access_level: 'public' | 'partial' | 'private'
  can_deliver: boolean
  delivery_notes?: string
  contact_info?: {
    phone?: string
    office_hours?: string
  }
}

interface SearchHistory {
  code: string
  timestamp: number
  result: 'success' | 'failed'
}

export default function OPCodeSearchPage() {
  const { user } = useAuth()
  const { hasPermission } = usePermission()
  const [searchCode, setSearchCode] = useState('')
  const [searchResult, setSearchResult] = useState<OPCodeResult | null>(null)
  const [searching, setSearching] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [searchHistory, setSearchHistory] = useState<SearchHistory[]>([])
  const [showPrivateInfo, setShowPrivateInfo] = useState(false)

  // 检查是否是信使
  const isCourier = hasPermission('courier.basic') || 
                   hasPermission('courier.intermediate') || 
                   hasPermission('courier.advanced') || 
                   hasPermission('courier.management')

  // 从本地存储加载搜索历史
  useEffect(() => {
    const stored = localStorage.getItem('opcode-search-history')
    if (stored) {
      try {
        const history = JSON.parse(stored).slice(0, 10) // 只保留最近10条
        setSearchHistory(history)
      } catch (e) {
        console.warn('Failed to load search history')
      }
    }
  }, [])

  // 保存搜索历史
  const saveToHistory = (code: string, success: boolean) => {
    const newHistory: SearchHistory[] = [
      { code: code.toUpperCase(), timestamp: Date.now(), result: success ? 'success' as const : 'failed' as const },
      ...searchHistory.filter(h => h.code !== code.toUpperCase())
    ].slice(0, 10)
    
    setSearchHistory(newHistory)
    localStorage.setItem('opcode-search-history', JSON.stringify(newHistory))
  }

  // 搜索OP Code
  const searchOPCode = async (code: string) => {
    if (!code.trim()) {
      setError('请输入OP Code')
      return
    }

    const cleanCode = code.trim().toUpperCase()
    
    // 验证格式
    if (!/^[A-Z0-9]{6}$/.test(cleanCode)) {
      setError('OP Code格式错误，应为6位字母数字组合')
      return
    }

    try {
      setSearching(true)
      setError(null)
      setSearchResult(null)

      const response = await apiClient.get(`/api/v1/opcode/${cleanCode}`)
      const result = ((response as any)?.data?.data || (response as any)?.data)?.data

      if (result) {
        setSearchResult(result)
        saveToHistory(cleanCode, true)
      } else {
        throw new Error('未找到对应地址')
      }
    } catch (err: any) {
      const errorMessage = err.message?.includes('404') ? 'OP Code不存在' : (err.message || '搜索失败')
      setError(errorMessage)
      saveToHistory(cleanCode, false)
    } finally {
      setSearching(false)
    }
  }

  // 处理搜索
  const handleSearch = () => {
    searchOPCode(searchCode)
  }

  // 处理回车搜索
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch()
    }
  }

  // 复制OP Code
  const copyOPCode = (code: string) => {
    navigator.clipboard.writeText(code)
    toast({
      title: '已复制',
      description: `OP Code ${code} 已复制到剪贴板`
    })
  }

  // 复制地址
  const copyAddress = () => {
    if (searchResult) {
      navigator.clipboard.writeText(searchResult.full_address)
      toast({
        title: '已复制',
        description: '完整地址已复制到剪贴板'
      })
    }
  }

  // 从历史记录搜索
  const searchFromHistory = (code: string) => {
    setSearchCode(code)
    searchOPCode(code)
  }

  // 获取访问级别信息
  const getAccessLevelInfo = (level: string) => {
    switch (level) {
      case 'public':
        return { label: '公开', color: 'bg-green-100 text-green-800', icon: Eye }
      case 'partial':
        return { label: '部分可见', color: 'bg-yellow-100 text-yellow-800', icon: EyeOff }
      case 'private':
        return { label: '私密', color: 'bg-red-100 text-red-800', icon: Shield }
      default:
        return { label: '未知', color: 'bg-gray-100 text-gray-800', icon: Info }
    }
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <BackButton href="/delivery-guide" />
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
              <Search className="h-8 w-8" />
              OP Code查询
            </h1>
            <p className="text-gray-600 mt-2">快速查询和验证OP Code地址信息</p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 主要搜索区域 */}
        <div className="lg:col-span-2 space-y-6">
          {/* 搜索输入 */}
          <Card>
            <CardHeader>
              <CardTitle>地址查询</CardTitle>
              <CardDescription>输入6位OP Code查询详细地址信息</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex gap-2">
                <div className="flex-1">
                  <Input
                    placeholder="输入OP Code，如：PK5F3D"
                    value={searchCode}
                    onChange={(e) => setSearchCode(e.target.value.toUpperCase())}
                    onKeyPress={handleKeyPress}
                    className="font-mono text-lg"
                    maxLength={6}
                  />
                </div>
                <Button 
                  onClick={handleSearch}
                  disabled={searching || !searchCode.trim()}
                >
                  {searching ? (
                    <>
                      <Clock className="w-4 h-4 mr-2 animate-spin" />
                      查询中...
                    </>
                  ) : (
                    <>
                      <Search className="w-4 h-4 mr-2" />
                      查询
                    </>
                  )}
                </Button>
              </div>

              {/* 格式提示 */}
              <div className="text-sm text-gray-600 bg-blue-50 p-3 rounded-lg">
                <div className="flex items-start gap-2">
                  <Info className="h-4 w-4 text-blue-600 mt-0.5" />
                  <div>
                    <p className="font-medium text-blue-900 mb-1">格式说明</p>
                    <p>OP Code由6位字母数字组成：</p>
                    <p><span className="font-mono">AA</span>(学校) + <span className="font-mono">BB</span>(区域) + <span className="font-mono">CC</span>(位置)</p>
                    <p className="mt-1 text-blue-700">示例：<span className="font-mono bg-white px-1">PK5F3D</span> = 北大 + 5号楼 + 303室</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* 搜索错误 */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* 搜索结果 */}
          {searchResult && (
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="flex items-center gap-2">
                    <MapPin className="h-5 w-5" />
                    查询结果
                  </CardTitle>
                  <div className="flex items-center gap-2">
                    {(() => {
                      const accessInfo = getAccessLevelInfo(searchResult.access_level)
                      return (
                        <Badge variant="outline" className={accessInfo.color}>
                          <accessInfo.icon className="w-3 h-3 mr-1" />
                          {accessInfo.label}
                        </Badge>
                      )
                    })()}
                    {searchResult.can_deliver ? (
                      <Badge className="bg-green-100 text-green-800">
                        <CheckCircle className="w-3 h-3 mr-1" />
                        可投递
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="bg-red-100 text-red-800">
                        <AlertCircle className="w-3 h-3 mr-1" />
                        限制投递
                      </Badge>
                    )}
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-6">
                {/* OP Code解析 */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="p-4 bg-blue-50 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="font-semibold text-blue-900">学校信息</h4>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => copyOPCode(searchResult.school_code)}
                        className="h-6 w-6 p-0"
                      >
                        <Copy className="h-3 w-3" />
                      </Button>
                    </div>
                    <p className="font-mono text-lg">{searchResult.school_code}</p>
                    <p className="text-sm text-blue-700">{searchResult.school_name}</p>
                  </div>
                  
                  <div className="p-4 bg-green-50 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="font-semibold text-green-900">区域信息</h4>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => copyOPCode(searchResult.area_code)}
                        className="h-6 w-6 p-0"
                      >
                        <Copy className="h-3 w-3" />
                      </Button>
                    </div>
                    <p className="font-mono text-lg">{searchResult.area_code}</p>
                    <p className="text-sm text-green-700">{searchResult.area_name}</p>
                  </div>
                  
                  <div className="p-4 bg-purple-50 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="font-semibold text-purple-900">位置信息</h4>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => copyOPCode(searchResult.location_code)}
                        className="h-6 w-6 p-0"
                      >
                        <Copy className="h-3 w-3" />
                      </Button>
                    </div>
                    <p className="font-mono text-lg">{searchResult.location_code}</p>
                    <p className="text-sm text-purple-700">{searchResult.location_name}</p>
                  </div>
                </div>

                {/* 完整地址 */}
                <div className="p-4 bg-gray-50 rounded-lg">
                  <div className="flex items-center justify-between mb-2">
                    <h4 className="font-semibold">完整地址</h4>
                    <Button variant="outline" size="sm" onClick={copyAddress}>
                      <Copy className="w-4 h-4 mr-2" />
                      复制地址
                    </Button>
                  </div>
                  <p className="text-lg">{searchResult.full_address}</p>
                </div>

                {/* 坐标信息 */}
                {searchResult.coordinates && (
                  <div className="p-4 bg-blue-50 rounded-lg">
                    <h4 className="font-semibold mb-2">地理坐标</h4>
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <p className="text-gray-600">纬度</p>
                        <p className="font-mono">{searchResult.coordinates.lat.toFixed(6)}</p>
                      </div>
                      <div>
                        <p className="text-gray-600">经度</p>
                        <p className="font-mono">{searchResult.coordinates.lng.toFixed(6)}</p>
                      </div>
                    </div>
                  </div>
                )}

                {/* 投递备注 */}
                {searchResult.delivery_notes && (
                  <Alert>
                    <Info className="h-4 w-4" />
                    <AlertDescription>
                      <strong>投递备注：</strong>{searchResult.delivery_notes}
                    </AlertDescription>
                  </Alert>
                )}

                {/* 联系信息（仅信使可见） */}
                {isCourier && searchResult.contact_info && (showPrivateInfo || searchResult.access_level !== 'private') && (
                  <div className="p-4 bg-yellow-50 rounded-lg border border-yellow-200">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="font-semibold text-yellow-900">联系信息</h4>
                      <Badge variant="outline" className="bg-yellow-100 text-yellow-800">
                        <Shield className="w-3 h-3 mr-1" />
                        信使专用
                      </Badge>
                    </div>
                    {searchResult.contact_info.phone && (
                      <div className="flex items-center gap-2 mb-1">
                        <Phone className="w-4 h-4 text-yellow-600" />
                        <span className="font-mono">{searchResult.contact_info.phone}</span>
                      </div>
                    )}
                    {searchResult.contact_info.office_hours && (
                      <div className="flex items-center gap-2">
                        <Clock className="w-4 h-4 text-yellow-600" />
                        <span>{searchResult.contact_info.office_hours}</span>
                      </div>
                    )}
                  </div>
                )}

                {/* 操作按钮 */}
                <div className="flex gap-2 pt-4 border-t">
                  {searchResult.coordinates && (
                    <Button variant="outline">
                      <Navigation className="w-4 h-4 mr-2" />
                      导航到此处
                    </Button>
                  )}
                  <Button variant="outline">
                    <QrCode className="w-4 h-4 mr-2" />
                    生成二维码
                  </Button>
                  <Button variant="outline">
                    <Map className="w-4 h-4 mr-2" />
                    查看地图
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        {/* 右侧边栏 */}
        <div className="space-y-6">
          {/* 搜索历史 */}
          {searchHistory.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">最近查询</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {searchHistory.map((item, index) => (
                  <div key={index} className="flex items-center justify-between">
                    <button
                      onClick={() => searchFromHistory(item.code)}
                      className="flex-1 text-left p-2 hover:bg-gray-50 rounded text-sm"
                    >
                      <span className="font-mono">{item.code}</span>
                      <span className="ml-2 text-xs text-gray-500">
                        {new Date(item.timestamp).toLocaleString('zh-CN', {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit'
                        })}
                      </span>
                    </button>
                    <div className="flex items-center gap-1">
                      {item.result === 'success' ? (
                        <CheckCircle className="w-3 h-3 text-green-500" />
                      ) : (
                        <AlertCircle className="w-3 h-3 text-red-500" />
                      )}
                    </div>
                  </div>
                ))}
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full"
                  onClick={() => {
                    setSearchHistory([])
                    localStorage.removeItem('opcode-search-history')
                  }}
                >
                  清除历史
                </Button>
              </CardContent>
            </Card>
          )}

          {/* 常用学校代码 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">常用学校代码</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {[
                { code: 'PK', name: '北京大学' },
                { code: 'QH', name: '清华大学' },
                { code: 'BD', name: '北京交通大学' },
                { code: 'CS', name: '中南大学' },
                { code: 'FD', name: '复旦大学' },
                { code: 'SJ', name: '上海交通大学' }
              ].map((school) => (
                <button
                  key={school.code}
                  onClick={() => setSearchCode(school.code)}
                  className="w-full text-left p-2 hover:bg-gray-50 rounded text-sm flex justify-between"
                >
                  <span>{school.name}</span>
                  <span className="font-mono text-gray-500">{school.code}</span>
                </button>
              ))}
            </CardContent>
          </Card>

          {/* 使用统计 */}
          {isCourier && (
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">今日使用情况</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span>查询次数</span>
                  <span className="font-bold">
                    {searchHistory.filter(h => 
                      new Date(h.timestamp).toDateString() === new Date().toDateString()
                    ).length}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span>成功率</span>
                  <span className="font-bold text-green-600">
                    {searchHistory.length > 0 
                      ? Math.round(searchHistory.filter(h => h.result === 'success').length / searchHistory.length * 100)
                      : 0}%
                  </span>
                </div>
              </CardContent>
            </Card>
          )}

          {/* 帮助信息 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">使用帮助</CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-gray-600 space-y-2">
              <p>• OP Code由6位字母数字组成</p>
              <p>• 前2位是学校代码</p>
              <p>• 中间2位是区域代码</p>
              <p>• 后2位是具体位置</p>
              <p>• 支持模糊匹配和历史记录</p>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}