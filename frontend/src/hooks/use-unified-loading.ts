/**
 * Unified Loading State Hook - 统一加载状态管理
 * Centralized loading state management for all async operations
 */

import { useState, useCallback, useRef } from 'react'
import { useUserStore } from '@/stores/user-store'

export interface LoadingOptions {
  showGlobalLoading?: boolean
  timeout?: number
  retries?: number
  retryDelay?: number
}

export interface LoadingState {
  isLoading: boolean
  error: string | null
  progress?: number
  operation?: string
}

/**
 * Hook for managing unified loading states
 */
export function useUnifiedLoading() {
  const [localLoading, setLocalLoading] = useState<LoadingState>({
    isLoading: false,
    error: null
  })

  const { loading: globalLoading, setLoading: setGlobalLoading } = useUserStore()
  const timeoutRef = useRef<NodeJS.Timeout>()

  /**
   * Execute async operation with unified loading management
   */
  const execute = useCallback(async <T>(
    operation: () => Promise<T>,
    options: LoadingOptions & {
      operationName?: string
      onProgress?: (progress: number) => void
    } = {}
  ): Promise<T> => {
    const {
      showGlobalLoading = false,
      timeout = 30000,
      retries = 0,
      retryDelay = 1000,
      operationName = 'Loading...',
      onProgress
    } = options

    const setLoadingState = showGlobalLoading ? setGlobalLoading : setLocalLoading

    // Clear any existing timeout
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    // Set loading state
    setLoadingState({
      isLoading: true,
      error: null,
      operation: operationName
    })

    // Set timeout
    if (timeout > 0) {
      timeoutRef.current = setTimeout(() => {
        setLoadingState(prev => ({
          ...prev,
          error: `Operation timed out after ${timeout}ms`
        }))
      }, timeout)
    }

    let lastError: Error | null = null
    let attempt = 0

    while (attempt <= retries) {
      try {
        // Progress callback for operation start
        onProgress?.(attempt > 0 ? 25 : 0)

        const result = await operation()

        // Clear timeout on success
        if (timeoutRef.current) {
          clearTimeout(timeoutRef.current)
        }

        // Progress callback for success
        onProgress?.(100)

        // Clear loading state
        setLoadingState({
          isLoading: false,
          error: null
        })

        return result
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error))
        
        if (attempt < retries) {
          // Progress callback for retry
          onProgress?.(25 + (attempt * 25))
          
          // Wait before retry
          await new Promise(resolve => setTimeout(resolve, retryDelay))
        }
        
        attempt++
      }
    }

    // All attempts failed
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    const errorMessage = lastError?.message || 'Operation failed'
    setLoadingState({
      isLoading: false,
      error: errorMessage
    })

    throw lastError
  }, [setGlobalLoading])

  /**
   * Clear loading state
   */
  const clearLoading = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
    
    setLocalLoading({
      isLoading: false,
      error: null
    })
  }, [])

  /**
   * Set loading state manually
   */
  const setLoading = useCallback((
    loading: boolean, 
    options: { error?: string; operation?: string } = {}
  ) => {
    setLocalLoading({
      isLoading: loading,
      error: options.error || null,
      operation: options.operation
    })
  }, [])

  return {
    // Local loading state
    localLoading,
    setLoading,
    clearLoading,
    
    // Global loading state (from user store)
    globalLoading,
    
    // Execution helper
    execute,
    
    // Combined loading state
    isLoading: localLoading.isLoading || globalLoading.isLoading,
    error: localLoading.error || globalLoading.error,
    operation: localLoading.operation || globalLoading.error
  }
}

/**
 * Hook for specific operation loading states
 */
export function useOperationLoading(operationName: string) {
  const [operationStates, setOperationStates] = useState<Record<string, LoadingState>>({})

  const setOperationLoading = useCallback((
    loading: boolean,
    error?: string
  ) => {
    setOperationStates(prev => ({
      ...prev,
      [operationName]: {
        isLoading: loading,
        error: error || null,
        operation: operationName
      }
    }))
  }, [operationName])

  const executeOperation = useCallback(async <T>(
    operation: () => Promise<T>,
    options: LoadingOptions = {}
  ): Promise<T> => {
    setOperationLoading(true)
    
    try {
      const result = await operation()
      setOperationLoading(false)
      return result
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error)
      setOperationLoading(false, errorMessage)
      throw error
    }
  }, [setOperationLoading])

  const clearOperationLoading = useCallback(() => {
    setOperationStates(prev => {
      const { [operationName]: _, ...rest } = prev
      return rest
    })
  }, [operationName])

  const currentState = operationStates[operationName] || {
    isLoading: false,
    error: null
  }

  return {
    ...currentState,
    setLoading: setOperationLoading,
    clearLoading: clearOperationLoading,
    execute: executeOperation
  }
}

/**
 * Hook for batch operations with progress tracking
 */
export function useBatchLoading() {
  const [batchState, setBatchState] = useState<{
    isLoading: boolean
    progress: number
    currentOperation: string | null
    completedOperations: string[]
    failedOperations: Array<{ operation: string; error: string }>
    totalOperations: number
  }>({
    isLoading: false,
    progress: 0,
    currentOperation: null,
    completedOperations: [],
    failedOperations: [],
    totalOperations: 0
  })

  const executeBatch = useCallback(async <T>(
    operations: Array<{
      name: string
      operation: () => Promise<T>
    }>,
    options: {
      continueOnError?: boolean
      onProgress?: (progress: number, current: string) => void
    } = {}
  ): Promise<Array<T | Error>> => {
    const { continueOnError = false, onProgress } = options
    const results: Array<T | Error> = []

    setBatchState({
      isLoading: true,
      progress: 0,
      currentOperation: null,
      completedOperations: [],
      failedOperations: [],
      totalOperations: operations.length
    })

    for (let i = 0; i < operations.length; i++) {
      const { name, operation } = operations[i]
      
      setBatchState(prev => ({
        ...prev,
        currentOperation: name,
        progress: (i / operations.length) * 100
      }))

      onProgress?.((i / operations.length) * 100, name)

      try {
        const result = await operation()
        results.push(result)
        
        setBatchState(prev => ({
          ...prev,
          completedOperations: [...prev.completedOperations, name]
        }))
      } catch (error) {
        const errorObj = error instanceof Error ? error : new Error(String(error))
        results.push(errorObj)
        
        setBatchState(prev => ({
          ...prev,
          failedOperations: [
            ...prev.failedOperations,
            { operation: name, error: errorObj.message }
          ]
        }))

        if (!continueOnError) {
          setBatchState(prev => ({
            ...prev,
            isLoading: false,
            progress: 100
          }))
          throw errorObj
        }
      }
    }

    setBatchState(prev => ({
      ...prev,
      isLoading: false,
      progress: 100,
      currentOperation: null
    }))

    onProgress?.(100, 'Completed')

    return results
  }, [])

  const clearBatch = useCallback(() => {
    setBatchState({
      isLoading: false,
      progress: 0,
      currentOperation: null,
      completedOperations: [],
      failedOperations: [],
      totalOperations: 0
    })
  }, [])

  return {
    ...batchState,
    execute: executeBatch,
    clear: clearBatch
  }
}

export default useUnifiedLoading