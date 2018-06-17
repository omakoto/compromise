package compstore

import (
	"bufio"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compenv"
	"github.com/omakoto/compromise/src/compromise/internal/compdebug"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/utils"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

var (
	cacheLock = &sync.Mutex{}
)

func CacheCandidates(candidates []compromise.Candidate) (e error) {
	if len(compenv.CacheFilename) == 0 {
		return nil
	}
	compdebug.Time("Saving to cache", func() {
		cacheLock.Lock()
		defer cacheLock.Unlock()

		f := compenv.CacheFilename
		err := os.MkdirAll(filepath.Dir(f), 0700)
		if err != nil {
			common.Warnf("unable to create directory for %s", f)
			e = errors.Wrap(err, "Unable to create directory")
			return
		}

		wr, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			e = errors.Wrap(err, "Unable to create file")
			return
		}
		defer wr.Close()

		bwr := bufio.NewWriter(wr)
		defer bwr.Flush()

		bwr.WriteString(strconv.Itoa(len(candidates)))
		bwr.WriteByte(0)
		for _, c := range candidates {
			c.Serialize(bwr)
		}
	})
	if e != nil {
		compdebug.Warnf("Cache save error: %s", e)
	}
	return
}

func LoadCandidates() (result []compromise.Candidate, e error) {
	if len(compenv.CacheFilename) == 0 {
		return nil, nil
	}
	compdebug.Time("Loading from cache", func() {
		cacheLock.Lock()
		defer cacheLock.Unlock()

		rd, err := os.OpenFile(compenv.CacheFilename, os.O_RDONLY, 0)
		if err != nil {
			e = errors.Wrap(err, "Unable to open file")
			return
		}
		defer rd.Close()

		brd := bufio.NewReader(rd)
		line, err := brd.ReadString(0)
		if err == nil {
			line = line[0 : len(line)-1]
			size := utils.ParseInt(line, 10, 0)

			result = make([]compromise.Candidate, 0, size)

			for i := 0; i < size; i++ {
				var c compromise.Candidate
				c, err = compromise.Deserialize(brd)
				if err != nil {
					break
				}
				result = append(result, c)
			}
		}

		if err == io.EOF {
			compdebug.Debugf("%d candidates loaded from cache\n", len(result))
			return
		}
		e = err
		return
	})
	if e != nil {
		compdebug.Warnf("Cache load error: %s", e)
	}
	return
}
