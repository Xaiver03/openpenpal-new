/**
 * Performance Monitoring Utility
 * 性能监控工具
 * 
 * Tracks web vitals, performance metrics, and provides insights
 * 追踪Web性能指标并提供洞察
 */

import { log } from '@/lib/utils/simple-logger'

interface PerformanceEntry {
  name: string
  value: number
  timestamp: number
  unit?: string
  category?: string
  metadata?: Record<string, any>
}

interface PerformanceReport {
  timestamp: number
  entries: PerformanceEntry[]
  summary: {
    first_input_delay: number
    cumulative_layout_shift: number
    largest_contentful_paint: number
    time_to_first_byte: number
    cache_hit_rate: number
    memory_usage_mb: number
  }
  webVitals: Record<string, number>
}

class PerformanceMonitor {
  private entries: PerformanceEntry[] = []
  private maxEntries: number = 1000
  private observers: Map<string, PerformanceObserver> = new Map()
  
  constructor() {
    if (typeof window !== 'undefined') {
      this.initializeObservers()
    }
  }

  /**
   * Initialize performance observers
   * 初始化性能观察器
   */
  private initializeObservers(): void {
    // Layout shift observer
    if ('PerformanceObserver' in window) {
      try {
        const clsObserver = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (entry.entryType === 'layout-shift' && !(entry as any).hadRecentInput) {
              this.recordMetric('cls', (entry as any).value, '', 'web_vitals')
            }
          }
        })
        clsObserver.observe({ entryTypes: ['layout-shift'] })
        this.observers.set('cls', clsObserver)
      } catch (e) {
        log.warn('Failed to initialize CLS observer', e)
      }

      // Largest contentful paint observer
      try {
        const lcpObserver = new PerformanceObserver((list) => {
          const entries = list.getEntries()
          const lastEntry = entries[entries.length - 1]
          this.recordMetric('lcp', lastEntry.startTime, 'ms', 'web_vitals')
        })
        lcpObserver.observe({ entryTypes: ['largest-contentful-paint'] })
        this.observers.set('lcp', lcpObserver)
      } catch (e) {
        log.warn('Failed to initialize LCP observer', e)
      }

      // First input delay observer
      try {
        const fidObserver = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (entry.entryType === 'first-input') {
              const delay = (entry as any).processingStart - entry.startTime
              this.recordMetric('fid', delay, 'ms', 'web_vitals')
            }
          }
        })
        fidObserver.observe({ entryTypes: ['first-input'] })
        this.observers.set('fid', fidObserver)
      } catch (e) {
        log.warn('Failed to initialize FID observer', e)
      }
    }
  }

  /**
   * Record a performance metric
   * 记录性能指标
   */
  recordMetric(
    name: string,
    value: number,
    unit: string = '',
    category: string = 'custom',
    metadata?: Record<string, any>
  ): void {
    const entry: PerformanceEntry = {
      name,
      value,
      timestamp: Date.now(),
      unit,
      category,
      metadata
    }

    this.entries.push(entry)

    // Enforce max entries limit
    if (this.entries.length > this.maxEntries) {
      this.entries = this.entries.slice(-this.maxEntries)
    }

    log.debug('Performance metric recorded', { name, value, unit, category })
  }

  /**
   * Record navigation timing
   * 记录导航计时
   */
  recordNavigationTiming(): void {
    if (typeof window === 'undefined' || !window.performance?.timing) return

    const timing = window.performance.timing
    const navigationStart = timing.navigationStart

    // Time to first byte
    if (timing.responseStart > 0) {
      this.recordMetric(
        'ttfb',
        timing.responseStart - navigationStart,
        'ms',
        'navigation'
      )
    }

    // DOM content loaded
    if (timing.domContentLoadedEventEnd > 0) {
      this.recordMetric(
        'dom_content_loaded',
        timing.domContentLoadedEventEnd - navigationStart,
        'ms',
        'navigation'
      )
    }

    // Page load complete
    if (timing.loadEventEnd > 0) {
      this.recordMetric(
        'page_load_complete',
        timing.loadEventEnd - navigationStart,
        'ms',
        'navigation'
      )
    }
  }

  /**
   * Track resource loading
   * 跟踪资源加载
   */
  trackResourceLoading(): void {
    if (typeof window === 'undefined' || !window.performance?.getEntriesByType) return

    const resources = window.performance.getEntriesByType('resource')
    
    resources.forEach(resource => {
      const resourceEntry = resource as PerformanceResourceTiming
      if (resourceEntry.duration > 100) { // Only track slow resources
        this.recordMetric(
          'slow_resource',
          resourceEntry.duration,
          'ms',
          'resources',
          {
            name: resourceEntry.name,
            type: resourceEntry.initiatorType,
            size: resourceEntry.transferSize
          }
        )
      }
    })
  }

  /**
   * Get memory usage
   * 获取内存使用情况
   */
  getMemoryUsage(): number {
    if (typeof window === 'undefined') return 0
    
    const performance = window.performance as any
    if (performance.memory) {
      return Math.round(performance.memory.usedJSHeapSize / 1048576) // Convert to MB
    }
    
    return 0
  }

  /**
   * Calculate cache hit rate
   * 计算缓存命中率
   */
  calculateCacheHitRate(): number {
    if (typeof window === 'undefined' || !window.performance?.getEntriesByType) return 0

    const resources = window.performance.getEntriesByType('resource') as PerformanceResourceTiming[]
    if (resources.length === 0) return 0

    const cachedResources = resources.filter(r => r.transferSize === 0 && r.decodedBodySize > 0)
    return (cachedResources.length / resources.length) * 100
  }

  /**
   * Generate performance report
   * 生成性能报告
   */
  generateReport(): PerformanceReport {
    const webVitals = this.getWebVitals()
    
    return {
      timestamp: Date.now(),
      entries: [...this.entries],
      summary: {
        first_input_delay: webVitals.FID,
        cumulative_layout_shift: webVitals.CLS,
        largest_contentful_paint: webVitals.LCP,
        time_to_first_byte: webVitals.TTFB,
        cache_hit_rate: this.calculateCacheHitRate(),
        memory_usage_mb: this.getMemoryUsage()
      },
      webVitals
    }
  }

  /**
   * Get web vitals
   * 获取Web核心指标
   */
  private getWebVitals(): Record<string, number> {
    const vitals: Record<string, number> = {
      FID: 0,
      CLS: 0,
      LCP: 0,
      TTFB: 0,
      FCP: 0,
      INP: 0
    }

    // Get latest values from entries
    this.entries.forEach(entry => {
      if (entry.category === 'web_vitals') {
        switch (entry.name) {
          case 'fid':
            vitals.FID = entry.value
            break
          case 'cls':
            vitals.CLS += entry.value // CLS is cumulative
            break
          case 'lcp':
            vitals.LCP = entry.value
            break
          case 'ttfb':
            vitals.TTFB = entry.value
            break
          case 'fcp':
            vitals.FCP = entry.value
            break
          case 'inp':
            vitals.INP = Math.max(vitals.INP, entry.value)
            break
        }
      }
    })

    return vitals
  }

  /**
   * Mark performance timing
   * 标记性能时间点
   */
  mark(markName: string): void {
    if (typeof window !== 'undefined' && window.performance?.mark) {
      window.performance.mark(markName)
    }
  }

  /**
   * Measure between marks
   * 测量标记之间的时间
   */
  measure(measureName: string, startMark: string, endMark?: string): void {
    if (typeof window !== 'undefined' && window.performance?.measure) {
      try {
        window.performance.measure(measureName, startMark, endMark)
        const measures = window.performance.getEntriesByName(measureName, 'measure')
        if (measures.length > 0) {
          const measure = measures[measures.length - 1]
          this.recordMetric(measureName, measure.duration, 'ms', 'user_timing')
        }
      } catch (e) {
        log.warn('Failed to measure performance', { measureName, error: e })
      }
    }
  }

  /**
   * Clear all entries
   * 清除所有条目
   */
  clear(): void {
    this.entries = []
  }

  /**
   * Destroy observers
   * 销毁观察器
   */
  destroy(): void {
    this.observers.forEach(observer => observer.disconnect())
    this.observers.clear()
  }
}

// Singleton instance
export const performanceMonitor = new PerformanceMonitor()

// Auto-record navigation timing when page loads
if (typeof window !== 'undefined') {
  window.addEventListener('load', () => {
    performanceMonitor.recordNavigationTiming()
    performanceMonitor.trackResourceLoading()
  })
}

export default performanceMonitor