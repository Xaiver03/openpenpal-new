/**
 * 欢迎横幅组件 - 首次登录用户引导
 * Welcome Banner Component for First-time Login Users
 */

'use client'

import { useState, useEffect } from 'react'
import { X, CheckCircle, ArrowRight } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/contexts/auth-context-new'
import { getRoleDescription } from '@/lib/auth/role-navigation'
import { getRoleDisplayName, type UserRole } from '@/constants/roles'

interface WelcomeBannerProps {
  onDismiss?: () => void
  autoHide?: boolean
  autoHideDelay?: number
}

export function WelcomeBanner({ 
  onDismiss, 
  autoHide = true, 
  autoHideDelay = 10000 
}: WelcomeBannerProps) {
  const { user } = useAuth()
  const [isVisible, setIsVisible] = useState(false)
  const [countdown, setCountdown] = useState(autoHideDelay / 1000)

  useEffect(() => {
    // 检查是否显示欢迎横幅
    const urlParams = new URLSearchParams(window.location.search)
    const showWelcome = urlParams.get('welcome') === 'true'
    
    if (showWelcome && user) {
      setIsVisible(true)
      
      // 清除URL参数
      const newUrl = window.location.pathname
      window.history.replaceState({}, '', newUrl)
      
      // 自动隐藏倒计时
      if (autoHide) {
        const interval = setInterval(() => {
          setCountdown(prev => {
            if (prev <= 1) {
              handleDismiss()
              return 0
            }
            return prev - 1
          })
        }, 1000)
        
        return () => clearInterval(interval)
      }
    }
  }, [user, autoHide, autoHideDelay])

  const handleDismiss = () => {
    setIsVisible(false)
    onDismiss?.()
  }

  if (!isVisible || !user) {
    return null
  }

  const roleDescription = getRoleDescription(user)
  
  const getWelcomeMessage = () => {
    switch (user.role) {
      case 'super_admin':
      case 'admin':
        return {
          title: `欢迎，${roleDescription}！`,
          message: '您现在可以管理系统用户、查看数据分析、配置系统设置等。',
          features: [
            '用户管理和权限控制',
            '系统数据统计分析', 
            '内容审核和管理',
            '系统配置和设置'
          ]
        }
      case 'courier':
      case 'courier_coordinator': 
      case 'senior_courier':
        return {
          title: `欢迎，${roleDescription}！`,
          message: '您现在可以管理信件投递、扫码收发、查看任务等。',
          features: [
            '扫码快速投递信件',
            '查看和管理投递任务',
            '管理下级信使（如适用）',
            '查看投递统计数据'
          ]
        }
      default:
        return {
          title: '欢迎来到 OpenPenPal！',
          message: '开始您的信件写作之旅，与朋友分享美好时光。',
          features: [
            '写信给朋友和同学',
            '浏览信件广场',
            '参观信件博物馆',
            '管理个人资料'
          ]
        }
    }
  }

  const welcomeContent = getWelcomeMessage()

  return (
    <Card className="mb-6 border-green-200 bg-green-50">
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-3">
              <CheckCircle className="h-5 w-5 text-green-600" />
              <h3 className="text-lg font-semibold text-green-800">
                {welcomeContent.title}
              </h3>
            </div>
            
            <p className="text-green-700 mb-4">
              {welcomeContent.message}
            </p>
            
            <div className="space-y-2 mb-4">
              <p className="text-sm font-medium text-green-800">您可以：</p>
              <ul className="space-y-1">
                {welcomeContent.features.map((feature, index) => (
                  <li key={index} className="flex items-center gap-2 text-sm text-green-700">
                    <ArrowRight className="h-3 w-3" />
                    {feature}
                  </li>
                ))}
              </ul>
            </div>
            
            {autoHide && countdown > 0 && (
              <p className="text-xs text-green-600">
                此消息将在 {countdown} 秒后自动消失
              </p>
            )}
          </div>
          
          <Button
            variant="ghost"
            size="sm"
            onClick={handleDismiss}
            className="text-green-600 hover:text-green-800 hover:bg-green-100"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
        
        <div className="mt-4 flex gap-2">
          <Button size="sm" onClick={handleDismiss}>
            开始使用
          </Button>
          <Button variant="outline" size="sm" onClick={handleDismiss}>
            了解更多
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

export default WelcomeBanner