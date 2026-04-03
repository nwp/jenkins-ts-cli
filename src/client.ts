import type { Credentials } from "./auth.ts";
import { normalizeUrl, jobPath } from "./paths.ts";

export class JenkinsClient {
  readonly baseUrl: string;
  private readonly authHeader: string;
  private crumb: { field: string; value: string } | null = null;
  private crumbFetched = false;

  constructor(serverUrl: string, credentials: Credentials) {
    this.baseUrl = normalizeUrl(serverUrl);
    this.authHeader =
      "Basic " + btoa(`${credentials.username}:${credentials.token}`);
  }

  private async fetchCrumb(): Promise<void> {
    if (this.crumbFetched) return;
    this.crumbFetched = true;

    try {
      const res = await fetch(`${this.baseUrl}/crumbIssuer/api/json`, {
        headers: { Authorization: this.authHeader },
      });
      if (res.ok) {
        const data = (await res.json()) as {
          crumbRequestField: string;
          crumb: string;
        };
        this.crumb = { field: data.crumbRequestField, value: data.crumb };
      }
    } catch {
      // CRUMB not available — proceed without it
    }
  }

  private crumbHeaders(): Record<string, string> {
    if (!this.crumb) return {};
    return { [this.crumb.field]: this.crumb.value };
  }

  async get(path: string, options?: { raw?: boolean }): Promise<Response> {
    const url = `${this.baseUrl}${path}`;
    const res = await fetch(url, {
      headers: { Authorization: this.authHeader },
    });
    if (!res.ok && !options?.raw) {
      throw await apiError(res);
    }
    return res;
  }

  async getJson<T = unknown>(path: string): Promise<T> {
    const res = await this.get(path);
    return (await res.json()) as T;
  }

  async getText(path: string): Promise<string> {
    const res = await this.get(path);
    return res.text();
  }

  async post(
    path: string,
    options?: {
      body?: string | FormData | Blob | ArrayBuffer | ReadableStream;
      contentType?: string;
      raw?: boolean;
    },
  ): Promise<Response> {
    await this.fetchCrumb();
    const url = `${this.baseUrl}${path}`;
    const headers: Record<string, string> = {
      Authorization: this.authHeader,
      ...this.crumbHeaders(),
    };
    if (options?.contentType) {
      headers["Content-Type"] = options.contentType;
    }
    const res = await fetch(url, {
      method: "POST",
      headers,
      body: options?.body,
    });
    if (!res.ok && !options?.raw) {
      throw await apiError(res);
    }
    return res;
  }

  async postForm(
    path: string,
    params: Record<string, string>,
  ): Promise<Response> {
    const body = new URLSearchParams(params).toString();
    return this.post(path, {
      body,
      contentType: "application/x-www-form-urlencoded",
    });
  }

  jobUrl(name: string): string {
    return `/${jobPath(name)}`;
  }

  async readStdin(): Promise<string> {
    const chunks: Uint8Array[] = [];
    for await (const chunk of Bun.stdin.stream()) {
      chunks.push(chunk);
    }
    return Buffer.concat(chunks).toString("utf-8");
  }
}

async function apiError(res: Response): Promise<Error> {
  let body = "";
  try {
    body = await res.text();
  } catch {}
  const status = res.status;
  const statusText = res.statusText;
  const msg = body
    ? `Jenkins API error ${status} (${statusText}): ${body.slice(0, 500)}`
    : `Jenkins API error ${status} (${statusText})`;
  return new Error(msg);
}
