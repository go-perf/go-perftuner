package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

type boundCheckRunner struct {
	messageRE *regexp.Regexp
}

func (r *boundCheckRunner) Init() {
	const location = `(.*:\d+:\d+)`
	const pat = location + `: Found Is(Slice)?InBounds`
	r.messageRE = regexp.MustCompile(pat)
}

func (r *boundCheckRunner) Run(pkg string) error {
	cmd := exec.Command("go", r.getCmd(pkg)...)
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

	if asJSON {
		marshalJSON(results)
		return nil
	}

	for _, r := range results {
		log.Printf("%s: slice/array has bound checks\n", r.Loc)
	}
	return nil
}

func (r *boundCheckRunner) getCmd(pkg string) []string {
	return goArgs(pkg, goArgsBuild, goArgsGcFlags("-d=ssa/check_bce/debug=1"))
}
