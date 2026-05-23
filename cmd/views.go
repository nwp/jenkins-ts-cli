package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

func newViewsCmd() *cobra.Command {
	views := &cobra.Command{
		Use:   "views",
		Short: "View management commands",
	}
	views.AddCommand(
		newGetViewCmd(),
		newCreateViewCmd(),
		newDeleteViewCmd(),
		newUpdateViewCmd(),
		newAddJobToViewCmd(),
		newRemoveJobFromViewCmd(),
	)
	return views
}

func newGetViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-view <name>",
		Short: "Get view configuration XML",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			xml, err := cl.GetText(cl.ViewURL(args[0]) + "/config.xml")
			if err != nil {
				return err
			}
			fmt.Print(xml)
			return nil
		},
	}
}

func newCreateViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-view <name>",
		Short: "Create a new view from XML on stdin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			xml, err := xmlFromStdin()
			if err != nil {
				return err
			}
			path := "/createView?name=" + url.QueryEscape(args[0])
			resp, err := cl.PostXML(path, xml)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("View %q created.\n", args[0])
			return nil
		},
	}
}

func newDeleteViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-view <name>",
		Short: "Delete a view",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			resp, err := cl.PostEmpty(cl.ViewURL(args[0]) + "/doDelete")
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("View %q deleted.\n", args[0])
			return nil
		},
	}
}

func newUpdateViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update-view <name>",
		Short: "Update view configuration from XML on stdin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			xml, err := xmlFromStdin()
			if err != nil {
				return err
			}
			resp, err := cl.PostXML(cl.ViewURL(args[0])+"/config.xml", xml)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("View %q updated.\n", args[0])
			return nil
		},
	}
}

func newAddJobToViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-job-to-view <view> <job>",
		Short: "Add a job to a view",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			path := cl.ViewURL(args[0]) + "/addJobToView?name=" + url.QueryEscape(args[1])
			resp, err := cl.PostEmpty(path)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Job %q added to view %q.\n", args[1], args[0])
			return nil
		},
	}
}

func newRemoveJobFromViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-job-from-view <view> <job>",
		Short: "Remove a job from a view",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			path := cl.ViewURL(args[0]) + "/removeJobFromView?name=" + url.QueryEscape(args[1])
			resp, err := cl.PostEmpty(path)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Job %q removed from view %q.\n", args[1], args[0])
			return nil
		},
	}
}
