'use client'

import { useEffect } from 'react'
import { useClientUserInitialization, useUserStore } from '@/stores/user-store'

interface UserInitializerProps {
  children: React.ReactNode
}

/**
 * Initializes user state from persisted storage after client-side hydration
 * This prevents hydration mismatch errors by ensuring user state is only loaded on the client
 */
export function UserInitializer({ children }: UserInitializerProps) {
  // Initialize user state on client side only
  useClientUserInitialization()
  
  // Rehydrate zustand persist store after mount
  useEffect(() => {
    useUserStore.persist.rehydrate()
  }, [])
  
  return <>{children}</>
}