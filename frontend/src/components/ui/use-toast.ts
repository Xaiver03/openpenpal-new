/**
 * Toast Hook
 * Toast通知钩子
 */

import { useState, useCallback } from 'react'

export interface ToastProps {
  id?: string
  title?: string
  description?: string
  variant?: 'default' | 'destructive' | 'success'
  duration?: number
  action?: React.ReactNode
}

interface Toast extends Required<Omit<ToastProps, 'id'>> {
  id: string
}

const TOAST_LIMIT = 5
const TOAST_REMOVE_DELAY = 5000

let count = 0

function genId() {
  count = (count + 1) % Number.MAX_VALUE
  return count.toString()
}

type ToasterToast = Toast

const toastTimeouts = new Map<string, ReturnType<typeof setTimeout>>()
const listeners: Array<(toasts: ToasterToast[]) => void> = []

let memoryState: ToasterToast[] = []

function dispatch(action: {
  type: 'ADD_TOAST' | 'UPDATE_TOAST' | 'DISMISS_TOAST' | 'REMOVE_TOAST'
  toast?: Partial<ToasterToast>
  toastId?: string
}) {
  switch (action.type) {
    case 'ADD_TOAST':
      if (action.toast) {
        const toast: ToasterToast = {
          id: genId(),
          title: '',
          description: '',
          variant: 'default',
          duration: TOAST_REMOVE_DELAY,
          action: null,
          ...action.toast
        }
        
        memoryState = [toast, ...memoryState].slice(0, TOAST_LIMIT)
        
        listeners.forEach((listener) => {
          listener(memoryState)
        })

        if (toast.duration !== Infinity) {
          toastTimeouts.set(
            toast.id,
            setTimeout(() => {
              dispatch({
                type: 'REMOVE_TOAST',
                toastId: toast.id
              })
            }, toast.duration)
          )
        }
      }
      break

    case 'UPDATE_TOAST':
      if (action.toast?.id) {
        memoryState = memoryState.map((t) =>
          t.id === action.toast!.id ? { ...t, ...action.toast } : t
        )
        listeners.forEach((listener) => {
          listener(memoryState)
        })
      }
      break

    case 'DISMISS_TOAST':
      if (action.toastId) {
        const timeoutId = toastTimeouts.get(action.toastId)
        if (timeoutId) {
          clearTimeout(timeoutId)
          toastTimeouts.delete(action.toastId)
        }

        memoryState = memoryState.map((t) =>
          t.id === action.toastId ? { ...t, open: false } : t
        )
        listeners.forEach((listener) => {
          listener(memoryState)
        })

        setTimeout(() => {
          dispatch({
            type: 'REMOVE_TOAST',
            toastId: action.toastId
          })
        }, 200)
      }
      break

    case 'REMOVE_TOAST':
      if (action.toastId) {
        memoryState = memoryState.filter((t) => t.id !== action.toastId)
        listeners.forEach((listener) => {
          listener(memoryState)
        })
      }
      break
  }
}

function reducer(state: ToasterToast[], action: any) {
  switch (action.type) {
    case 'ADD_TOAST':
      return [action.toast, ...state].slice(0, TOAST_LIMIT)
    case 'UPDATE_TOAST':
      return state.map((t) =>
        t.id === action.toast.id ? { ...t, ...action.toast } : t
      )
    case 'DISMISS_TOAST':
      const { toastId } = action
      return state.map((t) =>
        t.id === toastId || toastId === undefined
          ? {
              ...t,
              open: false,
            }
          : t
      )
    case 'REMOVE_TOAST':
      if (action.toastId === undefined) {
        return []
      }
      return state.filter((t) => t.id !== action.toastId)
  }
}

function useToast() {
  const [state, setState] = useState<ToasterToast[]>(memoryState)

  useState(() => {
    listeners.push(setState)
    return () => {
      const index = listeners.indexOf(setState)
      if (index > -1) {
        listeners.splice(index, 1)
      }
    }
  })

  return {
    toasts: state,
    toast: useCallback((props: ToastProps) => {
      dispatch({
        type: 'ADD_TOAST',
        toast: props
      })
    }, []),
    dismiss: useCallback((toastId?: string) => {
      dispatch({
        type: 'DISMISS_TOAST',
        toastId
      })
    }, [])
  }
}

// Simple toast function for immediate use
const toast = (props: ToastProps) => {
  dispatch({
    type: 'ADD_TOAST',
    toast: props
  })
}

export { useToast, toast }