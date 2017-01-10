package bitflow

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/golib"
	"vbom.ml/util/sortorder"
)

// FileGroup provides utility functionality when dealing with a group of files sharing
// the same directory, file prefix and file extension. It provides methods for listing,
// walking or deleting files that belong to that group.
type FileGroup struct {
	filename string
	dir      string
	prefix   string
	suffix   string
}

// NewFileGroup returns a new FileGroup instance. The filename parameter is parsed
// and split into directory, file name prefix and file extension. The file can also have no extension.
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

// BuildFilename returns a file belonging to the receiving group, with the added number
// as suffix. The suffix is added before the file extension, separated with a hyphen, like so:
//   dir1/dir2/filePrefix-<num>.ext
func (group *FileGroup) BuildFilename(num int) string {
	return group.BuildFilenameStr(strconv.Itoa(num))
}

// BuildFilenameStr returns a file belonging to the receiving group, with the added string
// as suffix. The suffix is added before the file extension, separated with a hyphen, like so:
//   dir1/dir2/filePrefix-<suffix>.ext
func (group *FileGroup) BuildFilenameStr(suffix string) string {
	filename := group.prefix
	if suffix == "" {
		if filename == "" {
			// Avoid filenames starting with dot and empty filenames.
			// This does not collide with FileSink.openNextNewFile and is also matched by FileRegex().
			filename = "0"
		}
	} else if filename != "" {
		filename = filename + "-"
	}
	filename += suffix + group.suffix
	return filepath.Join(group.dir, filename)
}

// FileRegex returns a regular expresion that matches filenames belonging to the receiving group.
// Only files with an optional numeric suffix are matched, e.g.:
//   dir1/dir2/filePrefix(-[0-9]+)?.ext
// For empty 'filePrefix':
//   dir1/dir2/[0-9]+.ext
func (group *FileGroup) FileRegex() *regexp.Regexp {
	prefix := "[0-9]+"
	if group.prefix != "" {
		prefix = regexp.QuoteMeta(group.prefix) + "(-" + prefix + ")?"
	}
	regex := "^" + prefix + regexp.QuoteMeta(group.suffix) + "$"
	return regexp.MustCompile(regex)
}

// StopWalking can be returned from the walk function parameter for WalkFiles to indicate,
// that the tree should not be walked any further down the current directory.
var StopWalking = errors.New("stop walking")

// WalkFiles walks all files that belong to the receiving FileGroup. It returnes the number
// of walked files and a non-nil error if there was an error while walking.
// The walk function parameter is called for every file, providing the file name and the
// respective os.FileInfo.
//
// WalkFiles walks all files that match the regular expression returnes by FileRegex().
//
// The files are walked in lexical order, which does not represent the order the files
// would be written by FileSink.
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

// AllFiles returns a slice of all files that belong to the receiving
// FileGroup, and a non-nil error if the list could not be determined.
// AllFiles returns all files matching the regular expression returned by
// FileRegex().
//
// The files are returned sorted in the order they would be written out by
// FileSink.
func (group *FileGroup) AllFiles() (all []string, err error) {
	basefile := group.BuildFilenameStr("")
	basefileFound := false
	_, err = group.WalkFiles(func(path string, _ os.FileInfo) error {
		if path == basefile {
			basefileFound = true
		} else {
			all = append(all, path)
		}
		return nil
	})

	// Natural sort: treat numbers as a human would
	sort.Sort(sortorder.Natural(all))

	if basefileFound {
		// Treat the first file specially, otherwise it is sorted wrong.
		all = append([]string{basefile}, all...)
	}
	if err == nil && len(all) == 0 {
		err = errors.New(os.ErrNotExist.Error() + ": " + group.filename)
	}
	return
}

// DeleteFiles tries to delete all files that belong to the receiving FileGroup
// and returns a non-nil error when deleting any of the files failed.
// DeleteFiles deletes all files matching the regular expression returned by
// FileRegex().
func (group *FileGroup) DeleteFiles() error {
	_, err := group.WalkFiles(func(path string, _ os.FileInfo) error {
		return os.Remove(path)
	})
	return err
}

// ==================== File data source ====================

// ReadingDirWarnDuration defines a timeout that triggers a warning printed to
// the logger when listing files using ListMatchingFiles takes too long.
// The warning will be printed only once per ListMatchingFiles invokation,
// and will not be printed if ListMatchingFiles returns in time.
const ReadingDirWarnDuration = 2000 * time.Millisecond

// FileSource is an implementation of UnmarshallingMetricSource that reads samples
// from one or more files. Various parameters control the behavior and performance
// of the FileSource.
type FileSource struct {
	AbstractUnmarshallingMetricSource

	// Filenames is a slice of all files that will be read by the FileSource in sequence.
	// For every Filename, the FileSource will not only read the file itself,
	// but also for all files that belong to the same FileGroup, as returned by:
	//   NewFileGroup(filename).AllFiles()
	Filenames []string

	// ReadFileGroups can be set to true to extend the input files to the associated
	// file groups. For an input file named 'data.bin', all files named 'data-[0-9]+.bin'
	// will be read as well. The file group for 'data' is 'data-[0-9]+', the file
	// group for '.bin' is '[0-9]+.bin'.
	ReadFileGroups bool

	// Robust can be set to true to allow errors when reading or parsing files,
	// and only print Warnings instead. This is useful if the files to be parsed
	// are mostly valid, but have garbage at the end.
	Robust bool

	// IoBuffer configures the buffer size for read files. It should be large enough
	// to allow multiple goroutines to parse the read data in parallel.
	IoBuffer int

	// ConvertFile is an optional hook for converting the filename to a custom string.
	// The custom string will then be passed to the ReadSampleHandler configured in
	// the Reader field, instead of simply using the filename.
	ConvertFilename func(string) string

	// KeepAlive makes this FileSource not close after all files have been read.
	// Instead, it will stay open without producing any more data.
	KeepAlive bool

	stream *SampleInputStream
	closed *golib.OneshotCondition
}

var fileSourceClosed = errors.New("file source is closed")

// String implements the MetricSource interface.
func (source *FileSource) String() string {
	if len(source.Filenames) == 1 {
		return fmt.Sprintf("FileSource(%v)", source.Filenames[0])
	} else {
		return fmt.Sprintf("FileSource(%v files)", len(source.Filenames))
	}
}

// Start implements the MetricSource interface. It starts reading all configured
// files in sequence using background goroutines. Depending on the Robust flag
// of the receiving FileSource, the reading exits after the first error, or continues
// until all configured files have been opened.
func (source *FileSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.closed = golib.NewOneshotCondition()
	var files []string
	if source.ReadFileGroups {
		for _, filename := range source.Filenames {
			group := NewFileGroup(filename)
			if groupFiles, err := group.AllFiles(); err != nil {
				source.CloseSink(wg)
				return golib.TaskFinishedError(err)
			} else {
				files = append(files, groupFiles...)
			}
		}
	} else {
		files = make([]string, len(source.Filenames))
		copy(files, source.Filenames)
	}
	if len(files) == 0 {
		source.CloseSink(wg)
		return golib.TaskFinishedError(errors.New("No files specified for FileSource"))
	} else if len(files) > 1 {
		log.Println("Reading", len(files), "files")
	}
	return source.readFilesKeepAlive(wg, files)
}

// This is a copy of golib.WaitErrFunc, but extended to implement the FileSource.KeepAlive flag.
// If KeepAlive is set, and in case the wait() func returns a nil-error, it will not trigger
// the StopChan immediately, but instead wait for the source.closed condition to be triggered first.
// In other words: after reading all files succesfully, this FileSource will wait for the Stop() call.
// In case of an error, it will still immediately shut down.
func (source *FileSource) readFilesKeepAlive(wg *sync.WaitGroup, files []string) golib.StopChan {
	if wg != nil {
		wg.Add(1)
	}
	finished := make(chan error, 1)
	go func() {
		if wg != nil {
			defer wg.Done()
		}
		err := source.readFiles(wg, files)
		if source.KeepAlive && err == nil {
			source.closed.Wait()
		}
		source.CloseSink(wg)
		source.closed.EnableOnly()
		finished <- err
		close(finished)
	}()
	return finished
}

// Stop implements the MetricSource interface. it stops all goroutines that are spawned
// for reading files and prints any errors to the logger. Calling it after the FileSource
// finished on its own will have no effect.
func (source *FileSource) Stop() {
	source.closed.Enable(func() {
		if source.stream != nil {
			if err := source.stream.Close(); err != nil && !isFileClosedError(err) {
				log.Errorln("Error closing file:", err)
			}
		}
	})
}

func (source *FileSource) readFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	var stream *SampleInputStream
	source.closed.IfNotEnabled(func() {
		stream = source.Reader.OpenBuffered(file, source.OutgoingSink, source.IoBuffer)
		source.stream = stream
	})
	if stream == nil {
		return fileSourceClosed
	}
	defer stream.Close() // Drop error
	name := file.Name()
	if converter := source.ConvertFilename; converter != nil {
		name = converter(name)
	}
	return stream.ReadNamedSamples(name)
}

func (source *FileSource) readFiles(wg *sync.WaitGroup, files []string) error {
	for _, filename := range files {
		err := source.readFile(filename)
		if err == fileSourceClosed {
			return nil
		} else if isFileClosedError(err) {
			continue
		} else if err != nil {
			if source.Robust {
				log.WithFields(log.Fields{"file": filename}).Warnln("Error reading file:", err)
				continue
			} else {
				return err
			}
		}
	}
	return nil
}

func isFileClosedError(err error) bool {
	// The file was most likely intentionally closed
	patherr, ok := err.(*os.PathError)
	return ok && patherr.Err == syscall.EBADF
}

// ListMatchingFiles is a convenience function that traverses a directory recursively
// and returnes all files that match a given regular expression. It prints a warning in the logger,
// if traversing the directory takes too long, see the ReadingDirWarnDuration constant.
func ListMatchingFiles(dir string, regexStr string) ([]string, error) {
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, err
	}
	finishedReading := false
	time.AfterFunc(ReadingDirWarnDuration, func() {
		if !finishedReading {
			log.Warnf("Reading directory \"%v\"...", dir)
		}
	})
	var result []string
	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && regex.MatchString(path) {
			result = append(result, path)
		}
		return nil
	})
	finishedReading = true
	return result, walkErr
}

// ==================== File data sink ====================

const (
	// MaxOutputFileErrors is the number of retries that are accepted before
	// giving up to open an output file. After each try, the output filename
	// will be changed.
	MaxOutputFileErrors = 5

	// MkdirsPermissions defines the permission bits used when creating new
	// directories for storing output files.
	MkdirsPermissions = 0755
)

// FileSink is an implementation of MetricSink that writes output Headers and Samples
// to a given file. Every time a new Header is received by the FileSink, a new file is opened
// using an automatically incremented number as suffix (see FileGroup). Other parameters
// define the parsing behavior of the FileSink.
type FileSink struct {
	// AbstractMarshallingMetricSink defines the Marshaller and SampleWriter that will
	// be used when writing Samples. See their documentation for further info.
	AbstractMarshallingMetricSink

	// Filename defines the file that will be used for writing Samples. Each time a new Header
	// is received be FileSink, a new file will be opened automatically. The filenames are built
	// by FileGroup.BuildFilename(), using an automatically incrementing integer suffix. The first
	// filename will not have any suffix, the second file will have suffix "-0", the second "-1", and so on.
	// If one of those files already exists, the suffix keeps incrementing, until a free slot is found.
	// If errors occurr while opening output files, a number of retries is attempted while incrementing
	// the suffix, until the number of error exceeds MaxOutputFileErrors. After this, the FileSink stops
	// and reports the last error. All intermediate errors are logged as warnings.
	Filename string

	// IoBuffer defines the output buffer when writing samples to a file. It should be large
	// enough to minize the number of write() calls in the operating system.
	IoBuffer int

	// CleanFiles can be set to true to delete all files that would potentially collide with output files.
	// In particular, this causes the following when starting the FileSink:
	//   NewFileGroup(sink.Filename).DeleteFiles()
	// When deleting these files fails, the FileSink stops and reports an error.
	CleanFiles bool

	checker  HeaderChecker
	group    FileGroup
	file_num int
	stream   *SampleOutputStream
	closed   *golib.OneshotCondition
}

// String implements the MetricSink interface.
func (sink *FileSink) String() string {
	return fmt.Sprintf("FileSink(%v)", sink.Filename)
}

// Start implements the MetricSink interface. It does not start any goroutines.
// It initialized the FileSink, prints some log messages, and depending on the
// CleanFiles flag tries to delete existing files that would conflict with the output file.
func (sink *FileSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.WithFields(log.Fields{"file": sink.Filename, "format": sink.Marshaller}).Println("Writing samples")
	sink.closed = golib.NewOneshotCondition()
	sink.group = NewFileGroup(sink.Filename)
	if sink.CleanFiles {
		if err := sink.group.DeleteFiles(); err != nil {
			return golib.TaskFinishedError(fmt.Errorf("Failed to clean result files: %v", err))
		}
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

// Close implements the MetricSink interface. It flushes and closes the currently open file.
// No more data should be written to Sample/Header after calling Close.
func (sink *FileSink) Close() {
	sink.closed.Enable(func() {
		if err := sink.flush(); err != nil {
			log.Errorln("Error closing file:", err)
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
		var file *os.File
		file, err = sink.openNextNewFile()
		if err == nil {
			sink.stream = sink.Writer.OpenBuffered(file, sink.Marshaller, sink.IoBuffer)
			log.WithField("file", file.Name()).Println("Opened file")
		}
	})
	return
}

func (sink *FileSink) openNextNewFile() (file *os.File, err error) {
	num_errors := 0
	file_num := sink.file_num
	for {
		var name string
		if file_num == 0 {
			name = sink.group.BuildFilenameStr("")
		} else {
			name = sink.group.BuildFilename(file_num)
		}
		file_num++

		if _, err = os.Stat(name); os.IsNotExist(err) {
			// File does not exist, try to open and create it

			dir := path.Dir(name)
			if _, err = os.Stat(dir); os.IsNotExist(err) {
				// Directory does not exist, try to create entire path
				err = os.MkdirAll(dir, MkdirsPermissions)
			}
			if err == nil {
				file, err = os.Create(name)
			}
		} else if err == nil {
			// File exists, try next one
			continue
		}

		if err == nil {
			sink.file_num = file_num
			return
		}
		log.WithField("file", name).Warnln("Failed to open output file:", err)
		num_errors++
		if num_errors >= MaxOutputFileErrors {
			return
		}
	}
}

// Sample writes a Sample to the current open file.
func (sink *FileSink) Sample(sample *Sample, header *Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	if sink.checker.HeaderChanged(header) {
		if err := sink.openNextFile(); err != nil {
			return err
		}
	}
	return sink.stream.Sample(sample, header)
}
