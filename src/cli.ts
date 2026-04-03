import { Command } from "commander";
import { resolveCredentials, type Credentials } from "./auth.ts";
import { JenkinsClient } from "./client.ts";
import { registerSystemCommands } from "./commands/system.ts";
import { registerConfigureCommands } from "./commands/configure.ts";
import { registerJobCommands } from "./commands/jobs.ts";
import { registerBuildCommands } from "./commands/builds.ts";
import { registerNodeCommands } from "./commands/nodes.ts";
import { registerViewCommands } from "./commands/views.ts";
import { registerPluginCommands } from "./commands/plugins.ts";
import { registerGroovyCommands } from "./commands/groovy.ts";
import { registerSkillCommands } from "./commands/skills.ts";

export interface GlobalOptions {
  s?: string;
  auth?: string;
  json: boolean;
}

export interface CommandContext {
  client: JenkinsClient;
  credentials: Credentials;
  json: boolean;
}

export function createProgram(): Command {
  const program = new Command();

  program
    .name("jenkins")
    .description("Jenkins CLI client")
    .version("0.2.0")
    .option("-s <url>", "Jenkins server URL")
    .option("--auth <credentials>", "username:apitoken")
    .option("--json", "output as JSON where supported", false);

  registerSystemCommands(program);
  registerConfigureCommands(program);
  registerJobCommands(program);
  registerBuildCommands(program);
  registerNodeCommands(program);
  registerViewCommands(program);
  registerPluginCommands(program);
  registerGroovyCommands(program);
  registerSkillCommands(program);

  return program;
}

export async function resolveContext(program: Command): Promise<CommandContext> {
  const opts = program.opts<GlobalOptions>();
  const serverUrl = opts.s || process.env.JENKINS_URL;
  if (!serverUrl) {
    throw new Error("Jenkins server URL required. Use -s <url> or set JENKINS_URL");
  }
  const credentials = await resolveCredentials(serverUrl, opts.auth);
  return {
    client: new JenkinsClient(serverUrl, credentials),
    credentials,
    json: opts.json,
  };
}
