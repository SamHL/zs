package main

import (
	"github.com/SamHL/zs/cmd"
)

// Version information set by build flags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	cmd.Execute()
}
