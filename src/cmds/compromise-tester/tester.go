package main

import (
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compfunc"
	"github.com/omakoto/compromise/src/compromise/compmain"
)

func init() {
	compfunc.Register("takeOne", takeOne)
}

func takeOne() compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		ret := make([]compromise.Candidate, 0)

		ret = append(ret, compromise.NewCandidate().SetValue("cooked").SetRaw(false))
		ret = append(ret, compromise.NewCandidate().SetValue("raw").SetRaw(true))

		ret = append(ret, compromise.NewCandidate().SetValue("cooked#test").SetRaw(false))
		ret = append(ret, compromise.NewCandidate().SetValue("raw#test").SetRaw(true))

		return ret
	})
}

func main() {
	compmain.Main(spec)
}

var spec = "//" + compromise.NewDirectives().SetSourceLocation().Tab(4).JSON() + `
@command tester

@loop
	@cand takeOne
	@cand takeFile
`
