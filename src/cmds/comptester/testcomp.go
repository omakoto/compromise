package main

import (
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compfunc"
	"github.com/omakoto/compromise/src/compromise/compmain"
	"os"
)

func init() {
	compfunc.Register("takeLazily", takeLazily)
}

func main() {
	compmain.HandleCompletionRaw(os.Args[1], os.Args[2:])
}

func takeLazily() compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		return nil
	})
}
