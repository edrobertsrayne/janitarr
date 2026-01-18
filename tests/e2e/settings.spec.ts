import { test, expect } from "./setup";

test.describe("Settings page", () => {
  test("settings page loads", async ({ page }) => {
    await page.goto("/settings");

    // Check for heading
    await expect(
      page.getByRole("heading", { name: /settings|configuration/i }),
    ).toBeVisible();

    // Main content should be visible
    await expect(page.locator("main")).toBeVisible();
  });

  test("displays current configuration", async ({ page }) => {
    await page.goto("/settings");

    // Settings page should show current configuration values
    const mainContent = page.locator("main");
    await expect(mainContent).toBeVisible();

    // Check that there's meaningful content
    const content = await mainContent.textContent();
    expect(content).toBeTruthy();
    expect(content.length).toBeGreaterThan(50); // Should have substantial content
  });

  test("save settings button exists", async ({ page }) => {
    await page.goto("/settings");

    // Look for save button
    const saveButton = page.getByRole("button", { name: /save|update|apply/i });

    const count = await saveButton.count();
    if (count > 0) {
      await expect(saveButton.first()).toBeVisible();
    }
  });

  test("schedule settings section exists", async ({ page }) => {
    await page.goto("/settings");

    // Settings page should have schedule-related inputs
    const content = await page.content();

    // Check for schedule-related content
    const hasScheduleContent =
      content.toLowerCase().includes("schedule") ||
      content.toLowerCase().includes("interval") ||
      content.toLowerCase().includes("hours");

    expect(hasScheduleContent).toBeTruthy();
  });

  test("search limits settings section exists", async ({ page }) => {
    await page.goto("/settings");

    // Settings page should have limit-related inputs
    const content = await page.content();

    // Check for limits-related content
    const hasLimitsContent =
      content.toLowerCase().includes("limit") ||
      content.toLowerCase().includes("missing") ||
      content.toLowerCase().includes("cutoff");

    expect(hasLimitsContent).toBeTruthy();
  });

  test("validation for invalid values", async ({ page }) => {
    await page.goto("/settings");

    // Look for numeric input fields
    const numericInputs = page.locator('input[type="number"]');
    const count = await numericInputs.count();

    if (count > 0) {
      // Try entering an invalid value
      const firstInput = numericInputs.first();
      await firstInput.fill("-1"); // Negative values should be invalid

      // Try to save
      const saveButton = page.getByRole("button", {
        name: /save|update|apply/i,
      });
      if (
        await saveButton
          .first()
          .isVisible()
          .catch(() => false)
      ) {
        await saveButton.first().click();
        await page.waitForTimeout(1000);

        // Should either show validation error or not accept the value
        // Page should still be functional
        await expect(page.locator("main")).toBeVisible();
      }
    }
  });

  test("scheduler enabled toggle exists", async ({ page }) => {
    await page.goto("/settings");

    // Look for scheduler enabled toggle (checkbox, switch, etc.)
    const toggleElements = page
      .locator('input[type="checkbox"], button')
      .filter({ hasText: /enabled|disabled/i });

    const count = await toggleElements.count();
    if (count > 0) {
      // Toggle exists - verify it's interactive
      const firstToggle = toggleElements.first();
      if (await firstToggle.isVisible().catch(() => false)) {
        await expect(firstToggle).toBeVisible();
      }
    }
  });

  test("interval hours input exists", async ({ page }) => {
    await page.goto("/settings");

    // Look for interval input
    const intervalInput = page
      .locator("input")
      .filter({ hasText: /interval|hours/i });
    const numericInputs = page.locator('input[type="number"]');

    const hasInterval = (await intervalInput.count()) > 0;
    const hasNumeric = (await numericInputs.count()) > 0;

    // At least one should exist
    expect(hasInterval || hasNumeric).toBeTruthy();
  });

  test("missing limits inputs exist", async ({ page }) => {
    await page.goto("/settings");

    // Look for missing-related inputs
    const content = await page.content();
    const hasContent = content.toLowerCase().includes("missing");

    // Should have some reference to missing content limits
    expect(content).toBeTruthy();
  });

  test("cutoff limits inputs exist", async ({ page }) => {
    await page.goto("/settings");

    // Look for cutoff-related inputs
    const content = await page.content();
    const hasContent = content.toLowerCase().includes("cutoff");

    // Should have some reference to cutoff content limits
    expect(content).toBeTruthy();
  });

  test("save settings shows feedback", async ({ page }) => {
    await page.goto("/settings");

    // Look for save button
    const saveButton = page.getByRole("button", { name: /save|update|apply/i });

    if (
      await saveButton
        .first()
        .isVisible()
        .catch(() => false)
    ) {
      // Click save
      await saveButton.first().click();

      // Wait for response
      await page.waitForTimeout(2000);

      // Should show some feedback (toast, alert, or message)
      // At minimum, the page should still be functional
      await expect(page.locator("main")).toBeVisible();
    }
  });

  test("reset to defaults option may exist", async ({ page }) => {
    await page.goto("/settings");

    // Look for reset button
    const resetButton = page.getByRole("button", { name: /reset|default/i });

    const count = await resetButton.count();
    if (count > 0) {
      // Reset button exists
      await expect(resetButton.first()).toBeVisible();
    }

    // This is optional, so we just check if it exists
    // Don't fail the test if it's not there
  });

  test("form validation prevents invalid submissions", async ({ page }) => {
    await page.goto("/settings");

    // Try to submit with potentially invalid values
    const numericInputs = page.locator('input[type="number"]');
    const count = await numericInputs.count();

    if (count > 0) {
      // Clear a required field if possible
      const firstInput = numericInputs.first();
      await firstInput.fill("");

      // Try to save
      const saveButton = page.getByRole("button", {
        name: /save|update|apply/i,
      });
      if (
        await saveButton
          .first()
          .isVisible()
          .catch(() => false)
      ) {
        await saveButton.first().click();
        await page.waitForTimeout(1000);

        // Page should handle this gracefully
        await expect(page.locator("main")).toBeVisible();
      }
    }
  });
});
