package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

func newGroovyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "groovy [script]",
		Short: "Execute a Groovy script via the Jenkins Script Console",
		Long: `Execute a Groovy script on the Jenkins Script Console.

Provide a file path to execute that file, '-' or no argument to read from stdin.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}

			var script string
			if len(args) == 0 || args[0] == "-" {
				script, err = xmlFromStdin()
				if err != nil {
					return err
				}
			} else {
				b, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("read script file: %w", err)
				}
				script = string(b)
			}

			resp, err := cl.PostForm("/scriptText", url.Values{"script": {script}})
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("read script output: %w", err)
			}
			fmt.Print(string(body))
			return nil
		},
	}
}
