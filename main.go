package main

import (
	"context"
	"flag"
	"fmt"
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
		r.Exit(err)
	}
}

var cmds = []acmd.Command{
	{
		Name:        "almostInlined",
		Alias:       "inl",
		Description: "find functions that cross inlining threshold just barely",
		ExecFunc: func(_ context.Context, args []string) error {
			if err := parseArgs("almostInlined", args); err != nil {
				return err
			}
			return run(&almostInlinedRunner{})
		},
	},
	{
		Name:        "escapedVariables",
		Alias:       "esc",
		Description: "find variables that are escaped to the heap",
		ExecFunc: func(_ context.Context, args []string) error {
			if err := parseArgs("escapedVariables", args); err != nil {
				return err
			}
			return run(&escapeAnalysisRunner{})
		},
	},
	{
		Name:        "boundChecks",
		Alias:       "bce",
		Description: "find slice/array that has bound check",
		ExecFunc: func(_ context.Context, args []string) error {
			if err := parseArgs("boundChecks", args); err != nil {
				return err
			}
			return run(&boundCheckRunner{})
		},
	},
	{
		Name:        "funcSize",
		Alias:       "fsize",
		Description: "list function machine code sizes in bytes",
		ExecFunc: func(_ context.Context, args []string) error {
			if err := parseArgs("funcSize", args); err != nil {
				return err
			}
			return run(&funcSizeRunner{})
		},
	},
	{
		Name:        "benchstat",
		Alias:       "bs",
		Description: "stricter benchstat with colors",
		ExecFunc: func(_ context.Context, args []string) error {
			return runBenchstat(args)
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
		// skip - and -- paramets
		if strings.HasPrefix(pkg, "-") {
			continue
		}
		if err := cmd.Run(pkg); err != nil {
			fmt.Printf("%s: %v", pkg, err)
		}
	}
	return nil
}

func parseArgs(cmd string, args []string) error {
	fset := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fset.StringVar(&flagMod, "mod", "", `-mod compiler flag(readonly|vendor)`)
	fset.BoolVar(&asJSON, "json", false, `return result as JSON`)
	return fset.Parse(args)
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
