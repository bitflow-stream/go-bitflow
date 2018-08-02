package main

import (
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
	"github.com/antongulenko/go-bitflow-pipeline/recovery"
	"github.com/antongulenko/go-bitflow-pipeline/regression"
	"github.com/antongulenko/go-bitflow-pipeline/steps"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
	"github.com/antongulenko/go-bitflow-pipeline/builder"
	"github.com/antongulenko/go-bitflow-pipeline/bitflowcli/script"
	"io"
	"encoding/json"
	"bytes"
)

var queryBuilder = script.NewProcessorRegistry()

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
	scriptFile := ""
	flag.StringVar(&scriptFile, "f", "", "File to read a Bitflow script from (alternative to providing the script on the command line)")

	plugin.RegisterPluginDataSource(&queryBuilder.Endpoints)
	register_analyses(queryBuilder)

	bitflow.RegisterGolibFlags()
	queryBuilder.Endpoints.RegisterFlags()
	flag.Parse()
	golib.ConfigureLogging()
	if *printCapabilities {
		PrintJsonCapabilities(queryBuilder, os.Stdout)
		return 0
	}
	if *printAnalyses {
		fmt.Printf("Available analysis steps:\n", )
		PrintAllAnalyses(queryBuilder)
		return 0
	}

	rawScript := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if scriptFile != "" && rawScript != "" {
		golib.Fatalln("Please provide a bitflow pipeline script either via -f or as parameter, not both.")
	}
	if scriptFile != "" {
		scriptBytes, err := ioutil.ReadFile(scriptFile)
		if err != nil {
			golib.Fatalf("Error reading bitflow script file %v: %v", scriptFile, err)
		}
		rawScript = string(scriptBytes)
	}
	if rawScript == "" {
		golib.Fatalln("Please provide a bitflow pipeline script via -f or directly as parameter.")
	}

	pipe, err := make_pipeline(rawScript)
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


func make_pipeline(rawScript string) (*pipeline.SamplePipeline, error) {
	s, err := script.NewAntlrBitflowParser(queryBuilder).ParseScript(rawScript)
	return s, err.NilOrError()
}

// Print
func PrintAllAnalyses(reg *script.ProcessorRegistry) {
	all := reg.GetAvailableProcessors()
	var buf bytes.Buffer
	for i, analysis := range all {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(" - ")
		buf.WriteString(analysis.Name)
		buf.WriteString(":\n")
		buf.WriteString("      ")
		buf.WriteString(analysis.Description)
	}
	fmt.Printf("Available analysis steps:\n%v\n", buf.String())
}

func PrintJsonCapabilities(reg *script.ProcessorRegistry, out io.Writer) error {
	all := reg.GetAvailableProcessors()
	data, err := json.Marshal(all)
	if err == nil {
		_, err = out.Write(data)
	}
	return err
}

func register_analyses(b builder.PipelineBuilder) {

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
}
