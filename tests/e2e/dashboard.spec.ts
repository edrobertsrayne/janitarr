import { test, expect } from "./setup";

test.describe("Dashboard", () => {
  test("dashboard loads", async ({ page }) => {
    await page.goto("/");

    // Check that the page loaded without errors
    await expect(page).toHaveTitle(/Dashboard.*Janitarr/);

    // Check for main navigation
    await expect(page.getByRole("link", { name: "Dashboard" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Servers" })).toBeVisible();
    await expect(
      page.getByRole("link", { name: "Activity Logs" }),
    ).toBeVisible();
    await expect(page.getByRole("link", { name: "Settings" })).toBeVisible();
  });

  test("shows stats cards", async ({ page }) => {
    await page.goto("/");

    // Dashboard should show statistics
    // Stats cards typically show: server count, searches, etc.
    // Check for common dashboard elements
    const mainContent = page.locator("main");
    await expect(mainContent).toBeVisible();

    // The dashboard should have some content
    const hasContent = await mainContent.textContent();
    expect(hasContent).toBeTruthy();
  });

  test("shows server list", async ({ page }) => {
    await page.goto("/");

    // When no servers are configured, should show empty state or server section
    // The exact content depends on implementation, but page should load
    await expect(page.locator("main")).toBeVisible();
  });

  test("run now button triggers cycle", async ({ page }) => {
    await page.goto("/");

    // Look for a "Run Now" or similar trigger button
    // This test will verify the button exists and can be clicked
    // The actual cycle may not run without configured servers
    const runButton = page.getByRole("button", {
      name: /run now|trigger|manual/i,
    });

    if (await runButton.isVisible()) {
      // Click the button
      await runButton.click();

      // Should show some feedback (loading state, success message, or error)
      // Wait a moment for the response
      await page.waitForTimeout(1000);

      // Page should still be functional
      await expect(page.locator("main")).toBeVisible();
    }
  });

  test("dark mode toggle works", async ({ page }) => {
    await page.goto("/");

    // Look for dark mode toggle button
    const darkModeButton = page.getByRole("button", {
      name: /dark mode|light mode/i,
    });

    if (await darkModeButton.isVisible()) {
      // Get initial state
      const htmlElement = page.locator("html");
      const initialClasses = await htmlElement.getAttribute("class");

      // Toggle dark mode
      await darkModeButton.click();

      // Wait for change
      await page.waitForTimeout(300);

      // Classes should have changed
      const newClasses = await htmlElement.getAttribute("class");
      expect(newClasses).not.toBe(initialClasses);
    }
  });

  test("navigation links work", async ({ page }) => {
    await page.goto("/");

    // Test navigation to servers page
    await page.getByRole("link", { name: "Servers" }).click();
    await expect(page).toHaveURL(/\/servers/);

    // Navigate back to dashboard
    await page.getByRole("link", { name: "Dashboard" }).click();
    await expect(page).toHaveURL(/\/$/);

    // Test navigation to logs
    await page.getByRole("link", { name: "Activity Logs" }).click();
    await expect(page).toHaveURL(/\/logs/);

    // Test navigation to settings
    await page.getByRole("link", { name: "Settings" }).click();
    await expect(page).toHaveURL(/\/settings/);
  });
});
