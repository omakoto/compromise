package parser

import (
	"github.com/omakoto/compromise-go/src/compromise"
	"github.com/omakoto/compromise-go/src/compromise/compast"
	"github.com/omakoto/compromise-go/src/compromise/internal/parser/tokenizer"
	"github.com/omakoto/go-common/src/common"
)

type parser struct {
	// Source string.
	source string

	// Last token, used for error message.
	lastToken *compast.Token

	directives *compromise.Directives
}

func Parse(spec string, d *compromise.Directives) *compast.Node {
	p := &parser{source: spec, directives: d}

	return p.parse()
}

func (p *parser) parse() *compast.Node {
	t := tokenizer.NewTokenizer(p.source, p.directives)

	lastLineStartColumn := 0
	columns := make([]int, 0) // Used to detect to go back indents.
	depth := 0

	root := compast.NewRoot()
	nodeStack := make([]*compast.Node, 0)
	nodeStack = append(nodeStack, root)

	// Loop over tokens...
	for {
		tok := t.NextToken()
		p.lastToken = tok

		if tok == nil {
			break // EOF
		}

		if tok.IndexInLine != 0 {
			panic(compromise.NewSpecErrorf(tok, "Unexpected token: %s", tok))
		}

		// Always beginning of a line when we're here.
		if tok.Column == lastLineStartColumn {
			// same depth
		} else if tok.Column > lastLineStartColumn {
			// child
			columns = append(columns, lastLineStartColumn)
			common.Debugf("Indent increased from %d to %d", lastLineStartColumn, tok.Column)
			depth++

			if len(nodeStack) <= depth {
				nodeStack = append(nodeStack, nil)
			}
		} else {
			// go up
			if len(columns) == 0 {
				panic("depth can't be 0")
			}
			common.Debugf("Indent decreased from %d to %d", lastLineStartColumn, tok.Column)

			// Go up to the right depth.
			for {
				prevColumn := columns[len(columns)-1]
				columns = columns[:len(columns)-1]
				if tok.Column > prevColumn {
					panic(compromise.NewSpecErrorf(tok, "inconsistent indent for token \"%s\", expected column is %d", tok.RawWord, prevColumn))
				}
				depth--
				if tok.Column == prevColumn {
					break
				}
			}
		}
		lastLineStartColumn = tok.Column

		common.Debugf("%3d> (%2d,%2d) [%d] [%s] [%s]\n", depth, tok.Line, tok.Column, tok.TokenType, tok.Word, tok.RawWord)

		var n *compast.Node

		//newNode := func(nodeType int, args ...*compast.Token) *compast.Node {
		//	return &compast.Node{nodeType: nodeType, Token: tok, Args: args}
		//}

		switch tok.TokenType {
		case compast.TokenCommand:
			switch tok.Word {
			case "command":
				if depth != 1 {
					panic(compromise.NewSpecError(tok, "@command must be at the toplevel"))
				}

				const err = "@command must be followed by a command name (any string) and optionally a label name (:name)"
				targetCommand := t.MustGetNextTokenInLine(compast.TokenLiteral, err)
				label := t.MaybeGetLabel()

				common.Debugf("* Directive: command %s then jump to %s", targetCommand.Word, label)

				n = compast.NewCommand(tok, targetCommand, label)

			case "label":
				if depth != 1 {
					panic(compromise.NewSpecError(tok, "@label must be at the toplevel"))
				}

				const err = "@label must be followed by a label name (:name)"
				label := t.MustGetNextTokenInLine(compast.TokenLabel, err)
				common.Debugf("* Label %s", label.Word)

				n = compast.NewLabel(tok, label)

			//case "jump":
			//	const err = "@jump must be followed by a label name (:name)"
			//	label := t.MustGetNextTokenInLine(compast.TokenLabel, err)
			//	common.Debugf("* Jump to %s", label.Word)
			//
			//	n = compast.NewJump(tok, label)

			case "call":
				const err = "@call must be followed by a label name (:name)"
				label := t.MustGetNextTokenInLine(compast.TokenLabel, err)
				common.Debugf("* Jump to %s", label.Word)

				n = compast.NewCall(tok, label)

			case "finish":
				common.Debugf("* Action: finish")

				n = compast.NewFinish(tok)

			case "loop":
				common.Debugf("* Action: loop")

				pattern, label := t.MaybeGetLiteralAndLabel()

				n = compast.NewLoop(tok, pattern, label)

			case "switch":
				common.Debugf("* Action: switch")

				pattern, label := t.MaybeGetLiteralAndLabel()

				n = compast.NewSwitch(tok, pattern, label)

			case "switchloop":
				common.Debugf("* Action: switch-loop")

				pattern, label := t.MaybeGetLiteralAndLabel()

				n = compast.NewSwitchLoop(tok, pattern, label)

			case "break":
				common.Debugf("* Action: break")

				label := t.MaybeGetLabel()

				n = compast.NewBreak(tok, label)

			case "continue":
				common.Debugf("* Action: continue")

				label := t.MaybeGetLabel()

				n = compast.NewContinue(tok, label)

			case "any":
				help := t.MaybeGetHelpToken()
				common.Debugf("* Candidate: any")

				n = compast.NewAny(tok, help)

			case "go_call":
				const err = "@go_call must be followed by a function name"
				funcName := t.MustGetNextTokenInLine(compast.TokenLiteral, err)
				args := t.MaybeGetArgsAndHelpToken()

				n = compast.NewGoCall(tok, funcName, args)

				common.Debugf("* Action: go_call to %s", funcName)
			case "cand":
				const err = "@cand must be followed by a function name"
				funcName := t.MustGetNextTokenInLine(compast.TokenLiteral, err)
				args := t.MaybeGetArgsAndHelpToken()

				n = compast.NewCandidate(tok, funcName, args)
				common.Debugf("* Action: candidates from %s", funcName)
			default:
				panic(compromise.NewSpecErrorf(tok, "unexpected command %s", tok))
			}
		case compast.TokenLiteral:
			help := t.MaybeGetHelpToken()

			common.Debugf("* Literal: %s with help %s", tok, help)

			n = compast.NewLiteral(tok, help)
		default:
			panic(compromise.NewSpecErrorf(tok, "Unexpected token: %s", tok))
		}

		nodeStack[depth-1].AddChild(n)
		nodeStack[depth] = n
	}
	p.ensureLabelsExist(root)
	return root
}

func (p *parser) ensureLabelsExist(n *compast.Node)  {
	if n == nil {
		return
	}
	if n.Label() != nil {
		if n.Root().GetLabeledNode(n.LabelWord(), n.SelfToken()) == nil {
			panic(compromise.NewSpecErrorf(n.SelfToken(), "label %s doesn't exist", n.LabelWord()))
		}
	}
	for c := n.Child() ; c != nil; c = c.Next() {
		p.ensureLabelsExist(c)
	}
}