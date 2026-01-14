/**
 * Tests for API client library
 */

import { describe, test, expect } from "bun:test";
import { normalizeUrl, validateUrl } from "../../src/lib/api-client";

describe("normalizeUrl", () => {
  test("adds http:// protocol if missing", () => {
    expect(normalizeUrl("localhost:7878")).toBe("http://localhost:7878");
    expect(normalizeUrl("192.168.1.1:8989")).toBe("http://192.168.1.1:8989");
  });

  test("preserves existing http:// protocol", () => {
    expect(normalizeUrl("http://localhost:7878")).toBe("http://localhost:7878");
  });

  test("preserves existing https:// protocol", () => {
    expect(normalizeUrl("https://example.com")).toBe("https://example.com");
  });

  test("removes trailing slashes", () => {
    expect(normalizeUrl("http://localhost:7878/")).toBe("http://localhost:7878");
    expect(normalizeUrl("http://localhost:7878///")).toBe("http://localhost:7878");
  });

  test("trims whitespace", () => {
    expect(normalizeUrl("  http://localhost:7878  ")).toBe("http://localhost:7878");
  });

  test("handles case-insensitive protocols", () => {
    expect(normalizeUrl("HTTP://localhost")).toBe("HTTP://localhost");
    expect(normalizeUrl("HTTPS://localhost")).toBe("HTTPS://localhost");
  });
});

describe("validateUrl", () => {
  test("accepts valid http URL", () => {
    const result = validateUrl("http://localhost:7878");
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe("http://localhost:7878");
    }
  });

  test("accepts valid https URL", () => {
    const result = validateUrl("https://example.com/radarr");
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe("https://example.com/radarr");
    }
  });

  test("normalizes URL without protocol", () => {
    const result = validateUrl("localhost:7878");
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe("http://localhost:7878");
    }
  });

  test("rejects completely invalid URLs", () => {
    const result = validateUrl("not a url at all!!!");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error).toContain("Invalid URL");
    }
  });

  test("accepts IP addresses", () => {
    const result = validateUrl("192.168.1.100:7878");
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe("http://192.168.1.100:7878");
    }
  });

  test("accepts localhost", () => {
    const result = validateUrl("localhost:8989");
    expect(result.success).toBe(true);
  });

  test("accepts domain names", () => {
    const result = validateUrl("radarr.example.com");
    expect(result.success).toBe(true);
  });
});
