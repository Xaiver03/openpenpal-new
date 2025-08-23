/**
 * Hook to provide stable timestamps that avoid hydration mismatches
 * 稳定时间戳 Hook，避免 SSR/CSR 水合不匹配
 */

import { useState, useEffect } from 'react'

/**
 * Returns a stable timestamp that is the same on both server and client
 * 返回在服务器和客户端相同的稳定时间戳
 */
export function useStableTimestamp(initialTimestamp?: string | Date): string {
  // Use a fixed timestamp initially to avoid hydration mismatch
  const [timestamp, setTimestamp] = useState<string>(() => {
    if (initialTimestamp) {
      return typeof initialTimestamp === 'string' 
        ? initialTimestamp 
        : initialTimestamp.toISOString()
    }
    // Return empty string during SSR to avoid mismatch
    return ''
  })

  useEffect(() => {
    // Only update timestamp on client side after hydration
    if (!timestamp && typeof window !== 'undefined') {
      setTimestamp(new Date().toISOString())
    }
  }, [timestamp])

  return timestamp
}

/**
 * Format date safely for SSR/CSR
 * 安全格式化日期，兼容 SSR/CSR
 */
export function useFormattedDate(
  date: string | Date | undefined,
  formatter: (date: Date) => string
): string {
  const [formattedDate, setFormattedDate] = useState<string>('')

  useEffect(() => {
    if (date && typeof window !== 'undefined') {
      const dateObj = typeof date === 'string' ? new Date(date) : date
      setFormattedDate(formatter(dateObj))
    }
  }, [date, formatter])

  return formattedDate
}

/**
 * Returns current time that updates only on client
 * 返回仅在客户端更新的当前时间
 */
export function useClientTime(interval?: number): Date | null {
  const [time, setTime] = useState<Date | null>(null)

  useEffect(() => {
    // Only run on client
    if (typeof window === 'undefined') return

    setTime(new Date())

    if (interval) {
      const timer = setInterval(() => {
        setTime(new Date())
      }, interval)

      return () => clearInterval(timer)
    }
  }, [interval])

  return time
}