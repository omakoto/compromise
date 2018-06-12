package compast

// AST (sort of...) for compromise.

import (
	"fmt"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/go-common/src/common"
	"math"
	"regexp"
	"strings"
)

const (
	TokenCommand = iota
	TokenLiteral
	TokenLabel
	TokenHelp

	// This is used as an expected type.
	TokenAny
)

var tokenTypeNames = []string{
	"Command",
	"Literal",
	"Label",
	"Help",
	"Any",
}

type Token struct {
	TokenType int
	Word      string
	RawWord   string

	IndexInLine int

	SourceFile string
	Line       int
	Column     int
}

var _ compromise.SourceLocation = ((*Token)(nil))

func (t *Token) String() string {
	return fmt.Sprintf("%q", t.RawWord)
}

func (t *Token) SourceLocation() (string, int, int) {
	return t.SourceFile, t.Line, t.Column
}

const (
	NodeRoot = iota
	NodeCommand
	NodeLabel
	NodeCall
	NodeFinish
	NodeLoop
	NodeSwitch
	NodeSwitchLoop
	NodeAny
	NodeBreak
	NodeContinue
	NodeGoCall
	NodeCandidate
	NodeLiteral
)

var nodeTypeNames = []string{
	"Root",
	"Command",
	"Label",
	"Call",
	"Finish",
	"Loop",
	"Switch",
	"SwitchLoop",
	"Any",
	"Break",
	"Continue",
	"GoCall",
	"Candidate",
	"Literal",
}

// Node implements a tree of Tokens. This tree is a basic AST of the completion spec.
type Node struct {
	// For debugging
	id    int
	depth int

	// For structuring a tree
	root        *Node
	parent      *Node
	child       *Node
	lastChild   *Node
	next        *Node
	numChildren int

	// Node information
	nodeType  int
	selfToken *Token

	// Parameters
	literal  *Token
	command  *Token
	pattern  *Token
	funcName *Token
	label    *Token
	help     *Token
	args     []*Token

	// Used to detect infinity loop.
	lastVisitedWordIndex int

	// Only root has the following items.
	labels        map[string]*Node
	commandJumpTo map[string]*Node
}

func (n *Node) NodeType() int {
	return n.nodeType
}

func (n *Node) NodeTypeString() string {
	return nodeTypeNames[n.nodeType]
}

func (n *Node) SelfToken() *Token {
	return n.selfToken
}

func (n *Node) maxChildren() int {
	switch n.nodeType {
	case NodeCommand, NodeCall, NodeFinish, NodeGoCall:
		return 0
	}
	return math.MaxInt32
}

// AddChild adds a child new to node n.
func (n *Node) AddChild(new *Node) {
	if new.parent != nil {
		panic(fmt.Sprintf("node %v already has parent set", new))
	}
	maxChildren := n.maxChildren()
	if n.numChildren+1 > maxChildren {
		if n.maxChildren() == 0 {
			panic(compromise.NewSpecErrorf(n.selfToken, "%s takes no children", n.selfToken))
		} else {
			panic(compromise.NewSpecErrorf(n.selfToken, "%s takes at most %d children", n.selfToken, maxChildren))
		}
	}
	switch new.nodeType {
	case NodeCommand:
		n.root.commandJumpTo[new.command.Word] = new
	case NodeLabel:
		n.root.labels[strings.ToLower(new.label.Word)] = new
	}

	if n.lastChild == nil {
		// First child of Node n.
		n.child = new
		n.lastChild = new
		n.numChildren++

		new.parent = n
		new.root = n.root
		new.depth = n.depth + 1
		return
	}
	// Node n already has a child.
	n.lastChild.addNext(new)
}

// AddNext adds a child new to the parent of node n.
func (n *Node) addNext(new *Node) {
	if n.next != nil {
		panic(fmt.Sprintf("node %v already has next set", n))
	}
	if new.parent != nil {
		panic(fmt.Sprintf("node %v already has parent set", new))
	}
	if n.nodeType == NodeLabel && new.nodeType != NodeLabel {
		panic(compromise.NewSpecErrorf(new.selfToken, "only @label can appear at the top level"))
	}

	n.next = new
	n.parent.lastChild = new
	n.parent.numChildren++

	new.parent = n.parent
	new.root = n.root
	new.depth = n.depth
}

func (n *Node) Root() *Node {
	return n.root
}

func (n *Node) Parent() *Node {
	return n.parent
}

func (n *Node) Child() *Node {
	return n.child
}

func (n *Node) Next() *Node {
	return n.next
}

// GetLabeledNode returns the NodeLabel with a given label. Only supported by a root node.
func (n *Node) GetLabeledNode(label string, referrer *Token) *Node {
	if n, ok := n.labels[strings.ToLower(label)]; ok {
		return n
	}
	panic(compromise.NewSpecErrorf(referrer, "undefined label :%s", label))
}

// GetStartNodeForCommand takes a command name given by completion and returns the starting NodeLabel
// for the command, taking @command's into account.
func (n *Node) GetStartNodeForCommand(command string) *Node {
	if commandNode, ok := n.commandJumpTo[command]; ok && commandNode.label != nil {
		return n.GetLabeledNode(commandNode.label.Word, commandNode.command).Child()
	}
	return n.Child()
}

func (n *Node) Literal() *Token {
	return n.literal
}

func (n *Node) Command() *Token {
	return n.command
}

func (n *Node) Pattern() *Token {
	return n.pattern
}

func (n *Node) PatternMatches(s string) bool {
	pat := n.pattern
	if pat == nil {
		return true
	}
	m, err := regexp.MatchString(pat.Word, s)
	if err != nil {
		panic(compromise.NewSpecErrorf(pat, "invalid regex %q", pat.Word))
	}
	return m
}

func (n *Node) FuncName() *Token {
	return n.funcName
}

func (n *Node) Label() *Token {
	return n.label
}

func (n *Node) LabelWord() string {
	if n.label != nil {
		return n.label.Word
	}
	return ""
}

func (n *Node) Help() *Token {
	return n.help
}

func (n *Node) SetHelp(b *compromise.CandidateBuilder) *compromise.CandidateBuilder {
	if n.Help() != nil {
		b.Help(n.help.Word)
	}
	return b
}

func (n *Node) UpdateLastVisitedWordIndex(index int) {
	if n.lastVisitedWordIndex == index {
		common.Panicf("Node %s already visited for index %d", n, index)
	}
	n.lastVisitedWordIndex = index
}

func (n *Node) AsCandidates() []compromise.Candidate {
	switch n.nodeType {
	case NodeAny:
		b := compromise.NewCandidateBuilder()
		return []compromise.Candidate{n.SetHelp(b).Force(true).Build()}
	case NodeLiteral:
		raw := n.Literal().RawWord
		if strings.HasPrefix(raw, "\"") || !strings.ContainsRune(raw, '|') {
			// Escaped word, or contains no special character. Just create a candidate.
			b := compromise.NewCandidateBuilder().Value(n.Literal().Word)
			return []compromise.Candidate{n.SetHelp(b).Build()}
		}
		fields := strings.Split(n.Literal().Word, "|")
		ret := make([]compromise.Candidate, 0, len(fields))
		for _, t := range fields {
			b := compromise.NewCandidateBuilder().Value(strings.Trim(t, " \t\r\n"))
			ret = append(ret, n.SetHelp(b).Build())
		}
		return ret
	default:
		common.Panicf("%v cannot be used as candidates", n)
		return nil
	}
}

func (n *Node) Args() []string {
	ret := make([]string, 0, len(n.args))
	for _, a := range n.args {
		ret = append(ret, a.Word)
	}
	return ret
}

// Return command name listed with @command.
func (n *Node) TargetCommands() []string {
	keys := make([]string, 0, len(n.commandJumpTo))
	for k := range n.commandJumpTo {
		keys = append(keys, k)
	}
	return keys
}
