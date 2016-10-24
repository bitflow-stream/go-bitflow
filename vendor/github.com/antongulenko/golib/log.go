package golib

import (
	"flag"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/chris-garrett/lfshook"
)

var (
	LogFile      string
	LogVerbose   bool
	LogQuiet     bool
	LogVeryQuiet bool

	// Package-wide logger, can be configured or disabled.
	Log = &log.Logger{
		Out: os.Stderr,
		Formatter: &log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.StampMilli,
		},
		Hooks: make(log.LevelHooks),
		Level: log.DebugLevel,
	}
)

func init() {
	// Configure logging output
	log.SetOutput(os.Stderr)
	log.SetFormatter(newLogFormatter())
	Log = log.StandardLogger()
}

// RegisterLogFlags registers flags for changing variables that will controll
// the log level and other logging parameters when calling ConfigureLogging().
func RegisterLogFlags() {
	flag.BoolVar(&LogVerbose, "v", false, "Enable verbose logging output")
	flag.BoolVar(&LogQuiet, "q", false, "Suppress logging output (except warnings and errors)")
	flag.BoolVar(&LogVeryQuiet, "qq", false, "Suppress logging output (except errors)")
	flag.StringVar(&LogFile, "log", "", "Redirect logs to a given file in addition to the console.")
}

// ConfigureLogging configures the logger based on the command-line flags defined in RegisterLogFlags.
// This function should be called early in every main package, preferably before any prior logging output,
// but after calling RegisterLogFlags() and flag.Parse().
func ConfigureLogging() {
	level := log.InfoLevel
	if LogVerbose {
		level = log.DebugLevel
	} else if LogVeryQuiet {
		level = log.ErrorLevel
	} else if LogQuiet {
		level = log.WarnLevel
	}
	log.SetLevel(level)
	if LogFile != "" {
		pathmap := make(lfshook.PathMap)
		for i := 0; i < 256; i++ {
			pathmap[log.Level(i)] = LogFile
		}
		hook := lfshook.NewHook(pathmap)
		formatter := newLogFormatter()
		hook.SetFormatter(formatter)
		// HACK: force the formatter to use colored output in the file
		formatter.DisableColors = false
		formatter.ForceColors = true
		log.AddHook(hook)
	}
}

func newLogFormatter() *log.TextFormatter {
	return &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.StampMilli,
	}
}
