/**
 * Responsive Design E2E Tests
 * 响应式设计端到端测试
 */

import { test, expect, devices } from '@playwright/test'
import { LoginPage } from '../pages/login.page'
import { DashboardPage } from '../pages/dashboard.page'
import { TEST_CREDENTIALS } from '../fixtures/test-users'

const mobileViewport = devices['iPhone 12'].viewport
const tabletViewport = { width: 768, height: 1024 }
const desktopViewport = { width: 1920, height: 1080 }

test.describe('Responsive Design Tests', () => {
  test('should display mobile navigation correctly', async ({ page }) => {
    await page.setViewportSize(mobileViewport!)
    
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    
    const dashboardPage = new DashboardPage(page)
    await dashboardPage.goto()
    
    // Mobile navigation should be visible
    const mobileMenu = page.getByTestId('mobile-menu-button')
    const mobileNav = page.getByTestId('mobile-navigation')
    
    await expect(mobileMenu).toBeVisible()
    
    // Open mobile menu
    await mobileMenu.click()
    await expect(mobileNav).toBeVisible()
    
    // Check navigation items
    await expect(mobileNav.getByTestId('nav-dashboard')).toBeVisible()
    await expect(mobileNav.getByTestId('nav-letters')).toBeVisible()
    await expect(mobileNav.getByTestId('nav-profile')).toBeVisible()
  })

  test('should adapt login form for mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport!)
    
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    
    // Form should be full width on mobile
    const loginForm = page.getByTestId('login-form')
    const formBounds = await loginForm.boundingBox()
    const viewportWidth = mobileViewport!.width
    
    expect(formBounds?.width).toBeGreaterThan(viewportWidth * 0.8)
    
    // Buttons should be full width
    const loginButton = loginPage.loginButton
    const buttonBounds = await loginButton.boundingBox()
    expect(buttonBounds?.width).toBeGreaterThan(viewportWidth * 0.8)
  })

  test('should display tablet layout correctly', async ({ page }) => {
    await page.setViewportSize(tabletViewport)
    
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    
    const dashboardPage = new DashboardPage(page)
    await dashboardPage.goto()
    
    // Tablet should show sidebar navigation
    const sidebarNav = page.getByTestId('sidebar-navigation')
    await expect(sidebarNav).toBeVisible()
    
    // Content area should be properly sized
    const contentArea = page.getByTestId('main-content')
    const contentBounds = await contentArea.boundingBox()
    
    expect(contentBounds?.width).toBeLessThan(tabletViewport.width)
    expect(contentBounds?.width).toBeGreaterThan(tabletViewport.width * 0.6)
  })

  test('should handle desktop layout', async ({ page }) => {
    await page.setViewportSize(desktopViewport)
    
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    
    const dashboardPage = new DashboardPage(page)
    await dashboardPage.goto()
    
    // Desktop should show full navigation
    const desktopNav = page.getByTestId('desktop-navigation')
    await expect(desktopNav).toBeVisible()
    
    // Should have multiple columns layout
    const leftColumn = page.getByTestId('left-column')
    const rightColumn = page.getByTestId('right-column')
    
    await expect(leftColumn).toBeVisible()
    await expect(rightColumn).toBeVisible()
  })

  test('should adapt letter writing interface for mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport!)
    
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    
    await page.goto('/letters/write')
    
    // Mobile editor should be full screen
    const editor = page.getByTestId('letter-editor')
    const editorBounds = await editor.boundingBox()
    
    expect(editorBounds?.width).toBeGreaterThan(mobileViewport!.width * 0.9)
    
    // Mobile toolbar should be scrollable
    const toolbar = page.getByTestId('editor-toolbar')
    await expect(toolbar).toHaveCSS('overflow-x', 'auto')
  })

  test('should handle orientation changes on mobile', async ({ page }) => {
    // Portrait
    await page.setViewportSize({ width: 375, height: 667 })
    
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    
    const portraitForm = page.getByTestId('login-form')
    const portraitBounds = await portraitForm.boundingBox()
    
    // Landscape
    await page.setViewportSize({ width: 667, height: 375 })
    await page.waitForTimeout(500) // Allow for reflow
    
    const landscapeForm = page.getByTestId('login-form')
    const landscapeBounds = await landscapeForm.boundingBox()
    
    // Form should adapt to landscape
    expect(landscapeBounds?.width).toBeGreaterThan(portraitBounds?.width!)
    expect(landscapeBounds?.height).toBeLessThan(portraitBounds?.height!)
  })

  test('should ensure touch targets are accessible on mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport!)
    
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    
    const dashboardPage = new DashboardPage(page)
    await dashboardPage.goto()
    
    // Check that interactive elements meet minimum touch target size (44px)
    const buttons = page.getByRole('button')
    const buttonCount = await buttons.count()
    
    for (let i = 0; i < buttonCount; i++) {
      const button = buttons.nth(i)
      const bounds = await button.boundingBox()
      
      if (bounds) {
        expect(bounds.height).toBeGreaterThanOrEqual(44)
        expect(bounds.width).toBeGreaterThanOrEqual(44)
      }
    }
  })

  test('should handle text scaling on mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport!)
    
    // Simulate larger text size
    await page.addStyleTag({
      content: `
        * {
          font-size: 1.2em !important;
        }
      `
    })
    
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    
    // Form should still be usable with larger text
    await expect(loginPage.usernameInput).toBeVisible()
    await expect(loginPage.passwordInput).toBeVisible()
    await expect(loginPage.loginButton).toBeVisible()
    
    // Elements should not overlap
    const usernameBox = await loginPage.usernameInput.boundingBox()
    const passwordBox = await loginPage.passwordInput.boundingBox()
    
    expect(passwordBox?.y).toBeGreaterThan((usernameBox?.y || 0) + (usernameBox?.height || 0))
  })

  test('should support landscape mode letter writing', async ({ page }) => {
    await page.setViewportSize({ width: 667, height: 375 }) // iPhone landscape
    
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    
    await page.goto('/letters/write')
    
    // In landscape, should have horizontal layout
    const titleInput = page.getByTestId('letter-title-input')
    const contentEditor = page.getByTestId('letter-content-editor')
    const toolbar = page.getByTestId('editor-toolbar')
    
    await expect(titleInput).toBeVisible()
    await expect(contentEditor).toBeVisible()
    await expect(toolbar).toBeVisible()
    
    // Editor should use available space efficiently
    const editorBounds = await contentEditor.boundingBox()
    expect(editorBounds?.width).toBeGreaterThan(400)
    expect(editorBounds?.height).toBeGreaterThan(200)
  })
})