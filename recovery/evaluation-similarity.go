package recovery

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/bitflow-script/reg"
	"github.com/bitflow-stream/go-bitflow-pipeline/steps"
	log "github.com/sirupsen/logrus"
)

func RegisterEventSimilarityEvaluation(b reg.ProcessorRegistry) {
	b.RegisterAnalysisParamsErr("recovery-event-similarity-evaluation",
		func(p *pipeline.SamplePipeline, params map[string]string) error {
			selection, err := NewRecommendingSelection(params)
			if err != nil {
				return err
			}
			collector := new(EvaluationDataCollector)
			collector.ParseRecoveryTags(params)
			p.Add((&pipeline.BatchProcessor{
				FlushTags:                   []string{collector.NodeNameTag, collector.StateTag, "anomaly"},
				SampleTimestampFlushTimeout: 5 * time.Second,
			}).Add(collector))
			proc := &EventSimilarityEvaluationProcessor{
				ConfigurableTags: collector.ConfigurableTags,
				selection:        selection,
				data:             make(map[string][]AnomalyData),
				samplesPerState:  make(map[string]int),
				SkipSamples:      reg.IntParam(params, "skip-samples", 0, false, &err),
				SamplesPerEvent:  reg.IntParam(params, "samples-per-event", 0, false, &err),
			}
			collector.StoreEvaluationEvent = proc.storeEvaluationEvent
			p.Add(proc)
			return err
		},
		"Recovery Engine based on recommendation system",
		reg.RequiredParams(append([]string{"skip-samples", "samples-per-event"}, TagParameterNames...)...),
		reg.OptionalParams(
			"max-curiosity", "curiosity-growth", "curiosity-growth-time",
			"similarity-scale-stddev", "similarity-scale-min-max", "similarity-scale-from", "similarity-scale-to", "similarity-normalize-similarities", // Recommending selection: similarity computation
		))
}

type EventSimilarityEvaluationProcessor struct {
	bitflow.NoopProcessor
	ConfigurableTags
	selection *RecommendingSelection

	data            map[string][]AnomalyData // state-name -> slice of events of that state
	samplesPerState map[string]int
	totalSamples    int
	totalEvents     int

	SamplesPerEvent int
	SkipSamples     int // in the beginning and in the end
}

func (p *EventSimilarityEvaluationProcessor) String() string {
	return fmt.Sprintf("event-similarity-evaluation (%v samples per event, skipping %v samples)", p.SamplesPerEvent, p.SkipSamples)
}

func (p *EventSimilarityEvaluationProcessor) storeEvaluationEvent(node, state string, samples []*bitflow.Sample, header *bitflow.Header) {
	if state != p.NormalStateValue {
		if len(samples) < 2*p.SkipSamples+p.SamplesPerEvent {
			log.Warnf("Skipping event state %v of node %v, because only %v samples are available", state, node)
			return
		}

		// Ignore a number of samples in the beginning and in the end
		samples := samples[p.SkipSamples : len(samples)-p.SkipSamples]
		skip := float64(len(samples)) / float64(p.SamplesPerEvent)

		// Select the required number of samples from the entire range of the event
		eventFeatures := p.data[state]
		for i := 0; i < p.SamplesPerEvent; i++ {
			sample := samples[int(float64(i)*skip)]
			newFeatures := SampleToAnomalyFeatures(sample, header)
			eventFeatures = append(eventFeatures, newFeatures)
			p.totalSamples++
			p.samplesPerState[state]++
		}
		p.data[state] = eventFeatures
		p.totalEvents++
	}
}

func (p *EventSimilarityEvaluationProcessor) Close() {
	p.performEvaluation()
	p.NoopProcessor.Close()
}

func (p *EventSimilarityEvaluationProcessor) performEvaluation() {
	log.Printf("Evaluation input data for %v nodes:", len(p.data))
	for name, num := range p.samplesPerState {
		log.Printf("%v: %v", name, num)
	}
	log.Printf("Evaluating %v anomaly states in %v samples. Sample counts:", p.totalEvents, p.totalSamples)
	for eventName := range p.data {
		log.Printf("%v: %v", eventName, len(p.data[eventName]))
	}

	var eval similarityEvaluator
	eval.evaluate(p.data, p.selection)
	eval.getResults().print(eval.eventNames)
}

type stats struct {
	steps.FeatureStats
}

func (s stats) String() string {
	return fmt.Sprintf("%.2f ±%.2f [%v]", s.Mean(), s.Stddev(), s.Len())
}

type similarityEvaluator struct {
	events     map[string][]AnomalyData
	eventNames []string
	results    [][][]float64
}

type similarityResults struct {
	results            [][]stats // cross-validation matrix
	self               []stats
	others             []stats
	selfAll, othersAll stats
}

func (p *similarityEvaluator) evaluate(events map[string][]AnomalyData, selection *RecommendingSelection) {
	p.events = events
	p.eventNames = make([]string, 0, len(events))
	var allEvents []AnomalyData
	for eventName, eventData := range events {
		p.eventNames = append(p.eventNames, eventName)
		allEvents = append(allEvents, eventData...)
	}
	sort.Strings(p.eventNames)

	// Normalize the input data, if enabled
	selection.preprocess(allEvents)

	var similaritySlices [][]float64
	p.results = make([][][]float64, len(events))
	for i, name1 := range p.eventNames {
		compareEvents := p.eventNames[i:]
		subResults := make([][]float64, len(compareEvents))
		for j, name2 := range compareEvents {
			events1 := events[name1]
			events2 := events[name2]
			log.Printf("Comparing %v (%v samples) and %v (%v samples)...", name1, len(events1), name2, len(events2))

			// Compute the similarity between every single sample-pair
			for _, event1 := range events1 {
				similarities := selection.computeSimilarities(events2, event1)
				subResults[j] = append(subResults[j], similarities...)
			}
		}
		similaritySlices = append(similaritySlices, subResults...)
		p.results[i] = subResults
	}

	// Normalize the output data, if enabled
	selection.normalizeSimilarities(similaritySlices...)
}

func (p *similarityEvaluator) getResults() (res similarityResults) {
	res.results = make([][]stats, len(p.results))
	res.self = make([]stats, len(p.results))
	res.others = make([]stats, len(p.results))
	res.selfAll = stats{}
	res.othersAll = stats{}
	for i := range p.eventNames {
		others := p.eventNames[i:]
		subResults := make([]stats, len(others))
		for j := range others {
			for _, simil := range p.results[i][j] {
				subResults[j].Push(simil)
				if j == 0 {
					res.self[i].Push(simil)
					res.selfAll.Push(simil)
				} else {
					res.others[i].Push(simil)
					res.othersAll.Push(simil)
					res.others[i+j].Push(simil)
				}
			}
		}
		res.results[i] = subResults
	}
	return
}

func (p similarityResults) print(eventNames []string) {
	wr := tabwriter.NewWriter(os.Stderr, 0, 10, 2, ' ', 0)
	for _, name := range eventNames {
		fmt.Fprintf(wr, "\t%v", name)
	}
	wr.Write([]byte("\t  self\t  others\n"))
	var selfAvg, selfStddev, othersAvg, othersStddev stats
	for i, subResults := range p.results {
		fmt.Fprintf(wr, eventNames[i])
		for j := 0; j < i; j++ {
			wr.Write([]byte{'\t'})
		}
		for _, res := range subResults {
			fmt.Fprintf(wr, "\t%v", res)
		}
		fmt.Fprintf(wr, "\t  %v\t  %v\n", p.self[i], p.others[i])
		selfAvg.Push(p.self[i].Mean())
		selfStddev.Push(p.self[i].Stddev())
		othersAvg.Push(p.others[i].Mean())
		othersStddev.Push(p.others[i].Stddev())
	}
	skip := strings.Repeat("\t", len(p.results)+1)
	fmt.Fprintf(wr, "%s= %v\t= %v\n", skip, p.selfAll, p.othersAll)
	fmt.Fprintf(wr, "%sø %.2f ±%.2f\tø %.2f ±%.2f\n", skip, selfAvg.Mean(), selfStddev.Mean(), othersAvg.Mean(), othersStddev.Mean())
	wr.Flush()
}
