package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nwp/jenkins-cli/internal/output"
)

func newJobsCmd() *cobra.Command {
	jobs := &cobra.Command{
		Use:   "jobs",
		Short: "Job management commands",
	}
	jobs.AddCommand(
		newListJobsCmd(),
		newGetJobCmd(),
		newCreateJobCmd(),
		newCopyJobCmd(),
		newDeleteJobCmd(),
		newUpdateJobCmd(),
		newReloadJobCmd(),
	)
	return jobs
}

// flatJobs recursively flattens the Jenkins job tree into a slice of full paths.
type jobItem struct {
	Name string    `json:"name"`
	URL  string    `json:"url"`
	Jobs []jobItem `json:"jobs"`
}

func collectJobNames(jobs []jobItem, prefix string) []string {
	var result []string
	for _, j := range jobs {
		fullName := j.Name
		if prefix != "" {
			fullName = prefix + "/" + j.Name
		}
		if len(j.Jobs) > 0 {
			result = append(result, collectJobNames(j.Jobs, fullName)...)
		} else {
			result = append(result, fullName)
		}
	}
	return result
}

func newListJobsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-jobs",
		Short: "List all jobs (recursive, including nested folders)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			var result struct {
				Jobs []jobItem `json:"jobs"`
			}
			tree := "jobs[name,url,jobs[name,url,jobs[name,url,jobs[name,url,jobs[name,url]]]]]"
			if err := c.GetJSON("/api/json?tree="+url.QueryEscape(tree), &result); err != nil {
				return err
			}
			names := collectJobNames(result.Jobs, "")
			return output.Print(names, flagJSON)
		},
	}
}

func newGetJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-job <name>",
		Short: "Get job configuration XML",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			xml, err := c.GetText(c.JobURL(args[0]) + "/config.xml")
			if err != nil {
				return err
			}
			fmt.Print(xml)
			return nil
		},
	}
}

func newCreateJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-job <name>",
		Short: "Create a new job from XML on stdin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			xml, err := xmlFromStdin()
			if err != nil {
				return err
			}
			path := "/createItem?name=" + url.QueryEscape(args[0])
			resp, err := c.PostXML(path, xml)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Job %q created.\n", args[0])
			return nil
		},
	}
}

func newCopyJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "copy-job <src> <dest>",
		Short: "Copy an existing job",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/createItem?name=%s&mode=copy&from=%s",
				url.QueryEscape(args[1]), url.QueryEscape(args[0]))
			resp, err := c.PostEmpty(path)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Job %q copied to %q.\n", args[0], args[1])
			return nil
		},
	}
}

func newDeleteJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-job <name>",
		Short: "Delete a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			resp, err := c.PostEmpty(c.JobURL(args[0]) + "/doDelete")
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Job %q deleted.\n", args[0])
			return nil
		},
	}
}

func newUpdateJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update-job <name>",
		Short: "Update job configuration from XML on stdin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			xml, err := xmlFromStdin()
			if err != nil {
				return err
			}
			resp, err := c.PostXML(c.JobURL(args[0])+"/config.xml", xml)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Job %q updated.\n", args[0])
			return nil
		},
	}
}

func newReloadJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reload-job <name>",
		Short: "Reload job configuration from disk",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient()
			if err != nil {
				return err
			}
			resp, err := c.PostEmpty(c.JobURL(args[0]) + "/reload")
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Job %q reloaded.\n", args[0])
			return nil
		},
	}
}

// splitParams parses "key=value" strings into a map.
func splitParams(params []string) (map[string]string, error) {
	m := make(map[string]string, len(params))
	for _, p := range params {
		idx := strings.IndexByte(p, '=')
		if idx < 0 {
			return nil, fmt.Errorf("invalid parameter %q: expected key=value", p)
		}
		m[p[:idx]] = p[idx+1:]
	}
	return m, nil
}
