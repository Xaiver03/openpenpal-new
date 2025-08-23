'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { 
  Home,
  Truck,
  QrCode,
  Navigation,
  User,
  Bell,
  Search,
  MapPin,
  Route,
  Building,
  BookOpen,
  Package,
  MessageCircle,
  Settings,
  Waves,
  Calendar
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuth } from '@/contexts/auth-context-new'
import { usePermission } from '@/hooks/use-permission'
import { Badge } from '@/components/ui/badge'

interface NavTab {
  id: string
  label: string
  icon: any
  href: string
  badge?: number
  requiresCourier?: boolean
  activePattern?: RegExp
}

export function MobileNavTabs() {
  const pathname = usePathname()
  const { user } = useAuth()
  const { hasPermission } = usePermission()
  const [notificationCount, setNotificationCount] = useState(0)

  // 检查是否是信使
  const isCourier = hasPermission('courier.basic') || 
                   hasPermission('courier.intermediate') || 
                   hasPermission('courier.advanced') || 
                   hasPermission('courier.management')

  // 检查是否在移动设备上
  const [isMobile, setIsMobile] = useState(false)

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768)
    }
    
    checkMobile()
    window.addEventListener('resize', checkMobile)
    
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  // 模拟通知数量（实际应该从状态管理或API获取）
  useEffect(() => {
    // 这里应该订阅通知状态
    setNotificationCount(3) // 模拟数据
  }, [])

  const navTabs: NavTab[] = [
    {
      id: 'home',
      label: '首页',
      icon: Home,
      href: '/',
      activePattern: /^\/$/
    },
    {
      id: 'scan',
      label: '扫码',
      icon: QrCode,
      href: '/courier/scan',
      requiresCourier: true,
      activePattern: /^\/courier\/(scan|tasks)/
    },
    {
      id: 'delivery',
      label: '投递',
      icon: Navigation,
      href: '/delivery-guide',
      activePattern: /^\/delivery-guide/
    },
    {
      id: 'courier',
      label: '信使',
      icon: Truck,
      href: '/courier',
      requiresCourier: true,
      activePattern: /^\/courier(?!\/scan)/
    },
    {
      id: 'profile',
      label: '我的',
      icon: User,
      href: user?.username ? `/u/${user.username}` : '/settings',
      badge: notificationCount,
      activePattern: /^\/(u\/|settings|profile)/
    }
  ]

  // 如果用户不是信使，显示不同的标签组合
  const userNavTabs: NavTab[] = [
    {
      id: 'home',
      label: '首页',
      icon: Home,
      href: '/',
      activePattern: /^\/$/
    },
    {
      id: 'drift-bottle',
      label: '漂流瓶',
      icon: Waves,
      href: '/drift-bottle',
      activePattern: /^\/drift-bottle/
    },
    {
      id: 'future-letter',
      label: '未来信',
      icon: Calendar,
      href: '/future-letter',
      activePattern: /^\/future-letter/
    },
    {
      id: 'write',
      label: '写信',
      icon: MessageCircle,
      href: '/letters/write',
      activePattern: /^\/letters\/write/
    },
    {
      id: 'profile',
      label: '我的',
      icon: User,
      href: user?.username ? `/u/${user.username}` : '/settings',
      badge: notificationCount,
      activePattern: /^\/(u\/|settings|profile)/
    }
  ]

  const currentTabs = isCourier ? navTabs : userNavTabs
  const visibleTabs = currentTabs.filter(tab => !tab.requiresCourier || isCourier)

  const isActive = (tab: NavTab) => {
    if (tab.activePattern) {
      return tab.activePattern.test(pathname)
    }
    return pathname === tab.href || pathname.startsWith(tab.href + '/')
  }

  // 只在移动端和有用户登录时显示
  if (!isMobile || !user) {
    return null
  }

  return (
    <>
      {/* 底部导航栏 */}
      <div className="fixed bottom-0 left-0 right-0 z-50 md:hidden">
        <div className="bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 border-t">
          <nav className="flex items-center justify-around px-2 py-1">
            {visibleTabs.map((tab) => {
              const Icon = tab.icon
              const active = isActive(tab)
              
              return (
                <Link
                  key={tab.id}
                  href={tab.href}
                  className={cn(
                    'flex flex-col items-center justify-center px-3 py-2 min-w-0 relative transition-colors',
                    active
                      ? 'text-primary'
                      : 'text-muted-foreground hover:text-foreground'
                  )}
                >
                  <div className="relative">
                    <Icon 
                      className={cn(
                        'h-5 w-5 mb-1',
                        active && 'scale-110 transition-transform'
                      )} 
                    />
                    {tab.badge && tab.badge > 0 && (
                      <Badge 
                        variant="destructive" 
                        className="absolute -top-2 -right-2 h-4 w-4 p-0 flex items-center justify-center text-xs"
                      >
                        {tab.badge > 9 ? '9+' : tab.badge}
                      </Badge>
                    )}
                  </div>
                  <span 
                    className={cn(
                      'text-xs truncate max-w-full',
                      active && 'font-medium'
                    )}
                  >
                    {tab.label}
                  </span>
                  {active && (
                    <div className="absolute top-0 left-1/2 transform -translate-x-1/2 w-6 h-0.5 bg-primary rounded-b-full" />
                  )}
                </Link>
              )
            })}
          </nav>
        </div>
      </div>

      {/* 底部占位空间，防止内容被导航栏遮挡 */}
      <div className="h-16 md:hidden" />
    </>
  )
}

// 移动端快速操作浮动按钮
export function MobileFAB() {
  const pathname = usePathname()
  const { hasPermission } = usePermission()
  const [isVisible, setIsVisible] = useState(true)
  const [lastScrollY, setLastScrollY] = useState(0)

  // 检查是否是信使
  const isCourier = hasPermission('courier.basic') || 
                   hasPermission('courier.intermediate') || 
                   hasPermission('courier.advanced') || 
                   hasPermission('courier.management')

  // 滚动时隐藏/显示FAB
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

  // 根据当前页面决定FAB功能
  const getFABConfig = () => {
    if (pathname.includes('/courier') && isCourier) {
      return {
        icon: QrCode,
        href: '/courier/scan',
        label: '扫码',
        color: 'bg-green-600 hover:bg-green-700'
      }
    }
    
    if (pathname.includes('/delivery-guide')) {
      return {
        icon: Navigation,
        href: '/delivery-guide/route-planner',
        label: '路线规划',
        color: 'bg-blue-600 hover:bg-blue-700'
      }
    }
    
    if (pathname === '/') {
      return {
        icon: MessageCircle,
        href: '/letters/write',
        label: '写信',
        color: 'bg-primary hover:bg-primary/90'
      }
    }
    
    return null
  }

  const fabConfig = getFABConfig()

  // 检查是否在移动设备上
  const [isMobile, setIsMobile] = useState(false)

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768)
    }
    
    checkMobile()
    window.addEventListener('resize', checkMobile)
    
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  if (!isMobile || !fabConfig) {
    return null
  }

  const Icon = fabConfig.icon

  return (
    <div 
      className={cn(
        'fixed right-4 bottom-20 z-40 transition-all duration-300 md:hidden',
        isVisible ? 'translate-y-0 opacity-100' : 'translate-y-2 opacity-0'
      )}
    >
      <Link
        href={fabConfig.href}
        className={cn(
          'flex items-center justify-center w-14 h-14 rounded-full shadow-lg text-white transition-all duration-200 active:scale-95',
          fabConfig.color
        )}
        aria-label={fabConfig.label}
      >
        <Icon className="h-6 w-6" />
      </Link>
    </div>
  )
}

// 移动端页面容器，自动调整底部边距
export function MobilePageContainer({ 
  children, 
  className 
}: { 
  children: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn('pb-16 md:pb-0', className)}>
      {children}
    </div>
  )
}