package main

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"regexp"
)

type almostInlinedRunner struct {
	flagMod   string
	asJSON    bool
	threshold int

	messageRE *regexp.Regexp
}

func (r *almostInlinedRunner) ExecCommand(_ context.Context, args []string) error {
	fset := flag.NewFlagSet("boundCheck", flag.ContinueOnError)
	fset.StringVar(&r.flagMod, "mod", "", `-mod compiler flag(readonly|vendor)`)
	fset.BoolVar(&r.asJSON, "json", false, `return result as JSON`)
	flag.IntVar(&r.threshold, "threshold", 10, `max inliner budget overflow threshold`)
	if err := fset.Parse(args); err != nil {
		return err
	}

	args = fset.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	const location = `(.*:\d+:\d+)`
	const function = `((?:\S*)?\w+)`
	const pat = location + `: .*? ` + function + `: function too complex: cost (\d+) exceeds budget (\d+)`
	r.messageRE = regexp.MustCompile(pat)

	for _, pkg := range args {
		if err := r.process(pkg); err != nil {
			fmt.Printf("%s: %v", pkg, err)
		}
	}
	return nil
}

func (r *almostInlinedRunner) process(pkg string) error {
	cmd := exec.Command("go", r.getCmdArgs(pkg)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}

	type almostInlinedResult struct {
		Loc  string `json:"loc"`
		Fn   string `json:"fn"`
		Cost int    `json:"cost"`
		Diff int    `json:"diff"`
	}
	results := []almostInlinedResult{}

	for _, submatches := range r.messageRE.FindAllStringSubmatch(string(out), -1) {
		loc := submatches[1]
		fn := submatches[2]
		cost := atoi(submatches[3])
		budget := atoi(submatches[4])
		diff := cost - budget

		if r.threshold == 0 || diff <= r.threshold {
			results = append(results, almostInlinedResult{
				Loc:  loc,
				Fn:   fn,
				Cost: cost,
				Diff: diff,
			})
		}
	}

	if r.asJSON {
		marshalJSON(results)
		return nil
	}

	for _, r := range results {
		fmt.Printf("%s: %s: budget exceeded by %d\n", r.Loc, r.Fn, r.Diff)
	}
	return nil
}

func (r *almostInlinedRunner) getCmdArgs(pkg string) []string {
	args := []string{"build", "-gcflags", "-m=2"}
	if r.flagMod != "" {
		args = append(args, "-mod", r.flagMod)
	}
	args = append(args, pkg)
	return args
}
