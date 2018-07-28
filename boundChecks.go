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
	const location = `(src/.*:\d+:\d+)`
	const function = `((?:\S*)?\w+)`
	const pat = location + `: Found Is(Slice)?InBounds`
	r.messageRE = regexp.MustCompile(pat)
}

func (r *boundCheckRunner) Run(pkg string) error {
	cmd := exec.Command("go", "build", "-gcflags", "-d=ssa/check_bce/debug=1", pkg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}

	for _, submatches := range r.messageRE.FindAllStringSubmatch(string(out), -1) {
		loc := submatches[1]
		log.Printf("%s: slice/array has bound checks\n", loc)
	}

	return nil
}
