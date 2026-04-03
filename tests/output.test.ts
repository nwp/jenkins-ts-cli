import { describe, expect, test } from "bun:test";
import { formatOutput } from "../src/output.ts";

describe("formatOutput", () => {
  test("json mode formats objects as JSON", () => {
    const result = formatOutput({ name: "test" }, true);
    expect(JSON.parse(result)).toEqual({ name: "test" });
  });

  test("json mode formats arrays as JSON", () => {
    const result = formatOutput([1, 2, 3], true);
    expect(JSON.parse(result)).toEqual([1, 2, 3]);
  });

  test("json mode formats strings as JSON", () => {
    const result = formatOutput("hello", true);
    expect(JSON.parse(result)).toBe("hello");
  });

  test("plain mode returns strings directly", () => {
    expect(formatOutput("hello", false)).toBe("hello");
  });

  test("plain mode joins array of strings with newlines", () => {
    expect(formatOutput(["a", "b", "c"], false)).toBe("a\nb\nc");
  });

  test("plain mode serializes objects as JSON", () => {
    const result = formatOutput({ name: "test" }, false);
    expect(JSON.parse(result)).toEqual({ name: "test" });
  });
});
