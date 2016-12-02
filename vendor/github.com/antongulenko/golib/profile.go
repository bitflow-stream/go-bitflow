package golib

import (
	"flag"
	"os"
	"runtime"
	"runtime/pprof"
)

var (
	CpuProfileFile = ""
	MemProfileFile = ""
)

func init() {
	flag.StringVar(&CpuProfileFile, "cpuprofile", CpuProfileFile, "Write cpu profile data to file")
	flag.StringVar(&MemProfileFile, "memprofile", MemProfileFile, "Write memory profile data to file")
}

// Usage: defer golib.ProfileCpu()()
// Performs both cpu and memory profiling, if enabled
func ProfileCpu() (stopProfiling func()) {
	var cpu, mem *os.File
	var err error
	if CpuProfileFile != "" {
		cpu, err = os.Create(CpuProfileFile)
		if err != nil {
			Log.Fatalln(err)
		}
		if err := pprof.StartCPUProfile(cpu); err != nil {
			Log.Fatalln("Unable to start CPU profile:", err)
		}
	}
	if MemProfileFile != "" {
		mem, err = os.Create(MemProfileFile)
		if err != nil {
			Log.Fatalln(err)
		}
	}
	return func() {
		if cpu != nil {
			pprof.StopCPUProfile()
		}
		if mem != nil {
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(mem); err != nil {
				Log.Warnln("Failed to write Memory profile:", err)
			}
			mem.Close()
		}
	}
}
