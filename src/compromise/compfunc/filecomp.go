package compfunc

import (
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compdebug"
	"github.com/omakoto/compromise/src/compromise/comptest"
	"github.com/omakoto/go-common/src/utils"
	"github.com/ungerik/go-dry"
	"os"
	"path"
	"regexp"
)

func init() {
	Register("TakeFile", TakeFile)
	Register("TakeDir", TakeDir)
}

func TakeFile(reFilenameMatcher string) compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		return fileCompFunc(prefix, reFilenameMatcher, true, nil)
	})
}

func TakeFileWithMapper(reFilenameMatcher string, mapper func(builder compromise.Candidate)) compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		return fileCompFunc(prefix, reFilenameMatcher, true, mapper)
	})
}

func TakeDir() compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		return fileCompFunc(prefix, "", false, nil)
	})
}

func fileCompFunc(prefix, reFilenameMatcher string, includeFiles bool, mapper func(builder compromise.Candidate)) []compromise.Candidate {
	prefixDir, prefixFile := path.Split(prefix)
	filenameRegexp := regexp.MustCompile(reFilenameMatcher)

	compdebug.Debugf("fileCompFunc: %q (%q + %q), %q, %v\n", prefix, prefixDir, prefixFile, reFilenameMatcher, includeFiles)

	files, err := os.ReadDir(utils.FirstNonEmpty(prefixDir, "."))
	if err != nil {
		compdebug.Debugf("Unable to read directory \"%s\": %s\n", prefixDir, err)
		return nil
	}

	conv := func(b compromise.Candidate) compromise.Candidate {
		if mapper != nil {
			mapper(b)
		}
		return b
	}

	ret := make([]compromise.Candidate, 0)

	for _, file := range files {
		baseName := file.Name()
		relPath := prefixDir + baseName
		isDir := dry.FileIsDir(relPath)

		compdebug.Debugf("  - %s [isdir=%v]\n", relPath, isDir)

		if !compromise.StringMatches(baseName, prefixFile) {
			continue
		}
		compdebug.Debug("      [prefix match]\n")

		if isDir {
			ret = append(ret, conv(compromise.NewCandidate().SetValue(relPath+"/").SetContinues(!comptest.IsEmptyDir(relPath))))
			continue
		}
		if includeFiles && len(filenameRegexp.FindStringIndex(baseName)) > 0 {
			ret = append(ret, conv(compromise.NewCandidate().SetValue(relPath)))
		}
	}
	compdebug.Dump("Filecopletion result=", ret)
	return ret
}
