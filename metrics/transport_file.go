package metrics

import (
	"log"
	"os"
	"sync"
)

type FileTransport struct {
	Filename string
	file     *os.File

	abstractSink
}

func (transport *FileTransport) Close() error {
	if transport.file != nil {
		return transport.file.Close()
	}
	return nil
}

// ==================== File data source ====================
type FileSource struct {
	FileTransport
}

func (source *FileSource) Start(wg *sync.WaitGroup, um Unmarshaller, sink MetricSink) (err error) {
	if source.file, err = os.Open(source.Filename); err == nil {
		simpleReadSamples(wg, source.file.Name(), source.file, um, sink)
	}
	return
}

// ==================== File data sink ====================
type FileSink struct {
	FileTransport
	abstractSink
}

func (sink *FileSink) Start(wg *sync.WaitGroup, marshaller Marshaller) (err error) {
	sink.marshaller = marshaller
	sink.file, err = os.Create(sink.Filename)
	return
}

func (sink *FileSink) Header(header Header) error {
	sink.header = header
	log.Printf("Writing %v metrics to %v\n", len(header), sink.file.Name())
	return sink.marshaller.WriteHeader(header, sink.file)
}

func (sink *FileSink) Sample(sample Sample) error {
	if err := sink.checkSample(sample); err != nil {
		return err
	}
	return sink.marshaller.WriteSample(sample, sink.file)
}
