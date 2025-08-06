'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Crown, 
  Settings, 
  Building, 
  School, 
  Home,
  ChevronUp,
  ChevronDown
} from 'lucide-react'
import { useCourierPermission } from '@/hooks/use-courier-permission'
import { cn } from '@/lib/utils'

export function ManagementFloatingButton() {
  const [isExpanded, setIsExpanded] = useState(false)
  const { 
    courierInfo,
    showManagementDashboard,
    getManagementDashboardPath,
    getCourierLevelName,
    isCourierLevel
  } = useCourierPermission()

  // 如果没有管理权限，不显示浮动按钮
  if (!showManagementDashboard() || !courierInfo) {
    return null
  }

  const levelConfig = {
    4: {
      icon: Crown,
      color: 'bg-purple-600 hover:bg-purple-700',
      borderColor: 'border-purple-200',
      textColor: 'text-purple-700',
      bgColor: 'bg-purple-50',
      title: '城市管理',
      description: '管理全市信使网络'
    },
    3: {
      icon: School,
      color: 'bg-amber-600 hover:bg-amber-700',
      borderColor: 'border-amber-200',
      textColor: 'text-amber-700',
      bgColor: 'bg-amber-50',
      title: '学校管理',
      description: '管理校内信使团队'
    },
    2: {
      icon: Building,
      color: 'bg-green-600 hover:bg-green-700',
      borderColor: 'border-green-200',
      textColor: 'text-green-700',
      bgColor: 'bg-green-50',
      title: '片区管理',
      description: '管理楼栋信使'
    },
    1: {
      icon: Home,
      color: 'bg-gray-600 hover:bg-gray-700',
      borderColor: 'border-gray-200',
      textColor: 'text-gray-700',
      bgColor: 'bg-gray-50',
      title: '个人任务',
      description: '查看个人任务'
    }
  }

  const config = levelConfig[courierInfo.level as keyof typeof levelConfig]
  const IconComponent = config.icon

  return (
    <div className="fixed bottom-6 right-6 z-50">
      {/* 展开的管理信息卡片 */}
      {isExpanded && (
        <div 
          className={cn(
            "mb-4 p-4 rounded-lg shadow-lg border backdrop-blur-sm",
            config.borderColor,
            config.bgColor
          )}
          style={{ width: '280px' }}
        >
          <div className="flex items-center gap-3 mb-3">
            <div className={cn("w-10 h-10 rounded-full flex items-center justify-center text-white", config.color)}>
              <IconComponent className="w-5 h-5" />
            </div>
            <div>
              <h3 className={cn("font-semibold", config.textColor)}>
                {config.title}
              </h3>
              <p className={cn("text-xs", config.textColor, "opacity-70")}>
                {config.description}
              </p>
            </div>
          </div>
          
          <div className="flex items-center justify-between">
            <Badge variant="outline" className={cn("text-xs", config.textColor)}>
              {getCourierLevelName()}
            </Badge>
            <Link href={getManagementDashboardPath()}>
              <Button size="sm" className={cn("text-white", config.color)}>
                进入管理
              </Button>
            </Link>
          </div>
        </div>
      )}

      {/* 主浮动按钮 */}
      <Button
        size="lg"
        className={cn(
          "rounded-full w-14 h-14 shadow-lg border-2 border-white",
          config.color
        )}
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex flex-col items-center">
          <IconComponent className="w-5 h-5 text-white" />
          {isExpanded ? (
            <ChevronDown className="w-3 h-3 text-white -mt-1" />
          ) : (
            <ChevronUp className="w-3 h-3 text-white -mt-1" />
          )}
        </div>
      </Button>
    </div>
  )
}