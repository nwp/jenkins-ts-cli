import { describe, expect, test } from "bun:test";
import { JenkinsClient } from "../src/client.ts";

describe("JenkinsClient", () => {
  test("jobUrl converts simple name", () => {
    const client = new JenkinsClient("https://ci.example.com", {
      username: "user",
      token: "token",
    });
    expect(client.jobUrl("my-job")).toBe("/job/my-job");
  });

  test("jobUrl converts nested folder path", () => {
    const client = new JenkinsClient("https://ci.example.com", {
      username: "user",
      token: "token",
    });
    expect(client.jobUrl("folder/sub/job")).toBe("/job/folder/job/sub/job/job");
  });

  test("baseUrl is normalized", () => {
    const client = new JenkinsClient("HTTPS://CI.Example.COM/jenkins/", {
      username: "user",
      token: "token",
    });
    expect(client.baseUrl).toBe("https://ci.example.com/jenkins");
  });
});
