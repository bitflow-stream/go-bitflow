package golib

import (
	"syscall"
	"unsafe"
)

type TerminalWindowSize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func GetTerminalSize() (TerminalWindowSize, error) {
	var ws TerminalWindowSize
	res, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)))
	if res < 0 {
		return TerminalWindowSize{}, errno
	}
	return ws, nil
}
