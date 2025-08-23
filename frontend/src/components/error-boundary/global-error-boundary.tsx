/**
 * Global Error Boundary System
 * 全局错误边界系统
 * 
 * Comprehensive error handling with fallback UI and error reporting
 * 全面的错误处理，包含回退UI和错误报告
 */

'use client'

import React, { Component, ReactNode, ErrorInfo } from 'react'
import { AlertTriangle, RotateCcw, Home, Bug, Copy } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import Link from 'next/link'

interface ErrorBoundaryState {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
  errorId: string
}

interface ErrorBoundaryProps {
  children: ReactNode
  fallback?: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
  resetOnPropsChange?: boolean
  level?: 'page' | 'component' | 'global'
}

export class GlobalErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  private retryTimeoutId: number | null = null

  constructor(props: ErrorBoundaryProps) {
    super(props)

    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: ''
    }
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return {
      hasError: true,
      error,
      errorId: `error-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({
      error,
      errorInfo,
    })

    // Log error to console in development
    if (process.env.NODE_ENV === 'development') {
      console.group('🚨 Error Boundary Caught an Error')
      console.error('Error:', error)
      console.error('Error Info:', errorInfo)
      console.error('Component Stack:', errorInfo.componentStack)
      console.groupEnd()
    }

    // Report error to external service in production
    this.reportError(error, errorInfo)

    // Call custom error handler
    this.props.onError?.(error, errorInfo)
  }

  componentDidUpdate(prevProps: ErrorBoundaryProps) {
    const { resetOnPropsChange } = this.props
    const { hasError } = this.state

    if (hasError && prevProps.resetOnPropsChange && resetOnPropsChange) {
      if (prevProps.children !== this.props.children) {
        this.resetError()
      }
    }
  }

  componentWillUnmount() {
    if (this.retryTimeoutId) {
      clearTimeout(this.retryTimeoutId)
    }
  }

  private reportError = async (error: Error, errorInfo: ErrorInfo) => {
    try {
      const errorReport = {
        id: this.state.errorId,
        message: error.message,
        stack: error.stack,
        componentStack: errorInfo.componentStack,
        url: typeof window !== 'undefined' ? window.location.href : '',
        userAgent: typeof window !== 'undefined' ? window.navigator.userAgent : '',
        timestamp: new Date().toISOString(),
        level: this.props.level || 'component'
      }

      // Log to browser console
      console.error('Error Report:', errorReport)

      // In production, send to error reporting service
      if (process.env.NODE_ENV === 'production' && typeof window !== 'undefined') {
        // Example: await fetch('/api/errors', { method: 'POST', body: JSON.stringify(errorReport) })
      }
    } catch (reportingError) {
      console.error('Failed to report error:', reportingError)
    }
  }

  private resetError = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: ''
    })
  }

  private retryWithDelay = () => {
    this.retryTimeoutId = window.setTimeout(() => {
      this.resetError()
    }, 1000)
  }

  private copyErrorDetails = () => {
    const { error, errorInfo, errorId } = this.state
    const errorDetails = {
      id: errorId,
      message: error?.message || 'Unknown error',
      stack: error?.stack || 'No stack trace available',
      componentStack: errorInfo?.componentStack || 'No component stack available',
      timestamp: new Date().toISOString()
    }

    navigator.clipboard.writeText(JSON.stringify(errorDetails, null, 2))
      .then(() => {
        // Could show a toast notification here
        console.log('Error details copied to clipboard')
      })
      .catch(() => {
        console.log('Failed to copy error details')
      })
  }

  render() {
    const { hasError, error, errorInfo, errorId } = this.state
    const { children, fallback, level = 'component' } = this.props

    if (hasError) {
      // Use custom fallback if provided
      if (fallback) {
        return fallback
      }

      // Different fallback UIs based on error level
      if (level === 'global') {
        return (
          <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
            <Card className="w-full max-w-md">
              <CardHeader className="text-center">
                <div className="mx-auto w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mb-4">
                  <AlertTriangle className="w-8 h-8 text-red-600" />
                </div>
                <CardTitle className="text-xl">应用程序出错</CardTitle>
                <CardDescription>
                  很抱歉，应用程序遇到了意外错误
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <Alert variant="destructive">
                  <Bug className="h-4 w-4" />
                  <AlertDescription>
                    {error?.message || '发生了未知错误'}
                  </AlertDescription>
                </Alert>

                <div className="space-y-2">
                  <Button
                    onClick={() => window.location.reload()}
                    className="w-full"
                    size="sm"
                  >
                    <RotateCcw className="w-4 h-4 mr-2" />
                    重新加载页面
                  </Button>

                  <Button
                    asChild
                    variant="outline"
                    className="w-full"
                    size="sm"
                  >
                    <Link href="/">
                      <Home className="w-4 h-4 mr-2" />
                      回到首页
                    </Link>
                  </Button>

                  {process.env.NODE_ENV === 'development' && (
                    <Button
                      onClick={this.copyErrorDetails}
                      variant="ghost"
                      className="w-full"
                      size="sm"
                    >
                      <Copy className="w-4 h-4 mr-2" />
                      复制错误详情
                    </Button>
                  )}
                </div>

                {process.env.NODE_ENV === 'development' && (
                  <div className="mt-4 p-3 bg-gray-100 rounded text-xs">
                    <Badge variant="outline" className="mb-2">
                      错误ID: {errorId}
                    </Badge>
                    <details className="mt-2">
                      <summary className="cursor-pointer text-sm font-medium">
                        技术详情
                      </summary>
                      <pre className="mt-2 text-xs overflow-auto">
                        {error?.stack}
                      </pre>
                    </details>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        )
      }

      if (level === 'page') {
        return (
          <div className="container mx-auto px-4 py-12">
            <Card className="max-w-lg mx-auto">
              <CardHeader className="text-center">
                <AlertTriangle className="w-12 h-12 text-yellow-600 mx-auto mb-4" />
                <CardTitle>页面加载失败</CardTitle>
                <CardDescription>
                  这个页面遇到了错误，请尝试刷新或返回
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <Button
                  onClick={this.retryWithDelay}
                  className="w-full"
                  size="sm"
                >
                  <RotateCcw className="w-4 h-4 mr-2" />
                  重试
                </Button>

                <Button
                  asChild
                  variant="outline"
                  className="w-full"
                  size="sm"
                >
                  <Link href="/">
                    <Home className="w-4 h-4 mr-2" />
                    返回首页
                  </Link>
                </Button>
              </CardContent>
            </Card>
          </div>
        )
      }

      // Component level error
      return (
        <Alert variant="destructive" className="my-4">
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription className="flex items-center justify-between">
            <span>组件加载失败</span>
            <Button
              onClick={this.resetError}
              variant="ghost"
              size="sm"
              className="ml-2"
            >
              <RotateCcw className="w-3 h-3 mr-1" />
              重试
            </Button>
          </AlertDescription>
        </Alert>
      )
    }

    return children
  }
}

// Higher-order component for wrapping components with error boundaries
export function withErrorBoundary<T extends {}>(
  WrappedComponent: React.ComponentType<T>,
  options: Omit<ErrorBoundaryProps, 'children'> = {}
) {
  const WithErrorBoundaryComponent = (props: T) => (
    <GlobalErrorBoundary {...options}>
      <WrappedComponent {...props} />
    </GlobalErrorBoundary>
  )

  WithErrorBoundaryComponent.displayName = `withErrorBoundary(${WrappedComponent.displayName || WrappedComponent.name})`

  return WithErrorBoundaryComponent
}

// Hook for manually reporting errors
export function useErrorReporting() {
  const reportError = (error: Error, context?: string) => {
    const errorReport = {
      id: `manual-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      message: error.message,
      stack: error.stack,
      context,
      url: window.location.href,
      timestamp: new Date().toISOString()
    }

    console.error('Manual Error Report:', errorReport)

    // Report to external service in production
    if (process.env.NODE_ENV === 'production') {
      // fetch('/api/errors', { method: 'POST', body: JSON.stringify(errorReport) })
    }
  }

  return { reportError }
}

export default GlobalErrorBoundary