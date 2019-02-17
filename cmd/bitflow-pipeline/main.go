package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/cmd"
)

const (
	fileFlag            = "f"
	BitflowScriptSuffix = ".bf"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <flags> <bitflow script>\nAll flags must be defined before the first non-flag parameter.\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	fix_arguments(&os.Args)
	os.Exit(do_main())
}

func fix_arguments(argsPtr *[]string) {
	if n := golib.ParseHashbangArgs(argsPtr); n > 0 {
		// Insert -f before the script file, if necessary
		args := *argsPtr
		if args[n-1] != "-"+fileFlag && args[n-1] != "--"+fileFlag {
			args = append(args, "") // Extend by one entry
			copy(args[n+1:], args[n:])
			args[n] = "-" + fileFlag
			*argsPtr = args
		}
	}
}

func do_main() int {
	var builder cmd.CmdPipelineBuilder
	scriptFile := ""
	flag.StringVar(&scriptFile, fileFlag, "", "File to read a Bitflow script from (alternative to providing the script on the command line)")
	builder.RegisterFlags()
	_, args := cmd.ParseFlags()
	rawScript, err := get_script(args, scriptFile)
	golib.Checkerr(err)

	pipe, err := builder.BuildPipeline(rawScript)
	golib.Checkerr(err)
	pipe = builder.PrintPipeline(pipe)
	if pipe == nil {
		return 0
	}
	defer golib.ProfileCpu()()
	return pipe.StartAndWait()
}

func get_script(parsedArgs []string, scriptFile string) (string, error) {
	if scriptFile != "" && len(parsedArgs) > 0 {
		return "", errors.New("Please provide a bitflow pipeline script either via -f or as parameter, not both.")
	}
	if len(parsedArgs) == 1 && strings.HasSuffix(parsedArgs[0], BitflowScriptSuffix) {
		// Special case when passing a single existing .bf file as positional argument: Treat it as a script file
		info, err := os.Stat(parsedArgs[0])
		if err == nil && info.Mode().IsRegular() {
			scriptFile = parsedArgs[0]
		}
	}
	var rawScript string
	if scriptFile != "" {
		scriptBytes, err := ioutil.ReadFile(scriptFile)
		if err != nil {
			return "", fmt.Errorf("Error reading bitflow script file %v: %v", scriptFile, err)
		}
		rawScript = string(scriptBytes)
	} else {
		rawScript = strings.TrimSpace(strings.Join(parsedArgs, " "))
	}
	if rawScript == "" {
		return "", errors.New("Please provide a bitflow pipeline script via -f or directly as parameter.")
	}
	return rawScript, nil
}
