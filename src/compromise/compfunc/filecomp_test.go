package compfunc

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFileCompFunc(t *testing.T) {
	tests := []struct {
		cwd             string
		prefix          string
		mask            string
		includeFiles    bool
		expectedFiles   []string
		unexpectedFiles []string
	}{
		// Too lazy, just use some "standard" files.
		{"/proc", "", ``, true, []string{"1/", "self/"}, []string{}},
		{"/proc", "self/", ``, true, []string{"self/limits ", "self/maps ", "self/exe ", "self/mem ", "self/cwd/", "self/fd/", "self/fdinfo/"}, []string{}},
		{"/proc", "self/e", ``, true, []string{"self/exe ", "self/environ "}, []string{"self/limits ", "self/maps ", "self/cwd/", "self/fd/", "self/fdinfo/"}},
		{"/proc", "self/en", ``, true, []string{"self/environ "}, []string{"self/exe ", "self/cwd/", "self/fd/", "self/fdinfo/"}},
		{"/proc", "self/", `n$`, true, []string{"self/environ ", "self/cwd/", "self/fd/", "self/fdinfo/"}, []string{"self/exe "}},
		{"/proc", "self/", `s$`, true, []string{"self/limits ", "self/maps ", "self/cwd/", "self/fd/", "self/fdinfo/"}, []string{"self/exe "}},
		{"/proc", "self/", `s$`, false, []string{"self/cwd/", "self/fd/", "self/fdinfo/"}, []string{"self/limits ", "self/maps ", "self/exe "}},

		{"/proc", "./self/", ``, true, []string{"./self/limits ", "./self/maps ", "./self/exe ", "./self/mem ", "./self/cwd/", "./self/fd/", "./self/fdinfo/"}, []string{}},
		{"/proc", "./self/e", ``, true, []string{"./self/exe ", "./self/environ "}, []string{"./self/limits ", "./self/maps ", "./self/cwd/", "./self/fd/", "./self/fdinfo/"}},
		{"/proc", "./self/en", ``, true, []string{"./self/environ "}, []string{"./self/exe ", "./self/cwd/", "./self/fd/", "./self/fdinfo/"}},
		{"/proc", "./self/", `n$`, true, []string{"./self/environ ", "./self/cwd/", "./self/fd/", "./self/fdinfo/"}, []string{"./self/exe "}},
		{"/proc", "./self/", `s$`, true, []string{"./self/limits ", "./self/maps ", "./self/cwd/", "./self/fd/", "./self/fdinfo/"}, []string{"./self/exe "}},
		{"/proc", "./self/", `s$`, false, []string{"./self/cwd/", "./self/fd/", "./self/fdinfo/"}, []string{"./self/limits ", "./self/maps ", "./self/exe "}},

		{"/proc", "././self/", ``, true, []string{"././self/limits ", "././self/maps ", "././self/exe ", "././self/mem ", "././self/cwd/", "././self/fd/", "././self/fdinfo/"}, []string{}},
		{"/proc", "././self/e", ``, true, []string{"././self/exe ", "././self/environ "}, []string{"././self/limits ", "././self/maps ", "././self/cwd/", "././self/fd/", "././self/fdinfo/"}},
		{"/proc", "././self/en", ``, true, []string{"././self/environ "}, []string{"././self/exe ", "././self/cwd/", "././self/fd/", "././self/fdinfo/"}},
		{"/proc", "././self/", `n$`, true, []string{"././self/environ ", "././self/cwd/", "././self/fd/", "././self/fdinfo/"}, []string{"././self/exe "}},
		{"/proc", "././self/", `s$`, true, []string{"././self/limits ", "././self/maps ", "././self/cwd/", "././self/fd/", "././self/fdinfo/"}, []string{"././self/exe "}},
		{"/proc", "././self/", `s$`, false, []string{"././self/cwd/", "././self/fd/", "././self/fdinfo/"}, []string{"././self/limits ", "././self/maps ", "././self/exe "}},

		{"/proc/1/", "../self/", ``, true, []string{"../self/limits ", "../self/maps ", "../self/exe ", "../self/mem ", "../self/cwd/", "../self/fd/", "../self/fdinfo/"}, []string{}},
		{"/proc/1/", "../self/e", ``, true, []string{"../self/exe ", "../self/environ "}, []string{"../self/limits ", "../self/maps ", "../self/cwd/", "../self/fd/", "../self/fdinfo/"}},
		{"/proc/1/", "../self/en", ``, true, []string{"../self/environ "}, []string{"../self/exe ", "../self/cwd/", "../self/fd/", "../self/fdinfo/"}},
		{"/proc/1/", "../self/", `n$`, true, []string{"../self/environ ", "../self/cwd/", "../self/fd/", "../self/fdinfo/"}, []string{"../self/exe "}},
		{"/proc/1/", "../self/", `s$`, true, []string{"../self/limits ", "../self/maps ", "../self/cwd/", "../self/fd/", "../self/fdinfo/"}, []string{"../self/exe "}},
		{"/proc/1/", "../self/", `s$`, false, []string{"../self/cwd/", "../self/fd/", "../self/fdinfo/"}, []string{"../self/limits ", "../self/maps ", "../self/exe "}},
	}
	for i, v := range tests {
		assert.NoError(t, os.Chdir(v.cwd), "chdir")
		res := toStrings(fileCompFunc(v.prefix, v.mask, v.includeFiles, nil))

		for _, e := range v.expectedFiles {
			assert.Contains(t, res, e, "#%d %q vs %q", i, e, res)
		}
		for _, e := range v.unexpectedFiles {
			assert.NotContains(t, res, e, "#%d %q vs %q", i, e, res)
		}
	}
}
