import type { Metadata, Viewport } from 'next'
import { Inter, Noto_Serif_SC } from 'next/font/google'
import './globals.css'
import { cn } from '@/lib/utils'
import { AuthProvider } from '@/contexts/auth-context-new'
import { AuthInitializer } from '@/components/providers/auth-initializer'
import { WebSocketErrorBoundary, PageErrorBoundary } from '@/components/error-boundary'
import { QueryProvider } from '@/components/providers/query-provider'
import { LazyWrapper } from '@/components/optimization/performance-wrapper'
import { TokenRefreshProvider } from '@/components/providers/token-refresh-provider'
import { TokenProvider } from '@/contexts/token-context'
import { UserInitializer } from '@/components/providers/user-initializer'
import { GlobalErrorBoundary } from '@/components/error-boundary/global-error-boundary'
import { PerformanceTracker } from '@/components/performance/performance-tracker'
import { preloadCriticalRoutes } from '@/lib/utils/lazy-loader'
import dynamic from 'next/dynamic'

const ClientBoundary = dynamic(
  () => import('@/components/providers/client-boundary').then(mod => ({ default: mod.ClientBoundary })),
  { 
    ssr: false,
    loading: () => <div className="relative flex min-h-screen flex-col"><main className="flex-1"></main></div>
  }
)

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-sans',
})

const notoSerifSC = Noto_Serif_SC({
  subsets: ['latin'],
  variable: '--font-serif',
  weight: ['400', '500', '600', '700'],
})

export const metadata: Metadata = {
  title: 'OpenPenPal 信使计划',
  description: '实体手写信 + 数字跟踪平台，重建校园社群的温度感知与精神连接',
  keywords: ['信件', '手写', '校园', '社交', '信使'],
  authors: [{ name: 'OpenPenPal Team' }],
  robots: 'index, follow',
  openGraph: {
    title: 'OpenPenPal 信使计划',
    description: '实体手写信 + 数字跟踪平台，重建校园社群的温度感知与精神连接',
    type: 'website',
    locale: 'zh_CN',
  },
  verification: {
    google: process.env.NEXT_PUBLIC_GOOGLE_SITE_VERIFICATION,
  },
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  themeColor: '#fefcf7',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="" />
        <link rel="dns-prefetch" href="//fonts.googleapis.com" />
        <meta name="format-detection" content="telephone=no" />
        {process.env.NODE_ENV === 'development' && (
          <script src="/debug-404.js" defer />
        )}
      </head>
      <body
        className={cn(
          'min-h-screen bg-letter-paper font-sans antialiased',
          inter.variable,
          notoSerifSC.variable
        )}
      >
        <GlobalErrorBoundary level="global">
          <PageErrorBoundary>
            <LazyWrapper enableLazyLoading={true}>
              <QueryProvider>
                <TokenProvider>
                  <AuthProvider>
                    <AuthInitializer>
                      <UserInitializer>
                        <TokenRefreshProvider>
                          <WebSocketErrorBoundary fallback={<div className="hidden"></div>}>
                            <ClientBoundary>
                              <PerformanceTracker 
                                pageName="app-root" 
                                trackInteractions={true}
                                reportInterval={60000}
                              />
                              {children}
                            </ClientBoundary>
                          </WebSocketErrorBoundary>
                        </TokenRefreshProvider>
                      </UserInitializer>
                    </AuthInitializer>
                  </AuthProvider>
                </TokenProvider>
              </QueryProvider>
            </LazyWrapper>
          </PageErrorBoundary>
        </GlobalErrorBoundary>
      </body>
    </html>
  )
}