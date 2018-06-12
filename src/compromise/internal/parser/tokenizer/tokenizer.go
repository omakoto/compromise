package tokenizer

import (
	"bytes"
	"github.com/omakoto/compromise-go/src/compromise"
	"github.com/omakoto/compromise-go/src/compromise/compast"
	"github.com/omakoto/go-common/src/textio"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"
)

type identDetector struct {
	inHelp    bool
	tokenizer *Tokenizer
	last      rune
}

// isIdentRune
// Tokens consist of any non-whitespace chars.
// If a token starts with #, the rest of the line is all "help".
// : and @ can be used only as the first character.
// "..." and `...` are allowed, and can be used to circumvent the above restrictions.
// Accepts go-style comments.
func (f *identDetector) isIdentRune(ch rune, i int) bool {
	defer func() {
		f.last = ch
	}()
	if ch < 0 {
		return false
	}
	if f.inHelp {
		if ch == '\n' && f.last != '\\' {
			f.inHelp = false
			return false
		}
		return true
	}
	if i == 0 {
		if ch == '#' {
			f.inHelp = true
			return true
		}
		if ch == '"' || ch == '`' || ch == '/' {
			return false
		}
	}
	if ch == '@' || ch == ':' {
		if i > 0 {
			panic(compromise.NewSpecError(f.tokenizer, "@ and : can only show up as the first character (consider quoting with \"...\")"))
		}
	}
	return !unicode.IsSpace(ch)
}

type Tokenizer struct {
	source        string
	scanner       *scanner.Scanner
	identDetector *identDetector

	current *compast.Token
	peek    *compast.Token
	last    *compast.Token

	lastLineNo  int
	indexInLine int

	directives *compromise.Directives
}

// NewTokenizer creates a new Tokenizer that tokenizes a completion spec string.
func NewTokenizer(source string, d *compromise.Directives) *Tokenizer {
	t := &Tokenizer{source: source, identDetector: &identDetector{}, directives: d}
	t.identDetector.tokenizer = t

	source = textio.ExpandTab(source, t.directives.TabWidth)

	t.scanner = &scanner.Scanner{}
	t.scanner.Init(bytes.NewBufferString(source))
	t.scanner.Mode = scanner.ScanIdents | scanner.ScanComments | scanner.ScanStrings | scanner.ScanRawStrings
	t.scanner.IsIdentRune = t.identDetector.isIdentRune

	return t
}

func (t *Tokenizer) SourceLocation() (string, int, int) {
	return t.directives.Filename, t.scanner.Line + t.directives.StartLine - 1, t.scanner.Column
}

func (t *Tokenizer) CurrentToken() *compast.Token {
	return t.current
}

func (t *Tokenizer) PushBack(tok *compast.Token) {
	if t.peek != nil {
		panic("pushBack called twice")
	}
	t.peek = tok
	t.current = t.last
	t.last = nil
}

var helpLineConcatRe = regexp.MustCompile(`\\\n\s*#\s*`)

// nextToken returns the next "token" in the completion spec string.
func (t *Tokenizer) NextToken() *compast.Token {
	// One token read ahead
	if t.peek != nil {
		r := t.peek
		t.peek = nil

		t.last = t.current
		t.current = r
		return r
	}
	for {
		tok := t.scanner.Scan()
		if tok == scanner.EOF {
			return nil
		}
		if tok != scanner.Comment {
			break
		}
	}
	var err error
	tokenType := compast.TokenLiteral

	rawWord := t.scanner.TokenText()
	word := rawWord
	first := rawWord[0]

	if len(rawWord) == 0 {
		panic("zero-length spec detected.")
	}
	switch first {
	case '@':
		if len(rawWord) == 1 {
			panic(compromise.NewSpecError(t, "missing function or command name after @"))
		}
		word = rawWord[1:]

		tokenType = compast.TokenCommand
	case ':':
		if len(rawWord) == 1 {
			panic(compromise.NewSpecError(t, "missing label name after :"))
		}
		word = rawWord[1:]
		tokenType = compast.TokenLabel
	case '#':
		word = strings.Trim(rawWord[1:], " \t\r\n")
		word = helpLineConcatRe.ReplaceAllString(word, "")
		tokenType = compast.TokenHelp
	case '"', '`':
		word, err = strconv.Unquote(rawWord)
		if err != nil {
			panic(compromise.NewSpecError(t, "invalid string "+rawWord))
		}
	}

	if t.lastLineNo != t.scanner.Line {
		t.indexInLine = 0
	} else {
		t.indexInLine++
	}

	t.last = t.current
	t.current = &compast.Token{
		TokenType:   tokenType,
		RawWord:     rawWord,
		Word:        word,
		Line:        t.scanner.Line + t.directives.StartLine - 1,
		Column:      t.scanner.Column,
		IndexInLine: t.indexInLine,
		SourceFile:  t.directives.Filename,
	}
	t.lastLineNo = t.scanner.Line
	return t.current
}

func (t *Tokenizer) GetNextTokenInLine() *compast.Token {
	if t.CurrentToken() == nil {
		panic("GetNextTokenInLine called when there's no last token")
	}
	line := t.CurrentToken().Line

	tok := t.NextToken()
	if tok == nil {
		return nil
	}
	if tok.Line == line {
		return tok
	}
	t.PushBack(tok)
	return nil
}

func (t *Tokenizer) MustGetNextTokenInLine(expectedType int, error string) *compast.Token {
	tok := t.GetNextTokenInLine()
	if tok == nil {
		panic(compromise.NewSpecError(t, error))
	}
	if expectedType != compast.TokenAny && tok.TokenType != expectedType {
		panic(compromise.NewSpecError(tok, error))
	}
	return tok
}

func (t *Tokenizer) MaybeGetArgsAndHelpToken() (args []*compast.Token) {
	for {
		tok := t.GetNextTokenInLine()
		if tok == nil {
			return
		}
		if tok.TokenType != compast.TokenLiteral && tok.TokenType != compast.TokenHelp {
			panic(compromise.NewSpecError(tok, "Only string literals or a help string (#...) may appear here"))
		}
		args = append(args, tok)
	}
}

func (t *Tokenizer) MaybeGetHelpToken() *compast.Token {
	tok := t.GetNextTokenInLine()
	if tok == nil {
		return nil
	}
	if tok.TokenType != compast.TokenHelp {
		panic(compromise.NewSpecError(tok, "Only a help string (#...) may appear here"))
	}

	return tok
}

func (t *Tokenizer) MaybeGetLabel() *compast.Token {
	tok := t.GetNextTokenInLine()
	if tok == nil {
		return nil
	}
	if tok.TokenType != compast.TokenLabel {
		panic(compromise.NewSpecError(tok, "Only a label (:...) may appear here"))
	}

	return tok
}

func (t *Tokenizer) MaybeGetLiteralAndLabel() (literal *compast.Token, label *compast.Token) {
	tok := t.GetNextTokenInLine()
	if tok == nil {
		return
	}
	if tok.TokenType == compast.TokenLabel {
		t.MustHaveNoTokenInLine()
		label = tok
		return
	}
	if tok.TokenType == compast.TokenLiteral {
		literal = tok
	}
	label = t.MaybeGetLabel()
	t.MustHaveNoTokenInLine()
	return
}

func (t *Tokenizer) MustHaveNoTokenInLine() {
	tok := t.GetNextTokenInLine()
	if tok != nil {
		panic(compromise.NewSpecError(tok, "excessive token detected"))
	}
}
