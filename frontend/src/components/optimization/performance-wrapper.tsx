'use client'

import { Suspense, lazy, memo, useCallback, useMemo } from 'react'
import { ErrorBoundary } from 'react-error-boundary'

/**
 * Performance optimization wrapper components
 * Implements code splitting, lazy loading, and error boundaries
 */

interface PerformanceWrapperProps {
  children: React.ReactNode
  fallback?: React.ReactNode
  errorFallback?: React.ComponentType<any>
  enableLazyLoading?: boolean
}

// Lazy loading wrapper with suspense
export const LazyWrapper = memo(function LazyWrapper({ 
  children, 
  fallback = <LoadingSpinner />,
  errorFallback = DefaultErrorFallback,
  enableLazyLoading = true 
}: PerformanceWrapperProps) {
  if (!enableLazyLoading) {
    return <>{children}</>
  }

  return (
    <ErrorBoundary FallbackComponent={errorFallback}>
      <Suspense fallback={fallback}>
        {children}
      </Suspense>
    </ErrorBoundary>
  )
})

// Default loading spinner
const LoadingSpinner = memo(function LoadingSpinner() {
  return (
    <div className="flex items-center justify-center py-8" role="status" aria-label="Loading">
      <div className="relative">
        <div className="w-8 h-8 border-4 border-amber-200 border-t-amber-600 rounded-full animate-spin"></div>
        <span className="sr-only">Loading...</span>
      </div>
    </div>
  )
})

// Default error fallback
function DefaultErrorFallback({ error, resetErrorBoundary }: any) {
  return (
    <div className="flex flex-col items-center justify-center py-8 px-4 text-center">
      <div className="text-red-600 mb-4">
        <svg className="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
        </svg>
      </div>
      <h3 className="text-lg font-semibold text-gray-900 mb-2">出现错误</h3>
      <p className="text-gray-600 mb-4 max-w-md">
        {error?.message || '加载组件时出现问题，请重试。'}
      </p>
      <button
        onClick={resetErrorBoundary}
        className="px-4 py-2 bg-amber-600 text-white rounded-md hover:bg-amber-700 transition-colors"
      >
        重试
      </button>
    </div>
  )
}

// Optimized image component with lazy loading
interface OptimizedImageProps {
  src: string
  alt: string
  width?: number
  height?: number
  className?: string
  priority?: boolean
  quality?: number
}

export const OptimizedImage = memo(function OptimizedImage({
  src,
  alt,
  width,
  height,
  className = '',
  priority = false,
  quality = 75
}: OptimizedImageProps) {
  const imageProps = useMemo(() => ({
    src,
    alt,
    width,
    height,
    className: `transition-opacity duration-300 ${className}`,
    loading: priority ? 'eager' : 'lazy' as 'eager' | 'lazy',
    decoding: 'async' as 'async',
    quality
  }), [src, alt, width, height, className, priority, quality])

  return (
    <img
      {...imageProps}
      onLoad={(e) => {
        e.currentTarget.style.opacity = '1'
      }}
      style={{ opacity: 0 }}
    />
  )
})

// Virtual scrolling for large lists
interface VirtualScrollProps {
  items: any[]
  itemHeight: number
  containerHeight: number
  renderItem: (item: any, index: number) => React.ReactNode
  className?: string
}

export const VirtualScroll = memo(function VirtualScroll({
  items,
  itemHeight,
  containerHeight,
  renderItem,
  className = ''
}: VirtualScrollProps) {
  const [startIndex, setStartIndex] = useState(0)
  
  const visibleCount = Math.ceil(containerHeight / itemHeight)
  const endIndex = Math.min(startIndex + visibleCount + 1, items.length)
  const visibleItems = items.slice(startIndex, endIndex)
  
  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    const scrollTop = e.currentTarget.scrollTop
    const newStartIndex = Math.floor(scrollTop / itemHeight)
    setStartIndex(newStartIndex)
  }, [itemHeight])
  
  const totalHeight = items.length * itemHeight
  const offsetY = startIndex * itemHeight
  
  return (
    <div
      className={`overflow-auto ${className}`}
      style={{ height: containerHeight }}
      onScroll={handleScroll}
    >
      <div style={{ height: totalHeight, position: 'relative' }}>
        <div style={{ transform: `translateY(${offsetY}px)` }}>
          {visibleItems.map((item, index) => 
            renderItem(item, startIndex + index)
          )}
        </div>
      </div>
    </div>
  )
})

// Debounced input for search
interface DebouncedInputProps {
  value: string
  onChange: (value: string) => void
  delay?: number
  placeholder?: string
  className?: string
}

export const DebouncedInput = memo(function DebouncedInput({
  value,
  onChange,
  delay = 300,
  placeholder,
  className = ''
}: DebouncedInputProps) {
  const [localValue, setLocalValue] = useState(value)
  
  useEffect(() => {
    setLocalValue(value)
  }, [value])
  
  useEffect(() => {
    const timer = setTimeout(() => {
      onChange(localValue)
    }, delay)
    
    return () => clearTimeout(timer)
  }, [localValue, delay, onChange])
  
  return (
    <input
      type="text"
      value={localValue}
      onChange={(e) => setLocalValue(e.target.value)}
      placeholder={placeholder}
      className={className}
    />
  )
})

// Intersection observer for infinite scroll
interface IntersectionTriggerProps {
  onIntersect: () => void
  threshold?: number
  rootMargin?: string
  children?: React.ReactNode
}

export const IntersectionTrigger = memo(function IntersectionTrigger({
  onIntersect,
  threshold = 0.1,
  rootMargin = '100px',
  children
}: IntersectionTriggerProps) {
  const ref = useRef<HTMLDivElement>(null)
  
  useEffect(() => {
    const element = ref.current
    if (!element) return
    
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          onIntersect()
        }
      },
      { threshold, rootMargin }
    )
    
    observer.observe(element)
    
    return () => observer.unobserve(element)
  }, [onIntersect, threshold, rootMargin])
  
  return (
    <div ref={ref} className="intersection-trigger">
      {children}
    </div>
  )
})

// Enhanced Performance metrics hook
export function usePerformanceMetrics() {
  const [metrics, setMetrics] = useState({
    loadTime: 0,
    renderTime: 0,
    interactionTime: 0,
    lcp: 0,
    fid: 0,
    cls: 0,
    ttfb: 0
  })
  
  useEffect(() => {
    // Measure initial load time and Core Web Vitals
    if (typeof window !== 'undefined' && 'performance' in window) {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming
      if (navigation) {
        setMetrics(prev => ({
          ...prev,
          loadTime: navigation.loadEventEnd - navigation.loadEventStart,
          ttfb: navigation.responseStart - navigation.fetchStart
        }))
      }

      // Monitor Core Web Vitals
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.entryType === 'largest-contentful-paint') {
            setMetrics(prev => ({ ...prev, lcp: entry.startTime }))
          }
          if (entry.entryType === 'first-input') {
            setMetrics(prev => ({ ...prev, fid: (entry as any).processingStart - entry.startTime }))
          }
          if (entry.entryType === 'layout-shift') {
            const entry_value = (entry as any).value
            if (!(entry as any).hadRecentInput) {
              setMetrics(prev => ({ ...prev, cls: prev.cls + entry_value }))
            }
          }
        }
      })

      try {
        observer.observe({ entryTypes: ['largest-contentful-paint'] })
        observer.observe({ entryTypes: ['first-input'] })
        observer.observe({ entryTypes: ['layout-shift'] })
      } catch (e) {
        console.warn('Some performance metrics not supported')
      }

      return () => observer.disconnect()
    }
  }, [])
  
  const measureRender = useCallback((componentName: string) => {
    const startTime = performance.now()
    
    return () => {
      const endTime = performance.now()
      const renderTime = endTime - startTime
      
      setMetrics(prev => ({
        ...prev,
        renderTime: renderTime
      }))
      
      if (process.env.NODE_ENV === 'development') {
        console.debug(`${componentName} render time: ${renderTime.toFixed(2)}ms`)
      }
    }
  }, [])

  const reportMetrics = useCallback(() => {
    if (process.env.NODE_ENV === 'development') {
      console.group('📊 Performance Report')
      console.log('🎯 Core Web Vitals:')
      console.log(`  LCP: ${metrics.lcp.toFixed(2)}ms`)
      console.log(`  FID: ${metrics.fid.toFixed(2)}ms`) 
      console.log(`  CLS: ${metrics.cls.toFixed(4)}`)
      console.log('⏱️ Load Metrics:')
      console.log(`  TTFB: ${metrics.ttfb.toFixed(2)}ms`)
      console.log(`  Load Time: ${metrics.loadTime.toFixed(2)}ms`)
      console.groupEnd()
    }
  }, [metrics])
  
  return { metrics, measureRender, reportMetrics }
}

// Lazy load pages
export const LazyPage = (importFunc: () => Promise<any>) => {
  return lazy(() => importFunc())
}

// Export common lazy-loaded components
// Note: These can be uncommented when the corresponding pages exist
// export const LazyCourierTasks = LazyPage(() => import('../../../app/courier/tasks/page'))
// export const LazyWritePage = LazyPage(() => import('../../../app/(main)/write/page'))
// export const LazyMailboxPage = LazyPage(() => import('../../../app/(main)/mailbox/page'))
// export const LazySettingsPage = LazyPage(() => import('../../../app/settings/page'))

import { useState, useEffect, useRef } from 'react'