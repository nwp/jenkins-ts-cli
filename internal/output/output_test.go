package output_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/nwp/jenkins-cli/internal/output"
)

// captureStdout captures anything written to os.Stdout during fn.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrint_String_Plain(t *testing.T) {
	got := captureStdout(t, func() {
		output.Print("hello world", false)
	})
	if got != "hello world\n" {
		t.Errorf("got %q", got)
	}
}

func TestPrint_String_AlreadyNewline(t *testing.T) {
	got := captureStdout(t, func() {
		output.Print("hello\n", false)
	})
	if got != "hello\n" {
		t.Errorf("got %q", got)
	}
}

func TestPrint_StringSlice_Plain(t *testing.T) {
	got := captureStdout(t, func() {
		output.Print([]string{"a", "b", "c"}, false)
	})
	if got != "a\nb\nc\n" {
		t.Errorf("got %q", got)
	}
}

func TestPrint_JSON_Mode(t *testing.T) {
	got := captureStdout(t, func() {
		output.Print([]string{"a", "b"}, true)
	})
	want := "[\n  \"a\",\n  \"b\"\n]\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrint_Object_Plain(t *testing.T) {
	got := captureStdout(t, func() {
		output.Print(map[string]string{"key": "val"}, false)
	})
	if len(got) == 0 {
		t.Error("expected non-empty output for object")
	}
}
