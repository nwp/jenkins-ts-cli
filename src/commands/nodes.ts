import type { Command } from "commander";
import type { JenkinsClient } from "../client.ts";
import { resolveContext } from "../cli.ts";
import { nodeUrl } from "../paths.ts";
import { readStdin } from "../stdin.ts";

async function waitForNodeState(
  client: JenkinsClient,
  name: string,
  targetOffline: boolean,
  timeout: number,
): Promise<boolean> {
  const url = nodeUrl(name);
  for (let i = 0; i < timeout; i++) {
    const status = await client.getJson<{ offline: boolean }>(`${url}/api/json?tree=offline`);
    if (status.offline === targetOffline) return true;
    await Bun.sleep(1000);
  }
  return false;
}

export function registerNodeCommands(program: Command): void {
  program
    .command("create-node")
    .description("Create a new node from XML on stdin")
    .argument("<name>", "node name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      const xml = await readStdin();
      await ctx.client.post(
        `/computer/doCreateItem?name=${encodeURIComponent(name)}&type=hudson.slaves.DumbSlave`,
        { body: xml, contentType: "application/xml" },
      );
      console.log(`Node '${name}' created.`);
    });

  program
    .command("delete-node")
    .description("Delete a node")
    .argument("<name>", "node name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.post(`${nodeUrl(name)}/doDelete`);
      console.log(`Node '${name}' deleted.`);
    });

  program
    .command("update-node")
    .description("Update node configuration from XML on stdin")
    .argument("<name>", "node name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      const xml = await readStdin();
      await ctx.client.post(`${nodeUrl(name)}/config.xml`, {
        body: xml,
        contentType: "application/xml",
      });
      console.log(`Node '${name}' updated.`);
    });

  program
    .command("connect-node")
    .description("Connect a node")
    .argument("<name>", "node name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.post(`${nodeUrl(name)}/launchSlaveAgent`);
      console.log(`Node '${name}' connection initiated.`);
    });

  program
    .command("disconnect-node")
    .description("Disconnect a node")
    .argument("<name>", "node name")
    .option("-m, --message <message>", "offline message", "")
    .action(async (name: string, opts: { message: string }) => {
      const ctx = await resolveContext(program);
      await ctx.client.postForm(`${nodeUrl(name)}/doDisconnect`, {
        offlineMessage: opts.message,
      });
      console.log(`Node '${name}' disconnected.`);
    });

  program
    .command("online-node")
    .description("Bring a node online")
    .argument("<name>", "node name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      const url = nodeUrl(name);
      const status = await ctx.client.getJson<{ offline: boolean }>(`${url}/api/json?tree=offline`);
      if (!status.offline) {
        console.log(`Node '${name}' is already online.`);
        return;
      }
      await ctx.client.post(`${url}/toggleOffline`);
      console.log(`Node '${name}' brought online.`);
    });

  program
    .command("offline-node")
    .description("Take a node offline")
    .argument("<name>", "node name")
    .option("-m, --message <message>", "offline message", "")
    .action(async (name: string, opts: { message: string }) => {
      const ctx = await resolveContext(program);
      const url = nodeUrl(name);
      const status = await ctx.client.getJson<{ offline: boolean }>(`${url}/api/json?tree=offline`);
      if (status.offline) {
        console.log(`Node '${name}' is already offline.`);
        return;
      }
      await ctx.client.postForm(`${url}/toggleOffline`, {
        offlineMessage: opts.message,
      });
      console.log(`Node '${name}' taken offline.`);
    });

  program
    .command("wait-node-online")
    .description("Wait for a node to come online")
    .argument("<name>", "node name")
    .option("--timeout <seconds>", "timeout in seconds", "60")
    .action(async (name: string, opts: { timeout: string }) => {
      const ctx = await resolveContext(program);
      const ok = await waitForNodeState(ctx.client, name, false, parseInt(opts.timeout, 10));
      if (ok) {
        console.log(`Node '${name}' is online.`);
      } else {
        console.error(`Timed out waiting for node '${name}' to come online.`);
        process.exit(1);
      }
    });

  program
    .command("wait-node-offline")
    .description("Wait for a node to go offline")
    .argument("<name>", "node name")
    .option("--timeout <seconds>", "timeout in seconds", "60")
    .action(async (name: string, opts: { timeout: string }) => {
      const ctx = await resolveContext(program);
      const ok = await waitForNodeState(ctx.client, name, true, parseInt(opts.timeout, 10));
      if (ok) {
        console.log(`Node '${name}' is offline.`);
      } else {
        console.error(`Timed out waiting for node '${name}' to go offline.`);
        process.exit(1);
      }
    });
}
