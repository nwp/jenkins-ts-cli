package keyring

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	gokeyring "github.com/zalando/go-keyring"

	"github.com/nwp/jenkins-cli/internal/paths"
)

const serviceName = "jenkins-cli"

// credentialsFile is the path to ~/.jenkins-cli/credentials.json which maps
// normalized server URLs to usernames. API tokens are stored in the OS keyring.
func credentialsFile() (string, error) {
	dir, err := paths.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

func readCredMap() (map[string]string, error) {
	f, err := credentialsFile()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(f)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read credentials file: %w", err)
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse credentials file: %w", err)
	}
	return m, nil
}

func writeCredMap(m map[string]string) error {
	f, err := credentialsFile()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f, data, 0o600)
}

// keyringAccount returns a consistent account key for a server + username pair.
func keyringAccount(normalizedURL, username string) string {
	return username + "@" + normalizedURL
}

// StoreCredentials persists username in credentials.json and token in the OS keyring.
func StoreCredentials(normalizedURL, username, token string) error {
	m, err := readCredMap()
	if err != nil {
		return err
	}
	m[normalizedURL] = username
	if err := writeCredMap(m); err != nil {
		return err
	}
	return gokeyring.Set(serviceName, keyringAccount(normalizedURL, username), token)
}

// GetStoredUsername returns the username stored for serverURL, or "" if not found.
func GetStoredUsername(normalizedURL string) (string, error) {
	m, err := readCredMap()
	if err != nil {
		return "", err
	}
	return m[normalizedURL], nil
}

// GetStoredToken returns the API token for the given server and username from the OS keyring.
func GetStoredToken(normalizedURL, username string) (string, error) {
	token, err := gokeyring.Get(serviceName, keyringAccount(normalizedURL, username))
	if errors.Is(err, gokeyring.ErrNotFound) {
		return "", nil
	}
	return token, err
}

// DeleteCredentials removes credentials from both credentials.json and the OS keyring.
func DeleteCredentials(normalizedURL string) error {
	m, err := readCredMap()
	if err != nil {
		return err
	}
	username, ok := m[normalizedURL]
	if !ok {
		return nil
	}
	delete(m, normalizedURL)
	if err := writeCredMap(m); err != nil {
		return err
	}
	_ = gokeyring.Delete(serviceName, keyringAccount(normalizedURL, username))
	return nil
}
