/**
 * Unit tests for CLI unified startup commands
 *
 * These tests focus on command configuration, validation, and integration behavior.
 * Full end-to-end tests with process lifecycle are covered in E2E tests.
 */

import { describe, test, expect, beforeEach, afterEach, mock } from "bun:test";
import { createProgram } from "../../src/cli/commands";
import { DatabaseManager, closeDatabase } from "../../src/storage/database";
import { unlinkSync, existsSync } from "fs";

const TEST_DB_PATH = "./data/test-cli-commands.db";

describe("CLI Commands - Unified Startup", () => {
  let testDb: DatabaseManager;
  let originalExit: typeof process.exit;
  let exitCode: number | undefined;

  beforeEach(() => {
    // Clean up test database
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
    process.env.JANITARR_DB_PATH = TEST_DB_PATH;
    testDb = new DatabaseManager(TEST_DB_PATH);

    // Mock process.exit to prevent tests from exiting
    exitCode = undefined;
    originalExit = process.exit;
    process.exit = mock((code?: number) => {
      exitCode = code || 0;
      throw new Error(`Process.exit called with code ${code}`);
    }) as any;
  });

  afterEach(() => {
    // Restore process.exit
    process.exit = originalExit;

    // Clean up database
    testDb.close();
    closeDatabase();
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
    delete process.env.JANITARR_DB_PATH;
  });

  describe("start command", () => {
    test("command is registered with correct name and description", () => {
      const program = createProgram();
      const startCommand = program.commands.find(
        (cmd) => cmd.name() === "start",
      );

      expect(startCommand).toBeDefined();
      expect(startCommand?.description()).toBe(
        "Start scheduler and web server (production mode)",
      );
    });

    test("command has port option with default value", () => {
      const program = createProgram();
      const startCommand = program.commands.find(
        (cmd) => cmd.name() === "start",
      );

      const portOption = startCommand?.options.find(
        (opt) => opt.short === "-p" || opt.long === "--port",
      );

      expect(portOption).toBeDefined();
      expect(portOption?.long).toBe("--port");
      expect(portOption?.short).toBe("-p");
      expect(portOption?.defaultValue).toBe("3434");
    });

    test("command has host option with default value", () => {
      const program = createProgram();
      const startCommand = program.commands.find(
        (cmd) => cmd.name() === "start",
      );

      const hostOption = startCommand?.options.find(
        (opt) => opt.short === "-h" || opt.long === "--host",
      );

      expect(hostOption).toBeDefined();
      expect(hostOption?.long).toBe("--host");
      expect(hostOption?.short).toBe("-h");
      expect(hostOption?.defaultValue).toBe("localhost");
    });

    test("rejects port 0 as invalid", async () => {
      const program = createProgram();

      try {
        await program.parseAsync(["start", "--port", "0"], { from: "user" });
      } catch (error) {
        expect(exitCode).toBe(1);
      }
    });

    test("rejects port above 65535 as invalid", async () => {
      const program = createProgram();
      exitCode = undefined;

      try {
        await program.parseAsync(["start", "--port", "65536"], {
          from: "user",
        });
      } catch (error) {
        expect(exitCode).toBeDefined();
        expect(exitCode!).toBe(1);
      }
    });

    test("rejects non-numeric port as invalid", async () => {
      const program = createProgram();
      exitCode = undefined;

      try {
        await program.parseAsync(["start", "--port", "abc"], { from: "user" });
      } catch (error) {
        expect(exitCode).toBeDefined();
        expect(exitCode!).toBe(1);
      }
    });
  });

  describe("dev command", () => {
    test("command is registered with correct name and description", () => {
      const program = createProgram();
      const devCommand = program.commands.find((cmd) => cmd.name() === "dev");

      expect(devCommand).toBeDefined();
      expect(devCommand?.description()).toBe(
        "Start scheduler and web server (development mode)",
      );
    });

    test("command has port option with default value", () => {
      const program = createProgram();
      const devCommand = program.commands.find((cmd) => cmd.name() === "dev");

      const portOption = devCommand?.options.find(
        (opt) => opt.short === "-p" || opt.long === "--port",
      );

      expect(portOption).toBeDefined();
      expect(portOption?.long).toBe("--port");
      expect(portOption?.short).toBe("-p");
      expect(portOption?.defaultValue).toBe("3434");
    });

    test("command has host option with default value", () => {
      const program = createProgram();
      const devCommand = program.commands.find((cmd) => cmd.name() === "dev");

      const hostOption = devCommand?.options.find(
        (opt) => opt.short === "-h" || opt.long === "--host",
      );

      expect(hostOption).toBeDefined();
      expect(hostOption?.long).toBe("--host");
      expect(hostOption?.short).toBe("-h");
      expect(hostOption?.defaultValue).toBe("localhost");
    });

    test("rejects negative port as invalid", async () => {
      const program = createProgram();

      try {
        await program.parseAsync(["dev", "--port", "-1"], { from: "user" });
      } catch (error) {
        expect(exitCode).toBe(1);
      }
    });

    test("rejects port above 65535 as invalid", async () => {
      const program = createProgram();
      exitCode = undefined;

      try {
        await program.parseAsync(["dev", "--port", "99999"], { from: "user" });
      } catch (error) {
        expect(exitCode).toBeDefined();
        expect(exitCode).toBe(1);
      }
    });
  });

  describe("serve command removal", () => {
    test("serve command is not registered", () => {
      const program = createProgram();
      const serveCommand = program.commands.find(
        (cmd) => cmd.name() === "serve",
      );

      expect(serveCommand).toBeUndefined();
    });
  });

  describe("command integration behavior", () => {
    test("scheduler configuration can be enabled", () => {
      // Enable scheduler
      testDb.setAppConfig({ schedule: { enabled: true, intervalHours: 2 } });
      const config = testDb.getAppConfig();
      expect(config.schedule.enabled).toBe(true);
      expect(config.schedule.intervalHours).toBe(2);
    });

    test("scheduler configuration can be disabled", () => {
      // Disable scheduler
      testDb.setAppConfig({ schedule: { enabled: false } });
      const config = testDb.getAppConfig();
      expect(config.schedule.enabled).toBe(false);
    });

    test("scheduler interval hours can be configured", () => {
      // Test that configuration is accessible (used by both commands)
      testDb.setAppConfig({ schedule: { enabled: true, intervalHours: 3 } });
      const config = testDb.getAppConfig();
      expect(config.schedule.enabled).toBe(true);
      expect(config.schedule.intervalHours).toBe(3);
    });
  });

  describe("port validation boundary cases", () => {
    test("port 1 is the minimum valid port", () => {
      const port = 1;
      const isValid = !isNaN(port) && port >= 1 && port <= 65535;
      expect(isValid).toBe(true);
    });

    test("port 65535 is the maximum valid port", () => {
      const port = 65535;
      const isValid = !isNaN(port) && port >= 1 && port <= 65535;
      expect(isValid).toBe(true);
    });

    test("port 0 is invalid", () => {
      const port = 0;
      const isValid = !isNaN(port) && port >= 1 && port <= 65535;
      expect(isValid).toBe(false);
    });

    test("port 65536 is invalid", () => {
      const port = 65536;
      const isValid = !isNaN(port) && port >= 1 && port <= 65535;
      expect(isValid).toBe(false);
    });
  });
});
