package golib

import (
	"log"
	"os"
)

var (
	ErrorExitHook    func()
	checkerr_exiting bool
)

func Checkerr(err error) {
	if err != nil {
		if checkerr_exiting {
			log.Println("Recursive Checkerr:", err)
			return
		}
		checkerr_exiting = true
		log.Println("Fatal Error:", err)
		if ErrorExitHook != nil {
			ErrorExitHook()
		}
		os.Exit(1)
	}
}

func Printerr(err error) {
	if err != nil {
		log.Println("Error:", err)
	}
}
