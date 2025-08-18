'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { 
  MapPin, 
  Building2, 
  Home, 
  Package,
  Users,
  TrendingUp,
  CheckCircle,
  AlertCircle
} from 'lucide-react'
import { apiClient } from '@/lib/api-client-enhanced'

interface OPCodeOverviewProps {
  courierInfo: any
}

export function OPCodeOverview({ courierInfo }: OPCodeOverviewProps) {
  const [stats, setStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchStats()
  }, [])

  const fetchStats = async () => {
    try {
      // 根据信使级别获取不同范围的统计数据
      const prefix = courierInfo.managedOPCodePrefix || courierInfo.zoneCode || ''
      const response = await apiClient.get(`/api/v1/opcode/stats/${prefix}`)
      
      if ((response.data as any).success) {
        setStats((response.data as any).data)
      }
    } catch (error) {
      console.error('Failed to fetch stats:', error)
      // 使用模拟数据
      setStats({
        totalSchools: courierInfo.level >= 4 ? 15 : 1,
        totalDistricts: courierInfo.level >= 3 ? 25 : courierInfo.level >= 2 ? 5 : 1,
        totalBuildings: courierInfo.level >= 2 ? 120 : 20,
        totalDeliveryPoints: 600,
        activeDeliveryPoints: 450,
        occupiedPoints: 320,
        availablePoints: 130,
        recentApplications: 12
      })
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-amber-600"></div>
      </div>
    )
  }

  const occupancyRate = stats ? (stats.occupiedPoints / stats.totalDeliveryPoints * 100).toFixed(1) : 0

  return (
    <div className="space-y-6">
      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {courierInfo.level >= 4 && (
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">管理学校</CardTitle>
              <Building2 className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats?.totalSchools || 0}</div>
              <p className="text-xs text-muted-foreground">
                所有学校编码
              </p>
            </CardContent>
          </Card>
        )}

        {courierInfo.level >= 3 && (
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">片区数量</CardTitle>
              <MapPin className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats?.totalDistricts || 0}</div>
              <p className="text-xs text-muted-foreground">
                管理的片区
              </p>
            </CardContent>
          </Card>
        )}

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">楼栋数量</CardTitle>
            <Home className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.totalBuildings || 0}</div>
            <p className="text-xs text-muted-foreground">
              管理的楼栋
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">投递点总数</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.totalDeliveryPoints || 0}</div>
            <p className="text-xs text-muted-foreground">
              {stats?.activeDeliveryPoints || 0} 个活跃
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 使用情况 */}
      <Card>
        <CardHeader>
          <CardTitle>投递点使用情况</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span>占用率</span>
              <span className="font-medium">{occupancyRate}%</span>
            </div>
            <Progress value={Number(occupancyRate)} className="h-2" />
          </div>
          
          <div className="grid grid-cols-3 gap-4 pt-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">{stats?.availablePoints || 0}</div>
              <p className="text-xs text-muted-foreground">可用</p>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-amber-600">{stats?.occupiedPoints || 0}</div>
              <p className="text-xs text-muted-foreground">已占用</p>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-600">
                {(stats?.totalDeliveryPoints || 0) - (stats?.activeDeliveryPoints || 0)}
              </div>
              <p className="text-xs text-muted-foreground">未激活</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 最近活动 */}
      {courierInfo.level >= 2 && (
        <Card>
          <CardHeader>
            <CardTitle>最近申请</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <AlertCircle className="h-5 w-5 text-amber-500" />
                  <div>
                    <p className="text-sm font-medium">待审核申请</p>
                    <p className="text-xs text-muted-foreground">需要您的审核</p>
                  </div>
                </div>
                <Badge variant="secondary">{stats?.recentApplications || 0}</Badge>
              </div>
              
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <TrendingUp className="h-5 w-5 text-green-500" />
                  <div>
                    <p className="text-sm font-medium">本月新增</p>
                    <p className="text-xs text-muted-foreground">新激活的投递点</p>
                  </div>
                </div>
                <Badge variant="outline">+28</Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 管理提示 */}
      <Card className="bg-amber-50 border-amber-200">
        <CardHeader>
          <CardTitle className="text-amber-900">管理提示</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="space-y-2 text-sm text-amber-800">
            {courierInfo.level >= 3 && (
              <li className="flex items-start gap-2">
                <CheckCircle className="h-4 w-4 mt-0.5 flex-shrink-0" />
                <span>定期检查片区编码分配，确保覆盖所有区域</span>
              </li>
            )}
            {courierInfo.level >= 2 && (
              <li className="flex items-start gap-2">
                <CheckCircle className="h-4 w-4 mt-0.5 flex-shrink-0" />
                <span>及时处理楼栋编码申请，保持数据准确性</span>
              </li>
            )}
            <li className="flex items-start gap-2">
              <CheckCircle className="h-4 w-4 mt-0.5 flex-shrink-0" />
              <span>监控投递点使用情况，优化资源分配</span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle className="h-4 w-4 mt-0.5 flex-shrink-0" />
              <span>保持与其他信使的沟通，确保编码系统协调一致</span>
            </li>
          </ul>
        </CardContent>
      </Card>
    </div>
  )
}