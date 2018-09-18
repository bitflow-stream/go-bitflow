package script // Bitflow

import (
	"fmt"
)

type token struct {
	Lit   string
	Start int
	End   int
}

func (tok token) Content() string {
	return stripQuotes(tok.Lit)
}

func (tok token) String() string {
	return fmt.Sprintf("[%v-%v] '%v'", tok.Start, tok.End, tok.Lit)
}

func convertParams(p map[token]token) map[string]string {
	params := make(map[string]string, len(p))
	for key, value := range p {
		params[key.Content()] = value.Content()
	}
	return params
}
