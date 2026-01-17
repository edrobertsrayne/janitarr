import { test, expect } from '@playwright/test';

test.describe('Logs page', () => {
  test('should display log entries and allow filtering', async ({ page }) => {
    // Mock the API endpoint for fetching logs
    await page.route('**/api/logs*', async (route) => { // Make handler async
      const url = new URL(route.request().url());
      const searchParam = url.searchParams.get('search') || '';

      const allLogs = [
        {
          id: 'log1',
          timestamp: new Date().toISOString(),
          level: 'info',
          message: 'Automation cycle started',
          category: 'automation',
          details: {},
        },
        {
          id: 'log2',
          timestamp: new Date().toISOString(),
          level: 'warn',
          message: 'Server "Radarr" connection failed',
          category: 'server',
          details: { serverName: 'Radarr' },
        },
        {
          id: 'log3',
          timestamp: new Date().toISOString(),
          level: 'info',
          message: 'Search triggered for item "Movie Title"',
          category: 'search',
          details: { itemName: 'Movie Title' },
        },
        {
          id: 'log4',
          timestamp: new Date().toISOString(),
          level: 'info',
          message: 'Automation cycle completed',
          category: 'automation',
          details: {},
        },
      ];

      const filteredLogs = allLogs.filter(log => log.message.toLowerCase().includes(searchParam.toLowerCase()));

      await route.fulfill({ // Use await here
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: filteredLogs,
        }),
      });
    });

    page.on('console', msg => {
      console.log(`PAGE CONSOLE: ${msg.text()}`);
    });

    await page.goto('/logs');
    // Wait for the GET /api/logs?... request to complete
    await page.waitForResponse(response => response.url().includes('/api/logs') && response.request().method() === 'GET');
    await page.waitForLoadState('domcontentloaded'); // Ensure DOM is fully parsed
    await page.waitForLoadState('domcontentloaded'); // Ensure DOM is fully parsed

    // Expect the logs heading to be visible
    await expect(page.getByRole('heading', { name: 'Logs' })).toBeVisible();

    // Expect all initial log entries to be visible using a more robust locator
    await expect(page.getByRole('listitem', { name: /Automation cycle started/ })).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('listitem', { name: /Server "Radarr" connection failed/ })).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('listitem', { name: /Search triggered for item "Movie Title"/ })).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('listitem', { name: /Automation cycle completed/ })).toBeVisible({ timeout: 15000 });

    // Test filtering
    await page.getByPlaceholder('Search logs').fill('connection failed');

    // Expect only the filtered log entry to be visible
    await expect(page.getByRole('listitem', { name: /Automation cycle started/ })).not.toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('listitem', { name: /Server "Radarr" connection failed/ })).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('listitem', { name: /Search triggered for item "Movie Title"/ })).not.toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('listitem', { name: /Automation cycle completed/ })).not.toBeVisible({ timeout: 15000 });

    // Clear filter
    await page.getByPlaceholder('Search logs').clear();
    await expect(page.getByRole('listitem', { name: /Automation cycle started/ })).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('listitem', { name: /Server "Radarr" connection failed/ })).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('listitem', { name: /Search triggered for item "Movie Title"/ })).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('listitem', { name: /Automation cycle completed/ })).toBeVisible({ timeout: 15000 });
  });
});
