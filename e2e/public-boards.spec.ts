import { test, expect } from '@playwright/test';
import { AuthHelper } from './utils/auth-helpers';
import { BoardHelper } from './utils/board-helpers';
import { testUsers, testBoards, testBoardItems } from './fixtures/test-data';

test.describe('Public Boards', () => {
  let authHelper: AuthHelper;
  let boardHelper: BoardHelper;

  test.describe('Public Board Access', () => {
    let publicBoardId: string;

    test.beforeEach(async ({ page }) => {
      authHelper = new AuthHelper(page);
      boardHelper = new BoardHelper(page);

      // Create a user and public board
      await authHelper.register(testUsers.validUser);
      publicBoardId = await boardHelper.createBoard(testBoards.publicBoard);

      // Add some content to the board
      await boardHelper.addPostItNote({
        ...testBoardItems.postItNote,
        content: 'Public evidence item'
      });

      await boardHelper.addSuspectCard({
        ...testBoardItems.suspectCard,
        content: 'Public suspect card'
      });

      // Logout to test public access
      await authHelper.logout();
    });

    test('should allow anonymous users to view public boards', async ({ page }) => {
      // Navigate to public board URL
      await page.goto(`/public/board/${publicBoardId}`);

      // Should be able to view the board without authentication
      await expect(page.locator('[data-testid="board-title"]')).toContainText(testBoards.publicBoard.title);
      await expect(page.locator('[data-testid="board-canvas"]')).toBeVisible();

      // Should see the board items
      await expect(page.locator('[data-testid="board-item"]:has-text("Public evidence item")')).toBeVisible();
      await expect(page.locator('[data-testid="board-item"]:has-text("Public suspect card")')).toBeVisible();
    });

    test('should not show edit controls for anonymous users', async ({ page }) => {
      await page.goto(`/public/board/${publicBoardId}`);

      // Edit controls should not be visible
      await expect(page.locator('[data-testid="add-postit-button"]')).not.toBeVisible();
      await expect(page.locator('[data-testid="add-suspect-button"]')).not.toBeVisible();
      await expect(page.locator('[data-testid="connection-mode-button"]')).not.toBeVisible();
      await expect(page.locator('[data-testid="delete-mode-button"]')).not.toBeVisible();
      await expect(page.locator('[data-testid="board-settings-button"]')).not.toBeVisible();
    });

    test('should not allow anonymous users to edit items', async ({ page }) => {
      await page.goto(`/public/board/${publicBoardId}`);

      // Try to double-click to edit an item
      await page.dblclick('[data-testid="board-item"]:has-text("Public evidence item")');

      // Edit input should not appear
      await expect(page.locator('[data-testid="item-edit-input"]')).not.toBeVisible();
    });

    test('should show public board indicator', async ({ page }) => {
      await page.goto(`/public/board/${publicBoardId}`);

      // Should show public board indicator
      await expect(page.locator('[data-testid="public-board-indicator"]')).toBeVisible();
      await expect(page.locator('[data-testid="public-board-indicator"]')).toContainText('Public Board');
    });

    test('should show login prompt for interaction', async ({ page }) => {
      await page.goto(`/public/board/${publicBoardId}`);

      // Should show a login prompt or button
      await expect(page.locator('[data-testid="login-to-edit-prompt"]')).toBeVisible();
      await expect(page.locator('[data-testid="login-button"]')).toBeVisible();
    });

    test('should redirect to login when clicking login button', async ({ page }) => {
      await page.goto(`/public/board/${publicBoardId}`);

      // Click login button
      await page.click('[data-testid="login-button"]');

      // Should redirect to login page
      await expect(page).toHaveURL('/login');
    });

    test('should handle non-existent public boards', async ({ page }) => {
      await page.goto('/public/board/non-existent-id');

      // Should show 404 or not found message
      await expect(page.locator('[data-testid="board-not-found"]')).toBeVisible();
      await expect(page.locator('[data-testid="board-not-found"]')).toContainText('Board not found');
    });

    test('should handle private boards accessed via public URL', async ({ page }) => {
      // Create a private board first
      await authHelper.login(testUsers.validUser);
      const privateBoardId = await boardHelper.createBoard(testBoards.privateBoard);
      await authHelper.logout();

      // Try to access private board via public URL
      await page.goto(`/public/board/${privateBoardId}`);

      // Should show access denied or not found
      await expect(page.locator('[data-testid="board-access-denied"]')).toBeVisible();
    });
  });

  test.describe('Public Board SEO and Sharing', () => {
    let publicBoardId: string;

    test.beforeEach(async ({ page }) => {
      authHelper = new AuthHelper(page);
      boardHelper = new BoardHelper(page);

      await authHelper.register(testUsers.validUser);
      publicBoardId = await boardHelper.createBoard(testBoards.publicBoard);
      await authHelper.logout();
    });

    test('should have proper meta tags for SEO', async ({ page }) => {
      await page.goto(`/public/board/${publicBoardId}`);

      // Check for proper meta tags
      const title = await page.title();
      expect(title).toContain(testBoards.publicBoard.title);

      // Check meta description
      const metaDescription = await page.locator('meta[name="description"]').getAttribute('content');
      expect(metaDescription).toContain(testBoards.publicBoard.description);

      // Check Open Graph tags
      const ogTitle = await page.locator('meta[property="og:title"]').getAttribute('content');
      expect(ogTitle).toContain(testBoards.publicBoard.title);

      const ogDescription = await page.locator('meta[property="og:description"]').getAttribute('content');
      expect(ogDescription).toContain(testBoards.publicBoard.description);
    });

    test('should have shareable URL', async ({ page }) => {
      await page.goto(`/public/board/${publicBoardId}`);

      // Should show share button
      await expect(page.locator('[data-testid="share-url-button"]')).toBeVisible();

      // Click share button
      await page.click('[data-testid="share-url-button"]');

      // Should show shareable URL
      await expect(page.locator('[data-testid="shareable-url"]')).toBeVisible();
      
      const shareableUrl = await page.locator('[data-testid="shareable-url"]').inputValue();
      expect(shareableUrl).toContain(`/public/board/${publicBoardId}`);
    });

    test('should copy URL to clipboard', async ({ page }) => {
      await page.goto(`/public/board/${publicBoardId}`);

      // Grant clipboard permissions
      await page.context().grantPermissions(['clipboard-write']);

      // Click copy URL button
      await page.click('[data-testid="copy-url-button"]');

      // Should show success message
      await expect(page.locator('[data-testid="url-copied-message"]')).toBeVisible();
    });
  });

  test.describe('Public Board Real-time Updates', () => {
    let publicBoardId: string;

    test('should show real-time updates to anonymous viewers', async ({ browser }) => {
      // Create two browser contexts - one for owner, one for anonymous viewer
      const ownerContext = await browser.newContext();
      const viewerContext = await browser.newContext();
      
      const ownerPage = await ownerContext.newPage();
      const viewerPage = await viewerContext.newPage();

      const ownerAuth = new AuthHelper(ownerPage);
      const ownerBoard = new BoardHelper(ownerPage);

      // Owner creates board and adds content
      await ownerAuth.register(testUsers.validUser);
      publicBoardId = await ownerBoard.createBoard(testBoards.publicBoard);

      // Anonymous viewer opens public board
      await viewerPage.goto(`/public/board/${publicBoardId}`);

      // Owner adds an item
      await ownerBoard.addPostItNote({
        ...testBoardItems.postItNote,
        content: 'Real-time public update'
      });

      // Anonymous viewer should see the update
      await expect(viewerPage.locator('[data-testid="board-item"]:has-text("Real-time public update")')).toBeVisible();

      // Owner moves an item
      await ownerBoard.dragItem('Real-time public update', { x: 300, y: 200 });

      // Viewer should see the movement (position update)
      await viewerPage.waitForTimeout(1000); // Wait for real-time update

      await ownerContext.close();
      await viewerContext.close();
    });

    test('should show active user count on public boards', async ({ browser }) => {
      const context1 = await browser.newContext();
      const context2 = await browser.newContext();
      
      const ownerPage = await context1.newPage();
      const viewerPage = await context2.newPage();

      const ownerAuth = new AuthHelper(ownerPage);
      const ownerBoard = new BoardHelper(ownerPage);

      // Owner creates board
      await ownerAuth.register(testUsers.validUser);
      publicBoardId = await ownerBoard.createBoard(testBoards.publicBoard);

      // Check initial viewer count
      await expect(ownerPage.locator('[data-testid="viewer-count"]')).toContainText('1 viewer');

      // Anonymous viewer joins
      await viewerPage.goto(`/public/board/${publicBoardId}`);

      // Owner should see increased viewer count
      await expect(ownerPage.locator('[data-testid="viewer-count"]')).toContainText('2 viewers');

      // Viewer should also see viewer count
      await expect(viewerPage.locator('[data-testid="viewer-count"]')).toContainText('2 viewers');

      await context1.close();
      await context2.close();
    });
  });

  test.describe('Public Board Performance', () => {
    test('should load public boards quickly', async ({ page }) => {
      // Create a board with many items
      authHelper = new AuthHelper(page);
      boardHelper = new BoardHelper(page);

      await authHelper.register(testUsers.validUser);
      const publicBoardId = await boardHelper.createBoard(testBoards.publicBoard);

      // Add multiple items
      for (let i = 0; i < 20; i++) {
        await boardHelper.addPostItNote({
          ...testBoardItems.postItNote,
          content: `Performance test item ${i}`,
          position: { x: 100 + (i % 5) * 100, y: 100 + Math.floor(i / 5) * 100 }
        });
      }

      await authHelper.logout();

      // Measure load time
      const startTime = Date.now();
      await page.goto(`/public/board/${publicBoardId}`);
      
      // Wait for board to be fully loaded
      await expect(page.locator('[data-testid="board-canvas"]')).toBeVisible();
      await expect(page.locator('[data-testid="board-item"]').first()).toBeVisible();
      
      const loadTime = Date.now() - startTime;
      
      // Should load within reasonable time (adjust threshold as needed)
      expect(loadTime).toBeLessThan(5000); // 5 seconds
    });

    test('should handle large number of concurrent viewers', async ({ browser }) => {
      // This test simulates multiple viewers but in a simplified way
      // In a real scenario, you'd want to test with actual load testing tools
      
      authHelper = new AuthHelper(await browser.newPage());
      boardHelper = new BoardHelper(authHelper.page);

      await authHelper.register(testUsers.validUser);
      const publicBoardId = await boardHelper.createBoard(testBoards.publicBoard);
      await authHelper.logout();

      // Create multiple viewer contexts
      const viewerPromises = [];
      for (let i = 0; i < 5; i++) {
        const context = await browser.newContext();
        const page = await context.newPage();
        
        viewerPromises.push(
          page.goto(`/public/board/${publicBoardId}`).then(() => {
            return expect(page.locator('[data-testid="board-canvas"]')).toBeVisible();
          })
        );
      }

      // All viewers should be able to load the board
      await Promise.all(viewerPromises);
    });
  });
});
