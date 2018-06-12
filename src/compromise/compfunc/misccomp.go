package compfunc

import (
	"bytes"
	"github.com/omakoto/compromise-go/src/compromise"
	"github.com/omakoto/compromise-go/src/compromise/internal/compdebug"
	"github.com/omakoto/go-common/src/common"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func init() {
	Register("TakeAny", TakeAny)
	Register("TakeInteger", TakeInteger)
}

// ReadLinesFromFile reads a file returns the lines in it. Lines starting with # will be ignored.
func ReadLinesFromFile(filename string) []string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}

	ret := make([]string, 0)
	for _, line := range bytes.Split(content, []byte("\n")) {
		l := bytes.Trim(line, " \t\r")
		if len(l) > 0 && l[0] != '#' {
			ret = append(ret, string(l))
		}
	}
	return ret
}

// ReadCandidatesFromFile reads a file and build a CandidateList from it, using each line as a
// single Candidate. Lines starting with # will be ignored.
func ReadCandidateListFromFile(filename string) compromise.CandidateList {
	return compromise.LazyCandidates(func(_ string) []compromise.Candidate {
		ret := make([]compromise.Candidate, 0)

		for _, line := range ReadLinesFromFile(filename) {
			line = strings.Trim(line, " \t\r")
			if len(line) > 0 && line[0] != '#' {
				ret = append(ret, compromise.NewCandidateBuilder().Value(string(line)).Build())
			}
		}
		return ret
	})
}

// BuildCandidateListFromCommand executes a command wih /bin/sh and build a CandidateList from the output,
// using each line as a single Candidate.
func BuildCandidateListFromCommand(command string) compromise.CandidateList {
	return BuildCandidateListFromCommandWithMap(command, nil)
}

// BuildCandidateListFromCommandWithMap executes a command wih /bin/sh and build a CandidateList from the output,
// using each line as a single Candidate. If mapFunc is given, it'll be applied to each line.
func BuildCandidateListFromCommandWithMap(command string, mapFunc func(line int, s string) string) compromise.CandidateList {
	return BuildCandidateListFromCommandWithBuilder(command, func(line int, s string, b *compromise.CandidateBuilder) {
		s = mapFunc(line, s)
		if len(s) > 0 {
			b.Value(s)
		}
	})
}

// BuildCandidateListFromCommandWithBuilder executes a command wih /bin/sh and build a CandidateList from the output,
// converting using each line into a single Candidate with mapFunc.
func BuildCandidateListFromCommandWithBuilder(command string, mapFunc func(line int, s string, b *compromise.CandidateBuilder)) compromise.CandidateList {
	return compromise.LazyCandidates(func(_ string) []compromise.Candidate {
		if mapFunc == nil {
			mapFunc = func(line int, s string, b *compromise.CandidateBuilder) {
				b.Value(s)
			}
		}
		compdebug.Debugf("Executing: %q\n", command)

		cmd := exec.Command("/bin/sh", "-c", command)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()

		if err != nil {
			common.Warnf("Command execution error: command=%q error=%s", command, err)
		}

		ret := make([]compromise.Candidate, 0)

		if output != nil {
			for i, line := range bytes.Split(output, []byte("\n")) {
				compdebug.Debugf(" ->%q\n", line)

				b := compromise.NewCandidateBuilder()
				mapFunc(i, string(line), b)
				c := b.Build()
				if len(c.Value()) > 0 {
					ret = append(ret, c)
				}
			}
		}

		return ret
	})
}

func AnyWithHelp(help string) compromise.Candidate {
	return compromise.NewCandidateBuilder().Force(true).Help(help).Build()
}

func TakeAny(help string) compromise.CandidateList {
	return compromise.OpenCandidates(AnyWithHelp(help))
}

func TakeInteger(ctx compromise.CompleteContext) compromise.CandidateList {
	return compromise.OpenCandidates()
}
