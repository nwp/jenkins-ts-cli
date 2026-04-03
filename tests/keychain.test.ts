import { describe, expect, test } from "bun:test";
import { normalizeUrl } from "../src/paths.ts";

describe("keychain URL normalization", () => {
  test("same URL variants normalize to same key", () => {
    const a = normalizeUrl("https://ci.example.com/jenkins/");
    const b = normalizeUrl("https://CI.Example.COM/jenkins");
    expect(a).toBe(b);
  });

  test("different paths produce different keys", () => {
    const a = normalizeUrl("https://ci.example.com/jenkins");
    const b = normalizeUrl("https://ci.example.com/ci");
    expect(a).not.toBe(b);
  });
});
