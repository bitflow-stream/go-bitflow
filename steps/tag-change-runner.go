package steps

import (
	"fmt"
	"strings"
	"sync"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

type TagChangeRunner struct {
	TagChangeListener
	Program        string
	Args           []string
	PreserveStdout bool

	commandWg sync.WaitGroup
}

func RegisterTagChangeRunner(b reg.ProcessorRegistry) {
	step := b.RegisterStep("run-on-tag-change", func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
		step := &TagChangeRunner{
			Program:        params["exec"].(string),
			Args:           params["args"].([]string),
			PreserveStdout: params["preserve-stdout"].(bool),
		}
		step.TagChangeListener.Callback = step
		step.TagChangeListener.ReadParameters(params)
		p.Add(step)
		return nil
	}, "Execute a program whenever values new values for a given tag are detected or expired. Parameters will be: [expired|updated] <tag> <all current tags>...").
		Required("exec", reg.String()).
		Optional("args", reg.List(reg.String()), []string{}).
		Optional("preserve-stdout", reg.Bool(), false)
	AddTagChangeListenerParams(step)
}

func (r *TagChangeRunner) String() string {
	return fmt.Sprintf("%v: Execute: %v %v", r.TagChangeListener.String(), r.Program, strings.Join(r.Args, " "))
}

func (r *TagChangeRunner) Expired(value string, allValues []string) error {
	r.run("expired", value, allValues)
	return nil
}

func (r *TagChangeRunner) Updated(value string, sample *bitflow.Sample, allValues []string) error {
	r.run("updated", value, allValues)
	return nil
}

func (r *TagChangeRunner) run(action string, updatedTag string, allValues []string) {
	args := append(r.Args, action, updatedTag)
	args = append(args, allValues...)

	// Wait for previous command to finish. Avoid spamming parallel commands.
	r.commandWg.Wait()

	// Build and start the command. Launch a goroutine that will log the result when the command is finished.
	command := &golib.Command{
		Program:        r.Program,
		Args:           args,
		ShortName:      fmt.Sprintf("%v tag %v=%v", action, r.Tag, updatedTag),
		PreserveStdout: r.PreserveStdout,
	}
	stopper := command.Start(&r.commandWg)
	if stopper.Stopped() {
		// Command failed to start
		log.Errorf("%v: Failed to launch command: %v", r, stopper.Err())
	} else {
		r.commandWg.Add(1)
		go r.logCommandExit(command, stopper)
		log.Printf("%v: Launching command: %v", r, command)
	}
}

func (r *TagChangeRunner) logCommandExit(command *golib.Command, stopper golib.StopChan) {
	defer r.commandWg.Done()
	stopper.Wait()
	if !command.Success() {
		log.Errorf("%v: Command finished: %v", r, command.StateString())
	}
}
