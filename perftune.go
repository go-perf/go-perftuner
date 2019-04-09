package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var asJSON bool

var subCommands = []*subCommand{
	{
		runner:    &almostInlinedRunner{},
		name:      "almostInlined",
		shortName: "inl",
		summary:   "find functions that cross inlining threshold just barely",
	},
	{
		runner:    &escapeAnalysisRunner{},
		name:      "escapedVariables",
		shortName: "esc",
		summary:   "find variables that are escaped to the heap",
	},
	{
		runner:    &boundCheckRunner{},
		name:      "boundChecks",
		shortName: "bce",
		summary:   "find slice/array that has bound check",
	},
}

type subCommandRunner interface {
	Init()
	Run(pkg string) error
}

type subCommand struct {
	runner    subCommandRunner
	name      string
	shortName string
	summary   string
}

func main() {
	log.SetFlags(0)

	argv := os.Args
	if len(argv) < 2 {
		terminate("not enough arguments, expected sub-command name", printUsage)
	}

	subIdx := 1 // [0] is program name
	sub := os.Args[subIdx]
	// Erase sub-command argument (index=1) to make it invisible for
	// sub commands themselves.
	os.Args = append(os.Args[:subIdx], os.Args[subIdx+1:]...)

	// Choose and run sub-command main.
	cmd := findSubCommand(sub)
	if cmd == nil {
		terminate("unknown sub-command: "+sub, printSupportedSubs)
	}
	flag.Usage = func() {
		log.Printf("usage: [flags] package...")
		flag.PrintDefaults()
	}

	flag.BoolVar(&asJSON, "json", false, `return result as JSON`)

	cmd.runner.Init()
	flag.Parse()

	for _, pkg := range flag.Args() {
		log.SetPrefix(sub + ": " + pkg + ": ")
		if err := cmd.runner.Run(pkg); err != nil {
			log.Printf("%s: %v", pkg, err)
		}
	}
}

// findSubCommand looks up subCommand by its name.
// Returns nil if requested command not found.
func findSubCommand(name string) *subCommand {
	for _, cmd := range subCommands {
		if cmd.name == name || cmd.shortName == name {
			return cmd
		}
	}
	return nil
}

func printUsage() {
	// TODO: implement me. For now, print supported commands.
	printSupportedSubs()
}

func printSupportedSubs() {
	stderrPrintf("Supported sub-commands:\n")
	for _, cmd := range subCommands {
		stderrPrintf("\t%s (or %s) - %s\n", cmd.name, cmd.shortName, cmd.summary)
	}
}

// terminate prints error specified by reason, runs optional printHelp
// function and then exists with non-zero status.
func terminate(reason string, printHelp func()) {
	stderrPrintf("error: %s\n", reason)
	if printHelp != nil {
		stderrPrintf("\n")
		printHelp()
	}
	os.Exit(1)
}

// stderrPrintf writes formatted message to stderr and checks for error
// making "not annoying at all" linters happy.
func stderrPrintf(format string, args ...interface{}) {
	_, err := fmt.Fprintf(os.Stderr, format, args...)
	if err != nil {
		panic(fmt.Sprintf("stderr write error: %v", err))
	}
}
