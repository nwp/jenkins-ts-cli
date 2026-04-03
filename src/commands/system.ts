import type { Command } from "commander";
import { resolveContext } from "../cli.ts";
import { printOutput } from "../output.ts";

export function registerSystemCommands(program: Command): void {
  program
    .command("version")
    .description("Show Jenkins version")
    .action(async () => {
      const ctx = await resolveContext(program);
      const res = await ctx.client.get("/api/json?tree=version", { raw: true });
      const version =
        res.headers.get("X-Jenkins") ??
        ((await res.json()) as { version: string }).version;
      printOutput(ctx.json ? { version } : version, ctx.json);
    });

  program
    .command("who-am-i")
    .description("Show current user credentials and permissions")
    .action(async () => {
      const ctx = await resolveContext(program);
      const data = await ctx.client.getJson<{
        id: string;
        fullName: string;
        authorities?: { authority: string }[];
      }>("/me/api/json");
      if (ctx.json) {
        printOutput(data, true);
      } else {
        console.log(`Authenticated as: ${data.fullName}`);
        if (data.authorities) {
          console.log("Authorities:");
          for (const a of data.authorities) {
            console.log(`  ${a.authority}`);
          }
        }
      }
    });

  program
    .command("quiet-down")
    .description("Put Jenkins into quiet mode")
    .action(async () => {
      const ctx = await resolveContext(program);
      await ctx.client.post("/quietDown");
      console.log("Jenkins is entering quiet mode.");
    });

  program
    .command("cancel-quiet-down")
    .description("Cancel quiet mode")
    .action(async () => {
      const ctx = await resolveContext(program);
      await ctx.client.post("/cancelQuietDown");
      console.log("Quiet mode cancelled.");
    });

  program
    .command("clear-queue")
    .description("Clear the build queue")
    .action(async () => {
      const ctx = await resolveContext(program);
      const queue = await ctx.client.getJson<{
        items: { id: number }[];
      }>("/queue/api/json");
      for (const item of queue.items) {
        await ctx.client.post(`/queue/cancelItem?id=${item.id}`, { raw: true });
      }
      console.log(`Cleared ${queue.items.length} item(s) from the queue.`);
    });

  program
    .command("reload-configuration")
    .description("Reload configuration from disk")
    .action(async () => {
      const ctx = await resolveContext(program);
      await ctx.client.post("/reload");
      console.log("Configuration reloaded.");
    });
}
