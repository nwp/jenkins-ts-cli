package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nwp/jenkins-cli/internal/keyring"
	"github.com/nwp/jenkins-cli/internal/output"
	"github.com/nwp/jenkins-cli/internal/paths"
)

func newConfigureCmd() *cobra.Command {
	cfg := &cobra.Command{
		Use:   "configure",
		Short: "Manage stored Jenkins credentials",
	}
	cfg.AddCommand(newConfigureSetCmd(), newConfigureShowCmd(), newConfigureClearCmd())
	return cfg
}

func newConfigureSetCmd() *cobra.Command {
	var serverURL, username, token string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Store credentials for a Jenkins server",
		RunE: func(cmd *cobra.Command, args []string) error {
			normalized, err := paths.NormalizeURL(serverURL)
			if err != nil {
				return err
			}
			if err := keyring.StoreCredentials(normalized, username, token); err != nil {
				return fmt.Errorf("store credentials: %w", err)
			}
			fmt.Printf("Credentials stored for %s\n", normalized)
			return nil
		},
	}
	cmd.Flags().StringVarP(&serverURL, "server", "s", "", "Jenkins server URL")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Jenkins username")
	cmd.Flags().StringVarP(&token, "token", "t", "", "Jenkins API token")
	cmd.MarkFlagRequired("server")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("token")
	return cmd
}

func newConfigureShowCmd() *cobra.Command {
	var serverURL string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show stored credentials for a Jenkins server",
		RunE: func(cmd *cobra.Command, args []string) error {
			normalized, err := paths.NormalizeURL(serverURL)
			if err != nil {
				return err
			}
			username, err := keyring.GetStoredUsername(normalized)
			if err != nil {
				return err
			}
			if username == "" {
				fmt.Println("No credentials stored for that server.")
				return nil
			}
			token, err := keyring.GetStoredToken(normalized, username)
			if err != nil {
				return err
			}
			tokenDisplay := "********"
			if token == "" {
				tokenDisplay = "(not found in keyring)"
			}
			type credDisplay struct {
				Username string `json:"username"`
				Token    string `json:"token"`
			}
			if flagJSON {
				return output.Print(credDisplay{Username: username, Token: tokenDisplay}, true)
			}
			fmt.Printf("Username: %s\n", username)
			fmt.Printf("Token:    %s\n", tokenDisplay)
			return nil
		},
	}
	cmd.Flags().StringVarP(&serverURL, "server", "s", "", "Jenkins server URL")
	cmd.MarkFlagRequired("server")
	return cmd
}

func newConfigureClearCmd() *cobra.Command {
	var serverURL string
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Remove stored credentials for a Jenkins server",
		RunE: func(cmd *cobra.Command, args []string) error {
			normalized, err := paths.NormalizeURL(serverURL)
			if err != nil {
				return err
			}
			if err := keyring.DeleteCredentials(normalized); err != nil {
				return fmt.Errorf("delete credentials: %w", err)
			}
			fmt.Printf("Credentials cleared for %s\n", normalized)
			return nil
		},
	}
	cmd.Flags().StringVarP(&serverURL, "server", "s", "", "Jenkins server URL")
	cmd.MarkFlagRequired("server")
	return cmd
}
