package main

import (
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compenv"
	"github.com/omakoto/compromise/src/compromise/compfunc"
	"github.com/omakoto/compromise/src/compromise/compmain"
	"path/filepath"
)

func init() {
	compfunc.Register("takeLunch", takeLunch)
}

func takeLunch() compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		var ret []compromise.Candidate
		for _, dev := range compfunc.ReadLinesFromFile(filepath.Join(compenv.Home, ".android-devices")) {
			for _, flavor := range []string{"eng", "userdebug"} {
				ret = append(ret, compromise.NewCandidate().SetValue(dev+"-"+flavor))
			}
		}
		return ret
	})
}

func main() {
	compmain.Main(spec)
}

var spec = "//" + compromise.NewDirectives().SetSourceLocation().Tab(4).JSON() + `
@command lunch
@command a-lunch

@cand takeLunch
`
