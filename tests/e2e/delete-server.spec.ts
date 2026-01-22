import { test, expect } from "./setup";

test.describe("Delete Server flow", () => {
  test("should show DaisyUI confirmation modal", async ({ page }) => {
    await page.goto("/servers");

    // Find delete button
    const deleteButton = page.getByRole("button", { name: /delete/i }).first();

    if (!(await deleteButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    await deleteButton.click();

    // DaisyUI modal should appear (not browser confirm)
    const modal = page.locator(".modal-open, dialog[open]");
    await expect(modal).toBeVisible();

    // Modal should have confirmation text
    await expect(page.getByText(/are you sure/i)).toBeVisible();

    // Modal should have Cancel and Delete buttons
    await expect(page.getByRole("button", { name: /cancel/i })).toBeVisible();
    await expect(modal.getByRole("button", { name: /delete/i })).toBeVisible();
  });

  test("should close modal on cancel without deleting", async ({ page }) => {
    await page.goto("/servers");

    const deleteButton = page.getByRole("button", { name: /delete/i }).first();

    if (!(await deleteButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    // Count servers before
    const serverCountBefore = await page.locator(".card").count();

    await deleteButton.click();

    // Wait for modal
    const modal = page.locator(".modal-open, dialog[open]");
    await expect(modal).toBeVisible();

    // Click cancel
    await page.getByRole("button", { name: /cancel/i }).click();

    // Modal should close
    await expect(modal).not.toBeVisible();

    // Server count should be unchanged
    const serverCountAfter = await page.locator(".card").count();
    expect(serverCountAfter).toBe(serverCountBefore);
  });

  test("should delete server on confirm", async ({ page }) => {
    await page.goto("/servers");

    const deleteButton = page.getByRole("button", { name: /delete/i }).first();

    if (!(await deleteButton.isVisible().catch(() => false))) {
      test.skip();
      return;
    }

    // Count servers before
    const serverCountBefore = await page.locator(".card").count();

    await deleteButton.click();

    // Wait for modal and confirm delete
    const modal = page.locator(".modal-open, dialog[open]");
    await expect(modal).toBeVisible();

    // Click the delete button inside the modal
    await modal.getByRole("button", { name: /delete/i }).click();

    // Wait for deletion to complete
    await page.waitForTimeout(1500);

    // Server count should decrease by 1
    const serverCountAfter = await page.locator(".card").count();
    expect(serverCountAfter).toBe(serverCountBefore - 1);
  });
});
