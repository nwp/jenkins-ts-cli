package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nwp/jenkins-cli/internal/auth"
	"github.com/nwp/jenkins-cli/internal/client"
	"github.com/nwp/jenkins-cli/internal/paths"
)

var (
	flagServer string
	flagAuth   string
	flagJSON   bool
)

var rootCmd = &cobra.Command{
	Use:   "jenkins",
	Short: "Jenkins CLI",
	Long:  "A command-line interface for Jenkins.",
	// Silence the default error/usage printing so we can format errors ourselves.
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root command.
func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
	return err
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagServer, "server", "s", "", "Jenkins server URL (or JENKINS_URL env var)")
	rootCmd.PersistentFlags().StringVar(&flagAuth, "auth", "", "username:token credentials")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "output as JSON")

	// Register all sub-commands
	rootCmd.AddCommand(
		newVersionCmd(),
		newWhoAmICmd(),
		newQuietDownCmd(),
		newCancelQuietDownCmd(),
		newClearQueueCmd(),
		newReloadConfigurationCmd(),
		newConfigureCmd(),
		newJobsCmd(),
		newBuildCmd(),
		newBuildsCmd(),
		newNodesCmd(),
		newViewsCmd(),
		newPluginsCmd(),
		newGroovyCmd(),
		newInstallSkillCmd(),
	)
}

// resolveClient normalizes the server URL, resolves credentials, and returns
// a ready-to-use Jenkins client. Commands call this at the start of their RunE.
func resolveClient() (*client.Client, error) {
	serverURL := flagServer
	if serverURL == "" {
		serverURL = os.Getenv("JENKINS_URL")
	}
	if serverURL == "" {
		return nil, errors.New("server URL is required: use -s <url> or set JENKINS_URL")
	}

	normalized, err := paths.NormalizeURL(serverURL)
	if err != nil {
		return nil, err
	}

	creds, err := auth.ResolveCredentials(auth.ResolveOptions{
		AuthFlag:  flagAuth,
		ServerURL: normalized,
	})
	if err != nil {
		return nil, err
	}

	return client.New(normalized, creds), nil
}

// normalizeArgs converts Java CLI-style single-dash long flags (e.g. -auth, -server)
// to double-dash equivalents so users familiar with the Jenkins Java CLI feel at home.
func NormalizeArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		switch a {
		case "-auth":
			out = append(out, "--auth")
		case "-server":
			out = append(out, "--server")
		case "-s":
			out = append(out, "-s")
		default:
			out = append(out, a)
		}
	}
	return out
}
