package script_go

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
// Any element can be Output, Step, Fork or Pipelines.
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

func (f Fork) Pos() Token {
	return f.Name
}

type Step struct {
	Name   Token
	Params map[Token]Token
}

func (step Step) Pos() Token {
	return step.Name
}

func (step Step) ParamsMap() map[string]string {
	res := make(map[string]string, len(step.Params))
	for key, value := range step.Params {
		res[key.Content()] = value.Content()
	}
	return res
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

type MultiInput struct {
	Pipelines
}
