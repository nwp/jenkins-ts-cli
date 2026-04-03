import { existsSync, mkdirSync, writeFileSync } from "fs";
import { join } from "path";
import type { Command } from "commander";
import { skillContent, commandsReference } from "../skills/content.ts";

type AgentType = "claude" | "copilot" | "codex";

const AGENT_SKILL_DIRS: Record<AgentType, string> = {
  claude: ".claude/skills/jenkins",
  copilot: ".github/skills/jenkins",
  codex: ".codex/skills/jenkins",
};

const AGENT_DETECT_DIRS: Record<AgentType, string> = {
  claude: ".claude",
  copilot: ".github",
  codex: ".codex",
};

interface InstallResult {
  agent: AgentType;
  path: string;
  action: "installed" | "skipped" | "updated";
}

export function registerSkillCommands(program: Command): void {
  program
    .command("install-skill")
    .description("Install AI agent skill for the Jenkins CLI into the current directory")
    .option("--force", "overwrite existing skill files", false)
    .option("--agent <type>", "target a specific agent (claude, copilot, codex) instead of auto-detecting")
    .action(async (opts: { force: boolean; agent?: string }) => {
      const cwd = process.cwd();

      if (opts.agent && !["claude", "copilot", "codex"].includes(opts.agent)) {
        console.error(`Unknown agent type: ${opts.agent}. Use claude, copilot, or codex.`);
        process.exit(1);
      }

      const agents: AgentType[] = opts.agent
        ? [opts.agent as AgentType]
        : detectAgents(cwd);

      if (agents.length === 0) {
        console.error("No AI agent directories detected in the current directory.");
        console.error("Checked for: .claude/, .github/, .codex/");
        console.error("Use --agent <claude|copilot|codex> to install for a specific agent.");
        process.exit(1);
      }

      const results: InstallResult[] = [];
      for (const agent of agents) {
        results.push(installSkill(cwd, agent, opts.force));
      }

      for (const r of results) {
        const label = r.agent.charAt(0).toUpperCase() + r.agent.slice(1);
        switch (r.action) {
          case "installed":
            console.log(`${label}: installed → ${r.path}`);
            break;
          case "updated":
            console.log(`${label}: updated → ${r.path}`);
            break;
          case "skipped":
            console.log(`${label}: already exists → ${r.path} (use --force to overwrite)`);
            break;
        }
      }
    });
}

function detectAgents(cwd: string): AgentType[] {
  const agents: AgentType[] = [];
  for (const [agent, dir] of Object.entries(AGENT_DETECT_DIRS) as [AgentType, string][]) {
    if (existsSync(join(cwd, dir))) agents.push(agent);
  }
  return agents;
}

function installSkill(cwd: string, agent: AgentType, force: boolean): InstallResult {
  const skillDir = join(cwd, AGENT_SKILL_DIRS[agent]);
  const skillPath = join(skillDir, "SKILL.md");
  const refsDir = join(skillDir, "references");
  const refsPath = join(refsDir, "commands.md");
  const existed = existsSync(skillPath);

  if (existed && !force) {
    return { agent, path: skillPath, action: "skipped" };
  }

  mkdirSync(refsDir, { recursive: true });
  writeFileSync(skillPath, skillContent(AGENT_SKILL_DIRS[agent]));
  writeFileSync(refsPath, commandsReference());

  return { agent, path: skillPath, action: existed ? "updated" : "installed" };
}
