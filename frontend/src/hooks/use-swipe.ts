import { useState, useEffect, useRef } from 'react'

interface TouchPosition {
  x: number
  y: number
}

interface SwipeConfig {
  minSwipeDistance?: number
  maxSwipeTime?: number
  preventDefaultTouchmoveEvent?: boolean
}

interface SwipeHandlers {
  onSwipeLeft?: () => void
  onSwipeRight?: () => void
  onSwipeUp?: () => void
  onSwipeDown?: () => void
  onTouchStart?: (position: TouchPosition) => void
  onTouchEnd?: (position: TouchPosition) => void
}

export function useSwipe(
  handlers: SwipeHandlers,
  config: SwipeConfig = {}
) {
  const {
    minSwipeDistance = 50,
    maxSwipeTime = 500,
    preventDefaultTouchmoveEvent = false
  } = config

  const [touchStart, setTouchStart] = useState<TouchPosition | null>(null)
  const [touchEnd, setTouchEnd] = useState<TouchPosition | null>(null)
  const touchStartTime = useRef<number>(0)

  const onTouchStart = (event: TouchEvent) => {
    const touch = event.targetTouches[0]
    const position = {
      x: touch.clientX,
      y: touch.clientY
    }
    setTouchEnd(null)
    setTouchStart(position)
    touchStartTime.current = Date.now()
    handlers.onTouchStart?.(position)
  }

  const onTouchMove = (event: TouchEvent) => {
    if (preventDefaultTouchmoveEvent) {
      event.preventDefault()
    }
  }

  const onTouchEnd = (event: TouchEvent) => {
    if (!touchStart) return
    
    const touch = event.changedTouches[0]
    const position = {
      x: touch.clientX,
      y: touch.clientY
    }
    setTouchEnd(position)
    handlers.onTouchEnd?.(position)

    const touchEndTime = Date.now()
    const swipeTime = touchEndTime - touchStartTime.current

    if (swipeTime > maxSwipeTime) return

    const distanceX = touchStart.x - position.x
    const distanceY = touchStart.y - position.y
    const isLeftSwipe = distanceX > minSwipeDistance
    const isRightSwipe = distanceX < -minSwipeDistance
    const isUpSwipe = distanceY > minSwipeDistance
    const isDownSwipe = distanceY < -minSwipeDistance

    // 判断是水平还是垂直滑动（取绝对值更大的方向）
    const isHorizontal = Math.abs(distanceX) > Math.abs(distanceY)

    if (isHorizontal) {
      if (isLeftSwipe) {
        handlers.onSwipeLeft?.()
      }
      if (isRightSwipe) {
        handlers.onSwipeRight?.()
      }
    } else {
      if (isUpSwipe) {
        handlers.onSwipeUp?.()
      }
      if (isDownSwipe) {
        handlers.onSwipeDown?.()
      }
    }
  }

  return {
    onTouchStart,
    onTouchMove,
    onTouchEnd,
    touchStart,
    touchEnd
  }
}

// React component hook wrapper
export function useSwipeElement(
  elementRef: React.RefObject<HTMLElement>,
  handlers: SwipeHandlers,
  config?: SwipeConfig
) {
  const swipe = useSwipe(handlers, config)

  useEffect(() => {
    const element = elementRef.current
    if (!element) return

    element.addEventListener('touchstart', swipe.onTouchStart, { passive: false })
    element.addEventListener('touchmove', swipe.onTouchMove, { passive: false })
    element.addEventListener('touchend', swipe.onTouchEnd, { passive: false })

    return () => {
      element.removeEventListener('touchstart', swipe.onTouchStart)
      element.removeEventListener('touchmove', swipe.onTouchMove)
      element.removeEventListener('touchend', swipe.onTouchEnd)
    }
  }, [elementRef, swipe])

  return swipe
}