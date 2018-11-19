package plugin

import "github.com/bitflow-stream/go-bitflow"

type SampleSourcePlugin interface {
	Start(params map[string]string, dataSink DataSink)
	Close()
}

type DataSink interface {
	Error(error)
	Close()
	Sample(sample *bitflow.Sample, header *bitflow.Header)
}
