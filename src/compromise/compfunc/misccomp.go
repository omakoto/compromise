package compfunc

import (
	"bytes"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compdebug"
	"github.com/omakoto/compromise/src/compromise/comptest"
	"io/ioutil"
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
				ret = append(ret, compromise.NewCandidate().SetValue(string(line)))
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
	return BuildCandidateListFromCommandWithBuilder(command, func(line int, s string, c compromise.Candidate) {
		if mapFunc != nil {
			s = mapFunc(line, s)
		}
		if len(s) > 0 {
			c.SetValue(s)
		}
	})
}

// BuildCandidateListFromCommandWithBuilder executes a command wih /bin/sh and build a CandidateList from the output,
// converting using each line into a single Candidate with mapFunc.
func BuildCandidateListFromCommandWithBuilder(command string, mapFunc func(line int, s string, c compromise.Candidate)) compromise.CandidateList {
	return compromise.LazyCandidates(func(_ string) []compromise.Candidate {
		if mapFunc == nil {
			mapFunc = func(line int, s string, c compromise.Candidate) {
				c.SetValue(s)
			}
		}
		output, _ := ExecAndGetStdout(command)

		return StringsToCandidates(strings.Split(string(output), "\n"), mapFunc)
	})
}

func StringsToCandidates(vals []string, mapFunc func(line int, s string, c compromise.Candidate)) []compromise.Candidate {
	ret := make([]compromise.Candidate, 0)

	for i, v := range vals {
		compdebug.Debugf(" ->%q\n", v)

		c := compromise.NewCandidate()
		mapFunc(i, v, c)
		if len(c.Value()) > 0 {
			ret = append(ret, c)
		}
	}
	return ret
}

func AnyWithHelp(help string) compromise.Candidate {
	return compromise.NewCandidate().SetForce(true).SetHelp(help)
}

func TakeAny(help string) compromise.CandidateList {
	return compromise.OpenCandidates(AnyWithHelp(help))
}

func TakeInteger() compromise.CandidateList {
	return compromise.NewCandidate().SetForce(true).SetHelp("INTEGER")
}

func ExecAndGetStdout(command string) ([]byte, error) {
	return comptest.ExecAndGetStdout(command)
}
