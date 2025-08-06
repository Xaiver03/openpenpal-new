'use client'

import React, { Component, ErrorInfo, ReactNode } from 'react'
import { AlertTriangle, RefreshCw, Bug, Home } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { log } from '@/utils/logger'

interface Props {
  children: ReactNode
  fallback?: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
  showDetails?: boolean
  level?: 'page' | 'component' | 'feature'
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
  errorId: string
}

/**
 * 通用错误边界组件
 * Generic Error Boundary Component
 * 
 * 功能：
 * - 捕获子组件中的 JavaScript 错误
 * - 显示友好的错误UI
 * - 记录错误日志
 * - 提供重试机制
 */
export class ErrorBoundary extends Component<Props, State> {
  private errorId: string

  constructor(props: Props) {
    super(props)
    
    this.errorId = `error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: this.errorId
    }
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    // 更新 state 使下一次渲染能够显示降级后的 UI
    return {
      hasError: true,
      error
    }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // 记录错误日志
    log.error('Error Boundary caught an error', {
      error: {
        name: error.name,
        message: error.message,
        stack: error.stack
      },
      errorInfo: {
        componentStack: errorInfo.componentStack
      },
      errorId: this.errorId,
      level: this.props.level || 'component',
      userAgent: navigator.userAgent,
      url: window.location.href,
      timestamp: new Date().toISOString()
    }, 'ErrorBoundary')

    // 更新状态
    this.setState({
      error,
      errorInfo
    })

    // 调用自定义错误处理函数
    if (this.props.onError) {
      this.props.onError(error, errorInfo)
    }

    // 发送错误报告到监控服务（如果有的话）
    if (typeof window !== 'undefined' && (window as any).gtag) {
      (window as any).gtag('event', 'exception', {
        description: error.message,
        fatal: false,
        error_id: this.errorId
      })
    }
  }

  handleRetry = () => {
    // 重置错误状态，重新渲染子组件
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: `error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    })
  }

  handleGoHome = () => {
    window.location.href = '/'
  }

  handleReload = () => {
    window.location.reload()
  }

  render() {
    if (this.state.hasError) {
      // 如果提供了自定义的降级UI，使用它
      if (this.props.fallback) {
        return this.props.fallback
      }

      // 根据错误边界级别显示不同的UI
      const level = this.props.level || 'component'
      
      return (
        <div className={`
          flex items-center justify-center p-4
          ${level === 'page' ? 'min-h-screen bg-gray-50' : ''}
          ${level === 'feature' ? 'min-h-[400px]' : ''}
          ${level === 'component' ? 'min-h-[200px]' : ''}
        `}>
          <Card className={`
            w-full max-w-lg border-red-200 
            ${level === 'page' ? 'shadow-lg' : 'shadow-sm'}
          `}>
            <CardHeader className="text-center">
              <div className="flex justify-center mb-4">
                <div className="p-3 bg-red-100 rounded-full">
                  <AlertTriangle className="w-8 h-8 text-red-600" />
                </div>
              </div>
              <CardTitle className="text-xl font-semibold text-gray-900">
                {level === 'page' ? '页面加载出错' : '组件渲染出错'}
              </CardTitle>
              <CardDescription className="text-gray-600">
                {level === 'page' 
                  ? '很抱歉，页面遇到一些问题无法正常显示'
                  : '此功能模块暂时无法正常工作，请稍后重试'
                }
              </CardDescription>
            </CardHeader>

            <CardContent className="space-y-4">
              {/* 错误信息 */}
              {this.props.showDetails && this.state.error && (
                <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                  <div className="flex items-start space-x-2">
                    <Bug className="w-4 h-4 text-red-500 mt-0.5 flex-shrink-0" />
                    <div className="space-y-1 text-sm">
                      <p className="font-medium text-red-800">
                        错误详情:
                      </p>
                      <p className="text-red-700 font-mono text-xs break-all">
                        {this.state.error.message}
                      </p>
                      <p className="text-red-600 text-xs">
                        错误ID: {this.state.errorId}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* 操作按钮 */}
              <div className="flex flex-col sm:flex-row gap-2 pt-2">
                <Button 
                  variant="outline" 
                  onClick={this.handleRetry}
                  className="flex-1"
                >
                  <RefreshCw className="w-4 h-4 mr-2" />
                  重新加载
                </Button>
                
                {level === 'page' && (
                  <>
                    <Button 
                      variant="outline" 
                      onClick={this.handleGoHome}
                      className="flex-1"
                    >
                      <Home className="w-4 h-4 mr-2" />
                      返回首页
                    </Button>
                    
                    <Button 
                      variant="outline" 
                      onClick={this.handleReload}
                      className="flex-1"
                    >
                      <RefreshCw className="w-4 h-4 mr-2" />
                      刷新页面
                    </Button>
                  </>
                )}
              </div>

              {/* 帮助信息 */}
              <div className="text-xs text-gray-500 text-center pt-2 border-t">
                如果问题持续出现，请联系技术支持或稍后再试
              </div>
            </CardContent>
          </Card>
        </div>
      )
    }

    return this.props.children
  }
}

/**
 * 页面级错误边界组件
 * Page-level Error Boundary Component
 */
export function PageErrorBoundary({ 
  children, 
  onError 
}: { 
  children: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void 
}) {
  return (
    <ErrorBoundary 
      level="page" 
      showDetails={process.env.NODE_ENV === 'development'}
      onError={onError}
    >
      {children}
    </ErrorBoundary>
  )
}

/**
 * 功能级错误边界组件
 * Feature-level Error Boundary Component
 */
export function FeatureErrorBoundary({ 
  children, 
  onError,
  fallback 
}: { 
  children: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
  fallback?: ReactNode
}) {
  return (
    <ErrorBoundary 
      level="feature" 
      showDetails={process.env.NODE_ENV === 'development'}
      onError={onError}
      fallback={fallback}
    >
      {children}
    </ErrorBoundary>
  )
}

/**
 * 组件级错误边界组件
 * Component-level Error Boundary Component
 */
export function ComponentErrorBoundary({ 
  children, 
  onError,
  fallback 
}: { 
  children: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
  fallback?: ReactNode
}) {
  return (
    <ErrorBoundary 
      level="component" 
      showDetails={process.env.NODE_ENV === 'development'}
      onError={onError}
      fallback={fallback}
    >
      {children}
    </ErrorBoundary>
  )
}

/**
 * WebSocket专用错误边界 (保持向后兼容)
 * WebSocket-specific Error Boundary (backward compatibility)
 */
export class WebSocketErrorBoundary extends Component<{
  children: ReactNode
  fallback?: ReactNode
}, State> {
  constructor(props: { children: ReactNode; fallback?: ReactNode }) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: `ws_error_${Date.now()}`
    }
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    log.error('WebSocket Error Boundary caught an error', {
      error,
      errorInfo
    }, 'WebSocketErrorBoundary')
  }

  render() {
    if (this.state.hasError) {
      return (
        this.props.fallback || (
          <div className="p-4 text-amber-700 bg-amber-50 rounded-md">
            <p>实时连接功能暂时不可用</p>
            <button 
              onClick={() => this.setState({ hasError: false, error: null, errorInfo: null, errorId: `ws_error_${Date.now()}` })}
              className="mt-2 px-4 py-2 bg-amber-600 text-white rounded hover:bg-amber-700 transition-colors"
            >
              重试
            </button>
          </div>
        )
      )
    }

    return this.props.children
  }
}

/**
 * 错误边界 Hook
 * Error Boundary Hook for functional components
 */
export function useErrorHandler() {
  const handleError = (error: Error, errorBoundary?: string) => {
    log.error('Manual error reported', {
      error: {
        name: error.name,
        message: error.message,
        stack: error.stack
      },
      errorBoundary,
      userAgent: navigator.userAgent,
      url: window.location.href,
      timestamp: new Date().toISOString()
    }, 'useErrorHandler')

    // 在开发环境下抛出错误以便错误边界捕获
    if (process.env.NODE_ENV === 'development') {
      throw error
    }
  }

  return { handleError }
}

export default ErrorBoundary