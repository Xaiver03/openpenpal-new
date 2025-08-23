/**
 * Performance Tracking Component
 * 性能追踪组件
 * 
 * Integrates with the performance monitor to track user interactions and page performance
 * 集成性能监控器，追踪用户交互和页面性能
 */

'use client'

import { useEffect, useRef, useState } from 'react'
import { performanceMonitor } from '@/lib/utils/performance-monitor'

interface PerformanceTrackerProps {
  pageName: string
  trackInteractions?: boolean
  reportInterval?: number
}

export function PerformanceTracker({ 
  pageName, 
  trackInteractions = true, 
  reportInterval = 30000 
}: PerformanceTrackerProps) {
  const startTimeRef = useRef<number>()
  const [isVisible, setIsVisible] = useState(true)

  useEffect(() => {
    // Mark page start
    startTimeRef.current = Date.now()
    performanceMonitor.mark(`${pageName}-start`)

    // Track page view
    performanceMonitor.recordMetric('page_view', 1, 'count', 'navigation', {
      page: pageName,
      timestamp: Date.now()
    })

    // Track initial page load performance
    if (typeof window !== 'undefined' && window.performance) {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming
      if (navigation) {
        performanceMonitor.recordMetric('page_load_time', navigation.loadEventEnd - navigation.fetchStart, 'ms', 'navigation', {
          page: pageName
        })
      }
    }

    return () => {
      // Mark page end and measure time spent
      if (startTimeRef.current) {
        const timeSpent = Date.now() - startTimeRef.current
        performanceMonitor.recordMetric('time_on_page', timeSpent, 'ms', 'user_interaction', {
          page: pageName
        })
        
        performanceMonitor.mark(`${pageName}-end`)
        performanceMonitor.measure(`${pageName}-session`, `${pageName}-start`, `${pageName}-end`)
      }
    }
  }, [pageName])

  // Track page visibility changes
  useEffect(() => {
    const handleVisibilityChange = () => {
      const wasVisible = isVisible
      const nowVisible = !document.hidden
      setIsVisible(nowVisible)

      if (wasVisible && !nowVisible) {
        // Page became hidden
        performanceMonitor.recordMetric('page_hidden', 1, 'count', 'user_interaction', {
          page: pageName,
          timestamp: Date.now()
        })
      } else if (!wasVisible && nowVisible) {
        // Page became visible
        performanceMonitor.recordMetric('page_visible', 1, 'count', 'user_interaction', {
          page: pageName,
          timestamp: Date.now()
        })
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange)
  }, [isVisible, pageName])

  // Track user interactions
  useEffect(() => {
    if (!trackInteractions) return

    const trackClick = (event: MouseEvent) => {
      const target = event.target as HTMLElement
      const tagName = target.tagName.toLowerCase()
      const id = target.id || ''
      const className = target.className || ''

      performanceMonitor.recordMetric('user_click', 1, 'count', 'user_interaction', {
        page: pageName,
        element: tagName,
        elementId: id,
        elementClass: className,
        timestamp: Date.now()
      })
    }

    const trackScroll = () => {
      const scrollDepth = Math.round((window.scrollY / (document.body.scrollHeight - window.innerHeight)) * 100)
      
      performanceMonitor.recordMetric('scroll_depth', scrollDepth, '%', 'user_interaction', {
        page: pageName,
        timestamp: Date.now()
      })
    }

    const trackKeypress = () => {
      performanceMonitor.recordMetric('keypress', 1, 'count', 'user_interaction', {
        page: pageName,
        timestamp: Date.now()
      })
    }

    // Throttled scroll tracking
    let scrollTimeout: number
    const throttledScroll = () => {
      if (scrollTimeout) clearTimeout(scrollTimeout)
      scrollTimeout = window.setTimeout(trackScroll, 500)
    }

    document.addEventListener('click', trackClick)
    window.addEventListener('scroll', throttledScroll, { passive: true })
    document.addEventListener('keydown', trackKeypress)

    return () => {
      document.removeEventListener('click', trackClick)
      window.removeEventListener('scroll', throttledScroll)
      document.removeEventListener('keydown', trackKeypress)
      if (scrollTimeout) clearTimeout(scrollTimeout)
    }
  }, [trackInteractions, pageName])

  // Periodic performance reporting
  useEffect(() => {
    const interval = setInterval(() => {
      // Record current memory usage
      const memoryUsage = performanceMonitor.getMemoryUsage()
      if (memoryUsage > 0) {
        performanceMonitor.recordMetric('memory_usage', memoryUsage, 'MB', 'performance', {
          page: pageName
        })
      }

      // Record cache hit rate
      const cacheHitRate = performanceMonitor.calculateCacheHitRate()
      if (cacheHitRate > 0) {
        performanceMonitor.recordMetric('cache_hit_rate', cacheHitRate, '%', 'performance', {
          page: pageName
        })
      }

      // Check for long tasks (>50ms)
      if (typeof window !== 'undefined' && 'PerformanceObserver' in window) {
        try {
          const observer = new PerformanceObserver((list) => {
            for (const entry of list.getEntries()) {
              if (entry.duration > 50) {
                performanceMonitor.recordMetric('long_task', entry.duration, 'ms', 'performance', {
                  page: pageName,
                  taskType: (entry as any).name || 'unknown'
                })
              }
            }
          })
          observer.observe({ entryTypes: ['longtask'] })
        } catch (e) {
          // Long task API not supported
        }
      }
    }, reportInterval)

    return () => clearInterval(interval)
  }, [pageName, reportInterval])

  return null // This is a tracking component, no UI needed
}

/**
 * Hook for tracking custom performance metrics
 * 用于追踪自定义性能指标的Hook
 */
export function usePerformanceTracking(componentName: string) {
  const startTimeRef = useRef<number>()

  useEffect(() => {
    startTimeRef.current = Date.now()
    performanceMonitor.mark(`${componentName}-mount`)

    return () => {
      if (startTimeRef.current) {
        const mountTime = Date.now() - startTimeRef.current
        performanceMonitor.recordMetric('component_mount_time', mountTime, 'ms', 'performance', {
          component: componentName
        })
      }
    }
  }, [componentName])

  const trackAction = (actionName: string, data?: Record<string, any>) => {
    performanceMonitor.recordMetric('user_action', 1, 'count', 'user_interaction', {
      component: componentName,
      action: actionName,
      ...data
    })
  }

  const measureOperation = async <T,>(operationName: string, operation: () => Promise<T>): Promise<T> => {
    const start = Date.now()
    performanceMonitor.mark(`${componentName}-${operationName}-start`)

    try {
      const result = await operation()
      const duration = Date.now() - start
      
      performanceMonitor.recordMetric('operation_duration', duration, 'ms', 'performance', {
        component: componentName,
        operation: operationName,
        success: true
      })

      return result
    } catch (error) {
      const duration = Date.now() - start
      
      performanceMonitor.recordMetric('operation_duration', duration, 'ms', 'performance', {
        component: componentName,
        operation: operationName,
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error'
      })

      throw error
    }
  }

  return {
    trackAction,
    measureOperation
  }
}

/**
 * Component for tracking form performance
 * 用于追踪表单性能的组件
 */
interface FormPerformanceTrackerProps {
  formName: string
  onSubmit?: () => void
}

export function FormPerformanceTracker({ formName, onSubmit }: FormPerformanceTrackerProps) {
  const startTimeRef = useRef<number>()
  const { trackAction } = usePerformanceTracking(`form-${formName}`)

  useEffect(() => {
    startTimeRef.current = Date.now()
    trackAction('form_started')

    const trackFormInteraction = (event: Event) => {
      const target = event.target as HTMLElement
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT') {
        trackAction('form_field_interaction', {
          fieldType: target.tagName.toLowerCase(),
          fieldName: target.getAttribute('name') || '',
          eventType: event.type
        })
      }
    }

    document.addEventListener('focus', trackFormInteraction, true)
    document.addEventListener('blur', trackFormInteraction, true)
    document.addEventListener('change', trackFormInteraction, true)

    return () => {
      document.removeEventListener('focus', trackFormInteraction, true)
      document.removeEventListener('blur', trackFormInteraction, true)
      document.removeEventListener('change', trackFormInteraction, true)

      if (startTimeRef.current) {
        const timeSpent = Date.now() - startTimeRef.current
        performanceMonitor.recordMetric('form_time_spent', timeSpent, 'ms', 'user_interaction', {
          form: formName
        })
      }
    }
  }, [formName, trackAction])

  useEffect(() => {
    if (onSubmit) {
      trackAction('form_submitted')
      
      if (startTimeRef.current) {
        const completionTime = Date.now() - startTimeRef.current
        performanceMonitor.recordMetric('form_completion_time', completionTime, 'ms', 'user_interaction', {
          form: formName
        })
      }
    }
  }, [onSubmit, formName, trackAction])

  return null
}

export default PerformanceTracker