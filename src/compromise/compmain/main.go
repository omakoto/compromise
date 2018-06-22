package compmain

// Contains the entry point.

import (
	"fmt"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compdebug"
	"github.com/omakoto/compromise/src/compromise/internal/adapters"
	"github.com/omakoto/compromise/src/compromise/internal/completer"
	"github.com/omakoto/compromise/src/compromise/internal/compstore"
	"github.com/omakoto/compromise/src/compromise/internal/parser"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/textio"
	"io"
	"io/ioutil"
	"os"
)

func RunWithFatalCatcher(f func()) {
	common.RunAndExitIfFailure(func() int {
		f()
		return 0
	})
}

func PrintInstallScript(spec string, commandsOverride ...string) {
	RunWithFatalCatcher(func() {
		opts := InstallOptions{In: os.Stdin, Out: os.Stdout}
		PrintInstallScriptRaw(spec, opts, commandsOverride...)
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

	opts := InstallOptions{In: os.Stdin, Out: os.Stdout}

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--compromise-list-commands" {
		args = args[1:]
		opts.ListCommandsOnly = true
	}

	PrintInstallScriptRaw(spec, opts, args...)
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

func getTargetCommands(original, override []string) []string {
	if len(override) == 0 {
		return original
	}
	if override[0] != "-" {
		return override
	}
	// Otherwise remove the ones in override.
	remove := make(map[string]string)
	for _, v := range override[1:] {
		remove[v] = v
	}
	ret := make([]string, 0)
	for _, v := range original {
		if _, ok := remove[v]; !ok {
			ret = append(ret, v)
		}
	}
	return ret
}

type InstallOptions struct {
	In               io.Reader
	Out              io.Writer
	ListCommandsOnly bool
}

func PrintInstallScriptRaw(spec string, opts InstallOptions, commandsOverride ...string) {
	runWithSpecCatcher(func() {
		// Parse the spec.
		directives := compromise.ExtractDirectives(spec)
		root := parser.Parse(spec, directives)

		commands := getTargetCommands(root.TargetCommands(), commandsOverride)

		if len(commands) == 0 {
			common.Fatal("spec doesn't contain any @commands; target commands must be passed as arguments")
		}

		if opts.ListCommandsOnly {
			for _, s := range commands {
				textio.BufferedStdout.WriteString(s)
				textio.BufferedStdout.WriteString("\n")
			}
			return
		}

		adapter := adapters.GetShellAdapter(opts.In, opts.Out)
		defer adapter.Finish()

		adapter.Install(commands, spec)
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
	HandleCompletionRaw(func() string {
		return loadFile(os.Args[2])
	}, os.Args[3:], os.Stdin, os.Stdout)
	return
}

func HandleCompletionRaw(specProducer func() string, args []string, in io.Reader, out io.Writer) {
	compdebug.Time("Total", func() {
		runWithSpecCatcher(func() {
			// Prepare shell adapter.
			adapter := adapters.GetShellAdapter(in, out)
			defer adapter.Finish()

			spec := specProducer()
			directives := compromise.ExtractDirectives(spec)

			cl := adapter.GetCommandLine(args)
			compstore.UpdateForInvocation(cl.RawWords(), cl.CursorIndex())

			// Run.
			e := compengine.NewEngine(adapter, cl, directives)
			compdebug.Time("Parse spec", func() {
				e.ParseSpec(spec)
			})
			e.Run()
		})
	})
}

func loadFile(path string) (ret string) {
	compdebug.Time("Load spec file", func() {
		data, err := ioutil.ReadFile(path)
		common.Checkf(err, "unable to read from %s", path)
		ret = string(data)
	})
	return
}
