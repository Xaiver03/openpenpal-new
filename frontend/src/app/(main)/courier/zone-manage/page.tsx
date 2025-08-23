/**
 * 二级信使管理页面 - 片区信使管理中心
 * Level 2 Courier Management Page - Zone Courier Management Center
 */

'use client'

import { useState, useEffect } from 'react'
import { CourierService } from '@/lib/api/courier-service'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Users, 
  MapPin, 
  TrendingUp, 
  Award, 
  Eye, 
  Edit, 
  UserPlus,
  Building,
  Home,
  Clock,
  Plus
} from 'lucide-react'
import { useCourierPermission } from '@/hooks/use-courier-permission'
import { SwipeableCard } from '@/components/ui/swipeable-card'
import { CreateCourierDialog } from '@/components/courier/CreateCourierDialog'
import { ManagementFloatingButton } from '@/components/courier/ManagementFloatingButton'
import { 
  ManagementPageLayout, 
  MANAGEMENT_CONFIGS, 
  createStatCard, 
  createFilterOptions, 
  createSortOptions,
  type StatCardData 
} from '@/components/courier/ManagementPageLayout'
import { 
  CourierPermissionGuard, 
  COURIER_PERMISSION_CONFIGS 
} from '@/components/courier/CourierPermissionGuard'
import { FeatureErrorBoundary } from '@/components/error-boundary'
import { CourierCenterNavigation } from '@/components/courier/CourierCenterNavigation'
import { log } from '@/utils/logger'

interface ZoneLevelCourier {
  id: string
  username: string
  buildingName: string
  buildingCode: string
  floorRange?: string
  roomRange?: string
  level: 1
  status: 'active' | 'pending' | 'frozen'
  points: number
  taskCount: number
  completedTasks: number
  averageRating: number
  joinDate: string
  lastActive: string
  contactInfo?: {
    phone?: string
    wechat?: string
  }
  workingHours?: {
    start: string
    end: string
    weekdays: number[]
  }
}

interface ZoneLevelStats {
  totalBuildings: number
  activeCouriers: number
  totalDeliveries: number
  pendingTasks: number
  averageRating: number
  successRate: number
}

export default function ZoneManagePage() {
  const [stats, setStats] = useState<ZoneLevelStats | null>(null)
  const [couriers, setCouriers] = useState<ZoneLevelCourier[]>([])
  const [filteredCouriers, setFilteredCouriers] = useState<ZoneLevelCourier[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [sortBy, setSortBy] = useState<string>('rating')
  const [showCreateDialog, setShowCreateDialog] = useState(false)

  const { 
    canCreateSubordinate, 
    canManageSubordinates, 
    canAssignTasks,
    getCourierLevelName 
  } = useCourierPermission()

  const loadData = async () => {
    setLoading(true)
    try {
        // 使用真实的API获取数据
        const courierStats = await CourierService.getCourierStats()
        
        // 映射API数据到页面所需格式
        setStats({
          totalBuildings: courierStats.teamStats?.totalMembers || 0,
          activeCouriers: courierStats.teamStats?.activeMembers || 0,  
          totalDeliveries: courierStats.teamStats?.totalDeliveries || 0,
          pendingTasks: courierStats.dailyStats?.pendingTasks || 0,
          averageRating: courierStats.courierInfo?.avgRating || 0,
          successRate: courierStats.courierInfo?.successRate || 0
        })

        // 获取下级信使列表
        try {
          const hierarchyInfo = await CourierService.getHierarchyInfo()
          console.log('获取层级信息成功:', hierarchyInfo)
          
          // 尝试获取下级信使列表
          const subordinatesResponse = await fetch('/api/v1/courier/subordinates', {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
          })
          
          if (subordinatesResponse.ok) {
            const subordinatesData = await subordinatesResponse.json()
            const apiCouriers = subordinatesData.data?.couriers || subordinatesData.couriers || []
            
            if (apiCouriers.length > 0) {
              // 转换API数据为UI格式 - 二级信使管理一级信使
              const uiCouriers: ZoneLevelCourier[] = apiCouriers
                .filter((courier: any) => courier.level === 1) 
                .map((courier: any) => ({
                  id: courier.id || `courier-${Date.now()}`,
                  username: courier.username || '未知用户',
                  buildingName: courier.building || '未分配楼栋',
                  buildingCode: courier.zone_code || 'UNASSIGNED',
                  floorRange: courier.floor_range || '1-10层',
                  roomRange: courier.room_range || '101-1020',
                  level: courier.level || 1,
                  status: courier.status === 'active' ? 'active' : 'pending',
                  points: courier.points || 0,
                  taskCount: (courier.completedTasks || 0) + (courier.currentTasks || 0),
                  completedTasks: courier.completedTasks || 0,
                  averageRating: courier.rating || 4.5,
                  joinDate: courier.createdAt || undefined,
                  lastActive: courier.lastActive || undefined,
                  contactInfo: courier.profile?.phone ? { 
                    phone: courier.profile.phone,
                    wechat: courier.profile.wechat
                  } : {},
                  workingHours: courier.workingHours
                }))
              setCouriers(uiCouriers)
              console.log('✅ 使用真实API数据:', uiCouriers)
              return
            }
          }
        } catch (error) {
          console.log('⚠️ API调用异常:', error)
        }

        // 如果API无数据或失败，设置空数组显示"暂无数据"状态
        setCouriers([])
    } catch (error) {
      log.error('Failed to load zone level management data', error, 'ZoneManagePage')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  useEffect(() => {
    let filtered = couriers.filter(courier => {
      const matchesSearch = courier.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           courier.buildingName.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           courier.buildingCode.toLowerCase().includes(searchTerm.toLowerCase())
      
      const matchesStatus = statusFilter === 'all' || courier.status === statusFilter
      
      return matchesSearch && matchesStatus
    })

    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'rating':
          return b.averageRating - a.averageRating
        case 'points':
          return b.points - a.points
        case 'tasks':
          return b.completedTasks - a.completedTasks
        case 'building':
          return a.buildingName.localeCompare(b.buildingName)
        case 'recent':
          return new Date(b.lastActive).getTime() - new Date(a.lastActive).getTime()
        default:
          return 0
      }
    })

    setFilteredCouriers(filtered)
  }, [couriers, searchTerm, statusFilter, sortBy])

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-600'
      case 'pending': return 'bg-yellow-600'
      case 'frozen': return 'bg-red-600'
      default: return 'bg-gray-600'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return '活跃'
      case 'pending': return '待审核'
      case 'frozen': return '冻结'
      default: return '未知'
    }
  }

  const formatLastActive = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffHours = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60))
    
    if (diffHours < 1) return '刚刚活跃'
    if (diffHours < 24) return `${diffHours}小时前`
    if (diffHours < 48) return '昨天'
    return `${Math.floor(diffHours / 24)}天前`
  }

  const handleCourierAction = (courierId: string, action: string) => {
    log.dev(`Action ${action} for courier ${courierId}`, { courierId, action }, 'ZoneManagePage')
  }

  const renderCourierCard = (courier: ZoneLevelCourier) => (
    <SwipeableCard
      key={courier.id}
      onSwipeLeft={() => handleCourierAction(courier.id, 'edit')}
      onSwipeRight={() => handleCourierAction(courier.id, 'assign')}
      className="mb-3 sm:mb-4"
    >
      <Card className="border-green-200 hover:border-green-300 transition-all duration-200 touch-manipulation">
        <CardContent className="p-3 sm:p-4">
          <div className="flex items-start justify-between mb-3">
            <div className="flex items-center space-x-2 sm:space-x-3">
              <div className={`p-1.5 sm:p-2 rounded-lg ${getStatusColor(courier.status)}`}>
                <Building className="w-3 h-3 sm:w-4 sm:h-4 text-white" />
              </div>
              <div className="min-w-0 flex-1">
                <h4 className="font-medium text-gray-900 truncate text-sm sm:text-base">
                  {courier.username}
                </h4>
                <p className="text-xs sm:text-sm text-gray-600 truncate">
                  {courier.buildingName} ({courier.buildingCode})
                </p>
                {courier.floorRange && (
                  <p className="text-xs text-gray-500 truncate">
                    {courier.floorRange} {courier.roomRange && `· ${courier.roomRange}`}
                  </p>
                )}
              </div>
            </div>
            <Badge 
              variant="secondary" 
              className={`${getStatusColor(courier.status)} text-white text-xs px-1.5 py-0.5 sm:px-2 sm:py-1`}
            >
              {getStatusText(courier.status)}
            </Badge>
          </div>

          {/* 统计信息 */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 sm:gap-4 mb-3">
            <div className="text-center">
              <div className="text-lg sm:text-xl font-bold text-green-600">{courier.points}</div>
              <div className="text-xs text-gray-500">积分</div>
            </div>
            <div className="text-center">
              <div className="text-lg sm:text-xl font-bold text-blue-600">{courier.completedTasks}</div>
              <div className="text-xs text-gray-500">完成任务</div>
            </div>
            <div className="text-center">
              <div className="text-lg sm:text-xl font-bold text-purple-600">{courier.averageRating.toFixed(1)}</div>
              <div className="text-xs text-gray-500">平均评分</div>
            </div>
            <div className="text-center">
              <div className="text-xs sm:text-sm text-gray-600">{formatLastActive(courier.lastActive)}</div>
              <div className="text-xs text-gray-500">最后活跃</div>
            </div>
          </div>

          {/* 工作时间和联系方式 */}
          {(courier.workingHours || courier.contactInfo) && (
            <div className="text-xs text-gray-500 space-y-1">
              {courier.workingHours && (
                <div className="flex items-center space-x-1">
                  <Clock className="w-3 h-3" />
                  <span>{courier.workingHours.start}-{courier.workingHours.end}</span>
                  <span>({courier.workingHours.weekdays.length}天/周)</span>
                </div>
              )}
              {courier.contactInfo?.phone && (
                <div className="flex items-center space-x-1">
                  <span>📱 {courier.contactInfo.phone}</span>
                  {courier.contactInfo.wechat && <span>· 微信: {courier.contactInfo.wechat}</span>}
                </div>
              )}
            </div>
          )}

          {/* 操作按钮 - 移动端优化 */}
          <div className="flex flex-wrap gap-1 sm:gap-2 mt-3">
            <Button
              size="sm"
              variant="outline"
              className="flex-1 sm:flex-none text-xs sm:text-sm touch-manipulation active:scale-95"
              onClick={() => handleCourierAction(courier.id, 'view')}
            >
              <Eye className="w-3 h-3 sm:w-4 sm:h-4 mr-1" />
              查看
            </Button>
            {canAssignTasks() && (
              <Button
                size="sm"
                variant="outline"
                className="flex-1 sm:flex-none text-xs sm:text-sm touch-manipulation active:scale-95"
                onClick={() => handleCourierAction(courier.id, 'assign')}
              >
                <UserPlus className="w-3 h-3 sm:w-4 sm:h-4 mr-1" />
                分配任务
              </Button>
            )}
            {canManageSubordinates() && (
              <Button
                size="sm"
                variant="outline"
                className="flex-1 sm:flex-none text-xs sm:text-sm touch-manipulation active:scale-95"
                onClick={() => handleCourierAction(courier.id, 'edit')}
              >
                <Edit className="w-3 h-3 sm:w-4 sm:h-4 mr-1" />
                编辑
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </SwipeableCard>
  )

  const statCards: StatCardData[] = stats ? [
    createStatCard(
      <Building className="w-6 h-6 sm:w-8 sm:h-8 text-green-600" />,
      stats.totalBuildings,
      '管理楼栋',
      'text-green-600'
    ),
    createStatCard(
      <Users className="w-6 h-6 sm:w-8 sm:h-8 text-blue-600" />,
      stats.activeCouriers,
      '活跃信使',
      'text-blue-600'
    ),
    createStatCard(
      <MapPin className="w-6 h-6 sm:w-8 sm:h-8 text-purple-600" />,
      stats.totalDeliveries,
      '总配送数',
      'text-purple-600'
    ),
    createStatCard(
      <Clock className="w-6 h-6 sm:w-8 sm:h-8 text-orange-600" />,
      stats.pendingTasks,
      '待处理任务',
      'text-orange-600'
    ),
    createStatCard(
      <Award className="w-6 h-6 sm:w-8 sm:h-8 text-amber-600" />,
      stats.averageRating.toFixed(1),
      '平均评分',
      'text-amber-600'
    ),
    createStatCard(
      <TrendingUp className="w-6 h-6 sm:w-8 sm:h-8 text-emerald-600" />,
      `${stats.successRate.toFixed(1)}%`,
      '成功率',
      'text-emerald-600'
    )
  ] : []

  const filterOptions = createFilterOptions([
    { value: 'active', label: '活跃' },
    { value: 'pending', label: '待审核' },
    { value: 'frozen', label: '冻结' }
  ])

  const sortOptions = createSortOptions([
    { value: 'rating', label: '按评分' },
    { value: 'points', label: '按积分' },
    { value: 'tasks', label: '按任务数' },
    { value: 'building', label: '按楼栋' },
    { value: 'recent', label: '按活跃度' }
  ])

  const courierListContent = (
    <>
      {filteredCouriers.length > 0 ? (
        filteredCouriers.map(renderCourierCard)
      ) : (
        <Card className="border-green-200">
          <CardContent className="p-6 sm:p-8 text-center">
            <Users className="w-12 h-12 sm:w-16 sm:h-16 text-green-400 mx-auto mb-4" />
            <h3 className="text-lg sm:text-xl font-medium text-gray-900 mb-2">暂无信使数据</h3>
            <p className="text-sm sm:text-base text-gray-600 mb-4">
              {searchTerm || statusFilter !== 'all' 
                ? '没有找到符合条件的信使，请尝试调整搜索条件'
                : '还没有添加一级信使，点击添加按钮开始管理'
              }
            </p>
            {canCreateSubordinate() && !searchTerm && statusFilter === 'all' && (
              <Button 
                onClick={() => setShowCreateDialog(true)}
                className="bg-green-600 hover:bg-green-700 text-white"
              >
                <Plus className="w-4 h-4 mr-2" />
                添加第一个信使
              </Button>
            )}
          </CardContent>
        </Card>
      )}
    </>
  )

  return (
    <FeatureErrorBoundary>
      <CourierPermissionGuard 
        config={COURIER_PERMISSION_CONFIGS.SECOND_LEVEL_MANAGEMENT}
        errorTitle="二级信使管理权限不足"
        errorDescription="只有三级及以上信使才能管理二级信使"
      >
        {/* 信使中心导航 */}
        <CourierCenterNavigation currentPage="zone" className="mb-6" />

        <ManagementPageLayout
        config={MANAGEMENT_CONFIGS.SECOND_LEVEL}
        stats={statCards}
        searchPlaceholder="搜索信使用户名、楼栋名称或编号..."
        searchValue={searchTerm}
        onSearchChange={setSearchTerm}
        filterOptions={filterOptions}
        filterValue={statusFilter}
        onFilterChange={setStatusFilter}
        sortOptions={sortOptions}
        sortValue={sortBy}
        onSortChange={setSortBy}
        canCreate={canCreateSubordinate()}
        createButtonText="添加信使"
        onCreateClick={() => setShowCreateDialog(true)}
        isLoading={loading}
      >
        {courierListContent}
      </ManagementPageLayout>

      {/* 创建信使对话框 */}
      {showCreateDialog && (
        <CreateCourierDialog
          open={showCreateDialog}
          onOpenChange={(open) => setShowCreateDialog(open)}
          targetLevel={1}
          onSuccess={() => {
            // Reload data after successful creation
            loadData()
            setShowCreateDialog(false)
          }}
        />
      )}

        {/* 管理浮动按钮 */}
        <ManagementFloatingButton />
      </CourierPermissionGuard>
    </FeatureErrorBoundary>
  )
}