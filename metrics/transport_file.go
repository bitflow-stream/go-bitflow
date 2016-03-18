package metrics

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/antongulenko/golib"
)

type FileTransport struct {
	Filename string
	prefix   string
	suffix   string
	file     *os.File
	stopped  *golib.OneshotCondition

	abstractSink
}

func (transport *FileTransport) Close() error {
	f := transport.file
	if f != nil {
		transport.file = nil
		return f.Close()
	}
	return nil
}

func (transport *FileTransport) Stop() {
	transport.stopped.Enable(func() {
		if err := transport.Close(); err != nil {
			log.Println("Error closing file:", err)
		}
	})
}

func (t *FileTransport) init() {
	t.stopped = golib.NewOneshotCondition()
	filename := t.Filename
	index := strings.LastIndex(filename, ".")
	if index < 0 {
		index = len(filename)
	}
	t.prefix, t.suffix = filename[:index], filename[index:]
}

func (t *FileTransport) buildFilename(num int) string {
	return fmt.Sprintf("%v-%v%v", t.prefix, num, t.suffix)
}

func (t *FileTransport) fileRegex() *regexp.Regexp {
	return regexp.MustCompile("^" + regexp.QuoteMeta(t.prefix) + "(-[0-9]+)?" + regexp.QuoteMeta(t.suffix) + "$")
}

func (t *FileTransport) walkFiles(walk func(os.FileInfo) error) error {
	dir := filepath.Dir(t.Filename)
	r := t.fileRegex()
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if t.stopped.Enabled() {
			return filepath.SkipDir
		}
		if !info.IsDir() && r.MatchString(filepath.Base(info.Name())) {
			return walk(info)
		}
		if info.IsDir() && path != dir {
			return filepath.SkipDir
		} else {
			return nil
		}
	})
}

// ==================== File data source ====================
type FileSource struct {
	unmarshallingMetricSource
	FileTransport
}

func (source *FileSource) String() string {
	return fmt.Sprintf("FileSource(%s)", source.Filename)
}

func (source *FileSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.init()
	if _, err := os.Stat(source.Filename); os.IsNotExist(err) {
		return source.iterateFiles(wg)
	} else if err == nil {
		return source.read(wg, source.Filename)
	} else {
		return golib.TaskFinishedError(err)
	}
}

func (source *FileSource) read(wg *sync.WaitGroup, filename string) golib.StopChan {
	var file *os.File
	var err error
	source.stopped.IfElseEnabled(func() {
		err = errors.New("FileSource already stopped")
	}, func() {
		file, err = os.Open(filename)
		source.file = file
	})
	if err == nil && file != nil {
		return simpleReadSamples(wg, file.Name(), file, source.Unmarshaller, source.Sink)
	} else {
		return golib.TaskFinishedError(err)
	}
}

func (source *FileSource) iterateFiles(wg *sync.WaitGroup) golib.StopChan {
	return golib.WaitErrFunc(wg, func() error {
		return source.walkFiles(func(info os.FileInfo) error {
			defer source.Close() // Ignore error
			return <-source.read(wg, info.Name())
		})
	})
}

// ==================== File data sink ====================
type FileSink struct {
	FileTransport
	num int
	abstractSink
}

func (source *FileSink) String() string {
	return fmt.Sprintf("FileSink(%s)", source.Filename)
}

func (sink *FileSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Writing", sink.marshaller, "samples to", sink.Filename)
	sink.init()
	if err := sink.cleanFiles(); err != nil {
		return golib.TaskFinishedError(fmt.Errorf("Failed to clean result files: %v", err))
	}
	sink.num = 0
	return nil
}

func (sink *FileSink) cleanFiles() error {
	return sink.walkFiles(func(info os.FileInfo) error {
		return os.Remove(info.Name())
	})
}

func (sink *FileSink) openNextFile() error {
	if err := sink.Close(); err != nil {
		log.Println("Error closing file:", err)
	}
	sink.num++
	name := sink.buildFilename(sink.num)
	var err error
	sink.stopped.IfElseEnabled(func() {
		err = errors.New("FileSink already stopped")
	}, func() {
		sink.file, err = os.Create(name)
		log.Println("Opened file", sink.file.Name())
	})
	return err
}

func (sink *FileSink) Header(header Header) error {
	newHeader := len(sink.header) != len(header)
	if !newHeader {
		for i, old := range sink.header {
			if old != header[i] {
				newHeader = true
				break
			}
		}
	}
	if newHeader {
		if err := sink.openNextFile(); err != nil {
			return err
		}
		sink.header = header
		return sink.marshaller.WriteHeader(header, sink.file)
	}
	return nil
}

func (sink *FileSink) Sample(sample Sample) error {
	if err := sink.checkSample(sample); err != nil {
		return err
	}
	return sink.marshaller.WriteSample(sample, sink.file)
}
