/**
 * Mobile Enhanced Card Component
 * 移动端增强卡片组件 - 提供更好的移动端交互体验
 */

'use client'

import { ReactNode, useState, useRef, useEffect } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

export interface MobileEnhancedCardProps {
  children: ReactNode
  className?: string
  
  // 点击交互
  onClick?: () => void
  onDoubleClick?: () => void
  
  // 长按交互
  onLongPress?: () => void
  longPressDelay?: number
  
  // 滑动交互
  enableSwipe?: boolean
  onSwipeLeft?: () => void
  onSwipeRight?: () => void
  onSwipeUp?: () => void
  onSwipeDown?: () => void
  swipeThreshold?: number
  
  // 拖拽交互
  enableDrag?: boolean
  onDragStart?: (e: TouchEvent | MouseEvent) => void
  onDragMove?: (e: TouchEvent | MouseEvent, deltaX: number, deltaY: number) => void
  onDragEnd?: (e: TouchEvent | MouseEvent, deltaX: number, deltaY: number) => void
  
  // 视觉反馈
  enablePressEffect?: boolean
  enableHoverEffect?: boolean
  enableFocusEffect?: boolean
  
  // 可访问性
  role?: string
  tabIndex?: number
  ariaLabel?: string
}

export function MobileEnhancedCard({
  children,
  className,
  
  onClick,
  onDoubleClick,
  
  onLongPress,
  longPressDelay = 500,
  
  enableSwipe = false,
  onSwipeLeft,
  onSwipeRight,
  onSwipeUp,
  onSwipeDown,
  swipeThreshold = 50,
  
  enableDrag = false,
  onDragStart,
  onDragMove,
  onDragEnd,
  
  enablePressEffect = true,
  enableHoverEffect = true,
  enableFocusEffect = true,
  
  role = "button",
  tabIndex = 0,
  ariaLabel
}: MobileEnhancedCardProps) {
  
  const cardRef = useRef<HTMLDivElement>(null)
  const [isPressed, setIsPressed] = useState(false)
  const [isDragging, setIsDragging] = useState(false)
  
  // 触摸状态
  const touchStartRef = useRef<{ x: number; y: number; time: number } | null>(null)
  const longPressTimerRef = useRef<NodeJS.Timeout | null>(null)
  const lastTapRef = useRef<number>(0)
  const dragStartPosRef = useRef<{ x: number; y: number } | null>(null)

  // 清理定时器
  const clearLongPressTimer = () => {
    if (longPressTimerRef.current) {
      clearTimeout(longPressTimerRef.current)
      longPressTimerRef.current = null
    }
  }

  // 处理触摸开始
  const handleTouchStart = (e: React.TouchEvent) => {
    const touch = e.touches[0]
    const now = Date.now()
    
    touchStartRef.current = {
      x: touch.clientX,
      y: touch.clientY,
      time: now
    }
    
    setIsPressed(true)
    
    // 长按检测
    if (onLongPress) {
      longPressTimerRef.current = setTimeout(() => {
        onLongPress()
        setIsPressed(false)
      }, longPressDelay)
    }
    
    // 拖拽开始
    if (enableDrag) {
      dragStartPosRef.current = { x: touch.clientX, y: touch.clientY }
      onDragStart?.(e.nativeEvent)
    }
  }

  // 处理触摸移动
  const handleTouchMove = (e: React.TouchEvent) => {
    if (!touchStartRef.current) return
    
    const touch = e.touches[0]
    const deltaX = touch.clientX - touchStartRef.current.x
    const deltaY = touch.clientY - touchStartRef.current.y
    
    // 如果移动距离超过阈值，取消长按
    if (Math.abs(deltaX) > 10 || Math.abs(deltaY) > 10) {
      clearLongPressTimer()
    }
    
    // 拖拽移动
    if (enableDrag && dragStartPosRef.current) {
      setIsDragging(true)
      onDragMove?.(e.nativeEvent, deltaX, deltaY)
    }
  }

  // 处理触摸结束
  const handleTouchEnd = (e: React.TouchEvent) => {
    clearLongPressTimer()
    setIsPressed(false)
    
    if (!touchStartRef.current) return
    
    const touch = e.changedTouches[0]
    const deltaX = touch.clientX - touchStartRef.current.x
    const deltaY = touch.clientY - touchStartRef.current.y
    const deltaTime = Date.now() - touchStartRef.current.time
    
    // 拖拽结束
    if (enableDrag && isDragging) {
      setIsDragging(false)
      onDragEnd?.(e.nativeEvent, deltaX, deltaY)
      dragStartPosRef.current = null
      touchStartRef.current = null
      return
    }
    
    // 滑动检测
    if (enableSwipe && (Math.abs(deltaX) > swipeThreshold || Math.abs(deltaY) > swipeThreshold)) {
      if (Math.abs(deltaX) > Math.abs(deltaY)) {
        // 水平滑动
        if (deltaX > 0 && onSwipeRight) {
          onSwipeRight()
        } else if (deltaX < 0 && onSwipeLeft) {
          onSwipeLeft()
        }
      } else {
        // 垂直滑动
        if (deltaY > 0 && onSwipeDown) {
          onSwipeDown()
        } else if (deltaY < 0 && onSwipeUp) {
          onSwipeUp()
        }
      }
      touchStartRef.current = null
      return
    }
    
    // 点击检测
    if (Math.abs(deltaX) < 10 && Math.abs(deltaY) < 10 && deltaTime < 300) {
      const now = Date.now()
      
      // 双击检测
      if (onDoubleClick && now - lastTapRef.current < 300) {
        onDoubleClick()
        lastTapRef.current = 0
      } else {
        // 单击
        if (onClick) {
          onClick()
        }
        lastTapRef.current = now
      }
    }
    
    touchStartRef.current = null
  }

  // 处理鼠标事件（桌面端）
  const handleMouseDown = (e: React.MouseEvent) => {
    if ('ontouchstart' in window) return // 跳过移动设备
    
    setIsPressed(true)
    
    if (enableDrag) {
      dragStartPosRef.current = { x: e.clientX, y: e.clientY }
      onDragStart?.(e.nativeEvent)
    }
  }

  const handleMouseUp = (e: React.MouseEvent) => {
    if ('ontouchstart' in window) return
    
    setIsPressed(false)
    
    if (enableDrag && isDragging) {
      const deltaX = e.clientX - (dragStartPosRef.current?.x || 0)
      const deltaY = e.clientY - (dragStartPosRef.current?.y || 0)
      
      setIsDragging(false)
      onDragEnd?.(e.nativeEvent, deltaX, deltaY)
      dragStartPosRef.current = null
    } else if (onClick) {
      onClick()
    }
  }

  const handleMouseMove = (e: React.MouseEvent) => {
    if ('ontouchstart' in window || !enableDrag || !dragStartPosRef.current) return
    
    const deltaX = e.clientX - dragStartPosRef.current.x
    const deltaY = e.clientY - dragStartPosRef.current.y
    
    if (!isDragging && (Math.abs(deltaX) > 5 || Math.abs(deltaY) > 5)) {
      setIsDragging(true)
    }
    
    if (isDragging) {
      onDragMove?.(e.nativeEvent, deltaX, deltaY)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      onClick?.()
    }
  }

  // 清理定时器
  useEffect(() => {
    return () => {
      clearLongPressTimer()
    }
  }, [])

  return (
    <Card
      ref={cardRef}
      className={cn(
        // 基础样式
        "relative overflow-hidden transition-all duration-200 ease-out",
        
        // 触摸优化
        "touch-manipulation select-none",
        
        // 交互状态
        enablePressEffect && isPressed && "scale-95 shadow-sm",
        enableHoverEffect && "hover:shadow-md hover:-translate-y-0.5",
        enableFocusEffect && "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2",
        
        // 拖拽状态
        isDragging && "cursor-grabbing scale-105 shadow-lg z-10",
        enableDrag && !isDragging && "cursor-grab",
        
        // 可交互指示
        (onClick || onDoubleClick || onLongPress || enableSwipe) && "cursor-pointer",
        
        className
      )}
      
      // 触摸事件
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
      
      // 鼠标事件
      onMouseDown={handleMouseDown}
      onMouseUp={handleMouseUp}
      onMouseMove={handleMouseMove}
      onMouseLeave={() => {
        setIsPressed(false)
        clearLongPressTimer()
      }}
      
      // 键盘事件
      onKeyDown={handleKeyDown}
      
      // 可访问性
      role={role}
      tabIndex={tabIndex}
      aria-label={ariaLabel}
    >
      <CardContent className="p-0">
        {children}
      </CardContent>
      
      {/* 波纹效果指示器 */}
      {(enableSwipe || enableDrag) && (
        <div className="absolute top-2 right-2 opacity-30">
          <div className="w-2 h-2 bg-gray-400 rounded-full animate-pulse"></div>
          <div className="w-2 h-2 bg-gray-400 rounded-full animate-pulse delay-75 mt-0.5"></div>
          <div className="w-2 h-2 bg-gray-400 rounded-full animate-pulse delay-150 mt-0.5"></div>
        </div>
      )}
      
      {/* 长按进度指示器 */}
      {onLongPress && isPressed && (
        <div className="absolute bottom-0 left-0 right-0 h-1 bg-blue-500 origin-left animate-pulse">
          <div 
            className="h-full bg-blue-600 transition-all duration-500 ease-linear"
            style={{ 
              width: isPressed ? '100%' : '0%',
              transitionDuration: `${longPressDelay}ms`
            }}
          />
        </div>
      )}
    </Card>
  )
}

// 预设配置的便捷组件
export function SwipeCard({ children, onSwipeLeft, onSwipeRight, ...props }: 
  Omit<MobileEnhancedCardProps, 'enableSwipe'> & {
    onSwipeLeft?: () => void
    onSwipeRight?: () => void
  }) {
  return (
    <MobileEnhancedCard
      enableSwipe
      onSwipeLeft={onSwipeLeft}
      onSwipeRight={onSwipeRight}
      {...props}
    >
      {children}
    </MobileEnhancedCard>
  )
}

export function DraggableCard({ children, onDragEnd, ...props }: 
  Omit<MobileEnhancedCardProps, 'enableDrag'> & {
    onDragEnd?: (e: TouchEvent | MouseEvent, deltaX: number, deltaY: number) => void
  }) {
  return (
    <MobileEnhancedCard
      enableDrag
      onDragEnd={onDragEnd}
      {...props}
    >
      {children}
    </MobileEnhancedCard>
  )
}

export function InteractiveCard({ children, onClick, onLongPress, ...props }: 
  Omit<MobileEnhancedCardProps, 'onClick' | 'onLongPress'> & {
    onClick?: () => void
    onLongPress?: () => void
  }) {
  return (
    <MobileEnhancedCard
      onClick={onClick}
      onLongPress={onLongPress}
      {...props}
    >
      {children}
    </MobileEnhancedCard>
  )
}

export default MobileEnhancedCard