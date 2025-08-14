/**
 * Enhanced Error Boundary - SOTA Implementation
 * 增强的错误边界 - 支持错误恢复、智能重试、用户反馈收集
 */

'use client'

import React, { Component, ReactNode, ErrorInfo } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Textarea } from '@/components/ui/textarea'
import { 
  AlertTriangle, 
  RefreshCw, 
  Bug, 
  MessageSquare, 
  Home,
  ChevronDown,
  ChevronUp
} from 'lucide-react'
import { ErrorHandler, ErrorContext } from '@/lib/utils/error-handler'

interface Props {
  children: ReactNode
  fallback?: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
  enableRecovery?: boolean
  enableFeedback?: boolean
  level?: 'page' | 'component' | 'feature'
  name?: string
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
  retryCount: number
  showDetails: boolean
  feedbackText: string
  feedbackSent: boolean
  isRecovering: boolean
}

const MAX_RETRY_COUNT = 3
const RECOVERY_DELAY = 1000

/**
 * Enhanced Error Boundary with recovery capabilities
 */
export class EnhancedErrorBoundary extends Component<Props, State> {
  private errorHandler: ErrorHandler
  private recoveryTimer: NodeJS.Timeout | null = null

  constructor(props: Props) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      retryCount: 0,
      showDetails: false,
      feedbackText: '',
      feedbackSent: false,
      isRecovering: false
    }
    
    this.errorHandler = new ErrorHandler()
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return {
      hasError: true,
      error
    }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({ errorInfo })

    // Report error using our error handler
    const context: ErrorContext = {
      component: this.props.name || 'EnhancedErrorBoundary',
      level: this.props.level || 'component',
      timestamp: new Date().toISOString(),
      userAgent: navigator.userAgent,
      url: window.location.href,
      userId: this.getCurrentUserId(),
      stackTrace: errorInfo.componentStack || undefined
    }

    this.errorHandler.handleError(error, context)

    // Call custom error handler if provided
    if (this.props.onError) {
      this.props.onError(error, errorInfo)
    }

    // Auto-recovery for component-level errors
    if (this.props.enableRecovery && this.props.level === 'component') {
      this.scheduleRecovery()
    }
  }

  private getCurrentUserId(): string | undefined {
    try {
      const userStr = localStorage.getItem('user')
      return userStr ? JSON.parse(userStr).id : undefined
    } catch {
      return undefined
    }
  }

  private scheduleRecovery = () => {
    if (this.state.retryCount < MAX_RETRY_COUNT) {
      this.setState({ isRecovering: true })
      
      this.recoveryTimer = setTimeout(() => {
        this.setState(prevState => ({
          hasError: false,
          error: null,
          errorInfo: null,
          retryCount: prevState.retryCount + 1,
          isRecovering: false
        }))
      }, RECOVERY_DELAY * (this.state.retryCount + 1))
    }
  }

  private handleRetry = () => {
    if (this.state.retryCount < MAX_RETRY_COUNT) {
      this.setState({
        hasError: false,
        error: null,
        errorInfo: null,
        retryCount: this.state.retryCount + 1
      })
    } else {
      // Force page reload as last resort
      window.location.reload()
    }
  }

  private handleGoHome = () => {
    window.location.href = '/'
  }

  private handleToggleDetails = () => {
    this.setState(prevState => ({
      showDetails: !prevState.showDetails
    }))
  }

  private handleFeedbackSubmit = async () => {
    const { error, errorInfo, feedbackText } = this.state
    
    try {
      // Submit feedback to error tracking service
      await this.errorHandler.submitFeedback({
        error: error?.message || 'Unknown error',
        stack: error?.stack,
        componentStack: errorInfo?.componentStack || undefined,
        feedback: feedbackText,
        url: window.location.href,
        timestamp: new Date().toISOString(),
        userId: this.getCurrentUserId()
      })

      this.setState({ feedbackSent: true })
    } catch (err) {
      console.error('Failed to submit feedback:', err)
    }
  }

  componentWillUnmount() {
    if (this.recoveryTimer) {
      clearTimeout(this.recoveryTimer)
    }
  }

  private renderErrorUI() {
    const { error, errorInfo, showDetails, retryCount, feedbackText, feedbackSent, isRecovering } = this.state
    const { level = 'component', enableRecovery = false, enableFeedback = false } = this.props
    
    const canRetry = retryCount < MAX_RETRY_COUNT
    const severity = this.getErrorSeverity(error)

    return (
      <Card className={`w-full max-w-2xl mx-auto ${this.getSeverityStyles(severity)}`}>
        <CardHeader>
          <div className="flex items-center gap-3">
            <AlertTriangle className={`h-6 w-6 ${this.getSeverityIconColor(severity)}`} />
            <div>
              <CardTitle className="text-lg">
                {this.getErrorTitle(severity, level)}
              </CardTitle>
              <CardDescription>
                {this.getErrorDescription(severity, level)}
              </CardDescription>
            </div>
          </div>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* Error message */}
          <Alert>
            <AlertDescription className="font-mono text-sm">
              {error?.message || 'An unknown error occurred'}
            </AlertDescription>
          </Alert>

          {/* Recovery indicator */}
          {isRecovering && (
            <Alert>
              <RefreshCw className="h-4 w-4 animate-spin" />
              <AlertDescription>
                Attempting automatic recovery... ({MAX_RETRY_COUNT - retryCount} attempts remaining)
              </AlertDescription>
            </Alert>
          )}

          {/* Action buttons */}
          <div className="flex flex-wrap gap-2">
            {enableRecovery && canRetry && !isRecovering && (
              <Button onClick={this.handleRetry} variant="default">
                <RefreshCw className="h-4 w-4 mr-2" />
                Retry {retryCount > 0 && `(${retryCount}/${MAX_RETRY_COUNT})`}
              </Button>
            )}

            {(!canRetry || !enableRecovery) && (
              <Button onClick={this.handleRetry} variant="default">
                <RefreshCw className="h-4 w-4 mr-2" />
                Refresh
              </Button>
            )}

            {level === 'page' && (
              <Button onClick={this.handleGoHome} variant="outline">
                <Home className="h-4 w-4 mr-2" />
                Go Home
              </Button>
            )}

            <Button 
              onClick={this.handleToggleDetails} 
              variant="ghost"
              size="sm"
            >
              <Bug className="h-4 w-4 mr-2" />
              {showDetails ? 'Hide' : 'Show'} Details
              {showDetails ? <ChevronUp className="h-4 w-4 ml-1" /> : <ChevronDown className="h-4 w-4 ml-1" />}
            </Button>
          </div>

          {/* Error details */}
          {showDetails && (
            <div className="space-y-3">
              <div className="bg-gray-50 p-3 rounded-md">
                <h4 className="font-semibold text-sm mb-2">Error Stack:</h4>
                <pre className="text-xs text-gray-700 whitespace-pre-wrap overflow-auto max-h-40">
                  {error?.stack || 'No stack trace available'}
                </pre>
              </div>

              {errorInfo?.componentStack && (
                <div className="bg-blue-50 p-3 rounded-md">
                  <h4 className="font-semibold text-sm mb-2">Component Stack:</h4>
                  <pre className="text-xs text-gray-700 whitespace-pre-wrap overflow-auto max-h-40">
                    {errorInfo.componentStack}
                  </pre>
                </div>
              )}
            </div>
          )}

          {/* Feedback section */}
          {enableFeedback && !feedbackSent && (
            <div className="border-t pt-4">
              <h4 className="font-semibold text-sm mb-2 flex items-center">
                <MessageSquare className="h-4 w-4 mr-2" />
                Help us improve
              </h4>
              <Textarea
                placeholder="What were you doing when this error occurred? Any additional context would be helpful..."
                value={feedbackText}
                onChange={(e) => this.setState({ feedbackText: e.target.value })}
                className="mb-2"
                rows={3}
              />
              <Button 
                onClick={this.handleFeedbackSubmit}
                disabled={!feedbackText.trim()}
                size="sm"
              >
                Send Feedback
              </Button>
            </div>
          )}

          {feedbackSent && (
            <Alert>
              <MessageSquare className="h-4 w-4" />
              <AlertDescription>
                Thank you for your feedback! Our team will review it.
              </AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>
    )
  }

  private getErrorSeverity(error: Error | null): 'low' | 'medium' | 'high' | 'critical' {
    if (!error) return 'low'
    
    const message = error.message.toLowerCase()
    
    if (message.includes('chunk') || message.includes('loading')) return 'low'
    if (message.includes('network') || message.includes('fetch')) return 'medium'  
    if (message.includes('reference') || message.includes('null')) return 'high'
    if (message.includes('security') || message.includes('permission')) return 'critical'
    
    return 'medium'
  }

  private getSeverityStyles(severity: string): string {
    switch (severity) {
      case 'low': return 'border-yellow-200 bg-yellow-50'
      case 'medium': return 'border-orange-200 bg-orange-50'
      case 'high': return 'border-red-200 bg-red-50'
      case 'critical': return 'border-red-500 bg-red-100'
      default: return 'border-gray-200'
    }
  }

  private getSeverityIconColor(severity: string): string {
    switch (severity) {
      case 'low': return 'text-yellow-600'
      case 'medium': return 'text-orange-600'
      case 'high': return 'text-red-600'
      case 'critical': return 'text-red-700'
      default: return 'text-gray-600'
    }
  }

  private getErrorTitle(severity: string, level: string): string {
    const levelText = level === 'page' ? 'Page' : level === 'feature' ? 'Feature' : 'Component'
    
    switch (severity) {
      case 'low': return `${levelText} temporarily unavailable`
      case 'medium': return `${levelText} error occurred`
      case 'high': return `Critical ${levelText.toLowerCase()} error`
      case 'critical': return `System error detected`
      default: return `${levelText} error`
    }
  }

  private getErrorDescription(severity: string, level: string): string {
    switch (severity) {
      case 'low': return 'This appears to be a temporary issue. Please try again.'
      case 'medium': return 'Something went wrong. We\'re working to fix this issue.'
      case 'high': return 'A serious error has occurred. Our team has been notified.'
      case 'critical': return 'A critical system error has been detected. Please contact support if this persists.'
      default: return 'An unexpected error occurred. Please try refreshing the page.'
    }
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }
      
      return (
        <div className="min-h-[400px] flex items-center justify-center p-4">
          {this.renderErrorUI()}
        </div>
      )
    }

    return this.props.children
  }
}

/**
 * HOC for wrapping components with enhanced error boundary
 */
export function withEnhancedErrorBoundary<P extends object>(
  WrappedComponent: React.ComponentType<P>,
  options: Omit<Props, 'children'> = {}
) {
  const WithErrorBoundary = (props: P) => (
    <EnhancedErrorBoundary {...options}>
      <WrappedComponent {...props} />
    </EnhancedErrorBoundary>
  )

  WithErrorBoundary.displayName = `withEnhancedErrorBoundary(${WrappedComponent.displayName || WrappedComponent.name})`
  
  return WithErrorBoundary
}

/**
 * Hook for manually triggering error boundary
 */
export function useErrorBoundary() {
  const throwError = React.useCallback((error: Error) => {
    throw error
  }, [])

  return { throwError }
}