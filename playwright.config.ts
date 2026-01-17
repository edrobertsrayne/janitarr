import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for Janitarr UI testing
 * Uses headless Chromium provided by devenv
 */
export default defineConfig({
  // Test directory - only E2E tests
  testDir: './tests/e2e/',

  // Maximum time one test can run
  timeout: 30 * 1000,

  // Run tests in files in parallel
  fullyParallel: true,

  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,

  // Retry on CI only
  retries: process.env.CI ? 2 : 0,

  // Opt out of parallel tests on CI
  workers: process.env.CI ? 1 : undefined,

  // Reporter to use
  reporter: 'html',

  // Shared settings for all the projects below
  use: {
    // Base URL for page.goto('/')
    baseURL: 'http://localhost:5173', // Point to UI dev server

    // Collect trace when retrying the failed test
    trace: 'on-first-retry',

    // Screenshot on failure
    screenshot: 'only-on-failure',
  },

  // Configure projects for Chromium
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        // Use devenv's Chromium if CHROMIUM_PATH is set
        // Otherwise Playwright will download its own
        ...(process.env.CHROMIUM_PATH && {
          channel: undefined,
          launchOptions: {
            executablePath: process.env.CHROMIUM_PATH,
          },
        }),
      },
    },
  ],

  // Run your local dev server before starting the tests
  webServer: [
    {
      command: 'bun src/index.ts serve', // Backend server
      url: 'http://localhost:3000', // Correct URL to match default server port
      reuseExistingServer: !process.env.CI,
      timeout: 120 * 1000,
      stderr: 'pipe',
      stdout: 'pipe'
    },
    {
      command: 'cd ui && node_modules/.bin/vite', // UI dev server (use direct path to vite)
      url: 'http://localhost:5173', // Vite's default port
      reuseExistingServer: !process.env.CI,
      timeout: 120 * 1000,
      stderr: 'pipe',
      stdout: 'pipe'
    }
  ],
});
