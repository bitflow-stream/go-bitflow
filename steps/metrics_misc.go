package steps

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

func RegisterSetCurrentTime(b reg.ProcessorRegistry) {
	b.RegisterStep("set_time",
		func(p *bitflow.SamplePipeline, _ map[string]string) error {
			p.Add(&bitflow.SimpleProcessor{
				Description: "reset time to now",
				Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
					sample.Time = time.Now()
					return sample, header, nil
				},
			})
			return nil
		},
		"Set the timestamp on every processed sample to the current time")
}

func RegisterAppendTimeDifference(b reg.ProcessorRegistry) {
	fieldName := "time-difference"
	var checker bitflow.HeaderChecker
	var outHeader *bitflow.Header
	var lastTime time.Time

	b.RegisterStep("append_latency",
		func(p *bitflow.SamplePipeline, _ map[string]string) error {
			p.Add(&bitflow.SimpleProcessor{
				Description: "Append time difference as metric " + fieldName,
				Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
					if checker.HeaderChanged(header) {
						outHeader = header.Clone(append(header.Fields, fieldName))
					}
					var diff float64
					if !lastTime.IsZero() {
						diff = float64(sample.Time.Sub(lastTime))
					}
					lastTime = sample.Time
					AppendToSample(sample, []float64{diff})
					return sample, outHeader, nil
				},
			})
			return nil
		},
		"Append the time difference to the previous sample as a metric")
}

func RegisterStripMetrics(b reg.ProcessorRegistry) {
	b.RegisterStep("strip",
		func(p *bitflow.SamplePipeline, _ map[string]string) error {
			p.Add(&bitflow.SimpleProcessor{
				Description: "remove metric values, keep timestamp and tags",
				Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
					return sample.Metadata().NewSample(nil), header.Clone(nil), nil
				},
			})
			return nil
		},
		"Remove all metrics, only keeping the timestamp and the tags of each sample")
}

func RegisterParseTags(b reg.ProcessorRegistry) {
	b.RegisterStep("parse_tags",
		func(p *bitflow.SamplePipeline, params map[string]string) error {
			var checker bitflow.HeaderChecker
			var outHeader *bitflow.Header
			var sorted bitflow.SortedStringPairs
			warnedMissingTags := make(map[string]bool)
			sorted.FillFromMap(params)
			sort.Sort(&sorted)

			p.Add(&bitflow.SimpleProcessor{
				Description: "Convert tags to metrics: " + sorted.String(),
				Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
					if checker.HeaderChanged(header) {
						outHeader = header.Clone(append(header.Fields, sorted.Keys...))
					}
					values := make([]float64, len(sorted.Values))
					for i, tag := range sorted.Values {
						var value float64
						if !sample.HasTag(tag) {
							if !warnedMissingTags[tag] {
								warnedMissingTags[tag] = true
								log.Warnf("Encountered sample missing tag '%v'. Using metric value 0 instead. This warning is printed once per tag.", tag)
							}
						} else {
							var err error
							value, err = strconv.ParseFloat(sample.Tag(tag), 64)
							if err != nil {
								return nil, nil, fmt.Errorf("Cloud not convert '%v' tag to float64: %v", tag, err)
							}
						}
						values[i] = value
					}
					AppendToSample(sample, values)
					return sample, outHeader, nil
				},
			})
			return nil
		},
		"Append metrics based on tag values. Keys are new metric names, values are tag names",
		reg.VariableParams())
}
