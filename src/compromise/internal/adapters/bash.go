package adapters

import (
	"bufio"
	"bytes"
	"github.com/davecgh/go-spew/spew"
	"github.com/mattn/go-isatty"
	"github.com/omakoto/compromise-go/src/compromise"
	"github.com/omakoto/compromise-go/src/compromise/compfunc"
	"github.com/omakoto/compromise-go/src/compromise/internal/compdebug"
	"github.com/omakoto/compromise-go/src/compromise/internal/compmisc"
	"github.com/omakoto/compromise-go/src/compromise/internal/compstore"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/fileutils"
	"github.com/omakoto/go-common/src/shell"
	"github.com/omakoto/go-common/src/utils"
	"github.com/ungerik/go-dry"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

const bashSectionSeparator = "\n-*-*-*-COMPROMISE-*-*-*-\n"

// bashVariables contains all (except arrays and hashes) shell/environmental variables from bash.
type bashVariables struct {
	name  string
	value string
}

// bashAdapter is the adapter between bash and compromise.
type bashAdapter struct {
	in  io.Reader
	out *bufio.Writer

	candidates []compromise.Candidate

	variables map[string]string

	// The "delta" of the cursor word from what readline think the current word is.
	bashDeltaFromReadline string
}

var _ ShellAdapter = ((*bashAdapter)(nil))

func newBashAdapter(rd io.Reader, wr io.Writer) *bashAdapter {
	a := &bashAdapter{in: rd, out: bufio.NewWriter(wr)}

	return a
}

// bashParameters is a template parameter
type bashParameters struct {
	FuncName             string
	ExecutableName       string
	SectionSeparator     string
	CommandNames         []string
	SkipBashBind         string
	SpecFile             string
	CompletionIgnoreCase string
	CompletionMapCase    string
}

func (p *bashParameters) Escape(arg string) string {
	return shell.Escape(arg)
}

func (p *bashParameters) Unescape(arg string) string {
	return shell.Unescape(arg)
}

func (a *bashAdapter) Install(targetCommandNames []string, spec string) {
	p := bashParameters{}
	p.FuncName = "__compromise_" + toShellSafeName(targetCommandNames[0]) + "_completion"
	path, err := filepath.Abs(common.MustGetExecutable())
	common.Checkf(err, "Abs failed")
	p.ExecutableName = path
	p.SectionSeparator = bashSectionSeparator
	p.CommandNames = targetCommandNames
	p.SkipBashBind = "0"
	if compmisc.BashSkipBind {
		p.SkipBashBind = "1"
	}
	p.CompletionIgnoreCase = "off"
	p.CompletionMapCase = "off"
	if compmisc.IgnoreCase {
		p.CompletionIgnoreCase = "on"
	}
	if compmisc.MapCase {
		p.CompletionMapCase = "on"
	}

	p.SpecFile = saveSpec(targetCommandNames[0], spec)

	tmpl, err := template.New("t").Parse(`
# Install this script into bash with the following command:
# . <({{.Escape .ExecutableName}}
  {{- range $command := .CommandNames}} {{$.Escape $command }}{{end}} )

# Always redraw the current line after completion.
# This also helps to catch double TAB presses.
if (( ! {{.SkipBashBind  }} )) ; then
  # bind '"\e:1": overwrite-mode' # Not used
  bind '"\e:2": complete'
  bind '"\e:3": redraw-current-line'
  bind '"\C-i": "\e:2\e:3"'
  bind 'set completion-ignore-case {{.CompletionIgnoreCase}}'
  bind 'set completion-map-case {{.CompletionMapCase}}'
  bind 'set show-all-if-ambiguous on'
  bind 'set show-all-if-unmodified on'
  bind 'set skip-completed-text on'
fi
	
# This feeds information within shell (e.g. shell variables)
# to compromise.
function __compromise_context_dumper {
  declare -p
  echo -n "{{.SectionSeparator}}"
  jobs
}

# Actual completion function.
function {{.FuncName}} {
  export COMP_POINT
  export COMP_LINE
  export COMP_TYPE
  export COMP_WORDBREAKS
  . <( __compromise_context_dumper |
      {{.Escape .ExecutableName}} --` + InvokeOption + ` {{.Escape .SpecFile}} \
	      "$COMP_CWORD" "${COMP_WORDS[@]}" 
  )
}

{{range $command := .CommandNames -}}
complete -o nospace -F {{$.FuncName}} -- {{$.Escape $command }}
{{end}}`)

	common.Check(err, "parse failed")
	common.Check(tmpl.Execute(a.out, &p), "execute failed")
}

// HasMenuCompletion returns false for bash.
func (a *bashAdapter) HasMenuCompletion() bool {
	return false
}

// Escape is a shell escape function for bash.
func (a *bashAdapter) Escape(arg string) string {
	return shell.Escape(arg)
}

// Escape is a shell unescape function for bash.
func (a *bashAdapter) Unescape(arg string) string {
	return shell.Unescape(utils.HomeExpanded(arg))
}

func (a *bashAdapter) GetCommandLine(args []string) *CommandLine {
	cword, err := strconv.Atoi(args[0])
	common.CheckPanic(err, "Atoi failed") // This is an internal error, so use panic.
	compWords := args[1:]

	ret := newCommandLine(a.Unescape, cword, compWords)

	ret.bashCompCword = cword
	ret.bashCompWords = compWords
	if cword <= len(compWords) {
		ret.bashCompCurrentWord = compWords[cword]
	}
	ret.bashCompPoint = int(utils.ParseInt(os.Getenv("COMP_POINT"), 10, -1))
	ret.bashCompLine = os.Getenv("COMP_LINE")
	ret.bashCompType = os.Getenv("COMP_TYPE")
	ret.bashCompWordbreaks = os.Getenv("COMP_WORDBREAKS")
	ret.bashParsedRawWords = shell.SplitToTokens(ret.bashCompLine)

	// For compatibility with zsh, and also for simplicity, we ignore $COMP_CWORD and $COMP_WORDS
	// and just split COMP_LINE by ourselves.
	// bashParsedRawWords already has split up tokens, so let's find the cursor index for
	// bashCompPoint.
	// However COMP_WORDBREAKS makes it hard... See below.
	lastIndex := 0
	parsedRawWords := make([]string, 0)
	var lastToken shell.Token
	for i, t := range ret.bashParsedRawWords {
		parsedRawWords = append(parsedRawWords, t.Word)
		if t.Index > ret.bashCompPoint {
			break
		}
		lastIndex = i
		lastToken = t
	}
	// If the cursor is after the last word, advance the index.
	if ret.bashCompPoint > lastToken.Index+len(lastToken.Word) {
		lastIndex++
	}

	ret.Replace(lastIndex, parsedRawWords)

	// Work around for COMP_WORDBREAKS.
	// See E13 on https://tiswww.case.edu/php/chet/bash/FAQ
	// Even though we ignore COMP_WORDBREAKS and get the current token by ourselves,
	// readline will *still* use the COMP_WORDBREAKS rule to replace the current word.
	//
	// * example1: echo a:b:c[TAB]
	// -> Only "c" is the current word by readline's view, so completion will be only performed on it.
	// So if we return "a:b:cXYZ", then readline will generate "a:b:a:b:cXYZ" because it only
	// replaces "c".
	//
	// * example2: echo a:b:[TAB]
	// -> Only ":" is the current word by readline's view, so completion will be only performed on it.
	// However if the current word is a break word, readline *will not* replace it, but only append.
	// So if we return "a:b:XYZ", then readline will generate "a:b:a:b:XYZ" because it *appends*
	// to ":". (not *replace* unlike above.)
	//
	// So, when we return candidates, we need to replace the "delta" part.
	// In example 1, we should only return "cXYZ", and in example 2, we should only return "XYZ"
	//
	// HOWEVER...
	// This means "force" candidates won't work.
	// Normally, it's possible to complete "filena" to "Filename", because readline will always
	// use a single candidate no matter what, even if it doesn't match the actual word.
	// However, we can't replace "filename:filena" with "Filename:Filename" because we can't replace
	// the "filename:" part because that's not the completion target.
	a.bashDeltaFromReadline = findDeltaFromReadline(ret.RawWordAtCursor(0), ret.bashCompCurrentWord, ret.bashCompWordbreaks, ret)

	return ret
}

func findDeltaFromReadline(ours, theirs, wordbreaks string, commandLineForLog *CommandLine) string {
	// If readline's current word is a break word (e.g. ":" or "::" or whatever...)
	// the whole "our" word needs to be stripped off from the final candidates.
	if len(theirs) > 0 && strings.IndexByte(wordbreaks, theirs[0]) >= 0 {
		return ours
	}
	// Otherwise, if theirs is a suffix of ours, then the leading part will be the delta.
	if strings.HasSuffix(ours, theirs) {
		return ours[0 : len(ours)-len(theirs)]
	}
	compdebug.Warnf("bashagent: Failed to find delta from readline: ours=%q theirs=%q\nCommandline=%s", ours, theirs, spew.Sdump(commandLineForLog))
	return ""
}

func (a *bashAdapter) StartCompletion(commandLine *CommandLine) {

	a.parseContext()

	// TODO Fix up command line? (we used to do in the ruby version...)

	if stdin, ok := a.in.(*os.File); !ok || !isatty.IsTerminal(stdin.Fd()) {
		all, err := ioutil.ReadAll(a.in)
		common.Checkf(err, "ReadAll(stdin) failed")
		dry.Nop(all) // TODO Receive from __compromise_context_passer
		/* Sample:
		   declare -x rvm_wrapper_name
		   declare -- script="override_gem"
		   declare -a chpwd_functions=([0]="__rvm_cd_functions_set")
		   declare -- __git_mergetools_common="diffuse diffmerge ecmerge emerge kdiff3 meld opendiff
		   tkdiff vimdiff gvimdiff xxdiff araxis p4merge bc3 codecompare
		   "
		   declare -A _xspecs=([freeamp]="!*.@(mp3|og[ag]|pls|m3u)" [bibtex]="!*.aux")
		   ***MARKER***
		   [1]   Stopped                 cat
		   [2]-  Stopped                 cat
		   [3]+  Stopped                 cat > /dev/null
		*/
	}
	a.out.WriteString(`local IFS=$'\\n'; COMPREPLY=(`)
	a.out.WriteByte('\n')
}

func (a *bashAdapter) MaybeOverrideCandidates(commandLine *CommandLine) []compromise.Candidate {
	// 1. First, if the current token begins with $, then do a variable name expansion.
	raw := commandLine.RawWordAtCursor(0)
	if strings.HasPrefix(raw, "$") {
		compdebug.Debugf("Variable expansion: raw word=%q\n", raw)
		varPat := regexp.MustCompile(`^.([_a-zA-Z0-9]*)(/?)`)
		m := varPat.FindStringSubmatch(raw)
		if len(m) == 3 {
			ret := make([]compromise.Candidate, 0)
			if len(m[2]) == 0 {
				// $NAM[TAB]
				// Do a variable name expansion. e.g. $PAT -> $PATH
				for key, _ := range a.variables {
					if compromise.StringMatches(key, m[1]) {
						ret = append(ret, compromise.NewCandidateBuilder().Value("$"+key).Continues(true).Force(true).Build())
					}
				}
			} else {
				// $NAME/[TAB] or
				// $NAME/....[TAB] -> If the variable contains a directory name, expand it.
				name := m[1]
				if val, ok := a.variables[name]; ok && fileutils.DirExists(val) {
					rest := shell.Unescape(raw[len(m[0]):])
					c := compromise.NewCandidateBuilder().Value(val + "/" + rest).Continues(true).Force(true).Build()
					ret = append(ret, c)
				}
			}
			if len(ret) > 0 {
				return ret
			}
		}
	}

	// 2. In bash, completion will be triggered even after a redirect operator. Detect it and switch
	// to filename completion.
	afterRedirect := false
	for i := 1; i <= commandLine.cursorIndex; i++ {
		if strings.IndexByte("<>", utils.StringByteAt(commandLine.RawWordAt(i), 0)) >= 0 {
			afterRedirect = true
			break
		}
	}
	if !afterRedirect {
		return nil
	}
	compdebug.Debugf("  Switching to file complete\n")
	return compfunc.TakeFile("").GetCandidate(commandLine.WordAtCursor(0))
}

func (a *bashAdapter) AddCandidate(c compromise.Candidate) {
	a.candidates = append(a.candidates, c)
}

func (a *bashAdapter) cutDeltaFromReadline(cand string) string {
	dlen := len(a.bashDeltaFromReadline)
	if dlen == 0 {
		return cand // return as is.
	}
	if strings.HasPrefix(cand, a.bashDeltaFromReadline) {
		return cand[dlen:]
	}
	// See GetCommandLine(). In this case... We can't replace the "delta" part,
	// so just throw away this candidate. TODO Is this really okay?
	compdebug.Warnf("Throwing away candidate %q (delta=%q)\n", cand, a.bashDeltaFromReadline)
	return ""
}

func (a *bashAdapter) printCandidate(c compromise.Candidate) bool {
	// Dump a candidate to stdout.
	val := c.Value()
	val = a.cutDeltaFromReadline(val)
	if len(val) == 0 {
		return false
	}

	if !c.Continues() {
		val += " "
	}
	// Output will be eval'ed, so need double-escaping unless it's raw.
	if c.Raw() {
		a.out.WriteString(val)
	} else {
		a.out.WriteString(a.Escape(val))
	}
	a.out.WriteByte('\n')
	return true
}

func (a *bashAdapter) EndCompletion() {
	sort.Slice(a.candidates, func(i, j int) bool {
		return a.candidates[i].Value() < a.candidates[j].Value()
	})

	// Show candidates on stdout for eval by bash.

	candCount := 0
	for _, c := range a.candidates {
		if a.printCandidate(c) {
			candCount++
			if candCount >= compmisc.MaxCandidates {
				break
			}
		}
	}

	store := compstore.Load()

	// Show help on stderr
	if candCount > 1 || candCount == 0 {
		// First, show help, for at most
		buf := bytes.NewBuffer(nil)

		helpCount := 0
		for _, c := range a.candidates {
			if !c.NeedsHelp() || len(c.Help()) == 0 {
				continue
			}
			helpCount++

			buf.WriteString("  ")
			if len(c.Value()) > 0 {
				buf.WriteString(c.Value())
			} else {
				buf.WriteString("<ANY>")
			}
			if len(c.Help()) > 0 {
				if compmisc.UseColor {
					buf.WriteString("\x1b[32m")
				}
				buf.WriteString(" : ")
				buf.WriteString(c.Help())
				if compmisc.UseColor {
					buf.WriteString("\x1b[0m")
				}
			}
			buf.WriteString("\n")

			if !store.IsDoublePress && helpCount >= compmisc.BashHelpMaxCandidates {
				if compmisc.UseColor {
					buf.WriteString("\x1b[34m")
				}
				buf.WriteString("  [Result omitted; hit tab twice to show all]")
				if compmisc.UseColor {
					buf.WriteString("\x1b[0m")
				}
				buf.WriteString("\n")
				break
			}
		}

		content := buf.Bytes()
		if len(content) > 0 {
			os.Stderr.WriteString("\n")
			os.Stderr.Write(content)
		}
	}

	a.out.WriteString(`) # End of COMPREPLY`)
	a.out.WriteByte('\n')
}

func (a *bashAdapter) Finish() {
	a.out.Flush()
}

// parseContext parses the content passed by __compromise_context_dumper that contains shell variables, etc.
func (a *bashAdapter) parseContext() {
	bytes, err := ioutil.ReadAll(os.Stdin)
	common.Check(err, "cannot read from stdin")

	str := string(bytes)
	split := strings.Split(str, bashSectionSeparator)
	if len(split) < 2 {
		compdebug.Warnf("stdin content=%q\n", spew.Sdump(str))
		panic("unable to decode stdin")
	}

	a.variables = make(map[string]string)

	tokens := shell.Split(split[0])
	i := 0
	for i+2 < len(tokens) {
		decl := tokens[i]
		if decl != "declare" {
			compdebug.Warnf("Unable to parse variables\n", str)
			break
		}
		flags := tokens[i+1]
		if strings.ContainsAny(flags, "aA") {
			i += 2
			if tokens[i] == "()" {
				i++
				continue
			}
			// It's an array or a hash.  Skip all until ")"
			for i < len(tokens) {
				t := tokens[i]
				i++
				if t == ")" {
					break
				}
			}
			continue
		}
		nameAndVal := strings.SplitN(tokens[i+2], "=", 2)
		val := ""
		if len(nameAndVal) == 2 {
			val = nameAndVal[1]
		}

		a.variables[nameAndVal[0]] = shell.Unescape(val)

		i += 3
	}
	compdebug.Dump("Variables=", a.variables)

	// TODO Parse jobs

}
