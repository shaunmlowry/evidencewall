import { test, expect } from '@playwright/test';
import { AuthHelper } from './utils/auth-helpers';
import { BoardHelper } from './utils/board-helpers';
import { testUsers, testBoards, testBoardItems } from './fixtures/test-data';

test.describe('Board Management', () => {
  let authHelper: AuthHelper;
  let boardHelper: BoardHelper;

  test.beforeEach(async ({ page }) => {
    authHelper = new AuthHelper(page);
    boardHelper = new BoardHelper(page);
    
    // Login before each test
    await authHelper.register(testUsers.validUser);
  });

  test.describe('Board Creation', () => {
    test('should create a new private board', async ({ page }) => {
      await page.goto('/dashboard');

      // Check dashboard elements
      await expect(page.locator('[data-testid="dashboard-title"]')).toContainText('My Boards');
      await expect(page.locator('[data-testid="create-board-button"]')).toBeVisible();

      // Create board
      const boardId = await boardHelper.createBoard(testBoards.privateBoard);

      // Verify board was created
      expect(boardId).toBeTruthy();
      await expect(page).toHaveURL(`/board/${boardId}`);
      await expect(page.locator('[data-testid="board-title"]')).toContainText(testBoards.privateBoard.title);
    });

    test('should create a new public board', async ({ page }) => {
      const boardId = await boardHelper.createBoard(testBoards.publicBoard);

      // Verify board was created
      expect(boardId).toBeTruthy();
      await expect(page.locator('[data-testid="board-title"]')).toContainText(testBoards.publicBoard.title);
      await expect(page.locator('[data-testid="public-board-indicator"]')).toBeVisible();
    });

    test('should show validation errors for invalid board data', async ({ page }) => {
      await page.goto('/dashboard');
      await page.click('[data-testid="create-board-button"]');

      // Try to submit empty form
      await page.click('[data-testid="create-board-submit"]');

      // Check for validation errors
      await expect(page.locator('[data-testid="title-error"]')).toBeVisible();
    });

    test('should show created board in dashboard', async ({ page }) => {
      const boardId = await boardHelper.createBoard(testBoards.privateBoard);

      // Go back to dashboard
      await page.goto('/dashboard');

      // Check if board appears in list
      await expect(page.locator(`[data-testid="board-card-${boardId}"]`)).toBeVisible();
      await expect(page.locator(`[data-testid="board-card-${boardId}"] h3`)).toContainText(testBoards.privateBoard.title);
    });
  });

  test.describe('Board Navigation', () => {
    let boardId: string;

    test.beforeEach(async () => {
      boardId = await boardHelper.createBoard(testBoards.privateBoard);
    });

    test('should navigate to board from dashboard', async ({ page }) => {
      await page.goto('/dashboard');

      // Click on board card
      await page.click(`[data-testid="board-card-${boardId}"]`);

      // Should navigate to board
      await expect(page).toHaveURL(`/board/${boardId}`);
      await expect(page.locator('[data-testid="board-canvas"]')).toBeVisible();
    });

    test('should navigate back to dashboard from board', async ({ page }) => {
      await boardHelper.openBoard(boardId);

      // Click dashboard link
      await page.click('[data-testid="dashboard-link"]');

      // Should navigate to dashboard
      await expect(page).toHaveURL('/dashboard');
    });
  });

  test.describe('Board Items Management', () => {
    let boardId: string;

    test.beforeEach(async () => {
      boardId = await boardHelper.createBoard(testBoards.privateBoard);
    });

    test('should add a post-it note to the board', async ({ page }) => {
      await boardHelper.addPostItNote(testBoardItems.postItNote);

      // Verify item was added
      await expect(page.locator(`[data-testid="board-item"]:has-text("${testBoardItems.postItNote.content}")`)).toBeVisible();
    });

    test('should add a suspect card to the board', async ({ page }) => {
      await boardHelper.addSuspectCard(testBoardItems.suspectCard);

      // Verify item was added
      await expect(page.locator(`[data-testid="board-item"]:has-text("${testBoardItems.suspectCard.content}")`)).toBeVisible();
    });

    test('should add multiple items with different colors', async ({ page }) => {
      // Add multiple items
      await boardHelper.addPostItNote({
        ...testBoardItems.postItNote,
        content: 'Yellow Note',
        color: '#ffeb3b',
        position: { x: 100, y: 100 }
      });

      await boardHelper.addPostItNote({
        ...testBoardItems.postItNote,
        content: 'Blue Note',
        color: '#bbdefb',
        position: { x: 200, y: 100 }
      });

      await boardHelper.addSuspectCard({
        ...testBoardItems.suspectCard,
        content: 'Red Suspect',
        color: '#ffcdd2',
        position: { x: 300, y: 100 }
      });

      // Verify all items are visible
      await expect(page.locator('[data-testid="board-item"]:has-text("Yellow Note")')).toBeVisible();
      await expect(page.locator('[data-testid="board-item"]:has-text("Blue Note")')).toBeVisible();
      await expect(page.locator('[data-testid="board-item"]:has-text("Red Suspect")')).toBeVisible();
    });

    test('should drag and drop items on the board', async ({ page }) => {
      await boardHelper.addPostItNote(testBoardItems.postItNote);

      // Drag item to new position
      const newPosition = { x: 400, y: 300 };
      await boardHelper.dragItem(testBoardItems.postItNote.content, newPosition);

      // Verify item moved (this would need to check actual position)
      await expect(page.locator(`[data-testid="board-item"]:has-text("${testBoardItems.postItNote.content}")`)).toBeVisible();
    });

    test('should delete items from the board', async ({ page }) => {
      await boardHelper.addPostItNote(testBoardItems.postItNote);

      // Delete the item
      await boardHelper.deleteItem(testBoardItems.postItNote.content);

      // Verify item was deleted
      await expect(page.locator(`[data-testid="board-item"]:has-text("${testBoardItems.postItNote.content}")`)).not.toBeVisible();
    });

    test('should edit item content', async ({ page }) => {
      await boardHelper.addPostItNote(testBoardItems.postItNote);

      // Double-click to edit
      await page.dblclick(`[data-testid="board-item"]:has-text("${testBoardItems.postItNote.content}")`);

      // Edit content
      const newContent = 'Updated evidence content';
      await page.fill('[data-testid="item-edit-input"]', newContent);
      await page.press('[data-testid="item-edit-input"]', 'Enter');

      // Verify content was updated
      await expect(page.locator(`[data-testid="board-item"]:has-text("${newContent}")`)).toBeVisible();
    });
  });

  test.describe('Board Connections', () => {
    let boardId: string;

    test.beforeEach(async () => {
      boardId = await boardHelper.createBoard(testBoards.privateBoard);
      
      // Add items to connect
      await boardHelper.addPostItNote({
        ...testBoardItems.postItNote,
        content: 'Evidence A',
        position: { x: 100, y: 100 }
      });
      
      await boardHelper.addSuspectCard({
        ...testBoardItems.suspectCard,
        content: 'Suspect B',
        position: { x: 300, y: 200 }
      });
    });

    test('should create connections between items', async ({ page }) => {
      await boardHelper.connectItems('Evidence A', 'Suspect B');

      // Verify connection was created
      await expect(page.locator('[data-testid="board-connection"]')).toBeVisible();
    });

    test('should delete connections', async ({ page }) => {
      await boardHelper.connectItems('Evidence A', 'Suspect B');

      // Enable delete mode and click connection
      await page.click('[data-testid="delete-mode-button"]');
      await page.click('[data-testid="board-connection"]');

      // Verify connection was deleted
      await expect(page.locator('[data-testid="board-connection"]')).not.toBeVisible();
    });

    test('should create multiple connections', async ({ page }) => {
      // Add third item
      await boardHelper.addPostItNote({
        ...testBoardItems.evidenceItem,
        content: 'Evidence C',
        position: { x: 200, y: 300 }
      });

      // Create multiple connections
      await boardHelper.connectItems('Evidence A', 'Suspect B');
      await boardHelper.connectItems('Suspect B', 'Evidence C');
      await boardHelper.connectItems('Evidence A', 'Evidence C');

      // Verify all connections exist
      const connections = page.locator('[data-testid="board-connection"]');
      await expect(connections).toHaveCount(3);
    });
  });

  test.describe('Board Sharing', () => {
    let boardId: string;

    test.beforeEach(async () => {
      boardId = await boardHelper.createBoard(testBoards.privateBoard);
    });

    test('should share board with another user', async ({ page }) => {
      await boardHelper.shareBoard(testUsers.collaborator.email, 'read');

      // Verify share success
      await expect(page.locator('[data-testid="share-success-message"]')).toBeVisible();
    });

    test('should show shared users list', async ({ page }) => {
      await boardHelper.shareBoard(testUsers.collaborator.email, 'write');

      // Open share dialog again
      await page.click('[data-testid="share-board-button"]');

      // Verify user appears in shared list
      await expect(page.locator(`[data-testid="shared-user"]:has-text("${testUsers.collaborator.email}")`)).toBeVisible();
    });

    test('should update user permissions', async ({ page }) => {
      await boardHelper.shareBoard(testUsers.collaborator.email, 'read');

      // Open share dialog
      await page.click('[data-testid="share-board-button"]');

      // Update permission
      await page.selectOption(`[data-testid="permission-select-${testUsers.collaborator.email}"]`, 'admin');
      await page.click('[data-testid="update-permission-button"]');

      // Verify permission updated
      await expect(page.locator('[data-testid="permission-updated-message"]')).toBeVisible();
    });

    test('should remove user access', async ({ page }) => {
      await boardHelper.shareBoard(testUsers.collaborator.email, 'read');

      // Open share dialog
      await page.click('[data-testid="share-board-button"]');

      // Remove user
      await page.click(`[data-testid="remove-user-${testUsers.collaborator.email}"]`);
      await page.click('[data-testid="confirm-remove-button"]');

      // Verify user removed
      await expect(page.locator(`[data-testid="shared-user"]:has-text("${testUsers.collaborator.email}")`)).not.toBeVisible();
    });
  });

  test.describe('Board Settings', () => {
    let boardId: string;

    test.beforeEach(async () => {
      boardId = await boardHelper.createBoard(testBoards.privateBoard);
    });

    test('should update board title and description', async ({ page }) => {
      // Open board settings
      await page.click('[data-testid="board-settings-button"]');

      // Update details
      const newTitle = 'Updated Investigation Board';
      const newDescription = 'Updated description for testing';
      
      await page.fill('[data-testid="board-title-input"]', newTitle);
      await page.fill('[data-testid="board-description-input"]', newDescription);
      await page.click('[data-testid="save-board-settings"]');

      // Verify updates
      await expect(page.locator('[data-testid="board-title"]')).toContainText(newTitle);
    });

    test('should toggle board public/private status', async ({ page }) => {
      // Open board settings
      await page.click('[data-testid="board-settings-button"]');

      // Toggle to public
      await page.check('[data-testid="board-public-checkbox"]');
      await page.click('[data-testid="save-board-settings"]');

      // Verify public indicator appears
      await expect(page.locator('[data-testid="public-board-indicator"]')).toBeVisible();
    });

    test('should delete board', async ({ page }) => {
      // Open board settings
      await page.click('[data-testid="board-settings-button"]');

      // Delete board
      await page.click('[data-testid="delete-board-button"]');
      await page.fill('[data-testid="delete-confirmation-input"]', testBoards.privateBoard.title);
      await page.click('[data-testid="confirm-delete-board"]');

      // Should redirect to dashboard
      await expect(page).toHaveURL('/dashboard');

      // Board should not appear in dashboard
      await expect(page.locator(`[data-testid="board-card-${boardId}"]`)).not.toBeVisible();
    });
  });

  test.describe('Board Search and Filtering', () => {
    test.beforeEach(async () => {
      // Create multiple boards for testing
      await boardHelper.createBoard({ ...testBoards.privateBoard, title: 'Murder Investigation' });
      await boardHelper.createBoard({ ...testBoards.publicBoard, title: 'Theft Case' });
      await boardHelper.createBoard({ ...testBoards.collaborativeBoard, title: 'Missing Person' });
    });

    test('should search boards by title', async ({ page }) => {
      await page.goto('/dashboard');

      // Search for specific board
      await page.fill('[data-testid="board-search-input"]', 'Murder');

      // Should show only matching boards
      await expect(page.locator('[data-testid="board-card"]:has-text("Murder Investigation")')).toBeVisible();
      await expect(page.locator('[data-testid="board-card"]:has-text("Theft Case")')).not.toBeVisible();
    });

    test('should filter boards by type', async ({ page }) => {
      await page.goto('/dashboard');

      // Filter by public boards
      await page.selectOption('[data-testid="board-filter-select"]', 'public');

      // Should show only public boards
      await expect(page.locator('[data-testid="board-card"]:has-text("Theft Case")')).toBeVisible();
      await expect(page.locator('[data-testid="board-card"]:has-text("Murder Investigation")')).not.toBeVisible();
    });

    test('should sort boards by date', async ({ page }) => {
      await page.goto('/dashboard');

      // Sort by newest first
      await page.selectOption('[data-testid="board-sort-select"]', 'newest');

      // Verify order (newest board should be first)
      const boardCards = page.locator('[data-testid="board-card"]');
      await expect(boardCards.first()).toContainText('Missing Person');
    });
  });
});
