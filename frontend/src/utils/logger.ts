/**
 * Production-Safe Logging Utility
 * 生产环境安全的日志工具
 */

export enum LogLevel {
  ERROR = 0,
  WARN = 1,
  INFO = 2,
  DEBUG = 3
}

export interface LogEntry {
  level: LogLevel
  message: string
  data?: any
  timestamp: string
  context?: string
}

class Logger {
  private level: LogLevel
  private isDevelopment: boolean
  private logs: LogEntry[] = []
  private maxLogs: number = 1000

  constructor() {
    this.isDevelopment = process.env.NODE_ENV === 'development'
    this.level = this.isDevelopment ? LogLevel.DEBUG : LogLevel.ERROR
  }

  private shouldLog(level: LogLevel): boolean {
    return level <= this.level
  }

  private createLogEntry(level: LogLevel, message: string, data?: any, context?: string): LogEntry {
    return {
      level,
      message,
      data,
      timestamp: new Date().toISOString(),
      context
    }
  }

  private addToHistory(entry: LogEntry): void {
    this.logs.push(entry)
    if (this.logs.length > this.maxLogs) {
      this.logs.shift()
    }
  }

  private formatMessage(level: LogLevel, message: string, context?: string): string {
    const levelStr = LogLevel[level]
    const contextStr = context ? `[${context}]` : ''
    return `${levelStr}${contextStr}: ${message}`
  }

  error(message: string, data?: any, context?: string): void {
    const entry = this.createLogEntry(LogLevel.ERROR, message, data, context)
    this.addToHistory(entry)

    if (this.shouldLog(LogLevel.ERROR)) {
      const formattedMessage = this.formatMessage(LogLevel.ERROR, message, context)
      if (data) {
        console.error(formattedMessage, data)
      } else {
        console.error(formattedMessage)
      }
    }
  }

  warn(message: string, data?: any, context?: string): void {
    const entry = this.createLogEntry(LogLevel.WARN, message, data, context)
    this.addToHistory(entry)

    if (this.shouldLog(LogLevel.WARN)) {
      const formattedMessage = this.formatMessage(LogLevel.WARN, message, context)
      if (data) {
        console.warn(formattedMessage, data)
      } else {
        console.warn(formattedMessage)
      }
    }
  }

  info(message: string, data?: any, context?: string): void {
    const entry = this.createLogEntry(LogLevel.INFO, message, data, context)
    this.addToHistory(entry)

    if (this.shouldLog(LogLevel.INFO)) {
      const formattedMessage = this.formatMessage(LogLevel.INFO, message, context)
      if (data) {
        console.info(formattedMessage, data)
      } else {
        console.info(formattedMessage)
      }
    }
  }

  debug(message: string, data?: any, context?: string): void {
    const entry = this.createLogEntry(LogLevel.DEBUG, message, data, context)
    this.addToHistory(entry)

    if (this.shouldLog(LogLevel.DEBUG)) {
      const formattedMessage = this.formatMessage(LogLevel.DEBUG, message, context)
      if (data) {
        console.log(formattedMessage, data)
      } else {
        console.log(formattedMessage)
      }
    }
  }

  // 开发环境专用方法
  devLog(message: string, data?: any, context?: string): void {
    if (this.isDevelopment) {
      this.debug(`🐛 ${message}`, data, context)
    }
  }

  // 性能日志
  performance(operation: string, duration: number, context?: string): void {
    this.debug(`⚡ ${operation} completed in ${duration}ms`, undefined, context)
  }

  // 获取日志历史
  getHistory(level?: LogLevel): LogEntry[] {
    if (level !== undefined) {
      return this.logs.filter(log => log.level === level)
    }
    return [...this.logs]
  }

  // 清除日志历史
  clearHistory(): void {
    this.logs = []
  }

  // 设置日志级别
  setLevel(level: LogLevel): void {
    this.level = level
  }

  // 获取当前日志级别
  getLevel(): LogLevel {
    return this.level
  }

  // 检查是否为开发环境
  isDev(): boolean {
    return this.isDevelopment
  }
}

// 创建全局logger实例
const logger = new Logger()

// 便捷函数
export const log = {
  error: (message: string, data?: any, context?: string) => logger.error(message, data, context),
  warn: (message: string, data?: any, context?: string) => logger.warn(message, data, context),
  info: (message: string, data?: any, context?: string) => logger.info(message, data, context),
  debug: (message: string, data?: any, context?: string) => logger.debug(message, data, context),
  dev: (message: string, data?: any, context?: string) => logger.devLog(message, data, context),
  perf: (operation: string, duration: number, context?: string) => logger.performance(operation, duration, context)
}

// 性能监控装饰器
export function logPerformance(context?: string) {
  return function (target: any, propertyName: string, descriptor: PropertyDescriptor) {
    const method = descriptor.value

    descriptor.value = async function (...args: any[]) {
      const start = performance.now()
      try {
        const result = await method.apply(this, args)
        const duration = performance.now() - start
        logger.performance(`${propertyName}`, duration, context)
        return result
      } catch (error) {
        const duration = performance.now() - start
        logger.error(`${propertyName} failed after ${duration}ms`, error, context)
        throw error
      }
    }

    return descriptor
  }
}

// 错误捕获装饰器
export function logErrors(context?: string) {
  return function (target: any, propertyName: string, descriptor: PropertyDescriptor) {
    const method = descriptor.value

    descriptor.value = async function (...args: any[]) {
      try {
        return await method.apply(this, args)
      } catch (error) {
        logger.error(`${propertyName} error`, error, context)
        throw error
      }
    }

    return descriptor
  }
}

// 导出logger实例和类型
export { logger, Logger }
export default logger