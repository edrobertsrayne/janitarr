import { test, expect, describe, beforeAll, afterAll, beforeEach } from "bun:test";
import * as path from 'path';
import { promises as fs } from 'fs';
import { getEncryptionKey, encryptApiKey, decryptApiKey, setKeyPathForTesting } from "../../src/lib/crypto";

describe("crypto functions", () => {
  const TEST_KEY_FILE_NAME = '.janitarr.test.key';
  const TEST_KEY_DIR = path.join(process.cwd(), 'data');
  const TEST_KEY_PATH = path.join(TEST_KEY_DIR, TEST_KEY_FILE_NAME);

  // Store the original KEY_PATH to restore it after all tests
  let originalKeyPath: string;

  beforeAll(() => {
    // Save original KEY_PATH and set test specific one
    originalKeyPath = path.join(process.cwd(), 'data', '.janitarr.key'); // Assuming default is .janitarr.key
    setKeyPathForTesting(TEST_KEY_PATH);
  });

  beforeEach(async () => {
    // Ensure test directory exists
    await fs.mkdir(TEST_KEY_DIR, { recursive: true });
    // Clean up any existing test key file before each test
    try {
      await fs.unlink(TEST_KEY_PATH);
    } catch (error: unknown) {
      if ((error as NodeJS.ErrnoException).code !== 'ENOENT') {
        console.error('Failed to clean up test key file before test:', error);
      }
    }
  });

  afterAll(async () => {
    // Clean up the test key file after all tests
    try {
      await fs.unlink(TEST_KEY_PATH);
    } catch (error: unknown) {
      if ((error as NodeJS.ErrnoException).code !== 'ENOENT') {
        console.error('Failed to clean up test key file after all tests:', error);
      }
    }
    // Restore original KEY_PATH
    setKeyPathForTesting(originalKeyPath);
  });

  test("should generate and retrieve a new encryption key if none exists", async () => {
    const key1 = await getEncryptionKey();
    expect(key1).toBeInstanceOf(CryptoKey);

    const keyData = await fs.readFile(TEST_KEY_PATH, 'utf8');
    expect(keyData).toBeString();
    expect(JSON.parse(keyData)).toHaveProperty('kty', 'oct'); // Octet string key type

    const key2 = await getEncryptionKey(); // Should retrieve the same cached key
    expect(key2).toBe(key1);
  });

  test("should encrypt and decrypt a plaintext string", async () => {
    const key = await getEncryptionKey();
    const plaintext = "mysecretapikey123";

    const encrypted = await encryptApiKey(plaintext, key);
    expect(encrypted).toBeString();
    expect(encrypted).not.toBe(plaintext); // Should not be plaintext

    const decrypted = await decryptApiKey(encrypted, key);
    expect(decrypted).toBe(plaintext);
  });

  test("should throw an error for invalid ciphertext format", async () => {
    const key = await getEncryptionKey();
    const invalidCiphertext = "invalid-format"; // Missing colon for IV:ciphertext

    await expect(decryptApiKey(invalidCiphertext, key)).rejects.toThrow(
      "Invalid ciphertext format"
    );
  });

  test("should correctly handle empty strings", async () => {
    const key = await getEncryptionKey();
    const plaintext = "";

    const encrypted = await encryptApiKey(plaintext, key);
    expect(encrypted).toBeString();

    const decrypted = await decryptApiKey(encrypted, key);
    expect(decrypted).toBe(plaintext);
  });

  test("should use different IVs for each encryption, resulting in different ciphertexts", async () => {
    const key = await getEncryptionKey();
    const plaintext = "same secret";

    const encrypted1 = await encryptApiKey(plaintext, key);
    const encrypted2 = await encryptApiKey(plaintext, key);

    expect(encrypted1).not.toBe(encrypted2); // Different IVs should produce different ciphertexts
  });

  test("should fail decryption with a different key", async () => {
    // Generate two different keys
    const key1 = await getEncryptionKey(); // This will use TEST_KEY_PATH

    // Temporarily change KEY_PATH to generate a second distinct key
    const TEMP_KEY_PATH_2 = path.join(TEST_KEY_DIR, '.janitarr.test.key.temp');
    setKeyPathForTesting(TEMP_KEY_PATH_2);
    const key2 = await getEncryptionKey();
    setKeyPathForTesting(TEST_KEY_PATH); // Restore for subsequent tests

    const plaintext = "another secret";
    const encrypted = await encryptApiKey(plaintext, key1);

    // Attempt to decrypt with key2
    await expect(decryptApiKey(encrypted, key2)).rejects.toThrow();

    // Clean up temp key file
    try {
      await fs.unlink(TEMP_KEY_PATH_2);
    } catch (error: unknown) {
      if ((error as NodeJS.ErrnoException).code !== 'ENOENT') {
        console.error('Failed to clean up temp test key file:', error);
      }
    }
  });

});
