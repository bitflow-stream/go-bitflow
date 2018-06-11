package pipeline

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

type SubprocessRunner struct {
	bitflow.NoopProcessor
	Cmd  string
	Args []string

	// Configurations for the input/output marshalling

	Reader     bitflow.SampleReader
	Writer     bitflow.SampleWriter
	Marshaller bitflow.Marshaller

	cmd    *exec.Cmd
	output *bitflow.WriterSink
	input  *bitflow.ReaderSource
	stderr bytes.Buffer
}

func (r *SubprocessRunner) Configure(marshallingFormat string, f *bitflow.EndpointFactory) error {
	r.Marshaller = bitflow.MarshallingFormat(marshallingFormat).Marshaller()
	if r.Marshaller == nil {
		return fmt.Errorf("Unknown marshalling format: %v", marshallingFormat)
	}
	r.Reader = f.Reader(nil)
	r.Writer = f.Writer()
	return nil
}

func (r *SubprocessRunner) Start(wg *sync.WaitGroup) golib.StopChan {
	if err := r.createProcess(); err != nil {
		return golib.NewStoppedChan(err)
	}

	var tasks golib.TaskGroup
	if r.input != nil {
		// (Optionally) start the input first
		tasks.Add(&bitflow.SourceTaskWrapper{r.input})
	}
	tasks.Add(&golib.NoopTask{
		Description: "",
		Chan:        golib.WaitErrFunc(wg, r.runProcess),
	}, &bitflow.ProcessorTaskWrapper{r.output})

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
		return err.NilOrError()
	})
}

func (r *SubprocessRunner) createProcess() error {
	r.cmd = exec.Command(r.Cmd, r.Args...)
	r.cmd.Stderr = &r.stderr
	desc := r.String()

	writePipe, err := r.cmd.StdinPipe()
	if err != nil {
		return err
	}
	r.output = &bitflow.WriterSink{
		Output:      writePipe,
		Description: desc,
	}
	r.output.Writer = r.Writer
	r.output.Marshaller = r.Marshaller

	if _, isEmpty := r.OutgoingSink.(*bitflow.DroppingSampleProcessor); r.OutgoingSink != nil && !isEmpty {
		readPipe, err := r.cmd.StdoutPipe()
		if err != nil {
			return err
		}
		r.input = &bitflow.ReaderSource{
			Input:       readPipe,
			Description: desc,
		}
		r.input.Reader = r.Reader
		r.input.SetSink(r.OutgoingSink)
	} else {
		log.Printf("%v: Not parsing subprocess output", r)
	}
	return nil
}

func (r *SubprocessRunner) runProcess() error {
	err := r.cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.Success() {
				err = nil
			} else {
				if r.stderr.Len() > 0 {
					log.Warnf("Stderr output of %v:", r)
					scanner := bufio.NewScanner(&r.stderr)
					scanner.Split(bufio.ScanLines)
					for scanner.Scan() {
						log.Warnln(" > " + scanner.Text())
					}
				}
				return fmt.Errorf("Subprocess '%v' exited abnormally (%v)", r.Cmd, exitErr.ProcessState.String())
			}
		}
	}
	return err
}

func (r *SubprocessRunner) String() string {
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

func (r *SubprocessRunner) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	return r.output.Sample(sample, header)
}

func (r *SubprocessRunner) Close() {
	r.output.Close()
	// TODO if the process won't close, try to kill it
	// r.cmd.Process.Kill()
}
