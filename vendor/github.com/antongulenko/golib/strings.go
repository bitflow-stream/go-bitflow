package golib

import (
	"bytes"

	"golang.org/x/text/width"

	"github.com/lunixbochs/vtclean"
)

// Return the number of normalized utf8-runes within the cleaned string.
// Clean means no terminal escape characters and no color codes.
func StringLength(str string) (strlen int) {
	str = vtclean.Clean(str, false)
	for str != "" {
		_, rest, width := ReadRune(str)
		str = rest
		strlen += width
	}
	// Alternative:
	// strlen = utf8.RuneCountInString(str)
	// Alternative:
	// var ia norm.Iter
	// ia.InitString(norm.NFKD, str)
	return
}

func ReadRune(input string) (theRune string, rest string, runeWidth int) {
	prop, size := width.LookupString(input)
	rest = input[size:]
	theRune = input[:size]
	switch prop.Kind() {
	case width.EastAsianFullwidth, width.EastAsianHalfwidth, width.EastAsianWide:
		runeWidth = 2
	default:
		runeWidth = 1
	}
	return
}

// iFrom and iTo are indices to normalized utf8-runes within the cleaned string.
// Clean means no terminal escape characters and no color codes.
// All runes and special codes will be preserved in the output string.
func Substring(str string, iFrom int, iTo int) string {

	// Find the start in the input string
	buf := bytes.NewBuffer(make([]byte, 0, len(str)))
	textWidth := 0
	cleanedLen := 0
	for str != "" && textWidth < iFrom {
		runeStr, rest, width := ReadRune(str)
		buf.Write([]byte(runeStr))

		cleaned := vtclean.Clean(buf.String(), false)
		if len(cleaned) > cleanedLen {
			// A visible rune was added
			cleanedLen = len(cleaned)
			textWidth += width
		}
		if textWidth >= iFrom {
			break
		}
		str = rest
	}

	iLen := iTo - iFrom
	possibleColor := false

	// Find the end in the input string
	to := 0
	suffix := str[:]
	buf.Reset()
	textWidth = 0
	cleanedLen = 0
	for str != "" && textWidth < iLen {
		runeStr, rest, width := ReadRune(str)
		buf.Write([]byte(runeStr))

		cleaned := vtclean.Clean(buf.String(), false)
		if len(cleaned) > cleanedLen {
			// A visible rune was added
			cleanedLen = len(cleaned)
			textWidth += width
		}
		if textWidth > iLen {
			break
		}
		str = rest
		to += len(runeStr)
	}

	suffix = suffix[:to]
	if possibleColor {
		// Might contain color codes, make sure to disable colors at the end
		suffix += "\033[0m"
	}
	return suffix
}
