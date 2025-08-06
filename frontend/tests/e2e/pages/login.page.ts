/**
 * Login Page Object Model
 * 登录页面对象模型
 */

import { Page, Locator, expect } from '@playwright/test'

export class LoginPage {
  readonly page: Page
  readonly usernameInput: Locator
  readonly passwordInput: Locator
  readonly loginButton: Locator
  readonly errorMessage: Locator
  readonly forgotPasswordLink: Locator
  readonly registerLink: Locator

  constructor(page: Page) {
    this.page = page
    this.usernameInput = page.getByTestId('username-input')
    this.passwordInput = page.getByTestId('password-input')
    this.loginButton = page.getByTestId('login-button')
    this.errorMessage = page.getByTestId('error-message')
    this.forgotPasswordLink = page.getByTestId('forgot-password-link')
    this.registerLink = page.getByTestId('register-link')
  }

  async goto() {
    await this.page.goto('/login')
    await this.page.waitForLoadState('networkidle')
  }

  async login(username: string, password: string) {
    await this.usernameInput.fill(username)
    await this.passwordInput.fill(password)
    await this.loginButton.click()
  }

  async expectLoginSuccess() {
    // Should redirect to dashboard or home page
    await expect(this.page).toHaveURL(/\/(dashboard|home|$)/)
    await this.page.waitForLoadState('networkidle')
  }

  async expectLoginError(message?: string) {
    await expect(this.errorMessage).toBeVisible()
    if (message) {
      await expect(this.errorMessage).toContainText(message)
    }
  }

  async expectFormValidation() {
    // Check that form validation is working
    await expect(this.usernameInput).toHaveAttribute('required')
    await expect(this.passwordInput).toHaveAttribute('required')
  }

  async clickRegisterLink() {
    await this.registerLink.click()
    await expect(this.page).toHaveURL('/register')
  }

  async clickForgotPasswordLink() {
    await this.forgotPasswordLink.click()
    await expect(this.page).toHaveURL('/forgot-password')
  }
}