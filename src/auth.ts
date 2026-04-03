import { getStoredUsername, getStoredToken, storeCredentials } from "./keychain.ts";

export interface Credentials {
  username: string;
  token: string;
}

export async function resolveCredentials(
  serverUrl: string,
  authFlag?: string,
): Promise<Credentials> {
  if (authFlag) {
    const [username, ...rest] = authFlag.split(":");
    const token = rest.join(":");
    if (!username || !token) {
      throw new Error("Invalid -auth format. Expected username:token");
    }
    return { username, token };
  }

  const envUser = process.env.JENKINS_USER_ID;
  const envToken = process.env.JENKINS_API_TOKEN;
  if (envUser && envToken) {
    return { username: envUser, token: envToken };
  }

  const storedUsername = await getStoredUsername(serverUrl);
  if (storedUsername) {
    const storedToken = await getStoredToken(serverUrl, storedUsername);
    if (storedToken) {
      return { username: storedUsername, token: storedToken };
    }
  }

  return promptAndStore(serverUrl);
}

async function promptAndStore(serverUrl: string): Promise<Credentials> {
  const username = await prompt("Jenkins username: ");
  if (!username) throw new Error("Username is required");

  const token = await prompt("Jenkins API token: ");
  if (!token) throw new Error("API token is required");

  await storeCredentials(serverUrl, username, token);
  return { username, token };
}

function prompt(message: string): Promise<string> {
  process.stdout.write(message);
  return new Promise((resolve) => {
    let data = "";
    process.stdin.setEncoding("utf-8");
    process.stdin.resume();
    process.stdin.once("data", (chunk: string) => {
      data = chunk.trim();
      process.stdin.pause();
      resolve(data);
    });
  });
}
