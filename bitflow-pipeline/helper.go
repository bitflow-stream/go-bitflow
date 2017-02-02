package main

import (
	"bufio"
	"bytes"
)

const nul = rune(0)

func SplitShellCommand(s string) []string {
	scanner := bufio.NewScanner(bytes.NewBuffer([]byte(s)))
	scanner.Split(bufio.ScanRunes)
	var res []string
	var buf bytes.Buffer
	quote := nul
	for scanner.Scan() {
		r := rune(scanner.Text()[0])
		flush := false
		switch quote {
		case nul:
			switch r {
			case ' ', '\t', '\r', '\n':
				flush = true
			case '"', '\'':
				quote = r
				flush = true
			}
		case '"', '\'':
			if r == quote {
				flush = true
				quote = nul
			}
		}

		if flush {
			if buf.Len() > 0 {
				res = append(res, buf.String())
				buf.Reset()
			}
		} else {
			buf.WriteRune(r)
		}
	}

	// Un-closed quotes are ignored
	if buf.Len() > 0 {
		res = append(res, buf.String())
	}
	return res
}
