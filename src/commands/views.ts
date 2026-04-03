import type { Command } from "commander";
import { resolveContext } from "../cli.ts";
import { printOutput } from "../output.ts";

export function registerViewCommands(program: Command): void {
  program
    .command("get-view")
    .description("Get view configuration XML")
    .argument("<name>", "view name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      const xml = await ctx.client.getText(
        `/view/${encodeURIComponent(name)}/config.xml`,
      );
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
      const xml = await ctx.client.readStdin();
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
      await ctx.client.post(`/view/${encodeURIComponent(name)}/doDelete`);
      console.log(`View '${name}' deleted.`);
    });

  program
    .command("update-view")
    .description("Update view configuration from XML on stdin")
    .argument("<name>", "view name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      const xml = await ctx.client.readStdin();
      await ctx.client.post(
        `/view/${encodeURIComponent(name)}/config.xml`,
        { body: xml, contentType: "application/xml" },
      );
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
        `/view/${encodeURIComponent(view)}/addJobToView?name=${encodeURIComponent(job)}`,
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
        `/view/${encodeURIComponent(view)}/removeJobFromView?name=${encodeURIComponent(job)}`,
      );
      console.log(`Job '${job}' removed from view '${view}'.`);
    });
}
