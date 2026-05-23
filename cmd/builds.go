package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/nwp/jenkins-cli/internal/console"
	"github.com/nwp/jenkins-cli/internal/output"
)

func newBuildsCmd() *cobra.Command {
	builds := &cobra.Command{
		Use:   "builds",
		Short: "Build management commands",
	}
	builds.AddCommand(
		newConsoleCmd(),
		newGetBuildCmd(),
		newWaitBuildCmd(),
		newListBuildsCmd(),
		newTailCmd(),
		newSetBuildDescriptionCmd(),
		newSetBuildDisplayNameCmd(),
		newDeleteBuildsCmd(),
		newListChangesCmd(),
	)
	return builds
}

func buildRef(args []string, idx int) string {
	if idx < len(args) && args[idx] != "" {
		return args[idx]
	}
	return "lastBuild"
}

func newConsoleCmd() *cobra.Command {
	var follow bool
	c := &cobra.Command{
		Use:   "console <name> [build]",
		Short: "Get console output for a build",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			jobURL := cl.JobURL(args[0])
			ref := buildRef(args, 1)
			if follow {
				return console.Follow(cl, jobURL, ref, os.Stdout)
			}
			text, err := console.GetText(cl, jobURL, ref)
			if err != nil {
				return err
			}
			fmt.Print(text)
			return nil
		},
	}
	c.Flags().BoolVar(&follow, "follow", false, "stream console output in real time")
	return c
}

func newGetBuildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-build <name> [build]",
		Short: "Get build details",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			var b struct {
				Number      int    `json:"number"`
				Result      string `json:"result"`
				Building    bool   `json:"building"`
				Duration    int64  `json:"duration"`
				Timestamp   int64  `json:"timestamp"`
				DisplayName string `json:"displayName"`
			}
			path := fmt.Sprintf("%s/%s/api/json", cl.JobURL(args[0]), buildRef(args, 1))
			if err := cl.GetJSON(path, &b); err != nil {
				return err
			}
			if flagJSON {
				return output.Print(b, true)
			}
			status := b.Result
			if b.Building {
				status = "BUILDING"
			}
			ts := time.Unix(b.Timestamp/1000, 0).Local().Format(time.RFC3339)
			fmt.Printf("#%d  %s  %s  %s\n", b.Number, b.DisplayName, status, ts)
			fmt.Printf("Duration: %s\n", formatDuration(b.Duration))
			return nil
		},
	}
}

func newWaitBuildCmd() *cobra.Command {
	var timeoutSec int
	c := &cobra.Command{
		Use:   "wait-build <name> [build]",
		Short: "Wait for a build to complete",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			return waitForBuild(cl, cl.JobURL(args[0]), buildRef(args, 1), timeoutSec)
		},
	}
	c.Flags().IntVar(&timeoutSec, "timeout", 0, "timeout in seconds (0 = no timeout)")
	return c
}

func newListBuildsCmd() *cobra.Command {
	var limit int
	c := &cobra.Command{
		Use:   "list-builds <name>",
		Short: "List recent builds for a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			var result struct {
				Builds []struct {
					Number      int    `json:"number"`
					Result      string `json:"result"`
					Building    bool   `json:"building"`
					Duration    int64  `json:"duration"`
					Timestamp   int64  `json:"timestamp"`
					DisplayName string `json:"displayName"`
				} `json:"builds"`
			}
			tree := fmt.Sprintf("builds[number,result,building,duration,timestamp,displayName]{0,%d}", limit)
			path := cl.JobURL(args[0]) + "/api/json?tree=" + url.QueryEscape(tree)
			if err := cl.GetJSON(path, &result); err != nil {
				return err
			}
			if flagJSON {
				return output.Print(result.Builds, true)
			}
			for _, b := range result.Builds {
				status := b.Result
				if b.Building {
					status = "BUILDING"
				}
				ts := time.Unix(b.Timestamp/1000, 0).Local().Format("2006-01-02 15:04:05")
				fmt.Printf("#%d  %-10s  %s  %s\n", b.Number, status, formatDuration(b.Duration), ts)
			}
			return nil
		},
	}
	c.Flags().IntVar(&limit, "limit", 10, "maximum number of builds to show")
	return c
}

func newTailCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tail <name> [build]",
		Short: "Stream new console output only (from current end position)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			return console.Tail(cl, cl.JobURL(args[0]), buildRef(args, 1), os.Stdout)
		},
	}
}

func newSetBuildDescriptionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-build-description <name> <build> <description>",
		Short: "Set the description of a build",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			path := fmt.Sprintf("%s/%s/submitDescription", cl.JobURL(args[0]), args[1])
			resp, err := cl.PostForm(path, url.Values{"description": {args[2]}})
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Println("Description updated.")
			return nil
		},
	}
}

func newSetBuildDisplayNameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-build-display-name <name> <build> <displayName>",
		Short: "Set the display name of a build",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			path := fmt.Sprintf("%s/%s/configSubmit", cl.JobURL(args[0]), args[1])
			resp, err := cl.PostForm(path, url.Values{"displayName": {args[2]}})
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Println("Display name updated.")
			return nil
		},
	}
}

func newDeleteBuildsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-builds <name> <builds>",
		Short: "Delete builds (comma-separated build numbers)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			jobURL := cl.JobURL(args[0])
			for _, num := range strings.Split(args[1], ",") {
				num = strings.TrimSpace(num)
				path := fmt.Sprintf("%s/%s/doDelete", jobURL, num)
				resp, err := cl.PostEmpty(path)
				if err != nil {
					return fmt.Errorf("delete build %s: %w", num, err)
				}
				drainClose(resp)
				fmt.Printf("Build #%s deleted.\n", num)
			}
			return nil
		},
	}
}

func newListChangesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-changes <name> [build]",
		Short: "List changes/commits in a build",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			var result struct {
				ChangeSets []struct {
					Items []struct {
						CommitID string `json:"commitId"`
						Msg      string `json:"msg"`
						Author   struct {
							FullName string `json:"fullName"`
						} `json:"author"`
					} `json:"items"`
				} `json:"changeSets"`
			}
			path := fmt.Sprintf("%s/%s/api/json?tree=changeSets[items[commitId,msg,author[fullName]]]",
				cl.JobURL(args[0]), buildRef(args, 1))
			if err := cl.GetJSON(path, &result); err != nil {
				return err
			}
			if flagJSON {
				return output.Print(result.ChangeSets, true)
			}
			for _, cs := range result.ChangeSets {
				for _, item := range cs.Items {
					id := item.CommitID
					if len(id) > 8 {
						id = id[:8]
					}
					fmt.Printf("%s  %s  %s\n", id, item.Author.FullName, item.Msg)
				}
			}
			return nil
		},
	}
}

// formatDuration formats a Jenkins build duration (milliseconds) as a human-readable string.
func formatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	secs := ms / 1000
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	mins := secs / 60
	secs = secs % 60
	if mins < 60 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	hours := mins / 60
	mins = mins % 60
	return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
}
