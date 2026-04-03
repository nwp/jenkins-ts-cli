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
