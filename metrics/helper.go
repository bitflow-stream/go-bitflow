package metrics

import "time"

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
