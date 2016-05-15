package golib

import "sync"

type OneshotCondition struct {
	cond    *sync.Cond
	enabled bool
}

func NewOneshotCondition() *OneshotCondition {
	return &OneshotCondition{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (cond *OneshotCondition) Enable(perform func()) {
	if cond == nil || cond.cond == nil {
		return
	}
	cond.cond.L.Lock()
	defer cond.cond.L.Unlock()
	if cond.enabled {
		return
	}
	if perform != nil {
		perform()
	}
	cond.enabled = true
	cond.cond.Broadcast()
}

func (cond *OneshotCondition) EnableOnly() {
	cond.Enable(nil)
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
	cond.cond.Wait()
	// No need to check enabled flag afterwards
	// because the condition cannot be undone.
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
