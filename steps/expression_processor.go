package steps

import (
	"bytes"

	"github.com/antongulenko/go-bitflow"
)

type ExpressionProcessor struct {
	bitflow.NoopProcessor
	Filter bool

	checker     bitflow.HeaderChecker
	expressions []*Expression
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
