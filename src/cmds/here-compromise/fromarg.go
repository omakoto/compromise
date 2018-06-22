package main

// Here compromise takes a spec from argument.

import (
	"fmt"
	"github.com/omakoto/compromise/src/compromise/compmain"
	"github.com/omakoto/go-common/src/common"
	"os"
)

func main() {
	compmain.MaybeHandleCompletion()
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s SPEC [commands...]\n", common.MustGetBinName())
		os.Exit(1)
	}
	compmain.PrintInstallScript(os.Args[1], os.Args[2:]...)
}
