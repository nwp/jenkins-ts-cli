package cmd

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/cobra"
)

func newNodesCmd() *cobra.Command {
	nodes := &cobra.Command{
		Use:   "nodes",
		Short: "Node/agent management commands",
	}
	nodes.AddCommand(
		newCreateNodeCmd(),
		newDeleteNodeCmd(),
		newUpdateNodeCmd(),
		newConnectNodeCmd(),
		newDisconnectNodeCmd(),
		newOnlineNodeCmd(),
		newOfflineNodeCmd(),
		newWaitNodeOnlineCmd(),
		newWaitNodeOfflineCmd(),
	)
	return nodes
}

func newCreateNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-node <name>",
		Short: "Create a new node from XML on stdin",
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
			path := "/computer/doCreateItem?name=" + url.QueryEscape(args[0])
			resp, err := cl.PostXML(path, xml)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Node %q created.\n", args[0])
			return nil
		},
	}
}

func newDeleteNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-node <name>",
		Short: "Delete a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			resp, err := cl.PostEmpty(cl.NodeURL(args[0]) + "/doDelete")
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Node %q deleted.\n", args[0])
			return nil
		},
	}
}

func newUpdateNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update-node <name>",
		Short: "Update node configuration from XML on stdin",
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
			resp, err := cl.PostXML(cl.NodeURL(args[0])+"/config.xml", xml)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Node %q updated.\n", args[0])
			return nil
		},
	}
}

func newConnectNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connect-node <name>",
		Short: "Connect a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			resp, err := cl.PostEmpty(cl.NodeURL(args[0]) + "/launchSlaveAgent")
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Node %q connect initiated.\n", args[0])
			return nil
		},
	}
}

func newDisconnectNodeCmd() *cobra.Command {
	var message string
	c := &cobra.Command{
		Use:   "disconnect-node <name>",
		Short: "Disconnect a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			params := url.Values{}
			if message != "" {
				params.Set("offlineMessage", message)
			}
			resp, err := cl.PostForm(cl.NodeURL(args[0])+"/doDisconnect", params)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Node %q disconnected.\n", args[0])
			return nil
		},
	}
	c.Flags().StringVarP(&message, "message", "m", "", "offline message")
	return c
}

func newOnlineNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "online-node <name>",
		Short: "Bring a node online",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			resp, err := cl.PostEmpty(cl.NodeURL(args[0]) + "/toggleOffline?offlineMessage=")
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Node %q brought online.\n", args[0])
			return nil
		},
	}
}

func newOfflineNodeCmd() *cobra.Command {
	var message string
	c := &cobra.Command{
		Use:   "offline-node <name>",
		Short: "Take a node offline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			params := url.Values{"offlineMessage": {message}}
			resp, err := cl.PostForm(cl.NodeURL(args[0])+"/toggleOffline", params)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Node %q taken offline.\n", args[0])
			return nil
		},
	}
	c.Flags().StringVarP(&message, "message", "m", "", "offline message")
	return c
}

func newWaitNodeOnlineCmd() *cobra.Command {
	var timeoutSec int
	c := &cobra.Command{
		Use:   "wait-node-online <name>",
		Short: "Wait for a node to come online",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			return waitNodeState(cl, args[0], false, timeoutSec)
		},
	}
	c.Flags().IntVar(&timeoutSec, "timeout", 60, "timeout in seconds")
	return c
}

func newWaitNodeOfflineCmd() *cobra.Command {
	var timeoutSec int
	c := &cobra.Command{
		Use:   "wait-node-offline <name>",
		Short: "Wait for a node to go offline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			return waitNodeState(cl, args[0], true, timeoutSec)
		},
	}
	c.Flags().IntVar(&timeoutSec, "timeout", 60, "timeout in seconds")
	return c
}

func waitNodeState(cl interface {
	GetJSON(string, any) error
	NodeURL(string) string
}, name string, wantOffline bool, timeoutSec int) error {
	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	for {
		var info struct {
			Offline bool `json:"offline"`
		}
		if err := cl.GetJSON(cl.NodeURL(name)+"/api/json?tree=offline", &info); err != nil {
			return err
		}
		if info.Offline == wantOffline {
			state := "online"
			if wantOffline {
				state = "offline"
			}
			fmt.Printf("Node %q is now %s.\n", name, state)
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for node %q", name)
		}
		time.Sleep(3 * time.Second)
	}
}
