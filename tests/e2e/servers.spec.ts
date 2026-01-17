import { test, expect } from '@playwright/test';

test.describe('Servers page', () => {
  test('should display "Servers" heading', async ({ page }) => {
    await page.goto('/servers');
    await expect(page.getByRole('heading', { name: 'Servers' })).toBeVisible();
  });
});
