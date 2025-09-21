import { test, expect } from '@playwright/test';
import { AuthHelper } from './utils/auth-helpers';
import { BoardHelper } from './utils/board-helpers';
import { testUsers, testBoards, testBoardItems } from './fixtures/test-data';

test.describe('Real-time Collaboration', () => {
  let authHelper1: AuthHelper;
  let authHelper2: AuthHelper;
  let boardHelper1: BoardHelper;
  let boardHelper2: BoardHelper;

  test.describe('Multi-user Collaboration', () => {
    test('should show real-time updates between users', async ({ browser }) => {
      // Create two browser contexts for two users
      const context1 = await browser.newContext();
      const context2 = await browser.newContext();
      
      const page1 = await context1.newPage();
      const page2 = await context2.newPage();

      authHelper1 = new AuthHelper(page1);
      authHelper2 = new AuthHelper(page2);
      boardHelper1 = new BoardHelper(page1);
      boardHelper2 = new BoardHelper(page2);

      // User 1 creates account and board
      await authHelper1.register(testUsers.validUser);
      const boardId = await boardHelper1.createBoard(testBoards.collaborativeBoard);

      // Share board with user 2
      await boardHelper1.shareBoard(testUsers.collaborator.email, 'write');

      // User 2 creates account
      await authHelper2.register(testUsers.collaborator);

      // User 2 opens the shared board
      await boardHelper2.openBoard(boardId);

      // User 1 adds an item
      await boardHelper1.addPostItNote({
        ...testBoardItems.postItNote,
        content: 'Real-time test item'
      });

      // User 2 should see the item appear in real-time
      await boardHelper2.waitForRealTimeUpdate('Real-time test item');

      // User 2 adds an item
      await boardHelper2.addSuspectCard({
        ...testBoardItems.suspectCard,
        content: 'Collaborative suspect'
      });

      // User 1 should see the item appear in real-time
      await boardHelper1.waitForRealTimeUpdate('Collaborative suspect');

      await context1.close();
      await context2.close();
    });

    test('should show live cursor tracking', async ({ browser }) => {
      const context1 = await browser.newContext();
      const context2 = await browser.newContext();
      
      const page1 = await context1.newPage();
      const page2 = await context2.newPage();

      authHelper1 = new AuthHelper(page1);
      authHelper2 = new AuthHelper(page2);
      boardHelper1 = new BoardHelper(page1);
      boardHelper2 = new BoardHelper(page2);

      // Setup users and board
      await authHelper1.register(testUsers.validUser);
      const boardId = await boardHelper1.createBoard(testBoards.collaborativeBoard);
      await boardHelper1.shareBoard(testUsers.collaborator.email, 'write');

      await authHelper2.register(testUsers.collaborator);
      await boardHelper2.openBoard(boardId);

      // Move cursor on page 1
      await page1.mouse.move(200, 200);

      // Page 2 should show user 1's cursor
      await expect(page2.locator(`[data-testid="user-cursor-${testUsers.validUser.email}"]`)).toBeVisible();

      // Move cursor on page 2
      await page2.mouse.move(300, 300);

      // Page 1 should show user 2's cursor
      await expect(page1.locator(`[data-testid="user-cursor-${testUsers.collaborator.email}"]`)).toBeVisible();

      await context1.close();
      await context2.close();
    });

    test('should handle concurrent item editing', async ({ browser }) => {
      const context1 = await browser.newContext();
      const context2 = await browser.newContext();
      
      const page1 = await context1.newPage();
      const page2 = await context2.newPage();

      authHelper1 = new AuthHelper(page1);
      authHelper2 = new AuthHelper(page2);
      boardHelper1 = new BoardHelper(page1);
      boardHelper2 = new BoardHelper(page2);

      // Setup users and board
      await authHelper1.register(testUsers.validUser);
      const boardId = await boardHelper1.createBoard(testBoards.collaborativeBoard);
      await boardHelper1.shareBoard(testUsers.collaborator.email, 'write');

      await authHelper2.register(testUsers.collaborator);
      await boardHelper2.openBoard(boardId);

      // User 1 adds an item
      await boardHelper1.addPostItNote({
        ...testBoardItems.postItNote,
        content: 'Editable item'
      });

      // Wait for item to appear on both pages
      await boardHelper2.waitForRealTimeUpdate('Editable item');

      // User 1 starts editing
      await page1.dblclick('[data-testid="board-item"]:has-text("Editable item")');

      // User 2 should see edit indicator
      await expect(page2.locator('[data-testid="item-being-edited"]')).toBeVisible();

      // User 1 finishes editing
      await page1.fill('[data-testid="item-edit-input"]', 'Updated by User 1');
      await page1.press('[data-testid="item-edit-input"]', 'Enter');

      // User 2 should see the updated content
      await expect(page2.locator('[data-testid="board-item"]:has-text("Updated by User 1")')).toBeVisible();

      await context1.close();
      await context2.close();
    });

    test('should handle user disconnection gracefully', async ({ browser }) => {
      const context1 = await browser.newContext();
      const context2 = await browser.newContext();
      
      const page1 = await context1.newPage();
      const page2 = await context2.newPage();

      authHelper1 = new AuthHelper(page1);
      authHelper2 = new AuthHelper(page2);
      boardHelper1 = new BoardHelper(page1);
      boardHelper2 = new BoardHelper(page2);

      // Setup users and board
      await authHelper1.register(testUsers.validUser);
      const boardId = await boardHelper1.createBoard(testBoards.collaborativeBoard);
      await boardHelper1.shareBoard(testUsers.collaborator.email, 'write');

      await authHelper2.register(testUsers.collaborator);
      await boardHelper2.openBoard(boardId);

      // Verify both users are connected
      await expect(page1.locator(`[data-testid="connected-user-${testUsers.collaborator.email}"]`)).toBeVisible();
      await expect(page2.locator(`[data-testid="connected-user-${testUsers.validUser.email}"]`)).toBeVisible();

      // User 2 disconnects
      await context2.close();

      // User 1 should see that user 2 is no longer connected
      await expect(page1.locator(`[data-testid="connected-user-${testUsers.collaborator.email}"]`)).not.toBeVisible();

      await context1.close();
    });
  });

  test.describe('Permission-based Collaboration', () => {
    test('should respect read-only permissions', async ({ browser }) => {
      const context1 = await browser.newContext();
      const context2 = await browser.newContext();
      
      const page1 = await context1.newPage();
      const page2 = await context2.newPage();

      authHelper1 = new AuthHelper(page1);
      authHelper2 = new AuthHelper(page2);
      boardHelper1 = new BoardHelper(page1);
      boardHelper2 = new BoardHelper(page2);

      // Setup users and board
      await authHelper1.register(testUsers.validUser);
      const boardId = await boardHelper1.createBoard(testBoards.collaborativeBoard);

      // Share with read-only permission
      await boardHelper1.shareBoard(testUsers.collaborator.email, 'read');

      await authHelper2.register(testUsers.collaborator);
      await boardHelper2.openBoard(boardId);

      // User 2 should not see edit controls
      await expect(page2.locator('[data-testid="add-postit-button"]')).not.toBeVisible();
      await expect(page2.locator('[data-testid="add-suspect-button"]')).not.toBeVisible();
      await expect(page2.locator('[data-testid="connection-mode-button"]')).not.toBeVisible();

      // User 1 adds an item
      await boardHelper1.addPostItNote(testBoardItems.postItNote);

      // User 2 should see the item but not be able to edit it
      await boardHelper2.waitForRealTimeUpdate(testBoardItems.postItNote.content);
      
      // Try to double-click to edit (should not work)
      await page2.dblclick(`[data-testid="board-item"]:has-text("${testBoardItems.postItNote.content}")`);
      await expect(page2.locator('[data-testid="item-edit-input"]')).not.toBeVisible();

      await context1.close();
      await context2.close();
    });

    test('should allow write permissions to edit', async ({ browser }) => {
      const context1 = await browser.newContext();
      const context2 = await browser.newContext();
      
      const page1 = await context1.newPage();
      const page2 = await context2.newPage();

      authHelper1 = new AuthHelper(page1);
      authHelper2 = new AuthHelper(page2);
      boardHelper1 = new BoardHelper(page1);
      boardHelper2 = new BoardHelper(page2);

      // Setup users and board
      await authHelper1.register(testUsers.validUser);
      const boardId = await boardHelper1.createBoard(testBoards.collaborativeBoard);

      // Share with write permission
      await boardHelper1.shareBoard(testUsers.collaborator.email, 'write');

      await authHelper2.register(testUsers.collaborator);
      await boardHelper2.openBoard(boardId);

      // User 2 should see edit controls
      await expect(page2.locator('[data-testid="add-postit-button"]')).toBeVisible();
      await expect(page2.locator('[data-testid="add-suspect-button"]')).toBeVisible();

      // User 2 should be able to add items
      await boardHelper2.addPostItNote({
        ...testBoardItems.postItNote,
        content: 'Added by collaborator'
      });

      // User 1 should see the new item
      await boardHelper1.waitForRealTimeUpdate('Added by collaborator');

      await context1.close();
      await context2.close();
    });

    test('should allow admin permissions to manage board', async ({ browser }) => {
      const context1 = await browser.newContext();
      const context2 = await browser.newContext();
      
      const page1 = await context1.newPage();
      const page2 = await context2.newPage();

      authHelper1 = new AuthHelper(page1);
      authHelper2 = new AuthHelper(page2);
      boardHelper1 = new BoardHelper(page1);
      boardHelper2 = new BoardHelper(page2);

      // Setup users and board
      await authHelper1.register(testUsers.validUser);
      const boardId = await boardHelper1.createBoard(testBoards.collaborativeBoard);

      // Share with admin permission
      await boardHelper1.shareBoard(testUsers.collaborator.email, 'admin');

      await authHelper2.register(testUsers.collaborator);
      await boardHelper2.openBoard(boardId);

      // User 2 should see admin controls
      await expect(page2.locator('[data-testid="board-settings-button"]')).toBeVisible();
      await expect(page2.locator('[data-testid="share-board-button"]')).toBeVisible();

      // User 2 should be able to modify board settings
      await page2.click('[data-testid="board-settings-button"]');
      await page2.fill('[data-testid="board-title-input"]', 'Updated by Admin');
      await page2.click('[data-testid="save-board-settings"]');

      // User 1 should see the updated title
      await expect(page1.locator('[data-testid="board-title"]')).toContainText('Updated by Admin');

      await context1.close();
      await context2.close();
    });
  });

  test.describe('Conflict Resolution', () => {
    test('should handle simultaneous item creation', async ({ browser }) => {
      const context1 = await browser.newContext();
      const context2 = await browser.newContext();
      
      const page1 = await context1.newPage();
      const page2 = await context2.newPage();

      authHelper1 = new AuthHelper(page1);
      authHelper2 = new AuthHelper(page2);
      boardHelper1 = new BoardHelper(page1);
      boardHelper2 = new BoardHelper(page2);

      // Setup users and board
      await authHelper1.register(testUsers.validUser);
      const boardId = await boardHelper1.createBoard(testBoards.collaborativeBoard);
      await boardHelper1.shareBoard(testUsers.collaborator.email, 'write');

      await authHelper2.register(testUsers.collaborator);
      await boardHelper2.openBoard(boardId);

      // Both users add items simultaneously
      const [, ] = await Promise.all([
        boardHelper1.addPostItNote({
          ...testBoardItems.postItNote,
          content: 'Item by User 1',
          position: { x: 100, y: 100 }
        }),
        boardHelper2.addPostItNote({
          ...testBoardItems.postItNote,
          content: 'Item by User 2',
          position: { x: 150, y: 100 }
        })
      ]);

      // Both items should be visible on both pages
      await expect(page1.locator('[data-testid="board-item"]:has-text("Item by User 1")')).toBeVisible();
      await expect(page1.locator('[data-testid="board-item"]:has-text("Item by User 2")')).toBeVisible();
      await expect(page2.locator('[data-testid="board-item"]:has-text("Item by User 1")')).toBeVisible();
      await expect(page2.locator('[data-testid="board-item"]:has-text("Item by User 2")')).toBeVisible();

      await context1.close();
      await context2.close();
    });

    test('should handle simultaneous item deletion', async ({ browser }) => {
      const context1 = await browser.newContext();
      const context2 = await browser.newContext();
      
      const page1 = await context1.newPage();
      const page2 = await context2.newPage();

      authHelper1 = new AuthHelper(page1);
      authHelper2 = new AuthHelper(page2);
      boardHelper1 = new BoardHelper(page1);
      boardHelper2 = new BoardHelper(page2);

      // Setup users and board
      await authHelper1.register(testUsers.validUser);
      const boardId = await boardHelper1.createBoard(testBoards.collaborativeBoard);
      await boardHelper1.shareBoard(testUsers.collaborator.email, 'write');

      await authHelper2.register(testUsers.collaborator);
      await boardHelper2.openBoard(boardId);

      // Add an item
      await boardHelper1.addPostItNote({
        ...testBoardItems.postItNote,
        content: 'Item to delete'
      });

      // Wait for item to appear on both pages
      await boardHelper2.waitForRealTimeUpdate('Item to delete');

      // Both users try to delete the same item
      await Promise.all([
        boardHelper1.deleteItem('Item to delete').catch(() => {}), // May fail if already deleted
        boardHelper2.deleteItem('Item to delete').catch(() => {}) // May fail if already deleted
      ]);

      // Item should be deleted on both pages
      await expect(page1.locator('[data-testid="board-item"]:has-text("Item to delete")')).not.toBeVisible();
      await expect(page2.locator('[data-testid="board-item"]:has-text("Item to delete")')).not.toBeVisible();

      await context1.close();
      await context2.close();
    });
  });
});
