package pipeline

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

const form_timeout = "timeout"

type HttpTagger struct {
	bitflow.AbstractProcessor
	taggingTimeout  time.Time
	loopCond        *sync.Cond
	currentHttpTags map[string]string

	Port int
}

func (tagger *HttpTagger) Start(wg *sync.WaitGroup) golib.StopChan {
	tagger.loopCond = sync.NewCond(new(sync.Mutex))
	mux := http.NewServeMux()
	mux.HandleFunc("/tag", tagger.handleTaggingRequest)
	server := http.Server{
		Addr:    ":" + strconv.Itoa(tagger.Port),
		Handler: mux,
	}

	// These routines are not added to the WaitGroup, because the server is not interruptible
	go tagger.checkTimeoutLoop()
	go func() {
		tagger.Error(server.ListenAndServe())
	}()
	return tagger.AbstractProcessor.Start(wg)
}

func (tagger *HttpTagger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	// The tags map is always rebuillt from scratch, no need to lock, just take the current reference
	if current := tagger.currentHttpTags; current != nil {
		header.HasTags = true
		for tag, val := range current {
			sample.SetTag(tag, val)
		}
	}
	return tagger.AbstractProcessor.Sample(sample, header)
}

func (tagger *HttpTagger) String() string {
	return fmt.Sprintf("HTTP tagger listening on port %v", tagger.Port)
}

func (tagger *HttpTagger) checkTimeoutLoop() {
	loopCond := tagger.loopCond
	loopCond.L.Lock()
	defer loopCond.L.Unlock()
	for {
		t := tagger.taggingTimeout.Sub(time.Now())
		if t <= 0 {
			if tags := tagger.currentHttpTags; tags != nil {
				log.Println("Timeout: Unsetting tags, previously:", tags)
				tagger.currentHttpTags = nil
			}
		} else {
			time.AfterFunc(t, func() {
				loopCond.L.Lock()
				defer loopCond.L.Unlock()
				loopCond.Broadcast()
			})
		}
		loopCond.Wait()
	}
}

func (tagger *HttpTagger) handleTaggingRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		tagger.handleTaggingPostRequest(w, r)
	case "GET":
		tagger.handleTaggingGetRequest(w, r)
	default:
		w.WriteHeader(405)
		w.Write([]byte("Only POST method is allowed, not " + r.Method))
	}
}

func (tagger *HttpTagger) handleTaggingGetRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(tagger.CurrentTagsStr() + "\n"))
}

func (tagger *HttpTagger) handleTaggingPostRequest(w http.ResponseWriter, r *http.Request) {
	timeout := r.FormValue(form_timeout)
	if timeout == "" {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Need URL parameter %s", form_timeout)))
		return
	}
	timeoutSec, err := strconv.Atoi(timeout)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Failed to parse %s parameter: %v", form_timeout, err)))
		return
	}
	newTags := make(map[string]string)
	for key, val := range r.Form {
		if key != form_timeout && len(val) > 0 {
			newTags[key] = val[0]
		}
	}

	tagger.loopCond.L.Lock()
	defer tagger.loopCond.L.Unlock()
	log.Println("Setting tags to", newTags, "previously:", tagger.currentHttpTags)
	tagger.currentHttpTags = newTags
	taggingDuration := time.Duration(timeoutSec) * time.Second
	tagger.taggingTimeout = time.Now().Add(taggingDuration)
	tagger.loopCond.Broadcast()

	w.WriteHeader(200)
	var content string
	if taggingDuration == 0 {
		content = "Unsetting tags"
	} else {
		content = fmt.Sprintf("%v for %v (until %v)", tagger.CurrentTagsStr(), taggingDuration, tagger.taggingTimeout.Format("2006-01-02 15:04:05.000"))
	}
	w.Write([]byte(content + "\n"))
}

func (tagger *HttpTagger) CurrentTagsStr() string {
	tags := tagger.currentHttpTags
	if tags == nil {
		return "Tags timed out or not configured"
	} else if len(tags) == 0 {
		return "No tags set"
	} else {
		return "Tags set to " + golib.FormatMap(tags)
	}
}

func (tagger *HttpTagger) HttpTagsSet() bool {
	return tagger.currentHttpTags != nil
}
