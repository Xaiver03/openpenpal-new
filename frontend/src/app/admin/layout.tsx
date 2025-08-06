'use client'

import { ProtectedRoute } from '@/components/auth/protected-route'

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <ProtectedRoute requiredRole={['school_admin', 'platform_admin', 'super_admin']}>
      {children}
    </ProtectedRoute>
  )
}