package pipeline

import (
	"container/list"
	"math"
	"sync"
	"time"

	bitflow "github.com/antongulenko/go-bitflow"
	log "github.com/sirupsen/logrus"
)

type SynchronizedSampleMerger struct {
	bitflow.NoopProcessor

	MergeTag        string
	MergeInterval   time.Duration
	ExpectedStreams int
	MergeSamples    func([]*bitflow.Sample, []*bitflow.Header) (*bitflow.Sample, *bitflow.Header)

	Description        string
	DebugQueueLengths  bool
	DebugWaitingQueues bool

	queues          map[string]*mergeQueue
	readyQueues     map[string]bool
	readyQueuesLock sync.Mutex
	start           time.Time
	end             time.Time
	samples         []queueElem // Reused slice to avoid reallocations
}

func (p *SynchronizedSampleMerger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if !sample.HasTag(p.MergeTag) {
		log.Warnln("Dropping sample without", p.MergeTag, "tag")
		return nil
	}
	if p.queues == nil {
		p.queues = make(map[string]*mergeQueue)
	}
	tag := sample.Tag(p.MergeTag)
	queue, ok := p.queues[tag]
	if !ok {
		queue = new(mergeQueue)
		queue.queue.Init()
		p.queues[tag] = queue
	}

	queue.push(sample, header)
	// Check if the new sample completes its queue, i.e. was recorded after the end of the current merge interval
	if p.canMergeSamples() && sample.Time.After(p.end) {
		p.readyQueuesLock.Lock()
		p.readyQueues[tag] = true
		allReady := len(p.readyQueues) >= len(p.queues)
		p.readyQueuesLock.Unlock()
		if allReady {
			// All queues have enough data, so output a sample and flush queues.
			sample, header = p.createSample()
			if sample != nil {
				if p.DebugQueueLengths {
					p.logQueueLengths()
				}
				return p.NoopProcessor.Sample(sample, header)
			} else if p.DebugWaitingQueues {
				p.logWaitingQueues()
			}
		}
	}
	return nil
}

func (p *SynchronizedSampleMerger) Close() {
	for {
		if !p.canMergeSamples() {
			break
		}
		sample, header := p.createSample()
		if sample == nil {
			break
		}
		if err := p.NoopProcessor.Sample(sample, header); err != nil {
			p.Error(err)
			break
		}
	}
	p.CloseSink()
}

func (p *SynchronizedSampleMerger) StreamClosed(name string) {
	p.readyQueuesLock.Lock()
	defer p.readyQueuesLock.Unlock()
	if queue, ok := p.queues[name]; ok {
		queue.closed = true
		p.readyQueues[name] = true
	}
}

func (p *SynchronizedSampleMerger) StreamOfSampleClosed(lastSample *bitflow.Sample, lastHeader *bitflow.Header) {
	p.StreamClosed(lastSample.Tag("client"))
}

func (p *SynchronizedSampleMerger) String() string {
	return p.Description
}

func (p *SynchronizedSampleMerger) logQueueLengths() {
	min := math.MaxInt32
	max := 0
	var avg float64
	for _, q := range p.queues {
		l := q.queue.Len()
		avg += float64(l) / float64(len(p.queues))
		if max < l {
			max = l
		}
		if min > l {
			min = l
		}
	}
	log.Printf("%v: Outputting merged sample, avg queue length: %v (min: %v max: %v)", p, avg, min, max)
}

func (p *SynchronizedSampleMerger) logWaitingQueues() {
	waitingQueues := make([]string, 0, len(p.queues))
	for name := range p.queues {
		if _, ok := p.readyQueues[name]; !ok {
			waitingQueues = append(waitingQueues, name)
		}
	}
	log.Printf("Waiting for %v queue(s): %v", len(waitingQueues), waitingQueues)
}

func (s *SynchronizedSampleMerger) canMergeSamples() bool {
	if len(s.queues) < s.ExpectedStreams {
		// Still waiting for every input stream to deliver at least one sample
		return false
	}
	if s.readyQueues == nil {
		// Lazy initialize
		s.readyQueues = make(map[string]bool)
	}
	if s.start.IsZero() {
		// We now received one sample on each queue, initialize the first merge window.
		s.storeStart()
		return false
	}
	return true
}

func (s *SynchronizedSampleMerger) createSample() (*bitflow.Sample, *bitflow.Header) {
	queueElements := s.collectQueuedSamples(s.queues)
	if len(queueElements) == 0 {
		return nil, nil
	}
	samples := make([]*bitflow.Sample, len(queueElements))
	headers := make([]*bitflow.Header, len(queueElements))
	for i, elem := range queueElements {
		samples[i] = elem.sample
		headers[i] = elem.header
	}
	outSample, outHeader := s.MergeSamples(samples, headers)

	// Store the start + end times for the next interval, record which queues already have enough data
	s.storeStart()
	s.readyQueuesLock.Lock()
	defer s.readyQueuesLock.Unlock()
	s.readyQueues = make(map[string]bool)
	for queueName, queue := range s.queues {
		// Some queues might already have enough samples for the next interval
		isReady := queue.closed
		if !isReady {
			newest := queue.peekNewest()
			if newest.sample != nil && newest.sample.Time.After(s.end) {
				isReady = true
			}
		}
		if isReady {
			s.readyQueues[queueName] = true
		}
	}
	return outSample, outHeader
}

func (s *SynchronizedSampleMerger) collectQueuedSamples(queues map[string]*mergeQueue) []queueElem {
	res := s.samples[:0]
	for _, queue := range queues {
		for {
			sample := queue.popYoungerThan(s.end)
			if sample.sample == nil {
				break
			}
			res = append(res, sample)
		}
	}
	s.samples = res // Reuse (possibly extended buffer) next time
	return res
}

func (s *SynchronizedSampleMerger) storeStart() {
	s.start = time.Time{}
	s.end = time.Time{}
	for _, queue := range s.queues {
		sample := queue.peek().sample
		if sample != nil && (s.start.IsZero() || sample.Time.Before(s.start)) {
			s.start = sample.Time
		}
	}
	s.end = s.start.Add(s.MergeInterval)
}

type mergeQueue struct {
	queue  list.List
	closed bool
}

type queueElem struct {
	sample *bitflow.Sample
	header *bitflow.Header
}

func (m *mergeQueue) push(sample *bitflow.Sample, header *bitflow.Header) {
	m.queue.PushBack(queueElem{sample, header})
}

func (m *mergeQueue) peek() queueElem {
	if elem := m.queue.Front(); elem != nil {
		return elem.Value.(queueElem)
	}
	return queueElem{}
}

func (m *mergeQueue) peekNewest() queueElem {
	if elem := m.queue.Back(); elem != nil {
		return elem.Value.(queueElem)
	}
	return queueElem{}
}

func (m *mergeQueue) popYoungerThan(endTime time.Time) queueElem {
	peekedElem := m.queue.Front()
	if peekedElem == nil {
		return queueElem{}
	}
	peeked := peekedElem.Value.(queueElem)
	if !peeked.sample.Time.After(endTime) {
		m.queue.Remove(peekedElem)
		return peeked
	}
	return queueElem{}
}
