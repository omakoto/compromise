package adapters

import (
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compenv"
	"github.com/omakoto/go-common/src/common"
)

func getUniqueName(command string) string {
	return toShellSafeName(common.MustGetBinName()) + "_" + toShellSafeName(command)
}

func getFuncName(command string) string {
	return "__compromise_" + getUniqueName(command) + "_completion"
}

type stringWriter interface {
	WriteString(s string) (n int, err error)
}

func AddDisplayString(c compromise.Candidate, bwr stringWriter) {
	if len(c.Value()) == 0 {
		bwr.WriteString("<ANY>")
	} else {
		bwr.WriteString(c.Value())
	}
	if len(c.Help()) > 0 {
		if compenv.UseColor {
			bwr.WriteString(compenv.HelpStartEscape)
		}
		bwr.WriteString(" : ")
		bwr.WriteString(c.Help())
		if compenv.UseColor {
			bwr.WriteString(compenv.HelpEndEscape)
		}
	}
}
