package pipeline

import (
	"bytes"
	"fmt"

	"github.com/Knetic/govaluate"
	"github.com/antongulenko/go-bitflow"
)

type SampleExpressionFilter struct {
	bitflow.AbstractProcessor
	checker     bitflow.HeaderChecker
	expressions []*expression
}

func (p *SampleExpressionFilter) Expr(expressionString string) error {
	newExpression := &expression{
		filter: p,
		vars:   make(map[string]bool),
	}
	expr, err := govaluate.NewEvaluableExpressionWithFunctions(expressionString, newExpression.makeFunctions())
	if err == nil {
		newExpression.expr = expr
		for _, variable := range expr.Vars() {
			newExpression.vars[variable] = true
		}
		p.expressions = append(p.expressions, newExpression)
	}
	return err
}

func (p *SampleExpressionFilter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	if res, err := p.evaluate(sample, header); err != nil {
		return err
	} else if res {
		return p.OutgoingSink.Sample(sample, header)
	}
	return nil
}

func (p *SampleExpressionFilter) MergeProcessor(otherProcessor bitflow.SampleProcessor) bool {
	if other, ok := otherProcessor.(*SampleExpressionFilter); !ok {
		return false
	} else {
		p.expressions = append(p.expressions, other.expressions...)
		return true
	}
}

func (p *SampleExpressionFilter) String() string {
	var str bytes.Buffer
	for _, expr := range p.expressions {
		if str.Len() > 0 {
			str.WriteString(" && ")
		}
		str.WriteString(expr.expr.String())
	}
	return "Expression filter: " + str.String()
}

func (p *SampleExpressionFilter) evaluate(sample *bitflow.Sample, header *bitflow.Header) (bool, error) {
	if p.checker.HeaderChanged(header) {
		for _, expr := range p.expressions {
			expr.updateHeader(header)
		}
	}
	for _, expr := range p.expressions {
		if res, err := expr.evaluate(sample); err != nil {
			return false, err
		} else if !res {
			return false, nil
		}
	}
	return true, nil
}

type expression struct {
	filter     *SampleExpressionFilter
	expr       *govaluate.EvaluableExpression
	vars       map[string]bool
	varIndices map[int]string
	sample     *bitflow.Sample
}

func (p *expression) updateHeader(header *bitflow.Header) error {
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
			return fmt.Errorf("%v: Variable %v cannot be resolved in header. Use tag() and has_tag() to access tags.", p.expr, variable)
		}
	}
	return nil
}

func (p *expression) evaluate(sample *bitflow.Sample) (bool, error) {
	parameters := p.makeParameters(sample)

	// Set the sample so that the functions in makeFunctions() can access its values.
	p.sample = sample
	defer func() {
		p.sample = nil
	}()
	result, err := p.expr.Evaluate(parameters)

	if err != nil {
		return false, err
	}
	if boolResult, ok := result.(bool); ok {
		return boolResult, nil
	} else {
		return false, fmt.Errorf("%v: Non-boolean result returned: %v (%T)", p.expr, result, result)
	}
}

func (p *expression) makeParameters(sample *bitflow.Sample) map[string]interface{} {
	parameters := make(map[string]interface{})
	for index, variable := range p.varIndices {
		parameters[variable] = float64(sample.Values[index])
	}
	return parameters
}

func (p *expression) makeFunctions() map[string]govaluate.ExpressionFunction {
	return map[string]govaluate.ExpressionFunction{
		"tag": p.makeStringFunction("tag", func(tag string) interface{} {
			return p.currentSample().Tag(tag)
		}),
		"has_tag": p.makeStringFunction("has_tag", func(tag string) interface{} {
			return p.currentSample().HasTag(tag)
		}),
	}
}

func (p *expression) makeStringFunction(funcName string, f func(arg string) interface{}) govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) == 1 {
			if strArg, ok := args[0].(string); ok {
				return f(strArg), nil
			}
		}
		return "", fmt.Errorf("%v() needs 1 string parameter, but received: %v", funcName, args)
	}
}

func (p *expression) currentSample() *bitflow.Sample {
	if p.sample == nil {
		panic("An expression function was called outside of evaluate()")
	}
	return p.sample
}
