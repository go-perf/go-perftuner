package main

import (
	"flag"
	"fmt"
	"os/exec"
	"regexp"
)

type almostInlinedRunner struct {
	threshold int

	messageRE *regexp.Regexp
}

func (r *almostInlinedRunner) Init() {
	flag.IntVar(&r.threshold, "threshold", 10, `max inliner budget overflow threshold`)

	const location = `(.*:\d+:\d+)`
	const function = `((?:\S*)?\w+)`
	const pat = location + `: .*? ` + function + `: function too complex: cost (\d+) exceeds budget (\d+)`
	r.messageRE = regexp.MustCompile(pat)
}

func (r *almostInlinedRunner) Run(pkg string) error {
	cmd := exec.Command("go", r.getCmd(pkg)...)
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

	if asJSON {
		marshalJSON(results)
		return nil
	}

	for _, r := range results {
		fmt.Printf("%s: %s: budget exceeded by %d\n", r.Loc, r.Fn, r.Diff)
	}
	return nil
}

func (r *almostInlinedRunner) getCmd(pkg string) []string {
	return goArgs(pkg, goArgsBuild, goArgsGcFlags("-m=2"))
}
