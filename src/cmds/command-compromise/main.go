package main

// Complete with output from an arbitrary command.

import (
	"flag"
	"fmt"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compfunc"
	"github.com/omakoto/compromise/src/compromise/compmain"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/shell"
	"os"
	"strconv"
)

var (
	takeDir  = flag.Bool("d", false, "Take a directory")
	takeFile = flag.Bool("f", false, "Take a file")
	command  = flag.String("c", "", "Command that generates the candidates. Current token is set to $1.")
)

func init() {
	compfunc.Register("takeCandidateFromCommand", takeCandidateFromCommand)
}

func takeCandidateFromCommand(ctx compromise.CompleteContext, arg string) compromise.CandidateList {
	tok := ctx.WordAtCursor(0)
	return compfunc.BuildCandidateListFromCommand(arg + " " + shell.Escape(tok))
}

func main() {
	compmain.MaybeHandleCompletion()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-f #Also take file] [-d #Also take dirctory] [-c #command that generates candidates] [commands...]\n", common.MustGetBinName())
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()

	spec := "@loop\n"
	spec += "  @switch\n"
	if *takeDir {
		spec += "    @cand takeDir\n"
	}
	if *takeFile {
		spec += "    @cand takeFile\n"
	}
	if *command != "" {
		spec += "    @cand takeCandidateFromCommand "
		spec += strconv.Quote(*command)
		spec += "\n"
	}

	compmain.PrintInstallScript(spec, flag.Args()...)
}
