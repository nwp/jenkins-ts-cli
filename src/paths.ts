import { homedir } from "os";
import { join } from "path";

export const CONFIG_DIR = join(homedir(), ".jenkins-cli");

export function jobPath(name: string): string {
  return name
    .split("/")
    .map((segment) => `job/${encodeURIComponent(segment)}`)
    .join("/");
}

export function nodeUrl(name: string): string {
  return `/computer/${encodeURIComponent(name)}`;
}

export function viewUrl(name: string): string {
  return `/view/${encodeURIComponent(name)}`;
}

export function normalizeUrl(url: string): string {
  const parsed = new URL(url);
  parsed.hostname = parsed.hostname.toLowerCase();
  parsed.protocol = parsed.protocol.toLowerCase();
  let normalized = parsed.toString();
  if (normalized.endsWith("/")) {
    normalized = normalized.slice(0, -1);
  }
  return normalized;
}
