package golib

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type Command struct {
	Name    string
	Proc    *os.Process
	Logfile string

	State    *os.ProcessState
	StateErr error

	processFinished *OneshotCondition
}

func openLogfile(dirname, filename string) (*os.File, error) {
	err := os.MkdirAll(dirname, os.FileMode(0775))
	if err != nil {
		return nil, err
	}
	logfile, err := ioutil.TempFile(dirname, filename)
	if err != nil {
		return nil, err
	}
	err = logfile.Truncate(0)
	if err != nil {
		return nil, err
	}
	return logfile, nil
}

func StartCommand(prog string, args []string, shortName, logdir, logfile string) (*Command, error) {
	cmd := exec.Command(prog, args...)
	if logfile != "" && logdir != "" {
		logF, err := openLogfile(logdir, logfile)
		if err != nil {
			return nil, err
		}
		logfile = logF.Name()
		cmd.Stdout = logF
		cmd.Stderr = logF
	} else {
		logfile = ""
		cmd.Stdout = nil
		cmd.Stderr = nil
	}

	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	command := &Command{
		Name:            shortName,
		processFinished: NewOneshotCondition(),
		Proc:            cmd.Process,
		Logfile:         logfile,
	}
	go command.waitForProcess()
	return command, nil
}

func (command *Command) waitForProcess() {
	state, err := command.Proc.Wait()
	if state == nil && err == nil {
		err = fmt.Errorf("No ProcState returned")
	}
	command.processFinished.Enable(func() {
		command.State, command.StateErr = state, err
	})
}

func (command *Command) IsFinished() bool {
	if err := command.checkStarted(); err != nil {
		return false
	}
	return command.processFinished.Enabled()
}

func (command *Command) checkStarted() error {
	if command == nil || command.Proc == nil {
		return fmt.Errorf("Command is nil")
	}
	return nil
}

func (command *Command) Stop() {
	if err := command.checkStarted(); err != nil {
		return
	}
	command.Proc.Signal(syscall.SIGHUP)
}

func (command *Command) Success() bool {
	return command.StateErr != nil || (command.State != nil && command.State.Success())
}

func (command *Command) StateString() string {
	if err := command.checkStarted(); err != nil {
		return err.Error()
	}
	if !command.IsFinished() {
		return fmt.Sprintf("%v (%v) running", command.Name, command.Proc.Pid)
	}
	if command.State == nil {
		return fmt.Sprintf("%v wait error: %s", command.Name, command.StateErr)
	} else {
		if command.State.Success() {
			return fmt.Sprintf("%v (%v) successful exit", command.Name, command.Proc.Pid)
		} else {
			return fmt.Sprintf("%v (%v) exit: %s", command.Name, command.Proc.Pid, command.State.String())
		}
	}
}

// ===== Implement Task interface

func (command *Command) String() string {
	return command.StateString() + " (" + command.Logfile + ")"
}

func (command *Command) Start(wg *sync.WaitGroup) StopChan {
	if err := command.checkStarted(); err != nil {
		return nil
	}
	return command.processFinished.Start(wg)
}
