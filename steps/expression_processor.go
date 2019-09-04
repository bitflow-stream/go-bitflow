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
		Required("expr", reg.String(), "Allows arithmetic and boolean operations. Can also perform them on sample fields "+
			"(e.g. \"field1\" [+,-,*,/] \"field2\").",
			"Following additional functions are implemented: ",
			"tag(string) string: Access tag by tag key (string). Returns tag string or empty string if key does not exist.",
			"",
			"has_tag(string) bool: Check existence of tag key.",
			"",
			"set_tag(string, string) bitflow.SampleAndHeader: Adds or replaces a tag at the current sample and returns the result.",
			"First argument is key, second is the tag value.",
			"",
			"timestamp() float64: Returns the timestamp of the current sample.",
			"",
			"now() float64: Returns the current local system time.",
			"",
			"num() float64: Returns the current number of processed samples.",
			"",
			"str(_) string: Converts an arbitrary argument to string.",
			"",
			"strToFloat(string) float64: Converts a string argument to float64.",
			"",
			"date_str(float64) string: Converts the timestamp argument to a string representation.",
			"",
			"set_timestamp(float64) bitflow.SampleAndHeader: Sets the timestamp of the current sample and returns the result.",
			"",
			"floor(float64) float64: Applies the floor operation on the argument.",
			"",
			"set(string, float64, optional: bitflow.SampleAndHeader) bitflow.SampleAndHeader:",
			" Sets or replaces the value (2nd argument) of the sample field (1st argument).",
			"Third argument is optional. If set, the setting is applied on the passed sample.",
			"Otherwise it is applied on the current sample.",
			"",
			"get_sample_and_header(): bitflow.SampleAndHeader: Returns the current sample.",
			"",
			"Note that arithmetic, boolean and function expressions can be combines as long as the arguments and return types match.",
			"Some examples: ",
			"'expr'='set_tag(\"my_system_time\", str(now()))'",
			"'expr'='set_timestamp(now())'",
			"'expr'='set(\"field3\", \"field1\" + \"field2\")'",
			"'expr'='set(\"field1\", \"field1\" * 10)'",
			"'expr'='set(\"field3\", \"field1\" * 10, set(\"field2\", now(), set_tag(\"new_tag\", \"awesome\")))'",
			"",
			"Currently the field to value mapping is done once before each sample is processed.",
			"Therefore, interdependent arithmetic operations produce possibly unexpected results.",
			"Example: 'expr'='set(\"field1\", \"field1\" + 10, set(\"field1\", 10))'",
			"The expected value for \"field1\" would be 20.",
			"However, the actual result would be the original value of \"field1\" + 10 or an error if \"field1\" does not exist in the sample.")
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
