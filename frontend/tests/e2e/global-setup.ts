/**
 * Playwright Global Setup
 * 全局测试环境设置
 */

import { chromium, FullConfig } from '@playwright/test'
import { TestDataManager } from '@/lib/auth/test-data-manager'

async function globalSetup(config: FullConfig) {
  console.log('🚀 Setting up E2E test environment...')

  // Initialize test data
  await TestDataManager.initializeTestUsers()
  
  // Create a browser instance for setup
  const browser = await chromium.launch()
  const context = await browser.newContext()
  const page = await context.newPage()

  // Pre-warm the application
  try {
    await page.goto('http://localhost:3000')
    await page.waitForLoadState('networkidle', { timeout: 30000 })
    console.log('✅ Application pre-warmed successfully')
  } catch (error) {
    console.warn('⚠️ Application pre-warm failed:', error)
  }

  await browser.close()
  console.log('✅ E2E test environment setup complete')
}

export default globalSetup