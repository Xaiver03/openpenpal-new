/**
 * Dashboard Page Object Model
 * 仪表板页面对象模型
 */

import { Page, Locator, expect } from '@playwright/test'

export class DashboardPage {
  readonly page: Page
  readonly userAvatar: Locator
  readonly userMenu: Locator
  readonly logoutButton: Locator
  readonly navigationMenu: Locator
  readonly writeLetterButton: Locator
  readonly lettersList: Locator
  readonly userProfile: Locator

  constructor(page: Page) {
    this.page = page
    this.userAvatar = page.getByTestId('user-avatar')
    this.userMenu = page.getByTestId('user-menu')
    this.logoutButton = page.getByTestId('logout-button')
    this.navigationMenu = page.getByTestId('navigation-menu')
    this.writeLetterButton = page.getByTestId('write-letter-button')
    this.lettersList = page.getByTestId('letters-list')
    this.userProfile = page.getByTestId('user-profile')
  }

  async goto() {
    await this.page.goto('/dashboard')
    await this.page.waitForLoadState('networkidle')
  }

  async expectDashboardLoaded() {
    await expect(this.navigationMenu).toBeVisible()
    await expect(this.userAvatar).toBeVisible()
  }

  async openUserMenu() {
    await this.userAvatar.click()
    await expect(this.userMenu).toBeVisible()
  }

  async logout() {
    await this.openUserMenu()
    await this.logoutButton.click()
    await expect(this.page).toHaveURL('/login')
  }

  async clickWriteLetter() {
    await this.writeLetterButton.click()
    await expect(this.page).toHaveURL('/letters/write')
  }

  async expectUserInfo(username: string) {
    await this.openUserMenu()
    await expect(this.userProfile).toContainText(username)
  }

  async expectLettersSection() {
    await expect(this.lettersList).toBeVisible()
  }
}