package bitflow

import (
	"io"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/golib"
	"github.com/antongulenko/golib/gotermBox"
)

// TODO comment the methods and types in this file

type ConsoleBoxSink struct {
	AbstractMetricSink
	gotermBox.CliLogBox
	UpdateInterval time.Duration

	updateTask *golib.LoopTask
	lock       sync.Mutex
	lastSample *Sample
	lastHeader *Header
}

// Should be called as early as possible to intercept all log messages
func (sink *ConsoleBoxSink) Init() {
	sink.CliLogBox.Init()
	sink.RegisterMessageHook()
}

func (sink *ConsoleBoxSink) String() string {
	return "ConsoleBoxSink"
}

func (sink *ConsoleBoxSink) Start(wg *sync.WaitGroup) golib.StopChan {
	sink.InterceptLogger()
	log.Println("Printing samples to table")
	sink.updateTask = golib.NewErrLoopTask("", func(stop golib.StopChan) error {
		if err := sink.updateBox(); err != nil {
			return err
		}
		select {
		case <-time.After(sink.UpdateInterval):
		case <-stop:
		}
		return nil
	})
	sink.updateTask.StopHook = func() {
		sink.updateBox()
		sink.RestoreLogger()
	}
	return sink.updateTask.Start(wg)
}

func (sink *ConsoleBoxSink) updateBox() (err error) {
	sink.Update(func(out io.Writer, textWidth int) {
		sink.lock.Lock()
		sample := sink.lastSample
		header := sink.lastHeader
		sink.lock.Unlock()
		if sample != nil && header != nil {
			marshaller := &TextMarshaller{
				TextWidth: textWidth,
			}
			err = marshaller.WriteSample(sample, header, out)
		}
	})
	return
}

func (sink *ConsoleBoxSink) Close() {
	sink.updateTask.Stop()
}

func (sink *ConsoleBoxSink) Sample(sample *Sample, header *Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	sink.lock.Lock()
	sink.lastSample = sample
	sink.lastHeader = header
	sink.lock.Unlock()
	return nil
}
