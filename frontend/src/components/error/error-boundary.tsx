'use client'

import React, { Component, ErrorInfo, ReactNode } from 'react'
import { AlertTriangle, RefreshCw, Home, Bug } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface Props {
  children: ReactNode
  fallback?: React.ComponentType<ErrorFallbackProps>
  onError?: (error: Error, errorInfo: ErrorInfo) => void
  resetKeys?: Array<string | number>
  resetOnPropsChange?: boolean
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
  errorId: string
}

export interface ErrorFallbackProps {
  error: Error | null
  errorInfo: ErrorInfo | null
  resetErrorBoundary: () => void
  errorId: string
}

export class ErrorBoundary extends Component<Props, State> {
  private resetTimeoutId: number | null = null

  constructor(props: Props) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: ''
    }
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    const errorId = `error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    return {
      hasError: true,
      error,
      errorId
    }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({
      error,
      errorInfo
    })

    // Call custom error handler
    if (this.props.onError) {
      this.props.onError(error, errorInfo)
    }

    // Log error to console in development
    if (process.env.NODE_ENV === 'development') {
      console.group('🚨 Error Boundary Caught an Error')
      console.error('Error:', error)
      console.error('Error Info:', errorInfo)
      console.error('Component Stack:', errorInfo.componentStack)
      console.groupEnd()
    }

    // Report error to monitoring service
    this.reportError(error, errorInfo)
  }

  componentDidUpdate(prevProps: Props) {
    const { resetKeys, resetOnPropsChange } = this.props
    const { hasError } = this.state

    if (hasError && prevProps.resetKeys !== resetKeys) {
      if (resetKeys?.some((resetKey, idx) => prevProps.resetKeys?.[idx] !== resetKey)) {
        this.resetErrorBoundary()
      }
    }

    if (hasError && resetOnPropsChange && prevProps.children !== this.props.children) {
      this.resetErrorBoundary()
    }
  }

  resetErrorBoundary = () => {
    if (this.resetTimeoutId) {
      clearTimeout(this.resetTimeoutId)
    }

    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: ''
    })
  }

  reportError = async (error: Error, errorInfo: ErrorInfo) => {
    try {
      // Report to error tracking service
      if (typeof window !== 'undefined') {
        // Example: Sentry, LogRocket, or custom analytics
        const errorReport = {
          message: error.message,
          stack: error.stack,
          componentStack: errorInfo.componentStack,
          url: window.location.href,
          userAgent: navigator.userAgent,
          timestamp: new Date().toISOString(),
          errorId: this.state.errorId
        }

        // Send to monitoring endpoint
        fetch('/api/errors/report', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(errorReport)
        }).catch(() => {
          // Fail silently if error reporting fails
        })
      }
    } catch (reportingError) {
      console.warn('Failed to report error:', reportingError)
    }
  }

  render() {
    if (this.state.hasError) {
      const FallbackComponent = this.props.fallback || DefaultErrorFallback
      
      return (
        <FallbackComponent
          error={this.state.error}
          errorInfo={this.state.errorInfo}
          resetErrorBoundary={this.resetErrorBoundary}
          errorId={this.state.errorId}
        />
      )
    }

    return this.props.children
  }
}

// Default error fallback component
function DefaultErrorFallback({ 
  error, 
  errorInfo, 
  resetErrorBoundary, 
  errorId 
}: ErrorFallbackProps) {
  const [showDetails, setShowDetails] = React.useState(false)
  const [copied, setCopied] = React.useState(false)

  const copyErrorDetails = async () => {
    const errorDetails = `
Error ID: ${errorId}
Error: ${error?.message || 'Unknown error'}
Stack: ${error?.stack || 'No stack trace'}
Component Stack: ${errorInfo?.componentStack || 'No component stack'}
URL: ${typeof window !== 'undefined' ? window.location.href : 'Unknown'}
Time: ${new Date().toISOString()}
    `.trim()

    try {
      await navigator.clipboard.writeText(errorDetails)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // Fallback for older browsers
      const textArea = document.createElement('textarea')
      textArea.value = errorDetails
      document.body.appendChild(textArea)
      textArea.select()
      document.execCommand('copy')
      document.body.removeChild(textArea)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className="min-h-[400px] flex items-center justify-center p-4">
      <Card className="w-full max-w-lg border-red-200 bg-red-50">
        <CardHeader className="text-center">
          <div className="mx-auto w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mb-4">
            <AlertTriangle className="w-8 h-8 text-red-600" />
          </div>
          <CardTitle className="text-red-900">出现了错误</CardTitle>
          <CardDescription className="text-red-700">
            很抱歉，应用遇到了意外错误。我们已经记录了这个问题。
          </CardDescription>
        </CardHeader>
        
        <CardContent className="space-y-4">
          {/* Error message */}
          <div className="p-3 bg-red-100 border border-red-200 rounded-md">
            <p className="text-sm text-red-800 font-mono">
              {error?.message || '未知错误'}
            </p>
          </div>

          {/* Error ID */}
          <div className="text-xs text-red-600 text-center">
            错误ID: {errorId}
          </div>

          {/* Action buttons */}
          <div className="flex flex-col sm:flex-row gap-2">
            <Button 
              onClick={resetErrorBoundary}
              className="flex-1 bg-red-600 hover:bg-red-700 text-white"
            >
              <RefreshCw className="w-4 h-4 mr-2" />
              重试
            </Button>
            
            <Button 
              onClick={() => window.location.href = '/'}
              variant="outline"
              className="flex-1 border-red-300 text-red-700 hover:bg-red-50"
            >
              <Home className="w-4 h-4 mr-2" />
              返回首页
            </Button>
          </div>

          {/* Details toggle */}
          <div className="border-t border-red-200 pt-4">
            <Button
              onClick={() => setShowDetails(!showDetails)}
              variant="ghost"
              size="sm"
              className="w-full text-red-600 hover:text-red-700 hover:bg-red-100"
            >
              <Bug className="w-4 h-4 mr-2" />
              {showDetails ? '隐藏' : '显示'}技术详情
            </Button>

            {showDetails && (
              <div className="mt-3 space-y-3">
                <div className="p-3 bg-gray-100 border rounded-md max-h-32 overflow-auto">
                  <pre className="text-xs text-gray-700 whitespace-pre-wrap">
                    {error?.stack || '无堆栈信息'}
                  </pre>
                </div>
                
                {errorInfo?.componentStack && (
                  <div className="p-3 bg-gray-100 border rounded-md max-h-32 overflow-auto">
                    <h4 className="text-xs font-semibold text-gray-700 mb-1">组件堆栈:</h4>
                    <pre className="text-xs text-gray-600 whitespace-pre-wrap">
                      {errorInfo.componentStack}
                    </pre>
                  </div>
                )}

                <Button
                  onClick={copyErrorDetails}
                  variant="outline"
                  size="sm"
                  className="w-full"
                  disabled={copied}
                >
                  {copied ? '已复制!' : '复制错误详情'}
                </Button>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

// Specific error fallbacks for different scenarios
export function NetworkErrorFallback({ resetErrorBoundary }: { resetErrorBoundary: () => void }) {
  return (
    <div className="text-center py-8">
      <div className="w-16 h-16 bg-orange-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <AlertTriangle className="w-8 h-8 text-orange-600" />
      </div>
      <h3 className="text-lg font-semibold text-gray-900 mb-2">网络连接问题</h3>
      <p className="text-gray-600 mb-4">请检查您的网络连接并重试</p>
      <Button onClick={resetErrorBoundary} className="bg-orange-600 hover:bg-orange-700">
        <RefreshCw className="w-4 h-4 mr-2" />
        重新加载
      </Button>
    </div>
  )
}

export function DataLoadErrorFallback({ resetErrorBoundary }: { resetErrorBoundary: () => void }) {
  return (
    <div className="text-center py-8">
      <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <AlertTriangle className="w-8 h-8 text-blue-600" />
      </div>
      <h3 className="text-lg font-semibold text-gray-900 mb-2">数据加载失败</h3>
      <p className="text-gray-600 mb-4">无法加载所需数据，请重试</p>
      <Button onClick={resetErrorBoundary} className="bg-blue-600 hover:bg-blue-700">
        <RefreshCw className="w-4 h-4 mr-2" />
        重新加载
      </Button>
    </div>
  )
}

// Error boundary hook for functional components
export function useErrorHandler() {
  return (error: Error, errorInfo?: ErrorInfo) => {
    console.error('Error caught by error handler:', error)
    
    // Report error
    if (typeof window !== 'undefined') {
      const errorReport = {
        message: error.message,
        stack: error.stack,
        url: window.location.href,
        timestamp: new Date().toISOString()
      }

      fetch('/api/errors/report', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(errorReport)
      }).catch(() => {
        // Fail silently
      })
    }
  }
}

// HOC for wrapping components with error boundary
export function withErrorBoundary<P extends object>(
  Component: React.ComponentType<P>,
  errorFallback?: React.ComponentType<ErrorFallbackProps>
) {
  const WrappedComponent = (props: P) => (
    <ErrorBoundary fallback={errorFallback}>
      <Component {...props} />
    </ErrorBoundary>
  )
  
  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`
  
  return WrappedComponent
}