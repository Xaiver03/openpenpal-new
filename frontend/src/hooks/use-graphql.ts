import { useState, useEffect, useCallback } from 'react'
import { graphqlClient } from '@/lib/graphql/client'

interface QueryState<T> {
  data: T | null
  loading: boolean
  error: string | null
}

interface MutationState<T> {
  data: T | null
  loading: boolean
  error: string | null
}

export function useQuery<T = any>(
  query: string,
  variables?: Record<string, any>,
  options?: {
    skip?: boolean
    pollInterval?: number
  }
): QueryState<T> & { refetch: () => void } {
  const [state, setState] = useState<QueryState<T>>({
    data: null,
    loading: !options?.skip,
    error: null
  })

  const executeQuery = useCallback(async () => {
    if (options?.skip) return

    setState(prev => ({ ...prev, loading: true, error: null }))

    try {
      const data = await graphqlClient.query<T>(query, variables)
      setState({ data, loading: false, error: null })
    } catch (error) {
      setState({ data: null, loading: false, error: error instanceof Error ? error.message : 'Unknown error' })
    }
  }, [query, variables, options?.skip])

  const refetch = useCallback(() => {
    executeQuery()
  }, [executeQuery])

  useEffect(() => {
    executeQuery()
  }, [executeQuery])

  // Polling
  useEffect(() => {
    if (options?.pollInterval && options.pollInterval > 0) {
      const interval = setInterval(executeQuery, options.pollInterval)
      return () => clearInterval(interval)
    }
  }, [executeQuery, options?.pollInterval])

  return {
    ...state,
    refetch
  }
}

export function useMutation<T = any>(
  mutation: string
): [
  (variables?: Record<string, any>) => Promise<T>,
  MutationState<T>
] {
  const [state, setState] = useState<MutationState<T>>({
    data: null,
    loading: false,
    error: null
  })

  const executeMutation = useCallback(async (variables?: Record<string, any>) => {
    setState({ data: null, loading: true, error: null })

    try {
      const data = await graphqlClient.mutate<T>(mutation, variables)
      setState({ data, loading: false, error: null })
      return data
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error'
      setState({ data: null, loading: false, error: errorMessage })
      throw error
    }
  }, [mutation])

  return [executeMutation, state]
}

// Specialized hooks for common operations
export function useAuth() {
  const { data: userData, loading: userLoading, error: userError, refetch } = useQuery(`
    query Me {
      me {
        id
        username
        email
        nickname
        role
        schoolCode
        isActive
        createdAt
        stats {
          totalLetters
          draftCount
          generatedCount
          deliveredCount
        }
      }
    }
  `)

  const [loginMutation, { loading: loginLoading, error: loginError }] = useMutation(`
    mutation Login($email: String!, $password: String!) {
      login(email: $email, password: $password) {
        token
        user {
          id
          username
          email
          nickname
          role
          schoolCode
          isActive
        }
      }
    }
  `)

  const [logoutMutation] = useMutation(`
    mutation Logout {
      logout
    }
  `)

  const login = useCallback(async (email: string, password: string) => {
    try {
      const result = await loginMutation({ email, password })
      if (result.login?.token) {
        graphqlClient.setAuthToken(result.login.token)
        if (typeof window !== 'undefined') {
          localStorage.setItem('auth_token', result.login.token)
        }
        await refetch() // Fetch user data after login
      }
      return result.login
    } catch (error) {
      throw error
    }
  }, [loginMutation, refetch])

  const logout = useCallback(async () => {
    try {
      await logoutMutation()
      graphqlClient.removeAuthToken()
      if (typeof window !== 'undefined') {
        localStorage.removeItem('auth_token')
      }
    } catch (error) {
      console.error('Logout error:', error)
    }
  }, [logoutMutation])

  // Initialize auth on mount
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('auth_token')
      if (token) {
        graphqlClient.setAuthToken(token)
        refetch()
      }
    }
  }, [refetch])

  return {
    user: userData?.me,
    loading: userLoading || loginLoading,
    error: userError || loginError,
    login,
    logout,
    refetch
  }
}

export function useMyTasks(filters?: {
  status?: string
  priority?: string
  limit?: number
  offset?: number
}) {
  return useQuery(`
    query MyTasks($status: CourierTaskStatus, $priority: TaskPriority, $limit: Int, $offset: Int) {
      myTasks(status: $status, priority: $priority, limit: $limit, offset: $offset) {
        nodes {
          id
          letterCode
          senderName
          senderPhone
          recipientHint
          targetLocation
          currentLocation
          priority
          status
          estimatedTime
          distance
          deadline
          instructions
          reward
          createdAt
          letter {
            id
            title
            user {
              nickname
            }
          }
        }
        totalCount
        hasNextPage
      }
    }
  `, filters)
}

export function useMyLetters(filters?: {
  status?: string
  limit?: number
  offset?: number
}) {
  return useQuery(`
    query MyLetters($status: LetterStatus, $limit: Int, $offset: Int) {
      letters(status: $status, limit: $limit, offset: $offset) {
        nodes {
          id
          title
          content
          status
          style
          code
          schoolCode
          createdAt
          updatedAt
        }
        totalCount
        hasNextPage
      }
    }
  `, filters)
}

export function useCreateLetter() {
  return useMutation(`
    mutation CreateLetter($input: CreateLetterInput!) {
      createLetter(input: $input) {
        id
        title
        content
        status
        style
        schoolCode
        createdAt
      }
    }
  `)
}

export function useAcceptTask() {
  return useMutation(`
    mutation AcceptTask($taskId: ID!) {
      acceptTask(taskId: $taskId) {
        id
        status
        updatedAt
      }
    }
  `)
}

export function useUpdateTaskStatus() {
  return useMutation(`
    mutation UpdateTaskStatus($taskId: ID!, $status: CourierTaskStatus!, $location: String, $notes: String) {
      updateTaskStatus(taskId: $taskId, status: $status, location: $location, notes: $notes) {
        id
        status
        currentLocation
        updatedAt
        letter {
          id
          status
        }
      }
    }
  `)
}