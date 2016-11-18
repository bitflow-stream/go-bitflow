package gotermBox

import (
	"io"

	"github.com/antongulenko/golib"
	"github.com/antongulenko/goterm"
)

type CliLogBox struct {
	*golib.LogBuffer
	NoUtf8        bool
	LogLines      int
	MessageBuffer int
}

func (box *CliLogBox) Init() {
	box.LogBuffer = golib.NewLogBuffer(box.MessageBuffer)
}

func (self *CliLogBox) Update(writeContent func(out io.Writer, width int)) {
	box := goterm.NewBox(100|goterm.PCT, 100|goterm.PCT, 0)
	box.Height -= 1 // Line with cursor
	var separator, dots string
	if self.NoUtf8 {
		box.Border = "- | - - - -"
		separator = "-"
		dots = "... "
	} else {
		box.Border = "═ ║ ╔ ╗ ╚ ╝"
		separator = "═"
		dots = "··· "
	}
	lines := box.Height - 3 // borders + separator

	counter := newlineCounter{out: box}
	if self.LogLines > 0 {
		counter.max_lines = lines - self.LogLines
	}
	writeContent(&counter, box.Width)
	lines -= counter.num

	if counter.num > 0 {
		i := 0
		if counter.truncated {
			box.Write([]byte(dots))
			i += len(dots)
		}
		for i := 0; i < box.Width; i++ {
			box.Write([]byte(separator))
		}
		box.Write([]byte("\n"))
	}
	self.PrintMessages(box, lines)
	goterm.MoveCursor(1, 1)
	goterm.Print(box)
	goterm.Flush()
}

type newlineCounter struct {
	out       io.Writer
	num       int
	max_lines int
	truncated bool
}

func (counter *newlineCounter) Write(data []byte) (total int, err error) {
	total = len(data)
	if counter.max_lines <= 0 || counter.num < counter.max_lines {
		written := 0
		start := 0
		for i, char := range data {
			if char == '\n' {
				counter.num++
				num, err := counter.out.Write(data[start : i+1])
				written += num
				start = i + 1
				if err != nil {
					return written, err
				}
				if counter.max_lines > 0 && counter.num >= counter.max_lines {
					return total, err
				}
			}
		}
		num, err := counter.out.Write(data[start:])
		written += num
		if err != nil {
			return written, err
		}
	} else {
		counter.truncated = true
	}
	return
}
