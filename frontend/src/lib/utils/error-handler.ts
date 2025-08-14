/**
 * SOTA Error Handling System
 * 
 * Provides comprehensive error handling with categorization, 
 * contextual logging, user-friendly messages, and recovery strategies.
 */

import { ApiError } from '../api-client'

// ================================
// Error Types and Interfaces
// ================================

export interface ErrorContext {
  component?: string
  action?: string
  level?: 'page' | 'component' | 'feature'
  userId?: string
  requestId?: string
  timestamp?: string
  userAgent?: string
  url?: string
  stackTrace?: string
  additionalData?: Record<string, any>
}

export interface ErrorReport {
  id: string
  type: ErrorType
  category: ErrorCategory
  severity: ErrorSeverity
  message: string
  userMessage: string
  stack?: string
  context: ErrorContext
  timestamp: string
  canRetry: boolean
  recoveryActions: string[]
}

export type ErrorType = 
  | 'API_ERROR'
  | 'NETWORK_ERROR' 
  | 'VALIDATION_ERROR'
  | 'AUTH_ERROR'
  | 'PERMISSION_ERROR'
  | 'BUSINESS_LOGIC_ERROR'
  | 'SYSTEM_ERROR'
  | 'UI_ERROR'

export type ErrorCategory = 
  | 'CRITICAL'
  | 'HIGH'
  | 'MEDIUM'
  | 'LOW'
  | 'INFO'

export type ErrorSeverity = 
  | 'FATAL'      // App cannot continue
  | 'ERROR'      // Feature broken, but app continues
  | 'WARNING'    // Potential issue, degraded experience
  | 'INFO'       // Informational, no impact

export interface RecoveryStrategy {
  type: 'retry' | 'redirect' | 'fallback' | 'ignore' | 'escalate'
  params?: Record<string, any>
  message?: string
}

// ================================
// Enhanced Error Handler
// ================================

export class ErrorHandler {
  private static instance: ErrorHandler
  private errorReports: Map<string, ErrorReport> = new Map()
  
  static getInstance(): ErrorHandler {
    if (!ErrorHandler.instance) {
      ErrorHandler.instance = new ErrorHandler()
    }
    return ErrorHandler.instance
  }
  
  /**
   * Handle error with comprehensive logging and user feedback
   */
  handleError(
    error: Error | ApiError | unknown,
    context: ErrorContext = {}
  ): ErrorReport {
    const errorId = this.generateErrorId()
    const timestamp = new Date().toISOString()
    const message = this.extractErrorMessage(error)
    
    // Create simplified report for now
    const report: ErrorReport = {
      id: errorId,
      type: 'SYSTEM_ERROR',
      category: 'MEDIUM',
      severity: 'ERROR',
      message,
      userMessage: '系统出现异常，请稍后重试',
      context: this.enrichContext(context),
      timestamp,
      canRetry: false,
      recoveryActions: ['retry']
    }
    
    this.storeErrorReport(report)
    this.logError(report)
    
    return report
  }
  
  private generateErrorId(): string {
    return `err_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }
  
  private extractErrorMessage(error: unknown): string {
    if (!error) return 'Unknown error occurred'
    if (typeof error === 'string') return error
    if (error instanceof Error) return error.message || 'Error occurred'
    return 'Unexpected error occurred'
  }
  
  private enrichContext(context: ErrorContext): ErrorContext {
    return {
      ...context,
      timestamp: new Date().toISOString(),
      userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : 'Unknown',
      url: typeof window !== 'undefined' ? window.location.href : 'Unknown'
    }
  }
  
  private storeErrorReport(report: ErrorReport): void {
    this.errorReports.set(report.id, report)
    
    // Limit stored reports
    if (this.errorReports.size > 100) {
      const firstKey = this.errorReports.keys().next().value
      if (firstKey) {
        this.errorReports.delete(firstKey)
      }
    }
  }
  
  private logError(report: ErrorReport): void {
    const logData = {
      id: report.id,
      type: report.type,
      severity: report.severity,
      message: report.message,
      context: report.context
    }
    
    switch (report.severity) {
      case 'FATAL':
        console.error('🔥 FATAL ERROR:', logData)
        break
      case 'ERROR':
        console.error('❌ ERROR:', logData)
        break
      case 'WARNING':
        console.warn('⚠️ WARNING:', logData)
        break
      case 'INFO':
        console.info('ℹ️ INFO:', logData)
        break
    }
  }
  
  /**
   * Get all error reports
   */
  getErrorReports(): ErrorReport[] {
    return Array.from(this.errorReports.values())
  }
  
  /**
   * Clear all error reports
   */
  clearErrors(): void {
    this.errorReports.clear()
  }
  
  /**
   * Submit feedback for error reports
   */
  async submitFeedback(feedbackData: {
    error: string
    stack?: string
    componentStack?: string
    feedback: string
    url: string
    timestamp: string
    userId?: string
  }): Promise<void> {
    try {
      // For now, just log the feedback (could be sent to analytics service)
      console.log('📧 User Feedback Submitted:', {
        error: feedbackData.error,
        feedback: feedbackData.feedback,
        context: {
          url: feedbackData.url,
          timestamp: feedbackData.timestamp,
          userId: feedbackData.userId
        }
      })
      
      // In production, this would send to an error tracking service
      // await fetch('/api/error-feedback', { method: 'POST', body: JSON.stringify(feedbackData) })
      
    } catch (err) {
      console.error('Failed to submit error feedback:', err)
      throw err
    }
  }
}

// ================================
// Convenience Functions
// ================================

/**
 * Quick error handling function
 */
export function handleError(
  error: Error | ApiError | unknown,
  context: ErrorContext = {}
): ErrorReport {
  return ErrorHandler.getInstance().handleError(error, context)
}

export default ErrorHandler