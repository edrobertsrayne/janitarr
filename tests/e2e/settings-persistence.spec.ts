import { test, expect } from "./setup";

test.describe("Settings persistence", () => {
  test("should save interval hours successfully", async ({ page }) => {
    await page.goto("/settings");

    // Find interval input by ID
    const intervalInput = page.locator("#interval");
    await expect(intervalInput).toBeVisible();

    // Change to a different value (12 hours)
    await intervalInput.clear();
    await intervalInput.fill("12");

    // Save settings
    const saveButton = page.getByRole("button", { name: /save settings/i });
    await saveButton.click();

    // Verify save succeeded
    await expect(page.getByText(/settings saved successfully/i)).toBeVisible({
      timeout: 10000,
    });
  });

  test("should save scheduler enabled state", async ({ page }) => {
    await page.goto("/settings");

    // Find scheduler enabled checkbox by ID
    const enabledCheckbox = page.locator("#enabled");
    await expect(enabledCheckbox).toBeVisible();

    // Get current state and toggle it
    const wasChecked = await enabledCheckbox.isChecked();
    await enabledCheckbox.click();

    // Verify it toggled
    const isNowChecked = await enabledCheckbox.isChecked();
    expect(isNowChecked).toBe(!wasChecked);

    // Save
    const saveButton = page.getByRole("button", { name: /save settings/i });
    await saveButton.click();

    // Verify save succeeded
    await expect(page.getByText(/settings saved successfully/i)).toBeVisible();
  });

  test("should save and update search limits", async ({ page }) => {
    await page.goto("/settings");

    // Test all four search limit fields
    const fields = [
      { id: "#missing-movies", value: "50" },
      { id: "#missing-episodes", value: "75" },
      { id: "#cutoff-movies", value: "30" },
      { id: "#cutoff-episodes", value: "40" },
    ];

    for (const field of fields) {
      const input = page.locator(field.id);
      await expect(input).toBeVisible();
      await input.clear();
      await input.fill(field.value);
    }

    // Save settings
    const saveButton = page.getByRole("button", { name: /save settings/i });
    await saveButton.click();

    // Verify save succeeded
    await expect(page.getByText(/settings saved successfully/i)).toBeVisible();

    // Verify values are still set after save
    for (const field of fields) {
      await expect(page.locator(field.id)).toHaveValue(field.value);
    }
  });

  test("should save log retention period", async ({ page }) => {
    await page.goto("/settings");

    // Find retention days select
    const retentionSelect = page.locator("#retention-days");
    await expect(retentionSelect).toBeVisible();

    // Change to a different value (14 days)
    await retentionSelect.selectOption("14");
    await expect(retentionSelect).toHaveValue("14");

    // Save settings
    const saveButton = page.getByRole("button", { name: /save settings/i });
    await saveButton.click();

    // Verify save succeeded
    await expect(page.getByText(/settings saved successfully/i)).toBeVisible();

    // Verify value is still set after save
    await expect(retentionSelect).toHaveValue("14");
  });

  test("should show save confirmation feedback", async ({ page }) => {
    await page.goto("/settings");

    // Find save button and click it
    const saveButton = page.getByRole("button", { name: /save settings/i });
    await expect(saveButton).toBeVisible();

    await saveButton.click();

    // Should show success message (loading state may be too fast to catch)
    await expect(page.getByText(/settings saved successfully/i)).toBeVisible({
      timeout: 5000,
    });

    // Success message should disappear after 3 seconds
    await expect(
      page.getByText(/settings saved successfully/i),
    ).not.toBeVisible({ timeout: 4000 });
  });

  test("should validate numeric input ranges", async ({ page }) => {
    await page.goto("/settings");

    // Interval field should enforce min/max
    const intervalInput = page.locator("#interval");
    await expect(intervalInput).toHaveAttribute("min", "1");
    await expect(intervalInput).toHaveAttribute("max", "168");
    await expect(intervalInput).toHaveAttribute("type", "number");
    await expect(intervalInput).toHaveAttribute("required");

    // Missing movies field should have limits
    const missingMoviesInput = page.locator("#missing-movies");
    await expect(missingMoviesInput).toHaveAttribute("min", "0");
    await expect(missingMoviesInput).toHaveAttribute("max", "1000");
    await expect(missingMoviesInput).toHaveAttribute("type", "number");
    await expect(missingMoviesInput).toHaveAttribute("required");

    // All numeric inputs should have proper constraints
    const numericInputs = await page.locator('input[type="number"]').all();
    expect(numericInputs.length).toBeGreaterThan(0);
  });

  test("should handle form submission with invalid values gracefully", async ({
    page,
  }) => {
    await page.goto("/settings");

    // Try to set interval to invalid value (out of range)
    const intervalInput = page.locator("#interval");
    await intervalInput.clear();

    // HTML5 validation should prevent submitting empty required field
    const saveButton = page.getByRole("button", { name: /save settings/i });
    await saveButton.click();

    // Form should not submit (no success message due to validation)
    // Page should still be on settings
    await expect(page).toHaveURL(/\/settings/);
  });
});
