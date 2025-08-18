'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  MapPin, 
  Building2, 
  Home, 
  Package,
  Plus,
  Settings,
  AlertCircle,
  Shield,
  Users,
  School
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context-new'
import { apiClient } from '@/lib/api-client-enhanced'
import { useRouter } from 'next/navigation'
import { CourierPermissionGuard } from '@/components/courier/CourierPermissionGuard'
import { DistrictManagement } from '@/components/courier/opcode/DistrictManagement'
import { BuildingManagement } from '@/components/courier/opcode/BuildingManagement'
import { DeliveryPointManagement } from '@/components/courier/opcode/DeliveryPointManagement'
import { OPCodeOverview } from '@/components/courier/opcode/OPCodeOverview'

export default function OPCodeManagePage() {
  const { user } = useAuth()
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [courierInfo, setCourierInfo] = useState<any>(null)
  const [activeTab, setActiveTab] = useState('overview')

  useEffect(() => {
    fetchCourierInfo()
  }, [])

  const fetchCourierInfo = async () => {
    try {
      const response = await apiClient.get('/api/v1/courier/profile')
      if ((response.data as any).success) {
        setCourierInfo((response.data as any).data)
      }
    } catch (error) {
      console.error('Failed to fetch courier info:', error)
    } finally {
      setLoading(false)
    }
  }

  const getCourierLevelName = (level: number) => {
    const levelNames = {
      1: '一级信使 - 楼栋投递员',
      2: '二级信使 - 片区管理员',
      3: '三级信使 - 学校协调员',
      4: '四级信使 - 城市总监'
    }
    return levelNames[level as keyof typeof levelNames] || '未知级别'
  }

  const getLevelColor = (level: number) => {
    const colors = {
      1: 'bg-blue-100 text-blue-800',
      2: 'bg-green-100 text-green-800',
      3: 'bg-purple-100 text-purple-800',
      4: 'bg-amber-100 text-amber-800'
    }
    return colors[level as keyof typeof colors] || 'bg-gray-100 text-gray-800'
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-amber-600"></div>
      </div>
    )
  }

  if (!courierInfo) {
    return (
      <div className="container max-w-4xl mx-auto px-4 py-8">
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            您不是信使，无法访问此页面
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <CourierPermissionGuard requiredLevel={1}>
      <div className="container max-w-7xl mx-auto px-4 py-8">
        {/* 页面标题 */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">OP Code 地理编码管理</h1>
              <p className="text-gray-600">
                管理您权限范围内的地理位置编码
              </p>
            </div>
            <Badge className={getLevelColor(courierInfo.level)}>
              <Shield className="w-4 h-4 mr-1" />
              {getCourierLevelName(courierInfo.level)}
            </Badge>
          </div>

          {/* 权限说明 */}
          <Alert className="bg-blue-50 border-blue-200">
            <AlertCircle className="h-4 w-4 text-blue-600" />
            <AlertDescription className="text-blue-800">
              {courierInfo.level === 4 && "作为城市总监，您可以管理所有学校的地理编码"}
              {courierInfo.level === 3 && `作为学校协调员，您可以管理 ${courierInfo.managedOPCodePrefix || courierInfo.zoneCode} 学校的所有片区和楼栋`}
              {courierInfo.level === 2 && `作为片区管理员，您可以管理 ${courierInfo.managedOPCodePrefix || courierInfo.zoneCode} 片区的所有楼栋和投递点`}
              {courierInfo.level === 1 && `作为楼栋投递员，您可以查看 ${courierInfo.managedOPCodePrefix || courierInfo.zoneCode} 楼栋的投递点`}
            </AlertDescription>
          </Alert>
        </div>

        {/* 功能标签页 */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="grid grid-cols-4 w-full max-w-2xl">
            <TabsTrigger value="overview">
              <MapPin className="w-4 h-4 mr-2" />
              概览
            </TabsTrigger>
            {courierInfo.level >= 3 && (
              <TabsTrigger value="districts">
                <Building2 className="w-4 h-4 mr-2" />
                片区管理
              </TabsTrigger>
            )}
            {courierInfo.level >= 2 && (
              <TabsTrigger value="buildings">
                <Home className="w-4 h-4 mr-2" />
                楼栋管理
              </TabsTrigger>
            )}
            <TabsTrigger value="delivery-points">
              <Package className="w-4 h-4 mr-2" />
              投递点管理
            </TabsTrigger>
          </TabsList>

          {/* 概览标签 */}
          <TabsContent value="overview">
            <OPCodeOverview courierInfo={courierInfo} />
          </TabsContent>

          {/* 片区管理标签 - L3/L4信使可见 */}
          {courierInfo.level >= 3 && (
            <TabsContent value="districts">
              <DistrictManagement courierInfo={courierInfo} />
            </TabsContent>
          )}

          {/* 楼栋管理标签 - L2及以上信使可见 */}
          {courierInfo.level >= 2 && (
            <TabsContent value="buildings">
              <BuildingManagement courierInfo={courierInfo} />
            </TabsContent>
          )}

          {/* 投递点管理标签 - 所有信使可见 */}
          <TabsContent value="delivery-points">
            <DeliveryPointManagement courierInfo={courierInfo} />
          </TabsContent>
        </Tabs>
      </div>
    </CourierPermissionGuard>
  )
}