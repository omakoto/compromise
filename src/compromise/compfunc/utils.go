package compfunc

import (
	"github.com/omakoto/compromise-go/src/compromise"
	"strings"
)

const Space = " \t\r\n\v"

func TrimSpace(s string) string {
	return strings.Trim(s, Space)
}

func FieldAt(s string, index int, ignoreLeadingSpace bool) string {
	if ignoreLeadingSpace {
		s = strings.TrimLeft(s, Space)
	}
	return GetField(strings.Fields(s), index)
}

func FieldAtWithSeparator(s string, index int, separator string, ignoreLeadingSpace bool) string {
	if ignoreLeadingSpace {
		s = strings.TrimLeft(s, Space)
	}
	return GetField(strings.Split(s, separator), index)
}

func GetField(fields []string, index int) string {
	if index < len(fields) {
		return fields[index]
	}
	return ""
}

func toStrings(ar []compromise.Candidate) []string {
	ret := make([]string, 0, len(ar))
	for _, a := range ar {
		v := a.Value()
		if !a.Continues() {
			v += " "
		}
		ret = append(ret, v)
	}
	return ret
}
