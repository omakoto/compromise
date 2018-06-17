package selectors

import "github.com/omakoto/compromise/src/compromise"

type Selector interface {
	Select(string, []compromise.Candidate) (compromise.Candidate, error)
}
