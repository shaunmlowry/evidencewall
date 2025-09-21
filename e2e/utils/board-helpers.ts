import { Page, expect } from '@playwright/test';
import { testBoards, testBoardItems } from '../fixtures/test-data';

export class BoardHelper {
  constructor(private page: Page) {}

  async createBoard(board = testBoards.privateBoard) {
    await this.page.goto('/dashboard');
    
    // Click create board button
    await this.page.click('[data-testid="create-board-button"]');
    
    // Fill board details
    await this.page.fill('[data-testid="board-title-input"]', board.title);
    await this.page.fill('[data-testid="board-description-input"]', board.description);
    
    if (board.isPublic) {
      await this.page.check('[data-testid="board-public-checkbox"]');
    }
    
    // Submit form
    await this.page.click('[data-testid="create-board-submit"]');
    
    // Wait for board to be created and navigate to board page
    await expect(this.page).toHaveURL(/\/board\/[a-f0-9-]+/);
    
    // Return the board ID from URL
    const url = this.page.url();
    const boardId = url.split('/board/')[1];
    return boardId;
  }

  async openBoard(boardId: string) {
    await this.page.goto(`/board/${boardId}`);
    await expect(this.page.locator('[data-testid="board-canvas"]')).toBeVisible();
  }

  async addPostItNote(item = testBoardItems.postItNote) {
    // Click add post-it button
    await this.page.click('[data-testid="add-post-it"]');
    
    // Wait for item to appear (it's created automatically)
    await expect(this.page.locator('[data-testid="board-item"]').last()).toBeVisible();
    
    // If we need to set specific content, double-click to edit
    if (item.content && item.content !== 'Evidence notes...\n') {
      const items = await this.page.locator('[data-testid="board-item"]').all();
      const lastItem = items[items.length - 1];
      await lastItem.dblclick();
      await this.page.fill('[data-testid="item-edit-input"]', item.content);
      await this.page.press('[data-testid="item-edit-input"]', 'Enter');
    }
  }

  async addSuspectCard(item = testBoardItems.suspectCard) {
    // Click add suspect button
    await this.page.click('[data-testid="add-suspect-card"]');
    
    // Wait for item to appear (it's created automatically)
    await expect(this.page.locator('[data-testid="board-item"]').last()).toBeVisible();
    
    // If we need to set specific content, double-click to edit
    if (item.content && item.content !== 'Suspect Name\nAge: Unknown\nLast seen:\nNotes:') {
      const items = await this.page.locator('[data-testid="board-item"]').all();
      const lastItem = items[items.length - 1];
      await lastItem.dblclick();
      await this.page.fill('[data-testid="item-edit-input"]', item.content);
      await this.page.press('[data-testid="item-edit-input"]', 'Enter');
    }
  }

  async connectItems(fromItemText: string, toItemText: string) {
    // Enable connection mode
    await this.page.click('[data-testid="connect-mode-button"]');
    
    // Click on first item
    await this.page.click(`[data-testid="board-item"]:has-text("${fromItemText}")`);
    
    // Click on second item
    await this.page.click(`[data-testid="board-item"]:has-text("${toItemText}")`);
    
    // Wait for connection to appear
    await expect(this.page.locator('[data-testid="board-connection"]')).toBeVisible();
    
    // Disable connection mode
    await this.page.click('[data-testid="connect-mode-button"]');
  }

  async dragItem(itemText: string, newPosition: { x: number; y: number }) {
    const item = this.page.locator(`[data-testid="board-item"]:has-text("${itemText}")`);
    
    // Get current position
    const box = await item.boundingBox();
    if (!box) throw new Error('Item not found');
    
    // Drag to new position
    await this.page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
    await this.page.mouse.down();
    await this.page.mouse.move(newPosition.x, newPosition.y);
    await this.page.mouse.up();
    
    // Wait for position update
    await this.page.waitForTimeout(500);
  }

  async deleteItem(itemText: string) {
    // Enable delete mode
    await this.page.click('[data-testid="delete-mode-button"]');
    
    // Click on item to delete
    await this.page.click(`[data-testid="board-item"]:has-text("${itemText}")`);
    
    // Confirm deletion if modal appears
    if (await this.page.locator('[data-testid="confirm-delete-button"]').isVisible()) {
      await this.page.click('[data-testid="confirm-delete-button"]');
    }
    
    // Wait for item to disappear
    await expect(this.page.locator(`[data-testid="board-item"]:has-text("${itemText}")`)).not.toBeVisible();
    
    // Disable delete mode
    await this.page.click('[data-testid="delete-mode-button"]');
  }

  async shareBoard(userEmail: string, permission: 'read' | 'write' | 'admin' = 'read') {
    // Click share button
    await this.page.click('[data-testid="share-board-button"]');
    
    // Fill user email
    await this.page.fill('[data-testid="share-email-input"]', userEmail);
    
    // Select permission level
    await this.page.selectOption('[data-testid="permission-select"]', permission);
    
    // Submit share
    await this.page.click('[data-testid="share-submit-button"]');
    
    // Wait for success message
    await expect(this.page.locator('[data-testid="share-success-message"]')).toBeVisible();
  }

  async getBoardItems() {
    const items = await this.page.locator('[data-testid="board-item"]').all();
    const itemData = [];
    
    for (const item of items) {
      const text = await item.textContent();
      const box = await item.boundingBox();
      itemData.push({
        text: text?.trim(),
        position: box ? { x: box.x, y: box.y } : null
      });
    }
    
    return itemData;
  }

  async waitForRealTimeUpdate(expectedItemText: string, timeout = 10000) {
    await expect(
      this.page.locator(`[data-testid="board-item"]:has-text("${expectedItemText}")`)
    ).toBeVisible({ timeout });
  }
}
