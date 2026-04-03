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
