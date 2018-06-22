package adapters

import (
	"bytes"
	"fmt"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/go-common/src/common"
	"io"
	"os"
	"path/filepath"
)

const (
	InvokeOption = "compromise-complete"
)

type ShellAdapter interface {
	Install(targetCommandNames []string, spec string)
	HasMenuCompletion() bool
	UseFzf() bool

	Escape(arg string) string
	Unescape(arg string) string

	GetCommandLine(args []string) *CommandLine

	StartCompletion(commandLine *CommandLine)
	MaybeOverrideCandidates(commandLine *CommandLine) []compromise.Candidate
	AddCandidate(candidate compromise.Candidate)

	EndCompletion()
	Finish()

	DefaultMaxCandidates() int
	DefaultMaxHelps() int
}

func toShellSafeName(commandName string) string {
	ret := bytes.NewBuffer(nil)
	for _, ch := range commandName {
		if ('0' <= ch && ch <= '9') ||
			('a' <= ch && ch <= 'z') ||
			('A' <= ch && ch <= 'Z') {
			ret.WriteRune(ch)
			continue
		}
		ret.WriteString(fmt.Sprintf("%%%04x", ch))
	}
	return ret.String()
}

func GetShellAdapter(rd io.Reader, wr io.Writer) ShellAdapter {
	shell := os.Getenv("COMPROMISE_SHELL")
	if shell == "" {
		shell = common.MustGetenv("SHELL")
	}
	switch filepath.Base(shell) {
	case "bash":
		return newBashAdapter(rd, wr)
	case "zsh":
		return newZshAdapter(rd, wr)
	case "tester":
		return newTesterAdapter(rd, wr)
	}
	common.Fatalf("Unknown shell %s", shell)
	return nil
}
