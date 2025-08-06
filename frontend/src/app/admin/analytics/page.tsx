'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { usePermission } from '@/hooks/use-permission'
import { 
  Users, 
  Mail, 
  TrendingUp, 
  Package, 
  Clock, 
  MapPin,
  BarChart3,
  PieChart,
  LineChart,
  Activity
} from 'lucide-react'

// 简单的图表组件
function SimpleLineChart({ data, title }: { data: number[], title: string }) {
  const max = Math.max(...data)
  const height = 80
  
  return (
    <div className="space-y-2">
      <div className="text-sm font-medium text-gray-700">{title}</div>
      <div className="relative h-20 bg-gradient-to-t from-blue-50 to-transparent rounded border">
        <svg className="w-full h-full" viewBox={`0 0 ${data.length * 20} ${height}`}>
          <polyline
            points={data.map((value, index) => `${index * 20},${height - (value / max) * (height - 10)}`).join(' ')}
            fill="none"
            stroke="#3b82f6"
            strokeWidth="2"
            className="drop-shadow-sm"
          />
          {data.map((value, index) => (
            <circle
              key={index}
              cx={index * 20}
              cy={height - (value / max) * (height - 10)}
              r="3"
              fill="#3b82f6"
              className="hover:r-4 transition-all cursor-pointer"
            />
          ))}
        </svg>
      </div>
    </div>
  )
}

function SimpleBarChart({ data, labels, title }: { data: number[], labels: string[], title: string }) {
  const max = Math.max(...data)
  
  return (
    <div className="space-y-2">
      <div className="text-sm font-medium text-gray-700">{title}</div>
      <div className="space-y-2">
        {data.map((value, index) => (
          <div key={index} className="flex items-center gap-2">
            <div className="w-16 text-xs text-gray-600 truncate">{labels[index]}</div>
            <div className="flex-1 bg-gray-100 rounded-full h-4 relative overflow-hidden">
              <div 
                className="bg-gradient-to-r from-blue-500 to-blue-600 h-full rounded-full transition-all duration-500 ease-out"
                style={{ width: `${(value / max) * 100}%` }}
              />
            </div>
            <div className="w-12 text-xs text-gray-800 font-medium text-right">{value}</div>
          </div>
        ))}
      </div>
    </div>
  )
}

function SimplePieChart({ data, labels, colors }: { data: number[], labels: string[], colors: string[] }) {
  const total = data.reduce((sum, value) => sum + value, 0)
  let currentAngle = 0
  
  return (
    <div className="flex items-center justify-center space-x-6">
      <div className="relative">
        <svg width="120" height="120" viewBox="0 0 120 120">
          {data.map((value, index) => {
            const percentage = value / total
            const startAngle = currentAngle
            currentAngle += percentage * 360
            
            const startAngleRad = (startAngle - 90) * (Math.PI / 180)
            const endAngleRad = (currentAngle - 90) * (Math.PI / 180)
            
            const x1 = 60 + 50 * Math.cos(startAngleRad)
            const y1 = 60 + 50 * Math.sin(startAngleRad)
            const x2 = 60 + 50 * Math.cos(endAngleRad)
            const y2 = 60 + 50 * Math.sin(endAngleRad)
            
            const largeArcFlag = percentage > 0.5 ? 1 : 0
            
            const pathData = [
              'M', 60, 60,
              'L', x1, y1,
              'A', 50, 50, 0, largeArcFlag, 1, x2, y2,
              'Z'
            ].join(' ')
            
            return (
              <path
                key={index}
                d={pathData}
                fill={colors[index]}
                className="hover:opacity-80 transition-opacity cursor-pointer"
              />
            )
          })}
        </svg>
      </div>
      <div className="space-y-2">
        {labels.map((label, index) => (
          <div key={index} className="flex items-center gap-2 text-xs">
            <div 
              className="w-3 h-3 rounded-full" 
              style={{ backgroundColor: colors[index] }}
            />
            <span className="text-gray-600">{label}</span>
            <span className="font-medium">{data[index]}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

export default function AnalyticsPage() {
  const { user } = usePermission()
  const [timeRange, setTimeRange] = useState('7d')
  const [isLoading, setIsLoading] = useState(true)

  // 模拟数据加载
  useEffect(() => {
    const timer = setTimeout(() => setIsLoading(false), 1000)
    return () => clearTimeout(timer)
  }, [timeRange])

  // 模拟数据
  const userGrowthData = [120, 140, 180, 220, 260, 300, 340]
  const letterVolumeData = [45, 52, 48, 61, 55, 67, 73]
  const courierPerformanceData = [85, 92, 88, 95, 90, 97, 89]
  
  const regionData = [156, 89, 67, 45, 23]
  const regionLabels = ['华东', '华南', '华北', '西南', '西北']
  
  const statusData = [2340, 1890, 567, 123]
  const statusLabels = ['已送达', '运输中', '待收取', '异常']
  const statusColors = ['#10b981', '#f59e0b', '#3b82f6', '#ef4444']

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <div className="container mx-auto px-4 py-8">
          <div className="animate-pulse space-y-6">
            <div className="h-8 bg-gray-200 rounded w-1/4"></div>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
              {[1,2,3,4].map(i => (
                <div key={i} className="h-32 bg-gray-200 rounded"></div>
              ))}
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        {/* 页面标题 */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-2">
            <BarChart3 className="w-8 h-8 text-purple-600" />
            数据分析
          </h1>
          <p className="text-gray-600 mt-2">平台运营数据和趋势分析</p>
        </div>

        {/* 时间范围选择 */}
        <div className="mb-6">
          <div className="flex gap-2">
            {[
              { key: '7d', label: '近7天' },
              { key: '30d', label: '近30天' },
              { key: '90d', label: '近3个月' },
              { key: '1y', label: '近1年' }
            ].map(option => (
              <button
                key={option.key}
                onClick={() => setTimeRange(option.key)}
                className={`px-3 py-1 text-sm rounded-md transition-colors ${
                  timeRange === option.key 
                    ? 'bg-purple-600 text-white' 
                    : 'bg-white text-gray-600 hover:bg-gray-100'
                }`}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        {/* 核心指标卡片 */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <Card className="hover:shadow-md transition-shadow">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-600 text-sm">总用户数</p>
                  <p className="text-2xl font-bold text-gray-900">12,456</p>
                  <div className="flex items-center gap-1 mt-1">
                    <TrendingUp className="w-3 h-3 text-green-500" />
                    <span className="text-xs text-green-600">+12.5%</span>
                  </div>
                </div>
                <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center">
                  <Users className="w-6 h-6 text-blue-600" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="hover:shadow-md transition-shadow">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-600 text-sm">信件总数</p>
                  <p className="text-2xl font-bold text-gray-900">45,789</p>
                  <div className="flex items-center gap-1 mt-1">
                    <TrendingUp className="w-3 h-3 text-green-500" />
                    <span className="text-xs text-green-600">+18.2%</span>
                  </div>
                </div>
                <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
                  <Mail className="w-6 h-6 text-green-600" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="hover:shadow-md transition-shadow">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-600 text-sm">活跃信使</p>
                  <p className="text-2xl font-bold text-gray-900">1,234</p>
                  <div className="flex items-center gap-1 mt-1">
                    <Activity className="w-3 h-3 text-orange-500" />
                    <span className="text-xs text-orange-600">+5.8%</span>
                  </div>
                </div>
                <div className="w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center">
                  <Package className="w-6 h-6 text-orange-600" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="hover:shadow-md transition-shadow">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-600 text-sm">平均送达时间</p>
                  <p className="text-2xl font-bold text-gray-900">2.4h</p>
                  <div className="flex items-center gap-1 mt-1">
                    <Clock className="w-3 h-3 text-purple-500" />
                    <span className="text-xs text-purple-600">-0.3h</span>
                  </div>
                </div>
                <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center">
                  <Clock className="w-6 h-6 text-purple-600" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* 数据分析面板 */}
        <Tabs defaultValue="overview" className="space-y-6">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="overview">概览</TabsTrigger>
            <TabsTrigger value="users">用户分析</TabsTrigger>
            <TabsTrigger value="letters">信件分析</TabsTrigger>
            <TabsTrigger value="couriers">信使分析</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <LineChart className="w-5 h-5" />
                    用户增长趋势
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <SimpleLineChart data={userGrowthData} title="日新增用户数" />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <BarChart3 className="w-5 h-5" />
                    信件投递量
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <SimpleLineChart data={letterVolumeData} title="日投递信件数" />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <MapPin className="w-5 h-5" />
                    地区分布
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <SimpleBarChart 
                    data={regionData} 
                    labels={regionLabels} 
                    title="用户地区分布" 
                  />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <PieChart className="w-5 h-5" />
                    信件状态分布
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <SimplePieChart 
                    data={statusData}
                    labels={statusLabels}
                    colors={statusColors}
                  />
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="users" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle>用户活跃度</CardTitle>
                  <CardDescription>最近7天用户活跃情况</CardDescription>
                </CardHeader>
                <CardContent>
                  <SimpleLineChart data={[1200, 1450, 1680, 1520, 1890, 2100, 1950]} title="日活跃用户" />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>用户角色分布</CardTitle>
                  <CardDescription>平台用户角色统计</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex justify-between items-center p-2 bg-gray-50 rounded">
                      <span className="text-sm text-gray-600">普通用户</span>
                      <Badge variant="secondary">10,234</Badge>
                    </div>
                    <div className="flex justify-between items-center p-2 bg-gray-50 rounded">
                      <span className="text-sm text-gray-600">一级信使</span>
                      <Badge variant="secondary">856</Badge>
                    </div>
                    <div className="flex justify-between items-center p-2 bg-gray-50 rounded">
                      <span className="text-sm text-gray-600">二级信使</span>
                      <Badge variant="secondary">234</Badge>
                    </div>
                    <div className="flex justify-between items-center p-2 bg-gray-50 rounded">
                      <span className="text-sm text-gray-600">三级信使</span>
                      <Badge variant="secondary">89</Badge>
                    </div>
                    <div className="flex justify-between items-center p-2 bg-gray-50 rounded">
                      <span className="text-sm text-gray-600">四级信使</span>
                      <Badge variant="secondary">43</Badge>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="letters" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle>信件投递趋势</CardTitle>
                  <CardDescription>每日信件投递量变化</CardDescription>
                </CardHeader>
                <CardContent>
                  <SimpleLineChart data={letterVolumeData} title="日投递量" />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>信件类型分布</CardTitle>
                  <CardDescription>不同类型信件统计</CardDescription>
                </CardHeader>
                <CardContent>
                  <SimplePieChart 
                    data={[3200, 2100, 1500, 890]}
                    labels={['手写信件', '明信片', '包裹', '其他']}
                    colors={['#3b82f6', '#10b981', '#f59e0b', '#ef4444']}
                  />
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="couriers" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle>信使绩效趋势</CardTitle>
                  <CardDescription>平均信使评分变化</CardDescription>
                </CardHeader>
                <CardContent>
                  <SimpleLineChart data={courierPerformanceData} title="平均评分" />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>信使排行榜</CardTitle>
                  <CardDescription>本月表现最佳信使</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {[
                      { name: '张小明', score: 98, orders: 156 },
                      { name: '李小红', score: 96, orders: 142 },
                      { name: '王小刚', score: 95, orders: 138 },
                      { name: '刘小芳', score: 94, orders: 135 },
                      { name: '陈小华', score: 93, orders: 129 },
                    ].map((courier, index) => (
                      <div key={courier.name} className="flex items-center gap-3 p-2 bg-gray-50 rounded">
                        <div className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ${
                          index === 0 ? 'bg-yellow-100 text-yellow-600' :
                          index === 1 ? 'bg-gray-100 text-gray-600' :
                          index === 2 ? 'bg-orange-100 text-orange-600' :
                          'bg-blue-100 text-blue-600'
                        }`}>
                          {index + 1}
                        </div>
                        <div className="flex-1">
                          <div className="text-sm font-medium">{courier.name}</div>
                          <div className="text-xs text-gray-600">{courier.orders} 单投递</div>
                        </div>
                        <Badge variant="secondary">{courier.score}</Badge>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}