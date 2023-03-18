package main

import (
	"github.com/cristalhq/acmd"
)

const version = "v0.0.0"

func main() {
	r := acmd.RunnerOf(cmds, acmd.Config{
		Version: version,
	})
	if err := r.Run(); err != nil {
		r.Exit(err)
	}
}

var cmds = []acmd.Command{
	{
		Name:        "almostInlined",
		Alias:       "inl",
		Description: "find functions that cross inlining threshold just barely",
		Exec:        &almostInlinedRunner{},
	},
	{
		Name:        "escapedVariables",
		Alias:       "esc",
		Description: "find variables that are escaped to the heap",
		Exec:        &escapeAnalysisRunner{},
	},
	{
		Name:        "boundChecks",
		Alias:       "bce",
		Description: "find slice/array that has bound check",
		Exec:        &boundCheckRunner{},
	},
	{
		Name:        "funcSize",
		Alias:       "fsize",
		Description: "list function machine code sizes in bytes",
		Exec:        &funcSizeRunner{},
	},
	{
		Name:        "benchstat",
		Alias:       "bstat",
		Description: "stricter benchstat with colors",
		Exec:        &benchstatRunner{},
	},
}
