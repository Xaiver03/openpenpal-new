'use client'

import { lazy, Suspense, ComponentType } from 'react'
import { Skeleton } from '@/components/ui/skeleton'

interface LazyLoadingProps {
  fallback?: React.ReactNode
  className?: string
}

/**
 * 高阶组件：为组件添加懒加载功能
 */
export function withLazyLoading(
  importFn: () => Promise<{ default: ComponentType<any> }>,
  fallback?: React.ReactNode
) {
  const LazyComponent = lazy(importFn)
  
  return function LazyLoadedComponent(props: any) {
    const { className, ...restProps } = props
    
    const defaultFallback = (
      <div className={className}>
        <Skeleton className="h-4 w-full mb-2" />
        <Skeleton className="h-4 w-3/4 mb-2" />
        <Skeleton className="h-4 w-1/2" />
      </div>
    )
    
    return (
      <Suspense fallback={fallback || defaultFallback}>
        <LazyComponent {...restProps} />
      </Suspense>
    )
  }
}

/**
 * 图片懒加载组件
 */
interface LazyImageProps {
  src: string
  alt: string
  className?: string
  fallback?: string
  onLoad?: () => void
  onError?: () => void
}

export function LazyImage({ 
  src, 
  alt, 
  className = '', 
  fallback = '/placeholder.svg',
  onLoad,
  onError 
}: LazyImageProps) {
  return (
    <img
      src={src}
      alt={alt}
      className={className}
      loading="lazy"
      decoding="async"
      onLoad={onLoad}
      onError={(e) => {
        const target = e.target as HTMLImageElement
        target.src = fallback
        onError?.()
      }}
    />
  )
}

/**
 * 内容懒加载容器
 */
interface LazyContentProps {
  children: React.ReactNode
  threshold?: number
  className?: string
  fallback?: React.ReactNode
}

export function LazyContent({ 
  children, 
  threshold = 0.1, 
  className = '',
  fallback 
}: LazyContentProps) {
  return (
    <div className={className}>
      <Suspense fallback={fallback || <Skeleton className="h-20 w-full" />}>
        {children}
      </Suspense>
    </div>
  )
}