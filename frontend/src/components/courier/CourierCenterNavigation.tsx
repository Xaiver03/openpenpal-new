/**
 * 信使中心导航组件 - 统一的级别管理导航
 */

'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  ArrowLeft, 
  Home, 
  Users, 
  Building, 
  School,
  Truck,
  ChevronRight,
  MapPin
} from 'lucide-react'
import { useCourierPermission } from '@/hooks/use-courier-permission'

interface CourierCenterNavigationProps {
  currentPage: 'home' | 'building' | 'zone' | 'school' | 'city'
  className?: string
}

export function CourierCenterNavigation({ 
  currentPage,
  className = ''
}: CourierCenterNavigationProps) {
  const router = useRouter()
  const { courierInfo, getCourierLevelName } = useCourierPermission()
  const courierLevel = courierInfo?.level || 0

  const navigationItems = [
    {
      key: 'home',
      title: '信使中心',
      description: '管理中心首页',
      icon: Home,
      href: '/courier',
      color: 'bg-blue-600',
      minLevel: 1
    },
    {
      key: 'building',
      title: '楼栋管理',
      description: '管理负责楼栋',
      icon: Building,
      href: '/courier/building-manage',
      color: 'bg-yellow-600',
      minLevel: 1,
      maxLevel: 1 // 只有一级信使显示
    },
    {
      key: 'zone', 
      title: '片区管理',
      description: '管理片区信使',
      icon: Truck,
      href: '/courier/zone-manage',
      color: 'bg-green-600',
      minLevel: 2,
      maxLevel: 2 // 只有二级信使显示
    },
    {
      key: 'school',
      title: '学校管理', 
      description: '管理校内信使',
      icon: School,
      href: '/courier/school-manage',
      color: 'bg-purple-600',
      minLevel: 3,
      maxLevel: 3 // 只有三级信使显示
    },
    {
      key: 'city',
      title: '城市管理',
      description: '管理城市信使',
      icon: Users,
      href: '/courier/city-manage', 
      color: 'bg-red-600',
      minLevel: 4,
      maxLevel: 4 // 只有四级信使显示
    }
  ]

  const availableItems = navigationItems.filter(item => {
    if (item.minLevel && courierLevel < item.minLevel) return false
    if (item.maxLevel && courierLevel > item.maxLevel) return false
    return true
  })

  const handleNavigation = (href: string) => {
    router.push(href)
  }

  const handleBackToCenter = () => {
    router.push('/courier')
  }

  const currentItem = navigationItems.find(item => item.key === currentPage)

  return (
    <Card className={`border-amber-200 bg-gradient-to-r from-amber-50 to-orange-50 ${className}`}>
      <CardContent className="p-4">
        {/* 顶部导航栏 */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-3">
            {currentPage !== 'home' && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleBackToCenter}
                className="flex items-center space-x-1"
              >
                <ArrowLeft className="w-4 h-4" />
                <span>返回中心</span>
              </Button>
            )}
            <div className="flex items-center space-x-2">
              {currentItem && (
                <>
                  <div className={`p-2 rounded-lg ${currentItem.color} text-white`}>
                    <currentItem.icon className="w-5 h-5" />
                  </div>
                  <div>
                    <h3 className="font-medium text-gray-900">{currentItem.title}</h3>
                    <p className="text-sm text-gray-600">{currentItem.description}</p>
                  </div>
                </>
              )}
            </div>
          </div>
          <Badge variant="secondary" className="bg-amber-100 text-amber-800">
            {getCourierLevelName()} (L{courierLevel})
          </Badge>
        </div>

        {/* 快速导航按钮 */}
        {currentPage !== 'home' && availableItems.length > 1 && (
          <div className="border-t pt-4">
            <div className="flex items-center mb-2">
              <MapPin className="w-4 h-4 text-amber-600 mr-2" />
              <span className="text-sm font-medium text-gray-700">快速切换</span>
            </div>
            <div className="flex flex-wrap gap-2">
              {availableItems
                .filter(item => item.key !== currentPage)
                .map((item) => (
                  <Button
                    key={item.key}
                    variant="outline"
                    size="sm"
                    onClick={() => handleNavigation(item.href)}
                    className="flex items-center space-x-1 text-xs"
                  >
                    <item.icon className="w-3 h-3" />
                    <span>{item.title}</span>
                    <ChevronRight className="w-3 h-3" />
                  </Button>
                ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}