package paths_test

import (
	"testing"

	"github.com/nwp/jenkins-cli/internal/paths"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"http://Jenkins.Example.com/", "http://jenkins.example.com"},
		{"HTTP://JENKINS.EXAMPLE.COM", "http://jenkins.example.com"},
		{"https://jenkins.example.com/jenkins/", "https://jenkins.example.com/jenkins"},
		{"https://jenkins.example.com/jenkins", "https://jenkins.example.com/jenkins"},
		{"jenkins.example.com", "http://jenkins.example.com"},
	}
	for _, tt := range tests {
		got, err := paths.NormalizeURL(tt.input)
		if err != nil {
			t.Errorf("NormalizeURL(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestJobPath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"myjob", "/job/myjob"},
		{"folder/myjob", "/job/folder/job/myjob"},
		{"folder/sub/myjob", "/job/folder/job/sub/job/myjob"},
		{"my job", "/job/my%20job"},
	}
	for _, tt := range tests {
		got := paths.JobPath(tt.name)
		if got != tt.want {
			t.Errorf("JobPath(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestNodeURL(t *testing.T) {
	if got := paths.NodeURL("my-agent"); got != "/computer/my-agent" {
		t.Errorf("NodeURL = %q", got)
	}
	if got := paths.NodeURL("agent with space"); got != "/computer/agent%20with%20space" {
		t.Errorf("NodeURL = %q", got)
	}
}

func TestViewURL(t *testing.T) {
	if got := paths.ViewURL("All"); got != "/view/All" {
		t.Errorf("ViewURL = %q", got)
	}
}
