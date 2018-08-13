package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/clustering/dbscan"
	"github.com/antongulenko/go-bitflow-pipeline/clustering/denstream"
	"github.com/antongulenko/go-bitflow-pipeline/evaluation"
	"github.com/antongulenko/go-bitflow-pipeline/http"
	"github.com/antongulenko/go-bitflow-pipeline/http_tags"
	"github.com/antongulenko/go-bitflow-pipeline/plugin"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/go-bitflow-pipeline/recovery"
	"github.com/antongulenko/go-bitflow-pipeline/regression"
	"github.com/antongulenko/go-bitflow-pipeline/steps"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <flags> <bitflow script>\nAll flags must be defined before the first non-flag parameter.\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	os.Exit(do_main())
}

func do_main() int {
	printAnalyses := flag.Bool("print-analyses", false, "Print a list of available analyses and exit.")
	printPipeline := flag.Bool("print-pipeline", false, "Print the parsed pipeline and exit. Can be used to verify the input script.")
	printCapabilities := flag.Bool("capabilities", false, "Print the capablities of this pipeline in JSON form and exit.")
	useNewScript := flag.Bool("new", false, "Use the new script parser for processing the input script.")
	scriptFile := ""
	flag.StringVar(&scriptFile, "f", "", "File to read a Bitflow script from (alternative to providing the script on the command line)")

	builder := query.NewPipelineBuilder()
	plugin.RegisterPluginDataSource(&builder.Endpoints)
	register_analyses(builder)

	bitflow.RegisterGolibFlags()
	builder.Endpoints.RegisterFlags()
	flag.Parse()
	golib.ConfigureLogging()
	if *printCapabilities {
		builder.PrintJsonCapabilities(os.Stdout)
		return 0
	}
	if *printAnalyses {
		fmt.Printf("Available analysis steps:\n%v\n", builder.PrintAllAnalyses())
		return 0
	}

	script := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if scriptFile != "" && script != "" {
		golib.Fatalln("Please provide a bitflow pipeline script either via -f or as parameter, not both.")
	}
	if scriptFile != "" {
		scriptBytes, err := ioutil.ReadFile(scriptFile)
		if err != nil {
			golib.Fatalf("Error reading bitflow script file $v: %v", scriptFile, err)
		}
		script = string(scriptBytes)
	}
	if script == "" {
		golib.Fatalln("Please provide a bitflow pipeline script via -f or directly as parameter.")
	}

	var pipe *pipeline.SamplePipeline
	var err error
	if *useNewScript {
		pipe, err = make_pipeline_new(builder, script)
	} else {
		pipe, err = make_pipeline_old(builder, script)
	}
	if err != nil {
		log.Errorln(err)
		golib.Fatalln("Use -print-analyses to print all available analysis steps.")
	}
	defer golib.ProfileCpu()()
	for _, str := range pipe.FormatLines() {
		log.Println(str)
	}
	if *printPipeline {
		return 0
	}
	return pipe.StartAndWait()
}

func make_pipeline_old(builder *query.PipelineBuilder, script string) (*pipeline.SamplePipeline, error) {
	parser := query.NewParser(bytes.NewReader([]byte(script)))
	pipe, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return builder.MakePipeline(pipe)
}

func make_pipeline_new(builder *query.PipelineBuilder, script string) (*pipeline.SamplePipeline, error) {
	return nil, fmt.Errorf("The new script parser is not yet implemented")
}

func register_analyses(b *query.PipelineBuilder) {

	// Control flow
	RegisterTaggingAnalyses(b)
	steps.RegisterNoop(b)
	steps.RegisterSleep(b)
	steps.RegisterForks(b)
	steps.RegisterExpression(b)
	steps.RegisterSubprocessRunner(b)
	steps.RegisterMergeHeaders(b)
	steps.RegisterGenericBatch(b)
	steps.RegisterDecouple(b)
	steps.RegisterDropErrorsStep(b)
	steps.RegisterResendStep(b)
	steps.RegisterPipelineRateSynchronizer(b)
	steps.RegisterSubpipelineStreamMerger(b)
	blockMgr := steps.NewBlockManager()
	blockMgr.RegisterBlockingProcessor(b)
	blockMgr.RegisterReleasingProcessor(b)

	// Data output
	steps.RegisterOutputFiles(b)
	steps.RegisterGraphiteOutput(b)
	steps.RegisterOpentsdbOutput(b)

	// Logging, output metadata
	steps.RegisterStoreStats(b)
	steps.RegisterLoggingSteps(b)

	// Visualization
	plotHttp.RegisterHttpPlotter(b)
	steps.RegisterPlot(b)

	// Basic Math
	steps.RegisterFFT(b)
	steps.RegisterRMS(b)
	regression.RegisterLinearRegression(b)
	regression.RegisterLinearRegressionBruteForce(b)
	steps.RegisterPCA(b)
	steps.RegisterPCAStore(b)
	steps.RegisterPCALoad(b)
	steps.RegisterPCALoadStream(b)
	steps.RegisterMinMaxScaling(b)
	steps.RegisterStandardizationScaling(b)
	steps.RegisterAggregateAvg(b)
	steps.RegisterAggregateSlope(b)

	// Clustering & Evaluation
	dbscan.RegisterDbscan(b)
	dbscan.RegisterDbscanParallel(b)
	denstream.RegisterDenstream(b)
	denstream.RegisterDenstreamLinear(b)
	denstream.RegisterDenstreamBirch(b)
	evaluation.RegisterAnomalyClusterTagger(b)
	evaluation.RegisterCitTagsPreprocessor(b)
	evaluation.RegisterAnomalySmoothing(b)
	evaluation.RegisterEventEvaluation(b)
	evaluation.RegisterBinaryEvaluation(b)

	// Filter samples
	steps.RegisterFilterExpression(b)
	steps.RegisterPickPercent(b)
	steps.RegisterPickHead(b)
	steps.RegisterSkipHead(b)
	steps.RegisterConvexHull(b)

	// Reorder samples
	steps.RegisterConvexHullSort(b)
	steps.RegisterSampleShuffler(b)
	steps.RegisterSampleSorter(b)

	// Metadata
	steps.RegisterSetCurrentTime(b)
	steps.RegisterTaggingProcessor(b)
	http_tags.RegisterHttpTagger(b)
	steps.RegisterInjectionInfoTagger(b)
	steps.RegisterPauseTagger(b)

	// Add/Remove/Rename/Reorder generic metrics
	steps.RegisterParseTags(b)
	steps.RegisterStripMetrics(b)
	steps.RegisterMetricMapper(b)
	steps.RegisterMetricRenamer(b)
	steps.RegisterIncludeMetricsFilter(b)
	steps.RegisterExcludeMetricsFilter(b)
	steps.RegisterVarianceMetricsFilter(b)

	// Special
	steps.RegisterSphere(b)
	steps.RegisterAppendTimeDifference(b)
	recovery.RegisterRecoveryEngine(b)
}
