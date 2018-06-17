package compromise

import (
	"github.com/omakoto/compromise/src/compromise/compenv"
	"strings"
)

var (
	converter = func(s string) string { return s }
)

func setConverter(f func(s string) string) {
	prev := converter
	converter = func(s string) string {
		return f(prev(s))
	}
}

func hyphenMapper(s string) string {
	return strings.Replace(s, "-", "_", -1)
}

func init() {
	if compenv.IgnoreCase {
		setConverter(strings.ToLower)
	}
	if compenv.MapCase {
		setConverter(hyphenMapper)
	}
}

func StringMatches(s, prefix string) bool {
	return strings.HasPrefix(converter(s), converter(prefix))
}
