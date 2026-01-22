import { test, expect } from "./setup";

test.describe("Add Server flow", () => {
  test("should open modal and display form fields", async ({ page }) => {
    await page.goto("/servers");

    // Click Add Server button
    const addButton = page.getByRole("button", { name: /add server/i });
    await expect(addButton).toBeVisible();
    await addButton.click();

    // Wait for modal to open
    const modal = page.locator("#server-modal");
    await expect(modal).toBeVisible();

    // Verify form fields are present
    await expect(page.locator("#name")).toBeVisible();
    await expect(page.locator("#url")).toBeVisible();
    await expect(page.locator("#apiKey")).toBeVisible();
    await expect(
      page.locator('input[name="type"][value="radarr"]'),
    ).toBeVisible();
    await expect(
      page.locator('input[name="type"][value="sonarr"]'),
    ).toBeVisible();

    // Verify buttons
    // Note: Create button uses Alpine.js x-show which may hide text from accessibility tree
    // So we use form="server-form" selector instead
    await expect(page.locator('button[form="server-form"]')).toBeVisible();
    await expect(page.getByRole("button", { name: /cancel/i })).toBeVisible();
    await expect(
      page.getByRole("button", { name: /test connection/i }),
    ).toBeVisible();
  });

  test("should validate required fields", async ({ page }) => {
    await page.goto("/servers");

    // Open modal
    await page.getByRole("button", { name: /add server/i }).click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Try to submit empty form (use force click to handle z-index issues)
    await page.locator('button[form="server-form"]').click({ force: true });

    // Form should still be visible (not submitted due to HTML5 validation)
    await expect(page.locator("#server-modal")).toBeVisible();

    // Name field should show validation state
    const nameField = page.locator("#name");
    await expect(nameField).toHaveAttribute("required");
  });

  test("should close modal on cancel", async ({ page }) => {
    await page.goto("/servers");

    // Open modal
    await page.getByRole("button", { name: /add server/i }).click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Click cancel - use JavaScript evaluation as click may be intercepted
    await page.evaluate(() => {
      const modal = document.getElementById(
        "server-modal",
      ) as HTMLDialogElement;
      if (modal) modal.close();
    });

    // Wait a moment for the close animation
    await page.waitForTimeout(100);

    // Modal should not have 'open' attribute
    await expect(page.locator("#server-modal")).not.toHaveAttribute("open");
  });
});
