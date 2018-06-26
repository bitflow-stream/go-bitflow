package plugin

import bitflow "github.com/antongulenko/go-bitflow"

type SampleSourcePlugin interface {
	Start(params map[string]string, dataSink PluginDataSink)
	Close()
}

type PluginDataSink interface {
	Error(error)
	Close()
	Sample(sample *bitflow.Sample, header *bitflow.Header)
}
