package metrics

import (
	"log"
	"net"
	"sync"

	"github.com/antongulenko/golib"
)

// ==================== TCP listener source ====================
type TCPListenerSource struct {
	unmarshallingMetricSource
	ListenEndpoint string
	loopTask       golib.Task
	listener       *net.TCPListener
	conn           *net.TCPConn
	stopped        *golib.OneshotCondition
}

func (source *TCPListenerSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.stopped = golib.NewOneshotCondition()
	endpoint, err := net.ResolveTCPAddr("tcp", source.ListenEndpoint)
	if err != nil {
		return golib.TaskFinishedError(err)
	}
	source.listener, err = net.ListenTCP("tcp", endpoint)
	if err != nil {
		return golib.TaskFinishedError(err)
	}
	source.loopTask = source.listen(wg)
	return source.loopTask.Start(wg)
}

func (source *TCPListenerSource) listen(wg *sync.WaitGroup) golib.Task {
	log.Println("Listening for incoming", source.Unmarshaller, "samples on", source.listener.Addr())
	return golib.LoopTask(func(_ golib.StopChan) {
		conn, err := source.listener.AcceptTCP()
		if err != nil {
			log.Println("Error accepting connection:", err)
		} else {
			if existing := source.conn; existing != nil {
				log.Println("Rejecting connection from %v, already connected to %v\n", conn.RemoteAddr(), existing.RemoteAddr())
				if err := conn.Close(); err != nil {
					log.Println("Error closing connection:", err)
				}
				return
			}
			source.stopped.IfElseEnabled(func() {
				_ = conn.Close() // Drop error
			}, func() {
				log.Println("Accepted connection from", conn.RemoteAddr())
				source.conn = conn
				wg.Add(1)
				go func() {
					defer func() {
						wg.Done()
					}()
					tcpReadSamples(conn, source.Unmarshaller, source.Sink)
				}()
			})
		}
	})
}

func (source *TCPListenerSource) Stop() {
	source.stopped.Enable(func() {
		source.loopTask.Stop()
		if listener := source.listener; listener != nil {
			_ = listener.Close() // Drop error
		}
		if conn := source.conn; conn != nil {
			_ = conn.Close() // Drop error
		}
	})
}

// ==================== TCP listener sink ====================
type TCPListenerSink struct {
	abstractSink
	Endpoint    string
	connections map[*tcpWriteConn]bool
	listener    *net.TCPListener
	loopTask    golib.Task
	stopped     *golib.OneshotCondition
}

func (sink *TCPListenerSink) Start(wg *sync.WaitGroup) golib.StopChan {
	if sink.connections == nil {
		sink.connections = make(map[*tcpWriteConn]bool)
	}
	sink.stopped = golib.NewOneshotCondition()

	addr, err := net.ResolveTCPAddr("tcp", sink.Endpoint)
	if err != nil {
		return golib.TaskFinishedError(err)
	}
	sink.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return golib.TaskFinishedError(err)
	}
	sink.loopTask = sink.listen(wg)
	return sink.loopTask.Start(wg)
}

func (sink *TCPListenerSink) Stop() {
	sink.stopped.Enable(func() {
		sink.loopTask.Stop()
		if listener := sink.listener; listener != nil {
			_ = listener.Close()
		}
		for conn := range sink.connections {
			conn.Stop()
		}
	})
}

func (sink *TCPListenerSink) listen(wg *sync.WaitGroup) golib.Task {
	log.Println("Listening for", sink.marshaller, "sample output connections on", sink.listener.Addr())
	return golib.LoopTask(func(_ golib.StopChan) {
		conn, err := sink.listener.AcceptTCP()
		if err != nil {
			log.Println("Error accepting connection:", err)
		} else {
			sink.stopped.IfElseEnabled(func() {
				_ = conn.Close()
			}, func() {
				writeConn := sink.writeConn(conn)
				wg.Add(1)
				go writeConn.run(wg)
				sink.connections[writeConn] = true
			})
		}
	})
}

func (sink *TCPListenerSink) Header(header Header) error {
	sink.header = header
	// Close all running connections, since we have to negotiate a new header.
	for conn := range sink.connections {
		conn.stop(nil)
	}
	return nil
}

func (sink *TCPListenerSink) Sample(sample Sample) error {
	if err := sink.checkSample(sample); err != nil {
		return err
	}
	for conn, _ := range sink.connections {
		if conn.conn == nil {
			// Clean up closed connections
			delete(sink.connections, conn)
			close(conn.samples)
			continue
		}
		conn.samples <- sample
	}
	return nil
}
