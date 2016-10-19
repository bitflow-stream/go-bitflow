package golib

import (
	"errors"
	"net"
	"sync"
)

func FirstIpAddress() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			// Loopback and disabled interfaces are not interesting
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				return v.IP, nil
			case *net.IPAddr:
				return v.IP, nil
			}
		}
	}
	return nil, errors.New("No valid network interfaces found")
}

// ==================== TCP listener task ====================
type TCPConnectionHandler func(wg *sync.WaitGroup, conn *net.TCPConn)

type TCPListenerTask struct {
	Handler        TCPConnectionHandler
	ListenEndpoint string
	StopHook       func()

	loopTask *LoopTask
	listener *net.TCPListener
}

func (task *TCPListenerTask) String() string {
	return "TCP listener " + task.ListenEndpoint
}

func (task *TCPListenerTask) Start(wg *sync.WaitGroup) StopChan {
	return task.ExtendedStart(nil, wg)
}

func (task *TCPListenerTask) ExtendedStart(start func(addr net.Addr), wg *sync.WaitGroup) StopChan {
	hook := task.StopHook
	defer func() {
		if hook != nil {
			hook()
		}
	}()

	endpoint, err := net.ResolveTCPAddr("tcp", task.ListenEndpoint)
	if err != nil {
		return TaskFinishedError(err)
	}
	task.listener, err = net.ListenTCP("tcp", endpoint)
	if err != nil {
		return TaskFinishedError(err)
	}
	if start != nil {
		start(task.listener.Addr())
	}
	hook = nil
	task.loopTask = task.listen(wg)
	return task.loopTask.Start(wg)
}

func (task *TCPListenerTask) listen(wg *sync.WaitGroup) *LoopTask {
	loop := NewLoopTask("tcp listener", func(_ StopChan) {
		if listener := task.listener; listener == nil {
			return
		} else {
			conn, err := listener.AcceptTCP()
			if err != nil {
				if task.listener != nil {
					Log.Errorln("Error accepting connection:", err)
				}
			} else {
				task.loopTask.IfElseEnabled(func() {
					_ = conn.Close() // Drop error
				}, func() {
					task.Handler(wg, conn)
				})
			}
		}
	})
	loop.StopHook = task.StopHook
	return loop
}

func (task *TCPListenerTask) Stop() {
	task.ExtendedStop(nil)
}

func (task *TCPListenerTask) ExtendedStop(stop func()) {
	if task.loopTask != nil {
		task.loopTask.Enable(func() {
			if listener := task.listener; listener != nil {
				task.listener = nil  // Will be checked when returning from AcceptTCP()
				_ = listener.Close() // Drop error
			}
			if stop != nil {
				stop()
			}
		})
	}
}

func (task *TCPListenerTask) IfRunning(do func()) {
	task.loopTask.IfNotEnabled(do)
}
