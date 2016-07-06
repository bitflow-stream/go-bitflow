package golib

import (
	"flag"
	"syscall"
)

var (
	ConfiguredOpenFilesLimit uint64
)

func init() {
	flag.Uint64Var(&ConfiguredOpenFilesLimit, "ofl", ConfiguredOpenFilesLimit,
		"Set to >0 for configuring the open files limit (only possible as root)")
}

func ConfigureOpenFilesLimit() {
	if ConfiguredOpenFilesLimit > 0 {
		if err := SetOpenFilesLimit(ConfiguredOpenFilesLimit); err != nil {
			Log.Println("Failed to set open files limit to %v: %v", ConfiguredOpenFilesLimit, err)
		} else {
			Log.Println("Successfully set open files limit to %v", ConfiguredOpenFilesLimit)
		}
	}
}

func SetOpenFilesLimit(ulimit uint64) error {
	rLimit := syscall.Rlimit{
		Max: ulimit,
		Cur: ulimit,
	}
	return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
}
