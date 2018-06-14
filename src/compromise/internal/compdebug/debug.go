package compdebug

// Debug/warning log functions.

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/omakoto/compromise/src/compromise/internal/compmisc"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/textio"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	lock = &sync.Mutex{}
	Out  io.WriteCloser

	indent               = 0
	cachedIndent *string = nil
	lineStart            = true
)

func Indent() {
	lock.Lock()
	defer lock.Unlock()

	indent++
	cachedIndent = nil
}

func Unindent() {
	lock.Lock()
	defer lock.Unlock()

	indent--
	cachedIndent = nil
}

func Dump(msg string, val interface{}) {
	if !compmisc.DebugEnabled {
		return
	}
	write(msg + spew.Sdump(val) + "\n")
}

func Debug(msg string) {
	if !compmisc.DebugEnabled {
		return
	}
	write(msg)
}

func Debugf(format string, args ...interface{}) {
	if !compmisc.DebugEnabled {
		return
	}
	write(fmt.Sprintf(format, args...))
}

func Warn(msg string) {
	write("WARNING: " + msg)
}

func Warnf(format string, args ...interface{}) {
	Warn(fmt.Sprintf(format, args...))
}

func Time(what string, f func()) {
	if compmisc.Time {
		start := time.Now()
		defer func() {
			end := time.Now()

			fmt.Fprintf(os.Stderr, "%s: %s: %v\n", common.MustGetBinName(), what, end.Sub(start))
		}()
	}

	f()
}

func CloseLog() {
	lock.Lock()
	defer lock.Unlock()

	if Out == nil {
		return
	}
	Out.Close()
	Out = nil
}

func write(s string) {
	lock.Lock()
	defer lock.Unlock()

	if Out == nil {
		var err error
		file := compmisc.LogFile
		Out, err = os.Create(file)
		if err != nil {
			common.Warnf("Unable to open \"%s\"", file)
			Out = os.Stderr
		}
	}
	if cachedIndent == nil {
		i := strings.Repeat("  ", indent)
		cachedIndent = &i
	}
	o := ""
	if lineStart {
		o = *cachedIndent
	}
	chomped, lf := textio.StringChomped(s)
	o = o + strings.Replace(chomped, "\n", "\n"+*cachedIndent, -1) + lf
	Out.Write([]byte(o))

	lineStart = len(lf) > 0
}
