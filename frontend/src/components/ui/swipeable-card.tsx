'use client'

import React, { useState, useRef } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useSwipeElement } from '@/hooks/use-swipe'
import { Eye, Edit, X } from 'lucide-react'

interface SwipeableCardProps {
  children: React.ReactNode
  onSwipeLeft?: () => void
  onSwipeRight?: () => void
  onView?: () => void
  onEdit?: () => void
  canEdit?: boolean
  className?: string
  swipeThreshold?: number
}

export function SwipeableCard({
  children,
  onSwipeLeft,
  onSwipeRight,
  onView,
  onEdit,
  canEdit = true,
  className = '',
  swipeThreshold = 100
}: SwipeableCardProps) {
  const [isSwipedLeft, setIsSwipedLeft] = useState(false)
  const [isSwipedRight, setIsSwipedRight] = useState(false)
  const [touchOffset, setTouchOffset] = useState(0)
  const cardRef = useRef<HTMLDivElement>(null)

  const resetSwipe = () => {
    setIsSwipedLeft(false)
    setIsSwipedRight(false)
    setTouchOffset(0)
  }

  useSwipeElement(cardRef, {
    onSwipeLeft: () => {
      if (onEdit && canEdit) {
        setIsSwipedLeft(true)
        onSwipeLeft?.()
      }
    },
    onSwipeRight: () => {
      if (onView) {
        setIsSwipedRight(true)
        onSwipeRight?.()
      }
    },
    onTouchStart: () => {
      resetSwipe()
    }
  }, { minSwipeDistance: swipeThreshold })

  return (
    <div className="relative overflow-hidden rounded-lg">
      {/* 右滑动作 - 查看 */}
      {isSwipedRight && (
        <div className="absolute inset-y-0 left-0 flex items-center justify-start bg-blue-500 text-white px-4 w-20 rounded-l-lg z-10">
          <Eye className="w-5 h-5" />
        </div>
      )}

      {/* 左滑动作 - 编辑 */}
      {isSwipedLeft && canEdit && (
        <div className="absolute inset-y-0 right-0 flex items-center justify-end bg-amber-500 text-white px-4 w-20 rounded-r-lg z-10">
          <Edit className="w-5 h-5" />
        </div>
      )}

      {/* 卡片内容 */}
      <Card
        ref={cardRef}
        className={`relative transition-all duration-200 ${
          isSwipedLeft ? 'transform -translate-x-20' : 
          isSwipedRight ? 'transform translate-x-20' : ''
        } ${className} touch-pan-y`}
        style={{
          transform: touchOffset !== 0 ? `translateX(${touchOffset}px)` : undefined
        }}
      >
        {children}
        
        {/* 滑动提示指示器 */}
        {(isSwipedLeft || isSwipedRight) && (
          <div className="absolute top-2 right-2 z-20">
            <Button
              variant="ghost"
              size="sm"
              onClick={resetSwipe}
              className="h-6 w-6 p-0 bg-white/80 hover:bg-white"
            >
              <X className="w-3 h-3" />
            </Button>
          </div>
        )}
      </Card>

      {/* 滑动操作按钮 */}
      {isSwipedRight && onView && (
        <div className="absolute inset-y-0 left-0 flex items-center justify-center w-20">
          <Button
            onClick={() => {
              onView()
              resetSwipe()
            }}
            className="bg-blue-500 hover:bg-blue-600 text-white rounded-full h-12 w-12 p-0"
          >
            <Eye className="w-5 h-5" />
          </Button>
        </div>
      )}

      {isSwipedLeft && onEdit && canEdit && (
        <div className="absolute inset-y-0 right-0 flex items-center justify-center w-20">
          <Button
            onClick={() => {
              onEdit()
              resetSwipe()
            }}
            className="bg-amber-500 hover:bg-amber-600 text-white rounded-full h-12 w-12 p-0"
          >
            <Edit className="w-5 h-5" />
          </Button>
        </div>
      )}
    </div>
  )
}