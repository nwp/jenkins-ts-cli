import type { JenkinsClient } from "./client.ts";

export async function followConsole(
  client: JenkinsClient,
  jobUrl: string,
  buildNumber: number | string,
): Promise<void> {
  let offset = 0;
  while (true) {
    const res = await client.get(
      `${jobUrl}/${buildNumber}/logText/progressiveText?start=${offset}`,
      { raw: true },
    );
    const text = await res.text();
    if (text) process.stdout.write(text);
    const newOffset = res.headers.get("X-Text-Size");
    if (newOffset) offset = parseInt(newOffset, 10);
    if (res.headers.get("X-More-Data") !== "true") break;
    await Bun.sleep(1000);
  }
}
