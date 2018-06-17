package compstore

import (
	"encoding/json"
	"github.com/omakoto/compromise/src/compromise/compenv"
	"github.com/omakoto/compromise/src/compromise/internal/compdebug"
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/go-common/src/utils"
	"github.com/ungerik/go-dry"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"
)

type Store struct {
	LastCommandLine           []string
	LastCursorIndex           int
	LastCompletionTime        time.Time
	CurrentCompletionTime     time.Time
	LastPwd                   string
	NumConsecutiveInvocations int
	IsDoublePress             bool
}

var (
	lock  = &sync.Mutex{}
	s     *Store
	clock = utils.NewClock()
)

func ensureLoadedLocked() {
	if s != nil {
		return
	}
	s = &Store{}

	f := compenv.StoreFilename
	if dry.FileExists(f) {
		data, err := dry.FileGetBytes(f)
		if err != nil {
			common.Warnf("unable to load %s", f)
			return
		}

		err = json.Unmarshal(data, s)
		if err != nil {
			common.Warnf("unable to parse %s", f)
			return
		}
	}
}

func saveLocked() {
	if s == nil {
		return
	}

	f := compenv.StoreFilename
	err := os.MkdirAll(filepath.Dir(f), 0700)
	if err != nil {
		common.Warnf("unable to create directory for %s", f)
		return
	}
	data, err := json.MarshalIndent(s, "", "  ")
	common.CheckPanice(err)

	err = ioutil.WriteFile(f, data, 0600)
	if err != nil {
		common.Warnf("unable to save %s", f)
		return
	}
}

func Load() *Store {
	lock.Lock()
	defer lock.Unlock()

	ensureLoadedLocked()
	return s
}

func UpdateForInvocation(commandLine []string, cursorIndex int) *Store {
	lock.Lock()
	defer lock.Unlock()

	ensureLoadedLocked()

	pwd, _ := os.Getwd()
	now := clock.Now()

	if s.LastPwd == pwd && s.LastCursorIndex == cursorIndex && reflect.DeepEqual(s.LastCommandLine, commandLine) {
		s.NumConsecutiveInvocations++
	} else {
		s.LastCursorIndex = cursorIndex
		s.LastCommandLine = commandLine
		s.NumConsecutiveInvocations = 1
	}

	s.IsDoublePress = s.NumConsecutiveInvocations > 1 && s.LastCompletionAge() <= compenv.DoublePressTimeout

	s.LastCompletionTime = s.CurrentCompletionTime
	s.LastPwd = pwd
	s.CurrentCompletionTime = now

	compdebug.Dump("Store updated", s)

	saveLocked()

	return s
}

func (s *Store) LastCompletionAge() time.Duration {
	return clock.Now().Sub(s.LastCompletionTime)
}
