'use client'

import * as React from 'react'
import { ErrorBoundary, ErrorBoundaryPropsWithFallback } from 'react-error-boundary'
import { AlertCircle, RefreshCw, Home } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { useRouter } from 'next/navigation'

/**
 * 默认错误回退组件
 */
function DefaultErrorFallback({ error, resetErrorBoundary }: { error: Error; resetErrorBoundary: () => void }) {
  const router = useRouter()
  
  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="max-w-md w-full">
        <CardHeader>
          <div className="flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-destructive" />
            <CardTitle>出错了</CardTitle>
          </div>
          <CardDescription>
            抱歉，页面遇到了一些问题。请尝试刷新页面或返回首页。
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg bg-muted p-4 font-mono text-sm">
            <p className="font-semibold text-destructive">错误信息：</p>
            <p className="mt-1 text-muted-foreground">{error.message}</p>
            {process.env.NODE_ENV === 'development' && error.stack && (
              <details className="mt-2">
                <summary className="cursor-pointer text-xs text-muted-foreground hover:text-foreground">
                  查看详细堆栈信息
                </summary>
                <pre className="mt-2 whitespace-pre-wrap text-xs text-muted-foreground">
                  {error.stack}
                </pre>
              </details>
            )}
          </div>
        </CardContent>
        <CardFooter className="flex gap-2">
          <Button onClick={resetErrorBoundary} variant="default" className="flex-1">
            <RefreshCw className="mr-2 h-4 w-4" />
            重试
          </Button>
          <Button onClick={() => router.push('/')} variant="outline" className="flex-1">
            <Home className="mr-2 h-4 w-4" />
            返回首页
          </Button>
        </CardFooter>
      </Card>
    </div>
  )
}

/**
 * 错误日志记录
 */
function logError(error: Error, errorInfo: any) {
  // 在生产环境中，这里应该发送到错误监控服务
  console.error('Error caught by error boundary:', error)
  console.error('Error info:', errorInfo)
  
  // TODO: 集成错误监控服务（如Sentry）
  // if (process.env.NODE_ENV === 'production') {
  //   Sentry.captureException(error, {
  //     contexts: {
  //       react: {
  //         componentStack: errorInfo.componentStack,
  //       },
  //     },
  //   })
  // }
}

/**
 * 高阶组件：为组件添加错误边界
 * 遵循CLAUDE.md原则：持续优化用户体验
 */
export function withErrorBoundary<P extends object>(
  Component: React.ComponentType<P>,
  errorBoundaryProps?: {
    onError?: (error: Error, info: any) => void
    onReset?: () => void
  }
) {
  const WrappedComponent = React.forwardRef<any, P>((props, ref) => {
    return (
      <ErrorBoundary
        FallbackComponent={DefaultErrorFallback}
        onError={errorBoundaryProps?.onError || logError}
        onReset={errorBoundaryProps?.onReset}
      >
        <Component {...(props as P)} ref={ref} />
      </ErrorBoundary>
    )
  })

  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name || 'Component'})`

  return WrappedComponent
}

/**
 * Hook：在函数组件中使用错误边界
 */
export function useErrorHandler() {
  const [error, setError] = React.useState<Error | null>(null)

  React.useEffect(() => {
    if (error) {
      throw error
    }
  }, [error])

  const resetError = React.useCallback(() => {
    setError(null)
  }, [])

  const captureError = React.useCallback((error: Error) => {
    setError(error)
  }, [])

  return { captureError, resetError }
}

/**
 * 使用示例：
 * 
 * // 1. 使用HOC
 * const SafeComponent = withErrorBoundary(MyComponent)
 * 
 * // 2. 自定义错误处理
 * const SafeComponent = withErrorBoundary(MyComponent, {
 *   FallbackComponent: CustomErrorFallback,
 *   onError: (error, info) => {
 *     console.log('Custom error handling', error)
 *   },
 *   onReset: () => {
 *     console.log('Error boundary reset')
 *   }
 * })
 * 
 * // 3. 在函数组件中使用
 * function MyComponent() {
 *   const { captureError } = useErrorHandler()
 *   
 *   const handleAsyncError = async () => {
 *     try {
 *       await riskyAsyncOperation()
 *     } catch (error) {
 *       captureError(error as Error)
 *     }
 *   }
 * }
 */