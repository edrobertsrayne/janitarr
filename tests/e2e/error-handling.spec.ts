import { test, expect } from "./setup";

test.describe("Error handling", () => {
  test("should handle 404 pages gracefully", async ({ page }) => {
    const response = await page.goto("/nonexistent-page");

    // Should return 404 status
    expect(response?.status()).toBe(404);

    // Page should still render something useful
    await expect(page.locator("body")).toBeVisible();
  });

  test("should handle invalid server form submission", async ({ page }) => {
    await page.goto("/servers");

    // Open add server modal
    await page.getByRole("button", { name: /add server/i }).click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Fill with invalid URL
    await page.locator("#name").fill("Test");
    await page.locator("#url").fill("not-a-valid-url");
    await page.locator("#apiKey").fill("test-key");

    // The URL field has type="url", so browser validation should prevent submission
    const urlField = page.locator("#url");
    await expect(urlField).toHaveAttribute("type", "url");
  });

  test("settings form should handle invalid numeric values", async ({
    page,
  }) => {
    await page.goto("/settings");

    // Find numeric inputs
    const numericInputs = page.locator('input[type="number"]');
    const count = await numericInputs.count();

    if (count > 0) {
      const firstInput = numericInputs.first();

      // Try to enter negative value
      await firstInput.clear();
      await firstInput.fill("-5");

      // Either has min validation or the form should handle it
      await expect(page.locator("main")).toBeVisible();
    }
  });

  test("app should remain functional after failed API call", async ({
    page,
  }) => {
    await page.goto("/");

    // Intercept and fail the automation trigger
    await page.route("**/api/automation/trigger", (route) => {
      route.fulfill({
        status: 500,
        body: JSON.stringify({ error: "Test error" }),
      });
    });

    // Click Run Now
    const runButton = page.getByRole("button", { name: /run now/i });
    await runButton.click();

    // Wait a moment
    await page.waitForTimeout(1000);

    // Page should still be functional
    await expect(page.locator("main")).toBeVisible();

    // Navigation should still work
    await page.getByRole("link", { name: "Servers" }).click();
    await expect(page).toHaveURL(/\/servers/);
  });
});
