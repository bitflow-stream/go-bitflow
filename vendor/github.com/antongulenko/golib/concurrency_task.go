package golib

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"reflect"
	"runtime/pprof"
	"sync"
	"time"
)

// ========= Task interface
type StopChan <-chan error

// Semantics: Start() may only be called once, but Stop() should be idempotent.
// One error must be sent on StopChan upon stopping. The error can be nil.
type Task interface {
	Start(wg *sync.WaitGroup) StopChan
	Stop()
	String() string // Tasks are frequently printed
}

type NoopTask struct {
	Chan        chan error
	Description string
}

func (task *NoopTask) Start(*sync.WaitGroup) StopChan {
	return task.Chan
}
func (task *NoopTask) Stop() {
	task.Chan <- nil
}
func (task *NoopTask) String() string {
	return fmt.Sprintf("Task(%s)", task.Description)
}

type CleanupTask struct {
	Cleanup     func()
	Description string
	once        sync.Once
}

func (task *CleanupTask) Start(*sync.WaitGroup) StopChan {
	return nil
}
func (task *CleanupTask) Stop() {
	task.once.Do(func() {
		if cleanup := task.Cleanup; cleanup != nil {
			cleanup()
		}
	})
}
func (task *CleanupTask) String() string {
	return fmt.Sprintf("Cleanup(%s)", task.Description)
}

type LoopTask struct {
	*OneshotCondition
	loop        func(stop StopChan) error
	Err         error
	Description string
	StopHook    func()
}

func (task *LoopTask) Start(wg *sync.WaitGroup) StopChan {
	cond := task.OneshotCondition
	if loop := task.loop; loop != nil {
		stop := WaitCondition(wg, cond)
		if wg != nil {
			wg.Add(1)
		}
		go func() {
			if wg != nil {
				defer wg.Done()
			}
			if hook := task.StopHook; hook != nil {
				defer hook()
			}
			for !cond.Enabled() {
				task.Err = loop(stop)
				if task.Err != nil {
					cond.EnableErr(task.Err)
				}
			}
		}()
	}
	return cond.Start(wg)
}

func (task *LoopTask) String() string {
	return fmt.Sprintf("LoopTask(%s)", task.Description)
}

func NewLoopTask(description string, loop func(stop StopChan)) *LoopTask {
	return NewErrLoopTask(description, func(stop StopChan) error {
		loop(stop)
		return nil
	})
}

func NewErrLoopTask(description string, loop func(stop StopChan) error) *LoopTask {
	return &LoopTask{
		OneshotCondition: NewOneshotCondition(),
		loop:             loop,
		Description:      description,
	}
}

// ========= Helpers to implement Task interface

func TaskFinished() StopChan {
	return TaskFinishedError(nil)
}

func TaskFinishedError(err error) StopChan {
	res := make(chan error, 1)
	res <- err
	return res
}

func WaitErrFunc(wg *sync.WaitGroup, wait func() error) StopChan {
	if wg != nil {
		wg.Add(1)
	}
	finished := make(chan error, 1)
	go func() {
		if wg != nil {
			defer wg.Done()
		}
		var err error
		if wait != nil {
			err = wait()
		}
		finished <- err
		close(finished)
	}()
	return finished
}

func WaitFunc(wg *sync.WaitGroup, wait func()) StopChan {
	return WaitErrFunc(wg, func() error {
		wait()
		return nil
	})
}

func WaitCondition(wg *sync.WaitGroup, cond *OneshotCondition) StopChan {
	if cond == nil {
		return nil
	}
	return WaitErrFunc(wg, func() error {
		cond.Wait()
		return cond.Err
	})
}

func WaitForAny(channels []StopChan) (int, error) {
	if len(channels) < 1 {
		return -1, nil
	}
	// Use reflect package to wait for any of the given channels
	var cases []reflect.SelectCase
	for _, ch := range channels {
		if ch != nil {
			refCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
			cases = append(cases, refCase)
		}
	}
	choice, result, _ := reflect.Select(cases)
	channels[choice] = nil // Already received
	if err, ok := result.Interface().(error); ok {
		return choice, err
	} else {
		return choice, nil
	}
}

func StartTasks(wg *sync.WaitGroup, tasks []Task) ([]Task, []StopChan) {
	channels := make([]StopChan, 0, len(tasks))
	waitingTasks := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		if channel := task.Start(wg); channel != nil {
			channels = append(channels, channel)
			waitingTasks = append(waitingTasks, task)
		}
	}
	return waitingTasks, channels
}

func WaitForAnyTask(wg *sync.WaitGroup, tasks []Task) (Task, error, []Task, []StopChan) {
	waitingTasks, channels := StartTasks(wg, tasks)
	choice, err := WaitForAny(channels)
	return waitingTasks[choice], err, waitingTasks, channels
}

func WaitForSetup(wg *sync.WaitGroup, setup func() error) StopChan {
	if wg != nil {
		wg.Add(1)
	}
	failed := make(chan error, 1)
	go func() {
		if wg != nil {
			defer wg.Done()
		}
		if setup != nil {
			if err := setup(); err != nil {
				failed <- err
				close(failed)
			}
		}
	}()
	return failed
}

// ========= Task Group

type TaskGroup struct {
	names  []string          // Track order of added new groups
	groups map[string][]Task // Groups will be stopped sequentially, but Tasks in one group in parallel
	all    []Task
}

func NewTaskGroup(tasks ...Task) *TaskGroup {
	group := &TaskGroup{
		groups: make(map[string][]Task),
	}
	for _, o := range tasks {
		group.Add(o)
	}
	return group
}

func (group *TaskGroup) AllTasks() []Task {
	return group.all
}

func (group *TaskGroup) Add(tasks ...Task) {
	group.AddNamed("default", tasks...)
}

func (group *TaskGroup) AddNamed(name string, tasks ...Task) {
	var list []Task
	if existingList, ok := group.groups[name]; ok {
		list = existingList
	} else {
		group.names = append(group.names, name)
	}
	for _, task := range tasks {
		if task != nil {
			group.all = append(group.all, tasks...)
			list = append(list, task)
		}
	}
	group.groups[name] = list
}

func (group *TaskGroup) StartTasks(wg *sync.WaitGroup) ([]Task, []StopChan) {
	return StartTasks(wg, group.all)
}

func (group *TaskGroup) WaitForAny(wg *sync.WaitGroup) (Task, error, []Task, []StopChan) {
	return WaitForAnyTask(wg, group.all)
}

func (group *TaskGroup) ReverseStop(printTasks bool) {
	for i := len(group.names) - 1; i >= 0; i-- {
		// Stop groups in reverse order
		var wg sync.WaitGroup
		tasks := group.groups[group.names[i]]
		for _, task := range tasks {
			// Stop tasks in one group in parallel
			wg.Add(1)
			go func(task Task) {
				defer wg.Done()
				if printTasks {
					Log.Println("Stopping", task)
				}
				task.Stop()
			}(task)
		}
		wg.Wait()
	}
}

func PrintErrors(inputs []StopChan, tasks []Task, printWait bool) (numErrors int) {
	for i, input := range inputs {
		if input != nil {
			if printWait {
				task := tasks[i]
				Log.Println("Waiting for", task)
			}
			if err := <-input; err != nil {
				numErrors++
				Log.Errorln(err)
			}
		}
	}
	return
}

func (group *TaskGroup) WaitAndStop(timeout time.Duration, printWait bool) (Task, int) {
	var wg sync.WaitGroup
	numErrors := 0
	reason, err, tasks, channels := group.WaitForAny(&wg)
	if err != nil {
		numErrors++
		Log.Errorln(err)
	}
	if timeout > 0 {
		time.AfterFunc(timeout, func() {
			pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
			panic("Waiting for stopping goroutines timed out")
		})
	}
	group.ReverseStop(printWait)
	wg.Wait()
	numErrors += PrintErrors(channels, tasks, printWait)
	return reason, numErrors
}

var (
	DefaultTaskStopTimeout   = time.Duration(0)
	DefaultPrintTaskStopWait = false
)

func init() {
	flag.BoolVar(&DefaultPrintTaskStopWait, "task_stop_print", DefaultPrintTaskStopWait, "Print tasks waited for when stopping (for debugging)")
	flag.DurationVar(&DefaultTaskStopTimeout, "task_stop_timeout", DefaultTaskStopTimeout, "Timeout duration when stopping and waiting for tasks to finish")
}

func (group *TaskGroup) PrintWaitAndStop() int {
	return group.TimeoutPrintWaitAndStop(DefaultTaskStopTimeout, DefaultPrintTaskStopWait)
}

func (group *TaskGroup) TimeoutPrintWaitAndStop(timeout time.Duration, printWait bool) (numErrors int) {
	reason, numErrors := group.WaitAndStop(timeout, printWait)
	Log.Debugln("Stopped because of", reason)
	return
}

// ========= Sources of interrupts by the user

func ExternalInterrupt() chan error {
	// This must be done after starting any subprocess that depends
	// the ignore-handler for SIGNIT provided by ./noint
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	stop := make(chan error, 2)
	go func() {
		defer signal.Stop(interrupt)
		select {
		case <-interrupt:
		case <-stop:
		}
		stop <- nil
	}()
	return stop
}

func UserInput() chan error {
	userinput := make(chan error, 2)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		_, err := reader.ReadString('\n')
		if err != nil {
			err = fmt.Errorf("Error reading user input: %v", err)
		}
		userinput <- err
	}()
	return userinput
}

func StdinClosed() chan error {
	closed := make(chan error, 2)
	go func() {
		_, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			err = fmt.Errorf("Error reading stdin: %v", err)
		}
		closed <- err
	}()
	return closed
}
