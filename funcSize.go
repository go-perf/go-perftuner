package main

import (
	"fmt"
	"os/exec"
	"regexp"
)

type funcSizeRunner struct {
	messageRE *regexp.Regexp
}

func (r *funcSizeRunner) Init() {
	r.messageRE = regexp.MustCompile(`(.*) STEXT.* size=(\d+)`)
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

	// TODO: add a CLI flag for the function name filtering?
	// Having to use a grep in 99% of use cases is not very convenient.
	for _, submatches := range r.messageRE.FindAllStringSubmatch(string(out), -1) {
		results = append(results, resultRow{
			Fn:   submatches[1],
			Size: submatches[2],
		})
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

func (r *funcSizeRunner) getCmd(pkg string) []string {
	return goArgs(pkg, goArgsBuild, goArgsGcFlags("-S"))
}
