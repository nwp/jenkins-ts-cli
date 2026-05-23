package cmd

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/cobra"

	"github.com/nwp/jenkins-cli/internal/console"
)

func newBuildCmd() *cobra.Command {
	var params []string
	var follow, wait bool

	c := &cobra.Command{
		Use:   "build <name>",
		Short: "Trigger a build",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			jobName := args[0]
			jobURL := cl.JobURL(jobName)

			var buildPath string
			if len(params) == 0 {
				buildPath = jobURL + "/build"
				resp, err := cl.PostEmpty(buildPath)
				if err != nil {
					return err
				}
				drainClose(resp)
			} else {
				paramMap, err := splitParams(params)
				if err != nil {
					return err
				}
				form := url.Values{}
				for k, v := range paramMap {
					form.Set(k, v)
				}
				buildPath = jobURL + "/buildWithParameters"
				resp, err := cl.PostForm(buildPath, form)
				if err != nil {
					return err
				}
				drainClose(resp)
			}

			fmt.Printf("Build triggered for %q.\n", jobName)

			if !follow && !wait {
				return nil
			}

			// Poll for the queued build number
			buildNumber, err := pollForBuildNumber(cl, jobURL)
			if err != nil {
				return fmt.Errorf("waiting for build to start: %w", err)
			}
			buildRef := fmt.Sprintf("%d", buildNumber)

			if follow {
				fmt.Printf("Streaming console output for build #%d...\n", buildNumber)
				return console.Follow(cl, jobURL, buildRef, cmd.OutOrStdout())
			}

			// --wait: poll for completion
			return waitForBuild(cl, jobURL, buildRef, 0)
		},
	}

	c.Flags().StringArrayVarP(&params, "param", "p", nil, "build parameter as key=value (repeatable)")
	c.Flags().BoolVar(&follow, "follow", false, "stream console output after triggering")
	c.Flags().BoolVar(&wait, "wait", false, "wait for build to complete")
	return c
}

// pollForBuildNumber waits until a new build appears on the job and returns its number.
func pollForBuildNumber(cl interface {
	GetJSON(string, any) error
	JobURL(string) string
}, jobURL string) (int, error) {
	type buildInfo struct {
		Number int `json:"number"`
	}
	type jobInfo struct {
		LastBuild buildInfo `json:"lastBuild"`
	}

	// Get current last build number before trigger
	var before jobInfo
	_ = cl.GetJSON(jobURL+"/api/json?tree=lastBuild[number]", &before)
	prevNumber := before.LastBuild.Number

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		var after jobInfo
		if err := cl.GetJSON(jobURL+"/api/json?tree=lastBuild[number]", &after); err != nil {
			return 0, err
		}
		if after.LastBuild.Number > prevNumber {
			return after.LastBuild.Number, nil
		}
		time.Sleep(2 * time.Second)
	}
	return 0, fmt.Errorf("timed out waiting for build to start")
}

// waitForBuild polls a build until it is no longer building.
// timeoutSec of 0 means no timeout.
func waitForBuild(cl interface {
	GetJSON(string, any) error
}, jobURL, buildRef string, timeoutSec int) error {
	type buildResult struct {
		Building bool   `json:"building"`
		Result   string `json:"result"`
		Number   int    `json:"number"`
	}

	deadline := time.Time{}
	if timeoutSec > 0 {
		deadline = time.Now().Add(time.Duration(timeoutSec) * time.Second)
	}

	for {
		var b buildResult
		if err := cl.GetJSON(fmt.Sprintf("%s/%s/api/json", jobURL, buildRef), &b); err != nil {
			return err
		}
		if !b.Building {
			fmt.Printf("Build #%d finished: %s\n", b.Number, b.Result)
			if b.Result != "SUCCESS" {
				return fmt.Errorf("build result: %s", b.Result)
			}
			return nil
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for build to complete")
		}
		time.Sleep(3 * time.Second)
	}
}
