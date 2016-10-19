package golib

import "sync"

type OneshotCondition struct {
	cond    *sync.Cond
	enabled bool
	Err     error
}

func NewOneshotCondition() *OneshotCondition {
	return &OneshotCondition{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (cond *OneshotCondition) EnableErrFunc(perform func() error) {
	if cond == nil || cond.cond == nil {
		return
	}
	cond.cond.L.Lock()
	defer cond.cond.L.Unlock()
	if cond.enabled {
		return
	}
	if perform != nil {
		cond.Err = perform()
	}
	cond.enabled = true
	cond.cond.Broadcast()
}

func (cond *OneshotCondition) Enable(perform func()) {
	cond.EnableErrFunc(func() error {
		if perform != nil {
			perform()
		}
		return nil
	})
}

func (cond *OneshotCondition) EnableErr(err error) {
	cond.EnableErrFunc(func() error {
		return err
	})
}

func (cond *OneshotCondition) EnableOnly() {
	cond.EnableErrFunc(nil)
}

func (cond *OneshotCondition) Enabled() bool {
	if cond == nil || cond.cond == nil {
		return false
	}
	cond.cond.L.Lock()
	defer cond.cond.L.Unlock()
	return cond.enabled
}

func (cond *OneshotCondition) Wait() {
	if cond == nil || cond.cond == nil {
		return
	}
	cond.cond.L.Lock()
	defer cond.cond.L.Unlock()
	for !cond.enabled {
		cond.cond.Wait()
	}
}

func (cond *OneshotCondition) IfEnabled(execute func()) {
	cond.cond.L.Lock()
	defer cond.cond.L.Unlock()
	if !cond.enabled {
		return
	}
	execute()
}

func (cond *OneshotCondition) IfNotEnabled(execute func()) {
	cond.cond.L.Lock()
	defer cond.cond.L.Unlock()
	if cond.enabled {
		return
	}
	execute()
}

func (cond *OneshotCondition) IfElseEnabled(enabled func(), disabled func()) {
	cond.cond.L.Lock()
	defer cond.cond.L.Unlock()
	if cond.enabled {
		enabled()
	} else {
		disabled()
	}
}

// ===== Implement Task interface

func (cond *OneshotCondition) Stop() {
	cond.EnableOnly()
}

func (cond *OneshotCondition) Start(wg *sync.WaitGroup) StopChan {
	return WaitCondition(wg, cond)
}
