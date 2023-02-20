package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cristalhq/acmd"
)

const version = "v0.0.0"

var (
	flagMod string
	asJSON  bool
)

func main() {
	flag.StringVar(&flagMod, "mod", "", `-mod compiler flag(readonly|vendor)`)
	flag.BoolVar(&asJSON, "json", false, `return result as JSON`)
	flag.Parse()

	r := acmd.RunnerOf(cmds, acmd.Config{
		Version: version,
	})
	if err := r.Run(); err != nil {
		fmt.Printf("go-perf: %v\n", err)
		os.Exit(1)
	}
}

var cmds = []acmd.Command{
	{
		Name:        "almostInlined",
		Alias:       "inl",
		Description: "find functions that cross inlining threshold just barely",
		ExecFunc: func(_ context.Context, _ []string) error {
			return run(&almostInlinedRunner{})
		},
	},
	{
		Name:        "escapedVariables",
		Alias:       "esc",
		Description: "find variables that are escaped to the heap",
		ExecFunc: func(_ context.Context, _ []string) error {
			return run(&escapeAnalysisRunner{})
		},
	},
	{
		Name:        "boundChecks",
		Alias:       "bce",
		Description: "find slice/array that has bound check",
		ExecFunc: func(_ context.Context, _ []string) error {
			return run(&boundCheckRunner{})
		},
	},
}

type subCommandRunner interface {
	Init()
	Run(pkg string) error
}

func run(cmd subCommandRunner) error {
	cmd.Init()
	for _, pkg := range flag.Args()[1:] {
		if err := cmd.Run(pkg); err != nil {
			fmt.Printf("%s: %v", pkg, err)
		}
	}
	return nil
}

type goArgFunc func() []string

func goArgs(pkg string, argFuncs ...goArgFunc) (args []string) {
	for _, f := range argFuncs {
		args = append(args, f()...)
	}
	args = append(args, goArgsMod()...)
	args = append(args, pkg)

	return
}

func goArgsBuild() []string {
	return []string{"build"}
}

func goArgsGcFlags(flags ...string) goArgFunc {
	return func() []string {
		return []string{"-gcflags", strings.Join(flags, " ")}
	}
}

func goArgsMod() (args []string) {
	if flagMod != "" {
		args = append(args, "-mod", flagMod)
	}
	return args
}
