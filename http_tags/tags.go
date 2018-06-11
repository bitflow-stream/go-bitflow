package http_tags

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/antongulenko/go-bitflow"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type HttpTagger struct {
	bitflow.NoopProcessor
	lock            sync.RWMutex
	currentHttpTags map[string]string
}

func NewHttpTagger(pathPrefix string, r *mux.Router) *HttpTagger {
	tagger := &HttpTagger{
		currentHttpTags: make(map[string]string),
	}
	r.HandleFunc(pathPrefix+"/tag/{name}", tagger.handleTagRequest).Methods("GET", "DELETE")
	r.HandleFunc(pathPrefix+"/tags", tagger.handleTagsRequest).Methods("GET", "PUT", "POST", "DELETE")
	return tagger
}

func NewStandaloneHttpTagger(pathPrefix string, endpoint string) *HttpTagger {
	router := mux.NewRouter()
	tagger := NewHttpTagger(pathPrefix, router)
	server := http.Server{
		Addr:    endpoint,
		Handler: router,
	}
	// Do not add this routine to any wait group, as it cannot be stopped
	go func() {
		tagger.Error(server.ListenAndServe())
	}()
	return tagger
}

func (tagger *HttpTagger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if current := tagger.currentHttpTags; current != nil {
		for tag, val := range current {
			sample.SetTag(tag, val)
		}
	}
	return tagger.NoopProcessor.Sample(sample, header)
}

func (tagger *HttpTagger) String() string {
	return "HTTP tagger"
}

func (tagger *HttpTagger) respondJson(w http.ResponseWriter, code int, obj interface{}) {
	content, err := json.Marshal(obj)
	if err != nil {
		data := []byte(fmt.Sprintf("Failed to marshall %T response to JSON (%v): %v", obj, err, obj))
		tagger.respond(w, http.StatusInternalServerError, data)
	} else {
		tagger.respond(w, code, content)
	}
}

func (tagger *HttpTagger) respondStr(w http.ResponseWriter, code int, data string) {
	tagger.respond(w, code, []byte(data))
}

func (tagger *HttpTagger) respond(w http.ResponseWriter, code int, data []byte) {
	w.WriteHeader(code)
	w.Write(data)
	w.Write([]byte("\n"))
}

func (tagger *HttpTagger) setMap(newMap map[string]string) map[string]string {
	// Always copy the map so that locking is not necessary for read operations
	tagger.lock.Lock()
	previous := tagger.currentHttpTags
	tagger.currentHttpTags = newMap
	tagger.lock.Unlock()
	log.Println("Setting tags to", newMap, "previously:", previous)
	return previous
}

func (tagger *HttpTagger) modifyMap(modify func(map[string]string)) {
	// Always copy the map so that locking is not necessary for read operations
	tagger.lock.Lock()
	defer tagger.lock.Unlock()
	newMap := make(map[string]string, len(tagger.currentHttpTags))
	for key, val := range tagger.currentHttpTags {
		newMap[key] = val
	}
	modify(newMap)
	log.Println("Setting tags to", newMap, "previously:", tagger.currentHttpTags)
	tagger.currentHttpTags = newMap
}

func (tagger *HttpTagger) handleTagRequest(w http.ResponseWriter, r *http.Request) {
	tags := tagger.currentHttpTags
	name := mux.Vars(r)["name"]
	val, ok := tags[name]
	if !ok {
		tagger.respondStr(w, http.StatusNotFound, "Tag not set: "+name)
		return
	}
	switch r.Method {
	case "GET":
	case "DELETE":
		tagger.modifyMap(func(m map[string]string) {
			delete(m, name)
		})
	}
	tagger.respondStr(w, http.StatusOK, val)
}

func (tagger *HttpTagger) handleTagsRequest(w http.ResponseWriter, r *http.Request) {
	fillMapFromRequest := func(m map[string]string, r *http.Request) {
		for key, val := range r.URL.Query() {
			if len(val) > 0 {
				m[key] = val[0]
			}
		}
	}

	var result map[string]string
	switch r.Method {
	case "GET":
		result = tagger.currentHttpTags
	case "PUT":
		tagger.modifyMap(func(tags map[string]string) {
			fillMapFromRequest(tags, r)
		})
		result = tagger.currentHttpTags
	case "POST":
		newTags := make(map[string]string)
		fillMapFromRequest(newTags, r)
		tagger.setMap(newTags)
		result = tagger.currentHttpTags
	case "DELETE":
		result = tagger.setMap(make(map[string]string))
	}
	tagger.respondJson(w, http.StatusOK, result)
}

func (tagger *HttpTagger) NumTags() int {
	return len(tagger.currentHttpTags)
}

func (tagger *HttpTagger) HasTags() bool {
	return tagger.NumTags() > 0
}

func (tagger *HttpTagger) Tags() map[string]string {
	return tagger.currentHttpTags
}
