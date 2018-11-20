package math

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/antongulenko/go-onlinestats"
	"github.com/antongulenko/golearn/base"
	"github.com/antongulenko/golearn/linear_models"
	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/reg"
	log "github.com/sirupsen/logrus"
)

type LinearRegression struct {
	SubHeader
	RegressionClassVar int
	Model              *linear_models.LinearRegression
	TrainData          *base.DenseInstances
}

type LinearRegressionBatchProcessor struct {
}

func (reg *LinearRegressionBatchProcessor) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	regression, err := NewLinearRegression(header, header.Fields)
	if err != nil {
		return nil, nil, err
	}
	if err := regression.Fit(samples); err != nil {
		return nil, nil, err
	}
	if mse, err := regression.MeanSquaredError(regression.TrainData); err != nil {
		log.Warnf("Failed to evaluate trained regression (%v): %v\n", regression.FormulaString(), err)
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
	num_routines := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(num_routines + 1)
	varCombinations := make(chan []int, num_routines*5)
	brute.resultChan = make(chan EvaluatedLinearRegression, num_routines*2)
	for i := 0; i < num_routines; i++ {
		go brute.computeRegressions(&wg, varCombinations, header, samples)
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

func (brute *LinearRegressionBruteForce) computeRegressions(wg *sync.WaitGroup, varChan <-chan []int, header *bitflow.Header, samples []*bitflow.Sample) {
	defer wg.Done()
	for vars := range varChan {
		var reg LinearRegression
		reg.Header = header
		reg.Vars = vars
		if err := reg.Fit(samples); err != nil {
			log.Warnf("Failed to fit regression (%v): %v\n", vars, err)
			continue
		}
		if !reg.IsValid() {
			atomic.AddUint64(&brute.invalidRegressions, 1)
			continue
		}
		if mse, err := reg.MeanSquaredError(reg.TrainData); err != nil {
			log.Warnf("Failed to evaluate trained regression (%v, %v): %v\n", vars, reg.FormulaString(), err)
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

func NewLinearRegression(header *bitflow.Header, fieldNames []string) (LinearRegression, error) {
	fieldNumbers := make([]int, len(fieldNames))
	var reg LinearRegression
	for i, searching := range fieldNames {
		found := false
		for j, field := range header.Fields {
			if field == searching {
				found = true
				fieldNumbers[i] = j
			}
		}
		if !found {
			return reg, fmt.Errorf("Could not find header field '%v'", searching)
		}
	}
	reg.Header = header
	reg.Vars = fieldNumbers
	return reg, nil
}

func RegisterLinearRegression(b reg.ProcessorRegistry) {
	create := func(p *pipeline.SamplePipeline) {
		p.Batch(new(LinearRegressionBatchProcessor))
	}
	b.RegisterAnalysis("regression", create, "Perform a linear regression analysis on a batch of samples", reg.SupportBatch())
}

func RegisterLinearRegressionBruteForce(b reg.ProcessorRegistry) {
	create := func(p *pipeline.SamplePipeline) {
		p.Batch(new(LinearRegressionBruteForce))
	}
	b.RegisterAnalysis("regression_brute", create, "In a batch of samples, perform a linear regression analysis for every possible combination of metrics", reg.SupportBatch())
}

func (reg *LinearRegression) FormulaString() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%v = %g", reg.Model.Cls.GetName(), reg.Model.Disturbance)
	coefficients := reg.Model.RegressionCoefficients
	for i, attr := range reg.Model.Attrs {
		fmt.Fprintf(&buf, " + %g %v", coefficients[i], attr.GetName())
	}
	return buf.String()
}

func (reg *LinearRegression) Fit(samples []*bitflow.Sample) error {
	if len(reg.Vars) < 2 {
		return fmt.Errorf("Need at least 2 variables for a linear regression, got %v", reg.Vars)
	}

	var err error
	reg.TrainData, err = reg.BuildInstances(reg.RegressionClassVar)
	if err != nil {
		return err
	}
	reg.FillInstances(samples, reg.TrainData)
	reg.Model = new(linear_models.LinearRegression)
	return reg.Model.Fit(reg.TrainData)
}

func (reg *LinearRegression) Predict(data base.FixedDataGrid) ([]float64, error) {
	prediction, err := reg.Model.Predict(data)
	if err != nil {
		return nil, err
	}
	classAttr, err := prediction.GetAttribute(prediction.AllClassAttributes()[0])
	if err != nil {
		return nil, err
	}
	_, num := data.Size()
	res := make([]float64, num)
	for i := range res {
		valBytes := prediction.Get(classAttr, i)
		res[i] = base.UnpackBytesToFloat(valBytes)
	}
	return res, nil
}

func (reg *LinearRegression) MeanSquaredError(data base.FixedDataGrid) (float64, error) {
	predictedValues, err := reg.Predict(data)
	if err != nil {
		return 0, err
	}
	referenceAttr, err := data.GetAttribute(reg.Model.Cls)
	if err != nil {
		return 0, err
	}
	var mse onlinestats.Running
	for i, predicted := range predictedValues {
		reference := base.UnpackBytesToFloat(data.Get(referenceAttr, i))
		diff := reference - predicted
		mse.Push(diff * diff)
	}
	return mse.Mean(), nil
}

func (reg *LinearRegression) IsValid() bool {
	non_zero := len(reg.Model.RegressionCoefficients) + 1
	if !pipeline.IsValidNumber(reg.Model.Disturbance) {
		return false
	}
	if reg.Model.Disturbance == 0 {
		non_zero--
	}
	for _, coefficients := range reg.Model.RegressionCoefficients {
		if !pipeline.IsValidNumber(coefficients) {
			return false
		}
		if coefficients == 0 {
			non_zero--
		}
	}
	if non_zero <= 0 {
		return false
	}
	return true
}

type SubHeader struct {
	Header *bitflow.Header
	Vars   []int
}

func (header SubHeader) BuildInstances(classAttribute int) (*base.DenseInstances, error) {
	data := base.NewDenseInstances()
	for i, fieldNum := range header.Vars {
		if fieldNum >= len(header.Header.Fields) || fieldNum < 0 {
			return nil, fmt.Errorf("Cannot access field nr. %v, header has %v fields", fieldNum, len(header.Header.Fields))
		}
		name := header.Header.Fields[fieldNum]
		attr := base.NewFloatAttribute(name)
		data.AddAttribute(attr)
		if i == classAttribute {
			data.AddClassAttribute(attr)
		}
	}
	return data, nil
}

func (header SubHeader) FillInstances(samples []*bitflow.Sample, instances *base.DenseInstances) {
	start, capacity := instances.Size()
	if capacity-start < len(samples) {
		instances.Extend(len(samples) - (capacity - start))
	}
	attributes := base.ResolveAllAttributes(instances)
	if len(attributes) != len(header.Vars) {
		panic("Number of attributes in instances does not match number of fields to fill in from samples")
	}

	for i, sample := range samples {
		for j, fieldNum := range header.Vars {
			val := sample.Values[fieldNum]
			valBytes := base.PackFloatToBytes(float64(val))
			instances.Set(attributes[j], start+i, valBytes)
		}
	}
}

func (header SubHeader) BuildFilledInstances(samples []*bitflow.Sample, classAttribute int) (*base.DenseInstances, error) {
	data, err := header.BuildInstances(classAttribute)
	if err != nil {
		return nil, err
	}
	header.FillInstances(samples, data)
	return data, nil
}
