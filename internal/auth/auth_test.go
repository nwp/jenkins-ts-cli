package auth_test

import (
	"os"
	"testing"

	"github.com/nwp/jenkins-cli/internal/auth"
)

func TestResolveCredentials_AuthFlag(t *testing.T) {
	creds, err := auth.ResolveCredentials(auth.ResolveOptions{
		AuthFlag:  "admin:mytoken",
		ServerURL: "http://jenkins.example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if creds.Username != "admin" || creds.Token != "mytoken" {
		t.Errorf("got %+v", creds)
	}
}

func TestResolveCredentials_AuthFlag_ColonInToken(t *testing.T) {
	// Token may itself contain colons; only split on first colon.
	creds, err := auth.ResolveCredentials(auth.ResolveOptions{
		AuthFlag: "admin:tok:en:with:colons",
	})
	if err != nil {
		t.Fatal(err)
	}
	if creds.Username != "admin" || creds.Token != "tok:en:with:colons" {
		t.Errorf("got %+v", creds)
	}
}

func TestResolveCredentials_AuthFlag_Invalid(t *testing.T) {
	_, err := auth.ResolveCredentials(auth.ResolveOptions{AuthFlag: "nocolon"})
	if err == nil {
		t.Error("expected error for missing colon")
	}
}

func TestResolveCredentials_EnvVars(t *testing.T) {
	t.Setenv("JENKINS_USER_ID", "envuser")
	t.Setenv("JENKINS_API_TOKEN", "envtoken")

	creds, err := auth.ResolveCredentials(auth.ResolveOptions{ServerURL: "http://j.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if creds.Username != "envuser" || creds.Token != "envtoken" {
		t.Errorf("got %+v", creds)
	}
}

func TestResolveCredentials_AuthFlagTakesPriorityOverEnv(t *testing.T) {
	t.Setenv("JENKINS_USER_ID", "envuser")
	t.Setenv("JENKINS_API_TOKEN", "envtoken")

	creds, err := auth.ResolveCredentials(auth.ResolveOptions{
		AuthFlag: "flaguser:flagtoken",
	})
	if err != nil {
		t.Fatal(err)
	}
	if creds.Username != "flaguser" {
		t.Errorf("auth flag should take priority, got %+v", creds)
	}
}

func TestResolveCredentials_NoCredsNoTTY(t *testing.T) {
	// Make sure JENKINS_ env vars are unset
	os.Unsetenv("JENKINS_USER_ID")
	os.Unsetenv("JENKINS_API_TOKEN")

	// Override TTY detection so the TUI never launches in tests
	orig := auth.IsTTY
	auth.IsTTY = func() bool { return false }
	defer func() { auth.IsTTY = orig }()

	// No stored creds (empty server URL), no TTY → expect error
	_, err := auth.ResolveCredentials(auth.ResolveOptions{ServerURL: ""})
	if err == nil {
		t.Error("expected error when no credentials available and no TTY")
	}
}
