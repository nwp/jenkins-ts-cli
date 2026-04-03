import type { Command } from "commander";
import { resolveContext } from "../cli.ts";
import { printOutput } from "../output.ts";

export function registerPluginCommands(program: Command): void {
  program
    .command("list-plugins")
    .description("List installed plugins")
    .action(async () => {
      const ctx = await resolveContext(program);
      const data = await ctx.client.getJson<{
        plugins: {
          shortName: string;
          version: string;
          active: boolean;
          enabled: boolean;
          longName: string;
        }[];
      }>(
        "/pluginManager/api/json?tree=plugins[shortName,version,active,enabled,longName]",
      );
      if (ctx.json) {
        printOutput(data.plugins, true);
      } else {
        for (const p of data.plugins) {
          const status = p.enabled ? (p.active ? "" : " (inactive)") : " (disabled)";
          console.log(`${p.shortName}:${p.version}${status}`);
        }
      }
    });

  program
    .command("install-plugin")
    .description("Install a plugin")
    .argument("<plugin>", "plugin ID (e.g., git@latest), URL, or local .hpi path")
    .action(async (plugin: string) => {
      const ctx = await resolveContext(program);

      if (plugin.endsWith(".hpi") || plugin.endsWith(".jpi")) {
        const file = Bun.file(plugin);
        if (!(await file.exists())) {
          throw new Error(`Plugin file not found: ${plugin}`);
        }
        const formData = new FormData();
        formData.append("name", new Blob([await file.arrayBuffer()]), file.name ?? "plugin.hpi");
        await ctx.client.post("/pluginManager/uploadPlugin", {
          body: formData,
        });
      } else if (plugin.startsWith("http://") || plugin.startsWith("https://")) {
        const xml = `<jenkins><install plugin="${escapeXml(plugin)}" /></jenkins>`;
        await ctx.client.post("/pluginManager/installNecessaryPlugins", {
          body: xml,
          contentType: "application/xml",
        });
      } else {
        const spec = plugin.includes("@") ? plugin : `${plugin}@latest`;
        const xml = `<jenkins><install plugin="${escapeXml(spec)}" /></jenkins>`;
        await ctx.client.post("/pluginManager/installNecessaryPlugins", {
          body: xml,
          contentType: "application/xml",
        });
      }

      console.log(`Plugin '${plugin}' installation initiated.`);
    });

  program
    .command("enable-plugin")
    .description("Enable a plugin")
    .argument("<name>", "plugin short name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.post(
        `/pluginManager/plugin/${encodeURIComponent(name)}/enable`,
      );
      console.log(`Plugin '${name}' enabled.`);
    });

  program
    .command("disable-plugin")
    .description("Disable a plugin")
    .argument("<name>", "plugin short name")
    .action(async (name: string) => {
      const ctx = await resolveContext(program);
      await ctx.client.post(
        `/pluginManager/plugin/${encodeURIComponent(name)}/disable`,
      );
      console.log(`Plugin '${name}' disabled.`);
    });
}

function escapeXml(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;");
}
