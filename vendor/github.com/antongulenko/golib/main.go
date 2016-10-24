package golib

import "os"

var (
	ErrorExitHook    func()
	checkerr_exiting bool
)

func Checkerr(err error) {
	if err != nil {
		if checkerr_exiting {
			Log.Warnln("Recursive Checkerr:", err)
			return
		}
		checkerr_exiting = true
		Log.Errorln("Fatal:", err)
		if ErrorExitHook != nil {
			ErrorExitHook()
		}
		os.Exit(1)
	}
}

func Printerr(err error) {
	if err != nil {
		Log.Errorln(err)
	}
}
