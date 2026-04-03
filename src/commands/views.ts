import type { Command } from "commander";
import { resolveContext } from "../cli.ts";
import { printOutput } from "../output.ts";
import { viewUrl } from "../paths.ts";
import { readStdin } from "../stdin.ts";

export function registerViewCommands(program: Command): void {
  program
    .command("get-view")
    .description("Get view configuration XML")
    .argument("<name>", "view name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      const xml = await ctx.client.getText(`${viewUrl(name)}/config.xml`);
      if (ctx.json) {
        printOutput({ name, xml }, true);
      } else {
        process.stdout.write(xml);
      }
    });

  program
    .command("create-view")
    .description("Create a new view from XML on stdin")
    .argument("<name>", "view name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      const xml = await readStdin();
      await ctx.client.post(
        `/createView?name=${encodeURIComponent(name)}`,
        { body: xml, contentType: "application/xml" },
      );
      console.log(`View '${name}' created.`);
    });

  program
    .command("delete-view")
    .description("Delete a view")
    .argument("<name>", "view name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.post(`${viewUrl(name)}/doDelete`);
      console.log(`View '${name}' deleted.`);
    });

  program
    .command("update-view")
    .description("Update view configuration from XML on stdin")
    .argument("<name>", "view name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      const xml = await readStdin();
      await ctx.client.post(`${viewUrl(name)}/config.xml`, {
        body: xml,
        contentType: "application/xml",
      });
      console.log(`View '${name}' updated.`);
    });

  program
    .command("add-job-to-view")
    .description("Add a job to a view")
    .argument("<view>", "view name")
    .argument("<job>", "job name")
    .action(async (view: string, job: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.post(
        `${viewUrl(view)}/addJobToView?name=${encodeURIComponent(job)}`,
      );
      console.log(`Job '${job}' added to view '${view}'.`);
    });

  program
    .command("remove-job-from-view")
    .description("Remove a job from a view")
    .argument("<view>", "view name")
    .argument("<job>", "job name")
    .action(async (view: string, job: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.post(
        `${viewUrl(view)}/removeJobFromView?name=${encodeURIComponent(job)}`,
      );
      console.log(`Job '${job}' removed from view '${view}'.`);
    });
}
