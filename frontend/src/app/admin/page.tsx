'use client'

import { usePermission } from '@/hooks/use-permission'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Users, Mail, BarChart, Settings, Shield, Brain, Truck } from 'lucide-react'
import Link from 'next/link'
import { WelcomeBanner } from '@/components/ui/welcome-banner'
import { useEffect, useState } from 'react'
import adminService from '@/lib/services/admin-service'

interface DashboardStats {
  totalUsers: number
  newUsersToday: number
  totalLetters: number
  lettersToday: number
  activeCouriers: number
  museumExhibits: number
  envelopeOrders: number
  totalNotifications: number
}

export default function AdminDashboard() {
  const { user, getRoleDisplayName } = usePermission()
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchStats = async () => {
      try {
        setLoading(true)
        const response = await adminService.getDashboardStats()
        setStats(response.data as unknown as DashboardStats)
        setError(null)
      } catch (err) {
        console.error('Failed to fetch dashboard stats:', err)
        setError('获取统计数据失败')
        // Fallback to initial values if API fails
        setStats({
          totalUsers: 0,
          newUsersToday: 0,
          totalLetters: 0,
          lettersToday: 0,
          activeCouriers: 0,
          museumExhibits: 0,
          envelopeOrders: 0,
          totalNotifications: 0
        })
      } finally {
        setLoading(false)
      }
    }

    fetchStats()
  }, [])

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
              {loading ? (
                <div className="text-2xl font-bold text-gray-400">加载中...</div>
              ) : error ? (
                <div className="text-2xl font-bold text-red-500">--</div>
              ) : (
                <div className="text-2xl font-bold">{stats?.totalUsers.toLocaleString() || 0}</div>
              )}
              <p className="text-gray-600">注册用户</p>
              {stats?.newUsersToday ? (
                <p className="text-sm text-green-600">今日新增: {stats.newUsersToday}</p>
              ) : null}
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              {loading ? (
                <div className="text-2xl font-bold text-gray-400">加载中...</div>
              ) : error ? (
                <div className="text-2xl font-bold text-red-500">--</div>
              ) : (
                <div className="text-2xl font-bold">{stats?.totalLetters.toLocaleString() || 0}</div>
              )}
              <p className="text-gray-600">投递信件</p>
              {stats?.lettersToday ? (
                <p className="text-sm text-green-600">今日投递: {stats.lettersToday}</p>
              ) : null}
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              {loading ? (
                <div className="text-2xl font-bold text-gray-400">加载中...</div>
              ) : error ? (
                <div className="text-2xl font-bold text-red-500">--</div>
              ) : (
                <div className="text-2xl font-bold">{stats?.activeCouriers.toLocaleString() || 0}</div>
              )}
              <p className="text-gray-600">活跃信使</p>
              {stats?.museumExhibits ? (
                <p className="text-sm text-blue-600">博物馆展品: {stats.museumExhibits}</p>
              ) : null}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}