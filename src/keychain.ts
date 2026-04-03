import { join } from "path";
import { mkdir, readFile, writeFile } from "fs/promises";
import { CONFIG_DIR, normalizeUrl } from "./paths.ts";

const CREDS_FILE = join(CONFIG_DIR, "credentials.json");
const SERVICE_LABEL = "jenkins-cli";

type CredentialsMap = Record<string, string>;

async function ensureConfigDir(): Promise<void> {
  await mkdir(CONFIG_DIR, { recursive: true });
}

async function readCredsMap(): Promise<CredentialsMap> {
  try {
    const data = await readFile(CREDS_FILE, "utf-8");
    return JSON.parse(data) as CredentialsMap;
  } catch {
    return {};
  }
}

async function writeCredsMap(map: CredentialsMap): Promise<void> {
  await ensureConfigDir();
  await writeFile(CREDS_FILE, JSON.stringify(map, null, 2), { mode: 0o600 });
}

export async function getStoredUsername(serverUrl: string): Promise<string | null> {
  const map = await readCredsMap();
  return map[normalizeUrl(serverUrl)] ?? null;
}

export async function storeCredentials(
  serverUrl: string,
  username: string,
  token: string,
): Promise<void> {
  const url = normalizeUrl(serverUrl);

  const map = await readCredsMap();
  map[url] = username;
  await writeCredsMap(map);

  const proc = Bun.spawn([
    "security",
    "add-internet-password",
    "-a", username,
    "-s", url,
    "-w", token,
    "-l", SERVICE_LABEL,
    "-U",
  ], { stdout: "ignore", stderr: "pipe" });

  const exitCode = await proc.exited;
  if (exitCode !== 0) {
    const stderr = await new Response(proc.stderr).text();
    throw new Error(`Failed to store credentials in Keychain: ${stderr.trim()}`);
  }
}

export async function getStoredToken(
  serverUrl: string,
  username: string,
): Promise<string | null> {
  const url = normalizeUrl(serverUrl);

  const proc = Bun.spawn([
    "security",
    "find-internet-password",
    "-a", username,
    "-s", url,
    "-l", SERVICE_LABEL,
    "-w",
  ], { stdout: "pipe", stderr: "pipe" });

  const exitCode = await proc.exited;
  if (exitCode !== 0) return null;

  const token = await new Response(proc.stdout).text();
  return token.trim() || null;
}

export async function deleteCredentials(serverUrl: string): Promise<boolean> {
  const url = normalizeUrl(serverUrl);
  const username = await getStoredUsername(url);
  if (!username) return false;

  const proc = Bun.spawn([
    "security",
    "delete-internet-password",
    "-a", username,
    "-s", url,
    "-l", SERVICE_LABEL,
  ], { stdout: "ignore", stderr: "ignore" });

  await proc.exited;

  const map = await readCredsMap();
  delete map[url];
  await writeCredsMap(map);

  return true;
}
