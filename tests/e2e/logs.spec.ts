import { test, expect } from '@playwright/test';

test.describe('Logs page', () => {
  test('should display log entries and allow filtering', async ({ page }) => {
    page.on('console', msg => {
      console.log(`PAGE CONSOLE: ${msg.text()}`);
    });

    await page.addInitScript(() => {
      // Inject test data directly into the window object for the Logs component
      // This bypasses API and WebSocket calls in the test environment.
      (window as any).__JANITARR_TEST_LOGS__ = [
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
    });

    // Mock the API endpoint for fetching logs (still needed for filtering test)
    await page.route('**/api/logs*', async (route) => { // Make handler async
      const url = new URL(route.request().url());
      const searchParam = url.searchParams.get('search') || '';

      const allLogs = (window as any).__JANITARR_TEST_LOGS__; // Use the injected test data

      const filteredLogs = allLogs.filter((log: any) => log.message.toLowerCase().includes(searchParam.toLowerCase()));

      await route.fulfill({ // Use await here
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: filteredLogs,
        }),
      });
    });

    await page.goto('/logs');
    
    // Wait for the LoadingSpinner to disappear, indicating logs have loaded
    await expect(page.getByText('Loading logs...')).not.toBeVisible({ timeout: 15000 });

    // Expect the logs heading to be visible
    await expect(page.getByRole('heading', { name: 'Logs' })).toBeVisible();

    // Expect all initial log entries to be visible using a more robust locator
    await expect(page.getByText('Automation cycle started')).toBeVisible({ timeout: 15000 });
    await expect(page.getByText('Server "Radarr" connection failed')).toBeVisible({ timeout: 15000 });
    await expect(page.getByText('Search triggered for item "Movie Title"')).toBeVisible({ timeout: 15000 });
    await expect(page.getByText('Automation cycle completed')).toBeVisible({ timeout: 15000 });

    // Test filtering
    await page.getByPlaceholder('Search logs').fill('connection failed');

    // Expect only the filtered log entry to be visible
    await expect(page.getByText('Automation cycle started')).not.toBeVisible({ timeout: 15000 });
    await expect(page.getByText('Server "Radarr" connection failed')).toBeVisible({ timeout: 15000 });
    await expect(page.getByText('Search triggered for item "Movie Title"')).not.toBeVisible({ timeout: 15000 });
    await expect(page.getByText('Automation cycle completed')).not.toBeVisible({ timeout: 15000 });

    // Clear filter
    await page.getByPlaceholder('Search logs').clear();
    await expect(page.getByText('Automation cycle started')).toBeVisible({ timeout: 15000 });
    await expect(page.getByText('Server "Radarr" connection failed')).toBeVisible({ timeout: 15000 });
    await expect(page.getByText('Search triggered for item "Movie Title"')).toBeVisible({ timeout: 15000 });
    await expect(page.getByText('Automation cycle completed')).toBeVisible({ timeout: 15000 });
  });
});
