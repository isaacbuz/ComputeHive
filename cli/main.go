package main

import (
	"github.com/computehive/cli/cmd"
)

// Build variables set by ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Set version information
	cmd.SetVersionInfo(version, commit, date, builtBy)
	
	// Execute the root command
	cmd.Execute()
} 