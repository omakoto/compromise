package compromise

// CompleteContext is a context for complete functions.
type CompleteContext interface {
	// Command returns the unescaped target executable command name.
	Command() string
	// RawCommand returns the raw target executable command name.
	RawCommand() string

	// WordAtCursor returns the unescaped word at cursor.
	WordAtCursor(offset int) string
	// RawWordAtCursor returns the raw word at cursor.
	RawWordAtCursor(offset int) string

	// WordAt returns the unescaped word at pc.
	WordAt(offset int) string
	// RawWordAt returns the raw word at pc.
	RawWordAt(offset int) string

	// BeforeCursor returns whether pc is after the cursor index.
	BeforeCursor() bool
	// AfterCursor returns whether pc is after the cursor index.
	AfterCursor() bool
	// AtCursor returns whether pc is equal to the cursor index.
	AtCursor() bool
}
