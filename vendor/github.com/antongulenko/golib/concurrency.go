package golib

import "time"

type Stopper struct {
	stopped   chan bool
	isStopped bool
	num       int
}

func NewStopper() *Stopper {
	return NewStopperLen(1)
}

func NewStopperLen(parallelStops int) *Stopper {
	return &Stopper{
		stopped: make(chan bool, parallelStops),
		num:     parallelStops,
	}
}

func (s *Stopper) Stop() {
	s.isStopped = true
	for i := 0; i < s.num; i++ {
		select {
		case s.stopped <- true:
		default:
			return
		}
	}
}

func (s *Stopper) Stopped(timeout time.Duration) bool {
	if s.IsStopped() {
		return true
	}
	select {
	case <-time.After(timeout):
		return false
	case <-s.stopped:
		s.Stop()
		return true
	}
}

func (s *Stopper) IsStopped() bool {
	return s.isStopped
}

// Whan receiving from this, always call Stop() afterwards!
func (s *Stopper) Wait() <-chan bool {
	return s.stopped
}
