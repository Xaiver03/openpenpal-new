/**
 * Lighthouse CI Configuration
 * 性能测试配置
 */

module.exports = {
  ci: {
    collect: {
      startServerCommand: 'npm start',
      startServerReadyPattern: 'ready on',
      startServerReadyTimeout: 30000,
      url: [
        'http://localhost:3000',
        'http://localhost:3000/login',
        'http://localhost:3000/register',
        'http://localhost:3000/dashboard'
      ],
      numberOfRuns: 3
    },
    assert: {
      assertions: {
        'categories:performance': ['warn', { minScore: 0.8 }],
        'categories:accessibility': ['error', { minScore: 0.9 }],
        'categories:best-practices': ['warn', { minScore: 0.85 }],
        'categories:seo': ['warn', { minScore: 0.8 }],
        'categories:pwa': ['warn', { minScore: 0.7 }],
        
        // Core Web Vitals
        'first-contentful-paint': ['warn', { maxNumericValue: 2000 }],
        'largest-contentful-paint': ['warn', { maxNumericValue: 2500 }],
        'first-meaningful-paint': ['warn', { maxNumericValue: 2000 }],
        'speed-index': ['warn', { maxNumericValue: 3000 }],
        'interactive': ['warn', { maxNumericValue: 3000 }],
        'cumulative-layout-shift': ['warn', { maxNumericValue: 0.1 }],
        
        // Resource optimization
        'unused-javascript': ['warn', { maxNumericValue: 40000 }],
        'unused-css-rules': ['warn', { maxNumericValue: 20000 }],
        'render-blocking-resources': 'warn',
        'uses-responsive-images': 'warn',
        'uses-optimized-images': 'warn',
        'uses-webp-images': 'warn',
        'uses-text-compression': 'warn',
        'uses-rel-preconnect': 'warn',
        'uses-rel-preload': 'warn',
        
        // Security and best practices
        'is-on-https': 'error',
        'uses-http2': 'warn',
        'no-vulnerable-libraries': 'error',
        'csp-xss': 'warn',
        
        // Accessibility
        'color-contrast': 'error',
        'image-alt': 'error',
        'label': 'error',
        'aria-valid-attr': 'error',
        'button-name': 'error',
        'link-name': 'error'
      }
    },
    upload: {
      target: 'temporary-public-storage'
    },
    server: {
      port: 9001,
      storage: './lhci-data'
    }
  }
}