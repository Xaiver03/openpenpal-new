'use client'

import React from 'react'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { Star } from 'lucide-react'
import { getCreditLevelName, getLevelProgress, getPointsToNextLevel } from '@/lib/api/credit'
import { LEVEL_UP_POINTS } from '@/types/credit'

interface CreditProgressBarProps {
  currentLevel: number
  totalPoints: number
  showLabels?: boolean
  showNextLevel?: boolean
  animated?: boolean
  className?: string
}

export function CreditProgressBar({
  currentLevel,
  totalPoints,
  showLabels = true,
  showNextLevel = true,
  animated = true,
  className = ''
}: CreditProgressBarProps) {
  const progress = getLevelProgress(totalPoints, currentLevel)
  const pointsToNext = getPointsToNextLevel(totalPoints, currentLevel)
  const currentLevelName = getCreditLevelName(currentLevel)
  const nextLevelName = getCreditLevelName(currentLevel + 1)
  
  const currentLevelPoints = LEVEL_UP_POINTS[currentLevel - 1] || 0
  const nextLevelPoints = LEVEL_UP_POINTS[currentLevel] || LEVEL_UP_POINTS[LEVEL_UP_POINTS.length - 1]
  
  const isMaxLevel = currentLevel >= LEVEL_UP_POINTS.length

  return (
    <div className={`space-y-3 ${className}`}>
      {/* 当前等级和积分 */}
      {showLabels && (
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Badge variant="secondary" className="text-xs">
              <Star className="h-3 w-3 mr-1" />
              {currentLevelName}
            </Badge>
            <span className="text-sm text-muted-foreground">
              {totalPoints.toLocaleString()} 积分
            </span>
          </div>
          
          {showNextLevel && !isMaxLevel && (
            <div className="text-sm text-muted-foreground">
              还需 {pointsToNext.toLocaleString()} 积分升级
            </div>
          )}
        </div>
      )}

      {/* 进度条 */}
      <div className="space-y-2">
        <Progress 
          value={isMaxLevel ? 100 : progress} 
          className={`h-3 ${animated ? 'transition-all duration-700 ease-out' : ''}`}
        />
        
        {/* 进度条标签 */}
        <div className="flex justify-between text-xs text-muted-foreground">
          <span>{currentLevelPoints.toLocaleString()}</span>
          {!isMaxLevel && (
            <span>{nextLevelPoints.toLocaleString()}</span>
          )}
        </div>
      </div>

      {/* 下一等级预览 */}
      {showNextLevel && !isMaxLevel && (
        <div className="flex items-center justify-center p-2 bg-muted/50 rounded-lg">
          <div className="text-center">
            <div className="text-sm font-medium">下一等级</div>
            <div className="text-xs text-muted-foreground">
              {nextLevelName} ({nextLevelPoints.toLocaleString()} 积分)
            </div>
          </div>
        </div>
      )}

      {/* 最高等级提示 */}
      {isMaxLevel && showNextLevel && (
        <div className="flex items-center justify-center p-3 bg-gradient-to-r from-amber-50 to-orange-50 dark:from-amber-950 dark:to-orange-950 rounded-lg border border-amber-200 dark:border-amber-800">
          <div className="text-center">
            <div className="text-sm font-medium text-amber-700 dark:text-amber-300">
              🎉 恭喜您已达到最高等级！
            </div>
            <div className="text-xs text-amber-600 dark:text-amber-400 mt-1">
              继续积累积分可以兑换更多奖励
            </div>
          </div>
        </div>
      )}

      {/* 等级里程碑 */}
      {showLabels && (
        <div className="grid grid-cols-3 gap-2 text-xs">
          {LEVEL_UP_POINTS.slice(1, 4).map((points, index) => {
            const level = index + 2
            const isAchieved = currentLevel >= level
            const isCurrent = currentLevel === level
            
            return (
              <div
                key={level}
                className={`text-center p-2 rounded transition-colors ${
                  isCurrent
                    ? 'bg-primary/10 border border-primary/20 text-primary'
                    : isAchieved
                    ? 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300'
                    : 'bg-muted/50 text-muted-foreground'
                }`}
              >
                <div className="font-medium">
                  {getCreditLevelName(level)}
                </div>
                <div className="text-xs opacity-75">
                  {points.toLocaleString()}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

export default CreditProgressBar