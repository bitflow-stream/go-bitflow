package sample

import (
	"bytes"
	"fmt"
	"io"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/golib"
)

const (
	text_date_format = "2006-01-02 15:04:05.999"

	// Defaults will be applied if value is <= 0
	text_marshaller_default_spacing = 3
	text_marshaller_default_width   = 200 // Automatic for stdin.
	text_marshaller_header_char     = '='
)

type TextMarshaller struct {
	TextWidth    int // Ignored if Columns is > 0
	Columns      int // Will be inferred from TextWidth if <= 0
	Spacing      int
	AssumeStdout bool
}

func (*TextMarshaller) String() string {
	return "text"
}

func (m *TextMarshaller) WriteHeader(header Header, writer io.Writer) error {
	return nil
}

func (m *TextMarshaller) WriteSample(sample Sample, header Header, writer io.Writer) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	headerStr := sample.Time.Format(text_date_format)
	if header.HasTags {
		var b bytes.Buffer
		b.WriteString(headerStr)
		b.WriteString(" (")
		b.WriteString(sample.TagString())
		b.WriteString(")")
		headerStr = b.String()
	}
	lines := make([]string, 0, len(sample.Values))
	for i, value := range sample.Values {
		line := fmt.Sprintf("%s = %.4f", header.Fields[i], value)
		lines = append(lines, line)
	}

	textWidth, columnWidths := m.calculateWidths(lines, writer)
	m.writeHeader(headerStr, textWidth, writer)
	m.writeLines(lines, columnWidths, writer)
	return nil
}

func (m *TextMarshaller) calculateWidths(lines []string, writer io.Writer) (textWidth int, columnWidths []int) {
	spacing := m.Spacing
	if spacing <= 0 {
		spacing = text_marshaller_default_spacing
	}
	if m.Columns > 0 {
		columnWidths = m.columnWidths(lines, m.Columns, spacing)
		for _, width := range columnWidths {
			textWidth += width
		}
	} else {
		if m.TextWidth > 0 {
			textWidth = m.TextWidth
		} else {
			textWidth = m.defaultTextWidth(writer)
		}
		columns := m.numberOfColumns(lines, textWidth, spacing)
		columnWidths = m.columnWidths(lines, columns, spacing)
	}
	return
}

func (m *TextMarshaller) defaultTextWidth(writer io.Writer) int {
	if m.AssumeStdout || writer == os.Stdout {
		if size, err := golib.GetTerminalSize(); err != nil {
			log.Warnln("Failed to get terminal size:", err)
			return 0
		} else if size.Col == 0 {
			log.Warnln("Terminal size returned as 0, using default:", text_marshaller_default_width)
			return text_marshaller_default_width
		} else {
			return int(size.Col)
		}
	} else {
		return text_marshaller_default_width
	}
}

func (m *TextMarshaller) numberOfColumns(lines []string, textWidth int, spacing int) int {
	columns := len(lines)
	columnCounter := 0
	width := 0
	for _, line := range lines {
		length := len(line) + spacing
		if width+length > textWidth {
			if columns > columnCounter {
				columns = columnCounter
				if columns <= 1 {
					break
				}
			}
			width = length
			columnCounter = 1
		} else {
			width += length
			columnCounter++
		}
	}
	return columns
}

func (m *TextMarshaller) columnWidths(lines []string, columns int, spacing int) (widths []int) {
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

func (m *TextMarshaller) writeHeader(header string, textWidth int, writer io.Writer) {
	extraSpace := textWidth - len(header)
	if extraSpace >= 4 {
		lineChars := (extraSpace - 2) / 2
		line := make([]byte, lineChars)
		for i := 0; i < lineChars; i++ {
			line[i] = text_marshaller_header_char
		}
		lineStr := string(line)
		fmt.Fprintln(writer, lineStr, header, lineStr)
	} else {
		fmt.Fprintln(writer, header)
	}
}

func (m *TextMarshaller) writeLines(lines []string, widths []int, writer io.Writer) {
	columns := len(widths)
	for i, line := range lines {
		writer.Write([]byte(line))
		col := i % columns
		if col >= columns-1 || i == len(lines)-1 {
			writer.Write([]byte("\n"))
		} else {
			extraSpace := widths[col] - len(line)
			for j := 0; j < extraSpace; j++ {
				writer.Write([]byte(" "))
			}
		}
	}
}
