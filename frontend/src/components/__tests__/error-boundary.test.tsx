/**
 * Unit tests for Error Boundary components
 * 错误边界组件单元测试
 */

import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {
  ErrorBoundary,
  PageErrorBoundary,
  FeatureErrorBoundary,
  ComponentErrorBoundary,
  useErrorHandler
} from '../error-boundary'

// Mock logger
jest.mock('@/utils/logger', () => ({
  log: {
    error: jest.fn()
  }
}))

// Component that throws an error
const ThrowingComponent = ({ shouldThrow = false }) => {
  if (shouldThrow) {
    throw new Error('Test error')
  }
  return <div>Normal component</div>
}

// Component that uses error handler hook
const ComponentWithErrorHandler = () => {
  const { handleError } = useErrorHandler()
  
  const throwError = () => {
    try {
      throw new Error('Manual error')
    } catch (error) {
      handleError(error as Error, 'test-boundary')
    }
  }

  return (
    <div>
      <button onClick={throwError}>Throw Error</button>
    </div>
  )
}

describe('ErrorBoundary', () => {
  // Suppress console.error during tests to avoid noise
  const originalConsoleError = console.error
  beforeAll(() => {
    console.error = jest.fn()
  })
  afterAll(() => {
    console.error = originalConsoleError
  })

  test('renders children when no error occurs', () => {
    render(
      <ErrorBoundary>
        <div>Normal content</div>
      </ErrorBoundary>
    )

    expect(screen.getByText('Normal content')).toBeInTheDocument()
  })

  test('renders error UI when error occurs', () => {
    render(
      <ErrorBoundary>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(screen.getByText('组件渲染出错')).toBeInTheDocument()
    expect(screen.getByText('此功能模块暂时无法正常工作，请稍后重试')).toBeInTheDocument()
  })

  test('renders custom fallback when provided', () => {
    const customFallback = <div>Custom error fallback</div>

    render(
      <ErrorBoundary fallback={customFallback}>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(screen.getByText('Custom error fallback')).toBeInTheDocument()
    expect(screen.queryByText('组件渲染出错')).not.toBeInTheDocument()
  })

  test('shows error details in development mode', () => {
    const originalEnv = process.env.NODE_ENV
    Object.defineProperty(process.env, 'NODE_ENV', { value: 'development', writable: true })

    render(
      <ErrorBoundary showDetails={true}>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(screen.getByText('错误详情:')).toBeInTheDocument()
    expect(screen.getByText('Test error')).toBeInTheDocument()

    Object.defineProperty(process.env, 'NODE_ENV', { value: originalEnv, writable: true })
  })

  test('hides error details in production mode', () => {
    const originalEnv = process.env.NODE_ENV
    Object.defineProperty(process.env, 'NODE_ENV', { value: 'production', writable: true })

    render(
      <ErrorBoundary showDetails={false}>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(screen.queryByText('错误详情:')).not.toBeInTheDocument()
    expect(screen.queryByText('Test error')).not.toBeInTheDocument()

    Object.defineProperty(process.env, 'NODE_ENV', { value: originalEnv, writable: true })
  })

  test('retry button resets error state', async () => {
    const user = userEvent.setup()
    
    const { rerender } = render(
      <ErrorBoundary>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(screen.getByText('组件渲染出错')).toBeInTheDocument()

    const retryButton = screen.getByText('重新加载')
    await user.click(retryButton)

    // Rerender with no error
    rerender(
      <ErrorBoundary>
        <ThrowingComponent shouldThrow={false} />
      </ErrorBoundary>
    )

    await waitFor(() => {
      expect(screen.getByText('Normal component')).toBeInTheDocument()
    })
  })

  test('calls onError callback when error occurs', () => {
    const onErrorMock = jest.fn()

    render(
      <ErrorBoundary onError={onErrorMock}>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(onErrorMock).toHaveBeenCalledWith(
      expect.objectContaining({ message: 'Test error' }),
      expect.objectContaining({ componentStack: expect.any(String) })
    )
  })

  test('logs error with proper context', () => {
    const { log } = require('@/utils/logger')

    render(
      <ErrorBoundary>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>
    )

    expect(log.error).toHaveBeenCalledWith(
      'Error Boundary caught an error',
      expect.objectContaining({
        error: expect.objectContaining({
          message: 'Test error'
        }),
        errorInfo: expect.objectContaining({
          componentStack: expect.any(String)
        }),
        level: 'component'
      }),
      'ErrorBoundary'
    )
  })

  describe('different levels', () => {
    test('page level shows page-specific UI', () => {
      render(
        <ErrorBoundary level="page">
          <ThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      )

      expect(screen.getByText('页面加载出错')).toBeInTheDocument()
      expect(screen.getByText('很抱歉，页面遇到一些问题无法正常显示')).toBeInTheDocument()
      expect(screen.getByText('返回首页')).toBeInTheDocument()
      expect(screen.getByText('刷新页面')).toBeInTheDocument()
    })

    test('feature level shows feature-specific UI', () => {
      render(
        <ErrorBoundary level="feature">
          <ThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      )

      expect(screen.getByText('组件渲染出错')).toBeInTheDocument()
      expect(screen.queryByText('返回首页')).not.toBeInTheDocument()
    })

    test('component level shows minimal UI', () => {
      render(
        <ErrorBoundary level="component">
          <ThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      )

      expect(screen.getByText('组件渲染出错')).toBeInTheDocument()
      expect(screen.queryByText('返回首页')).not.toBeInTheDocument()
    })
  })
})

describe('PageErrorBoundary', () => {
  test('renders with page-level configuration', () => {
    render(
      <PageErrorBoundary>
        <ThrowingComponent shouldThrow={true} />
      </PageErrorBoundary>
    )

    expect(screen.getByText('页面加载出错')).toBeInTheDocument()
    expect(screen.getByText('返回首页')).toBeInTheDocument()
  })

  test('calls onError callback', () => {
    const onErrorMock = jest.fn()

    render(
      <PageErrorBoundary onError={onErrorMock}>
        <ThrowingComponent shouldThrow={true} />
      </PageErrorBoundary>
    )

    expect(onErrorMock).toHaveBeenCalled()
  })
})

describe('FeatureErrorBoundary', () => {
  test('renders with feature-level configuration', () => {
    render(
      <FeatureErrorBoundary>
        <ThrowingComponent shouldThrow={true} />
      </FeatureErrorBoundary>
    )

    expect(screen.getByText('组件渲染出错')).toBeInTheDocument()
    expect(screen.queryByText('返回首页')).not.toBeInTheDocument()
  })

  test('uses custom fallback when provided', () => {
    const customFallback = <div>Feature error fallback</div>

    render(
      <FeatureErrorBoundary fallback={customFallback}>
        <ThrowingComponent shouldThrow={true} />
      </FeatureErrorBoundary>
    )

    expect(screen.getByText('Feature error fallback')).toBeInTheDocument()
  })
})

describe('ComponentErrorBoundary', () => {
  test('renders with component-level configuration', () => {
    render(
      <ComponentErrorBoundary>
        <ThrowingComponent shouldThrow={true} />
      </ComponentErrorBoundary>
    )

    expect(screen.getByText('组件渲染出错')).toBeInTheDocument()
  })

  test('uses custom fallback when provided', () => {
    const customFallback = <div>Component error fallback</div>

    render(
      <ComponentErrorBoundary fallback={customFallback}>
        <ThrowingComponent shouldThrow={true} />
      </ComponentErrorBoundary>
    )

    expect(screen.getByText('Component error fallback')).toBeInTheDocument()
  })
})

describe('useErrorHandler', () => {
  test('logs error and throws in development', () => {
    const originalEnv = process.env.NODE_ENV
    Object.defineProperty(process.env, 'NODE_ENV', { value: 'development', writable: true })

    render(
      <ErrorBoundary>
        <ComponentWithErrorHandler />
      </ErrorBoundary>
    )

    const button = screen.getByText('Throw Error')
    expect(() => {
      button.click()
    }).not.toThrow() // Error boundary should catch it

    Object.defineProperty(process.env, 'NODE_ENV', { value: originalEnv, writable: true })
  })

  test('logs error without throwing in production', () => {
    const originalEnv = process.env.NODE_ENV
    Object.defineProperty(process.env, 'NODE_ENV', { value: 'production', writable: true })
    const { log } = require('@/utils/logger')

    render(<ComponentWithErrorHandler />)

    const button = screen.getByText('Throw Error')
    button.click()

    expect(log.error).toHaveBeenCalledWith(
      'Manual error reported',
      expect.objectContaining({
        error: expect.objectContaining({
          message: 'Manual error'
        }),
        errorBoundary: 'test-boundary'
      }),
      'useErrorHandler'
    )

    Object.defineProperty(process.env, 'NODE_ENV', { value: originalEnv, writable: true })
  })
})

describe('Error Recovery', () => {
  test('error boundary recovers after successful retry', async () => {
    const user = userEvent.setup()
    let shouldThrow = true

    const RecoveringComponent = () => {
      if (shouldThrow) {
        throw new Error('Recoverable error')
      }
      return <div>Recovered component</div>
    }

    const { rerender } = render(
      <ErrorBoundary>
        <RecoveringComponent />
      </ErrorBoundary>
    )

    expect(screen.getByText('组件渲染出错')).toBeInTheDocument()

    // Simulate fixing the error
    shouldThrow = false

    const retryButton = screen.getByText('重新加载')
    await user.click(retryButton)

    rerender(
      <ErrorBoundary>
        <RecoveringComponent />
      </ErrorBoundary>
    )

    await waitFor(() => {
      expect(screen.getByText('Recovered component')).toBeInTheDocument()
    })
  })
})

describe('Error Boundary Accessibility', () => {
  test('error UI is accessible', () => {
    render(
      <ErrorBoundary>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>
    )

    // Check for proper heading structure
    expect(screen.getByRole('heading', { level: 2 })).toBeInTheDocument()
    
    // Check for actionable buttons
    const retryButton = screen.getByRole('button', { name: /重新加载/i })
    expect(retryButton).toBeInTheDocument()
    expect(retryButton).not.toBeDisabled()
  })

  test('error details are properly marked up', () => {
    render(
      <ErrorBoundary showDetails={true}>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>
    )

    const errorDetails = screen.getByText('错误详情:')
    expect(errorDetails).toBeInTheDocument()
  })
})