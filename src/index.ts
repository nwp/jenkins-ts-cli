#!/usr/bin/env bun
import { createProgram } from "./cli.ts";

// Normalize `-auth` to `--auth` for Java CLI compatibility
const argv = process.argv.map((arg) => (arg === "-auth" ? "--auth" : arg));

const program = createProgram();

try {
  await program.parseAsync(argv);
} catch (err) {
  const message = err instanceof Error ? err.message : String(err);
  console.error(`ERROR: ${message}`);
  process.exit(1);
}
