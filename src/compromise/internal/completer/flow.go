package compengine

// Defines a struct used with panic() for flow control purposes.

import (
	"fmt"
	"github.com/omakoto/compromise/src/compromise/compast"
	"github.com/omakoto/compromise/src/compromise/internal/compdebug"
)

// flowControl will be thrown with panic() with a flow control node such as NodeBreak for global escaping.
type flowControl struct {
	nodeType   int
	sourceNode *compast.Node // optional but may not be nil if it's not finish.
}

func newFlowControl(node *compast.Node) *flowControl {
	return &flowControl{node.NodeType(), node}
}

func panicFinish(reason string) {
	compdebug.Debugf("[finish: %s]\n", reason)
	panic(&flowControl{compast.NodeFinish, nil})
}

func panicFinishf(reasonFormat string, args ...interface{}) {
	panicFinish(fmt.Sprintf(reasonFormat, args...))
}

// catchLoopControl generates a filter to catch break/continue for a structure with a given label.
func catchLoopControl(myLabel string) func(fc *flowControl) bool {
	return func(fc *flowControl) bool {
		switch fc.nodeType {
		case compast.NodeBreak, compast.NodeContinue:
			toLabel := fc.sourceNode.LabelWord()
			return toLabel == myLabel || toLabel == ""
		}
		return false
	}
}

// catchLoopControl generates a filter to catch "return" (which is NodeLabel).
func catchReturn() func(fc *flowControl) bool {
	return func(fc *flowControl) bool {
		switch fc.nodeType {
		case compast.NodeLabel:
			return true
		}
		return false
	}
}

// Executes a function fun. If it catches a flowControl panic that matches the given filter, swallow and return it.
func doWithFlowControl(filter func(*flowControl) bool, fun func()) (ret *flowControl) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*flowControl); ok && (filter == nil || filter(e)) {
				ret = e
			} else {
				panic(r)
			}
		}
	}()
	fun()
	return nil
}
