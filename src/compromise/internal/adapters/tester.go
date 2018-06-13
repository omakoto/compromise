package adapters

import (
	"bufio"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/go-common/src/shell"
	"io"
	"sort"
)

type testerAdapter struct {
	in  io.Reader
	out *bufio.Writer

	candidates []compromise.Candidate
}

var _ ShellAdapter = ((*testerAdapter)(nil))

func newTesterAdapter(rd io.Reader, wr io.Writer) *testerAdapter {
	a := &testerAdapter{in: rd, out: bufio.NewWriter(wr)}

	return a
}

func (a *testerAdapter) Install(targetCommandNames []string, specFile string) {
	a.out.WriteString(specFile)
	a.out.WriteString("\n")
}

func (a *testerAdapter) HasMenuCompletion() bool {
	return false
}

func (a *testerAdapter) Escape(arg string) string {
	return shell.Escape(arg)
}

func (a *testerAdapter) Unescape(arg string) string {
	return shell.Unescape(arg)
}

func (a *testerAdapter) GetCommandLine(args []string) *CommandLine {
	// Ignore the index

	args = args[1:]

	return newCommandLine(a.Unescape, len(args)-1, args)
}

func (a *testerAdapter) StartCompletion(commandLine *CommandLine) {
}

func (a *testerAdapter) MaybeOverrideCandidates(commandLine *CommandLine) []compromise.Candidate {
	return nil
}

func (a *testerAdapter) AddCandidate(candidate compromise.Candidate) {
	a.candidates = append(a.candidates, candidate)
}

func (a *testerAdapter) EndCompletion() {
	sort.Slice(a.candidates, func(i, j int) bool {
		return a.candidates[i].Value() < a.candidates[j].Value()
	})
	for _, v := range a.candidates {
		a.out.WriteString(v.Value())
		if v.Continues() {
			a.out.WriteString("+")
		}
		a.out.WriteString("\n")
	}
}

func (a *testerAdapter) Finish() {
	a.out.Flush()
}
