package compmain

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func split(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func TestGetTargetCommands(t *testing.T) {
	tests := []struct {
		original string
		override string
		expected string
	}{
		{"a,b,c", "", "a,b,c"},
		{"a,b,c", "x", "x"},
		{"a,b,c", "x,y", "x,y"},
		{"a,b,c", "-,a", "b,c"},
		{"a,b,c", "-,b", "a,c"},
		{"a,b,c", "-,c", "a,b"},
		{"a,b,c", "-,a,b,c", ""},
	}
	for i, v := range tests {
		assert.Equal(t, v.expected, strings.Join(getTargetCommands(split(v.original), split(v.override)), ","), "#%d", i)
	}
}
