package main

import (
	"fmt"
	"log"
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
	cmd := exec.Command("go", "build", "-gcflags", "-m -m", pkg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}

	for _, submatches := range r.messageRE.FindAllStringSubmatch(string(out), -1) {
		loc := submatches[1]
		variable := submatches[2]
		log.Printf("%s: %s\n", loc, variable)
	}
	return nil
}
