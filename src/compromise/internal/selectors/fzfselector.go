package selectors

import (
	"bufio"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compdebug"
	"github.com/omakoto/compromise/src/compromise/compenv"
	"github.com/omakoto/compromise/src/compromise/internal/adapters"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/fileutils"
	"github.com/omakoto/go-common/src/shell"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	opts = append(opts, shell.Split(compenv.FzfOptions)...)
	opts = append(opts, "--with-nth", "2..", "-n", "1..") // Don't show and search the first field.

	// +s: Don't sort, because we candidates are already sorted.
	// +m: No multi selection.
	opts = append(opts, "+s", "+m", "--read0", "--print0", "--ansi")

	opts = append(opts, "-q", prefix)

	if !compenv.FzfFlip {
		opts = append(opts, "--tac")
	}

	// Start FZF.
	starter := func(path string) (*exec.Cmd, io.WriteCloser, io.Reader, error) {
		compdebug.Debugf("Starting fzf at %s...", path)
		cmd := exec.Command(path, opts...)
		cmd.Stderr = os.Stderr
		wr, err := cmd.StdinPipe()
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "StdinPipe failed")
		}
		rd, err := cmd.StdoutPipe()
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "StdoutPipe failed")
		}
		err = cmd.Start()
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "Start failed")
		}
		return cmd, wr, rd, nil
	}

	cmd, wr, rd, err := starter(compenv.FzfBinName)
	if err != nil {
		compdebug.Warnf("Unable to start fzf: %s", err)

		alt := filepath.Join(filepath.Dir(common.MustGetExecutable()), "fzf")
		if fileutils.FileExists(alt) {
			cmd, wr, rd, err = starter(alt)
		}
		if err != nil {
			compdebug.Warnf("Unable to start fzf: %s", err)
			return nil, err
		}
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

	compdebug.Debugf("%d candidates passed to FZF", len(candidates))

	defer func() {
		cmd.Wait()
	}()

	// Read the input
	brd := bufio.NewReader(rd)
	res, err := brd.ReadString(0)
	if err == io.EOF {
		return nil, nil // Canceled.
	} else if err != nil {
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
