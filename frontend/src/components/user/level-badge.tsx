import React from 'react'
import { cn } from '@/lib/utils'
import { Crown, Truck, Edit, Star, Medal, Award } from 'lucide-react'

interface LevelBadgeProps {
  type: 'writing' | 'courier'
  level: number
  className?: string
  showLabel?: boolean
}

const WRITING_LEVELS = {
  1: { name: '新手写手', color: 'bg-gray-100 text-gray-700', icon: Edit },
  2: { name: '熟练写手', color: 'bg-blue-100 text-blue-700', icon: Edit },
  3: { name: '优秀写手', color: 'bg-green-100 text-green-700', icon: Star },
  4: { name: '资深写手', color: 'bg-purple-100 text-purple-700', icon: Medal },
  5: { name: '大师写手', color: 'bg-yellow-100 text-yellow-700', icon: Award },
}

const COURIER_LEVELS = {
  1: { name: '楼栋信使', color: 'bg-amber-100 text-amber-700', icon: Truck },
  2: { name: '片区信使', color: 'bg-orange-100 text-orange-700', icon: Truck },
  3: { name: '校级信使', color: 'bg-red-100 text-red-700', icon: Crown },
  4: { name: '城市总代', color: 'bg-gradient-to-r from-yellow-400 to-orange-500 text-white', icon: Crown },
}

export function LevelBadge({ type, level, className, showLabel = true }: LevelBadgeProps) {
  const levelConfig = type === 'writing' ? WRITING_LEVELS : COURIER_LEVELS
  const config = levelConfig[level as keyof typeof levelConfig]
  
  if (!config) {
    return null
  }

  const Icon = config.icon

  return (
    <div className={cn(
      'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium',
      config.color,
      className
    )}>
      <Icon className="h-3 w-3" />
      {showLabel && (
        <span>
          {config.name}
          <span className="ml-1 opacity-75">Lv.{level}</span>
        </span>
      )}
      {!showLabel && <span>Lv.{level}</span>}
    </div>
  )
}

interface UserLevelDisplayProps {
  writingLevel?: number
  courierLevel?: number
  className?: string
  compact?: boolean
}

export function UserLevelDisplay({ 
  writingLevel, 
  courierLevel, 
  className, 
  compact = false 
}: UserLevelDisplayProps) {
  return (
    <div className={cn('flex items-center gap-2 flex-wrap', className)}>
      {writingLevel && writingLevel > 0 && (
        <LevelBadge 
          type="writing" 
          level={writingLevel} 
          showLabel={!compact}
        />
      )}
      {courierLevel && courierLevel > 0 && (
        <LevelBadge 
          type="courier" 
          level={courierLevel} 
          showLabel={!compact}
        />
      )}
    </div>
  )
}