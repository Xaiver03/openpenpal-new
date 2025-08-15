/**
 * Simple toast hook implementation
 * This is a minimal implementation to satisfy imports
 */

import { useState, useCallback } from 'react'

export interface Toast {
  id: string
  title?: string
  description?: string
  variant?: 'default' | 'destructive' | 'success'
}

export function useToast() {
  const [toasts, setToasts] = useState<Toast[]>([])

  const toast = useCallback((props: Omit<Toast, 'id'>) => {
    const id = Math.random().toString(36).substr(2, 9)
    const newToast = { ...props, id }
    setToasts(prev => [...prev, newToast])
    
    // Simple console fallback for now
    console.log('Toast:', props.title || props.description)
    
    // Auto remove after 3 seconds
    setTimeout(() => {
      setToasts(prev => prev.filter(t => t.id !== id))
    }, 3000)
    
    return { id, dismiss: () => setToasts(prev => prev.filter(t => t.id !== id)) }
  }, [])

  return {
    toast,
    toasts,
    dismiss: useCallback((id: string) => {
      setToasts(prev => prev.filter(t => t.id !== id))
    }, [])
  }
}