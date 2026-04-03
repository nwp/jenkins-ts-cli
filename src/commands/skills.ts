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

const VALID_AGENTS = Object.keys(AGENT_SKILL_DIRS) as AgentType[];

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

      if (opts.agent && !VALID_AGENTS.includes(opts.agent as AgentType)) {
        throw new Error(`Unknown agent type: ${opts.agent}. Use ${VALID_AGENTS.join(", ")}.`);
      }

      const agents: AgentType[] = opts.agent
        ? [opts.agent as AgentType]
        : detectAgents(cwd);

      if (agents.length === 0) {
        const dirs = VALID_AGENTS.map((a) => agentRootDir(a) + "/").join(", ");
        throw new Error(
          `No AI agent directories detected. Checked for: ${dirs}\n` +
          `Use --agent <${VALID_AGENTS.join("|")}> to install for a specific agent.`,
        );
      }

      const results = agents.map((agent) => installSkill(cwd, agent, opts.force));

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
  return VALID_AGENTS.filter((agent) =>
    existsSync(join(cwd, agentRootDir(agent))),
  );
}

function agentRootDir(agent: AgentType): string {
  return AGENT_SKILL_DIRS[agent].split("/")[0]!;
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
