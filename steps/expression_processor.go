package steps

import (
	"bytes"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

type ExpressionProcessor struct {
	bitflow.NoopProcessor
	Filter bool

	checker     bitflow.HeaderChecker
	expressions []*Expression
}

func RegisterExpression(b reg.ProcessorRegistry) {
	b.RegisterStep("do",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
			return add_expression(p, params, false)
		},
		"Execute the given expression on every sample").
		Required("expr", reg.String(), "Allows arithmetic operations on sample fields (e.g. \"field1\" [+,-,*,/] \"field2\").",
			"Following functions are implemented:",
			"tag(string) string: Access tag by tag key (string). Returns tag string or empty string if key does not exist.",
			"has_tag(string) bool: Check existence of tag ",
			"set_set(string, string) bitflow.SampleAndHeader:",
			"timestamp() float64: ",
			"now() float64:",
			"num() float64:",
			"str(_) string:",
			"strToFloat(string) float64:",
			"date_str(float64) string:",
			"set_timestamp(float64) bitflow.SampleAndHeader:",
			"floor(float64) float64:",
			"set(string, float64, bitflow.SampleAndHeader) bitflow.SampleAndHeader: ",
			"get_sample_and_header(): bitflow.SampleAndHeader: ")
}

func RegisterFilterExpression(b reg.ProcessorRegistry) {
	b.RegisterStep("filter",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
			return add_expression(p, params, true)
		},
		"Filter the samples based on a boolean expression").
		Required("expr", reg.String())
}

func add_expression(p *bitflow.SamplePipeline, params map[string]interface{}, filter bool) error {
	proc := &ExpressionProcessor{Filter: filter}
	err := proc.AddExpression(params["expr"].(string))
	if err == nil {
		p.Add(proc)
	}
	return err
}

func (p *ExpressionProcessor) AddExpression(expressionString string) error {
	expr, err := NewExpression(expressionString)
	if err != nil {
		return err
	}
	p.expressions = append(p.expressions, expr)
	return nil
}

func (p *ExpressionProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if outSample, outHeader, err := p.evaluate(sample, header); err != nil {
		return err
	} else if sample != nil || header != nil {
		return p.NoopProcessor.Sample(outSample, outHeader)
	}
	return nil
}

func (p *ExpressionProcessor) MergeProcessor(otherProcessor bitflow.SampleProcessor) bool {
	if other, ok := otherProcessor.(*ExpressionProcessor); !ok {
		return false
	} else {
		if other.Filter != p.Filter {
			return false
		}
		p.expressions = append(p.expressions, other.expressions...)
		return true
	}
}

func (p *ExpressionProcessor) String() string {
	var str bytes.Buffer
	for _, expr := range p.expressions {
		if str.Len() > 0 {
			if p.Filter {
				str.WriteString(" && ")
			} else {
				str.WriteString("; ")
			}
		}
		str.WriteString(expr.expr.String())
	}
	res := "Expression"
	if p.Filter {
		res += " filter"
	}
	return res + ": " + str.String()
}

func (p *ExpressionProcessor) evaluate(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
	if p.checker.HeaderChanged(header) {
		for _, expr := range p.expressions {
			if err := expr.UpdateHeader(header); err != nil {
				return nil, nil, err
			}
		}
	}
	outSample := sample
	outHeader := header
	var err error
	var res bool
	for _, expr := range p.expressions {
		if p.Filter {
			res, err = expr.EvaluateBool(outSample, outHeader)
			if err != nil {
				return nil, nil, err
			}
			if ! res {
				return nil, nil, nil
			}
		} else {
			outSample, outHeader, err = expr.Evaluate(outSample, outHeader)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return outSample, outHeader, err
}
