package steps

import (
	"container/list"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

func RegisterTagSynchronizer(b reg.ProcessorRegistry) {
	b.RegisterStep("synchronize_tags",
		func(p *bitflow.SamplePipeline, params map[string]string) error {
			var err error
			synchronizer := new(TagSynchronizer)
			synchronizer.StreamIdentifierTag = params["identifier"]
			synchronizer.ReferenceStream = params["reference"]
			synchronizer.NumTargetStreams = reg.IntParam(params, "num", 0, false, &err)
			if err == nil {
				p.Add(synchronizer)
			}
			return err
		},
		"Split samples into streams identified by a given tag,",
		reg.RequiredParams("identifier", "reference", "num"))
}

// This processor copies tags from a "reference" sample stream to a number of "target" sample streams.
// Streams are identified by the value of a given tag, where the reference stream holds a special value that must be given.
// The target streams can have arbitrary values. The tag synchronization is done by time: one reference sample affects
// all target samples after its timestamp, and before the timestamp of the follow-up reference sample. Target samples
// with timestamps BEFORE any reference sample are forwarded unmodified (with a warning).
// Target samples AFTER the last reference sample will receive the tags from the last reference sample.
// All streams are assumed to be sorted by time, arrive in parallel, and are forwarded in the same order.
type TagSynchronizer struct {
	bitflow.NoopProcessor

	StreamIdentifierTag string
	ReferenceStream     string
	NumTargetStreams    int

	lock               sync.Mutex
	warnedMissingTag   bool
	warnedEarlySamples map[string]bool
	referenceStream    []bitflow.SampleAndHeader
	streams            map[string]*targetStream
}

func (s *TagSynchronizer) Start(wg *sync.WaitGroup) golib.StopChan {
	s.streams = make(map[string]*targetStream)
	s.warnedEarlySamples = make(map[string]bool)
	return s.NoopProcessor.Start(wg)
}

func (s *TagSynchronizer) String() string {
	return fmt.Sprintf("Synchronize tags from %v=%v to %v other streams identified by tag %v",
		s.StreamIdentifierTag, s.ReferenceStream, s.NumTargetStreams, s.StreamIdentifierTag)
}

func (s *TagSynchronizer) Close() {
	err := s.synchronize()
	if err != nil {
		log.Errorf("Error during last synchronization in tag-synchronizer: %v", err)
	}

	var last *bitflow.Sample
	if len(s.referenceStream) == 0 {
		log.Warnf("Tag-synchronizer: No reference samples while closing, cannot synchronize tags for remaining streams")
	} else {
		last = s.referenceStream[len(s.referenceStream)-1].Sample
	}

	// All remaining samples receive the tags from the last reference sample
	for _, stream := range s.streams {
		for i := stream.samples.Front(); i != nil; i = i.Next() {
			sample := i.Value.(bitflow.SampleAndHeader)
			if last != nil {
				s.copyTags(last, sample.Sample)
			}
			stream.samples.Remove(i)
			err := s.NoopProcessor.Sample(sample.Sample, sample.Header)
			if err != nil {
				log.Errorf("Error forwarding last samples in tag-synchronizer: %v", err)
			}
		}
	}

	// Also flush remaining reference samples
	for i, sample := range s.referenceStream {
		s.referenceStream[i] = bitflow.SampleAndHeader{}
		err := s.NoopProcessor.Sample(sample.Sample, sample.Header)
		if err != nil {
			log.Errorf("Error forwarding last reference samples in tag-synchronizer: %v", err)
		}
	}

	s.NoopProcessor.Close()
}

func (s *TagSynchronizer) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if !sample.HasTag(s.StreamIdentifierTag) {
		if !s.warnedMissingTag {
			log.Warnf("Sample does not contain '%v' tag, cannot synchronize tags from '%v'='%v'", s.StreamIdentifierTag, s.StreamIdentifierTag, s.ReferenceStream)
			s.warnedMissingTag = true
		}
		return s.NoopProcessor.Sample(sample, header)
	} else {
		s.lock.Lock()
		defer s.lock.Unlock()

		s.pushSample(sample, header)
		return s.synchronize()
	}
}

func (s *TagSynchronizer) pushSample(sample *bitflow.Sample, header *bitflow.Header) {
	streamName := sample.Tag(s.StreamIdentifierTag)
	if streamName == s.ReferenceStream {
		s.referenceStream = append(s.referenceStream, bitflow.SampleAndHeader{
			Sample: sample,
			Header: header,
		})
	} else {
		stream, ok := s.streams[streamName]
		if !ok {
			stream = new(targetStream)
			s.streams[streamName] = stream
		}
		stream.samples.PushBack(bitflow.SampleAndHeader{
			Sample: sample,
			Header: header,
		})
	}
}

func (s *TagSynchronizer) synchronize() error {
	if len(s.referenceStream) == 0 {
		// No reference samples currently - cannot synchronize any tags
		return nil
	}

	var err golib.MultiError
	send := func(sample bitflow.SampleAndHeader) {
		err.Add(s.NoopProcessor.Sample(sample.Sample, sample.Header))
	}

	for streamName, stream := range s.streams {
		for i := stream.samples.Front(); i != nil; i = i.Next() {
			sample := i.Value.(bitflow.SampleAndHeader)
			index := sort.Search(len(s.referenceStream), func(i int) bool {
				return s.referenceStream[i].Time.After(sample.Time)
			})

			if index <= 0 {
				// Sample is either from before the reference stream starts, or this stream is unsorted.
				// Forward the sample without modifications
				if !s.warnedEarlySamples[streamName] {
					log.Warnf("Stream '%v'='%v' contains samples from before the start of the reference stream. Those samples will not synchronize any tags.", s.StreamIdentifierTag, streamName)
					s.warnedEarlySamples[streamName] = true
				}
			} else if index >= len(s.referenceStream) {
				// Sample from after the reference stream - this and all further samples cannot currently be handled
				break
			} else {
				s.copyTags(s.referenceStream[index-1].Sample, sample.Sample)
			}
			stream.position = sample.Time
			stream.samples.Remove(i)
			send(sample)
		}
	}

	// If we have already seen all required streams, flush all reference samples from the past
	if len(s.streams) >= s.NumTargetStreams {

		// Find the oldest target stream position
		var oldest time.Time
		for _, stream := range s.streams {
			if !stream.position.IsZero() && (oldest.IsZero() || stream.position.Before(oldest)) {
				oldest = stream.position
			}
		}

		// Flush unneeded samples from the reference stream. Can only flush samples that have been surpassed by ALL target streams.
		if !oldest.IsZero() {
			r := s.referenceStream
			for i, sample := range r {
				// This will preserve at least one reference sample
				if i >= len(r)-1 || r[i+1].Time.After(oldest) {
					break
				}
				if i < len(r)-1 {
					copy(r[i:], r[i+1:])
				}
				r = r[:len(r)-1]
				send(sample)
			}
			s.referenceStream = r
		}
	}
	return err.NilOrError()
}

func (s *TagSynchronizer) copyTags(from, to *bitflow.Sample) {
	oldRef := to.Tag(s.StreamIdentifierTag)
	to.AddTagsFrom(from)
	to.SetTag(s.StreamIdentifierTag, oldRef)
}

type targetStream struct {
	samples  list.List
	position time.Time
}
