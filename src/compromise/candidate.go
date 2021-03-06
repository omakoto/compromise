package compromise

import (
	"bufio"
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

	Serialize(wr *bufio.Writer)

	Deserialize(rd *bufio.Reader) error

	SetValue(value string) Candidate
	SetRaw(raw bool) Candidate
	SetHidden(hidden bool) Candidate
	SetContinues(continues bool) Candidate
	SetForce(force bool) Candidate
	SetHelp(help string) Candidate
}

type candidate struct {
	value     string
	raw       bool
	hidden    bool
	continues bool
	force     bool
	help      string
}

var _ Candidate = (*candidate)(nil)
var _ CandidateList = (*candidate)(nil)
var _ fmt.Stringer = (*candidate)(nil)

const (
	raw = 1 << iota
	hidden
	continues
	force
	help
)

func NewCandidate() Candidate {
	return &candidate{}
}

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

func (c *candidate) SetValue(value string) Candidate {
	c.value = value
	return c
}

func (c *candidate) SetRaw(raw bool) Candidate {
	c.raw = raw
	return c
}

func (c *candidate) SetHidden(hidden bool) Candidate {
	c.hidden = hidden
	return c
}

func (c *candidate) SetContinues(continues bool) Candidate {
	c.continues = continues
	return c
}

func (c *candidate) SetForce(force bool) Candidate {
	c.force = force
	return c
}

func (c *candidate) SetHelp(help string) Candidate {
	c.help = help
	return c
}

func (c *candidate) Matches(prefix string) bool {
	return c.Force() || StringMatches(c.value, prefix)
}

func (c *candidate) MatchesFully(target string) bool {
	return c.Force() || (c.value == target)
}

func (c *candidate) GetCandidate(prefix string) []Candidate {
	if c.Matches(prefix) {
		return []Candidate{c}
	}
	return nil
}

func (c *candidate) Serialize(wr *bufio.Writer) {
	wr.WriteString(c.value)
	wr.WriteByte(0)
	wr.WriteString(c.help)
	wr.WriteByte(0)

	var v byte
	if c.raw {
		v |= raw
	}
	if c.hidden {
		v |= hidden
	}
	if c.continues {
		v |= continues
	}
	if c.force {
		v |= force
	}
	wr.WriteByte(v)
}

func (c *candidate) Deserialize(rd *bufio.Reader) error {
	s, _ := rd.ReadString(0)
	c.value = s[0 : len(s)-1]

	s, _ = rd.ReadString(0)
	c.help = s[0 : len(s)-1]

	v, err := rd.ReadByte()
	if (v & raw) != 0 {
		c.raw = true
	}
	if (v & hidden) != 0 {
		c.hidden = true
	}
	if (v & continues) != 0 {
		c.continues = true
	}
	if (v & force) != 0 {
		c.force = true
	}
	return err
}

func Deserialize(rd *bufio.Reader) (Candidate, error) {
	ret := candidate{}
	err := ret.Deserialize(rd)
	return &ret, err
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
