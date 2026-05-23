package cmd

import "os"

// stdinReader returns os.Stdin. Extracted so tests can replace it.
var stdinReader = func() *os.File { return os.Stdin }
