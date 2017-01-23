package query

import "io"

const ExpectedPipelineStepError = "Expected pipeline step (identifier, string or '{')"

type Parser struct {
	s   *Scanner
	buf struct {
		tok        Token
		err        error
		isBuffered bool
	}
}

func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

type ParserError struct {
	Pos     Token
	Message string
}

func (e ParserError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "Unknown parser error"
	}
	if e.Pos.Type != ILLEGAL {
		msg += " (at " + e.Pos.String() + ")"
	}
	return msg
}

func (p *Parser) scanOne() (Token, error) {
	if p.buf.isBuffered {
		p.buf.isBuffered = false
		return p.buf.tok, p.buf.err
	}
	tok, err := p.s.Scan()
	p.buf.tok, p.buf.err = tok, err
	if err != nil {
		err = ParserError{
			Pos:     tok,
			Message: err.Error(),
		}
	}
	return tok, err
}

func (p *Parser) unscan() {
	p.buf.isBuffered = true
}

func (p *Parser) scan() (Token, error) {
	tok, err := p.scanOne()
	if err != nil {
		return tok, err
	}
	if tok.Type == WS {
		return p.scan()
	}
	return tok, err
}

func (p *Parser) scanOptional(expected TokenType) (Token, bool, error) {
	tok, err := p.scan()
	ok := tok.Type == expected
	if !ok && err == nil {
		p.unscan()
	}
	return tok, ok, err
}

func (p *Parser) scanRequired(expected TokenType, expectedStr string) (Token, error) {
	tok, err := p.scan()
	if err == nil && tok.Type != expected {
		err = ParserError{
			Pos:     tok,
			Message: "Expected '" + expectedStr + "'",
		}
	}
	return tok, err
}

func (p *Parser) Parse() (Pipelines, error) {
	pipes, err := p.parsePipelines(true, false)
	if err == nil {
		_, err = p.scanRequired(EOF, "EOF")
	}
	if err != nil {
		pipes = nil
	}
	return pipes, err
}

func (p *Parser) parsePipelines(isInput, isFork bool) (res Pipelines, err error) {
	for {
		var pipe Pipeline
		pipe, err = p.parsePipeline(isInput, isFork)
		if err == nil {
			err = pipe.Validate(isInput, isFork)
		}
		if err != nil {
			break
		}
		res = append(res, pipe)
		var sep bool
		if _, sep, err = p.scanOptional(SEP); !sep || err != nil {
			break
		}
	}
	return
}

func (p *Parser) parsePipeline(isInput, isFork bool) (res Pipeline, err error) {
	firstStep := true
	for {
		var step Node
		step, err = p.parseStep(isInput, isFork, firstStep)
		if err != nil {
			break
		}
		res = append(res, step)
		var next bool
		if _, next, err = p.scanOptional(NEXT); !next || err != nil {
			break
		}
		isInput = false
		firstStep = false
	}
	return
}

func (p *Parser) parseStep(isInput, isFork, firstStep bool) (Node, error) {
	tok, err := p.scan()
	if err != nil {
		return nil, err
	}
	switch tok.Type {
	case OPEN:
		return p.parseOpenedPipelines(isInput, false)
	case STR, QUOT_STR:
		params, haveParams, err := p.scanOptional(PARAM)
		if err != nil {
			return nil, err
		}
		if !haveParams {
			return p.parseEdges(tok, (isInput || isFork) && firstStep)
		}

		_, haveOpen, err := p.scanOptional(OPEN)
		if err != nil {
			return nil, err
		}
		if haveOpen {
			pipes, err := p.parseOpenedPipelines(isInput, true)
			return Fork{
				Name:      tok,
				Params:    params,
				Pipelines: pipes,
			}, err
		} else {
			return Step{
				Name:   tok,
				Params: params,
			}, nil
		}
	default:
		return nil, ParserError{
			Pos:     tok,
			Message: ExpectedPipelineStepError,
		}
	}
}

func (p *Parser) parseEdges(firstEdge Token, inputEdge bool) (Node, error) {
	edges := []Token{firstEdge}
	for {
		tok, err := p.scan()
		if err != nil {
			return nil, err
		}
		if tok.Type == STR || tok.Type == QUOT_STR {
			edges = append(edges, tok)
		} else {
			p.unscan()
			break
		}
	}
	var result Node
	if inputEdge {
		inputs := make(Inputs, len(edges))
		for i, edge := range edges {
			inputs[i] = Input{Name: edge}
		}
		if len(inputs) == 1 {
			result = inputs[0]
		} else {
			result = inputs
		}
	} else {
		if len(edges) > 1 {
			return nil, ParserError{
				Pos:     firstEdge,
				Message: "Multiple output edges are not allowed",
			}
		}
		result = Output{Name: edges[0]}
	}
	return result, nil
}

func (p *Parser) parseOpenedPipelines(isInput bool, isFork bool) (Pipelines, error) {
	pipes, err := p.parsePipelines(isInput, isFork)
	if err == nil {
		_, err = p.scanRequired(CLOSE, "}")
	}
	return pipes, err
}
