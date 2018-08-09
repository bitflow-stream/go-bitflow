package rest_input

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline/http"
	"github.com/antongulenko/golib"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type RestEndpointFactory struct {
	endpoints map[string]*RestEndpoint
}

func (s *RestEndpointFactory) GetEndpoint(endpointString string) *RestEndpoint {
	if s.endpoints == nil {
		s.endpoints = make(map[string]*RestEndpoint)
	}
	endpoint, ok := s.endpoints[endpointString]
	if !ok {
		endpoint = &RestEndpoint{
			engine:   golib.NewGinEngine(),
			endpoint: endpointString,
		}
		s.endpoints[endpointString] = endpoint
	}
	return endpoint
}

type RestEndpoint struct {
	engine    *gin.Engine
	paths     []string
	startOnce sync.Once
	endpoint  string
	stoppers  []golib.StopChan
}

func (endpoint *RestEndpoint) start() {
	endpoint.startOnce.Do(func() {
		go func() {
			// This routine cannot be interrupted
			log.Println("Listening for REST input on", endpoint.endpoint)
			for _, path := range endpoint.paths {
				log.Println(" -", path)
			}
			if err := endpoint.engine.Run(endpoint.endpoint); err != nil {
				for _, stopper := range endpoint.stoppers {
					stopper.StopErr(err)
				}
			}
		}()
	})
}

func (endpoint *RestEndpoint) serve(verb string, path string, logFile string, serve func(*gin.Context)) {
	pathStr := fmt.Sprintf("[%s] %s", verb, path)
	// TODO check if pathStr is already present in paths, raise error if so

	handlers := gin.HandlersChain{serve}
	if logFile != "" {
		handlers = append(handlers, plotHttp.LogGinRequests(logFile, true))
	}
	endpoint.engine.Handle(verb, path, handlers...)
	endpoint.paths = append(endpoint.paths, pathStr)
}

func (endpoint *RestEndpoint) NewDataSource(outgoingSampleBuffer int) *RestDataSource {
	stopper := golib.NewStopChan()
	endpoint.stoppers = append(endpoint.stoppers, stopper)
	return &RestDataSource{
		outgoing: make(chan bitflow.SampleAndHeader, outgoingSampleBuffer),
		endpoint: endpoint,
		stop:     stopper,
	}
}

type RestDataSource struct {
	bitflow.AbstractSampleSource

	outgoing chan bitflow.SampleAndHeader
	stop     golib.StopChan
	endpoint *RestEndpoint
	wg       *sync.WaitGroup
}

func (source *RestDataSource) Endpoint() string {
	return source.endpoint.endpoint
}

func (source *RestDataSource) Serve(verb string, path string, httpLogFile string, serve func(*gin.Context)) {
	source.endpoint.serve(verb, path, httpLogFile, serve)
}

func (source *RestDataSource) EmitSample(sample *bitflow.Sample, header *bitflow.Header) {
	source.outgoing <- bitflow.SampleAndHeader{sample, header}
}

func (source *RestDataSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.wg = wg
	wg.Add(1)
	go source.sinkSamples(wg)
	source.endpoint.start()
	return source.stop
}

func (source *RestDataSource) Close() {
	source.stop.StopFunc(func() {
		close(source.outgoing)
		source.AbstractSampleSource.CloseSinkParallel(source.wg)
	})
}

func (source *RestDataSource) sinkSamples(wg *sync.WaitGroup) {
	defer wg.Done()
	for sample := range source.outgoing {
		if err := source.GetSink().Sample(sample.Sample, sample.Header); err != nil {
			log.Errorln("Error sinking sample:", err)
		}
	}
}

type RestReplyHelpers struct {
}

func (h RestReplyHelpers) ReplySuccess(context *gin.Context, message string) {
	context.Writer.Header().Set("Content-Type", "text/plain")
	_, err := context.Writer.WriteString(message + "\n")
	if err != nil {
		log.Errorln("REST: Error sending success-reply to client:", err)
	}
}

func (h RestReplyHelpers) ReplyGenericError(context *gin.Context, errorMessage string) {
	log.Warnln("REST:", errorMessage)
	_, err := context.Writer.WriteString(errorMessage + "\n")
	context.Writer.WriteHeader(http.StatusBadRequest)
	if err != nil {
		log.Errorln("REST: Error sending error-reply to client:", err)
	}
}

func (h RestReplyHelpers) ReplyError(context *gin.Context, errorMessage string) {
	h.ReplyGenericError(context, "Wrong HTTP request: "+errorMessage)
}
