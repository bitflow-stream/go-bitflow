package golib

import (
	"bytes"

	"github.com/lunixbochs/vtclean"
	"golang.org/x/text/unicode/norm"
)

// Return the number of normalized utf8-runes within the cleaned string.
// Clean means no terminal escape characters and no color codes.
func StringLength(str string) (strlen int) {
	str = vtclean.Clean(str, false)
	var ia norm.Iter
	ia.InitString(norm.NFKD, str)
	for ; !ia.Done(); ia.Next() {
		strlen += 1
	}
	// Alternative:
	// strlen = utf8.RuneCountInString(str)
	return
}

// iFrom and iTo are indices to normalized utf8-runes within the cleaned string.
// Clean means no terminal escape characters and no color codes.
// All runes and special codes will be preserved in the output string.
func Substring(str string, iFrom int, iTo int) string {
	var ia norm.Iter

	// Find the start in the input string
	from := 0
	ia.InitString(norm.NFKD, str)
	buf := bytes.NewBuffer(make([]byte, 0, len(str)))
	numRunes := 0
	cleanedLen := 0
	for !ia.Done() {
		from = ia.Pos()
		cleaned := vtclean.Clean(buf.String(), false)
		if len(cleaned) > cleanedLen {
			cleanedLen = len(cleaned)
			numRunes++
		}
		if numRunes >= iFrom {
			break
		}
		part := ia.Next()
		buf.Write(part)
	}

	iLen := iTo - iFrom
	str = str[from:]
	possibleColor := false

	// Find the end in the input string
	to := len(str)
	ia.InitString(norm.NFKD, str)
	buf.Reset()
	numRunes = 0
	cleanedLen = 0
	for !ia.Done() {
		part := ia.Next()
		buf.Write(part)
		to = ia.Pos()
		bufStr := buf.String()
		cleaned := vtclean.Clean(bufStr, false)
		if len(cleaned) > cleanedLen {
			cleanedLen = len(cleaned)
			numRunes++
		}
		if len(cleaned) != len(bufStr) {
			possibleColor = true
		}
		if numRunes >= iLen {
			break
		}
	}

	str = str[:to]
	if possibleColor {
		// Might contain color codes, make sure to disable colors at the end
		str = str + "\033[0m"
	}
	return str
}
