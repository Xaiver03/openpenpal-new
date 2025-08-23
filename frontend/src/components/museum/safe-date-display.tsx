/**
 * Safe Date Display Component for Museum
 * 博物馆安全日期显示组件
 */

'use client'

import { useState, useEffect } from 'react'

interface SafeDateDisplayProps {
  date: string | Date
  className?: string
}

export function SafeDateDisplay({ date, className }: SafeDateDisplayProps) {
  const [displayText, setDisplayText] = useState<string>('加载中...')

  useEffect(() => {
    // Only calculate on client side
    if (typeof window === 'undefined') return

    const dateObj = typeof date === 'string' ? new Date(date) : date
    const now = new Date()
    const diffInMs = now.getTime() - dateObj.getTime()
    const diffInMinutes = Math.floor(diffInMs / (1000 * 60))
    const diffInHours = Math.floor(diffInMs / (1000 * 60 * 60))
    const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24))

    let text = ''
    if (diffInMinutes < 1) {
      text = '刚刚'
    } else if (diffInMinutes < 60) {
      text = `${diffInMinutes}分钟前`
    } else if (diffInHours < 24) {
      text = `${diffInHours}小时前`
    } else if (diffInDays < 30) {
      text = `${diffInDays}天前`
    } else {
      // For older dates, show the actual date
      text = dateObj.toLocaleDateString('zh-CN')
    }

    setDisplayText(text)
  }, [date])

  return <span className={className}>{displayText}</span>
}