import { describe, expect, test, afterEach } from "bun:test";
import { resolveCredentials } from "../src/auth.ts";

describe("resolveCredentials", () => {
  const originalEnv = { ...process.env };

  afterEach(() => {
    process.env.JENKINS_USER_ID = originalEnv.JENKINS_USER_ID;
    process.env.JENKINS_API_TOKEN = originalEnv.JENKINS_API_TOKEN;
  });

  test("auth flag takes priority", async () => {
    process.env.JENKINS_USER_ID = "envuser";
    process.env.JENKINS_API_TOKEN = "envtoken";
    const creds = await resolveCredentials("https://ci.example.com", "flaguser:flagtoken");
    expect(creds).toEqual({ username: "flaguser", token: "flagtoken" });
  });

  test("auth flag with colon in token", async () => {
    const creds = await resolveCredentials("https://ci.example.com", "user:token:with:colons");
    expect(creds).toEqual({ username: "user", token: "token:with:colons" });
  });

  test("invalid auth flag throws", async () => {
    expect(resolveCredentials("https://ci.example.com", "nocolon")).rejects.toThrow(
      "Invalid -auth format",
    );
  });

  test("env vars used when no auth flag", async () => {
    process.env.JENKINS_USER_ID = "envuser";
    process.env.JENKINS_API_TOKEN = "envtoken";
    const creds = await resolveCredentials("https://ci.example.com");
    expect(creds).toEqual({ username: "envuser", token: "envtoken" });
  });
});
