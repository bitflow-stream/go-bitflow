package bitflow_plugin_default_steps

import (
	"log"

	"github.com/bitflow-stream/go-bitflow/script/plugin"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/bitflow-stream/go-bitflow/steps"
	"github.com/bitflow-stream/go-bitflow/steps/math"
	"github.com/bitflow-stream/go-bitflow/steps/plot"
)

// This plugin is automatically loaded by the bitflow-pipeline tool, there is no need to actually compile
// it as a plugin.
// TODO in the future, it would be nice to add a mechanism for automatically build and discover plugins and turn this into a regular plugin
// If this is implemented, changes this package name to 'main'
func main() {
	log.Fatalln("This package is intended to be loaded as a plugin, not executed directly")
}

var Plugin plugin.BitflowPlugin = new(pluginImpl)

type pluginImpl struct {
}

func (*pluginImpl) Name() string {
	return "Default pipeline steps"
}

func (p *pluginImpl) Init(b reg.ProcessorRegistry) error {

	// Control flow
	steps.RegisterNoop(b)
	steps.RegisterDrop(b)
	steps.RegisterSleep(b)
	steps.RegisterForks(b)
	steps.RegisterExpression(b)
	steps.RegisterSubProcessRunner(b)
	steps.RegisterMergeHeaders(b)
	steps.RegisterDecouple(b)
	steps.RegisterDropErrorsStep(b)
	steps.RegisterResendStep(b)
	steps.RegisterFillUpStep(b)
	steps.RegisterPipelineRateSynchronizer(b)
	steps.RegisterSubpipelineStreamMerger(b)
	blockMgr := steps.NewBlockManager()
	blockMgr.RegisterBlockingProcessor(b)
	blockMgr.RegisterReleasingProcessor(b)
	steps.RegisterTagSynchronizer(b)
	steps.RegisterTagChangeRunner(b)

	// Data input
	steps.RegisterDynamicSource(b.Endpoints)
	steps.RegisterGeneratorSource(b.Endpoints)

	// Data output
	steps.RegisterConsoleBoxOutput(b.Endpoints)
	steps.RegisterOutputFiles(b)
	steps.RegisterGraphiteOutput(b)
	steps.RegisterOpentsdbOutput(b)

	// Data formats
	steps.RegisterPrometheusMarshaller(b.Endpoints)

	// Logging, output metadata
	steps.RegisterStoreStats(b)
	steps.RegisterLoggingSteps(b)

	// Visualization
	plot.RegisterHttpPlotter(b)
	plot.RegisterPlot(b)

	// Basic Math
	math.RegisterFFT(b)
	math.RegisterRMS(b)
	math.RegisterPCA(b)
	math.RegisterPCAStore(b)
	math.RegisterPCALoad(b)
	math.RegisterPCALoadStream(b)
	math.RegisterMinMaxScaling(b)
	math.RegisterStandardizationScaling(b)
	math.RegisterAggregateAvg(b)
	math.RegisterAggregateSlope(b)
	math.RegisterBatchFeatureStatsAggregators(b)
	math.RegisterBatchAggregators(b)

	// Filter samples
	steps.RegisterFilterExpression(b)
	steps.RegisterPickPercent(b)
	steps.RegisterPickHead(b)
	steps.RegisterSkipHead(b)
	steps.RegisterPickTail(b)
	steps.RegisterDropInvalid(b)
	math.RegisterConvexHull(b)
	steps.RegisterDuplicateTimestampFilter(b)

	// Reorder samples
	math.RegisterConvexHullSort(b)
	steps.RegisterSampleShuffler(b)
	steps.RegisterSampleSorter(b)

	// Metadata
	steps.RegisterSetCurrentTime(b)
	steps.RegisterTaggingProcessor(b)
	steps.RegisterTagMapping(b)
	steps.RegisterHttpTagger(b)
	steps.RegisterPauseTagger(b)

	// Add/Remove/Rename/Reorder generic metrics
	steps.RegisterParseTags(b)
	steps.RegisterStripMetrics(b)
	steps.RegisterMetricMapper(b)
	steps.RegisterMetricRenamer(b)
	steps.RegisterIncludeMetricsFilter(b)
	steps.RegisterExcludeMetricsFilter(b)
	steps.RegisterVarianceMetricsFilter(b)
	steps.RegisterMetricSplitter(b)

	// Special
	math.RegisterSphere(b)
	steps.RegisterAppendTimeDifference(b)

	return nil
}
