'use client'

import React, { createContext, useContext, useState, useCallback } from 'react'

interface TokenContextType {
  token: string | null
  userId: string | null
  setToken: (token: string | null) => void
  setUserId: (userId: string | null) => void
  clear: () => void
}

const TokenContext = createContext<TokenContextType | null>(null)

interface TokenProviderProps {
  children: React.ReactNode
}

// 独立的Token管理Provider，打破循环依赖
export function TokenProvider({ children }: TokenProviderProps) {
  const [token, setTokenState] = useState<string | null>(null)
  const [userId, setUserIdState] = useState<string | null>(null)

  const setToken = useCallback((newToken: string | null) => {
    setTokenState(newToken)
  }, [])

  const setUserId = useCallback((newUserId: string | null) => {
    setUserIdState(newUserId)
  }, [])

  const clear = useCallback(() => {
    setTokenState(null)
    setUserIdState(null)
  }, [])

  const value = {
    token,
    userId,
    setToken,
    setUserId,
    clear
  }

  return (
    <TokenContext.Provider value={value}>
      {children}
    </TokenContext.Provider>
  )
}

export function useToken() {
  const context = useContext(TokenContext)
  if (!context) {
    throw new Error('useToken must be used within TokenProvider')
  }
  return context
}