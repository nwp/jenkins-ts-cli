import type { Command } from "commander";
import { resolveContext, type GlobalOptions } from "../cli.ts";
import { resolveCredentials } from "../auth.ts";
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
        code = await ctx.client.readStdin();
      } else {
        const file = Bun.file(script);
        if (!(await file.exists())) {
          throw new Error(`Script file not found: ${script}`);
        }
        code = await file.text();
      }

      const res = await ctx.client.postForm("/scriptText", { script: code });
      const output = await res.text();
      if (output) process.stdout.write(output);
    });

  program
    .command("groovysh")
    .description("Start an interactive Groovy shell (requires Java)")
    .action(async () => {
      const opts = program.opts<GlobalOptions>();
      const serverUrl = opts.s || process.env.JENKINS_URL;
      if (!serverUrl) {
        throw new Error("Jenkins server URL required. Use -s <url> or set JENKINS_URL");
      }
      const credentials = await resolveCredentials(serverUrl, opts.auth);
      const ctx = await resolveContext(program);
      const exitCode = await runJar(ctx.client, credentials, ["groovysh"]);
      process.exit(exitCode);
    });
}
