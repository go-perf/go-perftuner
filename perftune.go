package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cristalhq/acmd"
)

const version = "v0.0.0"

func main() {
	if len(os.Args) < 2 {
		panic("not enough arguments, expected sub-command name")
	}

	log.SetFlags(0)
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
		Name:        "inl",
		Description: "find functions that cross inlining threshold just barely",
		Do: func(_ context.Context, _ []string) error {
			return run(&almostInlinedRunner{})
		},
	},
	{
		Name:        "esc",
		Description: "find variables that are escaped to the heap",
		Do: func(_ context.Context, _ []string) error {
			return run(&escapeAnalysisRunner{})
		},
	},
	{
		Name:        "bce",
		Description: "find slice/array that has bound check",
		Do: func(_ context.Context, _ []string) error {
			return run(&boundCheckRunner{})
		},
	},
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

var asJSON bool

type subCommandRunner interface {
	Init()
	Run(pkg string) error
}
