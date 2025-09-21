import { Page, expect } from '@playwright/test';
import { testUsers } from '../fixtures/test-data';

export class AuthHelper {
  constructor(private page: Page) {}

  async register(user = testUsers.validUser) {
    await this.page.goto('/register');
    
    await this.page.fill('[data-testid="name-input"]', user.name);
    await this.page.fill('[data-testid="email-input"]', user.email);
    await this.page.fill('[data-testid="password-input"]', user.password);
    await this.page.fill('[data-testid="confirm-password-input"]', user.password);
    
    await this.page.click('[data-testid="register-button"]');
    
    // Wait for successful registration
    await expect(this.page).toHaveURL('/dashboard');
  }

  async login(user = testUsers.validUser) {
    await this.page.goto('/login');
    
    await this.page.fill('[data-testid="email-input"]', user.email);
    await this.page.fill('[data-testid="password-input"]', user.password);
    
    await this.page.click('[data-testid="login-button"]');
    
    // Wait for successful login
    await expect(this.page).toHaveURL('/dashboard');
  }

  async logout() {
    // Click on user menu
    await this.page.click('[data-testid="user-menu-button"]');
    
    // Click logout option
    await this.page.click('[data-testid="logout-button"]');
    
    // Should redirect to login page
    await expect(this.page).toHaveURL('/login');
  }

  async isLoggedIn(): Promise<boolean> {
    try {
      // Check if we're on a protected route and not redirected to login
      await this.page.waitForURL('/dashboard', { timeout: 5000 });
      return true;
    } catch {
      return false;
    }
  }

  async ensureLoggedIn(user = testUsers.validUser) {
    if (!(await this.isLoggedIn())) {
      await this.login(user);
    }
  }

  async ensureLoggedOut() {
    if (await this.isLoggedIn()) {
      await this.logout();
    }
  }
}
