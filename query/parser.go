package query

import (
	"bytes"
	"errors"
	"fmt"
	"hash/adler32"
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
	if tok.Type == WS || tok.Type == COMMENT {
		return p.scan()
	}
	return tok, err
}

func (p *Parser) scanOptional(expected ...TokenType) (Token, bool, error) {
	tok, err := p.scan()
	ok := false
	for _, exp := range expected {
		if tok.Type == exp {
			ok = true
			break
		}
	}
	if !ok && err == nil {
		p.unscan()
	}
	return tok, ok, err
}

func (p *Parser) scanRequired(expectedStr string, expected ...TokenType) (Token, error) {
	tok, err := p.scan()
	if err == nil {
		for _, exp := range expected {
			if tok.Type == exp {
				return tok, err
			}
		}
		err = ParserError{
			Pos:     tok,
			Message: "Expected '" + expectedStr + "'",
		}
	}
	return tok, err
}

func (p *Parser) Parse() (Pipeline, error) {
	pipes, err := p.parsePipelines(true, false)
	if err == nil {
		_, err = p.scanRequired("EOF", EOF)
	}
	if err != nil {
		return nil, err
	} else {
		if len(pipes) == 1 {
			return pipes[0], nil
		} else {
			return Pipeline{pipes}, nil
		}
	}
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
	case OPEN, BRACKET_OPEN:
		return p.parseOpenedPipelines(tok, isInput, false)
	case STR, QUOT_STR:
		_, haveParams, err := p.scanOptional(PARAM_OPEN)
		if err != nil {
			return nil, err
		}
		if !haveParams {
			return p.parseInOutStep(tok, (isInput || isFork) && firstStep)
		}

		params, err := p.parseOpenedParams()
		if err != nil {
			return nil, err
		}
		openTok, haveOpen, err := p.scanOptional(OPEN, BRACKET_OPEN)
		if err != nil {
			return nil, err
		}
		if haveOpen {
			pipes, err := p.parseOpenedPipelines(openTok, isInput, true)
			return Fork{
				Step: Step{
					Name:   tok,
					Params: params,
				},
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

func (p *Parser) parseOpenedParams() (map[Token]Token, error) {
	if _, haveClose, err := p.scanOptional(PARAM_CLOSE); haveClose || err != nil {
		return nil, err
	}

	res := make(map[Token]Token)
	expect := 0
	var name, value Token
	var err error
	for {
		closed := false
		switch expect {
		case 0: // parameter name
			name, err = p.scanRequired("parameter name (string)", STR, QUOT_STR)
			expect = 1
		case 1: // PARAM_EQ
			_, err = p.scanRequired("=", PARAM_EQ)
			expect = 2
		case 2: // parameter value
			value, err = p.scanRequired("parameter value (string)", STR, QUOT_STR)
			expect = 3
		case 3: // PARAM_SEP or PARAM_CLOSE
			res[name] = value
			name = Token{}
			value = Token{}

			var tok Token
			tok, err = p.scan()
			if err != nil {
				break
			}
			switch tok.Type {
			case PARAM_SEP:
				expect = 0
			case PARAM_CLOSE:
				closed = true
			default:
				err = ParserError{
					Pos:     tok,
					Message: "Expected ',' or ')'",
				}
			}
		default:
			err = ParserError{
				Message: fmt.Sprintf("Unexpected 'expect' value: %v", expect),
			}
		}
		if closed || err != nil {
			break
		}
	}
	if len(res) == 0 {
		res = nil
	}
	return res, err
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

func (p *Parser) parseOpenedPipelines(openToken Token, isInput bool, isFork bool) (Pipelines, error) {
	fragments, err := p.parseOpenedPipelineFragments(openToken, isInput, isFork)
	if err != nil {
		return nil, err
	}

	pipeIndices := make(map[uint32]int) // Indices inside the pipes slice
	var defaultPipe Pipeline
	var pipes Pipelines

	for _, fragment := range fragments {
		switch fragment := fragment.(type) {
		case Pipeline:
			defaultPipe = append(defaultPipe, fragment...)
			for i := range pipes {
				pipes[i] = append(pipes[i], fragment...)
			}
		case Pipelines:
			for i, newPipe := range fragment {
				if input, hasInput := newPipe[0].(Input); hasInput {
					hash := hashInput(input)
					if pipeIndex, havePipe := pipeIndices[hash]; havePipe {
						pipes[pipeIndex] = append(pipes[pipeIndex], newPipe[1:]...) // The input is already there and should be the same as this one
					} else {
						if len(defaultPipe) > 0 {
							// Prepend the default pipe, but keep the input in front
							extendedPipe := make(Pipeline, 0, len(newPipe)+len(defaultPipe))
							extendedPipe = append(extendedPipe, newPipe[0])
							extendedPipe = append(extendedPipe, defaultPipe...)
							extendedPipe = append(extendedPipe, newPipe[1:]...)
							newPipe = extendedPipe
						}
						pipeIndices[hash] = len(pipes)
						pipes = append(pipes, newPipe)
					}
				} else {
					if i < len(pipes) {
						pipes[i] = append(pipes[i], newPipe...)
					} else {
						newPipe = append(defaultPipe, newPipe...)
						pipes = append(pipes, newPipe)
					}
				}
			}
		default:
			panic(fmt.Sprintf("parseOpenedPipelineFragments returned unexpected PipelineStep: %T %v", fragment, fragment))
		}
	}

	// If there are separate pipeline fragments intermixed in a forked pipeline, the separate fragments should also act as the default pipeline.
	// This requires constructing an artificial pipeline with an artificial input token.
	defaultInput := Input{Token{Type: STR, Lit: ""}}
	defaultHash := hashInput(defaultInput)
	needDefault := len(defaultPipe) > 0 && isFork && !isInput
	if _, haveDefault := pipeIndices[defaultHash]; needDefault && !haveDefault {
		defaultPipe = append(Pipeline{defaultInput}, defaultPipe...)
		pipes = append(pipes, defaultPipe)
	}

	return pipes, nil
}

func hashInput(input Input) uint32 {
	var buf bytes.Buffer
	for _, inputTok := range input {
		buf.WriteString(inputTok.Content())
		buf.WriteByte(0)
	}
	return adler32.Checksum(buf.Bytes())
}

func (p *Parser) parseOpenedPipelineFragments(openToken Token, isInput bool, isFork bool) ([]PipelineStep, error) {
	var result []PipelineStep
	for {
		fragment, err := p.parseOpenedPipelineFragment(openToken, isInput, isFork)
		if err != nil {
			return nil, err
		}
		result = append(result, fragment)

		var haveOpen bool
		openToken, haveOpen, err = p.scanOptional(OPEN, BRACKET_OPEN)
		if err != nil {
			return nil, err
		}
		if !haveOpen {
			break
		}
	}
	return result, nil
}

func (p *Parser) parseOpenedPipelineFragment(openToken Token, isInput bool, isFork bool) (PipelineStep, error) {
	switch openToken.Type {
	case OPEN:
		pipes, err := p.parsePipelines(isInput, isFork)
		if err == nil {
			_, err = p.scanRequired("}", CLOSE)
		}
		return pipes, err
	case BRACKET_OPEN:
		pipe, err := p.parsePipeline(isInput, false)
		if err == nil {
			_, err = p.scanRequired("]", BRACKET_CLOSE)
		}
		return pipe, err
	default:
		panic("parseOpenedPipelineFragments can only be called with OPEN or BRACKET_OPEN as openToken")
	}
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
			case Input:
				return ParserError{
					Pos:     node.Pos(),
					Message: "Intermediate pipeline step cannot be an input identifier",
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
