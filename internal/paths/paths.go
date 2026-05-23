package paths

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ConfigDir returns the path to ~/.jenkins-cli, creating it if necessary.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	dir := filepath.Join(home, ".jenkins-cli")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}
	return dir, nil
}

// NormalizeURL lowercases the scheme and host and removes any trailing slash.
func NormalizeURL(raw string) (string, error) {
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid URL %q: %w", raw, err)
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Path = strings.TrimRight(u.Path, "/")
	return u.String(), nil
}

// JobPath converts a job name (possibly slash-separated for nested folders) into
// the Jenkins URL path segment, e.g. "folder/job" → "/job/folder/job/job".
func JobPath(name string) string {
	parts := strings.Split(name, "/")
	var sb strings.Builder
	for _, p := range parts {
		sb.WriteString("/job/")
		sb.WriteString(url.PathEscape(p))
	}
	return sb.String()
}

// NodeURL returns the Jenkins URL path for an agent node.
func NodeURL(name string) string {
	return "/computer/" + url.PathEscape(name)
}

// ViewURL returns the Jenkins URL path for a view.
func ViewURL(name string) string {
	return "/view/" + url.PathEscape(name)
}
