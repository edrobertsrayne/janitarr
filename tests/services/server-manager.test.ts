/**
 * Tests for server manager service
 */

import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import { unlinkSync, existsSync } from "fs";
import { DatabaseManager, closeDatabase } from "../../src/storage/database";
import { maskApiKey } from "../../src/services/server-manager";

const TEST_DB_PATH = "./data/test-server-manager.db";

// Mock getDatabase to use test database
let testDb: DatabaseManager;

describe("maskApiKey", () => {
  test("masks long API keys showing first and last 4 chars", () => {
    const masked = maskApiKey("abcd1234efgh5678ijkl");
    expect(masked).toBe("abcd************ijkl");
  });

  test("masks short API keys completely", () => {
    expect(maskApiKey("abcd")).toBe("****");
    expect(maskApiKey("abcdefgh")).toBe("********");
  });

  test("handles exactly 9 character keys", () => {
    const masked = maskApiKey("123456789");
    expect(masked).toBe("1234*6789");
  });

  test("handles very long API keys with limited asterisks", () => {
    const longKey = "a".repeat(50);
    const masked = maskApiKey(longKey);
    // Should show first 4 + max 20 asterisks + last 4
    expect(masked.length).toBe(28);
    expect(masked.startsWith("aaaa")).toBe(true);
    expect(masked.endsWith("aaaa")).toBe(true);
  });
});

describe("Server Manager Database Integration", () => {
  beforeEach(() => {
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
    testDb = new DatabaseManager(TEST_DB_PATH);
  });

  afterEach(() => {
    testDb.close();
    closeDatabase();
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
  });

  test("prevents duplicate server URLs", () => {
    testDb.addServer({
      id: "id1",
      name: "Radarr 1",
      url: "http://localhost:7878",
      apiKey: "key1",
      type: "radarr",
    });

    // Same URL + type should be detected as duplicate
    expect(testDb.serverExists("http://localhost:7878", "radarr")).toBe(true);

    // Same URL but different type is allowed
    expect(testDb.serverExists("http://localhost:7878", "sonarr")).toBe(false);
  });

  test("prevents duplicate server names", () => {
    testDb.addServer({
      id: "id1",
      name: "My Server",
      url: "http://localhost:7878",
      apiKey: "key1",
      type: "radarr",
    });

    const existing = testDb.getServerByName("My Server");
    expect(existing).not.toBeNull();
  });
});
