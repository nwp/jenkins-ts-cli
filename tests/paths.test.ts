import { describe, expect, test } from "bun:test";
import { jobPath, normalizeUrl } from "../src/paths.ts";

describe("jobPath", () => {
  test("simple job name", () => {
    expect(jobPath("my-job")).toBe("job/my-job");
  });

  test("nested folder path", () => {
    expect(jobPath("folder/subfolder/my-job")).toBe(
      "job/folder/job/subfolder/job/my-job",
    );
  });

  test("single folder", () => {
    expect(jobPath("folder/job")).toBe("job/folder/job/job");
  });

  test("encodes special characters", () => {
    expect(jobPath("my folder/my job")).toBe(
      "job/my%20folder/job/my%20job",
    );
  });
});

describe("normalizeUrl", () => {
  test("strips trailing slash", () => {
    expect(normalizeUrl("https://ci.example.com/")).toBe(
      "https://ci.example.com",
    );
  });

  test("lowercases scheme and host", () => {
    expect(normalizeUrl("HTTPS://CI.Example.COM/jenkins")).toBe(
      "https://ci.example.com/jenkins",
    );
  });

  test("preserves path", () => {
    expect(normalizeUrl("https://ci.example.com/jenkins/")).toBe(
      "https://ci.example.com/jenkins",
    );
  });

  test("no trailing slash on bare host", () => {
    expect(normalizeUrl("https://ci.example.com")).toBe(
      "https://ci.example.com",
    );
  });
});
