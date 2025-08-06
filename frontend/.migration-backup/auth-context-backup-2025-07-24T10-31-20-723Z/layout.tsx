import type { Metadata, Viewport } from 'next'
import { Inter, Noto_Serif_SC } from 'next/font/google'
import './globals.css'
import { cn } from '@/lib/utils'
import { AuthProvider } from '@/contexts/auth-context'
import { WebSocketErrorBoundary, PageErrorBoundary } from '@/components/error-boundary'

// 动态加载WebSocketProvider以避免SSR问题
const WebSocketProvider = dynamic(
  () => import('@/contexts/websocket-context').then(mod => ({ default: mod.WebSocketProvider })),
  { 
    ssr: false,
    loading: () => <div className="hidden"></div> // 静默加载
  }
)
import { NotificationManager } from '@/components/realtime/notification-center'
import { LazyWrapper } from '@/components/optimization/performance-wrapper'
import dynamic from 'next/dynamic'

// 动态加载性能监控组件
const PerformanceMonitor = dynamic(
  () => import('@/components/optimization/performance-monitor').then(mod => ({ default: mod.PerformanceMonitor })),
  { ssr: false }
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
      </head>
      <body
        className={cn(
          'min-h-screen bg-letter-paper font-sans antialiased',
          inter.variable,
          notoSerifSC.variable
        )}
      >
        <PageErrorBoundary>
          <LazyWrapper enableLazyLoading={true}>
            <AuthProvider>
              <WebSocketErrorBoundary fallback={<div className="hidden"></div>}>
                <WebSocketProvider>
                  <div className="relative flex min-h-screen flex-col">
                    <main className="flex-1">{children}</main>
                    <NotificationManager />
                    <PerformanceMonitor />
                  </div>
                </WebSocketProvider>
              </WebSocketErrorBoundary>
            </AuthProvider>
          </LazyWrapper>
        </PageErrorBoundary>
      </body>
    </html>
  )
}