'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { WelcomeBanner } from '@/components/ui/welcome-banner'
import { 
  Send, 
  MapPin, 
  Star, 
  Users, 
  Clock,
  QrCode,
  Trophy,
  Heart,
  ArrowRight,
  Plus,
  Mail,
  CheckCircle,
  Settings,
  Crown,
  TrendingUp,
  UserCheck,
  Package,
  Award
} from 'lucide-react'
import { useCourierPermission } from '@/hooks/use-courier-permission'
import { useAuth } from '@/contexts/auth-context-new'
import { ManagementFloatingButton } from '@/components/courier/ManagementFloatingButton'

interface CourierStatus {
  isApplied: boolean
  status: 'pending' | 'approved' | 'rejected' | null
  level: number
  taskCount: number
  points: number
  zone: string
}

export default function CourierPage() {
  const { user } = useAuth()
  const { 
    courierInfo, 
    loading,
    showManagementDashboard, 
    getManagementDashboardPath, 
    getCourierLevelName 
  } = useCourierPermission()
  
  const [courierStatus, setCourierStatus] = useState<CourierStatus>({
    isApplied: false,
    status: null,
    level: 0,
    taskCount: 0,
    points: 0,
    zone: ''
  })

  // 根据用户角色设置信使状态
  useEffect(() => {
    if (user && !loading) {
      // 优先使用登录时返回的courierInfo
      const userCourierInfo = user.courierInfo
      
      // 如果是super_admin或者已经有信使信息，直接设置为已申请且通过
      if (user.role === 'super_admin' || courierInfo || userCourierInfo) {
        const finalCourierInfo = courierInfo || userCourierInfo
        setCourierStatus(prev => {
          // 只有当状态真正改变时才更新
          if (prev.status !== 'approved' || prev.level !== (finalCourierInfo?.level || 4)) {
            return {
              isApplied: true,
              status: 'approved',
              level: finalCourierInfo?.level || 4, // admin默认为最高级
              taskCount: finalCourierInfo?.taskCount || 0,
              points: finalCourierInfo?.points || 0,
              zone: finalCourierInfo?.zoneCode || '全城'
            }
          }
          return prev
        })
      } else if (user.role?.includes('courier')) {
        // 如果是信使角色但没有courierInfo，设置为待审核
        setCourierStatus(prev => {
          // 只有当状态真正改变时才更新
          if (prev.status !== 'pending') {
            return {
              isApplied: true,
              status: 'pending',
              level: 0,
              taskCount: 0,
              points: 0,
              zone: ''
            }
          }
          return prev
        })
      }
    }
  }, [user, courierInfo, loading])

  const features = [
    {
      icon: Send,
      title: '传递温暖',
      description: '成为连接心灵的桥梁，将每一封信安全送达',
      color: 'amber'
    },
    {
      icon: Star,
      title: '获得认可',
      description: '积累积分和等级，获得平台认证和奖励',
      color: 'orange'
    },
    {
      icon: Users,
      title: '社区归属',
      description: '加入温暖的信使大家庭，结识志同道合的朋友',
      color: 'yellow'
    },
    {
      icon: Heart,
      title: '帮助他人',
      description: '每一次投递都是在传递爱与关怀',
      color: 'red'
    }
  ]

  const quickActions = [
    {
      title: '扫码投递',
      description: '扫描信件二维码开始投递',
      href: '/courier/scan',
      icon: QrCode,
      color: 'bg-amber-600 hover:bg-amber-700'
    },
    {
      title: '任务中心',
      description: '查看待投递任务和历史记录',
      href: '/courier/tasks',
      icon: Mail,
      color: 'bg-orange-600 hover:bg-orange-700'
    },
    {
      title: '我的积分',
      description: '查看积分排行和奖励兑换',
      href: '/courier/points',
      icon: Trophy,
      color: 'bg-yellow-600 hover:bg-yellow-700'
    },
    {
      title: '晋升之路',
      description: '查看晋升要求和申请晋级',
      href: '/courier/growth',
      icon: TrendingUp,
      color: 'bg-green-600 hover:bg-green-700'
    }
  ]

  // 根据信使级别添加管理后台入口
  if (showManagementDashboard()) {
    quickActions.push({
      title: '管理后台',
      description: `${getCourierLevelName()}专属管理中心`,
      href: getManagementDashboardPath(),
      icon: courierInfo?.level === 4 ? Crown : Settings,
      color: courierInfo?.level === 4 ? 'bg-purple-600 hover:bg-purple-700' : 'bg-green-600 hover:bg-green-700'
    })
    
    // 3级及以上信使可以管理晋升申请
    if (courierInfo && courierInfo.level >= 3) {
      quickActions.push({
        title: '晋升管理',
        description: '审核下级信使晋升申请',
        href: '/courier/growth/manage',
        icon: UserCheck,
        color: 'bg-purple-600 hover:bg-purple-700'
      })
      
      // L3/L4信使专属：批量条码管理
      quickActions.push({
        title: '批量管理',
        description: '批量生成和管理条码系统',
        href: '/courier/batch',
        icon: Package,
        color: 'bg-amber-600 hover:bg-amber-700'
      })
    }
    
    // L2+ 信使可以访问积分管理
    if (courierInfo && courierInfo.level >= 2) {
      quickActions.push({
        title: '积分管理',
        description: '管理团队积分和奖励',
        href: '/courier/credit-manage',
        icon: Award,
        color: 'bg-purple-600 hover:bg-purple-700'
      })
    }
  }

  const CourierDashboard = () => (
    <div className="space-y-6">
      {/* 信使状态卡片 */}
      <Card className="border-amber-200 bg-gradient-to-r from-amber-50 to-orange-50">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Badge variant="secondary" className="bg-amber-600 text-white">
                  {user?.role === 'super_admin' ? '系统管理员' : `${courierInfo?.level || user?.courierInfo?.level || courierStatus.level}级信使`}
                </Badge>
                <span className="text-amber-900">信使控制台</span>
              </CardTitle>
              <CardDescription className="text-amber-700">
                覆盖区域: {courierInfo?.zoneCode || courierStatus.zone || '未设置'}
              </CardDescription>
            </div>
            <div className="text-right">
              <div className="text-2xl font-bold text-amber-900">{courierInfo?.points || courierStatus.points}</div>
              <div className="text-sm text-amber-600">积分</div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-4 text-center">
            <div>
              <div className="text-lg font-semibold text-amber-900">{courierInfo?.taskCount || courierStatus.taskCount}</div>
              <div className="text-sm text-amber-600">已投递</div>
            </div>
            <div>
              <div className="text-lg font-semibold text-amber-900">3</div>
              <div className="text-sm text-amber-600">待投递</div>
            </div>
            <div>
              <div className="text-lg font-semibold text-amber-900">98%</div>
              <div className="text-sm text-amber-600">完成率</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 管理权限提示卡片 */}
      {showManagementDashboard() && (
        <Card className="border-purple-200 bg-gradient-to-r from-purple-50 to-amber-50">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className={`w-12 h-12 rounded-full flex items-center justify-center ${
                  courierInfo?.level === 4 ? 'bg-purple-600' : 'bg-green-600'
                } text-white`}>
                  {courierInfo?.level === 4 ? (
                    <Crown className="w-6 h-6" />
                  ) : (
                    <Settings className="w-6 h-6" />
                  )}
                </div>
                <div>
                  <h3 className="font-semibold text-purple-900">
                    {getCourierLevelName()}管理权限
                  </h3>
                  <p className="text-sm text-purple-700">
                    您有权限管理下级信使和查看管理数据
                  </p>
                </div>
              </div>
              <Link href={getManagementDashboardPath()}>
                <Button className={`${
                  courierInfo?.level === 4 ? 'bg-purple-600 hover:bg-purple-700' : 'bg-green-600 hover:bg-green-700'
                } text-white`}>
                  进入管理后台
                  <ArrowRight className="w-4 h-4 ml-2" />
                </Button>
              </Link>
            </div>
          </CardContent>
        </Card>
      )}
      

      {/* 快速操作 */}
      <div className={`grid grid-cols-1 gap-4 ${quickActions.length <= 3 ? 'md:grid-cols-3' : 'md:grid-cols-2 lg:grid-cols-4'}`}>
        {quickActions.map((action) => {
          const Icon = action.icon
          return (
            <Link key={action.href} href={action.href}>
              <Card className="border-amber-200 hover:border-amber-400 hover:shadow-lg transition-all duration-200 cursor-pointer">
                <CardContent className="p-4 text-center">
                  <div className={`w-12 h-12 rounded-full ${action.color} flex items-center justify-center mx-auto mb-3`}>
                    <Icon className="w-6 h-6 text-white" />
                  </div>
                  <h3 className="font-semibold text-amber-900 mb-1">{action.title}</h3>
                  <p className="text-sm text-amber-700">{action.description}</p>
                </CardContent>
              </Card>
            </Link>
          )
        })}
      </div>
    </div>
  )

  const WelcomeSection = () => (
    <div className="space-y-8">
      {/* 欢迎标题 */}
      <div className="text-center">
        <div className="w-20 h-20 bg-amber-200 rounded-full flex items-center justify-center mx-auto mb-6">
          <Send className="w-10 h-10 text-amber-700" />
        </div>
        <h1 className="font-serif text-4xl font-bold text-amber-900 mb-4">
          成为 OpenPenPal 信使
        </h1>
        <p className="text-xl text-amber-700 max-w-2xl mx-auto leading-relaxed">
          你的每一次投递，都是在传递情感的旅程。
          加入我们，成为连接心灵的桥梁。
        </p>
      </div>

      {/* 信使特权 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {features.map((feature, index) => {
          const Icon = feature.icon
          return (
            <Card key={feature.title} className="border-amber-200 hover:border-amber-400 hover:shadow-lg transition-all duration-300">
              <CardContent className="p-6 text-center">
                <div className={`w-16 h-16 bg-${feature.color}-100 rounded-2xl flex items-center justify-center mx-auto mb-4`}>
                  <Icon className={`w-8 h-8 text-${feature.color}-600`} />
                </div>
                <h3 className="font-serif text-lg font-semibold text-amber-900 mb-2">
                  {feature.title}
                </h3>
                <p className="text-amber-700 text-sm leading-relaxed">
                  {feature.description}
                </p>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* 统计数据 */}
      <Card className="border-amber-200 bg-gradient-to-r from-amber-50 to-orange-50">
        <CardContent className="p-8">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 text-center">
            <div>
              <div className="text-3xl font-bold text-amber-900 mb-2">156</div>
              <div className="text-amber-700">活跃信使</div>
            </div>
            <div>
              <div className="text-3xl font-bold text-amber-900 mb-2">12</div>
              <div className="text-amber-700">覆盖学校</div>
            </div>
            <div>
              <div className="text-3xl font-bold text-amber-900 mb-2">2,847</div>
              <div className="text-amber-700">投递信件</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 申请流程 */}
      <Card className="border-amber-200">
        <CardHeader>
          <CardTitle className="text-center text-amber-900">申请流程</CardTitle>
          <CardDescription className="text-center">
            简单三步，开启信使之旅
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="text-center">
              <div className="w-12 h-12 bg-amber-600 text-white rounded-full flex items-center justify-center mx-auto mb-3 font-bold">
                1
              </div>
              <h3 className="font-semibold text-amber-900 mb-2">提交申请</h3>
              <p className="text-sm text-amber-700">填写基本信息和服务区域</p>
            </div>
            <div className="text-center">
              <div className="w-12 h-12 bg-amber-600 text-white rounded-full flex items-center justify-center mx-auto mb-3 font-bold">
                2
              </div>
              <h3 className="font-semibold text-amber-900 mb-2">审核通过</h3>
              <p className="text-sm text-amber-700">24小时内完成审核</p>
            </div>
            <div className="text-center">
              <div className="w-12 h-12 bg-amber-600 text-white rounded-full flex items-center justify-center mx-auto mb-3 font-bold">
                3
              </div>
              <h3 className="font-semibold text-amber-900 mb-2">开始投递</h3>
              <p className="text-sm text-amber-700">完成新手任务正式上线</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* CTA按钮 */}
      <div className="text-center">
        <Button asChild size="lg" className="bg-amber-600 hover:bg-amber-700 text-white font-serif text-xl px-12 py-6">
          <Link href="/courier/apply">
            <Plus className="mr-3 h-6 w-6" />
            申请成为信使
            <ArrowRight className="ml-3 h-6 w-6" />
          </Link>
        </Button>
        <p className="text-amber-600 text-sm mt-4">
          已有账号？<Link href="/courier/scan" className="underline hover:text-amber-800">直接扫码投递</Link>
        </p>
      </div>
    </div>
  )

  // 如果正在加载courierInfo，显示加载状态
  if (loading) {
    return (
      <div className="min-h-screen bg-amber-50">
        <div className="container max-w-6xl mx-auto px-4 py-8">
          <div className="text-center py-16">
            <div className="w-16 h-16 border-4 border-amber-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
            <h2 className="text-2xl font-bold text-amber-900 mb-2">加载信使信息中...</h2>
            <p className="text-amber-700">正在获取您的信使数据</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-6xl mx-auto px-4 py-8">
        {/* 顶部导航栏 */}
        <div className="flex justify-between items-center mb-6">
          <Link href="/">
            <Button variant="ghost" className="flex items-center gap-2 text-amber-700 hover:text-amber-900 hover:bg-amber-100">
              <ArrowRight className="w-4 h-4 rotate-180" />
              返回首页
            </Button>
          </Link>
          <div className="flex items-center gap-2">
            <span className="text-amber-700">欢迎，{user?.nickname || user?.username}</span>
          </div>
        </div>
        
        <WelcomeBanner />
        
        {courierStatus.isApplied && courierStatus.status === 'approved' ? (
          <CourierDashboard />
        ) : courierStatus.isApplied && courierStatus.status === 'pending' ? (
          // 申请审核中状态
          <div className="text-center py-16">
            <Clock className="w-16 h-16 text-amber-600 mx-auto mb-4" />
            <h2 className="text-2xl font-bold text-amber-900 mb-2">申请审核中</h2>
            <p className="text-amber-700 mb-6">
              我们正在审核您的申请，通常在24小时内完成。
              审核通过后会通过微信/短信通知您。
            </p>
            <Button asChild variant="outline" className="border-amber-300 text-amber-700 hover:bg-amber-50">
              <Link href="/profile">
                返回个人中心
              </Link>
            </Button>
          </div>
        ) : (
          <WelcomeSection />
        )}
      </div>
      
      {/* 管理浮动按钮 - 仅在已通过审核时显示 */}
      {courierStatus.isApplied && courierStatus.status === 'approved' && (
        <ManagementFloatingButton />
      )}
    </div>
  )
}