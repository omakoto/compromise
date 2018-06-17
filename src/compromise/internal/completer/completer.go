package compengine

// This is the core of the completion logic.

import (
	"fmt"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compast"
	"github.com/omakoto/compromise/src/compromise/compenv"
	"github.com/omakoto/compromise/src/compromise/compfunc"
	"github.com/omakoto/compromise/src/compromise/internal/adapters"
	"github.com/omakoto/compromise/src/compromise/internal/compdebug"
	"github.com/omakoto/compromise/src/compromise/internal/compstore"
	"github.com/omakoto/compromise/src/compromise/internal/parser"
	"github.com/omakoto/compromise/src/compromise/internal/selectors"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/utils"
	"sort"
	"sync/atomic"
)

type Engine struct {
	adapter     adapters.ShellAdapter
	commandLine *adapters.CommandLine

	currentIndex int

	astRoot *compast.Node

	candidates []compromise.Candidate

	directives *compromise.Directives
}

func NewEngine(adapter adapters.ShellAdapter, commandLine *adapters.CommandLine, d *compromise.Directives) *Engine {
	e := &Engine{
		adapter:     adapter,
		commandLine: commandLine,
		directives:  d,
	}
	compdebug.Dump("CommandLine=", commandLine)
	return e
}

func (e *Engine) ParseSpec(spec string) {
	ast := parser.Parse(spec, e.directives)
	if compenv.DebugEnabled {
		compdebug.Debugf("Spec=%s\n", ast.Dump(true))
	}
	e.astRoot = ast
}

func (e *Engine) Run() {
	compdebug.Debugf("Run() start\n")

	// Find the start node.
	e.adapter.StartCompletion(e.commandLine)
	defer e.adapter.EndCompletion()

	store := compstore.Load()
	cacheAge := store.LastCompletionAge()
	compdebug.Debugf("Cache age: %v\n", cacheAge)
	if store.NumConsecutiveInvocations > 1 && cacheAge <= compenv.CacheTimeout {
		cached, _ := compstore.LoadCandidates()
		if len(cached) > 0 {
			e.addCandidates(cached...)
		}
	}
	if e.candidates == nil {
		// Execute.
		compdebug.Time("Maybe override", func() {
			e.addCandidates(e.adapter.MaybeOverrideCandidates(e.commandLine)...)
		})
		if e.candidates != nil {
			compdebug.Dump("  -> Candidates overridden=", e.candidates)
		} else {
			compdebug.Time("Execute", func() {
				e.execute()
			})
		}

		// Push the result.
		if compenv.DebugEnabled {
			compdebug.Debug("Final candidates:\n")
			for _, c := range e.candidates {
				compdebug.Debugf("  %s\n", c)
			}
		}

		// Sort the result.
		sort.Slice(e.candidates, func(i, j int) bool {
			return e.candidates[i].Value() < e.candidates[j].Value()
		})

		// Cache the candidates.
		compstore.CacheCandidates(e.candidates)
	}

	if len(e.candidates) == 0 {
		return
	}

	// Maybe try FZF.
	if e.adapter.SupportsFzf() && compenv.UseFzf && compstore.Load().IsDoublePress {
		compdebug.Debug("Trying FZF\n")
		fzf := selectors.NewFzfSelector()
		selected, err := fzf.Select(e.commandLine.WordAtCursor(0), e.candidates)
		if err != nil {
			compdebug.Warnf("Unable to execute FZF: %s\n", err.Error())
		} else if selected != nil {
			// Selected by FZF. Use it.
			e.adapter.AddCandidate(selected)
			return
		} else {
			// FZF Launched, and the user didn't select any -> cancel, but go ahead and
			// let bash show all help / candidates.
			// TODO: Maybe only show help?
		}
	} else {
		compdebug.Debug("Not using FZF\n")
	}

	// Pass back to the shell.
	for _, c := range e.candidates {
		e.adapter.AddCandidate(c)
	}
}

func (e *Engine) addCandidates(candidates ...compromise.Candidate) {
	w := e.commandLine.WordAtCursor(0)
	for _, c := range candidates {
		compdebug.Debugf("  -> Candidate: %v", c)
		if c.Matches(w) {
			compdebug.Debug(" [Matched]")
			e.candidates = append(e.candidates, c)
		}
		compdebug.Debug("\n")
	}
}

func (e *Engine) advancePc(reason string) {
	e.commandLine.AdvancePc(1)
	compdebug.Debugf("[advance: %s]\n", reason)
}

func (e *Engine) collecting() bool {
	return e.commandLine.AtCursor()
}

// execute starts execution at the first node.
func (e *Engine) execute() {
	compdebug.Debugf("runInner() start\n")

	start := e.astRoot.GetStartNodeForCommand(e.commandLine.Command())
	e.commandLine.SetPc(1)

	f := doWithFlowControl(nil, func() {
		m := false
		e.executeNode(start, false, &m)
	})
	if f != nil && f.nodeType != compast.NodeFinish && f.nodeType != compast.NodeLabel {
		s := f.sourceNode
		panic(compromise.NewSpecErrorf(s.SelfToken(), "unexpected flow control %q (with label %q)", s.NodeTypeString(), s.LabelWord()))
	}
}

var lastDebugId int32 = -1

func debugId() int {
	return int(atomic.AddInt32(&lastDebugId, 1))
}

func (e *Engine) executeNode(n *compast.Node, inSwitch bool, matched *bool) {
	if n == nil {
		return
	}

	id := debugId()

	compdebug.Indent()
	defer compdebug.Unindent()

	cl := e.commandLine
	for ; !cl.AfterCursor() && n != nil; n = n.Next() {
		compdebug.Debugf("[#%d] At %q (%d/%d) : executing %s (in-switch=%v)\n", id, cl.RawWordAt(0), cl.Pc(), cl.CursorIndex(), n, inSwitch)

		n.UpdateLastVisitedWordIndex(e.commandLine.Pc())

		m := false
		collecting := e.collecting()

		if n.NodeType() == compast.NodeCommand {
			// Just skip and move to next. Don't advance PC.
			continue
		}

		utils.DoAndEnsure(func() {
			switch n.NodeType() {
			case compast.NodeLabel: // Note: for flow control purposes, it's used as return.
				panic(newFlowControl(n))

			case compast.NodeFinish, compast.NodeBreak, compast.NodeContinue:
				panic(newFlowControl(n))

			case compast.NodeSwitch:
				e.executeSwitchLoop(n, inSwitch, true, false, &m)
			case compast.NodeSwitchLoop:
				e.executeSwitchLoop(n, inSwitch, true, true, &m)
			case compast.NodeLoop:
				e.executeSwitchLoop(n, inSwitch, false, true, &m)

			case compast.NodeAny:
				e.executeAny(n, inSwitch, &m)
			case compast.NodeCandidate:
				e.executeCandidate(n, inSwitch, &m)
			case compast.NodeLiteral:
				e.executeLiteral(n, inSwitch, &m)

			case compast.NodeCall:
				e.executeCall(n, inSwitch, &m)
			case compast.NodeGoCall:
				e.executeGoCall(n, &m)
			default:
				panic(fmt.Errorf("unexpected node %s", n))
			}
		}, func() {
			compdebug.Debugf("[#%d] result=%v\n", id, m)
			if m {
				*matched = true
			}
		})

		if inSwitch {
			if collecting {
				compdebug.Debug("[next: in switch and collecting]\n")
				continue
			}
			if m {
				compdebug.Debug("[return: in switch and match found]\n")
				break
			}
			compdebug.Debug("[next: in switch and match not found]\n")
			continue
		}
		// Sequential.
		if !m {
			panicFinishf("[#%d] sequential and didn't match", id)
		}
		compdebug.Debug("[next: sequential and matched]\n")
	}
}

func (e *Engine) executeCall(n *compast.Node, inSwitch bool, matched *bool) {
	// @call: Jump to the target label, and then execute from it's next, until "return" is detected.

	label := n.LabelWord()
	target := e.astRoot.GetLabeledNode(label, n.Label()).Child()

	doWithFlowControl(catchReturn(), func() {
		e.executeNode(target, inSwitch, matched)
	})
}

func (e *Engine) executeAny(n *compast.Node, inSwitch bool, matched *bool) {
	// @any: Any matches any word but generates no candidates.
	//e.executeCandidateNode(n, inSwitch, compfunc.TakeAny, matched)
	e.executeCandidateNode(n, inSwitch, func() compromise.CandidateList {
		return compromise.OpenCandidates(n.AsCandidates()...)
	}, matched)
}

func (e *Engine) executeCandidate(n *compast.Node, inSwitch bool, matched *bool) {
	// @cand: lazily generate candidates if necessary. It matches any word.
	funcName := n.FuncName().Word

	e.executeCandidateNode(n, inSwitch, func() compromise.CandidateList {
		return compfunc.Invoke(funcName, e.commandLine, n.Args())
	}, matched)
}

func (e *Engine) executeLiteral(n *compast.Node, inSwitch bool, matched *bool) {
	// Literal (such as -f, etc): Emits itself as a candidate. Only the exact same word will match.
	e.executeCandidateNode(n, inSwitch, func() compromise.CandidateList {
		return compromise.StrictCandidates(n.AsCandidates()...)
	}, matched)
}

func (e *Engine) executeCandidateNode(n *compast.Node, inSwitch bool, genCands func() compromise.CandidateList, matched *bool) {
	common.OrPanicf(!e.commandLine.AfterCursor(), "must not be after cursor")

	curWord := e.commandLine.WordAt(0)

	if e.collecting() {
		e.addCandidates(genCands().GetCandidate(curWord)...)
		if !inSwitch {
			panicFinish("cursor word consumed")
		}
		*matched = true
		return
	}

	// Otherwise, if it has children, we need to go deeper.
	if genCands().MatchesFully(curWord) {
		e.advancePc("literal matched")

		// Note we don't need to propagate matched here. As long as this node matches,
		// we still report "match" to the caller.
		m := false
		*matched = true
		e.executeNode(n.Child(), false, &m)
		return
	}
	*matched = false
}

func (e *Engine) executeGoCall(n *compast.Node, matched *bool) {
	// @go_call: When we reach a @go_call, we always executes the function, but no states will change.
	funcName := n.FuncName().Word
	*matched = true // Always matches but don't advance PC.
	ret := compfunc.Invoke(funcName, e.commandLine, n.Args())
	if ret != nil {
		panic(compromise.NewSpecErrorf(n.FuncName(), "@go_call function %s must not return values", funcName))
	}
}

func (e *Engine) executeSwitchLoop(n *compast.Node, inSwitch, doSwitch bool, doLoop bool, matched *bool) {
	myLabel := n.LabelWord()

	id := debugId()

	compdebug.Indent()
	defer compdebug.Unindent()
	compdebug.Debugf("[#%d] executeSwitchLoop s=%v, l=%v node=%s (in-switch=%v)\n", id, doSwitch, doLoop, n, inSwitch)
	defer compdebug.Debugf("[#%d] executeSwitchLoop done\n", id)

	if n.Child() == nil {
		panic(compromise.NewSpecErrorf(n.SelfToken(), "%s must have at least one child", n.SelfToken()))
	}

	for !e.commandLine.AfterCursor() {
		// See if the current token is accepted by this loop.
		if e.commandLine.BeforeCursor() && !n.PatternMatches(e.commandLine.WordAt(0)) {
			compdebug.Debugf("| %q is not accepted\n", e.commandLine.WordAt(0))
			*matched = true
			break
		}

		startPc := e.commandLine.Pc()
		m := false
		collecting := e.collecting()

		var fc *flowControl
		utils.DoAndEnsure(func() {
			fc = doWithFlowControl(catchLoopControl(myLabel), func() {
				e.executeNode(n.Child(), doSwitch, &m)
			})
		}, func() {
			compdebug.Debugf("[#%d]  result=%v\n", id, m)
			if m {
				*matched = true
			}
		})

		if collecting {
			if !inSwitch && n.PatternMatches(e.commandLine.WordAt(0)) {
				panicFinish("cursor word consumed")
			}
			compdebug.Debugf("[#%d] still collecting, continuing to the caller...\n", id)
			break
		}
		if !doLoop {
			break // Not looping.
		}
		if fc != nil && fc.nodeType == compast.NodeBreak {
			compdebug.Debugf("[#%d] break detected\n", id)
			break
		}

		// If looping, ensure we always advance at least by one.
		if startPc == e.commandLine.Pc() {
			e.advancePc("forced in loop")
		}
		compdebug.Debugf("[#%d] loop n=%s\n", id, n)
	}
}
