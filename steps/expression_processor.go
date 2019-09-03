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
		Required("expr", reg.String())
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
	if res, err := p.evaluate(sample, header); err != nil {
		return err
	} else if res != nil {
		if sampleAndHeader, ok := res.(*bitflow.SampleAndHeader); ok {
			return p.NoopProcessor.Sample(sampleAndHeader.Sample, sampleAndHeader.Header)
		} else if keepSample, ok := res.(bool); ok && keepSample {
			return p.NoopProcessor.Sample(sample, header)
		} else {
			return p.NoopProcessor.Sample(sample, header)
		}
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

func (p *ExpressionProcessor) evaluate(sample *bitflow.Sample, header *bitflow.Header) (interface{}, error) {
	if p.checker.HeaderChanged(header) {
		for res, expr := range p.expressions {
			if err := expr.UpdateHeader(header); err != nil {
				return res, err
			}
		}
	}
	var result interface{}
	for _, expr := range p.expressions {
		if p.Filter {
			if res, err := expr.EvaluateBool(sample, header); err != nil {
				return res, err
			} else if !res {
				return res, nil
			}
		} else {
			if res, err := expr.Evaluate(sample, header); err != nil {
				return res, err
			} else {
				result = res
			}
		}
	}
	return result, nil
}
