'use client'

/**
 * Privacy Settings Page - Complete privacy control interface
 * 隐私设置页面 - 完整的隐私控制界面
 */

import React from 'react'
import { Container } from '@/components/ui/container'
import { PageHeader } from '@/components/ui/page-header'
import { PrivacySettings } from '@/components/profile/privacy-settings'

export default function PrivacySettingsPage() {
  return (
    <Container className="py-8">
      <PageHeader
        title="隐私设置"
        description="管理您的隐私偏好，控制谁可以看到您的个人资料和与您互动"
        className="mb-8"
      />
      
      <div className="max-w-4xl mx-auto">
        <PrivacySettings />
      </div>
    </Container>
  )
}