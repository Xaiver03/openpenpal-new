'use client'

import { useEffect, useCallback, useRef } from 'react'

/**
 * 页面性能监控Hook
 */
export function usePerformanceMonitor() {
  const metricsRef = useRef<any>({})

  useEffect(() => {
    // 监控Core Web Vitals
    const observer = new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        if (entry.entryType === 'largest-contentful-paint') {
          metricsRef.current.lcp = entry.startTime
        }
        if (entry.entryType === 'first-input') {
          metricsRef.current.fid = (entry as any).processingStart - entry.startTime
        }
        if (entry.entryType === 'layout-shift') {
          if (!(entry as any).hadRecentInput) {
            metricsRef.current.cls = (metricsRef.current.cls || 0) + (entry as any).value
          }
        }
      }
    })

    // 监控不同类型的性能指标
    try {
      observer.observe({ entryTypes: ['largest-contentful-paint'] })
      observer.observe({ entryTypes: ['first-input'] })
      observer.observe({ entryTypes: ['layout-shift'] })
    } catch (e) {
      // 某些浏览器可能不支持特定的entryTypes
      console.warn('Performance monitoring not fully supported')
    }

    return () => observer.disconnect()
  }, [])

  const getMetrics = useCallback(() => {
    return {
      ...metricsRef.current,
      // 添加Navigation Timing API数据
      navigationStart: performance.timing?.navigationStart,
      loadComplete: performance.timing?.loadEventEnd,
      domReady: performance.timing?.domContentLoadedEventEnd,
      // 计算关键时间
      ttfb: performance.timing?.responseStart - performance.timing?.navigationStart,
      domLoad: performance.timing?.domContentLoadedEventEnd - performance.timing?.navigationStart,
      pageLoad: performance.timing?.loadEventEnd - performance.timing?.navigationStart,
    }
  }, [])

  return { getMetrics }
}

/**
 * 防抖Hook
 */
export function useDebounce<T extends (...args: any[]) => any>(
  callback: T,
  delay: number
): T {
  const timeoutRef = useRef<NodeJS.Timeout>()

  const debouncedCallback = useCallback((...args: Parameters<T>) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
    
    timeoutRef.current = setTimeout(() => {
      callback(...args)
    }, delay)
  }, [callback, delay]) as T

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  return debouncedCallback
}

/**
 * 节流Hook
 */
export function useThrottle<T extends (...args: any[]) => any>(
  callback: T,
  delay: number
): T {
  const lastRun = useRef(Date.now())

  const throttledCallback = useCallback((...args: Parameters<T>) => {
    if (Date.now() - lastRun.current >= delay) {
      callback(...args)
      lastRun.current = Date.now()
    }
  }, [callback, delay]) as T

  return throttledCallback
}

/**
 * 交叉观察器Hook (用于懒加载)
 */
export function useIntersectionObserver(
  elementRef: React.RefObject<Element>,
  options: IntersectionObserverInit = {}
) {
  const { threshold = 0.1, rootMargin = '0px', root = null } = options
  const isIntersecting = useRef(false)

  useEffect(() => {
    const element = elementRef.current
    if (!element) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        isIntersecting.current = entry.isIntersecting
      },
      { threshold, rootMargin, root }
    )

    observer.observe(element)
    return () => observer.disconnect()
  }, [threshold, rootMargin, root])

  return isIntersecting.current
}

/**
 * 资源预加载Hook
 */
export function usePreload() {
  const preloadResource = useCallback((url: string, type: 'image' | 'script' | 'style' = 'image') => {
    const link = document.createElement('link')
    link.rel = 'preload'
    link.href = url
    
    switch (type) {
      case 'image':
        link.as = 'image'
        break
      case 'script':
        link.as = 'script'
        break
      case 'style':
        link.as = 'style'
        break
    }
    
    document.head.appendChild(link)
  }, [])

  const preloadImages = useCallback((urls: string[]) => {
    urls.forEach(url => preloadResource(url, 'image'))
  }, [preloadResource])

  return { preloadResource, preloadImages }
}

/**
 * 内存使用监控Hook
 */
export function useMemoryMonitor() {
  const getMemoryInfo = useCallback(() => {
    if ('memory' in performance) {
      const memory = (performance as any).memory
      return {
        usedJSHeapSize: memory.usedJSHeapSize,
        totalJSHeapSize: memory.totalJSHeapSize,
        jsHeapSizeLimit: memory.jsHeapSizeLimit,
        usage: memory.usedJSHeapSize / memory.jsHeapSizeLimit
      }
    }
    return null
  }, [])

  useEffect(() => {
    const interval = setInterval(() => {
      const memInfo = getMemoryInfo()
      if (memInfo && memInfo.usage > 0.9) {
        console.warn('High memory usage detected:', memInfo)
      }
    }, 30000) // 每30秒检查一次

    return () => clearInterval(interval)
  }, [getMemoryInfo])

  return { getMemoryInfo }
}