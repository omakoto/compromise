package main

import (
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compfunc"
	"github.com/omakoto/compromise/src/compromise/compmain"
)

func init() {
	compfunc.Register("takeLunch", takeLunch)
}

func takeLunch() compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		var ret []compromise.Candidate
		for _, dev := range compfunc.ReadLinesFromFile(".android-devices") {
			for _, flavor := range []string{"eng", "userdebug"} {
				ret = append(ret, compromise.NewCandidateBuilder().Value(dev+"-"+flavor).Build())
			}
		}
		return ret
	})
}

func main() {
	compmain.Main(spec)
}

var spec = "//" + compromise.NewDirectives().SetSourceLocation().Tab(4).Json() + `
@command lunch
@command a-lunch

@cand takeLunch
`
