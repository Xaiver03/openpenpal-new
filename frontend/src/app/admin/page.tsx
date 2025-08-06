'use client'

import { usePermission } from '@/hooks/use-permission'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Users, Mail, BarChart, Settings, Shield, Brain, Truck } from 'lucide-react'
import Link from 'next/link'
import { WelcomeBanner } from '@/components/ui/welcome-banner'

export default function AdminDashboard() {
  const { user, getRoleDisplayName } = usePermission()

  const adminCards = [
    {
      title: '用户管理',
      description: '管理平台用户和权限',
      icon: Users,
      href: '/admin/users',
      color: 'bg-blue-600'
    },
    {
      title: '信件管理',
      description: '查看和管理所有信件',
      icon: Mail,
      href: '/admin/letters',
      color: 'bg-green-600'
    },
    {
      title: '信使管理',
      description: '管理四级信使层级系统',
      icon: Truck,
      href: '/admin/couriers',
      color: 'bg-emerald-600'
    },
    {
      title: '内容审核',
      description: '管理内容审核和敏感词库',
      icon: Shield,
      href: '/admin/moderation',
      color: 'bg-red-600'
    },
    {
      title: '数据分析',
      description: '查看平台运营数据',
      icon: BarChart,
      href: '/admin/analytics',
      color: 'bg-purple-600'
    },
    {
      title: '系统设置',
      description: '配置系统参数',
      icon: Settings,
      href: '/admin/settings',
      color: 'bg-orange-600'
    },
    {
      title: 'AI管理',
      description: 'AI功能配置和监控',
      icon: Brain,
      href: '/admin/ai',
      color: 'bg-indigo-600'
    }
  ]

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <WelcomeBanner />
        
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">管理控制台</h1>
          <p className="text-gray-600 mt-2">
            欢迎，{user?.nickname} ({getRoleDisplayName()})
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-6">
          {adminCards.map((card) => {
            const Icon = card.icon
            return (
              <Link key={card.href} href={card.href}>
                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <div className={`w-12 h-12 rounded-lg ${card.color} flex items-center justify-center mb-4`}>
                      <Icon className="w-6 h-6 text-white" />
                    </div>
                    <CardTitle>{card.title}</CardTitle>
                    <CardDescription>{card.description}</CardDescription>
                  </CardHeader>
                </Card>
              </Link>
            )
          })}
        </div>

        {/* 快速统计 */}
        <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-6">
          <Card>
            <CardContent className="p-6">
              <div className="text-2xl font-bold">1,234</div>
              <p className="text-gray-600">注册用户</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="text-2xl font-bold">5,678</div>
              <p className="text-gray-600">投递信件</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="text-2xl font-bold">156</div>
              <p className="text-gray-600">活跃信使</p>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}