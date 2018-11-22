package steps

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type HttpTagger struct {
	bitflow.NoopProcessor
	lock            sync.RWMutex
	currentHttpTags map[string]string
	listenEndpoint  string
}

func NewHttpTagger(pathPrefix string, r *mux.Router) *HttpTagger {
	tagger := &HttpTagger{
		currentHttpTags: make(map[string]string),
		listenEndpoint:  pathPrefix,
	}
	r.HandleFunc(pathPrefix+"/tag/{name}", tagger.handleTagRequest).Methods("GET", "DELETE")
	r.HandleFunc(pathPrefix+"/tags", tagger.handleTagsRequest).Methods("GET", "PUT", "POST", "DELETE")
	return tagger
}

func NewStandaloneHttpTagger(pathPrefix string, endpoint string) *HttpTagger {
	router := mux.NewRouter()
	tagger := NewHttpTagger(pathPrefix, router)
	tagger.listenEndpoint = endpoint + pathPrefix
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

func RegisterHttpTagger(b reg.ProcessorRegistry) {
	create := func(p *bitflow.SamplePipeline, params map[string]string) error {
		tagger := NewStandaloneHttpTagger("/api", params["listen"])
		if defaultTagStr, ok := params["default"]; ok {
			parts := strings.Split(defaultTagStr, ",")
			for _, part := range parts {
				tagParts := strings.Split(part, "=")
				if len(tagParts) != 2 {
					return reg.ParameterError("default", errors.New("Format must be 'key1=value1,key2=value2,...'. Received: "+defaultTagStr))
				}
				tagger.currentHttpTags[tagParts[0]] = tagParts[1]
			}
		}
		p.Add(tagger)
		return nil
	}

	b.RegisterAnalysisParamsErr("listen_tags", create, "Listen for HTTP requests on the given port at /api/tag and /api/tags to configure tags", reg.RequiredParams("listen"), reg.OptionalParams("default"))
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
	tagger.lock.Lock()
	defer tagger.lock.Unlock()
	return fmt.Sprintf("HTTP tagger on %v (tags: %v)", tagger.listenEndpoint, tagger.currentHttpTags)
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
