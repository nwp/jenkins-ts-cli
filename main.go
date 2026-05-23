package main

import (
	"os"

	"github.com/nwp/jenkins-cli/cmd"
)

func main() {
	os.Args = cmd.NormalizeArgs(os.Args)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
