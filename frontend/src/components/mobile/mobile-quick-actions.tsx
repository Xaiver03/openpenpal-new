'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  QrCode,
  Navigation,
  MapPin,
  Route,
  Package,
  Clock,
  Zap,
  Phone,
  Search,
  Building,
  Truck,
  CheckCircle,
  AlertCircle,
  Star,
  Plus,
  Menu,
  X,
  Camera,
  MessageCircle,
  Target
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuth } from '@/contexts/auth-context-new'
import { usePermission } from '@/hooks/use-permission'
import Link from 'next/link'

interface QuickAction {
  id: string
  title: string
  description: string
  icon: any
  href: string
  badge?: number | string
  color: string
  requiresCourier?: boolean
  priority: 'high' | 'medium' | 'low'
}

interface MobileQuickActionsProps {
  className?: string
  variant?: 'compact' | 'full' | 'grid'
  showTitle?: boolean
}

export function MobileQuickActions({ 
  className, 
  variant = 'full',
  showTitle = true
}: MobileQuickActionsProps) {
  const { user } = useAuth()
  const { hasPermission } = usePermission()
  const [isExpanded, setIsExpanded] = useState(false)
  const [isMobile, setIsMobile] = useState(false)

  // 检查是否是信使
  const isCourier = hasPermission('courier.basic') || 
                   hasPermission('courier.intermediate') || 
                   hasPermission('courier.advanced') || 
                   hasPermission('courier.management')

  // 检查移动端
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768)
    }
    
    checkMobile()
    window.addEventListener('resize', checkMobile)
    
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  // 信使快速操作
  const courierActions: QuickAction[] = [
    {
      id: 'scan',
      title: '扫码收件',
      description: '扫描QR码更新状态',
      icon: QrCode,
      href: '/courier/scan',
      color: 'bg-green-500 hover:bg-green-600 text-white',
      priority: 'high',
      requiresCourier: true
    },
    {
      id: 'tasks',
      title: '我的任务',
      description: '查看待办投递任务',
      icon: Package,
      href: '/courier/tasks',
      badge: 3, // 模拟未完成任务数
      color: 'bg-blue-500 hover:bg-blue-600 text-white',
      priority: 'high',
      requiresCourier: true
    },
    {
      id: 'route',
      title: '路线规划',
      description: '优化投递路径',
      icon: Route,
      href: '/delivery-guide/route-planner',
      color: 'bg-purple-500 hover:bg-purple-600 text-white',
      priority: 'high'
    },
    {
      id: 'navigate',
      title: '地址导航',
      description: 'OP Code地址查询',
      icon: MapPin,
      href: '/delivery-guide/opcode-search',
      color: 'bg-orange-500 hover:bg-orange-600 text-white',
      priority: 'medium'
    },
    {
      id: 'building',
      title: '建筑导航',
      description: '室内定位导航',
      icon: Building,
      href: '/delivery-guide/building-nav',
      color: 'bg-indigo-500 hover:bg-indigo-600 text-white',
      priority: 'medium'
    },
    {
      id: 'emergency',
      title: '紧急联系',
      description: '客服和支持热线',
      icon: Phone,
      href: 'tel:400-000-0000',
      color: 'bg-red-500 hover:bg-red-600 text-white',
      priority: 'low'
    }
  ]

  // 普通用户快速操作
  const userActions: QuickAction[] = [
    {
      id: 'write',
      title: '写信',
      description: '创建新的手写信件',
      icon: MessageCircle,
      href: '/letters/write',
      color: 'bg-blue-500 hover:bg-blue-600 text-white',
      priority: 'high'
    },
    {
      id: 'search',
      title: '查询地址',
      description: 'OP Code地址查询',
      icon: Search,
      href: '/delivery-guide/opcode-search',
      color: 'bg-green-500 hover:bg-green-600 text-white',
      priority: 'high'
    },
    {
      id: 'track',
      title: '追踪信件',
      description: '查看投递进度',
      icon: Package,
      href: '/letters/track',
      color: 'bg-purple-500 hover:bg-purple-600 text-white',
      priority: 'medium'
    },
    {
      id: 'shop',
      title: '信封商城',
      description: '选购精美信封',
      icon: Plus,
      href: '/shop',
      color: 'bg-pink-500 hover:bg-pink-600 text-white',
      priority: 'medium'
    }
  ]

  const actions = isCourier ? courierActions : userActions
  const visibleActions = actions.filter(action => !action.requiresCourier || isCourier)

  // 根据变体决定显示的操作数量
  const getDisplayActions = () => {
    switch (variant) {
      case 'compact':
        return visibleActions.filter(a => a.priority === 'high').slice(0, 4)
      case 'grid':
        return visibleActions.slice(0, 6)
      case 'full':
      default:
        return isExpanded ? visibleActions : visibleActions.slice(0, 6)
    }
  }

  const displayActions = getDisplayActions()

  if (!isMobile) {
    return null // 只在移动端显示
  }

  return (
    <Card className={cn('border-0 shadow-lg', className)}>
      {showTitle && (
        <div className="px-4 pt-4 pb-2">
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-gray-900">
              {isCourier ? '信使工具' : '快速操作'}
            </h3>
            {variant === 'full' && visibleActions.length > 6 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsExpanded(!isExpanded)}
                className="text-xs"
              >
                {isExpanded ? (
                  <>
                    <X className="w-3 h-3 mr-1" />
                    收起
                  </>
                ) : (
                  <>
                    <Menu className="w-3 h-3 mr-1" />
                    更多
                  </>
                )}
              </Button>
            )}
          </div>
        </div>
      )}
      
      <CardContent className="p-4 pt-2">
        <div className={cn(
          'grid gap-3',
          variant === 'compact' ? 'grid-cols-2' : 'grid-cols-2 sm:grid-cols-3'
        )}>
          {displayActions.map((action) => {
            const Icon = action.icon
            
            return (
              <Link
                key={action.id}
                href={action.href}
                className="block"
              >
                <div className={cn(
                  'relative p-4 rounded-xl transition-all duration-200 active:scale-95',
                  action.color,
                  'flex flex-col items-center text-center space-y-2'
                )}>
                  <div className="relative">
                    <Icon className="h-6 w-6" />
                    {action.badge && (
                      <Badge 
                        variant="destructive" 
                        className="absolute -top-2 -right-2 h-5 w-5 p-0 flex items-center justify-center text-xs"
                      >
                        {typeof action.badge === 'number' && action.badge > 9 
                          ? '9+' 
                          : action.badge
                        }
                      </Badge>
                    )}
                  </div>
                  
                  <div className="space-y-0.5">
                    <p className="font-medium text-sm leading-tight">
                      {action.title}
                    </p>
                    <p className="text-xs opacity-90 leading-tight">
                      {action.description}
                    </p>
                  </div>

                  {/* 优先级指示器 */}
                  {action.priority === 'high' && (
                    <div className="absolute top-1 right-1">
                      <div className="w-2 h-2 bg-yellow-300 rounded-full animate-pulse" />
                    </div>
                  )}
                </div>
              </Link>
            )
          })}
        </div>

        {/* 快速状态指示 */}
        {isCourier && (
          <div className="mt-4 pt-3 border-t border-gray-200">
            <div className="flex items-center justify-between text-xs text-gray-600">
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 bg-green-500 rounded-full" />
                <span>在线</span>
              </div>
              <div className="flex items-center gap-1">
                <Clock className="w-3 h-3" />
                <span>今日: 5 任务</span>
              </div>
              <div className="flex items-center gap-1">
                <Star className="w-3 h-3 text-yellow-500" />
                <span>4.8 评分</span>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

// 紧急操作浮动按钮
export function EmergencyFAB() {
  const [isVisible, setIsVisible] = useState(true)
  const [lastScrollY, setLastScrollY] = useState(0)
  const { hasPermission } = usePermission()

  const isCourier = hasPermission('courier.basic') || 
                   hasPermission('courier.intermediate') || 
                   hasPermission('courier.advanced') || 
                   hasPermission('courier.management')

  // 滚动隐藏/显示
  useEffect(() => {
    const handleScroll = () => {
      const currentScrollY = window.scrollY
      
      if (currentScrollY > lastScrollY && currentScrollY > 100) {
        setIsVisible(false)
      } else {
        setIsVisible(true)
      }
      
      setLastScrollY(currentScrollY)
    }

    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [lastScrollY])

  // 检查移动端
  const [isMobile, setIsMobile] = useState(false)

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768)
    }
    
    checkMobile()
    window.addEventListener('resize', checkMobile)
    
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  if (!isMobile || !isCourier) {
    return null
  }

  return (
    <div 
      className={cn(
        'fixed left-4 bottom-20 z-40 transition-all duration-300 md:hidden',
        isVisible ? 'translate-y-0 opacity-100' : 'translate-y-2 opacity-0'
      )}
    >
      <Button
        asChild
        className="w-12 h-12 rounded-full bg-red-600 hover:bg-red-700 text-white shadow-lg border-0"
      >
        <a href="tel:400-000-0000" aria-label="紧急联系">
          <Phone className="h-5 w-5" />
        </a>
      </Button>
    </div>
  )
}

// 快速扫码按钮（仅信使）
export function QuickScanButton() {
  const { hasPermission } = usePermission()
  const [isMobile, setIsMobile] = useState(false)

  const isCourier = hasPermission('courier.basic') || 
                   hasPermission('courier.intermediate') || 
                   hasPermission('courier.advanced') || 
                   hasPermission('courier.management')

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768)
    }
    
    checkMobile()
    window.addEventListener('resize', checkMobile)
    
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  if (!isMobile || !isCourier) {
    return null
  }

  return (
    <div className="fixed bottom-4 left-1/2 transform -translate-x-1/2 z-50 md:hidden">
      <Button
        asChild
        className="w-16 h-16 rounded-full bg-gradient-to-r from-green-500 to-blue-500 hover:from-green-600 hover:to-blue-600 text-white shadow-xl border-0"
      >
        <Link href="/courier/scan" aria-label="快速扫码">
          <QrCode className="h-7 w-7" />
        </Link>
      </Button>
    </div>
  )
}