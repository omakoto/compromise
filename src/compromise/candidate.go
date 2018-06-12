package compromise

import (
	"fmt"
)

// Candidate represents a single candidate
type Candidate interface {
	CandidateList

	// Value returns a text to be completed. e.g. "-f", "--file", "filename.txt"
	Value() string

	// Raw returns whether it's a "raw" candidate -- e.g. whether it needs not to be escaped.
	// e.g. if a candidate value is `$HOME`, and if it's not a raw candidate, then `\$HOME` will be
	// inserted to command line. On the other hand, if it's raw, `$HOME` will be inserted.
	Raw() bool

	// Hidden returns whether it's a "hidden" candidate. TODO Explain
	Hidden() bool

	// Continues returns whether a candidate needs not to be followed by a space when inserted.
	Continues() bool

	// Force returns whether a candidate always needs to be included in the completion list
	// even if the current token is not a prefix of this candidate.
	Force() bool

	// Help returns a help string for a candidate.
	Help() string

	// Whether it needs to be shown in the help section or not
	// TODO Do we really need it?
	NeedsHelp() bool
}

type candidate struct {
	value     string
	raw       bool
	hidden    bool
	continues bool
	force     bool
	help      string
	needsHelp bool
}

var _ Candidate = (*candidate)(nil)
var _ CandidateList = (*candidate)(nil)
var _ fmt.Stringer = (*candidate)(nil)

func (c *candidate) String() string {
	return fmt.Sprintf("value=%q, raw=%v, continues=%v, hidden=%v, force=%v, help=%q", c.value, c.raw, c.continues, c.hidden, c.force, c.help)
}

func (c *candidate) Value() string {
	return c.value
}

func (c *candidate) Raw() bool {
	return c.raw
}

func (c *candidate) Hidden() bool {
	return c.hidden
}

func (c *candidate) Continues() bool {
	return c.continues
}

func (c *candidate) Force() bool {
	return c.force
}

func (c *candidate) Help() string {
	return c.help
}

func (c *candidate) NeedsHelp() bool {
	return c.needsHelp
}

func (c *candidate) Matches(prefix string) bool {
	return c.Force() || StringMatches(c.value, prefix)
}

func (c *candidate) GetCandidate(prefix string) []Candidate {
	if c.Matches(prefix) {
		return []Candidate{c}
	}
	return nil
}

//var reCandidateParser = regexp.MustCompile(`^([rhcf]*:)?\s*(\S+)(?:\s+#\s*(.*))?$`)
//
//func C(spec string) Candidate {
//	spec = strings.Trim(spec, " ")
//	if len(spec) == 0 {
//		return nil
//	}
//	subs := reCandidateParser.FindStringSubmatch(spec)
//	if subs == nil {
//		panic(fmt.Sprintf("Unable to parse spec \"%v\"", spec))
//	}
//
//	raw := strings.Contains(subs[1], "r")
//	hidden := strings.Contains(subs[1], "h")
//	continues := strings.Contains(subs[1], "c")
//	force := strings.Contains(subs[1], "f")
//
//	value := subs[2]
//	help := subs[3]
//
//	if len(value) == 0 {
//		panic(fmt.Sprintf("Spec \"%v\" doesn't contain value", spec))
//	}
//	return &candidate{
//		value:     value,
//		raw:       raw,
//		hidden:    hidden,
//		continues: continues,
//		force:     force,
//		help:      help,
//	}
//}

// candidate builder pattern code
type CandidateBuilder struct {
	candidate *candidate
}

func NewCandidateBuilder() *CandidateBuilder {
	b := &CandidateBuilder{candidate: &candidate{}}
	b.NeedsHelp(true) // set default.
	return b
}

func (b *CandidateBuilder) Value(value string) *CandidateBuilder {
	b.candidate.value = value
	return b
}

func (b *CandidateBuilder) Raw(raw bool) *CandidateBuilder {
	b.candidate.raw = raw
	return b
}

func (b *CandidateBuilder) Hidden(hidden bool) *CandidateBuilder {
	b.candidate.hidden = hidden
	return b
}

func (b *CandidateBuilder) Continues(continues bool) *CandidateBuilder {
	b.candidate.continues = continues
	return b
}

func (b *CandidateBuilder) Force(force bool) *CandidateBuilder {
	b.candidate.force = force
	return b
}

func (b *CandidateBuilder) Help(help string) *CandidateBuilder {
	b.candidate.help = help
	return b
}

func (b *CandidateBuilder) NeedsHelp(needsHelp bool) *CandidateBuilder {
	b.candidate.needsHelp = needsHelp
	return b
}

func (b *CandidateBuilder) Build() *candidate {
	return b.candidate
}
