import type { Command } from "commander";
import { resolveContext } from "../cli.ts";
import { readStdin } from "../stdin.ts";
import { runJar } from "../jar.ts";

export function registerGroovyCommands(program: Command): void {
  program
    .command("groovy")
    .description("Execute a Groovy script")
    .argument("[script]", "script file path, or '-' for stdin")
    .action(async (script: string | undefined) => {
      const ctx = await resolveContext(program);

      let code: string;
      if (!script || script === "-") {
        code = await readStdin();
      } else {
        try {
          code = await Bun.file(script).text();
        } catch {
          throw new Error(`Script file not found or unreadable: ${script}`);
        }
      }

      const res = await ctx.client.postForm("/scriptText", { script: code });
      const output = await res.text();
      if (output) process.stdout.write(output);
    });

  program
    .command("groovysh")
    .description("Start an interactive Groovy shell (requires Java)")
    .action(async () => {
      const ctx = await resolveContext(program);
      const exitCode = await runJar(ctx.client, ctx.credentials, ["groovysh"]);
      process.exit(exitCode);
    });
}
