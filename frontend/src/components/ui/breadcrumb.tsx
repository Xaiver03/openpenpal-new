'use client'

import { ChevronRight, Home } from 'lucide-react'
import Link from 'next/link'
import { cn } from '@/lib/utils'

export interface BreadcrumbItem {
  label: string
  href?: string
  icon?: React.ComponentType<{ className?: string }>
}

interface BreadcrumbProps {
  items: readonly BreadcrumbItem[]
  className?: string
  separator?: React.ReactNode
  showHome?: boolean
  homeHref?: string
}

export function Breadcrumb({ 
  items, 
  className,
  separator = <ChevronRight className="h-4 w-4 text-gray-400" />,
  showHome = true,
  homeHref = '/'
}: BreadcrumbProps) {
  const allItems = showHome 
    ? [{ label: '首页', href: homeHref, icon: Home }, ...items]
    : items

  return (
    <nav className={cn('flex items-center space-x-1 text-sm text-gray-600 mb-4', className)}>
      {allItems.map((item, index) => {
        const isLast = index === allItems.length - 1
        const Icon = item.icon

        return (
          <div key={index} className="flex items-center">
            {index > 0 && (
              <span className="mx-2">{separator}</span>
            )}
            
            {item.href && !isLast ? (
              <Link 
                href={item.href}
                className="flex items-center hover:text-blue-600 transition-colors"
              >
                {Icon && <Icon className="h-4 w-4 mr-1" />}
                {item.label}
              </Link>
            ) : (
              <span 
                className={cn(
                  'flex items-center',
                  isLast ? 'text-gray-900 font-medium' : 'text-gray-600'
                )}
              >
                {Icon && <Icon className="h-4 w-4 mr-1" />}
                {item.label}
              </span>
            )}
          </div>
        )
      })}
    </nav>
  )
}

// 预定义的面包屑配置，方便复用
export const ADMIN_BREADCRUMBS = {
  dashboard: [{ label: '管理控制台', href: '/admin' }],
  users: [
    { label: '管理控制台', href: '/admin' },
    { label: '用户管理', href: '/admin/users' }
  ],
  letters: [
    { label: '管理控制台', href: '/admin' },
    { label: '信件管理', href: '/admin/letters' }
  ],
  couriers: [
    { label: '管理控制台', href: '/admin' },
    { label: '信使管理', href: '/admin/couriers' }
  ],
  courierTasks: [
    { label: '管理控制台', href: '/admin' },
    { label: '信使管理', href: '/admin/couriers' },
    { label: '任务管理', href: '/admin/couriers/tasks' }
  ],
  analytics: [
    { label: '管理控制台', href: '/admin' },
    { label: '数据分析', href: '/admin/analytics' }
  ],
  settings: [
    { label: '管理控制台', href: '/admin' },
    { label: '系统设置', href: '/admin/settings' }
  ],
  moderation: [
    { label: '管理控制台', href: '/admin' },
    { label: '内容审核', href: '/admin/moderation' }
  ],
  ai: [
    { label: '管理控制台', href: '/admin' },
    { label: 'AI管理', href: '/admin/ai' }
  ],
  appointment: [
    { label: '管理控制台', href: '/admin' },
    { label: '用户任命', href: '/admin/appointment' }
  ]
} as const