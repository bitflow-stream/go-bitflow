package query

type Node interface {
	Pos() Token
}

type PipelineStep interface {
	Node
}

type Pipelines []Pipeline

func (p Pipelines) Pos() (res Token) {
	if len(p) > 0 {
		res = p[0].Pos()
	}
	return
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
type Pipeline []PipelineStep

func (p Pipeline) Pos() (res Token) {
	if len(p) > 0 {
		res = p[0].Pos()
	}
	return
}

type Fork struct {
	Step
	Pipelines Pipelines
}

func (p Fork) Pos() Token {
	return p.Name
}

type Step struct {
	Name   Token
	Params map[Token]Token
}

func (p Step) Pos() Token {
	return p.Name
}

// Data input (file, console, TCP, ...)
// Inside a Fork, this identifies the pipeline.
type Input []Token

func (p Input) Pos() (res Token) {
	if len(p) > 0 {
		res = p[0]
	}
	return
}

// Data output (file, console, TCP, ...)
type Output Token

func (p Output) Pos() Token {
	return Token(p)
}
