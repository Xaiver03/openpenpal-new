// Enhanced service worker with proper chunk handling
// Prevents timeout errors while maintaining basic SW functionality

console.log('⚡ Service Worker initialized - pass-through mode')

self.addEventListener('install', (event) => {
  console.log('SW: Install event - immediate activation')
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  console.log('SW: Activate event - clearing old caches')
  event.waitUntil(
    Promise.all([
      self.clients.claim(),
      caches.keys().then(cacheNames => {
        return Promise.all(
          cacheNames.map(cacheName => {
            console.log('SW: Deleting cache:', cacheName)
            return caches.delete(cacheName)
          })
        )
      })
    ])
  )
})

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url)
  
  // Always pass through requests - no caching or interception
  // This prevents chunk loading timeout issues
  if (url.pathname.includes('/_next/') || 
      url.pathname.includes('/api/') ||
      url.pathname.includes('.js') ||
      url.pathname.includes('.json')) {
    // Critical resources - immediate pass-through
    return
  }
  
  // Log non-critical fetches for debugging
  if (!url.pathname.includes('/_next/static/')) {
    console.log('SW: Fetch pass-through:', event.request.url)
  }
})

// Handle any errors in the service worker
self.addEventListener('error', (event) => {
  console.error('SW: Error occurred:', event.error)
})

// Ensure SW doesn't interfere with module loading
self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting()
  }
})