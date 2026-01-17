import { test, expect } from '@playwright/test';

test.describe('Add Server flow', () => {
  test('should allow a user to add a new server', async ({ page }) => {
    let serverAdded = false;

    // Mock the API endpoints
    await page.route('**/api/servers**', async (route) => { // Use async route handler
      const request = route.request();
      const method = request.method();
      const url = request.url();

      console.log('Intercepted:', method, url, 'serverAdded:', serverAdded);

      if (url.endsWith('/test')) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true, data: { success: true, message: 'Connection successful' } }),
        });
        return;
      }

      if (method === 'POST') {
        serverAdded = true;
        console.log('Server added, serverAdded is now true');
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: 'test-id',
              name: 'Test Server',
              type: 'radarr',
              url: 'http://localhost:7878',
              apiKey: 'test-api-key',
              enabled: true,
            },
          }),
        });
        return;
      }

      if (method === 'GET') {
        // Return an empty array if no server has been added yet,
        // otherwise return the new server.
        console.log('Returning server list based on serverAdded:', serverAdded);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: serverAdded ? [
              {
                id: 'test-id',
                name: 'Test Server',
                type: 'radarr',
                url: 'http://localhost:7878',
                apiKey: 'test-api-key',
                enabled: true,
              },
            ] : [],
          }),
        });
        return;
      }

      // Important: if no conditions are met, continue the request to the network
      await route.continue();
    });

    page.on('console', msg => {
      console.log(`PAGE CONSOLE: ${msg.text()}`);
    });

    await page.goto('/servers');
    await page.waitForURL('/servers'); // Wait for navigation to complete

    await page.getByRole('button', { name: 'Add Server' }).click();

    await page.getByLabel('Name').fill('Test Server');
    await page.getByLabel('Radarr').check();
    await page.getByLabel('URL').fill('http://localhost:7878');
    await page.getByLabel('API Key').fill('test-api-key');

    await page.getByRole('button', { name: 'Test Connection' }).click();

    await expect(page.getByText('Connection successful')).toBeVisible({ timeout: 15000 }); // Increased timeout

    await page.getByRole('button', { name: 'Create' }).click();

    // Switch to Card view for better heading detection
    await page.getByRole('button', { name: 'Card view' }).click();
    // Wait for the UI to re-fetch the server list after creation.
    // This needs to be robust as networkidle was problematic.
    // Instead of waiting for a specific network response, wait for the UI element to appear.
    await expect(page.getByRole('heading', { name: 'Test Server' })).toBeVisible({ timeout: 15000 }); // Increased timeout
  });
});