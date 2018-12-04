package script

import (
	"fmt"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/antongulenko/golib"
	internal "github.com/bitflow-stream/go-bitflow/script/script/internal"
)

type HintedSubscript struct {
	Index  int
	Script string
	Hints  map[string]string
}

type BitflowScriptScheduleParser struct {
	RecoverPanics bool
}

func (p *BitflowScriptScheduleParser) ParseScript(script string) ([]HintedSubscript, golib.MultiError) {
	listener := &AntlrBitflowScriptScheduleListener{
		script:       script,
		rawScript:    script,
		currentIndex: 1,
	}
	runParser(script, listener, p.RecoverPanics)
	listener.assertCleanState()
	return listener.sortedSubscripts(), listener.MultiError
}

// AntlrBitflowScriptScheduleListener is a complete listener for a parse tree produced by BitflowParser.
type AntlrBitflowScriptScheduleListener struct {
	internal.BaseBitflowListener
	antlrErrorListener

	currentIndex       int
	script             string
	rawScript          string
	schedulableScripts []HintedSubscript

	schedulingParams map[string]string
	propagationStart int
	propagatedHints  map[string]string
}

func (s *AntlrBitflowScriptScheduleListener) assertCleanState() {
	if s.schedulingParams != nil || s.propagationStart > 0 || s.propagatedHints != nil {
		s.pushError(fmt.Errorf("Unclean listener state after parsing: propagationStart %v, propagatedHints %v, scheduling params %v",
			s.propagationStart, s.propagatedHints, s.schedulingParams))
	}
}

func (s *AntlrBitflowScriptScheduleListener) sortedSubscripts() []HintedSubscript {
	res := make([]HintedSubscript, len(s.schedulableScripts))
	for _, s := range s.schedulableScripts {
		res[s.Index] = s
	}
	return res
}

// extractHintedSubscript takes a BaseParserRuleContext and replaces it's occurrence in the script with an index.
// The script is updated during parsing, but the antlr.BaseParserRuleContext's properties are static, including start and
// end position. Therefore extractHintedSubscript is keeping track of the offset between
func (s *AntlrBitflowScriptScheduleListener) addHintedSubscript(from, until int, hints map[string]string) {
	index := s.currentIndex
	s.currentIndex++
	runes := []rune(s.rawScript)
	subscript := runes[from : until+1]
	s.schedulableScripts = append(s.schedulableScripts, HintedSubscript{
		Hints:  hints,
		Index:  index,
		Script: string(subscript),
	})
	s.replaceSubscriptWithReference(from, until, index)
}

// extractHintedSubscript takes a BaseParserRuleContext and replaces it's occurrence in the script with an index.
// The script is updated during parsing, but the antlr.BaseParserRuleContext's properties are static, including start and
// end position. Therefore extractHintedSubscript is keeping track of the offset between
func (s *AntlrBitflowScriptScheduleListener) replaceSubscriptWithReference(from, until int, index int) {
	runes := []rune(s.script)

	reference := "{{" + strconv.Itoa(int(index)) + "}}"
	currentOffset := len([]rune(s.rawScript)) - len(runes)

	start := from - currentOffset
	stop := until - currentOffset + 1

	// HAVE: head of script -> current token -> tail of script
	// WANT: head of script -> {{reference}} -> tail of script
	head := append([]rune{}, runes[:start]...)      // head of script ->
	headRef := append(head, []rune(reference)...)   // head of script -> {{reference}}
	headRefTail := append(headRef, runes[stop:]...) // head of script -> {{reference}} -> tail of script
	updatedScript := headRefTail

	s.script = string(updatedScript)
}

func extractParentsStartStop(ctx *internal.SchedulingHintsContext) (start int, stop int) {
	var baseCtx *antlr.BaseParserRuleContext
	switch c := ctx.GetParent().(type) {
	case *internal.TransformContext:
		baseCtx = c.BaseParserRuleContext
	case *internal.ForkContext:
		baseCtx = c.BaseParserRuleContext
	case *internal.WindowContext:
		baseCtx = c.BaseParserRuleContext
	}
	start = baseCtx.GetStart().GetStart()
	stop = baseCtx.GetStop().GetStop()
	return start, stop
}

func (s *AntlrBitflowScriptScheduleListener) EnterSchedulingHints(ctx *internal.SchedulingHintsContext) {
	s.schedulingParams = make(map[string]string)
}

func (s *AntlrBitflowScriptScheduleListener) ExitSchedulingHints(ctx *internal.SchedulingHintsContext) {
	parentStart, parentStop := extractParentsStartStop(ctx)

	if s.propagatedHints != nil {
		// terminate ongoing scheduling-hint-propagation
		propHints := s.propagatedHints
		propStart := s.propagationStart
		s.propagatedHints = nil
		s.propagationStart = 0
		propStop := parentStart - 1 - len("->")
		s.addHintedSubscript(propStart, propStop, propHints)
	}

	hints := s.schedulingParams
	s.schedulingParams = nil
	if val, ok := hints["propagate-down"]; ok && val == "true" {
		// save state and propagate down these scheduling hints
		s.propagationStart = parentStart
		s.propagatedHints = hints
		return
	}

	// apply hints now
	s.addHintedSubscript(parentStart, parentStop, hints)
}

func (s *AntlrBitflowScriptScheduleListener) ExitSchedulingParameter(ctx *internal.SchedulingParameterContext) {
	param := ctx.Parameter().(*internal.ParameterContext)
	key := unwrapString(param.Name().(*internal.NameContext))
	val := unwrapString(param.Val().(*internal.ValContext))
	s.schedulingParams[key] = val
}

func (s *AntlrBitflowScriptScheduleListener) ExitScript(ctx *internal.ScriptContext) {
	if s.propagatedHints != nil {
		// terminate ongoing scheduling-hint-propagation
		propHints := s.propagatedHints
		propStart := s.propagationStart
		s.propagatedHints = nil
		s.propagationStart = 0
		propStop := len([]rune(s.rawScript)) - 1
		s.addHintedSubscript(propStart, propStop, propHints)
	}

	// add the root script with Index 0
	s.schedulableScripts = append(s.schedulableScripts, HintedSubscript{
		Script: s.script,
		Index:  0,
	})
}
