package bitflow

import (
	"fmt"
	"testing"

	"github.com/antongulenko/golib"
	"github.com/stretchr/testify/suite"
)

type PipelinePrinterTestSuite struct {
	golib.AbstractTestSuite
}

func TestPipelinePrinter(t *testing.T) {
	suite.Run(t, new(PipelinePrinterTestSuite))
}

var printer = IndentPrinter{
	OuterIndent:  "| ",
	InnerIndent:  "|-",
	FillerIndent: "  ",
	CornerIndent: "\\-",
}

type contained struct {
	name     string
	children []fmt.Stringer
}

func (c contained) ContainedStringers() []fmt.Stringer {
	return c.children
}

func (c contained) String() string {
	return c.name
}

func (s *PipelinePrinterTestSuite) TestIndentPrinter1() {
	obj := contained{"a", []fmt.Stringer{String("b"), nil, String("c")}}
	s.Equal(`a
|-b
|-(nil)
\-c`,
		printer.Print(obj))
}

func (s *PipelinePrinterTestSuite) TestIndentPrinterBig() {
	obj := contained{"a",
		[]fmt.Stringer{
			contained{"b",
				[]fmt.Stringer{String("c")}},
			contained{"d",
				[]fmt.Stringer{
					contained{"e",
						[]fmt.Stringer{String("f"), String("g")},
					},
					contained{"h",
						[]fmt.Stringer{String("i"), String("j")},
					},
				}},
			contained{"k",
				[]fmt.Stringer{String("h"), String("i")}},
		}}

	s.Equal(`a
|-b
| \-c
|-d
| |-e
| | |-f
| | \-g
| \-h
|   |-i
|   \-j
\-k
  |-h
  \-i`,
		printer.Print(obj))
}
