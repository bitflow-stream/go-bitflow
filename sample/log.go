// test doc
package sample

import (
	"flag"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/chris-garrett/lfshook"

	"github.com/antongulenko/golib"
)

var (
	logFile      string
	logVerbose   bool
	logQuiet     bool
	logVeryQuiet bool
)

func init() {
	// Configure logging output
	log.SetOutput(os.Stderr)
	log.SetFormatter(newLogFormatter())
	golib.Log = log.StandardLogger()

	flag.BoolVar(&logVerbose, "v", false, "Enable verbose logging output")
	flag.BoolVar(&logQuiet, "q", false, "Suppress logging output (except warnings and errors)")
	flag.BoolVar(&logVeryQuiet, "qq", false, "Suppress logging output (except errors)")
	flag.StringVar(&logFile, "log", "", "Redirect logs to a given file in addition to the console.")
}

// ConfigureLogging configures the logger based on the command-line flags defined in this package.
// This function should be called early in every main package, preferably before any prior logging output.
func ConfigureLogging() {
	level := log.InfoLevel
	if logVerbose {
		level = log.DebugLevel
	} else if logVeryQuiet {
		level = log.ErrorLevel
	} else if logQuiet {
		level = log.WarnLevel
	}
	log.SetLevel(level)
	if logFile != "" {
		pathmap := make(lfshook.PathMap)
		for i := 0; i < 256; i++ {
			pathmap[log.Level(i)] = logFile
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
