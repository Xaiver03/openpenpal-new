/**
 * Optimized Image Component
 * 优化图片组件
 * 
 * Provides lazy loading, WebP support, and responsive images
 * 提供懒加载、WebP支持和响应式图片
 */

'use client'

import { useState, useRef, useEffect } from 'react'
import Image, { ImageProps } from 'next/image'
import { cn } from '@/lib/utils'
import { Loader2, ImageIcon } from 'lucide-react'

interface OptimizedImageProps extends Omit<ImageProps, 'onLoad' | 'onError'> {
  fallbackSrc?: string
  showLoader?: boolean
  className?: string
  containerClassName?: string
  alt: string
  onLoad?: () => void
  onError?: () => void
}

export function OptimizedImage({
  src,
  alt,
  fallbackSrc = '/images/placeholder.png',
  showLoader = true,
  className,
  containerClassName,
  onLoad,
  onError,
  ...props
}: OptimizedImageProps) {
  const [isLoading, setIsLoading] = useState(true)
  const [hasError, setHasError] = useState(false)
  const [currentSrc, setCurrentSrc] = useState(src)
  const [isInView, setIsInView] = useState(false)
  const imgRef = useRef<HTMLDivElement>(null)

  // Intersection Observer for lazy loading
  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsInView(true)
          observer.unobserve(entry.target)
        }
      },
      {
        rootMargin: '50px' // Start loading 50px before the image comes into view
      }
    )

    if (imgRef.current) {
      observer.observe(imgRef.current)
    }

    return () => observer.disconnect()
  }, [])

  const handleLoad = () => {
    setIsLoading(false)
    setHasError(false)
    onLoad?.()
  }

  const handleError = () => {
    setIsLoading(false)
    setHasError(true)
    if (currentSrc !== fallbackSrc) {
      setCurrentSrc(fallbackSrc)
    }
    onError?.()
  }

  return (
    <div
      ref={imgRef}
      className={cn('relative overflow-hidden', containerClassName)}
    >
      {/* Loading state */}
      {isLoading && showLoader && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-100 dark:bg-gray-800">
          <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
        </div>
      )}

      {/* Error state */}
      {hasError && currentSrc === fallbackSrc && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-100 dark:bg-gray-800">
          <ImageIcon className="h-8 w-8 text-gray-400" />
        </div>
      )}

      {/* Actual image - only render when in view */}
      {isInView && (
        <Image
          {...props}
          src={currentSrc}
          alt={alt}
          className={cn(
            'transition-opacity duration-300',
            isLoading ? 'opacity-0' : 'opacity-100',
            className
          )}
          onLoad={handleLoad}
          onError={handleError}
          quality={85}
          placeholder="blur"
          blurDataURL="data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAYEBQYFBAYGBQYHBwYIChAKCgkJChQODwwQFxQYGBcUFhYaHSUfGhsjHBYWICwgIyYnKSopGR8tMC0oMCUoKSj/2wBDAQcHBwoIChMKChMoGhYaKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCj/wAARCAAIAAoDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAhEAACAQMDBQAAAAAAAAAAAAABAgMABAUGIWGRkqGx0f/EABUBAQEAAAAAAAAAAAAAAAAAAAMF/8QAGhEAAgIDAAAAAAAAAAAAAAAAAAECEgMRkf/aAAwDAQACEQMRAD8AltJagyeH0AthI5xdrLcNM91BF5pX2HaH9bcfaSXWGaRmknyLCJd5Rj+HbK6iKiUWUk8GYT8ggUKdTDcfKqV1H+zDCPGf/9k="
        />
      )}
    </div>
  )
}

/**
 * Avatar component with optimized loading
 * 优化加载的头像组件
 */
interface AvatarImageProps extends Omit<OptimizedImageProps, 'alt'> {
  name: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
}

export function AvatarImage({
  name,
  size = 'md',
  className,
  containerClassName,
  ...props
}: AvatarImageProps) {
  const sizeClasses = {
    sm: 'w-8 h-8',
    md: 'w-12 h-12',
    lg: 'w-16 h-16',
    xl: 'w-24 h-24'
  }

  const generateInitials = (name: string) => {
    return name
      .split(' ')
      .map(word => word[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  return (
    <div className={cn('relative', sizeClasses[size], containerClassName)}>
      <OptimizedImage
        {...props}
        alt={`${name}的头像`}
        className={cn('rounded-full object-cover', sizeClasses[size], className)}
        fallbackSrc=""
        onError={() => {
          // Show initials on error
        }}
      />
      
      {/* Fallback to initials */}
      <div className={cn(
        'absolute inset-0 flex items-center justify-center rounded-full bg-gradient-to-br from-blue-400 to-purple-500 text-white font-semibold',
        size === 'sm' ? 'text-xs' : size === 'md' ? 'text-sm' : size === 'lg' ? 'text-base' : 'text-lg'
      )}>
        {generateInitials(name)}
      </div>
    </div>
  )
}

/**
 * Responsive image for different screen sizes
 * 适用于不同屏幕尺寸的响应式图片
 */
interface ResponsiveImageProps extends OptimizedImageProps {
  sizes?: string
  breakpoints?: {
    mobile: string
    tablet: string
    desktop: string
  }
}

export function ResponsiveImage({
  sizes = '(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw',
  breakpoints,
  ...props
}: ResponsiveImageProps) {
  return (
    <OptimizedImage
      {...props}
      sizes={sizes}
      priority={false} // Let lazy loading handle this
    />
  )
}

export default OptimizedImage