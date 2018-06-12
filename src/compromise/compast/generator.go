package compast

// Set of functions to create ast Nodes.

import (
	"github.com/omakoto/compromise-go/src/compromise"
	"github.com/omakoto/compromise-go/src/compromise/compfunc"
	"sync/atomic"
)

func assertType(t *Token, tokenType int, argName string) *Token {
	if t == nil {
		panic(compromise.NewSpecErrorf(nil, "%s must not be nil", argName))
	}
	if t.TokenType != tokenType {
		panic(compromise.NewSpecErrorf(t, "%s must be of type %s", argName, tokenTypeNames[tokenType]))
	}
	return t
}

func assertTypeOrNil(t *Token, tokenType int, argName string) *Token {
	if t != nil {
		assertType(t, tokenType, argName)
	}
	return t
}

var lastId int32 = -1

func newNode(nodeType int, selfToken *Token) *Node {
	return &Node{
		depth: 0,
		id:    int(atomic.AddInt32(&lastId, 1)),

		nodeType:  nodeType,
		selfToken: selfToken,
	}
}

func NewRoot() *Node {
	n := newNode(NodeRoot, nil)
	n.labels = make(map[string]*Node)
	n.commandJumpTo = make(map[string]*Node)
	n.root = n
	return n
}

func NewCommand(this, command, label *Token) *Node {
	n := newNode(NodeCommand, assertType(this, TokenCommand, "this"))
	n.command = assertType(command, TokenLiteral, "command")
	n.label = assertTypeOrNil(label, TokenLabel, "label")
	return n
}

func NewLabel(this, label *Token) *Node {
	n := newNode(NodeLabel, assertType(this, TokenCommand, "this"))
	n.label = assertType(label, TokenLabel, "label")
	return n
}

func NewCall(this, label *Token) *Node {
	n := newNode(NodeCall, assertType(this, TokenCommand, "this"))
	n.label = assertType(label, TokenLabel, "label")
	return n
}

func NewBreak(this, label *Token) *Node {
	n := newNode(NodeBreak, assertType(this, TokenCommand, "this"))
	n.label = assertTypeOrNil(label, TokenLabel, "label")
	return n
}

func NewContinue(this, label *Token) *Node {
	n := newNode(NodeContinue, assertType(this, TokenCommand, "this"))
	n.label = assertTypeOrNil(label, TokenLabel, "label")
	return n
}

func NewFinish(this *Token) *Node {
	n := newNode(NodeFinish, assertType(this, TokenCommand, "this"))
	return n
}

func NewSwitch(this *Token, pattern, label *Token) *Node {
	n := newNode(NodeSwitch, assertType(this, TokenCommand, "this"))
	n.pattern = assertTypeOrNil(pattern, TokenLiteral, "pattern")
	n.label = assertTypeOrNil(label, TokenLabel, "label")
	return n
}

func NewAny(this *Token, help *Token) *Node {
	n := newNode(NodeAny, assertType(this, TokenCommand, "this"))
	n.help = assertTypeOrNil(help, TokenHelp, "help")
	return n
}

func NewLoop(this, pattern, label *Token) *Node {
	n := newNode(NodeLoop, assertType(this, TokenCommand, "this"))
	n.pattern = assertTypeOrNil(pattern, TokenLiteral, "pattern")
	n.label = assertTypeOrNil(label, TokenLabel, "label")
	return n
}

func NewSwitchLoop(this, pattern, label *Token) *Node {
	n := newNode(NodeSwitchLoop, assertType(this, TokenCommand, "this"))
	n.pattern = assertTypeOrNil(pattern, TokenLiteral, "pattern")
	n.label = assertTypeOrNil(label, TokenLabel, "label")
	return n
}

func NewGoCall(this, funcName *Token, args []*Token) *Node {
	return newGolangCallNode(NodeGoCall, this, funcName, args)
}

func NewCandidate(this, funcName *Token, args []*Token) *Node {
	return newGolangCallNode(NodeCandidate, this, funcName, args)
}

func newGolangCallNode(nodeType int, this, funcName *Token, args []*Token) *Node {
	n := newNode(nodeType, assertType(this, TokenCommand, "this"))
	n.funcName = assertType(funcName, TokenLiteral, "funcName")
	for _, a := range args {
		assertType(a, TokenLiteral, "args")
	}
	n.args = args

	if err := compfunc.Defined(n.funcName.Word); err != nil {
		panic(compromise.NewSpecError(funcName, err.Error()))
	}

	return n
}

func NewLiteral(this, help *Token) *Node {
	n := newNode(NodeLiteral, assertType(this, TokenLiteral, "this"))
	n.literal = this
	n.help = assertTypeOrNil(help, TokenHelp, "help")
	return n
}
