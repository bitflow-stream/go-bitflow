package query

import "errors"

type Node interface {
	Children() []Node
	Pos() Token
}

func posOfFirstChild(node Node) (res Token) {
	children := node.Children()
	if len(children) > 0 {
		res = children[0].Pos()
	}
	return
}

type LeafNode struct {
}

func (LeafNode) Children() []Node {
	return nil
}

type Pipelines []Pipeline

func (p Pipelines) Children() []Node {
	res := make([]Node, len(p))
	for i, n := range p {
		res[i] = n
	}
	return res
}

func (p Pipelines) Pos() (res Token) {
	return posOfFirstChild(p)
}

// Depending on their context, pipelines may contain different elements.
// Any intermediate element (not the first and not the last) must be a Step, Fork or Pipelines.
// The last element (if it is not the only element) can be any of that, and also an Output.
// The first element depends on the context:
//  - if the pipeline has no predecessor (does not receive data from another pipeline or step),
//    it can start with anything, including an Input or Inputs
//  - if the pipeline is inside a Fork, it must start with Input or Inputs
//  - if the pipeline is inside a multiplex fork (a Pipelines without a surrounding Fork, that also has a predecessor),
//    it cannot start with an Input or Inputs
type Pipeline []Node

func (p Pipeline) Children() []Node {
	return p
}

func (p Pipeline) Pos() (res Token) {
	return posOfFirstChild(p)
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
				case Input, Inputs, Output:
					return ParserError{
						Pos:     node.Pos(),
						Message: "Multiplexed pipeline cannot start with an identifier (string)",
					}
				}
			}
			if isFork {
				switch node.(type) {
				case Inputs, Input:
				default:
					return ParserError{
						Pos:     node.Pos(),
						Message: "Forked pipeline must start with a pipeline identifier (string)",
					}
				}
			}
		case len(p) - 1:
			switch node.(type) {
			case Input, Inputs:
				return errors.New("The last pipeline element cannot be an Input") // Should not occur when parsing
			}
		default:
			// Intermediate pipeline step
			switch node.(type) {
			case Inputs, Input, Output:
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

type Fork struct {
	Name      Token
	Params    Token
	Pipelines Pipelines
}

func (f Fork) Children() []Node {
	return []Node{f.Pipelines}
}

func (p Fork) Pos() Token {
	return p.Name
}

type Step struct {
	LeafNode
	Name   Token
	Params Token
}

func (p Step) Pos() Token {
	return p.Name
}

// Data input (file, console, TCP, ...)
// Inside a Fork, this identifies the pipeline.
type Input struct {
	LeafNode
	Name Token
}

func (p Input) Pos() Token {
	return p.Name
}

// Data output (file, console, TCP, ...)
type Output struct {
	LeafNode
	Name Token
}

func (p Output) Pos() Token {
	return p.Name
}

type Inputs []Input

func (e Inputs) Children() []Node {
	res := make([]Node, len(e))
	for e, n := range e {
		res[e] = n
	}
	return res
}

func (e Inputs) Pos() (res Token) {
	return posOfFirstChild(e)
}
