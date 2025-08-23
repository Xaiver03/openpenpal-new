/**
 * SafeTimestamp Component - 安全时间戳组件
 * Prevents hydration mismatch by only rendering timestamp on client
 * 通过仅在客户端渲染时间戳来防止水合不匹配
 */

'use client'

import { useState, useEffect } from 'react'

interface SafeTimestampProps {
  date?: string | Date
  format?: 'iso' | 'locale' | 'relative'
  fallback?: string
  className?: string
}

export function SafeTimestamp({ 
  date, 
  format = 'iso', 
  fallback = '',
  className 
}: SafeTimestampProps) {
  const [mounted, setMounted] = useState(false)
  const [formattedDate, setFormattedDate] = useState(fallback)

  useEffect(() => {
    setMounted(true)
    
    if (date) {
      const dateObj = typeof date === 'string' ? new Date(date) : date
      
      switch (format) {
        case 'iso':
          setFormattedDate(dateObj.toISOString())
          break
        case 'locale':
          setFormattedDate(dateObj.toLocaleString())
          break
        case 'relative':
          // Use a more stable approach for relative time
          // Import formatDistanceToNow dynamically to avoid SSR issues
          import('date-fns').then(({ formatDistanceToNow }) => {
            import('date-fns/locale').then(({ zhCN }) => {
              setFormattedDate(
                formatDistanceToNow(dateObj, {
                  addSuffix: true,
                  locale: zhCN
                })
              )
            })
          }).catch(() => {
            // Fallback to simple format if date-fns fails
            const now = new Date()
            const diff = now.getTime() - dateObj.getTime()
            const minutes = Math.floor(diff / 60000)
            const hours = Math.floor(diff / 3600000)
            const days = Math.floor(diff / 86400000)
            
            if (minutes < 1) {
              setFormattedDate('刚刚')
            } else if (minutes < 60) {
              setFormattedDate(`${minutes}分钟前`)
            } else if (hours < 24) {
              setFormattedDate(`${hours}小时前`)
            } else {
              setFormattedDate(`${days}天前`)
            }
          })
          break
      }
    }
  }, [date, format, fallback])

  // During SSR, return fallback to avoid mismatch
  if (!mounted) {
    return <span className={className}>{fallback}</span>
  }

  return <span className={className}>{formattedDate}</span>
}

/**
 * CurrentTime Component - 当前时间组件
 * Shows current time that updates periodically
 * 显示定期更新的当前时间
 */
export function CurrentTime({ 
  interval = 1000, 
  format = 'locale',
  className 
}: {
  interval?: number
  format?: 'iso' | 'locale'
  className?: string
}) {
  const [time, setTime] = useState<string>('')
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
    
    const updateTime = () => {
      const now = new Date()
      setTime(format === 'iso' ? now.toISOString() : now.toLocaleString())
    }
    
    updateTime()
    const timer = setInterval(updateTime, interval)
    
    return () => clearInterval(timer)
  }, [interval, format])

  if (!mounted) {
    return <span className={className}>--</span>
  }

  return <span className={className}>{time}</span>
}