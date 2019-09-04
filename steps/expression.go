package steps

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/bitflow-stream/go-bitflow/bitflow"
)

type Expression struct {
	expr          *govaluate.EvaluableExpression
	orderedVars   []string // Required to make the order of parameters traceable
	vars          map[string]bool
	indicesToVars map[int]string
	varsToIndices map[string]int
	sample        *bitflow.Sample
	header        *bitflow.Header
	num           int
}

func NewExpression(expressionString string) (*Expression, error) {
	expr := &Expression{
		vars: make(map[string]bool),
	}
	compiled, err := govaluate.NewEvaluableExpressionWithFunctions(expressionString, expr.makeFunctions())
	if err != nil {
		return nil, err
	}
	expr.expr = compiled
	for _, variable := range compiled.Vars() {
		expr.vars[variable] = true
		expr.orderedVars = append(expr.orderedVars, variable)
	}
	return expr, nil
}

func (p *Expression) UpdateHeader(header *bitflow.Header) error {
	resolvedVariables := make(map[string]bool)
	p.indicesToVars = make(map[int]string)

	for i, field := range header.Fields {
		if p.vars[field] {
			p.indicesToVars[i] = field
			resolvedVariables[field] = true
		}
	}

	for variable := range p.vars {
		if !resolvedVariables[variable] {
			return fmt.Errorf("%v: Variable %v cannot be resolved in header", p.expr, variable)
		}
	}
	p.varsToIndices = header.BuildIndex()

	return nil
}

func (p *Expression) Evaluate(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
	res, err := p.doEvaluate(sample, header)
	if outSampleAndHeader, ok := res.(bitflow.SampleAndHeader); ok {
		return outSampleAndHeader.Sample, outSampleAndHeader.Header, err
	}
	return sample, header, err
}

func (p *Expression) EvaluateBool(sample *bitflow.Sample, header *bitflow.Header) (bool, error) {
	result, err := p.doEvaluate(sample, header)
	if err != nil {
		return false, err
	}
	if boolResult, ok := result.(bool); ok {
		return boolResult, nil
	} else {
		return false, fmt.Errorf("%v: Non-boolean result returned: %v (%T)", p.expr, result, result)
	}
}

func (p *Expression) doEvaluate(sample *bitflow.Sample, header *bitflow.Header) (interface{}, error) {
	parameters := p.makeParameters(sample)
	p.sample = sample // Set the sample/header so that the functions in makeFunctions() can access its values
	p.header = header
	defer func() {
		p.sample = nil
		p.header = nil
	}()
	res, err := p.expr.Evaluate(parameters)
	p.num++
	return res, err
}

func (p *Expression) makeParameters(sample *bitflow.Sample) map[string]interface{} {
	parameters := make(map[string]interface{})
	for index, variable := range p.indicesToVars {
		parameters[variable] = float64(sample.Values[index])
	}
	return parameters
}

func (p *Expression) makeFunctions() map[string]govaluate.ExpressionFunction {
	return map[string]govaluate.ExpressionFunction{
		"tag": p.makeStringFunction("tag", 1, func(sample *bitflow.Sample, args ...string) (interface{}, error) {
			return sample.Tag(args[0]), nil
		}),
		"has_tag": p.makeStringFunction("has_tag", 1, func(sample *bitflow.Sample, args ...string) (interface{}, error) {
			return sample.HasTag(args[0]), nil
		}),
		"set_tag": p.makeStringFunction("set_tag", 2, func(sample *bitflow.Sample, args ...string) (interface{}, error) {
			sample.SetTag(args[0], args[1])
			return bitflow.SampleAndHeader{
				Sample: sample,
				Header: p.currentHeader(),
			}, nil
		}),
		"timestamp": p.makeStringFunction("timestamp", 0, func(sample *bitflow.Sample, _ ...string) (interface{}, error) {
			return float64(sample.Time.Unix()), nil
		}),
		// Dates are parsed automatically by the govaluate library if a date/time formatted string is encountered. The Unix() value is used.
		// If alternative date formats are required, this function can be added.
		// "date": p.makeStringFunction("date", 1, func(sample *bitflow.Sample, args ...string) (interface{}, error) {
		//  	date, err := time.Parse(bitflow.TextMarshallerDateFormat, args[0])
		//  	return float64(date.Unix()), fmt.Errorf("Cannot parse date (format: %v): %v", bitflow.TextMarshallerDateFormat, err)
		// }),
		"now": p.makeStringFunction("now", 0, func(_ *bitflow.Sample, _ ...string) (interface{}, error) {
			return float64(time.Now().Unix()), nil
		}),
		"num": p.makeStringFunction("num", 0, func(_ *bitflow.Sample, _ ...string) (interface{}, error) {
			return float64(p.num), nil
		}),
		"str": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) == 1 {
				return fmt.Sprintf("%v", arguments[0]), nil
			}
			return nil, fmt.Errorf("str() needs 1 parameter, but received: %v", printParamStrings(arguments))
		},
		"strToFloat": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) == 1 {
				if strValue, ok := arguments[0].(string); ok {
					if value, err := strconv.ParseFloat(strValue, 64); err == nil {
						return value, nil
					} else {
						return nil, fmt.Errorf("error parsing float64 from string %v", strValue)
					}
				}
			}
			return nil, fmt.Errorf("strToFloat() needs 1 string parameter, but received: %v", printParamStrings(arguments))
		},
		"date_str": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) == 1 {
				if numArg, err := p.argsToFloat64(arguments); err == nil {
					return time.Unix(int64(numArg[0]), 0).Format(bitflow.TextMarshallerDateFormat), nil
				}
			}
			return nil, fmt.Errorf("date_str() needs 1 float64 parameter, but received: %v", printParamStrings(arguments))
		},
		"set_timestamp": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) == 1 {
				if numArg, err := p.argsToFloat64(arguments); err == nil {
					outSampleAndHeader := bitflow.SampleAndHeader{
						Sample: p.currentSample(),
						Header: p.currentHeader(),
					}
					p.currentSample().Time = time.Unix(int64(numArg[0]), 0)
					return outSampleAndHeader, nil
				}
			}
			return nil, fmt.Errorf("set_timestamp() needs 1 float64 parameter, but received: %v", printParamStrings(arguments))
		},
		"floor": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) == 1 {
				if numArg, err := p.argsToFloat64(arguments); err == nil {
					return math.Floor(numArg[0]), nil
				}
			}
			return nil, fmt.Errorf("floor() needs 1 float64 parameter, but received: %v", printParamStrings(arguments))
		},
		"set": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) != 2 && len(arguments) != 3 {
				return nil, fmt.Errorf("following parametrization required: set(required(string), required(float64),"+
					" optional(bitflow.SampleAndHeader)), but received: %v", printParamStrings(arguments))
			}
			var ok bool
			outSampleAndHeader := bitflow.SampleAndHeader{
				Sample: p.currentSample(),
				Header: p.currentHeader(),
			}
			if len(arguments) == 3 {
				outSampleAndHeader, ok = arguments[2].(bitflow.SampleAndHeader)
				if !ok {
					return nil, fmt.Errorf("Error while parsing third sample and header argument: %v", arguments[2])
				}
			}
			value, ok := arguments[1].(float64)
			if !ok {
				return nil, fmt.Errorf("Error while parsing second float64 argument: %v", arguments[1])
			}
			field, ok := arguments[0].(string)
			if !ok {
				return nil, fmt.Errorf("Error while parsing first string argument: %v", arguments[0])
			}
			if ! outSampleAndHeader.Header.ContainsField(field) {
				outSampleAndHeader.Sample = p.currentSample().DeepClone()
				outSampleAndHeader.Header = p.currentHeader().Clone(p.currentHeader().Fields)

				outSampleAndHeader.Sample.Values = append(outSampleAndHeader.Sample.Values, bitflow.Value(value))
				outSampleAndHeader.Header.Fields = append(outSampleAndHeader.Header.Fields, field)
			} else {
				outSampleAndHeader.Sample.Values[p.varsToIndices[field]] = bitflow.Value(value)
			}

			return outSampleAndHeader, nil
		},
		"get_sample_and_header": func(arguments ...interface{}) (interface{}, error) {
			outSampleAndHeader := bitflow.SampleAndHeader{
				Sample: p.currentSample(),
				Header: p.currentHeader(),
			}
			return outSampleAndHeader, nil
		},
	}
}

func (p *Expression) argsToFloat64(arguments interface{}) ([]float64, error) {
	args, ok := arguments.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid argument format. array expected but received: %v", args)
	}
	floats := make([]float64, len(args))
	for i, arg := range args {
		floatArg, ok := arg.(float64)
		if ! ok {
			return nil, fmt.Errorf("failed to parse float64 parameter: %v", arg)
		}
		floats[i] = floatArg
	}
	return floats, nil
}

func (p *Expression) makeStringFunction(funcName string, numArgs int, f func(sample *bitflow.Sample, args ...string) (interface{}, error)) govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		sample := p.currentSample()
		if len(args) == numArgs {
			strArgs := make([]string, 0, numArgs)
			for _, arg := range args {
				if strArg, ok := arg.(string); ok {
					strArgs = append(strArgs, strArg)
				} else {
					break
				}
			}
			if len(strArgs) == numArgs {
				return f(sample, strArgs...)
			}
		}
		return nil, fmt.Errorf("%v() needs %v string parameter(s), but received: %v", funcName, numArgs, printParamStrings(args))
	}
}

func (p *Expression) currentSample() *bitflow.Sample {
	sample := p.sample
	if sample == nil {
		panic("An expression function was called outside of Expression.Evaluate()")
	}
	return sample
}

func (p *Expression) currentHeader() *bitflow.Header {
	header := p.header
	if header == nil {
		panic("An expression function was called outside of Expression.Evaluate()")
	}
	return header
}

func printParamStrings(values []interface{}) string {
	var buf bytes.Buffer
	for _, val := range values {
		if buf.Len() > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(&buf, "%v (%T)", val, val)
	}
	return buf.String()
}
