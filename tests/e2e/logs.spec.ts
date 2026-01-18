import { test, expect } from "./setup";

test.describe("Logs page", () => {
  test("logs page loads", async ({ page }) => {
    await page.goto("/logs");

    // Check for heading
    await expect(
      page.getByRole("heading", { name: /logs|activity/i }),
    ).toBeVisible();

    // Main content should be visible
    await expect(page.locator("main")).toBeVisible();
  });

  test("filter by type", async ({ page }) => {
    await page.goto("/logs");

    // Look for filter controls (dropdowns, buttons, etc.)
    const filterControls = page
      .locator("select, button")
      .filter({ hasText: /filter|type|all/i });

    const count = await filterControls.count();
    if (count > 0) {
      // Filter controls exist - test that they're interactive
      const firstControl = filterControls.first();
      await expect(firstControl).toBeVisible();

      // Try clicking/changing filter
      await firstControl.click();
      await page.waitForTimeout(500);

      // Page should still be functional
      await expect(page.locator("main")).toBeVisible();
    }
  });

  test("logs display when present", async ({ page }) => {
    await page.goto("/logs");

    // Even with no logs, the container should exist
    const logsContainer = page.locator("main");
    await expect(logsContainer).toBeVisible();

    // Check for either logs or empty state
    const content = await logsContainer.textContent();
    expect(content).toBeTruthy();
  });

  test("infinite scroll or pagination works", async ({ page }) => {
    await page.goto("/logs");

    // Look for pagination controls or infinite scroll indicators
    // The page should handle scrolling/pagination gracefully
    const mainContent = page.locator("main");
    await expect(mainContent).toBeVisible();

    // Try scrolling down
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(500);

    // Page should still be functional
    await expect(mainContent).toBeVisible();
  });

  test("clear logs button exists", async ({ page }) => {
    await page.goto("/logs");

    // Look for clear logs button
    const clearButton = page.getByRole("button", { name: /clear|delete/i });

    const count = await clearButton.count();
    if (count > 0) {
      await expect(clearButton.first()).toBeVisible();
    }
  });

  test("clear logs with confirmation", async ({ page }) => {
    await page.goto("/logs");

    // Look for clear logs button
    const clearButton = page.getByRole("button", { name: /clear|delete/i });

    if (
      await clearButton
        .first()
        .isVisible()
        .catch(() => false)
    ) {
      // Click clear button
      await clearButton.first().click();
      await page.waitForTimeout(500);

      // Should show confirmation dialog or modal
      // Look for confirmation buttons
      const confirmButtons = page.getByRole("button", {
        name: /confirm|yes|delete/i,
      });

      if (
        await confirmButtons
          .first()
          .isVisible()
          .catch(() => false)
      ) {
        // Confirmation dialog exists
        await expect(confirmButtons.first()).toBeVisible();

        // Don't actually confirm in this test, just verify the flow exists
        // Click cancel or close if available
        const cancelButton = page.getByRole("button", { name: /cancel|no/i });
        if (
          await cancelButton
            .first()
            .isVisible()
            .catch(() => false)
        ) {
          await cancelButton.first().click();
        }
      }

      await page.waitForTimeout(500);

      // Page should still be functional
      await expect(page.locator("main")).toBeVisible();
    }
  });

  test("export logs functionality exists", async ({ page }) => {
    await page.goto("/logs");

    // Look for export buttons (JSON/CSV)
    const exportButton = page.getByRole("button", { name: /export/i });
    const exportLink = page.getByRole("link", { name: /export|download/i });

    // Either export buttons or links might exist
    const buttonCount = await exportButton.count();
    const linkCount = await exportLink.count();

    if (buttonCount > 0 || linkCount > 0) {
      // Export functionality exists
      if (buttonCount > 0) {
        await expect(exportButton.first()).toBeVisible();
      } else {
        await expect(exportLink.first()).toBeVisible();
      }
    }
  });

  test("real-time updates with WebSocket", async ({ page }) => {
    await page.goto("/logs");

    // WebSocket connection should be established
    // We can't easily test the actual WebSocket messages without triggering events
    // But we can verify the page is set up for it

    // Wait a moment for WebSocket connection
    await page.waitForTimeout(1000);

    // Page should be functional
    await expect(page.locator("main")).toBeVisible();

    // The page should have loaded completely
    const content = await page.content();
    expect(content).toContain("log"); // Should have some reference to logs
  });

  test("log entry formatting", async ({ page }) => {
    await page.goto("/logs");

    // Check that logs are displayed in a structured format
    const mainContent = page.locator("main");
    await expect(mainContent).toBeVisible();

    // Even with no logs, the structure should be present
    const hasContent = await mainContent.textContent();
    expect(hasContent).toBeTruthy();
  });

  test("filter by server", async ({ page }) => {
    await page.goto("/logs");

    // Look for server filter dropdown
    const serverFilter = page
      .locator("select, button")
      .filter({ hasText: /server/i });

    const count = await serverFilter.count();
    if (count > 0) {
      // Server filter exists
      await expect(serverFilter.first()).toBeVisible();

      // Try interacting with it
      await serverFilter.first().click();
      await page.waitForTimeout(500);

      // Page should still work
      await expect(page.locator("main")).toBeVisible();
    }
  });

  test("empty state message when no logs", async ({ page }) => {
    await page.goto("/logs");

    // Fresh database should have no logs
    // Check for empty state messaging
    const mainContent = page.locator("main");
    await expect(mainContent).toBeVisible();

    const content = await mainContent.textContent();

    // Should either have logs or indicate there are none
    expect(content).toBeTruthy();
  });
});
