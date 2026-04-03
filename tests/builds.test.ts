import { describe, expect, test } from "bun:test";
import { formatDuration } from "../src/commands/builds.ts";

describe("formatDuration", () => {
  test("sub-second", () => {
    expect(formatDuration(0)).toBe("0ms");
    expect(formatDuration(500)).toBe("500ms");
    expect(formatDuration(999)).toBe("999ms");
  });

  test("seconds only", () => {
    expect(formatDuration(1000)).toBe("1s");
    expect(formatDuration(45000)).toBe("45s");
    expect(formatDuration(59999)).toBe("59s");
  });

  test("minutes and seconds", () => {
    expect(formatDuration(60000)).toBe("1m 0s");
    expect(formatDuration(90000)).toBe("1m 30s");
    expect(formatDuration(754000)).toBe("12m 34s");
  });

  test("hours, minutes, and seconds", () => {
    expect(formatDuration(3600000)).toBe("1h 0m 0s");
    expect(formatDuration(3661000)).toBe("1h 1m 1s");
    expect(formatDuration(7384000)).toBe("2h 3m 4s");
  });
});
