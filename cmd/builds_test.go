package cmd

import (
	"testing"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		ms   int64
		want string
	}{
		{0, "0ms"},
		{500, "500ms"},
		{999, "999ms"},
		{1000, "1s"},
		{59000, "59s"},
		{60000, "1m 0s"},
		{90000, "1m 30s"},
		{3600000, "1h 0m 0s"},
		{3661000, "1h 1m 1s"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.ms)
		if got != tt.want {
			t.Errorf("formatDuration(%d) = %q, want %q", tt.ms, got, tt.want)
		}
	}
}

func TestSplitParams(t *testing.T) {
	m, err := splitParams([]string{"key=value", "a=b=c"})
	if err != nil {
		t.Fatal(err)
	}
	if m["key"] != "value" {
		t.Errorf("expected key=value, got %v", m["key"])
	}
	// Value with = should be preserved
	if m["a"] != "b=c" {
		t.Errorf("expected a=b=c, got %v", m["a"])
	}
}

func TestSplitParams_Invalid(t *testing.T) {
	_, err := splitParams([]string{"noequalssign"})
	if err == nil {
		t.Error("expected error for missing =")
	}
}

func TestBuildRef(t *testing.T) {
	if buildRef([]string{"job"}, 1) != "lastBuild" {
		t.Error("expected lastBuild when no second arg")
	}
	if buildRef([]string{"job", "42"}, 1) != "42" {
		t.Error("expected 42")
	}
}
