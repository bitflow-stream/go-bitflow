package bitflow

import (
	"fmt"
	"io"
	"os"

	"github.com/antongulenko/golib/gotermBox"
)

const (
	// TextMarshallerDateFormat is the date format used by TextMarshaller to
	// print the timestamp of each sample.
	TextMarshallerDateFormat = "2006-01-02 15:04:05.999"

	// TextMarshallerDefaultSpacing is the default spacing between the columns
	// printed by TextMarshaller.
	TextMarshallerDefaultSpacing = 3

	// TextMarshallerHeaderChar is used as fill-character in the header line
	// preceding each sample marshalled by TextMarshaller.
	TextMarshallerHeaderChar = '='
)

// TextMarshaller marshals Headers and Samples to a human readable test format.
// It is mainly intended for easily readable output on the console. Headers are
// not printed separately. Every Sample is preceded by a header line containing
// the timestamp and tags. Afterwards, all values are printed in a aligned table
// in a key = value format. The width of the header line, the number of columns
// in the table, and the spacing between the columns in the table can be configured.
type TextMarshaller struct {
	// TextWidths sets the width of the header line and value table.
	// If Columns > 0, this value is ignored as the width is determined by the
	// number of columns. If this is 0, the width will be determined automatically:
	// If the output is a TTY (or if AssumeStdout is true), the width of the terminal
	// will be used. If it cannot be obtained, golib.GetTerminalSize() will return
	// a default value.
	TextWidth int

	// Columns can be set to > 0 to override TextWidth and set a fixed number of
	// columns in the table. Otherwise it will be computed automatically based
	// on TextWidth.
	Columns int

	// Set additional spacing between the columns of the output table. If <= 0, the
	// default value TextMarshallerDefaultSpacing will be used.
	Spacing int

	// If true, assume the output is a TTY and try to obtain the TextWidth from
	// the operating system.
	AssumeStdout bool
}

// ShouldCloseAfterFirstSample defines that text streams can stream without closing
func (TextMarshaller) ShouldCloseAfterFirstSample() bool {
	return false
}

// String implements the Marshaller interface.
func (TextMarshaller) String() string {
	return "text"
}

// WriteHeader implements the Marshaller interface. It is empty, because
// TextMarshaller prints a separate header for each Sample.
func (TextMarshaller) WriteHeader(header *Header, withTags bool, output io.Writer) error {
	return nil
}

// WriteSample implements the Marshaller interface. See the TextMarshaller godoc
// for information about the format.
func (m TextMarshaller) WriteSample(sample *Sample, header *Header, withTags bool, writer io.Writer) error {
	headerStr := sample.Time.Format(TextMarshallerDateFormat)
	if withTags {
		headerStr = fmt.Sprintf("%s (%s)", headerStr, sample.TagString())
	}
	lines := make([]string, 0, len(sample.Values))
	for i, value := range sample.Values {
		line := fmt.Sprintf("%s = %.4f", header.Fields[i], value)
		lines = append(lines, line)
	}

	textWidth, columnWidths := m.calculateWidths(lines, writer)
	if err := m.writeHeader(headerStr, textWidth, writer); err != nil {
		return err
	}
	return m.writeLines(lines, columnWidths, writer)
}

func (m TextMarshaller) calculateWidths(lines []string, writer io.Writer) (textWidth int, columnWidths []int) {
	spacing := m.Spacing
	if spacing <= 0 {
		spacing = TextMarshallerDefaultSpacing
	}
	if m.Columns > 0 {
		columnWidths = m.fixedColumnWidths(lines, m.Columns, spacing)
		for _, width := range columnWidths {
			textWidth += width
		}
	} else {
		if m.TextWidth > 0 {
			textWidth = m.TextWidth
		} else {
			textWidth = m.defaultTextWidth(writer)
		}
		columnWidths = m.variableColumnWidths(lines, textWidth, spacing)
	}
	return
}

func (m TextMarshaller) fixedColumnWidths(lines []string, columns int, spacing int) (widths []int) {
	widths = make([]int, columns)
	for i, line := range lines {
		col := i % columns
		length := len(line) + spacing
		if widths[col] < length {
			widths[col] = length
		}
	}
	return
}

func (m TextMarshaller) defaultTextWidth(writer io.Writer) int {
	if m.AssumeStdout || writer == os.Stdout {
		size := gotermBox.GetTerminalSize()
		return int(size.Col)
	} else {
		return int(gotermBox.DefaultTerminalWindowSize.Col)
	}
}

func (m TextMarshaller) variableColumnWidths(strings []string, textWidth int, spacing int) []int {
	columns := make([]int, len(strings))
	strLengths := make([]int, len(strings))
	for i, line := range strings {
		length := len(line) + spacing
		columns[i] = length
		strLengths[i] = length
	}
	for len(columns) > 1 {
		columns = columns[:len(columns)-1]
		for i, strLen := range strLengths {
			col := i % len(columns)
			if columns[col] < strLen {
				columns[col] = strLen
			}
		}
		lineLen := 0
		for _, strLen := range columns {
			lineLen += strLen
		}
		if lineLen <= textWidth {
			break
		}
	}
	return columns
}

func (m TextMarshaller) writeHeader(header string, textWidth int, writer io.Writer) (err error) {
	extraSpace := textWidth - len(header)
	if extraSpace >= 4 {
		lineChars := (extraSpace - 2) / 2
		line := make([]byte, lineChars)
		for i := 0; i < lineChars; i++ {
			line[i] = TextMarshallerHeaderChar
		}
		lineStr := string(line)
		_, err = fmt.Fprintln(writer, lineStr, header, lineStr)
	} else {
		_, err = fmt.Fprintln(writer, header)
	}
	return
}

func (m TextMarshaller) writeLines(lines []string, widths []int, writer io.Writer) error {
	for i, line := range lines {
		if err := m.writeLine(line, i, lines, widths, writer); err != nil {
			return err
		}
	}
	return nil
}

func (m TextMarshaller) writeLine(line string, lineIndex int, lines []string, widths []int, writer io.Writer) error {
	if _, err := writer.Write([]byte(line)); err != nil {
		return err
	}
	col := lineIndex % len(widths)
	if col >= len(widths)-1 || lineIndex == len(lines)-1 {
		if _, err := writer.Write([]byte("\n")); err != nil {
			return err
		}
	} else {
		extraSpace := widths[col] - len(line)
		for j := 0; j < extraSpace; j++ {
			if _, err := writer.Write([]byte(" ")); err != nil {
				return err
			}
		}
	}
	return nil
}
