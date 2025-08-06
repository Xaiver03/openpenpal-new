/**
 * 四级信使管理页面 - 城市信使管理中心
 * Level 4 Courier Management Page - City Courier Management Center
 */

'use client'

import { useState, useEffect } from 'react'
import { getCityStats, getCityCouriers, getCourierCandidates } from '@/lib/api/index'
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
  School,
  Crown,
  Clock,
  Plus
} from 'lucide-react'
import { useCourierPermission } from '@/hooks/use-courier-permission'
import { SwipeableCard } from '@/components/ui/swipeable-card'
import { CreateCourierDialog } from '@/components/courier/CreateCourierDialog'
import { ManagementFloatingButton } from '@/components/courier/ManagementFloatingButton'
import { PostcodeManagement } from '@/components/postcode/PostcodeManagement'
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
import { log } from '@/utils/logger'

interface CityLevelCourier {
  id: string
  username: string
  schoolName: string
  schoolCode: string
  zoneCount?: number
  coverage?: string
  level: 3
  status: 'active' | 'pending' | 'frozen'
  points: number
  taskCount: number
  completedTasks: number
  subordinateCount: number
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

interface CityLevelStats {
  totalSchools: number
  activeCouriers: number
  totalDeliveries: number
  pendingTasks: number
  averageRating: number
  successRate: number
}

export default function CityManagePage() {
  const [stats, setStats] = useState<CityLevelStats | null>(null)
  const [couriers, setCouriers] = useState<CityLevelCourier[]>([])
  const [filteredCouriers, setFilteredCouriers] = useState<CityLevelCourier[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [sortBy, setSortBy] = useState<string>('rating')
  const [showCreateDialog, setShowCreateDialog] = useState(false)

  const { 
    canCreateSubordinate, 
    canManageSubordinates, 
    canAssignTasks,
    getCourierLevelName,
    courierInfo
  } = useCourierPermission()
  
  // Debug logging
  console.log('🔍 City Manage Page Debug:', {
    courierInfo,
    canCreate: canCreateSubordinate(),
    canManage: canManageSubordinates(),
    canAssign: canAssignTasks()
  })

  const loadData = async () => {
    setLoading(true)
    try {
        const [statsResponse, couriersResponse] = await Promise.all([
          getCityStats().catch(() => null),
          getCityCouriers().catch(() => [])
        ])

        if (statsResponse?.success) {
          setStats(statsResponse.data as any)
        } else {
          setStats({
            totalSchools: 25,
            activeCouriers: 38,
            totalDeliveries: 12678,
            pendingTasks: 45,
            averageRating: 4.9,
            successRate: 97.2
          })
        }

        if (couriersResponse && 'success' in couriersResponse && couriersResponse.success) {
          setCouriers((couriersResponse as any).data || [])
        } else {
          const mockCouriers: CityLevelCourier[] = [
            {
              id: '1',
              username: 'university_peking_manager',
              schoolName: '北京大学',
              schoolCode: 'PKU_MAIN',
              zoneCount: 8,
              coverage: '燕园校区全区',
              level: 3,
              status: 'active',
              points: 1250,
              taskCount: 456,
              completedTasks: 445,
              subordinateCount: 8,
              averageRating: 4.95,
              joinDate: '2023-09-01',
              lastActive: '2024-01-24T07:45:00Z',
              contactInfo: {
                phone: '138****0001',
                wechat: 'pku_courier_head'
              },
              workingHours: {
                start: '06:00',
                end: '22:00',
                weekdays: [1, 2, 3, 4, 5, 6, 7]
              }
            },
            {
              id: '2',
              username: 'university_tsinghua_manager',
              schoolName: '清华大学',
              schoolCode: 'THU_MAIN',
              zoneCount: 6,
              coverage: '紫荆校区全区',
              level: 3,
              status: 'active',
              points: 1180,
              taskCount: 398,
              completedTasks: 390,
              subordinateCount: 6,
              averageRating: 4.88,
              joinDate: '2023-09-05',
              lastActive: '2024-01-24T08:20:00Z',
              contactInfo: {
                phone: '138****0002',
                wechat: 'thu_courier_head'
              },
              workingHours: {
                start: '06:30',
                end: '21:30',
                weekdays: [1, 2, 3, 4, 5, 6, 7]
              }
            },
            {
              id: '3',
              username: 'university_renda_manager',
              schoolName: '中国人民大学',
              schoolCode: 'RUC_MAIN',
              zoneCount: 4,
              coverage: '明德楼区域',
              level: 3,
              status: 'pending',
              points: 680,
              taskCount: 234,
              completedTasks: 225,
              subordinateCount: 4,
              averageRating: 4.72,
              joinDate: '2024-01-10',
              lastActive: '2024-01-23T19:30:00Z',
              contactInfo: {
                phone: '138****0003',
                wechat: 'ruc_courier_head'
              }
            },
            {
              id: '4',
              username: 'university_beishi_manager',
              schoolName: '北京师范大学',
              schoolCode: 'BNU_MAIN',
              zoneCount: 5,
              coverage: '北师大校区',
              level: 3,
              status: 'frozen',
              points: 420,
              taskCount: 189,
              completedTasks: 165,
              subordinateCount: 3,
              averageRating: 4.35,
              joinDate: '2023-11-15',
              lastActive: '2024-01-15T16:20:00Z',
              contactInfo: {
                phone: '138****0004'
              }
            }
          ]
          setCouriers(mockCouriers)
        }
      } catch (error) {
        log.error('Failed to load city level management data', error, 'CityManagePage')
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
                           courier.schoolName.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           courier.schoolCode.toLowerCase().includes(searchTerm.toLowerCase())
      
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
        case 'school':
          return a.schoolName.localeCompare(b.schoolName)
        case 'subordinates':
          return b.subordinateCount - a.subordinateCount
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
    log.dev(`Action ${action} for courier ${courierId}`, { courierId, action }, 'CityManagePage')
  }

  const renderCourierCard = (courier: CityLevelCourier) => (
    <SwipeableCard
      key={courier.id}
      onSwipeLeft={() => handleCourierAction(courier.id, 'edit')}
      onSwipeRight={() => handleCourierAction(courier.id, 'assign')}
      className="mb-3 sm:mb-4"
    >
      <Card className="border-purple-200 hover:border-purple-300 transition-all duration-200 touch-manipulation">
        <CardContent className="p-3 sm:p-4">
          <div className="flex items-start justify-between mb-3">
            <div className="flex items-center space-x-2 sm:space-x-3">
              <div className={`p-1.5 sm:p-2 rounded-lg ${getStatusColor(courier.status)}`}>
                <Crown className="w-3 h-3 sm:w-4 sm:h-4 text-white" />
              </div>
              <div className="min-w-0 flex-1">
                <h4 className="font-medium text-gray-900 truncate text-sm sm:text-base">
                  {courier.username}
                </h4>
                <p className="text-xs sm:text-sm text-gray-600 truncate">
                  {courier.schoolName} ({courier.schoolCode})
                </p>
                {courier.coverage && (
                  <p className="text-xs text-gray-500 truncate">
                    {courier.coverage} · {courier.zoneCount}个片区
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
              <div className="text-lg sm:text-xl font-bold text-purple-600">{courier.points}</div>
              <div className="text-xs text-gray-500">积分</div>
            </div>
            <div className="text-center">
              <div className="text-lg sm:text-xl font-bold text-green-600">{courier.completedTasks}</div>
              <div className="text-xs text-gray-500">完成任务</div>
            </div>
            <div className="text-center">
              <div className="text-lg sm:text-xl font-bold text-amber-600">{courier.averageRating.toFixed(2)}</div>
              <div className="text-xs text-gray-500">平均评分</div>
            </div>
            <div className="text-center">
              <div className="text-lg sm:text-xl font-bold text-blue-600">{courier.subordinateCount}</div>
              <div className="text-xs text-gray-500">下级信使</div>
            </div>
          </div>

          {/* 其他信息 */}
          <div className="text-xs text-gray-500 space-y-1 mb-3">
            <div className="flex items-center justify-between">
              <span>{formatLastActive(courier.lastActive)}</span>
              {courier.workingHours && (
                <span className="flex items-center space-x-1">
                  <Clock className="w-3 h-3" />
                  <span>{courier.workingHours.start}-{courier.workingHours.end}</span>
                </span>
              )}
            </div>
            {courier.contactInfo?.phone && (
              <div className="flex items-center space-x-1">
                <span>📱 {courier.contactInfo.phone}</span>
                {courier.contactInfo.wechat && <span>· 微信: {courier.contactInfo.wechat}</span>}
              </div>
            )}
          </div>

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
      <Crown className="w-6 h-6 sm:w-8 sm:h-8 text-purple-600" />,
      stats.totalSchools,
      '管理学校',
      'text-purple-600'
    ),
    createStatCard(
      <Users className="w-6 h-6 sm:w-8 sm:h-8 text-green-600" />,
      stats.activeCouriers,
      '活跃信使',
      'text-green-600'
    ),
    createStatCard(
      <MapPin className="w-6 h-6 sm:w-8 sm:h-8 text-blue-600" />,
      stats.totalDeliveries,
      '总配送数',
      'text-blue-600'
    ),
    createStatCard(
      <Clock className="w-6 h-6 sm:w-8 sm:h-8 text-orange-600" />,
      stats.pendingTasks,
      '待处理任务',
      'text-orange-600'
    ),
    createStatCard(
      <Award className="w-6 h-6 sm:w-8 sm:h-8 text-amber-600" />,
      stats.averageRating ? stats.averageRating.toFixed(1) : '0.0',
      '平均评分',
      'text-amber-600'
    ),
    createStatCard(
      <TrendingUp className="w-6 h-6 sm:w-8 sm:h-8 text-emerald-600" />,
      `${stats.successRate ? stats.successRate.toFixed(1) : '0.0'}%`,
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
    { value: 'school', label: '按学校' },
    { value: 'subordinates', label: '按下级数量' },
    { value: 'recent', label: '按活跃度' }
  ])

  const courierListContent = (
    <>
      {filteredCouriers.length > 0 ? (
        filteredCouriers.map(renderCourierCard)
      ) : (
        <Card className="border-purple-200">
          <CardContent className="p-6 sm:p-8 text-center">
            <Users className="w-12 h-12 sm:w-16 sm:h-16 text-purple-400 mx-auto mb-4" />
            <h3 className="text-lg sm:text-xl font-medium text-gray-900 mb-2">暂无信使数据</h3>
            <p className="text-sm sm:text-base text-gray-600 mb-4">
              {searchTerm || statusFilter !== 'all' 
                ? '没有找到符合条件的信使，请尝试调整搜索条件'
                : '还没有添加三级信使，点击添加按钮开始管理'
              }
            </p>
            {canCreateSubordinate() && !searchTerm && statusFilter === 'all' && (
              <Button 
                onClick={() => setShowCreateDialog(true)}
                className="bg-purple-600 hover:bg-purple-700 text-white"
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
        config={COURIER_PERMISSION_CONFIGS.THIRD_LEVEL_MANAGEMENT}
        errorTitle="城市管理权限不足"
        errorDescription="只有四级信使（城市总代）才能管理城市级信使"
      >
      <ManagementPageLayout
        config={MANAGEMENT_CONFIGS.FOURTH_LEVEL}
        stats={statCards}
        searchPlaceholder="搜索信使用户名、学校名称或编号..."
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
        additionalTabs={[
          {
            id: 'postcode',
            label: 'Postcode管理',
            content: <PostcodeManagement />
          }
        ]}
      >
        {courierListContent}
      </ManagementPageLayout>

      {/* 创建信使对话框 */}
      {showCreateDialog && (
        <CreateCourierDialog
          open={showCreateDialog}
          onOpenChange={(open) => setShowCreateDialog(open)}
          targetLevel={3}
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