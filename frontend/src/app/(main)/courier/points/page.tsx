'use client'

import { CourierCreditDashboard } from '@/components/courier/courier-credit-dashboard'

export default function CourierPointsPage() {
  return (
    <div className="min-h-screen bg-amber-50">
      <div className="container max-w-6xl mx-auto px-4 py-8">
        <CourierCreditDashboard />
      </div>
    </div>
  )
}