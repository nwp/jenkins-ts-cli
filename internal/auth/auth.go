package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nwp/jenkins-cli/internal/keyring"
)

// IsTTY reports whether stdin is a terminal. Overridable in tests.
var IsTTY = func() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// ResolveOptions holds inputs to ResolveCredentials.
type ResolveOptions struct {
	// AuthFlag is the raw value of --auth (username:token). May be empty.
	AuthFlag string
	// ServerURL is the normalized Jenkins URL used for keyring lookup.
	ServerURL string
}

// ResolveCredentials resolves credentials in priority order:
//  1. --auth flag (username:token)
//  2. JENKINS_USER_ID + JENKINS_API_TOKEN env vars
//  3. Stored credentials from OS keyring
//  4. Interactive TUI (if TTY) or error
func ResolveCredentials(opts ResolveOptions) (Credentials, error) {
	// 1. --auth flag
	if opts.AuthFlag != "" {
		idx := strings.Index(opts.AuthFlag, ":")
		if idx < 0 {
			return Credentials{}, errors.New("--auth must be in username:token format")
		}
		return Credentials{
			Username: opts.AuthFlag[:idx],
			Token:    opts.AuthFlag[idx+1:],
		}, nil
	}

	// 2. Environment variables
	userID := os.Getenv("JENKINS_USER_ID")
	apiToken := os.Getenv("JENKINS_API_TOKEN")
	if userID != "" && apiToken != "" {
		return Credentials{Username: userID, Token: apiToken}, nil
	}

	// 3. OS keyring
	if opts.ServerURL != "" {
		username, err := keyring.GetStoredUsername(opts.ServerURL)
		if err != nil {
			return Credentials{}, fmt.Errorf("keyring lookup failed: %w", err)
		}
		if username != "" {
			token, err := keyring.GetStoredToken(opts.ServerURL, username)
			if err != nil {
				return Credentials{}, fmt.Errorf("keyring token lookup failed: %w", err)
			}
			if token != "" {
				return Credentials{Username: username, Token: token}, nil
			}
		}
	}

	// 4. Interactive TUI (requires a TTY)
	if !IsTTY() {
		return Credentials{}, errors.New("no credentials found and stdin is not a terminal; " +
			"use --auth or set JENKINS_USER_ID and JENKINS_API_TOKEN, " +
			"or run 'jenkins configure set' to store credentials")
	}

	return RunTUI(opts.ServerURL)
}
