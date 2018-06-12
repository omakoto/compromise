package adapters

import (
	"bytes"
	"github.com/omakoto/compromise-go/src/compromise"
	"github.com/omakoto/go-common/src/common"
	"io"
	"os"
	"path/filepath"
	"fmt"
)

const (
	InvokeOption = "compromise-complete"
)

type ShellAdapter interface {
	Install(targetCommandNames []string, hereSpec string)
	HasMenuCompletion() bool

	Escape(arg string) string
	Unescape(arg string) string

	GetCommandLine(args []string) *CommandLine

	StartCompletion(commandLine *CommandLine)
	MaybeOverrideCandidates(commandLine *CommandLine) []compromise.Candidate
	AddCandidate(candidate compromise.Candidate)

	EndCompletion()
	Finish()
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
	}
	common.Fatalf("Unknown shell %s", shell)
	return nil
}
