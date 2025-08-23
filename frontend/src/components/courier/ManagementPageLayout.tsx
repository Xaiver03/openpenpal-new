'use client'

import { ReactNode } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { LucideIcon, Search, Plus } from 'lucide-react'
import { SafeTimestamp } from '@/components/ui/safe-timestamp'

interface StatCard {
  icon: LucideIcon
  value: string | number
  label: string
  color?: 'default' | 'green' | 'orange' | 'amber'
}

interface TabConfig {
  id: string
  label: string
  content: ReactNode
}

// Legacy interfaces for backward compatibility
interface ManagementPageLayoutProps {
  title: string
  description: string
  titleIcon: LucideIcon
  stats: StatCard[]
  tabs: TabConfig[]
  defaultTab?: string
  children?: ReactNode
}

// 统一的信使卡片组件
interface CourierCardProps {
  id: string
  username: string
  locationName: string
  locationCode: string
  level: 1 | 2 | 3 | 4
  status: 'active' | 'pending' | 'frozen'
  points: number
  taskCount: number
  subordinateCount?: number
  averageRating: number
  joinDate: string
  lastActive: string
  onView: () => void
  onEdit?: () => void
  canEdit?: boolean
  actions?: ReactNode
}

export function CourierCard({
  id,
  username,
  locationName,
  locationCode,
  level,
  status,
  points,
  taskCount,
  subordinateCount,
  averageRating,
  joinDate,
  lastActive,
  onView,
  onEdit,
  canEdit = false,
  actions
}: CourierCardProps) {
  const getLevelBadge = (level: number) => {
    const levelMap = {
      1: { label: '一级信使', color: 'bg-yellow-600' },
      2: { label: '二级信使', color: 'bg-orange-600' },
      3: { label: '三级信使', color: 'bg-amber-600' },
      4: { label: '四级信使', color: 'bg-red-600' }
    }
    return levelMap[level as keyof typeof levelMap] || levelMap[1]
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-100 text-green-800'
      case 'pending': return 'bg-yellow-100 text-yellow-800'
      case 'frozen': return 'bg-red-100 text-red-800'
      default: return 'bg-gray-100 text-gray-800'
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

  const levelBadge = getLevelBadge(level)

  return (
    <Card className="border-amber-200 hover:border-amber-400 transition-all touch-manipulation" onClick={onView}>
      <CardContent className="p-3 sm:p-4 md:p-6">
        <div className="flex flex-col sm:flex-row items-start gap-3 sm:gap-4">
          {/* 头像和基础信息 */}
          <div className="flex items-start gap-3 w-full sm:flex-1">
            <div className={`w-10 h-10 sm:w-12 sm:h-12 ${levelBadge.color} text-white rounded-full flex items-center justify-center font-bold text-sm sm:text-base flex-shrink-0`}>
              {username.charAt(0)}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex flex-wrap items-center gap-1 sm:gap-2 mb-2">
                <h3 className="font-semibold text-amber-900 text-sm sm:text-base truncate">{username}</h3>
                <Badge className={`${levelBadge.color} text-white text-xs`}>
                  {levelBadge.label}
                </Badge>
                <Badge className={`${getStatusColor(status)} text-xs`}>
                  {getStatusText(status)}
                </Badge>
              </div>
              
              {/* 详细信息 */}
              <div className="text-xs sm:text-sm text-amber-700 space-y-1 sm:space-y-1.5">
                <div className="flex items-center gap-1">
                  <span className="truncate">{locationName}</span>
                  <span className="text-amber-500">({locationCode})</span>
                </div>
                {subordinateCount !== undefined && (
                  <div className="flex items-center gap-1">
                    <span>管理 {subordinateCount} 位下级信使</span>
                  </div>
                )}
                <div className="grid grid-cols-2 sm:flex sm:items-center gap-2 sm:gap-4 mt-2">
                  <div className="flex items-center gap-1">
                    <span>{points} 积分</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <span>{taskCount} 个任务</span>
                  </div>
                  <div className="flex items-center gap-1 col-span-2 sm:col-span-1">
                    <span>评分: {averageRating}/5.0</span>
                  </div>
                </div>
                <div className="text-xs text-amber-600 flex items-center gap-1">
                  <span>入职:</span>
                  {joinDate ? (
                    <SafeTimestamp 
                      date={joinDate} 
                      format="locale" 
                      fallback="--"
                      className="inline"
                    />
                  ) : (
                    <span>--</span>
                  )}
                  <span>|</span>
                  <span>最后活跃:</span>
                  {lastActive ? (
                    <SafeTimestamp 
                      date={lastActive} 
                      format="locale" 
                      fallback="--"
                      className="inline"
                    />
                  ) : (
                    <span>--</span>
                  )}
                </div>
              </div>
            </div>
          </div>
          
          {/* 操作按钮或自定义操作 */}
          <div className="hidden sm:flex gap-2 sm:flex-col lg:flex-row">
            {actions || (
              <>
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    onView()
                  }}
                  className="px-3 py-1.5 text-xs sm:text-sm border border-amber-300 text-amber-700 hover:bg-amber-50 rounded touch-manipulation active:scale-95 transition-all"
                >
                  详情
                </button>
                {canEdit && onEdit && (
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      onEdit()
                    }}
                    className="px-3 py-1.5 text-xs sm:text-sm border border-amber-300 text-amber-700 hover:bg-amber-50 rounded touch-manipulation active:scale-95 transition-all"
                  >
                    编辑
                  </button>
                )}
              </>
            )}
          </div>
          
          {/* 移动端手势提示 */}
          <div className="sm:hidden text-xs text-amber-600 text-center w-full mt-2 opacity-60">
            点击查看详情 {canEdit && '| 长按编辑'}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

// Enhanced Management Page Layout with filters and search
export interface StatCardData {
  icon: ReactNode
  value: string | number
  label: string
  color?: string
}

export interface FilterOption {
  value: string
  label: string
}

export interface SortOption {
  value: string
  label: string
}

export interface ManagementConfig {
  title: string
  description: string
  icon: LucideIcon
}

// Configuration constants
export const MANAGEMENT_CONFIGS = {
  FIRST_LEVEL: {
    title: '一级信使管理',
    description: '管理楼栋配送信使',
    icon: Search
  },
  SECOND_LEVEL: {
    title: '二级信使管理', 
    description: '管理片区信使',
    icon: Search
  },
  THIRD_LEVEL: {
    title: '三级信使管理',
    description: '管理学校信使',
    icon: Search
  },
  FOURTH_LEVEL: {
    title: '四级信使管理',
    description: '管理城市信使',
    icon: Search
  }
}

// Utility functions
export function createStatCard(
  icon: ReactNode,
  value: string | number,
  label: string,
  color?: string
): StatCardData {
  return { icon, value, label, color }
}

export function createFilterOptions(options: FilterOption[]): FilterOption[] {
  return [{ value: 'all', label: '全部' }, ...options]
}

export function createSortOptions(options: SortOption[]): SortOption[] {
  return options
}

// Enhanced Management Page Layout Props
interface EnhancedManagementPageLayoutProps {
  config: ManagementConfig
  stats: StatCardData[]
  searchPlaceholder: string
  searchValue: string
  onSearchChange: (value: string) => void
  filterOptions: FilterOption[]
  filterValue: string
  onFilterChange: (value: string) => void
  sortOptions: SortOption[]
  sortValue: string
  onSortChange: (value: string) => void
  canCreate?: boolean
  createButtonText?: string
  onCreateClick?: () => void
  isLoading?: boolean
  children: ReactNode
  additionalTabs?: Array<{
    id: string
    label: string
    content: ReactNode
  }>
}

export function ManagementPageLayout({
  config,
  stats,
  searchPlaceholder,
  searchValue,
  onSearchChange,
  filterOptions,
  filterValue,
  onFilterChange,
  sortOptions,
  sortValue,
  onSortChange,
  canCreate = false,
  createButtonText = '创建',
  onCreateClick,
  isLoading = false,
  children
}: EnhancedManagementPageLayoutProps) {
  const Icon = config.icon

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-7xl mx-auto px-4 py-8">
        {/* 页面标题 */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <Icon className="w-8 h-8 text-amber-600" />
            <h1 className="text-3xl font-bold text-amber-900">{config.title}</h1>
          </div>
          <p className="text-amber-700">{config.description}</p>
        </div>

        {/* 统计卡片 - 响应式布局 */}
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-2 sm:gap-4 mb-6 sm:mb-8">
          {stats.map((stat, index) => (
            <Card key={index} className="border-amber-200 touch-manipulation">
              <CardContent className="p-2 sm:p-4 text-center">
                <div className="mb-1 sm:mb-2">
                  {stat.icon}
                </div>
                <div className={`text-lg sm:text-xl md:text-2xl font-bold ${stat.color || 'text-amber-900'}`}>
                  {typeof stat.value === 'number' ? stat.value.toLocaleString() : stat.value}
                </div>
                <div className="text-xs sm:text-sm text-amber-600">{stat.label}</div>
              </CardContent>
            </Card>
          ))}
        </div>

        {/* 搜索和过滤控件 */}
        <div className="mb-6 bg-white rounded-lg border border-amber-200 p-4">
          <div className="flex flex-col sm:flex-row gap-4">
            {/* 搜索框 */}
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
              <Input
                placeholder={searchPlaceholder}
                value={searchValue}
                onChange={(e) => onSearchChange(e.target.value)}
                className="pl-10"
              />
            </div>
            
            {/* 过滤器 */}
            <div className="flex gap-2">
              <Select value={filterValue} onValueChange={onFilterChange}>
                <SelectTrigger className="w-[120px]">
                  <SelectValue placeholder="状态" />
                </SelectTrigger>
                <SelectContent>
                  {filterOptions.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select value={sortValue} onValueChange={onSortChange}>
                <SelectTrigger className="w-[120px]">
                  <SelectValue placeholder="排序" />
                </SelectTrigger>
                <SelectContent>
                  {sortOptions.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              {canCreate && (
                <Button
                  onClick={onCreateClick}
                  className="bg-amber-600 hover:bg-amber-700 text-white"
                >
                  <Plus className="w-4 h-4 mr-2" />
                  {createButtonText}
                </Button>
              )}
            </div>
          </div>
        </div>

        {/* 主要内容区域 */}
        <div className="space-y-4">
          {isLoading ? (
            <div className="text-center py-8">
              <div className="text-amber-600">加载中...</div>
            </div>
          ) : (
            children
          )}
        </div>
      </div>
    </div>
  )
}