package selectors

import (
	"bufio"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compmisc"
	"github.com/omakoto/compromise/src/compromise/internal/adapters"
	"github.com/omakoto/compromise/src/compromise/internal/compdebug"
	"github.com/omakoto/go-common/src/shell"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type fzfSelector struct {
}

var _ Selector = (*fzfSelector)(nil)

func NewFzfSelector() Selector {
	return &fzfSelector{}
}

func (s *fzfSelector) Select(prefix string, candidates []compromise.Candidate) (compromise.Candidate, error) {
	opts := make([]string, 0)
	opts = append(opts, shell.Split(compmisc.FzfOptions)...)
	opts = append(opts, "--with-nth", "2..", "-n", "1..") // Don't show and search the first field.

	// +s: Don't sort, because we candidates are already sorted.
	// +m: No multi selection.
	opts = append(opts, "+s", "+m", "--read0", "--print0", "--ansi")

	opts = append(opts, "-q", prefix)

	if !compmisc.FzfFlip {
		opts = append(opts, "--tac")
	}

	// Start FZF.
	cmd := exec.Command(compmisc.FzfBinName, opts...)
	cmd.Stderr = os.Stderr
	wr, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "StdinPipe failed")
	}
	rd, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "StdoutPipe failed")
	}
	err = cmd.Start()
	if err != nil {
		return nil, errors.Wrap(err, "Start failed")
	}

	// Pass data to FZF.
	bwr := bufio.NewWriter(wr)
	for i, c := range candidates {
		bwr.WriteString(strconv.Itoa(i))
		bwr.WriteString(" ")
		adapters.AddDisplayString(c, bwr)
		bwr.WriteByte(0)
	}
	bwr.Flush()
	wr.Close()

	compdebug.Debugf("%d candidates passed to FZF")

	defer func() {
		cmd.Wait()
	}()

	// Read the input
	brd := bufio.NewReader(rd)
	res, err := brd.ReadString(0)
	if err != nil {
		return nil, errors.Wrap(err, "ReadString failed")
	}
	compdebug.Debugf("Result from FZF: %s\n", res)

	splits := strings.SplitN(res, " ", 2)
	if len(splits[0]) > 0 {
		if index, ok := strconv.Atoi(splits[0]); ok == nil && 0 <= index && index < len(candidates) {
			selected := candidates[index]
			compdebug.Debugf("Selected: %v\n", selected)
			return selected, nil
		}
	}

	return nil, nil
}
