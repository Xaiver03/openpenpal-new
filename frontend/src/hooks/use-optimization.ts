import { useState, useEffect, useCallback, useMemo, useRef } from 'react'

/**
 * Optimization hooks for performance and user experience
 */

// Debounce hook
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)

    return () => {
      clearTimeout(handler)
    }
  }, [value, delay])

  return debouncedValue
}

// Throttle hook
export function useThrottle<T>(value: T, limit: number): T {
  const [throttledValue, setThrottledValue] = useState<T>(value)
  const lastRan = useRef<number>(Date.now())

  useEffect(() => {
    const handler = setTimeout(() => {
      if (Date.now() - lastRan.current >= limit) {
        setThrottledValue(value)
        lastRan.current = Date.now()
      }
    }, limit - (Date.now() - lastRan.current))

    return () => {
      clearTimeout(handler)
    }
  }, [value, limit])

  return throttledValue
}

// Local storage with SSR safety
export function useLocalStorage<T>(
  key: string,
  initialValue: T
): [T, (value: T | ((val: T) => T)) => void] {
  const [storedValue, setStoredValue] = useState<T>(() => {
    if (typeof window === 'undefined') {
      return initialValue
    }
    try {
      const item = window.localStorage.getItem(key)
      return item ? JSON.parse(item) : initialValue
    } catch (error) {
      console.warn(`Error reading localStorage key "${key}":`, error)
      return initialValue
    }
  })

  const setValue = useCallback((value: T | ((val: T) => T)) => {
    try {
      const valueToStore = value instanceof Function ? value(storedValue) : value
      setStoredValue(valueToStore)
      if (typeof window !== 'undefined') {
        window.localStorage.setItem(key, JSON.stringify(valueToStore))
      }
    } catch (error) {
      console.warn(`Error setting localStorage key "${key}":`, error)
    }
  }, [key, storedValue])

  return [storedValue, setValue]
}

// Intersection observer hook
export function useIntersectionObserver(
  elementRef: React.RefObject<Element>,
  {
    threshold = 0,
    root = null,
    rootMargin = '0%',
    freezeOnceVisible = false
  }: {
    threshold?: number
    root?: Element | null
    rootMargin?: string
    freezeOnceVisible?: boolean
  } = {}
): boolean {
  const [isIntersecting, setIsIntersecting] = useState(false)

  useEffect(() => {
    const element = elementRef?.current

    if (!element) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        const isElementIntersecting = entry.isIntersecting
        if (!freezeOnceVisible || !isIntersecting) {
          setIsIntersecting(isElementIntersecting)
        }
      },
      { threshold, root, rootMargin }
    )

    observer.observe(element)

    return () => observer.unobserve(element)
  }, [elementRef, threshold, root, rootMargin, freezeOnceVisible, isIntersecting])

  return isIntersecting
}

// Media query hook
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(false)

  useEffect(() => {
    if (typeof window === 'undefined') return

    const media = window.matchMedia(query)
    if (media.matches !== matches) {
      setMatches(media.matches)
    }

    const listener = () => setMatches(media.matches)
    media.addEventListener('change', listener)

    return () => media.removeEventListener('change', listener)
  }, [matches, query])

  return matches
}

// Previous value hook
export function usePrevious<T>(value: T): T | undefined {
  const ref = useRef<T>()
  useEffect(() => {
    ref.current = value
  })
  return ref.current
}

// Online status hook
export function useOnlineStatus(): boolean {
  const [isOnline, setIsOnline] = useState(true)

  useEffect(() => {
    if (typeof window === 'undefined') return

    setIsOnline(navigator.onLine)

    const handleOnline = () => setIsOnline(true)
    const handleOffline = () => setIsOnline(false)

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  return isOnline
}

// Optimized search hook
export function useOptimizedSearch<T>(
  items: T[],
  searchTerm: string,
  searchFields: (keyof T)[],
  options: {
    debounceDelay?: number
    caseSensitive?: boolean
    minSearchLength?: number
  } = {}
) {
  const {
    debounceDelay = 300,
    caseSensitive = false,
    minSearchLength = 1
  } = options

  const debouncedSearchTerm = useDebounce(searchTerm, debounceDelay)

  const filteredItems = useMemo(() => {
    if (!debouncedSearchTerm || debouncedSearchTerm.length < minSearchLength) {
      return items
    }

    const searchValue = caseSensitive ? debouncedSearchTerm : debouncedSearchTerm.toLowerCase()

    return items.filter(item =>
      searchFields.some(field => {
        const fieldValue = item[field]
        if (typeof fieldValue === 'string') {
          const compareValue = caseSensitive ? fieldValue : fieldValue.toLowerCase()
          return compareValue.includes(searchValue)
        }
        return false
      })
    )
  }, [items, debouncedSearchTerm, searchFields, caseSensitive, minSearchLength])

  return {
    filteredItems,
    isSearching: searchTerm !== debouncedSearchTerm,
    searchTerm: debouncedSearchTerm
  }
}

// Optimized pagination hook
export function usePagination<T>(
  items: T[],
  itemsPerPage: number = 20
) {
  const [currentPage, setCurrentPage] = useState(1)

  const paginationData = useMemo(() => {
    const totalPages = Math.ceil(items.length / itemsPerPage)
    const startIndex = (currentPage - 1) * itemsPerPage
    const endIndex = startIndex + itemsPerPage
    const currentItems = items.slice(startIndex, endIndex)

    return {
      currentItems,
      currentPage,
      totalPages,
      totalItems: items.length,
      hasNextPage: currentPage < totalPages,
      hasPreviousPage: currentPage > 1,
      startIndex: startIndex + 1,
      endIndex: Math.min(endIndex, items.length)
    }
  }, [items, itemsPerPage, currentPage])

  const goToPage = useCallback((page: number) => {
    setCurrentPage(Math.max(1, Math.min(page, paginationData.totalPages)))
  }, [paginationData.totalPages])

  const nextPage = useCallback(() => {
    goToPage(currentPage + 1)
  }, [currentPage, goToPage])

  const previousPage = useCallback(() => {
    goToPage(currentPage - 1)
  }, [currentPage, goToPage])

  // Reset to first page when items change
  useEffect(() => {
    setCurrentPage(1)
  }, [items.length])

  return {
    ...paginationData,
    goToPage,
    nextPage,
    previousPage,
    setCurrentPage
  }
}

// Form validation hook
export function useFormValidation<T extends Record<string, any>>(
  initialValues: T,
  validationRules: Partial<Record<keyof T, (value: any) => string | null>>
) {
  const [values, setValues] = useState<T>(initialValues)
  const [errors, setErrors] = useState<Partial<Record<keyof T, string>>>({})
  const [touched, setTouched] = useState<Partial<Record<keyof T, boolean>>>({})

  const validate = useCallback((field?: keyof T) => {
    const newErrors: Partial<Record<keyof T, string>> = {}

    const fieldsToValidate = field ? [field] : Object.keys(validationRules) as (keyof T)[]

    fieldsToValidate.forEach(fieldName => {
      const rule = validationRules[fieldName]
      if (rule) {
        const error = rule(values[fieldName])
        if (error) {
          newErrors[fieldName] = error
        }
      }
    })

    if (field) {
      setErrors(prev => ({ ...prev, ...newErrors }))
    } else {
      setErrors(newErrors)
    }

    return Object.keys(newErrors).length === 0
  }, [values, validationRules])

  const setValue = useCallback((field: keyof T, value: any) => {
    setValues(prev => ({ ...prev, [field]: value }))
    setTouched(prev => ({ ...prev, [field]: true }))
    
    // Validate field immediately
    setTimeout(() => validate(field), 0)
  }, [validate])

  const reset = useCallback(() => {
    setValues(initialValues)
    setErrors({})
    setTouched({})
  }, [initialValues])

  const isValid = Object.keys(errors).length === 0
  const hasErrors = Object.keys(errors).length > 0

  return {
    values,
    errors,
    touched,
    isValid,
    hasErrors,
    setValue,
    validate,
    reset,
    setValues
  }
}

// Optimized data fetching hook
export function useOptimizedFetch<T>(
  url: string,
  options: {
    enabled?: boolean
    refetchInterval?: number
    cacheTime?: number
    retryCount?: number
  } = {}
) {
  const {
    enabled = true,
    refetchInterval,
    cacheTime = 5 * 60 * 1000, // 5 minutes
    retryCount = 3
  } = options

  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const retryCountRef = useRef(0)

  const fetchData = useCallback(async () => {
    if (!enabled) return

    setLoading(true)
    setError(null)

    try {
      const response = await fetch(url)
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const result = await response.json()
      setData(result)
      retryCountRef.current = 0
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An error occurred'
      setError(errorMessage)
      
      // Retry logic
      if (retryCountRef.current < retryCount) {
        retryCountRef.current++
        setTimeout(fetchData, 1000 * retryCountRef.current) // Exponential backoff
      }
    } finally {
      setLoading(false)
    }
  }, [url, enabled, retryCount])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  useEffect(() => {
    if (refetchInterval && enabled) {
      const interval = setInterval(fetchData, refetchInterval)
      return () => clearInterval(interval)
    }
  }, [fetchData, refetchInterval, enabled])

  return {
    data,
    loading,
    error,
    refetch: fetchData
  }
}

// Keyboard shortcuts hook
export function useKeyboardShortcuts(shortcuts: Record<string, () => void>) {
  useEffect(() => {
    if (typeof window === 'undefined') return

    const handleKeyDown = (event: KeyboardEvent) => {
      const { key, ctrlKey, metaKey, shiftKey, altKey } = event

      const modifiers = [
        ctrlKey && 'ctrl',
        metaKey && 'cmd',
        shiftKey && 'shift',
        altKey && 'alt'
      ].filter(Boolean).join('+')

      const shortcutKey = modifiers ? `${modifiers}+${key.toLowerCase()}` : key.toLowerCase()

      if (shortcuts[shortcutKey]) {
        event.preventDefault()
        shortcuts[shortcutKey]()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [shortcuts])
}