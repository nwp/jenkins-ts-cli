import type { Command } from "commander";
import { resolveContext } from "../cli.ts";
import { printOutput } from "../output.ts";
import { followConsole } from "../console.ts";

export function registerBuildCommands(program: Command): void {
  program
    .command("console")
    .description("Get console output for a build")
    .argument("<name>", "job name")
    .argument("[build]", "build number (default: lastBuild)", "lastBuild")
    .option("--follow", "follow output in real-time", false)
    .action(async (name: string, build: string, opts: { follow: boolean }) => {
      const ctx = await resolveContext(program);
      const jobUrl = ctx.client.jobUrl(name);

      if (opts.follow) {
        await followConsole(ctx.client, jobUrl, build);
      } else {
        const text = await ctx.client.getText(`${jobUrl}/${build}/consoleText`);
        process.stdout.write(text);
      }
    });

  program
    .command("get-build")
    .description("Get build details (status, result, duration, etc.)")
    .argument("<name>", "job name")
    .argument("[build]", "build number (default: lastBuild)", "lastBuild")
    .action(async (name: string, build: string) => {
      const ctx = await resolveContext(program);
      const data = await ctx.client.getJson<BuildInfo>(
        `${ctx.client.jobUrl(name)}/${build}/api/json?tree=number,result,building,duration,estimatedDuration,timestamp,displayName,description,url`,
      );
      if (ctx.json) {
        printOutput(data, true);
      } else {
        const status = data.building ? "BUILDING" : (data.result ?? "UNKNOWN");
        const duration = data.building
          ? formatDuration(Date.now() - data.timestamp) + " (so far)"
          : formatDuration(data.duration);
        console.log(`Build:    ${data.displayName ?? `#${data.number}`}`);
        console.log(`Status:   ${status}`);
        console.log(`Duration: ${duration}`);
        if (data.description) console.log(`Desc:     ${data.description}`);
        console.log(`URL:      ${data.url}`);
      }
      if (!data.building && data.result !== "SUCCESS") {
        process.exitCode = 1;
      }
    });

  program
    .command("wait-build")
    .description("Wait for a build to finish and report its result")
    .argument("<name>", "job name")
    .argument("[build]", "build number (default: lastBuild)", "lastBuild")
    .option("--timeout <seconds>", "maximum time to wait in seconds", "0")
    .action(async (name: string, build: string, opts: { timeout: string }) => {
      const ctx = await resolveContext(program);
      const jobUrl = ctx.client.jobUrl(name);
      const timeout = parseInt(opts.timeout, 10) * 1000;
      const start = Date.now();

      while (true) {
        const data = await ctx.client.getJson<BuildInfo>(
          `${jobUrl}/${build}/api/json?tree=number,result,building,duration,timestamp,displayName,url`,
        );
        if (!data.building) {
          const status = data.result ?? "UNKNOWN";
          if (ctx.json) {
            printOutput(data, true);
          } else {
            console.log(`Build ${data.displayName ?? `#${data.number}`} finished: ${status} (${formatDuration(data.duration)})`);
          }
          if (data.result !== "SUCCESS") process.exitCode = 1;
          return;
        }
        if (timeout > 0 && Date.now() - start >= timeout) {
          throw new Error(`Timed out after ${opts.timeout}s waiting for build to finish.`);
        }
        await Bun.sleep(3000);
      }
    });

  program
    .command("list-builds")
    .description("List recent builds for a job")
    .argument("<name>", "job name")
    .option("--limit <n>", "number of builds to show", "10")
    .action(async (name: string, opts: { limit: string }) => {
      const ctx = await resolveContext(program);
      const limit = parseInt(opts.limit, 10);
      const data = await ctx.client.getJson<{
        builds: BuildInfo[];
      }>(
        `${ctx.client.jobUrl(name)}/api/json?tree=builds[number,result,building,duration,timestamp,displayName,url]{0,${limit}}`,
      );
      const builds = data.builds ?? [];
      if (ctx.json) {
        printOutput(builds, true);
      } else if (builds.length === 0) {
        console.log("No builds.");
      } else {
        for (const b of builds) {
          const status = b.building ? "BUILDING" : (b.result ?? "UNKNOWN");
          const name = b.displayName ?? `#${b.number}`;
          const dur = b.building ? formatDuration(Date.now() - b.timestamp) : formatDuration(b.duration);
          console.log(`${name.padEnd(12)} ${status.padEnd(10)} ${dur}`);
        }
      }
    });

  program
    .command("tail")
    .description("Stream console output starting from the current position (new lines only)")
    .argument("<name>", "job name")
    .argument("[build]", "build number (default: lastBuild)", "lastBuild")
    .action(async (name: string, build: string) => {
      const ctx = await resolveContext(program);
      const jobUrl = ctx.client.jobUrl(name);

      const initialRes = await ctx.client.get(
        `${jobUrl}/${build}/logText/progressiveText?start=0`,
        { raw: true },
      );
      let offset = parseInt(initialRes.headers.get("X-Text-Size") ?? "0", 10);
      const moreData = initialRes.headers.get("X-More-Data") === "true";

      if (!moreData) {
        console.error("Build already finished. Use 'console' to view the full log.");
        return;
      }

      while (true) {
        const res = await ctx.client.get(
          `${jobUrl}/${build}/logText/progressiveText?start=${offset}`,
          { raw: true },
        );
        const text = await res.text();
        if (text) process.stdout.write(text);
        const newOffset = res.headers.get("X-Text-Size");
        if (newOffset) offset = parseInt(newOffset, 10);
        if (res.headers.get("X-More-Data") !== "true") break;
        await Bun.sleep(1000);
      }
    });

  program
    .command("set-build-description")
    .description("Set the description of a build")
    .argument("<name>", "job name")
    .argument("<build>", "build number")
    .argument("<description>", "description text")
    .action(async (name: string, build: string, description: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.postForm(
        `${ctx.client.jobUrl(name)}/${build}/submitDescription`,
        { description },
      );
      console.log(`Build #${build} description updated.`);
    });

  program
    .command("set-build-display-name")
    .description("Set the display name of a build")
    .argument("<name>", "job name")
    .argument("<build>", "build number")
    .argument("<displayName>", "display name")
    .action(async (name: string, build: string, displayName: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.postForm(
        `${ctx.client.jobUrl(name)}/${build}/configSubmit`,
        { displayName },
      );
      console.log(`Build #${build} display name updated.`);
    });

  program
    .command("delete-builds")
    .description("Delete builds")
    .argument("<name>", "job name")
    .argument("<builds>", "build numbers (comma-separated)")
    .action(async (name: string, builds: string) => {
      const ctx = await resolveContext(program);
      const jobUrl = ctx.client.jobUrl(name);
      const numbers = builds.split(",").map((s) => s.trim());
      await Promise.all(numbers.map((num) => ctx.client.post(`${jobUrl}/${num}/doDelete`)));
      console.log(`Deleted build(s): ${numbers.join(", ")}`);
    });

  program
    .command("list-changes")
    .description("List changes in a build")
    .argument("<name>", "job name")
    .argument("[build]", "build number (default: lastBuild)", "lastBuild")
    .action(async (name: string, build: string) => {
      const ctx = await resolveContext(program);
      const data = await ctx.client.getJson<{
        changeSet: {
          items: { msg: string; author: { fullName: string }; date?: string }[];
        };
      }>(
        `${ctx.client.jobUrl(name)}/${build}/api/json?tree=changeSet[items[msg,author[fullName],date]]`,
      );
      const items = data.changeSet?.items ?? [];
      if (ctx.json) {
        printOutput(items, true);
      } else if (items.length === 0) {
        console.log("No changes.");
      } else {
        for (const item of items) {
          console.log(`${item.author.fullName}: ${item.msg}`);
        }
      }
    });
}

interface BuildInfo {
  number: number;
  result: string | null;
  building: boolean;
  duration: number;
  estimatedDuration: number;
  timestamp: number;
  displayName: string | null;
  description: string | null;
  url: string;
}

export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  const seconds = Math.floor(ms / 1000) % 60;
  const minutes = Math.floor(ms / 60000) % 60;
  const hours = Math.floor(ms / 3600000);
  if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`;
  if (minutes > 0) return `${minutes}m ${seconds}s`;
  return `${seconds}s`;
}
