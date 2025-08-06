/**
 * Performance utilities and optimizations
 */

// Memoization utility
export function memoize<T extends (...args: any[]) => any>(fn: T): T {
  const cache = new Map()
  
  return ((...args: any[]) => {
    const key = JSON.stringify(args)
    
    if (cache.has(key)) {
      return cache.get(key)
    }
    
    const result = fn(...args)
    cache.set(key, result)
    
    // Limit cache size to prevent memory leaks
    if (cache.size > 100) {
      const firstKey = cache.keys().next().value
      cache.delete(firstKey)
    }
    
    return result
  }) as T
}

// Throttle function
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: NodeJS.Timeout | null = null
  let lastExecTime = 0
  
  return (...args: Parameters<T>) => {
    const currentTime = Date.now()
    
    if (currentTime - lastExecTime > delay) {
      func(...args)
      lastExecTime = currentTime
    } else {
      if (timeoutId) {
        clearTimeout(timeoutId)
      }
      
      timeoutId = setTimeout(() => {
        func(...args)
        lastExecTime = Date.now()
      }, delay - (currentTime - lastExecTime))
    }
  }
}

// Debounce function
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: NodeJS.Timeout | null = null
  
  return (...args: Parameters<T>) => {
    if (timeoutId) {
      clearTimeout(timeoutId)
    }
    
    timeoutId = setTimeout(() => {
      func(...args)
    }, delay)
  }
}

// Image preloader
export function preloadImages(urls: string[]): Promise<void[]> {
  return Promise.all(
    urls.map(url => {
      return new Promise<void>((resolve, reject) => {
        const img = new Image()
        img.onload = () => resolve()
        img.onerror = () => reject(new Error(`Failed to load image: ${url}`))
        img.src = url
      })
    })
  )
}

// Resource preloader
export function preloadResource(url: string, type: 'script' | 'style' | 'font' | 'image'): Promise<void> {
  return new Promise((resolve, reject) => {
    const link = document.createElement('link')
    link.rel = 'preload'
    link.href = url
    
    switch (type) {
      case 'script':
        link.as = 'script'
        break
      case 'style':
        link.as = 'style'
        break
      case 'font':
        link.as = 'font'
        link.crossOrigin = 'anonymous'
        break
      case 'image':
        link.as = 'image'
        break
    }
    
    link.onload = () => resolve()
    link.onerror = () => reject(new Error(`Failed to preload ${type}: ${url}`))
    
    document.head.appendChild(link)
  })
}

// Web Workers utility
export class WebWorkerManager {
  private workers: Map<string, Worker> = new Map()
  
  createWorker(name: string, scriptUrl: string): Worker {
    if (this.workers.has(name)) {
      return this.workers.get(name)!
    }
    
    const worker = new Worker(scriptUrl)
    this.workers.set(name, worker)
    
    return worker
  }
  
  terminateWorker(name: string): void {
    const worker = this.workers.get(name)
    if (worker) {
      worker.terminate()
      this.workers.delete(name)
    }
  }
  
  terminateAll(): void {
    this.workers.forEach(worker => worker.terminate())
    this.workers.clear()
  }
}

// Performance monitoring
export class PerformanceMonitor {
  private marks: Map<string, number> = new Map()
  private measures: Map<string, number> = new Map()
  
  mark(name: string): void {
    this.marks.set(name, performance.now())
    
    if (typeof performance.mark === 'function') {
      performance.mark(name)
    }
  }
  
  measure(name: string, startMark: string, endMark?: string): number {
    const startTime = this.marks.get(startMark)
    const endTime = endMark ? this.marks.get(endMark) : performance.now()
    
    if (startTime === undefined) {
      throw new Error(`Start mark "${startMark}" not found`)
    }
    
    if (endMark && endTime === undefined) {
      throw new Error(`End mark "${endMark}" not found`)
    }
    
    const duration = (endTime as number) - startTime
    this.measures.set(name, duration)
    
    if (typeof performance.measure === 'function') {
      if (endMark) {
        performance.measure(name, startMark, endMark)
      } else {
        performance.measure(name, startMark)
      }
    }
    
    return duration
  }
  
  getMeasure(name: string): number | undefined {
    return this.measures.get(name)
  }
  
  getAllMeasures(): Record<string, number> {
    return Object.fromEntries(this.measures)
  }
  
  clear(): void {
    this.marks.clear()
    this.measures.clear()
    
    if (typeof performance.clearMarks === 'function') {
      performance.clearMarks()
    }
    
    if (typeof performance.clearMeasures === 'function') {
      performance.clearMeasures()
    }
  }
}

// Critical resource hints
export function addResourceHints(resources: Array<{
  url: string
  type: 'preload' | 'prefetch' | 'preconnect' | 'dns-prefetch'
  as?: string
  crossorigin?: boolean
}>) {
  resources.forEach(({ url, type, as, crossorigin }) => {
    const link = document.createElement('link')
    link.rel = type
    link.href = url
    
    if (as) {
      link.setAttribute('as', as)
    }
    
    if (crossorigin) {
      link.crossOrigin = 'anonymous'
    }
    
    document.head.appendChild(link)
  })
}

// Bundle analyzer simulation
export function analyzeBundleSize(bundleName: string): Promise<{
  size: number
  gzipSize: number
  loadTime: number
}> {
  return new Promise((resolve) => {
    const start = performance.now()
    
    // Simulate bundle analysis
    fetch(`/_next/static/chunks/${bundleName}`)
      .then(response => {
        const size = parseInt(response.headers.get('content-length') || '0')
        const gzipSize = Math.floor(size * 0.3) // Estimate gzip compression
        const loadTime = performance.now() - start
        
        resolve({ size, gzipSize, loadTime })
      })
      .catch(() => {
        resolve({ size: 0, gzipSize: 0, loadTime: 0 })
      })
  })
}

// Memory usage monitor
export function getMemoryUsage(): {
  used: number
  total: number
  percentage: number
} | null {
  if (typeof window !== 'undefined' && 'memory' in performance) {
    const memory = (performance as any).memory
    return {
      used: memory.usedJSHeapSize,
      total: memory.totalJSHeapSize,
      percentage: (memory.usedJSHeapSize / memory.totalJSHeapSize) * 100
    }
  }
  
  return null
}

// Network information
export function getNetworkInfo(): {
  type: string
  downlink: number
  rtt: number
  effectiveType: string
} | null {
  if (typeof navigator !== 'undefined' && 'connection' in navigator) {
    const connection = (navigator as any).connection
    return {
      type: connection.type || 'unknown',
      downlink: connection.downlink || 0,
      rtt: connection.rtt || 0,
      effectiveType: connection.effectiveType || 'unknown'
    }
  }
  
  return null
}

// Service Worker utilities
export class ServiceWorkerManager {
  async register(scriptUrl: string): Promise<ServiceWorkerRegistration | null> {
    if ('serviceWorker' in navigator) {
      try {
        const registration = await navigator.serviceWorker.register(scriptUrl)
        console.log('Service Worker registered:', registration)
        return registration
      } catch (error) {
        console.error('Service Worker registration failed:', error)
        return null
      }
    }
    
    return null
  }
  
  async unregister(): Promise<boolean> {
    if ('serviceWorker' in navigator) {
      const registrations = await navigator.serviceWorker.getRegistrations()
      
      const unregisterPromises = registrations.map(registration => 
        registration.unregister()
      )
      
      const results = await Promise.all(unregisterPromises)
      return results.every(result => result)
    }
    
    return false
  }
  
  postMessage(data: any): void {
    if ('serviceWorker' in navigator && navigator.serviceWorker.controller) {
      navigator.serviceWorker.controller.postMessage(data)
    }
  }
}

// Intersection Observer utility
export function createIntersectionObserver(
  callback: IntersectionObserverCallback,
  options?: IntersectionObserverInit
): IntersectionObserver {
  return new IntersectionObserver(callback, {
    threshold: 0.1,
    rootMargin: '50px',
    ...options
  })
}

// Critical path CSS
export function inlineCriticalCSS(css: string): void {
  const style = document.createElement('style')
  style.textContent = css
  style.setAttribute('data-critical', 'true')
  document.head.insertBefore(style, document.head.firstChild)
}

// Lazy loading utility
export function createLazyLoader<T>(
  loader: () => Promise<T>,
  delay: number = 0
): Promise<T> {
  return new Promise((resolve, reject) => {
    setTimeout(async () => {
      try {
        const result = await loader()
        resolve(result)
      } catch (error) {
        reject(error)
      }
    }, delay)
  })
}

// Performance metrics singleton
export const performanceMonitor = new PerformanceMonitor()
export const webWorkerManager = new WebWorkerManager()
export const serviceWorkerManager = new ServiceWorkerManager()

// Web Vitals tracking
export function trackWebVitals(metric: any): void {
  if (process.env.NODE_ENV === 'development') {
    console.log(`Web Vital - ${metric.name}:`, metric.value)
  }
  
  // Send to analytics
  if (typeof window !== 'undefined' && 'gtag' in window) {
    (window as any).gtag('event', metric.name, {
      value: metric.value,
      metric_id: metric.id,
      metric_delta: metric.delta
    })
  }
}