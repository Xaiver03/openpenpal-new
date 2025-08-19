'use client'

import React from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  MapPin, 
  Package, 
  BarChart, 
  Settings,
  Users,
  Truck,
  Building,
  Home,
  School,
  FileText,
  ShieldCheck
} from 'lucide-react'
import Link from 'next/link'
import { useUser } from '@/hooks/use-user'

export default function CourierDashboard() {
  const router = useRouter()
  const { user } = useUser()

  // 获取信使级别
  const courierLevel = user?.role?.includes('courier_level') 
    ? parseInt(user.role.split('_')[2]) 
    : 0

  // 根据级别定义可用功能
  const getAvailableFeatures = () => {
    const baseFeatures = [
      {
        title: '任务管理',
        description: '查看和管理投递任务',
        icon: Package,
        href: '/courier/tasks',
        color: 'bg-blue-600'
      },
      {
        title: '数据统计',
        description: '查看投递数据和绩效',
        icon: BarChart,
        href: '/courier/analytics',
        color: 'bg-purple-600'
      }
    ]

    // L2及以上信使可以管理 OP Code
    if (courierLevel >= 2) {
      baseFeatures.push({
        title: 'OP Code 管理',
        description: '管理负责区域的编码',
        icon: MapPin,
        href: '/courier/opcode-manage',
        color: 'bg-green-600'
      })
    }

    // L3及以上信使可以管理下级信使
    if (courierLevel >= 3) {
      baseFeatures.push({
        title: '信使管理',
        description: '管理下级信使团队',
        icon: Users,
        href: '/courier/team',
        color: 'bg-orange-600'
      })
    }

    // L4信使有额外的城市级管理功能
    if (courierLevel === 4) {
      baseFeatures.push({
        title: '城市管理',
        description: '管理城市级配置',
        icon: Building,
        href: '/courier/city-manage',
        color: 'bg-red-600'
      })
    }

    return baseFeatures
  }

  // 获取级别信息
  const getLevelInfo = () => {
    switch (courierLevel) {
      case 1:
        return {
          icon: <Home className="h-6 w-6" />,
          title: '一级信使（楼栋投递员）',
          description: '负责具体楼栋的信件投递',
          color: 'bg-blue-500'
        }
      case 2:
        return {
          icon: <Truck className="h-6 w-6" />,
          title: '二级信使（片区管理员）',
          description: '管理片区投递点和投递任务分配',
          color: 'bg-green-500'
        }
      case 3:
        return {
          icon: <School className="h-6 w-6" />,
          title: '三级信使（学校协调员）',
          description: '管理学校级信使团队和区域编码',
          color: 'bg-purple-500'
        }
      case 4:
        return {
          icon: <Building className="h-6 w-6" />,
          title: '四级信使（城市总监）',
          description: '管理城市级信使网络和学校编码',
          color: 'bg-red-500'
        }
      default:
        return null
    }
  }

  const levelInfo = getLevelInfo()
  const features = getAvailableFeatures()

  if (!user || courierLevel === 0) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <ShieldCheck className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 mb-2">需要信使权限</h2>
            <p className="text-gray-600 mb-4">
              您需要成为信使才能访问此页面
            </p>
            <Button onClick={() => router.push('/')} variant="outline">
              返回首页
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 欢迎横幅 */}
      <div className="mb-8">
        <Card className={`${levelInfo?.color} text-white`}>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                {levelInfo?.icon}
                <div>
                  <CardTitle className="text-2xl">欢迎回来，{user.nickname}</CardTitle>
                  <CardDescription className="text-white/80">
                    {levelInfo?.title} - {levelInfo?.description}
                  </CardDescription>
                </div>
              </div>
              <Badge variant="secondary" className="text-lg px-4 py-2">
                管理范围: {user.managed_op_code_prefix || '未设置'}
              </Badge>
            </div>
          </CardHeader>
        </Card>
      </div>

      {/* 快速统计 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <Card>
          <CardContent className="pt-6">
            <div className="text-2xl font-bold">0</div>
            <p className="text-gray-600">今日投递</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-2xl font-bold">0</div>
            <p className="text-gray-600">待处理任务</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-2xl font-bold">0%</div>
            <p className="text-gray-600">投递成功率</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="text-2xl font-bold">0</div>
            <p className="text-gray-600">团队成员</p>
          </CardContent>
        </Card>
      </div>

      {/* 功能入口 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {features.map((feature) => {
          const Icon = feature.icon
          return (
            <Link key={feature.href} href={feature.href}>
              <Card className="hover:shadow-lg transition-shadow cursor-pointer h-full">
                <CardHeader>
                  <div className={`w-12 h-12 rounded-lg ${feature.color} flex items-center justify-center mb-4`}>
                    <Icon className="w-6 h-6 text-white" />
                  </div>
                  <CardTitle>{feature.title}</CardTitle>
                  <CardDescription>{feature.description}</CardDescription>
                </CardHeader>
              </Card>
            </Link>
          )
        })}
      </div>

      {/* 权限说明 */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            权限说明
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 text-sm text-gray-600">
            {courierLevel === 1 && (
              <>
                <p>• 查看和执行分配给您的投递任务</p>
                <p>• 扫描信件条码更新投递状态</p>
                <p>• 查看投递任务中的完整 OP Code</p>
              </>
            )}
            {courierLevel === 2 && (
              <>
                <p>• 审核和分配投递点编码（后两位）</p>
                <p>• 管理片区内的投递任务分配</p>
                <p>• 查看片区投递数据统计</p>
              </>
            )}
            {courierLevel === 3 && (
              <>
                <p>• 管理学校内的片区和楼栋编码（中间两位）</p>
                <p>• 创建和管理下级信使账号</p>
                <p>• 组织学校级活动和推广</p>
              </>
            )}
            {courierLevel === 4 && (
              <>
                <p>• 管理城市内的学校编码（前两位）</p>
                <p>• 统筹城市级信使网络</p>
                <p>• 制定城市投递策略和标准</p>
              </>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}