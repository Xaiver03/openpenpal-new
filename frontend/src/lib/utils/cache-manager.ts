/**
 * SOTA Cache Management System
 * 
 * Provides intelligent caching with TTL, LRU eviction,
 * memory management, and performance optimization.
 */

import React from 'react'
import { performanceMonitor } from './performance-monitor'

// ================================
// Cache Types and Interfaces
// ================================

export interface CacheEntry<T = any> {
  key: string
  value: T
  timestamp: number
  ttl: number
  hits: number
  size: number
  tags?: string[]
  metadata?: Record<string, any>
}

export interface CacheStats {
  total_entries: number
  total_size: number
  hit_count: number
  miss_count: number
  eviction_count: number
  hit_rate: number
  memory_usage: number
  oldest_entry: number
  newest_entry: number
}

export interface CacheConfig {
  maxSize: number
  defaultTTL: number
  maxMemoryUsage: number
  cleanupInterval: number
  enableMetrics: boolean
}

export type EvictionPolicy = 'lru' | 'lfu' | 'ttl' | 'size'

// ================================
// Advanced Cache Manager
// ================================

export class CacheManager {
  private static instance: CacheManager
  private cache: Map<string, CacheEntry> = new Map()
  private accessOrder: string[] = []
  private stats: CacheStats
  private config: CacheConfig
  private cleanupTimer?: NodeJS.Timeout

  constructor(config?: Partial<CacheConfig>) {
    this.config = {
      maxSize: 1000,
      defaultTTL: 5 * 60 * 1000, // 5 minutes
      maxMemoryUsage: 50 * 1024 * 1024, // 50MB
      cleanupInterval: 60 * 1000, // 1 minute
      enableMetrics: true,
      ...config
    }

    this.stats = {
      total_entries: 0,
      total_size: 0,
      hit_count: 0,
      miss_count: 0,
      eviction_count: 0,
      hit_rate: 0,
      memory_usage: 0,
      oldest_entry: 0,
      newest_entry: 0
    }

    this.startCleanupTimer()
  }

  static getInstance(config?: Partial<CacheConfig>): CacheManager {
    if (!CacheManager.instance) {
      CacheManager.instance = new CacheManager(config)
    }
    return CacheManager.instance
  }

  // ================================
  // Core Cache Operations
  // ================================

  /**
   * Set cache entry with intelligent storage
   */
  set<T>(
    key: string,
    value: T,
    ttl?: number,
    tags?: string[],
    metadata?: Record<string, any>
  ): void {
    const now = Date.now()
    const entryTTL = ttl || this.config.defaultTTL
    const size = this.calculateSize(value)

    // Remove existing entry if present
    if (this.cache.has(key)) {
      this.delete(key, false)
    }

    // Check if we need to make room
    this.ensureCapacity(size)

    // Create new entry
    const entry: CacheEntry<T> = {
      key,
      value,
      timestamp: now,
      ttl: entryTTL,
      hits: 0,
      size,
      tags,
      metadata
    }

    // Store entry
    this.cache.set(key, entry)
    this.updateAccessOrder(key)
    this.updateStats(entry, 'set')

    // Record performance metric
    if (this.config.enableMetrics) {
      performanceMonitor.recordMetric(
        'cache_set',
        1,
        'count',
        'cache',
        { key: this.sanitizeKey(key), size: size.toString() }
      )
    }
  }

  /**
   * Get cache entry with hit tracking
   */
  get<T>(key: string): T | null {
    const entry = this.cache.get(key) as CacheEntry<T> | undefined
    const now = Date.now()

    if (!entry) {
      this.stats.miss_count++
      this.updateHitRate()

      if (this.config.enableMetrics) {
        performanceMonitor.recordMetric(
          'cache_miss',
          1,
          'count',
          'cache',
          { key: this.sanitizeKey(key) }
        )
      }

      return null
    }

    // Check TTL expiration
    if (now > entry.timestamp + entry.ttl) {
      this.delete(key, false)
      this.stats.miss_count++
      this.updateHitRate()

      if (this.config.enableMetrics) {
        performanceMonitor.recordMetric(
          'cache_expired',
          1,
          'count',
          'cache',
          { key: this.sanitizeKey(key) }
        )
      }

      return null
    }

    // Update hit stats
    entry.hits++
    this.stats.hit_count++
    this.updateAccessOrder(key)
    this.updateHitRate()

    if (this.config.enableMetrics) {
      performanceMonitor.recordMetric(
        'cache_hit',
        1,
        'count',
        'cache',
        { key: this.sanitizeKey(key) }
      )
    }

    return entry.value
  }

  /**
   * Check if key exists and is valid
   */
  has(key: string): boolean {
    const entry = this.cache.get(key)
    if (!entry) return false

    const now = Date.now()
    if (now > entry.timestamp + entry.ttl) {
      this.delete(key, false)
      return false
    }

    return true
  }

  /**
   * Delete cache entry
   */
  delete(key: string, updateStats: boolean = true): boolean {
    const entry = this.cache.get(key)
    if (!entry) return false

    this.cache.delete(key)
    this.removeFromAccessOrder(key)

    if (updateStats) {
      this.updateStats(entry, 'delete')
    }

    return true
  }

  /**
   * Clear all cache entries
   */
  clear(): void {
    this.cache.clear()
    this.accessOrder = []
    this.resetStats()
  }

  // ================================
  // Advanced Operations
  // ================================

  /**
   * Get multiple keys at once
   */
  getMany<T>(keys: string[]): Map<string, T | null> {
    const results = new Map<string, T | null>()
    
    for (const key of keys) {
      results.set(key, this.get<T>(key))
    }

    return results
  }

  /**
   * Set multiple entries at once
   */
  setMany<T>(entries: Array<{
    key: string
    value: T
    ttl?: number
    tags?: string[]
  }>): void {
    for (const entry of entries) {
      this.set(entry.key, entry.value, entry.ttl, entry.tags)
    }
  }

  /**
   * Invalidate cache entries by tags
   */
  invalidateByTags(tags: string[]): number {
    let invalidated = 0

    for (const [key, entry] of this.cache.entries()) {
      if (entry.tags && entry.tags.some(tag => tags.includes(tag))) {
        this.delete(key)
        invalidated++
      }
    }

    return invalidated
  }

  /**
   * Invalidate cache entries matching pattern
   */
  invalidateByPattern(pattern: RegExp): number {
    let invalidated = 0

    for (const key of this.cache.keys()) {
      if (pattern.test(key)) {
        this.delete(key)
        invalidated++
      }
    }

    return invalidated
  }

  /**
   * Update TTL for existing entry
   */
  updateTTL(key: string, newTTL: number): boolean {
    const entry = this.cache.get(key)
    if (!entry) return false

    entry.ttl = newTTL
    entry.timestamp = Date.now() // Reset timestamp
    return true
  }

  /**
   * Get cache entry metadata
   */
  getMetadata(key: string): CacheEntry | null {
    return this.cache.get(key) || null
  }

  // ================================
  // Capacity Management
  // ================================

  private ensureCapacity(newEntrySize: number): void {
    // Check memory usage
    if (this.stats.total_size + newEntrySize > this.config.maxMemoryUsage) {
      this.evictBySize(newEntrySize)
    }

    // Check entry count
    if (this.cache.size >= this.config.maxSize) {
      this.evictByPolicy('lru')
    }
  }

  private evictByPolicy(policy: EvictionPolicy, count: number = 1): void {
    switch (policy) {
      case 'lru':
        this.evictLRU(count)
        break
      case 'lfu':
        this.evictLFU(count)
        break
      case 'ttl':
        this.evictExpired()
        break
      case 'size':
        this.evictLargest(count)
        break
    }
  }

  private evictLRU(count: number): void {
    for (let i = 0; i < count && this.accessOrder.length > 0; i++) {
      const oldestKey = this.accessOrder[0]
      this.delete(oldestKey)
      this.stats.eviction_count++
    }
  }

  private evictLFU(count: number): void {
    const entries = Array.from(this.cache.entries())
      .sort(([, a], [, b]) => a.hits - b.hits)

    for (let i = 0; i < Math.min(count, entries.length); i++) {
      this.delete(entries[i][0])
      this.stats.eviction_count++
    }
  }

  private evictExpired(): void {
    const now = Date.now()
    const expiredKeys: string[] = []

    for (const [key, entry] of this.cache.entries()) {
      if (now > entry.timestamp + entry.ttl) {
        expiredKeys.push(key)
      }
    }

    for (const key of expiredKeys) {
      this.delete(key)
      this.stats.eviction_count++
    }
  }

  private evictLargest(count: number): void {
    const entries = Array.from(this.cache.entries())
      .sort(([, a], [, b]) => b.size - a.size)

    for (let i = 0; i < Math.min(count, entries.length); i++) {
      this.delete(entries[i][0])
      this.stats.eviction_count++
    }
  }

  private evictBySize(requiredSize: number): void {
    while (this.stats.total_size + requiredSize > this.config.maxMemoryUsage) {
      const oldestKey = this.accessOrder[0]
      if (!oldestKey) break
      
      this.delete(oldestKey)
      this.stats.eviction_count++
    }
  }

  // ================================
  // Utility Methods
  // ================================

  private updateAccessOrder(key: string): void {
    // Remove from current position
    this.removeFromAccessOrder(key)
    // Add to end (most recently used)
    this.accessOrder.push(key)
  }

  private removeFromAccessOrder(key: string): void {
    const index = this.accessOrder.indexOf(key)
    if (index > -1) {
      this.accessOrder.splice(index, 1)
    }
  }

  private calculateSize(value: any): number {
    try {
      return JSON.stringify(value).length * 2 // Approximate UTF-16 size
    } catch {
      return 100 // Fallback size for non-serializable objects
    }
  }

  private sanitizeKey(key: string): string {
    // Remove sensitive information from key for metrics
    return key.replace(/[a-f0-9]{32,}/gi, '[hash]').substring(0, 50)
  }

  private updateStats(entry: CacheEntry, operation: 'set' | 'delete'): void {
    if (operation === 'set') {
      this.stats.total_entries++
      this.stats.total_size += entry.size
      this.stats.memory_usage = this.stats.total_size
      
      if (this.stats.oldest_entry === 0 || entry.timestamp < this.stats.oldest_entry) {
        this.stats.oldest_entry = entry.timestamp
      }
      
      if (entry.timestamp > this.stats.newest_entry) {
        this.stats.newest_entry = entry.timestamp
      }
    } else if (operation === 'delete') {
      this.stats.total_entries = Math.max(0, this.stats.total_entries - 1)
      this.stats.total_size = Math.max(0, this.stats.total_size - entry.size)
      this.stats.memory_usage = this.stats.total_size
    }
  }

  private updateHitRate(): void {
    const totalRequests = this.stats.hit_count + this.stats.miss_count
    this.stats.hit_rate = totalRequests > 0 
      ? (this.stats.hit_count / totalRequests) * 100 
      : 0
  }

  private resetStats(): void {
    this.stats = {
      total_entries: 0,
      total_size: 0,
      hit_count: 0,
      miss_count: 0,
      eviction_count: 0,
      hit_rate: 0,
      memory_usage: 0,
      oldest_entry: 0,
      newest_entry: 0
    }
  }

  // ================================
  // Cleanup and Maintenance
  // ================================

  private startCleanupTimer(): void {
    if (this.cleanupTimer) {
      clearInterval(this.cleanupTimer)
    }

    this.cleanupTimer = setInterval(() => {
      this.cleanup()
    }, this.config.cleanupInterval)
  }

  private cleanup(): void {
    const before = this.cache.size
    this.evictExpired()
    const after = this.cache.size

    if (this.config.enableMetrics && before > after) {
      performanceMonitor.recordMetric(
        'cache_cleanup',
        before - after,
        'count',
        'cache'
      )
    }
  }

  // ================================
  // Public API
  // ================================

  /**
   * Get current cache statistics
   */
  getStats(): CacheStats {
    return { ...this.stats }
  }

  /**
   * Get cache configuration
   */
  getConfig(): CacheConfig {
    return { ...this.config }
  }

  /**
   * Update cache configuration
   */
  updateConfig(newConfig: Partial<CacheConfig>): void {
    this.config = { ...this.config, ...newConfig }
    
    // Restart cleanup timer if interval changed
    if (newConfig.cleanupInterval) {
      this.startCleanupTimer()
    }
  }

  /**
   * Export cache state for debugging
   */
  exportState(): {
    entries: Array<{ key: string; metadata: CacheEntry }>
    stats: CacheStats
    config: CacheConfig
  } {
    return {
      entries: Array.from(this.cache.entries()).map(([key, entry]) => ({
        key,
        metadata: { ...entry, value: '[hidden]' } as any
      })),
      stats: this.getStats(),
      config: this.getConfig()
    }
  }

  /**
   * Cleanup resources
   */
  destroy(): void {
    if (this.cleanupTimer) {
      clearInterval(this.cleanupTimer)
    }
    this.clear()
  }
}

// ================================
// Specialized Cache Types
// ================================

/**
 * API Response Cache
 */
export class APICache extends CacheManager {
  constructor() {
    super({
      defaultTTL: 2 * 60 * 1000, // 2 minutes for API responses
      maxSize: 500,
      enableMetrics: true
    })
  }

  cacheResponse<T>(endpoint: string, method: string, params: any, response: T): void {
    const key = this.createAPIKey(endpoint, method, params)
    this.set(key, response, undefined, ['api', method.toLowerCase()])
  }

  getCachedResponse<T>(endpoint: string, method: string, params: any): T | null {
    const key = this.createAPIKey(endpoint, method, params)
    return this.get<T>(key)
  }

  invalidateEndpoint(endpoint: string): void {
    this.invalidateByPattern(new RegExp(`^api:${endpoint.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}`))
  }

  private createAPIKey(endpoint: string, method: string, params: any): string {
    const paramsHash = this.hashParams(params)
    return `api:${endpoint}:${method}:${paramsHash}`
  }

  private hashParams(params: any): string {
    try {
      return btoa(JSON.stringify(params)).substring(0, 16)
    } catch {
      return 'no-params'
    }
  }
}

/**
 * Component State Cache
 */
export class ComponentCache extends CacheManager {
  constructor() {
    super({
      defaultTTL: 10 * 60 * 1000, // 10 minutes
      maxSize: 200,
      enableMetrics: true
    })
  }

  cacheComponentState(componentName: string, state: any): void {
    this.set(`component:${componentName}`, state, undefined, ['component'])
  }

  getCachedComponentState(componentName: string): any {
    return this.get(`component:${componentName}`)
  }
}

// ================================
// React Hooks for Cache
// ================================

/**
 * Hook for API caching
 */
export function useAPICache() {
  const cache = React.useMemo(() => new APICache(), [])

  const getCached = React.useCallback(<T>(
    endpoint: string,
    method: string,
    params: any
  ): T | null => {
    return cache.getCachedResponse<T>(endpoint, method, params)
  }, [cache])

  const setCached = React.useCallback(<T>(
    endpoint: string,
    method: string,
    params: any,
    response: T
  ): void => {
    cache.cacheResponse(endpoint, method, params, response)
  }, [cache])

  const invalidate = React.useCallback((endpoint: string): void => {
    cache.invalidateEndpoint(endpoint)
  }, [cache])

  return { getCached, setCached, invalidate, stats: cache.getStats() }
}

// ================================
// Global Cache Instances
// ================================

export const globalCache = CacheManager.getInstance()
export const apiCache = new APICache()
export const componentCache = new ComponentCache()

export default CacheManager