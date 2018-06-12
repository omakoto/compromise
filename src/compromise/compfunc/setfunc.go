package compfunc

import (
	"github.com/omakoto/compromise-go/src/compromise"
	"github.com/omakoto/compromise-go/src/compromise/internal/compdebug"
)

func SetBool(target *bool, value bool) func() {
	return func() {
		*target = value
		compdebug.Debugf("  SetBool set %v to true\n", target)
	}
}

func SetString(target *string, value string) func() {
	return func() {
		*target = value
		compdebug.Debugf("  SetString set %v to %q\n", target, value)
	}
}

func SetLastSeenString(target *string) func(context compromise.CompleteContext) {
	return func(context compromise.CompleteContext) {
		s := context.WordAt(-1)
		*target = s
		compdebug.Debugf("  SetLastSeenString set %v to %q\n", target, s)
	}
}
