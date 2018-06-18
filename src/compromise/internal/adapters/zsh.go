package adapters

import (
	"bufio"
	"fmt"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compenv"
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
	SkipZshBind    string

	EvalStr string
}

func (p *zshParameters) Escape(arg string) string {
	return shell.Escape(arg)
}

func (p *zshParameters) Unescape(arg string) string {
	return shell.Unescape(arg)
}

func (a *zshAdapter) Install(targetCommandNames []string, specFile string) {
	p := zshParameters{}
	p.FuncName = getFuncName(targetCommandNames[0])
	path, err := filepath.Abs(common.MustGetExecutable())
	common.Checkf(err, "Abs failed")
	p.ExecutableName = path
	p.CommandNames = targetCommandNames
	p.SpecFile = specFile
	p.SkipZshBind = "0"
	if compenv.ZshSkipBind {
		p.SkipZshBind = "1"
	}

	command := []string{
		shell.Escape(p.ExecutableName),
		"--" + InvokeOption,
		shell.Escape(p.SpecFile),
		`"$(( $CURRENT - 1 ))"`,
		`"${words[@]}"`,
	}
	p.EvalStr = strings.Join(command, " ")

	tmpl, err := template.New("t").Parse(`
# Completion script generated by Compromise (https://github.com/omakoto/compromise)

if ! type compdef >&/dev/null ; then
  echo "compromise: 'compdef' not defined. Please perform minimum Zsh setup first." 1>&2  
else
  # Completion function.
  function {{.FuncName}} {
    eval "$( {{- .EvalStr -}} )"
	
	# Note we want to redraw the current line here, in case we did fzf. But how?
	# Fzf's completion binding solves it by making it a widget and binding [TAB] to it.
	# But if we do that too, that'd czonflict with fzf's.
	# One possible hacky workaround: https://stackoverflow.com/questions/48055589
	#
	# So for now, we just bind Alt+Shift+R to refresh command line.
  }
  
  compdef {{$.FuncName}}{{range $command := .CommandNames}} {{$.Escape $command }}{{end}}
	
  if (( ! {{.SkipZshBind  }} )) ; then
	bindkey '^[R' redisplay # Alt+Shift+R to refresh command line.
  fi

  if [[ "$COMPROMISE_QUIET" != 1 ]] ; then
    echo "Installed completion:"{{- range $command := .CommandNames}} {{$.Escape $command }}{{end}} 1>&2
  fi
fi
`)

	common.Check(err, "parse failed")
	common.Check(tmpl.Execute(a.out, &p), "execute failed")
}

func (a *zshAdapter) HasMenuCompletion() bool {
	return true
}

// Using FZF will make Zsh hang... But you might be able to find out a work around.
func (a *zshAdapter) SupportsFzf() bool {
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
	if len(a.candidates) < compenv.MaxCandidates {
		a.candidates = append(a.candidates, c)
	}
}

func (a *zshAdapter) printCandidate(c compromise.Candidate) bool {
	// Dump a candidate to stdout.
	val := c.Value()
	if len(val) == 0 {
		return false
	}

	val = shell.EscapeNoQuotes(val)
	if !c.Continues() {
		val += " "
	}
	if !c.Raw() {
		val = shell.EscapeNoQuotes(val)
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
