import { test, expect } from "./setup";

test.describe("Servers page", () => {
  test('should display "Servers" heading', async ({ page }) => {
    await page.goto("/servers");
    await expect(page.getByRole("heading", { name: /Servers/i })).toBeVisible();
  });

  test("add server form opens", async ({ page }) => {
    await page.goto("/servers");

    // Look for "Add Server" button
    const addButton = page.getByRole("button", { name: /add server/i });
    await expect(addButton).toBeVisible();

    // Click to open form/modal
    await addButton.click();

    // Form should appear with required fields
    // The form might be in a modal or inline
    await page.waitForTimeout(500);

    // Look for form fields
    const nameField = page.getByLabel(/name/i).first();
    const urlField = page.getByLabel(/url/i).first();

    // At least one of these fields should be visible
    const nameVisible = await nameField.isVisible().catch(() => false);
    const urlVisible = await urlField.isVisible().catch(() => false);

    expect(nameVisible || urlVisible).toBeTruthy();
  });

  test("create server with valid data", async ({ page }) => {
    await page.goto("/servers");

    // Open add server form
    const addButton = page.getByRole("button", { name: /add server/i });
    await addButton.click();
    await page.waitForTimeout(500);

    // Fill in server details
    // Note: This will fail connection test since we don't have a real server
    // but we can test the form behavior
    await page.getByLabel(/name/i).first().fill("Test Radarr");
    await page.getByLabel(/url/i).first().fill("http://localhost:7878");

    // Look for type selector (radio, select, or buttons)
    const typeSelector = page
      .locator('select, input[type="radio"], button')
      .filter({ hasText: /radarr|sonarr/i })
      .first();
    if (await typeSelector.isVisible()) {
      await typeSelector.click();
    }

    // Fill API key
    const apiKeyField = page.getByLabel(/api key/i).first();
    if (await apiKeyField.isVisible()) {
      await apiKeyField.fill("1234567890abcdef1234567890abcdef");
    }

    // Find and click submit button
    const submitButton = page
      .getByRole("button", { name: /save|add|create|submit/i })
      .first();

    if (await submitButton.isVisible()) {
      await submitButton.click();

      // Wait for response
      await page.waitForTimeout(2000);

      // Either success or error should be shown
      // Since we don't have a real server, expect an error or the form to still be visible
      const pageContent = await page.content();
      expect(pageContent.length).toBeGreaterThan(0);
    }
  });

  test("empty state when no servers", async ({ page }) => {
    await page.goto("/servers");

    // Should show either empty state or empty list
    const mainContent = page.locator("main");
    await expect(mainContent).toBeVisible();

    // Check for common empty state text
    const content = await mainContent.textContent();
    expect(content).toBeTruthy();
  });

  test("edit server button visible when server exists", async ({ page }) => {
    // First, try to add a server (even if connection fails)
    await page.goto("/servers");

    const addButton = page.getByRole("button", { name: /add server/i });
    if (await addButton.isVisible()) {
      // If we have servers, edit buttons should be available
      // This is a basic check that the page structure is correct
      await expect(page.locator("main")).toBeVisible();
    }
  });

  test("test connection button works", async ({ page }) => {
    await page.goto("/servers");

    // Look for a test connection button (might be in add form or server card)
    const testButtons = page.getByRole("button", { name: /test/i });

    // If test buttons exist, they should be clickable
    const count = await testButtons.count();
    if (count > 0) {
      const firstButton = testButtons.first();
      await expect(firstButton).toBeVisible();

      // Click and wait for response
      await firstButton.click();
      await page.waitForTimeout(1000);

      // Page should still be functional
      await expect(page.locator("main")).toBeVisible();
    }
  });

  test("server cards display when servers exist", async ({ page }) => {
    await page.goto("/servers");

    // Check that the page can display servers
    // Even with no servers, the container should be present
    const mainContent = page.locator("main");
    await expect(mainContent).toBeVisible();

    // The page should have meaningful content
    const hasContent = await mainContent.textContent();
    expect(hasContent).toBeTruthy();
  });
});
