package compmain

import (
	"bytes"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compfunc"
	"github.com/omakoto/compromise/src/compromise/internal/compdebug"
	"github.com/omakoto/compromise/src/compromise/internal/compmisc"
	"github.com/omakoto/go-common/src/shell"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"regexp"
)

var (
	stringHolder string
)

func takeLazily(args []string) compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		ret := make([]compromise.Candidate, 0)
		for _, a := range args {
			ret = append(ret, compromise.NewCandidateBuilder().Value(a).Build())
		}
		return ret
	})
}

func takeStatically(args []string) compromise.CandidateList {
	ret := make([]compromise.Candidate, 0)
	for _, a := range args {
		ret = append(ret, compromise.NewCandidateBuilder().Value(a).Build())
	}
	return compromise.StrictCandidates(ret...)
}

func takeHeldValue() compromise.CandidateList {
	ret := make([]compromise.Candidate, 0)
	ret = append(ret, compromise.NewCandidateBuilder().Value(stringHolder).Build())
	return compromise.StrictCandidates(ret...)
}

func init() {
	compmisc.DebugEnabled = true
	compmisc.LogFile = "/tmp/compromise-test.log"
	compdebug.CloseLog()
	os.Setenv("COMPROMISE_SHELL", "tester")

	compfunc.Register("takeLazily", takeLazily)
	compfunc.Register("takeStatically", takeStatically)

	compfunc.Register("setA", compfunc.SetString(&stringHolder, "A"))
	compfunc.Register("setB", compfunc.SetString(&stringHolder, "B"))
	compfunc.Register("setCurrent", compfunc.SetLastSeenString(&stringHolder))

	compfunc.Register("takeHeldValue", takeHeldValue)
}

func TestFull(t *testing.T) {
	testdir := "./tests"
	files, err := ioutil.ReadDir(testdir)
	if err != nil {
		t.Fatalf("can't open test file dir: %s", err)
		return
	}

	commentsStripper := regexp.MustCompile(`(\n|^)//.*?\n`)
	for _, f := range files {
		if !f.Mode().IsRegular() {
			continue
		}
		file := filepath.Join(testdir, f.Name())
		bindata, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatalf("can't open test file %s: %s", file, err)
			return
		}
		compdebug.Debugf("\n*** TEST %s ***\n", file)

		data := string(bindata)
		data = commentsStripper.ReplaceAllString(data, "\n")

		data = strings.TrimRight(data, " \t\n") + "\n"
		splits := strings.SplitN(data, "===\n", 3)

		if len(splits) != 3 {
			t.Fatalf("Invalid test file format in file %q", file)
		}

		spec := splits[0]
		commandLine := shell.Split(splits[1])
		expected := splits[2]

		buf := &bytes.Buffer{}

		HandleCompletionRaw(func() string {
			return "//" + compromise.NewDirectives().SetFilename(file).SetStartLine(0).Json() + "\n" + spec
		}, commandLine, nil, buf)

		result := buf.String()

		compare(t, file, expected, result)
	}
}

func compare(t *testing.T, test, a, b string) {
	if a == b {
		return
	}
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(b, a, false)

	t.Errorf("* Test failed\nFile %s:1\n%s\n", test, diffPrettyText(diffs))
}

func diffPrettyText(diffs []diffmatchpatch.Diff) string {
	var buff bytes.Buffer
	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			buff.WriteString("\x1b[32m[(+)")
			buff.WriteString(text)
			buff.WriteString("]\x1b[0m")
		case diffmatchpatch.DiffDelete:
			buff.WriteString("\x1b[31m[(-)")
			buff.WriteString(text)
			buff.WriteString("]\x1b[0m")
		case diffmatchpatch.DiffEqual:
			buff.WriteString(text)
		}
	}

	return buff.String()
}

func TestBad(t *testing.T) {
	testdir := "./bad"
	files, err := ioutil.ReadDir(testdir)
	if err != nil {
		t.Fatalf("can't open test file dir: %s", err)
		return
	}

	for _, f := range files {
		if !f.Mode().IsRegular() {
			continue
		}
		file := filepath.Join(testdir, f.Name())
		bindata, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatalf("can't open test file %s: %s", file, err)
			return
		}
		compdebug.Debugf("\n*** TEST %s ***\n", file)

		buf := &bytes.Buffer{}
		assert.Panics(t, func() {
			HandleCompletionRaw(func() string {
				return string(bindata)
			}, []string{"dummy"}, nil, buf)
		}, "File %s:1", file)
	}
}
