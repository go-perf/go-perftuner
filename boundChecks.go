package main

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"regexp"
)

type boundCheckRunner struct {
	flagMod string
	asJSON  bool

	messageRE *regexp.Regexp
}

func (r *boundCheckRunner) ExecCommand(_ context.Context, args []string) error {
	fset := flag.NewFlagSet("boundCheck", flag.ContinueOnError)
	fset.StringVar(&r.flagMod, "mod", "", `-mod compiler flag(readonly|vendor)`)
	fset.BoolVar(&r.asJSON, "json", false, `return result as JSON`)
	if err := fset.Parse(args); err != nil {
		return err
	}

	args = fset.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	r.messageRE = regexp.MustCompile(`(.*:\d+:\d+): Found Is(Slice)?InBounds`)

	for _, pkg := range args {
		if err := r.process(pkg); err != nil {
			fmt.Printf("%s: %v", pkg, err)
		}
	}
	return nil
}

func (r *boundCheckRunner) process(pkg string) error {
	cmd := exec.Command("go", r.getCmdArgs(pkg)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}

	type boundCheckResult struct {
		Loc string `json:"loc"`
	}
	results := []boundCheckResult{}

	for _, submatches := range r.messageRE.FindAllStringSubmatch(string(out), -1) {
		loc := submatches[1]
		results = append(results, boundCheckResult{
			Loc: loc,
		})
	}

	if r.asJSON {
		marshalJSON(results)
		return nil
	}

	for _, r := range results {
		fmt.Printf("%s: slice/array has bound checks\n", r.Loc)
	}
	return nil
}

func (r *boundCheckRunner) getCmdArgs(pkg string) []string {
	args := []string{"build", "-gcflags", "-d=ssa/check_bce/debug=1"}
	if r.flagMod != "" {
		args = append(args, "-mod", r.flagMod)
	}
	args = append(args, pkg)
	return args
}
