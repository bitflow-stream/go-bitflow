package pipeline

import (
	"bytes"
	"fmt"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/antongulenko/go-bitflow"
)

type Expression struct {
	expr       *govaluate.EvaluableExpression
	vars       map[string]bool
	varIndices map[int]string
	sample     *bitflow.Sample
	header     *bitflow.Header
	num        int
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
	}
	return expr, nil
}

func (p *Expression) UpdateHeader(header *bitflow.Header) error {
	resolvedVariables := make(map[string]bool)
	p.varIndices = make(map[int]string)

	for i, field := range header.Fields {
		if p.vars[field] {
			p.varIndices[i] = field
			resolvedVariables[field] = true
		}
	}

	for variable := range p.vars {
		if !resolvedVariables[variable] {
			return fmt.Errorf("%v: Variable %v cannot be resolved in header", p.expr, variable)
		}
	}
	return nil
}

func (p *Expression) Evaluate(sample *bitflow.Sample, header *bitflow.Header) (interface{}, error) {
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

func (p *Expression) EvaluateBool(sample *bitflow.Sample, header *bitflow.Header) (bool, error) {
	result, err := p.Evaluate(sample, header)
	if err != nil {
		return false, err
	}
	if boolResult, ok := result.(bool); ok {
		return boolResult, nil
	} else {
		return false, fmt.Errorf("%v: Non-boolean result returned: %v (%T)", p.expr, result, result)
	}
}

func (p *Expression) makeParameters(sample *bitflow.Sample) map[string]interface{} {
	parameters := make(map[string]interface{})
	for index, variable := range p.varIndices {
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
			p.currentHeader().HasTags = true
			return args[1], nil
		}),
		"timestamp": p.makeStringFunction("timestamp", 0, func(sample *bitflow.Sample, args ...string) (interface{}, error) {
			return float64(sample.Time.Unix()), nil
		}),
		// Dates are parsed automatically by the govaluate library if a date/time formatted string is encountered. The Unix() value is used.
		// If alternative date formats are required, this function can be added.
		//"date": p.makeStringFunction("date", 1, func(sample *bitflow.Sample, args ...string) (interface{}, error) {
		//	date, err := time.Parse(bitflow.TextMarshallerDateFormat, args[0])
		//	return float64(date.Unix()), fmt.Errorf("Cannot parse date (format: %v): %v", bitflow.TextMarshallerDateFormat, err)
		//}),
		"now": p.makeStringFunction("now", 0, func(sample *bitflow.Sample, args ...string) (interface{}, error) {
			return float64(time.Now().Unix()), nil
		}),
		"num": p.makeStringFunction("num", 0, func(sample *bitflow.Sample, args ...string) (interface{}, error) {
			return p.num, nil
		}),
		"str": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) == 1 {
				return fmt.Sprintf("%v", arguments[0]), nil
			}
			return nil, fmt.Errorf("str() needs 1 parameter, but received: %v", printParamStrings(arguments))
		},
		"date_str": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) == 1 {
				if numArg, ok := arguments[0].(float64); ok {
					return time.Unix(int64(numArg), 0).Format(bitflow.TextMarshallerDateFormat), nil
				}
			}
			return nil, fmt.Errorf("date_str() needs 1 float64 parameter, but received: %v", printParamStrings(arguments))
		},
		"set_timestamp": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) == 1 {
				if numArg, ok := arguments[0].(float64); ok {
					p.currentSample().Time = time.Unix(int64(numArg), 0)
					return arguments[0], nil
				}
			}
			return nil, fmt.Errorf("set_timestamp() needs 1 float64 parameter, but received: %v", printParamStrings(arguments))
		},
	}
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

func printParamStrings(vals []interface{}) string {
	var buf bytes.Buffer
	for _, val := range vals {
		if buf.Len() > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(&buf, "%v (%T)", val, val)
	}
	return buf.String()
}
