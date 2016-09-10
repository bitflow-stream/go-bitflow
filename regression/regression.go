package regression

import (
	"bytes"
	"fmt"
	"math"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/go-onlinestats"
	"github.com/antongulenko/golearn/base"
	"github.com/antongulenko/golearn/linear_models"
)

const (
	RegressionClassAttr = 0
)

type LinearRegression struct {
	SubHeader
	Model     linear_models.LinearRegression
	TrainData *base.DenseInstances
}

func NewLinearRegression(header *sample.Header, fieldNames []string) (LinearRegression, error) {
	fieldNums := make([]int, len(fieldNames))
	var reg LinearRegression
	for i, searching := range fieldNames {
		found := false
		for j, field := range header.Fields {
			if field == searching {
				found = true
				fieldNums[i] = j
			}
		}
		if !found {
			return reg, fmt.Errorf("Could not find header field '%v'", searching)
		}
	}
	reg.Header = *header
	reg.Vars = fieldNums
	return reg, nil
}

func (reg *LinearRegression) FormulaString() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%v = %g", reg.Model.Cls.GetName(), reg.Model.Disturbance)
	coeff := reg.Model.RegressionCoefficients
	for i, attr := range reg.Model.Attrs {
		fmt.Fprintf(&buf, " + %g %v", coeff[i], attr.GetName())
	}
	return buf.String()
}

func (reg *LinearRegression) Fit(samples []sample.Sample) error {
	if len(reg.Vars) < 2 {
		return fmt.Errorf("Need at least 2 variables for a linear regression, got %v", reg.Vars)
	}

	var err error
	reg.TrainData, err = reg.BuildInstances(RegressionClassAttr)
	if err != nil {
		return err
	}
	reg.FillInstances(samples, reg.TrainData)
	reg.Model = *(linear_models.NewLinearRegression())
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
	if !reg.IsValidNumber(reg.Model.Disturbance) {
		return false
	}
	if reg.Model.Disturbance == 0 {
		non_zero--
	}
	for _, coeff := range reg.Model.RegressionCoefficients {
		if !reg.IsValidNumber(coeff) {
			return false
		}
		if coeff == 0 {
			non_zero--
		}
	}
	if non_zero <= 0 {
		return false
	}
	return true
}

func (reg *LinearRegression) IsValidNumber(val float64) bool {
	return !math.IsNaN(val) && !math.IsInf(val, 0)
}
