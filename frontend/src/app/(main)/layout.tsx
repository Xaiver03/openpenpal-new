import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'
import { MobileNavTabs, MobileFAB } from '@/components/mobile/mobile-nav-tabs'
import { EmergencyFAB, QuickScanButton } from '@/components/mobile/mobile-quick-actions'

export default function MainLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="min-h-screen flex flex-col">
      <Header />
      <main className="flex-1 pb-16 md:pb-0">
        {children}
      </main>
      <Footer />
      <MobileNavTabs />
      <MobileFAB />
      <EmergencyFAB />
      <QuickScanButton />
    </div>
  )
}