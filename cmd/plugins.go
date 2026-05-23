package cmd

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nwp/jenkins-cli/internal/output"
)

func newPluginsCmd() *cobra.Command {
	plugins := &cobra.Command{
		Use:   "plugins",
		Short: "Plugin management commands",
	}
	plugins.AddCommand(
		newListPluginsCmd(),
		newInstallPluginCmd(),
		newEnablePluginCmd(),
		newDisablePluginCmd(),
	)
	return plugins
}

func newListPluginsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-plugins",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			var result struct {
				Plugins []struct {
					ShortName string `json:"shortName"`
					Version   string `json:"version"`
					Active    bool   `json:"active"`
					Enabled   bool   `json:"enabled"`
				} `json:"plugins"`
			}
			if err := cl.GetJSON("/pluginManager/api/json?depth=1", &result); err != nil {
				return err
			}
			if flagJSON {
				return output.Print(result.Plugins, true)
			}
			for _, p := range result.Plugins {
				status := "active"
				if !p.Active {
					status = "inactive"
				}
				if !p.Enabled {
					status = "disabled"
				}
				fmt.Printf("%-40s  %-15s  %s\n", p.ShortName, p.Version, status)
			}
			return nil
		},
	}
}

func newInstallPluginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install-plugin <plugin>",
		Short: "Install a plugin (id[@version], URL, or path to .hpi/.jpi file)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			plugin := args[0]

			// Local file upload (.hpi or .jpi)
			if strings.HasSuffix(plugin, ".hpi") || strings.HasSuffix(plugin, ".jpi") {
				return uploadPlugin(cl, plugin)
			}

			// URL install
			if strings.HasPrefix(plugin, "http://") || strings.HasPrefix(plugin, "https://") {
				xml := fmt.Sprintf(`<jenkins><install plugin="%s" /></jenkins>`, plugin)
				resp, err := cl.PostXML("/pluginManager/installNecessaryPlugins", xml)
				if err != nil {
					return err
				}
				drainClose(resp)
				fmt.Printf("Plugin install requested: %s\n", plugin)
				return nil
			}

			// id@version or just id
			xml := fmt.Sprintf(`<jenkins><install plugin="%s@default" /></jenkins>`, plugin)
			if strings.Contains(plugin, "@") {
				xml = fmt.Sprintf(`<jenkins><install plugin="%s" /></jenkins>`, plugin)
			}
			resp, err := cl.PostXML("/pluginManager/installNecessaryPlugins", xml)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Plugin %q install requested.\n", plugin)
			return nil
		},
	}
}

func uploadPlugin(cl interface {
	Post(string, io.Reader, string) (*http.Response, error)
}, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open plugin file: %w", err)
	}
	defer f.Close()

	var buf strings.Builder
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return err
	}
	mw.Close()

	resp, err := cl.Post("/pluginManager/uploadPlugin", strings.NewReader(buf.String()), mw.FormDataContentType())
	if err != nil {
		return err
	}
	drainClose(resp)
	fmt.Printf("Plugin %q uploaded.\n", filepath.Base(filePath))
	return nil
}

func newEnablePluginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable-plugin <name>",
		Short: "Enable a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			path := "/pluginManager/plugin/" + url.PathEscape(args[0]) + "/makeEnabled"
			resp, err := cl.PostEmpty(path)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Plugin %q enabled.\n", args[0])
			return nil
		},
	}
}

func newDisablePluginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable-plugin <name>",
		Short: "Disable a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := resolveClient()
			if err != nil {
				return err
			}
			path := "/pluginManager/plugin/" + url.PathEscape(args[0]) + "/makeDisabled"
			resp, err := cl.PostEmpty(path)
			if err != nil {
				return err
			}
			drainClose(resp)
			fmt.Printf("Plugin %q disabled.\n", args[0])
			return nil
		},
	}
}
