/**
 * Authentication E2E Tests
 * 认证端到端测试
 */

import { test, expect } from '@playwright/test'
import { LoginPage } from '../pages/login.page'
import { DashboardPage } from '../pages/dashboard.page'
import { TEST_CREDENTIALS } from '../fixtures/test-users'

test.describe('Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Ensure we start from a clean state
    await page.context().clearCookies()
    await page.context().clearPermissions()
  })

  test('should display login page correctly', async ({ page }) => {
    const loginPage = new LoginPage(page)
    await loginPage.goto()

    await expect(page).toHaveTitle(/登录|Login/)
    await expect(loginPage.usernameInput).toBeVisible()
    await expect(loginPage.passwordInput).toBeVisible()
    await expect(loginPage.loginButton).toBeVisible()
    await expect(loginPage.registerLink).toBeVisible()
  })

  test('should login with valid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page)
    const dashboardPage = new DashboardPage(page)

    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    
    await loginPage.expectLoginSuccess()
    await dashboardPage.expectDashboardLoaded()
    await dashboardPage.expectUserInfo(TEST_CREDENTIALS.valid.username)
  })

  test('should show error with invalid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page)

    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.invalid.username, TEST_CREDENTIALS.invalid.password)
    
    await loginPage.expectLoginError('用户名或密码错误')
    await expect(page).toHaveURL('/login')
  })

  test('should validate required fields', async ({ page }) => {
    const loginPage = new LoginPage(page)

    await loginPage.goto()
    await loginPage.expectFormValidation()
    
    // Try to submit with empty fields
    await loginPage.loginButton.click()
    
    // Should show validation errors or prevent submission
    await expect(page).toHaveURL('/login')
  })

  test('should navigate to register page', async ({ page }) => {
    const loginPage = new LoginPage(page)

    await loginPage.goto()
    await loginPage.clickRegisterLink()
    
    await expect(page).toHaveURL('/register')
  })

  test('should logout successfully', async ({ page }) => {
    const loginPage = new LoginPage(page)
    const dashboardPage = new DashboardPage(page)

    // Login first
    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    await loginPage.expectLoginSuccess()

    // Logout
    await dashboardPage.logout()
    
    await expect(page).toHaveURL('/login')
  })

  test('should redirect to login when accessing protected route', async ({ page }) => {
    // Try to access dashboard without logging in
    await page.goto('/dashboard')
    
    // Should redirect to login
    await expect(page).toHaveURL('/login')
  })

  test('should persist login state on page refresh', async ({ page }) => {
    const loginPage = new LoginPage(page)
    const dashboardPage = new DashboardPage(page)

    // Login
    await loginPage.goto()
    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    await loginPage.expectLoginSuccess()

    // Refresh page
    await page.reload()
    await dashboardPage.expectDashboardLoaded()
  })
})

test.describe('Security Tests', () => {
  test('should handle CSRF protection', async ({ page }) => {
    const loginPage = new LoginPage(page)

    await loginPage.goto()
    
    // Intercept login request to check CSRF token
    const loginPromise = page.waitForRequest(request => 
      request.url().includes('/api/auth/login') && request.method() === 'POST'
    )

    await loginPage.login(TEST_CREDENTIALS.valid.username, TEST_CREDENTIALS.valid.password)
    
    const loginRequest = await loginPromise
    const headers = loginRequest.headers()
    
    // Should have CSRF token in headers
    expect(headers['x-csrf-token']).toBeDefined()
  })

  test('should implement rate limiting', async ({ page }) => {
    const loginPage = new LoginPage(page)
    await loginPage.goto()

    // Make multiple rapid login attempts
    const attempts = []
    for (let i = 0; i < 6; i++) {
      attempts.push(
        loginPage.login(TEST_CREDENTIALS.invalid.username, TEST_CREDENTIALS.invalid.password)
      )
      await page.waitForTimeout(100) // Small delay between attempts
    }

    await Promise.all(attempts)

    // Should eventually show rate limit error
    const errorText = await loginPage.errorMessage.textContent()
    expect(errorText).toMatch(/限制|limit|too many/i)
  })
})