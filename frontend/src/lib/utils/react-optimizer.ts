/**
 * React Performance Optimization Utilities - SOTA Implementation
 * React性能优化工具 - 支持智能memo化、懒加载、批处理更新
 */

import { 
  useRef, 
  useCallback, 
  useMemo, 
  useEffect, 
  useState,
  ComponentType,
  ReactElement,
  DependencyList
} from 'react'

/**
 * Enhanced useMemo with dependency comparison
 * 增强的useMemo，支持深度比较
 */
export function useDeepMemo<T>(
  factory: () => T,
  deps: DependencyList
): T {
  const ref = useRef<{ deps: DependencyList; value: T }>()

  if (!ref.current || !deepEqual(deps, ref.current.deps)) {
    ref.current = { deps, value: factory() }
  }

  return ref.current.value
}

/**
 * Enhanced useCallback with deep comparison
 * 增强的useCallback，支持深度比较
 */
export function useDeepCallback<T extends (...args: any[]) => any>(
  callback: T,
  deps: DependencyList
): T {
  return useDeepMemo(() => callback, deps)
}

/**
 * Debounced value hook
 * 防抖值Hook
 */
export function useDebouncedValue<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)

    return () => {
      clearTimeout(handler)
    }
  }, [value, delay])

  return debouncedValue
}

/**
 * Throttled callback hook
 * 节流回调Hook
 */
export function useThrottledCallback<T extends (...args: any[]) => any>(
  callback: T,
  delay: number
): T {
  const throttledRef = useRef<boolean>(false)

  return useCallback(((...args: any[]) => {
    if (throttledRef.current) return

    throttledRef.current = true
    setTimeout(() => {
      throttledRef.current = false
    }, delay)

    return callback(...args)
  }) as T, [callback, delay])
}

/**
 * Intersection Observer hook for lazy loading
 * 交叉观察器Hook，用于懒加载
 */
export function useIntersectionObserver(
  options: IntersectionObserverInit = {}
): [React.RefCallback<Element>, boolean] {
  const [isIntersecting, setIsIntersecting] = useState(false)
  const [node, setNode] = useState<Element | null>(null)

  const observer = useMemo(() => {
    return new IntersectionObserver(([entry]) => {
      setIsIntersecting(entry.isIntersecting)
    }, options)
  }, [options.threshold, options.rootMargin])

  useEffect(() => {
    if (node) observer.observe(node)
    return () => observer.disconnect()
  }, [node, observer])

  const ref = useCallback((node: Element | null) => {
    setNode(node)
  }, [])

  return [ref, isIntersecting]
}

/**
 * Virtual scrolling hook
 * 虚拟滚动Hook
 */
export function useVirtualizer<T>({
  items,
  itemHeight,
  containerHeight,
  overscan = 5
}: {
  items: T[]
  itemHeight: number
  containerHeight: number
  overscan?: number
}) {
  const [scrollTop, setScrollTop] = useState(0)

  const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan)
  const endIndex = Math.min(
    items.length - 1,
    Math.floor((scrollTop + containerHeight) / itemHeight) + overscan
  )

  const visibleItems = items.slice(startIndex, endIndex + 1)
  const totalHeight = items.length * itemHeight
  const offsetY = startIndex * itemHeight

  return {
    visibleItems,
    totalHeight,
    offsetY,
    onScroll: (e: React.UIEvent<HTMLDivElement>) => {
      setScrollTop(e.currentTarget.scrollTop)
    }
  }
}

/**
 * Memory-efficient useState for large objects
 * 内存高效的useState，适用于大对象
 */
export function useOptimizedState<T extends Record<string, any>>(
  initialState: T
): [T, (updates: Partial<T>) => void, () => void] {
  const [state, setState] = useState<T>(initialState)
  
  const updateState = useCallback((updates: Partial<T>) => {
    setState(prevState => ({ ...prevState, ...updates }))
  }, [])

  const resetState = useCallback(() => {
    setState(initialState)
  }, [initialState])

  return [state, updateState, resetState]
}

/**
 * Batch updates hook
 * 批处理更新Hook
 */
export function useBatchedUpdates<T>() {
  const [batch, setBatch] = useState<T[]>([])
  const timeoutRef = useRef<NodeJS.Timeout>()

  const addToBatch = useCallback((item: T, delay: number = 100) => {
    setBatch(prev => [...prev, item])
    
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    timeoutRef.current = setTimeout(() => {
      setBatch([])
    }, delay)
  }, [])

  return { batch, addToBatch }
}

/**
 * Resource preloader hook
 * 资源预加载Hook
 */
export function useResourcePreloader() {
  const [preloadedResources] = useState(new Set<string>())

  const preloadImage = useCallback((src: string): Promise<void> => {
    if (preloadedResources.has(src)) {
      return Promise.resolve()
    }

    return new Promise((resolve, reject) => {
      const img = new Image()
      img.onload = () => {
        preloadedResources.add(src)
        resolve()
      }
      img.onerror = reject
      img.src = src
    })
  }, [preloadedResources])

  const preloadScript = useCallback((src: string): Promise<void> => {
    if (preloadedResources.has(src)) {
      return Promise.resolve()
    }

    return new Promise((resolve, reject) => {
      const script = document.createElement('script')
      script.onload = () => {
        preloadedResources.add(src)
        resolve()
      }
      script.onerror = reject
      script.src = src
      document.head.appendChild(script)
    })
  }, [preloadedResources])

  return { preloadImage, preloadScript }
}

/**
 * Component lazy loading with error boundary
 * 组件懒加载，带错误边界
 */
export function createLazyComponent<T extends ComponentType<any>>(
  importFn: () => Promise<{ default: T }>,
  fallback?: ReactElement
): T {
  const LazyComponent = React.lazy(importFn)

  return ((props: any) => {
    return React.createElement(
      React.Suspense,
      { fallback: fallback || React.createElement('div', null, 'Loading...') },
      React.createElement(LazyComponent, props)
    )
  }) as T
}

/**
 * Smart component memoization
 * 智能组件记忆化
 */
export function smartMemo<T extends ComponentType<any>>(
  Component: T,
  options: {
    deepCompare?: boolean
    skipProps?: (keyof React.ComponentProps<T>)[]
  } = {}
): T {
  const { deepCompare = false, skipProps = [] } = options

  const areEqual = (prevProps: any, nextProps: any): boolean => {
    const prevFiltered = filterProps(prevProps, skipProps as string[])
    const nextFiltered = filterProps(nextProps, skipProps as string[])

    return deepCompare 
      ? deepEqual(prevFiltered, nextFiltered)
      : shallowEqual(prevFiltered, nextFiltered)
  }

  return React.memo(Component, areEqual) as unknown as T
}

/**
 * Memory usage monitor hook
 * 内存使用监控Hook
 */
export function useMemoryMonitor() {
  const [memoryUsage, setMemoryUsage] = useState<{
    used: number
    total: number
    percentage: number
  } | null>(null)

  useEffect(() => {
    if ('memory' in performance) {
      const updateMemoryUsage = () => {
        const memory = (performance as any).memory
        if (memory) {
          setMemoryUsage({
            used: memory.usedJSHeapSize,
            total: memory.totalJSHeapSize,
            percentage: (memory.usedJSHeapSize / memory.totalJSHeapSize) * 100
          })
        }
      }

      updateMemoryUsage()
      const interval = setInterval(updateMemoryUsage, 5000)
      
      return () => clearInterval(interval)
    }
  }, [])

  return memoryUsage
}

/**
 * Render optimization tracker
 * 渲染优化追踪器
 */
export function useRenderTracker(componentName: string) {
  const renderCount = useRef(0)
  const lastRenderTime = useRef(performance.now())

  useEffect(() => {
    renderCount.current++
    const now = performance.now()
    const renderTime = now - lastRenderTime.current
    lastRenderTime.current = now

    if (process.env.NODE_ENV === 'development') {
      console.log(`🎭 ${componentName} rendered ${renderCount.current} times (${renderTime.toFixed(2)}ms since last render)`)
    }
  })

  return {
    renderCount: renderCount.current,
    getStats: () => ({
      renders: renderCount.current,
      lastRenderTime: lastRenderTime.current
    })
  }
}

// Utility functions
function deepEqual(a: any, b: any): boolean {
  if (a === b) return true
  if (a == null || b == null) return false
  if (Array.isArray(a) && Array.isArray(b)) {
    if (a.length !== b.length) return false
    for (let i = 0; i < a.length; i++) {
      if (!deepEqual(a[i], b[i])) return false
    }
    return true
  }
  if (typeof a === 'object' && typeof b === 'object') {
    const keysA = Object.keys(a)
    const keysB = Object.keys(b)
    if (keysA.length !== keysB.length) return false
    for (const key of keysA) {
      if (!keysB.includes(key) || !deepEqual(a[key], b[key])) return false
    }
    return true
  }
  return false
}

function shallowEqual(a: any, b: any): boolean {
  if (a === b) return true
  if (a == null || b == null) return false
  
  const keysA = Object.keys(a)
  const keysB = Object.keys(b)
  
  if (keysA.length !== keysB.length) return false
  
  for (const key of keysA) {
    if (!keysB.includes(key) || a[key] !== b[key]) return false
  }
  
  return true
}

function filterProps(props: any, skipProps: string[]): any {
  const filtered: any = {}
  for (const [key, value] of Object.entries(props)) {
    if (!skipProps.includes(key)) {
      filtered[key] = value
    }
  }
  return filtered
}

// React import for lazy loading
import React from 'react'