'use client'

import { ReactNode } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useCourierPermission } from '@/hooks/use-courier-permission'
import { Shield, ArrowLeft } from 'lucide-react'

interface CourierPermissionGuardProps {
  requiredLevel?: 1 | 2 | 3 | 4
  config?: PermissionConfig
  children: ReactNode
  fallbackContent?: ReactNode
  errorTitle?: string
  errorDescription?: string
}

interface PermissionConfig {
  requiredLevel: 1 | 2 | 3 | 4
  errorTitle: string
  errorDescription: string
}

// Permission configuration constants
export const COURIER_PERMISSION_CONFIGS = {
  FIRST_LEVEL_MANAGEMENT: {
    requiredLevel: 2 as const,
    errorTitle: '一级信使管理权限不足',
    errorDescription: '只有二级及以上信使才能管理一级信使'
  },
  SECOND_LEVEL_MANAGEMENT: {
    requiredLevel: 3 as const,
    errorTitle: '二级信使管理权限不足',
    errorDescription: '只有三级及以上信使才能管理二级信使'
  },
  THIRD_LEVEL_MANAGEMENT: {
    requiredLevel: 4 as const,
    errorTitle: '三级信使管理权限不足',
    errorDescription: '只有四级信使才能管理三级信使'
  },
  TASK_ASSIGNMENT: {
    requiredLevel: 2 as const,
    errorTitle: '任务分配权限不足',
    errorDescription: '只有二级及以上信使才能分配任务'
  }
}

export function CourierPermissionGuard({
  requiredLevel,
  config,
  children,
  fallbackContent,
  errorTitle,
  errorDescription
}: CourierPermissionGuardProps) {
  const { courierInfo, isCourierLevel, loading } = useCourierPermission()

  // Use config or individual parameters
  const actualRequiredLevel = config?.requiredLevel || requiredLevel
  const actualErrorTitle = config?.errorTitle || errorTitle || '访问权限不足'
  const actualErrorDescription = config?.errorDescription || errorDescription || '您没有足够的权限访问此页面'

  // 加载中状态
  if (loading) {
    return (
      <div className="min-h-screen bg-amber-50 flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Shield className="w-12 h-12 text-amber-600 mx-auto mb-4 animate-pulse" />
            <h2 className="text-xl font-semibold text-amber-900 mb-2">正在验证权限...</h2>
            <p className="text-amber-700">请稍候</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  // 权限检查失败
  if (!actualRequiredLevel || !courierInfo || !isCourierLevel(actualRequiredLevel)) {
    if (fallbackContent) {
      return <>{fallbackContent}</>
    }

    const levelNames = {
      1: '一级信使（楼栋负责人）',
      2: '二级信使（片区管理员）',
      3: '三级信使（校级管理员）',
      4: '四级信使（城市总代）'
    }

    return (
      <div className="min-h-screen bg-amber-50 flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Shield className="w-12 h-12 text-amber-600 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-amber-900 mb-2">{actualErrorTitle}</h2>
            <p className="text-amber-700 mb-4">
              {actualErrorDescription}
            </p>
            <div className="text-sm text-amber-600 mb-6">
              当前权限: {courierInfo ? `${courierInfo.level}级信使` : '未获取到信使信息'}
              {actualRequiredLevel && (
                <span className="block mt-1">需要: {levelNames[actualRequiredLevel]}</span>
              )}
            </div>
            <div className="flex flex-col sm:flex-row gap-2 justify-center">
              <Button 
                asChild 
                variant="outline" 
                className="border-amber-300 text-amber-700 touch-manipulation active:scale-95"
              >
                <a href="/courier">
                  <ArrowLeft className="w-4 h-4 mr-2" />
                  返回信使中心
                </a>
              </Button>
              <Button 
                asChild 
                className="bg-amber-600 hover:bg-amber-700 text-white touch-manipulation active:scale-95"
              >
                <a href="/courier/apply">申请权限升级</a>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  // 权限验证通过，渲染子组件
  return <>{children}</>
}

// 带有权限检查的高阶组件
export function withCourierPermission<P extends object>(
  Component: React.ComponentType<P>,
  requiredLevel: 1 | 2 | 3 | 4
) {
  return function WrappedComponent(props: P) {
    return (
      <CourierPermissionGuard requiredLevel={requiredLevel}>
        <Component {...props} />
      </CourierPermissionGuard>
    )
  }
}