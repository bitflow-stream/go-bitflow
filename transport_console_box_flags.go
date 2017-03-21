package bitflow

import (
	"time"

	"flag"
	"fmt"

	"sync"

	"github.com/antongulenko/golib/gotermBox"
)

const ConsoleBoxEndpoint = EndpointType("box")

var (
	ConsoleBoxSettings = gotermBox.CliLogBox{
		NoUtf8:        false,
		LogLines:      10,
		MessageBuffer: 500,
	}
	ConsoleBoxUpdateInterval    = 500 * time.Millisecond
	ConsoleBoxMinUpdateInterval = 50 * time.Millisecond

	// console_box_testMode is a flag used by tests to suppress initialization routines
	// that are not testable. It is a hack to keep the EndpointFactory easy to use
	// while making it testable.
	console_box_testMode bool

	registerConsoleBoxOnce sync.Once
)

func RegisterConsoleBoxOutput() {
	registerConsoleBoxOnce.Do(func() {
		var factory consoleBoxFactory
		CustomDataSinks[ConsoleBoxEndpoint] = factory.createConsoleBox
		CustomOutputFlags = append(CustomOutputFlags, factory.registerFlags)
	})
}

type consoleBoxFactory struct {
	ConsoleBoxNoImmediateScreenUpdate bool
}

func (factory *consoleBoxFactory) registerFlags(f *flag.FlagSet) {
	f.BoolVar(&factory.ConsoleBoxNoImmediateScreenUpdate, "slow-screen-updates", false, fmt.Sprintf("For console box output, don't update the screen on every sample, but only in intervals of %v", ConsoleBoxUpdateInterval))
}

func (factory *consoleBoxFactory) createConsoleBox(target string) (MetricSink, error) {
	if target != stdTransportTarget {
		return nil, fmt.Errorf("Transport '%v' can only be defined with target '%v'", ConsoleBoxEndpoint, stdTransportTarget)
	}
	sink := &ConsoleBoxSink{
		CliLogBoxTask: gotermBox.CliLogBoxTask{
			CliLogBox:         ConsoleBoxSettings,
			UpdateInterval:    ConsoleBoxUpdateInterval,
			MinUpdateInterval: ConsoleBoxMinUpdateInterval,
		},
		ImmediateScreenUpdate: !factory.ConsoleBoxNoImmediateScreenUpdate,
	}
	if !console_box_testMode {
		sink.Init()
	}
	return sink, nil
}
