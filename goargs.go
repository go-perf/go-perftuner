package main

import (
	"flag"
	"strings"
)

var (
	flagMod string
)

func init() {
	flag.StringVar(&flagMod, "mod", "", `-mod compiler flag(readonly|vendor)`)
}

type goArgFunc func() []string

func goArgs(pkg string, argFuncs ...goArgFunc) (args []string) {
	for _, f := range argFuncs {
		args = append(args, f()...)
	}
	args = append(args, goArgsMod()...)
	args = append(args, pkg)

	return
}

func goArgsBuild() []string {
	return []string{"build"}
}

func goArgsGcFlags(flags ...string) goArgFunc {
	return func() []string {
		return []string{"-gcflags", strings.Join(flags, " ")}
	}
}

func goArgsMod() (args []string) {
	if flagMod != "" {
		args = append(args, "-mod", flagMod)
	}

	return args
}
