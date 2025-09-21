import { test, expect } from '@playwright/test';
import { AuthHelper } from './utils/auth-helpers';
import { testUsers } from './fixtures/test-data';

test.describe('Authentication Flow', () => {
  let authHelper: AuthHelper;

  test.beforeEach(async ({ page }) => {
    authHelper = new AuthHelper(page);
  });

  test.describe('User Registration', () => {
    test('should register a new user successfully', async ({ page }) => {
      await page.goto('/register');

      // Check page elements
      await expect(page.locator('h1')).toContainText('Register');
      await expect(page.locator('[data-testid="name-input"]')).toBeVisible();
      await expect(page.locator('[data-testid="email-input"]')).toBeVisible();
      await expect(page.locator('[data-testid="password-input"]')).toBeVisible();
      await expect(page.locator('[data-testid="confirm-password-input"]')).toBeVisible();

      // Fill registration form
      await page.fill('[data-testid="name-input"]', testUsers.validUser.name);
      await page.fill('[data-testid="email-input"]', testUsers.validUser.email);
      await page.fill('[data-testid="password-input"]', testUsers.validUser.password);
      await page.fill('[data-testid="confirm-password-input"]', testUsers.validUser.password);

      // Submit form
      await page.click('[data-testid="register-button"]');

      // Should redirect to dashboard
      await expect(page).toHaveURL('/dashboard');
      await expect(page.locator('[data-testid="user-name"]')).toContainText(testUsers.validUser.name);
    });

    test('should show validation errors for invalid input', async ({ page }) => {
      await page.goto('/register');

      // Try to submit empty form
      await page.click('[data-testid="register-button"]');

      // Check for validation errors
      await expect(page.locator('[data-testid="name-error"]')).toBeVisible();
      await expect(page.locator('[data-testid="email-error"]')).toBeVisible();
      await expect(page.locator('[data-testid="password-error"]')).toBeVisible();
    });

    test('should show error for mismatched passwords', async ({ page }) => {
      await page.goto('/register');

      await page.fill('[data-testid="name-input"]', testUsers.validUser.name);
      await page.fill('[data-testid="email-input"]', testUsers.validUser.email);
      await page.fill('[data-testid="password-input"]', testUsers.validUser.password);
      await page.fill('[data-testid="confirm-password-input"]', 'DifferentPassword123!');

      await page.click('[data-testid="register-button"]');

      await expect(page.locator('[data-testid="confirm-password-error"]')).toContainText('Passwords do not match');
    });

    test('should show error for weak password', async ({ page }) => {
      await page.goto('/register');

      await page.fill('[data-testid="name-input"]', testUsers.validUser.name);
      await page.fill('[data-testid="email-input"]', testUsers.validUser.email);
      await page.fill('[data-testid="password-input"]', '123');
      await page.fill('[data-testid="confirm-password-input"]', '123');

      await page.click('[data-testid="register-button"]');

      await expect(page.locator('[data-testid="password-error"]')).toBeVisible();
    });

    test('should show error for duplicate email', async ({ page }) => {
      // First registration
      await authHelper.register(testUsers.validUser);
      await authHelper.logout();

      // Try to register with same email
      await page.goto('/register');
      await page.fill('[data-testid="name-input"]', 'Another User');
      await page.fill('[data-testid="email-input"]', testUsers.validUser.email);
      await page.fill('[data-testid="password-input"]', 'AnotherPassword123!');
      await page.fill('[data-testid="confirm-password-input"]', 'AnotherPassword123!');

      await page.click('[data-testid="register-button"]');

      await expect(page.locator('[data-testid="form-error"]')).toContainText('Email already exists');
    });
  });

  test.describe('User Login', () => {
    test.beforeEach(async () => {
      // Register a user for login tests
      await authHelper.register(testUsers.validUser);
      await authHelper.logout();
    });

    test('should login with valid credentials', async ({ page }) => {
      await page.goto('/login');

      // Check page elements
      await expect(page.locator('h1')).toContainText('Login');
      await expect(page.locator('[data-testid="email-input"]')).toBeVisible();
      await expect(page.locator('[data-testid="password-input"]')).toBeVisible();

      // Login
      await page.fill('[data-testid="email-input"]', testUsers.validUser.email);
      await page.fill('[data-testid="password-input"]', testUsers.validUser.password);
      await page.click('[data-testid="login-button"]');

      // Should redirect to dashboard
      await expect(page).toHaveURL('/dashboard');
      await expect(page.locator('[data-testid="user-name"]')).toContainText(testUsers.validUser.name);
    });

    test('should show error for invalid credentials', async ({ page }) => {
      await page.goto('/login');

      await page.fill('[data-testid="email-input"]', testUsers.validUser.email);
      await page.fill('[data-testid="password-input"]', 'WrongPassword123!');
      await page.click('[data-testid="login-button"]');

      await expect(page.locator('[data-testid="form-error"]')).toContainText('Invalid credentials');
      await expect(page).toHaveURL('/login');
    });

    test('should show validation errors for empty fields', async ({ page }) => {
      await page.goto('/login');

      await page.click('[data-testid="login-button"]');

      await expect(page.locator('[data-testid="email-error"]')).toBeVisible();
      await expect(page.locator('[data-testid="password-error"]')).toBeVisible();
    });

    test('should show error for non-existent user', async ({ page }) => {
      await page.goto('/login');

      await page.fill('[data-testid="email-input"]', 'nonexistent@example.com');
      await page.fill('[data-testid="password-input"]', 'SomePassword123!');
      await page.click('[data-testid="login-button"]');

      await expect(page.locator('[data-testid="form-error"]')).toContainText('Invalid credentials');
    });
  });

  test.describe('User Logout', () => {
    test.beforeEach(async () => {
      await authHelper.register(testUsers.validUser);
    });

    test('should logout successfully', async ({ page }) => {
      // Should be logged in and on dashboard
      await expect(page).toHaveURL('/dashboard');

      // Logout
      await authHelper.logout();

      // Should redirect to login page
      await expect(page).toHaveURL('/login');
    });

    test('should clear user session after logout', async ({ page }) => {
      await authHelper.logout();

      // Try to access protected route
      await page.goto('/dashboard');

      // Should redirect to login
      await expect(page).toHaveURL('/login');
    });
  });

  test.describe('Protected Routes', () => {
    test('should redirect to login when accessing protected routes without auth', async ({ page }) => {
      const protectedRoutes = ['/dashboard', '/board/123'];

      for (const route of protectedRoutes) {
        await page.goto(route);
        await expect(page).toHaveURL('/login');
      }
    });

    test('should allow access to protected routes when authenticated', async ({ page }) => {
      await authHelper.register(testUsers.validUser);

      await page.goto('/dashboard');
      await expect(page).toHaveURL('/dashboard');
    });

    test('should allow access to public routes without auth', async ({ page }) => {
      const publicRoutes = ['/login', '/register'];

      for (const route of publicRoutes) {
        await page.goto(route);
        await expect(page).toHaveURL(route);
      }
    });
  });

  test.describe('Session Persistence', () => {
    test('should maintain session across page reloads', async ({ page }) => {
      await authHelper.register(testUsers.validUser);

      // Reload page
      await page.reload();

      // Should still be logged in
      await expect(page).toHaveURL('/dashboard');
      await expect(page.locator('[data-testid="user-name"]')).toContainText(testUsers.validUser.name);
    });

    test('should maintain session across navigation', async ({ page }) => {
      await authHelper.register(testUsers.validUser);

      // Navigate to different pages
      await page.goto('/dashboard');
      await expect(page.locator('[data-testid="user-name"]')).toContainText(testUsers.validUser.name);

      // Navigate back to root
      await page.goto('/');
      await expect(page).toHaveURL('/dashboard'); // Should redirect to dashboard
    });
  });

  test.describe('User Profile', () => {
    test.beforeEach(async () => {
      await authHelper.register(testUsers.validUser);
    });

    test('should display user profile information', async ({ page }) => {
      await page.goto('/dashboard');

      // Click on user menu
      await page.click('[data-testid="user-menu-button"]');
      await page.click('[data-testid="profile-button"]');

      // Check profile information
      await expect(page.locator('[data-testid="profile-name"]')).toHaveValue(testUsers.validUser.name);
      await expect(page.locator('[data-testid="profile-email"]')).toHaveValue(testUsers.validUser.email);
    });

    test('should update user profile', async ({ page }) => {
      await page.goto('/dashboard');

      // Open profile
      await page.click('[data-testid="user-menu-button"]');
      await page.click('[data-testid="profile-button"]');

      // Update name
      const newName = 'Updated Test User';
      await page.fill('[data-testid="profile-name"]', newName);
      await page.click('[data-testid="save-profile-button"]');

      // Check success message
      await expect(page.locator('[data-testid="profile-success-message"]')).toBeVisible();

      // Verify update in UI
      await page.goto('/dashboard');
      await expect(page.locator('[data-testid="user-name"]')).toContainText(newName);
    });
  });
});
