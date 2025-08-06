/**
 * Playwright Global Teardown
 * 全局测试环境清理
 */

import { FullConfig } from '@playwright/test'

async function globalTeardown(config: FullConfig) {
  console.log('🧹 Cleaning up E2E test environment...')
  
  // Clean up test data if needed
  // This would typically involve database cleanup, file cleanup, etc.
  
  console.log('✅ E2E test environment cleanup complete')
}

export default globalTeardown