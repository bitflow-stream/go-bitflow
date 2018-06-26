package pipeline

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var printer = IndentPrinter{
	OuterIndent:  "| ",
	InnerIndent:  "|-",
	FillerIndent: "  ",
	CornerIndent: "\\-",
}

func TestIndentPrinter1(t *testing.T) {
	obj := contained{"a", []fmt.Stringer{String("b"), nil, String("c")}}
	assert.Equal(t,
		`a
|-b
|-(nil)
\-c`,
		printer.Print(obj))
}

func TestIndentPrinterBig(t *testing.T) {
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

	assert.Equal(t,
		`a
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

type contained struct {
	name     string
	children []fmt.Stringer
}

func (c contained) ContainedStringers() []fmt.Stringer {
	return []fmt.Stringer(c.children)
}

func (c contained) String() string {
	return c.name
}
