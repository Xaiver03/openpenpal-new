'use client'

import React, { useEffect, useCallback, useRef } from 'react'
import { useAuth } from '@/contexts/auth-context-new'
import { useUserStore } from '@/stores/user-store'
import { apiClient, TokenManager } from '@/lib/api-client'

interface TokenRefreshProviderProps {
  children: React.ReactNode
}

export function TokenRefreshProvider({ children }: TokenRefreshProviderProps) {
  const { user, logout, refreshUser } = useAuth()
  const { updateUser } = useUserStore()
  const refreshIntervalRef = useRef<NodeJS.Timeout | null>(null)
  const isRefreshingRef = useRef(false)

  const refreshToken = useCallback(async () => {
    if (!user || isRefreshingRef.current) return

    try {
      isRefreshingRef.current = true
      const response = await apiClient.post('/auth/refresh')
      
      if (response.success) {
        // Refresh user data from server (which includes the new token)
        await refreshUser()
      } else {
        console.error('Token refresh failed:', response.message)
        logout()
      }
    } catch (error) {
      console.error('Token refresh error:', error)
      // Don't logout on network errors, just log
    } finally {
      isRefreshingRef.current = false
    }
  }, [user, logout, refreshUser])

  const checkTokenExpiry = useCallback(async () => {
    const token = TokenManager.get()
    if (!token || !user) return

    // Check if token is expired
    if (TokenManager.isExpired(token)) {
      logout()
    } else {
      // Check if token is close to expiry (refresh 5 minutes before expiry)
      try {
        const parts = token.split('.')
        if (parts.length === 3) {
          const payload = JSON.parse(atob(parts[1]))
          if (payload.exp) {
            const expiryTime = payload.exp * 1000
            const currentTime = Date.now()
            const fiveMinutesInMs = 5 * 60 * 1000
            
            if (expiryTime - currentTime <= fiveMinutesInMs) {
              await refreshToken()
            }
          }
        }
      } catch (error) {
        console.error('Error checking token expiry:', error)
        await refreshToken()
      }
    }
  }, [user, logout, refreshToken])

  useEffect(() => {
    const token = TokenManager.get()
    if (!token || !user) {
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current)
        refreshIntervalRef.current = null
      }
      return
    }

    // Check token expiry immediately
    checkTokenExpiry()

    // Set up periodic checks every 5 minutes
    refreshIntervalRef.current = setInterval(() => {
      checkTokenExpiry()
    }, 5 * 60 * 1000)

    // Clean up on unmount
    return () => {
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current)
      }
    }
  }, [user, checkTokenExpiry])

  // Handle visibility change
  useEffect(() => {
    const handleVisibilityChange = () => {
      const token = TokenManager.get()
      if (!document.hidden && token && user) {
        checkTokenExpiry()
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [user, checkTokenExpiry])

  return <>{children}</>
}