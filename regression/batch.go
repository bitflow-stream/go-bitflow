package regression

import (
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/go-bitflow"
)

type LinearRegressionBatchProcessor struct {
}

func (reg *LinearRegressionBatchProcessor) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	regression, err := NewLinearRegression(header, header.Fields)
	if err != nil {
		return nil, nil, err
	}
	allSamples := make([]bitflow.Sample, len(samples))
	for i, sample := range samples {
		allSamples[i] = *sample
	}
	if err := regression.Fit(allSamples); err != nil {
		return nil, nil, err
	}
	if mse, err := regression.MeanSquaredError(regression.TrainData); err != nil {
		log.Warnln("Failed to evaluate trained regression (%v): %v", regression.FormulaString(), err)
	} else {
		log.Printf("Linear Regression MSE %g: %v", mse, regression.FormulaString())
	}
	return header, samples, nil
}

func (reg *LinearRegressionBatchProcessor) String() string {
	return "Linear Regression"
}

type EvaluatedLinearRegression struct {
	LinearRegression
	MSE float64
}

type LinearRegressionBruteForce struct {
	invalidRegressions uint64
	numCombinations    uint64
	resultChan         chan EvaluatedLinearRegression
	results            SortedRegressions
}

func (brute *LinearRegressionBruteForce) String() string {
	return "Linear Regression Brute Force"
}

func (brute *LinearRegressionBruteForce) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	allSamples := make([]bitflow.Sample, len(samples))
	for i, sample := range samples {
		allSamples[i] = *sample
	}

	num_routines := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(num_routines + 1)
	varCombinations := make(chan []int, num_routines*5)
	brute.resultChan = make(chan EvaluatedLinearRegression, num_routines*2)
	for i := 0; i < num_routines; i++ {
		go brute.computeRegressions(&wg, varCombinations, header, allSamples)
	}
	go brute.generateVarCombinations(&wg, header, varCombinations)
	var resultWg sync.WaitGroup
	resultWg.Add(1)
	go brute.handleResults(&resultWg)
	wg.Wait()
	close(brute.resultChan)
	resultWg.Wait()

	for _, reg := range brute.results {
		log.Printf("MSE %g: %v", reg.MSE, reg.FormulaString())
	}
	log.Println("Computed", brute.numCombinations, "regressions,", brute.invalidRegressions, "ignored,", len(brute.results), "valid")

	return header, samples, nil
}

func (brute *LinearRegressionBruteForce) generateVarCombinations(wg *sync.WaitGroup, header *bitflow.Header, varCombinations chan<- []int) {
	defer wg.Done()
	// Generate unique pairs of header fields
	// TODO try other approaches
	for i := range header.Fields {
		for j := i + 1; j < len(header.Fields); j++ {
			varCombinations <- []int{i, j}
			atomic.AddUint64(&brute.numCombinations, 1)
		}
	}
	close(varCombinations)
}

func (brute *LinearRegressionBruteForce) computeRegressions(wg *sync.WaitGroup, varChan <-chan []int, header *bitflow.Header, samples []bitflow.Sample) {
	defer wg.Done()
	for vars := range varChan {
		var reg LinearRegression
		reg.Header = *header
		reg.Vars = vars
		if err := reg.Fit(samples); err != nil {
			log.Warnln("Failed to fit regression (%v): %v", vars, err)
			continue
		}
		if !reg.IsValid() {
			atomic.AddUint64(&brute.invalidRegressions, 1)
			continue
		}
		if mse, err := reg.MeanSquaredError(reg.TrainData); err != nil {
			log.Warnln("Failed to evaluate trained regression (%v, %v): %v", vars, reg.FormulaString(), err)
			continue
		} else {
			brute.resultChan <- EvaluatedLinearRegression{reg, mse}
		}
	}
}

func (brute *LinearRegressionBruteForce) handleResults(wg *sync.WaitGroup) {
	defer wg.Done()
	for result := range brute.resultChan {
		result.TrainData = nil // Release input data since MSE is already computed
		brute.results = append(brute.results, result)
		sort.Sort(brute.results)
	}
}

type SortedRegressions []EvaluatedLinearRegression

func (slice SortedRegressions) Len() int {
	return len(slice)
}

func (slice SortedRegressions) Less(i, j int) bool {
	return slice[i].MSE < slice[j].MSE
}

func (slice SortedRegressions) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
