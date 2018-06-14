package compromise

import (
	"encoding/json"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/utils"
	"runtime"
	"strings"
)

type Directives struct {
	TabWidth  int    `json:"tab"`  // Tabs in a spec is assumed to be this many spaces.
	StartLine int    `json:"line"` // USed to override the number of a spec string
	Filename  string `json:"file"` // Filename where a spec is defined
}

func NewDirectives() *Directives {
	d := &Directives{}
	d.TabWidth = 8
	d.StartLine = 1
	return d
}

func (d *Directives) SetSourceLocation() *Directives {
	if _, file, line, ok := runtime.Caller(1); ok {
		d.Filename = file
		d.StartLine = line
	}
	return d
}

func (d *Directives) SetFilename(filename string) *Directives {
	d.Filename = filename
	return d
}

func (d *Directives) SetStartLine(lineNumber int) *Directives {
	d.StartLine = lineNumber
	return d
}

func (d *Directives) Tab(tabWidth int) *Directives {
	d.TabWidth = tabWidth
	return d
}

func (d *Directives) Json() string {
	buffer, err := json.Marshal(d)
	common.CheckPanic(err, "json.Marshal failed.")
	return string(buffer)
}

func (d *Directives) UnmarshalJson(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), d)
}

// ExtractDirectives options in a form of json from the first line of the spec string.
func ExtractDirectives(spec string) *Directives {
	ret := NewDirectives()
	if !strings.HasPrefix(spec, "//{") {
		return ret
	}
	directiveJson := spec[2:utils.IndexByteOrLen(spec, '\n')]
	err := ret.UnmarshalJson(directiveJson)
	if err != nil {
		panic(NewSpecErrorf(nil, "invalid parser directive in line 1 %s: %s", directiveJson, err.Error()))
	}
	return ret
}
