'use client'

import { useState, useEffect } from 'react'
import { usePermission } from '@/hooks/use-permission'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Users, Mail, BarChart, Settings, Shield, Brain, Truck, Loader2, RefreshCw, AlertCircle } from 'lucide-react'
import Link from 'next/link'
import { WelcomeBanner } from '@/components/ui/welcome-banner'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import AdminService from '@/lib/services/admin-service'
import { formatRelativeTime } from '@/lib/utils'

interface DashboardStats {
  users: {
    total: number
    active: number
    new_today: number
  }
  letters: {
    total: number
    sent_today: number
    in_transit: number
  }
  couriers: {
    total: number
    active: number
    performance: {
      success_rate: number
    }
  }
}

export default function EnhancedAdminDashboard() {
  const { user, getRoleDisplayName } = usePermission()
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())

  const adminCards = [
    {
      title: '用户管理',
      description: '管理平台用户和权限',
      icon: Users,
      href: '/admin/users',
      color: 'bg-blue-600',
      stat: stats?.users.total || 0,
      label: '注册用户'
    },
    {
      title: '信件管理',
      description: '查看和管理所有信件',
      icon: Mail,
      href: '/admin/letters',
      color: 'bg-green-600',
      stat: stats?.letters.total || 0,
      label: '总信件数'
    },
    {
      title: '信使管理',
      description: '管理四级信使层级系统',
      icon: Truck,
      href: '/admin/couriers',
      color: 'bg-emerald-600',
      stat: stats?.couriers.total || 0,
      label: '信使总数'
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

  useEffect(() => {
    loadDashboardStats()
  }, [])

  const loadDashboardStats = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await AdminService.getDashboardStats()
      
      if (response.success && response.data) {
        const systemStats = response.data
        // Transform the comprehensive stats to our simplified format
        setStats({
          users: {
            total: systemStats.users.total,
            active: systemStats.users.active,
            new_today: systemStats.users.new_today
          },
          letters: {
            total: systemStats.letters.total,
            sent_today: systemStats.letters.sent_today,
            in_transit: systemStats.letters.by_status?.in_transit || 0
          },
          couriers: {
            total: systemStats.couriers.total,
            active: systemStats.couriers.active,
            performance: {
              success_rate: systemStats.couriers.performance.success_rate
            }
          }
        })
      } else {
        throw new Error(response.message || '获取统计数据失败')
      }
    } catch (err) {
      console.error('Failed to load dashboard stats:', err)
      setError(err instanceof Error ? err.message : '加载数据失败')
      
      // 设置空数据而不是mock数据
      setStats({
        users: { total: 0, active: 0, new_today: 0 },
        letters: { total: 0, sent_today: 0, in_transit: 0 },
        couriers: { total: 0, active: 0, performance: { success_rate: 0 } }
      })
    } finally {
      setLoading(false)
      setLastRefresh(new Date())
    }
  }

  const handleRefresh = () => {
    loadDashboardStats()
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <WelcomeBanner />
        
        <div className="mb-8 flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">管理控制台</h1>
            <p className="text-gray-600 mt-2">
              欢迎，{user?.nickname} ({getRoleDisplayName()})
            </p>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-500">
              最后更新: {formatRelativeTime(lastRefresh)}
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={handleRefresh}
              disabled={loading}
            >
              <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
              刷新
            </Button>
          </div>
        </div>

        {error && (
          <Alert variant="destructive" className="mb-6">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {error}
              <Button
                variant="link"
                size="sm"
                onClick={handleRefresh}
                className="ml-2"
              >
                重试
              </Button>
            </AlertDescription>
          </Alert>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {adminCards.map((card) => {
            const Icon = card.icon
            return (
              <Link key={card.href} href={card.href}>
                <Card className="hover:shadow-lg transition-shadow cursor-pointer h-full">
                  <CardHeader>
                    <div className={`w-12 h-12 rounded-lg ${card.color} flex items-center justify-center mb-4`}>
                      <Icon className="w-6 h-6 text-white" />
                    </div>
                    <CardTitle>{card.title}</CardTitle>
                    <CardDescription>{card.description}</CardDescription>
                    {card.stat !== undefined && (
                      <div className="mt-2 pt-2 border-t">
                        <span className="text-2xl font-bold">{card.stat.toLocaleString()}</span>
                        <span className="text-sm text-gray-500 ml-2">{card.label}</span>
                      </div>
                    )}
                  </CardHeader>
                </Card>
              </Link>
            )
          })}
        </div>

        {/* 快速统计 */}
        <div className="mt-8 grid grid-cols-1 md:grid-cols-4 gap-6">
          <Card>
            <CardContent className="p-6">
              {loading ? (
                <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
              ) : (
                <>
                  <div className="text-2xl font-bold">{stats?.users.total.toLocaleString() || 0}</div>
                  <p className="text-gray-600">注册用户</p>
                  <p className="text-sm text-green-600">+{stats?.users.new_today || 0} 今日新增</p>
                </>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              {loading ? (
                <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
              ) : (
                <>
                  <div className="text-2xl font-bold">{stats?.letters.total.toLocaleString() || 0}</div>
                  <p className="text-gray-600">投递信件</p>
                  <p className="text-sm text-blue-600">{stats?.letters.sent_today || 0} 今日发送</p>
                </>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              {loading ? (
                <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
              ) : (
                <>
                  <div className="text-2xl font-bold">{stats?.couriers.active || 0}</div>
                  <p className="text-gray-600">活跃信使</p>
                  <p className="text-sm text-purple-600">总计 {stats?.couriers.total || 0} 人</p>
                </>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              {loading ? (
                <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
              ) : (
                <>
                  <div className="text-2xl font-bold">{stats?.couriers.performance.success_rate || 0}%</div>
                  <p className="text-gray-600">投递成功率</p>
                  <p className="text-sm text-orange-600">{stats?.letters.in_transit || 0} 运输中</p>
                </>
              )}
            </CardContent>
          </Card>
        </div>

        {/* API连接状态提示 */}
        <div className="mt-8">
          <Card className="bg-blue-50 border-blue-200">
            <CardContent className="p-4">
              <div className="flex items-center gap-2">
                <AlertCircle className="h-5 w-5 text-blue-600" />
                <div>
                  <p className="text-sm font-medium text-blue-900">API连接状态</p>
                  <p className="text-sm text-blue-700">
                    {loading ? '正在连接后端服务...' : 
                     error ? '使用备用数据显示（真实API连接失败）' : 
                     '已成功连接到后端API服务'}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}