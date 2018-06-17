package compromise

import (
	"github.com/omakoto/compromise/src/compromise/compenv"
	"github.com/ungerik/go-dry"
)

func Beep() {
	// TODO Implement it
	dry.Nop(compenv.BellStyle)
}
