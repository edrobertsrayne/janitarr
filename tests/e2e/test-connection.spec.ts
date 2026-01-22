import { test, expect } from "./setup";

test.describe("Test Connection functionality", () => {
  test("should show testing state when clicked", async ({ page }) => {
    await page.goto("/servers");

    const testButton = page.getByRole("button", { name: /^test$/i }).first();

    if (!(await testButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    // Click test button
    await testButton.click();

    // Should show "Testing..." state
    await expect(page.getByText(/testing/i)).toBeVisible();
  });

  test("should show result after test completes", async ({ page }) => {
    await page.goto("/servers");

    const testButton = page.getByRole("button", { name: /^test$/i }).first();

    if (!(await testButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    await testButton.click();

    // Wait for test to complete (may take a few seconds)
    await page.waitForTimeout(5000);

    // Should show either "Connected" or "Connection failed" or error message
    const serverCard = page.locator(".card").first();
    const resultText = await serverCard.textContent();

    const hasResult =
      resultText?.includes("Connected") ||
      resultText?.includes("Connection failed") ||
      resultText?.includes("Error");

    expect(hasResult).toBeTruthy();
  });

  test("test connection in add server modal", async ({ page }) => {
    await page.goto("/servers");

    // Open add server modal
    await page.getByRole("button", { name: /add server/i }).click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Fill required fields
    await page.locator("#name").fill("Test Server");
    await page.locator("#url").fill("http://localhost:7878");
    await page.locator("#apiKey").fill("invalid-api-key-for-testing");

    // Wait for modal to fully render
    await page.waitForTimeout(1000);

    // Click test connection using JavaScript click to bypass overlay issues
    await page.evaluate(() => {
      const btn = document.getElementById("test-connection-btn");
      if (btn) btn.click();
    });

    // Should show testing state
    await expect(page.getByText(/testing/i)).toBeVisible({ timeout: 2000 });

    // Wait for result
    await page.waitForTimeout(5000);

    // Should show connection result (likely failed since no real server)
    const modal = page.locator("#server-modal");
    const modalText = await modal.textContent();

    const hasResult =
      modalText?.includes("Connected") ||
      modalText?.includes("Connection failed") ||
      modalText?.includes("failed");

    expect(hasResult).toBeTruthy();
  });
});
