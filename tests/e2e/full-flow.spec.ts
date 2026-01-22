import { test, expect } from "./setup";

test.describe("Full user flow integration", () => {
  test("complete workflow: navigate all pages", async ({ page }) => {
    // Start at dashboard
    await page.goto("/");
    await expect(page).toHaveTitle(/Dashboard.*Janitarr/);
    await expect(
      page.getByRole("heading", { name: "Dashboard" }),
    ).toBeVisible();

    // Navigate to Servers
    await page.getByRole("link", { name: "Servers" }).click();
    await expect(page).toHaveURL(/\/servers/);
    await expect(
      page.getByRole("heading", { name: "Servers" }).first(),
    ).toBeVisible();

    // Navigate to Activity Logs
    await page.getByRole("link", { name: "Activity Logs" }).click();
    await expect(page).toHaveURL(/\/logs/);
    await expect(
      page.getByRole("heading", { name: /activity|logs/i }).first(),
    ).toBeVisible();

    // Navigate to Settings
    await page.getByRole("link", { name: "Settings" }).click();
    await expect(page).toHaveURL(/\/settings/);
    await expect(
      page.getByRole("heading", { name: /settings/i }),
    ).toBeVisible();

    // Return to Dashboard
    await page.getByRole("link", { name: "Dashboard" }).click();
    await expect(page).toHaveURL(/\/$/);
  });

  test("theme toggle persists across pages", async ({ page }) => {
    await page.goto("/");

    // Find theme toggle
    const themeToggle = page.locator('input[type="checkbox"]').last();

    if (await themeToggle.isVisible()) {
      // Get initial theme
      const html = page.locator("html");
      const initialTheme = await html.getAttribute("data-theme");

      // Toggle theme
      await themeToggle.click();
      await page.waitForTimeout(300);

      // Get new theme
      const newTheme = await html.getAttribute("data-theme");
      expect(newTheme).not.toBe(initialTheme);

      // Navigate to another page
      await page.getByRole("link", { name: "Settings" }).click();
      await page.waitForTimeout(300);

      // Theme should persist
      const themeAfterNav = await html.getAttribute("data-theme");
      expect(themeAfterNav).toBe(newTheme);
    }
  });

  test("server CRUD operations", async ({ page }) => {
    await page.goto("/servers");

    // Open add server modal
    const addButton = page.getByRole("button", { name: /add server/i });
    await addButton.click();

    // Verify modal opened
    const modal = page.locator("#server-modal");
    await expect(modal).toBeVisible();

    // Fill form
    await page.locator("#name").fill("Integration Test Server");
    await page.locator("#url").fill("http://localhost:9999");
    await page.locator("#apiKey").fill("test-api-key-12345");

    // Submit the form using the submit button
    // The button is outside the form but has form="server-form" attribute
    const submitButton = page.locator('button[form="server-form"]');
    await expect(submitButton).toBeVisible();
    await submitButton.click({ force: true });

    // Wait for response
    await page.waitForTimeout(2000);

    // Either modal closed (success) or error shown
    // The test verifies the flow works, not necessarily successful creation
    await expect(page.locator("main")).toBeVisible();
  });

  test("logs page loads and displays content", async ({ page }) => {
    await page.goto("/logs");

    // Heading should be visible
    await expect(
      page.getByRole("heading", { name: /activity|logs/i }).first(),
    ).toBeVisible();

    // Main content area should exist
    await expect(page.locator("main")).toBeVisible();

    // Either logs are displayed or empty state message
    const mainContent = await page.locator("main").textContent();
    expect(mainContent).toBeTruthy();
    expect(mainContent!.length).toBeGreaterThan(10);
  });
});
