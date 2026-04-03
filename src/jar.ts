import { homedir } from "os";
import { join } from "path";
import { mkdir } from "fs/promises";
import { existsSync } from "fs";
import type { JenkinsClient } from "./client.ts";

const CACHE_DIR = join(homedir(), ".jenkins-cli");
const JAR_PATH = join(CACHE_DIR, "jenkins-cli.jar");
const META_PATH = join(CACHE_DIR, "jar-meta.json");

interface JarMeta {
  etag?: string;
  lastModified?: string;
  serverUrl: string;
}

async function readMeta(): Promise<JarMeta | null> {
  try {
    const data = await Bun.file(META_PATH).text();
    return JSON.parse(data) as JarMeta;
  } catch {
    return null;
  }
}

async function writeMeta(meta: JarMeta): Promise<void> {
  await Bun.write(META_PATH, JSON.stringify(meta, null, 2));
}

async function findJava(): Promise<string | null> {
  const javaHome = process.env.JAVA_HOME;
  if (javaHome) {
    const path = join(javaHome, "bin", "java");
    if (existsSync(path)) return path;
  }

  const proc = Bun.spawn(["which", "java"], {
    stdout: "pipe",
    stderr: "ignore",
  });
  const exitCode = await proc.exited;
  if (exitCode === 0) {
    const path = (await new Response(proc.stdout).text()).trim();
    if (path) return path;
  }
  return null;
}

export async function ensureJar(client: JenkinsClient): Promise<string> {
  await mkdir(CACHE_DIR, { recursive: true });

  const meta = await readMeta();
  const jarExists = existsSync(JAR_PATH);

  if (jarExists && meta && meta.serverUrl === client.baseUrl) {
    const headers: Record<string, string> = {};
    if (meta.etag) headers["If-None-Match"] = meta.etag;
    if (meta.lastModified) headers["If-Modified-Since"] = meta.lastModified;

    const res = await fetch(`${client.baseUrl}/jnlpJars/jenkins-cli.jar`, {
      headers,
    });
    if (res.status === 304) return JAR_PATH;
    if (res.ok) {
      await Bun.write(JAR_PATH, res);
      await writeMeta({
        etag: res.headers.get("ETag") ?? undefined,
        lastModified: res.headers.get("Last-Modified") ?? undefined,
        serverUrl: client.baseUrl,
      });
      return JAR_PATH;
    }
  }

  const res = await fetch(`${client.baseUrl}/jnlpJars/jenkins-cli.jar`);
  if (!res.ok) {
    throw new Error(`Failed to download jenkins-cli.jar: ${res.status}`);
  }
  await Bun.write(JAR_PATH, res);
  await writeMeta({
    etag: res.headers.get("ETag") ?? undefined,
    lastModified: res.headers.get("Last-Modified") ?? undefined,
    serverUrl: client.baseUrl,
  });
  return JAR_PATH;
}

export async function runJar(
  client: JenkinsClient,
  credentials: { username: string; token: string },
  args: string[],
): Promise<number> {
  const java = await findJava();
  if (!java) {
    console.error(
      "Java is required for this command but was not found on PATH.\n" +
      "Install Java or set JAVA_HOME to use this command.",
    );
    return 1;
  }

  const jarPath = await ensureJar(client);
  const proc = Bun.spawn(
    [
      java,
      "-jar",
      jarPath,
      "-s",
      client.baseUrl,
      "-auth",
      `${credentials.username}:${credentials.token}`,
      ...args,
    ],
    {
      stdin: "inherit",
      stdout: "inherit",
      stderr: "inherit",
    },
  );
  return proc.exited;
}
