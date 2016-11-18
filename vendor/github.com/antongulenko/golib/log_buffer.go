package golib

import (
	"container/ring"
	"fmt"
	"io"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type LogBuffer struct {
	messages          *ring.Ring
	msgLock           sync.Mutex
	message_buffer    int
	originalLoggerOut io.Writer
}

func NewLogBuffer(message_buffer int) *LogBuffer {
	return &LogBuffer{
		messages:       ring.New(message_buffer),
		message_buffer: message_buffer,
	}
}

func (buf *LogBuffer) PushMessage(msg string) {
	buf.msgLock.Lock()
	defer buf.msgLock.Unlock()
	buf.messages.Value = msg
	buf.messages = buf.messages.Next()
}

func (buf *LogBuffer) PrintMessages(w io.Writer, max_num int) {
	if max_num <= 0 {
		return
	}
	msgStart := buf.messages
	if max_num < buf.message_buffer {
		msgStart = msgStart.Move(-max_num)
	}
	msgStart.Do(func(msg interface{}) {
		if msg != nil {
			fmt.Fprint(w, msg)
		}
	})
}

func (buf *LogBuffer) RegisterMessageHook() {
	log.StandardLogger().Hooks.Add(buf)
}

func (buf *LogBuffer) InterceptLogger() {
	buf.originalLoggerOut = log.StandardLogger().Out
	log.StandardLogger().Out = noopWriter{}
}

func (buf *LogBuffer) RestoreLogger() {
	log.StandardLogger().Out = buf.originalLoggerOut
}

func (buf *LogBuffer) Levels() []log.Level {
	res := make([]log.Level, 256)
	for i := 0; i < len(res); i++ {
		res[int(i)] = log.Level(i)
	}
	return res
}

func (buf *LogBuffer) Fire(entry *log.Entry) error {
	msg, err := log.StandardLogger().Formatter.Format(entry)
	if err != nil {
		return err
	}
	buf.PushMessage(string(msg))
	return nil
}

type noopWriter struct {
}

func (noopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
