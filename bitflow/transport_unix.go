// +build !windows

package bitflow

import (
	"os"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type fileVanishChecker struct {
	currentIno uint64
}

func (f *fileVanishChecker) setCurrentFile(name string) error {
	stat, err := os.Stat(name)
	if err == nil {
		f.currentIno = stat.Sys().(*syscall.Stat_t).Ino
	}
	return err
}

func (f *fileVanishChecker) hasFileVanished(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.WithField("file", name).Warn("Error stating opened output file:", err)
		return true
	} else {
		newIno := info.Sys().(*syscall.Stat_t).Ino
		if newIno != f.currentIno {
			log.WithField("file", name).Warnf("Output file inumber has changed, file was moved or vanished (%v -> %v)", f.currentIno, newIno)
			return true
		}
	}
	return false
}

// IsFileClosedError returns true, if the given error likely originates from intentionally
// closing a file, while it is still being read concurrently.
func IsFileClosedError(err error) bool {
	pathErr, ok := err.(*os.PathError)
	return ok && pathErr.Err == syscall.EBADF
}

func IsBrokenPipeError(err error) bool {
	if err == syscall.EPIPE {
		return true
	} else {
		if syscallErr, ok := err.(*os.SyscallError); ok && IsBrokenPipeError(syscallErr.Err) {
			return true
		}
	}
	return false
}
