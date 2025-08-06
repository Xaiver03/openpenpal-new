# API Client Migration Guide

## Overview

We are standardizing our API client implementation to use the new `api-client.ts` instead of the old `api.ts`. The new client provides better error handling, automatic retry logic, CSRF protection, and proper TypeScript types.

## Migration Steps

### 1. Update Imports

```typescript
// Old
import api from '@/lib/api'
import { getAuthToken, setAuthToken } from '@/lib/api'

// New
import { apiClient, TokenManager } from '@/lib/api-client'
```

### 2. Update API Calls

#### Authentication

```typescript
// Old
const result = await api.login(username, password)
if (result.success) {
  setAuthToken(result.data.token)
}

// New
const response = await apiClient.post('/auth/login', { username, password })
if (response.code === 0) {
  TokenManager.set(response.data.token)
}
```

#### User Profile

```typescript
// Old
const result = await api.getProfile()
const user = result.data

// New
const response = await apiClient.get('/users/me')
const user = response.data
```

#### Letters

```typescript
// Old
const result = await api.createLetter({
  content: 'Hello',
  style: 'classic'
})

// New
const response = await apiClient.post('/letters', {
  content: 'Hello',
  style: 'classic'
})
```

### 3. Update Response Handling

The new API client uses a standardized response format:

```typescript
interface StandardApiResponse<T> {
  code: number      // 0 for success, HTTP status for errors
  message: string
  data: T | null
  timestamp: string
}
```

```typescript
// Old
if (result.success) {
  // handle success
} else {
  // handle error: result.error
}

// New
if (response.code === 0) {
  // handle success
} else {
  // handle error: response.message
}
```

### 4. Error Handling

The new client throws `ApiError` objects with detailed information:

```typescript
try {
  const response = await apiClient.get('/users/me')
  // handle success
} catch (error) {
  if (error instanceof ApiError) {
    console.error(`API Error ${error.status}: ${error.message}`)
    // Handle specific error codes
    if (error.status === 401) {
      // Redirect to login
    }
  }
}
```

### 5. WebSocket Connection

```typescript
// Old
const ws = new WebSocket('ws://localhost:8080/ws')

// New
import { WebSocketManager } from '@/lib/api-client'
const wsManager = new WebSocketManager()
wsManager.connect()
```

### 6. Microservice APIs

The new client supports direct calls to microservices:

```typescript
// Write Service
const response = await apiClient.post('/letters/draft', data, {
  service: 'WRITE'
})

// Courier Service
const response = await apiClient.get('/courier/tasks', {
  service: 'COURIER'
})

// Admin Service
const response = await apiClient.get('/admin/users', {
  service: 'ADMIN'
})
```

## Features of New API Client

### 1. Automatic Token Management
- Tokens are automatically included in requests
- Token expiry is checked before requests
- Automatic token refresh (if implemented)

### 2. CSRF Protection
- CSRF tokens are automatically managed
- Double-submit cookie pattern

### 3. Retry Logic
- Failed requests are automatically retried (configurable)
- Exponential backoff for rate limiting

### 4. Request/Response Interceptors
- Automatic request transformation
- Response normalization
- Global error handling

### 5. TypeScript Support
- Full type safety for requests and responses
- Generic types for API responses
- Service-specific types

## Common Patterns

### Using with React Query

```typescript
import { useQuery, useMutation } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'

// Query
const { data, isLoading } = useQuery({
  queryKey: ['user', 'profile'],
  queryFn: async () => {
    const response = await apiClient.get('/users/me')
    return response.data
  }
})

// Mutation
const mutation = useMutation({
  mutationFn: (data: CreateLetterData) => apiClient.post('/letters', data),
  onSuccess: (response) => {
    if (response.code === 0) {
      // Handle success
    }
  }
})
```

### Handling Loading States

```typescript
const [isLoading, setIsLoading] = useState(false)

const fetchData = async () => {
  setIsLoading(true)
  try {
    const response = await apiClient.get('/letters')
    // Handle data
  } catch (error) {
    // Handle error
  } finally {
    setIsLoading(false)
  }
}
```

## Deprecation Timeline

1. **Phase 1 (Current)**: Both APIs work, old API shows console warnings
2. **Phase 2 (v1.1)**: Old API marked as deprecated in TypeScript
3. **Phase 3 (v2.0)**: Old API removed completely

## Need Help?

If you encounter any issues during migration:
1. Check the console for deprecation warnings
2. Refer to the TypeScript types for correct usage
3. See `src/lib/api-client.ts` for implementation details