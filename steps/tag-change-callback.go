package steps

import (
	"fmt"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

func AddTagChangeListenerParams(step *reg.RegisteredStep) {
	step.
		Required("tag", reg.String()).
		Optional("update-after", reg.Duration(), 2*time.Minute).
		Optional("expire-after", reg.Duration(), 15*time.Second).
		Optional("expire-on-close", reg.Bool(), false)
}

type TagChangeCallback interface {
	Expired(value string, allValues []string) error
	Updated(value string, sample *bitflow.Sample, allValues []string) error
}

type TagChangeListener struct {
	bitflow.AbstractSampleProcessor
	Tag               string
	UpdateInterval    time.Duration
	ExpirationTimeout time.Duration
	ExpireOnClose     bool
	Callback          TagChangeCallback

	loop        golib.LoopTask
	lastUpdated map[string]time.Time
	lastSeen    map[string]time.Time
	lock        sync.Mutex
}

func (t *TagChangeListener) ReadParameters(params map[string]interface{}) {
	t.Tag = params["tag"].(string)
	t.UpdateInterval = params["update-after"].(time.Duration)
	t.ExpirationTimeout = params["expire-after"].(time.Duration)
	t.ExpireOnClose = params["expire-on-close"].(bool)
}

func (t *TagChangeListener) String() string {
	return fmt.Sprintf("TagChangeListener (tag: %v, expiration: %v)", t.Tag, t.ExpirationTimeout)
}

func (t *TagChangeListener) Start(wg *sync.WaitGroup) golib.StopChan {
	t.lastUpdated = make(map[string]time.Time)
	t.lastSeen = make(map[string]time.Time)
	t.loop.Description = "Loop task of " + t.String()
	t.loop.Loop = t.loopExpireTagValues
	if t.ExpireOnClose {
		t.loop.StopHook = t.expireAllTagValues
	}
	return t.loop.Start(wg)
}

func (t *TagChangeListener) Close() {
	t.loop.Stop()
	t.AbstractSampleProcessor.CloseSink()
}

func (t *TagChangeListener) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	value := sample.Tag(t.Tag)
	if value != "" {
		// TODO instead of doing this here (potentially long-running), put it into a separate Goroutine
		t.handleTagValueUpdated(value, sample)
	}
	return t.AbstractSampleProcessor.GetSink().Sample(sample, header)
}

func (t *TagChangeListener) loopExpireTagValues(stop golib.StopChan) error {
	t.expireTagValues(stop)
	stop.WaitTimeout(t.ExpirationTimeout)
	return nil
}

func (t *TagChangeListener) expireTagValues(stop golib.StopChan) {
	t.lock.Lock()
	defer t.lock.Unlock()

	now := time.Now()
	for val, lastSeen := range t.lastSeen {
		diff := now.Sub(lastSeen)
		if diff >= t.ExpirationTimeout {
			log.Printf("Samples with tag %v=%v have not been seen for %v, expiring...", t.Tag, val, diff)
			t.handleTagValueExpired(val)
		}
		if stop.Stopped() {
			return
		}
	}
}

func (t *TagChangeListener) expireAllTagValues() {
	t.lock.Lock()
	defer t.lock.Unlock()

	log.Printf("Expiring %v value(s) of tag %v...", len(t.lastSeen), t.Tag)
	for val := range t.lastSeen {
		log.Printf("Expiring tag value %v=%v...", t.Tag, val)
		t.handleTagValueExpired(val)
	}
}

// t.lock must be locked when calling this
func (t *TagChangeListener) handleTagValueExpired(tagValue string) {
	if err := t.Callback.Expired(tagValue, t.getCurrentTags(tagValue, false)); err != nil {
		log.Errorf("Failed to handle expiration of tag %v=%v: %v", t.Tag, tagValue, err)
	} else {
		delete(t.lastUpdated, tagValue)
		delete(t.lastSeen, tagValue)
	}
}

func (t *TagChangeListener) handleTagValueUpdated(tagValue string, sample *bitflow.Sample) {
	last, ok := t.lastUpdated[tagValue]
	now := time.Now()
	updating := !ok || t.UpdateInterval <= 0 || now.Sub(last) > t.UpdateInterval
	if updating {
		err := t.Callback.Updated(tagValue, sample, t.getCurrentTags(tagValue, true))
		if err != nil {
			log.Errorln(err)
		}
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	t.lastSeen[tagValue] = now
	if updating {
		t.lastUpdated[tagValue] = now
	}
}

func (t *TagChangeListener) getCurrentTags(newTag string, added bool) []string {
	result := make([]string, 0, len(t.lastSeen)+1)
	for tag := range t.lastSeen {
		if !added && tag == newTag {
			continue
		}
		result = append(result, tag)
	}
	if added {
		result = append(result, newTag)
	}
	return result
}
