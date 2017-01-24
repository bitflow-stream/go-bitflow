package query

import (
	"errors"
	"io"
)

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
		var step PipelineStep
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

func (p *Parser) parseStep(isInput, isFork, firstStep bool) (PipelineStep, error) {
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
			return p.parseInOutStep(tok, (isInput || isFork) && firstStep)
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

func (p *Parser) parseInOutStep(firstStep Token, isInputStep bool) (PipelineStep, error) {
	steps := []Token{firstStep}
	for {
		tok, err := p.scan()
		if err != nil {
			return nil, err
		}
		if tok.Type == STR || tok.Type == QUOT_STR {
			steps = append(steps, tok)
		} else {
			p.unscan()
			break
		}
	}
	var result PipelineStep
	if isInputStep {
		result = Input(steps)
	} else {
		if len(steps) > 1 {
			return nil, ParserError{
				Pos:     firstStep,
				Message: "Multiple sequential outputs are not allowed",
			}
		}
		result = Output(steps[0])
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

func (p Pipeline) Validate(isInput bool, isFork bool) error {
	if len(p) == 0 {
		return errors.New("Empty pipeline is not allowed") // Should not occur when parsing
	}
	isMultiplex := !isInput && !isFork

	for i, node := range p {
		switch i {
		case 0:
			if isMultiplex && len(p) > 1 {
				switch node.(type) {
				case Input, Output:
					return ParserError{
						Pos:     node.Pos(),
						Message: "Multiplexed pipeline cannot start with an identifier (string)",
					}
				}
			}
			if isFork {
				if _, ok := node.(Input); !ok {
					return ParserError{
						Pos:     node.Pos(),
						Message: "Forked pipeline must start with a pipeline identifier (string)",
					}
				}
			}
		case len(p) - 1:
			if _, ok := node.(Input); ok {
				return errors.New("The last pipeline element cannot be an Input") // Should not occur when parsing
			}
		default:
			// Intermediate pipeline step
			switch node.(type) {
			case Input, Output:
				return ParserError{
					Pos:     node.Pos(),
					Message: "Intermediate pipeline step cannot be an input or output identifier",
				}
			}
		}
	}
	if isFork && len(p) < 2 {
		return ParserError{
			Pos:     p.Pos(),
			Message: "Forked pipeline must have at least one pipeline step",
		}
	}
	return nil
}
