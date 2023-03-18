package main

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"regexp"
)

type escapeAnalysisRunner struct {
	flagMod string
	asJSON  bool

	messageRE *regexp.Regexp
}

func (r *escapeAnalysisRunner) ExecCommand(_ context.Context, args []string) error {
	fset := flag.NewFlagSet("escapeAnalysis", flag.ContinueOnError)
	fset.StringVar(&r.flagMod, "mod", "", `-mod compiler flag(readonly|vendor)`)
	fset.BoolVar(&r.asJSON, "json", false, `return result as JSON`)
	if err := fset.Parse(args); err != nil {
		return err
	}

	args = fset.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	const location = `(.*:\d+:\d+)`
	const variable = `(.*)`
	const pat = location + `: ` + variable + `escapes to heap`
	r.messageRE = regexp.MustCompile(pat)

	for _, pkg := range args {
		if err := r.process(pkg); err != nil {
			fmt.Printf("%s: %v", pkg, err)
		}
	}
	return nil
}

func (r *escapeAnalysisRunner) Init() {}

func (r *escapeAnalysisRunner) Run(pkg string) error { return nil }

func (r *escapeAnalysisRunner) process(pkg string) error {
	cmd := exec.Command("go", r.getCmdArgs(pkg)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}

	type escapeAnalysisResult struct {
		Loc      string `json:"loc"`
		Variable string `json:"var"`
	}
	results := []escapeAnalysisResult{}

	for _, submatches := range r.messageRE.FindAllStringSubmatch(string(out), -1) {
		loc := submatches[1]
		variable := submatches[2]

		results = append(results, escapeAnalysisResult{
			Loc:      loc,
			Variable: variable,
		})
	}

	if r.asJSON {
		marshalJSON(results)
		return nil
	}

	for _, r := range results {
		fmt.Printf("%s: %s\n", r.Loc, r.Variable)
	}
	return nil
}

func (r *escapeAnalysisRunner) getCmdArgs(pkg string) []string {
	args := []string{"build", "-gcflags", "-m -m"}
	if r.flagMod != "" {
		args = append(args, "-mod", r.flagMod)
	}
	args = append(args, pkg)
	return args
}
