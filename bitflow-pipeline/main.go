package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/bitflow-script/reg"
	"github.com/bitflow-stream/go-bitflow-pipeline/bitflow-script/script"
	"github.com/bitflow-stream/go-bitflow-pipeline/bitflow-script/script_go"
	"github.com/bitflow-stream/go-bitflow-pipeline/clustering/dbscan"
	"github.com/bitflow-stream/go-bitflow-pipeline/clustering/denstream"
	"github.com/bitflow-stream/go-bitflow-pipeline/evaluation"
	"github.com/bitflow-stream/go-bitflow-pipeline/http"
	"github.com/bitflow-stream/go-bitflow-pipeline/http_tags"
	"github.com/bitflow-stream/go-bitflow-pipeline/plugin"
	"github.com/bitflow-stream/go-bitflow-pipeline/recovery"
	"github.com/bitflow-stream/go-bitflow-pipeline/regression"
	"github.com/bitflow-stream/go-bitflow-pipeline/steps"
	log "github.com/sirupsen/logrus"
)

const (
	fileFlag            = "f"
	BitflowScriptSuffix = ".bf"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <flags> <bitflow script>\nAll flags must be defined before the first non-flag parameter.\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	fix_arguments(&os.Args)
	os.Exit(do_main())
}

func fix_arguments(argsPtr *[]string) {
	if n := golib.ParseHashbangArgs(argsPtr); n > 0 {
		// Insert -f before the script file, if necessary
		args := *argsPtr
		if args[n-1] != "-"+fileFlag && args[n-1] != "--"+fileFlag {
			args = append(args, "") // Extend by one entry
			copy(args[n+1:], args[n:])
			args[n] = "-" + fileFlag
			*argsPtr = args
		}
	}
}

func do_main() int {
	printAnalyses := flag.Bool("print-analyses", false, "Print a list of available analyses and exit.")
	printPipeline := flag.Bool("print-pipeline", false, "Print the parsed pipeline and exit. Can be used to verify the input script.")
	printCapabilities := flag.Bool("capabilities", false, "Print the capabilities of this pipeline in JSON form and exit.")
	useNewScript := flag.Bool("new", false, "Use the new script parser for processing the input script.")
	scriptFile := ""
	flag.StringVar(&scriptFile, fileFlag, "", "File to read a Bitflow script from (alternative to providing the script on the command line)")

	registry := reg.NewProcessorRegistry()
	plugin.RegisterPluginDataSource(&registry.Endpoints)
	register_analyses(registry)
	bitflow.RegisterGolibFlags()
	registry.Endpoints.RegisterFlags()
	flag.Parse()
	golib.ConfigureLogging()
	if *printCapabilities {
		registry.PrintJsonCapabilities(os.Stdout)
		return 0
	}
	if *printAnalyses {
		fmt.Printf("Available analysis steps:\n%v\n", registry.PrintAllAnalyses())
		return 0
	}

	rawScript, err := get_script(flag.Args(), scriptFile)
	golib.Checkerr(err)
	make_pipeline := make_pipeline_old
	if *useNewScript {
		log.Println("Running using new ANTLR script implementation")
		make_pipeline = make_pipeline_new
	}
	pipe, err := make_pipeline(registry, rawScript)
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

func get_script(parsedArgs []string, scriptFile string) (string, error) {
	if scriptFile != "" && len(parsedArgs) > 0 {
		return "", errors.New("Please provide a bitflow pipeline script either via -f or as parameter, not both.")
	}
	if len(parsedArgs) == 1 && strings.HasSuffix(parsedArgs[0], BitflowScriptSuffix) {
		// Special case when passing a single existing .bf file as positional argument: Treat it as a script file
		info, err := os.Stat(parsedArgs[0])
		if err == nil && info.Mode().IsRegular() {
			scriptFile = parsedArgs[0]
		}
	}
	var rawScript string
	if scriptFile != "" {
		scriptBytes, err := ioutil.ReadFile(scriptFile)
		if err != nil {
			return "", fmt.Errorf("Error reading bitflow script file %v: %v", scriptFile, err)
		}
		rawScript = string(scriptBytes)
	} else {
		rawScript = strings.TrimSpace(strings.Join(parsedArgs, " "))
	}
	if rawScript == "" {
		return "", errors.New("Please provide a bitflow pipeline script via -f or directly as parameter.")
	}
	return rawScript, nil
}

func make_pipeline_old(registry reg.ProcessorRegistry, scriptStr string) (*pipeline.SamplePipeline, error) {
	queryBuilder := script_go.PipelineBuilder{registry}
	parser := script_go.NewParser(bytes.NewReader([]byte(scriptStr)))
	pipe, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return queryBuilder.MakePipeline(pipe)
}

func make_pipeline_new(registry reg.ProcessorRegistry, scriptStr string) (*pipeline.SamplePipeline, error) {
	s, err := script.NewAntlrBitflowParser(registry).ParseScript(scriptStr)
	return s, err.NilOrError()
}

func register_analyses(b reg.ProcessorRegistry) {

	// Control flow
	steps.RegisterNoop(b)
	steps.RegisterDrop(b)
	steps.RegisterSleep(b)
	steps.RegisterForks(b)
	steps.RegisterExpression(b)
	steps.RegisterSubprocessRunner(b)
	steps.RegisterMergeHeaders(b)
	steps.RegisterGenericBatch(b)
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
	steps.RegisterDuplicateTimestampFilter(b)

	// Reorder samples
	steps.RegisterConvexHullSort(b)
	steps.RegisterSampleShuffler(b)
	steps.RegisterSampleSorter(b)

	// Metadata
	steps.RegisterSetCurrentTime(b)
	steps.RegisterTaggingProcessor(b)
	http_tags.RegisterHttpTagger(b)
	steps.RegisterTargetTagSplitter(b)
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
	recovery.RegisterEventSimilarityEvaluation(b)
}
