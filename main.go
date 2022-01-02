package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cristalhq/acmd"
)

const version = "v0.0.0"

var (
	flagMod string
	tags    string
	asJSON  bool
)

func main() {
	if len(os.Args) < 2 {
		panic("not enough arguments, expected sub-command name")
	}

	log.SetFlags(0)
	flag.StringVar(&flagMod, "mod", "", `-mod compiler flag(readonly|vendor)`)
	flag.StringVar(&tags, "tags", "", `-tags compiler flag`)
	flag.BoolVar(&asJSON, "json", false, `return result as JSON`)
	flag.Parse()

	r := acmd.RunnerOf(cmds, acmd.Config{
		Version: version,
	})
	if err := r.Run(); err != nil {
		log.Fatal(fmt.Errorf("dbumper: %w", err))
	}
}

var cmds = []acmd.Command{
	{
		Name:        "almostInlined",
		Alias:       "inl",
		Description: "find functions that cross inlining threshold just barely",
		Do: func(_ context.Context, _ []string) error {
			return run(&almostInlinedRunner{})
		},
	},
	{
		Name:        "escapedVariables",
		Alias:       "esc",
		Description: "find variables that are escaped to the heap",
		Do: func(_ context.Context, _ []string) error {
			return run(&escapeAnalysisRunner{})
		},
	},
	{
		Name:        "boundChecks",
		Alias:       "bce",
		Description: "find slice/array that has bound check",
		Do: func(_ context.Context, _ []string) error {
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
			log.Printf("%s: %v", pkg, err)
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
	args = append(args, goTags()...)
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

func goTags() (args []string) {
	if tags != "" {
		args = append(args, "-tags", tags)
	}
	return args
}
