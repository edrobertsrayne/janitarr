import { test as base } from "@playwright/test";
import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from "url";

/**
 * E2E test setup for Janitarr
 *
 * This file provides setup and teardown for E2E tests:
 * - Resets the test database between tests
 * - Provides helper functions for common operations
 */

// Get __dirname equivalent for ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Database path for testing
const TEST_DB_PATH = path.join(__dirname, "../../data/janitarr.db");
const TEST_KEY_PATH = path.join(__dirname, "../../data/.janitarr.key");

/**
 * Reset the database by deleting it
 * The server will recreate it on next startup
 */
export function resetDatabase() {
  try {
    if (fs.existsSync(TEST_DB_PATH)) {
      fs.unlinkSync(TEST_DB_PATH);
    }
    if (fs.existsSync(TEST_KEY_PATH)) {
      fs.unlinkSync(TEST_KEY_PATH);
    }
  } catch (error) {
    console.warn("Could not reset database:", error);
  }
}

/**
 * Extended test fixture with database reset
 */
export const test = base.extend({
  page: async ({ page }, use) => {
    // Reset database before each test
    resetDatabase();

    // Wait a moment for the server to recreate the database
    await new Promise((resolve) => setTimeout(resolve, 500));

    await use(page);
  },
});

export { expect } from "@playwright/test";
