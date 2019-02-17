package cmd

import (
	"flag"

	"github.com/antongulenko/golib"
)

func ParseFlags() (*flag.FlagSet, []string) {
	golib.RegisterFlags(golib.FlagsAll)
	previousFlags, args := golib.ParseFlags()
	golib.ConfigureLogging()
	return previousFlags, args
}
