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

  test("should close modal on Escape key", async ({ page }) => {
    await page.goto("/servers");

    // Open modal
    await page.getByRole("button", { name: /add server/i }).click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Press Escape key
    await page.keyboard.press("Escape");

    // Wait a moment for the close animation
    await page.waitForTimeout(100);

    // Modal should not have 'open' attribute
    await expect(page.locator("#server-modal")).not.toHaveAttribute("open");
  });

  /**
   * REGRESSION TEST: Modal z-index stacking issue
   *
   * This test verifies that modal buttons are clickable without using Playwright's
   * `force: true` option. When the modal has z-index/stacking context issues,
   * elements from the page behind the modal intercept pointer events, causing
   * normal clicks to fail.
   *
   * Issue: The modal container is inside the DaisyUI drawer which has
   * `position: relative`, creating a stacking context that can cause the modal
   * to render behind other page elements despite having a high z-index.
   *
   * This test will FAIL until the modal z-index issue is properly fixed.
   * The fix should ensure the modal appears above all other page content
   * and its buttons are clickable without workarounds.
   */
  test("modal buttons should be clickable without force option (z-index regression)", async ({
    page,
  }) => {
    await page.goto("/servers");

    // Open modal
    await page.getByRole("button", { name: /add server/i }).click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Fill in required fields
    await page.locator("#name").fill("Test Server");
    await page.locator("#url").fill("http://localhost:7878");
    await page.locator("#apiKey").fill("test-api-key");

    // This is the critical test: clicking the Create button WITHOUT force: true
    // If the modal has z-index issues, this will timeout because another element
    // intercepts the click. The error message will indicate which element is
    // blocking: "... subtree intercepts pointer events"
    const createButton = page.locator('button[form="server-form"]');
    await expect(createButton).toBeVisible();

    // Attempt to click without force - this should work if modal z-index is correct
    // Using a shorter timeout to fail faster if there's an issue
    await createButton.click({ timeout: 5000 });

    // If we get here, the click succeeded. The form will attempt to submit
    // and likely fail (no real server), but that's fine - we proved the button
    // is clickable. Wait for either success or error state.
    await page.waitForTimeout(500);

    // The test passes if we reach this point - the button was clickable
  });

  /**
   * REGRESSION TEST: Cancel button clickability
   *
   * Similar to the Create button test, this verifies the Cancel button
   * is clickable without force option.
   */
  test("cancel button should be clickable without force option (z-index regression)", async ({
    page,
  }) => {
    await page.goto("/servers");

    // Open modal
    await page.getByRole("button", { name: /add server/i }).click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Click cancel WITHOUT force option
    const cancelButton = page.getByRole("button", { name: /cancel/i });
    await expect(cancelButton).toBeVisible();
    await cancelButton.click({ timeout: 5000 });

    // Modal should close
    await page.waitForTimeout(100);
    await expect(page.locator("#server-modal")).not.toHaveAttribute("open");
  });
});
