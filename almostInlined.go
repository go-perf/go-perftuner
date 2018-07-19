package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

type almostInlinedRunner struct {
	threshold int

	messageRE *regexp.Regexp
}

func (r *almostInlinedRunner) Init() {
	flag.IntVar(&r.threshold, "threshold", 10, `max inliner budget overflow threshold`)

	const location = `(src/.*:\d+:\d+)`
	const function = `((?:\S*)?\w+)`
	const pat = location + `: .*? ` + function + `: function too complex: cost (\d+) exceeds budget (\d+)`
	r.messageRE = regexp.MustCompile(pat)
}

func (r *almostInlinedRunner) Run(pkg string) error {
	cmd := exec.Command("go", "build", "-gcflags", "-m=2", pkg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}

	for _, submatches := range r.messageRE.FindAllStringSubmatch(string(out), -1) {
		loc := submatches[1]
		fn := submatches[2]
		cost := atoi(submatches[3])
		budget := atoi(submatches[4])
		diff := cost - budget
		if diff <= r.threshold {
			log.Printf("%s: %s: budget exceeded by %d\n", loc, fn, diff)
		}
	}

	return nil
}
