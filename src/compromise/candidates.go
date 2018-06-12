package compromise

type CandidateListGenerator func(ctx CompleteContext, args []string) CandidateList

func filter(candidates []Candidate, prefix string) []Candidate {
	ret := make([]Candidate, 0)

	for _, c := range candidates {
		if c.Matches(prefix) {
			ret = append(ret, c)
		}
	}
	return ret
}

// CandidateList represents a list of Candidate's.
type CandidateList interface {
	GetCandidate(prefix string) []Candidate

	Matches(word string) bool
}

// OpenCandidates generates an "open" CandidateList from a given list of Candidate's.
// it's "open" because candidates are considered to be non-exhaustive and any strings are
// considered to be potential matches.
func OpenCandidates(candidates ...Candidate) CandidateList {
	return &staticCandidates{candidates, false}
}

// StrictCandidates generates an "strict" CandidateList from a given list of Candidate's.
// it's "strict" because candidates are considered to be exhaustive and other strings aren't
// considered to be potential matches.
func StrictCandidates(candidates ...Candidate) CandidateList {
	return &staticCandidates{candidates, true}
}

// LazyCandidates generates a CandidateList from a given list of Candidate's.
func LazyCandidates(generator func(prefix string) []Candidate) CandidateList {
	return &lazyCandidates{generator}
}

type staticCandidates struct {
	candidates []Candidate
	strict     bool
}

var _ CandidateList = (*staticCandidates)(nil)

func (s *staticCandidates) GetCandidate(prefix string) []Candidate {
	return filter(s.candidates, prefix)
}

func (s *staticCandidates) Matches(word string) bool {
	if !s.strict {
		return true
	}
	for _, s := range s.candidates {
		if s.Matches(word) {
			return true
		}
	}
	return false
}

type lazyCandidates struct {
	generator func(prefix string) []Candidate
}

var _ CandidateList = (*lazyCandidates)(nil)

func (s *lazyCandidates) GetCandidate(prefix string) []Candidate {
	return filter(s.generator(prefix), prefix)
}

func (s *lazyCandidates) Matches(word string) bool {
	// Lazy candidates always assumes any non-empty string matches one of the candidates.
	return len(word) > 0
}
