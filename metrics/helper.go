package metrics

import (
	"bytes"
	"fmt"
	"time"
)

type Stopper struct {
	stopped   chan bool
	num       int
	isStopped bool
}

func NewStopper(numRoutines int) *Stopper {
	return &Stopper{
		stopped: make(chan bool, numRoutines),
		num:     numRoutines,
	}
}

func (s *Stopper) Stop() {
	s.isStopped = true
	for i := 0; i < s.num; i++ {
		s.stopped <- true
	}
}

func (s *Stopper) Stopped(timeout time.Duration) bool {
	select {
	case <-time.After(timeout):
		return false
	case <-s.stopped:
		return true
	}
}

func (s *Stopper) IsStopped() bool {
	return s.isStopped
}

type MultiError []error

func (err MultiError) NilOrError() error {
	if len(err) == 0 {
		return nil
	} else if len(err) == 1 {
		return err[0]
	}
	return err
}

func (err *MultiError) Add(errOrNil error) {
	if err != nil && errOrNil != nil {
		*err = append(*err, errOrNil)
	}
}

func (err MultiError) Error() string {
	switch len(err) {
	case 0:
		return "No error"
	case 1:
		return err[0].Error()
	default:
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "Multiple errors:\n")
		for i, e := range err {
			fmt.Fprintf(&buf, "\t%v. %v\n", i+1, e)
		}
		return buf.String()
	}
}
