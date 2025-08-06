# Frontend API Proxy Fix

## The Problem
The frontend was making API calls to itself (localhost:3000) instead of the backend services, causing 404 errors.

## Root Cause
The `apiClient` was initialized with an empty base URL:
```javascript
export const apiClient = new ApiClient('')
```

This made all API calls relative to the current domain (frontend), not the backend.

## Solution
Added Next.js rewrites configuration to proxy API requests:

```javascript
// next.config.js
async rewrites() {
  return [
    {
      source: '/api/:path*',
      destination: 'http://localhost:8000/api/:path*',
    },
  ];
},
```

## How It Works
1. Frontend makes request to `/api/v1/museum/entries`
2. Next.js intercepts and rewrites to `http://localhost:8000/api/v1/museum/entries`
3. Gateway (port 8000) receives request and forwards to appropriate backend service
4. Response flows back through the same path

## Result
✅ All API calls now properly routed through the gateway
✅ No more 404 errors in the browser console
✅ Museum, AI, and Letter features working correctly

## Note
The frontend server may need to be restarted for the configuration change to take effect:
```bash
# In the frontend terminal, press Ctrl+C then:
cd frontend && npm run dev
```