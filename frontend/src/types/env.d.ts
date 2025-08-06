declare namespace NodeJS {
  interface ProcessEnv {
    // API URLs
    NEXT_PUBLIC_API_URL: string
    NEXT_PUBLIC_WS_URL: string
    NEXT_PUBLIC_GATEWAY_URL?: string
    
    // Microservice URLs
    NEXT_PUBLIC_WRITE_SERVICE_URL?: string
    NEXT_PUBLIC_COURIER_SERVICE_URL?: string
    NEXT_PUBLIC_ADMIN_SERVICE_URL?: string
    NEXT_PUBLIC_OCR_SERVICE_URL?: string
    
    // App Config
    NEXT_PUBLIC_APP_NAME: string
    NEXT_PUBLIC_APP_VERSION?: string
    NEXT_PUBLIC_ENVIRONMENT: 'development' | 'production' | 'test'
    
    // Feature Flags
    NEXT_PUBLIC_ENABLE_DEBUG?: string
    NEXT_PUBLIC_ENABLE_ANALYTICS?: string
    NEXT_PUBLIC_ENABLE_PWA?: string
    
    // Third Party Services
    NEXT_PUBLIC_GA_ID?: string
    NEXT_PUBLIC_SENTRY_DSN?: string
    NEXT_PUBLIC_CLARITY_ID?: string
    NEXT_PUBLIC_GOOGLE_SITE_VERIFICATION?: string
    
    // Build Time Variables
    NODE_ENV: 'development' | 'production' | 'test'
    ANALYZE?: string
  }
}