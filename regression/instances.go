package regression

import (
	"fmt"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golearn/base"
)

type SubHeader struct {
	sample.Header
	Vars []int
}

func (header SubHeader) BuildInstances(classAttribute int) (*base.DenseInstances, error) {
	data := base.NewDenseInstances()
	for i, fieldNum := range header.Vars {
		if fieldNum >= len(header.Fields) || fieldNum < 0 {
			return nil, fmt.Errorf("Cannot access field nr. %v, header has %v fields", fieldNum, len(header.Fields))
		}
		name := header.Fields[fieldNum]
		attr := base.NewFloatAttribute(name)
		data.AddAttribute(attr)
		if i == classAttribute {
			data.AddClassAttribute(attr)
		}
	}
	return data, nil
}

func (header SubHeader) FillInstances(samples []data2go.Sample, instances *base.DenseInstances) {
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
			val := data2go.Values[fieldNum]
			valBytes := base.PackFloatToBytes(float64(val))
			instances.Set(attributes[j], start+i, valBytes)
		}
	}
}

func (header SubHeader) BuildFilledInstances(samples []data2go.Sample, classAttribute int) (*base.DenseInstances, error) {
	data, err := header.BuildInstances(classAttribute)
	if err != nil {
		return nil, err
	}
	header.FillInstances(samples, data)
	return data, nil
}
