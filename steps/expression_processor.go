package steps

import (
	"bytes"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

type ExpressionProcessor struct {
	bitflow.NoopProcessor
	Filter bool

	checker     bitflow.HeaderChecker
	expressions []*Expression
}

func RegisterExpression(b *query.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("do",
		func(p *pipeline.SamplePipeline, params map[string]string) error {
			return add_expression(p, params, false)
		},
		"Execute the given expression on every sample", []string{"expr"})
}

func RegisterFilterExpression(b *query.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("filter",
		func(p *pipeline.SamplePipeline, params map[string]string) error {
			return add_expression(p, params, true)
		},
		"Filter the samples based on a boolean expression", []string{"expr"})
}

func add_expression(p *pipeline.SamplePipeline, params map[string]string, filter bool) error {
	proc := &ExpressionProcessor{Filter: filter}
	err := proc.AddExpression(params["expr"])
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
	} else if res {
		return p.NoopProcessor.Sample(sample, header)
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

func (p *ExpressionProcessor) evaluate(sample *bitflow.Sample, header *bitflow.Header) (bool, error) {
	if p.checker.HeaderChanged(header) {
		for _, expr := range p.expressions {
			if err := expr.UpdateHeader(header); err != nil {
				return false, err
			}
		}
	}
	for _, expr := range p.expressions {
		if p.Filter {
			if res, err := expr.EvaluateBool(sample, header); err != nil {
				return false, err
			} else if !res {
				return false, nil
			}
		} else {
			if _, err := expr.Evaluate(sample, header); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}
