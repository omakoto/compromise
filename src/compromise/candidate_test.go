package compromise

//import (
//	"github.com/stretchr/testify/assert"
//	"testing"
//)
//
//func TestParseCandidate(t *testing.T) {
//	tests := []struct {
//		input string
//
//		value     string
//		raw       bool
//		hidden    bool
//		continues bool
//		force     bool
//		help      string
//	}{
//		{"-", "-", false, false, false, false, ""},
//		{"-f", "-f", false, false, false, false, ""},
//		{"--flags", "--flags", false, false, false, false, ""},
//
//		{"--flags #HELP", "--flags", false, false, false, false, "HELP"},
//		{"--flags#HELP", "--flags#HELP", false, false, false, false, ""},
//		{":--flags  #  HELP", "--flags", false, false, false, false, "HELP"},
//		{"r:--flags  #  HELP", "--flags", true, false, false, false, "HELP"},
//		{"r: --flags  #  HELP", "--flags", true, false, false, false, "HELP"},
//		{"h:--flags  #  HELP", "--flags", false, true, false, false, "HELP"},
//		{"c:--flags  #  HELP", "--flags", false, false, true, false, "HELP"},
//		{"f:--flags  #  HELP", "--flags", false, false, false, true, "HELP"},
//		{"rhcf:-f", "-f", true, true, true, true, ""},
//	}
//	for _, v := range tests {
//		c := parseCandidate(v.input)
//		assert.Equal(t, v.value, c.Value(), v.input)
//		assert.Equal(t, v.raw, c.Raw(), v.input)
//		assert.Equal(t, v.hidden, c.Hidden(), v.input)
//		assert.Equal(t, v.continues, c.Continues(), v.input)
//		assert.Equal(t, v.force, c.Force(), v.input)
//		assert.Equal(t, v.help, c.Help(), v.input)
//	}
//}
//
//func TestCPanic(t *testing.T) {
//	tests := []struct {
//		input         string
//		expectedPanic string
//	}{
//		{"", "Spec \"\" doesn't contain value"},
//		{"\nx", "Unable to parse spec \"\nx\""},
//	}
//	for _, v := range tests {
//		assert.PanicsWithValue(t, v.expectedPanic, func() { C(v.input) }, v.input)
//	}
//}
