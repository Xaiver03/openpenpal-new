'use client'

import React from 'react'
import { Badge } from '@/components/ui/badge'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Star, Crown, Award, Trophy } from 'lucide-react'
import { getCreditLevelName, getLevelProgress, getPointsToNextLevel } from '@/lib/api/credit'

interface CreditLevelBadgeProps {
  level: number
  totalPoints: number
  showTooltip?: boolean
  variant?: 'default' | 'secondary' | 'outline'
  size?: 'sm' | 'default' | 'lg'
  className?: string
}

export function CreditLevelBadge({
  level,
  totalPoints,
  showTooltip = true,
  variant = 'default',
  size = 'default',
  className = ''
}: CreditLevelBadgeProps) {
  const levelName = getCreditLevelName(level)
  const progress = getLevelProgress(totalPoints, level)
  const pointsToNext = getPointsToNextLevel(totalPoints, level)
  
  // 根据等级选择图标和颜色
  const getLevelIcon = (level: number) => {
    if (level >= 6) return <Crown className="h-3 w-3" />
    if (level >= 4) return <Trophy className="h-3 w-3" />
    if (level >= 2) return <Award className="h-3 w-3" />
    return <Star className="h-3 w-3" />
  }
  
  const getLevelColor = (level: number) => {
    if (level >= 6) return 'bg-gradient-to-r from-yellow-400 to-orange-500 text-white'
    if (level >= 4) return 'bg-gradient-to-r from-purple-400 to-pink-500 text-white'
    if (level >= 2) return 'bg-gradient-to-r from-blue-400 to-indigo-500 text-white'
    return 'bg-gradient-to-r from-green-400 to-blue-500 text-white'
  }

  const badge = (
    <Badge 
      variant={variant}
      className={`
        ${getLevelColor(level)}
        ${size === 'sm' ? 'text-xs px-2 py-0.5' : size === 'lg' ? 'text-sm px-3 py-1' : 'text-sm px-2 py-1'}
        font-semibold
        ${className}
      `}
    >
      {getLevelIcon(level)}
      <span className="ml-1">{levelName}</span>
    </Badge>
  )

  if (!showTooltip) {
    return badge
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          {badge}
        </TooltipTrigger>
        <TooltipContent side="bottom" className="max-w-xs">
          <div className="space-y-2">
            <div className="font-semibold text-center">{levelName} (等级 {level})</div>
            <div className="text-sm">
              <div>总积分: {totalPoints.toLocaleString()}</div>
              {pointsToNext > 0 ? (
                <>
                  <div>升级还需: {pointsToNext.toLocaleString()} 积分</div>
                  <div className="mt-2">
                    <div className="bg-gray-200 rounded-full h-2 dark:bg-gray-700">
                      <div 
                        className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                        style={{ width: `${progress}%` }}
                      />
                    </div>
                    <div className="text-xs text-center mt-1">{progress.toFixed(1)}%</div>
                  </div>
                </>
              ) : (
                <div className="text-amber-500 font-medium">🎉 已达到最高等级！</div>
              )}
            </div>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}

export default CreditLevelBadge