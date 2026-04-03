export function skillContent(skillDir: string): string {
  return `---
name: jenkins-cli
description: >
  How to interact with Jenkins CI/CD servers using the jenkins-ts-cli command-line tool.
  Use this skill whenever the user needs to trigger builds, check build status, stream logs,
  manage jobs/nodes/views/plugins, or automate any Jenkins workflow. Triggers on: Jenkins,
  CI/CD, build pipeline, deploy, job status, build logs, Jenkins server, continuous integration.
---

# Jenkins CLI

A TypeScript CLI for Jenkins that uses the REST API. Compiles to a single binary with no runtime dependencies.

## Authentication

Set these environment variables (recommended for non-interactive use):

\`\`\`bash
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USER_ID=admin
export JENKINS_API_TOKEN=your-token
\`\`\`

When \`JENKINS_URL\` is set, the \`-s\` flag can be omitted from all commands. Credentials can also be passed inline with \`-auth user:token\`.

## Command Syntax

\`\`\`text
jenkins [-s <url>] [-auth <user:token>] [--json] <command> [args...]
\`\`\`

The \`--json\` flag returns structured JSON on commands that support it. Use it for programmatic consumption.

## Common Agent Workflows

### Trigger a build and wait for completion

\`\`\`bash
jenkins build my-job --wait
\`\`\`

Exit code 0 means SUCCESS. Exit code 1 means the build failed or an error occurred.

### Trigger a build and stream console output

\`\`\`bash
jenkins build my-job --follow
\`\`\`

### Trigger a parameterized build

\`\`\`bash
jenkins build my-job -p ENV=production -p VERSION=1.2.3 --wait
\`\`\`

### Check the console output of the last build

\`\`\`bash
jenkins console my-job
\`\`\`

### Check console output of a specific build number

\`\`\`bash
jenkins console my-job 42
\`\`\`

### Check build status (pass/fail) without streaming logs

\`\`\`bash
jenkins get-build my-job
\`\`\`

Exit code 0 means SUCCESS. Exit code 1 means the build failed or is still building. Use \`--json\` for full details (result, duration, URL, etc.).

\`\`\`bash
jenkins --json get-build my-job 42
\`\`\`

### Wait for an externally triggered build to finish

\`\`\`bash
jenkins wait-build my-pr-job 147
\`\`\`

Blocks until the build completes, then exits 0 for SUCCESS or 1 for failure. Useful for monitoring builds triggered by webhooks (e.g., PR builds). Add \`--timeout 300\` to limit wait time to 5 minutes.

### List recent builds for a job

\`\`\`bash
jenkins list-builds my-job
jenkins --json list-builds my-job --limit 25
\`\`\`

### Stream only new console output (skip history)

\`\`\`bash
jenkins tail my-job
\`\`\`

Like \`console --follow\` but starts from the current position. Useful for joining long-running builds mid-stream.

### List all jobs (including those in folders)

\`\`\`bash
jenkins --json list-jobs
\`\`\`

### Get build changelog

\`\`\`bash
jenkins --json list-changes my-job 42
\`\`\`

## Folder-Aware Paths

Jobs inside Jenkins folders use slash-separated paths:

\`\`\`bash
jenkins build my-folder/my-subfolder/my-job --wait
jenkins console my-folder/my-job 5
jenkins get-job my-folder/my-job
\`\`\`

## JSON Output

These commands support \`--json\` for structured output:

- \`list-jobs\`, \`list-plugins\`, \`who-am-i\`, \`version\`, \`get-build\`, \`wait-build\`, \`list-builds\`, \`list-changes\`, \`get-view\`

Streaming commands (\`console\`, \`build --follow\`, \`groovy\`) always output plain text.

## Error Handling

All commands exit with code 0 on success and code 1 on any error. Error messages are printed to stderr. When automating, check the exit code to determine success or failure.

## Command Reference

Read the full command reference for detailed usage of all 33+ commands:
\`\`\`bash
cat ${skillDir}/references/commands.md
\`\`\`
`;
}

export function commandsReference(): string {
  return `# Jenkins CLI \u2014 Command Reference

## System

| Command | Description |
|---|---|
| \`jenkins version\` | Show Jenkins server version |
| \`jenkins who-am-i\` | Show authenticated user and permissions |
| \`jenkins quiet-down\` | Put Jenkins into quiet mode (no new builds) |
| \`jenkins cancel-quiet-down\` | Cancel quiet mode |
| \`jenkins clear-queue\` | Cancel all queued builds |
| \`jenkins reload-configuration\` | Reload configuration from disk |

## Jobs

| Command | Description |
|---|---|
| \`jenkins list-jobs\` | List all jobs (recursively traverses folders) |
| \`jenkins get-job <name>\` | Print job config XML to stdout |
| \`jenkins create-job <name> < config.xml\` | Create job from XML on stdin |
| \`jenkins update-job <name> < config.xml\` | Update job config from XML on stdin |
| \`jenkins copy-job <src> <dest>\` | Copy a job |
| \`jenkins delete-job <name>\` | Delete a job |
| \`jenkins reload-job <name>\` | Reload job config from disk |

## Builds

| Command | Description |
|---|---|
| \`jenkins build <name>\` | Trigger a build |
| \`jenkins build <name> -p key=val\` | Parameterized build (repeatable) |
| \`jenkins build <name> --follow\` | Trigger and stream console output |
| \`jenkins build <name> --wait\` | Trigger and wait for completion |
| \`jenkins get-build <name> [number]\` | Get build status, result, duration (exits 1 if not SUCCESS) |
| \`jenkins wait-build <name> [number]\` | Wait for build to finish, report result (exits 1 if not SUCCESS) |
| \`jenkins wait-build <name> --timeout 300\` | Wait with timeout in seconds |
| \`jenkins list-builds <name> [--limit N]\` | List recent builds (default 10) with status and duration |
| \`jenkins console <name>\` | Print console output (last build) |
| \`jenkins console <name> <number>\` | Print console output for build #N |
| \`jenkins console <name> --follow\` | Stream console output in real-time |
| \`jenkins tail <name> [number]\` | Stream new console output only (skip history) |
| \`jenkins set-build-description <name> <n> "text"\` | Set build description |
| \`jenkins set-build-display-name <name> <n> "text"\` | Set build display name |
| \`jenkins delete-builds <name> 40,41,42\` | Delete builds (comma-separated) |
| \`jenkins list-changes <name> <n>\` | List SCM changes in a build |

## Nodes

| Command | Description |
|---|---|
| \`jenkins create-node <name> < config.xml\` | Create node from XML on stdin |
| \`jenkins update-node <name> < config.xml\` | Update node config from XML on stdin |
| \`jenkins delete-node <name>\` | Delete a node |
| \`jenkins connect-node <name>\` | Initiate node connection |
| \`jenkins disconnect-node <name> [-m "reason"]\` | Disconnect a node |
| \`jenkins online-node <name>\` | Bring node online |
| \`jenkins offline-node <name> [-m "reason"]\` | Take node offline |
| \`jenkins wait-node-online <name> [--timeout 120]\` | Wait for node to come online |
| \`jenkins wait-node-offline <name> [--timeout 120]\` | Wait for node to go offline |

## Views

| Command | Description |
|---|---|
| \`jenkins get-view <name>\` | Print view config XML to stdout |
| \`jenkins create-view <name> < config.xml\` | Create view from XML on stdin |
| \`jenkins update-view <name> < config.xml\` | Update view config from XML on stdin |
| \`jenkins delete-view <name>\` | Delete a view |
| \`jenkins add-job-to-view <view> <job>\` | Add a job to a view |
| \`jenkins remove-job-from-view <view> <job>\` | Remove a job from a view |

## Plugins

| Command | Description |
|---|---|
| \`jenkins list-plugins\` | List installed plugins |
| \`jenkins install-plugin <id[@version]>\` | Install by plugin ID (defaults to @latest) |
| \`jenkins install-plugin <url>\` | Install from URL |
| \`jenkins install-plugin ./file.hpi\` | Install from local file |
| \`jenkins enable-plugin <name>\` | Enable a plugin |
| \`jenkins disable-plugin <name>\` | Disable a plugin |

## Groovy

| Command | Description |
|---|---|
| \`jenkins groovy script.groovy\` | Execute a Groovy script file |
| \`echo 'code' \\| jenkins groovy -\` | Execute Groovy from stdin |
| \`jenkins groovysh\` | Interactive Groovy shell (requires Java) |

## Credential Management

| Command | Description |
|---|---|
| \`jenkins configure set -s <url> -u <user> -t <token>\` | Store credentials |
| \`jenkins configure show -s <url>\` | Show stored credentials |
| \`jenkins configure clear -s <url>\` | Clear stored credentials |
`;
}
