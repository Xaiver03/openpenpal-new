'use client'

import { Button } from '@/components/ui/button'
import Link from 'next/link'
import { Shield, Crown, MapPin } from 'lucide-react'
import { useAuth, usePermissions, useCourier } from '@/stores/user-store'
import { getCourierLevelManagementPath } from '@/constants/roles'

export function AuthenticatedSections() {
  const { isAuthenticated } = useAuth()
  const { canAccessAdmin } = usePermissions()
  const { courierInfo, isCourier, levelName } = useCourier()
  
  // Only render authenticated sections after hydration
  if (!isAuthenticated) {
    return null
  }
  
  return (
    <>
      {/* Courier Growth Section */}
      {isCourier && (
        <section className="border-b border-gray-200">
          <div className="container py-24 sm:py-32 animate-fadeIn">
            <div className="max-w-3xl mx-auto text-center">
              <Crown className="h-16 w-16 text-yellow-600 mx-auto mb-6" />
              <h2 className="text-3xl font-serif tracking-tight text-gray-900 sm:text-4xl mb-6">
                信使成长之路
              </h2>
              
              <div className="bg-gradient-to-r from-yellow-50 to-amber-50 rounded-2xl p-8 mb-8">
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
                  <div className="text-center">
                    <p className="text-sm text-gray-600">当前等级</p>
                    <p className="text-2xl font-semibold text-gray-900">
                      {levelName || 'L0'}
                    </p>
                  </div>
                  <div className="text-center">
                    <p className="text-sm text-gray-600">完成任务</p>
                    <p className="text-2xl font-semibold text-gray-900">
                      {courierInfo?.completedTasks || 0}
                    </p>
                  </div>
                  <div className="text-center">
                    <p className="text-sm text-gray-600">信使积分</p>
                    <p className="text-2xl font-semibold text-gray-900">
                      {courierInfo?.points || 0}
                    </p>
                  </div>
                  <div className="text-center">
                    <p className="text-sm text-gray-600">服务评分</p>
                    <p className="text-2xl font-semibold text-gray-900">
                      {courierInfo?.averageRating?.toFixed(1) || '5.0'}⭐
                    </p>
                  </div>
                </div>
                
                <div className="flex flex-col sm:flex-row gap-4 justify-center">
                  <Link href="/courier/tasks">
                    <Button size="lg" className="group">
                      <MapPin className="mr-2 h-5 w-5" />
                      查看任务
                      <span className="ml-2 group-hover:translate-x-1 transition-transform">→</span>
                    </Button>
                  </Link>
                  <Link href={getCourierLevelManagementPath(courierInfo?.level || 1)}>
                    <Button size="lg" variant="outline">
                      管理中心
                    </Button>
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </section>
      )}
      
      {/* Admin Section */}
      {canAccessAdmin() && (
        <section className="border-b border-gray-200">
          <div className="container py-24 sm:py-32 animate-fadeIn">
            <div className="max-w-3xl mx-auto text-center">
              <Shield className="h-16 w-16 text-blue-600 mx-auto mb-6" />
              <h2 className="text-3xl font-serif tracking-tight text-gray-900 sm:text-4xl mb-6">
                管理员中心
              </h2>
              <p className="text-lg text-gray-600 mb-8">
                管理平台数据，维护社区秩序，让每一封信都安全送达
              </p>
              <Link href="/admin">
                <Button size="lg" variant="secondary">
                  进入管理后台
                </Button>
              </Link>
            </div>
          </div>
        </section>
      )}
    </>
  )
}

export function CourierCTA() {
  const { isAuthenticated } = useAuth()
  const { isCourier } = useCourier()
  
  // Only show for non-courier authenticated users
  if (!isAuthenticated || isCourier) {
    return null
  }
  
  return (
    <Link href="/courier/apply">
      <Button 
        size="lg" 
        variant="outline"
        className="group"
      >
        <Crown className="mr-2 h-5 w-5 text-yellow-600" />
        成为信使
        <span className="ml-2 group-hover:translate-x-1 transition-transform">→</span>
      </Button>
    </Link>
  )
}