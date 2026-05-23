package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/nwp/jenkins-cli/internal/keyring"
)

// Credentials holds a Jenkins username and API token.
type Credentials struct {
	Username string
	Token    string
}

// RunTUI presents an interactive huh form to collect server URL (optional if
// already known), username, and API token. On success it stores the credentials
// and returns them. If serverURL is non-empty it is pre-filled and locked.
func RunTUI(serverURL string) (Credentials, error) {
	var username, token string
	urlLabel := serverURL
	if urlLabel == "" {
		urlLabel = "https://jenkins.example.com"
	}

	fields := []huh.Field{
		huh.NewInput().
			Title("Username").
			Description("Jenkins username").
			Value(&username).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("username is required")
				}
				return nil
			}),
		huh.NewInput().
			Title("API Token").
			Description("Jenkins API token (User > Configure > Add new Token)").
			EchoMode(huh.EchoModePassword).
			Value(&token).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("API token is required")
				}
				return nil
			}),
	}

	form := huh.NewForm(huh.NewGroup(fields...)).
		WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return Credentials{}, fmt.Errorf("auth form cancelled: %w", err)
	}

	creds := Credentials{
		Username: strings.TrimSpace(username),
		Token:    strings.TrimSpace(token),
	}

	if err := keyring.StoreCredentials(serverURL, creds.Username, creds.Token); err != nil {
		// Non-fatal: warn but continue — user can re-authenticate later
		fmt.Printf("warning: could not store credentials: %v\n", err)
	}

	return creds, nil
}
