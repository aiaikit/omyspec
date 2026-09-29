// Command specify is the entry point for the Go CLI.
//
// Phase 1 mirrors src/specify_cli/__main__.py:argv[0] is rewritten
// to "specify" when invoked via `go run`, so usage text reads correctly.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
