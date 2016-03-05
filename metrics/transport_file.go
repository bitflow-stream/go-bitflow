package metrics

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type FileTransport struct {
	Filename string
	prefix   string
	suffix   string
	file     *os.File

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

func (t *FileTransport) init() {
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

func (source *FileSource) Start(wg *sync.WaitGroup, sink MetricSink) error {
	source.init()
	if _, err := os.Stat(source.Filename); os.IsNotExist(err) {
		source.iterateFiles(wg, sink)
		return nil
	} else if err == nil {
		return source.read(wg, source.Filename, sink)
	} else {
		return err
	}
}

func (source *FileSource) read(wg *sync.WaitGroup, filename string, sink MetricSink) (err error) {
	if source.file, err = os.Open(filename); err == nil {
		simpleReadSamples(wg, source.file.Name(), source.file, source.Unmarshaller, sink)
	}
	return
}

func (source *FileSource) iterateFiles(wg *sync.WaitGroup, sink MetricSink) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := source.fileRegex()
		source.walkFiles(func(info os.FileInfo) error {
			if r.MatchString(filepath.Base(info.Name())) {
				var readWg sync.WaitGroup
				source.read(&readWg, info.Name(), sink)
				readWg.Wait()
				defer source.Close() // Ignore error
			}
			return nil
		})
	}()
}

// ==================== File data sink ====================
type FileSink struct {
	FileTransport
	num int
	abstractSink
}

func (sink *FileSink) Start(wg *sync.WaitGroup, marshaller Marshaller) error {
	log.Println("Writing", marshaller, "samples to", sink.Filename)
	sink.init()
	sink.cleanFiles()
	sink.num = 0
	sink.marshaller = marshaller
	return nil
}

func (sink *FileSink) cleanFiles() error {
	return sink.walkFiles(func(info os.FileInfo) error {
		return os.Remove(info.Name())
	})
}

func (sink *FileSink) openNextFile() error {
	if sink.file != nil {
		if err := sink.file.Close(); err != nil {
			log.Println("Error closing file:", err)
		}
	}
	sink.num++
	name := sink.buildFilename(sink.num)
	var err error
	sink.file, err = os.Create(name)
	log.Println("Opened file", sink.file.Name())
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
	}
	sink.header = header
	return sink.marshaller.WriteHeader(header, sink.file)
}

func (sink *FileSink) Sample(sample Sample) error {
	if err := sink.checkSample(sample); err != nil {
		return err
	}
	return sink.marshaller.WriteSample(sample, sink.file)
}
