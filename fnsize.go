package main

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"regexp"
)

type funcSizeRunner struct {
	flagMod string
	asJSON  bool

	messageRE *regexp.Regexp
	filterRE  *regexp.Regexp
}

func (r *funcSizeRunner) ExecCommand(_ context.Context, args []string) error {
	fset := flag.NewFlagSet("funcSize", flag.ContinueOnError)
	fset.StringVar(&r.flagMod, "mod", "", `-mod compiler flag(readonly|vendor)`)
	fset.BoolVar(&r.asJSON, "json", false, `return result as JSON`)
	filter := fset.String("filter", "", "regex to filter the results")
	if err := fset.Parse(args); err != nil {
		return err
	}

	args = fset.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	r.messageRE = regexp.MustCompile(`(.*) STEXT.* size=(\d+)`)
	if *filter != "" {
		r.filterRE = regexp.MustCompile(*filter)
	}

	for _, pkg := range args {
		if err := r.process(pkg); err != nil {
			fmt.Printf("%s: %v", pkg, err)
		}
	}
	return nil
}

func (r *funcSizeRunner) process(pkg string) error {
	cmd := exec.Command("go", r.getCmdArgs(pkg)...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}

	type resultRow struct {
		Fn   string `json:"fn"`
		Size string `json:"size"`
	}
	results := []resultRow{}

	for _, submatches := range r.messageRE.FindAllStringSubmatch(string(out), -1) {
		fn, size := submatches[1], submatches[2]

		if r.passesFilter(fn) {
			results = append(results, resultRow{
				Fn:   fn,
				Size: size,
			})
		}
	}

	if r.asJSON {
		marshalJSON(results)
		return nil
	}

	for _, r := range results {
		fmt.Printf("%s: %s bytes\n", r.Fn, r.Size)
	}
	return nil
}

func (r *funcSizeRunner) passesFilter(fn string) bool {
	return r.filterRE == nil || r.filterRE.MatchString(fn)
}

func (r *funcSizeRunner) getCmdArgs(pkg string) []string {
	args := []string{"build", "-gcflags", "-S"}
	if r.flagMod != "" {
		args = append(args, "-mod", r.flagMod)
	}
	args = append(args, pkg)
	return args
}
