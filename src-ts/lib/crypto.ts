import * as path from "path";
import { promises as fs } from "fs";

const ALGORITHM_NAME = "AES-GCM";
const KEY_SIZE = 256;
const IV_LENGTH = 12; // AES-GCM recommends 12-byte IV
const KEY_FILE = ".janitarr.key"; // Using dotfile to hide it
const KEY_DIR = path.join(process.cwd(), "data");
export let KEY_PATH = path.join(KEY_DIR, KEY_FILE);

export function setKeyPathForTesting(newPath: string) {
  KEY_PATH = newPath;
  cachedKey = null; // Clear cached key when path changes
}

let cachedKey: CryptoKey | null = null;

async function ensureKeyDirectoryExists(): Promise<void> {
  await fs.mkdir(KEY_DIR, { recursive: true });
}

export async function getEncryptionKey(): Promise<CryptoKey> {
  if (cachedKey) {
    return cachedKey;
  }

  await ensureKeyDirectoryExists();

  try {
    const keyData = await fs.readFile(KEY_PATH, "utf8");
    const jwk = JSON.parse(keyData);
    cachedKey = await crypto.subtle.importKey(
      "jwk",
      jwk,
      { name: ALGORITHM_NAME, length: KEY_SIZE },
      true, // extractable
      ["encrypt", "decrypt"],
    );
    return cachedKey;
  } catch (error: unknown) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") {
      // Key file not found, generate a new one
      const newKey = await crypto.subtle.generateKey(
        {
          name: ALGORITHM_NAME,
          length: KEY_SIZE,
        },
        true, // extractable
        ["encrypt", "decrypt"],
      );

      const jwk = await crypto.subtle.exportKey("jwk", newKey);
      await fs.writeFile(KEY_PATH, JSON.stringify(jwk), "utf8");
      cachedKey = newKey;
      return newKey;
    }
    throw new Error(
      `Failed to get encryption key: ${(error as Error).message}`,
    );
  }
}

export async function encryptApiKey(
  plaintext: string,
  key: CryptoKey,
): Promise<string> {
  const iv = crypto.getRandomValues(new Uint8Array(IV_LENGTH));
  const encoded = new TextEncoder().encode(plaintext);

  const ciphertext = await crypto.subtle.encrypt(
    {
      name: ALGORITHM_NAME,
      iv: iv,
    },
    key,
    encoded,
  );

  // Combine IV and ciphertext for storage, separated by a delimiter
  const ivBase64 = Buffer.from(iv).toString("base64");
  const ciphertextBase64 = Buffer.from(ciphertext).toString("base64");

  return `${ivBase64}:${ciphertextBase64}`;
}

export async function decryptApiKey(
  ciphertextWithIv: string,
  key: CryptoKey,
): Promise<string> {
  const [ivBase64, ciphertextBase64] = ciphertextWithIv.split(":");

  if (!ivBase64 || !ciphertextBase64) {
    throw new Error("Invalid ciphertext format");
  }

  const iv = Buffer.from(ivBase64, "base64");
  const ciphertext = Buffer.from(ciphertextBase64, "base64");

  const decrypted = await crypto.subtle.decrypt(
    {
      name: ALGORITHM_NAME,
      iv: iv,
    },
    key,
    ciphertext,
  );

  return new TextDecoder().decode(decrypted);
}
