package compmain

// Contains the entry point.

import (
	"fmt"
	"github.com/omakoto/compromise-go/src/compromise"
	"github.com/omakoto/compromise-go/src/compromise/internal/adapters"
	"github.com/omakoto/compromise-go/src/compromise/internal/compdebug"
	"github.com/omakoto/compromise-go/src/compromise/internal/completer"
	"github.com/omakoto/compromise-go/src/compromise/internal/compmisc"
	"github.com/omakoto/compromise-go/src/compromise/internal/compstore"
	"github.com/omakoto/compromise-go/src/compromise/internal/parser"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/textio"
	"io/ioutil"
	"os"
	"strings"
)

func RunWithFatalCatcher(f func()) {
	common.RunAndExitIfFailure(func() int {
		f()
		return 0
	})
}

func PrintInstallScript(spec string, commandsOverride ...string) {
	RunWithFatalCatcher(func() {
		PrintInstallScriptRaw(spec, false, commandsOverride...)
	})
}

func MaybeHandleCompletion() {
	ret := false
	RunWithFatalCatcher(func() {
		ret = MaybeHandleCompletionRaw()
	})
	if ret {
		os.Exit(0)
	}
	return
}

// Call this from main() with a spec.
func Main(spec string) {
	RunWithFatalCatcher(func() {
		MainRaw(spec)
	})
}

func MainRaw(spec string) {
	if MaybeHandleCompletionRaw() {
		common.ExitSuccess()
	}

	listOnly := false
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--compromise-list-commands" {
		args = args[1:]
		listOnly = true
	}

	PrintInstallScriptRaw(spec, listOnly, args...)
}

func runWithSpecCatcher(f func()) {
	// Detect a SpecError panic and convert it to an error
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(compromise.SpecError); ok {
				msg := ""
				msg += fmt.Sprintf("invalid spec: %s", e.Message)
				if e.Location != nil {
					file, line, column := e.Location.SourceLocation()
					msg += fmt.Sprintf(" at %s:%d:%d", file, line, column)
				}
				common.Fatal(msg)
			} else {
				panic(r)
			}
		}
	}()
	f()
}

func PrintInstallScriptRaw(spec string, listCommandsOnly bool, commandsOverride ...string) {
	runWithSpecCatcher(func() {
		// Parse the spec.
		directives := compromise.ExtractDirectives(spec)
		root := parser.Parse(spec, directives)
		commands := root.TargetCommands()
		if len(commandsOverride) > 0 {
			commands = commandsOverride
		}

		if len(commands) == 0 {
			common.Fatal("spec doesn't contain any @commands; target commands must be passed as arguments")
		}

		if listCommandsOnly {
			for _, s := range commands {
				textio.BufferedStdout.WriteString(s)
				textio.BufferedStdout.WriteString("\n")
			}
			return
		}

		adapter := adapters.GetShellAdapter(os.Stdin, os.Stdout)
		defer adapter.Finish()

		adapter.Install(commands, spec)

		if compmisc.Verbose {
			fmt.Fprintf(os.Stderr, "Installed completion for %s\n", strings.Join(commands, " "))
		}
	})
}

func MaybeHandleCompletionRaw() (ret bool) {
	if len(os.Args) < 2 || os.Args[1] != "--"+adapters.InvokeOption {
		return false
	}
	ret = true
	if len(os.Args) < 5 {
		common.Fatalf("not enough arguments. %d given", len(os.Args))
	}
	compdebug.Time("total", func() {
		runWithSpecCatcher(func() {
			// Read spec.
			spec := loadFile(os.Args[2])

			// Prepare shell adapter.
			adapter := adapters.GetShellAdapter(os.Stdin, os.Stdout)
			defer adapter.Finish()

			directives := compromise.ExtractDirectives(spec)

			cl := adapter.GetCommandLine(os.Args[3:])
			compstore.UpdateForInvocation(cl.RawWords(), cl.CursorIndex())

			// Run.
			e := compengine.NewEngine(adapter, cl, directives)
			compdebug.Time("parse", func() {
				e.ParseSpec(spec)
			})
			e.Run()
		})
	})
	return
}

func loadFile(path string) (ret string) {
	compdebug.Time("load", func() {
		data, err := ioutil.ReadFile(path)
		common.Checkf(err, "unable to read from %s", path)
		ret = string(data)
	})
	return
}
