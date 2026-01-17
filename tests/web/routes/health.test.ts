/**
 * Unit tests for health check endpoint
 */

import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import { handleHealthCheck } from "../../../src/web/routes/health";
import { DatabaseManager, closeDatabase } from "../../../src/storage/database";
import { HttpStatus } from "../../../src/web/types";
import { unlinkSync } from "fs";

const TEST_DB_PATH = ":memory:";

describe("Health Check Endpoint", () => {
  let db: DatabaseManager;

  beforeEach(() => {
    db = new DatabaseManager(TEST_DB_PATH);
  });

  afterEach(() => {
    closeDatabase();
  });

  test("returns degraded status with scheduler disabled", async () => {
    // Explicitly disable scheduler
    db.setAppConfig({ schedule: { enabled: false } });

    const response = await handleHealthCheck(db);

    expect(response.status).toBe(HttpStatus.OK);

    const data = await response.json();
    expect(data.status).toBe("degraded"); // Scheduler disabled
    expect(data.timestamp).toBeDefined();
    expect(data.services.webServer.status).toBe("ok");
    expect(data.services.scheduler.status).toBe("disabled");
    expect(data.services.scheduler.isRunning).toBe(false);
    expect(data.services.scheduler.isCycleActive).toBe(false);
    expect(data.services.scheduler.nextRun).toBeNull();
    expect(data.database.status).toBe("ok");
  });

  test("returns error status when scheduler enabled but not running", async () => {
    // Enable scheduler but don't start it
    db.setAppConfig({ schedule: { enabled: true } });

    const response = await handleHealthCheck(db);

    expect(response.status).toBe(HttpStatus.SERVICE_UNAVAILABLE);

    const data = await response.json();
    expect(data.status).toBe("error"); // Scheduler enabled but not running
    expect(data.services.scheduler.status).toBe("error");
    expect(data.services.scheduler.isRunning).toBe(false);
  });

  test("response includes all required fields", async () => {
    const response = await handleHealthCheck(db);
    const data = await response.json();

    // Check top-level fields
    expect(data).toHaveProperty("status");
    expect(data).toHaveProperty("timestamp");
    expect(data).toHaveProperty("services");
    expect(data).toHaveProperty("database");

    // Check services structure
    expect(data.services).toHaveProperty("webServer");
    expect(data.services).toHaveProperty("scheduler");

    // Check scheduler fields
    expect(data.services.scheduler).toHaveProperty("status");
    expect(data.services.scheduler).toHaveProperty("isRunning");
    expect(data.services.scheduler).toHaveProperty("isCycleActive");
    expect(data.services.scheduler).toHaveProperty("nextRun");
  });

  test("timestamp is valid ISO 8601 format", async () => {
    const response = await handleHealthCheck(db);
    const data = await response.json();

    const timestamp = new Date(data.timestamp);
    expect(timestamp.toISOString()).toBe(data.timestamp);
  });

  test("returns JSON content type", async () => {
    const response = await handleHealthCheck(db);

    const contentType = response.headers.get("Content-Type");
    expect(contentType).toBe("application/json");
  });

  test("database status is ok when accessible", async () => {
    const response = await handleHealthCheck(db);
    const data = await response.json();

    expect(data.database.status).toBe("ok");
  });
});
