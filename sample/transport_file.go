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
	"syscall"

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
	var base string
	group.dir, base = filepath.Split(filename)
	if group.dir == "" {
		group.dir = "."
	}
	index := strings.LastIndex(base, ".")
	if index < 0 {
		group.prefix, group.suffix = base, ""
	} else {
		group.prefix, group.suffix = base[:index], base[index:]
	}
	return
}

func (group *FileGroup) BuildFilename(num int) string {
	base := fmt.Sprintf("%v-%v%v", group.prefix, num, group.suffix)
	return filepath.Join(group.dir, base)
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

// ==================== File data source ====================
type FileSource struct {
	AbstractUnmarshallingMetricSource
	Reader    SampleReader
	Filenames []string
	stream    *SampleInputStream
	closed    *golib.OneshotCondition
}

var fileSourceClosed = errors.New("file source is closed")

func (source *FileSource) String() string {
	if len(source.Filenames) == 1 {
		return fmt.Sprintf("FileSource(%v)", source.Filenames[0])
	} else {
		return fmt.Sprintf("FileSource(%v files)", len(source.Filenames))
	}
}

func (source *FileSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.closed = golib.NewOneshotCondition()
	var files []string
	for _, filename := range source.Filenames {
		group := NewFileGroup(filename)
		if groupFiles, err := group.AllFiles(); err != nil {
			source.CloseSink(wg)
			return golib.TaskFinishedError(err)
		} else {
			files = append(files, groupFiles...)
		}
	}
	if len(files) == 0 {
		source.CloseSink(wg)
		return golib.TaskFinishedError(errors.New("No files specified for FileSource"))
	} else if len(files) > 1 {
		log.Println("Reading", len(files), "files")
	}
	return golib.WaitErrFunc(wg, func() error {
		return source.readFiles(wg, files)
	})
}

func (source *FileSource) Stop() {
	source.closed.Enable(func() {
		if source.stream != nil {
			if err := source.stream.Close(); err != nil && !isFileClosedError(err) {
				log.Println("Error closing file:", err)
			}
		}
	})
}

func (source *FileSource) readFile(filename string) (err error) {
	if file, openErr := os.Open(filename); err != nil {
		err = openErr
	} else {
		var stream *SampleInputStream
		source.closed.IfNotEnabled(func() {
			stream = source.Reader.Open(file, source.Unmarshaller, source.OutgoingSink)
			source.stream = stream
		})
		if stream == nil {
			err = fileSourceClosed
		} else {
			defer stream.Close() // Drop error
			err = stream.ReadNamedSamples(file.Name())
		}
	}
	return
}

func (source *FileSource) readFiles(wg *sync.WaitGroup, files []string) error {
	defer source.CloseSink(wg)
	defer source.closed.EnableOnly()
	for _, filename := range files {
		err := source.readFile(filename)
		if isFileClosedError(err) || err == fileSourceClosed {
			return nil
		} else if err != nil {
			return err
		}
	}
	return nil
}

func isFileClosedError(err error) bool {
	// The file was most likely intentionally closed
	patherr, ok := err.(*os.PathError)
	return ok && patherr.Err == syscall.EBADF
}

// ==================== File data sink ====================
type FileSink struct {
	AbstractMarshallingMetricSink
	Filename string

	group      FileGroup
	lastHeader Header
	file_num   int
	stream     *SampleOutputStream
	closed     *golib.OneshotCondition
}

func (sink *FileSink) String() string {
	return fmt.Sprintf("FileSink(%v)", sink.Filename)
}

func (sink *FileSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Writing", sink.Marshaller, "samples to", sink.Filename)
	sink.closed = golib.NewOneshotCondition()
	sink.group = NewFileGroup(sink.Filename)
	if err := sink.group.DeleteFiles(); err != nil {
		return golib.TaskFinishedError(fmt.Errorf("Failed to clean result files: %v", err))
	}
	sink.file_num = 0
	return nil
}

func (sink *FileSink) flush() error {
	if sink.stream != nil {
		return sink.stream.Close()
	}
	return nil
}

func (sink *FileSink) Close() {
	sink.closed.Enable(func() {
		if err := sink.flush(); err != nil {
			log.Println("Error closing file:", err)
		}
	})
}

func (sink *FileSink) openNextFile() (err error) {
	sink.closed.IfElseEnabled(func() {
		err = errors.New(sink.String() + " is closed")
	}, func() {
		if err = sink.flush(); err != nil {
			return
		}
		sink.file_num++
		name := sink.group.BuildFilename(sink.file_num)

		file, err := os.Create(name)
		if err == nil {
			sink.stream = sink.Writer.OpenBuffered(file, sink.Marshaller)
			log.Println("Opened file", file.Name())
		}
	})
	return
}

func (sink *FileSink) Header(header Header) error {
	if !header.Equals(&sink.lastHeader) {
		if err := sink.openNextFile(); err != nil {
			return err
		}
		sink.lastHeader = header
		return sink.stream.Header(header)
	}
	return nil
}

func (sink *FileSink) Sample(sample Sample, header Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	return sink.stream.Sample(sample)
}
