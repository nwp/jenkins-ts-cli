package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/nwp/jenkins-cli/internal/skills"
)

type agentConfig struct {
	name string
	dir  string
}

var supportedAgents = []agentConfig{
	{name: "claude", dir: ".claude"},
	{name: "copilot", dir: ".github"},
	{name: "codex", dir: ".codex"},
}

func newInstallSkillCmd() *cobra.Command {
	var force bool
	var agentType string

	c := &cobra.Command{
		Use:   "install-skill",
		Short: "Install the Jenkins CLI AI agent skill into the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			var targets []agentConfig
			if agentType != "" {
				for _, a := range supportedAgents {
					if a.name == agentType {
						targets = []agentConfig{a}
						break
					}
				}
				if len(targets) == 0 {
					return fmt.Errorf("unknown agent type %q (supported: claude, copilot, codex)", agentType)
				}
			} else {
				// Auto-detect
				for _, a := range supportedAgents {
					if _, err := os.Stat(filepath.Join(cwd, a.dir)); err == nil {
						targets = append(targets, a)
					}
				}
				if len(targets) == 0 {
					return fmt.Errorf("no supported AI agent directories found (.claude/, .github/, .codex/); use --agent to specify one")
				}
			}

			for _, a := range targets {
				status, err := installSkillFor(cwd, a, force)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: install for %s: %v\n", a.name, err)
					continue
				}
				fmt.Printf("[%s] %s\n", a.name, status)
			}
			return nil
		},
	}

	c.Flags().BoolVar(&force, "force", false, "overwrite existing skill files")
	c.Flags().StringVar(&agentType, "agent", "", "target agent (claude, copilot, codex)")
	return c
}

func installSkillFor(cwd string, a agentConfig, force bool) (string, error) {
	skillDir := filepath.Join(cwd, a.dir, "skills", "jenkins")
	refDir := filepath.Join(skillDir, "references")

	if err := os.MkdirAll(refDir, 0o755); err != nil {
		return "", fmt.Errorf("create skill directory: %w", err)
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	refFile := filepath.Join(refDir, "commands.md")

	skillStatus := "installed"
	if _, err := os.Stat(skillFile); err == nil {
		if !force {
			skillStatus = "skipped (already exists)"
		} else {
			skillStatus = "updated"
		}
	}

	if skillStatus != "skipped (already exists)" {
		if err := os.WriteFile(skillFile, []byte(skills.SkillContent(skillDir)), 0o644); err != nil {
			return "", fmt.Errorf("write SKILL.md: %w", err)
		}
		if err := os.WriteFile(refFile, []byte(skills.CommandsReference()), 0o644); err != nil {
			return "", fmt.Errorf("write commands.md: %w", err)
		}
	}

	return skillStatus, nil
}
