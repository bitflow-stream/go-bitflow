package denstream

import (
	"fmt"
	"strconv"
	"time"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/bitflow-script/reg"
	"github.com/antongulenko/go-bitflow-pipeline/clustering"
	log "github.com/sirupsen/logrus"
)

var _ bitflow.SampleProcessor = new(ClusterProcessor)

const (
	ClusterTag   = "cluster"
	RadiusMetric = "radius"
)

type ClusterProcessor struct {
	bitflow.NoopProcessor
	Clusterer

	// If set to >0, will log the denstream clusterer state every OutputStateModulo samples
	OutputStateModulo  int
	CreateClusterSpace func(numDimensions int) ClusterSpace

	// If set to true, this processing step will output all clusters as special samples when closing.
	// The first metric will be the radius of the cluster, followed by the cluster center in every dimension.
	FlushClustersAtClose bool

	TrainTag      string // If TrainTag is not empty, only certain samples actually modify the clusters. Other samples are tagged with the cluter ID.
	TrainTagValue string // If TrainTagValue is empty, samples that have the TrainTag tag are used for training (value irrelevant). Otherwise, the tag value must match TrainTagValue.

	numDimensions       int
	processedSamples    int
	lastProcessedSample *bitflow.Sample
	lastProcessedHeader *bitflow.Header
}

func (p *ClusterProcessor) String() string {
	return fmt.Sprintf("denstream (λ=%v, ε=%v, βµ=%v)", p.HistoryFading, p.Epsilon, p.MaxOutlierWeight)
}

func (p *ClusterProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if p.numDimensions == 0 {
		p.numDimensions = len(sample.Values)
		log.Printf("Initializing denstream processor to %v dimensions", p.numDimensions)
		p.pClusters = p.CreateClusterSpace(p.numDimensions)
		p.oClusters = p.CreateClusterSpace(p.numDimensions)
	} else if p.numDimensions != len(sample.Values) {
		return fmt.Errorf("Denstream has already been initialized to %v dimensions, cannot handle sample with %v dimensions",
			p.numDimensions, len(sample.Values))
	}

	point := make([]float64, len(sample.Values))
	for i, v := range sample.Values {
		point[i] = float64(v)
	}
	var clusterId int
	if p.shouldTrain(sample) {
		clusterId = p.Insert(point, sample.Time)
	} else {
		clusterId = p.GetCluster(point)
	}

	if p.OutputStateModulo > 0 && p.processedSamples%p.OutputStateModulo == 0 {
		log.Println("Denstream processed", p.processedSamples, "points in:", p.Clusterer.String())
	}
	p.processedSamples++

	sample.SetTag(ClusterTag, strconv.Itoa(clusterId))
	p.lastProcessedSample = sample
	p.lastProcessedHeader = header
	return p.NoopProcessor.Sample(sample, header)
}

func (p *ClusterProcessor) shouldTrain(sample *bitflow.Sample) bool {
	switch {
	case p.TrainTag == "":
		return true
	case !sample.HasTag(p.TrainTag):
		return false
	}
	return p.TrainTagValue == "" || sample.Tag(p.TrainTag) == p.TrainTagValue
}

func (p *ClusterProcessor) Close() {
	if p.FlushClustersAtClose && p.lastProcessedSample != nil {
		newFields := make([]string, len(p.lastProcessedHeader.Fields)+1)
		newFields[0] = RadiusMetric
		copy(newFields[1:], p.lastProcessedHeader.Fields)
		header := p.lastProcessedHeader.Clone(newFields)

		var err error
		p.oClusters.ClustersDo(func(cluster clustering.SphericalCluster) {
			p.outputCluster(cluster, "outlier", header, &err)
		})
		p.pClusters.ClustersDo(func(cluster clustering.SphericalCluster) {
			p.outputCluster(cluster, "real", header, &err)
		})
	}
	p.NoopProcessor.Close()
}

func (p *ClusterProcessor) outputCluster(cluster clustering.SphericalCluster, clusterType string, header *bitflow.Header, err *error) {
	if *err != nil {
		return
	}
	center := cluster.Center()
	values := make([]bitflow.Value, len(center)+1)
	values[0] = bitflow.Value(cluster.Radius())
	for i, centerDim := range center {
		values[i+1] = bitflow.Value(centerDim)
	}

	sample := &bitflow.Sample{
		Time:   p.lastProcessedSample.Time,
		Values: values,
	}
	sample.SetTag("cluster-type", clusterType)
	sample.SetTag("cluster-id", strconv.Itoa(cluster.Id()))
	*err = p.GetSink().Sample(sample, header)
	if *err != nil {
		log.Errorln("Error outputting cluster centers:", *err)
	}
}

func create_denstream_step(p *pipeline.SamplePipeline, params map[string]string, createClusterSpace func(numDimensions int) ClusterSpace) (err error) {
	eps := 0.1
	if epsStr, ok := params["eps"]; ok {
		eps, err = strconv.ParseFloat(epsStr, 64)
		if err != nil {
			err = reg.ParameterError("eps", err)
			return
		}

	}
	enableEpsTuning := false
	if enableEpsTuningStr, ok := params["enableEpsTuning"]; ok {
		enableEpsTuning, err = strconv.ParseBool(enableEpsTuningStr)
		if err != nil {
			err = reg.ParameterError("enableEpsTuning", err)
			return
		}
	}
	if enableEpsTuning == false && eps == 0.0 {
		eps = 0.1
	}

	epsColdStart := false
	if epsColdStartStr, ok := params["epsColdStart"]; ok {
		epsColdStart, err = strconv.ParseBool(epsColdStartStr)
		if err != nil {
			err = reg.ParameterError("epsColdStart", err)
			return
		}
	}
	if enableEpsTuning == true && epsColdStart == false && eps == 0.0 {
		eps = 0.1
	}
	// else if enableEpsTuning == true && epsColdStart == true {
	// 	eps = 0.0
	// }

	minNumClusters := 20
	if minNumClustersStr, ok := params["minNumClusters"]; ok {
		minNumClusters, err = strconv.Atoi(minNumClustersStr)
		if err != nil {
			err = reg.ParameterError("minNumClusters", err)
			return
		}
	}

	lambda := 0.0000000001 // 0.0001 -> Decay check every 37 minutes
	lambdaStr, hasLambda := params["lambda"]
	if hasLambda {
		lambda, err = strconv.ParseFloat(lambdaStr, 64)
		if err != nil {
			err = reg.ParameterError("lambda", err)
			return
		}
	}

	maxOutlierWeight := 5.0
	if maxOutlierWeightStr, ok := params["maxOutlierWeight"]; ok {
		maxOutlierWeight, err = strconv.ParseFloat(maxOutlierWeightStr, 64)
		if err != nil {
			err = reg.ParameterError("maxOutlierWeight", err)
			return
		}
	}

	debug := 0
	if debugStr, ok := params["debug"]; ok {
		debug, err = strconv.Atoi(debugStr)
		if err != nil {
			err = reg.ParameterError("debug", err)
			return
		}
	}

	outputClusters := false
	if outputClustersStr, ok := params["output-clusters"]; ok {
		outputClusters, err = strconv.ParseBool(outputClustersStr)
		if err != nil {
			err = reg.ParameterError("output-clusters", err)
			return
		}
	}

	clust := &ClusterProcessor{
		Clusterer: Clusterer{
			HistoryFading:    lambda,
			MaxOutlierWeight: maxOutlierWeight,
			Epsilon:          eps,
			EnableEpsTuning:  enableEpsTuning,
			ColdStart:        epsColdStart,
			MinNumClusters:   minNumClusters,
		},
		OutputStateModulo:    debug,
		FlushClustersAtClose: outputClusters,
		TrainTag:             params["trainTag"],
		TrainTagValue:        params["trainTagValue"],
		CreateClusterSpace:   createClusterSpace,
	}

	if decayTimeStr, ok := params["decay"]; ok {
		var decayTime time.Duration
		decayTime, err = time.ParseDuration(decayTimeStr)
		if err != nil {
			err = reg.ParameterError("decay", err)
			return
		}
		if hasLambda {
			return fmt.Errorf("Cannot define both 'lambda' and 'decay' parameters")
		}
		clust.SetDecayTimeUnit(decayTime)
	}

	p.Add(clust)
	return nil
}

var optionalParameters = []string{"eps", "lambda", "maxOutlierWeight", "debug", "decay", "output-clusters", "trainTag", "trainTagValue", "enableEpsTuning", "epsColdStart", "minNumClusters"}

func RegisterDenstream(b reg.ProcessorRegistry) {
	create := func(p *pipeline.SamplePipeline, params map[string]string) (err error) {
		return create_denstream_step(p, params, func(numDimensions int) ClusterSpace {
			return NewRtreeClusterSpace(numDimensions, 25, 50)
		})
	}
	b.RegisterAnalysisParamsErr("denstream_rtree", create, "Perform a denstream clustering on the data stream. Clusters organzied in r-tree.", reg.OptionalParams(optionalParameters...))
}

func RegisterDenstreamLinear(b reg.ProcessorRegistry) {
	create := func(p *pipeline.SamplePipeline, params map[string]string) (err error) {
		return create_denstream_step(p, params, func(numDimensions int) ClusterSpace {
			return NewLinearClusterSpace()
		})
	}
	b.RegisterAnalysisParamsErr("denstream_linear", create, "Perform a denstream clustering on the data stream. Clusters searched linearly.", reg.OptionalParams(optionalParameters...))
}

func RegisterDenstreamBirch(b reg.ProcessorRegistry) {
	create := func(p *pipeline.SamplePipeline, params map[string]string) (err error) {
		return create_denstream_step(p, params, func(numDimensions int) ClusterSpace {
			return NewBirchTreeClusterSpace(numDimensions)
		})
	}
	b.RegisterAnalysisParamsErr("denstream_birch", create, "Perform a denstream clustering on the data stream. Clusters managed by BIRCH tree structure.", reg.OptionalParams(optionalParameters...))
}
