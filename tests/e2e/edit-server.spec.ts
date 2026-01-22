import { test, expect } from "./setup";

test.describe("Edit Server flow", () => {
  test.beforeEach(async ({ page }) => {
    // Ensure at least one server exists by navigating to servers page
    // The test database should have a server from previous test runs
    await page.goto("/servers");
  });

  test("should open edit modal with pre-filled data", async ({ page }) => {
    // Find and click Edit button on first server card
    const editButton = page.getByRole("button", { name: /edit/i }).first();

    // Skip if no servers exist
    if (!(await editButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    await editButton.click();

    // Wait for modal to open
    const modal = page.locator("#server-modal");
    await expect(modal).toBeVisible();

    // Verify modal title indicates edit mode
    await expect(
      page.getByRole("heading", { name: /edit server/i }),
    ).toBeVisible();

    // Verify fields are pre-filled (not empty)
    const nameField = page.locator("#name");
    await expect(nameField).toBeVisible();
    const nameValue = await nameField.inputValue();
    expect(nameValue.length).toBeGreaterThan(0);

    // Verify Update button (not Create)
    await expect(page.getByRole("button", { name: /update/i })).toBeVisible();
  });

  test("should show enabled checkbox in edit mode", async ({ page }) => {
    const editButton = page.getByRole("button", { name: /edit/i }).first();

    if (!(await editButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    await editButton.click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Enabled checkbox should be visible in edit mode
    const enabledCheckbox = page.locator("#enabled");
    await expect(enabledCheckbox).toBeVisible();
  });

  test("should allow editing server name", async ({ page }) => {
    const editButton = page.getByRole("button", { name: /edit/i }).first();

    if (!(await editButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    await editButton.click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Clear and type new name
    const nameField = page.locator("#name");
    await nameField.clear();
    await nameField.fill("Updated Server Name");

    // Verify the value changed
    await expect(nameField).toHaveValue("Updated Server Name");
  });

  test("should close edit modal on cancel without saving", async ({ page }) => {
    const editButton = page.getByRole("button", { name: /edit/i }).first();

    if (!(await editButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    // Get original server name from card
    const serverCard = page.locator(".card").first();
    const originalName = await serverCard.locator(".card-title").textContent();

    await editButton.click();
    await expect(page.locator("#server-modal")).toBeVisible();

    // Change the name
    const nameField = page.locator("#name");
    await nameField.clear();
    await nameField.fill("Should Not Save");

    // Cancel
    await page.getByRole("button", { name: /cancel/i }).click();
    await expect(page.locator("#server-modal")).not.toBeVisible();

    // Original name should still be displayed
    await expect(serverCard.locator(".card-title")).toHaveText(originalName!);
  });
});
