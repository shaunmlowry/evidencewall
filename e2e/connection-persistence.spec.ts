import { test, expect } from '@playwright/test';
import { AuthHelper } from './utils/auth-helpers';
import { BoardHelper } from './utils/board-helpers';
import { testUsers, testBoards, testBoardItems } from './fixtures/test-data';

test.describe('Board Connection Persistence', () => {
  let authHelper: AuthHelper;
  let boardHelper: BoardHelper;
  let boardId: string;

  test.beforeEach(async ({ page }) => {
    authHelper = new AuthHelper(page);
    boardHelper = new BoardHelper(page);
    
    // Login and create a board with items
    await authHelper.register(testUsers.validUser);
    boardId = await boardHelper.createBoard(testBoards.privateBoard);
    
    // Add two items to connect
    await boardHelper.addPostItNote({
      ...testBoardItems.postItNote,
      content: 'Evidence Item A',
      position: { x: 100, y: 100 }
    });
    
    await boardHelper.addSuspectCard({
      ...testBoardItems.suspectCard,
      content: 'Suspect Item B',
      position: { x: 300, y: 200 }
    });
  });

  test('should preserve connections when navigating away and back to board', async ({ page }) => {
    // Create a connection between the items
    await boardHelper.connectItems('Evidence Item A', 'Suspect Item B');

    // Verify connection exists
    await expect(page.locator('[data-testid="board-connection"]')).toBeVisible();
    
    // Get connection count
    const initialConnectionCount = await page.locator('[data-testid="board-connection"]').count();
    expect(initialConnectionCount).toBe(1);

    // Navigate away from the board
    await page.goto('/dashboard');
    await expect(page.locator('[data-testid="dashboard-title"]')).toContainText('My Boards');

    // Navigate back to the board
    await page.click(`[data-testid="board-card-${boardId}"]`);
    await expect(page).toHaveURL(`/board/${boardId}`);

    // Wait for board to load
    await expect(page.locator('[data-testid="board-canvas"]')).toBeVisible();
    await expect(page.locator('[data-testid="board-item"]:has-text("Evidence Item A")')).toBeVisible();
    await expect(page.locator('[data-testid="board-item"]:has-text("Suspect Item B")')).toBeVisible();

    // Verify connection is still there after navigation
    await expect(page.locator('[data-testid="board-connection"]')).toBeVisible();
    
    const finalConnectionCount = await page.locator('[data-testid="board-connection"]').count();
    expect(finalConnectionCount).toBe(1);
  });

  test('should preserve multiple connections when navigating away and back', async ({ page }) => {
    // Add a third item
    await boardHelper.addPostItNote({
      ...testBoardItems.postItNote,
      content: 'Evidence Item C',
      position: { x: 200, y: 300 }
    });

    // Create multiple connections
    await boardHelper.connectItems('Evidence Item A', 'Suspect Item B');
    await boardHelper.connectItems('Suspect Item B', 'Evidence Item C');
    await boardHelper.connectItems('Evidence Item A', 'Evidence Item C');

    // Verify all connections exist
    await expect(page.locator('[data-testid="board-connection"]')).toHaveCount(3);

    // Navigate away and back
    await page.goto('/dashboard');
    await page.click(`[data-testid="board-card-${boardId}"]`);

    // Wait for board to load
    await expect(page.locator('[data-testid="board-canvas"]')).toBeVisible();

    // Verify all connections are still there
    await expect(page.locator('[data-testid="board-connection"]')).toHaveCount(3);
  });

  test('should preserve connections after browser refresh', async ({ page }) => {
    // Create a connection
    await boardHelper.connectItems('Evidence Item A', 'Suspect Item B');
    await expect(page.locator('[data-testid="board-connection"]')).toBeVisible();

    // Refresh the page
    await page.reload();

    // Wait for board to load after refresh
    await expect(page.locator('[data-testid="board-canvas"]')).toBeVisible();
    await expect(page.locator('[data-testid="board-item"]:has-text("Evidence Item A")')).toBeVisible();
    await expect(page.locator('[data-testid="board-item"]:has-text("Suspect Item B")')).toBeVisible();

    // Verify connection is still there after refresh
    await expect(page.locator('[data-testid="board-connection"]')).toBeVisible();
  });

  test('should maintain connection visual properties after navigation', async ({ page }) => {
    // Create a connection
    await boardHelper.connectItems('Evidence Item A', 'Suspect Item B');

    // Get connection visual properties
    const connection = page.locator('[data-testid="board-connection"]');
    await expect(connection).toBeVisible();
    
    const initialStroke = await connection.getAttribute('stroke');
    const initialStrokeWidth = await connection.getAttribute('stroke-width');

    // Navigate away and back
    await page.goto('/dashboard');
    await page.click(`[data-testid="board-card-${boardId}"]`);
    await expect(page.locator('[data-testid="board-canvas"]')).toBeVisible();

    // Verify connection visual properties are preserved
    const restoredConnection = page.locator('[data-testid="board-connection"]');
    await expect(restoredConnection).toBeVisible();
    
    const finalStroke = await restoredConnection.getAttribute('stroke');
    const finalStrokeWidth = await restoredConnection.getAttribute('stroke-width');

    expect(finalStroke).toBe(initialStroke);
    expect(finalStrokeWidth).toBe(initialStrokeWidth);
  });
});
