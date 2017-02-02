package query

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type TokenType int

const (
	// Special
	ILLEGAL TokenType = iota
	EOF
	WS

	// Literals
	STR
	QUOT_STR // Surrounded by one of: " ' `

	// Parameters
	PARAM_OPEN  // (
	PARAM_CLOSE // )
	PARAM_SEP   // ,
	PARAM_EQ    // =

	// Misc & operators
	SEP   // ;
	NEXT  // ->
	OPEN  // {
	CLOSE // }

	eof = rune(0)
)

var (
	ErrorMissingQuote = "Unexpected EOF, missing closing %v quote"
)

// These runes interrupt a non-quoted string
// The '-' rune is handled specially because it is part of the two-rune token '->'
var specialRunes = map[rune]bool{
	';':  true,
	'{':  true,
	'}':  true,
	'"':  true,
	'\'': true,
	'`':  true,
	'(':  true,
	')':  true,
	'=':  true,
	',':  true,
	' ':  true,
	'\t': true,
	'\n': true,
}

type Token struct {
	Type  TokenType
	Lit   string
	Start int
	End   int
}

func (tok Token) Content() string {
	lit := tok.Lit
	if tok.Type == QUOT_STR {
		lit = lit[1 : len(tok.Lit)-1] // Strip quotes
	}
	return lit
}

func (tok Token) String() string {
	typ := tok.Type.String()
	lit := tok.Content()
	if typ == lit {
		typ = ""
	} else {
		typ = ", " + typ
	}
	return fmt.Sprintf("[%v-%v%v] '%v'", tok.Start, tok.End, typ, tok.Lit)
}

func (t TokenType) String() (s string) {
	switch t {
	case ILLEGAL:
		s = "ILLEGAL"
	case EOF:
		s = "EOF"
	case WS:
		s = "WS"
	case STR:
		s = "STR"
	case QUOT_STR:
		s = "QUOT_STR"
	case PARAM_OPEN:
		s = "("
	case PARAM_CLOSE:
		s = ")"
	case PARAM_EQ:
		s = "="
	case PARAM_SEP:
		s = ","
	case SEP:
		s = ";"
	case NEXT:
		s = "->"
	case OPEN:
		s = "{"
	case CLOSE:
		s = "}"
	default:
		s = fmt.Sprintf("UNKNOWN_TOKEN_TYPE(%v)", t)
	}
	return s
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isSpecial(ch rune) bool {
	return specialRunes[ch]
}

type Scanner struct {
	r   *bufio.Reader
	pos int

	// For allowing two consecutive unread() operations to support the '->' token
	buf  [2]rune
	nbuf int
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	switch s.nbuf {
	case 0:
		ch, _, err := s.r.ReadRune()
		s.buf[0], s.buf[1] = ch, s.buf[0]
		if err != nil {
			return eof
		}
		s.pos++
		return ch
	case 1, 2:
		s.pos++
		s.nbuf--
		return s.buf[s.nbuf]
	default:
		panic("Too many Scanner.unread() operations have been made")
		return eof
	}
}

func (s *Scanner) unread() {
	if s.nbuf >= 2 {
		panic("Cannot perform more than 2 Scanner.unread() operations")
	}
	s.pos--
	s.nbuf++
}

func (s *Scanner) Scan() (Token, error) {
	start_pos := s.pos
	ch := s.read()

	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace(), nil
	}

	tok := Token{
		Start: start_pos,
		End:   s.pos,
	}
	tok.Lit = string(ch)

	switch ch {
	case eof:
		tok.Type = EOF
		return tok, nil
	case '{':
		tok.Type = OPEN
		return tok, nil
	case '}':
		tok.Type = CLOSE
		return tok, nil
	case ';':
		tok.Type = SEP
		return tok, nil
	case '(':
		tok.Type = PARAM_OPEN
		return tok, nil
	case ')':
		tok.Type = PARAM_CLOSE
		return tok, nil
	case '=':
		tok.Type = PARAM_EQ
		return tok, nil
	case ',':
		tok.Type = PARAM_SEP
		return tok, nil
	case '"', '`', '\'':
		s.unread()
		return s.scanQuotedStr(ch)
	default:
		s.unread()
		return s.scanDirectStr(), nil
	}
}

func (s *Scanner) scanWhitespace() Token {
	tok := Token{
		Type:  WS,
		Start: s.pos,
	}

	var buf bytes.Buffer
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}
	tok.End = s.pos
	tok.Lit = buf.String()
	return tok
}

func (s *Scanner) scanQuotedStr(quoteRune rune) (tok Token, err error) {
	tok.Type = QUOT_STR
	tok.Start = s.pos

	var buf bytes.Buffer
	buf.WriteRune(s.read()) // Current character is the opening quote
	for {
		if ch := s.read(); ch == eof {
			err = fmt.Errorf(ErrorMissingQuote, string(quoteRune))
			break
		} else {
			buf.WriteRune(ch)
			if ch == quoteRune {
				break
			}
		}
	}
	tok.End = s.pos
	tok.Lit = buf.String()
	return
}

func (s *Scanner) scanDirectStr() Token {
	tok := Token{
		Type:  STR,
		Start: s.pos,
	}
	var buf bytes.Buffer
	for {
		ch := s.read()
		if ch == eof {
			break
		} else if ch == '-' {
			ch2 := s.read()
			if ch2 == '>' {
				if buf.Len() == 0 {
					tok.Type = NEXT
					buf.WriteRune(ch)
					buf.WriteRune(ch2)
				} else {
					// Direct string is interrupted by a complete '->' token
					s.unread()
					s.unread()
				}
				break
			} else {
				s.unread()
			}
		} else if isSpecial(ch) {
			s.unread()
			break
		}
		buf.WriteRune(ch)
	}
	tok.End = s.pos
	tok.Lit = buf.String()
	return tok
}
