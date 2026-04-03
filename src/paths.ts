export function jobPath(name: string): string {
  return name
    .split("/")
    .map((segment) => `job/${encodeURIComponent(segment)}`)
    .join("/");
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
