package regression

import (
	"bytes"
	"fmt"

	"github.com/antongulenko/go-onlinestats"
	"github.com/antongulenko/golearn/base"
	"github.com/antongulenko/golearn/linear_models"
	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/bitflow-script/reg"
)

type LinearRegression struct {
	SubHeader
	RegressionClassVar int
	Model              *linear_models.LinearRegression
	TrainData          *base.DenseInstances
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
