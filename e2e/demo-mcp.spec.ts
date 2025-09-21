import { test, expect } from '@playwright/test';

test.describe('Playwright MCP Server Demo', () => {
  test('should demonstrate MCP server integration', async ({ page }) => {
    // This test demonstrates how to use the Playwright MCP server
    // The MCP server provides browser automation capabilities
    
    // Navigate to a simple page for demonstration
    await page.goto('https://example.com');
    
    // Take a screenshot using MCP server capabilities
    await page.screenshot({ path: 'demo-screenshot.png' });
    
    // Verify page title
    await expect(page).toHaveTitle(/Example Domain/);
    
    // Demonstrate form interaction
    // Note: This is a demo - example.com doesn't have forms
    // In real tests, you'd interact with your application's forms
    
    // Get page content
    const content = await page.textContent('body');
    expect(content).toContain('Example Domain');
    
    // Demonstrate navigation
    const url = page.url();
    expect(url).toBe('https://example.com/');
  });

  test('should demonstrate browser automation features', async ({ page }) => {
    await page.goto('https://httpbin.org/forms/post');
    
    // Fill out a form
    await page.fill('input[name="custname"]', 'Test User');
    await page.fill('input[name="custtel"]', '123-456-7890');
    await page.fill('input[name="custemail"]', 'test@example.com');
    await page.selectOption('select[name="size"]', 'large');
    await page.check('input[name="topping"][value="bacon"]');
    
    // Submit form
    await page.click('input[type="submit"]');
    
    // Verify submission
    await expect(page.locator('pre')).toContainText('Test User');
  });

  test('should demonstrate network interception', async ({ page }) => {
    // Intercept network requests
    const responses = [];
    page.on('response', response => {
      responses.push({
        url: response.url(),
        status: response.status()
      });
    });
    
    await page.goto('https://jsonplaceholder.typicode.com/posts/1');
    
    // Verify we captured network activity
    expect(responses.length).toBeGreaterThan(0);
    
    // Check JSON response
    const jsonContent = await page.textContent('pre');
    const data = JSON.parse(jsonContent);
    expect(data).toHaveProperty('id', 1);
    expect(data).toHaveProperty('title');
  });

  test('should demonstrate mobile viewport testing', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    await page.goto('https://example.com');
    
    // Verify mobile layout
    const viewport = page.viewportSize();
    expect(viewport.width).toBe(375);
    expect(viewport.height).toBe(667);
    
    // Take mobile screenshot
    await page.screenshot({ path: 'mobile-demo.png' });
  });

  test('should demonstrate JavaScript evaluation', async ({ page }) => {
    await page.goto('https://example.com');
    
    // Evaluate JavaScript in the browser
    const title = await page.evaluate(() => document.title);
    expect(title).toBe('Example Domain');
    
    // Modify page content
    await page.evaluate(() => {
      document.body.style.backgroundColor = 'lightblue';
    });
    
    // Verify modification
    const bgColor = await page.evaluate(() => 
      getComputedStyle(document.body).backgroundColor
    );
    expect(bgColor).toBe('rgb(173, 216, 230)'); // lightblue in RGB
  });

  test('should demonstrate file upload simulation', async ({ page }) => {
    await page.goto('https://httpbin.org/forms/post');
    
    // Create a temporary file for upload testing
    const fileContent = 'This is a test file for upload demonstration';
    
    // Note: In real tests, you'd use actual file paths
    // This is just demonstrating the API
    const fileInput = page.locator('input[type="file"]');
    
    if (await fileInput.count() > 0) {
      // File upload would be handled here
      console.log('File input found - upload simulation would happen here');
    }
  });

  test('should demonstrate accessibility testing', async ({ page }) => {
    await page.goto('https://example.com');
    
    // Check for accessibility landmarks
    const main = page.locator('main, [role="main"]');
    const headings = page.locator('h1, h2, h3, h4, h5, h6');
    
    // Verify semantic structure
    await expect(headings.first()).toBeVisible();
    
    // Check for alt text on images
    const images = page.locator('img');
    const imageCount = await images.count();
    
    for (let i = 0; i < imageCount; i++) {
      const img = images.nth(i);
      const alt = await img.getAttribute('alt');
      // In a real test, you'd assert alt text exists
      console.log(`Image ${i} alt text:`, alt);
    }
  });
});
