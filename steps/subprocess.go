package steps

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

type SubProcessRunner struct {
	bitflow.NoopProcessor
	Cmd  string
	Args []string

	// StderrPrefix enables immediate logging of each line of the stderr of the sub-process. These lines are
	// prefixed with the given string. If StderrPrefix is left empty, the output is instead collected in an in-memory
	// buffer, and logged when the process exits (regardless of the exit condition).
	StderrPrefix string

	// Configurations for the input/output marshalling

	Reader     bitflow.SampleReader
	Writer     bitflow.SampleWriter
	Marshaller bitflow.Marshaller

	cmd    *exec.Cmd
	output *bitflow.WriterSink
	input  *bitflow.ReaderSource

	stderr       bytes.Buffer
	stderrLogger *LoggingWriter
}

func RegisterExecutable(b reg.ProcessorRegistry, description string) error {
	descriptionParts := strings.Split(description, ";")
	const expectedParts = 3
	if len(descriptionParts) != expectedParts {
		return fmt.Errorf("Wrong format for external executable (have %v part(s)), need %v parts:"+
			" <short name>;<executable path>;<initial arguments>", len(descriptionParts), expectedParts)
	}
	name := descriptionParts[0]
	executablePath := descriptionParts[1]
	initialArgs := SplitShellCommand(descriptionParts[2])

	create := func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
		// Assemble the command line parameters: <initial args> <extra args> -step <step name> -args <step args>
		args := initialArgs
		stepName := params["step"].(string)
		stepArgs := golib.FormatSortedMap(params["args"].(map[string]string))
		extraArgs := params["exe-args"].([]string)
		args = append(args, extraArgs...)
		args = append(args, "-step", stepName, "-args", stepArgs)
		stderrPrefix := ""
		if !params["buffer-stderr"].(bool) {
			stderrPrefix = fmt.Sprintf("[%v/%v] ", name, stepName)
		}
		runner := &SubProcessRunner{
			Cmd:          executablePath,
			Args:         args,
			StderrPrefix: stderrPrefix,
		}

		format := params["format"].(string)
		factory, err := b.Endpoints.CloneWithParams(params["endpoint-config"].(map[string]string))
		if err != nil {
			return fmt.Errorf("Error parsing endpoint parameters: %v", err)
		}
		if err = runner.Configure(format, factory); err != nil {
			return err
		}
		p.Add(runner)
		return nil
	}
	b.RegisterStep(name, create,
		fmt.Sprintf("Start as a sub-process: %v\nInitial arguments: %v", executablePath, initialArgs)).
		Required("step", reg.String(), "The name of the specific step to be executed").
		Optional("args", reg.Map(reg.String()), map[string]string{}, "Arguments for the step").
		Optional("exe-args", reg.List(reg.String()), []string{}, "Extra command line arguments for the sub process").
		Optional("format", reg.String(), "bin", "Data marshalling format used for exchanging samples on the standard in/out streams").
		Optional("endpoint-config", reg.Map(reg.String()), map[string]string{}, "Extra configuration parameters for the endpoint factory").
		Optional("buffer-stderr", reg.Bool(), false, "Instead of always logging the stderr output of the sub-process, collect it and output it after the process exits")

	return nil
}

func RegisterSubProcessRunner(b reg.ProcessorRegistry) {
	create := func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
		cmd := SplitShellCommand(params["cmd"].(string))
		format := params["format"].(string)
		factory, err := b.Endpoints.CloneWithParams(params["endpoint-config"].(map[string]string))
		if err != nil {
			return fmt.Errorf("Error parsing endpoint parameters: %v", err)
		}

		runner := &SubProcessRunner{
			Cmd:  cmd[0],
			Args: cmd[1:],
		}
		if err = runner.Configure(format, factory); err != nil {
			return err
		}
		p.Add(runner)
		return nil
	}
	b.RegisterStep("sub-process", create,
		"Start a sub-process for processing samples. Samples will be sent/received over std in/out in the given format.").
		Required("cmd", reg.String()).
		Optional("format", reg.String(), "bin").
		Optional("endpoint-config", reg.Map(reg.String()), map[string]string{})
}

func (r *SubProcessRunner) Configure(marshallingFormat string, f *bitflow.EndpointFactory) error {
	format := bitflow.MarshallingFormat(marshallingFormat)
	var err error
	r.Marshaller, err = f.CreateMarshaller(format)
	if err != nil {
		return err
	}
	if r.Marshaller == nil {
		return fmt.Errorf("Unknown marshalling format: %v", marshallingFormat)
	}
	r.Reader = f.Reader(nil)
	r.Writer = f.Writer()
	return nil
}

func (r *SubProcessRunner) Start(wg *sync.WaitGroup) golib.StopChan {
	if err := r.createProcess(); err != nil {
		return golib.NewStoppedChan(err)
	}

	var tasks golib.TaskGroup
	if r.input != nil {
		// (Optionally) start the input first
		tasks.Add(&bitflow.SourceTaskWrapper{SampleSource: r.input})
	}
	tasks.Add(&golib.NoopTask{
		Description: "",
		Chan:        golib.WaitErrFunc(wg, r.runProcess),
	}, &bitflow.ProcessorTaskWrapper{SampleProcessor: r.output})

	channels := tasks.StartTasks(wg)
	return golib.WaitErrFunc(wg, func() error {
		golib.WaitForAny(channels)

		// Try to stop everything
		if r.input != nil {
			r.input.Close()
		}
		r.Close()

		err := tasks.CollectMultiError(channels)

		// After everything is shut down: forward the close call
		r.CloseSink()

		if err := err.NilOrError(); err != nil {
			return fmt.Errorf("Error in %v: %v", r, err)
		}
		return nil
	})
}

func (r *SubProcessRunner) createProcess() error {
	r.cmd = exec.Command(r.Cmd, r.Args...)
	if r.StderrPrefix != "" {
		r.stderrLogger = &LoggingWriter{Prefix: r.StderrPrefix}
		r.cmd.Stderr = r.stderrLogger
	} else {
		r.cmd.Stderr = &r.stderr
	}
	desc := r.String()

	writePipe, err := r.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("Failed to attach to standard input of subprocess: %v", err)
	}
	r.output = &bitflow.WriterSink{
		Output:      writePipe,
		Description: desc,
	}
	r.output.Writer = r.Writer
	r.output.Marshaller = r.Marshaller
	r.output.SetSink(new(bitflow.DroppingSampleProcessor))

	if _, isEmpty := r.GetSink().(*bitflow.DroppingSampleProcessor); r.GetSink() != nil && !isEmpty {
		readPipe, err := r.cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("Failed to attach to standard output of subprocess: %v", err)
		}
		r.input = &bitflow.ReaderSource{
			Input:       readPipe,
			Description: desc,
		}
		r.input.Reader = r.Reader
		r.input.SetSink(r.GetSink())
	} else {
		log.Printf("%v: Not parsing subprocess output", r)
	}
	return nil
}

func (r *SubProcessRunner) runProcess() error {
	err := r.cmd.Run()
	exitErr, isExitErr := err.(*exec.ExitError)
	success := err == nil || (isExitErr && exitErr.Success())

	if r.stderr.Len() > 0 {
		level := log.ErrorLevel
		if success {
			level = log.InfoLevel
		}
		log.StandardLogger().Logf(level, "Stderr output of %v:", r)
		scanner := bufio.NewScanner(&r.stderr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			log.StandardLogger().Logln(level, " > "+scanner.Text())
		}
	}
	if r.stderrLogger != nil {
		r.stderrLogger.Flush()
	}

	if success {
		return nil
	} else if isExitErr {
		return fmt.Errorf("Subprocess '%v' exited abnormally (%v)", r.Cmd, exitErr.ProcessState.String())
	} else if err != nil {
		return fmt.Errorf("Error executing subprocess: %v", err)
	} else {
		return nil
	}
}

func (r *SubProcessRunner) String() string {
	var args bytes.Buffer
	for _, arg := range r.Args {
		if !strings.ContainsRune(arg, ' ') {
			args.WriteString(" ")
			args.WriteString(arg)
		} else if strings.ContainsRune(arg, '"') {
			args.WriteString(" '")
			args.WriteString(arg)
			args.WriteString("'")
		} else {
			args.WriteString(" \"")
			args.WriteString(arg)
			args.WriteString("\"")
		}
	}
	return fmt.Sprintf("Subprocess [%v%s]", r.Cmd, args.String())
}

func (r *SubProcessRunner) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	return r.output.Sample(sample, header)
}

func (r *SubProcessRunner) Close() {
	r.output.Close()
	// TODO if the process won't close, try to kill it
	// r.cmd.Process.Kill()
}

type LoggingWriter struct {
	Prefix string
	buf    []byte
}

func (w *LoggingWriter) Write(p []byte) (int, error) {
	n := len(p)
	for len(p) > 0 {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			break
		}
		if len(w.buf) > 0 {
			log.Printf("%v%s%s\n", w.Prefix, w.buf, p[:i])
			w.buf = w.buf[0:0]
		} else {
			log.Printf("%v%s\n", w.Prefix, p[:i])
		}
		p = p[i+1:]
	}

	// Copy the unterminated line until the next Write
	w.buf = append(w.buf, p...)
	return n, nil
}

func (w *LoggingWriter) Flush() {
	if len(w.buf) > 0 {
		log.Printf("%s\n", w.buf)
		w.buf = w.buf[0:0]
	}
}
