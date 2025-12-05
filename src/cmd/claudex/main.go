package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Version is set at build time via -ldflags
var Version = "dev"

// stringSlice implements flag.Value to allow multiple --doc flags
type stringSlice []string

func (s *stringSlice) String() string     { return strings.Join(*s, ":") }
func (s *stringSlice) Set(v string) error { *s = append(*s, v); return nil }

var noOverwrite = flag.Bool("no-overwrite", false, "skip overwriting existing .claude files")
var showVersion = flag.Bool("version", false, "print version and exit")
var docPaths stringSlice

func init() {
	flag.Var(&docPaths, "doc", "documentation path for agent context (can be specified multiple times)")
}

func main() {
	app := NewApp()

	if err := app.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
