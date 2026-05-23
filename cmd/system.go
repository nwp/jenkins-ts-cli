package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nwp/jenkins-cli/internal/output"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show Jenkins server version",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			resp, err := c.Get("/api/json")
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			version := resp.Header.Get("X-Jenkins")
			if version == "" {
				// Fall back to parsing JSON
				var info struct {
					Version string `json:"version"`
				}
				json.NewDecoder(resp.Body).Decode(&info)
				version = info.Version
			}
			return output.Print(version, flagJSON)
		},
	}
}

func newWhoAmICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "who-am-i",
		Short: "Show current user info and authorities",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			var info struct {
				ID          string   `json:"id"`
				FullName    string   `json:"fullName"`
				Authorities []string `json:"authorities"`
			}
			if err := c.GetJSON("/me/api/json", &info); err != nil {
				return err
			}
			if flagJSON {
				return output.Print(info, true)
			}
			fmt.Printf("Authenticated as: %s\n", info.ID)
			fmt.Printf("Authorities:\n")
			for _, a := range info.Authorities {
				fmt.Printf("  %s\n", a)
			}
			return nil
		},
	}
}

func newQuietDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "quiet-down",
		Short: "Put Jenkins into quiet mode (no new builds)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			resp, err := c.PostEmpty("/quietDown")
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Println("Jenkins is now in quiet mode.")
			return nil
		},
	}
}

func newCancelQuietDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel-quiet-down",
		Short: "Cancel quiet mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			resp, err := c.PostEmpty("/cancelQuietDown")
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Println("Quiet mode cancelled.")
			return nil
		},
	}
}

func newClearQueueCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear-queue",
		Short: "Clear the build queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			resp, err := c.PostEmpty("/queue/cancelAll")
			if err != nil {
				// Fallback for Jenkins versions that don't support /queue/cancelAll.
				resp, err = c.PostEmpty("/queue/clear")
				if err != nil {
					return err
				}
			}
			drainClose(resp)
			fmt.Println("Build queue cleared.")
			return nil
		},
	}
}

func newReloadConfigurationCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reload-configuration",
		Short: "Reload configuration from disk",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			resp, err := c.PostEmpty("/reload")
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Println("Configuration reloaded.")
			return nil
		},
	}
}

// drainClose reads and closes a response body safely.
func drainClose(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// xmlFromStdin reads XML from stdin (used for create/update operations).
func xmlFromStdin() (string, error) {
	b, err := io.ReadAll(stdinReader())
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}
