/**
 * SOTA Performance Monitoring System
 * 
 * Provides comprehensive performance tracking, metrics collection,
 * and optimization insights for the OpenPenPal frontend application.
 */

import React from 'react'

// ================================
// Performance Metrics Types
// ================================

export interface PerformanceMetric {
  name: string
  value: number
  unit: 'ms' | 'kb' | 'count' | '%'
  timestamp: number
  category: MetricCategory
  tags?: Record<string, string>
}

export type MetricCategory = 
  | 'render'
  | 'network'
  | 'memory'
  | 'user_interaction'
  | 'api'
  | 'cache'
  | 'navigation'

export interface PerformanceReport {
  id: string
  session_id: string
  metrics: PerformanceMetric[]
  summary: PerformanceSummary
  timestamp: string
  user_agent: string
  page_url: string
}

export interface PerformanceSummary {
  total_render_time: number
  average_api_response_time: number
  memory_usage_mb: number
  cache_hit_rate: number
  error_rate: number
  page_load_time: number
  time_to_first_byte: number
  largest_contentful_paint: number
  cumulative_layout_shift: number
  first_input_delay: number
}

// ================================
// Performance Monitor Class
// ================================

export class PerformanceMonitor {
  private static instance: PerformanceMonitor
  private metrics: PerformanceMetric[] = []
  private timers: Map<string, number> = new Map()
  private observers: Map<string, PerformanceObserver> = new Map()
  private sessionId: string
  private isEnabled: boolean = true
  private readonly maxMetrics = 1000

  constructor() {
    this.sessionId = this.generateSessionId()
    this.initializeObservers()
  }

  static getInstance(): PerformanceMonitor {
    if (!PerformanceMonitor.instance) {
      PerformanceMonitor.instance = new PerformanceMonitor()
    }
    return PerformanceMonitor.instance
  }

  // ================================
  // Core Timing Methods
  // ================================

  /**
   * Start timing an operation
   */
  startTimer(name: string, tags?: Record<string, string>): void {
    if (!this.isEnabled) return
    
    this.timers.set(name, performance.now())
    
    // Store tags for later use
    if (tags) {
      this.timers.set(`${name}_tags`, tags as any)
    }
  }

  /**
   * End timing and record metric
   */
  endTimer(name: string, category: MetricCategory = 'render'): number {
    if (!this.isEnabled) return 0
    
    const startTime = this.timers.get(name)
    if (!startTime) {
      console.warn(`Timer ${name} was not started`)
      return 0
    }

    const duration = performance.now() - startTime
    const tags = (this.timers.get(`${name}_tags`) as any) as Record<string, string> || {}

    this.addMetric({
      name,
      value: duration,
      unit: 'ms',
      timestamp: Date.now(),
      category,
      tags
    })

    this.timers.delete(name)
    this.timers.delete(`${name}_tags`)

    return duration
  }

  /**
   * Record a custom metric
   */
  recordMetric(
    name: string, 
    value: number, 
    unit: PerformanceMetric['unit'] = 'count',
    category: MetricCategory = 'user_interaction',
    tags?: Record<string, string>
  ): void {
    if (!this.isEnabled) return

    this.addMetric({
      name,
      value,
      unit,
      timestamp: Date.now(),
      category,
      tags
    })
  }

  /**
   * Measure memory usage
   */
  recordMemoryUsage(): void {
    if (!this.isEnabled || !('memory' in performance)) return

    const memory = (performance as any).memory
    
    this.addMetric({
      name: 'memory_used',
      value: Math.round(memory.usedJSHeapSize / 1024 / 1024),
      unit: 'kb',
      timestamp: Date.now(),
      category: 'memory'
    })

    this.addMetric({
      name: 'memory_limit',
      value: Math.round(memory.jsHeapSizeLimit / 1024 / 1024),
      unit: 'kb', 
      timestamp: Date.now(),
      category: 'memory'
    })
  }

  // ================================
  // Specialized Monitoring
  // ================================

  /**
   * Monitor React component render performance
   */
  wrapComponent<T extends React.ComponentType<any>>(
    Component: T,
    componentName: string
  ): T {
    if (!this.isEnabled) return Component

    return (React.forwardRef((props: any, ref: any) => {
      const startTime = React.useRef<number>()
      
      React.useLayoutEffect(() => {
        startTime.current = performance.now()
      })
      
      React.useEffect(() => {
        if (startTime.current) {
          const renderTime = performance.now() - startTime.current
          this.recordMetric(
            `component_render_${componentName}`,
            renderTime,
            'ms',
            'render',
            { component: componentName }
          )
        }
      })

      return React.createElement(Component, { ...props, ref })
    }) as unknown) as T
  }

  /**
   * Monitor API call performance
   */
  async wrapApiCall<T>(
    apiCall: () => Promise<T>,
    endpoint: string,
    method: string = 'GET'
  ): Promise<T> {
    if (!this.isEnabled) return apiCall()

    const startTime = performance.now()
    
    try {
      const result = await apiCall()
      const duration = performance.now() - startTime
      
      this.recordMetric(
        `api_call_${endpoint}`,
        duration,
        'ms',
        'api',
        { endpoint, method, status: 'success' }
      )
      
      return result
    } catch (error) {
      const duration = performance.now() - startTime
      
      this.recordMetric(
        `api_call_${endpoint}`,
        duration,
        'ms',
        'api',
        { endpoint, method, status: 'error' }
      )
      
      throw error
    }
  }

  /**
   * Monitor user interactions
   */
  recordUserInteraction(
    action: string,
    element?: string,
    duration?: number
  ): void {
    this.recordMetric(
      `user_${action}`,
      duration || 1,
      duration ? 'ms' : 'count',
      'user_interaction',
      { action, element: element || 'unknown' }
    )
  }

  // ================================
  // Web Vitals Integration
  // ================================

  /**
   * Initialize performance observers for Web Vitals
   */
  private initializeObservers(): void {
    if (typeof window === 'undefined' || !('PerformanceObserver' in window)) {
      return
    }

    // Largest Contentful Paint
    this.createObserver('largest-contentful-paint', (entries) => {
      const lcpEntry = entries[entries.length - 1]
      this.recordMetric(
        'largest_contentful_paint',
        lcpEntry.startTime,
        'ms',
        'render'
      )
    })

    // First Input Delay
    this.createObserver('first-input', (entries) => {
      entries.forEach((entry: any) => {
        this.recordMetric(
          'first_input_delay',
          entry.processingStart - entry.startTime,
          'ms',
          'user_interaction'
        )
      })
    })

    // Cumulative Layout Shift
    this.createObserver('layout-shift', (entries) => {
      let clsScore = 0
      entries.forEach((entry: any) => {
        if (!entry.hadRecentInput) {
          clsScore += entry.value
        }
      })
      
      if (clsScore > 0) {
        this.recordMetric(
          'cumulative_layout_shift',
          clsScore,
          '%',
          'render'
        )
      }
    })
  }

  private createObserver(
    entryType: string,
    callback: (entries: PerformanceEntry[]) => void
  ): void {
    try {
      const observer = new PerformanceObserver((list) => {
        callback(list.getEntries())
      })
      
      observer.observe({ entryTypes: [entryType] })
      this.observers.set(entryType, observer)
    } catch (error) {
      console.warn(`Failed to create performance observer for ${entryType}:`, error)
    }
  }

  // ================================
  // Data Management
  // ================================

  private addMetric(metric: PerformanceMetric): void {
    this.metrics.push(metric)

    // Limit stored metrics to prevent memory issues
    if (this.metrics.length > this.maxMetrics) {
      this.metrics = this.metrics.slice(-this.maxMetrics / 2)
    }
  }

  private generateSessionId(): string {
    return `perf_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  // ================================
  // Reporting and Analysis
  // ================================

  /**
   * Generate performance report
   */
  generateReport(): PerformanceReport {
    const summary = this.calculateSummary()
    
    return {
      id: `report_${Date.now()}`,
      session_id: this.sessionId,
      metrics: [...this.metrics],
      summary,
      timestamp: new Date().toISOString(),
      user_agent: navigator.userAgent,
      page_url: window.location.href
    }
  }

  private calculateSummary(): PerformanceSummary {
    const renderMetrics = this.metrics.filter(m => m.category === 'render')
    const apiMetrics = this.metrics.filter(m => m.category === 'api')
    const memoryMetrics = this.metrics.filter(m => m.category === 'memory')

    return {
      total_render_time: renderMetrics.reduce((sum, m) => sum + m.value, 0),
      average_api_response_time: apiMetrics.length > 0 
        ? apiMetrics.reduce((sum, m) => sum + m.value, 0) / apiMetrics.length 
        : 0,
      memory_usage_mb: this.getLatestMetricValue('memory_used') || 0,
      cache_hit_rate: this.calculateCacheHitRate(),
      error_rate: this.calculateErrorRate(),
      page_load_time: this.getPageLoadTime(),
      time_to_first_byte: this.getTTFB(),
      largest_contentful_paint: this.getLatestMetricValue('largest_contentful_paint') || 0,
      cumulative_layout_shift: this.getLatestMetricValue('cumulative_layout_shift') || 0,
      first_input_delay: this.getLatestMetricValue('first_input_delay') || 0
    }
  }

  private getLatestMetricValue(name: string): number | null {
    const metric = this.metrics
      .filter(m => m.name === name)
      .sort((a, b) => b.timestamp - a.timestamp)[0]
    
    return metric ? metric.value : null
  }

  private calculateCacheHitRate(): number {
    const cacheMetrics = this.metrics.filter(m => m.category === 'cache')
    const hits = cacheMetrics.filter(m => m.tags?.status === 'hit').length
    const total = cacheMetrics.length
    
    return total > 0 ? (hits / total) * 100 : 0
  }

  private calculateErrorRate(): number {
    const apiMetrics = this.metrics.filter(m => m.category === 'api')
    const errors = apiMetrics.filter(m => m.tags?.status === 'error').length
    const total = apiMetrics.length
    
    return total > 0 ? (errors / total) * 100 : 0
  }

  private getPageLoadTime(): number {
    if (typeof window === 'undefined' || !window.performance?.timing) {
      return 0
    }
    
    const timing = window.performance.timing
    return timing.loadEventEnd - timing.navigationStart
  }

  private getTTFB(): number {
    if (typeof window === 'undefined' || !window.performance?.timing) {
      return 0
    }
    
    const timing = window.performance.timing
    return timing.responseStart - timing.navigationStart
  }

  // ================================
  // Control Methods
  // ================================

  /**
   * Enable/disable performance monitoring
   */
  setEnabled(enabled: boolean): void {
    this.isEnabled = enabled
  }

  /**
   * Clear all metrics
   */
  clearMetrics(): void {
    this.metrics = []
  }

  /**
   * Get current metrics
   */
  getMetrics(): PerformanceMetric[] {
    return [...this.metrics]
  }

  /**
   * Cleanup observers
   */
  cleanup(): void {
    this.observers.forEach(observer => observer.disconnect())
    this.observers.clear()
    this.metrics = []
    this.timers.clear()
  }
}

// ================================
// React Hooks for Performance Monitoring
// ================================

/**
 * Hook for component performance monitoring
 */
export function usePerformanceMonitor(componentName: string) {
  React.useEffect(() => {
    const monitor = PerformanceMonitor.getInstance()
    monitor.startTimer(`component_mount_${componentName}`)
    
    return () => {
      monitor.endTimer(`component_mount_${componentName}`, 'render')
    }
  }, [componentName])

  const recordInteraction = React.useCallback((action: string, element?: string) => {
    const monitor = PerformanceMonitor.getInstance()
    monitor.recordUserInteraction(action, element)
  }, [])

  return { recordInteraction }
}

/**
 * Hook for API performance monitoring
 */
export function useApiPerformanceMonitor() {
  const wrapApiCall = React.useCallback(async <T>(
    apiCall: () => Promise<T>,
    endpoint: string,
    method?: string
  ): Promise<T> => {
    const monitor = PerformanceMonitor.getInstance()
    return monitor.wrapApiCall(apiCall, endpoint, method)
  }, [])

  return { wrapApiCall }
}

// ================================
// Convenience Functions
// ================================

/**
 * Quick performance timing function
 */
export function measurePerformance<T>(
  operation: () => T,
  name: string,
  category?: MetricCategory
): T {
  const monitor = PerformanceMonitor.getInstance()
  monitor.startTimer(name)
  
  try {
    const result = operation()
    monitor.endTimer(name, category)
    return result
  } catch (error) {
    monitor.endTimer(name, category)
    throw error
  }
}

/**
 * Create performance-aware component wrapper
 */
export function withPerformanceMonitoring<P extends object>(
  Component: React.ComponentType<P>,
  componentName: string
): React.ComponentType<P> {
  const monitor = PerformanceMonitor.getInstance()
  return monitor.wrapComponent(Component, componentName)
}

// Global instance
export const performanceMonitor = PerformanceMonitor.getInstance()

// Auto-start memory monitoring
if (typeof window !== 'undefined') {
  setInterval(() => {
    performanceMonitor.recordMemoryUsage()
  }, 30000) // Every 30 seconds
}

export default PerformanceMonitor