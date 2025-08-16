// Minimal service worker to override any auto-registration
// This prevents the 408 timeout errors from the previous sw.js

console.log('🚫 Service Worker disabled - no caching or network interception')

self.addEventListener('install', (event) => {
  console.log('SW: Install event - skipping waitUntil')
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  console.log('SW: Activate event - claiming clients')
  event.waitUntil(self.clients.claim())
})

self.addEventListener('fetch', (event) => {
  // Just pass through all requests without any caching or modification
  // This prevents the 408 timeout issues
  console.log('SW: Fetch pass-through:', event.request.url)
})