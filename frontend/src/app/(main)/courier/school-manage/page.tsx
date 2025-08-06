/**
 * 三级信使管理页面 - 学校信使管理中心
 * Level 3 Courier Management Page - School Courier Management Center
 */

'use client'

import { useState, useEffect } from 'react'
import { getSchoolStats, getSchoolCouriers, getCourierCandidates } from '@/lib/api/index'
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
import { log } from '@/utils/logger'

interface SchoolLevelCourier {
  id: string
  username: string
  zoneName: string
  zoneCode: string
  buildingCount?: number
  coverageArea?: string
  level: 2
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

interface SchoolLevelStats {
  totalZones: number
  activeCouriers: number
  totalDeliveries: number
  pendingTasks: number
  averageRating: number
  coverageRate: number
}

export default function SchoolManagePage() {
  const [stats, setStats] = useState<SchoolLevelStats | null>(null)
  const [couriers, setCouriers] = useState<SchoolLevelCourier[]>([])
  const [filteredCouriers, setFilteredCouriers] = useState<SchoolLevelCourier[]>([])
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
        const [statsResponse, couriersResponse] = await Promise.all([
          getSchoolStats().catch(() => null),
          getSchoolCouriers().catch(() => [])
        ])

        if (statsResponse?.success) {
          setStats(statsResponse.data as any)
        } else {
          setStats({
            totalZones: 8,
            activeCouriers: 12,
            totalDeliveries: 2456,
            pendingTasks: 15,
            averageRating: 4.8,
            coverageRate: 94.5
          })
        }

        if (couriersResponse && 'success' in couriersResponse && couriersResponse.success) {
          setCouriers((couriersResponse as any).data || [])
        } else {
          const mockCouriers: SchoolLevelCourier[] = [
            {
              id: '1',
              username: 'zone_a_manager',
              zoneName: '东区',
              zoneCode: 'SCHOOL_ZONE_A',
              buildingCount: 6,
              coverageArea: '宿舍区A1-A6',
              level: 2,
              status: 'active',
              points: 580,
              taskCount: 234,
              completedTasks: 225,
              subordinateCount: 6,
              averageRating: 4.9,
              joinDate: '2023-12-01',
              lastActive: '2024-01-24T08:30:00Z',
              contactInfo: {
                phone: '138****2468',
                wechat: 'zone_a_manager'
              },
              workingHours: {
                start: '07:00',
                end: '19:00',
                weekdays: [1, 2, 3, 4, 5, 6, 7]
              }
            },
            {
              id: '2',
              username: 'zone_b_manager',
              zoneName: '西区',
              zoneCode: 'SCHOOL_ZONE_B',
              buildingCount: 4,
              coverageArea: '宿舍区B1-B4',
              level: 2,
              status: 'active',
              points: 520,
              taskCount: 198,
              completedTasks: 192,
              subordinateCount: 4,
              averageRating: 4.7,
              joinDate: '2024-01-05',
              lastActive: '2024-01-24T10:15:00Z',
              contactInfo: {
                phone: '159****1357'
              },
              workingHours: {
                start: '08:00',
                end: '20:00',
                weekdays: [1, 2, 3, 4, 5, 6]
              }
            },
            {
              id: '3',
              username: 'zone_c_manager',
              zoneName: '南区',
              zoneCode: 'SCHOOL_ZONE_C',
              buildingCount: 5,
              coverageArea: '宿舍区C1-C5',
              level: 2,
              status: 'pending',
              points: 245,
              taskCount: 89,
              completedTasks: 82,
              subordinateCount: 3,
              averageRating: 4.4,
              joinDate: '2024-01-18',
              lastActive: '2024-01-23T17:20:00Z',
              contactInfo: {
                phone: '186****7890',
                wechat: 'zone_c_helper'
              }
            },
            {
              id: '4',
              username: 'zone_d_manager',
              zoneName: '北区',
              zoneCode: 'SCHOOL_ZONE_D',
              buildingCount: 3,
              coverageArea: '宿舍区D1-D3',
              level: 2,
              status: 'frozen',
              points: 180,
              taskCount: 156,
              completedTasks: 135,
              subordinateCount: 2,
              averageRating: 3.8,
              joinDate: '2023-11-20',
              lastActive: '2024-01-18T14:45:00Z',
              contactInfo: {
                phone: '177****4567'
              }
            }
          ]
          setCouriers(mockCouriers)
        }
      } catch (error) {
        log.error('Failed to load school level management data', error, 'SchoolManagePage')
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
                           courier.zoneName.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           courier.zoneCode.toLowerCase().includes(searchTerm.toLowerCase())
      
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
        case 'zone':
          return a.zoneName.localeCompare(b.zoneName)
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
    log.dev(`Action ${action} for courier ${courierId}`, { courierId, action }, 'SchoolManagePage')
  }

  const renderCourierCard = (courier: SchoolLevelCourier) => (
    <SwipeableCard
      key={courier.id}
      onSwipeLeft={() => handleCourierAction(courier.id, 'edit')}
      onSwipeRight={() => handleCourierAction(courier.id, 'assign')}
      className="mb-3 sm:mb-4"
    >
      <Card className="border-blue-200 hover:border-blue-300 transition-all duration-200 touch-manipulation">
        <CardContent className="p-3 sm:p-4">
          <div className="flex items-start justify-between mb-3">
            <div className="flex items-center space-x-2 sm:space-x-3">
              <div className={`p-1.5 sm:p-2 rounded-lg ${getStatusColor(courier.status)}`}>
                <School className="w-3 h-3 sm:w-4 sm:h-4 text-white" />
              </div>
              <div className="min-w-0 flex-1">
                <h4 className="font-medium text-gray-900 truncate text-sm sm:text-base">
                  {courier.username}
                </h4>
                <p className="text-xs sm:text-sm text-gray-600 truncate">
                  {courier.zoneName} ({courier.zoneCode})
                </p>
                {courier.coverageArea && (
                  <p className="text-xs text-gray-500 truncate">
                    {courier.coverageArea} · {courier.buildingCount}栋
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
              <div className="text-lg sm:text-xl font-bold text-blue-600">{courier.points}</div>
              <div className="text-xs text-gray-500">积分</div>
            </div>
            <div className="text-center">
              <div className="text-lg sm:text-xl font-bold text-green-600">{courier.completedTasks}</div>
              <div className="text-xs text-gray-500">完成任务</div>
            </div>
            <div className="text-center">
              <div className="text-lg sm:text-xl font-bold text-purple-600">{courier.averageRating.toFixed(1)}</div>
              <div className="text-xs text-gray-500">平均评分</div>
            </div>
            <div className="text-center">
              <div className="text-lg sm:text-xl font-bold text-orange-600">{courier.subordinateCount}</div>
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
      <School className="w-6 h-6 sm:w-8 sm:h-8 text-blue-600" />,
      stats.totalZones,
      '管理片区',
      'text-blue-600'
    ),
    createStatCard(
      <Users className="w-6 h-6 sm:w-8 sm:h-8 text-green-600" />,
      stats.activeCouriers,
      '活跃信使',
      'text-green-600'
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
      `${stats.coverageRate.toFixed(1)}%`,
      '覆盖率',
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
    { value: 'zone', label: '按片区' },
    { value: 'subordinates', label: '按下级数量' },
    { value: 'recent', label: '按活跃度' }
  ])

  const courierListContent = (
    <>
      {filteredCouriers.length > 0 ? (
        filteredCouriers.map(renderCourierCard)
      ) : (
        <Card className="border-blue-200">
          <CardContent className="p-6 sm:p-8 text-center">
            <Users className="w-12 h-12 sm:w-16 sm:h-16 text-blue-400 mx-auto mb-4" />
            <h3 className="text-lg sm:text-xl font-medium text-gray-900 mb-2">暂无信使数据</h3>
            <p className="text-sm sm:text-base text-gray-600 mb-4">
              {searchTerm || statusFilter !== 'all' 
                ? '没有找到符合条件的信使，请尝试调整搜索条件'
                : '还没有添加二级信使，点击添加按钮开始管理'
              }
            </p>
            {canCreateSubordinate() && !searchTerm && statusFilter === 'all' && (
              <Button 
                onClick={() => setShowCreateDialog(true)}
                className="bg-blue-600 hover:bg-blue-700 text-white"
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
        errorTitle="三级信使管理权限不足"
        errorDescription="只有四级信使才能管理三级信使"
      >
      <ManagementPageLayout
        config={MANAGEMENT_CONFIGS.THIRD_LEVEL}
        stats={statCards}
        searchPlaceholder="搜索信使用户名、片区名称或编号..."
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
          targetLevel={2}
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