package query

import (
	"bufio"
	"bytes"
	"errors"
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
	QUOT_STR
	PARAM

	// Misc & operators
	SEP   // ;
	NEXT  // ->
	OPEN  // {
	CLOSE // }

	eof = rune(0)
)

var (
	ErrorMissingQuote          = errors.New("Unexpected EOF, missing closing '\"'")
	ErrorMissingClosingBracket = errors.New("Unexpected EOF, missing closing ')'")
	ErrorMissingNext           = errors.New("Expected '->'")
)

// These runes interrupt a non-quoted string
var specialRunes = map[rune]bool{
	';':  true,
	'{':  true,
	'}':  true,
	'-':  true,
	'(':  true,
	'"':  true,
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
	if tok.Type == QUOT_STR || tok.Type == PARAM {
		lit = lit[1 : len(tok.Lit)-1] // Strip quotes or brackets
	}
	return lit
}

func (tok Token) String() string {
	return fmt.Sprintf("[%v-%v, %v] '%v'", tok.Start, tok.End, tok.Type, tok.Lit)
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
	case PARAM:
		s = "PARAM"
	case SEP:
		s = "SEP"
	case NEXT:
		s = "NEXT"
	case OPEN:
		s = "OPEN"
	case CLOSE:
		s = "CLOSE"
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
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	s.pos++
	return ch
}

func (s *Scanner) unread() {
	s.pos--
	_ = s.r.UnreadRune()
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
	case '-':
		ch2 := s.read()
		tok.Type = NEXT
		tok.End = s.pos
		tok.Lit += string(ch2)
		if ch2 == '>' {
			return tok, nil
		} else {
			return tok, ErrorMissingNext
		}
	case '(':
		s.unread()
		return s.scanParams()
	case '"':
		s.unread()
		return s.scanQuotedStr()
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

func (s *Scanner) scanParams() (tok Token, err error) {
	tok.Type = PARAM
	tok.Start = s.pos

	var buf bytes.Buffer
	buf.WriteRune(s.read()) // Current character is '('
	for {
		if ch := s.read(); ch == eof {
			err = ErrorMissingClosingBracket
			break
		} else {
			buf.WriteRune(ch)
			if ch == ')' {
				break
			}
		}
	}
	tok.End = s.pos
	tok.Lit = buf.String()
	return
}

func (s *Scanner) scanQuotedStr() (tok Token, err error) {
	tok.Type = QUOT_STR
	tok.Start = s.pos

	var buf bytes.Buffer
	buf.WriteRune(s.read()) // Current character is '"'
	for {
		if ch := s.read(); ch == eof {
			err = ErrorMissingQuote
			break
		} else {
			buf.WriteRune(ch)
			if ch == '"' {
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
		if ch := s.read(); ch == eof {
			break
		} else if isSpecial(ch) {
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
