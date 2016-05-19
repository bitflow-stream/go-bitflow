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

// ==================== File group ====================
type FileGroup struct {
	filename string
	dir      string
	prefix   string
	suffix   string
}

func NewFileGroup(filename string) (group FileGroup) {
	group.filename = filename
	group.dir = filepath.Dir(filename)
	base := filepath.Base(filename)
	index := strings.LastIndex(base, ".")
	if index < 0 {
		group.prefix, group.suffix = base, ""
	} else {
		group.prefix, group.suffix = base[:index], base[index:]
	}
	return
}

func (group *FileGroup) BuildFilename(num int) string {
	return fmt.Sprintf("%v-%v%v", group.prefix, num, group.suffix)
}

func (group *FileGroup) FileRegex() *regexp.Regexp {
	return regexp.MustCompile("^" + regexp.QuoteMeta(group.prefix) + "(-[0-9]+)?" + regexp.QuoteMeta(group.suffix) + "$")
}

var StopWalking = errors.New("stop walking")

func (group *FileGroup) WalkFiles(walk func(string, os.FileInfo) error) (num int, err error) {
	r := group.FileRegex()
	stopped := false
	err = filepath.Walk(group.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if stopped {
			return filepath.SkipDir
		}
		if !info.IsDir() && r.MatchString(filepath.Base(info.Name())) {
			num++
			err := walk(path, info)
			if err == StopWalking {
				stopped = true
			}
			return err
		}
		if info.IsDir() && path != group.dir {
			return filepath.SkipDir
		} else {
			return nil
		}
	})
	if err == StopWalking {
		err = nil
	}
	return
}

func (group *FileGroup) AllFiles() (all []string, err error) {
	_, err = group.WalkFiles(func(path string, _ os.FileInfo) error {
		all = append(all, path)
		return nil
	})
	if err == nil && len(all) == 0 {
		err = errors.New(os.ErrNotExist.Error() + ": " + group.filename)
	}
	return
}

func (group *FileGroup) DeleteFiles() error {
	_, err := group.WalkFiles(func(path string, _ os.FileInfo) error {
		return os.Remove(path)
	})
	return err
}

// ==================== File transport ====================
type FileTransport struct {
	AbstractMarshallingMetricSink
	file    *os.File
	stopped *golib.OneshotCondition
}

func (t *FileTransport) CloseFile() error {
	if f := t.file; f != nil {
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
}

// ==================== File data source ====================
type FileSource struct {
	AbstractUnmarshallingMetricSource
	FileTransport
	Filenames []string
}

func (source *FileSource) String() string {
	if len(source.Filenames) == 1 {
		return fmt.Sprintf("FileSource(%v)", source.Filenames[0])
	} else {
		return fmt.Sprintf("FileSource(%v files)", len(source.Filenames))
	}
}

func (source *FileSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.init()
	var files []string
	for _, filename := range source.Filenames {
		group := NewFileGroup(filename)
		if groupFiles, err := group.AllFiles(); err != nil {
			return golib.TaskFinishedError(err)
		} else {
			files = append(files, groupFiles...)
		}
	}
	if len(files) == 0 {
		return golib.TaskFinishedError(errors.New("No files specified for FileSource"))
	} else if len(files) > 1 {
		log.Println("Reading", len(files), "files")
	}
	return golib.WaitErrFunc(wg, func() error {
		return source.readFiles(wg, files)
	})
}

func (source *FileSource) Stop() {
	source.close()
}

func (source *FileSource) read(filename string) (err error) {
	var file *os.File
	source.stopped.IfElseEnabled(func() {
		err = errors.New("FileSource already stopped")
	}, func() {
		file, err = os.Open(filename)
		source.file = file
	})
	if err == nil {
		err = readSamplesNamed(file.Name(), file, source.Unmarshaller, source.OutgoingSink)
	}
	return
}

func (source *FileSource) readFiles(wg *sync.WaitGroup, files []string) (err error) {
	defer source.CloseSink(wg)
	for _, filename := range files {
		err = source.read(filename)
		_ = source.CloseFile() // Drop error
		if err != nil {
			break
		}
	}
	return
}

// ==================== File data sink ====================
type FileSink struct {
	FileTransport
	AbstractMarshallingMetricSink
	Filename string

	group      FileGroup
	lastHeader Header
	num        int
}

func (sink *FileSink) String() string {
	return fmt.Sprintf("FileSink(%v)", sink.Filename)
}

func (sink *FileSink) Start(wg *sync.WaitGroup) golib.StopChan {
	sink.init()
	log.Println("Writing", sink.Marshaller, "samples to", sink.Filename)
	sink.group = NewFileGroup(sink.Filename)
	if err := sink.group.DeleteFiles(); err != nil {
		return golib.TaskFinishedError(fmt.Errorf("Failed to clean result files: %v", err))
	}
	sink.num = 0
	return nil
}

func (source *FileSink) Close() {
	source.close()
}

func (sink *FileSink) openNextFile() error {
	if err := sink.CloseFile(); err != nil {
		log.Println("Error closing file:", err)
	}
	sink.num++
	name := sink.group.BuildFilename(sink.num)
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
	if !header.Equals(&sink.lastHeader) {
		if err := sink.openNextFile(); err != nil {
			return err
		}
		sink.lastHeader = header
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
