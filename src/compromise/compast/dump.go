package compast

// Dump an AST tree.

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
)

func (n *Node) String() string {
	return n.Dump(false)
}

// Dump returns a human readable dump of a Node tree.
func (n *Node) Dump(multiLine bool) string {
	b := bytes.NewBuffer(nil)
	n.dumpInner("", bufio.NewWriter(b), multiLine).Flush()
	return b.String()
}

func (n *Node) dumpInner(indent string, wr *bufio.Writer, multiLine bool) *bufio.Writer {
	wr.WriteString(indent)
	wr.WriteString(fmt.Sprintf("#%d [%d] ", n.id, n.depth))
	wr.WriteString(nodeTypeNames[n.nodeType])
	wr.WriteString(":")

	dumpField := func(f *Token, name string) {
		if f == nil {
			return
		}
		wr.WriteString(" ")
		wr.WriteString(name)
		wr.WriteString("=")
		wr.WriteString(strconv.Quote(f.Word))
	}
	dumpField(n.literal, "literal")
	dumpField(n.command, "command")
	dumpField(n.pattern, "pattern")
	dumpField(n.funcName, "funcName")
	dumpField(n.label, "label")
	dumpField(n.help, "help")

	if len(n.args) > 0 {
		wr.WriteString(" args=[")
		first := true
		for _, arg := range n.args {
			if arg == nil {
				continue
			}
			if !first {
				wr.WriteString(", ")
			} else {
				first = false
			}
			wr.WriteString(strconv.Quote(arg.Word))
		}
		wr.WriteString("]")
	}

	if multiLine {
		wr.WriteString("\n")
		if n.child != nil {
			n.child.dumpInner(indent+"  ", wr, multiLine)
		}
		if n.next != nil {
			n.next.dumpInner(indent, wr, multiLine)
		}
	}
	return wr
}
