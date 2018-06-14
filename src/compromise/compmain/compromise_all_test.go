package compmain

import (
	"bytes"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/internal/compdebug"
	"github.com/omakoto/compromise/src/compromise/internal/compmisc"
	"github.com/omakoto/go-common/src/shell"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFull(t *testing.T) {
	testdir := "./tests"
	files, err := ioutil.ReadDir(testdir)
	if err != nil {
		t.Fatalf("can't open test file dir: %s", err)
		return
	}

	compmisc.DebugEnabled = true
	os.Setenv("COMPROMISE_SHELL", "tester")

	for _, f := range files {
		file := filepath.Join(testdir, f.Name())
		bindata, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatalf("can't open test file %s: %s", file, err)
			return
		}
		compdebug.Debugf("\n*** TEST %s ***\n", file)

		data := string(bindata)
		splits := strings.SplitN(data, "===\n", 3)

		spec := splits[0]
		commandLine := shell.Split(splits[1])
		expected := splits[2]

		buf := &bytes.Buffer{}

		HandleCompletionRaw(func() string {
			return spec
		}, commandLine, nil, buf)

		result := buf.String()

		compare(t, f.Name(), expected, result)
	}
}

func compare(t *testing.T, test, a, b string) {
	if a == b {
		return
	}
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(b, a, false)

	t.Errorf("* Test %q failed:\n%s\n", test, diffPrettyText(diffs))
}

func diffPrettyText(diffs []diffmatchpatch.Diff) string {
	var buff bytes.Buffer
	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			buff.WriteString("[(+)")
			buff.WriteString(text)
			buff.WriteString("]")
		case diffmatchpatch.DiffDelete:
			buff.WriteString("[(-)")
			buff.WriteString(text)
			buff.WriteString("]")
		case diffmatchpatch.DiffEqual:
			buff.WriteString(text)
		}
	}

	return buff.String()
}

func takeLazily() compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		return nil
	})
}
