package adapters

import (
	"github.com/omakoto/compromise-go/src/compromise"
	"github.com/omakoto/go-common/src/shell"
	"github.com/omakoto/go-common/src/utils"
	"github.com/ungerik/go-dry"
)

// CommandLine holds the words in the command line and other contextual information.
type CommandLine struct {
	unescape func(string) string // Unescaping function

	rawWords []string // Original words from the command line.
	words    []string // Unescaped.

	// Index at which the cursor is, when completion started.
	cursorIndex int

	// pc stands for "program counter" -- index of the word that we're now looking at
	// when executing completion.
	pc int

	// Bash specific variables. We keep them here mostly so they'll be dumped in the debug log.
	bashCompCword         int      // Index given by readline as COMP_CWORD
	bashCompWords         []string // Words given by readline as COMP_WORDS (split up with COMP_WORDBREAKS)
	bashCompCurrentWord   string   // COMP_WORDS[COMP_CWORD]
	bashCompPoint         int
	bashCompLine          string
	bashCompType          string
	bashCompWordbreaks    string
	bashParsedRawWords    []shell.Token
	bashParsedWords       []string
	bashDeltaFromReadline string
}

var _ compromise.CompleteContext = (*CommandLine)(nil)

func newCommandLine(unescape func(string) string, cursorIndex int, rawWords []string) *CommandLine {
	return (&CommandLine{unescape: unescape}).Replace(cursorIndex, rawWords)
}

func (c *CommandLine) Replace(cursorIndex int, rawWords []string) *CommandLine {
	c.cursorIndex = cursorIndex
	c.rawWords = rawWords
	c.words = dry.StringMap(c.unescape, rawWords)
	return c
}

// CursorIndex returns the index of the word at the cursor.
func (c *CommandLine) CursorIndex() int {
	return c.cursorIndex
}

// Return raw words.
func (c *CommandLine) RawWords() []string {
	return c.rawWords
}

// Pc returns the current pc (program counter).
func (c *CommandLine) Pc() int {
	return c.pc
}

// SetPc set pc to n.
func (c *CommandLine) SetPc(n int) {
	c.pc = utils.Clip(n, 0, c.WordLen()+1)
}

// AdvancePc increments pc by n.
func (c *CommandLine) AdvancePc(n int) {
	c.pc += n
}

// AtCursor returns whether pc is equal to the cursor index.
func (c *CommandLine) AtCursor() bool {
	return c.pc == c.cursorIndex
}

// AfterCursor returns whether pc is after the cursor index.
func (c *CommandLine) AfterCursor() bool {
	return c.pc > c.cursorIndex
}

// BeforeCursor returns whether pc is before the cursor index.
func (c *CommandLine) BeforeCursor() bool {
	return c.pc < c.cursorIndex
}

// WordLen returns the number of words in the command line.
func (c *CommandLine) WordLen() int {
	return c.cursorIndex + 1
}

// WordAtIndex returns the unescaped word at a given index (0-based).
func (c *CommandLine) WordAtIndex(i int) string {
	if i < 0 || i >= len(c.words) {
		return ""
	}
	if c.AfterCursor() {
		return ""
	}
	return c.words[i]
}

// RawWordAtIndex returns the raw word at a given index (0-based).
func (c *CommandLine) RawWordAtIndex(i int) string {
	if i < 0 || i >= len(c.words) {
		return ""
	}
	if c.AfterCursor() {
		return ""
	}
	return c.rawWords[i]
}

// WordAtCursor returns the unescaped word at cursor.
func (c *CommandLine) WordAtCursor(offset int) string {
	return c.WordAtIndex(c.cursorIndex + offset)
}

// RawWordAtCursor returns the raw word at cursor.
func (c *CommandLine) RawWordAtCursor(offset int) string {
	return c.RawWordAtIndex(c.cursorIndex + offset)
}

// WordAt returns the unescaped word at pc.
func (c *CommandLine) WordAt(offset int) string {
	return c.WordAtIndex(c.pc + offset)
}

// RawWordAt returns the raw word at pc.
func (c *CommandLine) RawWordAt(offset int) string {
	return c.RawWordAtIndex(c.pc + offset)
}

// Command returns the unescaped target executable command name.
func (c *CommandLine) Command() string {
	return c.WordAtIndex(0)
}

// RawCommand returns the raw target executable command name.
func (c *CommandLine) RawCommand() string {
	return c.WordAtIndex(0)
}
