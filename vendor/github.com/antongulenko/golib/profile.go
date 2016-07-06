package golib

import (
	"flag"
	"os"
	"runtime/pprof"
)

var CpuProfileFile = ""

func init() {
	flag.StringVar(&CpuProfileFile, "cpuprofile", CpuProfileFile, "Write cpu profile data to file")
}

// Usage: defer golib.ProfileCpu()()
func ProfileCpu() (stopProfiling func()) {
	if CpuProfileFile != "" {
		f, err := os.Create(CpuProfileFile)
		if err != nil {
			Log.Fatalln(err)
		} else {
			pprof.StartCPUProfile(f)
			return pprof.StopCPUProfile
		}
	}
	return func() {}
}
