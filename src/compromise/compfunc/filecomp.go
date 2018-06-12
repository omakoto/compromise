package compfunc

import (
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/internal/compdebug"
	"github.com/omakoto/go-common/src/utils"
	"github.com/ungerik/go-dry"
	"io/ioutil"
	"path"
	"regexp"
)

func init() {
	Register("TakeFile", TakeFile)
	Register("TakeDir", TakeDir)
}

func TakeFile(reFilenameMatcher string) compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		return fileCompFunc(prefix, reFilenameMatcher, true)
	})
}

func TakeDir() compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		return fileCompFunc(prefix, "", false)
	})
}

func IsEmptyDir(path string) bool {
	files, _ := ioutil.ReadDir(path + "/")
	return len(files) == 0
}

func fileCompFunc(prefix, reFilenameMatcher string, includeFiles bool) []compromise.Candidate {
	prefixDir, prefixFile := path.Split(prefix)
	filenameRegexp := regexp.MustCompile(reFilenameMatcher)

	compdebug.Debugf("fileCompFunc: %q (%q + %q), %q, %v\n", prefix, prefixDir, prefixFile, reFilenameMatcher, includeFiles)

	files, err := ioutil.ReadDir(utils.FirstNonEmpty(prefixDir, "."))
	if err != nil {
		compdebug.Debugf("Unable to read directory \"%s\": %s\n", prefixDir, err)
		return nil
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
			ret = append(ret,
				compromise.NewCandidateBuilder().
					Value(relPath+"/").Continues(!IsEmptyDir(relPath)).NeedsHelp(false).Build())
			continue
		}
		if includeFiles && len(filenameRegexp.FindStringIndex(baseName)) > 0 {
			ret = append(ret, compromise.NewCandidateBuilder().Value(relPath).NeedsHelp(false).Build())
		}
	}
	compdebug.Dump("Filecopletion result=", ret)
	return ret
}
