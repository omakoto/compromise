package adapters

import (
	"bufio"
	"fmt"
	"github.com/omakoto/compromise-go/src/compromise"
	"github.com/omakoto/compromise-go/src/compromise/internal/compmisc"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/fileutils"
	"github.com/omakoto/go-common/src/shell"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

/*
zshAdapter is an interface to Zsh.

See:
http://www.csse.uwa.edu.au/programming/linux/zsh-doc/zsh_23.html
https://linux.die.net/man/1/zshcompsys
http://zsh.sourceforge.net/Guide/zshguide06.html
https://linux.die.net/man/1/zshcompwid (for compadd command)

Note zsh always seems to do variable expansion, so we don't have to do it,
unlike BashAgent.

NOT FULLY TESTED YET.
*/
type zshAdapter struct {
	in  io.Reader
	out *bufio.Writer

	candidates []compromise.Candidate
}

var _ ShellAdapter = ((*zshAdapter)(nil))

func newZshAdapter(rd io.Reader, wr io.Writer) *zshAdapter {
	a := &zshAdapter{in: rd, out: bufio.NewWriter(wr)}

	return a
}

// zshParameters is a template parameter
type zshParameters struct {
	FuncName       string
	ExecutableName string
	CommandNames   []string
	SpecFile       string
}

func (p *zshParameters) Escape(arg string) string {
	return shell.Escape(arg)
}

func (p *zshParameters) Unescape(arg string) string {
	return shell.Unescape(arg)
}

func (a *zshAdapter) Install(targetCommandNames []string, spec string) {
	p := zshParameters{}
	p.FuncName = "__compromise_" + toShellSafeName(targetCommandNames[0]) + "_completion"
	path, err := filepath.Abs(common.MustGetExecutable())
	common.Checkf(err, "Abs failed")
	p.ExecutableName = path
	p.CommandNames = targetCommandNames
	p.SpecFile = saveSpec(targetCommandNames[0], spec)

	tmpl, err := template.New("t").Parse(`
# Install this script into bash with the following command:
# . <({{.Escape .ExecutableName}}
  {{- range $command := .CommandNames}} {{$.Escape $command }}{{end}} )

# Actual completion function.
function {{.FuncName}} {
  . <({{.Escape .ExecutableName}} --` + InvokeOption + ` {{.Escape .SpecFile}} \
      "$(( $CURRENT - 1 ))" "${words[@]}" )
}

{{range $command := .CommandNames -}}
compdef {{$.FuncName}} {{$.Escape $command }}
{{end}}`)

	common.Check(err, "parse failed")
	common.Check(tmpl.Execute(a.out, &p), "execute failed")
}

func (a *zshAdapter) HasMenuCompletion() bool {
	return true
}

func (a *zshAdapter) Escape(arg string) string {
	return shell.Escape(arg)
}

func (a *zshAdapter) Unescape(arg string) string {
	return shell.Unescape(arg)
}

func (a *zshAdapter) GetCommandLine(args []string) *CommandLine {
	cursorIndex, err := strconv.Atoi(args[0])
	common.CheckPanic(err, "Atoi failed") // This is an internal error, so use panic.
	rawWords := args[1:]

	return newCommandLine(a.Unescape, cursorIndex, rawWords)
}

func (a *zshAdapter) StartCompletion(commandLine *CommandLine) {
	// zsh doesn't need it
}

func (a *zshAdapter) MaybeOverrideCandidates(commandLine *CommandLine) []compromise.Candidate {
	return nil // zsh doesn't need it
}

func (a *zshAdapter) AddCandidate(c compromise.Candidate) {
	if len(a.candidates) < compmisc.MaxCandidates {
		a.candidates = append(a.candidates, c)
	}
}

func (a *zshAdapter) printCandidate(c compromise.Candidate) bool {
	// Dump a candidate to stdout.
	val := c.Value()
	if len(val) == 0 {
		return false
	}

	val = a.Escape(val)
	if !c.Continues() {
		val += " "
	}
	if !c.Raw() {
		val = a.Escape(val)
	}

	// -S '' tells zsh not to add a space afterward. (because we do it by ourselves.)
	// -Q prevents zsh from quoting metacharacters in the results, which we do too.
	// -f treats the result as filenames.
	// -U suppress filtering by zsh

	fileopt := ""
	if fileutils.FileExists(c.Value()) {
		fileopt = "-f"
	}

	desc := c.Value()
	if len(c.Help()) > 0 {
		descPadLen := (len(desc)/10 + 1) * 10
		if descPadLen < 20 {
			descPadLen = 20
		}
		desc = desc + strings.Repeat(" ", (descPadLen-len(desc))) + ": " + c.Help()
	}

	a.out.WriteString(fmt.Sprintf(`
		COMPROMISE_D=(%s)
		compadd -S '' -Q -U %s -d COMPROMISE_D -- %s
	`, a.Escape(desc), fileopt, val))

	a.out.WriteByte('\n')
	return true
}

func (a *zshAdapter) EndCompletion() {
	for _, c := range a.candidates {
		a.printCandidate(c)
	}
}

func (a *zshAdapter) Finish() {
	a.out.Flush()
}
