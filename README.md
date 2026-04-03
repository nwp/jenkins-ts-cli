# jenkins-ts-cli

A TypeScript reimplementation of the [Jenkins CLI](https://www.jenkins.io/doc/book/managing/cli/) built with [Bun](https://bun.sh). Compiles to a single self-contained binary with no runtime dependencies. Designed for both interactive use and consumption by AI coding agents (Claude Code, Codex, Copilot CLI).

## Quick Start

```bash
# Install dependencies
bun install

# Run directly
bun run src/index.ts -s https://jenkins.example.com version

# Or compile to a standalone binary
bun run build
./jenkins -s https://jenkins.example.com version
```

## Installation

**Requirements:** [Bun](https://bun.sh) v1.2+ (for building). The compiled binary has no runtime dependencies.

```bash
git clone https://github.com/nwp/jenkins-ts-cli.git
cd jenkins-ts-cli
bun install
bun run build
cp jenkins /usr/local/bin/jenkins
```

## Authentication

Credentials are resolved in the following order (first match wins):

1. **`-auth` flag** (or `--auth`) — passed directly on the command line as `username:apitoken`
2. **Environment variables** — `JENKINS_USER_ID` and `JENKINS_API_TOKEN`
3. **macOS Keychain** — looked up by the normalized server URL
4. **Interactive prompt** — asks for username and API token, then stores them in the Keychain for future use

### Keychain Storage

On first use with a new server URL, the CLI prompts for credentials and stores them in the macOS Keychain via the `security` command. Subsequent invocations with the same URL retrieve credentials automatically.

Credentials are keyed by the normalized server URL (lowercased scheme+host, trailing slash stripped, path preserved). A username-to-URL mapping is stored at `~/.jenkins-cli/credentials.json`.

```bash
# Manually manage stored credentials
jenkins configure set -s https://jenkins.example.com -u admin -t your-api-token
jenkins configure show -s https://jenkins.example.com
jenkins configure clear -s https://jenkins.example.com
```

### For AI Agents

The recommended approach for non-interactive use is environment variables:

```bash
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USER_ID=admin
export JENKINS_API_TOKEN=your-token
jenkins version
```

When `JENKINS_URL` is set, the `-s` flag can be omitted.

## Usage

```text
jenkins -s <url> [-auth <user:token>] [--json] <command> [args...]
```

### Global Options

| Option | Description |
|---|---|
| `-s <url>` | Jenkins server URL (or set `JENKINS_URL`) |
| `-auth <user:token>` | Authentication credentials (single dash, matching the Java CLI) |
| `--json` | Output structured JSON where supported |
| `-V, --version` | Show CLI version |

### Commands

#### System

```bash
jenkins -s URL version                # Show Jenkins server version
jenkins -s URL who-am-i               # Show authenticated user and permissions
jenkins -s URL quiet-down             # Put Jenkins into quiet mode (no new builds)
jenkins -s URL cancel-quiet-down      # Cancel quiet mode
jenkins -s URL clear-queue            # Cancel all queued builds
jenkins -s URL reload-configuration   # Reload configuration from disk
```

#### Jobs

```bash
jenkins -s URL list-jobs                          # List all jobs (recursively traverses folders)
jenkins -s URL get-job my-job                     # Print job config XML to stdout
jenkins -s URL get-job my-folder/my-job           # Folder-aware path syntax
jenkins -s URL create-job my-job < config.xml     # Create job from XML on stdin
jenkins -s URL update-job my-job < config.xml     # Update job config from XML on stdin
jenkins -s URL copy-job src-job dest-job          # Copy a job
jenkins -s URL delete-job my-job                  # Delete a job
jenkins -s URL reload-job my-job                  # Reload job config from disk
```

#### Builds

```bash
jenkins -s URL build my-job                                # Trigger a build
jenkins -s URL build my-job -p key1=val1 -p key2=val2      # Parameterized build
jenkins -s URL build my-job --follow                       # Trigger and stream console output
jenkins -s URL build my-job --wait                         # Trigger and wait for completion
jenkins -s URL console my-job                              # Print console output (last build)
jenkins -s URL console my-job 42                           # Print console output for build #42
jenkins -s URL console my-job --follow                     # Stream console output in real-time
jenkins -s URL set-build-description my-job 42 "Fixed it"  # Set build description
jenkins -s URL set-build-display-name my-job 42 "v1.0"     # Set build display name
jenkins -s URL delete-builds my-job 40,41,42               # Delete builds (comma-separated)
jenkins -s URL list-changes my-job 42                      # List SCM changes in a build
```

#### Nodes

```bash
jenkins -s URL create-node my-agent < node-config.xml  # Create node from XML on stdin
jenkins -s URL update-node my-agent < node-config.xml  # Update node config from XML on stdin
jenkins -s URL delete-node my-agent                    # Delete a node
jenkins -s URL connect-node my-agent                   # Initiate node connection
jenkins -s URL disconnect-node my-agent                # Disconnect a node
jenkins -s URL disconnect-node my-agent -m "reason"    # Disconnect with message
jenkins -s URL online-node my-agent                    # Bring node online
jenkins -s URL offline-node my-agent                   # Take node offline
jenkins -s URL offline-node my-agent -m "maintenance"  # Take offline with message
jenkins -s URL wait-node-online my-agent               # Wait for node to come online
jenkins -s URL wait-node-online my-agent --timeout 120 # Custom timeout (default: 60s)
jenkins -s URL wait-node-offline my-agent              # Wait for node to go offline
```

#### Views

```bash
jenkins -s URL get-view my-view                           # Print view config XML to stdout
jenkins -s URL create-view my-view < view-config.xml      # Create view from XML on stdin
jenkins -s URL update-view my-view < view-config.xml      # Update view config from XML on stdin
jenkins -s URL delete-view my-view                        # Delete a view
jenkins -s URL add-job-to-view my-view my-job             # Add a job to a view
jenkins -s URL remove-job-from-view my-view my-job        # Remove a job from a view
```

#### Plugins

```bash
jenkins -s URL list-plugins                    # List installed plugins
jenkins -s URL install-plugin git              # Install by plugin ID (defaults to @latest)
jenkins -s URL install-plugin git@4.14.0       # Install specific version
jenkins -s URL install-plugin https://...hpi   # Install from URL
jenkins -s URL install-plugin ./my-plugin.hpi  # Install from local file
jenkins -s URL enable-plugin git               # Enable a plugin
jenkins -s URL disable-plugin git              # Disable a plugin
```

#### Groovy

```bash
jenkins -s URL groovy script.groovy            # Execute a Groovy script file
echo 'println Jenkins.instance.version' | jenkins -s URL groovy -   # Execute from stdin
jenkins -s URL groovysh                        # Interactive Groovy shell (requires Java)
```

### JSON Output

Commands that return structured data support the `--json` flag:

```bash
jenkins -s URL --json list-jobs
jenkins -s URL --json list-plugins
jenkins -s URL --json who-am-i
jenkins -s URL --json version
jenkins -s URL --json list-changes my-job 42
jenkins -s URL --json get-view my-view
```

Streaming commands (`console`, `build --follow`, `groovy`, `groovysh`) always output plain text.

## Departures from the Official Jenkins CLI

This CLI reimplements the Jenkins CLI using the REST API rather than the binary WebSocket protocol (PlainCLIProtocol) used by the official `jenkins-cli.jar`. This has several consequences:

### Behavioral Differences

| Area | Official CLI (Java) | This CLI |
|---|---|---|
| **Transport** | Binary WebSocket protocol (PlainCLIProtocol) over `/cli/ws` | REST API (`/api/json`, form POSTs, XML config endpoints) |
| **Console streaming** | True real-time streaming over WebSocket | Polling via `/logText/progressiveText` with ~1 second intervals |
| **`groovysh`** | Native binary protocol support | Delegates to the official `jenkins-cli.jar` via Java subprocess |
| **`session-id`** | Returns the CLI protocol session ID | Not implemented (CLI protocol concept with no REST equivalent) |
| **`-auth` flag** | Single-dash `-auth` natively | Accepts both `-auth` and `--auth` (normalized internally) |
| **`--json` flag** | Not supported | Added for AI agent consumption |
| **`configure` command** | Not present | Added for Keychain credential management |
| **SSH authentication** | Supported via `-ssh` flag | Not supported |
| **Credential storage** | No built-in credential storage | macOS Keychain integration with auto-prompt on first use |
| **Folder traversal** | Opaque (handled by server) | `list-jobs` recursively traverses folders in parallel |
| **`--follow` polling** | Instant streaming | ~1 second polling interval |
| **Exit codes** | Specific codes per error type | 0 for success, 1 for all errors |

### Not Implemented

- **`session-id`** — This command returns the session identifier of the binary CLI protocol channel. Since this CLI uses REST, there is no session to report.
- **SSH authentication** — The `-ssh` flag and SSH key-based auth are not supported. Use `-auth` with a username and API token instead.
- **WebSocket/HTTP binary transport** — The PlainCLIProtocol is not implemented. All commands go through the REST API except `groovysh`, which delegates to the official JAR.

### Added Features

- **`--json` flag** — Structured JSON output on supported commands for programmatic consumption.
- **`configure` command** — Manage stored credentials (`set`, `show`, `clear`) without running a real Jenkins command.
- **macOS Keychain integration** — Credentials stored securely and resolved automatically per server URL.
- **`JENKINS_URL` environment variable** — Allows omitting `-s` entirely when the env var is set.
- **Parallel operations** — Queue clearing, build deletion, and folder traversal run concurrently.

### `groovysh` and the JAR Dependency

The `groovysh` command requires an interactive REPL session over the binary WebSocket protocol, which cannot be replicated via REST. When invoked, this CLI:

1. Downloads `jenkins-cli.jar` from the connected Jenkins server (`<url>/jnlpJars/jenkins-cli.jar`)
2. Caches it at `~/.jenkins-cli/jenkins-cli.jar` (validates freshness via `ETag`/`Last-Modified`)
3. Invokes it via `java -jar` with credentials passed at spawn time

This requires Java to be installed and on your `PATH` (or `JAVA_HOME` set). If Java is not available, the command fails with a clear error message. All other commands work without Java.

## Project Structure

```text
src/
├── index.ts          # Entry point, argv preprocessing (-auth → --auth)
├── cli.ts            # Commander program definition, global flags, command registration
├── auth.ts           # Credential resolution chain (flag → env → keychain → prompt)
├── keychain.ts       # macOS Keychain wrapper via `security` CLI
├── client.ts         # Jenkins REST client (fetch, CRUMB handling, error formatting)
├── paths.ts          # URL normalization, folder-aware path helpers, shared constants
├── console.ts        # Shared progressive console output polling
├── stdin.ts          # Shared stdin reader
├── output.ts         # Plain text / JSON output formatter
├── jar.ts            # JAR download, caching, version checking, Java subprocess
└── commands/
    ├── system.ts     # version, who-am-i, quiet-down, cancel-quiet-down, clear-queue, reload-configuration
    ├── configure.ts  # configure set/show/clear
    ├── jobs.ts       # list-jobs, get-job, create-job, copy-job, delete-job, update-job, reload-job, build
    ├── builds.ts     # console, set-build-description, set-build-display-name, delete-builds, list-changes
    ├── nodes.ts      # create-node, delete-node, update-node, connect-node, disconnect-node, online/offline-node, wait-node-*
    ├── views.ts      # get-view, create-view, delete-view, update-view, add/remove-job-to/from-view
    ├── plugins.ts    # list-plugins, install-plugin, enable-plugin, disable-plugin
    └── groovy.ts     # groovy (REST), groovysh (JAR subprocess)

tests/
├── auth.test.ts      # Credential resolution priority
├── client.test.ts    # URL normalization, job path construction
├── keychain.test.ts  # Keychain URL key consistency
├── output.test.ts    # JSON and plain text formatting
├── paths.test.ts     # Folder path encoding, URL normalization
└── integration/      # Opt-in integration tests (require live Jenkins)
    └── smoke.test.ts
```

## Development

```bash
# Install dependencies
bun install

# Run from source
bun run src/index.ts -s https://jenkins.example.com version

# Type check
bun run typecheck

# Run tests
bun test

# Compile to binary
bun run build
```

### Running Integration Tests

Integration tests require a live Jenkins instance and are skipped by default. To run them:

```bash
JENKINS_TEST_URL=https://jenkins.example.com \
JENKINS_TEST_USER=admin \
JENKINS_TEST_TOKEN=your-api-token \
bun test tests/integration/
```

### Architecture Notes for Contributors

**REST-first approach.** Almost every command is implemented via the Jenkins REST API (`/api/json` for reads, form POSTs and XML config endpoints for writes). This avoids the complexity of the binary WebSocket protocol. The only exception is `groovysh`, which delegates to the official JAR.

**CRUMB handling.** CSRF protection tokens are fetched once per invocation from `/crumbIssuer/api/json` and injected into all POST requests. If the crumb issuer is disabled (returns 404/403), requests proceed without a crumb.

**Folder-aware paths.** Jobs in folders use a slash-separated syntax (`folder/subfolder/my-job`) which is translated to the Jenkins URL convention (`/job/folder/job/subfolder/job/my-job`). The `jobPath()` helper in `paths.ts` handles this.

**Credential isolation.** The `CommandContext` object exposes both the `JenkinsClient` (for REST calls) and raw `Credentials` (for JAR subprocess invocation). This avoids double-resolving credentials.

**Adding a new command.** Create or extend a command file in `src/commands/`, register it in `cli.ts` via a `registerXxxCommands(program)` function, and follow the existing pattern of calling `resolveContext(program)` at the start of each action handler.

### Dependencies

| Dependency | Purpose |
|---|---|
| `commander` | CLI argument parsing |
| `@types/bun` | Bun type definitions (dev) |
| `typescript` | Type checking (dev) |

No other runtime dependencies. HTTP is handled by Bun's built-in `fetch`, subprocesses by `Bun.spawn`, and credential storage by the macOS `security` CLI.

## License

MIT
