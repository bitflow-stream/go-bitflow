package sample

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

	AbstractMarshallingMetricSink
}

func (t *FileTransport) CloseFile() error {
	f := t.file
	if f != nil {
		t.file = nil
		return f.Close()
	}
	return nil
}

func (t *FileTransport) close() {
	t.stopped.Enable(func() {
		if err := t.CloseFile(); err != nil {
			log.Println("Error closing file:", err)
		}
	})
}

func (t *FileTransport) init() {
	t.stopped = golib.NewOneshotCondition()
	filename := t.Filename
	index := strings.LastIndex(filename, ".")
	if index < 0 {
		t.prefix, t.suffix = filename, ""
	} else {
		t.prefix, t.suffix = filename[:index], filename[index:]
	}
}

func (t *FileTransport) buildFilename(num int) string {
	return fmt.Sprintf("%v-%v%v", t.prefix, num, t.suffix)
}

func (t *FileTransport) fileRegex() *regexp.Regexp {
	return regexp.MustCompile("^" + regexp.QuoteMeta(t.prefix) + "(-[0-9]+)?" + regexp.QuoteMeta(t.suffix) + "$")
}

func (t *FileTransport) walkFiles(walk func(os.FileInfo) error) (int, error) {
	dir := filepath.Dir(t.Filename)
	r := t.fileRegex()
	num := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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
	return num, err
}

// ==================== File data source ====================
type FileSource struct {
	AbstractUnmarshallingMetricSource
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
		return source.read(wg, source.Filename, true)
	} else {
		return golib.TaskFinishedError(err)
	}
}

func (source *FileSource) Stop() {
	source.close()
}

func (source *FileSource) read(wg *sync.WaitGroup, filename string, closeSink bool) golib.StopChan {
	var file *os.File
	var err error
	source.stopped.IfElseEnabled(func() {
		err = errors.New("FileSource already stopped")
	}, func() {
		file, err = os.Open(filename)
		source.file = file
	})
	if err == nil && file != nil {
		return golib.WaitErrFunc(wg, func() error {
			err := readSamplesNamed(file.Name(), file, source.Unmarshaller, source.OutgoingSink)
			if closeSink {
				source.CloseSink()
			}
			return err
		})
	} else {
		if closeSink {
			source.CloseSink()
		}
		return golib.TaskFinishedError(err)
	}
}

func (source *FileSource) iterateFiles(wg *sync.WaitGroup) golib.StopChan {
	return golib.WaitErrFunc(wg, func() error {
		num, err := source.walkFiles(func(info os.FileInfo) error {
			defer source.CloseFile() // Ignore error
			return <-source.read(wg, info.Name(), false)
		})
		if err == nil && num <= 0 {
			err = errors.New(os.ErrNotExist.Error() + ": " + source.Filename)
		}
		source.CloseSink()
		return err
	})
}

// ==================== File data sink ====================
type FileSink struct {
	FileTransport
	AbstractMarshallingMetricSink
	LastHeader Header
	num        int
}

func (source *FileSink) String() string {
	return fmt.Sprintf("FileSink(%s)", source.Filename)
}

func (sink *FileSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Writing", sink.Marshaller, "samples to", sink.Filename)
	sink.init()
	if err := sink.cleanFiles(); err != nil {
		return golib.TaskFinishedError(fmt.Errorf("Failed to clean result files: %v", err))
	}
	sink.num = 0
	return nil
}

func (source *FileSink) Close() {
	source.close()
}

func (sink *FileSink) cleanFiles() error {
	_, err := sink.walkFiles(func(info os.FileInfo) error {
		return os.Remove(info.Name())
	})
	return err
}

func (sink *FileSink) openNextFile() error {
	if err := sink.CloseFile(); err != nil {
		log.Println("Error closing file:", err)
	}
	sink.num++
	name := sink.buildFilename(sink.num)
	var err error
	sink.stopped.IfElseEnabled(func() {
		err = errors.New("FileSink already stopped")
	}, func() {
		sink.file, err = os.Create(name)
		if err == nil {
			log.Println("Opened file", sink.file.Name())
		}
	})
	return err
}

func (sink *FileSink) Header(header Header) error {
	newHeader := len(sink.LastHeader.Fields) != len(header.Fields)
	if !newHeader {
		for i, old := range sink.LastHeader.Fields {
			if old != header.Fields[i] {
				newHeader = true
				break
			}
		}
	}
	if newHeader {
		if err := sink.openNextFile(); err != nil {
			return err
		}
		sink.LastHeader = header
		return sink.Marshaller.WriteHeader(header, sink.file)
	}
	return nil
}

func (sink *FileSink) Sample(sample Sample, header Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	return sink.Marshaller.WriteSample(sample, header, sink.file)
}
