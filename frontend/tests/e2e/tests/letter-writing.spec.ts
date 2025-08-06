/**
 * Letter Writing E2E Tests
 * 写信功能端到端测试
 */

import { test, expect } from '@playwright/test'
import { LoginPage } from '../pages/login.page'
import { DashboardPage } from '../pages/dashboard.page'
import { TEST_CREDENTIALS } from '../fixtures/test-users'

test.describe('Letter Writing Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    await loginPage.expectLoginSuccess()
  })

  test('should navigate to letter writing page', async ({ page }) => {
    const dashboardPage = new DashboardPage(page)
    
    await dashboardPage.goto()
    await dashboardPage.clickWriteLetter()
    
    await expect(page).toHaveURL('/letters/write')
    await expect(page.getByTestId('letter-editor')).toBeVisible()
  })

  test('should create a new letter', async ({ page }) => {
    await page.goto('/letters/write')
    
    const titleInput = page.getByTestId('letter-title-input')
    const contentEditor = page.getByTestId('letter-content-editor')
    const styleSelector = page.getByTestId('letter-style-selector')
    const saveButton = page.getByTestId('save-letter-button')
    
    await titleInput.fill('测试信件标题')
    await contentEditor.fill('这是一封测试信件的内容。希望这封信能够成功保存和发送。')
    await styleSelector.selectOption('classic')
    
    await saveButton.click()
    
    // Should show success message
    await expect(page.getByTestId('success-message')).toBeVisible()
    await expect(page.getByTestId('success-message')).toContainText('保存成功')
  })

  test('should validate letter content', async ({ page }) => {
    await page.goto('/letters/write')
    
    const titleInput = page.getByTestId('letter-title-input')
    const contentEditor = page.getByTestId('letter-content-editor')
    const saveButton = page.getByTestId('save-letter-button')
    
    // Try to save empty letter
    await saveButton.click()
    
    // Should show validation errors
    await expect(page.getByTestId('title-error')).toBeVisible()
    await expect(page.getByTestId('content-error')).toBeVisible()
    
    // Fill with invalid data
    await titleInput.fill('x') // Too short
    await contentEditor.fill('short') // Too short
    await saveButton.click()
    
    await expect(page.getByTestId('title-error')).toContainText('标题长度')
    await expect(page.getByTestId('content-error')).toContainText('内容长度')
  })

  test('should save letter as draft', async ({ page }) => {
    await page.goto('/letters/write')
    
    const titleInput = page.getByTestId('letter-title-input')
    const contentEditor = page.getByTestId('letter-content-editor')
    const saveDraftButton = page.getByTestId('save-draft-button')
    
    await titleInput.fill('草稿信件')
    await contentEditor.fill('这是一封草稿信件的内容。')
    
    await saveDraftButton.click()
    
    await expect(page.getByTestId('draft-saved-message')).toBeVisible()
  })

  test('should preview letter before sending', async ({ page }) => {
    await page.goto('/letters/write')
    
    const titleInput = page.getByTestId('letter-title-input')
    const contentEditor = page.getByTestId('letter-content-editor')
    const previewButton = page.getByTestId('preview-button')
    
    const title = '预览测试信件'
    const content = '这是预览功能的测试内容。'
    
    await titleInput.fill(title)
    await contentEditor.fill(content)
    
    await previewButton.click()
    
    // Should open preview modal or page
    const previewModal = page.getByTestId('letter-preview-modal')
    await expect(previewModal).toBeVisible()
    await expect(previewModal.getByTestId('preview-title')).toContainText(title)
    await expect(previewModal.getByTestId('preview-content')).toContainText(content)
  })

  test('should handle letter submission', async ({ page }) => {
    await page.goto('/letters/write')
    
    const titleInput = page.getByTestId('letter-title-input')
    const contentEditor = page.getByTestId('letter-content-editor')
    const recipientInput = page.getByTestId('recipient-input')
    const submitButton = page.getByTestId('submit-letter-button')
    
    await titleInput.fill('提交测试信件')
    await contentEditor.fill('这是一封准备提交的测试信件内容。希望能够成功发送到收件人。')
    await recipientInput.fill('recipient@example.com')
    
    // Monitor network request
    const submitPromise = page.waitForRequest(request => 
      request.url().includes('/api/letters') && request.method() === 'POST'
    )
    
    await submitButton.click()
    
    const submitRequest = await submitPromise
    expect(submitRequest).toBeTruthy()
    
    // Should show success message and redirect
    await expect(page.getByTestId('submission-success')).toBeVisible()
    await expect(page).toHaveURL(/\/letters\/success|\/dashboard/)
  })

  test('should support different letter styles', async ({ page }) => {
    await page.goto('/letters/write')
    
    const styleSelector = page.getByTestId('letter-style-selector')
    const letterPreview = page.getByTestId('letter-style-preview')
    
    // Test different styles
    const styles = ['classic', 'modern', 'elegant', 'casual']
    
    for (const style of styles) {
      await styleSelector.selectOption(style)
      await expect(letterPreview).toHaveAttribute('data-style', style)
      
      // Verify style-specific elements are visible
      await expect(page.getByTestId(`style-${style}-elements`)).toBeVisible()
    }
  })

  test('should auto-save letter content', async ({ page }) => {
    await page.goto('/letters/write')
    
    const titleInput = page.getByTestId('letter-title-input')
    const contentEditor = page.getByTestId('letter-content-editor')
    const autoSaveIndicator = page.getByTestId('auto-save-indicator')
    
    await titleInput.fill('自动保存测试')
    
    // Wait for auto-save to trigger
    await expect(autoSaveIndicator).toContainText('正在保存')
    await expect(autoSaveIndicator).toContainText('已保存')
    
    await contentEditor.fill('测试自动保存功能的内容。')
    
    // Should auto-save after typing
    await page.waitForTimeout(2000) // Wait for auto-save delay
    await expect(autoSaveIndicator).toContainText('已保存')
  })
})

test.describe('Letter Management', () => {
  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    await loginPage.expectLoginSuccess()
  })

  test('should display user letters list', async ({ page }) => {
    await page.goto('/letters')
    
    const lettersList = page.getByTestId('letters-list')
    const lettersTable = page.getByTestId('letters-table')
    
    await expect(lettersList).toBeVisible()
    await expect(lettersTable).toBeVisible()
    
    // Should have column headers
    await expect(page.getByText('标题')).toBeVisible()
    await expect(page.getByText('状态')).toBeVisible()
    await expect(page.getByText('创建时间')).toBeVisible()
  })

  test('should filter letters by status', async ({ page }) => {
    await page.goto('/letters')
    
    const statusFilter = page.getByTestId('status-filter')
    const lettersList = page.getByTestId('letters-list')
    
    // Filter by draft
    await statusFilter.selectOption('draft')
    await page.waitForLoadState('networkidle')
    
    // All visible letters should be drafts
    const letterItems = lettersList.getByTestId('letter-item')
    const count = await letterItems.count()
    
    for (let i = 0; i < count; i++) {
      const status = letterItems.nth(i).getByTestId('letter-status')
      await expect(status).toContainText('草稿')
    }
  })

  test('should search letters by title', async ({ page }) => {
    await page.goto('/letters')
    
    const searchInput = page.getByTestId('letters-search-input')
    const lettersList = page.getByTestId('letters-list')
    
    await searchInput.fill('测试')
    await page.waitForLoadState('networkidle')
    
    // All visible letters should contain the search term
    const letterTitles = lettersList.getByTestId('letter-title')
    const count = await letterTitles.count()
    
    for (let i = 0; i < count; i++) {
      const title = await letterTitles.nth(i).textContent()
      expect(title).toContain('测试')
    }
  })
})