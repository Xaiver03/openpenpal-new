'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { 
  Mail, 
  Send, 
  Inbox, 
  User, 
  Menu, 
  X,
  Bell,
  Settings,
  LogOut,
  Shield,
  Truck,
  Crown,
  Brain
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useAuth, usePermissions, useCourier } from '@/stores/user-store'
import { getCourierLevelManagementPath } from '@/constants/roles'
import { useUserBasicInfo, useAuthActions } from '@/hooks/use-optimized-subscriptions'
import { NotificationCenter } from '@/components/realtime/notification-center'
import { SimpleWebSocketIndicator } from '@/components/realtime/websocket-status'
import { CourierTestPanel } from '@/components/debug/courier-test-panel'

interface HeaderProps {
  className?: string
}

export function Header({ className }: HeaderProps) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const router = useRouter()
  
  // Optimized state subscriptions
  const { username, nickname, avatar, isAuthenticated } = useUserBasicInfo()
  const { logout } = useAuthActions()
  const { canAccessAdmin } = usePermissions()
  const { courierInfo, isCourier, levelName } = useCourier()
  
  // Debug logging (可以通过开发者面板查看)
  if (process.env.NODE_ENV === 'development' && isAuthenticated) {
    console.group('🔍 Header State Debug')
    console.log('Authentication:', { isAuthenticated, username })
    console.log('Courier Info:', { isCourier, courierInfo, levelName })
    console.log('Admin Access:', canAccessAdmin())
    console.groupEnd()
  }
  
  // Create user object for backward compatibility
  const user = { username, nickname, avatar }

  const navItems = [
    { href: '/write', label: '写信去', icon: Mail },
    { href: '/ai', label: '云锦传驿', icon: Brain },
    { href: '/plaza', label: '写作广场', icon: Send },
    { href: '/museum', label: '信件博物馆', icon: Inbox },
    { href: '/shop', label: '信封商城', icon: Mail },
  ]

  const userMenuItems = [
    { href: '/profile', label: '个人档案', icon: User },
    { href: '/settings', label: '设置', icon: Settings },
    ...(isCourier ? [{ href: '/courier', label: '信使中心', icon: Truck }] : []),
    ...(courierInfo && courierInfo.level >= 2 ? [{
      href: getCourierLevelManagementPath(courierInfo.level),
      label: courierInfo.level === 4 ? '城市管理' : 
             courierInfo.level === 3 ? '学校管理' : 
             courierInfo.level === 2 ? '片区管理' : '信使管理',
      icon: courierInfo?.level === 4 ? Crown : Settings
    }] : []),
    // Only show admin console for non-courier admin roles
    ...(canAccessAdmin() && !isCourier ? [{ href: '/admin', label: '管理控制台', icon: Shield }] : []),
  ]

  const handleLogout = async () => {
    await logout()
    router.push('/')
  }

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen)
  }

  return (
    <header className={cn(
      'sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60',
      className
    )}>
      <div className="container flex h-16 items-center justify-between px-4">
        {/* Logo */}
        <Link href="/" className="flex items-center space-x-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
            <Mail className="h-5 w-5" />
          </div>
          <span className="font-serif text-xl font-bold text-letter-ink">
            OpenPenPal
          </span>
        </Link>

        {/* Desktop Navigation */}
        <nav className="hidden md:flex items-center space-x-1">
          {navItems.map((item) => {
            const Icon = item.icon
            return (
              <Link
                key={item.href}
                href={item.href}
                className="flex items-center space-x-2 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
              >
                <Icon className="h-4 w-4" />
                <span>{item.label}</span>
              </Link>
            )
          })}
        </nav>

        {/* Desktop User Menu */}
        <div className="hidden md:flex items-center space-x-2">
          {isAuthenticated ? (
            <>
              <SimpleWebSocketIndicator className="mr-2" />
              <NotificationCenter />
              
              <div className="flex items-center space-x-1">
                {userMenuItems.map((item) => {
                  const Icon = item.icon
                  return (
                    <Button
                      key={item.href}
                      variant="ghost"
                      size="sm"
                      onClick={() => router.push(item.href)}
                      className="flex items-center space-x-2"
                    >
                      <Icon className="h-4 w-4" />
                      <span className="hidden xl:inline">{item.label}</span>
                    </Button>
                  )
                })}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleLogout}
                  className="flex items-center space-x-2 text-red-600 hover:text-red-700"
                >
                  <LogOut className="h-4 w-4" />
                  <span className="hidden xl:inline">退出</span>
                </Button>
              </div>
            </>
          ) : (
            <div className="flex items-center space-x-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => router.push('/login')}
              >
                登录
              </Button>
              <Button
                size="sm"
                onClick={() => router.push('/register')}
              >
                注册
              </Button>
            </div>
          )}
        </div>

        {/* Mobile Menu Button */}
        <Button
          variant="ghost"
          size="icon"
          className="md:hidden"
          onClick={toggleMobileMenu}
        >
          {isMobileMenuOpen ? (
            <X className="h-5 w-5" />
          ) : (
            <Menu className="h-5 w-5" />
          )}
        </Button>
      </div>

      {/* Mobile Navigation */}
      {isMobileMenuOpen && (
        <div className="md:hidden border-t bg-background">
          <div className="container px-4 py-4">
            <nav className="flex flex-col space-y-2">
              {navItems.map((item) => {
                const Icon = item.icon
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className="flex items-center space-x-3 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    <Icon className="h-4 w-4" />
                    <span>{item.label}</span>
                  </Link>
                )
              })}
              <div className="border-t pt-2 mt-2">
                {isAuthenticated ? (
                  <>
                    {userMenuItems.map((item) => {
                      const Icon = item.icon
                      return (
                        <Link
                          key={item.href}
                          href={item.href}
                          className="flex items-center space-x-3 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
                          onClick={() => setIsMobileMenuOpen(false)}
                        >
                          <Icon className="h-4 w-4" />
                          <span>{item.label}</span>
                        </Link>
                      )
                    })}
                    <button
                      onClick={() => {
                        handleLogout()
                        setIsMobileMenuOpen(false)
                      }}
                      className="flex items-center space-x-3 rounded-md px-3 py-2 text-sm font-medium text-red-600 hover:bg-red-50 w-full text-left"
                    >
                      <LogOut className="h-4 w-4" />
                      <span>退出登录</span>
                    </button>
                  </>
                ) : (
                  <>
                    <Link
                      href="/login"
                      className="flex items-center space-x-3 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
                      onClick={() => setIsMobileMenuOpen(false)}
                    >
                      <User className="h-4 w-4" />
                      <span>登录</span>
                    </Link>
                    <Link
                      href="/register"
                      className="flex items-center space-x-3 rounded-md px-3 py-2 text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90"
                      onClick={() => setIsMobileMenuOpen(false)}
                    >
                      <User className="h-4 w-4" />
                      <span>注册</span>
                    </Link>
                  </>
                )}
              </div>
            </nav>
          </div>
        </div>
      )}
      
      {/* 开发者调试面板 */}
      <CourierTestPanel />
    </header>
  )
}