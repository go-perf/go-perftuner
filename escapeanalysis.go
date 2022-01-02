package main

import (
	"fmt"
	"os/exec"
	"regexp"
)

type escapeAnalysisRunner struct {
	messageRE *regexp.Regexp
}

func (r *escapeAnalysisRunner) Init() {
	const location = `(.*:\d+:\d+)`
	const variable = `(.*)`
	const pat = location + `: ` + variable + `escapes to heap`
	r.messageRE = regexp.MustCompile(pat)
}

func (r *escapeAnalysisRunner) Run(pkg string) error {
	cmd := exec.Command("go", r.getCmd(pkg)...)
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

	if asJSON {
		marshalJSON(results)
		return nil
	}

	for _, r := range results {
		fmt.Printf("%s: %s\n", r.Loc, r.Variable)
	}
	return nil
}

func (r *escapeAnalysisRunner) getCmd(pkg string) []string {
	return goArgs(pkg, goArgsBuild, goArgsGcFlags("-m -m"))
}
