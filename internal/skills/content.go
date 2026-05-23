package skills

import "fmt"

// SkillContent returns the SKILL.md content for the given skill directory path.
func SkillContent(skillDir string) string {
	return fmt.Sprintf(`---
name: jenkins
description: Interact with Jenkins CI/CD servers using the jenkins CLI.
triggers:
  - jenkins
  - CI
  - pipeline
  - build
  - job
---

# Jenkins CLI Skill

This skill lets you interact with a Jenkins server using the %[1]s CLI.

## Authentication

Set the following environment variables before running commands:

- `+"`JENKINS_URL`"+` — Jenkins server URL (e.g. https://jenkins.example.com)
- `+"`JENKINS_USER_ID`"+` — Jenkins username
- `+"`JENKINS_API_TOKEN`"+` — Jenkins API token (User > Configure > Add new Token)

Or use the `+"`--auth`"+` flag: `+"`--auth username:token`"+`

Or store credentials interactively: `+"`jenkins configure set -s $JENKINS_URL -u $USER -t $TOKEN`"+`

## Common Workflows

### Trigger a build
`+"```"+`
jenkins build my-job -s $JENKINS_URL
`+"```"+`

### Trigger with parameters
`+"```"+`
jenkins build my-job -p BRANCH=main -p ENV=prod -s $JENKINS_URL
`+"```"+`

### Stream console output
`+"```"+`
jenkins console my-job lastBuild --follow -s $JENKINS_URL
# or trigger and follow in one step:
jenkins build my-job --follow -s $JENKINS_URL
`+"```"+`

### Wait for a build to complete
`+"```"+`
jenkins wait-build my-job -s $JENKINS_URL
jenkins build my-job --wait -s $JENKINS_URL
`+"```"+`

### List jobs (including nested folders)
`+"```"+`
jenkins jobs list-jobs -s $JENKINS_URL
`+"```"+`

### Nested folder jobs — use slash-separated paths
`+"```"+`
jenkins build "folder/sub-folder/my-job" -s $JENKINS_URL
jenkins console "folder/my-job" lastBuild -s $JENKINS_URL
`+"```"+`

### JSON output for scripting
`+"```"+`
jenkins builds get-build my-job --json -s $JENKINS_URL
jenkins plugins list-plugins --json -s $JENKINS_URL
`+"```"+`

### Execute Groovy script
`+"```"+`
jenkins groovy myscript.groovy -s $JENKINS_URL
echo 'println "hello"' | jenkins groovy - -s $JENKINS_URL
`+"```"+`

## Full Command Reference

See [commands.md](%[1]s/references/commands.md) for the full command reference.
`, skillDir)
}

// CommandsReference returns the full commands reference markdown table.
func CommandsReference() string {
	return `# Jenkins CLI — Command Reference

## Global Flags

| Flag | Description |
|------|-------------|
| ` + "`-s <url>`" + ` | Jenkins server URL (or ` + "`JENKINS_URL`" + ` env var) |
| ` + "`--auth <user:token>`" + ` | Credentials (username:apitoken) |
| ` + "`--json`" + ` | Output as JSON where supported |

## System

| Command | Description |
|---------|-------------|
| ` + "`jenkins version`" + ` | Show Jenkins server version |
| ` + "`jenkins who-am-i`" + ` | Show current user and authorities |
| ` + "`jenkins quiet-down`" + ` | Enter quiet mode (no new builds) |
| ` + "`jenkins cancel-quiet-down`" + ` | Cancel quiet mode |
| ` + "`jenkins clear-queue`" + ` | Clear the build queue |
| ` + "`jenkins reload-configuration`" + ` | Reload config from disk |

## Credentials

| Command | Description |
|---------|-------------|
| ` + "`jenkins configure set -s <url> -u <user> -t <token>`" + ` | Store credentials |
| ` + "`jenkins configure show -s <url>`" + ` | Show stored credentials |
| ` + "`jenkins configure clear -s <url>`" + ` | Remove stored credentials |

## Jobs

| Command | Description |
|---------|-------------|
| ` + "`jenkins jobs list-jobs`" + ` | List all jobs (recursive) |
| ` + "`jenkins jobs get-job <name>`" + ` | Get job config XML |
| ` + "`jenkins jobs create-job <name>`" + ` | Create job from XML on stdin |
| ` + "`jenkins jobs copy-job <src> <dest>`" + ` | Copy a job |
| ` + "`jenkins jobs delete-job <name>`" + ` | Delete a job |
| ` + "`jenkins jobs update-job <name>`" + ` | Update job from XML on stdin |
| ` + "`jenkins jobs reload-job <name>`" + ` | Reload job config from disk |

## Builds

| Command | Description |
|---------|-------------|
| ` + "`jenkins build <name> [-p key=val]`" + ` | Trigger a build |
| ` + "`jenkins build <name> --follow`" + ` | Trigger and stream console |
| ` + "`jenkins build <name> --wait`" + ` | Trigger and wait for completion |
| ` + "`jenkins builds console <name> [build]`" + ` | Get console output |
| ` + "`jenkins builds console <name> [build] --follow`" + ` | Stream console |
| ` + "`jenkins builds get-build <name> [build]`" + ` | Get build details |
| ` + "`jenkins builds wait-build <name> [build]`" + ` | Wait for completion |
| ` + "`jenkins builds list-builds <name> [--limit N]`" + ` | List recent builds |
| ` + "`jenkins builds tail <name> [build]`" + ` | Tail new console output |
| ` + "`jenkins builds set-build-description <name> <build> <desc>`" + ` | Set description |
| ` + "`jenkins builds set-build-display-name <name> <build> <name>`" + ` | Set display name |
| ` + "`jenkins builds delete-builds <name> <nums>`" + ` | Delete builds (comma-separated) |
| ` + "`jenkins builds list-changes <name> [build]`" + ` | List build changes |

## Nodes

| Command | Description |
|---------|-------------|
| ` + "`jenkins nodes create-node <name>`" + ` | Create node from XML on stdin |
| ` + "`jenkins nodes delete-node <name>`" + ` | Delete a node |
| ` + "`jenkins nodes update-node <name>`" + ` | Update node from XML on stdin |
| ` + "`jenkins nodes connect-node <name>`" + ` | Connect a node |
| ` + "`jenkins nodes disconnect-node <name> [-m msg]`" + ` | Disconnect a node |
| ` + "`jenkins nodes online-node <name>`" + ` | Bring node online |
| ` + "`jenkins nodes offline-node <name> [-m msg]`" + ` | Take node offline |
| ` + "`jenkins nodes wait-node-online <name>`" + ` | Wait for node to come online |
| ` + "`jenkins nodes wait-node-offline <name>`" + ` | Wait for node to go offline |

## Views

| Command | Description |
|---------|-------------|
| ` + "`jenkins views get-view <name>`" + ` | Get view config XML |
| ` + "`jenkins views create-view <name>`" + ` | Create view from XML on stdin |
| ` + "`jenkins views delete-view <name>`" + ` | Delete a view |
| ` + "`jenkins views update-view <name>`" + ` | Update view from XML on stdin |
| ` + "`jenkins views add-job-to-view <view> <job>`" + ` | Add job to view |
| ` + "`jenkins views remove-job-from-view <view> <job>`" + ` | Remove job from view |

## Plugins

| Command | Description |
|---------|-------------|
| ` + "`jenkins plugins list-plugins`" + ` | List installed plugins |
| ` + "`jenkins plugins install-plugin <id[@ver]|url|file>`" + ` | Install a plugin |
| ` + "`jenkins plugins enable-plugin <name>`" + ` | Enable a plugin |
| ` + "`jenkins plugins disable-plugin <name>`" + ` | Disable a plugin |

## Groovy

| Command | Description |
|---------|-------------|
| ` + "`jenkins groovy [file|-]`" + ` | Execute Groovy via Script Console |
`
}
