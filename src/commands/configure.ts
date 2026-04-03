import type { Command } from "commander";
import { storeCredentials, getStoredUsername, getStoredToken, deleteCredentials } from "../keychain.ts";
import { normalizeUrl } from "../paths.ts";

export function registerConfigureCommands(program: Command): void {
  const configure = program
    .command("configure")
    .description("Manage stored Jenkins credentials");

  configure
    .command("set")
    .description("Store credentials for a Jenkins server")
    .requiredOption("-s <url>", "Jenkins server URL")
    .requiredOption("-u, --username <username>", "Jenkins username")
    .requiredOption("-t, --token <token>", "Jenkins API token")
    .action(async (opts: { s: string; username: string; token: string }) => {
      await storeCredentials(opts.s, opts.username, opts.token);
      console.log(`Credentials stored for ${normalizeUrl(opts.s)}`);
    });

  configure
    .command("show")
    .description("Show stored credentials for a Jenkins server")
    .requiredOption("-s <url>", "Jenkins server URL")
    .action(async (opts: { s: string }) => {
      const url = normalizeUrl(opts.s);
      const username = await getStoredUsername(url);
      if (!username) {
        console.log(`No credentials stored for ${url}`);
        return;
      }
      const token = await getStoredToken(url, username);
      console.log(`Server:   ${url}`);
      console.log(`Username: ${username}`);
      console.log(`Token:    ${token ? "********" : "(not found in Keychain)"}`);
    });

  configure
    .command("clear")
    .description("Remove stored credentials for a Jenkins server")
    .requiredOption("-s <url>", "Jenkins server URL")
    .action(async (opts: { s: string }) => {
      const url = normalizeUrl(opts.s);
      const removed = await deleteCredentials(url);
      if (removed) {
        console.log(`Credentials removed for ${url}`);
      } else {
        console.log(`No credentials found for ${url}`);
      }
    });
}
