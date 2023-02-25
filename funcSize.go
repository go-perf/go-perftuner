package main

import (
	"fmt"
	"os/exec"
	"regexp"
)

type funcSizeRunner struct {
	messageRE *regexp.Regexp
	filterRE  *regexp.Regexp
}

func (r *funcSizeRunner) Init() {
	r.messageRE = regexp.MustCompile(`(.*) STEXT.* size=(\d+)`)
	if filter != "" {
		r.filterRE = regexp.MustCompile(filter)
	}
}

func (r *funcSizeRunner) Run(pkg string) error {
	cmd := exec.Command("go", r.getCmd(pkg)...)

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

	if asJSON {
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

func (r *funcSizeRunner) getCmd(pkg string) []string {
	return goArgs(pkg, goArgsBuild, goArgsGcFlags("-S"))
}
